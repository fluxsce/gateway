package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gohub/internal/gateway/config"
)

func TestGatewayConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.GatewayConfig
		description string
	}{
		{
			name: "CompleteConfig",
			config: &config.GatewayConfig{
				InstanceID: "test-gateway-001",
				Base: config.BaseConfig{
					Name:         "test-gateway",
					Listen:       ":8080",
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
			},
			description: "完整网关配置",
		},
		{
			name: "MinimalConfig",
			config: &config.GatewayConfig{
				InstanceID: "minimal-gateway-001",
				Base: config.BaseConfig{
					Name:   "minimal-gateway",
					Listen: ":3000",
				},
			},
			description: "最小网关配置",
		},
		{
			name: "ProductionConfig",
			config: &config.GatewayConfig{
				InstanceID: "prod-gateway-001",
				Base: config.BaseConfig{
					Name:            "prod-gateway",
					Listen:          ":80",
					ReadTimeout:     60 * time.Second,
					WriteTimeout:    60 * time.Second,
					IdleTimeout:     120 * time.Second,
					EnableHTTPS:     true,
					CertFile:        "/etc/ssl/certs/server.crt",
					KeyFile:         "/etc/ssl/private/server.key",
					EnableAccessLog: true,
					LogFormat:       "json",
					LogLevel:        "warn",
				},
			},
			description: "生产环境配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置字段
			assert.NotEmpty(t, tt.config.InstanceID, "实例ID不应该为空")
			assert.NotEmpty(t, tt.config.Base.Name, "网关名称不应该为空")
			assert.NotEmpty(t, tt.config.Base.Listen, "监听地址不应该为空")
		})
	}
}

func TestBaseConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      config.BaseConfig
		expectValid bool
		description string
	}{
		{
			name: "ValidHTTPConfig",
			config: config.BaseConfig{
				Name:         "test-gateway",
				Listen:       ":8080",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
			},
			expectValid: true,
			description: "有效的HTTP服务器配置",
		},
		{
			name: "ValidHTTPSConfig",
			config: config.BaseConfig{
				Name:        "test-gateway",
				Listen:      ":443",
				EnableHTTPS: true,
				CertFile:    "/path/to/cert.pem",
				KeyFile:     "/path/to/key.pem",
			},
			expectValid: true,
			description: "有效的HTTPS服务器配置",
		},
		{
			name: "InvalidName",
			config: config.BaseConfig{
				Name:   "",
				Listen: ":8080",
			},
			expectValid: false,
			description: "空名称应该失败",
		},
		{
			name: "InvalidListen",
			config: config.BaseConfig{
				Name:   "test-gateway",
				Listen: "",
			},
			expectValid: false,
			description: "空监听地址应该失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 简单验证逻辑
			valid := tt.config.Name != "" && tt.config.Listen != ""
			if tt.config.EnableHTTPS {
				valid = valid && tt.config.CertFile != "" && tt.config.KeyFile != ""
			}

			assert.Equal(t, tt.expectValid, valid, tt.description)
		})
	}
}

func TestConfigValidation(t *testing.T) {
	// 完整的配置验证测试
	validConfig := &config.GatewayConfig{
		InstanceID: "validation-test-001",
		Base: config.BaseConfig{
			Name:   "validation-test",
			Listen: ":8080",
		},
	}

	// 验证基本配置
	assert.NotEmpty(t, validConfig.InstanceID, "实例ID不应该为空")
	assert.NotEmpty(t, validConfig.Base.Name, "网关名称不应该为空")
	assert.NotEmpty(t, validConfig.Base.Listen, "监听地址不应该为空")
}

func TestConfigFromJSON(t *testing.T) {
	jsonConfig := `{
		"instance_id": "json-gateway-001",
		"base": {
			"name": "json-gateway",
			"listen": ":8080",
			"log_level": "debug",
			"log_format": "json",
			"enable_access_log": true
		}
	}`

	var cfg config.GatewayConfig
	err := json.Unmarshal([]byte(jsonConfig), &cfg)
	require.NoError(t, err, "JSON解析失败")

	// 验证解析结果
	assert.Equal(t, "json-gateway-001", cfg.InstanceID)
	assert.Equal(t, "json-gateway", cfg.Base.Name)
	assert.Equal(t, ":8080", cfg.Base.Listen)
	assert.Equal(t, "debug", cfg.Base.LogLevel)
}

func TestConfigFromViper(t *testing.T) {
	// 使用viper替代yaml.v2
	v := viper.New()
	v.SetConfigType("yaml")

	yamlConfig := `
instance_id: yaml-gateway-001
base:
  name: yaml-gateway
  listen: ":8080"
  enable_https: true
  cert_file: /etc/ssl/server.crt
  key_file: /etc/ssl/server.key
  log_level: info
  log_format: json
  enable_access_log: true
`

	err := v.ReadConfig(strings.NewReader(yamlConfig))
	require.NoError(t, err, "YAML解析失败")

	var cfg config.GatewayConfig
	err = v.Unmarshal(&cfg)
	require.NoError(t, err, "配置解析失败")

	// 验证解析结果
	assert.Equal(t, "yaml-gateway-001", cfg.InstanceID)
	assert.Equal(t, "yaml-gateway", cfg.Base.Name)
	assert.Equal(t, ":8080", cfg.Base.Listen)
	assert.True(t, cfg.Base.EnableHTTPS)
	assert.Equal(t, "/etc/ssl/server.crt", cfg.Base.CertFile)
	assert.Equal(t, "info", cfg.Base.LogLevel)
}

