package proxy

import (
	"time"
)

// 包级别的静态函数 - 避免重复代码

// getConfigValue 从配置映射中获取值，支持多种命名方式
// keys: 按优先级排序的键名列表，通常驼峰命名在前，下划线命名在后
func getConfigValue(configMap map[string]interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if value, exists := configMap[key]; exists {
			return value
		}
	}
	return nil
}

// parseIntFromNumber 从interface{}中解析整数，支持int、int64、float64类型
func parseIntFromNumber(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	default:
		return 0
	}
}

// parseInt64FromNumber 从interface{}中解析int64，支持int、int64、float64类型
func parseInt64FromNumber(value interface{}) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	default:
		return 0
	}
}

// parseDurationFromNumber 从interface{}中解析时间长度，支持int、float64类型（按秒计算）
func parseDurationFromNumber(value interface{}) time.Duration {
	switch v := value.(type) {
	case int:
		return time.Duration(v) * time.Second
	case int64:
		return time.Duration(v) * time.Second
	case float64:
		return time.Duration(v) * time.Second
	case float32:
		return time.Duration(v) * time.Second
	default:
		return 0
	}
}

// HTTPConfigParser HTTP配置解析器
type HTTPConfigParser struct{}

// NewHTTPConfigParser 创建HTTP配置解析器
func NewHTTPConfigParser() *HTTPConfigParser {
	return &HTTPConfigParser{}
}

