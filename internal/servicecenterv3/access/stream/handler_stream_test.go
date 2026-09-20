package stream

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"gateway/internal/servicecenterv3/config"
	"gateway/internal/servicecenterv3/infra/cache"
	"gateway/internal/servicecenterv3/model"
	"gateway/internal/servicecenterv3/naming"
	pb "gateway/internal/servicecenterv3/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func startStream(t *testing.T) (pb.ServiceCenterStreamClient, *naming.App, func()) {
	t.Helper()
	c := cache.New("center-1")
	c.SetNamespace(&model.Namespace{
		TenantID:           "t1",
		CenterInstanceName: "center-1",
		NamespaceID:        "ns",
		Active:             true,
	})
	nam := naming.New("center-1", nil, c)
	cfgApp := config.New("center-1", nil, c)
	h := NewHandler(&model.CenterInstance{InstanceName: "center-1", TenantID: "t1"}, nam, cfgApp, nil)

	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer()
	pb.RegisterServiceCenterStreamServer(gs, h)
	go func() { _ = gs.Serve(lis) }()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	stop := func() {
		_ = conn.Close()
		gs.Stop()
		_ = lis.Close()
	}
	return pb.NewServiceCenterStreamClient(conn), nam, stop
}

func openSession(t *testing.T, cli pb.ServiceCenterStreamClient) pb.ServiceCenterStream_ConnectClient {
	t.Helper()
	stream, err := openSessionErr(cli)
	if err != nil {
		t.Fatal(err)
	}
	return stream
}

func openSessionErr(cli pb.ServiceCenterStreamClient) (pb.ServiceCenterStream_ConnectClient, error) {
	stream, err := cli.Connect(context.Background())
	if err != nil {
		return nil, err
	}
	if err := stream.Send(&pb.ClientMessage{
		RequestId:   "hs",
		MessageType: pb.ClientMessageType_CLIENT_HANDSHAKE,
		Message: &pb.ClientMessage_Handshake{Handshake: &pb.ClientHandshake{
			NamespaceId:     "ns",
			ProtocolVersion: ProtocolVersion,
			Metadata:        &pb.ClientMetadata{ClientId: "c"},
		}},
	}); err != nil {
		return nil, err
	}
	msg, err := recvUntilErr(stream, func(m *pb.ServerMessage) bool {
		return m.GetMessageType() == pb.ServerMessageType_SERVER_HANDSHAKE
	})
	if err != nil {
		return nil, err
	}
	if !msg.GetHandshake().GetSuccess() {
		return nil, fmt.Errorf("handshake: %+v", msg.GetHandshake())
	}
	return stream, nil
}

