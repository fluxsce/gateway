package loader_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gateway/internal/gateway/config"
	"gateway/internal/gateway/handler/router"
	"gateway/internal/gateway/loader"
)

func TestNewGatewayConfigFactory(t *testing.T) {
	tests := []struct {
		name   string
		source loader.ConfigSource
	}{
		{"YAML源", loader.ConfigSourceYAML},
		{"JSON源", loader.ConfigSourceJSON},
		{"数据库源", loader.ConfigSourceDB},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := loader.NewGatewayConfigFactory(tt.source)
			assert.NotNil(t, factory)
		})
	}
}

func TestGatewayConfigFactory_LoadConfig(t *testing.T) {
	t.Run("加载YAML配置", func(t *testing.T) {
		factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

		// 使用简单的测试配置文件
		configPath := filepath.Join("testdata", "simple_test_config.yaml")

		cfg, err := factory.LoadConfig(configPath)
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// 验证配置的基本字段
		assert.Equal(t, "Simple Test Gateway", cfg.Base.Name)
		assert.Equal(t, ":8080", cfg.Base.Listen)
		assert.Equal(t, "simple-router", cfg.Router.ID)
		assert.Equal(t, "Simple Router", cfg.Router.Name)
		assert.True(t, cfg.Router.Enabled)
	})

	t.Run("加载不存在的配置文件", func(t *testing.T) {
		factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

		cfg, err := factory.LoadConfig("nonexistent.yaml")
		require.NoError(t, err) // 应该返回默认配置，不报错
		assert.NotNil(t, cfg)

		// 验证是否为默认配置
		defaultCfg := config.DefaultGatewayConfig
		assert.Equal(t, defaultCfg.Base.Name, cfg.Base.Name)
	})

	t.Run("加载空配置路径", func(t *testing.T) {
		factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

		cfg, err := factory.LoadConfig("")
		require.NoError(t, err) // 应该返回默认配置
		assert.NotNil(t, cfg)
	})

	t.Run("不支持的配置源", func(t *testing.T) {
		factory := loader.NewGatewayConfigFactory("unsupported")

		cfg, err := factory.LoadConfig("test.yaml")
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "不支持的配置源")
	})
}

func TestGatewayConfigFactory_LoadConfigFromBytes(t *testing.T) {
	factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

	t.Run("从字节数组加载有效配置", func(t *testing.T) {
		yamlData := `
base:
  name: "Test Gateway"
  listen: ":9090"
  read_timeout: 30s
  write_timeout: 30s
router:
  id: "test-router"
  name: "Test Router"
`

		cfg, err := factory.LoadConfigFromBytes([]byte(yamlData), "yaml")
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "Test Gateway", cfg.Base.Name)
		assert.Equal(t, ":9090", cfg.Base.Listen)
		assert.Equal(t, "test-router", cfg.Router.ID)
	})

	t.Run("从字节数组加载无效配置", func(t *testing.T) {
		invalidYaml := `
base:
  name: "Test Gateway
  listen: ":9090"
invalid_yaml_format
`

		cfg, err := factory.LoadConfigFromBytes([]byte(invalidYaml), "yaml")
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestGatewayConfigFactory_ExportConfigToYAML(t *testing.T) {
	factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

	t.Run("导出有效配置到YAML", func(t *testing.T) {
		cfg := &config.GatewayConfig{
			Base: config.BaseConfig{
				Name:   "Test Gateway",
				Listen: ":8080",
			},
			Router: router.RouterConfig{
				ID:   "test-router",
				Name: "Test Router",
			},
		}

		data, err := factory.ExportConfigToYAML(cfg)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// 验证导出的YAML可以重新解析
		parsedCfg, err := factory.LoadConfigFromBytes(data, "yaml")
		require.NoError(t, err)
		assert.Equal(t, cfg.Base.Name, parsedCfg.Base.Name)
		assert.Equal(t, cfg.Base.Listen, parsedCfg.Base.Listen)
	})

	t.Run("导出空配置", func(t *testing.T) {
		data, err := factory.ExportConfigToYAML(nil)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "配置不能为空")
	})
}

func TestGatewayConfigFactory_ExportConfigToYAMLFile(t *testing.T) {
	factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

	t.Run("导出配置到文件", func(t *testing.T) {
		cfg := &config.GatewayConfig{
			Base: config.BaseConfig{
				Name:   "Test Gateway",
				Listen: ":8080",
			},
		}

		// 创建临时文件路径
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test_config.yaml")

		err := factory.ExportConfigToYAMLFile(cfg, filePath)
		require.NoError(t, err)

		// 验证文件是否创建
		assert.FileExists(t, filePath)

		// 验证文件内容
		loadedCfg, err := factory.LoadConfig(filePath)
		require.NoError(t, err)
		assert.Equal(t, cfg.Base.Name, loadedCfg.Base.Name)
		assert.Equal(t, cfg.Base.Listen, loadedCfg.Base.Listen)
	})

	t.Run("导出到无效路径", func(t *testing.T) {
		cfg := &config.GatewayConfig{}

		err := factory.ExportConfigToYAMLFile(cfg, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "文件路径不能为空")
	})

	t.Run("导出空配置到文件", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test_config.yaml")

		err := factory.ExportConfigToYAMLFile(nil, filePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "配置不能为空")
	})
}

