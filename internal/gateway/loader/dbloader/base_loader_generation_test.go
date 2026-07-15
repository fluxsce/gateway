package dbloader

import (
	"testing"
	"time"

	"gateway/internal/gateway/config"
)

func TestBuildBaseConfigMapsGenerationDrainTimeout(t *testing.T) {
	httpPort := 8080
	loader := &BaseConfigLoader{}
	base := loader.BuildBaseConfig(&GatewayInstanceRecord{
		BindAddress:               "127.0.0.1",
		HTTPPort:                  &httpPort,
		GracefulShutdownTimeoutMs: 2500,
	})

	if base.GracefulShutdownTimeout != 2500*time.Millisecond {
		t.Fatalf("GracefulShutdownTimeout = %s, want 2.5s", base.GracefulShutdownTimeout)
	}
}

func TestBuildBaseConfigDefaultsGenerationDrainTimeout(t *testing.T) {
	httpPort := 8080
	loader := &BaseConfigLoader{}
	base := loader.BuildBaseConfig(&GatewayInstanceRecord{
		BindAddress: "127.0.0.1",
		HTTPPort:    &httpPort,
	})

	if base.GracefulShutdownTimeout != config.DefaultGatewayConfig.Base.GracefulShutdownTimeout {
		t.Fatalf("GracefulShutdownTimeout = %s, want %s",
			base.GracefulShutdownTimeout, config.DefaultGatewayConfig.Base.GracefulShutdownTimeout)
	}
}

func TestBuildBaseConfigMapsCapacityLimits(t *testing.T) {
	httpPort := 8080
	loader := &BaseConfigLoader{}
	base := loader.BuildBaseConfig(&GatewayInstanceRecord{
		BindAddress:    "127.0.0.1",
		HTTPPort:       &httpPort,
		MaxConnections: 25000,
		MaxWorkers:     2500,
	})

	if base.MaxConnections != 25000 || base.MaxWorkers != 2500 {
		t.Fatalf("capacity limits = connections:%d workers:%d, want 25000 and 2500",
			base.MaxConnections, base.MaxWorkers)
	}
}
