package dbloader

import "testing"

func TestMapCircuitBreakerRecord(t *testing.T) {
	msg := "unavailable"
	cacheName := `{"cache_name":"cb-redis"}`
	cfg := mapCircuitBreakerRecord(CircuitBreakerConfigRecord{
		ErrorRatePercent:    30,
		MinimumRequests:     5,
		HalfOpenMaxRequests: 2,
		SlowCallThreshold:   800,
		SlowCallRatePercent: 40,
		OpenTimeoutSeconds:  15,
		WindowSizeSeconds:   20,
		ErrorStatusCode:     502,
		ErrorMessage:        &msg,
		StorageType:         "redis",
		StorageConfig:       &cacheName,
	})
	if cfg == nil || !cfg.Enabled {
		t.Fatal("映射结果应启用")
	}
	if cfg.ErrorRatePercent != 30 || cfg.MinimumRequests != 5 {
		t.Fatalf("阈值映射错误: %+v", cfg)
	}
	if cfg.StorageConfig["cache_name"] != "cb-redis" {
		t.Fatalf("storageConfig 解析错误: %+v", cfg.StorageConfig)
	}
	if cfg.ErrorMessage != msg {
		t.Fatalf("错误信息映射错误: %s", cfg.ErrorMessage)
	}
}
