package tracing

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"math/rand/v2"
	"strings"
)

const (
	// HeaderTraceparent 是 W3C Trace Context 的主传递头。
	HeaderTraceparent = "traceparent"
	// HeaderTracestate 是 W3C Trace Context 的厂商状态头。
	HeaderTracestate = "tracestate"
)

// Traceparent 表示已解析的 W3C Trace Context 主传递头。
// 零值非法，使用前须调用 Valid。由 ParseTraceparent / NewTraceparent / ContinueTraceparent 构造。
type Traceparent struct {
	// Version 协议版本，当前固定为 00。
	Version string
	// TraceID 128-bit 链路 ID，32 位小写十六进制。
	TraceID string
	// ParentID 当前 Span ID，16 位小写十六进制。
	ParentID string
	// Flags 采样标志，01 表示采样，00 表示不采样。
	Flags string
}

// Valid 判断是否为合法的 W3C traceparent。全零 TraceID / ParentID 视为非法。
func (t Traceparent) Valid() bool {
	return len(t.TraceID) == 32 && len(t.ParentID) == 16 && len(t.Flags) == 2 && t.TraceID != strings.Repeat("0", 32) && t.ParentID != strings.Repeat("0", 16)
}

// Sampled 返回采样标志是否置位。
func (t Traceparent) Sampled() bool {
	return t.Valid() && (t.Flags == "01" || strings.HasSuffix(t.Flags, "1"))
}

// String 格式化为 traceparent 头值。
func (t Traceparent) String() string {
	version := t.Version
	if version == "" {
		version = "00"
	}
	return version + "-" + t.TraceID + "-" + t.ParentID + "-" + t.Flags
}

// ParseTraceparent 解析 W3C traceparent。格式非法或校验失败时返回零值，调用方用 Valid 判断。
func ParseTraceparent(value string) Traceparent {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 4 {
		return Traceparent{}
	}
	version, traceID, parentID, flags := parts[0], strings.ToLower(parts[1]), strings.ToLower(parts[2]), strings.ToLower(parts[3])
	if len(version) != 2 || len(traceID) != 32 || len(parentID) != 16 || len(flags) != 2 {
		return Traceparent{}
	}
	if !isHex(traceID) || !isHex(parentID) || !isHex(flags) || !isHex(version) {
		return Traceparent{}
	}
	tp := Traceparent{Version: version, TraceID: traceID, ParentID: parentID, Flags: flags}
	if !tp.Valid() {
		return Traceparent{}
	}
	return tp
}

// NewTraceparent 创建根 Trace。sampled 为 true 时 flags 置为 01。
func NewTraceparent(sampled bool) Traceparent {
	return Traceparent{
		Version:  "00",
		TraceID:  randomHex(16),
		ParentID: randomHex(8),
		Flags:    flagsValue(sampled),
	}
}

// ContinueTraceparent 继承父 TraceID 与采样标志，生成新的 Span ID。parent 非法时新建未采样根 Trace。
func ContinueTraceparent(parent Traceparent) Traceparent {
	if !parent.Valid() {
		return NewTraceparent(false)
	}
	return Traceparent{
		Version:  "00",
		TraceID:  parent.TraceID,
		ParentID: randomHex(8),
		Flags:    parent.Flags,
	}
}

func flagsValue(sampled bool) string {
	if sampled {
		return "01"
	}
	return "00"
}

// shouldSample 按 [0,1] 采样率做头采样。rate>=1 全采，rate<=0 不采。
func shouldSample(rate float64) bool {
	if rate >= 1 {
		return true
	}
	if rate <= 0 {
		return false
	}
	return rand.Float64() < rate
}

// randomHex 生成 nBytes 字节的小写十六进制。优先 crypto/rand，失败时回落 math/rand。
func randomHex(nBytes int) string {
	buf := make([]byte, nBytes)
	if _, err := cryptorand.Read(buf); err != nil {
		for i := range buf {
			buf[i] = byte(rand.UintN(256))
		}
	}
	return hex.EncodeToString(buf)
}

// isHex 判断是否为小写十六进制（解析前已 ToLower）。
func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
