package statichost

import (
	"bytes"
	"compress/gzip"
	"io"
	"mime"
	"path"
	"strings"

	"gateway/internal/gateway/core"
)

const (
	gzipMinBytes int64 = 1024
	gzipMaxBytes int64 = 1 << 20
)

var gzippableExt = map[string]struct{}{
	".css":  {},
	".htm":  {},
	".html": {},
	".js":   {},
	".json": {},
	".map":  {},
	".mjs":  {},
	".svg":  {},
	".txt":  {},
	".xml":  {},
}

// shouldGzipOnTheFly 在没有预压缩文件时，对中等大小的文本再压 gzip。
// Range 请求、已有 Content-Encoding、过小或过大的文件不压，以免和 206 冲突。
func shouldGzipOnTheFly(ctx *core.Context, snap *Snapshot, lookupPath, encoding string, size int64) bool {
	if snap == nil || !snap.EnableGzip || encoding != "" {
		return false
	}
	if size < gzipMinBytes || size > gzipMaxBytes {
		return false
	}
	if ctx == nil || ctx.Request == nil {
		return false
	}
	if strings.TrimSpace(ctx.Request.Header.Get("Range")) != "" {
		return false
	}
	if !acceptsEncoding(ctx.Request.Header.Get("Accept-Encoding"), "gzip") {
		return false
	}
	return isGzippableLookup(lookupPath)
}

func isGzippableLookup(lookupPath string) bool {
	ext := strings.ToLower(path.Ext(cleanURLPath(lookupPath)))
	if _, ok := gzippableExt[ext]; ok {
		return true
	}
	ctype := mime.TypeByExtension(ext)
	if ctype == "" {
		return false
	}
	if idx := strings.IndexByte(ctype, ';'); idx >= 0 {
		ctype = strings.TrimSpace(ctype[:idx])
	}
	ctype = strings.ToLower(ctype)
	return strings.HasPrefix(ctype, "text/") ||
		strings.Contains(ctype, "javascript") ||
		strings.Contains(ctype, "json") ||
		strings.Contains(ctype, "xml") ||
		ctype == "image/svg+xml"
}

func gzipContent(content io.ReadSeeker, size int64) ([]byte, error) {
	if content == nil {
		return nil, io.ErrUnexpectedEOF
	}
	if _, err := content.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(writer, io.LimitReader(content, size)); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
