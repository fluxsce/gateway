package dbloader

import (
	"context"
	"encoding/json"
	"fmt"

	"gateway/internal/gateway/handler/circuitbreaker"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
)

// LoadServiceCircuitBreakerConfig 按服务 ID 加载活动熔断配置。
// 无行时返回 nil, nil，由运行时回退 DefaultCircuitBreakerConfig。
func (loader *LimiterServiceLoader) LoadServiceCircuitBreakerConfig(ctx context.Context, serviceId string) (*circuitbreaker.CircuitBreakerConfig, error) {
	if serviceId == "" {
		return nil, nil
	}

	baseQuery := `
		SELECT tenantId, circuitBreakerConfigId, routeConfigId, targetServiceId, breakerName,
		       errorRatePercent, minimumRequests, halfOpenMaxRequests,
		       slowCallThreshold, slowCallRatePercent, openTimeoutSeconds, windowSizeSeconds,
		       errorStatusCode, errorMessage, storageType, storageConfig, configPriority
		FROM HUB_GW_CIRCUIT_BREAKER_CONFIG
		WHERE tenantId = ? AND targetServiceId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
	`

	pagination := sqlutils.NewPaginationInfo(1, 1)
	dbType := sqlutils.GetDatabaseType(loader.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, fmt.Errorf("构建熔断配置分页查询失败: %w", err)
	}

	allArgs := append([]interface{}{loader.tenantId, serviceId}, paginationArgs...)
	var records []CircuitBreakerConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询服务熔断配置失败: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	return mapCircuitBreakerRecord(records[0]), nil
}

// mapCircuitBreakerRecord 把表记录映射为运行时熔断配置。
func mapCircuitBreakerRecord(record CircuitBreakerConfigRecord) *circuitbreaker.CircuitBreakerConfig {
	cfg := &circuitbreaker.CircuitBreakerConfig{
		Enabled:             true,
		ErrorRatePercent:    record.ErrorRatePercent,
		MinimumRequests:     record.MinimumRequests,
		HalfOpenMaxRequests: record.HalfOpenMaxRequests,
		SlowCallThreshold:   record.SlowCallThreshold,
		SlowCallRatePercent: record.SlowCallRatePercent,
		OpenTimeoutSeconds:  record.OpenTimeoutSeconds,
		WindowSizeSeconds:   record.WindowSizeSeconds,
		ErrorStatusCode:     record.ErrorStatusCode,
		StorageType:         record.StorageType,
		StorageConfig:       parseCircuitBreakerStorageConfig(record.StorageConfig),
	}
	if record.ErrorMessage != nil {
		cfg.ErrorMessage = *record.ErrorMessage
	}
	return cfg
}

// parseCircuitBreakerStorageConfig 解析 storageConfig JSON 为 cache 实例参数。
func parseCircuitBreakerStorageConfig(raw *string) map[string]string {
	result := make(map[string]string)
	if raw == nil || *raw == "" {
		return result
	}
	var typed map[string]string
	if err := json.Unmarshal([]byte(*raw), &typed); err == nil {
		return typed
	}
	var generic map[string]interface{}
	if err := json.Unmarshal([]byte(*raw), &generic); err != nil {
		return result
	}
	for key, value := range generic {
		result[key] = fmt.Sprint(value)
	}
	return result
}
