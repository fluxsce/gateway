package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"gateway/internal/gateway/core"
)

// forwardAssist 单次发往上游的转发辅助：请求体按需缓冲或流过，发出头深拷一份脱离活的 proxyReq。
// proxyRequest 用 Body() 构造上游请求，defer 里 CommitBody 回填 request_body、CloneHeader 给上下文和从表。
// 会重试或记请求体时整包缓冲，request_body 始终是完整转发体；截断只在写库做。
// 多服务只用 CloneHeader，每条上游各拷一份只当参数传入，不写共享 ctx。
type forwardAssist struct {
	reader     io.Reader // 交给 http.NewRequest 的上游 body；无 body 时为 nil
	contentLen int64     // 已知长度；-1 表示未知（客户端 chunked），0 表示空
	counter    *byteCounter
	full       []byte // 整包缓冲后的转发体；流式路径为 nil
	recorded   bool   // 是否要把完整转发体写入 request_body
}

// byteCounter 只统计 TeeReader 流过的字节，不保存正文。
// 用于流式且 Content-Length 未知时，发送后再填 backend 的 requestSize。
type byteCounter struct {
	n int64
}

// Write 计入长度并吃下全部输入，避免 TeeReader 因写入失败中断转发。
func (c *byteCounter) Write(p []byte) (int, error) {
	c.n += int64(len(p))
	return len(p), nil
}

// prepareForwardAssist 按重试和记日志开关构造发给上游的请求体。
//
//   - maxRetries>0 或 record：ReadAll 整包，重置 req.Body，便于重试再读；contentLen 用实际字节数。
//   - 两都不开且长度已知：reader 指向原 Body，不拷贝。
//   - 两都不开且长度未知：TeeReader 只计数，仍不存正文。
//
// 不按 HTTP 方法判断：GET/DELETE 只要 Body 不是 nil/NoBody，记日志或重试时同样整包缓冲。
// req 为 nil、Body 为 nil 或 http.NoBody 时返回空的 forwardAssist，不报错。
func prepareForwardAssist(req *http.Request, maxRetries int, record bool) (*forwardAssist, error) {
	fb := &forwardAssist{contentLen: -1, recorded: record}
	if req == nil || req.Body == nil || req.Body == http.NoBody {
		return fb, nil
	}
	if req.ContentLength >= 0 {
		fb.contentLen = req.ContentLength
	}

	// 重试要能把同一份体再发给下一个节点；记日志要完整 request_body，截断留给写库。
	if maxRetries > 0 || record {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("读取请求体失败: %w", err)
		}
		fb.full = bodyBytes
		fb.contentLen = int64(len(bodyBytes))
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		fb.reader = bytes.NewReader(bodyBytes)
		return fb, nil
	}

	// 未知长度时 NewRequest 不会带 Content-Length，发送后用计数填 requestSize。
	if fb.contentLen < 0 {
		fb.counter = &byteCounter{}
		fb.reader = io.TeeReader(req.Body, fb.counter)
		return fb, nil
	}
	fb.reader = req.Body
	return fb, nil
}

// Body 返回交给上游的 Reader，无 body 时为 nil。
func (f *forwardAssist) Body() io.Reader {
	if f == nil {
		return nil
	}
	return f.reader
}

// ContentLength 返回已知转发体长度；-1 表示未知。
func (f *forwardAssist) ContentLength() int64 {
	if f == nil {
		return -1
	}
	return f.contentLen
}

// KnownSize 返回已确定的转发体字节数，供后端追踪的 requestSize 使用。
// 整包缓冲用实际长度；否则用客户端声明的 Content-Length；再否则用流过的计数。
// 空 body 或尚未读完的未知长度返回 0。须在上游读完 body 之后再取计数才准。
func (f *forwardAssist) KnownSize() int {
	if f == nil {
		return 0
	}
	if f.full != nil {
		return len(f.full)
	}
	if f.contentLen > 0 {
		return int(f.contentLen)
	}
	if f.counter != nil && f.counter.n > 0 {
		return int(f.counter.n)
	}
	return 0
}

// CommitBody 把完整转发体写入 ctx 的 request_body，供 WriteLog / 后端追踪读取。
// 未开启记录或未整包缓冲时不写。不在这里按 MaxBodySizeBytes 截断。
func (f *forwardAssist) CommitBody(ctx *core.Context) {
	if f == nil || !f.recorded || ctx == nil || f.full == nil {
		return
	}
	ctx.Set("request_body", f.full)
}

// CloneHeader 深拷一份转发头，脱离活的 proxyReq.Header。
// 单服务上下文和后端追踪 job 共用这一份；多服务只把这一份当参数传入，不写共享 ctx，避免群发互相覆盖。
// 接收者可为 nil 或零值，不依赖请求体状态。
func (f *forwardAssist) CloneHeader(h http.Header) http.Header {
	if len(h) == 0 {
		return nil
	}
	out := make(http.Header, len(h))
	for k, v := range h {
		out[k] = append([]string(nil), v...)
	}
	return out
}
