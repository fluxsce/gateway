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
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0062/models"
)

type TunnelServiceDAO struct {
	db database.Database
}

func NewTunnelServiceDAO(db database.Database) *TunnelServiceDAO {
	return &TunnelServiceDAO{db: db}
}

// QueryTunnelServices 查询服务列表（分页）
func (dao *TunnelServiceDAO) QueryTunnelServices(req *models.TunnelServiceQueryRequest) ([]*types.TunnelService, int, error) {
	ctx := context.Background()

	// 构建WHERE条件
	whereConditions := []string{"1=1"}
	args := []interface{}{}

	if req.TenantId != "" {
		whereConditions = append(whereConditions, "tenantId = ?")
		args = append(args, req.TenantId)
	}

	if req.TunnelClientId != "" {
		whereConditions = append(whereConditions, "tunnelClientId = ?")
		args = append(args, req.TunnelClientId)
	}

	if req.ServiceName != "" {
		whereConditions = append(whereConditions, "serviceName LIKE ?")
		args = append(args, "%"+req.ServiceName+"%")
	}

	if req.ServiceType != "" {
		whereConditions = append(whereConditions, "serviceType = ?")
		args = append(args, req.ServiceType)
	}

	if req.ServiceStatus != "" {
		whereConditions = append(whereConditions, "serviceStatus = ?")
		args = append(args, req.ServiceStatus)
	}

	if req.UserId != "" {
		whereConditions = append(whereConditions, "userId = ?")
		args = append(args, req.UserId)
	}

	if req.ActiveFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, req.ActiveFlag)
	}

	if req.Keyword != "" {
		whereConditions = append(whereConditions, "(serviceName LIKE ? OR localAddress LIKE ?)")
		keyword := "%" + req.Keyword + "%"
		args = append(args, keyword, keyword)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 构建基础查询
	baseQuery := fmt.Sprintf(`
		SELECT tunnelServiceId, tenantId, tunnelClientId, userId, serviceName, serviceDescription,
		       serviceType, localAddress, localPort, remotePort, customDomains, subDomain,
		       useEncryption, useCompression, bandwidthLimit, maxConnections,
		       serviceStatus, registeredTime, lastActiveTime, connectionCount, totalConnections, totalTraffic,
		       addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_SERVICE
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
		logger.Error("查询服务总数失败", "error", err)
		errMsg := "查询服务总数失败: " + err.Error()
		return nil, 0, huberrors.NewError(errMsg)
	}

	total := countResult.Count
	if total == 0 {
		return []*types.TunnelService{}, 0, nil
	}

	// 分页查询
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	allArgs := append(args, paginationArgs...)

	services := []*types.TunnelService{}
	err = dao.db.Query(ctx, &services, paginatedQuery, allArgs, true)
	if err != nil {
		logger.Error("查询服务列表失败", "error", err)
		errMsg := "查询服务列表失败: " + err.Error()
		return nil, 0, huberrors.NewError(errMsg)
	}

	return services, total, nil
}

// GetTunnelService 获取服务详情
func (dao *TunnelServiceDAO) GetTunnelService(tunnelServiceId string) (*types.TunnelService, error) {
	ctx := context.Background()

	querySQL := `
		SELECT tunnelServiceId, tenantId, tunnelClientId, userId, serviceName, serviceDescription,
		       serviceType, localAddress, localPort, remotePort, customDomains, subDomain,
		       useEncryption, useCompression, bandwidthLimit, maxConnections,
		       serviceStatus, registeredTime, lastActiveTime, connectionCount, totalConnections, totalTraffic,
		       addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_SERVICE
		WHERE tunnelServiceId = ?
	`

	service := &types.TunnelService{}
	err := dao.db.QueryOne(ctx, service, querySQL, []interface{}{tunnelServiceId}, true)
	if err != nil {
		logger.Error("查询服务详情失败", "tunnelServiceId", tunnelServiceId, "error", err)
		errMsg := "查询服务详情失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	return service, nil
}

// helper function: convert *string to string
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// helper function: convert *int to int
func intValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// CreateTunnelService 创建服务
func (dao *TunnelServiceDAO) CreateTunnelService(service *types.TunnelService) (*types.TunnelService, error) {
	ctx := context.Background()

	// 生成服务ID
	if service.TunnelServiceId == "" {
		service.TunnelServiceId = "service-" + random.GenerateRandomString(16)
	}

	// 检查服务名称唯一性
	exists, err := dao.checkServiceNameExists(service.ServiceName, "")
	if err != nil {
		return nil, err
	}
	if exists {
		errMsg := fmt.Sprintf("服务名称 '%s' 已存在", service.ServiceName)
		return nil, huberrors.NewError(errMsg)
	}

	// 检查客户端是否存在且在线
	clientExists, err := dao.checkClientExists(service.TunnelClientId)
	if err != nil {
		return nil, err
	}
	if !clientExists {
		return nil, huberrors.NewError("关联的客户端不存在或未激活")
	}

	// 设置默认值
	now := time.Now()
	service.AddTime = now
	service.EditTime = now
	service.RegisteredTime = now
	service.CurrentVersion = 1
	if service.ActiveFlag == "" {
		service.ActiveFlag = "Y"
	}
	if service.ServiceStatus == "" {
		service.ServiceStatus = types.ServiceStatusInactive
	}
	if service.UseEncryption == "" {
		service.UseEncryption = "N"
	}
	if service.UseCompression == "" {
		service.UseCompression = "N"
	}
	if service.MaxConnections == nil || *service.MaxConnections == 0 {
		maxConn := 100
		service.MaxConnections = &maxConn
	}

	// 插入数据库
	insertSQL := `
		INSERT INTO HUB_TUNNEL_SERVICE (
			tunnelServiceId, tenantId, tunnelClientId, userId, serviceName, serviceDescription,
			serviceType, localAddress, localPort, remotePort, customDomains, subDomain,
			useEncryption, useCompression, bandwidthLimit, maxConnections,
			serviceStatus, registeredTime, lastActiveTime, connectionCount, totalConnections, totalTraffic,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	_, err = dao.db.Exec(ctx, insertSQL, []interface{}{
		service.TunnelServiceId, service.TenantId, service.TunnelClientId, service.UserId, service.ServiceName, service.ServiceDescription,
		service.ServiceType, service.LocalAddress, service.LocalPort, service.RemotePort, service.CustomDomains, service.SubDomain,
		service.UseEncryption, service.UseCompression, service.BandwidthLimit, service.MaxConnections,
		service.ServiceStatus, service.RegisteredTime, service.LastActiveTime, service.ConnectionCount, service.TotalConnections, service.TotalTraffic,
		service.AddTime, service.AddWho, service.EditTime, service.EditWho, service.OprSeqFlag, service.CurrentVersion, service.ActiveFlag, service.NoteText, service.ExtProperty,
	}, false)

	if err != nil {
		logger.Error("创建服务失败", "serviceName", service.ServiceName, "error", err)
		errMsg := "创建服务失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	// 更新客户端的服务计数
	err = dao.updateClientServiceCount(service.TunnelClientId, 1)
	if err != nil {
		logger.Warn("更新客户端服务计数失败", "tunnelClientId", service.TunnelClientId, "error", err)
	}

	logger.Info("创建服务成功", "tunnelServiceId", service.TunnelServiceId, "serviceName", service.ServiceName)
	return service, nil
}

// UpdateTunnelService 更新服务
func (dao *TunnelServiceDAO) UpdateTunnelService(service *types.TunnelService) (*types.TunnelService, error) {
	ctx := context.Background()

	// 检查服务是否存在
	existing, err := dao.GetTunnelService(service.TunnelServiceId)
	if err != nil {
		return nil, err
	}

	// 检查服务名称唯一性（排除自己）
	if service.ServiceName != existing.ServiceName {
		exists, err := dao.checkServiceNameExists(service.ServiceName, service.TunnelServiceId)
		if err != nil {
			return nil, err
		}
		if exists {
			errMsg := fmt.Sprintf("服务名称 '%s' 已存在", service.ServiceName)
			return nil, huberrors.NewError(errMsg)
		}
	}

	// 更新时间和版本号
	service.EditTime = time.Now()
	service.CurrentVersion = existing.CurrentVersion + 1

	updateSQL := `
		UPDATE HUB_TUNNEL_SERVICE SET
			tunnelClientId = ?, serviceName = ?, serviceDescription = ?, serviceType = ?,
			localAddress = ?, localPort = ?, remotePort = ?, customDomains = ?, subDomain = ?,
			useEncryption = ?, useCompression = ?, bandwidthLimit = ?, maxConnections = ?,
			editTime = ?, editWho = ?, currentVersion = ?,
			activeFlag = ?, noteText = ?, extProperty = ?
		WHERE tunnelServiceId = ? AND currentVersion = ?
	`

	rowsAffected, err := dao.db.Exec(ctx, updateSQL, []interface{}{
		service.TunnelClientId, service.ServiceName, service.ServiceDescription, service.ServiceType,
		service.LocalAddress, service.LocalPort, service.RemotePort, service.CustomDomains, service.SubDomain,
		service.UseEncryption, service.UseCompression, service.BandwidthLimit, service.MaxConnections,
		service.EditTime, service.EditWho, service.CurrentVersion,
		service.ActiveFlag, service.NoteText, service.ExtProperty,
		service.TunnelServiceId, existing.CurrentVersion,
	}, false)

	if err != nil {
		logger.Error("更新服务失败", "tunnelServiceId", service.TunnelServiceId, "error", err)
		errMsg := "更新服务失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	if rowsAffected == 0 {
		return nil, huberrors.NewError("更新失败，服务可能已被其他用户修改，请刷新后重试")
	}

	logger.Info("更新服务成功", "tunnelServiceId", service.TunnelServiceId)
	return dao.GetTunnelService(service.TunnelServiceId)
}

// DeleteTunnelService 删除服务（物理删除）
func (dao *TunnelServiceDAO) DeleteTunnelService(tunnelServiceId, editWho string) (*types.TunnelService, error) {
	ctx := context.Background()

	// 检查服务是否存在
	service, err := dao.GetTunnelService(tunnelServiceId)
	if err != nil {
		return nil, err
	}

	// 先注销服务（如果服务已注册）
	clientManager := getTunnelClientManager()
	if clientManager != nil {
		tunnelClient := clientManager.GetClient(service.TunnelClientId)
		if tunnelClient != nil && tunnelClient.IsConnected() {
			// 尝试注销服务（如果服务已注册，则注销；如果未注册，忽略错误）
			if unregisterErr := tunnelClient.UnregisterService(ctx, tunnelServiceId); unregisterErr != nil {
				logger.Warn("删除服务前注销失败，继续删除", "tunnelServiceId", tunnelServiceId, "error", unregisterErr.Error())
				// 继续执行删除，不中断流程
			}
		}
	}

	// 物理删除
	deleteSQL := `
		DELETE FROM HUB_TUNNEL_SERVICE
		WHERE tunnelServiceId = ?
	`

	_, err = dao.db.Exec(ctx, deleteSQL, []interface{}{tunnelServiceId}, false)
	if err != nil {
		logger.Error("删除服务失败", "tunnelServiceId", tunnelServiceId, "error", err)
		errMsg := "删除服务失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	// 更新客户端的服务计数
	err = dao.updateClientServiceCount(service.TunnelClientId, -1)
	if err != nil {
		logger.Warn("更新客户端服务计数失败", "tunnelClientId", service.TunnelClientId, "error", err)
	}

	logger.Info("删除服务成功", "tunnelServiceId", tunnelServiceId)
	return service, nil
}

// GetServiceStats 获取服务统计信息
func (dao *TunnelServiceDAO) GetServiceStats() (*models.TunnelServiceStats, error) {
	ctx := context.Background()

	querySQL := `
		SELECT 
			COUNT(*) as totalServices,
			SUM(CASE WHEN serviceStatus = 'active' THEN 1 ELSE 0 END) as activeServices,
			SUM(CASE WHEN serviceStatus = 'inactive' THEN 1 ELSE 0 END) as inactiveServices,
			SUM(CASE WHEN serviceStatus = 'error' THEN 1 ELSE 0 END) as errorServices,
			SUM(totalConnections) as totalConnections,
			SUM(totalTraffic) as totalTraffic
		FROM HUB_TUNNEL_SERVICE
	`

	stats := &models.TunnelServiceStats{}
	err := dao.db.QueryOne(ctx, stats, querySQL, []interface{}{}, true)
	if err != nil {
		logger.Error("查询服务统计信息失败", "error", err)
		errMsg := "查询服务统计信息失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	return stats, nil
}

// checkServiceNameExists 检查服务名称是否存在
func (dao *TunnelServiceDAO) checkServiceNameExists(serviceName, excludeServiceId string) (bool, error) {
	ctx := context.Background()

	querySQL := `
		SELECT COUNT(*) as count
		FROM HUB_TUNNEL_SERVICE 
		WHERE serviceName = ?
	`
	args := []interface{}{serviceName}

	if excludeServiceId != "" {
		querySQL += " AND tunnelServiceId != ?"
		args = append(args, excludeServiceId)
	}

	var countResult struct {
		Count int `db:"count"`
	}
	err := dao.db.QueryOne(ctx, &countResult, querySQL, args, true)
	if err != nil {
		logger.Error("检查服务名称唯一性失败", "serviceName", serviceName, "error", err)
		errMsg := "检查服务名称唯一性失败: " + err.Error()
		return false, huberrors.NewError(errMsg)
	}

	return countResult.Count > 0, nil
}

// checkClientExists 检查客户端是否存在
func (dao *TunnelServiceDAO) checkClientExists(tunnelClientId string) (bool, error) {
	ctx := context.Background()

	querySQL := `
		SELECT COUNT(*) as count
		FROM HUB_TUNNEL_CLIENT 
		WHERE tunnelClientId = ?
	`

	var countResult struct {
		Count int `db:"count"`
	}
	err := dao.db.QueryOne(ctx, &countResult, querySQL, []interface{}{tunnelClientId}, true)
	if err != nil {
		logger.Error("检查客户端是否存在失败", "tunnelClientId", tunnelClientId, "error", err)
		errMsg := "检查客户端是否存在失败: " + err.Error()
		return false, huberrors.NewError(errMsg)
	}

	return countResult.Count > 0, nil
}

// updateClientServiceCount 更新客户端的服务计数
func (dao *TunnelServiceDAO) updateClientServiceCount(tunnelClientId string, delta int) error {
	ctx := context.Background()

	updateSQL := `
		UPDATE HUB_TUNNEL_CLIENT
		SET serviceCount = serviceCount + ?, editTime = ?, currentVersion = currentVersion + 1
		WHERE tunnelClientId = ?
	`

	_, err := dao.db.Exec(ctx, updateSQL, []interface{}{delta, time.Now(), tunnelClientId}, false)
	if err != nil {
		errMsg := "更新客户端服务计数失败: " + err.Error()
		return huberrors.NewError(errMsg)
	}

	return nil
}

// RegisterService 注册服务到隧道管理器
func (dao *TunnelServiceDAO) RegisterService(ctx context.Context, tunnelServiceId string) (*types.TunnelService, error) {
	// 1. 查询服务信息
	service, err := dao.GetTunnelService(tunnelServiceId)
	if err != nil {
		return nil, huberrors.NewError("查询服务信息失败: " + err.Error())
	}

	if service == nil {
		return nil, huberrors.NewError("服务不存在")
	}

	if service.ActiveFlag != "Y" {
		return nil, huberrors.NewError("服务未激活，无法注册")
	}

	// 2. 获取隧道客户端管理器
	clientManager := getTunnelClientManager()
	if clientManager == nil {
		return nil, huberrors.NewError("隧道客户端管理器未初始化")
	}

	// 3. 获取对应的客户端实例
	tunnelClient := clientManager.GetClient(service.TunnelClientId)
	if tunnelClient == nil {
		return nil, huberrors.NewError("客户端未找到或未连接")
	}

	// 4. 调用客户端的 RegisterService 方法（客户端层会更新服务状态）
	if err := tunnelClient.RegisterService(ctx, service); err != nil {
		logger.Error("注册服务到客户端失败", "tunnelServiceId", tunnelServiceId, "error", err)
		return nil, huberrors.NewError("注册服务失败: " + err.Error())
	}

	logger.Info("服务注册成功", "tunnelServiceId", tunnelServiceId, "serviceName", service.ServiceName)

	// 查询最新的服务详情返回
	updatedService, err := dao.GetTunnelService(tunnelServiceId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取服务详情失败")
	}

	return updatedService, nil
}

// UnregisterService 从隧道管理器注销服务
func (dao *TunnelServiceDAO) UnregisterService(ctx context.Context, tunnelServiceId string) (*types.TunnelService, error) {
	// 1. 查询服务信息
	service, err := dao.GetTunnelService(tunnelServiceId)
	if err != nil {
		return nil, huberrors.NewError("查询服务信息失败: " + err.Error())
	}

	if service == nil {
		return nil, huberrors.NewError("服务不存在")
	}

	// 2. 获取隧道客户端管理器
	clientManager := getTunnelClientManager()
	if clientManager == nil {
		return nil, huberrors.NewError("隧道客户端管理器未初始化")
	}

	// 3. 获取对应的客户端实例
	tunnelClient := clientManager.GetClient(service.TunnelClientId)
	if tunnelClient == nil {
		return nil, huberrors.NewError("客户端未找到或未连接")
	}

	// 4. 调用客户端的 UnregisterService 方法（客户端层会更新服务状态）
	if err := tunnelClient.UnregisterService(ctx, tunnelServiceId); err != nil {
		logger.Error("从客户端注销服务失败", "tunnelServiceId", tunnelServiceId, "error", err)
		return nil, huberrors.NewError("注销服务失败: " + err.Error())
	}

	logger.Info("服务注销成功", "tunnelServiceId", tunnelServiceId, "serviceName", service.ServiceName)

	// 查询最新的服务详情返回
	updatedService, err := dao.GetTunnelService(tunnelServiceId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取服务详情失败")
	}

	return updatedService, nil
}
