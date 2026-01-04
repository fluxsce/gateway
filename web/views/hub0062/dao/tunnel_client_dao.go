package dao

import (
	"fmt"
	"strings"
	"time"

	"gateway/internal/tunnel/client"
	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0062/models"

	"github.com/gin-gonic/gin"
)

type TunnelClientDAO struct {
	db database.Database
}

func NewTunnelClientDAO(db database.Database) *TunnelClientDAO {
	return &TunnelClientDAO{db: db}
}

// QueryTunnelClients 查询客户端列表（分页）
func (dao *TunnelClientDAO) QueryTunnelClients(ginCtx *gin.Context, req *models.TunnelClientQueryRequest) ([]*types.TunnelClient, int, error) {
	ctx := ginCtx.Request.Context()

	// 构建WHERE条件
	whereConditions := []string{"1=1"}
	args := []interface{}{}

	if req.ClientName != "" {
		whereConditions = append(whereConditions, "clientName LIKE ?")
		args = append(args, "%"+req.ClientName+"%")
	}

	if req.ConnectionStatus != "" {
		whereConditions = append(whereConditions, "connectionStatus = ?")
		args = append(args, req.ConnectionStatus)
	}

	if req.UserId != "" {
		whereConditions = append(whereConditions, "userId = ?")
		args = append(args, req.UserId)
	}

	if req.ServerAddress != "" {
		whereConditions = append(whereConditions, "serverAddress = ?")
		args = append(args, req.ServerAddress)
	}

	if req.ActiveFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, req.ActiveFlag)
	}

	if req.Keyword != "" {
		whereConditions = append(whereConditions, "(clientName LIKE ? OR clientIpAddress LIKE ?)")
		keyword := "%" + req.Keyword + "%"
		args = append(args, keyword, keyword)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 构建基础查询
	baseQuery := fmt.Sprintf(`
		SELECT tunnelClientId, tenantId, userId, clientName, clientDescription, 
		       clientVersion, operatingSystem, clientIpAddress, clientMacAddress, serverAddress, serverPort,
		       authToken, tlsEnable, autoReconnect, maxRetries, retryInterval,
		       reconnectCount, totalConnectTime, heartbeatInterval, heartbeatTimeout, lastHeartbeat,
		       connectionStatus, lastConnectTime, lastDisconnectTime, serviceCount,
		       clientConfig, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_CLIENT
		WHERE %s
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
	err = dao.db.QueryOne(ctx, &countResult, countQuery, args, true)
	if err != nil {
		logger.Error("查询客户端总数失败", "error", err)
		errMsg := "查询客户端总数失败: " + err.Error()
		return nil, 0, huberrors.NewError(errMsg)
	}

	total := countResult.Count
	if total == 0 {
		return []*types.TunnelClient{}, 0, nil
	}

	// 分页查询
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	allArgs := append(args, paginationArgs...)

	clients := []*types.TunnelClient{}
	err = dao.db.Query(ctx, &clients, paginatedQuery, allArgs, true)
	if err != nil {
		logger.Error("查询客户端列表失败", "error", err)
		errMsg := "查询客户端列表失败: " + err.Error()
		return nil, 0, huberrors.NewError(errMsg)
	}

	return clients, total, nil
}

// GetTunnelClient 获取客户端详情
func (dao *TunnelClientDAO) GetTunnelClient(ginCtx *gin.Context, tunnelClientId string) (*types.TunnelClient, error) {
	ctx := ginCtx.Request.Context()

	querySQL := `
		SELECT tunnelClientId, tenantId, userId, clientName, clientDescription, 
		       clientVersion, operatingSystem, clientIpAddress, clientMacAddress, serverAddress, serverPort,
		       authToken, tlsEnable, autoReconnect, maxRetries, retryInterval,
		       reconnectCount, totalConnectTime, heartbeatInterval, heartbeatTimeout, lastHeartbeat,
		       connectionStatus, lastConnectTime, lastDisconnectTime, serviceCount,
		       clientConfig, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_CLIENT
		WHERE tunnelClientId = ?
	`

	client := &types.TunnelClient{}
	err := dao.db.QueryOne(ctx, client, querySQL, []interface{}{tunnelClientId}, true)
	if err != nil {
		logger.Error("查询客户端详情失败", "tunnelClientId", tunnelClientId, "error", err)
		errMsg := "查询客户端详情失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	return client, nil
}

// CreateTunnelClient 创建客户端
func (dao *TunnelClientDAO) CreateTunnelClient(ginCtx *gin.Context, client *types.TunnelClient) (*types.TunnelClient, error) {
	ctx := ginCtx.Request.Context()

	// 生成客户端ID
	if client.TunnelClientId == "" {
		client.TunnelClientId = "client-" + random.GenerateRandomString(16)
	}

	// 生成认证令牌
	if client.AuthToken == "" {
		client.AuthToken = random.GenerateRandomString(32)
	}

	// 检查客户端名称唯一性
	exists, err := dao.checkClientNameExists(ginCtx, client.ClientName, client.TenantId, "")
	if err != nil {
		return nil, err
	}
	if exists {
		errMsg := fmt.Sprintf("客户端名称 '%s' 已存在", client.ClientName)
		return nil, huberrors.NewError(errMsg)
	}

	// 设置默认值
	now := time.Now()
	client.AddTime = now
	client.EditTime = now
	client.CurrentVersion = 1
	if client.ActiveFlag == "" {
		client.ActiveFlag = "Y"
	}
	if client.ConnectionStatus == "" {
		client.ConnectionStatus = "disconnected"
	}
	if client.TlsEnable == "" {
		client.TlsEnable = "N"
	}
	if client.AutoReconnect == "" {
		client.AutoReconnect = "Y"
	}
	if client.MaxRetries == 0 {
		client.MaxRetries = 5
	}
	if client.RetryInterval == 0 {
		client.RetryInterval = 20
	}
	if client.HeartbeatInterval == 0 {
		client.HeartbeatInterval = 30
	}
	if client.HeartbeatTimeout == 0 {
		client.HeartbeatTimeout = 90
	}
	if client.ServiceCount == 0 {
		client.ServiceCount = 0
	}

	// 插入数据库（使用便捷方法）
	_, err = dao.db.Insert(ctx, "HUB_TUNNEL_CLIENT", client, true)
	if err != nil {
		logger.Error("创建客户端失败", "clientName", client.ClientName, "error", err)
		errMsg := "创建客户端失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	logger.Info("创建客户端成功", "tunnelClientId", client.TunnelClientId, "clientName", client.ClientName)
	return client, nil
}

// UpdateTunnelClient 更新客户端
func (dao *TunnelClientDAO) UpdateTunnelClient(ginCtx *gin.Context, client *types.TunnelClient) (*types.TunnelClient, error) {
	ctx := ginCtx.Request.Context()

	// 检查客户端是否存在
	existing, err := dao.GetTunnelClient(ginCtx, client.TunnelClientId)
	if err != nil {
		return nil, err
	}

	// 检查客户端名称唯一性（排除自己）
	if client.ClientName != existing.ClientName {
		exists, err := dao.checkClientNameExists(ginCtx, client.ClientName, client.TenantId, client.TunnelClientId)
		if err != nil {
			return nil, err
		}
		if exists {
			errMsg := fmt.Sprintf("客户端名称 '%s' 已存在", client.ClientName)
			return nil, huberrors.NewError(errMsg)
		}
	}

	// 更新时间和版本号
	client.EditTime = time.Now()
	client.CurrentVersion = existing.CurrentVersion + 1

	updateSQL := `
		UPDATE HUB_TUNNEL_CLIENT SET
			clientName = ?, clientDescription = ?, clientVersion = ?, operatingSystem = ?,
			clientIpAddress = ?, clientMacAddress = ?, serverAddress = ?, serverPort = ?,
			authToken = ?, tlsEnable = ?, autoReconnect = ?, maxRetries = ?, retryInterval = ?,
			heartbeatInterval = ?, heartbeatTimeout = ?, clientConfig = ?,
			editTime = ?, editWho = ?, currentVersion = ?,
			activeFlag = ?, noteText = ?, extProperty = ?
		WHERE tunnelClientId = ? AND currentVersion = ?
	`

	rowsAffected, err := dao.db.Exec(ctx, updateSQL, []interface{}{
		client.ClientName, client.ClientDescription, client.ClientVersion, client.OperatingSystem,
		client.ClientIpAddress, client.ClientMacAddress, client.ServerAddress, client.ServerPort,
		client.AuthToken, client.TlsEnable, client.AutoReconnect, client.MaxRetries, client.RetryInterval,
		client.HeartbeatInterval, client.HeartbeatTimeout, client.ClientConfig,
		client.EditTime, client.EditWho, client.CurrentVersion,
		client.ActiveFlag, client.NoteText, client.ExtProperty,
		client.TunnelClientId, existing.CurrentVersion,
	}, false)

	if err != nil {
		logger.Error("更新客户端失败", "tunnelClientId", client.TunnelClientId, "error", err)
		errMsg := "更新客户端失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	if rowsAffected == 0 {
		return nil, huberrors.NewError("更新失败，客户端可能已被其他用户修改，请刷新后重试")
	}

	logger.Info("更新客户端成功", "tunnelClientId", client.TunnelClientId)
	return dao.GetTunnelClient(ginCtx, client.TunnelClientId)
}

// DeleteTunnelClient 删除客户端（物理删除）
func (dao *TunnelClientDAO) DeleteTunnelClient(ginCtx *gin.Context, tunnelClientId, editWho string) (*types.TunnelClient, error) {
	ctx := ginCtx.Request.Context()

	// 检查客户端是否存在
	client, err := dao.GetTunnelClient(ginCtx, tunnelClientId)
	if err != nil {
		return nil, err
	}

	// 检查是否有关联的服务
	if client.ServiceCount > 0 {
		return nil, huberrors.NewError("客户端下还有关联的服务，无法删除")
	}

	// 物理删除
	deleteSQL := `
		DELETE FROM HUB_TUNNEL_CLIENT
		WHERE tunnelClientId = ?
	`

	_, err = dao.db.Exec(ctx, deleteSQL, []interface{}{tunnelClientId}, false)
	if err != nil {
		logger.Error("删除客户端失败", "tunnelClientId", tunnelClientId, "error", err)
		errMsg := "删除客户端失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	logger.Info("删除客户端成功", "tunnelClientId", tunnelClientId)
	return client, nil
}

// GetClientStats 获取客户端统计信息
func (dao *TunnelClientDAO) GetClientStats(ginCtx *gin.Context) (*models.TunnelClientStats, error) {
	ctx := ginCtx.Request.Context()

	querySQL := `
		SELECT 
			COUNT(*) as totalClients,
			SUM(CASE WHEN connectionStatus = 'connected' THEN 1 ELSE 0 END) as connectedClients,
			SUM(CASE WHEN connectionStatus = 'disconnected' THEN 1 ELSE 0 END) as disconnectedClients,
			SUM(CASE WHEN connectionStatus = 'error' THEN 1 ELSE 0 END) as errorClients,
			SUM(serviceCount) as totalServices,
			SUM(reconnectCount) as totalReconnects
		FROM HUB_TUNNEL_CLIENT
	`

	stats := &models.TunnelClientStats{}
	err := dao.db.QueryOne(ctx, stats, querySQL, []interface{}{}, true)
	if err != nil {
		logger.Error("查询客户端统计信息失败", "error", err)
		errMsg := "查询客户端统计信息失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	return stats, nil
}

// StartClient 启动客户端（连接到服务器）
func (dao *TunnelClientDAO) StartClient(ginCtx *gin.Context, tunnelClientId string) (*types.TunnelClient, error) {
	ctx := ginCtx.Request.Context()

	// 检查客户端是否存在
	_, err := dao.GetTunnelClient(ginCtx, tunnelClientId)
	if err != nil {
		return nil, err
	}

	// 调用TunnelClientManager启动客户端
	clientManager := getTunnelClientManager()
	if clientManager == nil {
		return nil, huberrors.NewError("隧道客户端管理器未初始化")
	}

	err = clientManager.Start(ctx, tunnelClientId)
	if err != nil {
		logger.Error("启动客户端失败", "tunnelClientId", tunnelClientId, "error", err)
		return nil, huberrors.WrapError(err, "启动客户端失败")
	}

	logger.Info("启动客户端成功", "tunnelClientId", tunnelClientId)

	// 查询最新的客户端详情返回
	client, err := dao.GetTunnelClient(ginCtx, tunnelClientId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取客户端详情失败")
	}

	return client, nil
}

// StopClient 停止客户端（断开连接）
func (dao *TunnelClientDAO) StopClient(ginCtx *gin.Context, tunnelClientId string) (*types.TunnelClient, error) {
	ctx := ginCtx.Request.Context()

	// 检查客户端是否存在
	_, err := dao.GetTunnelClient(ginCtx, tunnelClientId)
	if err != nil {
		return nil, err
	}

	// 调用TunnelClientManager停止客户端
	clientManager := getTunnelClientManager()
	if clientManager == nil {
		return nil, huberrors.NewError("隧道客户端管理器未初始化")
	}

	err = clientManager.Stop(ctx, tunnelClientId)
	if err != nil {
		logger.Error("停止客户端失败", "tunnelClientId", tunnelClientId, "error", err)
		return nil, huberrors.WrapError(err, "停止客户端失败")
	}

	logger.Info("停止客户端成功", "tunnelClientId", tunnelClientId)

	// 查询最新的客户端详情返回
	client, err := dao.GetTunnelClient(ginCtx, tunnelClientId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取客户端详情失败")
	}

	return client, nil
}

// RestartClient 重启客户端（重新连接）
func (dao *TunnelClientDAO) RestartClient(ginCtx *gin.Context, tunnelClientId string) (*types.TunnelClient, error) {
	ctx := ginCtx.Request.Context()

	// 检查客户端是否存在
	_, err := dao.GetTunnelClient(ginCtx, tunnelClientId)
	if err != nil {
		return nil, err
	}

	// 调用TunnelClientManager重启客户端（先停止再启动）
	clientManager := getTunnelClientManager()
	if clientManager == nil {
		return nil, huberrors.NewError("隧道客户端管理器未初始化")
	}

	// 先停止
	err = clientManager.Stop(ctx, tunnelClientId)
	if err != nil {
		logger.Warn("停止客户端时出错（可能未运行）", "tunnelClientId", tunnelClientId, "error", err)
	}

	// 再启动
	err = clientManager.Start(ctx, tunnelClientId)
	if err != nil {
		logger.Error("重启客户端失败", "tunnelClientId", tunnelClientId, "error", err)
		return nil, huberrors.WrapError(err, "重启客户端失败")
	}

	logger.Info("重启客户端成功", "tunnelClientId", tunnelClientId)

	// 查询最新的客户端详情返回
	client, err := dao.GetTunnelClient(ginCtx, tunnelClientId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取客户端详情失败")
	}

	return client, nil
}

// checkClientNameExists 检查客户端名称是否存在
func (dao *TunnelClientDAO) checkClientNameExists(ginCtx *gin.Context, clientName, tenantId, excludeClientId string) (bool, error) {
	ctx := ginCtx.Request.Context()

	querySQL := `
		SELECT COUNT(*) as count
		FROM HUB_TUNNEL_CLIENT 
		WHERE clientName = ? AND tenantId = ?
	`
	args := []interface{}{clientName, tenantId}

	if excludeClientId != "" {
		querySQL += " AND tunnelClientId != ?"
		args = append(args, excludeClientId)
	}

	var countResult struct {
		Count int `db:"count"`
	}
	err := dao.db.QueryOne(ctx, &countResult, querySQL, args, true)
	if err != nil {
		logger.Error("检查客户端名称唯一性失败", "clientName", clientName, "error", err)
		errMsg := "检查客户端名称唯一性失败: " + err.Error()
		return false, huberrors.NewError(errMsg)
	}

	return countResult.Count > 0, nil
}

// getTunnelClientManager 获取全局隧道客户端管理器实例
func getTunnelClientManager() *client.TunnelClientManager {
	return client.GetTunnelClientManager()
}
