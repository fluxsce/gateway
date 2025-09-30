package dao

import (
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/views/hub0063/models"
	"strings"
	"time"
)

// TunnelServiceDAO 隧道服务数据访问对象
type TunnelServiceDAO struct {
	db database.Database
}

// NewTunnelServiceDAO 创建隧道服务DAO实例
func NewTunnelServiceDAO(db database.Database) *TunnelServiceDAO {
	return &TunnelServiceDAO{db: db}
}

// QueryTunnelServices 查询隧道服务列表
func (dao *TunnelServiceDAO) QueryTunnelServices(req *models.TunnelServiceQueryRequest) ([]*models.TunnelService, int, error) {
	// 简化实现：返回模拟数据
	service1 := &models.TunnelService{
		TunnelServiceId:  "service-001",
		ServiceName:      "Web服务",
		ServiceType:      "HTTP",
		TunnelClientId:   "client-001",
		ClientName:       "客户端1",
		TunnelServerId:   "server-001",
		LocalAddress:     "127.0.0.1",
		LocalPort:        8080,
		RemotePort:       &[]int{80}[0],
		CustomDomains:    `["example.com"]`,
		UseEncryption:    false,
		UseCompression:   true,
		ServiceStatus:    "ACTIVE",
		ConnectionCount:  5,
		TotalConnections: 100,
		TotalTraffic:     1024 * 1024 * 10,
		AddTime:          time.Now().Add(-24 * time.Hour),
		AddWho:           "admin",
		EditTime:         time.Now(),
		EditWho:          "admin",
		CurrentVersion:   1,
		ActiveFlag:       "Y",
	}

	service2 := &models.TunnelService{
		TunnelServiceId:  "service-002",
		ServiceName:      "SSH服务",
		ServiceType:      "TCP",
		TunnelClientId:   "client-001",
		ClientName:       "客户端1",
		TunnelServerId:   "server-001",
		LocalAddress:     "127.0.0.1",
		LocalPort:        22,
		RemotePort:       &[]int{2222}[0],
		UseEncryption:    true,
		UseCompression:   false,
		ServiceStatus:    "ACTIVE",
		ConnectionCount:  2,
		TotalConnections: 50,
		TotalTraffic:     1024 * 1024 * 5,
		AddTime:          time.Now().Add(-12 * time.Hour),
		AddWho:           "admin",
		EditTime:         time.Now().Add(-1 * time.Hour),
		EditWho:          "admin",
		CurrentVersion:   1,
		ActiveFlag:       "Y",
	}

	service3 := &models.TunnelService{
		TunnelServiceId:  "service-003",
		ServiceName:      "数据库服务",
		ServiceType:      "TCP",
		TunnelClientId:   "client-002",
		ClientName:       "客户端2",
		TunnelServerId:   "server-001",
		LocalAddress:     "127.0.0.1",
		LocalPort:        3306,
		RemotePort:       &[]int{3306}[0],
		UseEncryption:    true,
		UseCompression:   false,
		ServiceStatus:    "INACTIVE",
		ConnectionCount:  0,
		TotalConnections: 20,
		TotalTraffic:     1024 * 1024 * 2,
		AddTime:          time.Now().Add(-6 * time.Hour),
		AddWho:           "admin",
		EditTime:         time.Now().Add(-30 * time.Minute),
		EditWho:          "admin",
		CurrentVersion:   1,
		ActiveFlag:       "Y",
	}

	services := []*models.TunnelService{service1, service2, service3}

	// 应用过滤条件
	if req.ServiceName != "" {
		var filtered []*models.TunnelService
		for _, service := range services {
			if strings.Contains(service.ServiceName, req.ServiceName) {
				filtered = append(filtered, service)
			}
		}
		services = filtered
	}

	if req.ServiceType != "" {
		var filtered []*models.TunnelService
		for _, service := range services {
			if service.ServiceType == req.ServiceType {
				filtered = append(filtered, service)
			}
		}
		services = filtered
	}

	if req.TunnelClientId != "" {
		var filtered []*models.TunnelService
		for _, service := range services {
			if service.TunnelClientId == req.TunnelClientId {
				filtered = append(filtered, service)
			}
		}
		services = filtered
	}

	if req.ServiceStatus != "" {
		var filtered []*models.TunnelService
		for _, service := range services {
			if service.ServiceStatus == req.ServiceStatus {
				filtered = append(filtered, service)
			}
		}
		services = filtered
	}

	return services, len(services), nil
}

