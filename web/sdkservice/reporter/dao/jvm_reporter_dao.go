package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
	"gateway/web/sdkservice/reporter/models"
	"strings"
	"time"
)

// JvmReporterDao JVM监控数据上报DAO
//
// 时间处理说明：
// 1. 所有时间值都作为参数传递给SQL，不使用数据库函数（如NOW()、SYSDATE等）
// 2. 这样可以确保跨数据库兼容性（MySQL、Oracle、SQLite等）
// 3. 时间戳转换使用time.UnixMilli()将毫秒时间戳转为time.Time类型
// 4. 当前时间使用time.Now()获取，然后作为参数传递
// 5. 需要时间格式化时可使用 gateway/pkg/utils/ctime 包提供的工具函数
type JvmReporterDao struct {
	db database.Database
}

// NewJvmReporterDao 创建JVM Reporter DAO实例
func NewJvmReporterDao(db database.Database) *JvmReporterDao {
	return &JvmReporterDao{
		db: db,
	}
}

// SaveJvmMonitoringData 保存JVM监控数据
// jvmResourceId由应用端传入，用于标识唯一的JVM实例，便于快速检索和更新
func (dao *JvmReporterDao) SaveJvmMonitoringData(tenantId, serviceGroupId, groupName string, req *models.JvmReportRequest) error {
	ctx := context.Background()

	jvmInfo := &req.JvmResourceInfo

	// 1. 使用应用端传入的jvmResourceId（不再生成）
	jvmResourceId := req.JvmResourceId
	if jvmResourceId == "" {
		return fmt.Errorf("jvmResourceId不能为空")
	}

	// 2. 保存JVM资源主表记录
	if err := dao.insertJvmResource(ctx, tenantId, serviceGroupId, groupName, jvmResourceId, req); err != nil {
		return fmt.Errorf("保存JVM资源记录失败: %w", err)
	}

	// 3. 保存堆内存记录
	if jvmInfo.HeapMemory != nil {
		if err := dao.insertMemory(ctx, tenantId, jvmResourceId, "HEAP", jvmInfo.HeapMemory); err != nil {
			return fmt.Errorf("保存堆内存记录失败: %w", err)
		}
	}

	// 4. 保存非堆内存记录
	if jvmInfo.NonHeapMemory != nil {
		if err := dao.insertMemory(ctx, tenantId, jvmResourceId, "NON_HEAP", jvmInfo.NonHeapMemory); err != nil {
			return fmt.Errorf("保存非堆内存记录失败: %w", err)
		}
	}

	// 5. 保存内存池列表
	if jvmInfo.MemoryPools != nil && len(jvmInfo.MemoryPools) > 0 {
		if err := dao.insertMemoryPools(ctx, tenantId, jvmResourceId, jvmInfo.MemoryPools); err != nil {
			return fmt.Errorf("保存内存池记录失败: %w", err)
		}
	}

	// 6. 保存GC快照（单条汇总记录）
	if jvmInfo.GarbageCollector != nil {
		if err := dao.insertGCSnapshot(ctx, tenantId, jvmResourceId, jvmInfo.GarbageCollector); err != nil {
			return fmt.Errorf("保存GC快照记录失败: %w", err)
		}
	}

	// 7. 保存线程信息
	if jvmInfo.ThreadInfo != nil {
		jvmThreadId := random.Generate32BitRandomString()
		if err := dao.insertThread(ctx, tenantId, jvmResourceId, jvmThreadId, jvmInfo.ThreadInfo); err != nil {
			return fmt.Errorf("保存线程信息失败: %w", err)
		}

		// 7.1 保存线程状态统计
		if jvmInfo.ThreadInfo.ThreadStateStats != nil {
			if err := dao.insertThreadState(ctx, tenantId, jvmResourceId, jvmThreadId, jvmInfo.ThreadInfo.ThreadStateStats); err != nil {
				return fmt.Errorf("保存线程状态统计失败: %w", err)
			}
		}

		// 7.2 保存死锁信息
		if jvmInfo.ThreadInfo.DeadlockInfo != nil {
			if err := dao.insertDeadlock(ctx, tenantId, jvmResourceId, jvmThreadId, jvmInfo.ThreadInfo.DeadlockInfo); err != nil {
				return fmt.Errorf("保存死锁信息失败: %w", err)
			}
		}
	}

	// 8. 保存类加载信息
	if jvmInfo.ClassLoadingInfo != nil {
		if err := dao.insertClassLoading(ctx, tenantId, jvmResourceId, jvmInfo.ClassLoadingInfo); err != nil {
			return fmt.Errorf("保存类加载信息失败: %w", err)
		}
	}

	logger.Info("JVM监控数据保存成功",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"groupName", groupName,
		"jvmResourceId", jvmResourceId,
		"applicationName", req.ApplicationName,
		"hostIpAddress", req.HostIpAddress)

	return nil
}

