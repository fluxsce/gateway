package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0061/models"
)

// StaticServerDAO 静态服务器数据访问对象
type StaticServerDAO struct {
	db database.Database
}

// NewStaticServerDAO 创建静态服务器DAO实例
func NewStaticServerDAO(db database.Database) *StaticServerDAO {
	return &StaticServerDAO{db: db}
}

// QueryStaticServers 查询静态服务器列表
func (dao *StaticServerDAO) QueryStaticServers(req *models.StaticServerQueryRequest) ([]*types.TunnelStaticServer, int, error) {
	ctx := context.Background()

	// 构建查询条件
	whereClause := "WHERE 1=1"
	var params []interface{}

	if req.ActiveFlag != "" {
		whereClause += " AND activeFlag = ?"
		params = append(params, req.ActiveFlag)
	} else {
		whereClause += " AND activeFlag = 'Y'"
	}

	if req.ServerStatus != "" {
		whereClause += " AND serverStatus = ?"
		params = append(params, req.ServerStatus)
	}

	if req.ServerType != "" {
		whereClause += " AND serverType = ?"
		params = append(params, req.ServerType)
	}

	if req.ServerName != "" {
		whereClause += " AND serverName LIKE ?"
		params = append(params, "%"+req.ServerName+"%")
	}

	if req.ServerDescription != "" {
		whereClause += " AND serverDescription LIKE ?"
		params = append(params, "%"+req.ServerDescription+"%")
	}

	if req.ListenAddress != "" {
		whereClause += " AND listenAddress = ?"
		params = append(params, req.ListenAddress)
	}

	if req.ListenPort > 0 {
		whereClause += " AND listenPort = ?"
		params = append(params, req.ListenPort)
	}

	// 构建基础查询
	baseQuery := fmt.Sprintf(`
		SELECT tunnelStaticServerId, tenantId, serverName, serverDescription,
			listenAddress, listenPort, serverType, maxConnections,
			connectionTimeout, readTimeout, writeTimeout,
			tlsEnable, tlsCertFile, tlsKeyFile, tlsCaFile,
			logLevel, logFile, serverStatus, startTime, stopTime,
			currentConnectionCount, totalConnectionCount, totalBytesReceived, totalBytesSent,
			healthCheckType, healthCheckUrl, healthCheckInterval, healthCheckTimeout, healthCheckMaxFailures,
			loadBalanceType, serverConfig,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_STATIC_SERVER %s
		ORDER BY editTime DESC
	`, whereClause)

	// 查询总数
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务器总数失败")
	}

	if countResult.Count == 0 {
		return []*types.TunnelStaticServer{}, 0, nil
	}

	// 分页查询
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	allArgs := append(params, paginationArgs...)

	var servers []*types.TunnelStaticServer
	err = dao.db.Query(ctx, &servers, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务器数据失败")
	}

	// 查询每个服务器的节点数量
	for _, server := range servers {
		nodeCount, _ := dao.getNodeCount(ctx, server.TunnelStaticServerId)
		server.NodeCount = nodeCount
	}

	return servers, countResult.Count, nil
}

// getNodeCount 获取服务器的节点数量
func (dao *StaticServerDAO) getNodeCount(ctx context.Context, serverId string) (int, error) {
	query := `SELECT COUNT(*) FROM HUB_TUNNEL_STATIC_NODE WHERE tunnelStaticServerId = ? AND activeFlag = 'Y'`
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{serverId}, true)
	if err != nil {
		return 0, err
	}
	return result.Count, nil
}

// GetStaticServer 获取静态服务器详情
func (dao *StaticServerDAO) GetStaticServer(ctx context.Context, serverId, tenantId string) (*types.TunnelStaticServer, error) {
	query := `
		SELECT tunnelStaticServerId, tenantId, serverName, serverDescription,
			listenAddress, listenPort, serverType, maxConnections,
			connectionTimeout, readTimeout, writeTimeout,
			tlsEnable, tlsCertFile, tlsKeyFile, tlsCaFile,
			logLevel, logFile, serverStatus, startTime, stopTime,
			currentConnectionCount, totalConnectionCount, totalBytesReceived, totalBytesSent,
			healthCheckType, healthCheckUrl, healthCheckInterval, healthCheckTimeout, healthCheckMaxFailures,
			loadBalanceType, serverConfig,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_STATIC_SERVER
		WHERE tunnelStaticServerId = ?
	`

	server := &types.TunnelStaticServer{}
	err := dao.db.QueryOne(ctx, server, query, []interface{}{serverId}, true)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "获取服务器信息失败")
	}

	// 获取节点数量
	server.NodeCount, _ = dao.getNodeCount(ctx, serverId)

	return server, nil
}

