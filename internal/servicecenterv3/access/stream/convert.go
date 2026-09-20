package stream

import (
	"gateway/internal/servicecenterv3/model"
	pb "gateway/internal/servicecenterv3/proto"
)

// serviceFromPB 把流协议 Service 转为领域对象；ns 优先于载荷内 namespace。
func serviceFromPB(ns string, in *pb.Service) *model.Service {
	if in == nil {
		return nil
	}
	if ns == "" {
		ns = in.GetNamespaceId()
	}
	group := in.GetGroupName()
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	svc := &model.Service{
		NamespaceID:      ns,
		GroupName:        group,
		ServiceName:      in.GetServiceName(),
		Version:          in.GetServiceVersion(),
		Description:      in.GetServiceDescription(),
		Metadata:         in.GetMetadata(),
		Tags:             in.GetTags(),
		ProtectThreshold: in.GetProtectThreshold(),
	}
	if node := in.GetNode(); node != nil {
		svc.Nodes = []*model.Node{nodeFromPB(ns, node)}
	}
	return svc
}

// nodeFromPB 把流协议 Node 转为 Node。ephemeral 未设置视为临时。
func nodeFromPB(ns string, in *pb.Node) *model.Node {
	if in == nil {
		return nil
	}
	if ns == "" {
		ns = in.GetNamespaceId()
	}
	group := in.GetGroupName()
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	ephemeral := true
	if in.Ephemeral != nil {
		ephemeral = in.GetEphemeral()
	}
	return &model.Node{
		NodeID:        in.GetNodeId(),
		NamespaceID:   ns,
		GroupName:     group,
		ServiceName:   in.GetServiceName(),
		IP:            in.GetIpAddress(),
		Port:          int(in.GetPortNumber()),
		Weight:        in.GetWeight(),
		Ephemeral:     ephemeral,
		Status:        nodeStatusFromPB(in.GetInstanceStatus()),
		HealthyStatus: healthyStatusFromPB(in.GetHealthyStatus()),
		Metadata:      in.GetMetadata(),
	}
}

// nodeToPB 把领域节点转为流协议 Node。
func nodeToPB(in *model.Node) *pb.Node {
	if in == nil {
		return nil
	}
	ephemeral := in.Ephemeral
	return &pb.Node{
		NodeId:         in.NodeID,
		NamespaceId:    in.NamespaceID,
		GroupName:      in.GroupName,
		ServiceName:    in.ServiceName,
		IpAddress:      in.IP,
		PortNumber:     int32(in.Port),
		Weight:         in.Weight,
		Ephemeral:      &ephemeral,
		InstanceStatus: nodeStatusToPB(in.Status),
		HealthyStatus:  healthyStatusToPB(in.HealthyStatus),
		Metadata:       in.Metadata,
	}
}

// serviceToPB 把领域服务转为流协议 Service。读路径不填 node。
func serviceToPB(in *model.Service) *pb.Service {
	if in == nil {
		return nil
	}
	return &pb.Service{
		NamespaceId:        in.NamespaceID,
		GroupName:          in.GroupName,
		ServiceName:        in.ServiceName,
		ServiceType:        "INTERNAL",
		ServiceVersion:     in.Version,
		ServiceDescription: in.Description,
		ProtectThreshold:   in.ProtectThreshold,
		Metadata:           in.Metadata,
		Tags:               in.Tags,
	}
}

// nodesToPB 批量转换节点列表。
func nodesToPB(list []*model.Node) []*pb.Node {
	out := make([]*pb.Node, 0, len(list))
	for _, node := range list {
		out = append(out, nodeToPB(node))
	}
	return out
}

// namingEventToPB 把命名变更转为客户端可消费的 ServiceChangeEvent。
func namingEventToPB(ev model.NamingEvent) *pb.ServiceChangeEvent {
	return &pb.ServiceChangeEvent{
		EventType:   namingEventFromModel(ev.Type),
		Timestamp:   ev.Timestamp.UnixMilli(),
		NamespaceId: ev.NamespaceID,
		GroupName:   ev.GroupName,
		ServiceName: ev.ServiceName,
		Service:     serviceToPB(ev.Service),
		Nodes:       nodesToPB(nodesOf(ev.Service)),
		ChangedNode: nodeToPB(ev.ChangedNode),
	}
}

func nodesOf(svc *model.Service) []*model.Node {
	if svc == nil {
		return nil
	}
	return svc.Nodes
}

func releaseToPB(rel *model.ConfigRelease) *pb.ConfigData {
	if rel == nil {
		return nil
	}
	return &pb.ConfigData{
		NamespaceId:   rel.NamespaceID,
		GroupName:     rel.GroupName,
		ConfigDataId:  rel.DataID,
		ContentType:   rel.ContentType,
		ConfigContent: rel.Content,
		ContentMd5:    rel.MD5,
		ConfigDesc:    rel.Description,
		ConfigVersion: rel.Version,
	}
}

