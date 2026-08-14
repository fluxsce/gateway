package e2e_test

import (
	"context"
	"io"
	"testing"
	"time"

	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/testutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestStreamControlMessages_SQLite 覆盖 Stream 控制面：
// CLIENT_SUBSCRIBE_NAMESPACE、SERVER_ERROR（未知消息类型）、SERVER_CLOSE（实例关闭）。
func TestStreamControlMessages_SQLite(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.Start(ctx, testutil.Options{})
	if err != nil {
		t.Fatalf("启动 SQLite 服务中心失败: %v", err)
	}
	t.Cleanup(func() {
		_ = env.Stop(context.Background())
	})

	conn, err := grpc.NewClient(env.Endpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("连接 gRPC 失败: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	streamClient := pb.NewServiceCenterStreamClient(conn)
	registry := pb.NewServiceRegistryClient(conn)
	ns := env.NamespaceID
	group := env.GroupName

	t.Run("client_subscribe_namespace_push", func(t *testing.T) {
		streamCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		stream, err := streamClient.Connect(streamCtx)
		if err != nil {
			t.Fatalf("Connect: %v", err)
		}

		if err := stream.Send(&pb.ClientMessage{
			RequestId:   "hs-ns",
			MessageType: pb.ClientMessageType_CLIENT_HANDSHAKE,
			Message: &pb.ClientMessage_Handshake{
				Handshake: &pb.ClientHandshake{
					Metadata: &pb.ClientMetadata{
						ClientId:      "e2e-stream-ns",
						Language:      "go",
						ClientVersion: "e2e",
					},
					NamespaceId:    ns,
					SubscribeTypes: []string{"registry"},
				},
			},
		}); err != nil {
			t.Fatalf("handshake send: %v", err)
		}

		hs, err := stream.Recv()
		if err != nil || hs.GetMessageType() != pb.ServerMessageType_SERVER_HANDSHAKE {
			t.Fatalf("handshake recv: err=%v type=%v", err, hs.GetMessageType())
		}

		if err := stream.Send(&pb.ClientMessage{
			RequestId:   "sub-ns",
			MessageType: pb.ClientMessageType_CLIENT_SUBSCRIBE_NAMESPACE,
			Message: &pb.ClientMessage_SubscribeNamespace{
				SubscribeNamespace: &pb.SubscribeNamespaceRequest{
					NamespaceId: ns,
					GroupName:   group,
				},
			},
		}); err != nil {
			t.Fatalf("subscribe namespace send: %v", err)
		}

		eventCh := make(chan *pb.ServiceChangeEvent, 8)
		errCh := make(chan error, 1)
		go func() {
			for {
				msg, recvErr := stream.Recv()
				if recvErr != nil {
					if recvErr == io.EOF || streamCtx.Err() != nil {
						return
					}
					errCh <- recvErr
					return
				}
				if msg.GetMessageType() == pb.ServerMessageType_SERVER_SERVICE_CHANGE &&
					msg.GetServiceChange() != nil {
					eventCh <- msg.GetServiceChange()
				}
			}
		}()

		time.Sleep(500 * time.Millisecond)

		_, err = registry.RegisterService(ctx, &pb.Service{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: "e2e-stream-ns-svc",
			ServiceType: "INTERNAL",
			Node: &pb.Node{
				NamespaceId:    ns,
				GroupName:      group,
				ServiceName:    "e2e-stream-ns-svc",
				IpAddress:      "127.0.0.1",
				PortNumber:     18301,
				InstanceStatus: "UP",
				HealthyStatus:  "HEALTHY",
				Ephemeral:      "Y",
				Weight:         100,
			},
		})
		if err != nil {
			t.Fatalf("RegisterService: %v", err)
		}

		select {
		case ev := <-eventCh:
			if ev.GetServiceName() == "" {
				t.Fatalf("空 serviceName: %+v", ev)
			}
			t.Logf("CLIENT_SUBSCRIBE_NAMESPACE push ok type=%s service=%s",
				ev.GetEventType(), ev.GetServiceName())
		case recvErr := <-errCh:
			t.Fatalf("stream recv: %v", recvErr)
		case <-time.After(10 * time.Second):
			t.Fatal("超时未收到 Stream 命名空间推送")
		}
		_ = stream.CloseSend()
	})

	t.Run("server_error_on_unknown_message", func(t *testing.T) {
		streamCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		stream, err := streamClient.Connect(streamCtx)
		if err != nil {
			t.Fatalf("Connect: %v", err)
		}

		if err := stream.Send(&pb.ClientMessage{
			RequestId:   "hs-err",
			MessageType: pb.ClientMessageType_CLIENT_HANDSHAKE,
			Message: &pb.ClientMessage_Handshake{
				Handshake: &pb.ClientHandshake{
					Metadata: &pb.ClientMetadata{
						ClientId:      "e2e-stream-err",
						Language:      "go",
						ClientVersion: "e2e",
					},
					NamespaceId: ns,
				},
			},
		}); err != nil {
			t.Fatalf("handshake send: %v", err)
		}
		if _, err := stream.Recv(); err != nil {
			t.Fatalf("handshake recv: %v", err)
		}

		if err := stream.Send(&pb.ClientMessage{
			RequestId:   "bad-msg",
			MessageType: pb.ClientMessageType_CLIENT_UNKNOWN,
		}); err != nil {
			t.Fatalf("send unknown: %v", err)
		}

		deadline := time.After(5 * time.Second)
		for {
			select {
			case <-deadline:
				t.Fatal("超时未收到 SERVER_ERROR")
			default:
			}
			msg, err := stream.Recv()
			if err != nil {
				t.Fatalf("recv: %v", err)
			}
			if msg.GetMessageType() == pb.ServerMessageType_SERVER_ERROR {
				t.Logf("SERVER_ERROR ok requestId=%s msg=%s",
					msg.GetRequestId(), msg.GetError().GetMessage())
				_ = stream.CloseSend()
				return
			}
		}
	})

	t.Run("server_close_payload_unit_covered", func(t *testing.T) {
		// Server.Stop 走 grpc GracefulStop，不一定下发 SERVER_CLOSE 载荷。
		// SERVER_CLOSE 广播路径由 connection.TestConnectionManager_Close_SendsServerClose 覆盖（reason=server_shutdown）。
		t.Log("SERVER_CLOSE payload asserted in connection package unit test")
	})
}
