package syssetting

import (
	"strings"
	"sync"

	"gateway/pkg/logger"
	"gateway/pkg/security"
)

const defaultTenantID = "default"

// store 进程内缓存，按租户保存已加载的环境设置。未加载时 Get 返回默认值。
type store struct {
	mu           sync.RWMutex
	retention    map[string]RetentionSettings
	retentionJob map[string]RetentionJobSettings
	webTimeout   map[string]WebTimeoutSettings
	envVars      map[string]map[string]string
}

var global = &store{
	retention:    make(map[string]RetentionSettings),
	retentionJob: make(map[string]RetentionJobSettings),
	webTimeout:   make(map[string]WebTimeoutSettings),
	envVars:      make(map[string]map[string]string),
}

var (
	timeoutApplierMu   sync.RWMutex
	httpTimeoutApplier func(seconds int)
)

// RegisterHTTPTimeoutApplier 由 Web 服务注册，用于把接口超时同步到 http.Server。
// 保存 webTimeout 或启动时写入缓存后会回调 seconds。
func RegisterHTTPTimeoutApplier(fn func(seconds int)) {
	timeoutApplierMu.Lock()
	httpTimeoutApplier = fn
	timeoutApplierMu.Unlock()
}

func notifyHTTPTimeout(seconds int) {
	timeoutApplierMu.RLock()
	fn := httpTimeoutApplier
	timeoutApplierMu.RUnlock()
	if fn != nil {
		fn(seconds)
	}
}

func normTenant(tenantId string) string {
	if tenantId == "" {
		return defaultTenantID
	}
	return tenantId
}

// GetRetention 返回租户归档策略。未加载过该租户时返回默认值。
func GetRetention(tenantId string) RetentionSettings {
	tenantId = normTenant(tenantId)
	global.mu.RLock()
	defer global.mu.RUnlock()
	if v, ok := global.retention[tenantId]; ok {
		return v
	}
	return DefaultRetention()
}

// GetRetentionJob 返回租户归档任务调度。未加载过该租户时返回默认值。
func GetRetentionJob(tenantId string) RetentionJobSettings {
	tenantId = normTenant(tenantId)
	global.mu.RLock()
	defer global.mu.RUnlock()
	if v, ok := global.retentionJob[tenantId]; ok {
		return v
	}
	return DefaultRetentionJob()
}

// GetWebTimeout 返回租户 Web 超时。未加载过该租户时返回默认值。
func GetWebTimeout(tenantId string) WebTimeoutSettings {
	tenantId = normTenant(tenantId)
	global.mu.RLock()
	defer global.mu.RUnlock()
	if v, ok := global.webTimeout[tenantId]; ok {
		return v
	}
	return DefaultWebTimeout()
}

// GetWebCipher 返回全集群共用的密文传输设置。优先 default 租户已启用且有钥的行，
// 否则取任意已加载租户中带私钥的一行。登录尚不知租户，必须走这里。
func GetWebCipher() WebTimeoutSettings {
	global.mu.RLock()
	defer global.mu.RUnlock()
	return pickWebCipher(global.webTimeout)
}

func pickWebCipher(rows map[string]WebTimeoutSettings) WebTimeoutSettings {
	var fallback WebTimeoutSettings
	if v, ok := rows[defaultTenantID]; ok && strings.TrimSpace(v.PrivateKey) != "" {
		if v.CipherEnabled {
			return v
		}
		fallback = v
	}
	for _, v := range rows {
		if v.CipherEnabled && strings.TrimSpace(v.PrivateKey) != "" {
			return v
		}
		if strings.TrimSpace(fallback.PrivateKey) == "" && strings.TrimSpace(v.PrivateKey) != "" {
			fallback = v
		}
	}
	if strings.TrimSpace(fallback.PrivateKey) != "" {
		return fallback
	}
	return DefaultWebTimeout()
}

// PutRetention 写入租户归档策略缓存，保存成功后由 hub0009 调用。
func PutRetention(tenantId string, v RetentionSettings) {
	tenantId = normTenant(tenantId)
	global.mu.Lock()
	defer global.mu.Unlock()
	global.retention[tenantId] = mergeRetention(v)
}

