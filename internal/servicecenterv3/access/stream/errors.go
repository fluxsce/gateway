package stream

import (
	"errors"

	"gateway/internal/servicecenterv3/contract"
)

var (
	errHandshakeRequired = errors.New("PROTOCOL_HANDSHAKE_REQUIRED")
	errUnsupported       = errors.New("UNSUPPORTED")
)

// errorCode 把哨兵错误映射为 SDK 可稳定解析的 code，不要匹配文案。
func errorCode(err error) (code, message string) {
	if err == nil {
		return "OK", ""
	}
	message = err.Error()
	switch {
	case errors.Is(err, errHandshakeRequired):
		return "PROTOCOL_HANDSHAKE_REQUIRED", "第一条业务消息必须是 CLIENT_HANDSHAKE"
	case errors.Is(err, errUnsupported):
		return "UNSUPPORTED", message
	case errors.Is(err, contract.ErrInvalidContext):
		return "INVALID_CONTEXT", message
	case errors.Is(err, contract.ErrInvalidArgument):
		return "INVALID_ARGUMENT", message
	case errors.Is(err, contract.ErrNamespaceNotFound):
		return "NAMESPACE_NOT_FOUND", message
	case errors.Is(err, contract.ErrNamespaceDisabled):
		return "NAMESPACE_DISABLED", message
	case errors.Is(err, contract.ErrNamespaceForbidden):
		return "NAMESPACE_FORBIDDEN", message
	case errors.Is(err, contract.ErrQuotaExceeded):
		return "QUOTA_EXCEEDED", message
	case errors.Is(err, contract.ErrTenantMismatch):
		return "TENANT_MISMATCH", message
	case errors.Is(err, contract.ErrServiceNotFound):
		return "SERVICE_NOT_FOUND", message
	case errors.Is(err, contract.ErrNodeNotFound):
		return "NODE_NOT_FOUND", message
	case errors.Is(err, contract.ErrNodeExists):
		return "NODE_EXISTS", message
	case errors.Is(err, contract.ErrLiveUnavailable):
		return "LIVE_UNAVAILABLE", message
	case errors.Is(err, contract.ErrConfigNotFound):
		return "CONFIG_NOT_FOUND", message
	case errors.Is(err, contract.ErrReleaseNotFound):
		return "RELEASE_NOT_FOUND", message
	case errors.Is(err, contract.ErrCenterNotFound):
		return "CENTER_NOT_FOUND", message
	case errors.Is(err, contract.ErrCenterNotRunning):
		return "CENTER_NOT_RUNNING", message
	default:
		return "ERROR", message
	}
}
