package model

// 中心实例运行状态。取值稳定，管理面与库表 status 列共用。
const (
	CenterStatusStopped  = "STOPPED"
	CenterStatusStarting = "STARTING"
	CenterStatusRunning  = "RUNNING"
	CenterStatusStopping = "STOPPING"
	CenterStatusError    = "ERROR"
)

// CenterInstance 是一条独立监听加一套隔离运行时，不是已注册的业务节点。
// Y/N 开关字段沿用既有库表约定，避免本版本改列类型。
//
// 运行时键是 InstanceName + Environment：多套环境不要复用同一实例名，
// 或必须在管理面每次调用都带 environment。下列字段均由 stream.Server /
// Evictor / 鉴权拦截器消费，不再只入库。
type CenterInstance struct {
	TenantID              string // 归属租户，创建时由 CallContext 写入；数据面令牌租户必须与此一致
	InstanceName          string // 实例名，与 Environment 组成进程内运行时键
	Environment           string // 环境标签（DEVELOPMENT/STAGING/PRODUCTION），库表联合键，也是信任边界的一部分
	ListenAddress         string // 数据面监听地址，空视为 0.0.0.0
	ListenPort            int    // 数据面 gRPC 端口，与网关 Web 端口分离
	MaxRecvMsgSize        int    // 单条接收上限（字节），<=0 时 16MiB
	MaxSendMsgSize        int    // 单条发送上限（字节），<=0 时 16MiB
	KeepAliveTime         int    // 服务端 keepalive 探测间隔（秒）
	KeepAliveTimeout      int    // keepalive 超时（秒）
	KeepAliveMinTime      int    // 客户端 ping 最小间隔（秒）
	PermitWithoutStream   string // 无活动流时是否允许 ping，Y/N
	MaxConnectionIdle     int    // 空闲连接最大存活（秒），0 表示不限制
	MaxConnectionAge      int    // 连接最大寿命（秒），0 表示不限制
	MaxConnectionAgeGrace int    // 强制断开宽限期（秒）
	EnableReflection      string // 是否打开 gRPC reflection，Y/N
	EnableTLS             string // 是否启用 TLS，Y/N
	CertStorageType       string // 证书来源：FILE / DATABASE
	CertFilePath          string // 证书文件路径（FILE）
	KeyFilePath           string // 私钥文件路径（FILE）
	CertContent           string // 证书内容（DATABASE）
	KeyContent            string // 私钥内容（DATABASE）
	CertChainContent      string // mTLS 客户端 CA：PEM 文本或 PEM 文件路径
	CertPassword          string // 私钥口令
	EnableMTLS            string // 是否双向 TLS，Y/N；Y 时必须配置 CertChainContent
	MaxConcurrentStreams  int    // 单连接最大并发流，>0 才设置
	ReadBufferSize        int    // gRPC 读缓冲（字节），>0 才设置
	WriteBufferSize       int    // gRPC 写缓冲（字节），>0 才设置
	HealthCheckInterval   int    // Evictor 扫描间隔（秒），<=0 时 30s
	HealthCheckTimeout    int    // 心跳超时判定（秒），<=0 时 15s
	EnableAuth            string // 数据面是否强制鉴权，Y/N
	IPWhitelist           string // IP/CIDR 白名单，逗号或 JSON 数组；非空则只放行名单内地址
	IPBlacklist           string // IP/CIDR 黑名单，逗号或 JSON 数组；命中即拒绝（先于白名单）
	ExtProperty           string // 扩展 JSON：alert* 开关
	Status                string // 运行状态，见 CenterStatus*
	StatusMessage         string // 状态说明
	ActiveFlag            string // 是否启用定义，Y/N
}

// RuntimeKey 返回进程内池键（instanceName:environment），避免同名实例在不同环境下互相覆盖。
func (c *CenterInstance) RuntimeKey() string {
	if c == nil {
		return ""
	}
	if c.Environment == "" {
		return c.InstanceName
	}
	return c.InstanceName + ":" + c.Environment
}

// Alert 解析 ExtProperty 中的告警配置。
func (c *CenterInstance) Alert() AlertConfig {
	if c == nil {
		return DefaultAlertConfig()
	}
	return ParseAlertConfig(c.ExtProperty)
}

// ListenFingerprint 用于判断监听/TLS/鉴权是否变化（变化则 Reload 必须停端口再起）。
func (c *CenterInstance) ListenFingerprint() string {
	if c == nil {
		return ""
	}
	return c.ListenAddress + "|" + itoa(c.ListenPort) + "|" + c.EnableTLS + "|" +
		c.EnableMTLS + "|" + c.CertStorageType + "|" + c.EnableAuth + "|" +
		c.IPWhitelist + "|" + c.IPBlacklist + "|" + itoa(c.MaxRecvMsgSize) + "|" +
		itoa(c.MaxSendMsgSize) + "|" + itoa(c.ReadBufferSize) + "|" + itoa(c.WriteBufferSize) + "|" +
		itoa(c.HealthCheckInterval) + "|" + itoa(c.HealthCheckTimeout)
}

// ListenEndpoint 返回 host:port，供管理面与日志使用。
func (c *CenterInstance) ListenEndpoint() string {
	if c == nil {
		return ""
	}
	return formatEndpoint(c.ListenAddress, c.ListenPort)
}

// formatEndpoint 拼 host:port，空 host 视为 0.0.0.0。
func formatEndpoint(host string, port int) string {
	if host == "" {
		host = "0.0.0.0"
	}
	return host + ":" + itoa(port)
}

// itoa 将整数转为十进制字符串，避免为端口格式化引入 strconv。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
