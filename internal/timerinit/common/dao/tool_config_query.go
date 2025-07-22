package dao

import (
	"context"
	"fmt"
	"gohub/internal/types/tooltypes"
	"gohub/pkg/database"
	"strings"
	"time"
)

// ToolConfigQuery 定义工具配置查询条件结构
type ToolConfigQuery struct {
	// 基础查询条件
	ToolConfigId    *string `json:"toolConfigId,omitempty"`
	TenantId        *string `json:"tenantId,omitempty"`
	ToolName        *string `json:"toolName,omitempty"`
	ToolType        *string `json:"toolType,omitempty"`
	ToolVersion     *string `json:"toolVersion,omitempty"`
	ConfigName      *string `json:"configName,omitempty"`
	
	// 分组查询
	ConfigGroupId   *string `json:"configGroupId,omitempty"`
	ConfigGroupName *string `json:"configGroupName,omitempty"`
	
	// 连接配置查询
	HostAddress     *string `json:"hostAddress,omitempty"`
	PortNumber      *int    `json:"portNumber,omitempty"`
	ProtocolType    *string `json:"protocolType,omitempty"`
	
	// 认证配置查询
	AuthType        *string `json:"authType,omitempty"`
	UserName        *string `json:"userName,omitempty"`
	
	// 状态和控制查询
	ConfigStatus    *string `json:"configStatus,omitempty"`
	ConfigStatusList []string `json:"configStatusList,omitempty"`
	DefaultFlag     *string `json:"defaultFlag,omitempty"`
	ActiveFlag      *string `json:"activeFlag,omitempty"`
	
	// 优先级查询
	MinPriorityLevel *int `json:"minPriorityLevel,omitempty"`
	MaxPriorityLevel *int `json:"maxPriorityLevel,omitempty"`
	PriorityLevel    *int `json:"priorityLevel,omitempty"`
	
	// 安全配置查询
	EncryptionType  *string `json:"encryptionType,omitempty"`
	
	// 创建和修改时间查询
	AddTimeFrom     *time.Time `json:"addTimeFrom,omitempty"`
	AddTimeTo       *time.Time `json:"addTimeTo,omitempty"`
	EditTimeFrom    *time.Time `json:"editTimeFrom,omitempty"`
	EditTimeTo      *time.Time `json:"editTimeTo,omitempty"`
	AddWho          *string    `json:"addWho,omitempty"`
	EditWho         *string    `json:"editWho,omitempty"`
	
	// 模糊查询
	ToolNameLike         *string `json:"toolNameLike,omitempty"`
	ConfigNameLike       *string `json:"configNameLike,omitempty"`
	ConfigDescriptionLike *string `json:"configDescriptionLike,omitempty"`
	HostAddressLike      *string `json:"hostAddressLike,omitempty"`
	UserNameLike         *string `json:"userNameLike,omitempty"`
	NoteTextLike         *string `json:"noteTextLike,omitempty"`
	
	// 特殊查询条件
	OnlyEnabled     *bool      `json:"onlyEnabled,omitempty"`     // 只返回启用的配置
	OnlyDefault     *bool      `json:"onlyDefault,omitempty"`     // 只返回默认配置
	ToolTypeList    []string   `json:"toolTypeList,omitempty"`    // 工具类型列表
	AuthTypeList    []string   `json:"authTypeList,omitempty"`    // 认证类型列表
	
	// 排序和分页
	OrderBy         string     `json:"orderBy,omitempty"`         // 排序字段
	OrderDirection  string     `json:"orderDirection,omitempty"`  // ASC/DESC
	PageNum         int        `json:"pageNum,omitempty"`         // 页码，从1开始
	PageSize        int        `json:"pageSize,omitempty"`        // 每页大小
	Offset          int        `json:"offset,omitempty"`          // 偏移量
	Limit           int        `json:"limit,omitempty"`           // 限制数量
}

// ToolConfigQueryResult 查询结果结构
type ToolConfigQueryResult struct {
	Configs    []tooltypes.ToolConfig `json:"configs"`
	Total      int64                  `json:"total"`
	PageNum    int                    `json:"pageNum"`
	PageSize   int                    `json:"pageSize"`
	TotalPages int                    `json:"totalPages"`
}

// ToolConfigDAO 工具配置数据访问对象
type ToolConfigDAO struct {
	db database.Database
}

// NewToolConfigDAO 创建工具配置DAO实例
func NewToolConfigDAO(db database.Database) *ToolConfigDAO {
	return &ToolConfigDAO{db: db}
}

