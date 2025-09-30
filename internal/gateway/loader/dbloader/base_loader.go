package dbloader

import (
	"context"
	"fmt"
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
		if instance.CertStorageType == "FILE" {
			if instance.CertFilePath != nil {
				baseConfig.CertFile = *instance.CertFilePath
			}
			if instance.KeyFilePath != nil {
				baseConfig.KeyFile = *instance.KeyFilePath
			}
		} else if instance.CertStorageType == "DATABASE" {
			// TODO: 处理数据库存储的证书内容
			// 这里需要将证书内容写入临时文件或者使用内存证书
		}
	}

	return baseConfig
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
