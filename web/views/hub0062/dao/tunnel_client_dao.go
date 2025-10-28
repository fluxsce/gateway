package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0062/models"
)

type TunnelClientDAO struct {
	db database.Database
}

func NewTunnelClientDAO(db database.Database) *TunnelClientDAO {
	return &TunnelClientDAO{db: db}
}

// QueryTunnelClients 查询客户端列表（分页）
func (dao *TunnelClientDAO) QueryTunnelClients(req *models.TunnelClientQueryRequest) ([]*models.TunnelClient, int, error) {
	ctx := context.Background()

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
	} else {
		whereConditions = append(whereConditions, "activeFlag = 'Y'")
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
		       clientVersion, operatingSystem, clientIpAddress, serverAddress, serverPort,
		       authToken, tlsEnable, autoReconnect, maxRetries, retryInterval,
		       reconnectCount, totalConnectTime, heartbeatInterval, heartbeatTimeout, lastHeartbeat,
		       connectionStatus, lastConnectTime, lastDisconnectTime, disconnectReason, serviceCount,
		       addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
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
		return []*models.TunnelClient{}, 0, nil
	}

	// 分页查询
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	allArgs := append(args, paginationArgs...)

	clients := []*models.TunnelClient{}
	err = dao.db.Query(ctx, &clients, paginatedQuery, allArgs, true)
	if err != nil {
		logger.Error("查询客户端列表失败", "error", err)
		errMsg := "查询客户端列表失败: " + err.Error()
		return nil, 0, huberrors.NewError(errMsg)
	}

	return clients, total, nil
}