// PutRetentionJob 写入租户归档任务缓存，保存成功后由 hub0009 调用。
func PutRetentionJob(tenantId string, v RetentionJobSettings) {
	tenantId = normTenant(tenantId)
	global.mu.Lock()
	defer global.mu.Unlock()
	global.retentionJob[tenantId] = mergeRetentionJob(v)
}

// PutWebTimeout 写入租户 Web 超时缓存，保存成功后由 hub0009 调用。
// 写入后通知已注册的 HTTP 服务端，使 Read/Write 超时与 axios 使用同一秒数。
func PutWebTimeout(tenantId string, v WebTimeoutSettings) {
	tenantId = normTenant(tenantId)
	global.mu.Lock()
	merged := mergeWebTimeout(v)
	global.webTimeout[tenantId] = merged
	global.mu.Unlock()
	notifyHTTPTimeout(merged.RequestTimeoutSeconds)
	activateWebCipher()
}

// PutEnvVars 写入租户全局环境变量缓存。密文在写入时解密，运行时按明文展开。
func PutEnvVars(tenantId string, v EnvVarsSettings) {
	tenantId = normTenant(tenantId)
	resolved := decodeEnvVarMap(v)
	global.mu.Lock()
	defer global.mu.Unlock()
	global.envVars[tenantId] = resolved
}

// GetEnvVar 按名称读取已解密的变量值。不存在时 ok 为 false。
func GetEnvVar(tenantId, name string) (string, bool) {
	if name == "" {
		return "", false
	}
	tenantId = normTenant(tenantId)
	global.mu.RLock()
	defer global.mu.RUnlock()
	vars, ok := global.envVars[tenantId]
	if !ok {
		return "", false
	}
	val, ok := vars[name]
	return val, ok
}

// ApplyGroup 按分组编码把 JSON 写入进程缓存。未知分组忽略。
// webTimeout 会回调已注册的 HTTP 超时同步。
func ApplyGroup(tenantId, groupCode, content string) {
	switch groupCode {
	case GroupRetention:
		PutRetention(tenantId, ParseRetention(content))
	case GroupRetentionJob:
		PutRetentionJob(tenantId, ParseRetentionJob(content))
	case GroupWebTimeout:
		PutWebTimeout(tenantId, ParseWebTimeout(content))
	case GroupEnvVars:
		PutEnvVars(tenantId, ParseEnvVars(content))
	}
}

// KnownTenantIDs 返回缓存中出现过的租户，并保证包含 default，供清理任务遍历。
func KnownTenantIDs() []string {
	global.mu.RLock()
	defer global.mu.RUnlock()
	seen := map[string]struct{}{defaultTenantID: {}}
	for k := range global.retention {
		seen[k] = struct{}{}
	}
	for k := range global.retentionJob {
		seen[k] = struct{}{}
	}
	for k := range global.webTimeout {
		seen[k] = struct{}{}
	}
	for k := range global.envVars {
		seen[k] = struct{}{}
	}
	ids := make([]string, 0, len(seen))
	for k := range seen {
		ids = append(ids, k)
	}
	return ids
}

// activateWebCipher 按全局密文设置安装或清空包装钥。PutWebTimeout 之后调用。
func activateWebCipher() {
	s := GetWebCipher()
	if !s.CipherEnabled {
		security.SetPasswordWrap(nil)
		return
	}
	pemStr := decodeWebCipherPrivateKey(s.PrivateKey)
	if pemStr == "" {
		logger.Error("已开启密文传输但私钥不可用，拒绝明文口令")
		security.SetPasswordWrap(nil)
		return
	}
	w, err := security.LoadPasswordWrapPEM(pemStr)
	if err != nil {
		logger.Error("加载 Web 传输私钥失败", "error", err.Error())
		security.SetPasswordWrap(nil)
		return
	}
	security.SetPasswordWrap(w)
}

// GetRetentionJobForSchedule 归档 Job 的进程级调度。优先 default 租户已缓存值，否则用任意已加载租户。
func GetRetentionJobForSchedule() RetentionJobSettings {
	global.mu.RLock()
	defer global.mu.RUnlock()
	if v, ok := global.retentionJob[defaultTenantID]; ok {
		return v
	}
	for _, v := range global.retentionJob {
		return v
	}
	return DefaultRetentionJob()
}