// insertJvmResource 插入或更新JVM资源主表记录
// 先尝试UPDATE，如果记录不存在（影响行数=0）则INSERT
// 这种方式兼容MySQL、Oracle、SQLite等多种数据库
func (dao *JvmReporterDao) insertJvmResource(ctx context.Context, tenantId, serviceGroupId, groupName, jvmResourceId string, req *models.JvmReportRequest) error {
	jvmInfo := &req.JvmResourceInfo

	// 转换系统属性为JSON
	systemPropertiesJson := ""
	if jvmInfo.SystemProperties != nil {
		if jsonBytes, err := json.Marshal(jvmInfo.SystemProperties); err == nil {
			systemPropertiesJson = string(jsonBytes)
		}
	}

	// 获取当前时间，避免在SQL中使用数据库函数
	now := time.Now()

	// 转换时间戳为time.Time类型，用于数据库参数
	collectionTime := jvmInfo.ToCollectionTime()
	startTime := jvmInfo.ToStartTime()

	// 1. 先尝试UPDATE现有记录
	updateSql := `
		UPDATE HUB_MONITOR_JVM_RESOURCE SET
			groupName = ?,
			applicationName = ?,
			hostName = ?,
			hostIpAddress = ?,
			collectionTime = ?,
			jvmStartTime = ?,
			jvmUptimeMs = ?,
			healthyFlag = ?,
			healthGrade = ?,
			requiresAttentionFlag = ?,
			summaryText = ?,
			systemPropertiesJson = ?,
			editTime = ?,
			currentVersion = currentVersion + 1
		WHERE tenantId = ? AND serviceGroupId = ? AND jvmResourceId = ?
	`

	rowsAffected, err := dao.db.Exec(ctx, updateSql, []interface{}{
		groupName,
		req.ApplicationName,
		req.HostName,
		req.HostIpAddress,
		collectionTime,
		startTime,
		jvmInfo.Uptime,
		boolToFlag(jvmInfo.Healthy),
		jvmInfo.HealthGrade,
		boolToFlag(jvmInfo.RequiresAttention),
		jvmInfo.Summary,
		systemPropertiesJson,
		now,
		tenantId,
		serviceGroupId,
		jvmResourceId,
	}, true)

	if err != nil {
		logger.Error("更新JVM资源记录失败",
			"error", err,
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"jvmResourceId", jvmResourceId)
		return err
	}

	// 2. 如果UPDATE没有影响任何行（记录不存在），则INSERT
	if rowsAffected == 0 {
		insertSql := `
			INSERT INTO HUB_MONITOR_JVM_RESOURCE (
				jvmResourceId, tenantId, serviceGroupId, applicationName, groupName, 
				hostName, hostIpAddress, collectionTime, jvmStartTime, jvmUptimeMs, 
				healthyFlag, healthGrade, requiresAttentionFlag, summaryText, 
				systemPropertiesJson, addTime, editTime, currentVersion, activeFlag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err = dao.db.Exec(ctx, insertSql, []interface{}{
			jvmResourceId,
			tenantId,
			serviceGroupId,
			req.ApplicationName,
			groupName,
			req.HostName,
			req.HostIpAddress,
			collectionTime,
			startTime,
			jvmInfo.Uptime,
			boolToFlag(jvmInfo.Healthy),
			jvmInfo.HealthGrade,
			boolToFlag(jvmInfo.RequiresAttention),
			jvmInfo.Summary,
			systemPropertiesJson,
			now,
			now,
			1,
			"Y",
		}, true)

		if err != nil {
			logger.Error("插入JVM资源记录失败",
				"error", err,
				"tenantId", tenantId,
				"serviceGroupId", serviceGroupId,
				"jvmResourceId", jvmResourceId)
			return err
		}

		logger.Debug("首次上报，创建JVM资源记录",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"groupName", groupName,
			"jvmResourceId", jvmResourceId)
	} else {
		logger.Debug("更新JVM资源记录",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"groupName", groupName,
			"jvmResourceId", jvmResourceId,
			"rowsAffected", rowsAffected)
	}

	return nil
}

// insertMemory 插入内存记录
func (dao *JvmReporterDao) insertMemory(ctx context.Context, tenantId, jvmResourceId, memoryType string, memory *models.MemoryInfo) error {
	sql := `
		INSERT INTO HUB_MONITOR_JVM_MEMORY (
			jvmMemoryId, tenantId, jvmResourceId, memoryType,
			initMemoryBytes, usedMemoryBytes, committedMemoryBytes, maxMemoryBytes,
			usagePercent, healthyFlag, collectionTime,
			addTime, editTime, currentVersion, activeFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	collectionTime := time.UnixMilli(memory.CollectionTime)

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		random.Generate32BitRandomString(),
		tenantId,
		jvmResourceId,
		memoryType,
		memory.Init,
		memory.Used,
		memory.Committed,
		memory.Max,
		memory.UsagePercent * 100, // 转换为百分比
		boolToFlag(memory.IsHealthy()),
		collectionTime,
		now,
		now,
		1,
		"Y",
	}, true)

	return err
}

