package logwrite

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/logwrite/asyncq"
	"gateway/pkg/logger"
)

const (
	// accessLogSubmitTimeout 单条 WriteLog 的独立超时，与原先 go 里 30s 一致。
	accessLogSubmitTimeout = 30 * time.Second
	// accessLogDrainWait 关写入器前等待队列排空的上限。
	accessLogDrainWait = 3 * time.Second
	// accessLogSubmitQueueSize 提交队列容量。约 6.7k QPS 时可缓冲近 10 秒突发，避免稍一堵就丢。
	accessLogSubmitQueueSize = 65536
	minAccessLogWorkers      = 4
	maxAccessLogWorkers      = 32
)

// accessLogJob 访问日志提交任务。
// ctx 在入队前必须已经完成 snapshot 且 Request/Writer 为 nil，worker 只使用 ctx.data 与时间字段。
type accessLogJob struct {
	instanceID string
	ctx        *core.Context
}

// AccessLogSubmitStats 访问日志提交队列的当前快照，供排查是否堆积或丢弃。
type AccessLogSubmitStats struct {
	QueueCap    int    // 队列容量
	Queued      int    // 当前排队未消费条数
	Workers     int    // 固定消费者数量
	Inflight    int64  // 正在 WriteLog 的条数
	Dropped     uint64 // 等待后仍无法入队而丢弃的累计条数
	EnqueueWait string // 队列满时的短暂等待
}

// accessLogPendingSet 按实例统计已入队或正在 WriteLog 的条数，供关写入器前等待。
type accessLogPendingSet struct {
	byInstance sync.Map
}

var (
	accessLogSubmitOnce    sync.Once
	accessLogSubmitQ       chan accessLogJob
	accessLogSubmitWorkers int
	accessLogDropped       atomic.Uint64
	accessLogInflight      atomic.Int64
	accessLogPending       accessLogPendingSet
)

// SubmitWriteLog 将已快照的访问日志交给固定 worker 调用 WriteLog。
//
// 调用方约定（与 finishRequest 对齐，缺一不可）：
//  1. 已在 ServeHTTP 返回前调用 snapshotHTTPData，日志字段在 ctx.data
//  2. 已将 ctx.Request、ctx.Writer 置 nil，防止 worker 在请求结束后误访问
//  3. 本函数不读 HTTP 对象；静态成功命中直接返回，不占队列
//
// 写入语义与原先 go WriteLog 相同：主表字段、端口重放 UpdateAccessLog、静态跳过。
// 响应已写出，入队不等待：满则立刻丢弃并计数，避免拖住 HTTP/1.1 复用连接上的下一条。
func SubmitWriteLog(instanceID string, gatewayCtx *core.Context) {
	if instanceID == "" || gatewayCtx == nil {
		return
	}
	// 只读 ctx.data 中的静态标记与状态码，不访问 Request/Writer。
	if shouldSkipStaticAccessLog(gatewayCtx) {
		return
	}
	ensureAccessLogSubmit()
	// 入队的是 *Context 指针；安全性依赖调用方已断掉 HTTP 引用。
	offerAccessLogJob(accessLogJob{instanceID: instanceID, ctx: gatewayCtx})
}

// GetAccessLogSubmitStats 返回提交队列容量、积压、worker 数和丢弃计数。
// 队列未启动时 QueueCap 为规划容量，Queued 为 0。
func GetAccessLogSubmitStats() AccessLogSubmitStats {
	stats := AccessLogSubmitStats{
		QueueCap:    accessLogSubmitQueueSize,
		Workers:     accessLogSubmitWorkers,
		Inflight:    accessLogInflight.Load(),
		Dropped:     accessLogDropped.Load(),
		EnqueueWait: "0s",
	}
	if accessLogSubmitQ != nil {
		stats.QueueCap = cap(accessLogSubmitQ)
		stats.Queued = len(accessLogSubmitQ)
		if stats.Workers == 0 {
			stats.Workers = accessLogSubmitWorkerCount()
		}
	} else {
		stats.Workers = accessLogSubmitWorkerCount()
	}
	return stats
}

// offerAccessLogJob 把任务交给消费者：立即入队，满则丢弃。
func offerAccessLogJob(job accessLogJob) {
	accessLogPending.add(job.instanceID, 1)
	// 响应已写出，入队不再短等：避免 HTTP/1.1 复用连接上的下一条被空等拖住。
	if asyncq.OfferWait(accessLogSubmitQ, job, 0) {
		return
	}
	accessLogPending.add(job.instanceID, -1)
	n := accessLogDropped.Add(1)
	if n == 1 || n%1000 == 0 {
		logger.Warn("访问日志提交队列已满，丢弃新日志",
			"dropped", n,
			"queueCap", cap(accessLogSubmitQ),
			"queued", len(accessLogSubmitQ),
			"inflight", accessLogInflight.Load(),
			"instanceID", job.instanceID)
	}
}

