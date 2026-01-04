// Package static 实现静态代理核心功能
// 负责单个静态服务器的端口转发和负载均衡
package static

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// staticProxy 静态代理实现
// 管理一个静态服务器的端口转发和多个后端节点
type staticProxy struct {
	// 配置信息
	server           *types.TunnelStaticServer                 // 静态服务器配置（包含节点列表）
	loadBalancer     LoadBalancer                              // 负载均衡器
	staticServerRepo *storage.TunnelStaticServerRepositoryImpl // 静态服务器存储接口
	staticNodeRepo   *storage.TunnelStaticNodeRepositoryImpl   // 静态节点存储接口

	// 网络监听
	listener net.Listener // 网络监听器

	// 互斥锁
	mutex sync.RWMutex

	// 上下文控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// newStaticProxy 创建静态代理实例
//
// 参数:
//   - server: 静态服务器配置（包含节点列表）
//   - staticServerRepo: 静态服务器存储接口
//   - staticNodeRepo: 静态节点存储接口
//
// 返回:
//   - *staticProxy: 静态代理实例
func newStaticProxy(server *types.TunnelStaticServer, staticServerRepo *storage.TunnelStaticServerRepositoryImpl, staticNodeRepo *storage.TunnelStaticNodeRepositoryImpl) *staticProxy {
	// 根据服务器配置的负载均衡类型创建负载均衡器
	algorithm := "round_robin" // 默认算法
	if server.LoadBalanceType != nil && *server.LoadBalanceType != "" {
		algorithm = *server.LoadBalanceType
	}
	loadBalancer := createLoadBalancer(algorithm)

	// 注意：ctx 和 cancel 在 Start() 时创建，确保每次启动都是新的 context
	return &staticProxy{
		server:           server,
		loadBalancer:     loadBalancer,
		staticServerRepo: staticServerRepo,
		staticNodeRepo:   staticNodeRepo,
	}
}

// Start 启动静态代理
// 创建监听器并开始接受连接
//
// 参数:
//   - ctx: 用于数据库操作的上下文（不影响代理生命周期）
func (p *staticProxy) Start(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.server.ServerStatus == "running" {
		return fmt.Errorf("static proxy already running")
	}

	// 检查是否有可用节点
	if p.server.Nodes == nil || len(p.server.Nodes) == 0 {
		return fmt.Errorf("no nodes available for static proxy")
	}

	// 重新创建内部 context（确保之前的 cancel 不会影响新启动）
	p.ctx, p.cancel = context.WithCancel(context.Background())

	// 创建监听器
	address := net.JoinHostPort(p.server.ListenAddress, fmt.Sprintf("%d", p.server.ListenPort))
	listener, err := net.Listen(p.server.ServerType, address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	p.listener = listener
	now := time.Now()
	p.server.ServerStatus = "running"
	p.server.StartTime = &now

	// 更新数据库（使用外部 ctx，允许调用方控制超时）
	if err := p.staticServerRepo.Update(ctx, p.server); err != nil {
		logger.Error("Failed to update server status in database", map[string]interface{}{
			"serverID": p.server.TunnelStaticServerId,
			"error":    err.Error(),
		})
		// 继续启动流程，不因数据库更新失败而中断
	}

	// 启动健康检查（使用内部 p.ctx，跟随代理生命周期）
	p.startHealthCheck(p.ctx)

	// 启动监听 goroutine（使用内部 p.ctx，跟随代理生命周期）
	p.wg.Add(1)
	go p.acceptConnections(p.ctx)

	logger.Info("Static proxy started", map[string]interface{}{
		"serverID":      p.server.TunnelStaticServerId,
		"serverName":    p.server.ServerName,
		"listenAddress": address,
		"nodeCount":     len(p.server.Nodes),
	})

	return nil
}

// Stop 停止静态代理
// 关闭监听器并等待所有连接完成
//
// 参数:
//   - ctx: 用于数据库操作的上下文（不影响停止流程）
func (p *staticProxy) Stop(ctx context.Context) error {
	p.mutex.Lock()

	if p.server.ServerStatus == "stopped" {
		p.mutex.Unlock()
		return nil
	}

	p.server.ServerStatus = "stopping"

	// 取消内部 context，通知所有 goroutine 停止
	if p.cancel != nil {
		p.cancel()
	}

	// 关闭监听器
	if p.listener != nil {
		if err := p.listener.Close(); err != nil {
			logger.Error("Failed to close listener", map[string]interface{}{
				"serverID": p.server.TunnelStaticServerId,
				"error":    err.Error(),
			})
		}
	}

	// 释放锁，避免在等待 goroutine 时持有锁
	p.mutex.Unlock()

	// 等待所有 goroutine 完成（不持有锁）
	p.wg.Wait()

	// 重新获取锁，更新状态
	p.mutex.Lock()
	now := time.Now()
	p.server.ServerStatus = "stopped"
	p.server.StopTime = &now
	serverID := p.server.TunnelStaticServerId
	p.mutex.Unlock()

	// 更新数据库（不持有锁，使用外部 ctx，允许调用方控制超时）
	if err := p.staticServerRepo.Update(ctx, p.server); err != nil {
		logger.Error("Failed to update server status in database", map[string]interface{}{
			"serverID": serverID,
			"error":    err.Error(),
		})
		// 继续停止流程，不因数据库更新失败而中断
	}

	logger.Info("Static proxy stopped", map[string]interface{}{
		"serverID": serverID,
	})

	return nil
}

// ReloadConfig 重新加载配置
// 如果监听地址或端口改变，需要重启代理
func (p *staticProxy) ReloadConfig(ctx context.Context, server *types.TunnelStaticServer) error {
	p.mutex.Lock()

	// 检查是否需要重启
	needRestart := false
	if p.server.ServerStatus == "running" {
		oldAddress := net.JoinHostPort(p.server.ListenAddress, fmt.Sprintf("%d", p.server.ListenPort))
		newAddress := net.JoinHostPort(server.ListenAddress, fmt.Sprintf("%d", server.ListenPort))
		needRestart = (oldAddress != newAddress)
	}

	// 更新服务器配置
	p.server = server
	p.mutex.Unlock()

	// 如果需要重启，释放锁后再执行（避免重入死锁）
	if needRestart {
		if err := p.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop proxy for reload: %w", err)
		}
		if err := p.Start(ctx); err != nil {
			return fmt.Errorf("failed to start proxy after reload: %w", err)
		}
	}

	return nil
}