// CreateStaticServer 创建静态服务器
func (dao *StaticServerDAO) CreateStaticServer(ctx context.Context, server *types.TunnelStaticServer) error {
	// 设置默认值
	if server.ActiveFlag == "" {
		server.ActiveFlag = "Y"
	}
	if server.ServerStatus == "" {
		server.ServerStatus = "stopped"
	}
	if server.ServerType == "" {
		server.ServerType = "tcp"
	}
	if server.ListenAddress == "" {
		server.ListenAddress = "0.0.0.0"
	}
	if server.MaxConnections == 0 {
		server.MaxConnections = 1000
	}
	if server.ConnectionTimeout == 0 {
		server.ConnectionTimeout = 30
	}
	if server.ReadTimeout == 0 {
		server.ReadTimeout = 60
	}
	if server.WriteTimeout == 0 {
		server.WriteTimeout = 60
	}
	if server.TlsEnable == "" {
		server.TlsEnable = "N"
	}
	if server.LogLevel == "" {
		server.LogLevel = "info"
	}

	// 检查端口冲突
	conflict, err := dao.CheckPortConflict(ctx, server.ListenAddress, server.ListenPort, server.ServerType, "")
	if err != nil {
		return huberrors.WrapError(err, "检查端口冲突失败")
	}
	if conflict {
		errMsg := fmt.Sprintf("端口已被占用: %s:%d (%s)", server.ListenAddress, server.ListenPort, server.ServerType)
		return huberrors.NewError(errMsg)
	}

	// 检查服务器名称唯一性
	exists, err := dao.CheckServerNameExists(ctx, server.ServerName, "")
	if err != nil {
		return huberrors.WrapError(err, "检查服务器名称存在性失败")
	}
	if exists {
		return huberrors.NewError("服务器名称已存在: " + server.ServerName)
	}

	// 设置时间
	now := time.Now()
	server.AddTime = now
	server.EditTime = now

	// 插入数据库
	_, err = dao.db.Insert(ctx, "HUB_TUNNEL_STATIC_SERVER", server, true)
	if err != nil {
		return huberrors.WrapError(err, "创建服务器失败")
	}

	logger.Info("创建静态服务器成功", "serverId", server.TunnelStaticServerId, "serverName", server.ServerName)
	return nil
}

// UpdateStaticServer 更新静态服务器
func (dao *StaticServerDAO) UpdateStaticServer(ctx context.Context, server *types.TunnelStaticServer) error {
	// 检查服务器是否存在
	existingServer, err := dao.GetStaticServer(ctx, server.TunnelStaticServerId, server.TenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取服务器信息失败")
	}
	if existingServer == nil {
		return huberrors.NewError("服务器不存在")
	}

	// 检查端口冲突（排除自己）
	conflict, err := dao.CheckPortConflict(ctx, server.ListenAddress, server.ListenPort, server.ServerType, server.TunnelStaticServerId)
	if err != nil {
		return huberrors.WrapError(err, "检查端口冲突失败")
	}
	if conflict {
		errMsg := fmt.Sprintf("端口已被占用: %s:%d (%s)", server.ListenAddress, server.ListenPort, server.ServerType)
		return huberrors.NewError(errMsg)
	}

	// 检查服务器名称唯一性（排除自己）
	if server.ServerName != existingServer.ServerName {
		exists, err := dao.CheckServerNameExists(ctx, server.ServerName, server.TunnelStaticServerId)
		if err != nil {
			return huberrors.WrapError(err, "检查服务器名称存在性失败")
		}
		if exists {
			return huberrors.NewError("服务器名称已存在: " + server.ServerName)
		}
	}

	// 更新版本号和时间
	server.CurrentVersion = existingServer.CurrentVersion + 1
	server.EditTime = time.Now()

	// 更新数据库
	whereClause := "tunnelStaticServerId = ?"
	args := []interface{}{server.TunnelStaticServerId}

	_, err = dao.db.Update(ctx, "HUB_TUNNEL_STATIC_SERVER", server, whereClause, args, true, true)
	if err != nil {
		return huberrors.WrapError(err, "更新服务器失败")
	}

	logger.Info("更新静态服务器成功", "serverId", server.TunnelStaticServerId)
	return nil
}

