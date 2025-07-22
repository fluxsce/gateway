package logwrite

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// WriteTaskExecutionLog 静态方法：写入任务执行日志
// 从任务配置和执行结果中提取信息，创建日志记录并写入数据库
func WriteTaskExecutionLog(ctx context.Context, taskConfig interface{}, taskResult interface{}, maxRetries int, tenantId, schedulerId string) error {
	// 从配置中获取默认数据库连接
	defaultDbName := config.GetString("database.default", "default")
	db := database.GetConnection(defaultDbName)
	if db == nil {
		return fmt.Errorf("未找到数据库连接: %s", defaultDbName)
	}

	// 创建日志记录
	log := createExecutionLog(taskConfig, taskResult, maxRetries, tenantId, schedulerId)

	// 设置默认值
	setLogDefaults(log)

	// 插入日志记录
	_, err := db.Insert(ctx, log.TableName(), log, true)
	if err != nil {
		logger.Error("写入执行日志失败", "executionId", log.ExecutionId, "error", err)
		return fmt.Errorf("写入执行日志失败: %w", err)
	}

	return nil
}

// createExecutionLog 创建执行日志记录
func createExecutionLog(taskConfig interface{}, taskResult interface{}, maxRetries int, tenantId, schedulerId string) *TimerExecutionLog {
	now := time.Now()

	// 设置默认租户ID
	if tenantId == "" {
		tenantId = "DEFAULT"
	}

	// 提取任务信息
	taskID := getStringField(taskConfig, "ID")
	taskName := getStringField(taskConfig, "Name")
	params := getField(taskConfig, "Params")

	// 提取执行结果信息
	startTime := getTimeField(taskResult, "StartTime")
	endTime := getTimeField(taskResult, "EndTime")
	duration := getDurationField(taskResult, "Duration")
	error := getStringField(taskResult, "Error")
	retryCount := getIntField(taskResult, "RetryCount")
	result := getField(taskResult, "Result")

	// 如果开始时间为空，使用当前时间
	if startTime.IsZero() {
		startTime = now
	}

	// 计算结束时间和时长
	var endTimePtr *time.Time
	var durationMs *int64

	if !endTime.IsZero() {
		endTimePtr = &endTime
		duration := endTime.Sub(startTime).Milliseconds()
		durationMs = &duration
	} else if duration > 0 {
		calculatedEndTime := startTime.Add(duration)
		endTimePtr = &calculatedEndTime
		durationValue := duration.Milliseconds()
		durationMs = &durationValue
	}

	// 判断执行是否成功
	success := error == ""
	executionStatus := int(StatusCompleted) // 使用基础int类型
	if !success {
		executionStatus = int(StatusFailed)
	}

	resultSuccess := "Y"
	if !success {
		resultSuccess = "N"
	}

	// 序列化参数和结果
	var paramsStr, resultStr *string
	if params != nil {
		if paramsBytes, err := json.Marshal(params); err == nil {
			paramsString := string(paramsBytes)
			paramsStr = &paramsString
		}
	}
	if result != nil {
		if resultBytes, err := json.Marshal(result); err == nil {
			resultString := string(resultBytes)
			resultStr = &resultString
		}
	}

	// 设置错误信息
	var errorMessage *string
	if error != "" {
		errorMessage = &error
	}

	// 设置任务名称和调度器ID
	var taskNamePtr *string
	if taskName != "" {
		taskNamePtr = &taskName
	}

	var schedulerIdPtr *string
	if schedulerId != "" {
		schedulerIdPtr = &schedulerId
	}

	// 设置日志信息（使用基础string类型）
	logLevel := string(LogLevelInfo)
	logMessage := "任务执行完成"
	if !success {
		logLevel = string(LogLevelError)
		logMessage = "任务执行失败"
	}
	logTimestamp := now

	return &TimerExecutionLog{
		ExecutionId:         generateExecutionId(),
		TenantId:            tenantId,
		TaskId:              taskID,
		TaskName:            taskNamePtr,
		SchedulerId:         schedulerIdPtr,
		ExecutionStartTime:  startTime,
		ExecutionEndTime:    endTimePtr,
		ExecutionDurationMs: durationMs,
		ExecutionStatus:     executionStatus,
		ResultSuccess:       resultSuccess,
		ErrorMessage:        errorMessage,
		RetryCount:          retryCount,
		MaxRetryCount:       maxRetries,
		ExecutionParams:     paramsStr,
		ExecutionResult:     resultStr,
		LogLevel:            &logLevel,
		LogMessage:          &logMessage,
		LogTimestamp:        &logTimestamp,
		AddTime:             now,
		EditTime:            now,
		AddWho:              "SYSTEM",
		EditWho:             "SYSTEM",
		OprSeqFlag:          fmt.Sprintf("LOG_%d", now.UnixNano()),
		CurrentVersion:      1,
		ActiveFlag:          "Y",
	}
}

// setLogDefaults 设置日志默认值
func setLogDefaults(log *TimerExecutionLog) {
	now := time.Now()

	if log.AddTime.IsZero() {
		log.AddTime = now
	}
	if log.EditTime.IsZero() {
		log.EditTime = now
	}
	if log.AddWho == "" {
		log.AddWho = "SYSTEM"
	}
	if log.EditWho == "" {
		log.EditWho = "SYSTEM"
	}
	if log.OprSeqFlag == "" {
		log.OprSeqFlag = fmt.Sprintf("LOG_%d", now.UnixNano())
	}
	if log.CurrentVersion == 0 {
		log.CurrentVersion = 1
	}
	if log.ActiveFlag == "" {
		log.ActiveFlag = "Y"
	}
	if log.ResultSuccess == "" {
		log.ResultSuccess = "N"
	}
}

// 辅助函数：通过反射获取字符串字段
func getStringField(obj interface{}, fieldName string) string {
	if obj == nil {
		return ""
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName(fieldName)
		if field.IsValid() && field.Kind() == reflect.String {
			return field.String()
		}
	}
	return ""
}

// 辅助函数：通过反射获取时间字段
func getTimeField(obj interface{}, fieldName string) time.Time {
	if obj == nil {
		return time.Time{}
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName(fieldName)
		if field.IsValid() {
			if t, ok := field.Interface().(time.Time); ok {
				return t
			}
		}
	}
	return time.Time{}
}

// 辅助函数：通过反射获取时长字段
func getDurationField(obj interface{}, fieldName string) time.Duration {
	if obj == nil {
		return 0
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName(fieldName)
		if field.IsValid() {
			if d, ok := field.Interface().(time.Duration); ok {
				return d
			}
		}
	}
	return 0
}

// 辅助函数：通过反射获取整数字段
func getIntField(obj interface{}, fieldName string) int {
	if obj == nil {
		return 0
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName(fieldName)
		if field.IsValid() && field.Kind() == reflect.Int {
			return int(field.Int())
		}
	}
	return 0
}

// 辅助函数：通过反射获取任意字段
func getField(obj interface{}, fieldName string) interface{} {
	if obj == nil {
		return nil
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName(fieldName)
		if field.IsValid() {
			return field.Interface()
		}
	}
	return nil
}

// generateExecutionId 生成执行ID
// 使用高并发安全的随机数生成器确保唯一性
func generateExecutionId() string {
	// 使用32位唯一字符串生成器，确保高并发下的唯一性
	// 格式：EXEC_前缀 + 32位唯一字符串
	return random.Generate32BitRandomString()
}
