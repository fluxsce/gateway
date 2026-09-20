package stream

import "strconv"

// pageOf 按 pageSize / pageToken 切片。pageSize<=0 时返回全部。
// token 是下一页起点下标的十进制表示。
func pageOf[T any](all []T, pageSize int32, token string) (out []T, next string, total int32) {
	total = int32(len(all))
	if pageSize <= 0 {
		return all, "", total
	}
	offset := 0
	if token != "" {
		if n, err := strconv.Atoi(token); err == nil && n > 0 {
			offset = n
		}
	}
	if offset >= len(all) {
		return []T{}, "", total
	}
	end := offset + int(pageSize)
	if end >= len(all) {
		return all[offset:], "", total
	}
	return all[offset:end], strconv.Itoa(end), total
}
