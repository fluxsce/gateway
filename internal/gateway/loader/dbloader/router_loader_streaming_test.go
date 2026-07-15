package dbloader

import (
	"testing"
	"time"
)

func TestBuildRouteConfigMapsStreamingFields(t *testing.T) {
	rewritePath := "/stream/events"
	config := buildRouteConfig(RouteConfigRecord{
		RouteConfigId:   "route-1",
		RouteName:       "stream-route",
		RoutePath:       "/events",
		StripPathPrefix: "Y",
		RewritePath:     &rewritePath,
		EnableWebsocket: "Y",
		TimeoutMs:       0,
		RetryCount:      3,
		RetryIntervalMs: 250,
		RoutePriority:   10,
		ActiveFlag:      "Y",
	})

	if !config.StripPathPrefix || config.RewritePath != rewritePath || !config.EnableWebSocket {
		t.Fatalf("流式路由字段映射不正确: %+v", config)
	}
	if config.Timeout != 0 {
		t.Fatalf("timeoutMs=0应保留为不覆盖代理总超时: %v", config.Timeout)
	}
	if config.RetryCount != 3 || config.RetryInterval != 250*time.Millisecond {
		t.Fatalf("重试配置映射不正确: %+v", config)
	}
	if config.OverrideProxyTimeout {
		t.Fatal("未配置元数据时不应默认覆盖代理超时/重试")
	}
}

func TestMetadataEnabledFlagParsesOverrideProxyTimeout(t *testing.T) {
	if !metadataEnabledFlag(map[string]interface{}{"overrideProxyTimeout": "Y"}, "overrideProxyTimeout") {
		t.Fatal("Y 应解析为开启覆盖")
	}
	if !metadataEnabledFlag(map[string]interface{}{"override_proxy_timeout": "Y"}, "overrideProxyTimeout", "override_proxy_timeout") {
		t.Fatal("override_proxy_timeout=Y 应解析为开启覆盖")
	}
	for _, value := range []interface{}{"N", "y", "true", true, 1, 1.0, ""} {
		if metadataEnabledFlag(map[string]interface{}{"overrideProxyTimeout": value}, "overrideProxyTimeout") {
			t.Fatalf("%v 不应开启覆盖，仅允许 Y", value)
		}
	}
}
