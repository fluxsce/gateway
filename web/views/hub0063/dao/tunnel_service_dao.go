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
	"gateway/web/views/hub0063/models"
)

type TunnelServiceDAO struct {
	db database.Database
}

func NewTunnelServiceDAO(db database.Database) *TunnelServiceDAO {
	return &TunnelServiceDAO{db: db}
}

// QueryTunnelServices 查询服务列表（分页）
func (dao *TunnelServiceDAO) QueryTunnelServices(req *models.TunnelServiceQueryRequest) ([]*models.TunnelService, int, error) {
	ctx := context.Background()

	// 构建WHERE条件
	whereConditions := []string{"1=1"}
	args := []interface{}{}

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
	} else {
		whereConditions = append(whereConditions, "activeFlag = 'Y'")
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
		return []*models.TunnelService{}, 0, nil
	}

	// 分页查询
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	allArgs := append(args, paginationArgs...)

	services := []*models.TunnelService{}
	err = dao.db.Query(ctx, &services, paginatedQuery, allArgs, true)
	if err != nil {
		logger.Error("查询服务列表失败", "error", err)
		errMsg := "查询服务列表失败: " + err.Error()
		return nil, 0, huberrors.NewError(errMsg)
	}

	return services, total, nil
}

// GetTunnelService 获取服务详情
func (dao *TunnelServiceDAO) GetTunnelService(tunnelServiceId string) (*models.TunnelService, error) {
	ctx := context.Background()

	querySQL := `
		SELECT tunnelServiceId, tenantId, tunnelClientId, userId, serviceName, serviceDescription,
		       serviceType, localAddress, localPort, remotePort, customDomains, subDomain,
		       useEncryption, useCompression, bandwidthLimit, maxConnections,
		       serviceStatus, registeredTime, lastActiveTime, connectionCount, totalConnections, totalTraffic,
		       addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_SERVICE
		WHERE tunnelServiceId = ? AND activeFlag = 'Y'
	`

	service := &models.TunnelService{}
	err := dao.db.QueryOne(ctx, service, querySQL, []interface{}{tunnelServiceId}, true)
	if err != nil {
		logger.Error("查询服务详情失败", "tunnelServiceId", tunnelServiceId, "error", err)
		errMsg := "查询服务详情失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	return service, nil
}

// CreateTunnelService 创建服务
func (dao *TunnelServiceDAO) CreateTunnelService(service *models.TunnelService) (*models.TunnelService, error) {
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
		service.ServiceStatus = "active"
	}
	if service.UseEncryption == "" {
		service.UseEncryption = "N"
	}
	if service.UseCompression == "" {
		service.UseCompression = "N"
	}
	if service.MaxConnections == 0 {
		service.MaxConnections = 100
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
func (dao *TunnelServiceDAO) UpdateTunnelService(service *models.TunnelService) (*models.TunnelService, error) {
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
			serviceName = ?, serviceDescription = ?, serviceType = ?,
			localAddress = ?, localPort = ?, remotePort = ?, customDomains = ?, subDomain = ?,
			useEncryption = ?, useCompression = ?, bandwidthLimit = ?, maxConnections = ?,
			editTime = ?, editWho = ?, currentVersion = ?,
			noteText = ?, extProperty = ?
		WHERE tunnelServiceId = ? AND activeFlag = 'Y' AND currentVersion = ?
	`

	rowsAffected, err := dao.db.Exec(ctx, updateSQL, []interface{}{
		service.ServiceName, service.ServiceDescription, service.ServiceType,
		service.LocalAddress, service.LocalPort, service.RemotePort, service.CustomDomains, service.SubDomain,
		service.UseEncryption, service.UseCompression, service.BandwidthLimit, service.MaxConnections,
		service.EditTime, service.EditWho, service.CurrentVersion,
		service.NoteText, service.ExtProperty,
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

// DeleteTunnelService 删除服务（逻辑删除）
func (dao *TunnelServiceDAO) DeleteTunnelService(tunnelServiceId, editWho string) (*models.TunnelService, error) {
	ctx := context.Background()

	// 检查服务是否存在
	service, err := dao.GetTunnelService(tunnelServiceId)
	if err != nil {
		return nil, err
	}

	// 逻辑删除
	deleteSQL := `
		UPDATE HUB_TUNNEL_SERVICE
		SET activeFlag = 'N', editTime = ?, editWho = ?, currentVersion = currentVersion + 1
		WHERE tunnelServiceId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, deleteSQL, []interface{}{time.Now(), editWho, tunnelServiceId}, false)
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
	service.ActiveFlag = "N"
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
		WHERE activeFlag = 'Y'
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

// GetServicesByClient 按客户端查询服务列表
func (dao *TunnelServiceDAO) GetServicesByClient(req *models.ServicesByClientRequest) ([]*models.TunnelService, error) {
	ctx := context.Background()

	whereConditions := []string{"tunnelClientId = ?", "activeFlag = 'Y'"}
	args := []interface{}{req.TunnelClientId}

	if req.ServiceStatus != "" {
		whereConditions = append(whereConditions, "serviceStatus = ?")
		args = append(args, req.ServiceStatus)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	querySQL := fmt.Sprintf(`
		SELECT tunnelServiceId, tenantId, tunnelClientId, userId, serviceName, serviceDescription,
		       serviceType, localAddress, localPort, remotePort, customDomains, subDomain,
		       useEncryption, useCompression, bandwidthLimit, maxConnections,
		       serviceStatus, registeredTime, lastActiveTime, connectionCount, totalConnections, totalTraffic,
		       addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_SERVICE
		WHERE %s
		ORDER BY registeredTime DESC
	`, whereClause)

	services := []*models.TunnelService{}
	err := dao.db.Query(ctx, &services, querySQL, args, true)
	if err != nil {
		logger.Error("查询客户端服务列表失败", "tunnelClientId", req.TunnelClientId, "error", err)
		errMsg := "查询客户端服务列表失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	return services, nil
}

// AllocateRemotePort 分配远程端口
func (dao *TunnelServiceDAO) AllocateRemotePort(tunnelServiceId string, preferredPort int) (*models.AllocatePortResponse, error) {
	ctx := context.Background()

	// 检查服务是否存在
	service, err := dao.GetTunnelService(tunnelServiceId)
	if err != nil {
		return nil, err
	}

	// 检查服务类型是否需要端口
	if service.ServiceType != "tcp" && service.ServiceType != "udp" {
		return nil, huberrors.NewError("只有TCP/UDP类型的服务需要分配远程端口")
	}

	// 如果已经分配了端口，直接返回
	if service.RemotePort != nil && *service.RemotePort > 0 {
		return &models.AllocatePortResponse{
			TunnelServiceId: tunnelServiceId,
			RemotePort:      *service.RemotePort,
		}, nil
	}

	// 分配端口
	allocatedPort, err := dao.findAvailablePort(preferredPort)
	if err != nil {
		return nil, err
	}

	// 更新服务的远程端口
	updateSQL := `
		UPDATE HUB_TUNNEL_SERVICE
		SET remotePort = ?, editTime = ?, currentVersion = currentVersion + 1
		WHERE tunnelServiceId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, updateSQL, []interface{}{allocatedPort, time.Now(), tunnelServiceId}, false)
	if err != nil {
		logger.Error("分配远程端口失败", "tunnelServiceId", tunnelServiceId, "error", err)
		errMsg := "分配远程端口失败: " + err.Error()
		return nil, huberrors.NewError(errMsg)
	}

	logger.Info("分配远程端口成功", "tunnelServiceId", tunnelServiceId, "remotePort", allocatedPort)
	return &models.AllocatePortResponse{
		TunnelServiceId: tunnelServiceId,
		RemotePort:      allocatedPort,
	}, nil
}

// ReleaseRemotePort 释放远程端口
func (dao *TunnelServiceDAO) ReleaseRemotePort(tunnelServiceId string) error {
	ctx := context.Background()

	// 检查服务是否存在
	_, err := dao.GetTunnelService(tunnelServiceId)
	if err != nil {
		return err
	}

	// 释放端口
	updateSQL := `
		UPDATE HUB_TUNNEL_SERVICE
		SET remotePort = NULL, editTime = ?, currentVersion = currentVersion + 1
		WHERE tunnelServiceId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, updateSQL, []interface{}{time.Now(), tunnelServiceId}, false)
	if err != nil {
		logger.Error("释放远程端口失败", "tunnelServiceId", tunnelServiceId, "error", err)
		errMsg := "释放远程端口失败: " + err.Error()
		return huberrors.NewError(errMsg)
	}

	logger.Info("释放远程端口成功", "tunnelServiceId", tunnelServiceId)
	return nil
}

// EnableService 启用服务
func (dao *TunnelServiceDAO) EnableService(tunnelServiceId, editWho string) error {
	return dao.updateServiceStatus(tunnelServiceId, "active", editWho)
}

// DisableService 禁用服务
func (dao *TunnelServiceDAO) DisableService(tunnelServiceId, editWho string) error {
	return dao.updateServiceStatus(tunnelServiceId, "inactive", editWho)
}

// updateServiceStatus 更新服务状态
func (dao *TunnelServiceDAO) updateServiceStatus(tunnelServiceId, status, editWho string) error {
	ctx := context.Background()

	updateSQL := `
		UPDATE HUB_TUNNEL_SERVICE
		SET serviceStatus = ?, editTime = ?, editWho = ?, currentVersion = currentVersion + 1
		WHERE tunnelServiceId = ? AND activeFlag = 'Y'
	`

	_, err := dao.db.Exec(ctx, updateSQL, []interface{}{status, time.Now(), editWho, tunnelServiceId}, false)
	if err != nil {
		logger.Error("更新服务状态失败", "tunnelServiceId", tunnelServiceId, "status", status, "error", err)
		errMsg := "更新服务状态失败: " + err.Error()
		return huberrors.NewError(errMsg)
	}

	logger.Info("更新服务状态成功", "tunnelServiceId", tunnelServiceId, "status", status)
	return nil
}

// checkServiceNameExists 检查服务名称是否存在
func (dao *TunnelServiceDAO) checkServiceNameExists(serviceName, excludeServiceId string) (bool, error) {
	ctx := context.Background()

	querySQL := `
		SELECT COUNT(*) as count
		FROM HUB_TUNNEL_SERVICE 
		WHERE serviceName = ? AND activeFlag = 'Y'
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

// checkClientExists 检查客户端是否存在且激活
func (dao *TunnelServiceDAO) checkClientExists(tunnelClientId string) (bool, error) {
	ctx := context.Background()

	querySQL := `
		SELECT COUNT(*) as count
		FROM HUB_TUNNEL_CLIENT 
		WHERE tunnelClientId = ? AND activeFlag = 'Y'
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
		WHERE tunnelClientId = ? AND activeFlag = 'Y'
	`

	_, err := dao.db.Exec(ctx, updateSQL, []interface{}{delta, time.Now(), tunnelClientId}, false)
	if err != nil {
		errMsg := "更新客户端服务计数失败: " + err.Error()
		return huberrors.NewError(errMsg)
	}

	return nil
}

// findAvailablePort 查找可用端口
func (dao *TunnelServiceDAO) findAvailablePort(preferredPort int) (int, error) {
	ctx := context.Background()

	// 端口范围：10000-60000
	startPort := 10000
	endPort := 60000

	// 如果指定了首选端口，先检查是否可用
	if preferredPort >= startPort && preferredPort <= endPort {
		available, err := dao.isPortAvailable(preferredPort)
		if err != nil {
			return 0, err
		}
		if available {
			return preferredPort, nil
		}
	}

	// 查询已使用的端口
	querySQL := `
		SELECT remotePort
		FROM HUB_TUNNEL_SERVICE
		WHERE remotePort IS NOT NULL AND activeFlag = 'Y'
		ORDER BY remotePort
	`

	var usedPorts []struct {
		RemotePort int `db:"remotePort"`
	}
	err := dao.db.Query(ctx, &usedPorts, querySQL, []interface{}{}, true)
	if err != nil {
		logger.Error("查询已使用端口失败", "error", err)
		errMsg := "查询已使用端口失败: " + err.Error()
		return 0, huberrors.NewError(errMsg)
	}

	// 构建已使用端口的集合
	usedPortMap := make(map[int]bool)
	for _, p := range usedPorts {
		usedPortMap[p.RemotePort] = true
	}

	// 查找第一个可用端口
	for port := startPort; port <= endPort; port++ {
		if !usedPortMap[port] {
			return port, nil
		}
	}

	return 0, huberrors.NewError("没有可用的端口")
}

// isPortAvailable 检查端口是否可用
func (dao *TunnelServiceDAO) isPortAvailable(port int) (bool, error) {
	ctx := context.Background()

	querySQL := `
		SELECT COUNT(*) as count
		FROM HUB_TUNNEL_SERVICE
		WHERE remotePort = ? AND activeFlag = 'Y'
	`

	var countResult struct {
		Count int `db:"count"`
	}
	err := dao.db.QueryOne(ctx, &countResult, querySQL, []interface{}{port}, true)
	if err != nil {
		logger.Error("检查端口可用性失败", "port", port, "error", err)
		errMsg := "检查端口可用性失败: " + err.Error()
		return false, huberrors.NewError(errMsg)
	}

	return countResult.Count == 0, nil
}
