package contract

import "context"

// ctxKey 避免 context 键与其他包冲突。
type ctxKey int

const callContextKey ctxKey = 1

// CallContext 是 L3 每次调用必带的上下文。
//
// TenantID 只从鉴权（或鉴权关闭时的 CenterInstance.TenantID）写入，
// 客户端载荷不得覆盖。CenterInstanceName + Environment 选择隔离运行时
// （多环境信任边界在实例，不在命名空间）。
// NamespaceID 是 Naming/Config 的工作空间。OperatorID 仅审计。
type CallContext struct {
	// TenantID 租户，来自鉴权，禁止客户端自报。
	TenantID string
	// CenterInstanceName 中心实例名，对应一条监听与一套缓存。
	CenterInstanceName string
	// Environment 环境标签，与 CenterInstanceName 组成运行时键；管理面应传入。
	Environment string
	// NamespaceID 命名空间。Admin 调用可空，Naming/Config 不可空。
	NamespaceID string
	// AllowedNamespaces 令牌绑定的命名空间；空表示该实例下全部。
	AllowedNamespaces []string
	// OperatorID 操作者，写入审计字段，不参与隔离键。
	OperatorID string
	// ConnectionID 数据面连接。订阅与临时实例绑定用，管理面调用可空。
	ConnectionID string
}

// AllowsNamespace 判断 ns 是否在令牌授权范围内。空名单表示不限制。
func (c CallContext) AllowsNamespace(ns string) bool {
	if len(c.AllowedNamespaces) == 0 {
		return true
	}
	if ns == "" {
		return false
	}
	for _, id := range c.AllowedNamespaces {
		if id == ns {
			return true
		}
	}
	return false
}

// RequireCenter 要求 TenantID 与 CenterInstanceName 均非空。
func (c CallContext) RequireCenter() error {
	if c.TenantID == "" || c.CenterInstanceName == "" {
		return ErrInvalidContext
	}
	return nil
}

// RequireNamespace 在 RequireCenter 之上还要求 NamespaceID 非空。
func (c CallContext) RequireNamespace() error {
	if err := c.RequireCenter(); err != nil {
		return err
	}
	if c.NamespaceID == "" {
		return ErrInvalidContext
	}
	return nil
}

// WithCallContext 把 CallContext 放入 ctx，供拦截器与处理函数传递。
func WithCallContext(ctx context.Context, cc CallContext) context.Context {
	return context.WithValue(ctx, callContextKey, cc)
}

// FromContext 读取可选的 CallContext，不校验字段。
func FromContext(ctx context.Context) (CallContext, bool) {
	if ctx == nil {
		return CallContext{}, false
	}
	cc, ok := ctx.Value(callContextKey).(CallContext)
	return cc, ok
}

// MustFromContext 读取 CallContext 并执行 RequireCenter。
func MustFromContext(ctx context.Context) (CallContext, error) {
	cc, ok := FromContext(ctx)
	if !ok {
		return CallContext{}, ErrInvalidContext
	}
	if err := cc.RequireCenter(); err != nil {
		return CallContext{}, err
	}
	return cc, nil
}