// BuildWhere 构建WHERE条件语句
func (q *ToolConfigQuery) BuildWhere() (string, []interface{}) {
	var conditions []string
	var args []interface{}
	
	// 基础条件
	if q.ToolConfigId != nil {
		conditions = append(conditions, "toolConfigId = ?")
		args = append(args, *q.ToolConfigId)
	}
	
	if q.TenantId != nil {
		conditions = append(conditions, "tenantId = ?")
		args = append(args, *q.TenantId)
	}
	
	if q.ToolName != nil {
		conditions = append(conditions, "toolName = ?")
		args = append(args, *q.ToolName)
	}
	
	if q.ToolType != nil {
		conditions = append(conditions, "toolType = ?")
		args = append(args, *q.ToolType)
	}
	
	if q.ToolVersion != nil {
		conditions = append(conditions, "toolVersion = ?")
		args = append(args, *q.ToolVersion)
	}
	
	if q.ConfigName != nil {
		conditions = append(conditions, "configName = ?")
		args = append(args, *q.ConfigName)
	}
	
	// 分组条件
	if q.ConfigGroupId != nil {
		conditions = append(conditions, "configGroupId = ?")
		args = append(args, *q.ConfigGroupId)
	}
	
	if q.ConfigGroupName != nil {
		conditions = append(conditions, "configGroupName = ?")
		args = append(args, *q.ConfigGroupName)
	}
	
	// 连接配置条件
	if q.HostAddress != nil {
		conditions = append(conditions, "hostAddress = ?")
		args = append(args, *q.HostAddress)
	}
	
	if q.PortNumber != nil {
		conditions = append(conditions, "portNumber = ?")
		args = append(args, *q.PortNumber)
	}
	
	if q.ProtocolType != nil {
		conditions = append(conditions, "protocolType = ?")
		args = append(args, *q.ProtocolType)
	}
	
	// 认证配置条件
	if q.AuthType != nil {
		conditions = append(conditions, "authType = ?")
		args = append(args, *q.AuthType)
	}
	
	if q.UserName != nil {
		conditions = append(conditions, "userName = ?")
		args = append(args, *q.UserName)
	}
	
	// 状态和控制条件
	if q.ConfigStatus != nil {
		conditions = append(conditions, "configStatus = ?")
		args = append(args, *q.ConfigStatus)
	}
	
	if len(q.ConfigStatusList) > 0 {
		placeholders := make([]string, len(q.ConfigStatusList))
		for i, status := range q.ConfigStatusList {
			placeholders[i] = "?"
			args = append(args, status)
		}
		conditions = append(conditions, fmt.Sprintf("configStatus IN (%s)", strings.Join(placeholders, ",")))
	}
	
	if q.DefaultFlag != nil {
		conditions = append(conditions, "defaultFlag = ?")
		args = append(args, *q.DefaultFlag)
	}
	
	if q.ActiveFlag != nil {
		conditions = append(conditions, "activeFlag = ?")
		args = append(args, *q.ActiveFlag)
	}
	
	// 优先级条件
	if q.PriorityLevel != nil {
		conditions = append(conditions, "priorityLevel = ?")
		args = append(args, *q.PriorityLevel)
	}
	
	if q.MinPriorityLevel != nil {
		conditions = append(conditions, "priorityLevel >= ?")
		args = append(args, *q.MinPriorityLevel)
	}
	
	if q.MaxPriorityLevel != nil {
		conditions = append(conditions, "priorityLevel <= ?")
		args = append(args, *q.MaxPriorityLevel)
	}
	
	// 安全配置条件
	if q.EncryptionType != nil {
		conditions = append(conditions, "encryptionType = ?")
		args = append(args, *q.EncryptionType)
	}
	
	// 创建和修改时间条件
	if q.AddTimeFrom != nil {
		conditions = append(conditions, "addTime >= ?")
		args = append(args, *q.AddTimeFrom)
	}
	
	if q.AddTimeTo != nil {
		conditions = append(conditions, "addTime <= ?")
		args = append(args, *q.AddTimeTo)
	}
	
	if q.EditTimeFrom != nil {
		conditions = append(conditions, "editTime >= ?")
		args = append(args, *q.EditTimeFrom)
	}
	
	if q.EditTimeTo != nil {
		conditions = append(conditions, "editTime <= ?")
		args = append(args, *q.EditTimeTo)
	}
	
	if q.AddWho != nil {
		conditions = append(conditions, "addWho = ?")
		args = append(args, *q.AddWho)
	}
	
	if q.EditWho != nil {
		conditions = append(conditions, "editWho = ?")
		args = append(args, *q.EditWho)
	}
	
	// 模糊查询条件
	if q.ToolNameLike != nil {
		conditions = append(conditions, "toolName LIKE ?")
		args = append(args, "%"+*q.ToolNameLike+"%")
	}
	
	if q.ConfigNameLike != nil {
		conditions = append(conditions, "configName LIKE ?")
		args = append(args, "%"+*q.ConfigNameLike+"%")
	}
	
	if q.ConfigDescriptionLike != nil {
		conditions = append(conditions, "configDescription LIKE ?")
		args = append(args, "%"+*q.ConfigDescriptionLike+"%")
	}
	
	if q.HostAddressLike != nil {
		conditions = append(conditions, "hostAddress LIKE ?")
		args = append(args, "%"+*q.HostAddressLike+"%")
	}
	
	if q.UserNameLike != nil {
		conditions = append(conditions, "userName LIKE ?")
		args = append(args, "%"+*q.UserNameLike+"%")
	}
	
	if q.NoteTextLike != nil {
		conditions = append(conditions, "noteText LIKE ?")
		args = append(args, "%"+*q.NoteTextLike+"%")
	}
	
	// 特殊条件
	if q.OnlyEnabled != nil && *q.OnlyEnabled {
		conditions = append(conditions, "configStatus = ?")
		args = append(args, tooltypes.ConfigStatusEnabled)
	}
	
	if q.OnlyDefault != nil && *q.OnlyDefault {
		conditions = append(conditions, "defaultFlag = ?")
		args = append(args, tooltypes.DefaultFlagYes)
	}
	
	if len(q.ToolTypeList) > 0 {
		placeholders := make([]string, len(q.ToolTypeList))
		for i, toolType := range q.ToolTypeList {
			placeholders[i] = "?"
			args = append(args, toolType)
		}
		conditions = append(conditions, fmt.Sprintf("toolType IN (%s)", strings.Join(placeholders, ",")))
	}
	
	if len(q.AuthTypeList) > 0 {
		placeholders := make([]string, len(q.AuthTypeList))
		for i, authType := range q.AuthTypeList {
			placeholders[i] = "?"
			args = append(args, authType)
		}
		conditions = append(conditions, fmt.Sprintf("authType IN (%s)", strings.Join(placeholders, ",")))
	}
	
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}
	
	return where, args
}

