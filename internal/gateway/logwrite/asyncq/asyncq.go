package asyncq

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// EnqueueWait 队列满时再等这一小段，给消费者腾坑；仍满才丢。
	// 访问日志提交发生在 SetResponseTime 之后，短等不计入主表总时间。
	EnqueueWait = 20 * time.Millisecond
	// DefaultWriteTimeout 异步刷盘/插库没有调用方超时时使用的上限。
	DefaultWriteTimeout = 30 * time.Second
)

// Offer 把 item 放入有界通道：先立刻投递，满则再等 EnqueueWait，仍满返回 false。
func Offer[T any](ch chan T, item T) bool {
	return OfferWait(ch, item, EnqueueWait)
}

// OfferWait 与 Offer 相同，等待时间由调用方指定。wait<=0 时满则立刻失败。
func OfferWait[T any](ch chan T, item T, wait time.Duration) bool {
	if ch == nil {
		return false
	}
	select {
	case ch <- item:
		return true
	default:
	}
	if wait <= 0 {
		return false
	}
	timer := time.NewTimer(wait)
	select {
	case ch <- item:
		timer.Stop()
		return true
	case <-timer.C:
		return false
	}
}

// Drain 非阻塞取出通道内现有元素并交给 handle，返回处理条数。
func Drain[T any](ch chan T, handle func(T)) int {
	if ch == nil || handle == nil {
		return 0
	}
	count := 0
	for {
		select {
		case item := <-ch:
			handle(item)
			count++
		default:
			return count
		}
	}
}

// Take 在锁内拿走缓冲区全部元素并换上空切片，调用方应在解锁后写库。
func Take[T any](mu *sync.Mutex, buf *[]T, capHint int) []T {
	if mu == nil || buf == nil {
		return nil
	}
	if capHint < 1 {
		capHint = 1
	}
	mu.Lock()
	defer mu.Unlock()
	if len(*buf) == 0 {
		return nil
	}
	out := *buf
	*buf = make([]T, 0, capHint)
	return out
}

// RestorePrepend 把写失败的一批插回缓冲头部，下次 Flush 优先重试。
// 超过 max 时截掉尾部（更新的未刷条目），避免存储长时间不可用时内存涨死。
// 返回因超限丢掉的条数。
func RestorePrepend[T any](mu *sync.Mutex, buf *[]T, batch []T, max, capHint int) int {
	if mu == nil || buf == nil || len(batch) == 0 {
		return 0
	}
	if max < 1 {
		max = len(batch)
	}
	if capHint < 1 {
		capHint = max
	}
	mu.Lock()
	defer mu.Unlock()
	combined := make([]T, 0, len(batch)+len(*buf))
	combined = append(combined, batch...)
	combined = append(combined, *buf...)
	dropped := 0
	if len(combined) > max {
		dropped = len(combined) - max
		combined = combined[:max]
	}
	if cap := cap(combined); cap < capHint && dropped == 0 {
		grown := make([]T, len(combined), capHint)
		copy(grown, combined)
		combined = grown
	}
	*buf = combined
	return dropped
}

// AppendCapped 在冷却期内只追加、不触发满批直写；已达 max 则拒绝新条目并返回 1。
func AppendCapped[T any](mu *sync.Mutex, buf *[]T, item T, max int) int {
	if mu == nil || buf == nil {
		return 1
	}
	if max < 1 {
		max = 1
	}
	mu.Lock()
	defer mu.Unlock()
	if len(*buf) >= max {
		return 1
	}
	*buf = append(*buf, item)
	return 0
}

// ShouldDeferFlush 为真时满批直写应改为只入缓冲，等定时 Flush 重试，避免存储故障时热循环。
func ShouldDeferFlush(retryAfter *atomic.Int64) bool {
	if retryAfter == nil {
		return false
	}
	return time.Now().UnixNano() < retryAfter.Load()
}

// MarkRetryAfter 从现在起 delay 内不再满批直写。delay<=0 时按 1s。
func MarkRetryAfter(retryAfter *atomic.Int64, delay time.Duration) {
	if retryAfter == nil {
		return
	}
	if delay <= 0 {
		delay = time.Second
	}
	retryAfter.Store(time.Now().Add(delay).UnixNano())
}

// RetryableWriteError 网络、超时、选主失败可回灌；重复键等数据错误不可回灌，以免死循环。
func RetryableWriteError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "duplicate") || strings.Contains(msg, "e11000") {
		return false
	}
	for _, key := range []string{
		"server selection",
		"timeout",
		"deadline",
		"connection",
		"network",
		"i/o",
		"reset",
		"refused",
		"unavailable",
		"eof",
	} {
		if strings.Contains(msg, key) {
			return true
		}
	}
	return false
}

// AppendTakeIfFull 追加一条；达到 limit 时拿走整批供调用方在锁外写入。
func AppendTakeIfFull[T any](mu *sync.Mutex, buf *[]T, item T, limit, capHint int) []T {
	if mu == nil || buf == nil {
		return nil
	}
	if limit < 1 {
		limit = 1
	}
	if capHint < 1 {
		capHint = limit
	}
	mu.Lock()
	defer mu.Unlock()
	*buf = append(*buf, item)
	if len(*buf) < limit {
		return nil
	}
	out := *buf
	*buf = make([]T, 0, capHint)
	return out
}

// WriteContext 为刷盘/插库构造超时 context。timeout<=0 时用 DefaultWriteTimeout。
func WriteContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultWriteTimeout
	}
	return context.WithTimeout(context.Background(), timeout)
}

// EnsureWriteCtxTimeout 保证写库/刷盘有超时：已有 deadline 则沿用，否则套 timeout。
func EnsureWriteCtxTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultWriteTimeout
	}
	if ctx != nil && ctx.Err() == nil {
		if _, ok := ctx.Deadline(); ok {
			return ctx, func() {}
		}
		return context.WithTimeout(ctx, timeout)
	}
	return WriteContext(timeout)
}

// EnsureWriteCtx 保证写库/刷盘有超时：已有 deadline 则沿用，否则套 DefaultWriteTimeout。
func EnsureWriteCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return EnsureWriteCtxTimeout(ctx, DefaultWriteTimeout)
}
