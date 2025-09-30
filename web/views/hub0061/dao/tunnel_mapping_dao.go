package dao

import (
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/views/hub0061/models"
	"strings"
	"time"
)

// TunnelMappingDAO 隧道映射数据访问对象
type TunnelMappingDAO struct {
	db database.Database
}

// NewTunnelMappingDAO 创建隧道映射DAO实例
func NewTunnelMappingDAO(db database.Database) *TunnelMappingDAO {
	return &TunnelMappingDAO{db: db}
}

// QueryTunnelMappings 查询隧道映射列表
func (dao *TunnelMappingDAO) QueryTunnelMappings(req *models.TunnelMappingQueryRequest) ([]*models.TunnelMapping, int, error) {
	// 简化实现：返回模拟数据
	mapping1 := &models.TunnelMapping{
		TunnelMappingId:   "mapping-001",
		MappingName:       "Web端口映射",
		MappingType:       "PORT",
		TunnelServerId:    "server-001",
		ServerName:        "主服务器",
		ExternalPort:      &[]int{8080}[0],
		InternalPort:      &[]int{80}[0],
		Protocol:          &[]string{"TCP"}[0],
		EnableSSL:         false,
		RateLimitEnabled:  true,
		RequestsPerSecond: &[]int{100}[0],
		CacheEnabled:      true,
		CacheTTL:          &[]int{300}[0],
		LogEnabled:        true,
		LogLevel:          &[]string{"INFO"}[0],
		MappingStatus:     "ACTIVE",
		TotalRequests:     1000,
		TotalTraffic:      1024 * 1024 * 100,
		ActiveConnections: 5,
		AddTime:           time.Now().Add(-24 * time.Hour),
		AddWho:            "admin",
		EditTime:          time.Now(),
		EditWho:           "admin",
		CurrentVersion:    1,
		ActiveFlag:        "Y",
	}

	mapping2 := &models.TunnelMapping{
		TunnelMappingId:   "mapping-002",
		MappingName:       "域名映射",
		MappingType:       "DOMAIN",
		TunnelServerId:    "server-001",
		ServerName:        "主服务器",
		ExternalDomain:    &[]string{"api.example.com"}[0],
		InternalHost:      &[]string{"192.168.1.100"}[0],
		InternalPort2:     &[]int{3000}[0],
		EnableSSL:         true,
		ForceHTTPS:        true,
		RateLimitEnabled:  false,
		CacheEnabled:      false,
		LogEnabled:        true,
		LogLevel:          &[]string{"DEBUG"}[0],
		MappingStatus:     "ACTIVE",
		TotalRequests:     500,
		TotalTraffic:      1024 * 1024 * 50,
		ActiveConnections: 3,
		AddTime:           time.Now().Add(-12 * time.Hour),
		AddWho:            "admin",
		EditTime:          time.Now().Add(-1 * time.Hour),
		EditWho:           "admin",
		CurrentVersion:    1,
		ActiveFlag:        "Y",
	}

	mappings := []*models.TunnelMapping{mapping1, mapping2}

	// 应用过滤条件
	if req.MappingName != "" {
		var filtered []*models.TunnelMapping
		for _, mapping := range mappings {
			if strings.Contains(mapping.MappingName, req.MappingName) {
				filtered = append(filtered, mapping)
			}
		}
		mappings = filtered
	}

	if req.MappingType != "" {
		var filtered []*models.TunnelMapping
		for _, mapping := range mappings {
			if mapping.MappingType == req.MappingType {
				filtered = append(filtered, mapping)
			}
		}
		mappings = filtered
	}

	if req.TunnelServerId != "" {
		var filtered []*models.TunnelMapping
		for _, mapping := range mappings {
			if mapping.TunnelServerId == req.TunnelServerId {
				filtered = append(filtered, mapping)
			}
		}
		mappings = filtered
	}

	if req.MappingStatus != "" {
		var filtered []*models.TunnelMapping
		for _, mapping := range mappings {
			if mapping.MappingStatus == req.MappingStatus {
				filtered = append(filtered, mapping)
			}
		}
		mappings = filtered
	}

	return mappings, len(mappings), nil
}

// GetTunnelMapping 获取隧道映射详情
func (dao *TunnelMappingDAO) GetTunnelMapping(tunnelMappingId string) (*models.TunnelMapping, error) {
	// 简化实现：返回模拟数据
	if tunnelMappingId == "mapping-001" {
		return &models.TunnelMapping{
			TunnelMappingId:   "mapping-001",
			MappingName:       "Web端口映射",
			MappingType:       "PORT",
			TunnelServerId:    "server-001",
			ServerName:        "主服务器",
			ExternalPort:      &[]int{8080}[0],
			InternalPort:      &[]int{80}[0],
			Protocol:          &[]string{"TCP"}[0],
			EnableSSL:         false,
			RateLimitEnabled:  true,
			RequestsPerSecond: &[]int{100}[0],
			CacheEnabled:      true,
			CacheTTL:          &[]int{300}[0],
			LogEnabled:        true,
			LogLevel:          &[]string{"INFO"}[0],
			MappingStatus:     "ACTIVE",
			TotalRequests:     1000,
			TotalTraffic:      1024 * 1024 * 100,
			ActiveConnections: 5,
			AddTime:           time.Now().Add(-24 * time.Hour),
			AddWho:            "admin",
			EditTime:          time.Now(),
			EditWho:           "admin",
			CurrentVersion:    1,
			ActiveFlag:        "Y",
		}, nil
	}

	return nil, fmt.Errorf("隧道映射不存在: %s", tunnelMappingId)
}

