package random

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GenerateRandomString 生成指定长度的随机字符串（大写字母和数字）
func GenerateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%0*d", length, time.Now().Nanosecond()%int(pow10(length)))
	}
	
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	
	return string(b)
}

// pow10 计算10的n次方
func pow10(n int) int64 {
	result := int64(1)
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
} 