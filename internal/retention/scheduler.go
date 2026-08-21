package retention

import (
	"context"
	"sync"
	"time"

	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/syssetting"
)

const (
	// defaultTick 调度循环醒来间隔：争抢/续期租约并读取归档任务配置。
	defaultTick = time.Minute
	// defaultInterval 未配置开始时刻且间隔非法时的 Job 执行间隔兜底值。
	defaultInterval = 60 * time.Minute
	// defaultLeaseTTL 清理租约有效期。持有节点按 tick 续期，过期后其他节点可抢占。
	defaultLeaseTTL = 3 * time.Minute
)

// Scheduler 租户级数据生命周期调度器，随进程启动并在关闭时停止。
// 各节点都运行循环，但通过 HUB_RETENTION_LEASE 保证同一时刻通常只有一个节点执行 Job。
// 循环按 tick 醒来；是否真正清理由环境设置 retentionJob 决定（启停、间隔、每天开始时刻）。
// lastRun 仅保存在本进程内存，进程重启后会丢失。关闭「启用归档」只跳过 Job，不停止本调度器。
type Scheduler struct {
	db       database.Database
	owner    string
	jobs     []Job
	tick     time.Duration
	interval time.Duration
	leaseTTL time.Duration

	mu      sync.Mutex
	lastRun time.Time
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewScheduler 创建调度器。db 为空时 Start 不会启动循环。
// owner 为空则使用节点 ID，作为租约持有者标识。jobs 在到期时按注册顺序串行执行。
func NewScheduler(db database.Database, owner string, jobs []Job) *Scheduler {
	if owner == "" {
		owner = config.GetNodeId()
	}
	return &Scheduler{
		db:       db,
		owner:    owner,
		jobs:     jobs,
		tick:     defaultTick,
		interval: defaultInterval,
		leaseTTL: defaultLeaseTTL,
	}
}

// Start 启动后台调度循环。s 为 nil 或 db 为空时直接返回；已启动时重复调用无效。
// 循环绑定 parent 的派生 context，parent 取消或调用 Stop 都会让循环退出。
func (s *Scheduler) Start(parent context.Context) {
	if s == nil || s.db == nil {
		return
	}
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel
	s.mu.Unlock()

	s.wg.Add(1)
	go s.loop(ctx)
	logger.Info("生命周期调度已启动", "owner", s.owner, "jobs", len(s.jobs), "interval", s.interval)
}

// Stop 取消调度循环并等待 goroutine 退出。进程关闭时由 StopRetention 调用。
// 正在执行的 Job.Run 会因 context 取消而尽快返回；未启动或已停止时为空操作。
func (s *Scheduler) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	s.wg.Wait()
	logger.Info("生命周期调度已停止")
}

// loop 先尝试跑一轮，再按 tick 周期调用 maybeRun，直到 ctx 取消。
func (s *Scheduler) loop(ctx context.Context) {
	defer s.wg.Done()
	s.maybeRun(ctx)
	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.maybeRun(ctx)
		}
	}
}

// maybeRun 尝试成为租约持有者后读取归档任务配置，到期则串行执行全部 Job。
// 未抢到租约、任务已停用或尚未到期时立即返回。lastRun 在本轮 Job 全部调用后再更新。
func (s *Scheduler) maybeRun(ctx context.Context) {
	if !TryAcquireLease(ctx, s.db, s.owner, s.leaseTTL) {
		return
	}
	ReloadSettings(ctx, s.db)
	cfg := syssetting.GetRetentionJobForSchedule()
	if !cfg.Enabled {
		return
	}
	interval := time.Duration(cfg.IntervalMinutes) * time.Minute
	if interval <= 0 {
		interval = s.interval
	}
	s.mu.Lock()
	last := s.lastRun
	s.mu.Unlock()
	if !jobDue(time.Now(), last, interval, cfg.StartTime) {
		return
	}
	for _, job := range s.jobs {
		if job == nil {
			continue
		}
		if err := job.Run(ctx); err != nil {
			logger.Error("生命周期任务失败", "job", job.Name(), "error", err)
		}
	}
	s.mu.Lock()
	s.lastRun = time.Now()
	s.mu.Unlock()
}

// jobDue 判断当前是否应执行一轮清理。
// startHHMM 可解析时：每个自然日在该时刻之后最多执行一次，忽略 interval。
// 未配置开始时刻时：仅按 interval；lastRun 为零表示尚未执行过，立即到期。
func jobDue(now, lastRun time.Time, interval time.Duration, startHHMM string) bool {
	hour, minute, ok := syssetting.ParseJobStartClock(startHHMM)
	if ok {
		start := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if now.Before(start) {
			return false
		}
		// 今天该时刻之后已经跑过则等到下一个自然日
		if !lastRun.IsZero() && !lastRun.Before(start) {
			return false
		}
		return true
	}
	if lastRun.IsZero() {
		return true
	}
	return now.Sub(lastRun) >= interval
}
