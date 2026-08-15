package circuitbreaker

import "errors"

// errUpstreamServerError 上游返回 5xx 时的熔断失败原因。
var errUpstreamServerError = errors.New("upstream returned 5xx")

// IsFailureStatus 判断上游 HTTP 状态是否计入节点熔断失败。
// 5xx 表示上游自身故障；4xx 是调用方问题，不摘除节点。
func IsFailureStatus(statusCode int) bool {
	return statusCode >= 500 && statusCode <= 599
}

// AttemptFailed 判断本次上游尝试是否计入熔断失败。
// 连接/超时错误与上游 5xx 都算失败。
func AttemptFailed(err error, statusCode int) bool {
	return err != nil || IsFailureStatus(statusCode)
}

// AttemptError 返回写入熔断的错误；5xx 且无传输错误时用统一原因。
func AttemptError(err error, statusCode int) error {
	if err != nil {
		return err
	}
	if IsFailureStatus(statusCode) {
		return errUpstreamServerError
	}
	return nil
}