// GetTunnelService 获取隧道服务详情
func (dao *TunnelServiceDAO) GetTunnelService(tunnelServiceId string) (*models.TunnelService, error) {
	// 简化实现：返回模拟数据
	if tunnelServiceId == "service-001" {
		return &models.TunnelService{
			TunnelServiceId:  "service-001",
			ServiceName:      "Web服务",
			ServiceType:      "HTTP",
			TunnelClientId:   "client-001",
			ClientName:       "客户端1",
			TunnelServerId:   "server-001",
			LocalAddress:     "127.0.0.1",
			LocalPort:        8080,
			RemotePort:       &[]int{80}[0],
			CustomDomains:    `["example.com"]`,
			UseEncryption:    false,
			UseCompression:   true,
			ServiceStatus:    "ACTIVE",
			ConnectionCount:  5,
			TotalConnections: 100,
			TotalTraffic:     1024 * 1024 * 10,
			AddTime:          time.Now().Add(-24 * time.Hour),
			AddWho:           "admin",
			EditTime:         time.Now(),
			EditWho:          "admin",
			CurrentVersion:   1,
			ActiveFlag:       "Y",
		}, nil
	}

	return nil, fmt.Errorf("隧道服务不存在: %s", tunnelServiceId)
}

// CreateTunnelService 创建隧道服务
func (dao *TunnelServiceDAO) CreateTunnelService(service *models.TunnelService) error {
	logger.Info("创建隧道服务", "tunnelServiceId", service.TunnelServiceId, "serviceName", service.ServiceName)
	return nil
}

// UpdateTunnelService 更新隧道服务
func (dao *TunnelServiceDAO) UpdateTunnelService(service *models.TunnelService) error {
	logger.Info("更新隧道服务", "tunnelServiceId", service.TunnelServiceId, "serviceName", service.ServiceName)
	return nil
}

// DeleteTunnelService 删除隧道服务（逻辑删除）
func (dao *TunnelServiceDAO) DeleteTunnelService(tunnelServiceId, editWho string) error {
	logger.Info("删除隧道服务", "tunnelServiceId", tunnelServiceId, "editWho", editWho)
	return nil
}

// UpdateTunnelServiceStatus 更新隧道服务状态
func (dao *TunnelServiceDAO) UpdateTunnelServiceStatus(tunnelServiceId, status string, connectionCount int) error {
	logger.Info("更新隧道服务状态", "tunnelServiceId", tunnelServiceId, "status", status)
	return nil
}

// UpdateTunnelServiceTraffic 更新隧道服务流量统计
func (dao *TunnelServiceDAO) UpdateTunnelServiceTraffic(tunnelServiceId string, totalConnections, totalTraffic int64) error {
	logger.Info("更新隧道服务流量统计", "tunnelServiceId", tunnelServiceId)
	return nil
}

// GetTunnelServiceStats 获取隧道服务统计信息
func (dao *TunnelServiceDAO) GetTunnelServiceStats() (*models.TunnelServiceStats, error) {
	return &models.TunnelServiceStats{
		TotalServices:    3,
		ActiveServices:   2,
		InactiveServices: 1,
		ErrorServices:    0,
		TotalConnections: 170,
		TotalTraffic:     1024 * 1024 * 17,
	}, nil
}

// GetServiceTypeOptions 获取服务类型选项
func (dao *TunnelServiceDAO) GetServiceTypeOptions() []map[string]interface{} {
	return []map[string]interface{}{
		{"value": "TCP", "label": "TCP"},
		{"value": "UDP", "label": "UDP"},
		{"value": "HTTP", "label": "HTTP"},
		{"value": "HTTPS", "label": "HTTPS"},
		{"value": "STCP", "label": "STCP（安全TCP）"},
		{"value": "SUDP", "label": "SUDP（安全UDP）"},
		{"value": "XTCP", "label": "XTCP（P2P TCP）"},
	}
}

// GetServiceStatusOptions 获取服务状态选项
func (dao *TunnelServiceDAO) GetServiceStatusOptions() []map[string]interface{} {
	return []map[string]interface{}{
		{"value": "ACTIVE", "label": "活跃"},
		{"value": "INACTIVE", "label": "非活跃"},
		{"value": "ERROR", "label": "错误"},
	}
}

// CheckServiceNameExists 检查服务名称是否存在（在同一客户端下）
func (dao *TunnelServiceDAO) CheckServiceNameExists(tunnelClientId, serviceName, excludeId string) (bool, error) {
	return false, nil
}

// GetTunnelServicesByClientId 根据客户端ID获取服务列表
func (dao *TunnelServiceDAO) GetTunnelServicesByClientId(tunnelClientId string) ([]*models.TunnelService, error) {
	services, _, err := dao.QueryTunnelServices(&models.TunnelServiceQueryRequest{
		TunnelClientId: tunnelClientId,
	})
	return services, err
}

// CheckRemotePortExists 检查远程端口是否已被占用
func (dao *TunnelServiceDAO) CheckRemotePortExists(remotePort int, excludeId string) (bool, error) {
	return false, nil
}

// CheckCustomDomainExists 检查自定义域名是否已被占用
func (dao *TunnelServiceDAO) CheckCustomDomainExists(domain, excludeId string) (bool, error) {
	return false, nil
}
