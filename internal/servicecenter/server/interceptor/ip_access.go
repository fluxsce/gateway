package interceptor

import (
	"context"
	"fmt"
	"net"
	"strings"

	"gateway/internal/servicecenter/types"
	"gateway/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// IPAccessInterceptor IP 访问控制拦截器
// 根据配置的白名单和黑名单检查客户端 IP
type IPAccessInterceptor struct {
	configProvider ConfigProvider
}

// NewIPAccessInterceptor 创建 IP 访问控制拦截器
func NewIPAccessInterceptor(configProvider ConfigProvider) *IPAccessInterceptor {
	return &IPAccessInterceptor{
		configProvider: configProvider,
	}
}

// UnaryServerInterceptor 返回 Unary IP 访问控制拦截器
func (i *IPAccessInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		config := i.configProvider.GetConfig()

		// 如果未启用认证，跳过 IP 检查
		if config.EnableAuth != "Y" {
			return handler(ctx, req)
		}

		// 获取客户端 IP
		clientIP, err := getClientIP(ctx)
		if err != nil {
			logger.Warn("无法获取客户端 IP", "error", err, "method", info.FullMethod)
			return handler(ctx, req) // 无法获取 IP 时，继续处理（宽松策略）
		}

		// 检查 IP 访问权限
		if !i.checkIPAccess(clientIP, config) {
			logger.Warn("IP 访问被拒绝",
				"clientIP", clientIP,
				"method", info.FullMethod,
				"instanceName", config.InstanceName)
			return nil, status.Errorf(codes.PermissionDenied, "访问被拒绝：IP %s 未授权", clientIP)
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor 返回 Stream IP 访问控制拦截器
func (i *IPAccessInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		config := i.configProvider.GetConfig()

		if config.EnableAuth != "Y" {
			return handler(srv, ss)
		}

		clientIP, err := getClientIP(ss.Context())
		if err != nil {
			logger.Warn("无法获取客户端 IP", "error", err, "method", info.FullMethod)
			return handler(srv, ss)
		}

		if !i.checkIPAccess(clientIP, config) {
			logger.Warn("IP 访问被拒绝（流式）",
				"clientIP", clientIP,
				"method", info.FullMethod,
				"instanceName", config.InstanceName)
			return status.Errorf(codes.PermissionDenied, "访问被拒绝：IP %s 未授权", clientIP)
		}

		return handler(srv, ss)
	}
}

// checkIPAccess 检查 IP 是否有访问权限
// 规则：
//  1. 如果在黑名单中，拒绝访问
//  2. 如果配置了白名单，只允许白名单中的 IP 访问
//  3. 如果没有配置白名单，允许所有不在黑名单中的 IP 访问
func (i *IPAccessInterceptor) checkIPAccess(clientIP string, config *types.InstanceConfig) bool {
	// 解析黑名单
	if config.IpBlacklist != "" {
		blacklist := parseIPList(config.IpBlacklist)
		if contains(blacklist, clientIP) {
			return false // 在黑名单中，拒绝访问
		}
	}

	// 解析白名单
	if config.IpWhitelist != "" {
		whitelist := parseIPList(config.IpWhitelist)
		if !contains(whitelist, clientIP) {
			return false // 不在白名单中，拒绝访问
		}
	}

	return true // 通过检查，允许访问
}

// getClientIP 从 context 中获取客户端 IP
func getClientIP(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("无法从 context 获取 peer 信息")
	}

	// 解析 IP 地址
	addr := p.Addr.String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// 如果没有端口，直接返回地址
		return addr, nil
	}
	return host, nil
}

// parseIPList 解析 IP 列表（JSON 数组或逗号分隔）
func parseIPList(ipListStr string) []string {
	// 简单实现：按逗号分隔
	// TODO: 支持 JSON 数组格式
	// TODO: 支持 CIDR 格式（如 192.168.1.0/24）
	if ipListStr == "" {
		return []string{}
	}
	ips := strings.Split(ipListStr, ",")
	result := make([]string, 0, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			result = append(result, ip)
		}
	}
	return result
}

// contains 检查字符串是否在列表中
func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

// ================================================================================
// TODO: 扩展 IP 访问控制功能
// ================================================================================

// CIDRMatcher CIDR 格式匹配器（待实现）
type CIDRMatcher struct {
	networks []*net.IPNet
}

// MatchCIDR 检查 IP 是否匹配 CIDR 规则（待实现）
func (c *CIDRMatcher) MatchCIDR(ip string) bool {
	// TODO: 实现 CIDR 匹配
	// 1. 解析 IP 地址
	// 2. 遍历所有 CIDR 规则
	// 3. 检查 IP 是否在任何 CIDR 范围内
	logger.Debug("CIDR 匹配待实现", "ip", ip)
	return false
}

// IPRateLimiter IP 级别的限流器（待实现）
type IPRateLimiter struct {
	limits map[string]int // IP -> 每秒请求数
}

// CheckRateLimit 检查 IP 是否超过限流（待实现）
func (r *IPRateLimiter) CheckRateLimit(ip string) (bool, error) {
	// TODO: 实现 IP 级别限流
	// 1. 记录每个 IP 的请求次数
	// 2. 检查是否超过限流阈值
	// 3. 支持滑动窗口或令牌桶算法
	logger.Debug("IP 限流待实现", "ip", ip)
	return true, nil
}
