package reqhand

import (
	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/filter"
)

// PreserveOriginalRequestInfoIfNeeded 统一的静态方法：检查过滤器并保存原始请求信息
// 返回是否保存了原始信息
func PreserveOriginalRequestInfoIfNeeded(ctx *core.Context, filters []filter.Filter) bool {
	// 检查是否有修改类过滤器
	hasModifiers := HasModificationFilters(filters)

	if !hasModifiers {
		return false
	}

	// 检查是否需要保存请求头
	needsHeaders := NeedsHeaderPreservation(filters)

	// 保存原始信息
	PreserveOriginalRequestInfo(ctx, needsHeaders)
	return true
}

// PreserveOriginalRequestInfo 静态方法：保存原始请求信息到上下文
func PreserveOriginalRequestInfo(ctx *core.Context, needsHeaders bool) {
	req := ctx.Request
	if req == nil {
		return
	}

	// 保存原始请求信息到上下文
	ctx.Set(constants.ContextKeyOriginalMethod, req.Method)
	ctx.Set(constants.ContextKeyOriginalURLPath, req.URL.Path)
	ctx.Set(constants.ContextKeyOriginalQueryString, req.URL.RawQuery)

	// 保存原始请求头（深拷贝，仅在需要时）
	if needsHeaders {
		originalHeaders := make(map[string][]string)
		for name, values := range req.Header {
			originalHeaders[name] = append([]string{}, values...)
		}
		ctx.Set(constants.ContextKeyOriginalHeaders, originalHeaders)
	}
}

// HasModificationFilters 静态方法：检查过滤器列表是否包含修改类过滤器
func HasModificationFilters(filters []filter.Filter) bool {
	for _, f := range filters {
		if IsModificationFilterType(f.GetType()) {
			return true
		}
	}
	return false
}

// NeedsHeaderPreservation 静态方法：检查是否需要保存请求头
func NeedsHeaderPreservation(filters []filter.Filter) bool {
	for _, f := range filters {
		if f.GetType() == filter.HeaderFilterType {
			return true
		}
	}
	return false
}

// IsModificationFilterType 静态方法：判断过滤器类型是否为修改类
func IsModificationFilterType(filterType filter.FilterType) bool {
	switch filterType {
	case filter.QueryParamFilterType,
		filter.HeaderFilterType,
		filter.BodyFilterType,
		filter.URLFilterType,
		filter.StripFilterType,
		filter.RewriteFilterType,
		filter.MethodFilterType:
		return true
	default:
		return false
	}
}

// SnapshotHTTPData 快照HTTP请求和响应数据到上下文
// 用于异步日志记录，在 ServeHTTP 返回前调用
// 保存请求和响应的关键信息，避免在异步goroutine中访问已回收的HTTP对象
// 与 PreserveOriginalRequestInfo 保持一致，使用常量键名
func SnapshotHTTPData(ctx *core.Context) {
	if ctx.Request != nil {
		// 如果还没有保存原始方法，则保存（与 PreserveOriginalRequestInfo 保持一致）
		if _, exists := ctx.Get(constants.ContextKeyOriginalMethod); !exists {
			ctx.Set(constants.ContextKeyOriginalMethod, ctx.Request.Method)
		}

		// 如果还没有保存原始路径，则保存
		if _, exists := ctx.Get(constants.ContextKeyOriginalURLPath); !exists {
			ctx.Set(constants.ContextKeyOriginalURLPath, ctx.Request.URL.Path)
		}

		// 如果还没有保存原始查询字符串，则保存
		if _, exists := ctx.Get(constants.ContextKeyOriginalQueryString); !exists {
			ctx.Set(constants.ContextKeyOriginalQueryString, ctx.Request.URL.RawQuery)
		}

		// 保存快照请求信息（用于异步日志记录）
		// 注意：Method、Path、QueryString、Headers、URI 已通过原始信息保存，不需要重复保存
		ctx.Set(constants.ContextKeySnapshotRequestProto, ctx.Request.Proto)
		ctx.Set(constants.ContextKeySnapshotRequestHost, ctx.Request.Host)
		ctx.Set(constants.ContextKeySnapshotRequestRemoteAddr, ctx.Request.RemoteAddr)

		// 缓存请求大小（从 Content-Length 获取）
		if ctx.Request.ContentLength >= 0 {
			ctx.Set(constants.ContextKeySnapshotRequestSize, int(ctx.Request.ContentLength))
		} else {
			ctx.Set(constants.ContextKeySnapshotRequestSize, -1) // 未知大小
		}

		// 深拷贝请求头（保持原始键名，不修改）
		if ctx.Request.Header != nil {
			requestHeaders := make(map[string][]string, len(ctx.Request.Header))
			for key, values := range ctx.Request.Header {
				valuesCopy := make([]string, len(values))
				copy(valuesCopy, values)
				requestHeaders[key] = valuesCopy
			}
			// 如果还没有保存原始请求头，则保存（与 PreserveOriginalRequestInfo 保持一致）
			if _, exists := ctx.Get(constants.ContextKeyOriginalHeaders); !exists {
				ctx.Set(constants.ContextKeyOriginalHeaders, requestHeaders)
			}
		}
	}

	// 从 Writer 中提取并缓存响应头（保持原始键名，不修改）
	if ctx.Writer != nil {
		respHeaders := ctx.Writer.Header()
		if respHeaders != nil {
			responseHeaders := make(map[string][]string, len(respHeaders))
			for key, values := range respHeaders {
				valuesCopy := make([]string, len(values))
				copy(valuesCopy, values)
				responseHeaders[key] = valuesCopy
			}
			ctx.Set(constants.ContextKeySnapshotResponseHeaders, responseHeaders)
		}
	}
}
