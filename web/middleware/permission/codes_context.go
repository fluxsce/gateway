package permission

import (
	"context"
	"sync"
)

// userResourceCodesBagKey 请求级资源码袋的 context key。
type userResourceCodesBagKey struct{}

// userResourceCodesBag 同一次请求内复用 ListUserResourceCodes 的结果。
// 只挂在该请求的 context 上，请求结束即释放，不跨请求、不跨实例。
type userResourceCodesBag struct {
	once     sync.Once
	userId   string
	tenantId string
	codes    map[string]struct{}
	err      error
}

// WithUserResourceCodesBag 在 ctx 上挂一个资源码袋。
// 已挂过则原样返回，便于模块中间件和按钮中间件各调一次。
func WithUserResourceCodesBag(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Value(userResourceCodesBagKey{}).(*userResourceCodesBag); ok {
		return ctx
	}
	return context.WithValue(ctx, userResourceCodesBagKey{}, &userResourceCodesBag{})
}

// userResourceCodesBagFrom 取出请求上的资源码袋，未挂过返回 false。
func userResourceCodesBagFrom(ctx context.Context) (*userResourceCodesBag, bool) {
	if ctx == nil {
		return nil, false
	}
	bag, ok := ctx.Value(userResourceCodesBagKey{}).(*userResourceCodesBag)
	return bag, ok && bag != nil
}

// getOrLoad 首次按 userId+tenantId 查库，同一次请求内后续直接返回。
// 身份与首次不一致时不复用，避免串码。
func (b *userResourceCodesBag) getOrLoad(userId, tenantId string, load func() (map[string]struct{}, error)) (map[string]struct{}, error) {
	if b == nil {
		return load()
	}
	b.once.Do(func() {
		b.userId = userId
		b.tenantId = tenantId
		b.codes, b.err = load()
	})
	if b.userId != userId || b.tenantId != tenantId {
		return load()
	}
	return b.codes, b.err
}
