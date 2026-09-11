package asyncq

import (
	"context"
	"sync"
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
