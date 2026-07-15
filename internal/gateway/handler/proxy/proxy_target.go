package proxy

import (
	"path"
	"strings"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
)

// buildTargetQuery 合并目标节点地址与客户端请求的查询参数。
//
// 处理规则：
//  1. 节点地址(后台配置)未携带查询参数：原样使用客户端请求的查询参数。
//  2. 客户端请求未携带查询参数：直接使用节点地址中的查询参数。
//  3. 两者都存在：以节点地址参数为准覆盖客户端同名参数；客户端独有的参数追加保留；
//     节点地址独有的参数也会保留（因为整体以节点地址查询串为基础）。
//
// 节点地址查询串原样保留（不重新排序或重新编码），避免破坏签名类参数
// （如 sign、timestamp、apptoken）的原始顺序与编码。
func buildTargetQuery(targetRawQuery, requestRawQuery string) string {
	if targetRawQuery == "" {
		return requestRawQuery
	}
	if requestRawQuery == "" {
		return targetRawQuery
	}

	// 收集节点地址中已配置的参数名，用于判断客户端同名参数是否需要被覆盖（丢弃）
	targetKeys := make(map[string]struct{})
	for _, pair := range strings.Split(targetRawQuery, "&") {
		if pair == "" {
			continue
		}
		key := pair
		if idx := strings.IndexByte(pair, '='); idx >= 0 {
			key = pair[:idx]
		}
		targetKeys[key] = struct{}{}
	}

	// 以节点地址查询串为基础（保持原始顺序与编码），再追加客户端中节点地址未配置的参数。
	merged := targetRawQuery
	for _, pair := range strings.Split(requestRawQuery, "&") {
		if pair == "" {
			continue
		}
		key := pair
		if idx := strings.IndexByte(pair, '='); idx >= 0 {
			key = pair[:idx]
		}
		// 同名参数由节点地址覆盖，跳过客户端携带的该参数
		if _, exists := targetKeys[key]; exists {
			continue
		}
		merged += "&" + pair
	}
	return merged
}

// buildTargetPath 构建代理请求路径 - 简化的nginx proxy_pass处理方式。
// HTTP 与 WebSocket 共用同一套路径规则，避免两套入口拼接语义分叉。
//
// 处理规则：
// 1. 目标路径为空或只有斜杠：使用请求地址
// 2. 前缀不一样：直接使用目标地址
// 3. 前缀一样：处理重复拼接问题
//
// 额外可选策略（历史上库字段存在但未进入运行时；开启后改变转发路径）：
// - rewritePath 非空：直接使用重写路径，不再拼接请求路径
// - stripPathPrefix=Y：先去掉已匹配路由前缀，再按上述 nginx 规则处理
func buildTargetPath(ctx *core.Context, targetPath string) string {
	requestPath := ctx.Request.URL.Path

	// 路由显式重写路径：整段替换，不回落到 nginx 拼接逻辑
	if value, exists := ctx.Get(constants.ContextKeyRouteRewritePath); exists {
		if rewritePath, ok := value.(string); ok && rewritePath != "" {
			return cleanTargetPath(rewritePath)
		}
	}

	// 剥离已匹配路由前缀后再参与后续拼接；仅当 stripPathPrefix 开启时生效。
	// 使用路径段边界，避免 /api 误匹配 /apix。
	if stripPrefix, exists := ctx.GetBool(constants.ContextKeyRouteStripPathPrefix); exists && stripPrefix {
		if stripped, ok := stripMatchedRoutePrefix(requestPath, ctx.GetMatchedPath()); ok {
			requestPath = stripped
			// 目标有具体路径时：节点 path + 去前缀后的请求剩余段
			if targetPath != "" && targetPath != "/" {
				return cleanTargetPath(strings.TrimSuffix(targetPath, "/") + requestPath)
			}
			// 目标为空或根路径：落到下方原始规则，但使用已去前缀的 requestPath
		}
	}

	// 记住原始路径的斜杠状态
	originalTargetHasSlash := strings.HasSuffix(targetPath, "/")
	originalRequestHasSlash := strings.HasSuffix(requestPath, "/")

	// 清理路径
	targetPath = cleanTargetPath(targetPath)
	requestPath = cleanTargetPath(requestPath)

	// 1. 目标路径为空或只有斜杠：使用请求地址，但要保留原始请求路径的斜杠状态
	if targetPath == "" || targetPath == "/" {
		// 如果原始请求路径以斜杠结尾且清理后不是根路径，需要恢复斜杠
		if originalRequestHasSlash && requestPath != "/" {
			return requestPath + "/"
		}
		return requestPath
	}

	// 2. 前缀不一样：直接使用目标地址
	if !hasTargetPathPrefix(targetPath, requestPath) {
		// 如果原始目标路径有斜杠，保留它
		if originalTargetHasSlash && !strings.HasSuffix(targetPath, "/") {
			return targetPath + "/"
		}
		return targetPath
	}

	// 3. 前缀一样：处理重复拼接问题
	// 特殊情况：如果路径完全相同，直接返回目标路径
	if targetPath == requestPath {
		if originalTargetHasSlash && !strings.HasSuffix(targetPath, "/") {
			return targetPath + "/"
		}
		return targetPath
	}

	// 如果请求路径以目标路径为前缀，直接返回请求路径避免重复
	if strings.HasPrefix(requestPath, targetPath) {
		return requestPath
	}

	// 否则根据是否有斜杠决定拼接方式
	if originalTargetHasSlash {
		// 目标路径原本有斜杠，直接拼接
		if requestPath == "/" {
			return targetPath + "/"
		}
		return targetPath + requestPath
	}
	// 目标路径不以/结尾，直接拼接
	return targetPath + requestPath
}