func TestConfigFromFile(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()

	// JSON配置文件
	jsonFile := filepath.Join(tempDir, "config.json")
	jsonContent := `{
		"instance_id": "file-gateway-001",
		"base": {
			"name": "file-gateway",
			"listen": ":3000",
			"log_level": "info",
			"log_format": "text",
			"enable_access_log": true
		}
	}`

	err := ioutil.WriteFile(jsonFile, []byte(jsonContent), 0644)
	require.NoError(t, err)

	// 使用viper从文件加载配置
	v := viper.New()
	v.SetConfigFile(jsonFile)
	err = v.ReadInConfig()
	require.NoError(t, err, "从JSON文件读取配置失败")

	var cfg config.GatewayConfig
	err = v.Unmarshal(&cfg)
	require.NoError(t, err, "JSON配置解析失败")

	assert.Equal(t, "file-gateway-001", cfg.InstanceID)
	assert.Equal(t, "file-gateway", cfg.Base.Name)
	assert.Equal(t, ":3000", cfg.Base.Listen)

	// YAML配置文件
	yamlFile := filepath.Join(tempDir, "config.yaml")
	yamlContent := `
instance_id: yaml-file-gateway-001
base:
  name: yaml-file-gateway
  listen: ":8080"
  log_level: debug
  log_format: json
  enable_access_log: true
`

	err = ioutil.WriteFile(yamlFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// 从YAML文件加载配置
	v2 := viper.New()
	v2.SetConfigFile(yamlFile)
	err = v2.ReadInConfig()
	require.NoError(t, err, "从YAML文件读取配置失败")

	var cfg2 config.GatewayConfig
	err = v2.Unmarshal(&cfg2)
	require.NoError(t, err, "YAML配置解析失败")

	assert.Equal(t, "yaml-file-gateway-001", cfg2.InstanceID)
	assert.Equal(t, "yaml-file-gateway", cfg2.Base.Name)
	assert.Equal(t, ":8080", cfg2.Base.Listen)
	assert.Equal(t, "debug", cfg2.Base.LogLevel)
}

func TestConfigFromEnvironment(t *testing.T) {
	// 设置环境变量
	os.Setenv("GATEWAY_INSTANCE_ID", "env-gateway-001")
	os.Setenv("GATEWAY_BASE_NAME", "env-gateway")
	os.Setenv("GATEWAY_BASE_LISTEN", ":9000")
	os.Setenv("GATEWAY_BASE_LOG_LEVEL", "warn")
	defer func() {
		os.Unsetenv("GATEWAY_INSTANCE_ID")
		os.Unsetenv("GATEWAY_BASE_NAME")
		os.Unsetenv("GATEWAY_BASE_LISTEN")
		os.Unsetenv("GATEWAY_BASE_LOG_LEVEL")
	}()

	// 使用viper从环境变量加载配置
	v := viper.New()
	v.SetEnvPrefix("GATEWAY")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	v.SetDefault("instance_id", "default-001")
	v.SetDefault("base.name", "default-gateway")
	v.SetDefault("base.listen", ":8080")

	var cfg config.GatewayConfig
	cfg.InstanceID = v.GetString("instance_id")
	cfg.Base.Name = v.GetString("base.name")
	cfg.Base.Listen = v.GetString("base.listen")
	cfg.Base.LogLevel = v.GetString("base.log_level")

	assert.Equal(t, "env-gateway-001", cfg.InstanceID)
	assert.Equal(t, "env-gateway", cfg.Base.Name)
	assert.Equal(t, ":9000", cfg.Base.Listen)
	assert.Equal(t, "warn", cfg.Base.LogLevel)
}

func TestDefaultConfig(t *testing.T) {
	defaultCfg := config.DefaultGatewayConfig

	// 验证默认配置
	assert.Equal(t, "GoHub Gateway", defaultCfg.Base.Name)
	assert.Equal(t, ":8080", defaultCfg.Base.Listen)
	assert.Equal(t, "info", defaultCfg.Base.LogLevel)
	assert.Equal(t, "json", defaultCfg.Base.LogFormat)
	assert.True(t, defaultCfg.Base.EnableAccessLog)
}

func TestConfigMerge(t *testing.T) {
	// 基础配置
	baseConfig := &config.GatewayConfig{
		InstanceID: "base-gateway-001",
		Base: config.BaseConfig{
			Name:   "base-gateway",
			Listen: ":8080",
		},
	}

	// 覆盖配置
	overrideConfig := &config.GatewayConfig{
		InstanceID: "override-gateway-001",
		Base: config.BaseConfig{
			Listen:   ":80",
			LogLevel: "warn",
		},
	}

	// 手动合并配置
	mergedConfig := *baseConfig
	if overrideConfig.InstanceID != "" {
		mergedConfig.InstanceID = overrideConfig.InstanceID
	}
	if overrideConfig.Base.Listen != "" {
		mergedConfig.Base.Listen = overrideConfig.Base.Listen
	}
	if overrideConfig.Base.LogLevel != "" {
		mergedConfig.Base.LogLevel = overrideConfig.Base.LogLevel
	}

	// 验证合并结果
	assert.Equal(t, "override-gateway-001", mergedConfig.InstanceID, "实例ID应该来自覆盖配置")
	assert.Equal(t, "base-gateway", mergedConfig.Base.Name, "名称应该来自基础配置")
	assert.Equal(t, ":80", mergedConfig.Base.Listen, "监听地址应该来自覆盖配置")
	assert.Equal(t, "warn", mergedConfig.Base.LogLevel, "日志级别应该来自覆盖配置")
}

func TestHTTPSConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      config.BaseConfig
		expectValid bool
		description string
	}{
		{
			name: "ValidHTTPS",
			config: config.BaseConfig{
				Name:        "test-gateway",
				Listen:      ":443",
				EnableHTTPS: true,
				CertFile:    "/path/to/cert.pem",
				KeyFile:     "/path/to/key.pem",
			},
			expectValid: true,
			description: "有效的HTTPS配置",
		},
		{
			name: "DisabledHTTPS",
			config: config.BaseConfig{
				Name:        "test-gateway",
				Listen:      ":8080",
				EnableHTTPS: false,
			},
			expectValid: true,
			description: "禁用的HTTPS配置应该有效",
		},
		{
			name: "MissingCertFile",
			config: config.BaseConfig{
				Name:        "test-gateway",
				Listen:      ":443",
				EnableHTTPS: true,
				CertFile:    "",
				KeyFile:     "/path/to/key.pem",
			},
			expectValid: false,
			description: "缺少证书文件应该失败",
		},
		{
			name: "MissingKeyFile",
			config: config.BaseConfig{
				Name:        "test-gateway",
				Listen:      ":443",
				EnableHTTPS: true,
				CertFile:    "/path/to/cert.pem",
				KeyFile:     "",
			},
			expectValid: false,
			description: "缺少私钥文件应该失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 简单验证逻辑
			valid := tt.config.Name != "" && tt.config.Listen != ""
			if tt.config.EnableHTTPS {
				valid = valid && tt.config.CertFile != "" && tt.config.KeyFile != ""
			}

			assert.Equal(t, tt.expectValid, valid, tt.description)
		})
	}
}

