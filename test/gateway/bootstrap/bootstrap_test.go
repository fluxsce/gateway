package bootstrap

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gateway/internal/gateway/bootstrap"
	"gateway/internal/gateway/loader"
)

// TestRealConfigBootstrap 使用真实配置文件测试网关的加载和启动
func TestRealConfigBootstrap(t *testing.T) {
	// 不使用 t.Parallel() 避免并行测试

	// 获取测试配置文件路径
	configPath := filepath.Join("testdata", "real_gateway_config.yaml")

	t.Run("LoadAndStartWithRealConfig", func(t *testing.T) {
		// 设置更长的测试超时
		if testing.Short() {
			t.Skip("跳过长时间运行的测试")
		}

		// 使用配置工厂加载真实配置文件
		factory := loader.NewGatewayConfigFactory(loader.ConfigSourceYAML)
		cfg, err := factory.LoadConfig(configPath)
		require.NoError(t, err, "加载真实配置文件失败")
		require.NotNil(t, cfg, "配置不应该为nil")

		// 创建网关
		gatewayFactory := bootstrap.NewGatewayFactory()
		gateway, err := gatewayFactory.CreateGateway(cfg, configPath)
		require.NoError(t, err, "创建网关失败")
		require.NotNil(t, gateway, "网关不应该为nil")

		// 验证初始状态
		assert.False(t, gateway.IsRunning(), "网关不应该处于运行状态")

		// 创建信号通道用于手动控制测试结束
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// 启动网关 - 直接调用，不再需要协程
		fmt.Println("网关启动中，按 Ctrl+C 停止测试...")
		err = gateway.Start()
		if err != nil {
			fmt.Printf("网关启动失败: %v", err)
		}

		// 验证网关已启动
		assert.True(t, gateway.IsRunning(), "网关应该已经启动")

		fmt.Printf("网关已成功启动，监听端口: %s\n", cfg.Base.Listen)
		fmt.Printf("测试路由: /api/v1/users/1, /api/v1/orders/1, /api/v1/products/1")
		fmt.Printf("按 Ctrl+C 停止测试...")

		// 等待手动终止信号
		<-sigCh
		fmt.Printf("收到终止信号，正在停止网关...")

		// 停止网关
		err = gateway.Stop()
		assert.NoError(t, err, "停止网关失败")
		assert.False(t, gateway.IsRunning(), "网关应该已停止")

		fmt.Printf("网关已成功停止")
	})
}
