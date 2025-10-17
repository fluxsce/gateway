package models

import (
	"encoding/json"
	"time"
)

// JvmReportRequest JVM监控数据上报请求
type JvmReportRequest struct {
	// 应用标识信息
	JvmResourceId   string `json:"jvmResourceId" binding:"required"`   // JVM资源唯一标识（由应用端生成，如：app-name_host-ip_pid）
	ApplicationName string `json:"applicationName" binding:"required"` // 应用名称
	HostName        string `json:"hostName"`                           // 主机名
	HostIpAddress   string `json:"hostIpAddress"`                      // 主机IP地址

	// JVM资源信息
	JvmResourceInfo JvmResourceInfo `json:"jvmResourceInfo" binding:"required"` // JVM资源信息
}

// JvmResourceInfo JVM资源信息数据模型
type JvmResourceInfo struct {
	CollectionTime    int64             `json:"collectionTime"`    // 数据采集时间戳（毫秒）
	StartTime         int64             `json:"startTime"`         // JVM启动时间戳（毫秒）
	Uptime            int64             `json:"uptime"`            // JVM运行时长（毫秒）
	HeapMemory        *MemoryInfo       `json:"heapMemory"`        // 堆内存信息
	NonHeapMemory     *MemoryInfo       `json:"nonHeapMemory"`     // 非堆内存信息
	MemoryPools       []MemoryPoolInfo  `json:"memoryPools"`       // 内存池信息列表
	GarbageCollector  *GarbageCollector `json:"garbageCollector"`  // GC快照信息（汇总所有GC收集器）
	ThreadInfo        *ThreadInfo       `json:"threadInfo"`        // 线程信息
	ClassLoadingInfo  *ClassLoadingInfo `json:"classLoadingInfo"`  // 类加载信息
	SystemProperties  map[string]string `json:"systemProperties"`  // JVM系统属性
	Healthy           bool              `json:"healthy"`           // JVM整体健康状况
	HealthGrade       string            `json:"healthGrade"`       // JVM健康等级
	RequiresAttention bool              `json:"requiresAttention"` // 是否需要立即关注
	Summary           string            `json:"summary"`           // 监控摘要信息
}

// MemoryInfo 内存信息数据模型
type MemoryInfo struct {
	CollectionTime int64   `json:"collectionTime"` // 数据采集时间戳
	Init           int64   `json:"init"`           // 初始内存大小（字节）
	Used           int64   `json:"used"`           // 已使用内存大小（字节）
	Committed      int64   `json:"committed"`      // 已提交内存大小（字节）
	Max            int64   `json:"max"`            // 最大内存大小（字节）
	UsagePercent   float64 `json:"usagePercent"`   // 内存使用率（0.0-1.0）
}

// MemoryPoolInfo 内存池信息数据模型
type MemoryPoolInfo struct {
	CollectionTime                    int64       `json:"collectionTime"`                    // 数据采集时间戳
	Name                              string      `json:"name"`                              // 内存池名称
	Type                              string      `json:"type"`                              // 内存池类型（HEAP/NON_HEAP）
	Usage                             *MemoryInfo `json:"usage"`                             // 当前内存使用情况
	PeakUsage                         *MemoryInfo `json:"peakUsage"`                         // 峰值内存使用情况
	UsageThresholdSupported           bool        `json:"usageThresholdSupported"`           // 是否支持使用阈值监控
	UsageThreshold                    int64       `json:"usageThreshold"`                    // 使用阈值（字节）
	UsageThresholdCount               int64       `json:"usageThresholdCount"`               // 使用阈值超越次数
	CollectionUsageThresholdSupported bool        `json:"collectionUsageThresholdSupported"` // 是否支持收集使用量监控
}