// insertMemoryPools 插入内存池列表
func (dao *JvmReporterDao) insertMemoryPools(ctx context.Context, tenantId, jvmResourceId string, pools []models.MemoryPoolInfo) error {
	for _, pool := range pools {
		if err := dao.insertMemoryPool(ctx, tenantId, jvmResourceId, &pool); err != nil {
			return err
		}
	}
	return nil
}

// insertMemoryPool 插入单个内存池记录
func (dao *JvmReporterDao) insertMemoryPool(ctx context.Context, tenantId, jvmResourceId string, pool *models.MemoryPoolInfo) error {
	sql := `
		INSERT INTO HUB_MONITOR_JVM_MEM_POOL (
			memoryPoolId, tenantId, jvmResourceId, poolName, poolType, poolCategory,
			currentInitBytes, currentUsedBytes, currentCommittedBytes, currentMaxBytes, currentUsagePercent,
			peakInitBytes, peakUsedBytes, peakCommittedBytes, peakMaxBytes, peakUsagePercent,
			usageThresholdSupported, usageThresholdBytes, usageThresholdCount, collectionUsageSupported,
			healthyFlag, collectionTime,
			addTime, editTime, currentVersion, activeFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	currentUsagePercent := 0.0
	if pool.Usage != nil {
		currentUsagePercent = pool.Usage.UsagePercent * 100
	}

	peakUsagePercent := 0.0
	if pool.PeakUsage != nil {
		peakUsagePercent = pool.PeakUsage.UsagePercent * 100
	}

	now := time.Now()
	collectionTime := time.UnixMilli(pool.CollectionTime)

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		random.Generate32BitRandomString(),
		tenantId,
		jvmResourceId,
		pool.Name,
		pool.Type,
		pool.GetCategoryLabel(),
		getMemoryValue(pool.Usage, "Init"),
		getMemoryValue(pool.Usage, "Used"),
		getMemoryValue(pool.Usage, "Committed"),
		getMemoryValue(pool.Usage, "Max"),
		currentUsagePercent,
		getMemoryValue(pool.PeakUsage, "Init"),
		getMemoryValue(pool.PeakUsage, "Used"),
		getMemoryValue(pool.PeakUsage, "Committed"),
		getMemoryValue(pool.PeakUsage, "Max"),
		peakUsagePercent,
		boolToFlag(pool.UsageThresholdSupported),
		pool.UsageThreshold,
		pool.UsageThresholdCount,
		boolToFlag(pool.CollectionUsageThresholdSupported),
		"Y", // 默认健康
		collectionTime,
		now,
		now,
		1,
		"Y",
	}, true)

	return err
}

// insertGCSnapshot 插入GC快照记录（每次采集插入一条汇总记录）
func (dao *JvmReporterDao) insertGCSnapshot(ctx context.Context, tenantId, jvmResourceId string, gc *models.GarbageCollector) error {
	sql := `
		INSERT INTO HUB_MONITOR_JVM_GC (
			gcSnapshotId, tenantId, jvmResourceId, 
			collectionCount, collectionTimeMs,
			s0c, s1c, s0u, s1u,
			ec, eu,
			oc, ou,
			mc, mu,
			ccsc, ccsu,
			ygc, ygct, fgc, fgct, gct,
			collectionTime,
			addTime, editTime, currentVersion, activeFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	collectionTime := time.UnixMilli(gc.CollectionTime)

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		random.Generate32BitRandomString(),
		tenantId,
		jvmResourceId,
		gc.CollectionCount,
		gc.CollectionTimeMs,
		gc.S0C, gc.S1C, gc.S0U, gc.S1U,
		gc.EC, gc.EU,
		gc.OC, gc.OU,
		gc.MC, gc.MU,
		gc.CCSC, gc.CCSU,
		gc.YGC, gc.YGCT, gc.FGC, gc.FGCT, gc.GCT,
		collectionTime,
		now,
		now,
		1,
		"Y",
	}, true)

	return err
}

