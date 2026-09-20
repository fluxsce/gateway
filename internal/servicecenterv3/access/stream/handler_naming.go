package stream

import (
	"context"

	"gateway/internal/servicecenterv3/contract"
	pb "gateway/internal/servicecenterv3/proto"
)

// registerService 只注册服务定义；附带 Node 时再注册该节点并绑定连接。
func (h *Handler) registerService(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetRegisterService()
	if in == nil || in.GetServiceName() == "" {
		return contract.ErrInvalidArgument
	}
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	svc := serviceFromPB(cc.NamespaceID, in)
	if err := h.naming.RegisterService(ctx, cc, svc); err != nil {
		return err
	}
	nodeID := ""
	if in.GetNode() != nil {
		inst := nodeFromPB(cc.NamespaceID, in.GetNode())
		inst.ServiceName = svc.ServiceName
		inst.GroupName = svc.GroupName
		inst.ConnectionID = st.id
		if err := h.naming.RegisterNode(ctx, cc, inst); err != nil {
			return err
		}
		nodeID = inst.NodeID
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_REGISTER_SERVICE,
		Message:     &pb.ServerMessage_RegisterService{RegisterService: &pb.RegisterServiceResponse{Success: true, NodeId: nodeID}},
	})
}

// unregisterService 数据面只摘本连接在该服务下的节点，不误伤同服务其它实例。
func (h *Handler) unregisterService(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetUnregisterService()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	if err := h.naming.DeregisterService(ctx, cc, in.GetGroupName(), in.GetServiceName()); err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_UNREGISTER_SERVICE,
		Message:     &pb.ServerMessage_UnregisterService{UnregisterService: &pb.RegistryResponse{Success: true}},
	})
}

// registerNode 把节点挂到 Node.serviceName 对应的服务下并绑定当前连接。
func (h *Handler) registerNode(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetRegisterNode()
	if in == nil || in.GetServiceName() == "" || in.GetIpAddress() == "" || in.GetPortNumber() <= 0 {
		return contract.ErrInvalidArgument
	}
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	inst := nodeFromPB(cc.NamespaceID, in)
	inst.ConnectionID = st.id
	if err := h.naming.RegisterNode(ctx, cc, inst); err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_REGISTER_NODE,
		Message:     &pb.ServerMessage_RegisterNode{RegisterNode: &pb.RegisterNodeResponse{Success: true, NodeId: inst.NodeID}},
	})
}

// unregisterNode 按 nodeId 下线。
func (h *Handler) unregisterNode(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetUnregisterNode()
	cc := h.callContext(ctx, "", st)
	if err := h.naming.DeregisterNode(ctx, cc, in.GetNodeId()); err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_UNREGISTER_NODE,
		Message:     &pb.ServerMessage_UnregisterNode{UnregisterNode: &pb.RegistryResponse{Success: true}},
	})
}

// discover 列出节点，可只返回健康实例。
func (h *Handler) discover(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetDiscoverNodes()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	list, err := h.naming.ListNodes(ctx, cc, in.GetGroupName(), in.GetServiceName(), in.GetHealthyOnly())
	if err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_DISCOVER_NODES,
		Message:     &pb.ServerMessage_DiscoverNodes{DiscoverNodes: &pb.DiscoverNodesResponse{Success: true, Nodes: nodesToPB(list)}},
	})
}

// getService 读取服务定义及实例快照。
func (h *Handler) getService(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetGetService()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	svc, err := h.naming.GetService(ctx, cc, in.GetGroupName(), in.GetServiceName())
	if err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_GET_SERVICE,
		Message: &pb.ServerMessage_GetService{GetService: &pb.GetServiceResponse{
			Success: true,
			Service: serviceToPB(svc),
			Nodes:   nodesToPB(svc.Nodes),
		}},
	})
}

// listServices 列出命名空间下服务。
func (h *Handler) listServices(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetListServices()
	ns := ""
	group := ""
	if in != nil {
		ns, group = in.GetNamespaceId(), in.GetGroupName()
	}
	cc := h.callContext(ctx, ns, st)
	list, err := h.naming.ListServices(ctx, cc, group)
	if err != nil {
		return err
	}
	all := make([]*pb.Service, 0, len(list))
	for _, svc := range list {
		all = append(all, serviceToPB(svc))
	}
	pageSize, token := int32(0), ""
	if in != nil {
		pageSize, token = in.GetPageSize(), in.GetPageToken()
	}
	out, next, total := pageOf(all, pageSize, token)
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_LIST_SERVICES,
		Message:     &pb.ServerMessage_ListServices{ListServices: &pb.ListServicesResponse{Success: true, Services: out, NextPageToken: next, Total: total}},
	})
}

// heartbeat 刷新 LastBeatTime。
func (h *Handler) heartbeat(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetHeartbeat()
	ns := ""
	if in.GetService() != nil {
		ns = in.GetService().GetNamespaceId()
	}
	cc := h.callContext(ctx, ns, st)
	svc := serviceFromPB(cc.NamespaceID, in.GetService())
	if svc != nil {
		for _, inst := range svc.Nodes {
			if inst != nil {
				inst.ConnectionID = st.id
			}
		}
	}
	if err := h.naming.Heartbeat(ctx, cc, in.GetNodeId(), svc); err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_HEARTBEAT,
		Message:     &pb.ServerMessage_Heartbeat{Heartbeat: &pb.RegistryResponse{Success: true}},
	})
}

// subscribeServices 订阅指定服务，绑定连接并回 ACK。
func (h *Handler) subscribeServices(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetSubscribeServices()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	if err := h.naming.SubscribeServices(ctx, cc, in.GetGroupName(), in.GetServiceNames(), st.namingCh); err != nil {
		return err
	}
	h.naming.BindClient(st.namingCh, st.id)
	return h.sendAck(st, stream, msg.GetRequestId())
}

// subscribeNamespace 订阅命名空间并回 ACK。
func (h *Handler) subscribeNamespace(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetSubscribeNamespace()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	if err := h.naming.SubscribeNamespace(ctx, cc, in.GetGroupName(), st.namingCh); err != nil {
		return err
	}
	h.naming.BindClient(st.namingCh, st.id)
	return h.sendAck(st, stream, msg.GetRequestId())
}

// unsubscribe 取消本连接全部命名订阅并回 ACK。
func (h *Handler) unsubscribe(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	cc := h.callContext(ctx, "", st)
	h.naming.Unsubscribe(ctx, cc, st.namingCh)
	return h.sendAck(st, stream, msg.GetRequestId())
}

// unsubscribeServices 取消本连接上指定服务的订阅并回 ACK。
func (h *Handler) unsubscribeServices(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetUnsubscribeServices()
	ns, group := "", ""
	var names []string
	if in != nil {
		ns, group, names = in.GetNamespaceId(), in.GetGroupName(), in.GetServiceNames()
	}
	cc := h.callContext(ctx, ns, st)
	h.naming.UnsubscribeServices(ctx, cc, group, names, st.namingCh)
	return h.sendAck(st, stream, msg.GetRequestId())
}