func TestStreamPeerOfflineNotifiesSubscriber(t *testing.T) {
	cli, nam, stop := startStream(t)
	defer stop()

	watcher := openSession(t, cli)
	mustSend(t, watcher, &pb.ClientMessage{
		RequestId:   "sub",
		MessageType: pb.ClientMessageType_CLIENT_SUBSCRIBE_SERVICES,
		Message: &pb.ClientMessage_SubscribeServices{SubscribeServices: &pb.SubscribeServicesRequest{
			NamespaceId:  "ns",
			GroupName:    "g",
			ServiceNames: []string{"svc"},
		}},
	})
	_ = recvUntil(t, watcher, func(m *pb.ServerMessage) bool {
		return m.GetMessageType() == pb.ServerMessageType_SERVER_ACK
	})

	provider := openSession(t, cli)
	mustSend(t, provider, &pb.ClientMessage{
		RequestId:   "reg",
		MessageType: pb.ClientMessageType_CLIENT_REGISTER_NODE,
		Message: &pb.ClientMessage_RegisterNode{RegisterNode: &pb.Node{
			NamespaceId: "ns",
			GroupName:   "g",
			ServiceName: "svc",
			IpAddress:   "10.0.0.8",
			PortNumber:  8080,
		}},
	})
	reg := recvUntil(t, provider, func(m *pb.ServerMessage) bool {
		return m.GetMessageType() == pb.ServerMessageType_SERVER_REGISTER_NODE
	})
	nodeID := reg.GetRegisterNode().GetNodeId()
	if nodeID == "" {
		t.Fatal("empty nodeId")
	}

	added := recvUntil(t, watcher, func(m *pb.ServerMessage) bool {
		return m.GetMessageType() == pb.ServerMessageType_SERVER_SERVICE_CHANGE &&
			m.GetServiceChange().GetEventType() == pb.NamingEventType_NAMING_EVENT_NODE_REGISTERED
	})
	if added.GetServiceChange().GetChangedNode().GetNodeId() != nodeID {
		t.Fatalf("added node %s", added.GetServiceChange().GetChangedNode().GetNodeId())
	}

	_ = provider.CloseSend()
	evicted := recvUntil(t, watcher, func(m *pb.ServerMessage) bool {
		return m.GetMessageType() == pb.ServerMessageType_SERVER_SERVICE_CHANGE &&
			m.GetServiceChange().GetEventType() == pb.NamingEventType_NAMING_EVENT_NODE_EVICTED
	})
	if evicted.GetServiceChange().GetChangedNode().GetNodeId() != nodeID {
		t.Fatalf("evicted node %s", evicted.GetServiceChange().GetChangedNode().GetNodeId())
	}
	if _, ok := nam.Cache().GetNode(nodeID); ok {
		t.Fatal("ephemeral node should be gone after provider disconnect")
	}
}

func TestStreamConcurrentRegister(t *testing.T) {
	cli, nam, stop := startStream(t)
	defer stop()

	const n = 16
	done := make(chan string, n)
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			stream, err := openSessionErr(cli)
			if err != nil {
				errCh <- err
				return
			}
			if err := stream.Send(&pb.ClientMessage{
				RequestId:   "r",
				MessageType: pb.ClientMessageType_CLIENT_REGISTER_NODE,
				Message: &pb.ClientMessage_RegisterNode{RegisterNode: &pb.Node{
					NamespaceId: "ns",
					GroupName:   "g",
					ServiceName: "svc",
					IpAddress:   "10.0.1.1",
					PortNumber:  int32(9000 + i),
				}},
			}); err != nil {
				errCh <- err
				return
			}
			msg, err := recvUntilErr(stream, func(m *pb.ServerMessage) bool {
				return m.GetMessageType() == pb.ServerMessageType_SERVER_REGISTER_NODE ||
					m.GetMessageType() == pb.ServerMessageType_SERVER_ERROR
			})
			if err != nil {
				errCh <- err
				return
			}
			if msg.GetMessageType() == pb.ServerMessageType_SERVER_ERROR {
				errCh <- io.ErrUnexpectedEOF
				return
			}
			done <- msg.GetRegisterNode().GetNodeId()
		}()
	}
	ids := map[string]struct{}{}
	for i := 0; i < n; i++ {
		select {
		case id := <-done:
			if id == "" {
				t.Fatal("empty node id")
			}
			ids[id] = struct{}{}
		case err := <-errCh:
			t.Fatal(err)
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent register timeout")
		}
	}
	if len(ids) != n {
		t.Fatalf("want %d unique nodes, got %d", n, len(ids))
	}
	for id := range ids {
		if _, ok := nam.Cache().GetNode(id); !ok {
			t.Fatalf("instance %s missing from cache", id)
		}
	}
}

