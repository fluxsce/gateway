package loader_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gateway/internal/gateway/config"
	"gateway/internal/gateway/handler/auth"
	"gateway/internal/gateway/handler/cors"
	"gateway/internal/gateway/handler/limiter"
	"gateway/internal/gateway/handler/proxy"
	"gateway/internal/gateway/handler/router"
	"gateway/internal/gateway/handler/security"
	"gateway/internal/gateway/loader"
)

func TestNewYAMLConfigLoader(t *testing.T) {
	t.Run("创建YAML配置加载器", func(t *testing.T) {
		yamlLoader := loader.NewYAMLConfigLoader()
		assert.NotNil(t, yamlLoader)
	})
}

func TestYAMLConfigLoader_LoadConfig(t *testing.T) {
	yamlLoader := loader.NewYAMLConfigLoader()

	t.Run("加载简单的YAML配置文件", func(t *testing.T) {
		configPath := filepath.Join("testdata", "simple_test_config.yaml")

		cfg, err := yamlLoader.LoadConfig(configPath)
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// 验证配置的基本字段
		assert.Equal(t, "Simple Test Gateway", cfg.Base.Name)
		assert.Equal(t, ":8080", cfg.Base.Listen)
		assert.Equal(t, "simple-router", cfg.Router.ID)
		assert.Equal(t, "Simple Router", cfg.Router.Name)
		assert.True(t, cfg.Router.Enabled)
	})

	t.Run("加载不存在的YAML文件", func(t *testing.T) {
		cfg, err := yamlLoader.LoadConfig("nonexistent.yaml")
		require.NoError(t, err) // 应该返回默认配置，不报错
		assert.NotNil(t, cfg)

		// 验证是否为默认配置
		defaultCfg := config.DefaultGatewayConfig
		assert.Equal(t, defaultCfg.Base.Name, cfg.Base.Name)
	})

	t.Run("加载空路径", func(t *testing.T) {
		cfg, err := yamlLoader.LoadConfig("")
		require.NoError(t, err) // 应该返回默认配置
		assert.NotNil(t, cfg)
	})

	t.Run("加载无效的YAML文件", func(t *testing.T) {
		configPath := filepath.Join("testdata", "invalid_config.yaml")

		cfg, err := yamlLoader.LoadConfig(configPath)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("加载格式错误的YAML文件", func(t *testing.T) {
		// 创建格式错误的YAML文件
		tempDir := t.TempDir()
		badYAMLPath := filepath.Join(tempDir, "bad.yaml")

		badYAML := `
base:
  name: "Test Gateway
  listen: ":8080"
invalid_yaml: [unclosed bracket
`
		err := os.WriteFile(badYAMLPath, []byte(badYAML), 0644)
		require.NoError(t, err)

		cfg, err := yamlLoader.LoadConfig(badYAMLPath)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestYAMLConfigLoader_ValidateConfig(t *testing.T) {
	yamlLoader := loader.NewYAMLConfigLoader()

	t.Run("验证有效的YAML配置", func(t *testing.T) {
		configPath := filepath.Join("testdata", "simple_test_config.yaml")
		err := yamlLoader.ValidateConfig(configPath)
		assert.NoError(t, err)
	})

	t.Run("验证不存在的文件", func(t *testing.T) {
		err := yamlLoader.ValidateConfig("nonexistent.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "YAML配置文件验证失败")
	})

	t.Run("验证空路径", func(t *testing.T) {
		err := yamlLoader.ValidateConfig("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "YAML配置文件验证失败")
	})

	t.Run("验证格式错误的YAML文件", func(t *testing.T) {
		// 创建格式错误的YAML文件
		tempDir := t.TempDir()
		badYAMLPath := filepath.Join(tempDir, "bad.yaml")

		badYAML := `
base:
  name: "Test Gateway
  listen: ":8080"
invalid_yaml_structure
`
		err := os.WriteFile(badYAMLPath, []byte(badYAML), 0644)
		require.NoError(t, err)

		err = yamlLoader.ValidateConfig(badYAMLPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "YAML配置文件解析失败")
	})

	t.Run("验证错误扩展名的文件", func(t *testing.T) {
		tempDir := t.TempDir()
		txtFilePath := filepath.Join(tempDir, "config.txt")

		err := os.WriteFile(txtFilePath, []byte("test"), 0644)
		require.NoError(t, err)

		err = yamlLoader.ValidateConfig(txtFilePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "YAML配置源需要.yaml或.yml文件")
	})
}

func TestYAMLConfigLoader_GetSupportedExtensions(t *testing.T) {
	yamlLoader := loader.NewYAMLConfigLoader()

	t.Run("获取支持的扩展名", func(t *testing.T) {
		extensions := yamlLoader.GetSupportedExtensions()
		assert.Len(t, extensions, 2)
		assert.Contains(t, extensions, ".yaml")
		assert.Contains(t, extensions, ".yml")
	})
}

func TestYAMLConfigLoader_IsValidExtension(t *testing.T) {
	yamlLoader := loader.NewYAMLConfigLoader()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"有效的.yaml扩展名", "config.yaml", true},
		{"有效的.yml扩展名", "config.yml", true},
		{"无效的.json扩展名", "config.json", false},
		{"无效的.txt扩展名", "config.txt", false},
		{"无扩展名", "config", false},
		{"带路径的有效扩展名", "/path/to/config.yaml", true},
		{"带路径的无效扩展名", "/path/to/config.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := yamlLoader.IsValidExtension(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestYAMLConfigLoader_ParseYAMLString(t *testing.T) {
	yamlLoader := loader.NewYAMLConfigLoader()

	t.Run("解析有效的YAML字符串", func(t *testing.T) {
		yamlString := `
base:
  name: "Test Gateway"
  listen: ":8080"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_body_size: 10485760
  enable_https: false
  use_gin: true
  enable_access_log: true
  log_format: "json"
  log_level: "info"
  enable_gzip: true

router:
  id: "test-router"
  name: "Test Router"
  enabled: true

proxy:
  id: "test-proxy"
  name: "Test Proxy"
  enabled: true

auth:
  enabled: false
  strategy: "no_auth"

cors:
  enabled: true
  allow_origins: ["*"]
  allow_methods: ["GET", "POST", "PUT", "DELETE"]
  allow_headers: ["Content-Type", "Authorization"]
  max_age: 86400
`

		cfg, err := yamlLoader.ParseYAMLString(yamlString)
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// 验证解析结果
		assert.Equal(t, "Test Gateway", cfg.Base.Name)
		assert.Equal(t, ":8080", cfg.Base.Listen)
		assert.Equal(t, 30*time.Second, cfg.Base.ReadTimeout)
		assert.Equal(t, "test-router", cfg.Router.ID)
		assert.Equal(t, "Test Router", cfg.Router.Name)
		assert.True(t, cfg.Router.Enabled)
		assert.False(t, cfg.Auth.Enabled)
		assert.Equal(t, auth.StrategyNoAuth, cfg.Auth.Strategy)
		assert.True(t, cfg.CORS.Enabled)
		assert.Contains(t, cfg.CORS.AllowOrigins, "*")
	})

	t.Run("解析无效的YAML字符串", func(t *testing.T) {
		invalidYAML := `
base:
  name: "Test Gateway
  listen: ":8080"
invalid_yaml: [unclosed
`

		cfg, err := yamlLoader.ParseYAMLString(invalidYAML)
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "解析YAML字符串失败")
	})

	t.Run("解析空YAML字符串", func(t *testing.T) {
		cfg, err := yamlLoader.ParseYAMLString("")
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// 应该被默认值填充
		defaultCfg := config.DefaultGatewayConfig
		assert.Equal(t, defaultCfg.Base.Name, cfg.Base.Name)
		assert.Equal(t, defaultCfg.Base.Listen, cfg.Base.Listen)
	})

	t.Run("解析部分YAML配置", func(t *testing.T) {
		partialYAML := `
base:
  name: "Partial Gateway"
router:
  id: "partial-router"
`

		cfg, err := yamlLoader.ParseYAMLString(partialYAML)
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// 验证自定义值
		assert.Equal(t, "Partial Gateway", cfg.Base.Name)
		assert.Equal(t, "partial-router", cfg.Router.ID)

		// 验证默认值被填充
		defaultCfg := config.DefaultGatewayConfig
		assert.Equal(t, defaultCfg.Base.Listen, cfg.Base.Listen)
		assert.Equal(t, defaultCfg.Base.ReadTimeout, cfg.Base.ReadTimeout)
	})
}

func TestYAMLConfigLoader_ExportConfigToYAML(t *testing.T) {
	yamlLoader := loader.NewYAMLConfigLoader()

	t.Run("导出有效配置到YAML", func(t *testing.T) {
		cfg := &config.GatewayConfig{
			Base: config.BaseConfig{
				Name:         "Export Test Gateway",
				Listen:       ":9090",
				ReadTimeout:  45 * time.Second,
				WriteTimeout: 45 * time.Second,
				IdleTimeout:  180 * time.Second,
				MaxBodySize:  20 * 1024 * 1024, // 20MB
				EnableHTTPS:  true,
				UseGin:       true,
				LogFormat:    "json",
				LogLevel:     "debug",
				EnableGzip:   false,
			},
			Router: router.RouterConfig{
				ID:      "export-router",
				Name:    "Export Router",
				Enabled: true,
			},
			Proxy: proxy.ProxyConfig{
				ID:      "export-proxy",
				Name:    "Export Proxy",
				Enabled: true,
			},
		}

		yamlString, err := yamlLoader.ExportConfigToYAML(cfg)
		require.NoError(t, err)
		assert.NotEmpty(t, yamlString)

		// 验证导出的YAML包含关键信息
		assert.Contains(t, yamlString, "Export Test Gateway")
		assert.Contains(t, yamlString, ":9090")
		assert.Contains(t, yamlString, "export-router")
		assert.Contains(t, yamlString, "Export Router")

		// 验证导出的YAML可以重新解析
		parsedCfg, err := yamlLoader.ParseYAMLString(yamlString)
		require.NoError(t, err)
		assert.Equal(t, cfg.Base.Name, parsedCfg.Base.Name)
		assert.Equal(t, cfg.Base.Listen, parsedCfg.Base.Listen)
		assert.Equal(t, cfg.Router.ID, parsedCfg.Router.ID)
	})

	t.Run("导出空配置", func(t *testing.T) {
		yamlString, err := yamlLoader.ExportConfigToYAML(nil)
		assert.Error(t, err)
		assert.Empty(t, yamlString)
		assert.Contains(t, err.Error(), "导出YAML配置失败")
	})

	t.Run("导出默认配置", func(t *testing.T) {
		defaultCfg := config.DefaultGatewayConfig
		cfg := &defaultCfg

		yamlString, err := yamlLoader.ExportConfigToYAML(cfg)
		require.NoError(t, err)
		assert.NotEmpty(t, yamlString)

		// 验证默认配置的关键字段
		assert.Contains(t, yamlString, cfg.Base.Name)
		assert.Contains(t, yamlString, cfg.Base.Listen)
	})
}

func TestYAMLConfigLoader_SaveConfigToFile(t *testing.T) {
	yamlLoader := loader.NewYAMLConfigLoader()

	t.Run("保存配置到YAML文件", func(t *testing.T) {
		cfg := &config.GatewayConfig{
			Base: config.BaseConfig{
				Name:   "Save Test Gateway",
				Listen: ":7070",
			},
			Router: router.RouterConfig{
				ID:   "save-router",
				Name: "Save Router",
			},
		}

		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "save_test.yaml")

		err := yamlLoader.SaveConfigToFile(cfg, configPath)
		require.NoError(t, err)

		// 验证文件是否创建
		assert.FileExists(t, configPath)

		// 验证文件内容
		loadedCfg, err := yamlLoader.LoadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, cfg.Base.Name, loadedCfg.Base.Name)
		assert.Equal(t, cfg.Base.Listen, loadedCfg.Base.Listen)
		assert.Equal(t, cfg.Router.ID, loadedCfg.Router.ID)
	})

	t.Run("保存配置到不存在的目录", func(t *testing.T) {
		cfg := &config.GatewayConfig{
			Base: config.BaseConfig{
				Name: "Test Gateway",
			},
		}

		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "nested", "dir", "config.yaml")

		err := yamlLoader.SaveConfigToFile(cfg, configPath)
		require.NoError(t, err)

		// 验证文件和目录是否创建
		assert.FileExists(t, configPath)
	})

	t.Run("保存空配置", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "empty.yaml")

		err := yamlLoader.SaveConfigToFile(nil, configPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "导出YAML配置失败")
	})
}

func TestYAMLConfigLoader_ReloadConfig(t *testing.T) {
	yamlLoader := loader.NewYAMLConfigLoader()

	t.Run("重新加载配置", func(t *testing.T) {
		// 创建初始配置文件
		initialCfg := &config.GatewayConfig{
			Base: config.BaseConfig{
				Name:   "Initial Gateway",
				Listen: ":8080",
			},
		}

		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "reload_test.yaml")

		err := yamlLoader.SaveConfigToFile(initialCfg, configPath)
		require.NoError(t, err)

		// 第一次加载
		cfg1, err := yamlLoader.LoadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, "Initial Gateway", cfg1.Base.Name)

		// 修改配置文件
		modifiedCfg := &config.GatewayConfig{
			Base: config.BaseConfig{
				Name:   "Modified Gateway",
				Listen: ":9090",
			},
		}

		err = yamlLoader.SaveConfigToFile(modifiedCfg, configPath)
		require.NoError(t, err)

		// 重新加载
		cfg2, err := yamlLoader.ReloadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, "Modified Gateway", cfg2.Base.Name)
		assert.Equal(t, ":9090", cfg2.Base.Listen)
	})

	t.Run("重新加载不存在的文件", func(t *testing.T) {
		cfg, err := yamlLoader.ReloadConfig("nonexistent.yaml")
		require.NoError(t, err) // 应该返回默认配置
		assert.NotNil(t, cfg)
	})
}

