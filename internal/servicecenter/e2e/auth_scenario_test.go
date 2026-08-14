package e2e_test

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"testing"
	"time"

	"gateway/internal/servicecenter/server/interceptor"
	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/testutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TestAuthScenarios_SQLite 认证开启后的正反路径：
// 无凭证拒绝、Basic 错/对、API Token 错/对、JWT 错/过期/对。
func TestAuthScenarios_SQLite(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.Start(ctx, testutil.Options{EnableAuth: true})
	if err != nil {
		t.Fatalf("启动认证环境失败: %v", err)
	}
	t.Cleanup(func() {
		_ = env.Stop(context.Background())
	})

	conn, err := grpc.NewClient(env.Endpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	registry := pb.NewServiceRegistryClient(conn)

	register := func(callCtx context.Context, serviceName string) (*pb.RegisterServiceResponse, error) {
		return registry.RegisterService(callCtx, &pb.Service{
			NamespaceId: env.NamespaceID,
			GroupName:   env.GroupName,
			ServiceName: serviceName,
			ServiceType: "INTERNAL",
			Node: &pb.Node{
				NamespaceId:    env.NamespaceID,
				GroupName:      env.GroupName,
				ServiceName:    serviceName,
				IpAddress:      "127.0.0.1",
				PortNumber:     19001,
				InstanceStatus: "UP",
				HealthyStatus:  "HEALTHY",
				Ephemeral:      "Y",
				Weight:         100,
			},
		})
	}

	t.Run("reject_without_credentials", func(t *testing.T) {
		_, err := register(ctx, "auth-deny-no-cred")
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("期望 Unauthenticated, got code=%v err=%v", status.Code(err), err)
		}
	})

	t.Run("reject_wrong_basic", func(t *testing.T) {
		callCtx := withBasicAuth(ctx, env.AuthUserID, "wrong-password")
		_, err := register(callCtx, "auth-deny-basic")
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("期望 Unauthenticated, got code=%v err=%v", status.Code(err), err)
		}
	})

	t.Run("accept_basic", func(t *testing.T) {
		callCtx := withBasicAuth(ctx, env.AuthUserID, env.AuthPassword)
		resp, err := register(callCtx, "auth-ok-basic")
		if err != nil || !resp.GetSuccess() {
			t.Fatalf("Basic 应成功: err=%v success=%v msg=%s", err, resp.GetSuccess(), resp.GetMessage())
		}
		t.Logf("Basic ok nodeId=%s", resp.GetNodeId())
	})

	t.Run("reject_wrong_api_token", func(t *testing.T) {
		callCtx := withBearer(ctx, "definitely-not-a-valid-token")
		_, err := register(callCtx, "auth-deny-token")
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("期望 Unauthenticated, got code=%v err=%v", status.Code(err), err)
		}
	})

	t.Run("accept_api_token", func(t *testing.T) {
		callCtx := withBearer(ctx, env.AuthAPIToken)
		resp, err := register(callCtx, "auth-ok-token")
		if err != nil || !resp.GetSuccess() {
			t.Fatalf("API Token 应成功: err=%v success=%v msg=%s", err, resp.GetSuccess(), resp.GetMessage())
		}
		t.Logf("API Token ok nodeId=%s", resp.GetNodeId())
	})

	t.Run("reject_expired_jwt", func(t *testing.T) {
		tok, err := interceptor.SignJWT(env.AuthUserID, env.TenantID, env.AuthJwtSecret, env.AuthJwtIssuer, -time.Minute)
		if err != nil {
			t.Fatalf("SignJWT: %v", err)
		}
		callCtx := withBearer(ctx, tok)
		_, err = register(callCtx, "auth-deny-jwt-exp")
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("期望 Unauthenticated, got code=%v err=%v", status.Code(err), err)
		}
	})

	t.Run("reject_jwt_wrong_secret", func(t *testing.T) {
		tok, err := interceptor.SignJWT(env.AuthUserID, env.TenantID, "wrong-secret", env.AuthJwtIssuer, time.Hour)
		if err != nil {
			t.Fatalf("SignJWT: %v", err)
		}
		callCtx := withBearer(ctx, tok)
		_, err = register(callCtx, "auth-deny-jwt-sig")
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("期望 Unauthenticated, got code=%v err=%v", status.Code(err), err)
		}
	})

	t.Run("accept_jwt", func(t *testing.T) {
		tok, err := interceptor.SignJWT(env.AuthUserID, env.TenantID, env.AuthJwtSecret, env.AuthJwtIssuer, time.Hour)
		if err != nil {
			t.Fatalf("SignJWT: %v", err)
		}
		callCtx := withBearer(ctx, tok)
		resp, err := register(callCtx, "auth-ok-jwt")
		if err != nil || !resp.GetSuccess() {
			t.Fatalf("JWT 应成功: err=%v success=%v msg=%s", err, resp.GetSuccess(), resp.GetMessage())
		}
		t.Logf("JWT ok nodeId=%s", resp.GetNodeId())
	})

	t.Run("stream_reject_without_credentials", func(t *testing.T) {
		stream, err := pb.NewServiceCenterStreamClient(conn).Connect(ctx)
		if err != nil {
			t.Fatalf("Connect: %v", err)
		}
		err = stream.Send(&pb.ClientMessage{
			RequestId:   "hs",
			MessageType: pb.ClientMessageType_CLIENT_HANDSHAKE,
			Message: &pb.ClientMessage_Handshake{
				Handshake: &pb.ClientHandshake{
					Metadata: &pb.ClientMetadata{
						ClientId:      "auth-stream-no-cred",
						Language:      "go",
						ClientVersion: "e2e",
					},
					NamespaceId: env.NamespaceID,
				},
			},
		})
		if err == nil {
			_, err = stream.Recv()
		}
		// 拦截器返回 Unauthenticated；部分时机表现为流直接 EOF/Canceled
		code := status.Code(err)
		ok := err != nil && (code == codes.Unauthenticated || code == codes.Canceled || errors.Is(err, io.EOF))
		if !ok {
			t.Fatalf("Stream 无认证应被拒绝, got code=%v err=%v", code, err)
		}
		t.Logf("stream without credentials rejected: code=%v err=%v", code, err)
	})

	t.Run("stream_accept_basic", func(t *testing.T) {
		callCtx := withBasicAuth(ctx, env.AuthUserID, env.AuthPassword)
		stream, err := pb.NewServiceCenterStreamClient(conn).Connect(callCtx)
		if err != nil {
			t.Fatalf("Connect: %v", err)
		}
		if err := stream.Send(&pb.ClientMessage{
			RequestId:   "hs-ok",
			MessageType: pb.ClientMessageType_CLIENT_HANDSHAKE,
			Message: &pb.ClientMessage_Handshake{
				Handshake: &pb.ClientHandshake{
					Metadata: &pb.ClientMetadata{
						ClientId:      "auth-stream-basic",
						Language:      "go",
						ClientVersion: "e2e",
					},
					NamespaceId: env.NamespaceID,
				},
			},
		}); err != nil {
			t.Fatalf("handshake send: %v", err)
		}
		msg, err := stream.Recv()
		if err != nil {
			t.Fatalf("handshake recv: %v", err)
		}
		if msg.GetMessageType() != pb.ServerMessageType_SERVER_HANDSHAKE {
			t.Fatalf("期望 SERVER_HANDSHAKE, got %v", msg.GetMessageType())
		}
		t.Logf("stream Basic handshake ok connectionId=%s", msg.GetHandshake().GetConnectionId())
		_ = stream.CloseSend()
	})
}

func withBasicAuth(ctx context.Context, userID, password string) context.Context {
	encoded := base64.StdEncoding.EncodeToString([]byte(userID + ":" + password))
	md := metadata.Pairs("authorization", "Basic "+encoded)
	return metadata.NewOutgoingContext(ctx, md)
}

func withBearer(ctx context.Context, token string) context.Context {
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewOutgoingContext(ctx, md)
}
