package model

// AlertConfig 来自 CenterInstance.ExtProperty，与中台 hub0040 字段一一对应。
// 未写的开关使用下列默认：启动失败/异常停止/健康检查/驱逐/同步失败/配置变更为开，
// 注册/注销/订阅/断连为关（高频）。AlertEnabled 默认关，避免空 ExtProperty 误告警。
type AlertConfig struct {
	AlertEnabled           bool
	ChannelName            string
	AlertOnStartFailure    bool
	AlertOnStopAbnormal    bool
	AlertOnHealthCheckFail bool
	AlertOnNodeEviction    bool
	NodeEvictionThreshold  int
	AlertOnSyncFailure     bool
	AlertOnNodeRegister    bool
	AlertOnNodeUnregister  bool
	AlertOnSubscribeNotify bool
	AlertOnConfigChange    bool
	AlertOnConnectionLost  bool
}

// DefaultAlertConfig 返回与前端未勾选时一致的默认值。
func DefaultAlertConfig() AlertConfig {
	return AlertConfig{
		AlertOnStartFailure:    true,
		AlertOnStopAbnormal:    true,
		AlertOnHealthCheckFail: true,
		AlertOnNodeEviction:    true,
		NodeEvictionThreshold:  5,
		AlertOnSyncFailure:     true,
		AlertOnConfigChange:    true,
	}
}

// ParseAlertConfig 从 ExtProperty 解析告警开关。与 hub0040 extProperty.alert* 字段对齐。
func ParseAlertConfig(extProperty string) AlertConfig {
	cfg := DefaultAlertConfig()
	m := ParseExtProperty(extProperty)
	if len(m) == 0 {
		return cfg
	}
	cfg.AlertEnabled = ExtYN(m, "alertEnabled", false)
	cfg.ChannelName = ExtString(m, "channelName")
	cfg.AlertOnStartFailure = ExtYN(m, "alertOnStartFailure", cfg.AlertOnStartFailure)
	cfg.AlertOnStopAbnormal = ExtYN(m, "alertOnStopAbnormal", cfg.AlertOnStopAbnormal)
	cfg.AlertOnHealthCheckFail = ExtYN(m, "alertOnHealthCheckFail", cfg.AlertOnHealthCheckFail)
	cfg.AlertOnNodeEviction = ExtYN(m, "alertOnNodeEviction", cfg.AlertOnNodeEviction)
	cfg.NodeEvictionThreshold = ExtInt(m, "nodeEvictionThreshold", cfg.NodeEvictionThreshold)
	cfg.AlertOnSyncFailure = ExtYN(m, "alertOnSyncFailure", cfg.AlertOnSyncFailure)
	cfg.AlertOnNodeRegister = ExtYN(m, "alertOnNodeRegister", cfg.AlertOnNodeRegister)
	cfg.AlertOnNodeUnregister = ExtYN(m, "alertOnNodeUnregister", cfg.AlertOnNodeUnregister)
	cfg.AlertOnSubscribeNotify = ExtYN(m, "alertOnSubscribeNotify", cfg.AlertOnSubscribeNotify)
	cfg.AlertOnConfigChange = ExtYN(m, "alertOnConfigChange", cfg.AlertOnConfigChange)
	cfg.AlertOnConnectionLost = ExtYN(m, "alertOnConnectionLost", cfg.AlertOnConnectionLost)
	return cfg
}

// TokenScope 是不透明 Bearer 令牌在 ExtProperty 中的绑定范围。
// instanceName 建议必填：多环境时令牌不得跨中心实例使用。
// namespaceIds 空表示该实例下全部命名空间。
type TokenScope struct {
	InstanceName string
	Environment  string
	NamespaceIDs []string
}

// ParseTokenScope 从 HUB_SERVICE_AUTH_TOKEN.extProperty 读取实例/命名空间绑定。
func ParseTokenScope(extProperty string) TokenScope {
	m := ParseExtProperty(extProperty)
	return TokenScope{
		InstanceName: ExtString(m, "instanceName"),
		Environment:  ExtString(m, "environment"),
		NamespaceIDs: ExtStringList(m, "namespaceIds"),
	}
}
