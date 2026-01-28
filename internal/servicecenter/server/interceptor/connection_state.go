package interceptor

import (
	"context"
	"sync"

	"gateway/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// ConnectionStateListener 连接状态监听器接口
// 用于监听连接建立和断开事件（参考 Nacos 2.0 设计）
type ConnectionStateListener interface {
	// OnConnect 连接建立时调用
	OnConnect(clientID string, ctx context.Context)
	// OnDisconnect 连接断开时调用
	OnDisconnect(clientID string)
}

// ConnectionStateInterceptor 连接状态拦截器
// 参考 Nacos 2.0 设计：在 gRPC 层面监听连接建立和断开事件
//
// 工作原理：
//   - Stream RPC（Bidirectional Streaming）：通过 stream.Context() 监听连接状态
//   - 连接建立：在 StreamServerInterceptor 中调用 OnConnect
//   - 连接断开：监听 stream.Context().Done()，调用 OnDisconnect
//   - 这是实时感知连接断开的主要方式（参考 Nacos 2.0）
//   - Unary RPC：无法实时感知连接断开
//   - context 在调用结束后会被取消，不能用来检测连接状态
//   - 建议使用 Bidirectional Streaming RPC（如 ConnectionManagement）进行连接管理和心跳
type ConnectionStateInterceptor struct {
	listener ConnectionStateListener
	// 跟踪已建立的连接（避免重复通知）
	connections sync.Map // map[string]bool
}

// NewConnectionStateInterceptor 创建连接状态拦截器
func NewConnectionStateInterceptor(listener ConnectionStateListener) *ConnectionStateInterceptor {
	return &ConnectionStateInterceptor{
		listener: listener,
	}
}

// 注意：不提供 UnaryServerInterceptor
// Unary RPC 无法实时感知连接断开（context 在调用结束后会被取消）
// 连接状态监听只通过 Stream RPC 实现（参考 Nacos 2.0 设计）
// 如果需要连接管理和心跳，请使用 Bidirectional Streaming RPC（如 ConnectionManagement）

// StreamServerInterceptor 返回 Stream 连接状态拦截器
func (c *ConnectionStateInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		clientID, ok := getClientID(ctx)
		if ok && c.listener != nil {
			// 检查是否是新连接（避免重复通知）
			key := "stream:" + clientID
			if _, exists := c.connections.LoadOrStore(key, true); !exists {
				// 新连接，通知监听器
				c.listener.OnConnect(clientID, ctx)
				logger.Debug("检测到 Stream RPC 连接建立",
					"clientID", clientID,
					"method", info.FullMethod)
			}

			// 启动 goroutine 监听连接断开
			go func() {
				<-ctx.Done()
				// 连接断开，通知监听器
				c.listener.OnDisconnect(clientID)
				// 清理连接记录
				c.connections.Delete(key)
				logger.Info("检测到 Stream RPC 连接断开",
					"clientID", clientID,
					"method", info.FullMethod)
			}()
		}

		// 执行实际的 Stream 处理
		return handler(srv, ss)
	}
}

// getClientID 从 context 中获取客户端标识（IP:port）
func getClientID(ctx context.Context) (string, bool) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", false
	}
	return p.Addr.String(), true
}
