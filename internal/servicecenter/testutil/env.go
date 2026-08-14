// Package testutil 提供服务中心端到端测试环境：SQLite 建库、种子数据与 gRPC 实例启停。
package testutil

import (
	"context"
	"embed"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gateway/internal/servicecenter/cache"
	"gateway/internal/servicecenter/manager"
	"gateway/pkg/database"
	"gateway/pkg/database/dbtypes"
	_ "gateway/pkg/database/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

const (
	// DefaultTenantID 测试租户
	DefaultTenantID = "default"
	// DefaultInstanceName 测试实例名
	DefaultInstanceName = "service-center-e2e"
	// DefaultEnvironment 测试环境
	DefaultEnvironment = "DEVELOPMENT"
	// DefaultGroupName 默认分组
	DefaultGroupName = "e2e-group"

	// DefaultAuthUserID Basic / JWT / API Token 关联用户
	DefaultAuthUserID = "e2e-user"
	// DefaultAuthPassword Basic 明文密码（库内明文，ValidateUser 支持）
	DefaultAuthPassword = "e2e-pass"
	// DefaultAuthAPIToken Bearer API Token 明文值
	DefaultAuthAPIToken = "e2e-api-token-7f3c9a1b"
	// DefaultAuthJwtSecret JWT HS256 密钥（写入实例 extProperty.authJwtSecret）
	DefaultAuthJwtSecret = "e2e-jwt-secret-change-me"
	// DefaultAuthJwtIssuer JWT issuer
	DefaultAuthJwtIssuer = "service-center"
)

// Env 已启动的服务中心测试环境
type Env struct {
	DB           database.Database
	Manager      *manager.ServiceCenterManager
	TempDir      string
	Host         string
	Port         int
	NamespaceID  string
	GroupName    string
	TenantID     string
	InstanceName string
	Environment  string

	// 认证相关（仅 EnableAuth=true 时有值）
	EnableAuth    bool
	AuthUserID    string
	AuthPassword  string
	AuthAPIToken  string
	AuthJwtSecret string
	AuthJwtIssuer string

	// TLS 相关（仅 EnableTLS=true 时有值）
	EnableTLS  bool
	EnableMTLS bool
	TLS        *TLSMaterial
}

// Options 启动选项
type Options struct {
	// NamespaceID 命名空间；空则自动生成
	NamespaceID string
	// GroupName 分组；空则用 DefaultGroupName
	GroupName string
	// Host 监听地址；空则 127.0.0.1
	Host string
	// EnableAuth 为 true 时开启实例认证并种子用户 / API Token / JWT 密钥
	EnableAuth bool
	// EnableTLS 为 true 时生成自签证书并开启服务端 TLS
	EnableTLS bool
	// EnableMTLS 为 true 时要求客户端证书（隐含 EnableTLS）
	EnableMTLS bool
}

// Start 使用临时 SQLite 文件启动服务中心，监听本机空闲端口。
// 默认无 TLS、无认证；由 Options.EnableAuth / EnableTLS / EnableMTLS 控制。
func Start(ctx context.Context, opts Options) (*Env, error) {
	if opts.EnableMTLS {
		opts.EnableTLS = true
	}
	if opts.Host == "" {
		opts.Host = "127.0.0.1"
	}
	if opts.GroupName == "" {
		opts.GroupName = DefaultGroupName
	}
	if opts.NamespaceID == "" {
		opts.NamespaceID = fmt.Sprintf("ns_e2e_%d", time.Now().UnixNano())
	}

	port, err := freePort()
	if err != nil {
		return nil, fmt.Errorf("分配空闲端口失败: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "servicecenter_e2e_*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	var tlsMat *TLSMaterial
	if opts.EnableTLS {
		certDir := filepath.Join(tempDir, "certs")
		mat, genErr := GenerateTLSMaterial(certDir)
		if genErr != nil {
			_ = os.RemoveAll(tempDir)
			return nil, fmt.Errorf("生成 TLS 证书失败: %w", genErr)
		}
		tlsMat = mat
	}

	dbPath := filepath.Join(tempDir, "servicecenter.db")
	db, err := openSQLite(dbPath)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	if err := applySchema(ctx, db); err != nil {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	if err := seedInstanceAndNamespace(ctx, db, port, opts, tlsMat); err != nil {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	// 清空进程内全局缓存，避免多次 Start 互相污染
	cache.GetGlobalCache().Clear(ctx)

	mgr := manager.NewServiceCenterManager(db)
	if err := mgr.LoadAllInstancesFromDB(ctx, DefaultTenantID); err != nil {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
		return nil, fmt.Errorf("加载实例失败: %w", err)
	}

	if err := mgr.StartInstance(ctx, DefaultTenantID, DefaultInstanceName, DefaultEnvironment); err != nil {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
		return nil, fmt.Errorf("启动实例失败: %w", err)
	}

	// 等待端口可连
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, dialErr := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", opts.Host, port), 200*time.Millisecond)
		if dialErr == nil {
			_ = conn.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	env := &Env{
		DB:           db,
		Manager:      mgr,
		TempDir:      tempDir,
		Host:         opts.Host,
		Port:         port,
		NamespaceID:  opts.NamespaceID,
		GroupName:    opts.GroupName,
		TenantID:     DefaultTenantID,
		InstanceName: DefaultInstanceName,
		Environment:  DefaultEnvironment,
		EnableAuth:   opts.EnableAuth,
		EnableTLS:    opts.EnableTLS,
		EnableMTLS:   opts.EnableMTLS,
		TLS:          tlsMat,
	}
	if opts.EnableAuth {
		env.AuthUserID = DefaultAuthUserID
		env.AuthPassword = DefaultAuthPassword
		env.AuthAPIToken = DefaultAuthAPIToken
		env.AuthJwtSecret = DefaultAuthJwtSecret
		env.AuthJwtIssuer = DefaultAuthJwtIssuer
	}
	return env, nil
}

// Stop 停止实例并清理临时库
func (e *Env) Stop(ctx context.Context) error {
	if e == nil {
		return nil
	}
	var firstErr error
	if e.Manager != nil {
		if err := e.Manager.StopInstance(ctx, e.InstanceName); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if e.DB != nil {
		if err := e.DB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if e.TempDir != "" {
		_ = os.RemoveAll(e.TempDir)
	}
	cache.GetGlobalCache().Clear(ctx)
	return firstErr
}

// Endpoint 返回 host:port
func (e *Env) Endpoint() string {
	return fmt.Sprintf("%s:%d", e.Host, e.Port)
}

func openSQLite(dbPath string) (database.Database, error) {
	cfg := &database.DbConfig{
		Driver:  database.DriverSQLite,
		Name:    fmt.Sprintf("sc_e2e_%d", time.Now().UnixNano()),
		Enabled: true,
		DSN:     dbPath,
		Pool: dbtypes.PoolConfig{
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 1800,
		},
		Log: dbtypes.LogConfig{
			Enable:        false,
			SlowThreshold: 200,
		},
	}
	db, err := database.Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("打开 SQLite 失败: %w", err)
	}
	if err := db.Ping(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("Ping SQLite 失败: %w", err)
	}
	return db, nil
}

func applySchema(ctx context.Context, db database.Database) error {
	raw, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("读取 schema.sql 失败: %w", err)
	}
	// SQLite 驱动通常一次执行一条语句
	for _, stmt := range splitSQLStatements(string(raw)) {
		if _, err := db.Exec(ctx, stmt, nil, true); err != nil {
			return fmt.Errorf("执行 DDL 失败: %w\nSQL: %s", err, stmt)
		}
	}
	return nil
}

func seedInstanceAndNamespace(ctx context.Context, db database.Database, port int, opts Options, tlsMat *TLSMaterial) error {
	enableAuth := "N"
	extProperty := ""
	if opts.EnableAuth {
		enableAuth = "Y"
		extProperty = fmt.Sprintf(`{"authJwtSecret":"%s","authJwtIssuer":"%s"}`,
			DefaultAuthJwtSecret, DefaultAuthJwtIssuer)
	}

	enableTLS := "N"
	enableMTLS := "N"
	certFile := ""
	keyFile := ""
	certChain := ""
	if opts.EnableTLS && tlsMat != nil {
		enableTLS = "Y"
		certFile = tlsMat.ServerCert
		keyFile = tlsMat.ServerKey
		certChain = tlsMat.CACertFile
	}
	if opts.EnableMTLS {
		enableMTLS = "Y"
	}

	instanceSQL := `
INSERT INTO HUB_SERVICE_INSTANCE (
    tenantId, instanceName, environment, serverType,
    listenAddress, listenPort,
    maxRecvMsgSize, maxSendMsgSize,
    keepAliveTime, keepAliveTimeout, keepAliveMinTime, permitWithoutStream,
    maxConnectionIdle, maxConnectionAge, maxConnectionAgeGrace,
    enableReflection, enableTLS, certStorageType, certFilePath, keyFilePath, certChainContent, enableMTLS,
    maxConcurrentStreams, readBufferSize, writeBufferSize,
    healthCheckInterval, healthCheckTimeout,
    instanceStatus, statusMessage, enableAuth, extProperty,
    addWho, editWho, oprSeqFlag, currentVersion, activeFlag, noteText
) VALUES (
    ?, ?, ?, 'GRPC',
    ?, ?,
    16777216, 16777216,
    30, 10, 5, 'Y',
    0, 0, 20,
    'Y', ?, 'FILE', ?, ?, ?, ?,
    250, 32768, 32768,
    30, 5,
    'STOPPED', 'e2e', ?, ?,
    'e2e', 'e2e', 'E2E', 1, 'Y', 'sqlite e2e instance'
)`
	if _, err := db.Exec(ctx, instanceSQL, []interface{}{
		DefaultTenantID, DefaultInstanceName, DefaultEnvironment,
		opts.Host, port,
		enableTLS, certFile, keyFile, certChain, enableMTLS,
		enableAuth, extProperty,
	}, true); err != nil {
		return fmt.Errorf("插入实例失败: %w", err)
	}

	nsSQL := `
INSERT INTO HUB_SERVICE_NAMESPACE (
    namespaceId, tenantId, instanceName, environment,
    namespaceName, namespaceDescription,
    serviceQuotaLimit, configQuotaLimit,
    addWho, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (?, ?, ?, ?, ?, ?, 200, 200, 'e2e', 'e2e', 'E2E', 1, 'Y')`
	if _, err := db.Exec(ctx, nsSQL, []interface{}{
		opts.NamespaceID, DefaultTenantID, DefaultInstanceName, DefaultEnvironment,
		opts.NamespaceID, "e2e namespace",
	}, true); err != nil {
		return fmt.Errorf("插入命名空间失败: %w", err)
	}

	if opts.EnableAuth {
		if err := seedAuthCredentials(ctx, db); err != nil {
			return err
		}
	}
	return nil
}

func seedAuthCredentials(ctx context.Context, db database.Database) error {
	userSQL := `
INSERT INTO HUB_USER (
    userId, tenantId, userName, password, realName, deptId,
    statusFlag, userExpireDate, pwdErrorCount,
    addWho, editWho, oprSeqFlag, currentVersion, activeFlag, noteText
) VALUES (?, ?, ?, ?, ?, 'D00000001', 'Y', '2099-12-31 23:59:59', 0,
    'e2e', 'e2e', 'E2E', 1, 'Y', 'e2e auth user')`
	if _, err := db.Exec(ctx, userSQL, []interface{}{
		DefaultAuthUserID, DefaultTenantID, DefaultAuthUserID, DefaultAuthPassword, "E2E User",
	}, true); err != nil {
		return fmt.Errorf("插入认证用户失败: %w", err)
	}

	tokenSQL := `
INSERT INTO HUB_SERVICE_AUTH_TOKEN (
    tokenId, tenantId, tokenValue, userId, tokenName, expireTime, statusFlag,
    addWho, editWho, oprSeqFlag, currentVersion, activeFlag, noteText
) VALUES (?, ?, ?, ?, 'e2e-api-token', NULL, 'Y',
    'e2e', 'e2e', 'E2E', 1, 'Y', 'e2e api token')`
	if _, err := db.Exec(ctx, tokenSQL, []interface{}{
		"tok-e2e-001", DefaultTenantID, DefaultAuthAPIToken, DefaultAuthUserID,
	}, true); err != nil {
		return fmt.Errorf("插入 API Token 失败: %w", err)
	}
	return nil
}

func freePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}

func splitSQLStatements(script string) []string {
	// 先去掉行注释，避免注释中的分号被误判为语句结束
	var cleaned strings.Builder
	for _, line := range strings.Split(script, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "--") {
			continue
		}
		cleaned.WriteString(line)
		cleaned.WriteByte('\n')
	}

	parts := strings.Split(cleaned.String(), ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		stmt := strings.TrimSpace(p)
		if stmt != "" {
			out = append(out, stmt)
		}
	}
	return out
}
