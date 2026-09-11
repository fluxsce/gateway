package proxy

import (
	"time"

	"gateway/internal/gateway/core"
)

// waitRetryInterval 等待重试间隔，并把实际等待计入上下文。
// 这段墙钟属于「等下一次尝试」，不是网关计算/过滤时间，主表网关耗时会扣除它。
// 返回 false 表示请求已取消，调用方应停止重试。
func waitRetryInterval(ctx *core.Context, retryTimeout time.Duration) bool {
	if ctx == nil || ctx.Request == nil {
		return false
	}
	start := time.Now()
	select {
	case <-ctx.Request.Context().Done():
		ctx.AddRetryWait(time.Since(start))
		return false
	case <-time.After(retryTimeout):
		ctx.AddRetryWait(time.Since(start))
		return true
	}
}