// insertThread 插入线程信息
func (dao *JvmReporterDao) insertThread(ctx context.Context, tenantId, jvmResourceId, jvmThreadId string, thread *models.ThreadInfo) error {
	// 转换潜在问题列表为JSON
	potentialIssuesJson := ""
	if thread.PotentialIssues != nil {
		if jsonBytes, err := json.Marshal(thread.PotentialIssues); err == nil {
			potentialIssuesJson = string(jsonBytes)
		}
	}

	sql := `
		INSERT INTO HUB_MONITOR_JVM_THREAD (
			jvmThreadId, tenantId, jvmResourceId, currentThreadCount, daemonThreadCount, userThreadCount,
			peakThreadCount, totalStartedThreadCount, threadGrowthRatePercent, daemonThreadRatioPercent,
			cpuTimeSupported, cpuTimeEnabled, memoryAllocSupported, memoryAllocEnabled,
			contentionSupported, contentionEnabled, healthyFlag, healthGrade, requiresAttentionFlag,
			potentialIssuesJson, collectionTime,
			addTime, editTime, currentVersion, activeFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	collectionTime := time.UnixMilli(thread.CollectionTime)

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		jvmThreadId,
		tenantId,
		jvmResourceId,
		thread.ThreadCount,
		thread.DaemonThreadCount,
		thread.UserThreadCount,
		thread.PeakThreadCount,
		thread.TotalStartedThreadCount,
		thread.ThreadGrowthRate * 100,
		thread.DaemonThreadRatio * 100,
		boolToFlag(thread.ThreadCpuTimeSupported),
		boolToFlag(thread.ThreadCpuTimeEnabled),
		boolToFlag(thread.ThreadMemoryAllocSupported),
		boolToFlag(thread.ThreadMemoryAllocEnabled),
		boolToFlag(thread.ThreadContentionSupported),
		boolToFlag(thread.ThreadContentionEnabled),
		boolToFlag(thread.Healthy),
		thread.HealthGrade,
		"N", // 默认不需要关注
		potentialIssuesJson,
		collectionTime,
		now,
		now,
		1,
		"Y",
	}, true)

	return err
}

// insertThreadState 插入线程状态统计
func (dao *JvmReporterDao) insertThreadState(ctx context.Context, tenantId, jvmResourceId, jvmThreadId string, stats *models.ThreadStateStats) error {
	// 计算比例
	activeRatio := 0.0
	blockedRatio := 0.0
	waitingRatio := 0.0

	if stats.TotalThreadCount > 0 {
		activeRatio = float64(stats.RunnableThreadCount) / float64(stats.TotalThreadCount) * 100
		blockedRatio = float64(stats.BlockedThreadCount) / float64(stats.TotalThreadCount) * 100
		waitingRatio = float64(stats.WaitingThreadCount+stats.TimedWaitingThreadCount) / float64(stats.TotalThreadCount) * 100
	}

	// 判断健康等级
	healthGrade := "EXCELLENT"
	healthyFlag := "Y"
	if blockedRatio > 20 || waitingRatio > 80 {
		healthGrade = "POOR"
		healthyFlag = "N"
	} else if blockedRatio > 15 || waitingRatio > 70 {
		healthGrade = "FAIR"
	} else if blockedRatio > 10 || waitingRatio > 60 {
		healthGrade = "GOOD"
	}

	sql := `
		INSERT INTO HUB_MONITOR_JVM_THR_STATE (
			threadStateId, tenantId, jvmThreadId, jvmResourceId,
			newThreadCount, runnableThreadCount, blockedThreadCount, waitingThreadCount,
			timedWaitingThreadCount, terminatedThreadCount, totalThreadCount,
			activeThreadRatioPercent, blockedThreadRatioPercent, waitingThreadRatioPercent,
			healthyFlag, healthGrade, collectionTime,
			addTime, editTime, currentVersion, activeFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	collectionTime := time.UnixMilli(stats.CollectionTime)

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		random.Generate32BitRandomString(),
		tenantId,
		jvmThreadId,
		jvmResourceId,
		stats.NewThreadCount,
		stats.RunnableThreadCount,
		stats.BlockedThreadCount,
		stats.WaitingThreadCount,
		stats.TimedWaitingThreadCount,
		stats.TerminatedThreadCount,
		stats.TotalThreadCount,
		activeRatio,
		blockedRatio,
		waitingRatio,
		healthyFlag,
		healthGrade,
		collectionTime,
		now,
		now,
		1,
		"Y",
	}, true)

	return err
}