// DeleteStaticServer 删除静态服务器
func (dao *StaticServerDAO) DeleteStaticServer(ctx context.Context, serverId, tenantId string) error {
	// 检查服务器是否存在
	server, err := dao.GetStaticServer(ctx, serverId, tenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取服务器信息失败")
	}
	if server == nil {
		return huberrors.NewError("服务器不存在")
	}

	// 检查是否有关联的节点
	nodeCount, err := dao.getNodeCount(ctx, serverId)
	if err != nil {
		return huberrors.WrapError(err, "检查关联节点失败")
	}
	if nodeCount > 0 {
		return huberrors.NewError(fmt.Sprintf("服务器下还有 %d 个节点，请先删除节点", nodeCount))
	}

	// 物理删除
	deleteSQL := `DELETE FROM HUB_TUNNEL_STATIC_SERVER WHERE tunnelStaticServerId = ?`
	_, err = dao.db.Exec(ctx, deleteSQL, []interface{}{serverId}, false)
	if err != nil {
		return huberrors.WrapError(err, "删除服务器失败")
	}

	logger.Info("删除静态服务器成功", "serverId", serverId)
	return nil
}

// CheckPortConflict 检查端口冲突
func (dao *StaticServerDAO) CheckPortConflict(ctx context.Context, listenAddress string, listenPort int, serverType, excludeId string) (bool, error) {
	whereClause := "listenAddress = ? AND listenPort = ? AND serverType = ? AND activeFlag = 'Y'"
	args := []interface{}{listenAddress, listenPort, serverType}

	if excludeId != "" {
		whereClause += " AND tunnelStaticServerId != ?"
		args = append(args, excludeId)
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM HUB_TUNNEL_STATIC_SERVER WHERE %s", whereClause)

	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, query, args, true)
	if err != nil {
		return false, huberrors.WrapError(err, "检查端口冲突失败")
	}

	return result.Count > 0, nil
}

// CheckServerNameExists 检查服务器名称是否存在
func (dao *StaticServerDAO) CheckServerNameExists(ctx context.Context, serverName, excludeId string) (bool, error) {
	whereClause := "serverName = ? AND activeFlag = 'Y'"
	args := []interface{}{serverName}

	if excludeId != "" {
		whereClause += " AND tunnelStaticServerId != ?"
		args = append(args, excludeId)
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM HUB_TUNNEL_STATIC_SERVER WHERE %s", whereClause)

	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, query, args, true)
	if err != nil {
		return false, huberrors.WrapError(err, "检查服务器名称存在性失败")
	}

	return result.Count > 0, nil
}

// GetStaticServerStats 获取静态服务器统计信息
func (dao *StaticServerDAO) GetStaticServerStats(ctx context.Context) (*models.StaticServerStats, error) {
	// 查询总服务器数
	totalQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_STATIC_SERVER WHERE activeFlag = 'Y'`
	var totalResult struct {
		Count int `db:"COUNT(*)"`
	}
	dao.db.QueryOne(ctx, &totalResult, totalQuery, nil, true)

	// 查询运行中服务器数
	runningQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_STATIC_SERVER WHERE activeFlag = 'Y' AND serverStatus = 'running'`
	var runningResult struct {
		Count int `db:"COUNT(*)"`
	}
	dao.db.QueryOne(ctx, &runningResult, runningQuery, nil, true)

	// 查询流量统计
	statsQuery := `
		SELECT COALESCE(SUM(totalConnectionCount), 0) as totalConnections,
		       COALESCE(SUM(totalBytesReceived), 0) as totalBytesReceived,
		       COALESCE(SUM(totalBytesSent), 0) as totalBytesSent
		FROM HUB_TUNNEL_STATIC_SERVER
		WHERE activeFlag = 'Y'
	`
	var statsResult struct {
		TotalConnections   int64 `db:"totalConnections"`
		TotalBytesReceived int64 `db:"totalBytesReceived"`
		TotalBytesSent     int64 `db:"totalBytesSent"`
	}
	dao.db.QueryOne(ctx, &statsResult, statsQuery, nil, true)

	return &models.StaticServerStats{
		TotalServers:       totalResult.Count,
		RunningServers:     runningResult.Count,
		StoppedServers:     totalResult.Count - runningResult.Count,
		TotalConnections:   statsResult.TotalConnections,
		TotalBytesReceived: statsResult.TotalBytesReceived,
		TotalBytesSent:     statsResult.TotalBytesSent,
	}, nil
}
