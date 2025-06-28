package dao

import (
	"context"
	"gohub/pkg/database"
	"gohub/web/views/hubplugin/common/models"
)

// ConfigGroupDao 工具配置分组数据访问对象
type ConfigGroupDao struct {
	db database.Database
}

// NewConfigGroupDao 创建工具配置分组数据访问对象
func NewConfigGroupDao(db database.Database) *ConfigGroupDao {
	return &ConfigGroupDao{
		db: db,
	}
}

// Add 添加工具配置分组
func (d *ConfigGroupDao) Add(ctx context.Context, configGroup *models.ToolConfigGroup) (int64, error) {
	return d.db.Insert(ctx, configGroup.TableName(), configGroup, true)
}

// GetById 根据ID获取工具配置分组
func (d *ConfigGroupDao) GetById(ctx context.Context, tenantId, configGroupId string) (*models.ToolConfigGroup, error) {
	configGroup := &models.ToolConfigGroup{}
	query := "SELECT * FROM " + configGroup.TableName() + " WHERE tenantId = ? AND configGroupId = ? AND activeFlag = 'Y'"
	err := d.db.QueryOne(ctx, configGroup, query, []interface{}{tenantId, configGroupId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return configGroup, nil
}

// Update 更新工具配置分组
func (d *ConfigGroupDao) Update(ctx context.Context, configGroup *models.ToolConfigGroup) (int64, error) {
	query := "UPDATE " + configGroup.TableName() + " SET " +
		"groupName = ?, groupDescription = ?, parentGroupId = ?, groupLevel = ?, " +
		"groupPath = ?, groupType = ?, sortOrder = ?, groupIcon = ?, groupColor = ?, " +
		"accessLevel = ?, allowedUsers = ?, allowedRoles = ?, editTime = ?, editWho = ?, " +
		"oprSeqFlag = ?, noteText = ?, extProperty = ?, currentVersion = currentVersion + 1 " +
		"WHERE tenantId = ? AND configGroupId = ? AND activeFlag = 'Y'"

	args := []interface{}{
		configGroup.GroupName, configGroup.GroupDescription, configGroup.ParentGroupId,
		configGroup.GroupLevel, configGroup.GroupPath, configGroup.GroupType,
		configGroup.SortOrder, configGroup.GroupIcon, configGroup.GroupColor,
		configGroup.AccessLevel, configGroup.AllowedUsers, configGroup.AllowedRoles,
		configGroup.EditTime, configGroup.EditWho, configGroup.OprSeqFlag,
		configGroup.NoteText, configGroup.ExtProperty,
		configGroup.TenantId, configGroup.ConfigGroupId,
	}

	return d.db.Exec(ctx, query, args, true)
}

// Delete 删除工具配置分组（逻辑删除）
func (d *ConfigGroupDao) Delete(ctx context.Context, tenantId, configGroupId, operatorId string) (int64, error) {
	query := "UPDATE " + (&models.ToolConfigGroup{}).TableName() + " SET activeFlag = 'N', editWho = ?, editTime = NOW() " +
		"WHERE tenantId = ? AND configGroupId = ? AND activeFlag = 'Y'"
	return d.db.Exec(ctx, query, []interface{}{operatorId, tenantId, configGroupId}, true)
}

// Query 查询工具配置分组列表
func (d *ConfigGroupDao) Query(ctx context.Context, params map[string]interface{}, page, pageSize int) ([]*models.ToolConfigGroup, int64, error) {
	var configGroups []*models.ToolConfigGroup

	// 构建查询条件
	whereClause := "WHERE activeFlag = 'Y' "
	args := []interface{}{}

	// 租户ID条件（必需）
	if tenantId, ok := params["tenantId"].(string); ok && tenantId != "" {
		whereClause += "AND tenantId = ? "
		args = append(args, tenantId)
	}

	// 分组名称模糊查询
	if groupName, ok := params["groupName"].(string); ok && groupName != "" {
		whereClause += "AND groupName LIKE ? "
		args = append(args, "%"+groupName+"%")
	}

	// 父分组ID
	if parentGroupId, ok := params["parentGroupId"].(string); ok && parentGroupId != "" {
		whereClause += "AND parentGroupId = ? "
		args = append(args, parentGroupId)
	}

	// 分组类型
	if groupType, ok := params["groupType"].(string); ok && groupType != "" {
		whereClause += "AND groupType = ? "
		args = append(args, groupType)
	}

	// 计算总记录数
	countQuery := "SELECT COUNT(*) FROM " + (&models.ToolConfigGroup{}).TableName() + " " + whereClause
	var result struct {
		Count int64 `db:"COUNT(*)"`
	}
	err := d.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, err
	}
	total := result.Count

	if total == 0 {
		return []*models.ToolConfigGroup{}, 0, nil
	}

	// 查询数据
	query := "SELECT * FROM " + (&models.ToolConfigGroup{}).TableName() + " " + whereClause +
		"ORDER BY sortOrder ASC, addTime DESC LIMIT ?, ?"
	args = append(args, (page-1)*pageSize, pageSize)

	err = d.db.Query(ctx, &configGroups, query, args, true)
	if err != nil {
		return nil, 0, err
	}

	return configGroups, total, nil
}

// GetByParentId 根据父分组ID获取子分组列表
func (d *ConfigGroupDao) GetByParentId(ctx context.Context, tenantId, parentGroupId string) ([]*models.ToolConfigGroup, error) {
	var configGroups []*models.ToolConfigGroup
	query := "SELECT * FROM " + (&models.ToolConfigGroup{}).TableName() + 
		" WHERE tenantId = ? AND parentGroupId = ? AND activeFlag = 'Y' ORDER BY sortOrder ASC, addTime DESC"
	args := []interface{}{tenantId, parentGroupId}

	err := d.db.Query(ctx, &configGroups, query, args, true)
	if err != nil {
		return nil, err
	}

	return configGroups, nil
} 