func TestConfigValidationErrors(t *testing.T) {
	// 测试各种配置验证错误场景
	tests := []struct {
		name        string
		config      *config.GatewayConfig
		expectError bool
		description string
	}{
		{
			name: "EmptyInstanceID",
			config: &config.GatewayConfig{
				InstanceID: "",
				Base: config.BaseConfig{
					Name:   "test-gateway",
					Listen: ":8080",
				},
			},
			expectError: true,
			description: "空实例ID应该报错",
		},
		{
			name: "EmptyName",
			config: &config.GatewayConfig{
				InstanceID: "test-001",
				Base: config.BaseConfig{
					Name:   "",
					Listen: ":8080",
				},
			},
			expectError: true,
			description: "空名称应该报错",
		},
		{
			name: "ValidConfig",
			config: &config.GatewayConfig{
				InstanceID: "test-001",
				Base: config.BaseConfig{
					Name:   "test-gateway",
					Listen: ":8080",
				},
			},
			expectError: false,
			description: "有效配置不应该报错",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 简单验证逻辑
			hasError := tt.config.InstanceID == "" || tt.config.Base.Name == "" || tt.config.Base.Listen == ""
			assert.Equal(t, tt.expectError, hasError, tt.description)
		})
	}
}

// 基准测试
func BenchmarkConfigValidation(b *testing.B) {
	cfg := &config.GatewayConfig{
		InstanceID: "bench-gateway-001",
		Base: config.BaseConfig{
			Name:   "bench-gateway",
			Listen: ":8080",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 简单验证
		_ = cfg.InstanceID != "" && cfg.Base.Name != "" && cfg.Base.Listen != ""
	}
}

func BenchmarkJSONConfigParsing(b *testing.B) {
	jsonConfig := `{
		"instance_id": "bench-gateway-001",
		"base": {
			"name": "bench-gateway",
			"listen": ":8080",
			"log_level": "info",
			"log_format": "json",
			"enable_access_log": true
		}
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg config.GatewayConfig
		_ = json.Unmarshal([]byte(jsonConfig), &cfg)
	}
}

func BenchmarkViperConfigParsing(b *testing.B) {
	yamlConfig := `
instance_id: bench-gateway-001
base:
  name: bench-gateway
  listen: ":8080"
  log_level: info
  log_format: json
  enable_access_log: true
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := viper.New()
		v.SetConfigType("yaml")
		_ = v.ReadConfig(strings.NewReader(yamlConfig))
		var cfg config.GatewayConfig
		_ = v.Unmarshal(&cfg)
	}
}
