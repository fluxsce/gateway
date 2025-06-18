package loader

import (
	"encoding/json"
	"fmt"
	"os"

	"gohub/internal/gateway/config"
)

// JSONConfigLoader JSON配置加载器
type JSONConfigLoader struct {
	factory *GatewayConfigFactory
}

// NewJSONConfigLoader 创建JSON配置加载器
func NewJSONConfigLoader() *JSONConfigLoader {
	return &JSONConfigLoader{
		factory: NewGatewayConfigFactory(ConfigSourceJSON),
	}
}

// LoadConfig 从JSON文件加载配置
func (j *JSONConfigLoader) LoadConfig(configPath string) (*config.GatewayConfig, error) {
	return j.factory.LoadConfig(configPath)
}

// ValidateConfig 验证JSON配置
func (j *JSONConfigLoader) ValidateConfig(configPath string) error {
	if err := j.factory.ValidateConfigFile(configPath); err != nil {
		return fmt.Errorf("JSON配置文件验证失败: %w", err)
	}

	// 尝试解析配置文件
	_, err := j.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("JSON配置文件解析失败: %w", err)
	}

	return nil
}

// GetSupportedExtensions 获取支持的文件扩展名
func (j *JSONConfigLoader) GetSupportedExtensions() []string {
	return []string{".json"}
}

// IsValidExtension 检查文件扩展名是否有效
func (j *JSONConfigLoader) IsValidExtension(configPath string) bool {
	for _, ext := range j.GetSupportedExtensions() {
		if len(configPath) >= len(ext) && configPath[len(configPath)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// ParseJSONString 从JSON字符串解析配置
func (j *JSONConfigLoader) ParseJSONString(jsonString string) (*config.GatewayConfig, error) {
	cfg := &config.GatewayConfig{}
	if err := json.Unmarshal([]byte(jsonString), cfg); err != nil {
		return nil, fmt.Errorf("解析JSON字符串失败: %w", err)
	}

	// 合并默认配置
	j.factory.mergeDefaultConfig(cfg)

	return cfg, nil
}

// ExportConfigToJSON 将配置导出为JSON格式
func (j *JSONConfigLoader) ExportConfigToJSON(cfg *config.GatewayConfig, indent bool) (string, error) {
	var data []byte
	var err error

	if indent {
		data, err = json.MarshalIndent(cfg, "", "  ")
	} else {
		data, err = json.Marshal(cfg)
	}

	if err != nil {
		return "", fmt.Errorf("导出JSON配置失败: %w", err)
	}

	return string(data), nil
}

// SaveConfigToFile 将配置保存到JSON文件
func (j *JSONConfigLoader) SaveConfigToFile(cfg *config.GatewayConfig, configPath string, indent bool) error {
	jsonData, err := j.ExportConfigToJSON(cfg, indent)
	if err != nil {
		return fmt.Errorf("导出JSON配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(jsonData), 0644); err != nil {
		return fmt.Errorf("保存JSON配置文件失败: %w", err)
	}

	return nil
}

// ValidateJSONString 验证JSON字符串格式
func (j *JSONConfigLoader) ValidateJSONString(jsonString string) error {
	cfg := &config.GatewayConfig{}
	if err := json.Unmarshal([]byte(jsonString), cfg); err != nil {
		return fmt.Errorf("JSON格式验证失败: %w", err)
	}
	return nil
}

// CompareConfigs 比较两个配置是否相同
func (j *JSONConfigLoader) CompareConfigs(cfg1, cfg2 *config.GatewayConfig) (bool, error) {
	json1, err := j.ExportConfigToJSON(cfg1, false)
	if err != nil {
		return false, fmt.Errorf("序列化配置1失败: %w", err)
	}

	json2, err := j.ExportConfigToJSON(cfg2, false)
	if err != nil {
		return false, fmt.Errorf("序列化配置2失败: %w", err)
	}

	return json1 == json2, nil
}

// ReloadConfig 重新加载JSON配置
func (j *JSONConfigLoader) ReloadConfig(configPath string) (*config.GatewayConfig, error) {
	return j.LoadConfig(configPath)
}