// garbageCollector GC快照数据模型（jstat -gc 风格，汇总所有GC收集器）
// 与Java端 GarbageCollectorInfo 对应，每次采集上报一条汇总记录
type GarbageCollector struct {
	// 时间戳
	CollectionTime int64 `json:"collectionTime"` // 数据采集时间戳

	// GC累积统计（从JVM启动到当前采集时刻，所有GC收集器汇总）
	CollectionCount  int64 `json:"collectionCount"`  // GC总次数（累积）
	CollectionTimeMs int64 `json:"collectionTimeMs"` // GC总耗时（毫秒，累积）

	// ===== jstat -gc 风格的内存区域数据（单位：KB） =====

	// Survivor区
	S0C int64 `json:"s0c"` // Survivor 0 区容量（KB）
	S1C int64 `json:"s1c"` // Survivor 1 区容量（KB）
	S0U int64 `json:"s0u"` // Survivor 0 区使用量（KB）
	S1U int64 `json:"s1u"` // Survivor 1 区使用量（KB）

	// Eden区
	EC int64 `json:"ec"` // Eden 区容量（KB）
	EU int64 `json:"eu"` // Eden 区使用量（KB）

	// Old区
	OC int64 `json:"oc"` // Old 区容量（KB）
	OU int64 `json:"ou"` // Old 区使用量（KB）

	// Metaspace
	MC int64 `json:"mc"` // Metaspace 容量（KB）
	MU int64 `json:"mu"` // Metaspace 使用量（KB）

	// 压缩类空间
	CCSC int64 `json:"ccsc"` // 压缩类空间容量（KB）
	CCSU int64 `json:"ccsu"` // 压缩类空间使用量（KB）

	// GC统计（jstat -gc 格式）
	YGC  int64   `json:"ygc"`  // 年轻代GC次数
	YGCT float64 `json:"ygct"` // 年轻代GC总时间（秒）
	FGC  int64   `json:"fgc"`  // Full GC次数
	FGCT float64 `json:"fgct"` // Full GC总时间（秒）
	GCT  float64 `json:"gct"`  // 总GC时间（秒）
}

// ThreadInfo 线程信息数据模型
type ThreadInfo struct {
	CollectionTime             int64             `json:"collectionTime"`             // 数据采集时间戳
	ThreadCount                int               `json:"threadCount"`                // 当前线程数
	DaemonThreadCount          int               `json:"daemonThreadCount"`          // 守护线程数
	UserThreadCount            int               `json:"userThreadCount"`            // 用户线程数
	PeakThreadCount            int               `json:"peakThreadCount"`            // 峰值线程数
	TotalStartedThreadCount    int64             `json:"totalStartedThreadCount"`    // 总创建线程数
	ThreadStateStats           *ThreadStateStats `json:"threadStateStats"`           // 线程状态分布统计
	DeadlockInfo               *DeadlockInfo     `json:"deadlockInfo"`               // 死锁检测信息
	ThreadCpuTimeSupported     bool              `json:"threadCpuTimeSupported"`     // 线程CPU时间监控是否支持
	ThreadCpuTimeEnabled       bool              `json:"threadCpuTimeEnabled"`       // 线程CPU时间监控是否启用
	ThreadMemoryAllocSupported bool              `json:"threadMemoryAllocSupported"` // 线程内存分配监控是否支持
	ThreadMemoryAllocEnabled   bool              `json:"threadMemoryAllocEnabled"`   // 线程内存分配监控是否启用
	ThreadContentionSupported  bool              `json:"threadContentionSupported"`  // 线程争用监控是否支持
	ThreadContentionEnabled    bool              `json:"threadContentionEnabled"`    // 线程争用监控是否启用
	ThreadGrowthRate           float64           `json:"threadGrowthRate"`           // 线程增长率
	DaemonThreadRatio          float64           `json:"daemonThreadRatio"`          // 守护线程比例
	HealthGrade                string            `json:"healthGrade"`                // 线程健康等级
	Healthy                    bool              `json:"healthy"`                    // 是否健康
	PotentialIssues            []string          `json:"potentialIssues"`            // 潜在问题列表
}

// ThreadStateStats 线程状态统计数据模型
type ThreadStateStats struct {
	CollectionTime          int64 `json:"collectionTime"`          // 数据采集时间戳
	NewThreadCount          int   `json:"newThreadCount"`          // NEW状态线程数
	RunnableThreadCount     int   `json:"runnableThreadCount"`     // RUNNABLE状态线程数
	BlockedThreadCount      int   `json:"blockedThreadCount"`      // BLOCKED状态线程数
	WaitingThreadCount      int   `json:"waitingThreadCount"`      // WAITING状态线程数
	TimedWaitingThreadCount int   `json:"timedWaitingThreadCount"` // TIMED_WAITING状态线程数
	TerminatedThreadCount   int   `json:"terminatedThreadCount"`   // TERMINATED状态线程数
	TotalThreadCount        int   `json:"totalThreadCount"`        // 总线程数
}

// DeadlockInfo 死锁检测信息数据模型
type DeadlockInfo struct {
	CollectionTime       int64    `json:"collectionTime"`       // 数据采集时间戳
	HasDeadlock          bool     `json:"hasDeadlock"`          // 是否检测到死锁
	DeadlockThreadCount  int      `json:"deadlockThreadCount"`  // 死锁线程数量
	DeadlockThreadIds    []int64  `json:"deadlockThreadIds"`    // 死锁线程ID列表
	DeadlockThreadNames  []string `json:"deadlockThreadNames"`  // 死锁线程名称列表
	DetectionTime        int64    `json:"detectionTime"`        // 死锁检测时间戳
	DeadlockDuration     int64    `json:"deadlockDuration"`     // 死锁持续时间（毫秒）
	Severity             string   `json:"severity"`             // 死锁严重程度
	AffectedThreadGroups int      `json:"affectedThreadGroups"` // 死锁影响的线程组数量
	Description          string   `json:"description"`          // 死锁描述信息
	RecommendedAction    string   `json:"recommendedAction"`    // 建议的解决方案
}