// insertDeadlock 插入死锁信息
func (dao *JvmReporterDao) insertDeadlock(ctx context.Context, tenantId, jvmResourceId, jvmThreadId string, deadlock *models.DeadlockInfo) error {
	// 转换线程ID列表为字符串
	threadIdsStr := ""
	if deadlock.DeadlockThreadIds != nil {
		ids := make([]string, len(deadlock.DeadlockThreadIds))
		for i, id := range deadlock.DeadlockThreadIds {
			ids[i] = fmt.Sprintf("%d", id)
		}
		threadIdsStr = strings.Join(ids, ",")
	}

	// 转换线程名称列表为字符串
	threadNamesStr := ""
	if deadlock.DeadlockThreadNames != nil {
		threadNamesStr = strings.Join(deadlock.DeadlockThreadNames, ",")
	}

	// 确定告警级别
	alertLevel := "INFO"
	requiresAction := "N"
	if deadlock.HasDeadlock {
		switch deadlock.Severity {
		case "LOW":
			alertLevel = "WARNING"
		case "MEDIUM":
			alertLevel = "ERROR"
		case "HIGH":
			alertLevel = "CRITICAL"
			requiresAction = "Y"
		case "CRITICAL":
			alertLevel = "EMERGENCY"
			requiresAction = "Y"
		}
	}

	// 处理可空的死锁检测时间（如果为0则使用nil，避免MySQL的0000-00-00错误）
	var detectionTime interface{}
	if deadlock.DetectionTime > 0 {
		ts := time.UnixMilli(deadlock.DetectionTime)
		detectionTime = ts
	} else {
		detectionTime = nil
	}

	now := time.Now()
	collectionTime := time.UnixMilli(deadlock.CollectionTime)

	sql := `
		INSERT INTO HUB_MONITOR_JVM_DEADLOCK (
			deadlockId, tenantId, jvmThreadId, jvmResourceId, hasDeadlockFlag,
			deadlockThreadCount, deadlockThreadIds, deadlockThreadNames,
			severityLevel, severityDescription, affectedThreadGroups,
			detectionTime, deadlockDurationMs, collectionTime,
			descriptionText, recommendedAction, alertLevel, requiresActionFlag,
			addTime, editTime, currentVersion, activeFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		random.Generate32BitRandomString(),
		tenantId,
		jvmThreadId,
		jvmResourceId,
		boolToFlag(deadlock.HasDeadlock),
		deadlock.DeadlockThreadCount,
		threadIdsStr,
		threadNamesStr,
		deadlock.Severity,
		getSeverityDescription(deadlock.Severity),
		deadlock.AffectedThreadGroups,
		detectionTime,
		deadlock.DeadlockDuration,
		collectionTime,
		deadlock.Description,
		deadlock.RecommendedAction,
		alertLevel,
		requiresAction,
		now,
		now,
		1,
		"Y",
	}, true)

	return err
}

// insertClassLoading 插入类加载信息
func (dao *JvmReporterDao) insertClassLoading(ctx context.Context, tenantId, jvmResourceId string, classLoading *models.ClassLoadingInfo) error {
	// 转换潜在问题列表为JSON
	potentialIssuesJson := ""
	if classLoading.PotentialIssues != nil {
		if jsonBytes, err := json.Marshal(classLoading.PotentialIssues); err == nil {
			potentialIssuesJson = string(jsonBytes)
		}
	}

	// 判断是否需要关注
	requiresAttention := "N"
	if classLoading.HealthGrade == "POOR" || classLoading.LoadedClassCount > 80000 {
		requiresAttention = "Y"
	}

	loadingRatePerHour := 0.0
	loadingEfficiency := 0.0
	memoryEfficiency := ""
	loaderHealth := ""

	if classLoading.Performance != nil {
		loadingRatePerHour = classLoading.Performance.LoadingRatePerHour
		loadingEfficiency = classLoading.Performance.LoadingEfficiency
		memoryEfficiency = classLoading.Performance.MemoryEfficiency
		loaderHealth = classLoading.Performance.LoaderHealth
	}

	now := time.Now()
	collectionTime := time.UnixMilli(classLoading.CollectionTime)

	sql := `
		INSERT INTO HUB_MONITOR_JVM_CLASS (
			classLoadingId, tenantId, jvmResourceId, loadedClassCount, totalLoadedClassCount,
			unloadedClassCount, classUnloadRatePercent, classRetentionRatePercent,
			verboseClassLoading, loadingRatePerHour, loadingEfficiency, memoryEfficiency,
			loaderHealth, healthyFlag, healthGrade, requiresAttentionFlag,
			potentialIssuesJson, collectionTime,
			addTime, editTime, currentVersion, activeFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		random.Generate32BitRandomString(),
		tenantId,
		jvmResourceId,
		classLoading.LoadedClassCount,
		classLoading.TotalLoadedClassCount,
		classLoading.UnloadedClassCount,
		classLoading.ClassUnloadRate * 100,
		classLoading.ClassRetentionRate * 100,
		boolToFlag(classLoading.VerboseClassLoading),
		loadingRatePerHour,
		loadingEfficiency,
		memoryEfficiency,
		loaderHealth,
		boolToFlag(classLoading.Healthy),
		classLoading.HealthGrade,
		requiresAttention,
		potentialIssuesJson,
		collectionTime,
		now,
		now,
		1,
		"Y",
	}, true)

	return err
}