func configEventToPB(ev model.ConfigEvent) *pb.ConfigChangeEvent {
	md5 := ""
	if ev.Release != nil {
		md5 = ev.Release.MD5
	}
	return &pb.ConfigChangeEvent{
		EventType:    configEventFromModel(ev.Type),
		Timestamp:    ev.Timestamp.UnixMilli(),
		NamespaceId:  ev.NamespaceID,
		GroupName:    ev.GroupName,
		ConfigDataId: ev.DataID,
		Config:       releaseToPB(ev.Release),
		ContentMd5:   md5,
	}
}

func draftFromPB(ns string, in *pb.ConfigData) *model.ConfigDraft {
	if in == nil {
		return nil
	}
	if ns == "" {
		ns = in.GetNamespaceId()
	}
	group := in.GetGroupName()
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	return &model.ConfigDraft{
		NamespaceID: ns,
		GroupName:   group,
		DataID:      in.GetConfigDataId(),
		Content:     in.GetConfigContent(),
		ContentType: in.GetContentType(),
		Description: in.GetConfigDesc(),
	}
}

func nodeStatusFromPB(in pb.InstanceStatus) string {
	switch in {
	case pb.InstanceStatus_INSTANCE_STATUS_DOWN:
		return model.NodeDown
	case pb.InstanceStatus_INSTANCE_STATUS_STARTING:
		return model.NodeStarting
	case pb.InstanceStatus_INSTANCE_STATUS_OUT_OF_SERVICE:
		return model.NodeOutOfService
	default:
		return model.NodeUP
	}
}

func nodeStatusToPB(in string) pb.InstanceStatus {
	switch in {
	case model.NodeDown:
		return pb.InstanceStatus_INSTANCE_STATUS_DOWN
	case model.NodeStarting:
		return pb.InstanceStatus_INSTANCE_STATUS_STARTING
	case model.NodeOutOfService:
		return pb.InstanceStatus_INSTANCE_STATUS_OUT_OF_SERVICE
	default:
		return pb.InstanceStatus_INSTANCE_STATUS_UP
	}
}

func healthyStatusFromPB(in pb.HealthyStatus) string {
	switch in {
	case pb.HealthyStatus_HEALTHY_STATUS_UNHEALTHY:
		return model.Unhealthy
	case pb.HealthyStatus_HEALTHY_STATUS_UNKNOWN:
		return model.Unknown
	default:
		return model.Healthy
	}
}

func healthyStatusToPB(in string) pb.HealthyStatus {
	switch in {
	case model.Unhealthy:
		return pb.HealthyStatus_HEALTHY_STATUS_UNHEALTHY
	case model.Unknown:
		return pb.HealthyStatus_HEALTHY_STATUS_UNKNOWN
	default:
		return pb.HealthyStatus_HEALTHY_STATUS_HEALTHY
	}
}

func namingEventFromModel(t string) pb.NamingEventType {
	switch t {
	case model.EventServiceAdded:
		return pb.NamingEventType_NAMING_EVENT_SERVICE_ADDED
	case model.EventServiceUpdated:
		return pb.NamingEventType_NAMING_EVENT_SERVICE_UPDATED
	case model.EventServiceDeleted:
		return pb.NamingEventType_NAMING_EVENT_SERVICE_DELETED
	case model.EventNodeRegistered:
		return pb.NamingEventType_NAMING_EVENT_NODE_REGISTERED
	case model.EventNodeUpdated:
		return pb.NamingEventType_NAMING_EVENT_NODE_UPDATED
	case model.EventNodeDeregistered:
		return pb.NamingEventType_NAMING_EVENT_NODE_DEREGISTERED
	case model.EventNodeEvicted:
		return pb.NamingEventType_NAMING_EVENT_NODE_EVICTED
	case model.EventNodeOffline:
		return pb.NamingEventType_NAMING_EVENT_NODE_OFFLINE
	default:
		return pb.NamingEventType_NAMING_EVENT_SERVICE_UPDATED
	}
}

func configEventFromModel(t string) pb.ConfigEventType {
	switch t {
	case model.EventConfigDeleted:
		return pb.ConfigEventType_CONFIG_EVENT_DELETED
	case model.EventConfigRolledBack:
		return pb.ConfigEventType_CONFIG_EVENT_ROLLED_BACK
	default:
		return pb.ConfigEventType_CONFIG_EVENT_PUBLISHED
	}
}

func configChangeTypeToPB(t string) pb.ConfigChangeType {
	switch t {
	case "CREATE", "ADD":
		return pb.ConfigChangeType_CONFIG_CHANGE_TYPE_ADD
	case "DELETE":
		return pb.ConfigChangeType_CONFIG_CHANGE_TYPE_DELETE
	default:
		return pb.ConfigChangeType_CONFIG_CHANGE_TYPE_UPDATE
	}
}
