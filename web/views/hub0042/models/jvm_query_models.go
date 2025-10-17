package models

import "time"

// ===============================
// 通用查询请求和响应
// ===============================

// BaseQueryRequest 通用查询请求
type BaseQueryRequest struct {
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId"`                            // 租户ID
	ServiceGroupId    string `json:"serviceGroupId" form:"serviceGroupId" query:"serviceGroupId"`          // 服务分组ID
	GroupName         string `json:"groupName" form:"groupName" query:"groupName"`                         // 分组名称
	JvmResourceId     string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId"`             // JVM资源ID
	ApplicationName   string `json:"applicationName" form:"applicationName" query:"applicationName"`       // 应用名称
	HostIpAddress     string `json:"hostIpAddress" form:"hostIpAddress" query:"hostIpAddress"`             // 主机IP
	StartTime         string `json:"startTime" form:"startTime" query:"startTime"`                         // 开始时间（字符串格式，支持: 2006-01-02 15:04:05）
	EndTime           string `json:"endTime" form:"endTime" query:"endTime"`                               // 结束时间（字符串格式，支持: 2006-01-02 15:04:05）
	HealthyFlag       string `json:"healthyFlag" form:"healthyFlag" query:"healthyFlag"`                   // 健康标记(Y/N)
	PageNum           int    `json:"pageNum" form:"pageNum" query:"pageNum"`                               // 页码
	PageSize          int    `json:"pageSize" form:"pageSize" query:"pageSize"`                            // 每页大小
	OrderBy           string `json:"orderBy" form:"orderBy" query:"orderBy"`                               // 排序字段
	OrderDirection    string `json:"orderDirection" form:"orderDirection" query:"orderDirection"`          // 排序方向(ASC/DESC)
	RequiresAttention string `json:"requiresAttention" form:"requiresAttention" query:"requiresAttention"` // 是否需要关注
}

// PageInfo 分页信息
type PageInfo struct {
	PageNum    int   `json:"pageNum"`    // 当前页码
	PageSize   int   `json:"pageSize"`   // 每页大小
	TotalCount int64 `json:"totalCount"` // 总记录数
	TotalPages int   `json:"totalPages"` // 总页数
}

// ===============================
// 1. JVM资源监控查询 (HUB_MONITOR_JVM_RESOURCE)
// ===============================

// JvmResourceQueryRequest JVM资源查询请求
type JvmResourceQueryRequest struct {
	BaseQueryRequest
}

