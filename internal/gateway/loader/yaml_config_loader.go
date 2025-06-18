package loader

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"gohub/internal/gateway/config"
)

// YAMLConfigLoader YAML配置加载器
type YAMLConfigLoader struct {
	factory *GatewayConfigFactory
}

// NewYAMLConfigLoader 创建YAML配置加载器
func NewYAMLConfigLoader() *YAMLConfigLoader {
	return &YAMLConfigLoader{
		factory: NewGatewayConfigFactory(ConfigSourceYAML),
	}
}

// LoadConfig 从YAML文件加载配置
func (y *YAMLConfigLoader) LoadConfig(configPath string) (*config.GatewayConfig, error) {
	return y.factory.LoadConfig(configPath)
}

// ValidateConfig 验证YAML配置
func (y *YAMLConfigLoader) ValidateConfig(configPath string) error {
	if err := y.factory.ValidateConfigFile(configPath); err != nil {
		return fmt.Errorf("YAML配置文件验证失败: %w", err)
	}

	// 尝试解析配置文件
	_, err := y.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("YAML配置文件解析失败: %w", err)
	}

	return nil
}

// GetSupportedExtensions 获取支持的文件扩展名
func (y *YAMLConfigLoader) GetSupportedExtensions() []string {
	return []string{".yaml", ".yml"}
}

// IsValidExtension 检查文件扩展名是否有效
func (y *YAMLConfigLoader) IsValidExtension(configPath string) bool {
	for _, ext := range y.GetSupportedExtensions() {
		if len(configPath) >= len(ext) && configPath[len(configPath)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// ParseYAMLString 从YAML字符串解析配置
func (y *YAMLConfigLoader) ParseYAMLString(yamlString string) (*config.GatewayConfig, error) {
	cfg := &config.GatewayConfig{}
	if err := yaml.Unmarshal([]byte(yamlString), cfg); err != nil {
		return nil, fmt.Errorf("解析YAML字符串失败: %w", err)
	}

	// 合并默认配置
	y.factory.mergeDefaultConfig(cfg)

	return cfg, nil
}

// ExportConfigToYAML 将配置导出为YAML格式
func (y *YAMLConfigLoader) ExportConfigToYAML(cfg *config.GatewayConfig) (string, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("导出YAML配置失败: %w", err)
	}

	return string(data), nil
}

// SaveConfigToFile 将配置保存到YAML文件
func (y *YAMLConfigLoader) SaveConfigToFile(cfg *config.GatewayConfig, configPath string) error {
	yamlData, err := y.ExportConfigToYAML(cfg)
	if err != nil {
		return fmt.Errorf("导出YAML配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(yamlData), 0644); err != nil {
		return fmt.Errorf("保存YAML配置文件失败: %w", err)
	}

	return nil
}

// ReloadConfig 重新加载YAML配置
func (y *YAMLConfigLoader) ReloadConfig(configPath string) (*config.GatewayConfig, error) {
	return y.LoadConfig(configPath)
}
