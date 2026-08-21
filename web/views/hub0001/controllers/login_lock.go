package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"gateway/pkg/cache"
	"gateway/pkg/logger"
	"strconv"
	"strings"
	"time"
)

const (
	loginFailLimit      = 5
	loginFailWindow     = 15 * time.Minute
	loginFailPrefix     = "login:fail:"
	pwdChangeFailPrefix = "pwdchange:fail:"
	loginCooldownFirst  = 30 * time.Second
	loginCooldownSecond = time.Minute
	loginCooldownMax    = 2 * time.Minute
)

var (
	// ErrAccountLocked 登录冷却中。对外文案带剩余秒数，本哨兵仅用于 errors.Is。
	ErrAccountLocked = errors.New("登录冷却中，请稍后再试")
)

// loginFailRecord 同一账号的失败计数与冷却截止时间。
// 计数在 15 分钟无失败后过期；冷却时长随次数递增，冷却结束后计数仍在，再错则加长。
type loginFailRecord struct {
	Count       int   `json:"n"`
	LockedUntil int64 `json:"until"`
}

// loginFailStore 登录失败计数存储。default 缓存即可，单测可替换。
type loginFailStore interface {
	GetString(ctx context.Context, key string) (string, error)
	SetString(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// LoginLockService 按 userId 做渐进登录冷却，不按 IP 锁定。
// 使用 default 缓存：内存模式下按节点生效，切 Redis 后全集群共享。
type LoginLockService struct {
	store  loginFailStore
	ttl    time.Duration
	prefix string
}

// NewLoginLockService 创建登录锁定服务。
func NewLoginLockService() *LoginLockService {
	var store loginFailStore
	if mgr := cache.GetGlobalManager(); mgr != nil {
		if c := mgr.GetCache("default"); c != nil {
			store = c
		}
	}
	return &LoginLockService{
		store:  store,
		ttl:    loginFailWindow,
		prefix: loginFailPrefix,
	}
}

// NewPasswordChangeLockService 创建改密失败冷却，与登录失败计数隔离。
func NewPasswordChangeLockService() *LoginLockService {
	s := NewLoginLockService()
	s.prefix = pwdChangeFailPrefix
	return s
}

// Check 返回是否处于冷却及剩余时间。缓存不可用时放行。
func (s *LoginLockService) Check(ctx context.Context, userId string) (locked bool, remaining time.Duration) {
	rec := s.load(ctx, userId)
	if rec == nil {
		return false, 0
	}
	remaining = remainingUntil(rec.LockedUntil)
	return remaining > 0, remaining
}

// RecordFailure 记录一次凭据失败。若因此进入冷却，返回剩余冷却时间。
func (s *LoginLockService) RecordFailure(ctx context.Context, userId string) time.Duration {
	if s == nil || s.store == nil {
		return 0
	}
	key := s.key(userId)
	if key == "" {
		return 0
	}
	rec := s.load(ctx, userId)
	if rec == nil {
		rec = &loginFailRecord{}
	}
	if remainingUntil(rec.LockedUntil) > 0 {
		return remainingUntil(rec.LockedUntil)
	}
	rec.Count++
	if delay := delayForCount(rec.Count); delay > 0 {
		rec.LockedUntil = time.Now().Add(delay).Unix()
	} else {
		rec.LockedUntil = 0
	}
	s.save(ctx, key, rec)
	return remainingUntil(rec.LockedUntil)
}

// Clear 登录成功后清除失败计数。
func (s *LoginLockService) Clear(ctx context.Context, userId string) {
	if s == nil || s.store == nil {
		return
	}
	key := s.key(userId)
	if key == "" {
		return
	}
	if err := s.store.Delete(ctx, key); err != nil {
		logger.WarnWithTrace(ctx, "清除登录失败计数失败", "error", err)
	}
}

func (s *LoginLockService) load(ctx context.Context, userId string) *loginFailRecord {
	if s == nil || s.store == nil {
		return nil
	}
	key := s.key(userId)
	if key == "" {
		return nil
	}
	raw, err := s.store.GetString(ctx, key)
	if err != nil {
		logger.WarnWithTrace(ctx, "读取登录失败计数失败，本次跳过锁定检查", "error", err)
		return nil
	}
	if raw == "" {
		return nil
	}
	var rec loginFailRecord
	if json.Unmarshal([]byte(raw), &rec) == nil && rec.Count > 0 {
		return &rec
	}
	if n, convErr := strconv.Atoi(raw); convErr == nil && n > 0 {
		return &loginFailRecord{Count: n}
	}
	return nil
}

func (s *LoginLockService) save(ctx context.Context, key string, rec *loginFailRecord) {
	payload, err := json.Marshal(rec)
	if err != nil {
		logger.WarnWithTrace(ctx, "序列化登录失败计数失败", "error", err)
		return
	}
	if err := s.store.SetString(ctx, key, string(payload), s.ttl); err != nil {
		logger.WarnWithTrace(ctx, "写入登录失败计数失败", "error", err)
	}
}

func (s *LoginLockService) key(userId string) string {
	id := strings.ToLower(strings.TrimSpace(userId))
	if id == "" {
		return ""
	}
	p := s.prefix
	if p == "" {
		p = loginFailPrefix
	}
	return p + id
}

func delayForCount(count int) time.Duration {
	switch {
	case count < loginFailLimit:
		return 0
	case count == loginFailLimit:
		return loginCooldownFirst
	case count == loginFailLimit+1:
		return loginCooldownSecond
	default:
		return loginCooldownMax
	}
}

func remainingUntil(untilUnix int64) time.Duration {
	if untilUnix <= 0 {
		return 0
	}
	d := time.Until(time.Unix(untilUnix, 0))
	if d < time.Second {
		if d > 0 {
			return time.Second
		}
		return 0
	}
	return d
}

// RemainSeconds 将冷却剩余时间转为至少 1 的秒数，供接口返回。
func RemainSeconds(remaining time.Duration) int {
	if remaining <= 0 {
		return 0
	}
	sec := int((remaining + time.Second - 1) / time.Second)
	if sec < 1 {
		return 1
	}
	return sec
}
