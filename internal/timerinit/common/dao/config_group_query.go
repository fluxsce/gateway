package dao

import (
	"context"
	"fmt"
	"gateway/internal/types/tooltypes"
	"gateway/pkg/database"
	"strings"
	"time"
)

// ToolConfigGroupQuery 定义工具配置分组查询条件结构
type ToolConfigGroupQuery struct {
	// 基础查询条件
	ConfigGroupId *string `json:"configGroupId,omitempty"`
	TenantId      *string `json:"tenantId,omitempty"`
	GroupName     *string `json:"groupName,omitempty"`
	ParentGroupId *string `json:"parentGroupId,omitempty"`
	GroupType     *string `json:"groupType,omitempty"`

	// 层级查询
	GroupLevel    *int `json:"groupLevel,omitempty"`
	MinGroupLevel *int `json:"minGroupLevel,omitempty"`
	MaxGroupLevel *int `json:"maxGroupLevel,omitempty"`

	// 排序查询
	SortOrder    *int `json:"sortOrder,omitempty"`
	MinSortOrder *int `json:"minSortOrder,omitempty"`
	MaxSortOrder *int `json:"maxSortOrder,omitempty"`

	// 外观查询
	GroupIcon  *string `json:"groupIcon,omitempty"`
	GroupColor *string `json:"groupColor,omitempty"`

	// 权限查询
	AccessLevel     *string  `json:"accessLevel,omitempty"`
	AccessLevelList []string `json:"accessLevelList,omitempty"`

	// 状态查询
	ActiveFlag *string `json:"activeFlag,omitempty"`

	// 创建和修改时间查询
	AddTimeFrom  *time.Time `json:"addTimeFrom,omitempty"`
	AddTimeTo    *time.Time `json:"addTimeTo,omitempty"`
	EditTimeFrom *time.Time `json:"editTimeFrom,omitempty"`
	EditTimeTo   *time.Time `json:"editTimeTo,omitempty"`
	AddWho       *string    `json:"addWho,omitempty"`
	EditWho      *string    `json:"editWho,omitempty"`

	// 模糊查询
	GroupNameLike        *string `json:"groupNameLike,omitempty"`
	GroupDescriptionLike *string `json:"groupDescriptionLike,omitempty"`
	GroupPathLike        *string `json:"groupPathLike,omitempty"`
	NoteTextLike         *string `json:"noteTextLike,omitempty"`

	// 特殊查询条件
	OnlyRootGroups *bool    `json:"onlyRootGroups,omitempty"` // 只返回根分组
	OnlyLeafGroups *bool    `json:"onlyLeafGroups,omitempty"` // 只返回叶子分组
	OnlyPublic     *bool    `json:"onlyPublic,omitempty"`     // 只返回公开分组
	OnlyPrivate    *bool    `json:"onlyPrivate,omitempty"`    // 只返回私有分组
	GroupTypeList  []string `json:"groupTypeList,omitempty"`  // 分组类型列表

	// 排序和分页
	OrderBy        string `json:"orderBy,omitempty"`        // 排序字段
	OrderDirection string `json:"orderDirection,omitempty"` // ASC/DESC
	PageNum        int    `json:"pageNum,omitempty"`        // 页码，从1开始
	PageSize       int    `json:"pageSize,omitempty"`       // 每页大小
	Offset         int    `json:"offset,omitempty"`         // 偏移量
	Limit          int    `json:"limit,omitempty"`          // 限制数量
}

// ToolConfigGroupQueryResult 查询结果结构
type ToolConfigGroupQueryResult struct {
	Groups     []tooltypes.ToolConfigGroup `json:"groups"`
	Total      int64                       `json:"total"`
	PageNum    int                         `json:"pageNum"`
	PageSize   int                         `json:"pageSize"`
	TotalPages int                         `json:"totalPages"`
}

// ToolConfigGroupDAO 工具配置分组数据访问对象
type ToolConfigGroupDAO struct {
	db database.Database
}

