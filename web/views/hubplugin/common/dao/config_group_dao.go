package dao

import (
	"context"
	"gohub/pkg/database"
	"gohub/pkg/database/sqlutils"
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
	query := "SELECT * FROM " + configGroup.TableName() + " WHERE tenantId = ? AND configGroupId = ?"
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
		"WHERE tenantId = ? AND configGroupId = ?"

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

// Delete 删除工具配置分组（物理删除）
func (d *ConfigGroupDao) Delete(ctx context.Context, tenantId, configGroupId, operatorId string) (int64, error) {
	query := "DELETE FROM " + (&models.ToolConfigGroup{}).TableName() + 
		" WHERE tenantId = ? AND configGroupId = ?"
	return d.db.Exec(ctx, query, []interface{}{tenantId, configGroupId}, true)
}

// Query 查询工具配置分组列表
func (d *ConfigGroupDao) Query(ctx context.Context, params map[string]interface{}, page, pageSize int) ([]*models.ToolConfigGroup, int64, error) {
	var configGroups []*models.ToolConfigGroup

	// 构建查询条件
	whereClause := "WHERE 1=1 "
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

	// 构建基础查询语句
	baseQuery := "SELECT * FROM " + (&models.ToolConfigGroup{}).TableName() + " " + whereClause + "ORDER BY sortOrder ASC, addTime DESC"

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, err
	}

	// 执行统计查询
	var result struct {
		Count int64 `db:"COUNT(*)"`
	}
	err = d.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, err
	}
	total := result.Count

	if total == 0 {
		return []*models.ToolConfigGroup{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(d.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, err
	}

	// 合并查询参数
	allArgs := append(args, paginationArgs...)

	// 执行分页查询
	err = d.db.Query(ctx, &configGroups, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, err
	}

	return configGroups, total, nil
}

// GetByParentId 根据父分组ID获取子分组列表
func (d *ConfigGroupDao) GetByParentId(ctx context.Context, tenantId, parentGroupId string) ([]*models.ToolConfigGroup, error) {
	var configGroups []*models.ToolConfigGroup
	query := "SELECT * FROM " + (&models.ToolConfigGroup{}).TableName() + 
		" WHERE tenantId = ? AND parentGroupId = ? ORDER BY sortOrder ASC, addTime DESC"
	args := []interface{}{tenantId, parentGroupId}

	err := d.db.Query(ctx, &configGroups, query, args, true)
	if err != nil {
		return nil, err
	}

	return configGroups, nil
} 