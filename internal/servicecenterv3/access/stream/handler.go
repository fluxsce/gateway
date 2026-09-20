package stream

import (
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/alert"
	"gateway/internal/servicecenterv3/model"
	"gateway/internal/servicecenterv3/naming"
	pb "gateway/internal/servicecenterv3/proto"
	"gateway/pkg/utils/random"

	"google.golang.org/grpc/peer"
)

// ProtocolVersion 是数据面协议号，后续 SDK v3 握手必须带此值。
const ProtocolVersion = "v3"

// Handler 实现 ServiceCenterStreamServer，按 opcode 一对一路由到 Naming/Config。
type Handler struct {
	pb.UnimplementedServiceCenterStreamServer
	naming  contract.Naming
	config  contract.Config
	session Session
	cfg     *model.CenterInstance
	liveMu  sync.Mutex
	live    map[string]liveStream
}

type liveStream struct {
	st     *connState
	stream pb.ServiceCenterStream_ConnectServer
}

// connState 是单条 Connect 流的会话，不跨连接共享。
type connState struct {
	id        string
	clientID  string
	clientIP  string
	defaultNS string
	ready     bool
	drain     atomic.Bool // NotifyClose / 网关停机，OnClientLost 不得当 SDK 离开来驱逐
	namingCh  chan model.NamingEvent
	configCh  chan model.ConfigEvent
	sendMu    sync.Mutex
}

// NewHandler 构造流处理器。session 用于连接登记。
func NewHandler(cfg *model.CenterInstance, naming contract.Naming, config contract.Config, session Session) *Handler {
	return &Handler{naming: naming, config: config, session: session, cfg: cfg, live: map[string]liveStream{}}
}

// send 按连接串行写出，避免同一条流并发 Send，不阻塞其它连接。
func (h *Handler) send(st *connState, stream pb.ServiceCenterStream_ConnectServer, msg *pb.ServerMessage) error {
	st.sendMu.Lock()
	defer st.sendMu.Unlock()
	return stream.Send(msg)
}

// callContext 从鉴权取出租户与实例名；namespace 优先请求，其次握手默认值。
func (h *Handler) callContext(ctx context.Context, namespaceID string, st *connState) contract.CallContext {
	cc, _ := contract.FromContext(ctx)
	if cc.TenantID == "" && h.cfg != nil {
		cc.TenantID = h.cfg.TenantID
	}
	if cc.CenterInstanceName == "" && h.cfg != nil {
		cc.CenterInstanceName = h.cfg.InstanceName
	}
	if cc.Environment == "" && h.cfg != nil {
		cc.Environment = h.cfg.Environment
	}
	switch {
	case namespaceID != "":
		cc.NamespaceID = namespaceID
	case st != nil && st.defaultNS != "":
		cc.NamespaceID = st.defaultNS
	}
	if st != nil {
		cc.ConnectionID = st.id
	}
	return cc
}

// Connect 是数据面唯一 RPC。先握手，再处理业务；断连先取消订阅再关通道。
func (h *Handler) Connect(stream pb.ServiceCenterStream_ConnectServer) error {
	ctx := stream.Context()
	clientIP := ""
	if p, ok := peer.FromContext(ctx); ok {
		if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
			clientIP = host
		}
	}
	st := &connState{
		id:       random.Generate32BitRandomString(),
		clientIP: clientIP,
		// 256 只是缓冲；满了由 publish 阻塞等待，不丢事件。
		namingCh: make(chan model.NamingEvent, 256),
		configCh: make(chan model.ConfigEvent, 256),
	}
	info := &model.ConnectionInfo{ConnectionID: st.id, ClientIP: clientIP, LastActive: time.Now()}
	if h.session != nil {
		h.session.Add(info)
	}
	h.track(st, stream)
	fwdCtx, cancelFwd := context.WithCancel(ctx)
	fwdDone := make(chan struct{})
	go func() {
		defer close(fwdDone)
		h.forwardEvents(fwdCtx, stream, st)
	}()

	defer func() {
		bound := 0
		if n, ok := h.naming.(*naming.App); ok {
			bound = n.BoundNodeCount(st.id)
		}
		reason := model.DisconnectClientLost
		if st.drain.Load() {
			reason = model.DisconnectServerDrain
		}
		alert.ConnectionLost(h.cfg, alert.ConnectionInfo{
			ConnectionID:        st.id,
			ClientID:            st.clientID,
			ClientIP:            clientIP,
			RegisteredNodeCount: bound,
			Reason:              reason,
		})
		h.naming.OnClientLost(context.Background(), st.id, reason)
		h.config.OnClientLost(context.Background(), st.id)
		h.untrack(st.id)
		if h.session != nil {
			h.session.Remove(st.id)
		}
		cancelFwd()
		close(st.namingCh)
		close(st.configCh)
		<-fwdDone
	}()

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if h.session != nil {
			h.session.Touch(st.id)
		}
		if err := h.route(stream, ctx, st, msg); err != nil {
			_ = h.sendError(st, stream, msg.GetRequestId(), err)
		}
	}
}

