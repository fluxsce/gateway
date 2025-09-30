package dao

import (
	"context"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0060/models"
	"strings"
	"time"
)

// TunnelServerDAO 隧道服务器数据访问对象
type TunnelServerDAO struct {
	db database.Database
}

// NewTunnelServerDAO 创建隧道服务器DAO实例
func NewTunnelServerDAO(db database.Database) *TunnelServerDAO {
	return &TunnelServerDAO{db: db}
}

// QueryTunnelServers 查询隧道服务器列表
func (dao *TunnelServerDAO) QueryTunnelServers(req *models.TunnelServerQueryRequest) ([]*models.TunnelServer, int, error) {
	ctx := context.Background()

	// 构建查询条件
	whereClause := "WHERE 1=1"
	var params []interface{}

	if req.ServerName != "" {
		whereClause += " AND serverName LIKE ?"
		params = append(params, "%"+req.ServerName+"%")
	}

	if req.ServerAddress != "" {
		whereClause += " AND controlAddress LIKE ?"
		params = append(params, "%"+req.ServerAddress+"%")
	}

	if req.ServerStatus != "" {
		whereClause += " AND serverStatus = ?"
		params = append(params, req.ServerStatus)
	}

	if req.ActiveFlag != "" {
		whereClause += " AND activeFlag = ?"
		params = append(params, req.ActiveFlag)
	}

	if req.Keyword != "" {
		whereClause += " AND (serverName LIKE ? OR controlAddress LIKE ? OR serverDescription LIKE ?)"
		keyword := "%" + req.Keyword + "%"
		params = append(params, keyword, keyword, keyword)
	}

	// 构建基础查询语句
	baseQuery := fmt.Sprintf(`
		SELECT tunnelServerId, tenantId, serverName, serverDescription,
			controlAddress, controlPort, dashboardPort, vhostHttpPort, vhostHttpsPort,
			maxClients, tokenAuth, authToken, tlsEnable, tlsCertFile, tlsKeyFile,
			heartbeatInterval, heartbeatTimeout, logLevel, maxPortsPerClient, allowPorts,
			serverStatus, startTime, configVersion,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_SERVER %s
		ORDER BY editTime DESC
	`, whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, params, true)
	if err != nil {
		logger.Error("查询隧道服务器总数失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "查询隧道服务器总数失败")
	}

	// 如果没有记录，直接返回空列表
	if countResult.Count == 0 {
		return []*models.TunnelServer{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	// 执行分页查询
	var servers []*models.TunnelServer
	err = dao.db.Query(ctx, &servers, paginatedQuery, allArgs, true)
	if err != nil {
		logger.Error("查询隧道服务器数据失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "查询隧道服务器数据失败")
	}

	return servers, countResult.Count, nil
}

// GetTunnelServer 获取隧道服务器详情
func (dao *TunnelServerDAO) GetTunnelServer(tunnelServerId string) (*models.TunnelServer, error) {
	ctx := context.Background()

	// 构建查询语句
	query := `
		SELECT tunnelServerId, tenantId, serverName, serverDescription,
			controlAddress, controlPort, dashboardPort, vhostHttpPort, vhostHttpsPort,
			maxClients, tokenAuth, authToken, tlsEnable, tlsCertFile, tlsKeyFile,
			heartbeatInterval, heartbeatTimeout, logLevel, maxPortsPerClient, allowPorts,
			serverStatus, startTime, configVersion,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_SERVER
		WHERE tunnelServerId = ?
	`

	server := &models.TunnelServer{}
	err := dao.db.QueryOne(ctx, server, query, []interface{}{tunnelServerId}, true)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "not found") {
			return nil, huberrors.WrapError(err, "隧道服务器不存在")
		}
		return nil, huberrors.WrapError(err, "获取隧道服务器信息失败")
	}

	return server, nil
}

// CreateTunnelServer 创建隧道服务器
func (dao *TunnelServerDAO) CreateTunnelServer(server *models.TunnelServer) (*models.TunnelServer, error) {
	ctx := context.Background()

	// 设置默认值
	if server.ActiveFlag == "" {
		server.ActiveFlag = "Y"
	}
	if server.ServerStatus == "" {
		server.ServerStatus = "stopped"
	}
	if server.TokenAuth == "" {
		server.TokenAuth = "Y"
	}
	if server.TlsEnable == "" {
		server.TlsEnable = "N"
	}
	if server.HeartbeatInterval == 0 {
		server.HeartbeatInterval = 30
	}
	if server.HeartbeatTimeout == 0 {
		server.HeartbeatTimeout = 90
	}
	if server.LogLevel == "" {
		server.LogLevel = "info"
	}
	if server.ControlAddress == "" {
		server.ControlAddress = "0.0.0.0"
	}

	// 检查服务器名称是否重复
	exists, err := dao.CheckServerNameExists(server.ServerName, "")
	if err != nil {
		return nil, huberrors.WrapError(err, "检查服务器名称存在性失败")
	}
	if exists {
		return nil, huberrors.WrapError(fmt.Errorf("服务器名称重复"), "服务器名称已存在")
	}

	// 设置创建时间和修改时间
	now := time.Now()
	server.AddTime = now
	server.EditTime = now

	// 插入数据库
	_, err = dao.db.Insert(ctx, "HUB_TUNNEL_SERVER", server, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "创建隧道服务器失败")
	}

	logger.Info("创建隧道服务器成功", "tunnelServerId", server.TunnelServerId, "serverName", server.ServerName)

	// 返回创建后的服务器信息
	createdServer, err := dao.GetTunnelServer(server.TunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取创建后的服务器信息失败")
	}

	return createdServer, nil
}

// UpdateTunnelServer 更新隧道服务器
func (dao *TunnelServerDAO) UpdateTunnelServer(server *models.TunnelServer) (*models.TunnelServer, error) {
	ctx := context.Background()

	// 检查服务器是否存在
	existingServer, err := dao.GetTunnelServer(server.TunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取隧道服务器信息失败")
	}

	// 检查服务器名称是否重复
	if server.ServerName != existingServer.ServerName {
		exists, err := dao.CheckServerNameExists(server.ServerName, server.TunnelServerId)
		if err != nil {
			return nil, huberrors.WrapError(err, "检查服务器名称存在性失败")
		}
		if exists {
			return nil, huberrors.WrapError(fmt.Errorf("服务器名称重复"), "服务器名称已存在")
		}
	}

	// 更新版本号和修改时间
	server.CurrentVersion = existingServer.CurrentVersion + 1
	server.EditTime = time.Now()

	// 更新数据库
	whereClause := "tunnelServerId = ?"
	args := []interface{}{server.TunnelServerId}

	_, err = dao.db.Update(ctx, "HUB_TUNNEL_SERVER", server, whereClause, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "更新隧道服务器失败")
	}

	logger.Info("更新隧道服务器成功", "tunnelServerId", server.TunnelServerId, "serverName", server.ServerName)

	// 返回更新后的服务器信息
	updatedServer, err := dao.GetTunnelServer(server.TunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取更新后的服务器信息失败")
	}

	return updatedServer, nil
}