// ensureAccessLogSubmit 启动一次有界队列和固定 worker，进程内只做一遍。
func ensureAccessLogSubmit() {
	accessLogSubmitOnce.Do(func() {
		accessLogSubmitQ = make(chan accessLogJob, accessLogSubmitQueueSize)
		accessLogSubmitWorkers = accessLogSubmitWorkerCount()
		for i := 0; i < accessLogSubmitWorkers; i++ {
			go accessLogSubmitWorker()
		}
		logger.Info("访问日志提交队列已启动",
			"workers", accessLogSubmitWorkers,
			"queueCap", accessLogSubmitQueueSize,
			"enqueueWait", "0s")
	})
}

// accessLogSubmitWorkerCount 按 GOMAXPROCS 取消费者个数，夹在 4～32。
func accessLogSubmitWorkerCount() int {
	n := runtime.GOMAXPROCS(0)
	if n < minAccessLogWorkers {
		return minAccessLogWorkers
	}
	if n > maxAccessLogWorkers {
		return maxAccessLogWorkers
	}
	return n
}

// accessLogSubmitWorker 从队列取出任务并调用 runAccessLogJob，直到进程退出。
func accessLogSubmitWorker() {
	for job := range accessLogSubmitQ {
		runAccessLogJob(job)
	}
}

// runAccessLogJob 在固定 worker 中执行原 WriteLog。
// 此时 ServeHTTP 可能已返回：禁止使用 job.ctx.Request / job.ctx.Writer。
// HTTP 请求的 context 可能已被 finishRequest 之后的 Cancel 取消，
// 因此换成独立 timeout context，避免写库被已结束的请求 context 打断。
func runAccessLogJob(job accessLogJob) {
	accessLogInflight.Add(1)
	defer accessLogInflight.Add(-1)
	defer accessLogPending.add(job.instanceID, -1)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in access log writer", "error", r, "instanceID", job.instanceID)
		}
	}()

	// 不复用请求 context：r.Context() 在 ServeHTTP 返回后会取消。
	logCtx, cancel := context.WithTimeout(context.Background(), accessLogSubmitTimeout)
	defer cancel()

	originalCtx := job.ctx.Ctx
	job.ctx.Ctx = logCtx
	defer func() {
		job.ctx.Ctx = originalCtx
	}()

	// WriteLog / buildAccessLogFromContext 只读快照与 ctx.data，见包头异步安全规则。
	if err := WriteLog(job.instanceID, job.ctx); err != nil {
		logger.Error("Failed to write access log", "error", err, "instanceID", job.instanceID)
	}
}

// WaitAccessLogSubmitIdle 等待指定实例已入队和在途的访问日志、后端追踪写完。
// UnregisterLogWriter / CloseLogWriter 在摘槽前调用。
func WaitAccessLogSubmitIdle(instanceID string, wait time.Duration) {
	accessLogPending.waitIdle(instanceID, wait)
}

// drainAccessLogSubmit 等待访问日志与后端追踪提交队列都空闲，供 CloseAllLogWriters 使用。
func drainAccessLogSubmit(wait time.Duration) {
	if wait <= 0 {
		wait = accessLogDrainWait
	}
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		if accessLogSubmitIdle() && backendTraceSubmitIdle() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	accessQueued, accessInflight := 0, int64(0)
	if accessLogSubmitQ != nil {
		accessQueued = len(accessLogSubmitQ)
		accessInflight = accessLogInflight.Load()
	}
	backendQueued, backendInflight := 0, int64(0)
	if backendTraceSubmitQ != nil {
		backendQueued = len(backendTraceSubmitQ)
		backendInflight = backendTraceInflight.Load()
	}
	logger.Warn("等待日志提交队列排空超时",
		"accessQueued", accessQueued,
		"accessInflight", accessInflight,
		"backendQueued", backendQueued,
		"backendInflight", backendInflight)
}

func accessLogSubmitIdle() bool {
	return accessLogSubmitQ == nil ||
		(len(accessLogSubmitQ) == 0 && accessLogInflight.Load() == 0)
}

func (p *accessLogPendingSet) add(instanceID string, delta int64) {
	if p == nil || instanceID == "" || delta == 0 {
		return
	}
	actual, _ := p.byInstance.LoadOrStore(instanceID, &atomic.Int64{})
	actual.(*atomic.Int64).Add(delta)
}

func (p *accessLogPendingSet) count(instanceID string) int64 {
	if p == nil || instanceID == "" {
		return 0
	}
	actual, ok := p.byInstance.Load(instanceID)
	if !ok {
		return 0
	}
	return actual.(*atomic.Int64).Load()
}

func (p *accessLogPendingSet) waitIdle(instanceID string, wait time.Duration) {
	if instanceID == "" {
		return
	}
	if wait <= 0 {
		wait = accessLogDrainWait
	}
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		if p.count(instanceID) == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	logger.Warn("等待实例访问日志提交排空超时",
		"instanceID", instanceID,
		"pending", p.count(instanceID))
}
