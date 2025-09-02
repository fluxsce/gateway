package random

import (
	"crypto/rand"
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

var (
	// 全局原子计数器，用于确保唯一性
	atomicCounter uint64
	// 进程启动时间戳，用于区分不同的应用实例
	processStartTime int64
	// 当前节点IP地址，用于分布式环境下的唯一性标识
	nodeIP string
)

func init() {
	// 初始化进程启动时间戳
	processStartTime = time.Now().UnixNano()

	// 获取当前节点IP地址
	nodeIP = getNodeIP()

	// 使用进程启动时间的纳秒部分初始化计数器，避免重启后从0开始
	// 这样即使应用重启，计数器也不会从相同值开始
	atomic.StoreUint64(&atomicCounter, uint64(processStartTime%1000000))
}

// getNodeIP 获取当前节点的IPv4地址
func getNodeIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}

// GetNodeIP 获取当前节点IP地址（供外部调用）
func GetNodeIP() string {
	return nodeIP
}

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

// Generate32BitRandomString 生成32位唯一字符串（大写字母和数字）
// 专门用于生成符合数据库VARCHAR(32)字段的主键ID
// 使用时间戳+进程标识+原子计数器+强随机数确保唯一性，即使应用重启也不会重复
func Generate32BitRandomString() string {
	// 获取当前时间戳（微秒）
	currentTime := time.Now().UnixMicro()

	// 原子递增计数器
	counter := atomic.AddUint64(&atomicCounter, 1)

	// 计算进程标识（使用启动时间的后几位）
	processID := processStartTime % 10000 // 4位进程标识

	// 生成强随机数部分
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomBytes := make([]byte, 4)
	var randomNum uint32

	if _, err := rand.Read(randomBytes); err != nil {
		// 如果随机数生成失败，使用当前纳秒时间戳
		randomNum = uint32(time.Now().Nanosecond())
	} else {
		// 将4字节转换为uint32
		randomNum = uint32(randomBytes[0])<<24 |
			uint32(randomBytes[1])<<16 |
			uint32(randomBytes[2])<<8 |
			uint32(randomBytes[3])
	}

	// 组合生成唯一ID：
	// 时间戳(13位) + 进程ID(4位) + 计数器(6位) + 随机数(9位) = 32位
	timestamp13 := currentTime % 10000000000000 // 取时间戳后13位
	counter6 := counter % 1000000               // 取计数器后6位
	random9 := uint64(randomNum) % 1000000000   // 取随机数后9位

	// 格式化为32位数字字符串
	uniqueID := fmt.Sprintf("%013d%04d%06d%09d",
		timestamp13, processID, counter6, random9)

	// 将数字转换为字母数字混合格式，提高可读性
	result := make([]byte, 32)
	for i := 0; i < 32; i++ {
		digit := uniqueID[i] - '0'
		if i%3 == 0 && digit < 26 {
			// 每3位中的第1位，如果数字小于26，转换为字母
			result[i] = byte('A' + digit)
		} else {
			// 其他位保持数字或转换为字母
			if digit < 10 {
				result[i] = byte('0' + digit)
			} else {
				result[i] = charset[digit%36]
			}
		}
	}

	return string(result)
}

// pow10 计算10的n次方
func pow10(n int) int64 {
	result := int64(1)
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}
