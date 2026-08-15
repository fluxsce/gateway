package dao

import (
	"context"
	"errors"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0022/models"
)

// CircuitBreakerConfigDAO 熔断配置数据访问对象。
type CircuitBreakerConfigDAO struct {
	db database.Database
}

// NewCircuitBreakerConfigDAO 创建熔断配置 DAO。
func NewCircuitBreakerConfigDAO(db database.Database) *CircuitBreakerConfigDAO {
	return &CircuitBreakerConfigDAO{db: db}
}

// GetByTargetServiceId 按服务 ID 取一条熔断配置，含已停用行以便重新启用。
func (dao *CircuitBreakerConfigDAO) GetByTargetServiceId(ctx context.Context, tenantId, serviceId string) (*models.CircuitBreakerConfig, error) {
	if serviceId == "" {
		return nil, errors.New("targetServiceId不能为空")
	}
	query := `
		SELECT * FROM HUB_GW_CIRCUIT_BREAKER_CONFIG
		WHERE tenantId = ? AND targetServiceId = ?
		ORDER BY activeFlag DESC, configPriority ASC
	`
	var config models.CircuitBreakerConfig
	err := dao.db.QueryOne(ctx, &config, query, []interface{}{tenantId, serviceId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询服务熔断配置失败")
	}
	config.KeyStrategy = circuitBreakerKeyStrategyNode
	return &config, nil
}

// UpsertByTargetServiceId 按服务 ID 新增或更新熔断配置，并置为活动。
func (dao *CircuitBreakerConfigDAO) UpsertByTargetServiceId(ctx context.Context, config *models.CircuitBreakerConfig, operatorId string) error {
	if config == nil {
		return errors.New("熔断配置不能为空")
	}
	if config.TargetServiceId == "" {
		return errors.New("targetServiceId不能为空")
	}

	existing, err := dao.GetByTargetServiceId(ctx, config.TenantId, config.TargetServiceId)
	if err != nil {
		return err
	}

	now := time.Now()
	applyCircuitBreakerDefaults(config)
	config.ActiveFlag = "Y"
	config.EditTime = now
	config.EditWho = operatorId
	config.OprSeqFlag = random.Generate32BitRandomString()

	if existing == nil {
		if config.CircuitBreakerConfigId == "" {
			config.CircuitBreakerConfigId = random.GenerateUniqueStringWithPrefix("CB", 32)
		}
		config.AddTime = now
		config.AddWho = operatorId
		config.CurrentVersion = 1
		_, err = dao.db.Insert(ctx, "HUB_GW_CIRCUIT_BREAKER_CONFIG", config, true)
		if err != nil {
			return huberrors.WrapError(err, "添加熔断配置失败")
		}
		return nil
	}

	config.CircuitBreakerConfigId = existing.CircuitBreakerConfigId
	config.AddTime = existing.AddTime
	config.AddWho = existing.AddWho
	config.CurrentVersion = existing.CurrentVersion + 1
	sql := `
		UPDATE HUB_GW_CIRCUIT_BREAKER_CONFIG SET
			breakerName = ?, keyStrategy = ?, errorRatePercent = ?, minimumRequests = ?,
			halfOpenMaxRequests = ?, slowCallThreshold = ?, slowCallRatePercent = ?,
			openTimeoutSeconds = ?, windowSizeSeconds = ?, errorStatusCode = ?, errorMessage = ?,
			storageType = ?, storageConfig = ?, configPriority = ?, noteText = ?,
			editTime = ?, editWho = ?, currentVersion = ?, oprSeqFlag = ?, activeFlag = ?
		WHERE tenantId = ? AND circuitBreakerConfigId = ? AND currentVersion = ?
	`
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		config.BreakerName, config.KeyStrategy, config.ErrorRatePercent, config.MinimumRequests,
		config.HalfOpenMaxRequests, config.SlowCallThreshold, config.SlowCallRatePercent,
		config.OpenTimeoutSeconds, config.WindowSizeSeconds, config.ErrorStatusCode, config.ErrorMessage,
		config.StorageType, config.StorageConfig, config.ConfigPriority, config.NoteText,
		config.EditTime, config.EditWho, config.CurrentVersion, config.OprSeqFlag, config.ActiveFlag,
		config.TenantId, config.CircuitBreakerConfigId, existing.CurrentVersion,
	}, true)
	if err != nil {
		return huberrors.WrapError(err, "更新熔断配置失败")
	}
	if result == 0 {
		return errors.New("熔断配置已被其他用户修改，请刷新后重试")
	}
	return nil
}

// DeactivateByTargetServiceId 关闭服务熔断时停用对应配置行。
func (dao *CircuitBreakerConfigDAO) DeactivateByTargetServiceId(ctx context.Context, tenantId, serviceId, operatorId string) error {
	existing, err := dao.GetByTargetServiceId(ctx, tenantId, serviceId)
	if err != nil || existing == nil {
		return err
	}
	if existing.ActiveFlag == "N" {
		return nil
	}
	sql := `
		UPDATE HUB_GW_CIRCUIT_BREAKER_CONFIG SET
			activeFlag = 'N', editTime = ?, editWho = ?, currentVersion = ?, oprSeqFlag = ?
		WHERE tenantId = ? AND circuitBreakerConfigId = ?
	`
	_, err = dao.db.Exec(ctx, sql, []interface{}{
		time.Now(), operatorId, existing.CurrentVersion + 1, random.Generate32BitRandomString(),
		tenantId, existing.CircuitBreakerConfigId,
	}, true)
	if err != nil {
		return huberrors.WrapError(err, "停用熔断配置失败")
	}
	return nil
}

// DeleteByTargetServiceId 删除服务时同步删除熔断配置。
func (dao *CircuitBreakerConfigDAO) DeleteByTargetServiceId(ctx context.Context, tenantId, serviceId string) error {
	if serviceId == "" {
		return nil
	}
	sql := `DELETE FROM HUB_GW_CIRCUIT_BREAKER_CONFIG WHERE tenantId = ? AND targetServiceId = ?`
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, serviceId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除熔断配置失败")
	}
	return nil
}

func applyCircuitBreakerDefaults(config *models.CircuitBreakerConfig) {
	if config.BreakerName == "" {
		config.BreakerName = "cb-" + config.TargetServiceId
	}
	config.KeyStrategy = circuitBreakerKeyStrategyNode
	if config.ErrorRatePercent <= 0 {
		config.ErrorRatePercent = 50
	}
	if config.MinimumRequests <= 0 {
		config.MinimumRequests = 10
	}
	if config.HalfOpenMaxRequests <= 0 {
		config.HalfOpenMaxRequests = 3
	}
	if config.SlowCallThreshold <= 0 {
		config.SlowCallThreshold = 60000
	}
	if config.SlowCallRatePercent <= 0 {
		config.SlowCallRatePercent = 50
	}
	if config.OpenTimeoutSeconds <= 0 {
		config.OpenTimeoutSeconds = 60
	}
	if config.WindowSizeSeconds <= 0 {
		config.WindowSizeSeconds = 60
	}
	if config.ErrorStatusCode == 0 {
		config.ErrorStatusCode = 503
	}
	if config.ErrorMessage == "" {
		config.ErrorMessage = "Service temporarily unavailable due to circuit breaker"
	}
	if config.StorageType == "" {
		config.StorageType = "memory"
	}
}

const circuitBreakerKeyStrategyNode = "node"
