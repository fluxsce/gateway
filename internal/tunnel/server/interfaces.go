// Package server 定义隧道服务端核心接口
// 基于FRP架构，实现控制端口和数据端口分离
package server

import (
	"context"
	"net"
	"time"

	"gateway/internal/tunnel/types"
)

// TunnelServer 隧道服务器接口
type TunnelServer interface {
	// Start 启动隧道服务器
	Start(ctx context.Context) error

	// Stop 停止隧道服务器
	Stop(ctx context.Context) error

	// GetStatus 获取服务器状态
	GetStatus() ServerStatus

	// GetConfig 获取服务器配置
	GetConfig() *types.TunnelServer

	// RegisterClient 注册客户端
	RegisterClient(ctx context.Context, client *types.TunnelClient) error

	// UnregisterClient 注销客户端
	UnregisterClient(ctx context.Context, clientID string) error

	// GetConnectedClients 获取已连接的客户端
	GetConnectedClients() []*types.TunnelClient

	// BroadcastMessage 广播消息给所有客户端
	BroadcastMessage(ctx context.Context, message []byte) error

	// GetServiceRegistry 获取服务注册器
	GetServiceRegistry() ServiceRegistry
}

// ControlServer 控制服务器接口（处理客户端控制连接）
type ControlServer interface {
	// Start 启动控制服务器
	Start(ctx context.Context, address string, port int) error

	// Stop 停止控制服务器
	Stop(ctx context.Context) error

	// HandleConnection 处理客户端连接
	HandleConnection(ctx context.Context, conn net.Conn) error

	// SendMessageToClient 向指定客户端发送消息
	SendMessageToClient(clientID string, message *types.ControlMessage) error

	// GetActiveConnections 获取活跃连接数
	GetActiveConnections() int

	// SetProxyServer 设置代理服务器引用，用于处理数据连接
	SetProxyServer(proxyServer ProxyServer)
}

// ProxyServer 反向代理服务器接口（处理隧道数据转发）
type ProxyServer interface {
	// StartProxy 启动指定服务的反向代理
	StartProxy(ctx context.Context, config *ProxyConfig) error

	// StopProxy 停止指定的反向代理服务
	StopProxy(ctx context.Context, proxyID string) error

	// Stop 停止反向代理服务器
	Stop(ctx context.Context) error

	// GetActiveProxies 获取活跃的反向代理服务
	GetActiveProxies() []*ProxyInfo

	// HandleProxyConnection 处理反向代理连接
	HandleProxyConnection(ctx context.Context, conn net.Conn, proxyID string) error

	// HandleClientDataConnection 处理客户端数据连接
	HandleClientDataConnection(ctx context.Context, conn net.Conn, connectionID string, clientID string) error
}

// ServiceRegistry 服务注册器接口
type ServiceRegistry interface {
	// RegisterService 注册服务
	RegisterService(ctx context.Context, clientID string, service *types.TunnelService) error

	// UnregisterService 注销服务
	UnregisterService(ctx context.Context, serviceID string) error

	// GetService 获取服务
	GetService(ctx context.Context, serviceID string) (*types.TunnelService, error)

	// GetServicesByClient 获取客户端的所有服务
	GetServicesByClient(ctx context.Context, clientID string) ([]*types.TunnelService, error)

	// AllocatePort 分配端口
	AllocatePort(ctx context.Context, serviceType string, preferPort *int) (int, error)

	// ReleasePort 释放端口
	ReleasePort(ctx context.Context, port int) error

	// ValidateServiceConfig 验证服务配置
	ValidateServiceConfig(ctx context.Context, service *types.TunnelService) error
}

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	// SelectNode 选择最优节点
	SelectNode(ctx context.Context, nodes []*types.TunnelServerNode) (*types.TunnelServerNode, error)

	// UpdateNodeStats 更新节点统计信息
	UpdateNodeStats(ctx context.Context, nodeID string, stats *NodeStats) error

	// GetAlgorithm 获取负载均衡算法
	GetAlgorithm() string
}

// 数据结构定义

// ServerStatus 服务器状态
type ServerStatus struct {
	Status           string    `json:"status"`
	StartTime        time.Time `json:"startTime"`
	Uptime           int64     `json:"uptime"`
	ConnectedClients int       `json:"connectedClients"`
	TotalTraffic     int64     `json:"totalTraffic"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	ProxyID       string                 `json:"proxyId"`
	ProxyType     string                 `json:"proxyType"`
	ListenAddress string                 `json:"listenAddress"`
	ListenPort    int                    `json:"listenPort"`
	TargetAddress string                 `json:"targetAddress"`
	TargetPort    int                    `json:"targetPort"`
	Options       map[string]interface{} `json:"options"`
}

// ProxyInfo 代理信息
type ProxyInfo struct {
	ProxyID           string    `json:"proxyId"`
	ProxyType         string    `json:"proxyType"`
	ListenAddress     string    `json:"listenAddress"`
	ListenPort        int       `json:"listenPort"`
	Status            string    `json:"status"`
	StartTime         time.Time `json:"startTime"`
	ActiveConnections int       `json:"activeConnections"`
	TotalConnections  int64     `json:"totalConnections"`
	TotalTraffic      int64     `json:"totalTraffic"`
}

// NodeStats 节点统计
type NodeStats struct {
	ConnectionCount int     `json:"connectionCount"`
	TotalTraffic    int64   `json:"totalTraffic"`
	AverageLatency  float64 `json:"averageLatency"`
	ErrorRate       float64 `json:"errorRate"`
	CpuUsage        float64 `json:"cpuUsage"`
	MemoryUsage     float64 `json:"memoryUsage"`
	LoadAverage     float64 `json:"loadAverage"`
}

// 消息类型定义

// ControlMessage 控制消息
type ControlMessage struct {
	Type      string                 `json:"type"`
	SessionID string                 `json:"sessionId"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// 消息类型常量
const (
	// 控制消息类型
	MessageTypeAuth              = "auth"
	MessageTypeHeartbeat         = "heartbeat"
	MessageTypeRegisterService   = "register_service"
	MessageTypeUnregisterService = "unregister_service"
	MessageTypeNewProxy          = "new_proxy"
	MessageTypeCloseProxy        = "close_proxy"
	MessageTypeError             = "error"
	MessageTypeResponse          = "response"

	// 代理类型
	ProxyTypeTCP   = "tcp"
	ProxyTypeUDP   = "udp"
	ProxyTypeHTTP  = "http"
	ProxyTypeHTTPS = "https"
	ProxyTypeSTCP  = "stcp"
	ProxyTypeSUDP  = "sudp"
	ProxyTypeXTCP  = "xtcp"
)
