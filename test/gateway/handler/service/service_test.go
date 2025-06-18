package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gohub/internal/gateway/handler/service"
)

func TestServiceConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *service.ServiceConfig
		description string
	}{
		{
			name: "BasicServiceConfig",
			config: &service.ServiceConfig{
				ID:       "test-service",
				Name:     "测试服务",
				Strategy: service.RoundRobin,
				Nodes: []*service.NodeConfig{
					{
						ID:      "node1",
						URL:     "http://localhost:8001",
						Weight:  1,
						Health:  true,
						Enabled: true,
					},
				},
			},
			description: "基本服务配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置字段
			assert.NotEmpty(t, tt.config.ID, "ID不应该为空")
			assert.NotEmpty(t, tt.config.Name, "名称不应该为空")
			assert.NotEmpty(t, string(tt.config.Strategy), "策略不应该为空")
			assert.NotEmpty(t, tt.config.Nodes, "节点列表不应该为空")
		})
	}
}

func TestLoadBalancerAlgorithms(t *testing.T) {
	algorithms := []service.Strategy{
		service.RoundRobin,
		service.WeightedRoundRobin,
		service.LeastConn,
		service.IPHash,
		service.Random,
	}

	for _, algorithm := range algorithms {
		t.Run(string(algorithm), func(t *testing.T) {
			// 验证算法常量不为空
			assert.NotEmpty(t, string(algorithm), "算法常量不应该为空")
		})
	}
}

func TestServiceManager(t *testing.T) {
	manager := service.NewServiceManager()

	svc := &service.ServiceConfig{
		ID:       "test-service",
		Name:     "测试服务",
		Strategy: service.RoundRobin,
		Nodes: []*service.NodeConfig{
			{
				ID:      "node1",
				URL:     "http://localhost:8001",
				Weight:  1,
				Health:  true,
				Enabled: true,
			},
		},
	}

	// 添加服务
	err := manager.AddService(svc)
	require.NoError(t, err, "添加服务失败")

	// 获取服务
	config, exists := manager.GetService("test-service")
	require.True(t, exists, "服务应该存在")
	require.NotNil(t, config, "服务配置不应该为nil")
	assert.Equal(t, "test-service", config.ID, "服务ID不匹配")

	// 列出服务
	services := manager.ListServices()
	require.Len(t, services, 1, "应该有1个服务")

	// 移除服务
	err = manager.RemoveService("test-service")
	require.NoError(t, err, "移除服务失败")

	// 验证服务已移除
	_, exists = manager.GetService("test-service")
	assert.False(t, exists, "服务应该已经被移除")
}
