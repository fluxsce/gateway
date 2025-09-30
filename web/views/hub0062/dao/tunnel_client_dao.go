package dao

import (
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/views/hub0062/models"
	"strings"
	"time"
)

// TunnelClientDAO 隧道客户端数据访问对象
type TunnelClientDAO struct {
	db database.Database
}

// NewTunnelClientDAO 创建隧道客户端DAO实例
func NewTunnelClientDAO(db database.Database) *TunnelClientDAO {
	return &TunnelClientDAO{db: db}
}

// QueryTunnelClients 查询隧道客户端列表
func (dao *TunnelClientDAO) QueryTunnelClients(req *models.TunnelClientQueryRequest) ([]*models.TunnelClient, int, error) {
	// 简化实现：返回模拟数据
	client1 := &models.TunnelClient{
		TunnelClientId:     "client-001",
		ClientName:         "客户端1",
		ClientAddress:      "192.168.1.200",
		TunnelServerId:     "server-001",
		ServerName:         "主服务器",
		AuthToken:          "client-token-123",
		HeartbeatInterval:  30,
		MaxRetries:         3,
		RetryInterval:      5,
		ClientStatus:       "CONNECTED",
		LastConnectTime:    &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
		LastHeartbeat:      &[]time.Time{time.Now().Add(-1 * time.Minute)}[0],
		ReconnectCount:     2,
		RegisteredServices: 3,
		ActiveProxies:      2,
		TotalTraffic:       1024 * 1024 * 50,
		AddTime:            time.Now().Add(-24 * time.Hour),
		AddWho:             "admin",
		EditTime:           time.Now(),
		EditWho:            "admin",
		CurrentVersion:     1,
		ActiveFlag:         "Y",
	}

	client2 := &models.TunnelClient{
		TunnelClientId:     "client-002",
		ClientName:         "客户端2",
		ClientAddress:      "192.168.1.201",
		TunnelServerId:     "server-001",
		ServerName:         "主服务器",
		AuthToken:          "client-token-456",
		HeartbeatInterval:  30,
		MaxRetries:         3,
		RetryInterval:      5,
		ClientStatus:       "DISCONNECTED",
		LastConnectTime:    &[]time.Time{time.Now().Add(-2 * time.Hour)}[0],
		LastHeartbeat:      &[]time.Time{time.Now().Add(-10 * time.Minute)}[0],
		ReconnectCount:     5,
		RegisteredServices: 1,
		ActiveProxies:      0,
		TotalTraffic:       1024 * 1024 * 20,
		AddTime:            time.Now().Add(-12 * time.Hour),
		AddWho:             "admin",
		EditTime:           time.Now().Add(-2 * time.Hour),
		EditWho:            "admin",
		CurrentVersion:     1,
		ActiveFlag:         "Y",
	}

	clients := []*models.TunnelClient{client1, client2}

	// 应用过滤条件
	if req.ClientName != "" {
		var filtered []*models.TunnelClient
		for _, client := range clients {
			if strings.Contains(client.ClientName, req.ClientName) {
				filtered = append(filtered, client)
			}
		}
		clients = filtered
	}

	if req.TunnelServerId != "" {
		var filtered []*models.TunnelClient
		for _, client := range clients {
			if client.TunnelServerId == req.TunnelServerId {
				filtered = append(filtered, client)
			}
		}
		clients = filtered
	}

	if req.ClientStatus != "" {
		var filtered []*models.TunnelClient
		for _, client := range clients {
			if client.ClientStatus == req.ClientStatus {
				filtered = append(filtered, client)
			}
		}
		clients = filtered
	}

	return clients, len(clients), nil
}