// BuildOrderBy 构建ORDER BY语句
func (q *ToolConfigQuery) BuildOrderBy() string {
	if q.OrderBy == "" {
		return "ORDER BY priorityLevel DESC, addTime DESC"
	}
	
	direction := "ASC"
	if strings.ToUpper(q.OrderDirection) == "DESC" {
		direction = "DESC"
	}
	
	return fmt.Sprintf("ORDER BY %s %s", q.OrderBy, direction)
}

// BuildPagination 构建分页语句
func (q *ToolConfigQuery) BuildPagination() string {
	if q.Limit > 0 {
		if q.Offset > 0 {
			return fmt.Sprintf("LIMIT %d OFFSET %d", q.Limit, q.Offset)
		}
		return fmt.Sprintf("LIMIT %d", q.Limit)
	}
	
	if q.PageSize > 0 {
		offset := 0
		if q.PageNum > 1 {
			offset = (q.PageNum - 1) * q.PageSize
		}
		return fmt.Sprintf("LIMIT %d OFFSET %d", q.PageSize, offset)
	}
	
	return ""
}

// QueryConfigs 查询工具配置列表
func (dao *ToolConfigDAO) QueryConfigs(ctx context.Context, query *ToolConfigQuery) (*ToolConfigQueryResult, error) {
	tableName := (&tooltypes.ToolConfig{}).TableName()
	
	// 构建查询条件
	where, args := query.BuildWhere()
	orderBy := query.BuildOrderBy()
	pagination := query.BuildPagination()
	
	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", tableName, where)
	var total int64
	err := dao.db.QueryOne(ctx, &total, countSQL, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询工具配置总数失败: %w", err)
	}
	
	// 查询数据
	selectSQL := fmt.Sprintf("SELECT * FROM %s %s %s %s", tableName, where, orderBy, pagination)
	var configs []tooltypes.ToolConfig
	err = dao.db.Query(ctx, &configs, selectSQL, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询工具配置列表失败: %w", err)
	}
	
	// 计算分页信息
	result := &ToolConfigQueryResult{
		Configs: configs,
		Total:   total,
	}
	
	if query.PageSize > 0 {
		result.PageNum = query.PageNum
		if result.PageNum == 0 {
			result.PageNum = 1
		}
		result.PageSize = query.PageSize
		result.TotalPages = int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	}
	
	return result, nil
}

// GetConfigById 根据ID获取工具配置
func (dao *ToolConfigDAO) GetConfigById(ctx context.Context, tenantId, toolConfigId string) (*tooltypes.ToolConfig, error) {
	tableName := (&tooltypes.ToolConfig{}).TableName()
	sql := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND toolConfigId = ? AND activeFlag = ?", tableName)
	
	var config tooltypes.ToolConfig
	err := dao.db.QueryOne(ctx, &config, sql, []interface{}{tenantId, toolConfigId, tooltypes.ActiveFlagYes}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询工具配置失败: %w", err)
	}
	
	return &config, nil
}

