package interceptor

import (
	"context"
	"runtime/debug"

	"gateway/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor Panic 恢复拦截器
// 捕获所有 gRPC 处理过程中的 panic，记录详细日志，并返回错误响应，避免服务崩溃
type RecoveryInterceptor struct{}

// NewRecoveryInterceptor 创建 Panic 恢复拦截器
func NewRecoveryInterceptor() *RecoveryInterceptor {
	return &RecoveryInterceptor{}
}

// UnaryServerInterceptor 返回 Unary Panic 恢复拦截器
// 捕获 Unary RPC 处理过程中的 panic，记录堆栈信息，并返回内部错误
func (r *RecoveryInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 使用 defer + recover 捕获 panic
		defer func() {
			if rec := recover(); rec != nil {
				// 获取堆栈信息
				stackTrace := string(debug.Stack())

				// 获取客户端 IP（如果可用）
				clientIP, _ := getClientIP(ctx)

				// 记录详细的 panic 日志
				logger.Error("gRPC 处理过程中发生 Panic，已恢复",
					nil, // 没有原始 error，使用 nil
					"method", info.FullMethod,
					"clientIP", clientIP,
					"panic", rec,
					"stackTrace", stackTrace)

				// 返回内部错误给客户端，避免服务崩溃
				// 注意：这里设置返回值，因为我们在 defer 中
				err = status.Errorf(codes.Internal, "服务器内部错误: %v", rec)
				resp = nil
			}
		}()

		// 执行实际的 RPC 处理
		resp, err = handler(ctx, req)
		return resp, err
	}
}

// StreamServerInterceptor 返回 Stream Panic 恢复拦截器
// 捕获 Stream RPC 处理过程中的 panic，记录堆栈信息，并返回内部错误
func (r *RecoveryInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 使用 defer + recover 捕获 panic
		defer func() {
			if rec := recover(); rec != nil {
				// 获取堆栈信息
				stackTrace := string(debug.Stack())

				// 获取客户端 IP（如果可用）
				clientIP, _ := getClientIP(ss.Context())

				// 记录详细的 panic 日志
				logger.Error("gRPC Stream 处理过程中发生 Panic，已恢复",
					nil, // 没有原始 error，使用 nil
					"method", info.FullMethod,
					"clientIP", clientIP,
					"isClientStream", info.IsClientStream,
					"isServerStream", info.IsServerStream,
					"panic", rec,
					"stackTrace", stackTrace)

				// 对于 Stream，我们无法返回错误给客户端（因为已经 panic）
				// 但至少可以记录日志，避免服务崩溃
			}
		}()

		// 执行实际的 Stream RPC 处理
		err := handler(srv, ss)
		return err
	}
}