// 辅助函数

func boolToFlag(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}

func getMemoryValue(memory *models.MemoryInfo, field string) int64 {
	if memory == nil {
		return 0
	}

	switch field {
	case "Init":
		return memory.Init
	case "Used":
		return memory.Used
	case "Committed":
		return memory.Committed
	case "Max":
		return memory.Max
	}

	return 0
}

func getSeverityDescription(severity string) string {
	descriptions := map[string]string{
		"LOW":      "影响少量线程，对系统影响较小",
		"MEDIUM":   "影响中等数量线程，可能导致部分功能受影响",
		"HIGH":     "影响大量线程，系统性能严重下降",
		"CRITICAL": "系统可能完全阻塞，需要立即处理",
	}

	if desc, ok := descriptions[severity]; ok {
		return desc
	}
	return ""
}

// ==========================================
// 应用监控数据保存方法
// ==========================================

// SaveAppMonitoringData 保存应用监控数据
// 使用与JVM监控相同的主表（HUB_MONITOR_JVM_RESOURCE），应用数据存储到 HUB_MONITOR_APP_DATA 表
func (dao *JvmReporterDao) SaveAppMonitoringData(tenantId, serviceGroupId, groupName string, req *models.AppReportRequest) error {
	ctx := context.Background()

	// 1. 使用应用端传入的jvmResourceId（不再生成）
	jvmResourceId := req.JvmResourceId
	if jvmResourceId == "" {
		return fmt.Errorf("jvmResourceId不能为空")
	}

	// 2. 复用现有的JVM资源管理方法，创建一个最小化的JVM报告请求
	// 这样可以确保主表记录存在，并且复用现有的插入/更新逻辑
	minimalJvmReq := &models.JvmReportRequest{
		JvmResourceId:   req.JvmResourceId,
		ApplicationName: req.ApplicationName,
		HostName:        req.HostName,
		HostIpAddress:   req.HostIpAddress,
		JvmResourceInfo: models.JvmResourceInfo{
			CollectionTime:    req.ApplicationResourceInfo.CollectionTime,
			StartTime:         time.Now().UnixMilli(), // 使用当前时间作为默认值
			Uptime:            0,
			Healthy:           req.ApplicationResourceInfo.Healthy,
			HealthGrade:       "UNKNOWN",
			RequiresAttention: false,
			Summary:           "应用监控数据上报创建的基础记录",
			SystemProperties:  make(map[string]string),
		},
	}

	// 复用现有的insertJvmResource方法确保主表记录存在
	if err := dao.insertJvmResource(ctx, tenantId, serviceGroupId, groupName, jvmResourceId, minimalJvmReq); err != nil {
		return fmt.Errorf("确保JVM资源记录存在失败: %w", err)
	}

	// 3. 批量保存应用监控数据
	if req.ApplicationResourceInfo.ThirdPartyMonitorData != nil && len(req.ApplicationResourceInfo.ThirdPartyMonitorData) > 0 {
		if err := dao.batchInsertThirdPartyMonitorData(ctx, tenantId, jvmResourceId, req.ApplicationResourceInfo.ThirdPartyMonitorData); err != nil {
			return fmt.Errorf("保存应用监控数据失败: %w", err)
		}
	}

	logger.Info("应用监控数据保存成功",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"groupName", groupName,
		"jvmResourceId", jvmResourceId,
		"applicationName", req.ApplicationName,
		"hostIpAddress", req.HostIpAddress,
		"dataCount", len(req.ApplicationResourceInfo.ThirdPartyMonitorData))

	return nil
}