func TestGatewayConfigFactory_ExportDefaultConfigToYAML(t *testing.T) {
	factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

	t.Run("导出默认配置", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "default_config.yaml")

		err := factory.ExportDefaultConfigToYAML(filePath)
		require.NoError(t, err)

		// 验证文件是否创建
		assert.FileExists(t, filePath)

		// 验证文件内容
		loadedCfg, err := factory.LoadConfig(filePath)
		require.NoError(t, err)

		defaultCfg := config.DefaultGatewayConfig
		assert.Equal(t, defaultCfg.Base.Name, loadedCfg.Base.Name)
		assert.Equal(t, defaultCfg.Base.Listen, loadedCfg.Base.Listen)
	})
}

func TestGatewayConfigFactory_SetAndGetConfigValue(t *testing.T) {
	factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

	t.Run("设置和获取配置值", func(t *testing.T) {
		key := "test.key"
		value := "test_value"

		factory.SetConfigValue(key, value)
		retrievedValue := factory.GetConfigValue(key)

		assert.Equal(t, value, retrievedValue)
	})

	t.Run("获取不存在的配置值", func(t *testing.T) {
		value := factory.GetConfigValue("nonexistent.key")
		assert.Nil(t, value)
	})
}

func TestGatewayConfigFactory_GetConfigAsMap(t *testing.T) {
	factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

	t.Run("获取配置Map", func(t *testing.T) {
		configMap := factory.GetConfigAsMap()
		assert.NotNil(t, configMap)

		// 验证默认值是否存在
		assert.Contains(t, configMap, "base")
		assert.Contains(t, configMap, "cors")
		assert.Contains(t, configMap, "auth")
	})
}

func TestGatewayConfigFactory_ValidateConfigFile(t *testing.T) {
	factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

	t.Run("验证有效的YAML文件", func(t *testing.T) {
		configPath := filepath.Join("testdata", "simple_test_config.yaml")
		err := factory.ValidateConfigFile(configPath)
		assert.NoError(t, err)
	})

	t.Run("验证不存在的文件", func(t *testing.T) {
		err := factory.ValidateConfigFile("nonexistent.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "配置文件不存在")
	})

	t.Run("验证空路径", func(t *testing.T) {
		err := factory.ValidateConfigFile("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "配置文件路径不能为空")
	})

	t.Run("验证错误的文件扩展名", func(t *testing.T) {
		// 创建临时文件
		tempDir := t.TempDir()
		wrongExtFile := filepath.Join(tempDir, "config.txt")
		err := os.WriteFile(wrongExtFile, []byte("test"), 0644)
		require.NoError(t, err)

		err = factory.ValidateConfigFile(wrongExtFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "YAML配置源需要.yaml或.yml文件")
	})
}

func TestGatewayConfigFactory_WatchConfig(t *testing.T) {
	factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

	t.Run("监听配置文件变化", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "watch_config.yaml")

		initialConfig := `
base:
  name: "Initial Gateway"
  listen: ":8080"
`
		err := os.WriteFile(configPath, []byte(initialConfig), 0644)
		require.NoError(t, err)

		// 首先加载配置文件
		cfg, err := factory.LoadConfig(configPath)
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// 设置监听
		callbackCalled := false
		err = factory.WatchConfig(func(cfg *config.GatewayConfig) {
			callbackCalled = true
		})
		assert.NoError(t, err)

		// 注意：实际的文件变化监听测试比较复杂
		// 这里只测试WatchConfig方法本身不报错
		assert.False(t, callbackCalled) // 还没有变化，所以回调未被调用
	})
}

func TestGatewayConfigFactory_MergeDefaultConfig(t *testing.T) {
	factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

	t.Run("合并默认配置", func(t *testing.T) {
		// 创建一个部分配置
		cfg := &config.GatewayConfig{
			Base: config.BaseConfig{
				Name: "Custom Gateway",
				// 其他字段留空，应该使用默认值
			},
		}

		// 通过导出再加载的方式测试配置合并
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.yaml")

		err := factory.ExportConfigToYAMLFile(cfg, configPath)
		require.NoError(t, err)

		loadedCfg, err := factory.LoadConfig(configPath)
		require.NoError(t, err)

		// 验证自定义值保持不变
		assert.Equal(t, "Custom Gateway", loadedCfg.Base.Name)

		// 验证默认值被填充
		defaultCfg := config.DefaultGatewayConfig
		assert.Equal(t, defaultCfg.Base.Listen, loadedCfg.Base.Listen)
		assert.Equal(t, defaultCfg.Base.ReadTimeout, loadedCfg.Base.ReadTimeout)
		assert.Equal(t, defaultCfg.Base.WriteTimeout, loadedCfg.Base.WriteTimeout)
	})
}

