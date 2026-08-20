package permission

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gateway/pkg/cache"
	"gateway/pkg/logger"
)

// PermissionCache 用户权限码缓存。
// 以 tenantId+userId 为键存放资源编码列表，供接口鉴权避免每次打库。
type PermissionCache struct {
	store      cache.Cache
	keyPrefix  string
	defaultTTL time.Duration
}

// NewPermissionCache 使用全局缓存实例创建权限缓存。
// store 为 nil 时所有读写变为空操作，鉴权回源数据库。
func NewPermissionCache(store cache.Cache) *PermissionCache {
	return &PermissionCache{
		store:      store,
		keyPrefix:  "permission:codes:",
		defaultTTL: 30 * time.Minute,
	}
}

// NewPermissionCacheFromGlobal 从默认缓存实例创建权限缓存。
func NewPermissionCacheFromGlobal() *PermissionCache {
	return NewPermissionCache(cache.GetDefaultCache())
}

func (pc *PermissionCache) userCodesKey(tenantId, userId string) string {
	return fmt.Sprintf("%s%s:%s", pc.keyPrefix, tenantId, userId)
}

// GetUserCodes 读取用户已授权的资源编码。
// 第二个返回值 false 表示未命中或缓存不可用，调用方应查库。
func (pc *PermissionCache) GetUserCodes(ctx context.Context, tenantId, userId string) ([]string, bool) {
	if pc == nil || pc.store == nil || tenantId == "" || userId == "" {
		return nil, false
	}
	raw, err := pc.store.Get(ctx, pc.userCodesKey(tenantId, userId))
	if err != nil || len(raw) == 0 {
		return nil, false
	}
	var codes []string
	if err := json.Unmarshal(raw, &codes); err != nil {
		logger.Error("解析用户权限缓存失败", "error", err, "tenantId", tenantId, "userId", userId)
		return nil, false
	}
	return codes, true
}

// SetUserCodes 写入用户权限编码列表。
func (pc *PermissionCache) SetUserCodes(ctx context.Context, tenantId, userId string, codes []string) {
	if pc == nil || pc.store == nil || tenantId == "" || userId == "" {
		return
	}
	if codes == nil {
		codes = []string{}
	}
	raw, err := json.Marshal(codes)
	if err != nil {
		logger.Error("序列化用户权限缓存失败", "error", err, "tenantId", tenantId, "userId", userId)
		return
	}
	if err := pc.store.Set(ctx, pc.userCodesKey(tenantId, userId), raw, pc.defaultTTL); err != nil {
		logger.Error("写入用户权限缓存失败", "error", err, "tenantId", tenantId, "userId", userId)
	}
}

// DeleteUserCodes 删除指定用户的权限缓存，下次鉴权回源数据库。
func (pc *PermissionCache) DeleteUserCodes(ctx context.Context, tenantId, userId string) {
	if pc == nil || pc.store == nil || tenantId == "" || userId == "" {
		return
	}
	if err := pc.store.Delete(ctx, pc.userCodesKey(tenantId, userId)); err != nil {
		logger.Error("删除用户权限缓存失败", "error", err, "tenantId", tenantId, "userId", userId)
	}
}

// DeleteUsersCodes 批量删除用户权限缓存。
func (pc *PermissionCache) DeleteUsersCodes(ctx context.Context, tenantId string, userIds []string) {
	for _, userId := range userIds {
		pc.DeleteUserCodes(ctx, tenantId, userId)
	}
}