// CreateTunnelMapping 创建隧道映射
func (dao *TunnelMappingDAO) CreateTunnelMapping(mapping *models.TunnelMapping) error {
	logger.Info("创建隧道映射", "tunnelMappingId", mapping.TunnelMappingId, "mappingName", mapping.MappingName)
	return nil
}

// UpdateTunnelMapping 更新隧道映射
func (dao *TunnelMappingDAO) UpdateTunnelMapping(mapping *models.TunnelMapping) error {
	logger.Info("更新隧道映射", "tunnelMappingId", mapping.TunnelMappingId, "mappingName", mapping.MappingName)
	return nil
}

// DeleteTunnelMapping 删除隧道映射（逻辑删除）
func (dao *TunnelMappingDAO) DeleteTunnelMapping(tunnelMappingId, editWho string) error {
	logger.Info("删除隧道映射", "tunnelMappingId", tunnelMappingId, "editWho", editWho)
	return nil
}

// UpdateTunnelMappingStatus 更新隧道映射状态
func (dao *TunnelMappingDAO) UpdateTunnelMappingStatus(tunnelMappingId, status string, activeConnections int, errorMessage *string) error {
	logger.Info("更新隧道映射状态", "tunnelMappingId", tunnelMappingId, "status", status)
	return nil
}

// UpdateTunnelMappingTraffic 更新隧道映射流量统计
func (dao *TunnelMappingDAO) UpdateTunnelMappingTraffic(tunnelMappingId string, totalRequests, totalTraffic int64) error {
	logger.Info("更新隧道映射流量统计", "tunnelMappingId", tunnelMappingId)
	return nil
}

// GetTunnelMappingStats 获取隧道映射统计信息
func (dao *TunnelMappingDAO) GetTunnelMappingStats() (*models.TunnelMappingStats, error) {
	return &models.TunnelMappingStats{
		TotalMappings:    2,
		ActiveMappings:   2,
		InactiveMappings: 0,
		ErrorMappings:    0,
		PortMappings:     1,
		DomainMappings:   1,
		TotalRequests:    1500,
		TotalTraffic:     1024 * 1024 * 150,
	}, nil
}

// GetMappingTypeOptions 获取映射类型选项
func (dao *TunnelMappingDAO) GetMappingTypeOptions() []map[string]interface{} {
	return []map[string]interface{}{
		{"value": "PORT", "label": "端口映射"},
		{"value": "DOMAIN", "label": "域名映射"},
		{"value": "SUBDOMAIN", "label": "子域名映射"},
	}
}

// GetMappingStatusOptions 获取映射状态选项
func (dao *TunnelMappingDAO) GetMappingStatusOptions() []map[string]interface{} {
	return []map[string]interface{}{
		{"value": "ACTIVE", "label": "活跃"},
		{"value": "INACTIVE", "label": "非活跃"},
		{"value": "ERROR", "label": "错误"},
	}
}

// GetProtocolOptions 获取协议选项
func (dao *TunnelMappingDAO) GetProtocolOptions() []map[string]interface{} {
	return []map[string]interface{}{
		{"value": "TCP", "label": "TCP"},
		{"value": "UDP", "label": "UDP"},
	}
}

// CheckPortExists 检查端口是否已被占用
func (dao *TunnelMappingDAO) CheckPortExists(externalPort int, protocol, excludeId string) (bool, error) {
	// 简化实现：假设端口8080已被占用
	if externalPort == 8080 && excludeId == "" {
		return true, nil
	}
	return false, nil
}

// CheckDomainExists 检查域名是否已被占用
func (dao *TunnelMappingDAO) CheckDomainExists(domain, excludeId string) (bool, error) {
	// 简化实现：假设api.example.com已被占用
	if domain == "api.example.com" && excludeId == "" {
		return true, nil
	}
	return false, nil
}

// GetPortUsageList 获取端口使用列表
func (dao *TunnelMappingDAO) GetPortUsageList() ([]*models.PortUsageInfo, error) {
	return []*models.PortUsageInfo{
		{
			Port:        8080,
			Protocol:    "TCP",
			MappingName: "Web端口映射",
			Status:      "ACTIVE",
		},
		{
			Port:        3306,
			Protocol:    "TCP",
			MappingName: "数据库端口映射",
			Status:      "INACTIVE",
		},
	}, nil
}

// GetDomainUsageList 获取域名使用列表
func (dao *TunnelMappingDAO) GetDomainUsageList() ([]*models.DomainUsageInfo, error) {
	return []*models.DomainUsageInfo{
		{
			Domain:      "api.example.com",
			MappingName: "域名映射",
			Status:      "ACTIVE",
			EnableSSL:   true,
		},
		{
			Domain:      "test.example.com",
			MappingName: "测试域名映射",
			Status:      "INACTIVE",
			EnableSSL:   false,
		},
	}, nil
}