// GetTunnelClient 获取客户端详情
func (dao *TunnelClientDAO) GetTunnelClient(tunnelClientId string) (*models.TunnelClient, error) {
	ctx := context.Background()

	querySQL := `
		SELECT tunnelClientId, tenantId, userId, clientName, clientDescription, 
		       clientVersion, operatingSystem, clientIpAddress, serverAddress, serverPort,
		       authToken, tlsEnable, autoReconnect, maxRetries, retryInterval,
		       reconnectCount, totalConnectTime, heartbeatInterval, heartbeatTimeout, lastHeartbeat,
		       connectionStatus, lastConnectTime, lastDisconnectTime, disconnectReason, serviceCount,
		       addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_CLIENT
		WHERE tunnelClientId = ? AND activeFlag = 'Y'
	`

	client := &models.TunnelClient{}
	err := dao.db.QueryOne(ctx, client, querySQL, []interface{}{tunnelClientId}, true)
	if err != nil {
		logger.Error("查询客户端详情失败", "tunnelClientId", tunnelClientId, "error", err)
		errMsg := "查询客户端详情失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	return client, nil
}

// CreateTunnelClient 创建客户端
func (dao *TunnelClientDAO) CreateTunnelClient(client *models.TunnelClient) (*models.TunnelClient, error) {
	ctx := context.Background()

	// 生成客户端ID
	if client.TunnelClientId == "" {
		client.TunnelClientId = "client-" + random.GenerateRandomString(16)
	}

	// 生成认证令牌
	if client.AuthToken == "" {
		client.AuthToken = random.GenerateRandomString(32)
	}

	// 检查客户端名称唯一性
	exists, err := dao.checkClientNameExists(client.ClientName, client.TenantId, "")
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

	// 插入数据库
	insertSQL := `
		INSERT INTO HUB_TUNNEL_CLIENT (
			tunnelClientId, tenantId, userId, clientName, clientDescription,
			clientVersion, operatingSystem, clientIpAddress, serverAddress, serverPort,
			authToken, tlsEnable, autoReconnect, maxRetries, retryInterval,
			reconnectCount, totalConnectTime, heartbeatInterval, heartbeatTimeout, lastHeartbeat,
			connectionStatus, lastConnectTime, lastDisconnectTime, disconnectReason, serviceCount,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	_, err = dao.db.Exec(ctx, insertSQL, []interface{}{
		client.TunnelClientId, client.TenantId, client.UserId, client.ClientName, client.ClientDescription,
		client.ClientVersion, client.OperatingSystem, client.ClientIpAddress, client.ServerAddress, client.ServerPort,
		client.AuthToken, client.TlsEnable, client.AutoReconnect, client.MaxRetries, client.RetryInterval,
		client.ReconnectCount, client.TotalConnectTime, client.HeartbeatInterval, client.HeartbeatTimeout, client.LastHeartbeat,
		client.ConnectionStatus, client.LastConnectTime, client.LastDisconnectTime, client.DisconnectReason, client.ServiceCount,
		client.AddTime, client.AddWho, client.EditTime, client.EditWho, client.OprSeqFlag, client.CurrentVersion, client.ActiveFlag, client.NoteText, client.ExtProperty,
	}, false)

	if err != nil {
		logger.Error("创建客户端失败", "clientName", client.ClientName, "error", err)
		errMsg := "创建客户端失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	logger.Info("创建客户端成功", "tunnelClientId", client.TunnelClientId, "clientName", client.ClientName)
	return client, nil
}

// UpdateTunnelClient 更新客户端
func (dao *TunnelClientDAO) UpdateTunnelClient(client *models.TunnelClient) (*models.TunnelClient, error) {
	ctx := context.Background()

	// 检查客户端是否存在
	existing, err := dao.GetTunnelClient(client.TunnelClientId)
	if err != nil {
		return nil, err
	}

	// 检查客户端名称唯一性（排除自己）
	if client.ClientName != existing.ClientName {
		exists, err := dao.checkClientNameExists(client.ClientName, client.TenantId, client.TunnelClientId)
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
			clientIpAddress = ?, serverAddress = ?, serverPort = ?, tlsEnable = ?,
			autoReconnect = ?, maxRetries = ?, retryInterval = ?, heartbeatInterval = ?,
			heartbeatTimeout = ?, editTime = ?, editWho = ?, currentVersion = ?,
			noteText = ?, extProperty = ?
		WHERE tunnelClientId = ? AND activeFlag = 'Y' AND currentVersion = ?
	`

	rowsAffected, err := dao.db.Exec(ctx, updateSQL, []interface{}{
		client.ClientName, client.ClientDescription, client.ClientVersion, client.OperatingSystem,
		client.ClientIpAddress, client.ServerAddress, client.ServerPort, client.TlsEnable,
		client.AutoReconnect, client.MaxRetries, client.RetryInterval, client.HeartbeatInterval,
		client.HeartbeatTimeout, client.EditTime, client.EditWho, client.CurrentVersion,
		client.NoteText, client.ExtProperty,
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
	return dao.GetTunnelClient(client.TunnelClientId)
}

// DeleteTunnelClient 删除客户端（逻辑删除）
func (dao *TunnelClientDAO) DeleteTunnelClient(tunnelClientId, editWho string) (*models.TunnelClient, error) {
	ctx := context.Background()

	// 检查客户端是否存在
	client, err := dao.GetTunnelClient(tunnelClientId)
	if err != nil {
		return nil, err
	}

	// 检查是否有关联的服务
	if client.ServiceCount > 0 {
		return nil, huberrors.NewError("客户端下还有关联的服务，无法删除")
	}

	// 逻辑删除
	deleteSQL := `
		UPDATE HUB_TUNNEL_CLIENT
		SET activeFlag = 'N', editTime = ?, editWho = ?, currentVersion = currentVersion + 1
		WHERE tunnelClientId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, deleteSQL, []interface{}{time.Now(), editWho, tunnelClientId}, false)
	if err != nil {
		logger.Error("删除客户端失败", "tunnelClientId", tunnelClientId, "error", err)
		errMsg := "删除客户端失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	logger.Info("删除客户端成功", "tunnelClientId", tunnelClientId)
	client.ActiveFlag = "N"
	return client, nil
}

// GetClientStats 获取客户端统计信息
func (dao *TunnelClientDAO) GetClientStats() (*models.TunnelClientStats, error) {
	ctx := context.Background()

	querySQL := `
		SELECT 
			COUNT(*) as totalClients,
			SUM(CASE WHEN connectionStatus = 'connected' THEN 1 ELSE 0 END) as connectedClients,
			SUM(CASE WHEN connectionStatus = 'disconnected' THEN 1 ELSE 0 END) as disconnectedClients,
			SUM(CASE WHEN connectionStatus = 'error' THEN 1 ELSE 0 END) as errorClients,
			SUM(serviceCount) as totalServices,
			SUM(reconnectCount) as totalReconnects
		FROM HUB_TUNNEL_CLIENT
		WHERE activeFlag = 'Y'
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

// GetClientStatus 获取客户端实时状态
func (dao *TunnelClientDAO) GetClientStatus(tunnelClientId string) (*models.ClientStatusResponse, error) {
	ctx := context.Background()

	querySQL := `
		SELECT tunnelClientId, clientName, connectionStatus, lastConnectTime, lastHeartbeat,
		       serviceCount, totalConnectTime, reconnectCount
		FROM HUB_TUNNEL_CLIENT
		WHERE tunnelClientId = ? AND activeFlag = 'Y'
	`

	status := &models.ClientStatusResponse{}
	err := dao.db.QueryOne(ctx, status, querySQL, []interface{}{tunnelClientId}, true)
	if err != nil {
		logger.Error("查询客户端状态失败", "tunnelClientId", tunnelClientId, "error", err)
		errMsg := "查询客户端状态失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	return status, nil
}

// ResetAuthToken 重置客户端认证令牌
func (dao *TunnelClientDAO) ResetAuthToken(tunnelClientId, editWho string) (*models.ResetAuthTokenResponse, error) {
	ctx := context.Background()

	// 检查客户端是否存在
	_, err := dao.GetTunnelClient(tunnelClientId)
	if err != nil {
		return nil, err
	}

	// 生成新的认证令牌
	newToken := random.GenerateRandomString(32)

	updateSQL := `
		UPDATE HUB_TUNNEL_CLIENT
		SET authToken = ?, editTime = ?, editWho = ?, currentVersion = currentVersion + 1
		WHERE tunnelClientId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, updateSQL, []interface{}{newToken, time.Now(), editWho, tunnelClientId}, false)
	if err != nil {
		logger.Error("重置认证令牌失败", "tunnelClientId", tunnelClientId, "error", err)
		errMsg := "重置认证令牌失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	logger.Info("重置认证令牌成功", "tunnelClientId", tunnelClientId)
	return &models.ResetAuthTokenResponse{
		TunnelClientId: tunnelClientId,
		AuthToken:      newToken,
	}, nil
}

// DisconnectClient 强制断开客户端连接
func (dao *TunnelClientDAO) DisconnectClient(tunnelClientId, reason, editWho string) error {
	ctx := context.Background()

	// 检查客户端是否存在
	client, err := dao.GetTunnelClient(tunnelClientId)
	if err != nil {
		return err
	}

	if client.ConnectionStatus != "connected" {
		return huberrors.NewError("客户端未连接，无需断开")
	}

	// 更新连接状态
	updateSQL := `
		UPDATE HUB_TUNNEL_CLIENT
		SET connectionStatus = 'disconnected', 
		    lastDisconnectTime = ?, 
		    disconnectReason = ?,
		    editTime = ?, 
		    editWho = ?, 
		    currentVersion = currentVersion + 1
		WHERE tunnelClientId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, updateSQL, []interface{}{time.Now(), reason, time.Now(), editWho, tunnelClientId}, false)
	if err != nil {
		logger.Error("断开客户端连接失败", "tunnelClientId", tunnelClientId, "error", err)
		errMsg := "断开客户端连接失败: " + err.Error()
		return huberrors.NewError(errMsg)
	}

	logger.Info("断开客户端连接成功", "tunnelClientId", tunnelClientId, "reason", reason)
	return nil
}

// BatchEnableClients 批量启用客户端
func (dao *TunnelClientDAO) BatchEnableClients(clientIds []string, editWho string) (*models.BatchOperationResponse, error) {
	ctx := context.Background()
	response := &models.BatchOperationResponse{
		SuccessCount: 0,
		FailedCount:  0,
		FailedIds:    []string{},
	}

	for _, clientId := range clientIds {
		updateSQL := `
			UPDATE HUB_TUNNEL_CLIENT
			SET activeFlag = 'Y', editTime = ?, editWho = ?, currentVersion = currentVersion + 1
			WHERE tunnelClientId = ?
		`

		_, err := dao.db.Exec(ctx, updateSQL, []interface{}{time.Now(), editWho, clientId}, false)
		if err != nil {
			logger.Error("启用客户端失败", "tunnelClientId", clientId, "error", err)
			response.FailedCount++
			response.FailedIds = append(response.FailedIds, clientId)
		} else {
			response.SuccessCount++
		}
	}

	logger.Info("批量启用客户端完成", "successCount", response.SuccessCount, "failedCount", response.FailedCount)
	return response, nil
}

// BatchDisableClients 批量禁用客户端
func (dao *TunnelClientDAO) BatchDisableClients(clientIds []string, editWho string) (*models.BatchOperationResponse, error) {
	ctx := context.Background()
	response := &models.BatchOperationResponse{
		SuccessCount: 0,
		FailedCount:  0,
		FailedIds:    []string{},
	}

	for _, clientId := range clientIds {
		updateSQL := `
			UPDATE HUB_TUNNEL_CLIENT
			SET activeFlag = 'N', editTime = ?, editWho = ?, currentVersion = currentVersion + 1
			WHERE tunnelClientId = ?
		`

		_, err := dao.db.Exec(ctx, updateSQL, []interface{}{time.Now(), editWho, clientId}, false)
		if err != nil {
			logger.Error("禁用客户端失败", "tunnelClientId", clientId, "error", err)
			response.FailedCount++
			response.FailedIds = append(response.FailedIds, clientId)
		} else {
			response.SuccessCount++
		}
	}

	logger.Info("批量禁用客户端完成", "successCount", response.SuccessCount, "failedCount", response.FailedCount)
	return response, nil
}

// checkClientNameExists 检查客户端名称是否存在
func (dao *TunnelClientDAO) checkClientNameExists(clientName, tenantId, excludeClientId string) (bool, error) {
	ctx := context.Background()

	querySQL := `
		SELECT COUNT(*) as count
		FROM HUB_TUNNEL_CLIENT 
		WHERE clientName = ? AND tenantId = ? AND activeFlag = 'Y'
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