// CreateConfig 创建工具配置
func (dao *ToolConfigDAO) CreateConfig(ctx context.Context, config *tooltypes.ToolConfig) error {
	// 设置活动标记为Y
	config.ActiveFlag = tooltypes.ActiveFlagYes
	
	tableName := (&tooltypes.ToolConfig{}).TableName()
	_, err := dao.db.Insert(ctx, tableName, config, true)
	if err != nil {
		return fmt.Errorf("创建工具配置失败: %w", err)
	}
	return nil
}

// UpdateConfig 更新工具配置
func (dao *ToolConfigDAO) UpdateConfig(ctx context.Context, config *tooltypes.ToolConfig) error {
	tableName := (&tooltypes.ToolConfig{}).TableName()
	where := "tenantId = ? AND toolConfigId = ?"
	args := []interface{}{config.TenantId, config.ToolConfigId}
	
	_, err := dao.db.Update(ctx, tableName, config, where, args, true)
	if err != nil {
		return fmt.Errorf("更新工具配置失败: %w", err)
	}
	return nil
}

// DeleteConfig 删除工具配置（软删除）
func (dao *ToolConfigDAO) DeleteConfig(ctx context.Context, tenantId, toolConfigId string) error {
	tableName := (&tooltypes.ToolConfig{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = 'N' WHERE tenantId = ? AND toolConfigId = ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, toolConfigId}, true)
	if err != nil {
		return fmt.Errorf("删除工具配置失败: %w", err)
	}
	return nil
}

// GetConfigsByType 根据工具类型获取配置列表
func (dao *ToolConfigDAO) GetConfigsByType(ctx context.Context, tenantId, toolType string) ([]tooltypes.ToolConfig, error) {
	onlyEnabled := true
	activeFlag := tooltypes.ActiveFlagYes
	
	query := &ToolConfigQuery{
		TenantId:    &tenantId,
		ToolType:    &toolType,
		ActiveFlag:  &activeFlag,
		OnlyEnabled: &onlyEnabled,
		OrderBy:     "priorityLevel",
		OrderDirection: "DESC",
	}
	
	result, err := dao.QueryConfigs(ctx, query)
	if err != nil {
		return nil, err
	}
	
	return result.Configs, nil
}

// GetDefaultConfig 获取指定工具类型的默认配置
func (dao *ToolConfigDAO) GetDefaultConfig(ctx context.Context, tenantId, toolType string) (*tooltypes.ToolConfig, error) {
	onlyDefault := true
	onlyEnabled := true
	activeFlag := tooltypes.ActiveFlagYes
	
	query := &ToolConfigQuery{
		TenantId:    &tenantId,
		ToolType:    &toolType,
		ActiveFlag:  &activeFlag,
		OnlyDefault: &onlyDefault,
		OnlyEnabled: &onlyEnabled,
		Limit:       1,
	}
	
	result, err := dao.QueryConfigs(ctx, query)
	if err != nil {
		return nil, err
	}
	
	if len(result.Configs) == 0 {
		return nil, nil
	}
	
	return &result.Configs[0], nil
}

// GetConfigsByGroup 根据配置分组获取配置列表
func (dao *ToolConfigDAO) GetConfigsByGroup(ctx context.Context, tenantId, configGroupId string) ([]tooltypes.ToolConfig, error) {
	onlyEnabled := true
	activeFlag := tooltypes.ActiveFlagYes
	
	query := &ToolConfigQuery{
		TenantId:      &tenantId,
		ConfigGroupId: &configGroupId,
		ActiveFlag:    &activeFlag,
		OnlyEnabled:   &onlyEnabled,
		OrderBy:       "priorityLevel",
		OrderDirection: "DESC",
	}
	
	result, err := dao.QueryConfigs(ctx, query)
	if err != nil {
		return nil, err
	}
	
	return result.Configs, nil
}

// GetEnabledConfigs 获取所有启用的配置
func (dao *ToolConfigDAO) GetEnabledConfigs(ctx context.Context, tenantId string) ([]tooltypes.ToolConfig, error) {
	onlyEnabled := true
	activeFlag := tooltypes.ActiveFlagYes
	
	query := &ToolConfigQuery{
		TenantId:    &tenantId,
		ActiveFlag:  &activeFlag,
		OnlyEnabled: &onlyEnabled,
		OrderBy:     "toolType",
		OrderDirection: "ASC",
	}
	
	result, err := dao.QueryConfigs(ctx, query)
	if err != nil {
		return nil, err
	}
	
	return result.Configs, nil
} 