// DeleteTunnelServer 删除隧道服务器（逻辑删除）
func (dao *TunnelServerDAO) DeleteTunnelServer(tunnelServerId, editWho string) (*models.TunnelServer, error) {
	ctx := context.Background()

	// 获取服务器信息
	server, err := dao.GetTunnelServer(tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取隧道服务器信息失败")
	}

	// 逻辑删除（设置activeFlag为'N'）
	server.ActiveFlag = "N"
	server.EditTime = time.Now()
	server.EditWho = editWho
	server.CurrentVersion++

	// 更新数据库
	whereClause := "tunnelServerId = ?"
	args := []interface{}{tunnelServerId}

	_, err = dao.db.Update(ctx, "HUB_TUNNEL_SERVER", server, whereClause, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "删除隧道服务器失败")
	}

	logger.Info("删除隧道服务器成功", "tunnelServerId", tunnelServerId, "editWho", editWho)

	// 返回更新后的服务器信息
	updatedServer, err := dao.GetTunnelServer(tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取更新后的服务器信息失败")
	}

	return updatedServer, nil
}

// UpdateTunnelServerStatus 更新隧道服务器状态
func (dao *TunnelServerDAO) UpdateTunnelServerStatus(tunnelServerId, status string) (*models.TunnelServer, error) {
	ctx := context.Background()

	// 获取服务器信息
	server, err := dao.GetTunnelServer(tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取隧道服务器信息失败")
	}

	// 更新状态信息
	server.ServerStatus = status
	server.EditTime = time.Now()
	server.CurrentVersion++

	// 如果是启动状态，记录启动时间
	if status == "running" {
		now := time.Now()
		server.StartTime = &now
	}

	// 更新数据库
	whereClause := "tunnelServerId = ?"
	args := []interface{}{tunnelServerId}

	_, err = dao.db.Update(ctx, "HUB_TUNNEL_SERVER", server, whereClause, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "更新隧道服务器状态失败")
	}

	logger.Info("更新隧道服务器状态成功", "tunnelServerId", tunnelServerId, "status", status)

	// 返回更新后的服务器信息
	updatedServer, err := dao.GetTunnelServer(tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取更新后的服务器信息失败")
	}

	return updatedServer, nil
}

