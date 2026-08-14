package e2e_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"testing"

	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/testutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// TestTLSScenarios_SQLite 单向 TLS：明文应失败，信任 CA 后可注册。
func TestTLSScenarios_SQLite(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.Start(ctx, testutil.Options{EnableTLS: true})
	if err != nil {
		t.Fatalf("启动 TLS 环境失败: %v", err)
	}
	t.Cleanup(func() { _ = env.Stop(context.Background()) })
	if env.TLS == nil {
		t.Fatal("TLS material missing")
	}

	t.Run("plaintext_should_fail", func(t *testing.T) {
		conn, err := grpc.NewClient(env.Endpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		defer conn.Close()
		_, err = pb.NewServiceRegistryClient(conn).RegisterService(ctx, &pb.Service{
			NamespaceId: env.NamespaceID,
			GroupName:   env.GroupName,
			ServiceName: "tls-plain",
			ServiceType: "INTERNAL",
		})
		if err == nil {
			t.Fatal("明文连接 TLS 服务端应失败")
		}
		t.Logf("plaintext rejected: %v", err)
	})

	t.Run("tls_with_ca_ok", func(t *testing.T) {
		creds, err := clientTLSCreds(env.TLS.CACertFile, "")
		if err != nil {
			t.Fatalf("clientTLSCreds: %v", err)
		}
		conn, err := grpc.NewClient(env.Endpoint(), grpc.WithTransportCredentials(creds))
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		defer conn.Close()

		resp, err := pb.NewServiceRegistryClient(conn).RegisterService(ctx, &pb.Service{
			NamespaceId: env.NamespaceID,
			GroupName:   env.GroupName,
			ServiceName: "tls-ok-svc",
			ServiceType: "INTERNAL",
			Node: &pb.Node{
				NamespaceId:    env.NamespaceID,
				GroupName:      env.GroupName,
				ServiceName:    "tls-ok-svc",
				IpAddress:      "127.0.0.1",
				PortNumber:     19201,
				InstanceStatus: "UP",
				HealthyStatus:  "HEALTHY",
				Ephemeral:      "Y",
				Weight:         100,
			},
		})
		if err != nil || !resp.GetSuccess() {
			t.Fatalf("TLS 注册失败: err=%v success=%v msg=%s", err, resp.GetSuccess(), resp.GetMessage())
		}
		t.Logf("TLS ok nodeId=%s ca=%s", resp.GetNodeId(), env.TLS.CACertFile)
	})
}

// TestMTLSScenarios_SQLite 双向 TLS：无客户端证书失败，携带客户端证书成功。
func TestMTLSScenarios_SQLite(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.Start(ctx, testutil.Options{EnableMTLS: true})
	if err != nil {
		t.Fatalf("启动 mTLS 环境失败: %v", err)
	}
	t.Cleanup(func() { _ = env.Stop(context.Background()) })

	t.Run("tls_without_client_cert_should_fail", func(t *testing.T) {
		creds, err := clientTLSCreds(env.TLS.CACertFile, "")
		if err != nil {
			t.Fatalf("clientTLSCreds: %v", err)
		}
		conn, err := grpc.NewClient(env.Endpoint(), grpc.WithTransportCredentials(creds))
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		defer conn.Close()
		_, err = pb.NewServiceRegistryClient(conn).RegisterService(ctx, &pb.Service{
			NamespaceId: env.NamespaceID,
			GroupName:   env.GroupName,
			ServiceName: "mtls-no-client",
			ServiceType: "INTERNAL",
		})
		if err == nil {
			t.Fatal("mTLS 无客户端证书应失败")
		}
		t.Logf("mTLS without client cert rejected: %v", err)
	})

	t.Run("mtls_with_client_cert_ok", func(t *testing.T) {
		creds, err := clientMTLSCreds(env.TLS.CACertFile, env.TLS.ClientCert, env.TLS.ClientKey)
		if err != nil {
			t.Fatalf("clientMTLSCreds: %v", err)
		}
		conn, err := grpc.NewClient(env.Endpoint(), grpc.WithTransportCredentials(creds))
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		defer conn.Close()

		resp, err := pb.NewServiceRegistryClient(conn).RegisterService(ctx, &pb.Service{
			NamespaceId: env.NamespaceID,
			GroupName:   env.GroupName,
			ServiceName: "mtls-ok-svc",
			ServiceType: "INTERNAL",
			Node: &pb.Node{
				NamespaceId:    env.NamespaceID,
				GroupName:      env.GroupName,
				ServiceName:    "mtls-ok-svc",
				IpAddress:      "127.0.0.1",
				PortNumber:     19202,
				InstanceStatus: "UP",
				HealthyStatus:  "HEALTHY",
				Ephemeral:      "Y",
				Weight:         100,
			},
		})
		if err != nil || !resp.GetSuccess() {
			t.Fatalf("mTLS 注册失败: err=%v success=%v msg=%s", err, resp.GetSuccess(), resp.GetMessage())
		}
		t.Logf("mTLS ok nodeId=%s", resp.GetNodeId())
	})
}

func clientTLSCreds(caFile, serverName string) (credentials.TransportCredentials, error) {
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("invalid CA PEM")
	}
	cfg := &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
	}
	if serverName != "" {
		cfg.ServerName = serverName
	}
	return credentials.NewTLS(cfg), nil
}

func clientMTLSCreds(caFile, clientCert, clientKey string) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, err
	}
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("invalid CA PEM")
	}
	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		MinVersion:   tls.VersionTLS12,
	}), nil
}