func TestGetSupportedConfigSources(t *testing.T) {
	sources := loader.GetSupportedConfigSources()

	assert.Len(t, sources, 3)
	assert.Contains(t, sources, loader.ConfigSourceYAML)
	assert.Contains(t, sources, loader.ConfigSourceJSON)
	assert.Contains(t, sources, loader.ConfigSourceDB)
}

func TestGetConfigSourceDescription(t *testing.T) {
	tests := []struct {
		source      loader.ConfigSource
		expectsDesc bool
	}{
		{loader.ConfigSourceYAML, true},
		{loader.ConfigSourceJSON, true},
		{loader.ConfigSourceDB, true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			desc := loader.GetConfigSourceDescription(tt.source)
			if tt.expectsDesc {
				assert.NotEqual(t, "未知配置源", desc)
				assert.NotEmpty(t, desc)
			} else {
				assert.Equal(t, "未知配置源", desc)
			}
		})
	}
}

func TestGatewayConfigFactory_Integration(t *testing.T) {
	t.Run("完整的配置生命周期测试", func(t *testing.T) {
		factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

		// 1. 创建配置
		cfg := &config.GatewayConfig{
			Base: config.BaseConfig{
				Name:         "Integration Test Gateway",
				Listen:       ":9999",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
			},
			Router: router.RouterConfig{
				ID:      "integration-router",
				Name:    "Integration Router",
				Enabled: true,
			},
		}

		// 2. 导出到文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "integration_config.yaml")

		err := factory.ExportConfigToYAMLFile(cfg, configPath)
		require.NoError(t, err)

		// 3. 从文件加载
		loadedCfg, err := factory.LoadConfig(configPath)
		require.NoError(t, err)

		// 4. 验证配置
		assert.Equal(t, cfg.Base.Name, loadedCfg.Base.Name)
		assert.Equal(t, cfg.Base.Listen, loadedCfg.Base.Listen)
		assert.Equal(t, cfg.Router.ID, loadedCfg.Router.ID)
		assert.Equal(t, cfg.Router.Name, loadedCfg.Router.Name)

		// 5. 修改配置
		loadedCfg.Base.Name = "Modified Gateway"

		// 6. 重新导出
		modifiedPath := filepath.Join(tempDir, "modified_config.yaml")
		err = factory.ExportConfigToYAMLFile(loadedCfg, modifiedPath)
		require.NoError(t, err)

		// 7. 验证修改
		finalCfg, err := factory.LoadConfig(modifiedPath)
		require.NoError(t, err)
		assert.Equal(t, "Modified Gateway", finalCfg.Base.Name)
	})
}

func TestGatewayConfigFactory_LoadRealConfigAndExport(t *testing.T) {
	t.Run("加载真实配置文件并导出", func(t *testing.T) {
		factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)

		// 加载真实的配置文件
		realConfigPath := filepath.Join("testdata", "real_gateway_config.yaml")

		cfg, err := factory.LoadConfig(realConfigPath)
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// 验证配置的关键字段
		assert.Equal(t, "gateway-001", cfg.InstanceID)
		assert.Equal(t, "Gateway API Gateway", cfg.Base.Name)
		assert.Equal(t, ":38080", cfg.Base.Listen)
		assert.Equal(t, "default-router", cfg.Router.ID)
		assert.Equal(t, "Default Router", cfg.Router.Name)
		assert.True(t, cfg.Router.Enabled)

		// 创建export目录
		exportDir := filepath.Join(".", "export")
		err = os.MkdirAll(exportDir, 0755)
		require.NoError(t, err)

		// 导出原始配置
		originalExportPath := filepath.Join(exportDir, "original_real_gateway_config.yaml")
		err = factory.ExportConfigToYAMLFile(cfg, originalExportPath)
		require.NoError(t, err)

		// 验证导出文件存在
		assert.FileExists(t, originalExportPath)

		// 重新加载导出的配置进行验证
		reloadedCfg, err := factory.LoadConfig(originalExportPath)
		require.NoError(t, err)

		// 验证重新加载的配置与原配置一致
		assert.Equal(t, cfg.InstanceID, reloadedCfg.InstanceID)
		assert.Equal(t, cfg.Base.Name, reloadedCfg.Base.Name)
		assert.Equal(t, cfg.Base.Listen, reloadedCfg.Base.Listen)
		assert.Equal(t, cfg.Router.ID, reloadedCfg.Router.ID)
		assert.Equal(t, cfg.Router.Name, reloadedCfg.Router.Name)

		// 导出默认配置用于对比
		defaultExportPath := filepath.Join(exportDir, "default_gateway_config.yaml")
		err = factory.ExportDefaultConfigToYAML(defaultExportPath)
		require.NoError(t, err)
		assert.FileExists(t, defaultExportPath)
	})
}
