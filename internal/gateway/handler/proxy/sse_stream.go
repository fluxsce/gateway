package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
)

const (
	sseDisconnectCompleted    = "completed"
	sseDisconnectClientClosed = "client_closed"
	sseDisconnectUpstream     = "upstream_error"
	sseDisconnectDownstream   = "downstream_error"
)

// sseStreamer 将上游事件流实时转发到客户端，并管理滚动写截止时间和断开语义。
// 当 bodySampleLimit>0 时，仅缓存流前缀用于访问/后端日志，避免把整条无限流落入内存。
type sseStreamer struct {
	bufferSize      int
	sendTimeout     time.Duration
	bodySampleLimit int
}

// newSSEStreamer 创建SSE流式转发器。
// bodySampleLimit 为响应体采样上限字节数；0 表示不采样报文体。
func newSSEStreamer(bufferSize int, sendTimeout time.Duration, bodySampleLimit int) *sseStreamer {
	if bufferSize <= 0 {
		bufferSize = 1024
	}
	if bodySampleLimit < 0 {
		bodySampleLimit = 0
	}
	return &sseStreamer{
		bufferSize:      bufferSize,
		sendTimeout:     sendTimeout,
		bodySampleLimit: bodySampleLimit,
	}
}

// Stream 转发单个SSE响应，直到上游结束、客户端断开或发生真实传输错误。
// 结束后写入转发字节数、断开原因；若启用采样则写入截断后的 response_body。
func (s *sseStreamer) Stream(ctx *core.Context, resp *http.Response) error {
	s.copyHeaders(ctx.Writer.Header(), resp.Header)
	if ctx.Writer.Header().Get("Content-Type") == "" {
		ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	}
	if ctx.Writer.Header().Get("Cache-Control") == "" {
		ctx.Writer.Header().Set("Cache-Control", "no-store")
	}

	controller := http.NewResponseController(ctx.Writer)
	// 清除http.Server的绝对WriteTimeout，后续只在单次写入期间设置滚动截止时间。
	_ = controller.SetWriteDeadline(time.Time{})
	ctx.Writer.WriteHeader(resp.StatusCode)
	ctx.SetResponded()
	if err := controller.Flush(); err != nil {
		return fmt.Errorf("刷新SSE响应头失败: %w", err)
	}

	buffer := make([]byte, s.bufferSize)
	var bytesStreamed int64
	var bodySample []byte
	disconnectType := sseDisconnectCompleted
	defer func() {
		ctx.Set(constants.ContextKeySSEBytesStreamed, bytesStreamed)
		ctx.Set(constants.ContextKeySSEDisconnectType, disconnectType)
		ctx.Set(constants.ContextKeyResponseSize, clampInt64ToInt(bytesStreamed))
		// 仅在配置要求记录响应体时保留流前缀，供异步访问日志/后端追踪读取。
		if s.bodySampleLimit > 0 && len(bodySample) > 0 {
			ctx.Set("response_body", bodySample)
		}
	}()

	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			if s.bodySampleLimit > 0 && len(bodySample) < s.bodySampleLimit {
				remain := s.bodySampleLimit - len(bodySample)
				if remain > n {
					remain = n
				}
				bodySample = append(bodySample, buffer[:remain]...)
			}
			if s.sendTimeout > 0 {
				_ = controller.SetWriteDeadline(time.Now().Add(s.sendTimeout))
			}
			_, writeErr := ctx.Writer.Write(buffer[:n])
			_ = controller.SetWriteDeadline(time.Time{})
			if writeErr != nil {
				if isRequestCancellation(ctx.Request.Context(), writeErr) {
					disconnectType = sseDisconnectClientClosed
					return nil
				}
				disconnectType = sseDisconnectDownstream
				err := fmt.Errorf("写入SSE响应失败: %w", writeErr)
				ctx.AddError(err)
				return err
			}
			bytesStreamed += int64(n)
			if err := controller.Flush(); err != nil {
				if isRequestCancellation(ctx.Request.Context(), err) {
					disconnectType = sseDisconnectClientClosed
					return nil
				}
				disconnectType = sseDisconnectDownstream
				flushErr := fmt.Errorf("刷新SSE响应失败: %w", err)
				ctx.AddError(flushErr)
				return flushErr
			}
		}

		if readErr == nil {
			continue
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if isRequestCancellation(ctx.Request.Context(), readErr) {
			disconnectType = sseDisconnectClientClosed
			return nil
		}
		disconnectType = sseDisconnectUpstream
		err := fmt.Errorf("读取SSE上游响应失败: %w", readErr)
		ctx.AddError(err)
		return err
	}
}

func (s *sseStreamer) copyHeaders(dst, src http.Header) {
	for name, values := range src {
		if isHopByHopHeader(name) || name == "Content-Length" ||
			name == "Access-Control-Allow-Origin" {
			continue
		}
		for _, value := range values {
			dst.Add(name, value)
		}
	}
}

func isRequestCancellation(requestCtx context.Context, err error) bool {
	return requestCtx.Err() != nil ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}

// clampInt64ToInt 将字节统计收敛为访问日志使用的 int，溢出时取 MaxInt。
func clampInt64ToInt(v int64) int {
	if v < 0 {
		return 0
	}
	if v > int64(math.MaxInt) {
		return math.MaxInt
	}
	return int(v)
}
