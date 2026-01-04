// Package static 实现连接处理逻辑
// 负责接受连接、选择节点和转发数据
package static

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// acceptConnections 接受连接
// 持续监听并接受新的客户端连接
//
// 参数:
//   - ctx: 代理内部 context，用于控制生命周期
func (p *staticProxy) acceptConnections(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 设置接受超时，避免阻塞
			if tcpListener, ok := p.listener.(*net.TCPListener); ok {
				tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
			}

			conn, err := p.listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				if p.server.ServerStatus == "stopping" || p.server.ServerStatus == "stopped" {
					return
				}
				logger.Error("Failed to accept connection", map[string]interface{}{
					"serverID": p.server.TunnelStaticServerId,
					"error":    err.Error(),
				})
				continue
			}

			// 确保连接在出错时被关闭
			connToHandle := conn

			// 处理连接
			p.wg.Add(1)
			go func(conn net.Conn) {
				defer p.wg.Done()
				defer conn.Close() // 确保连接在处理完成后被关闭
				if err := p.HandleConnection(ctx, conn); err != nil {
					logger.Error("Failed to handle connection", map[string]interface{}{
						"serverID": p.server.TunnelStaticServerId,
						"error":    err.Error(),
					})
				}
			}(connToHandle)
		}
	}
}

// HandleConnection 处理连接
// 选择节点并转发数据
func (p *staticProxy) HandleConnection(ctx context.Context, conn net.Conn) error {
	// 使用原子操作更新服务器连接计数
	// 注意：CurrentConnectionCount 是 int 类型，需要使用 mutex 保护
	// TotalConnectionCount 是 int64 类型，可以使用原子操作
	p.mutex.Lock()
	p.server.CurrentConnectionCount++
	p.mutex.Unlock()
	atomic.AddInt64(&p.server.TotalConnectionCount, 1)

	defer func() {
		p.mutex.Lock()
		p.server.CurrentConnectionCount--
		p.mutex.Unlock()
	}()

	// 选择节点（使用负载均衡）
	node, err := p.selectNode(ctx)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to select node: %w", err)
	}

	// 转发连接
	return p.forwardConnection(ctx, conn, node)
}

// selectNode 选择节点（使用负载均衡）
// 从健康的活跃节点中选择一个最优节点
func (p *staticProxy) selectNode(ctx context.Context) (*types.TunnelStaticNode, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// 直接使用负载均衡器选择节点，负载均衡器内部会处理节点过滤
	return p.loadBalancer.SelectNode(ctx, p.server)
}

// forwardConnection 转发连接
// 建立到目标节点的连接并双向转发数据
// 支持 SSH 等长连接场景，使用 TCP 半关闭正确处理连接终止
func (p *staticProxy) forwardConnection(ctx context.Context, clientConn net.Conn, node *types.TunnelStaticNode) error {
	// 连接到目标节点
	targetAddr := net.JoinHostPort(node.TargetAddress, fmt.Sprintf("%d", node.TargetPort))
	connectionTimeout := time.Duration(p.server.ConnectionTimeout) * time.Second
	targetConn, err := net.DialTimeout(node.ProxyType, targetAddr, connectionTimeout)
	if err != nil {
		// 更新节点失败统计
		// FailureCount 是 int 类型，LastFailureTime 是指针类型，都需要 mutex 保护
		p.mutex.Lock()
		node.FailureCount++
		now := time.Now()
		node.LastFailureTime = &now
		p.mutex.Unlock()
		return fmt.Errorf("failed to connect to target %s: %w", targetAddr, err)
	}

	// 确保在函数退出时关闭连接
	defer targetConn.Close()

	// 更新节点连接计数
	// CurrentConnectionCount 是 int 类型，需要 mutex 保护
	// TotalConnectionCount 是 int64 类型，可以使用原子操作
	p.mutex.Lock()
	node.CurrentConnectionCount++
	p.mutex.Unlock()
	atomic.AddInt64(&node.TotalConnectionCount, 1)

	defer func() {
		p.mutex.Lock()
		node.CurrentConnectionCount--
		p.mutex.Unlock()
	}()

	// 双向转发数据，使用带方向标识的结构
	type copyResult struct {
		direction string // "c2t" = client to target, "t2c" = target to client
		bytes     int64
		err       error
	}
	resultChan := make(chan copyResult, 2)

	// 客户端 -> 目标（接收数据）
	// 对于 SSH 等长连接，io.Copy 会持续转发直到客户端关闭连接
	go func() {
		bytes, err := io.Copy(targetConn, clientConn)
		// 客户端发送完毕，使用半关闭通知目标服务器
		// 这对 SSH 是安全的：SSH 服务器收到 FIN 后会正常处理
		if tcpConn, ok := targetConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
		resultChan <- copyResult{direction: "c2t", bytes: bytes, err: err}
	}()

	// 目标 -> 客户端（发送数据）
	// 对于 SSH 等长连接，io.Copy 会持续转发直到服务器关闭连接
	go func() {
		bytes, err := io.Copy(clientConn, targetConn)
		// 目标响应完毕，使用半关闭通知客户端
		if tcpConn, ok := clientConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
		resultChan <- copyResult{direction: "t2c", bytes: bytes, err: err}
	}()

	// 等待两个方向都完成或 context 取消
	// 对于 SSH 长连接，会一直等待直到用户断开或服务停止
	var firstErr error
	completed := 0

	for completed < 2 {
		select {
		case <-ctx.Done():
			// context 取消时（如服务停止），强制关闭连接以中断 io.Copy
			// 对于 SSH 连接，这会导致客户端收到连接重置
			clientConn.Close()
			targetConn.Close()
			// 继续等待 goroutine 完成，确保资源正确释放
			for completed < 2 {
				<-resultChan
				completed++
			}
			return ctx.Err()

		case result := <-resultChan:
			completed++
			// 使用原子操作更新流量统计
			if result.direction == "c2t" {
				atomic.AddInt64(&p.server.TotalBytesReceived, result.bytes)
			} else {
				atomic.AddInt64(&p.server.TotalBytesSent, result.bytes)
			}

			// 记录第一个非 EOF 错误
			// 注意：对于正常关闭的 SSH 连接，err 通常是 nil 或 io.EOF
			if result.err != nil && result.err != io.EOF && firstErr == nil {
				firstErr = result.err
			}
		}
	}

	return firstErr
}
