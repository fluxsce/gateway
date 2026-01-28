package dao

import (
	"context"
	"fmt"
	"gateway/internal/tunnel/server"
	"gateway/internal/tunnel/types"
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
func (dao *TunnelServerDAO) QueryTunnelServers(ctx context.Context, req *models.TunnelServerQueryRequest) ([]*types.TunnelServer, int, error) {

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
		return []*types.TunnelServer{}, 0, nil
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
	var servers []*types.TunnelServer
	err = dao.db.Query(ctx, &servers, paginatedQuery, allArgs, true)
	if err != nil {
		logger.Error("查询隧道服务器数据失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "查询隧道服务器数据失败")
	}

	return servers, countResult.Count, nil
}

// GetTunnelServer 获取隧道服务器详情
func (dao *TunnelServerDAO) GetTunnelServer(ctx context.Context, tunnelServerId string) (*types.TunnelServer, error) {

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

	server := &types.TunnelServer{}
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
func (dao *TunnelServerDAO) CreateTunnelServer(ctx context.Context, server *types.TunnelServer) (*types.TunnelServer, error) {

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
	exists, err := dao.CheckServerNameExists(ctx, server.ServerName, "")
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
	createdServer, err := dao.GetTunnelServer(ctx, server.TunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取创建后的服务器信息失败")
	}

	return createdServer, nil
}

// UpdateTunnelServer 更新隧道服务器
func (dao *TunnelServerDAO) UpdateTunnelServer(ctx context.Context, server *types.TunnelServer) (*types.TunnelServer, error) {

	// 检查服务器是否存在
	existingServer, err := dao.GetTunnelServer(ctx, server.TunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取隧道服务器信息失败")
	}

	// 检查服务器名称是否重复
	if server.ServerName != existingServer.ServerName {
		exists, err := dao.CheckServerNameExists(ctx, server.ServerName, server.TunnelServerId)
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

	_, err = dao.db.Update(ctx, "HUB_TUNNEL_SERVER", server, whereClause, args, true, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "更新隧道服务器失败")
	}

	logger.Info("更新隧道服务器成功", "tunnelServerId", server.TunnelServerId, "serverName", server.ServerName)

	// 返回更新后的服务器信息
	updatedServer, err := dao.GetTunnelServer(ctx, server.TunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取更新后的服务器信息失败")
	}

	return updatedServer, nil
}

// DeleteTunnelServer 删除隧道服务器（物理删除）
func (dao *TunnelServerDAO) DeleteTunnelServer(ctx context.Context, tunnelServerId, editWho string) (*types.TunnelServer, error) {
	// 获取服务器信息
	server, err := dao.GetTunnelServer(ctx, tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取隧道服务器信息失败")
	}

	// 物理删除
	deleteSQL := `
		DELETE FROM HUB_TUNNEL_SERVER
		WHERE tunnelServerId = ?
	`

	_, err = dao.db.Exec(ctx, deleteSQL, []interface{}{tunnelServerId}, false)
	if err != nil {
		return nil, huberrors.WrapError(err, "删除隧道服务器失败")
	}

	logger.Info("删除隧道服务器成功", "tunnelServerId", tunnelServerId, "editWho", editWho)

	return server, nil
}

// UpdateTunnelServerStatus 更新隧道服务器状态
func (dao *TunnelServerDAO) UpdateTunnelServerStatus(ctx context.Context, tunnelServerId, status string) (*types.TunnelServer, error) {
	// 获取服务器信息
	server, err := dao.GetTunnelServer(ctx, tunnelServerId)
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

	_, err = dao.db.Update(ctx, "HUB_TUNNEL_SERVER", server, whereClause, args, true, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "更新隧道服务器状态失败")
	}

	logger.Info("更新隧道服务器状态成功", "tunnelServerId", tunnelServerId, "status", status)

	// 返回更新后的服务器信息
	updatedServer, err := dao.GetTunnelServer(ctx, tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取更新后的服务器信息失败")
	}

	return updatedServer, nil
}

// GetTunnelServerStats 获取隧道服务器统计信息
func (dao *TunnelServerDAO) GetTunnelServerStats(ctx context.Context) (*models.TunnelServerStats, error) {

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

	// 查询运行中服务器数量
	runningQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_SERVER WHERE activeFlag = 'Y' AND serverStatus = 'running'`
	var runningResult CountResult
	err = dao.db.QueryOne(ctx, &runningResult, runningQuery, nil, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询运行中服务器数量失败")
	}
	runningServers := runningResult.Count

	// 查询已停止服务器数量
	stoppedQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_SERVER WHERE activeFlag = 'Y' AND serverStatus = 'stopped'`
	var stoppedResult CountResult
	err = dao.db.QueryOne(ctx, &stoppedResult, stoppedQuery, nil, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询已停止服务器数量失败")
	}
	stoppedServers := stoppedResult.Count

	// 查询错误服务器数量
	errorQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_SERVER WHERE activeFlag = 'Y' AND serverStatus = 'error'`
	var errorResult CountResult
	err = dao.db.QueryOne(ctx, &errorResult, errorQuery, nil, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询错误服务器数量失败")
	}
	errorServers := errorResult.Count

	// 查询客户端总数
	clientQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_CLIENT WHERE activeFlag = 'Y'`
	var clientResult CountResult
	totalClients := 0
	err = dao.db.QueryOne(ctx, &clientResult, clientQuery, nil, true)
	if err == nil {
		totalClients = clientResult.Count
	}
	// 如果表不存在或查询失败，保持默认值0

	// 查询服务总数（作为连接数）
	serviceQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_SERVICE WHERE activeFlag = 'Y'`
	var serviceResult CountResult
	totalConnections := 0
	err = dao.db.QueryOne(ctx, &serviceResult, serviceQuery, nil, true)
	if err == nil {
		totalConnections = serviceResult.Count
	}
	// 如果表不存在或查询失败，保持默认值0

	return &models.TunnelServerStats{
		TotalServers:     totalServers,
		RunningServers:   runningServers,
		StoppedServers:   stoppedServers,
		ErrorServers:     errorServers,
		TotalClients:     totalClients,
		TotalConnections: totalConnections,
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
func (dao *TunnelServerDAO) CheckServerNameExists(ctx context.Context, serverName, excludeId string) (bool, error) {

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
func (dao *TunnelServerDAO) GetTunnelServerList(ctx context.Context) ([]*types.TunnelServer, error) {

	// 只查询活跃的服务器，并只返回必要的字段
	query := `
		SELECT tunnelServerId, serverName, controlAddress, controlPort, serverStatus
		FROM HUB_TUNNEL_SERVER
		WHERE activeFlag = 'Y'
		ORDER BY serverName ASC
	`

	// 执行查询
	var servers []*types.TunnelServer
	err := dao.db.Query(ctx, &servers, query, nil, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取隧道服务器列表失败")
	}

	return servers, nil
}

// StartTunnelServer 启动隧道服务器
func (dao *TunnelServerDAO) StartTunnelServer(ctx context.Context, tunnelServerId string) (*types.TunnelServer, error) {

	// 获取隧道服务端管理器
	serverManager := getTunnelServerManager()
	if serverManager == nil {
		return nil, huberrors.NewError("隧道服务端管理器未初始化")
	}

	// 调用隧道服务端管理器启动服务器
	err := serverManager.Start(ctx, tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "启动隧道服务器失败")
	}

	logger.Info("启动隧道服务器成功", "tunnelServerId", tunnelServerId)

	// 查询最新的服务器详情返回
	server, err := dao.GetTunnelServer(ctx, tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取服务器详情失败")
	}

	return server, nil
}

// StopTunnelServer 停止隧道服务器
func (dao *TunnelServerDAO) StopTunnelServer(ctx context.Context, tunnelServerId string) (*types.TunnelServer, error) {

	// 获取隧道服务端管理器
	serverManager := getTunnelServerManager()
	if serverManager == nil {
		return nil, huberrors.NewError("隧道服务端管理器未初始化")
	}

	// 调用隧道服务端管理器停止服务器
	err := serverManager.Stop(ctx, tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "停止隧道服务器失败")
	}

	logger.Info("停止隧道服务器成功", "tunnelServerId", tunnelServerId)

	// 查询最新的服务器详情返回
	server, err := dao.GetTunnelServer(ctx, tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取服务器详情失败")
	}

	return server, nil
}

// RestartTunnelServer 重启隧道服务器
func (dao *TunnelServerDAO) RestartTunnelServer(ctx context.Context, tunnelServerId string) (*types.TunnelServer, error) {

	// 获取隧道服务端管理器
	serverManager := getTunnelServerManager()
	if serverManager == nil {
		return nil, huberrors.NewError("隧道服务端管理器未初始化")
	}

	logger.Info("开始重启隧道服务器", "tunnelServerId", tunnelServerId)

	// 先停止服务器
	err := serverManager.Stop(ctx, tunnelServerId)
	if err != nil {
		logger.Warn("停止服务器失败，尝试继续启动", "tunnelServerId", tunnelServerId, "error", err)
		// 不返回错误，继续尝试启动
	}

	// 再启动服务器（会自动从数据库加载最新配置）
	err = serverManager.Start(ctx, tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "重启隧道服务器失败")
	}

	logger.Info("重启隧道服务器成功", "tunnelServerId", tunnelServerId)

	// 查询最新的服务器详情返回
	server, err := dao.GetTunnelServer(ctx, tunnelServerId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取服务器详情失败")
	}

	return server, nil
}

// ReloadTunnelServerConfig 重新加载隧道服务器配置
func (dao *TunnelServerDAO) ReloadTunnelServerConfig(ctx context.Context, tunnelServerId string) error {

	// 获取隧道服务端管理器
	serverManager := getTunnelServerManager()
	if serverManager == nil {
		return huberrors.NewError("隧道服务端管理器未初始化")
	}

	// 获取服务器实例
	server := serverManager.GetServer(tunnelServerId)
	if server == nil {
		return huberrors.NewError("服务器不存在或未加载: %s", tunnelServerId)
	}

	// 调用隧道服务端管理器重新加载配置
	// Reload 方法会自动从数据库加载最新配置
	err := serverManager.Reload(ctx, server.GetConfig())
	if err != nil {
		return huberrors.WrapError(err, "重新加载隧道服务器配置失败")
	}

	logger.Info("重新加载隧道服务器配置成功", "tunnelServerId", tunnelServerId)
	return nil
}

// getTunnelServerManager 获取隧道服务端管理器实例
func getTunnelServerManager() *server.TunnelServerManager {
	return server.GetTunnelServerManager()
}