// stripMatchedRoutePrefix 按路径段边界去掉已匹配路由前缀。
// 例如匹配 /api 时可剥 /api/users，但不会误剥 /apix/users。
func stripMatchedRoutePrefix(requestPath, matchedPath string) (string, bool) {
	if matchedPath == "" {
		return requestPath, false
	}
	matchedPath = cleanTargetPath(matchedPath)
	if matchedPath == "/" {
		return requestPath, false
	}
	cleanedRequest := cleanTargetPath(requestPath)
	if !hasTargetPathPrefix(matchedPath, cleanedRequest) {
		return requestPath, false
	}
	basePath := strings.TrimSuffix(matchedPath, "/")
	remaining := strings.TrimPrefix(cleanedRequest, basePath)
	if remaining == "" {
		remaining = "/"
	}
	if !strings.HasPrefix(remaining, "/") {
		remaining = "/" + remaining
	}
	return remaining, true
}

// hasTargetPathPrefix 检查目标路径和请求路径是否有相同前缀
func hasTargetPathPrefix(targetPath, requestPath string) bool {
	// 获取目标路径的基础部分（去掉结尾斜杠）
	basePath := strings.TrimSuffix(targetPath, "/")

	// 特殊情况：如果目标路径是根路径，只有请求路径也是根路径才算相同前缀
	if basePath == "" {
		return requestPath == "/"
	}

	// 如果请求路径不以目标路径开头，前缀不同
	if !strings.HasPrefix(requestPath, basePath) {
		return false
	}

	// 检查路径边界：确保匹配的是完整的路径段
	// 例如："/ap" 不应该匹配 "/api/v1"
	if len(requestPath) > len(basePath) {
		nextChar := requestPath[len(basePath)]
		return nextChar == '/'
	}

	// 请求路径长度等于或小于目标路径，认为是相同前缀
	return true
}

// cleanTargetPath 清理路径格式
func cleanTargetPath(value string) string {
	if value == "" {
		return "/"
	}

	// 确保以 / 开头
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}

	// 使用 path.Clean 清理路径
	return path.Clean(value)
}
