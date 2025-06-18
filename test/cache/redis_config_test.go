package cache

import (
	"testing"
	"time"

	"gohub/pkg/cache/redis"
)

func TestRedisConfigDefaults(t *testing.T) {
	config := &redis.RedisConfig{}
	config.SetDefaults()

	if config.Port != 6379 {
		t.Errorf("Expected default port 6379, got %d", config.Port)
	}

	if config.DB != 0 {
		t.Errorf("Expected default DB 0, got %d", config.DB)
	}

	if config.PoolSize != 100 {
		t.Errorf("Expected default PoolSize 100, got %d", config.PoolSize)
	}

	if config.MinIdleConns != 10 {
		t.Errorf("Expected default MinIdleConns 10, got %d", config.MinIdleConns)
	}

	if config.MaxIdleConns != 100 {
		t.Errorf("Expected default MaxIdleConns 100, got %d", config.MaxIdleConns)
	}

	if config.MaxActiveConns != 100 {
		t.Errorf("Expected default MaxActiveConns 100, got %d", config.MaxActiveConns)
	}

	if config.IdleTimeout != 1800000 {
		t.Errorf("Expected default IdleTimeout 1800000, got %d", config.IdleTimeout)
	}

	if config.DialTimeout != 5*time.Second {
		t.Errorf("Expected default DialTimeout 5s, got %v", config.DialTimeout)
	}

	if config.ReadTimeout != 3*time.Second {
		t.Errorf("Expected default ReadTimeout 3s, got %v", config.ReadTimeout)
	}

	if config.WriteTimeout != 3*time.Second {
		t.Errorf("Expected default WriteTimeout 3s, got %v", config.WriteTimeout)
	}

	if config.PoolTimeout != 4*time.Second {
		t.Errorf("Expected default PoolTimeout 4s, got %v", config.PoolTimeout)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected default MaxRetries 3, got %d", config.MaxRetries)
	}

	if config.MinRetryBackoff != 8*time.Millisecond {
		t.Errorf("Expected default MinRetryBackoff 8ms, got %v", config.MinRetryBackoff)
	}

	if config.MaxRetryBackoff != 512*time.Millisecond {
		t.Errorf("Expected default MaxRetryBackoff 512ms, got %v", config.MaxRetryBackoff)
	}
}

func TestRedisConfigValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      redis.RedisConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: redis.RedisConfig{
				Host: "localhost",
				Port: 6379,
				DB:   0,
			},
			expectError: false,
		},
		{
			name: "missing host",
			config: redis.RedisConfig{
				Port: 6379,
				DB:   0,
			},
			expectError: true,
		},
		{
			name: "invalid port - too low",
			config: redis.RedisConfig{
				Host: "localhost",
				Port: 0,
				DB:   0,
			},
			expectError: true,
		},
		{
			name: "invalid port - too high",
			config: redis.RedisConfig{
				Host: "localhost",
				Port: 70000,
				DB:   0,
			},
			expectError: true,
		},
		{
			name: "invalid db - negative",
			config: redis.RedisConfig{
				Host: "localhost",
				Port: 6379,
				DB:   -1,
			},
			expectError: true,
		},
		{
			name: "invalid db - too high",
			config: redis.RedisConfig{
				Host: "localhost",
				Port: 6379,
				DB:   16,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestRedisConfigGetAddress(t *testing.T) {
	config := &redis.RedisConfig{
		Host: "192.168.1.100",
		Port: 6380,
	}

	expected := "192.168.1.100:6380"
	actual := config.GetAddress()

	if actual != expected {
		t.Errorf("Expected address %s, got %s", expected, actual)
	}
}

func TestRedisConfigGetIdleTimeoutDuration(t *testing.T) {
	config := &redis.RedisConfig{
		IdleTimeout: 5000, // 5000毫秒
	}

	expected := 5 * time.Second
	actual := config.GetIdleTimeoutDuration()

	if actual != expected {
		t.Errorf("Expected duration %v, got %v", expected, actual)
	}
}

func TestRedisConfigString(t *testing.T) {
	config := &redis.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "secret",
		DB:       1,
		PoolSize: 50,
	}

	str := config.String()
	expected := "Redis{Host:localhost, Port:6379, Password:***, DB:1, PoolSize:50}"

	if str != expected {
		t.Errorf("Expected string %s, got %s", expected, str)
	}

	// 测试无密码情况
	config.Password = ""
	str = config.String()
	expected = "Redis{Host:localhost, Port:6379, Password:, DB:1, PoolSize:50}"

	if str != expected {
		t.Errorf("Expected string %s, got %s", expected, str)
	}
}

func TestRedisConfigGetType(t *testing.T) {
	config := &redis.RedisConfig{}
	if config.GetType() != "redis" {
		t.Errorf("Expected type 'redis', got %s", config.GetType())
	}
} 