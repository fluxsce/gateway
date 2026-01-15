package dao

import (
	"context"
	"errors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0007/models"
)

// ServerInfoDAO 系统节点信息数据访问对象
type ServerInfoDAO struct {
	db database.Database
}

// NewServerInfoDAO 创建系统节点信息DAO
func NewServerInfoDAO(db database.Database) *ServerInfoDAO {
	return &ServerInfoDAO{
		db: db,
	}
}

// GetServerInfoById 根据节点ID获取系统节点信息
func (dao *ServerInfoDAO) GetServerInfoById(ctx context.Context, metricServerId, tenantId string) (*models.ServerInfo, error) {
	if metricServerId == "" || tenantId == "" {
		return nil, errors.New("metricServerId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_METRIC_SERVER_INFO 
		WHERE metricServerId = ? AND tenantId = ?
	`

	var serverInfo models.ServerInfo
	err := dao.db.QueryOne(ctx, &serverInfo, query, []interface{}{metricServerId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询服务器信息失败")
	}

	return &serverInfo, nil
}

// ListServerInfos 获取系统节点信息列表（支持条件查询）
func (dao *ServerInfoDAO) ListServerInfos(ctx context.Context, tenantId string, query *models.ServerInfoQuery, page, pageSize int) ([]*models.ServerInfo, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建查询条件
	whereClause := "WHERE tenantId = ?"
	var params []interface{}
	params = append(params, tenantId)

	// 构建查询条件，只有当字段不为空时才添加对应条件
	if query != nil {
		if !empty.IsEmpty(query.Hostname) {
			whereClause += " AND hostname LIKE ?"
			params = append(params, "%"+query.Hostname+"%")
		}
		if !empty.IsEmpty(query.OsType) {
			whereClause += " AND osType = ?"
			params = append(params, query.OsType)
		}
		if !empty.IsEmpty(query.ServerType) {
			whereClause += " AND serverType = ?"
			params = append(params, query.ServerType)
		}
		if !empty.IsEmpty(query.IpAddress) {
			whereClause += " AND ipAddress LIKE ?"
			params = append(params, "%"+query.IpAddress+"%")
		}
		if !empty.IsEmpty(query.ServerLocation) {
			whereClause += " AND serverLocation LIKE ?"
			params = append(params, "%"+query.ServerLocation+"%")
		}
		// 只有当 activeFlag 不为空时才添加查询条件，否则不处理
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, query.ActiveFlag)
		}
	}

	// 基础查询语句
	baseQuery := `
		SELECT * FROM HUB_METRIC_SERVER_INFO
	` + whereClause + `
		ORDER BY lastUpdateTime DESC
	`

	// 构建计数查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建计数查询失败")
	}

	// 执行计数查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务器信息总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.ServerInfo{}, 0, nil
	}

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数：基础查询参数 + 分页参数
	queryArgs := params
	queryArgs = append(queryArgs, paginationArgs...)

	// 执行分页查询
	var serverInfos []*models.ServerInfo
	err = dao.db.Query(ctx, &serverInfos, paginatedQuery, queryArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务器信息列表失败")
	}

	return serverInfos, total, nil
}
