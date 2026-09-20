package naming

import (
	"context"
	"regexp"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/model"
)

// nodeID 是注册表主键（库 / Cache / Redis 都按 tenantId + nodeId 定位），不是服务名。
//
// 发号：RegisterNode 时空着由本机生成 32 位串；SDK 记下后心跳、注销、重连都带回来。
// 业务有稳定实例号（机器码、Pod UID）可以自己填，服务端原样用。
//
// 自带 ID 的约束：
//   - 字符集 [A-Za-z0-9._:/-]，长度 1～64（validNodeID）
//   - 不得占一个仍存活、且身份不同的实例（rejectTakenNodeID → ErrNodeExists）
//
// 身份 = 命名空间 + 组 + 服务名 + ip + port。同身份覆盖视为重连；心跳已过宽限视为可回收。
// 只比 LastBeatTime 刷新不算身份变化（liveIdentityChanged），避免对账把心跳当 UPDATED 刷订阅。

var nodeIDPattern = regexp.MustCompile(`^[A-Za-z0-9._:/-]{1,64}$`)

func validNodeID(id string) bool {
	return id != "" && nodeIDPattern.MatchString(id)
}

func sameNodeIdentity(a, b *model.Node) bool {
	if a == nil || b == nil {
		return false
	}
	return a.IP == b.IP && a.Port == b.Port && a.ServiceName == b.ServiceName &&
		a.NamespaceID == b.NamespaceID && a.GroupName == b.GroupName
}

func liveIdentityChanged(prev, next *model.Node) bool {
	if prev == nil || next == nil {
		return next != nil
	}
	return prev.IP != next.IP || prev.Port != next.Port ||
		prev.Status != next.Status || prev.HealthyStatus != next.HealthyStatus ||
		prev.ServiceName != next.ServiceName || prev.GroupName != next.GroupName
}

// rejectTakenNodeID 拒绝「自带 nodeId 且撞上另一个还活着的不同实例」。
// 先看本机 Cache，再看活视图。同身份或心跳过期则放行。
func (a *App) rejectTakenNodeID(ctx context.Context, inst *model.Node) error {
	if inst == nil || inst.NodeID == "" {
		return nil
	}
	existing, ok := a.cache.GetNode(inst.NodeID)
	if !ok {
		existing = a.liveGet(ctx, inst)
		ok = existing != nil
	}
	if !ok {
		return nil
	}
	if sameNodeIdentity(existing, inst) {
		return nil
	}
	if !persistentBeatFresh(existing.LastBeatTime, a.persistentHealthGrace()) {
		return nil
	}
	return contract.ErrNodeExists
}
