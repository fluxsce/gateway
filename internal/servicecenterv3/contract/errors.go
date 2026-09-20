package contract

import "errors"

// 哨兵错误是进程内稳定面的一部分，调用方必须用 errors.Is 判断。
// 文案可以本地化或加包装前缀，标识码（CENTER_NOT_FOUND 等）不要改。
var (
	// ErrInvalidContext 表示 TenantID / CenterInstanceName / NamespaceID 缺失。
	ErrInvalidContext = errors.New("servicecenterv3: tenantId 或 centerInstanceName 缺失")
	// ErrCenterNotFound 表示中心实例不存在或未加载。
	ErrCenterNotFound = errors.New("servicecenterv3: CENTER_NOT_FOUND")
	// ErrCenterNotRunning 表示实例已配置但未监听。
	ErrCenterNotRunning = errors.New("servicecenterv3: CENTER_NOT_RUNNING")
	// ErrCenterAlreadyExists 表示同名中心实例已存在。
	ErrCenterAlreadyExists = errors.New("servicecenterv3: CENTER_ALREADY_EXISTS")
	// ErrNamespaceNotFound 表示命名空间不存在或不属于当前中心实例。
	ErrNamespaceNotFound = errors.New("servicecenterv3: NAMESPACE_NOT_FOUND")
	// ErrNamespaceDisabled 表示命名空间已停用。
	ErrNamespaceDisabled = errors.New("servicecenterv3: NAMESPACE_DISABLED")
	// ErrServiceNotFound 表示服务定义不存在。
	ErrServiceNotFound = errors.New("servicecenterv3: SERVICE_NOT_FOUND")
	// ErrNodeNotFound 表示服务节点不存在。
	ErrNodeNotFound = errors.New("servicecenterv3: NODE_NOT_FOUND")
	// ErrConfigNotFound 表示草稿或已发布配置不存在。
	ErrConfigNotFound = errors.New("servicecenterv3: CONFIG_NOT_FOUND")
	// ErrReleaseNotFound 表示指定版本的发布记录不存在。
	ErrReleaseNotFound = errors.New("servicecenterv3: RELEASE_NOT_FOUND")
	// ErrInvalidArgument 表示参数不合法（空名、非法端口等）。
	ErrInvalidArgument = errors.New("servicecenterv3: INVALID_ARGUMENT")
	// ErrNamespaceForbidden 表示令牌未授权该命名空间。
	ErrNamespaceForbidden = errors.New("servicecenterv3: NAMESPACE_FORBIDDEN")
	// ErrQuotaExceeded 表示命名空间服务或配置配额已满。
	ErrQuotaExceeded = errors.New("servicecenterv3: QUOTA_EXCEEDED")
	// ErrTenantMismatch 表示令牌租户与中心实例租户不一致。
	ErrTenantMismatch = errors.New("servicecenterv3: TENANT_MISMATCH")
	// ErrViewNotReady 表示命名视图还在从库或副本追齐，当前空名单不能当「没有节点」。
	ErrViewNotReady = errors.New("servicecenterv3: VIEW_NOT_READY")
	// ErrNodeExists 表示客户端自带的 nodeId 已被另一个仍存活的实例占用。
	ErrNodeExists = errors.New("servicecenterv3: NODE_EXISTS")
	// ErrLiveUnavailable 表示集群活视图写入失败，调用方可重试注销。
	ErrLiveUnavailable = errors.New("servicecenterv3: LIVE_UNAVAILABLE")
)

// IsNotFound 判断 err 是否为「资源不存在」类哨兵（中心/命名空间/服务/节点/配置/发布）。
func IsNotFound(err error) bool {
	return errors.Is(err, ErrCenterNotFound) ||
		errors.Is(err, ErrNamespaceNotFound) ||
		errors.Is(err, ErrServiceNotFound) ||
		errors.Is(err, ErrNodeNotFound) ||
		errors.Is(err, ErrConfigNotFound) ||
		errors.Is(err, ErrReleaseNotFound)
}

// IsInvalid 判断 err 是否为上下文或参数不合法。
func IsInvalid(err error) bool {
	return errors.Is(err, ErrInvalidContext) || errors.Is(err, ErrInvalidArgument)
}