// NewToolConfigGroupDAO 创建工具配置分组DAO实例
func NewToolConfigGroupDAO(db database.Database) *ToolConfigGroupDAO {
	return &ToolConfigGroupDAO{db: db}
}

// BuildWhere 构建WHERE条件语句
func (q *ToolConfigGroupQuery) BuildWhere() (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// 基础条件
	if q.ConfigGroupId != nil {
		conditions = append(conditions, "configGroupId = ?")
		args = append(args, *q.ConfigGroupId)
	}

	if q.TenantId != nil {
		conditions = append(conditions, "tenantId = ?")
		args = append(args, *q.TenantId)
	}

	if q.GroupName != nil {
		conditions = append(conditions, "groupName = ?")
		args = append(args, *q.GroupName)
	}

	if q.ParentGroupId != nil {
		conditions = append(conditions, "parentGroupId = ?")
		args = append(args, *q.ParentGroupId)
	}

	if q.GroupType != nil {
		conditions = append(conditions, "groupType = ?")
		args = append(args, *q.GroupType)
	}

	// 层级条件
	if q.GroupLevel != nil {
		conditions = append(conditions, "groupLevel = ?")
		args = append(args, *q.GroupLevel)
	}

	if q.MinGroupLevel != nil {
		conditions = append(conditions, "groupLevel >= ?")
		args = append(args, *q.MinGroupLevel)
	}

	if q.MaxGroupLevel != nil {
		conditions = append(conditions, "groupLevel <= ?")
		args = append(args, *q.MaxGroupLevel)
	}

	// 排序条件
	if q.SortOrder != nil {
		conditions = append(conditions, "sortOrder = ?")
		args = append(args, *q.SortOrder)
	}

	if q.MinSortOrder != nil {
		conditions = append(conditions, "sortOrder >= ?")
		args = append(args, *q.MinSortOrder)
	}

	if q.MaxSortOrder != nil {
		conditions = append(conditions, "sortOrder <= ?")
		args = append(args, *q.MaxSortOrder)
	}

	// 外观条件
	if q.GroupIcon != nil {
		conditions = append(conditions, "groupIcon = ?")
		args = append(args, *q.GroupIcon)
	}

	if q.GroupColor != nil {
		conditions = append(conditions, "groupColor = ?")
		args = append(args, *q.GroupColor)
	}

	// 权限条件
	if q.AccessLevel != nil {
		conditions = append(conditions, "accessLevel = ?")
		args = append(args, *q.AccessLevel)
	}

	if len(q.AccessLevelList) > 0 {
		placeholders := make([]string, len(q.AccessLevelList))
		for i, level := range q.AccessLevelList {
			placeholders[i] = "?"
			args = append(args, level)
		}
		conditions = append(conditions, fmt.Sprintf("accessLevel IN (%s)", strings.Join(placeholders, ",")))
	}

	// 状态条件
	if q.ActiveFlag != nil {
		conditions = append(conditions, "activeFlag = ?")
		args = append(args, *q.ActiveFlag)
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
	if q.GroupNameLike != nil {
		conditions = append(conditions, "groupName LIKE ?")
		args = append(args, "%"+*q.GroupNameLike+"%")
	}

	if q.GroupDescriptionLike != nil {
		conditions = append(conditions, "groupDescription LIKE ?")
		args = append(args, "%"+*q.GroupDescriptionLike+"%")
	}

	if q.GroupPathLike != nil {
		conditions = append(conditions, "groupPath LIKE ?")
		args = append(args, "%"+*q.GroupPathLike+"%")
	}

	if q.NoteTextLike != nil {
		conditions = append(conditions, "noteText LIKE ?")
		args = append(args, "%"+*q.NoteTextLike+"%")
	}

	// 特殊条件
	if q.OnlyRootGroups != nil && *q.OnlyRootGroups {
		conditions = append(conditions, "(parentGroupId IS NULL OR parentGroupId = '')")
	}

	if q.OnlyLeafGroups != nil && *q.OnlyLeafGroups {
		// 叶子节点：没有子分组的分组
		conditions = append(conditions, fmt.Sprintf(`
			NOT EXISTS (
				SELECT 1 FROM %s child 
				WHERE child.parentGroupId = %s.configGroupId 
				AND child.activeFlag = 'Y'
			)
		`, (&tooltypes.ToolConfigGroup{}).TableName(), (&tooltypes.ToolConfigGroup{}).TableName()))
	}

	if q.OnlyPublic != nil && *q.OnlyPublic {
		conditions = append(conditions, "accessLevel = ?")
		args = append(args, tooltypes.AccessLevelPublic)
	}

	if q.OnlyPrivate != nil && *q.OnlyPrivate {
		conditions = append(conditions, "accessLevel = ?")
		args = append(args, tooltypes.AccessLevelPrivate)
	}

	if len(q.GroupTypeList) > 0 {
		placeholders := make([]string, len(q.GroupTypeList))
		for i, groupType := range q.GroupTypeList {
			placeholders[i] = "?"
			args = append(args, groupType)
		}
		conditions = append(conditions, fmt.Sprintf("groupType IN (%s)", strings.Join(placeholders, ",")))
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	return where, args
}

// BuildOrderBy 构建ORDER BY语句
func (q *ToolConfigGroupQuery) BuildOrderBy() string {
	if q.OrderBy == "" {
		return "ORDER BY groupLevel ASC, sortOrder ASC, addTime DESC"
	}

	direction := "ASC"
	if strings.ToUpper(q.OrderDirection) == "DESC" {
		direction = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", q.OrderBy, direction)
}

// BuildPagination 构建分页语句
func (q *ToolConfigGroupQuery) BuildPagination() string {
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

// QueryGroups 查询工具配置分组列表
func (dao *ToolConfigGroupDAO) QueryGroups(ctx context.Context, query *ToolConfigGroupQuery) (*ToolConfigGroupQueryResult, error) {
	tableName := (&tooltypes.ToolConfigGroup{}).TableName()

	// 构建查询条件
	where, args := query.BuildWhere()
	orderBy := query.BuildOrderBy()
	pagination := query.BuildPagination()

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", tableName, where)
	var total int64
	err := dao.db.QueryOne(ctx, &total, countSQL, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询配置分组总数失败: %w", err)
	}

	// 查询数据
	selectSQL := fmt.Sprintf("SELECT * FROM %s %s %s %s", tableName, where, orderBy, pagination)
	var groups []tooltypes.ToolConfigGroup
	err = dao.db.Query(ctx, &groups, selectSQL, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询配置分组列表失败: %w", err)
	}

	// 计算分页信息
	result := &ToolConfigGroupQueryResult{
		Groups: groups,
		Total:  total,
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

// GetGroupById 根据ID获取配置分组
func (dao *ToolConfigGroupDAO) GetGroupById(ctx context.Context, tenantId, configGroupId string) (*tooltypes.ToolConfigGroup, error) {
	tableName := (&tooltypes.ToolConfigGroup{}).TableName()
	sql := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND configGroupId = ?", tableName)

	var group tooltypes.ToolConfigGroup
	err := dao.db.QueryOne(ctx, &group, sql, []interface{}{tenantId, configGroupId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询配置分组失败: %w", err)
	}

	return &group, nil
}

// CreateGroup 创建配置分组
func (dao *ToolConfigGroupDAO) CreateGroup(ctx context.Context, group *tooltypes.ToolConfigGroup) error {
	tableName := (&tooltypes.ToolConfigGroup{}).TableName()
	_, err := dao.db.Insert(ctx, tableName, group, true)
	if err != nil {
		return fmt.Errorf("创建配置分组失败: %w", err)
	}
	return nil
}

// UpdateGroup 更新配置分组
func (dao *ToolConfigGroupDAO) UpdateGroup(ctx context.Context, group *tooltypes.ToolConfigGroup) error {
	tableName := (&tooltypes.ToolConfigGroup{}).TableName()
	where := "tenantId = ? AND configGroupId = ?"
	args := []interface{}{group.TenantId, group.ConfigGroupId}

	_, err := dao.db.Update(ctx, tableName, group, where, args, true)
	if err != nil {
		return fmt.Errorf("更新配置分组失败: %w", err)
	}
	return nil
}

// DeleteGroup 删除配置分组（软删除）
func (dao *ToolConfigGroupDAO) DeleteGroup(ctx context.Context, tenantId, configGroupId string) error {
	tableName := (&tooltypes.ToolConfigGroup{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = 'N' WHERE tenantId = ? AND configGroupId = ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, configGroupId}, true)
	if err != nil {
		return fmt.Errorf("删除配置分组失败: %w", err)
	}
	return nil
}

// GetRootGroups 获取根分组列表
func (dao *ToolConfigGroupDAO) GetRootGroups(ctx context.Context, tenantId string) ([]tooltypes.ToolConfigGroup, error) {
	onlyRootGroups := true
	activeFlag := tooltypes.ActiveFlagYes

	query := &ToolConfigGroupQuery{
		TenantId:       &tenantId,
		ActiveFlag:     &activeFlag,
		OnlyRootGroups: &onlyRootGroups,
		OrderBy:        "sortOrder",
		OrderDirection: "ASC",
	}

	result, err := dao.QueryGroups(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Groups, nil
}

// GetChildGroups 获取子分组列表
func (dao *ToolConfigGroupDAO) GetChildGroups(ctx context.Context, tenantId, parentGroupId string) ([]tooltypes.ToolConfigGroup, error) {
	activeFlag := tooltypes.ActiveFlagYes

	query := &ToolConfigGroupQuery{
		TenantId:       &tenantId,
		ParentGroupId:  &parentGroupId,
		ActiveFlag:     &activeFlag,
		OrderBy:        "sortOrder",
		OrderDirection: "ASC",
	}

	result, err := dao.QueryGroups(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Groups, nil
}

// GetGroupsByType 根据分组类型获取分组列表
func (dao *ToolConfigGroupDAO) GetGroupsByType(ctx context.Context, tenantId, groupType string) ([]tooltypes.ToolConfigGroup, error) {
	activeFlag := tooltypes.ActiveFlagYes

	query := &ToolConfigGroupQuery{
		TenantId:       &tenantId,
		GroupType:      &groupType,
		ActiveFlag:     &activeFlag,
		OrderBy:        "groupLevel",
		OrderDirection: "ASC",
	}

	result, err := dao.QueryGroups(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Groups, nil
}

// GetPublicGroups 获取公开分组列表
func (dao *ToolConfigGroupDAO) GetPublicGroups(ctx context.Context, tenantId string) ([]tooltypes.ToolConfigGroup, error) {
	onlyPublic := true
	activeFlag := tooltypes.ActiveFlagYes

	query := &ToolConfigGroupQuery{
		TenantId:       &tenantId,
		ActiveFlag:     &activeFlag,
		OnlyPublic:     &onlyPublic,
		OrderBy:        "groupLevel",
		OrderDirection: "ASC",
	}

	result, err := dao.QueryGroups(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Groups, nil
}

// GetGroupTree 获取分组树形结构
func (dao *ToolConfigGroupDAO) GetGroupTree(ctx context.Context, tenantId string) ([]tooltypes.ToolConfigGroup, error) {
	activeFlag := tooltypes.ActiveFlagYes

	query := &ToolConfigGroupQuery{
		TenantId:       &tenantId,
		ActiveFlag:     &activeFlag,
		OrderBy:        "groupLevel",
		OrderDirection: "ASC",
	}

	result, err := dao.QueryGroups(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Groups, nil
}
