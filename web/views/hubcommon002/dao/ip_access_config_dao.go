package dao

import (
	"context"
	"errors"
	"fmt"
	"gohub/pkg/database"
	"gohub/pkg/database/sqlutils"
	"gohub/pkg/utils/huberrors"
	"gohub/pkg/utils/random"
	"gohub/web/views/hubcommon002/models"
	"strings"
	"time"
)

// IpAccessConfigDAO IP访问控制配置数据访问对象
type IpAccessConfigDAO struct {
	db database.Database
}

// NewIpAccessConfigDAO 创建IP访问控制配置DAO
func NewIpAccessConfigDAO(db database.Database) *IpAccessConfigDAO {
	return &IpAccessConfigDAO{
		db: db,
	}
}

// generateIpAccessConfigId 生成IP访问配置ID
// 格式：IP + YYYYMMDD + HHMMSS + 4位随机数
// 示例：IP20240615143022A1B2
func (dao *IpAccessConfigDAO) generateIpAccessConfigId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("IP%s%s", timeStr, randomStr)
}



// isIpAccessConfigIdExists 检查IP访问配置ID是否已存在
func (dao *IpAccessConfigDAO) isIpAccessConfigIdExists(ctx context.Context, ipAccessConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_IP_ACCESS_CONFIG WHERE ipAccessConfigId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{ipAccessConfigId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueIpAccessConfigId 生成唯一的IP访问配置ID
func (dao *IpAccessConfigDAO) generateUniqueIpAccessConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		ipAccessConfigId := dao.generateIpAccessConfigId()
		
		exists, err := dao.isIpAccessConfigIdExists(ctx, ipAccessConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查IP访问配置ID是否存在失败")
		}
		
		if !exists {
			return ipAccessConfigId, nil
		}
		
		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一IP访问配置ID失败，已达到最大尝试次数")
}

// AddIpAccessConfig 添加IP访问控制配置
func (dao *IpAccessConfigDAO) AddIpAccessConfig(ctx context.Context, config *models.IpAccessConfig, operatorId string) error {
	if config.TenantId == "" || config.SecurityConfigId == "" {
		return errors.New("tenantId和securityConfigId不能为空")
	}

	// 自动生成IP访问配置ID
	if config.IpAccessConfigId == "" {
		generatedId, err := dao.generateUniqueIpAccessConfigId(ctx)
		if err != nil {
			return huberrors.WrapError(err, "生成IP访问配置ID失败")
		}
		config.IpAccessConfigId = generatedId
	}

	// 设置自动填充字段
	now := time.Now()
	config.AddTime = now
	config.AddWho = operatorId
	config.EditTime = now
	config.EditWho = operatorId
	config.OprSeqFlag = config.IpAccessConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	config.CurrentVersion = 1
	config.ActiveFlag = "Y"

	// 设置默认值
	if config.TrustXForwardedFor == "" {
		config.TrustXForwardedFor = "Y"
	}
	if config.TrustXRealIp == "" {
		config.TrustXRealIp = "Y"
	}
	if config.DefaultPolicy == "" {
		config.DefaultPolicy = "allow"
	}

	_, err := dao.db.Insert(ctx, "HUB_GW_IP_ACCESS_CONFIG", config, true)
	if err != nil {
		return huberrors.WrapError(err, "添加IP访问控制配置失败")
	}

	return nil
}

// GetIpAccessConfigBySecurityConfigId 根据安全配置ID获取IP访问控制配置
func (dao *IpAccessConfigDAO) GetIpAccessConfigBySecurityConfigId(ctx context.Context, securityConfigId, tenantId string) (*models.IpAccessConfig, error) {
	if securityConfigId == "" || tenantId == "" {
		return nil, errors.New("securityConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_IP_ACCESS_CONFIG 
		WHERE securityConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var config models.IpAccessConfig
	err := dao.db.QueryOne(ctx, &config, query, []interface{}{securityConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询IP访问控制配置失败")
	}

	return &config, nil
}

// GetIpAccessConfigById 根据IP访问配置ID获取配置
func (dao *IpAccessConfigDAO) GetIpAccessConfigById(ctx context.Context, ipAccessConfigId, tenantId string) (*models.IpAccessConfig, error) {
	if ipAccessConfigId == "" || tenantId == "" {
		return nil, errors.New("ipAccessConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_IP_ACCESS_CONFIG 
		WHERE ipAccessConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var config models.IpAccessConfig
	err := dao.db.QueryOne(ctx, &config, query, []interface{}{ipAccessConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询IP访问控制配置失败")
	}

	return &config, nil
}

// UpdateIpAccessConfig 更新IP访问控制配置
func (dao *IpAccessConfigDAO) UpdateIpAccessConfig(ctx context.Context, config *models.IpAccessConfig, operatorId string) error {
	if config.SecurityConfigId == "" || config.TenantId == "" {
		return errors.New("securityConfigId和tenantId不能为空")
	}

	// 首先获取当前配置
	currentConfig, err := dao.GetIpAccessConfigBySecurityConfigId(ctx, config.SecurityConfigId, config.TenantId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("IP访问控制配置不存在")
	}

	// 更新修改信息
	config.IpAccessConfigId = currentConfig.IpAccessConfigId
	config.EditTime = time.Now()
	config.EditWho = operatorId
	config.CurrentVersion = currentConfig.CurrentVersion + 1
	config.OprSeqFlag = config.IpAccessConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	sql := `
		UPDATE HUB_GW_IP_ACCESS_CONFIG SET
			configName = ?, defaultPolicy = ?, whitelistIps = ?, blacklistIps = ?,
			whitelistCidrs = ?, blacklistCidrs = ?, trustXForwardedFor = ?, trustXRealIp = ?,
			reserved1 = ?, reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?,
			extProperty = ?, noteText = ?, editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?
		WHERE ipAccessConfigId = ? AND tenantId = ? AND currentVersion = ?
	`

	result, err := dao.db.Exec(ctx, sql, []interface{}{
		config.ConfigName, config.DefaultPolicy, config.WhitelistIps, config.BlacklistIps,
		config.WhitelistCidrs, config.BlacklistCidrs, config.TrustXForwardedFor, config.TrustXRealIp,
		config.Reserved1, config.Reserved2, config.Reserved3, config.Reserved4, config.Reserved5,
		config.ExtProperty, config.NoteText, config.EditTime, config.EditWho, config.OprSeqFlag, config.CurrentVersion,
		config.IpAccessConfigId, config.TenantId, currentConfig.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新IP访问控制配置失败")
	}

	if result == 0 {
		return errors.New("IP访问控制配置数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteIpAccessConfig 删除IP访问控制配置
func (dao *IpAccessConfigDAO) DeleteIpAccessConfig(ctx context.Context, securityConfigId, tenantId, operatorId string) error {
	if securityConfigId == "" || tenantId == "" {
		return errors.New("securityConfigId和tenantId不能为空")
	}

	sql := `
		UPDATE HUB_GW_IP_ACCESS_CONFIG SET
			activeFlag = 'N', editTime = ?, editWho = ?
		WHERE securityConfigId = ? AND tenantId = ?
	`

	result, err := dao.db.Exec(ctx, sql, []interface{}{time.Now(), operatorId, securityConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除IP访问控制配置失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的IP访问控制配置")
	}

	return nil
}

// ListIpAccessConfigs 获取IP访问控制配置列表
func (dao *IpAccessConfigDAO) ListIpAccessConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.IpAccessConfig, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_IP_ACCESS_CONFIG WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY addTime DESC"

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, []interface{}{tenantId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询IP访问控制配置总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.IpAccessConfig{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append([]interface{}{tenantId}, paginationArgs...)

	// 执行分页查询
	var configs []*models.IpAccessConfig
	err = dao.db.Query(ctx, &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询IP访问控制配置列表失败")
	}

	return configs, total, nil
} 