// ClassLoadingInfo 类加载信息数据模型
type ClassLoadingInfo struct {
	CollectionTime        int64                    `json:"collectionTime"`        // 数据采集时间戳
	LoadedClassCount      int                      `json:"loadedClassCount"`      // 当前已加载类数量
	TotalLoadedClassCount int64                    `json:"totalLoadedClassCount"` // 总加载类数量
	UnloadedClassCount    int64                    `json:"unloadedClassCount"`    // 已卸载类数量
	ClassUnloadRate       float64                  `json:"classUnloadRate"`       // 类卸载率
	ClassRetentionRate    float64                  `json:"classRetentionRate"`    // 类保留率
	VerboseClassLoading   bool                     `json:"verboseClassLoading"`   // 是否启用了详细的类加载输出
	HealthGrade           string                   `json:"healthGrade"`           // 类加载健康等级
	Healthy               bool                     `json:"healthy"`               // 是否健康
	Performance           *ClassLoadingPerformance `json:"performance"`           // 类加载性能指标
	PotentialIssues       []string                 `json:"potentialIssues"`       // 潜在问题列表
}

// ClassLoadingPerformance 类加载性能指标
type ClassLoadingPerformance struct {
	LoadingRatePerHour float64 `json:"loadingRatePerHour"` // 每小时平均类加载数量
	LoadingEfficiency  float64 `json:"loadingEfficiency"`  // 类加载效率（加载成功率）
	MemoryEfficiency   string  `json:"memoryEfficiency"`   // 内存使用效率评估
	LoaderHealth       string  `json:"loaderHealth"`       // 类加载器健康状况
}

// ToCollectionTime 将时间戳转换为time.Time
func (j *JvmResourceInfo) ToCollectionTime() time.Time {
	return time.UnixMilli(j.CollectionTime)
}

// ToStartTime 将JVM启动时间戳转换为time.Time
func (j *JvmResourceInfo) ToStartTime() time.Time {
	return time.UnixMilli(j.StartTime)
}

// IsHealthy 判断内存是否健康
func (m *MemoryInfo) IsHealthy() bool {
	if m.Max <= 0 {
		// 最大内存为-1或0时，基于已提交内存判断
		if m.Committed > 0 {
			return float64(m.Used)/float64(m.Committed) < 0.8
		}
		return true
	}
	return m.UsagePercent < 0.8
}

// GetCategoryLabel 获取内存池的分类标签
// 支持各种GC类型的内存池：G1、Parallel、CMS、Serial、Shenandoah、ZGC等
func (m *MemoryPoolInfo) GetCategoryLabel() string {
	name := m.Name
	if name == "" {
		return "其他"
	}

	lowerName := toLower(name)

	// 1. 年轻代（Eden、Survivor）
	// 支持: G1 Eden Space, PS Eden Space, Par Eden Space, Eden Space, G1 Survivor Space, PS Survivor Space, etc.
	if contains(lowerName, []string{"eden", "survivor", "young", "nursery", "new generation"}) {
		return "年轻代"
	}

	// 2. 老年代（Old Gen、Tenured）
	// 支持: G1 Old Gen, PS Old Gen, CMS Old Gen, Tenured Gen, Old Generation, etc.
	if contains(lowerName, []string{"old gen", "old generation", "tenured", "cms old", "ps old"}) {
		return "老年代"
	}

	// 3. 元数据空间（Metaspace、Compressed Class Space）
	// JDK 8+: Metaspace 替代了 PermGen
	if containsSingle(lowerName, "metaspace") {
		return "元数据空间"
	}
	if contains(lowerName, []string{"compressed class", "class space"}) {
		return "元数据空间"
	}

	// 4. 代码缓存（Code Cache、CodeHeap分段）
	// JDK 9+ 将Code Cache分为三个区域：
	// - CodeHeap 'non-nmethods': 非方法代码（JVM内部使用）
	// - CodeHeap 'profiled nmethods': 轻度优化的JIT编译代码（C1编译器）
	// - CodeHeap 'non-profiled nmethods': 完全优化的JIT编译代码（C2编译器）
	if contains(lowerName, []string{"codeheap", "code heap", "code cache", "codecache"}) {
		return "代码缓存"
	}
	// 明确匹配带引号的CodeHeap名称
	if contains(lowerName, []string{"non-nmethods", "profiled nmethods", "non-profiled nmethods"}) {
		return "代码缓存"
	}

	// 5. 永久代（PermGen - JDK 7及以前版本）
	if contains(lowerName, []string{"perm gen", "permgen", "permanent generation", "perm space"}) {
		return "永久代"
	}

	// 6. ZGC特殊内存池
	if contains(lowerName, []string{"zgc", "z heap"}) {
		return "ZGC堆"
	}

	// 7. Shenandoah特殊内存池
	if containsSingle(lowerName, "shenandoah") {
		return "Shenandoah堆"
	}

	return "其他"
}

