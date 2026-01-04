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
	// 节点ID（从IP地址计算，用于集群环境下的唯一性）
	nodeID uint32
)

func init() {
	// 初始化进程启动时间戳
	processStartTime = time.Now().UnixNano()

	// 获取当前节点IP地址
	nodeIP = getNodeIP()

	// 计算节点ID（从IP地址计算，用于集群环境）
	// 注意：此时processStartTime已初始化，可以用于计算
	nodeID = calculateNodeID(nodeIP, processStartTime)

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

// calculateNodeID 从IP地址计算节点ID（用于集群环境下的唯一性）
// 不同节点的IP地址不同，确保集群环境下生成的ID唯一
func calculateNodeID(ip string, startTime int64) uint32 {
	// 将IP地址转换为数字标识
	// 例如：192.168.1.100 -> 192*256^3 + 168*256^2 + 1*256 + 100
	var hash uint32
	var num uint32 = 0
	for _, c := range ip {
		if c == '.' {
			hash = hash*256 + num
			num = 0
		} else if c >= '0' && c <= '9' {
			num = num*10 + uint32(c-'0')
		}
	}
	hash = hash*256 + num

	// 如果IP解析失败（如127.0.0.1），使用字符串哈希
	if hash == 0 || hash == 2130706433 { // 127.0.0.1的数值
		for _, c := range ip {
			hash = hash*31 + uint32(c)
		}
		// 结合进程启动时间，确保不同进程有不同的节点ID
		hash = hash ^ uint32(startTime%1000000)
	}

	return hash
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
// 使用时间戳+节点ID+原子计数器+随机数确保唯一性，支持集群环境
// 内部调用 GenerateUniqueStringWithPrefix("", 32) 实现
func Generate32BitRandomString() string {
	return GenerateUniqueStringWithPrefix("", 32)
}

// GenerateUniqueStringWithPrefix 生成带前缀的唯一字符串（支持集群环境）
// 参数:
//   - prefix: 前缀字符串（如 "IP", "GW" 等），可以为空
//   - totalLength: 总长度（包括前缀），最小值为前缀长度+10，建议不超过64
//
// 返回:
//   - 唯一字符串，格式为 prefix + 唯一随机字符串
//
// 说明:
//
//	使用时间戳(微秒) + 节点ID + 原子计数器 + 随机数确保唯一性
//	在集群环境下，不同节点的节点ID不同，确保全局唯一性
func GenerateUniqueStringWithPrefix(prefix string, totalLength int) string {
	prefixLen := len(prefix)
	if totalLength <= prefixLen {
		return prefix
	}

	// 计算随机部分长度
	randomLength := totalLength - prefixLen
	if randomLength < 10 {
		randomLength = 10 // 最小10位保证唯一性
	}
	if randomLength > 64 {
		randomLength = 64 // 最大64位
	}

	// 获取当前时间戳（微秒）- 只调用一次
	currentTime := time.Now().UnixMicro()

	// 原子递增计数器
	counter := atomic.AddUint64(&atomicCounter, 1)

	// 生成随机数（简化处理，使用计数器和时间戳混合）
	// 在高并发下，crypto/rand可能有锁竞争，这里使用混合方案保证性能
	randomNum := uint32(counter) ^ uint32(currentTime) ^ nodeID

	// 根据长度动态分配各部分位数
	// 时间戳(8-12位) + 节点ID(4-6位) + 计数器(4-6位) + 随机数(剩余位数)
	var timestampBits, nodeBits, counterBits, randomBits int

	if randomLength >= 32 {
		// 长ID：时间戳(12位) + 节点ID(6位) + 计数器(6位) + 随机数(剩余)
		timestampBits = 12
		nodeBits = 6
		counterBits = 6
		randomBits = randomLength - 24
	} else if randomLength >= 20 {
		// 中等ID：时间戳(10位) + 节点ID(4位) + 计数器(4位) + 随机数(剩余)
		timestampBits = 10
		nodeBits = 4
		counterBits = 4
		randomBits = randomLength - 18
	} else {
		// 短ID：时间戳(8位) + 节点ID(4位) + 计数器(2位) + 随机数(剩余)
		timestampBits = 8
		nodeBits = 4
		counterBits = 2
		randomBits = randomLength - 14
		if randomBits < 2 {
			randomBits = 2
		}
	}

	// 计算各部分的值
	timestampVal := uint64(currentTime) % pow10Uint64(timestampBits)
	nodeVal := uint64(nodeID) % pow10Uint64(nodeBits)
	counterVal := counter % pow10Uint64(counterBits)
	randomVal := uint64(randomNum) % pow10Uint64(randomBits)

	// 格式化为数字字符串
	uniqueID := fmt.Sprintf("%0*d%0*d%0*d%0*d",
		timestampBits, timestampVal,
		nodeBits, nodeVal,
		counterBits, counterVal,
		randomBits, randomVal)

	// 如果长度超过，截取；如果不足，用计数器扩展填充
	if len(uniqueID) > randomLength {
		uniqueID = uniqueID[:randomLength]
	} else if len(uniqueID) < randomLength {
		// 用计数器扩展填充，确保唯一性
		padLen := randomLength - len(uniqueID)
		// 使用计数器的扩展部分生成填充字符
		padSeed := counter * 1000000 // 扩展计数器范围
		for i := 0; i < padLen; i++ {
			uniqueID += fmt.Sprintf("%d", padSeed%10)
			padSeed = padSeed/10 + counter // 确保每次不同
		}
	}

	// 将数字转换为字母数字混合格式
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, randomLength)
	for i := 0; i < len(uniqueID) && i < randomLength; i++ {
		digit := uniqueID[i] - '0'
		if i%3 == 0 && digit < 26 {
			result[i] = byte('A' + digit)
		} else if digit < 10 {
			result[i] = byte('0' + digit)
		} else {
			result[i] = charset[digit%36]
		}
	}

	// 组合前缀和随机部分
	if prefixLen == 0 {
		return string(result)
	}
	return prefix + string(result)
}

// pow10Uint64 计算10的n次方（返回uint64）
func pow10Uint64(n int) uint64 {
	result := uint64(1)
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}

// pow10 计算10的n次方
func pow10(n int) int64 {
	result := int64(1)
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}
