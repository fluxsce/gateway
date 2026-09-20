package stream

import (
	"context"
	"encoding/base64"
	"net"
	"strings"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/store"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// chainInterceptors 组装 panic 恢复、IP 名单与鉴权。
// IP 名单只要配置了就生效，不依赖 EnableAuth：它是实例级网络边界。
// EnableAuth 为 Y 时写入租户；关闭时用实例自身 TenantID。令牌租户必须等于实例租户。
func chainInterceptors(cfg *model.CenterInstance, st *store.Store) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("servicecenterv3 stream panic", nil, "recover", r)
				err = status.Error(codes.Internal, "internal error")
			}
		}()
		ctx := ss.Context()
		if cfg != nil {
			if ipErr := checkIPAccess(ctx, cfg); ipErr != nil {
				return ipErr
			}
		}
		if cfg != nil && model.IsY(cfg.EnableAuth) && st != nil {
			authenticated, aerr := authenticate(ctx, cfg, st)
			if aerr != nil {
				return aerr
			}
			ctx = authenticated
		} else if cfg != nil {
			cc := contract.CallContext{
				TenantID:           cfg.TenantID,
				CenterInstanceName: cfg.InstanceName,
				Environment:        cfg.Environment,
			}
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				if vals := md.Get("x-operator-id"); len(vals) > 0 {
					cc.OperatorID = vals[0]
				}
			}
			ctx = contract.WithCallContext(ctx, cc)
		}
		return handler(srv, &wrappedStream{ServerStream: ss, ctx: ctx})
	}
}

// wrappedStream 用鉴权后的 ctx 覆盖原始流 Context。
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context 返回写入 CallContext 后的上下文。
func (w *wrappedStream) Context() context.Context { return w.ctx }

// checkIPAccess 先黑名单后白名单。名单空则不限制。支持单 IP 与 CIDR。
func checkIPAccess(ctx context.Context, cfg *model.CenterInstance) error {
	if cfg.IPBlacklist == "" && cfg.IPWhitelist == "" {
		return nil
	}
	clientIP := clientIPFromCtx(ctx)
	if clientIP == "" {
		logger.Warn("servicecenterv3 无法解析客户端 IP，跳过 IP 名单", "instance", cfg.InstanceName)
		return nil
	}
	if matchIPList(clientIP, cfg.IPBlacklist) {
		logger.Warn("servicecenterv3 IP 命中黑名单", "ip", clientIP, "instance", cfg.InstanceName)
		return status.Errorf(codes.PermissionDenied, "IP %s is not allowed", clientIP)
	}
	if cfg.IPWhitelist != "" && !matchIPList(clientIP, cfg.IPWhitelist) {
		logger.Warn("servicecenterv3 IP 不在白名单", "ip", clientIP, "instance", cfg.InstanceName)
		return status.Errorf(codes.PermissionDenied, "IP %s is not allowed", clientIP)
	}
	return nil
}

func clientIPFromCtx(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p == nil || p.Addr == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(p.Addr.String())
	if err != nil {
		return p.Addr.String()
	}
	return host
}

// matchIPList 解析逗号或 JSON 数组，条目可以是 IP 或 CIDR。
func matchIPList(ip, raw string) bool {
	items := model.SplitCSV(raw)
	parsed := net.ParseIP(ip)
	for _, item := range items {
		if item == ip {
			return true
		}
		if parsed != nil && strings.Contains(item, "/") {
			if _, network, err := net.ParseCIDR(item); err == nil && network.Contains(parsed) {
				return true
			}
		}
	}
	return false
}

// authenticate 解析 authorization：Basic 或服务端颁发的 Bearer 令牌，写入 CallContext。
func authenticate(ctx context.Context, cfg *model.CenterInstance, st *store.Store) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing authentication")
	}
	headers := md.Get("authorization")
	if len(headers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authentication token")
	}
	header := headers[0]
	var tenantID, operatorID string
	var allowedNS []string
	switch {
	case strings.HasPrefix(header, "Basic "):
		user, err := authBasic(ctx, st, header)
		if err != nil {
			return nil, err
		}
		tenantID, operatorID = user.TenantId, user.UserId
	case strings.HasPrefix(header, "Bearer "):
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		tk, err := st.Auth.ValidateToken(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid authentication token")
		}
		if bindErr := bindTokenToInstance(cfg, tk); bindErr != nil {
			return nil, bindErr
		}
		tenantID, operatorID = tk.TenantId, tk.UserId
		allowedNS = model.ParseTokenScope(tk.ExtProperty).NamespaceIDs
	default:
		return nil, status.Error(codes.Unauthenticated, "unsupported authentication type")
	}
	if tenantID == "" {
		tenantID = cfg.TenantID
	}
	if cfg.TenantID != "" && tenantID != cfg.TenantID {
		return nil, status.Error(codes.PermissionDenied, contract.ErrTenantMismatch.Error())
	}
	cc := contract.CallContext{
		TenantID:           tenantID,
		CenterInstanceName: cfg.InstanceName,
		Environment:        cfg.Environment,
		OperatorID:         operatorID,
		AllowedNamespaces:  allowedNS,
	}
	return contract.WithCallContext(ctx, cc), nil
}

// bindTokenToInstance 校验不透明令牌是否绑定当前中心实例。
// extProperty.instanceName 为空视为历史令牌（仅租户匹配）；生产令牌应填写实例名。
func bindTokenToInstance(cfg *model.CenterInstance, tk *store.AuthToken) error {
	scope := model.ParseTokenScope(tk.ExtProperty)
	if scope.InstanceName != "" && scope.InstanceName != cfg.InstanceName {
		return status.Error(codes.PermissionDenied, "token is not bound to this center instance")
	}
	if scope.Environment != "" && scope.Environment != cfg.Environment {
		return status.Error(codes.PermissionDenied, "token is not bound to this environment")
	}
	return nil
}

// authBasic 解析 Basic 头并校验用户名密码。
func authBasic(ctx context.Context, st *store.Store, header string) (*store.User, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(header, "Basic "))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication format")
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication format")
	}
	user, err := st.Auth.ValidateUser(ctx, parts[0], parts[1])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user or password")
	}
	return user, nil
}