// contains 辅助函数：检查字符串是否包含任意关键词（假设str已经是小写）
func contains(str string, keywords []string) bool {
	for _, keyword := range keywords {
		if containsSingle(str, keyword) {
			return true
		}
	}
	return false
}

// containsSingle 辅助函数：检查字符串是否包含单个关键词
func containsSingle(str, keyword string) bool {
	// 简单实现，实际应使用strings.Contains
	for i := 0; i <= len(str)-len(keyword); i++ {
		if str[i:i+len(keyword)] == keyword {
			return true
		}
	}
	return false
}

// toLower 辅助函数：转换为小写
func toLower(str string) string {
	result := make([]byte, len(str))
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// ==========================================
// 应用监控数据模型（对应 HUB_MONITOR_APP_DATA 表）
// ==========================================

// AppReportRequest 应用监控数据上报请求
// 对应Java端ResourceReporter.createApplicationReportRequest方法的数据结构
type AppReportRequest struct {
	// 应用标识信息
	JvmResourceId   string `json:"jvmResourceId" binding:"required"`   // JVM资源唯一标识（由应用端生成，如：app-name_host-ip_pid）
	ApplicationName string `json:"applicationName" binding:"required"` // 应用名称
	HostName        string `json:"hostName"`                           // 主机名
	HostIpAddress   string `json:"hostIpAddress"`                      // 主机IP地址

	// JVM资源信息
	ApplicationResourceInfo ApplicationResourceInfo `json:"applicationResourceInfo" binding:"required"` // JVM资源信息
}

// ApplicationResourceInfo JVM资源信息数据模型
type ApplicationResourceInfo struct {
	CollectionTime       int64 `json:"collectionTime"`       // 数据采集时间戳（毫秒）
	Healthy              bool  `json:"healthy"`              // 整体健康状况
	ProviderCount        int   `json:"providerCount"`        // 提供者数量
	HealthyProviderCount int   `json:"healthyProviderCount"` // 健康提供者数量
	// 应用监控数据 - 可能是不同的结构
	ThirdPartyMonitorData []ThirdPartyMonitorData `json:"thirdPartyMonitorData" binding:"required"` // 第三方监控数据列表
}

// ThirdPartyMonitorData 第三方监控数据（对应Java端的ThirdPartyMonitorData类）
// 对应数据库表：HUB_MONITOR_APP_DATA
// JSON字段名称与数据库字段名称保持一致（非驼峰命名）
type ThirdPartyMonitorData struct {
	// 基本标识信息
	DataType     string `json:"dataType"`               // 监控类型（THREAD_POOL, CONNECTION_POOL, CUSTOM_METRIC等）
	DataName     string `json:"dataName"`               // 监控名称（如：线程池名称、指标名称等）
	DataCategory string `json:"dataCategory,omitempty"` // 数据分类（如：业务线程池/IO线程池）

	// 监控数据（具体的监控指标，通常是嵌套结构）
	DataJson interface{} `json:"dataJson"` // 监控数据，JSON格式，包含具体的监控指标和值

	// 核心指标（从DataJson中提取的关键指标，便于查询和索引）
	PrimaryValue   *float64 `json:"primaryValue,omitempty"`   // 主要指标值（如：使用率、数量等）
	SecondaryValue *float64 `json:"secondaryValue,omitempty"` // 次要指标值（如：最大值、平均值等）
	StatusValue    string   `json:"statusValue,omitempty"`    // 状态值（如：NORMAL, HIGH_LOAD, OVERLOADED等）

	// 时间信息
	CollectionTime int64 `json:"collectionTime"` // 采集时间戳（毫秒）

	// 健康状态
	HealthyFlag           string `json:"healthyFlag"`           // 健康标记（Y健康,N异常）
	HealthGrade           string `json:"healthGrade,omitempty"` // 健康等级（EXCELLENT/GOOD/FAIR/POOR/CRITICAL）
	RequiresAttentionFlag string `json:"requiresAttentionFlag"` // 是否需要立即关注（Y是,N否）

	// 标签和元数据
	TagsJson interface{} `json:"tagsJson,omitempty"` // 标签信息，JSON格式（可以是map[string]string或JSON字符串）
}

// ==========================================
// ThirdPartyMonitorData 辅助方法
// ==========================================

// ToCollectionTime 将时间戳转换为time.Time
func (t *ThirdPartyMonitorData) ToCollectionTime() time.Time {
	return time.UnixMilli(t.CollectionTime)
}

// RequiresAttention 判断是否需要关注
func (t *ThirdPartyMonitorData) RequiresAttention() bool {
	return t.RequiresAttentionFlag == "Y"
}

// GetDataJsonAsString 将DataJson转换为JSON字符串用于存储
func (t *ThirdPartyMonitorData) GetDataJsonAsString() string {
	if t.DataJson == nil {
		return "{}"
	}

	// 如果已经是字符串，直接返回
	if str, ok := t.DataJson.(string); ok {
		return str
	}

	// 否则序列化为JSON
	if jsonBytes, err := json.Marshal(t.DataJson); err == nil {
		return string(jsonBytes)
	}

	return "{}"
}

// GetTagsJsonAsString 将TagsJson转换为JSON字符串用于存储
func (t *ThirdPartyMonitorData) GetTagsJsonAsString() string {
	if t.TagsJson == nil {
		return ""
	}

	// 如果已经是字符串，直接返回
	if str, ok := t.TagsJson.(string); ok {
		return str
	}

	// 否则序列化为JSON
	if jsonBytes, err := json.Marshal(t.TagsJson); err == nil {
		return string(jsonBytes)
	}

	return ""
}

// ExtractPrimaryValue 从DataJson中提取主要指标值
// 根据不同的DataType提取不同的指标
func (t *ThirdPartyMonitorData) ExtractPrimaryValue() *float64 {
	if t.PrimaryValue != nil {
		return t.PrimaryValue
	}

	// 如果没有显式设置，尝试从DataJson中提取
	if t.DataJson == nil {
		return nil
	}

	dataMap, ok := t.DataJson.(map[string]interface{})
	if !ok {
		return nil
	}

	// 根据DataType提取不同的主要指标
	var key string
	switch t.DataType {
	case "THREAD_POOL":
		key = "activeThreadRatioPercent" // 活跃线程比例
	case "CONNECTION_POOL":
		key = "activeConnectionRatioPercent" // 活跃连接比例
	case "CUSTOM_METRIC":
		key = "metricValue" // 指标值
	default:
		// 尝试通用字段
		if val, exists := dataMap["usagePercent"]; exists {
			if floatVal, ok := val.(float64); ok {
				return &floatVal
			}
		}
		return nil
	}

	if val, exists := dataMap[key]; exists {
		if floatVal, ok := val.(float64); ok {
			return &floatVal
		}
	}

	return nil
}

// ExtractSecondaryValue 从DataJson中提取次要指标值
func (t *ThirdPartyMonitorData) ExtractSecondaryValue() *float64 {
	if t.SecondaryValue != nil {
		return t.SecondaryValue
	}

	if t.DataJson == nil {
		return nil
	}

	dataMap, ok := t.DataJson.(map[string]interface{})
	if !ok {
		return nil
	}

	// 根据DataType提取不同的次要指标
	var key string
	switch t.DataType {
	case "THREAD_POOL":
		key = "queueUsageRatioPercent" // 队列使用比例
	case "CONNECTION_POOL":
		key = "poolUtilizationPercent" // 连接池利用率
	case "CUSTOM_METRIC":
		key = "metricAvg" // 平均值
	default:
		return nil
	}

	if val, exists := dataMap[key]; exists {
		if floatVal, ok := val.(float64); ok {
			return &floatVal
		}
	}

	return nil
}

// ExtractStatusValue 从DataJson中提取状态值
func (t *ThirdPartyMonitorData) ExtractStatusValue() string {
	if t.StatusValue != "" {
		return t.StatusValue
	}

	if t.DataJson == nil {
		return ""
	}

	dataMap, ok := t.DataJson.(map[string]interface{})
	if !ok {
		return ""
	}

	// 尝试从常见字段中提取状态
	statusKeys := []string{"status", "poolStatus", "connectionStatus", "state"}
	for _, key := range statusKeys {
		if val, exists := dataMap[key]; exists {
			if strVal, ok := val.(string); ok {
				return strVal
			}
		}
	}

	return ""
}
