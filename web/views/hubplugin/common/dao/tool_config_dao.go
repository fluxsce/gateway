package dao

import (
	"context"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/web/views/hubplugin/common/models"
	"strings"
)

// ToolConfigDao 工具配置数据访问对象
type ToolConfigDao struct {
	db database.Database
}

// NewToolConfigDao 创建工具配置数据访问对象
func NewToolConfigDao(db database.Database) *ToolConfigDao {
	return &ToolConfigDao{
		db: db,
	}
}

// ptrToString 安全地将字符串指针转换为接口值
func ptrToString(ptr *string) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}

// ptrToInt 安全地将整数指针转换为接口值
func ptrToInt(ptr *int) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}

// Add 添加工具配置
func (d *ToolConfigDao) Add(ctx context.Context, toolConfig *models.ToolConfig) (int64, error) {
	return d.db.Insert(ctx, toolConfig.TableName(), toolConfig, true)
}

// GetById 根据ID获取工具配置
func (d *ToolConfigDao) GetById(ctx context.Context, tenantId, toolConfigId string) (*models.ToolConfig, error) {
	toolConfig := &models.ToolConfig{}
	query := "SELECT * FROM " + toolConfig.TableName() + " WHERE tenantId = ? AND toolConfigId = ?"
	err := d.db.QueryOne(ctx, toolConfig, query, []interface{}{tenantId, toolConfigId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toolConfig, nil
}

// Update 更新工具配置
func (d *ToolConfigDao) Update(ctx context.Context, toolConfig *models.ToolConfig) (int64, error) {
	// 构建动态更新语句，跳过nil字段
	var setParts []string
	var args []interface{}

	// 基础字段（总是更新）
	setParts = append(setParts, "toolName = ?", "toolType = ?", "configName = ?")
	args = append(args, toolConfig.ToolName, toolConfig.ToolType, toolConfig.ConfigName)

	// 可选字段（根据是否为nil决定是否更新）
	if toolConfig.ToolVersion != nil {
		setParts = append(setParts, "toolVersion = ?")
		args = append(args, ptrToString(toolConfig.ToolVersion))
	}

	if toolConfig.ConfigDescription != nil {
		setParts = append(setParts, "configDescription = ?")
		args = append(args, ptrToString(toolConfig.ConfigDescription))
	}

	if toolConfig.ConfigGroupId != nil {
		setParts = append(setParts, "configGroupId = ?")
		args = append(args, ptrToString(toolConfig.ConfigGroupId))
	}

	if toolConfig.ConfigGroupName != nil {
		setParts = append(setParts, "configGroupName = ?")
		args = append(args, ptrToString(toolConfig.ConfigGroupName))
	}

	if toolConfig.HostAddress != nil {
		setParts = append(setParts, "hostAddress = ?")
		args = append(args, ptrToString(toolConfig.HostAddress))
	}

	if toolConfig.PortNumber != nil {
		setParts = append(setParts, "portNumber = ?")
		args = append(args, ptrToInt(toolConfig.PortNumber))
	}

	if toolConfig.ProtocolType != nil {
		setParts = append(setParts, "protocolType = ?")
		args = append(args, ptrToString(toolConfig.ProtocolType))
	}

	if toolConfig.AuthType != nil {
		setParts = append(setParts, "authType = ?")
		args = append(args, ptrToString(toolConfig.AuthType))
	}

	if toolConfig.UserName != nil {
		setParts = append(setParts, "userName = ?")
		args = append(args, ptrToString(toolConfig.UserName))
	}

	// 敏感字段：只有当不为nil时才更新
	if toolConfig.PasswordEncrypted != nil {
		setParts = append(setParts, "passwordEncrypted = ?")
		args = append(args, ptrToString(toolConfig.PasswordEncrypted))
	}

	if toolConfig.KeyFilePath != nil {
		setParts = append(setParts, "keyFilePath = ?")
		args = append(args, ptrToString(toolConfig.KeyFilePath))
	}

	if toolConfig.KeyFileContent != nil {
		setParts = append(setParts, "keyFileContent = ?")
		args = append(args, ptrToString(toolConfig.KeyFileContent))
	}

	if toolConfig.ConfigParameters != nil {
		setParts = append(setParts, "configParameters = ?")
		args = append(args, ptrToString(toolConfig.ConfigParameters))
	}

	if toolConfig.EnvironmentVariables != nil {
		setParts = append(setParts, "environmentVariables = ?")
		args = append(args, ptrToString(toolConfig.EnvironmentVariables))
	}

	if toolConfig.CustomSettings != nil {
		setParts = append(setParts, "customSettings = ?")
		args = append(args, ptrToString(toolConfig.CustomSettings))
	}

	// 状态字段（总是更新）
	setParts = append(setParts, "configStatus = ?", "defaultFlag = ?")
	args = append(args, toolConfig.ConfigStatus, toolConfig.DefaultFlag)

	if toolConfig.PriorityLevel != nil {
		setParts = append(setParts, "priorityLevel = ?")
		args = append(args, ptrToInt(toolConfig.PriorityLevel))
	}

	if toolConfig.EncryptionType != nil {
		setParts = append(setParts, "encryptionType = ?")
		args = append(args, ptrToString(toolConfig.EncryptionType))
	}

	if toolConfig.EncryptionKey != nil {
		setParts = append(setParts, "encryptionKey = ?")
		args = append(args, ptrToString(toolConfig.EncryptionKey))
	}

	// 系统字段（总是更新）
	setParts = append(setParts, "editTime = ?", "editWho = ?", "oprSeqFlag = ?", "currentVersion = currentVersion + 1")
	args = append(args, toolConfig.EditTime, toolConfig.EditWho, toolConfig.OprSeqFlag)

	if toolConfig.NoteText != nil {
		setParts = append(setParts, "noteText = ?")
		args = append(args, ptrToString(toolConfig.NoteText))
	}

	if toolConfig.ExtProperty != nil {
		setParts = append(setParts, "extProperty = ?")
		args = append(args, ptrToString(toolConfig.ExtProperty))
	}

	// 构建完整的UPDATE语句
	query := "UPDATE " + toolConfig.TableName() + " SET " + strings.Join(setParts, ", ") +
		" WHERE tenantId = ? AND toolConfigId = ? "

	// 添加WHERE条件的参数
	args = append(args, toolConfig.TenantId, toolConfig.ToolConfigId)

	return d.db.Exec(ctx, query, args, true)
}

// Delete 删除工具配置（物理删除）
func (d *ToolConfigDao) Delete(ctx context.Context, tenantId, toolConfigId, operatorId string) (int64, error) {
	query := "DELETE FROM " + (&models.ToolConfig{}).TableName() +
		" WHERE tenantId = ? AND toolConfigId = ?"
	return d.db.Exec(ctx, query, []interface{}{tenantId, toolConfigId}, true)
}

// Query 查询工具配置列表
func (d *ToolConfigDao) Query(ctx context.Context, params map[string]interface{}, page, pageSize int) ([]*models.ToolConfig, int64, error) {
	var toolConfigs []*models.ToolConfig

	// 构建基础查询条件
	whereClause := "WHERE 1=1 "
	args := []interface{}{}

	// 租户ID条件（必需）
	if tenantId, ok := params["tenantId"].(string); ok && tenantId != "" {
		whereClause += "AND tenantId = ? "
		args = append(args, tenantId)
	}

	// 工具名称模糊查询
	if toolName, ok := params["toolName"].(string); ok && toolName != "" {
		whereClause += "AND toolName LIKE ? "
		args = append(args, "%"+toolName+"%")
	}

	// 工具类型
	if toolType, ok := params["toolType"].(string); ok && toolType != "" {
		whereClause += "AND toolType = ? "
		args = append(args, toolType)
	}

	// 配置名称模糊查询
	if configName, ok := params["configName"].(string); ok && configName != "" {
		whereClause += "AND configName LIKE ? "
		args = append(args, "%"+configName+"%")
	}

	// 配置分组ID
	if configGroupId, ok := params["configGroupId"].(string); ok && configGroupId != "" {
		whereClause += "AND configGroupId = ? "
		args = append(args, configGroupId)
	}

	// 主机地址模糊查询
	if hostAddress, ok := params["hostAddress"].(string); ok && hostAddress != "" {
		whereClause += "AND hostAddress LIKE ? "
		args = append(args, "%"+hostAddress+"%")
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM " + (&models.ToolConfig{}).TableName() + " " + whereClause + "ORDER BY addTime DESC"

	// 使用标准化的COUNT查询构建
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, err
	}

	// 计算总记录数
	var result struct {
		Count int64 `db:"COUNT(*)"`
	}
	err = d.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, err
	}
	total := result.Count

	if total == 0 {
		return []*models.ToolConfig{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(d.db)

	// 使用标准化的分页查询构建
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, err
	}

	// 合并查询参数
	allArgs := append(args, paginationArgs...)

	// 查询数据
	err = d.db.Query(ctx, &toolConfigs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, err
	}

	return toolConfigs, total, nil
}

// GetByGroupId 根据分组ID获取工具配置列表
func (d *ToolConfigDao) GetByGroupId(ctx context.Context, tenantId, configGroupId string) ([]*models.ToolConfig, error) {
	var toolConfigs []*models.ToolConfig
	query := "SELECT * FROM " + (&models.ToolConfig{}).TableName() +
		" WHERE tenantId = ? AND configGroupId = ? ORDER BY addTime DESC"
	args := []interface{}{tenantId, configGroupId}

	err := d.db.Query(ctx, &toolConfigs, query, args, true)
	if err != nil {
		return nil, err
	}

	return toolConfigs, nil
}

// TestConnection 测试连接配置
func (d *ToolConfigDao) TestConnection(ctx context.Context, toolConfig *models.ToolConfig) error {
	// TODO: 实现SFTP连接测试逻辑
	// 这里可以调用SFTP客户端进行连接测试
	return nil
}