// GetTunnelClient 获取隧道客户端详情
func (dao *TunnelClientDAO) GetTunnelClient(tunnelClientId string) (*models.TunnelClient, error) {
	// 简化实现：返回模拟数据
	if tunnelClientId == "client-001" {
		return &models.TunnelClient{
			TunnelClientId:     "client-001",
			ClientName:         "客户端1",
			ClientAddress:      "192.168.1.200",
			TunnelServerId:     "server-001",
			ServerName:         "主服务器",
			AuthToken:          "client-token-123",
			HeartbeatInterval:  30,
			MaxRetries:         3,
			RetryInterval:      5,
			ClientStatus:       "CONNECTED",
			LastConnectTime:    &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
			LastHeartbeat:      &[]time.Time{time.Now().Add(-1 * time.Minute)}[0],
			ReconnectCount:     2,
			RegisteredServices: 3,
			ActiveProxies:      2,
			TotalTraffic:       1024 * 1024 * 50,
			AddTime:            time.Now().Add(-24 * time.Hour),
			AddWho:             "admin",
			EditTime:           time.Now(),
			EditWho:            "admin",
			CurrentVersion:     1,
			ActiveFlag:         "Y",
		}, nil
	}

	return nil, fmt.Errorf("隧道客户端不存在: %s", tunnelClientId)
}

// CreateTunnelClient 创建隧道客户端
func (dao *TunnelClientDAO) CreateTunnelClient(client *models.TunnelClient) error {
	logger.Info("创建隧道客户端", "tunnelClientId", client.TunnelClientId, "clientName", client.ClientName)
	return nil
}

// UpdateTunnelClient 更新隧道客户端
func (dao *TunnelClientDAO) UpdateTunnelClient(client *models.TunnelClient) error {
	logger.Info("更新隧道客户端", "tunnelClientId", client.TunnelClientId, "clientName", client.ClientName)
	return nil
}

// DeleteTunnelClient 删除隧道客户端（逻辑删除）
func (dao *TunnelClientDAO) DeleteTunnelClient(tunnelClientId, editWho string) error {
	logger.Info("删除隧道客户端", "tunnelClientId", tunnelClientId, "editWho", editWho)
	return nil
}

// UpdateTunnelClientStatus 更新隧道客户端状态
func (dao *TunnelClientDAO) UpdateTunnelClientStatus(tunnelClientId, status string, registeredServices, activeProxies int, totalTraffic int64) error {
	logger.Info("更新隧道客户端状态", "tunnelClientId", tunnelClientId, "status", status)
	return nil
}

// UpdateTunnelClientConnection 更新隧道客户端连接信息
func (dao *TunnelClientDAO) UpdateTunnelClientConnection(tunnelClientId, status string) error {
	logger.Info("更新隧道客户端连接信息", "tunnelClientId", tunnelClientId, "status", status)
	return nil
}

// GetTunnelClientStats 获取隧道客户端统计信息
func (dao *TunnelClientDAO) GetTunnelClientStats() (*models.TunnelClientStats, error) {
	return &models.TunnelClientStats{
		TotalClients:        2,
		ConnectedClients:    1,
		DisconnectedClients: 1,
		TotalServices:       4,
		ActiveServices:      2,
	}, nil
}

// GetClientStatusOptions 获取客户端状态选项
func (dao *TunnelClientDAO) GetClientStatusOptions() []map[string]interface{} {
	return []map[string]interface{}{
		{"value": "CONNECTED", "label": "已连接"},
		{"value": "DISCONNECTED", "label": "已断开"},
		{"value": "RECONNECTING", "label": "重连中"},
	}
}

// CheckClientNameExists 检查客户端名称是否存在
func (dao *TunnelClientDAO) CheckClientNameExists(clientName, excludeId string) (bool, error) {
	return false, nil
}

// GetTunnelClientsByServerId 根据服务器ID获取客户端列表
func (dao *TunnelClientDAO) GetTunnelClientsByServerId(tunnelServerId string) ([]*models.TunnelClient, error) {
	clients, _, err := dao.QueryTunnelClients(&models.TunnelClientQueryRequest{
		TunnelServerId: tunnelServerId,
	})
	return clients, err
}
