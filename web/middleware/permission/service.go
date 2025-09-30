package permission

import (
	"context"
	"fmt"
	"gateway/pkg/database"
)

// PermissionService 权限服务
type PermissionService struct {
	dao *PermissionDAOExtended
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
