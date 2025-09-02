package proxyutils

import (
	"testing"
)

func TestParseServiceMetadata(t *testing.T) {
	tests := []struct {
		name        string
		metadata    map[string]string
		expectError bool
		expected    *RegistryServiceMetadata
	}{
		{
			name: "完整的注册中心元数据",
			metadata: map[string]string{
				"tenantId":       "default",
				"serviceGroupId": "group001",
				"serviceName":    "user-service",
				"discoveryType":  "REGISTRY",
				"groupName":      "用户服务组",
			},
			expectError: false,
			expected: &RegistryServiceMetadata{
				TenantID:       "default",
				ServiceGroupID: "group001",
				ServiceName:    "user-service",
				DiscoveryType:  "REGISTRY",
				GroupName:      "用户服务组",
			},
		},
		{
			name: "使用下划线字段名",
			metadata: map[string]string{
				"tenant_id":        "test",
				"service_group_id": "test-group",
				"service_name":     "test-service",
				"discovery_type":   "REGISTRY",
			},
			expectError: false,
			expected: &RegistryServiceMetadata{
				TenantID:       "test",
				ServiceGroupID: "test-group",
				ServiceName:    "test-service",
				DiscoveryType:  "REGISTRY",
			},
		},
		{
			name: "缺少服务名称",
			metadata: map[string]string{
				"tenantId":       "default",
				"serviceGroupId": "group001",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name:        "空元数据",
			metadata:    nil,
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseServiceMetadata(tt.metadata)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望错误但没有发生错误")
				}
				return
			}

			if err != nil {
				t.Errorf("意外错误: %v", err)
				return
			}

			if result == nil {
				t.Errorf("结果为空")
				return
			}

			if result.TenantID != tt.expected.TenantID {
				t.Errorf("TenantID 不匹配: got %s, want %s", result.TenantID, tt.expected.TenantID)
			}

			if result.ServiceGroupID != tt.expected.ServiceGroupID {
				t.Errorf("ServiceGroupID 不匹配: got %s, want %s", result.ServiceGroupID, tt.expected.ServiceGroupID)
			}

			if result.ServiceName != tt.expected.ServiceName {
				t.Errorf("ServiceName 不匹配: got %s, want %s", result.ServiceName, tt.expected.ServiceName)
			}
		})
	}
}

func TestIsRegistryService(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]string
		expected bool
	}{
		{
			name: "明确指定为注册中心服务",
			metadata: map[string]string{
				"tenantId":      "default",
				"serviceName":   "user-service",
				"discoveryType": "REGISTRY",
			},
			expected: true,
		},
		{
			name: "有注册中心字段但未指定类型",
			metadata: map[string]string{
				"tenantId":    "default",
				"serviceName": "user-service",
			},
			expected: true,
		},
		{
			name: "明确指定为静态服务",
			metadata: map[string]string{
				"discoveryType": "STATIC",
			},
			expected: false,
		},
		{
			name:     "空元数据",
			metadata: nil,
			expected: false,
		},
		{
			name: "无相关字段",
			metadata: map[string]string{
				"other": "value",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRegistryService(tt.metadata)
			if result != tt.expected {
				t.Errorf("IsRegistryService() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateRegistryMetadata(t *testing.T) {
	tests := []struct {
		name        string
		metadata    *RegistryServiceMetadata
		expectError bool
	}{
		{
			name: "有效的元数据",
			metadata: &RegistryServiceMetadata{
				TenantID:       "default",
				ServiceGroupID: "group001",
				ServiceName:    "user-service",
			},
			expectError: false,
		},
		{
			name:        "空元数据",
			metadata:    nil,
			expectError: true,
		},
		{
			name: "缺少服务名称",
			metadata: &RegistryServiceMetadata{
				TenantID:       "default",
				ServiceGroupID: "group001",
			},
			expectError: true,
		},
		{
			name: "缺少租户ID",
			metadata: &RegistryServiceMetadata{
				ServiceGroupID: "group001",
				ServiceName:    "user-service",
			},
			expectError: true,
		},
		{
			name: "缺少服务组ID",
			metadata: &RegistryServiceMetadata{
				TenantID:    "default",
				ServiceName: "user-service",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryMetadata(tt.metadata)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望错误但没有发生错误")
				}
			} else {
				if err != nil {
					t.Errorf("意外错误: %v", err)
				}
			}
		})
	}
}

// BenchmarkParseServiceMetadata 基准测试元数据解析性能
func BenchmarkParseServiceMetadata(b *testing.B) {
	metadata := map[string]string{
		"tenantId":       "default",
		"serviceGroupId": "group001",
		"serviceName":    "user-service",
		"discoveryType":  "REGISTRY",
		"groupName":      "用户服务组",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseServiceMetadata(metadata)
	}
}

// BenchmarkIsRegistryService 基准测试服务类型判断性能
func BenchmarkIsRegistryService(b *testing.B) {
	metadata := map[string]string{
		"tenantId":      "default",
		"serviceName":   "user-service",
		"discoveryType": "REGISTRY",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsRegistryService(metadata)
	}
}