// GetTunnelServerStats 获取隧道服务器统计信息
func (dao *TunnelServerDAO) GetTunnelServerStats() (*models.TunnelServerStats, error) {
	ctx := context.Background()

	// 查询总服务器数量
	totalQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_SERVER WHERE activeFlag = 'Y'`
	type CountResult struct {
		Count int `db:"COUNT(*)"`
	}
	var totalResult CountResult
	err := dao.db.QueryOne(ctx, &totalResult, totalQuery, nil, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询服务器总数失败")
	}
	totalServers := totalResult.Count

	// 查询在线服务器数量
	onlineQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_SERVER WHERE activeFlag = 'Y' AND serverStatus = 'running'`
	var onlineResult CountResult
	err = dao.db.QueryOne(ctx, &onlineResult, onlineQuery, nil, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询在线服务器数量失败")
	}
	onlineServers := onlineResult.Count

	// 查询客户端总数
	clientQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_CLIENT WHERE activeFlag = 'Y'`
	var clientResult CountResult
	totalClients := 0
	err = dao.db.QueryOne(ctx, &clientResult, clientQuery, nil, true)
	if err == nil {
		totalClients = clientResult.Count
	}
	// 如果表不存在或查询失败，保持默认值0

	// 查询服务总数
	serviceQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_SERVICE WHERE activeFlag = 'Y'`
	var serviceResult CountResult
	totalServices := 0
	err = dao.db.QueryOne(ctx, &serviceResult, serviceQuery, nil, true)
	if err == nil {
		totalServices = serviceResult.Count
	}
	// 如果表不存在或查询失败，保持默认值0

	return &models.TunnelServerStats{
		TotalServers:   totalServers,
		OnlineServers:  onlineServers,
		OfflineServers: totalServers - onlineServers,
		TotalClients:   totalClients,
		TotalServices:  totalServices,
	}, nil
}

// GetServerStatusOptions 获取服务器状态选项
func (dao *TunnelServerDAO) GetServerStatusOptions() []map[string]interface{} {
	// 这个方法返回固定的状态选项，不需要数据库查询
	return []map[string]interface{}{
		{"value": "running", "label": "运行中"},
		{"value": "stopped", "label": "已停止"},
		{"value": "error", "label": "错误"},
	}
}

// CheckServerNameExists 检查服务器名称是否存在
func (dao *TunnelServerDAO) CheckServerNameExists(serverName, excludeId string) (bool, error) {
	ctx := context.Background()

	// 构建查询条件
	whereClause := "serverName = ? AND activeFlag = 'Y'"
	args := []interface{}{serverName}

	// 如果有排除ID，添加到条件中
	if excludeId != "" {
		whereClause += " AND tunnelServerId != ?"
		args = append(args, excludeId)
	}

	// 构建查询语句
	query := fmt.Sprintf("SELECT COUNT(*) FROM HUB_TUNNEL_SERVER WHERE %s", whereClause)

	// 执行查询
	type CountResult struct {
		Count int `db:"COUNT(*)"`
	}
	var result CountResult
	err := dao.db.QueryOne(ctx, &result, query, args, true)
	if err != nil {
		return false, huberrors.WrapError(err, "检查服务器名称存在性失败")
	}

	return result.Count > 0, nil
}

// GetTunnelServerList 获取隧道服务器列表（用于下拉选择）
func (dao *TunnelServerDAO) GetTunnelServerList() ([]*models.TunnelServer, error) {
	ctx := context.Background()

	// 只查询活跃的服务器，并只返回必要的字段
	query := `
		SELECT tunnelServerId, serverName, controlAddress, controlPort, serverStatus
		FROM HUB_TUNNEL_SERVER
		WHERE activeFlag = 'Y'
		ORDER BY serverName ASC
	`

	// 执行查询
	var servers []*models.TunnelServer
	err := dao.db.Query(ctx, &servers, query, nil, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取隧道服务器列表失败")
	}

	return servers, nil
}
