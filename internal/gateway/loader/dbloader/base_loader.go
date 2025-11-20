package dbloader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gateway/internal/gateway/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// BaseConfigLoader 基础配置加载器
type BaseConfigLoader struct {
	db       database.Database
	tenantId string
}

// NewBaseConfigLoader 创建基础配置加载器
func NewBaseConfigLoader(db database.Database, tenantId string) *BaseConfigLoader {
	return &BaseConfigLoader{
		db:       db,
		tenantId: tenantId,
	}
}

// LoadGatewayInstance 加载网关实例配置
func (loader *BaseConfigLoader) LoadGatewayInstance(ctx context.Context, instanceId string) (*GatewayInstanceRecord, error) {
	query := `
		SELECT tenantId, gatewayInstanceId, instanceName, instanceDesc, bindAddress,
		       httpPort, httpsPort, tlsEnabled, certStorageType, certFilePath,
		       keyFilePath, certContent, keyContent, certChainContent, certPassword,
		       maxConnections, readTimeoutMs, writeTimeoutMs, idleTimeoutMs, maxHeaderBytes,
		       maxWorkers, keepAliveEnabled, tcpKeepAliveEnabled, gracefulShutdownTimeoutMs,
		       enableHttp2, tlsVersion, tlsCipherSuites, disableGeneralOptionsHandler,
		       logConfigId, healthStatus, lastHeartbeatTime, instanceMetadata, activeFlag
		FROM HUB_GW_INSTANCE 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
	`

	var instance GatewayInstanceRecord
	err := loader.db.QueryOne(ctx, &instance, query, []interface{}{loader.tenantId, instanceId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询网关实例失败: %w", err)
	}

	return &instance, nil
}

// BuildBaseConfig 构建基础配置
func (loader *BaseConfigLoader) BuildBaseConfig(instance *GatewayInstanceRecord) config.BaseConfig {
	// 确定监听地址
	var listen string
	if instance.TLSEnabled == "Y" && instance.HTTPSPort != nil {
		listen = fmt.Sprintf("%s:%d", instance.BindAddress, *instance.HTTPSPort)
	} else if instance.HTTPPort != nil {
		listen = fmt.Sprintf("%s:%d", instance.BindAddress, *instance.HTTPPort)
	} else {
		// 默认使用HTTP 8080端口
		listen = fmt.Sprintf("%s:8080", instance.BindAddress)
	}

	baseConfig := config.BaseConfig{
		Listen:          listen,
		Name:            instance.InstanceName,
		ReadTimeout:     time.Duration(instance.ReadTimeoutMs) * time.Millisecond,
		WriteTimeout:    time.Duration(instance.WriteTimeoutMs) * time.Millisecond,
		IdleTimeout:     time.Duration(instance.IdleTimeoutMs) * time.Millisecond,
		MaxBodySize:     int64(instance.MaxHeaderBytes),
		EnableHTTPS:     instance.TLSEnabled == "Y",
		UseGin:          instance.DisableGeneralOptionsHandler != "Y", // 如果禁用了通用OPTIONS处理器，则不使用Gin
		EnableAccessLog: true,
		LogFormat:       "json",
		LogLevel:        "info",
		EnableGzip:      true,
	}

	// 处理TLS相关配置
	if instance.TLSEnabled == "Y" {
		// 设置私钥密码（如果有）
		if instance.CertPassword != nil && *instance.CertPassword != "" {
			baseConfig.KeyPassword = *instance.CertPassword
		}

		if instance.CertStorageType == "FILE" {
			// 文件存储模式：直接使用配置的文件路径
			if instance.CertFilePath != nil {
				baseConfig.CertFile = *instance.CertFilePath
			}
			if instance.KeyFilePath != nil {
				baseConfig.KeyFile = *instance.KeyFilePath
			}
		} else if instance.CertStorageType == "DATABASE" {
			// 数据库存储模式：将证书内容写入临时文件
			if err := loader.writeCertificatesToFiles(instance, &baseConfig); err != nil {
				logger.Error("写入数据库证书到文件失败", err, "instanceId", instance.InstanceId)
			}
		}
	}

	return baseConfig
}

// writeCertificatesToFiles 将数据库中的证书内容写入临时文件
func (loader *BaseConfigLoader) writeCertificatesToFiles(instance *GatewayInstanceRecord, baseConfig *config.BaseConfig) error {
	// 创建临时目录用于存储证书文件
	tempDir := filepath.Join(os.TempDir(), "gateway-certs", loader.tenantId, instance.InstanceId)
	if err := os.MkdirAll(tempDir, 0700); err != nil {
		return fmt.Errorf("创建证书临时目录失败: %w", err)
	}

	// 写入证书文件
	if instance.CertContent != nil && *instance.CertContent != "" {
		// 使用数据库中保存的原始文件名，如果没有则使用默认名称
		certFileName := "cert.pem"
		if instance.CertFilePath != nil && *instance.CertFilePath != "" {
			// 提取文件名（去除可能的路径前缀）
			certFileName = filepath.Base(*instance.CertFilePath)
		}

		certPath := filepath.Join(tempDir, certFileName)
		if err := os.WriteFile(certPath, []byte(*instance.CertContent), 0600); err != nil {
			return fmt.Errorf("写入证书文件失败: %w", err)
		}
		baseConfig.CertFile = certPath
		logger.Debug("证书文件已写入", "path", certPath, "fileName", certFileName, "instanceId", instance.InstanceId)
	}

	// 写入私钥文件
	if instance.KeyContent != nil && *instance.KeyContent != "" {
		// 使用数据库中保存的原始文件名，如果没有则使用默认名称
		keyFileName := "key.pem"
		if instance.KeyFilePath != nil && *instance.KeyFilePath != "" {
			// 提取文件名（去除可能的路径前缀）
			keyFileName = filepath.Base(*instance.KeyFilePath)
		}

		keyPath := filepath.Join(tempDir, keyFileName)
		if err := os.WriteFile(keyPath, []byte(*instance.KeyContent), 0600); err != nil {
			return fmt.Errorf("写入私钥文件失败: %w", err)
		}
		baseConfig.KeyFile = keyPath
		logger.Debug("私钥文件已写入", "path", keyPath, "fileName", keyFileName, "instanceId", instance.InstanceId)
	}

	// 如果有证书链内容，也写入文件（可选）
	if instance.CertChainContent != nil && *instance.CertChainContent != "" {
		chainPath := filepath.Join(tempDir, "chain.pem")
		if err := os.WriteFile(chainPath, []byte(*instance.CertChainContent), 0600); err != nil {
			logger.Warn("写入证书链文件失败", "error", err, "instanceId", instance.InstanceId)
		} else {
			logger.Debug("证书链文件已写入", "path", chainPath, "instanceId", instance.InstanceId)
		}
	}

	return nil
}

// UpdateGatewayHealthStatus 更新网关实例健康状态（静态方法）
// 使用默认数据库连接更新指定租户和实例的健康状态
func UpdateGatewayHealthStatus(tenantId, instanceId, healthStatus, errorMsg string) {
	// 获取默认数据库连接
	db := database.GetDefaultConnection()
	if db == nil {
		logger.Warn("无法获取默认数据库连接，跳过健康状态更新", "instanceId", instanceId)
		return
	}

	// 使用超时上下文避免阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 生成当前时间，不使用数据库函数
	now := time.Now()

	// 限制错误信息长度，避免超出reserved1字段限制（通常100字符）
	if len(errorMsg) > 100 {
		errorMsg = errorMsg[:97] + "..."
	}

	var query string
	var args []interface{}

	// 根据是否有错误信息构建不同的SQL
	if errorMsg != "" {
		query = `
			UPDATE HUB_GW_INSTANCE 
			SET healthStatus = ?, lastHeartbeatTime = ?, reserved1 = ?
			WHERE tenantId = ? AND gatewayInstanceId = ?
		`
		args = []interface{}{healthStatus, now, errorMsg, tenantId, instanceId}
	} else {
		query = `
			UPDATE HUB_GW_INSTANCE 
			SET healthStatus = ?, lastHeartbeatTime = ?, reserved1 = NULL
			WHERE tenantId = ? AND gatewayInstanceId = ?
		`
		args = []interface{}{healthStatus, now, tenantId, instanceId}
	}

	// 执行更新
	_, err := db.Exec(ctx, query, args, true)
	if err != nil {
		logger.Warn("更新网关实例健康状态失败", "error", err, "instanceId", instanceId, "healthStatus", healthStatus)
	} else {
		logger.Debug("网关实例健康状态已更新", "instanceId", instanceId, "healthStatus", healthStatus)
	}
}