// batchInsertThirdPartyMonitorData 批量插入第三方监控数据
// 使用批量插入提高性能，特别是在数据量较大时
func (dao *JvmReporterDao) batchInsertThirdPartyMonitorData(ctx context.Context, tenantId, jvmResourceId string, dataList []models.ThirdPartyMonitorData) error {
	if len(dataList) == 0 {
		return nil
	}

	// 构建批量插入SQL
	baseSql := `
		INSERT INTO HUB_MONITOR_APP_DATA (
			appDataId, tenantId, jvmResourceId, dataType, dataName, dataCategory,
			dataJson, primaryValue, secondaryValue, statusValue,
			healthyFlag, healthGrade, requiresAttentionFlag, tagsJson, collectionTime,
			addTime, editTime, currentVersion, activeFlag
		) VALUES `

	// 准备批量插入的值和参数
	var valuePlaceholders []string
	var args []interface{}
	now := time.Now()

	for _, data := range dataList {
		// 获取监控数据JSON字符串
		dataJsonStr := data.GetDataJsonAsString()

		// 提取核心指标（如果后端没有提供，则尝试自动提取）
		primaryValue := data.ExtractPrimaryValue()
		secondaryValue := data.ExtractSecondaryValue()
		statusValue := data.ExtractStatusValue()

		// 添加占位符
		valuePlaceholders = append(valuePlaceholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		// 添加参数
		args = append(args,
			random.Generate32BitRandomString(), // appDataId
			tenantId,                           // tenantId
			jvmResourceId,                      // jvmResourceId
			data.DataType,                      // dataType
			data.DataName,                      // dataName
			data.DataCategory,                  // dataCategory
			dataJsonStr,                        // dataJson
			primaryValue,                       // primaryValue
			secondaryValue,                     // secondaryValue
			statusValue,                        // statusValue
			data.HealthyFlag,                   // healthyFlag
			data.HealthGrade,                   // healthGrade
			data.RequiresAttentionFlag,         // requiresAttentionFlag
			data.GetTagsJsonAsString(),         // tagsJson
			data.ToCollectionTime(),            // collectionTime
			now,                                // addTime
			now,                                // editTime
			1,                                  // currentVersion
			"Y",                                // activeFlag
		)
	}

	// 构建完整的SQL语句
	fullSql := baseSql + strings.Join(valuePlaceholders, ", ")

	// 执行批量插入
	_, err := dao.db.Exec(ctx, fullSql, args, true)
	if err != nil {
		logger.Error("批量插入第三方监控数据失败",
			"error", err,
			"tenantId", tenantId,
			"jvmResourceId", jvmResourceId,
			"dataCount", len(dataList))
		return err
	}

	logger.Debug("批量插入第三方监控数据成功",
		"tenantId", tenantId,
		"jvmResourceId", jvmResourceId,
		"dataCount", len(dataList))

	return nil
}