// ParseConfig 解析HTTP配置 - 支持驼峰命名和下划线命名
func (p *HTTPConfigParser) ParseConfig(configMap map[string]interface{}, httpConfig *HTTPProxyConfig) {
	// timeout 字段解析
	if timeout := getConfigValue(configMap, "timeout"); timeout != nil {
		if timeoutStr, ok := timeout.(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				httpConfig.Timeout = d
			}
		}
		// 支持整型和浮点型按秒转换
		if duration := parseDurationFromNumber(timeout); duration > 0 {
			httpConfig.Timeout = duration
		}
	}
	
	// sendTimeout 字段解析
	if sendTimeout := getConfigValue(configMap, "sendTimeout", "send_timeout"); sendTimeout != nil {
		if timeoutStr, ok := sendTimeout.(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				httpConfig.SendTimeout = d
			}
		}
		// 支持整型和浮点型按秒转换
		if duration := parseDurationFromNumber(sendTimeout); duration > 0 {
			httpConfig.SendTimeout = duration
		}
	}
	
	// readTimeout 字段解析
	if readTimeout := getConfigValue(configMap, "readTimeout", "read_timeout"); readTimeout != nil {
		if timeoutStr, ok := readTimeout.(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				httpConfig.ReadTimeout = d
			}
		}
		// 支持整型和浮点型按秒转换
		if duration := parseDurationFromNumber(readTimeout); duration > 0 {
			httpConfig.ReadTimeout = duration
		}
	}
	
	// connectTimeout 字段解析
	if connectTimeout := getConfigValue(configMap, "connectTimeout", "connect_timeout"); connectTimeout != nil {
		if timeoutStr, ok := connectTimeout.(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				httpConfig.ConnectTimeout = d
			}

		}
		// 支持整型和浮点型按秒转换
		if duration := parseDurationFromNumber(connectTimeout); duration > 0 {
			httpConfig.ConnectTimeout = duration
		}
	}

	// followRedirects 字段解析
	if followRedirects := getConfigValue(configMap, "followRedirects", "follow_redirects"); followRedirects != nil {
		if b, ok := followRedirects.(bool); ok {
			httpConfig.FollowRedirects = b
		}
	}

	// keepAlive 字段解析
	if keepAlive := getConfigValue(configMap, "keepAlive", "keep_alive"); keepAlive != nil {
		if b, ok := keepAlive.(bool); ok {
			httpConfig.KeepAlive = b
		}
	}

	// maxIdleConns 字段解析
	if maxIdleConns := getConfigValue(configMap, "maxIdleConns", "max_idle_conns"); maxIdleConns != nil {
		if i := parseIntFromNumber(maxIdleConns); i > 0 {
			httpConfig.MaxIdleConns = i
		}
	}

	// idleConnTimeout 字段解析
	if idleConnTimeout := getConfigValue(configMap, "idleConnTimeout", "idle_conn_timeout"); idleConnTimeout != nil {
		if timeoutStr, ok := idleConnTimeout.(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				httpConfig.IdleConnTimeout = d
			}
		}
		// 支持整型和浮点型按秒转换
		if duration := parseDurationFromNumber(idleConnTimeout); duration > 0 {
			httpConfig.IdleConnTimeout = duration
		}
	}

	// copyResponseBody 字段解析
	if copyResponseBody := getConfigValue(configMap, "copyResponseBody", "copy_response_body"); copyResponseBody != nil {
		if b, ok := copyResponseBody.(bool); ok {
			httpConfig.CopyResponseBody = b
		}
	}

	// bufferSize 字段解析
	if bufferSize := getConfigValue(configMap, "bufferSize", "buffer_size"); bufferSize != nil {
		if i := parseIntFromNumber(bufferSize); i > 0 {
			httpConfig.BufferSize = i
		}
	}

	// maxBufferSize 字段解析
	if maxBufferSize := getConfigValue(configMap, "maxBufferSize", "max_buffer_size"); maxBufferSize != nil {
		if i := parseIntFromNumber(maxBufferSize); i > 0 {
			httpConfig.MaxBufferSize = i
		}
	}

	// retryCount 字段解析
	if retryCount := getConfigValue(configMap, "retryCount", "retry_count"); retryCount != nil {
		if i := parseIntFromNumber(retryCount); i >= 0 {
			httpConfig.RetryCount = i
		}
	}

	// retryTimeout 字段解析
	if retryTimeout := getConfigValue(configMap, "retryTimeout", "retry_timeout"); retryTimeout != nil {
		if timeoutStr, ok := retryTimeout.(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				httpConfig.RetryTimeout = d
			}
		}
		// 支持整型和浮点型按秒转换
		if duration := parseDurationFromNumber(retryTimeout); duration > 0 {
			httpConfig.RetryTimeout = duration
		}
	}
	
	// 解析头部配置
	if setHeaders := getConfigValue(configMap, "setHeaders", "set_headers"); setHeaders != nil {
		if headers, ok := setHeaders.(map[string]interface{}); ok {
			httpConfig.SetHeaders = make(map[string]string)
			for k, v := range headers {
				if str, ok := v.(string); ok {
					httpConfig.SetHeaders[k] = str
				}
			}
		}
	}
	
	if passHeaders := getConfigValue(configMap, "passHeaders", "pass_headers"); passHeaders != nil {
		if headers, ok := passHeaders.([]interface{}); ok {
			httpConfig.PassHeaders = make([]string, 0, len(headers))
			for _, h := range headers {
				if str, ok := h.(string); ok {
					httpConfig.PassHeaders = append(httpConfig.PassHeaders, str)
				}
			}
		}
	}
	
	if hideHeaders := getConfigValue(configMap, "hideHeaders", "hide_headers"); hideHeaders != nil {
		if headers, ok := hideHeaders.([]interface{}); ok {
			httpConfig.HideHeaders = make([]string, 0, len(headers))
			for _, h := range headers {
				if str, ok := h.(string); ok {
					httpConfig.HideHeaders = append(httpConfig.HideHeaders, str)
				}
			}
		}
	}
	
	// 解析高级选项
	if httpVersion := getConfigValue(configMap, "httpVersion", "http_version"); httpVersion != nil {
		if str, ok := httpVersion.(string); ok {
			httpConfig.HTTPVersion = str
		}
	}
	
	if preserveHost := getConfigValue(configMap, "preserveHost", "preserve_host"); preserveHost != nil {
		if b, ok := preserveHost.(bool); ok {
			httpConfig.PreserveHost = b
		}
	}
	
	if addXForwardedFor := getConfigValue(configMap, "addXForwardedFor", "add_x_forwarded_for"); addXForwardedFor != nil {
		if b, ok := addXForwardedFor.(bool); ok {
			httpConfig.AddXForwardedFor = b
		}
	}
	
	if addXRealIP := getConfigValue(configMap, "addXRealIP", "add_x_real_ip"); addXRealIP != nil {
		if b, ok := addXRealIP.(bool); ok {
			httpConfig.AddXRealIP = b
		}
	}
	
	if addXForwardedProto := getConfigValue(configMap, "addXForwardedProto", "add_x_forwarded_proto"); addXForwardedProto != nil {
		if b, ok := addXForwardedProto.(bool); ok {
			httpConfig.AddXForwardedProto = b
		}
	}
	
	// 解析nginx风格配置
	if proxyBuffering := getConfigValue(configMap, "proxyBuffering", "proxy_buffering"); proxyBuffering != nil {
		if b, ok := proxyBuffering.(bool); ok {
			httpConfig.ProxyBuffering = b
		}
	}
	
	// 解析TLS配置
	if tlsInsecureSkipVerify := getConfigValue(configMap, "tlsInsecureSkipVerify", "tls_insecure_skip_verify"); tlsInsecureSkipVerify != nil {
		if b, ok := tlsInsecureSkipVerify.(bool); ok {
			httpConfig.TLSInsecureSkipVerify = b
		}
	}
	
	if tlsMinVersion := getConfigValue(configMap, "tlsMinVersion", "tls_min_version"); tlsMinVersion != nil {
		if str, ok := tlsMinVersion.(string); ok {
			httpConfig.TLSMinVersion = str
		}
	}
	
	if tlsMaxVersion := getConfigValue(configMap, "tlsMaxVersion", "tls_max_version"); tlsMaxVersion != nil {
		if str, ok := tlsMaxVersion.(string); ok {
			httpConfig.TLSMaxVersion = str
		}
	}
	
	if tlsServerName := getConfigValue(configMap, "tlsServerName", "tls_server_name"); tlsServerName != nil {
		if str, ok := tlsServerName.(string); ok {
			httpConfig.TLSServerName = str
		}
	}
}