func TestStreamNotifyClose(t *testing.T) {
	c := cache.New("center-1")
	c.SetNamespace(&model.Namespace{
		TenantID:           "t1",
		CenterInstanceName: "center-1",
		NamespaceID:        "ns",
		Active:             true,
	})
	nam := naming.New("center-1", nil, c)
	cfgApp := config.New("center-1", nil, c)
	h := NewHandler(&model.CenterInstance{InstanceName: "center-1", TenantID: "t1"}, nam, cfgApp, nil)

	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer()
	pb.RegisterServiceCenterStreamServer(gs, h)
	go func() { _ = gs.Serve(lis) }()
	defer gs.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	stream := openSession(t, pb.NewServiceCenterStreamClient(conn))
	h.NotifyClose("server_shutdown", "bye", 5)
	msg := recvUntil(t, stream, func(m *pb.ServerMessage) bool {
		return m.GetMessageType() == pb.ServerMessageType_SERVER_CLOSE
	})
	if msg.GetClose().GetReason() != "server_shutdown" {
		t.Fatalf("close reason=%s", msg.GetClose().GetReason())
	}
}

func TestStreamServerDrainKeepsRegisteredNode(t *testing.T) {
	c := cache.New("center-1")
	c.SetNamespace(&model.Namespace{
		TenantID:           "t1",
		CenterInstanceName: "center-1",
		NamespaceID:        "ns",
		Active:             true,
	})
	nam := naming.New("center-1", nil, c)
	cfgApp := config.New("center-1", nil, c)
	h := NewHandler(&model.CenterInstance{InstanceName: "center-1", TenantID: "t1"}, nam, cfgApp, nil)

	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer()
	pb.RegisterServiceCenterStreamServer(gs, h)
	go func() { _ = gs.Serve(lis) }()
	defer gs.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	stream := openSession(t, pb.NewServiceCenterStreamClient(conn))
	mustSend(t, stream, &pb.ClientMessage{
		RequestId:   "reg",
		MessageType: pb.ClientMessageType_CLIENT_REGISTER_NODE,
		Message: &pb.ClientMessage_RegisterNode{RegisterNode: &pb.Node{
			NamespaceId: "ns",
			GroupName:   "g",
			ServiceName: "svc",
			IpAddress:   "10.0.0.9",
			PortNumber:  8081,
		}},
	})
	reg := recvUntil(t, stream, func(m *pb.ServerMessage) bool {
		return m.GetMessageType() == pb.ServerMessageType_SERVER_REGISTER_NODE
	})
	nodeID := reg.GetRegisterNode().GetNodeId()
	if nodeID == "" {
		t.Fatal("empty nodeId")
	}
	h.NotifyClose("server_shutdown", "restart", 1)
	_ = stream.CloseSend()
	deadline := time.Now().Add(2 * time.Second)
	var got *model.Node
	var ok bool
	for time.Now().Before(deadline) {
		got, ok = nam.Cache().GetNode(nodeID)
		if ok && got.ConnectionID == "" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !ok {
		t.Fatal("server drain must not evict the node; SDK will reconnect")
	}
	if got.ConnectionID != "" {
		t.Fatalf("drain should unbind connection, got %s", got.ConnectionID)
	}
}

func mustSend(t *testing.T, stream pb.ServiceCenterStream_ConnectClient, msg *pb.ClientMessage) {
	t.Helper()
	if err := stream.Send(msg); err != nil {
		t.Fatal(err)
	}
}

func recvUntil(t *testing.T, stream pb.ServiceCenterStream_ConnectClient, ok func(*pb.ServerMessage) bool) *pb.ServerMessage {
	t.Helper()
	msg, err := recvUntilErr(stream, ok)
	if err != nil {
		t.Fatal(err)
	}
	return msg
}

func recvUntilErr(stream pb.ServiceCenterStream_ConnectClient, ok func(*pb.ServerMessage) bool) (*pb.ServerMessage, error) {
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		type result struct {
			msg *pb.ServerMessage
			err error
		}
		ch := make(chan result, 1)
		go func() {
			m, err := stream.Recv()
			ch <- result{m, err}
		}()
		select {
		case r := <-ch:
			if r.err != nil {
				return nil, r.err
			}
			if ok(r.msg) {
				return r.msg, nil
			}
		case <-time.After(time.Until(deadline)):
			return nil, fmt.Errorf("recv timeout")
		}
	}
	return nil, fmt.Errorf("recv timeout")
}