func TestYAMLConfigLoader_Integration(t *testing.T) {
	t.Run("完整的YAML配置生命周期测试", func(t *testing.T) {
		yamlLoader := loader.NewYAMLConfigLoader()

		// 1. 创建复杂的配置
		originalCfg := &config.GatewayConfig{
			Base: config.BaseConfig{
				Name:            "Integration Test Gateway",
				Listen:          ":8888",
				ReadTimeout:     60 * time.Second,
				WriteTimeout:    60 * time.Second,
				IdleTimeout:     300 * time.Second,
				MaxBodySize:     50 * 1024 * 1024, // 50MB
				EnableHTTPS:     true,
				UseGin:          true,
				EnableAccessLog: true,
				LogFormat:       "json",
				LogLevel:        "info",
				EnableGzip:      true,
			},
			Router: router.RouterConfig{
				ID:      "integration-router",
				Name:    "Integration Router",
				Enabled: true,
			},
			Proxy: proxy.ProxyConfig{
				ID:      "integration-proxy",
				Name:    "Integration Proxy",
				Enabled: true,
			},
			Security: security.SecurityConfig{
				ID:      "integration-security",
				Enabled: true,
			},
			Auth: auth.AuthConfig{
				Enabled:  true,
				Strategy: auth.StrategyJWT,
			},
			CORS: cors.CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{"https://example.com", "https://api.example.com"},
				AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
				AllowHeaders: []string{"Content-Type", "Authorization", "X-API-Key"},
				MaxAge:       7200,
			},
			RateLimit: limiter.RateLimitConfig{
				Enabled:         true,
				Algorithm:       limiter.AlgorithmSlidingWindow,
				Rate:            1000,
				Burst:           200,
				ErrorStatusCode: 429,
				ErrorMessage:    "Too many requests",
			},
		}

		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "integration_test.yaml")

		// 2. 保存配置到文件
		err := yamlLoader.SaveConfigToFile(originalCfg, configPath)
		require.NoError(t, err)
		assert.FileExists(t, configPath)

		// 3. 验证配置文件
		err = yamlLoader.ValidateConfig(configPath)
		require.NoError(t, err)

		// 4. 加载配置
		loadedCfg, err := yamlLoader.LoadConfig(configPath)
		require.NoError(t, err)

		// 5. 验证基础配置
		assert.Equal(t, originalCfg.Base.Name, loadedCfg.Base.Name)
		assert.Equal(t, originalCfg.Base.Listen, loadedCfg.Base.Listen)
		assert.Equal(t, originalCfg.Base.ReadTimeout, loadedCfg.Base.ReadTimeout)
		assert.Equal(t, originalCfg.Base.MaxBodySize, loadedCfg.Base.MaxBodySize)
		assert.Equal(t, originalCfg.Base.EnableHTTPS, loadedCfg.Base.EnableHTTPS)

		// 6. 验证模块配置
		assert.Equal(t, originalCfg.Router.ID, loadedCfg.Router.ID)
		assert.Equal(t, originalCfg.Router.Name, loadedCfg.Router.Name)
		assert.Equal(t, originalCfg.Router.Enabled, loadedCfg.Router.Enabled)

		// 7. 验证认证配置
		assert.Equal(t, originalCfg.Auth.Enabled, loadedCfg.Auth.Enabled)
		assert.Equal(t, originalCfg.Auth.Strategy, loadedCfg.Auth.Strategy)

		// 8. 验证CORS配置
		assert.Equal(t, originalCfg.CORS.Enabled, loadedCfg.CORS.Enabled)
		assert.Equal(t, originalCfg.CORS.AllowOrigins, loadedCfg.CORS.AllowOrigins)
		assert.Equal(t, originalCfg.CORS.AllowMethods, loadedCfg.CORS.AllowMethods)
		assert.Equal(t, originalCfg.CORS.MaxAge, loadedCfg.CORS.MaxAge)

		// 9. 验证限流配置
		assert.Equal(t, originalCfg.RateLimit.Enabled, loadedCfg.RateLimit.Enabled)
		assert.Equal(t, originalCfg.RateLimit.Algorithm, loadedCfg.RateLimit.Algorithm)
		assert.Equal(t, originalCfg.RateLimit.Rate, loadedCfg.RateLimit.Rate)
		assert.Equal(t, originalCfg.RateLimit.Burst, loadedCfg.RateLimit.Burst)

		// 11. 导出为YAML字符串
		yamlString, err := yamlLoader.ExportConfigToYAML(loadedCfg)
		require.NoError(t, err)
		assert.NotEmpty(t, yamlString)

		// 12. 从字符串重新解析
		reparsedCfg, err := yamlLoader.ParseYAMLString(yamlString)
		require.NoError(t, err)

		// 13. 验证重新解析的配置
		assert.Equal(t, loadedCfg.Base.Name, reparsedCfg.Base.Name)
		assert.Equal(t, loadedCfg.Auth.Strategy, reparsedCfg.Auth.Strategy)
		assert.Equal(t, loadedCfg.CORS.AllowOrigins, reparsedCfg.CORS.AllowOrigins)

		// 14. 修改配置并重新保存
		reparsedCfg.Base.Name = "Modified Integration Gateway"
		reparsedCfg.Base.Listen = ":7777"

		modifiedPath := filepath.Join(tempDir, "modified_integration.yaml")
		err = yamlLoader.SaveConfigToFile(reparsedCfg, modifiedPath)
		require.NoError(t, err)

		// 15. 重新加载验证修改
		finalCfg, err := yamlLoader.ReloadConfig(modifiedPath)
		require.NoError(t, err)
		assert.Equal(t, "Modified Integration Gateway", finalCfg.Base.Name)
		assert.Equal(t, ":7777", finalCfg.Base.Listen)

		// 其他配置应该保持不变
		assert.Equal(t, originalCfg.Auth.Strategy, finalCfg.Auth.Strategy)
		assert.Equal(t, originalCfg.CORS.AllowOrigins, finalCfg.CORS.AllowOrigins)
		assert.Equal(t, originalCfg.RateLimit.Algorithm, finalCfg.RateLimit.Algorithm)
	})
}