// WebSocketConfigParser WebSocket配置解析器
type WebSocketConfigParser struct{}

// NewWebSocketConfigParser 创建WebSocket配置解析器
func NewWebSocketConfigParser() *WebSocketConfigParser {
	return &WebSocketConfigParser{}
}

// ParseConfig 解析WebSocket配置 - 支持驼峰命名和下划线命名
func (p *WebSocketConfigParser) ParseConfig(configMap map[string]interface{}, wsConfig *WebSocketConfig) {
	// enabled 字段解析
	if enabled := getConfigValue(configMap, "enabled"); enabled != nil {
		if b, ok := enabled.(bool); ok {
			wsConfig.Enabled = b
		}
	}

	// pingInterval 字段解析
	if pingInterval := getConfigValue(configMap, "pingInterval", "ping_interval"); pingInterval != nil {
		if intervalStr, ok := pingInterval.(string); ok {
			if d, err := time.ParseDuration(intervalStr); err == nil {
				wsConfig.PingInterval = d
			}
		}
		// 支持整型和浮点型按秒转换
		if duration := parseDurationFromNumber(pingInterval); duration > 0 {
			wsConfig.PingInterval = duration
		}
	}

	// pongTimeout 字段解析
	if pongTimeout := getConfigValue(configMap, "pongTimeout", "pong_timeout"); pongTimeout != nil {
		if timeoutStr, ok := pongTimeout.(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				wsConfig.PongTimeout = d
			}
		}
		// 支持整型和浮点型按秒转换
		if duration := parseDurationFromNumber(pongTimeout); duration > 0 {
			wsConfig.PongTimeout = duration
		}
	}

	// writeTimeout 字段解析
	if writeTimeout := getConfigValue(configMap, "writeTimeout", "write_timeout"); writeTimeout != nil {
		if timeoutStr, ok := writeTimeout.(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				wsConfig.WriteTimeout = d
			}
		}
		// 支持整型和浮点型按秒转换
		if duration := parseDurationFromNumber(writeTimeout); duration > 0 {
			wsConfig.WriteTimeout = duration
		}
	}

	// readTimeout 字段解析
	if readTimeout := getConfigValue(configMap, "readTimeout", "read_timeout"); readTimeout != nil {
		if timeoutStr, ok := readTimeout.(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				wsConfig.ReadTimeout = d
			}
		}
		// 支持整型和浮点型按秒转换
		if duration := parseDurationFromNumber(readTimeout); duration > 0 {
			wsConfig.ReadTimeout = duration
		}
	}

	// maxMessageSize 字段解析
	if maxMessageSize := getConfigValue(configMap, "maxMessageSize", "max_message_size"); maxMessageSize != nil {
		if i := parseInt64FromNumber(maxMessageSize); i > 0 {
			wsConfig.MaxMessageSize = i
		}
	}

	// readBufferSize 字段解析
	if readBufferSize := getConfigValue(configMap, "readBufferSize", "read_buffer_size"); readBufferSize != nil {
		if i := parseIntFromNumber(readBufferSize); i > 0 {
			wsConfig.ReadBufferSize = i
		}
	}

	// writeBufferSize 字段解析
	if writeBufferSize := getConfigValue(configMap, "writeBufferSize", "write_buffer_size"); writeBufferSize != nil {
		if i := parseIntFromNumber(writeBufferSize); i > 0 {
			wsConfig.WriteBufferSize = i
		}
	}

	// enableCompression 字段解析
	if enableCompression := getConfigValue(configMap, "enableCompression", "enable_compression"); enableCompression != nil {
		if b, ok := enableCompression.(bool); ok {
			wsConfig.EnableCompression = b
		}
	}

	// subprotocols 字段解析
	if subprotocols := getConfigValue(configMap, "subprotocols"); subprotocols != nil {
		if protocols, ok := subprotocols.([]interface{}); ok {
			wsConfig.Subprotocols = make([]string, 0, len(protocols))
			for _, protocol := range protocols {
				if str, ok := protocol.(string); ok {
					wsConfig.Subprotocols = append(wsConfig.Subprotocols, str)
				}
			}
		}
	}
} 