// JvmResourceResponse JVM资源响应 (对应 HUB_MONITOR_JVM_RESOURCE 表)
type JvmResourceResponse struct {
	// 主键和租户信息
	JvmResourceId  string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId" db:"jvmResourceId"`     // JVM资源ID（应用生成）
	TenantId       string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                         // 租户ID
	ServiceGroupId string `json:"serviceGroupId" form:"serviceGroupId" query:"serviceGroupId" db:"serviceGroupId"` // 服务分组ID

	// 应用标识信息
	ApplicationName string `json:"applicationName" form:"applicationName" query:"applicationName" db:"applicationName"` // 应用名称
	GroupName       string `json:"groupName" form:"groupName" query:"groupName" db:"groupName"`                         // 分组名称
	HostName        string `json:"hostName" form:"hostName" query:"hostName" db:"hostName"`                             // 主机名
	HostIpAddress   string `json:"hostIpAddress" form:"hostIpAddress" query:"hostIpAddress" db:"hostIpAddress"`         // 主机IP

	// 时间相关字段
	CollectionTime time.Time `json:"collectionTime" form:"collectionTime" query:"collectionTime" db:"collectionTime"` // 数据采集时间
	JvmStartTime   time.Time `json:"jvmStartTime" form:"jvmStartTime" query:"jvmStartTime" db:"jvmStartTime"`         // JVM启动时间
	JvmUptimeMs    int64     `json:"jvmUptimeMs" form:"jvmUptimeMs" query:"jvmUptimeMs" db:"jvmUptimeMs"`             // JVM运行时长(毫秒)

	// 健康状态字段
	HealthyFlag           string `json:"healthyFlag" form:"healthyFlag" query:"healthyFlag" db:"healthyFlag"`                                         // 整体健康标记(Y健康,N异常)
	HealthGrade           string `json:"healthGrade" form:"healthGrade" query:"healthGrade" db:"healthGrade"`                                         // 健康等级(EXCELLENT/GOOD/FAIR/POOR)
	RequiresAttentionFlag string `json:"requiresAttentionFlag" form:"requiresAttentionFlag" query:"requiresAttentionFlag" db:"requiresAttentionFlag"` // 是否需要立即关注(Y是,N否)
	SummaryText           string `json:"summaryText" form:"summaryText" query:"summaryText" db:"summaryText"`                                         // 监控摘要信息

	// 系统属性
	SystemPropertiesJson string `json:"systemPropertiesJson" form:"systemPropertiesJson" query:"systemPropertiesJson" db:"systemPropertiesJson"` // JVM系统属性JSON

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// JvmResourceListResponse JVM资源列表响应
type JvmResourceListResponse struct {
	PageInfo PageInfo              `json:"pageInfo"` // 分页信息
	List     []JvmResourceResponse `json:"list"`     // 数据列表
}

// ===============================
// 2. 内存信息查询 (HUB_MONITOR_JVM_MEMORY)
// ===============================

// MemoryQueryRequest 内存查询请求
type MemoryQueryRequest struct {
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId"`                // 租户ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId"` // JVM资源ID
	MemoryType    string `json:"memoryType" form:"memoryType" query:"memoryType"`          // 内存类型(HEAP/NON_HEAP)
	StartTime     string `json:"startTime" form:"startTime" query:"startTime"`             // 开始时间（字符串格式）
	EndTime       string `json:"endTime" form:"endTime" query:"endTime"`                   // 结束时间（字符串格式）
	Limit         int    `json:"limit" form:"limit" query:"limit"`                         // 查询数量限制
}

// MemoryResponse 内存响应 (对应 HUB_MONITOR_JVM_MEMORY 表)
type MemoryResponse struct {
	// 主键和关联字段
	JvmMemoryId   string `json:"jvmMemoryId" form:"jvmMemoryId" query:"jvmMemoryId" db:"jvmMemoryId"`         // 内存记录ID
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                     // 租户ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId" db:"jvmResourceId"` // 关联的JVM资源ID

	// 内存类型
	MemoryType string `json:"memoryType" form:"memoryType" query:"memoryType" db:"memoryType"` // 内存类型(HEAP/NON_HEAP)

	// 内存使用情况（字节）
	InitMemoryBytes      int64 `json:"initMemoryBytes" form:"initMemoryBytes" query:"initMemoryBytes" db:"initMemoryBytes"`                     // 初始内存(字节)
	UsedMemoryBytes      int64 `json:"usedMemoryBytes" form:"usedMemoryBytes" query:"usedMemoryBytes" db:"usedMemoryBytes"`                     // 已使用内存(字节)
	CommittedMemoryBytes int64 `json:"committedMemoryBytes" form:"committedMemoryBytes" query:"committedMemoryBytes" db:"committedMemoryBytes"` // 已提交内存(字节)
	MaxMemoryBytes       int64 `json:"maxMemoryBytes" form:"maxMemoryBytes" query:"maxMemoryBytes" db:"maxMemoryBytes"`                         // 最大内存(字节)

	// 计算指标
	UsagePercent float64 `json:"usagePercent" form:"usagePercent" query:"usagePercent" db:"usagePercent"` // 使用率(百分比)
	HealthyFlag  string  `json:"healthyFlag" form:"healthyFlag" query:"healthyFlag" db:"healthyFlag"`     // 健康标记(Y健康,N异常)

	// 时间字段
	CollectionTime time.Time `json:"collectionTime" form:"collectionTime" query:"collectionTime" db:"collectionTime"` // 采集时间

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// ===============================
// 3. 内存池查询 (HUB_MONITOR_JVM_MEM_POOL)
// ===============================

// MemoryPoolQueryRequest 内存池查询请求
type MemoryPoolQueryRequest struct {
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId"`                // 租户ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId"` // JVM资源ID
	PoolType      string `json:"poolType" form:"poolType" query:"poolType"`                // 内存池类型(HEAP/NON_HEAP)
	PoolCategory  string `json:"poolCategory" form:"poolCategory" query:"poolCategory"`    // 内存池分类
	StartTime     string `json:"startTime" form:"startTime" query:"startTime"`             // 开始时间（字符串格式）
	EndTime       string `json:"endTime" form:"endTime" query:"endTime"`                   // 结束时间（字符串格式）
	Limit         int    `json:"limit" form:"limit" query:"limit"`                         // 查询数量限制
}

// MemoryPoolResponse 内存池响应 (对应 HUB_MONITOR_JVM_MEM_POOL 表)
type MemoryPoolResponse struct {
	// 主键和关联字段
	MemoryPoolId  string `json:"memoryPoolId" form:"memoryPoolId" query:"memoryPoolId" db:"memoryPoolId"`     // 内存池ID
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                     // 租户ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId" db:"jvmResourceId"` // 关联的JVM资源ID

	// 内存池基本信息
	PoolName     string `json:"poolName" form:"poolName" query:"poolName" db:"poolName"`                 // 内存池名称
	PoolType     string `json:"poolType" form:"poolType" query:"poolType" db:"poolType"`                 // 内存池类型(HEAP/NON_HEAP)
	PoolCategory string `json:"poolCategory" form:"poolCategory" query:"poolCategory" db:"poolCategory"` // 内存池分类

	// 当前使用情况
	CurrentInitBytes      int64   `json:"currentInitBytes" form:"currentInitBytes" query:"currentInitBytes" db:"currentInitBytes"`                     // 当前初始内存(字节)
	CurrentUsedBytes      int64   `json:"currentUsedBytes" form:"currentUsedBytes" query:"currentUsedBytes" db:"currentUsedBytes"`                     // 当前已使用内存(字节)
	CurrentCommittedBytes int64   `json:"currentCommittedBytes" form:"currentCommittedBytes" query:"currentCommittedBytes" db:"currentCommittedBytes"` // 当前已提交内存(字节)
	CurrentMaxBytes       int64   `json:"currentMaxBytes" form:"currentMaxBytes" query:"currentMaxBytes" db:"currentMaxBytes"`                         // 当前最大内存(字节)
	CurrentUsagePercent   float64 `json:"currentUsagePercent" form:"currentUsagePercent" query:"currentUsagePercent" db:"currentUsagePercent"`         // 当前使用率(百分比)

	// 峰值使用情况
	PeakInitBytes      int64   `json:"peakInitBytes" form:"peakInitBytes" query:"peakInitBytes" db:"peakInitBytes"`                     // 峰值初始内存(字节)
	PeakUsedBytes      int64   `json:"peakUsedBytes" form:"peakUsedBytes" query:"peakUsedBytes" db:"peakUsedBytes"`                     // 峰值已使用内存(字节)
	PeakCommittedBytes int64   `json:"peakCommittedBytes" form:"peakCommittedBytes" query:"peakCommittedBytes" db:"peakCommittedBytes"` // 峰值已提交内存(字节)
	PeakMaxBytes       int64   `json:"peakMaxBytes" form:"peakMaxBytes" query:"peakMaxBytes" db:"peakMaxBytes"`                         // 峰值最大内存(字节)
	PeakUsagePercent   float64 `json:"peakUsagePercent" form:"peakUsagePercent" query:"peakUsagePercent" db:"peakUsagePercent"`         // 峰值使用率(百分比)

	// 阈值监控
	UsageThresholdSupported  string `json:"usageThresholdSupported" form:"usageThresholdSupported" query:"usageThresholdSupported" db:"usageThresholdSupported"`     // 是否支持使用阈值监控(Y是,N否)
	UsageThresholdBytes      int64  `json:"usageThresholdBytes" form:"usageThresholdBytes" query:"usageThresholdBytes" db:"usageThresholdBytes"`                     // 使用阈值(字节)
	UsageThresholdCount      int64  `json:"usageThresholdCount" form:"usageThresholdCount" query:"usageThresholdCount" db:"usageThresholdCount"`                     // 使用阈值超越次数
	CollectionUsageSupported string `json:"collectionUsageSupported" form:"collectionUsageSupported" query:"collectionUsageSupported" db:"collectionUsageSupported"` // 是否支持收集使用量监控(Y是,N否)

	// 健康状态
	HealthyFlag string `json:"healthyFlag" form:"healthyFlag" query:"healthyFlag" db:"healthyFlag"` // 内存池健康标记(Y健康,N异常)

	// 时间字段
	CollectionTime time.Time `json:"collectionTime" form:"collectionTime" query:"collectionTime" db:"collectionTime"` // 采集时间

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// ===============================
// 4. GC快照查询 (HUB_MONITOR_JVM_GC)
// ===============================

// GCSnapshotQueryRequest GC快照查询请求
type GCSnapshotQueryRequest struct {
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId"`                // 租户ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId"` // JVM资源ID（必填）
	StartTime     string `json:"startTime" form:"startTime" query:"startTime"`             // 开始时间（字符串格式）
	EndTime       string `json:"endTime" form:"endTime" query:"endTime"`                   // 结束时间（字符串格式）
	Limit         int    `json:"limit" form:"limit" query:"limit"`                         // 查询数量限制
}

// GCSnapshotResponse GC快照响应 (对应 HUB_MONITOR_JVM_GC 表)
type GCSnapshotResponse struct {
	// 主键和关联字段
	GcSnapshotId  string `json:"gcSnapshotId" form:"gcSnapshotId" query:"gcSnapshotId" db:"gcSnapshotId"`     // GC快照ID
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                     // 租户ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId" db:"jvmResourceId"` // 关联的JVM资源ID

	// GC累积统计
	CollectionCount  int64 `json:"collectionCount" form:"collectionCount" query:"collectionCount" db:"collectionCount"`     // GC总次数(累积)
	CollectionTimeMs int64 `json:"collectionTimeMs" form:"collectionTimeMs" query:"collectionTimeMs" db:"collectionTimeMs"` // GC总耗时(毫秒,累积)

	// Survivor区 (jstat -gc 风格, 单位: KB)
	S0C int64 `json:"s0c" form:"s0c" query:"s0c" db:"s0c"` // Survivor 0 容量(KB)
	S1C int64 `json:"s1c" form:"s1c" query:"s1c" db:"s1c"` // Survivor 1 容量(KB)
	S0U int64 `json:"s0u" form:"s0u" query:"s0u" db:"s0u"` // Survivor 0 使用量(KB)
	S1U int64 `json:"s1u" form:"s1u" query:"s1u" db:"s1u"` // Survivor 1 使用量(KB)

	// Eden区
	EC int64 `json:"ec" form:"ec" query:"ec" db:"ec"` // Eden 容量(KB)
	EU int64 `json:"eu" form:"eu" query:"eu" db:"eu"` // Eden 使用量(KB)

	// Old区
	OC int64 `json:"oc" form:"oc" query:"oc" db:"oc"` // Old 容量(KB)
	OU int64 `json:"ou" form:"ou" query:"ou" db:"ou"` // Old 使用量(KB)

	// Metaspace
	MC int64 `json:"mc" form:"mc" query:"mc" db:"mc"` // Metaspace 容量(KB)
	MU int64 `json:"mu" form:"mu" query:"mu" db:"mu"` // Metaspace 使用量(KB)

	// 压缩类空间
	CCSC int64 `json:"ccsc" form:"ccsc" query:"ccsc" db:"ccsc"` // 压缩类空间容量(KB)
	CCSU int64 `json:"ccsu" form:"ccsu" query:"ccsu" db:"ccsu"` // 压缩类空间使用量(KB)

	// GC统计 (jstat -gc 格式)
	YGC  int64   `json:"ygc" form:"ygc" query:"ygc" db:"ygc"`     // 年轻代GC次数
	YGCT float64 `json:"ygct" form:"ygct" query:"ygct" db:"ygct"` // 年轻代GC总时间(秒)
	FGC  int64   `json:"fgc" form:"fgc" query:"fgc" db:"fgc"`     // Full GC次数
	FGCT float64 `json:"fgct" form:"fgct" query:"fgct" db:"fgct"` // Full GC总时间(秒)
	GCT  float64 `json:"gct" form:"gct" query:"gct" db:"gct"`     // 总GC时间(秒)

	// 时间戳信息
	CollectionTime time.Time `json:"collectionTime" form:"collectionTime" query:"collectionTime" db:"collectionTime"` // 采集时间

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// GCTrendAnalysisResponse GC趋势分析响应 (计算字段，非数据库表)
type GCTrendAnalysisResponse struct {
	CollectionTime    time.Time `json:"collectionTime"`    // 采集时间
	GcCountIncrease   int64     `json:"gcCountIncrease"`   // GC次数增量
	GcTimeIncrease    int64     `json:"gcTimeIncrease"`    // GC耗时增量(毫秒)
	YgcIncrease       int64     `json:"ygcIncrease"`       // 年轻代GC增量
	FgcIncrease       int64     `json:"fgcIncrease"`       // Full GC增量
	IntervalSeconds   int64     `json:"intervalSeconds"`   // 采集间隔(秒)
	GcFrequencyPerMin float64   `json:"gcFrequencyPerMin"` // GC频率(次/分钟)
}

// ===============================
// 5. 线程监控查询 (HUB_MONITOR_JVM_THREAD)
// ===============================

// ThreadQueryRequest 线程查询请求
type ThreadQueryRequest struct {
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId"`                // 租户ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId"` // JVM资源ID
	HealthyFlag   string `json:"healthyFlag" form:"healthyFlag" query:"healthyFlag"`       // 健康标记
	StartTime     string `json:"startTime" form:"startTime" query:"startTime"`             // 开始时间（字符串格式）
	EndTime       string `json:"endTime" form:"endTime" query:"endTime"`                   // 结束时间（字符串格式）
	Limit         int    `json:"limit" form:"limit" query:"limit"`                         // 查询数量限制
}

// ThreadResponse 线程响应 (对应 HUB_MONITOR_JVM_THREAD 表)
type ThreadResponse struct {
	// 主键和关联字段
	JvmThreadId   string `json:"jvmThreadId" form:"jvmThreadId" query:"jvmThreadId" db:"jvmThreadId"`         // 线程记录ID
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                     // 租户ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId" db:"jvmResourceId"` // 关联的JVM资源ID

	// 基础线程统计
	CurrentThreadCount      int   `json:"currentThreadCount" form:"currentThreadCount" query:"currentThreadCount" db:"currentThreadCount"`                     // 当前线程数
	DaemonThreadCount       int   `json:"daemonThreadCount" form:"daemonThreadCount" query:"daemonThreadCount" db:"daemonThreadCount"`                         // 守护线程数
	UserThreadCount         int   `json:"userThreadCount" form:"userThreadCount" query:"userThreadCount" db:"userThreadCount"`                                 // 用户线程数
	PeakThreadCount         int   `json:"peakThreadCount" form:"peakThreadCount" query:"peakThreadCount" db:"peakThreadCount"`                                 // 峰值线程数
	TotalStartedThreadCount int64 `json:"totalStartedThreadCount" form:"totalStartedThreadCount" query:"totalStartedThreadCount" db:"totalStartedThreadCount"` // 总启动线程数

	// 性能指标
	ThreadGrowthRatePercent  float64 `json:"threadGrowthRatePercent" form:"threadGrowthRatePercent" query:"threadGrowthRatePercent" db:"threadGrowthRatePercent"`     // 线程增长率(百分比)
	DaemonThreadRatioPercent float64 `json:"daemonThreadRatioPercent" form:"daemonThreadRatioPercent" query:"daemonThreadRatioPercent" db:"daemonThreadRatioPercent"` // 守护线程比例(百分比)

	// 监控功能支持状态
	CpuTimeSupported     string `json:"cpuTimeSupported" form:"cpuTimeSupported" query:"cpuTimeSupported" db:"cpuTimeSupported"`                 // CPU时间监控是否支持(Y是,N否)
	CpuTimeEnabled       string `json:"cpuTimeEnabled" form:"cpuTimeEnabled" query:"cpuTimeEnabled" db:"cpuTimeEnabled"`                         // CPU时间监控是否启用(Y是,N否)
	MemoryAllocSupported string `json:"memoryAllocSupported" form:"memoryAllocSupported" query:"memoryAllocSupported" db:"memoryAllocSupported"` // 内存分配监控是否支持(Y是,N否)
	MemoryAllocEnabled   string `json:"memoryAllocEnabled" form:"memoryAllocEnabled" query:"memoryAllocEnabled" db:"memoryAllocEnabled"`         // 内存分配监控是否启用(Y是,N否)
	ContentionSupported  string `json:"contentionSupported" form:"contentionSupported" query:"contentionSupported" db:"contentionSupported"`     // 争用监控是否支持(Y是,N否)
	ContentionEnabled    string `json:"contentionEnabled" form:"contentionEnabled" query:"contentionEnabled" db:"contentionEnabled"`             // 争用监控是否启用(Y是,N否)

	// 健康状态
	HealthyFlag           string `json:"healthyFlag" form:"healthyFlag" query:"healthyFlag" db:"healthyFlag"`                                         // 线程健康标记(Y健康,N异常)
	HealthGrade           string `json:"healthGrade" form:"healthGrade" query:"healthGrade" db:"healthGrade"`                                         // 线程健康等级(EXCELLENT/GOOD/FAIR/POOR)
	RequiresAttentionFlag string `json:"requiresAttentionFlag" form:"requiresAttentionFlag" query:"requiresAttentionFlag" db:"requiresAttentionFlag"` // 是否需要立即关注(Y是,N否)
	PotentialIssuesJson   string `json:"potentialIssuesJson" form:"potentialIssuesJson" query:"potentialIssuesJson" db:"potentialIssuesJson"`         // 潜在问题列表JSON

	// 时间字段
	CollectionTime time.Time `json:"collectionTime" form:"collectionTime" query:"collectionTime" db:"collectionTime"` // 采集时间

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// ===============================
// 6. 线程状态统计查询 (HUB_MONITOR_JVM_THR_STATE)
// ===============================

// ThreadStateQueryRequest 线程状态查询请求
type ThreadStateQueryRequest struct {
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId"`                // 租户ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId"` // JVM资源ID
	JvmThreadId   string `json:"jvmThreadId" form:"jvmThreadId" query:"jvmThreadId"`       // JVM线程ID
	StartTime     string `json:"startTime" form:"startTime" query:"startTime"`             // 开始时间（字符串格式）
	EndTime       string `json:"endTime" form:"endTime" query:"endTime"`                   // 结束时间（字符串格式）
	Limit         int    `json:"limit" form:"limit" query:"limit"`                         // 查询数量限制
}

// ThreadStateResponse 线程状态响应 (对应 HUB_MONITOR_JVM_THR_STATE 表)
type ThreadStateResponse struct {
	// 主键和关联字段
	ThreadStateId string `json:"threadStateId" form:"threadStateId" query:"threadStateId" db:"threadStateId"` // 线程状态记录ID
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                     // 租户ID
	JvmThreadId   string `json:"jvmThreadId" form:"jvmThreadId" query:"jvmThreadId" db:"jvmThreadId"`         // 关联的JVM线程记录ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId" db:"jvmResourceId"` // 关联的JVM资源ID

	// 线程状态分布
	NewThreadCount          int `json:"newThreadCount" form:"newThreadCount" query:"newThreadCount" db:"newThreadCount"`                                     // NEW状态线程数
	RunnableThreadCount     int `json:"runnableThreadCount" form:"runnableThreadCount" query:"runnableThreadCount" db:"runnableThreadCount"`                 // RUNNABLE状态线程数
	BlockedThreadCount      int `json:"blockedThreadCount" form:"blockedThreadCount" query:"blockedThreadCount" db:"blockedThreadCount"`                     // BLOCKED状态线程数
	WaitingThreadCount      int `json:"waitingThreadCount" form:"waitingThreadCount" query:"waitingThreadCount" db:"waitingThreadCount"`                     // WAITING状态线程数
	TimedWaitingThreadCount int `json:"timedWaitingThreadCount" form:"timedWaitingThreadCount" query:"timedWaitingThreadCount" db:"timedWaitingThreadCount"` // TIMED_WAITING状态线程数
	TerminatedThreadCount   int `json:"terminatedThreadCount" form:"terminatedThreadCount" query:"terminatedThreadCount" db:"terminatedThreadCount"`         // TERMINATED状态线程数
	TotalThreadCount        int `json:"totalThreadCount" form:"totalThreadCount" query:"totalThreadCount" db:"totalThreadCount"`                             // 总线程数

	// 比例指标
	ActiveThreadRatioPercent  float64 `json:"activeThreadRatioPercent" form:"activeThreadRatioPercent" query:"activeThreadRatioPercent" db:"activeThreadRatioPercent"`     // 活跃线程比例(百分比)
	BlockedThreadRatioPercent float64 `json:"blockedThreadRatioPercent" form:"blockedThreadRatioPercent" query:"blockedThreadRatioPercent" db:"blockedThreadRatioPercent"` // 阻塞线程比例(百分比)
	WaitingThreadRatioPercent float64 `json:"waitingThreadRatioPercent" form:"waitingThreadRatioPercent" query:"waitingThreadRatioPercent" db:"waitingThreadRatioPercent"` // 等待状态线程比例(百分比)

	// 健康状态
	HealthyFlag string `json:"healthyFlag" form:"healthyFlag" query:"healthyFlag" db:"healthyFlag"` // 线程状态健康标记(Y健康,N异常)
	HealthGrade string `json:"healthGrade" form:"healthGrade" query:"healthGrade" db:"healthGrade"` // 健康等级(EXCELLENT/GOOD/FAIR/POOR)

	// 时间字段
	CollectionTime time.Time `json:"collectionTime" form:"collectionTime" query:"collectionTime" db:"collectionTime"` // 采集时间

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// ===============================
// 7. 死锁查询 (HUB_MONITOR_JVM_DEADLOCK)
// ===============================

// DeadlockQueryRequest 死锁查询请求
type DeadlockQueryRequest struct {
	TenantId        string `json:"tenantId" form:"tenantId" query:"tenantId"`                      // 租户ID
	JvmResourceId   string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId"`       // JVM资源ID
	HasDeadlockFlag string `json:"hasDeadlockFlag" form:"hasDeadlockFlag" query:"hasDeadlockFlag"` // 是否有死锁(Y/N)
	SeverityLevel   string `json:"severityLevel" form:"severityLevel" query:"severityLevel"`       // 严重程度
	StartTime       string `json:"startTime" form:"startTime" query:"startTime"`                   // 开始时间（字符串格式）
	EndTime         string `json:"endTime" form:"endTime" query:"endTime"`                         // 结束时间（字符串格式）
	Limit           int    `json:"limit" form:"limit" query:"limit"`                               // 查询数量限制
}

// DeadlockResponse 死锁响应 (对应 HUB_MONITOR_JVM_DEADLOCK 表)
type DeadlockResponse struct {
	// 主键和关联字段
	DeadlockId    string `json:"deadlockId" form:"deadlockId" query:"deadlockId" db:"deadlockId"`             // 死锁ID
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                     // 租户ID
	JvmThreadId   string `json:"jvmThreadId" form:"jvmThreadId" query:"jvmThreadId" db:"jvmThreadId"`         // 关联的JVM线程记录ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId" db:"jvmResourceId"` // 关联的JVM资源ID

	// 死锁基本信息
	HasDeadlockFlag     string `json:"hasDeadlockFlag" form:"hasDeadlockFlag" query:"hasDeadlockFlag" db:"hasDeadlockFlag"`                 // 是否检测到死锁(Y是,N否)
	DeadlockThreadCount int    `json:"deadlockThreadCount" form:"deadlockThreadCount" query:"deadlockThreadCount" db:"deadlockThreadCount"` // 死锁线程数量
	DeadlockThreadIds   string `json:"deadlockThreadIds" form:"deadlockThreadIds" query:"deadlockThreadIds" db:"deadlockThreadIds"`         // 死锁线程ID列表(逗号分隔)
	DeadlockThreadNames string `json:"deadlockThreadNames" form:"deadlockThreadNames" query:"deadlockThreadNames" db:"deadlockThreadNames"` // 死锁线程名称列表(逗号分隔)

	// 死锁严重程度
	SeverityLevel        string `json:"severityLevel" form:"severityLevel" query:"severityLevel" db:"severityLevel"`                             // 严重程度(LOW/MEDIUM/HIGH/CRITICAL)
	SeverityDescription  string `json:"severityDescription" form:"severityDescription" query:"severityDescription" db:"severityDescription"`     // 严重程度描述
	AffectedThreadGroups int    `json:"affectedThreadGroups" form:"affectedThreadGroups" query:"affectedThreadGroups" db:"affectedThreadGroups"` // 影响的线程组数量

	// 时间信息
	DetectionTime      time.Time `json:"detectionTime" form:"detectionTime" query:"detectionTime" db:"detectionTime"`                     // 死锁检测时间
	DeadlockDurationMs int64     `json:"deadlockDurationMs" form:"deadlockDurationMs" query:"deadlockDurationMs" db:"deadlockDurationMs"` // 死锁持续时间(毫秒)
	CollectionTime     time.Time `json:"collectionTime" form:"collectionTime" query:"collectionTime" db:"collectionTime"`                 // 数据采集时间

	// 诊断信息
	DescriptionText    string `json:"descriptionText" form:"descriptionText" query:"descriptionText" db:"descriptionText"`             // 死锁描述信息
	RecommendedAction  string `json:"recommendedAction" form:"recommendedAction" query:"recommendedAction" db:"recommendedAction"`     // 建议的解决方案
	AlertLevel         string `json:"alertLevel" form:"alertLevel" query:"alertLevel" db:"alertLevel"`                                 // 告警级别(INFO/WARNING/ERROR/CRITICAL/EMERGENCY)
	RequiresActionFlag string `json:"requiresActionFlag" form:"requiresActionFlag" query:"requiresActionFlag" db:"requiresActionFlag"` // 是否需要立即处理(Y是,N否)

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// ===============================
// 8. 类加载监控查询 (HUB_MONITOR_JVM_CLASS)
// ===============================

// ClassLoadingQueryRequest 类加载查询请求
type ClassLoadingQueryRequest struct {
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId"`                // 租户ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId"` // JVM资源ID
	HealthyFlag   string `json:"healthyFlag" form:"healthyFlag" query:"healthyFlag"`       // 健康标记
	StartTime     string `json:"startTime" form:"startTime" query:"startTime"`             // 开始时间（字符串格式）
	EndTime       string `json:"endTime" form:"endTime" query:"endTime"`                   // 结束时间（字符串格式）
	Limit         int    `json:"limit" form:"limit" query:"limit"`                         // 查询数量限制
}

// ClassLoadingResponse 类加载响应 (对应 HUB_MONITOR_JVM_CLASS 表)
type ClassLoadingResponse struct {
	// 主键和关联字段
	ClassLoadingId string `json:"classLoadingId" form:"classLoadingId" query:"classLoadingId" db:"classLoadingId"` // 类加载ID
	TenantId       string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                         // 租户ID
	JvmResourceId  string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId" db:"jvmResourceId"`     // 关联的JVM资源ID

	// 类加载统计
	LoadedClassCount      int   `json:"loadedClassCount" form:"loadedClassCount" query:"loadedClassCount" db:"loadedClassCount"`                     // 当前已加载类数量
	TotalLoadedClassCount int64 `json:"totalLoadedClassCount" form:"totalLoadedClassCount" query:"totalLoadedClassCount" db:"totalLoadedClassCount"` // 总加载类数量
	UnloadedClassCount    int64 `json:"unloadedClassCount" form:"unloadedClassCount" query:"unloadedClassCount" db:"unloadedClassCount"`             // 已卸载类数量

	// 比例指标
	ClassUnloadRatePercent    float64 `json:"classUnloadRatePercent" form:"classUnloadRatePercent" query:"classUnloadRatePercent" db:"classUnloadRatePercent"`             // 类卸载率(百分比)
	ClassRetentionRatePercent float64 `json:"classRetentionRatePercent" form:"classRetentionRatePercent" query:"classRetentionRatePercent" db:"classRetentionRatePercent"` // 类保留率(百分比)

	// 配置状态
	VerboseClassLoading string `json:"verboseClassLoading" form:"verboseClassLoading" query:"verboseClassLoading" db:"verboseClassLoading"` // 是否启用详细类加载输出(Y是,N否)

	// 性能指标
	LoadingRatePerHour float64 `json:"loadingRatePerHour" form:"loadingRatePerHour" query:"loadingRatePerHour" db:"loadingRatePerHour"` // 每小时平均类加载数量
	LoadingEfficiency  float64 `json:"loadingEfficiency" form:"loadingEfficiency" query:"loadingEfficiency" db:"loadingEfficiency"`     // 类加载效率
	MemoryEfficiency   string  `json:"memoryEfficiency" form:"memoryEfficiency" query:"memoryEfficiency" db:"memoryEfficiency"`         // 内存使用效率评估
	LoaderHealth       string  `json:"loaderHealth" form:"loaderHealth" query:"loaderHealth" db:"loaderHealth"`                         // 类加载器健康状况

	// 健康状态
	HealthyFlag           string `json:"healthyFlag" form:"healthyFlag" query:"healthyFlag" db:"healthyFlag"`                                         // 类加载健康标记(Y健康,N异常)
	HealthGrade           string `json:"healthGrade" form:"healthGrade" query:"healthGrade" db:"healthGrade"`                                         // 健康等级(EXCELLENT/GOOD/FAIR/POOR)
	RequiresAttentionFlag string `json:"requiresAttentionFlag" form:"requiresAttentionFlag" query:"requiresAttentionFlag" db:"requiresAttentionFlag"` // 是否需要立即关注(Y是,N否)
	PotentialIssuesJson   string `json:"potentialIssuesJson" form:"potentialIssuesJson" query:"potentialIssuesJson" db:"potentialIssuesJson"`         // 潜在问题列表JSON

	// 时间字段
	CollectionTime time.Time `json:"collectionTime" form:"collectionTime" query:"collectionTime" db:"collectionTime"` // 采集时间

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// ===============================
// 统计和概览
// ===============================

// JvmOverviewRequest JVM概览查询请求
type JvmOverviewRequest struct {
	TenantId        string `json:"tenantId" form:"tenantId" query:"tenantId"`                      // 租户ID
	ApplicationName string `json:"applicationName" form:"applicationName" query:"applicationName"` // 应用名称（可选）
}

// JvmOverviewResponse JVM概览响应 (统计数据，非数据库表)
type JvmOverviewResponse struct {
	TotalInstances     int     `json:"totalInstances"`     // JVM实例总数
	HealthyInstances   int     `json:"healthyInstances"`   // 健康实例数
	UnhealthyInstances int     `json:"unhealthyInstances"` // 异常实例数
	AttentionRequired  int     `json:"attentionRequired"`  // 需要关注实例数
	AvgHeapUsage       float64 `json:"avgHeapUsage"`       // 平均堆内存使用率
	AvgThreadCount     float64 `json:"avgThreadCount"`     // 平均线程数
	DeadlockCount      int     `json:"deadlockCount"`      // 死锁数量
	HighGcFrequency    int     `json:"highGcFrequency"`    // 高GC频率实例数
	ApplicationCount   int     `json:"applicationCount"`   // 应用数量
}

// ===============================
// 应用监控数据查询 (HUB_MONITOR_APP_DATA)
// ===============================

// QueryAppMonitorDataRequest 应用监控数据查询请求
type QueryAppMonitorDataRequest struct {
	PageIndex             int    `json:"pageIndex" form:"pageIndex" query:"pageIndex" binding:"min=1"`                     // 页码
	PageSize              int    `json:"pageSize" form:"pageSize" query:"pageSize" binding:"min=1,max=100"`                // 每页大小
	TenantId              string `json:"tenantId" form:"tenantId" query:"tenantId"`                                        // 租户ID
	JvmResourceId         string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId"`                         // JVM资源ID
	ApplicationName       string `json:"applicationName" form:"applicationName" query:"applicationName"`                   // 应用名称
	DataType              string `json:"dataType" form:"dataType" query:"dataType"`                                        // 数据类型
	DataName              string `json:"dataName" form:"dataName" query:"dataName"`                                        // 数据名称
	DataCategory          string `json:"dataCategory" form:"dataCategory" query:"dataCategory"`                            // 数据分类
	HealthyFlag           string `json:"healthyFlag" form:"healthyFlag" query:"healthyFlag"`                               // 健康标记
	HealthGrade           string `json:"healthGrade" form:"healthGrade" query:"healthGrade"`                               // 健康等级
	RequiresAttentionFlag string `json:"requiresAttentionFlag" form:"requiresAttentionFlag" query:"requiresAttentionFlag"` // 是否需要关注
	StartTime             string `json:"startTime" form:"startTime" query:"startTime"`                                     // 开始时间（字符串格式）
	EndTime               string `json:"endTime" form:"endTime" query:"endTime"`                                           // 结束时间（字符串格式）
	OrderBy               string `json:"orderBy" form:"orderBy" query:"orderBy"`                                           // 排序字段
	OrderDirection        string `json:"orderDirection" form:"orderDirection" query:"orderDirection"`                      // 排序方向
}

// AppMonitorDataResponse 应用监控数据响应 (对应 HUB_MONITOR_APP_DATA 表)
type AppMonitorDataResponse struct {
	// 主键和关联字段
	AppDataId     string `json:"appDataId" form:"appDataId" query:"appDataId" db:"appDataId"`                 // 应用监控数据ID
	TenantId      string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                     // 租户ID
	JvmResourceId string `json:"jvmResourceId" form:"jvmResourceId" query:"jvmResourceId" db:"jvmResourceId"` // 关联的JVM资源ID

	// 数据分类标识
	DataType     string `json:"dataType" form:"dataType" query:"dataType" db:"dataType"`                 // 数据类型
	DataName     string `json:"dataName" form:"dataName" query:"dataName" db:"dataName"`                 // 数据名称
	DataCategory string `json:"dataCategory" form:"dataCategory" query:"dataCategory" db:"dataCategory"` // 数据分类

	// 监控数据
	DataJson string `json:"dataJson" form:"dataJson" query:"dataJson" db:"dataJson"` // 监控数据JSON

	// 核心指标
	PrimaryValue   *float64 `json:"primaryValue" form:"primaryValue" query:"primaryValue" db:"primaryValue"`         // 主要指标值
	SecondaryValue *float64 `json:"secondaryValue" form:"secondaryValue" query:"secondaryValue" db:"secondaryValue"` // 次要指标值
	StatusValue    string   `json:"statusValue" form:"statusValue" query:"statusValue" db:"statusValue"`             // 状态值

	// 健康状态
	HealthyFlag           string `json:"healthyFlag" form:"healthyFlag" query:"healthyFlag" db:"healthyFlag"`                                         // 健康标记
	HealthGrade           string `json:"healthGrade" form:"healthGrade" query:"healthGrade" db:"healthGrade"`                                         // 健康等级
	RequiresAttentionFlag string `json:"requiresAttentionFlag" form:"requiresAttentionFlag" query:"requiresAttentionFlag" db:"requiresAttentionFlag"` // 是否需要关注

	// 标签和维度
	TagsJson string `json:"tagsJson" form:"tagsJson" query:"tagsJson" db:"tagsJson"` // 标签信息JSON

	// 时间字段
	CollectionTime time.Time `json:"collectionTime" form:"collectionTime" query:"collectionTime" db:"collectionTime"` // 数据采集时间

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}