// forwardEvents 把订阅/Watch 事件推给客户端；通道关闭后退出。
func (h *Handler) forwardEvents(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-st.namingCh:
			if !ok {
				return
			}
			_ = h.send(st, stream, &pb.ServerMessage{
				MessageType: pb.ServerMessageType_SERVER_SERVICE_CHANGE,
				Message:     &pb.ServerMessage_ServiceChange{ServiceChange: namingEventToPB(ev)},
			})
		case ev, ok := <-st.configCh:
			if !ok {
				return
			}
			_ = h.send(st, stream, &pb.ServerMessage{
				MessageType: pb.ServerMessageType_SERVER_CONFIG_CHANGE,
				Message:     &pb.ServerMessage_ConfigChange{ConfigChange: configEventToPB(ev)},
			})
		}
	}
}

// route 按 opcode 一对一分发。除 HANDSHAKE/PING 外必须先握手。
func (h *Handler) route(stream pb.ServiceCenterStream_ConnectServer, ctx context.Context, st *connState, msg *pb.ClientMessage) error {
	switch msg.GetMessageType() {
	case pb.ClientMessageType_CLIENT_HANDSHAKE:
		return h.handshake(stream, st, msg)
	case pb.ClientMessageType_CLIENT_PING:
		return h.pong(stream, st, msg)
	}
	if !st.ready {
		return errHandshakeRequired
	}
	switch msg.GetMessageType() {
	case pb.ClientMessageType_CLIENT_REGISTER_SERVICE:
		return h.registerService(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_UNREGISTER_SERVICE:
		return h.unregisterService(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_REGISTER_NODE:
		return h.registerNode(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_UNREGISTER_NODE:
		return h.unregisterNode(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_DISCOVER_NODES:
		return h.discover(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_HEARTBEAT:
		return h.heartbeat(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_GET_SERVICE:
		return h.getService(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_LIST_SERVICES:
		return h.listServices(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_SUBSCRIBE_SERVICES:
		return h.subscribeServices(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_SUBSCRIBE_NAMESPACE:
		return h.subscribeNamespace(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_UNSUBSCRIBE:
		return h.unsubscribe(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_UNSUBSCRIBE_SERVICES:
		return h.unsubscribeServices(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_GET_CONFIG:
		return h.getConfig(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_GET_DRAFT:
		return h.getDraft(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_SAVE_CONFIG:
		return h.saveDraft(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_PUBLISH_CONFIG:
		return h.publish(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_DELETE_CONFIG:
		return h.deleteConfig(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_LIST_CONFIGS:
		return h.listConfigs(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_WATCH_CONFIG:
		return h.watchConfig(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_UNWATCH_CONFIG:
		return h.unwatch(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_UNWATCH_CONFIGS:
		return h.unwatchConfigs(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_GET_CONFIG_HISTORY:
		return h.configHistory(ctx, stream, st, msg)
	case pb.ClientMessageType_CLIENT_ROLLBACK_CONFIG:
		return h.rollback(ctx, stream, st, msg)
	default:
		return errUnsupported
	}
}

// handshake 登记 clientId 与默认命名空间，并回协议版本。
func (h *Handler) handshake(stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetHandshake()
	if in != nil {
		if md := in.GetMetadata(); md != nil {
			st.clientID = md.GetClientId()
		}
		st.defaultNS = in.GetNamespaceId()
	}
	if st.defaultNS != "" {
		cc := h.callContext(stream.Context(), st.defaultNS, st)
		if !cc.AllowsNamespace(st.defaultNS) {
			return contract.ErrNamespaceForbidden
		}
	}
	st.ready = true
	if h.session != nil {
		h.session.Add(&model.ConnectionInfo{
			ConnectionID: st.id,
			ClientID:     st.clientID,
			ClientIP:     st.clientIP,
			NamespaceID:  st.defaultNS,
			LastActive:   time.Now(),
		})
	}
	cc := h.callContext(stream.Context(), "", st)
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_HANDSHAKE,
		Message: &pb.ServerMessage_Handshake{Handshake: &pb.ServerHandshake{
			Success:         true,
			ConnectionId:    st.id,
			ServerTime:      time.Now().UnixMilli(),
			TenantId:        cc.TenantID,
			ProtocolVersion: ProtocolVersion,
			ServerInfo: map[string]string{
				"protocol": ProtocolVersion,
			},
		}},
	})
}

// pong 回显客户端时间戳。
func (h *Handler) pong(stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	ts := int64(0)
	if ping := msg.GetPing(); ping != nil {
		ts = ping.GetTimestamp()
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_PONG,
		Message:     &pb.ServerMessage_Pong{Pong: &pb.ServerPong{Timestamp: time.Now().UnixMilli(), ClientTimestamp: ts}},
	})
}

func (h *Handler) sendAck(st *connState, stream pb.ServiceCenterStream_ConnectServer, requestID string) error {
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   requestID,
		MessageType: pb.ServerMessageType_SERVER_ACK,
		Message:     &pb.ServerMessage_Ack{Ack: &pb.Ack{Success: true, Code: "OK"}},
	})
}

func (h *Handler) sendError(st *connState, stream pb.ServiceCenterStream_ConnectServer, requestID string, err error) error {
	code, msg := errorCode(err)
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   requestID,
		MessageType: pb.ServerMessageType_SERVER_ERROR,
		Message:     &pb.ServerMessage_Error{Error: &pb.ErrorResponse{Code: code, Message: msg}},
	})
}

func (h *Handler) track(st *connState, stream pb.ServiceCenterStream_ConnectServer) {
	if st == nil {
		return
	}
	h.liveMu.Lock()
	h.live[st.id] = liveStream{st: st, stream: stream}
	h.liveMu.Unlock()
}

func (h *Handler) untrack(connectionID string) {
	h.liveMu.Lock()
	delete(h.live, connectionID)
	h.liveMu.Unlock()
}

// NotifyClose 向当前全部数据面连接下发 SERVER_CLOSE。
// reason=server_shutdown 时标记 drain，流结束后 OnClientLost 交还 Owner 而不驱逐节点。
func (h *Handler) NotifyClose(reason, message string, grace int32) {
	if reason == "" {
		reason = "server_shutdown"
	}
	msg := &pb.ServerMessage{
		MessageType: pb.ServerMessageType_SERVER_CLOSE,
		Message: &pb.ServerMessage_Close{Close: &pb.ServerCloseNotification{
			Reason:      reason,
			Message:     message,
			GracePeriod: grace,
		}},
	}
	h.liveMu.Lock()
	conns := make([]liveStream, 0, len(h.live))
	for _, c := range h.live {
		conns = append(conns, c)
	}
	h.liveMu.Unlock()
	drain := model.IsServerDrain(reason)
	for _, c := range conns {
		if drain {
			c.st.drain.Store(true)
		}
		_ = h.send(c.st, c.stream, msg)
	}
}
