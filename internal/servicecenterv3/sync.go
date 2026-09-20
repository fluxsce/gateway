package servicecenterv3

import (
	"gateway/internal/servicecenterv3/config"
	"gateway/internal/servicecenterv3/model"
	"gateway/internal/servicecenterv3/naming"
)

// SetLocalGatewayID 记录当前网关进程，供 Owner 标记与 Evictor 跳过副本。
func SetLocalGatewayID(gatewayID string) {
	naming.LocalGatewayID = gatewayID
}

// ApplyRemoteNaming 应用其它副本的命名变更：只改 Cache 并推给本机订阅者，不写库、不回复制。
func ApplyRemoteNaming(ev model.NamingEvent) {
	n := namingApp(ev.CenterInstanceName, ev.Environment)
	if n == nil {
		return
	}
	n.ApplyRemote(ev)
}

// ApplyRemoteNamespace 应用其它副本的命名空间变更：只改 Cache，不写库、不回复制。
func ApplyRemoteNamespace(ev model.NamespaceEvent) {
	pool := GetPool()
	if pool == nil {
		return
	}
	inst, ok := pool.FindInstance(ev.CenterInstanceName, ev.Environment)
	if !ok {
		return
	}
	if ev.Type == model.EventNamespaceDeleted {
		inst.DropNamespace(ev.NamespaceID)
		return
	}
	if ev.Namespace != nil {
		inst.SyncNamespace(ev.Namespace)
	}
}

// ApplyRemoteConfig 应用其它副本的已发布配置变更。
func ApplyRemoteConfig(ev model.ConfigEvent) {
	c := configApp(ev.CenterInstanceName, ev.Environment)
	if c == nil {
		return
	}
	c.ApplyRemote(ev)
}

func namingApp(centerName, environment string) *naming.App {
	pool := GetPool()
	if pool == nil {
		return nil
	}
	inst, ok := pool.FindInstance(centerName, environment)
	if !ok || !inst.IsRunning() {
		return nil
	}
	n, _ := inst.Naming().(*naming.App)
	return n
}

func configApp(centerName, environment string) *config.App {
	pool := GetPool()
	if pool == nil {
		return nil
	}
	inst, ok := pool.FindInstance(centerName, environment)
	if !ok || !inst.IsRunning() {
		return nil
	}
	c, _ := inst.ConfigService().(*config.App)
	return c
}
