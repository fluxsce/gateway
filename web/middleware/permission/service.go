package permission

import (
	"context"
	"fmt"
	"strings"

	"gateway/pkg/database"
)

// RoleCodeSuperAdmin 内置超级管理员角色编码，授权不可被自身卸掉。
const RoleCodeSuperAdmin = "ROLE_SUPER_ADMIN"

// PermissionService 权限服务
type PermissionService struct {
	dao   *PermissionDAOExtended
	cache *PermissionCache
}

// NewPermissionService 创建权限服务
// 参数:
//
//	db: 数据库连接实例
//
// 返回:
//
//	*PermissionService: 权限服务实例
func NewPermissionService(db database.Database) *PermissionService {
	return &PermissionService{
		dao: NewPermissionDAOExtended(db),
		// 暂不启用缓存：多实例下失效不一致，鉴权直接查库。
		cache: nil,
	}
}

// CheckPermission 检查用户权限，这是唯一的权限校验方法，默认必须进行用户权限校验
// 参数:
//
//	ctx: 上下文对象
//	req: 权限检查请求，包含用户ID、租户ID和各种权限检查类型
//
// 返回:
//
//	*PermissionCheckResponse: 权限检查响应，包含检查结果、数据权限范围和详细信息
//	error: 错误信息，成功时为nil
func (ps *PermissionService) CheckPermission(ctx context.Context, req *PermissionCheckRequest) (*PermissionCheckResponse, error) {
	// 验证请求参数
	if err := ps.validateRequest(req); err != nil {
		return &PermissionCheckResponse{
			HasPermission: false,
			Message:       fmt.Sprintf("参数验证失败: %v", err),
		}, nil
	}

	// 执行权限检查
	return ps.dao.CheckComplexPermission(ctx, req)
}

// HasModuleAccess 检查用户是否拥有指定 MODULE，或任一该模块下的按钮（如 hub0020:search）。
// 目录里还没有该 MODULE 时视为未纳入权限树，只要求已登录（放行）。
// 未授予模块且没有任何子资源时拒绝。
func (ps *PermissionService) HasModuleAccess(ctx context.Context, userId, tenantId, resourceCode string) (bool, error) {
	exists, err := ps.dao.ModuleResourceExists(ctx, tenantId, resourceCode)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}
	codes, err := ps.loadUserResourceCodes(ctx, userId, tenantId)
	if err != nil {
		return false, err
	}
	if _, ok := codes[resourceCode]; ok {
		return true, nil
	}
	prefix := resourceCode + ":"
	for code := range codes {
		if strings.HasPrefix(code, prefix) {
			return true, nil
		}
	}
	return false, nil
}

// HasAnyButtonAccess 检查用户是否拥有候选按钮码中的任意一个。
// 路由已声明 RequireButton 时：目录无此 BUTTON 或用户未授予，一律拒绝，避免拼错码被当成未配置而放行。
func (ps *PermissionService) HasAnyButtonAccess(ctx context.Context, userId, tenantId string, buttonCodes []string) (bool, error) {
	if len(buttonCodes) == 0 {
		return true, nil
	}

	existing, err := ps.dao.FilterExistingButtonCodes(ctx, tenantId, buttonCodes)
	if err != nil {
		return false, err
	}
	// 声明了按钮但资源表对不上，按无权限处理
	if len(existing) == 0 {
		return false, nil
	}

	codes, err := ps.loadUserResourceCodes(ctx, userId, tenantId)
	if err != nil {
		return false, err
	}
	for _, code := range existing {
		if _, ok := codes[code]; ok {
			return true, nil
		}
	}
	return false, nil
}

// InvalidateUserCache 删除指定用户的权限缓存。
func (ps *PermissionService) InvalidateUserCache(ctx context.Context, userId, tenantId string) {
	if ps == nil || ps.cache == nil {
		return
	}
	ps.cache.DeleteUserCodes(ctx, tenantId, userId)
}

// InvalidateUsersCache 批量删除用户权限缓存。
func (ps *PermissionService) InvalidateUsersCache(ctx context.Context, userIds []string, tenantId string) {
	if ps == nil || ps.cache == nil {
		return
	}
	ps.cache.DeleteUsersCodes(ctx, tenantId, userIds)
}

// loadUserResourceCodes 读取用户已授权资源编码。
// 请求 context 上挂了码袋时，同一次请求只查库一次。
func (ps *PermissionService) loadUserResourceCodes(ctx context.Context, userId, tenantId string) (map[string]struct{}, error) {
	if bag, ok := userResourceCodesBagFrom(ctx); ok {
		return bag.getOrLoad(userId, tenantId, func() (map[string]struct{}, error) {
			return ps.queryUserResourceCodes(ctx, userId, tenantId)
		})
	}
	return ps.queryUserResourceCodes(ctx, userId, tenantId)
}

// queryUserResourceCodes 从数据库读取用户已授权资源编码并转为集合。
func (ps *PermissionService) queryUserResourceCodes(ctx context.Context, userId, tenantId string) (map[string]struct{}, error) {
	list, err := ps.dao.ListUserResourceCodes(ctx, userId, tenantId)
	if err != nil {
		return nil, err
	}
	return toCodeSet(list), nil
}

func toCodeSet(codes []string) map[string]struct{} {
	set := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		if code != "" {
			set[code] = struct{}{}
		}
	}
	return set
}

// validateRequest 验证权限检查请求参数的合法性
func (ps *PermissionService) validateRequest(req *PermissionCheckRequest) error {
	if req.UserId == "" {
		return fmt.Errorf("用户ID不能为空")
	}
	if req.TenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}

	// 至少需要提供一种权限检查类型
	hasCheckType := req.ModuleCode != "" || req.ResourceCode != "" ||
		req.ButtonCode != "" || (req.ResourcePath != "" && req.Method != "")

	if !hasCheckType {
		return fmt.Errorf("至少需要提供一种权限检查类型：模块代码、资源代码、按钮代码或资源路径+方法")
	}

	return nil
}
