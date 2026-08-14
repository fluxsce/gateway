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

// TestBusinessScenarios_SQLite 真实业务场景：SQLite 服务端 + gRPC 客户端。
// 覆盖注册发现、命名空间订阅推送、配置发布与监听、注销。
func TestBusinessScenarios_SQLite(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.Start(ctx, testutil.Options{})
	if err != nil {
		t.Fatalf("启动 SQLite 服务中心失败: %v", err)
	}
	t.Cleanup(func() {
		if stopErr := env.Stop(context.Background()); stopErr != nil {
			t.Logf("停止环境失败: %v", stopErr)
		}
	})

	conn, err := grpc.NewClient(env.Endpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("连接 gRPC 失败: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	registry := pb.NewServiceRegistryClient(conn)
	configClient := pb.NewConfigCenterClient(conn)

	const serviceName = "e2e-order-service"
	ns := env.NamespaceID
	group := env.GroupName

	t.Run("register_heartbeat_discover", func(t *testing.T) {
		regResp, err := registry.RegisterService(ctx, &pb.Service{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
			ServiceType: "INTERNAL",
			Node: &pb.Node{
				NamespaceId:    ns,
				GroupName:      group,
				ServiceName:    serviceName,
				IpAddress:      "127.0.0.1",
				PortNumber:     8081,
				InstanceStatus: "UP",
				HealthyStatus:  "HEALTHY",
				Ephemeral:      "Y",
				Weight:         100,
			},
		})
		if err != nil {
			t.Fatalf("RegisterService: %v", err)
		}
		if !regResp.GetSuccess() || regResp.GetNodeId() == "" {
			t.Fatalf("注册失败: success=%v msg=%s", regResp.GetSuccess(), regResp.GetMessage())
		}
		nodeID := regResp.GetNodeId()

		hb, err := registry.Heartbeat(ctx, &pb.HeartbeatRequest{
			NodeId: nodeID,
			Service: &pb.Service{
				NamespaceId: ns,
				GroupName:   group,
				ServiceName: serviceName,
				ServiceType: "INTERNAL",
				Node: &pb.Node{
					NodeId:         nodeID,
					NamespaceId:    ns,
					GroupName:      group,
					ServiceName:    serviceName,
					IpAddress:      "127.0.0.1",
					PortNumber:     8081,
					InstanceStatus: "UP",
					HealthyStatus:  "HEALTHY",
					Ephemeral:      "Y",
					Weight:         100,
				},
			},
		})
		if err != nil || hb == nil || !hb.GetSuccess() {
			msg := ""
			if hb != nil {
				msg = hb.GetMessage()
			}
			t.Fatalf("Heartbeat: err=%v success=%v msg=%s", err, hb != nil && hb.GetSuccess(), msg)
		}

		getResp, err := registry.GetService(ctx, &pb.ServiceKey{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
		})
		if err != nil || !getResp.GetSuccess() {
			t.Fatalf("GetService: err=%v success=%v", err, getResp.GetSuccess())
		}
		if len(getResp.GetNodes()) == 0 {
			t.Fatal("GetService 应返回节点列表")
		}

		discover, err := registry.DiscoverNodes(ctx, &pb.DiscoverNodesRequest{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
			HealthyOnly: true,
		})
		if err != nil || !discover.GetSuccess() || len(discover.GetNodes()) == 0 {
			t.Fatalf("DiscoverNodes: err=%v success=%v nodes=%d", err, discover.GetSuccess(), len(discover.GetNodes()))
		}

		t.Logf("registered nodeId=%s nodes=%d", nodeID, len(discover.GetNodes()))
	})

	t.Run("subscribe_namespace_push", func(t *testing.T) {
		subCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		stream, err := registry.SubscribeNamespace(subCtx, &pb.SubscribeNamespaceRequest{
			NamespaceId: ns,
			GroupName:   group,
		})
		if err != nil {
			t.Fatalf("SubscribeNamespace: %v", err)
		}

		eventCh := make(chan *pb.ServiceChangeEvent, 8)
		errCh := make(chan error, 1)
		go func() {
			for {
				ev, recvErr := stream.Recv()
				if recvErr != nil {
					if recvErr == io.EOF || subCtx.Err() != nil {
						return
					}
					errCh <- recvErr
					return
				}
				eventCh <- ev
			}
		}()

		// 等待服务端完成 SubscribeNamespace 注册，避免推送早于订阅就绪被丢弃
		time.Sleep(500 * time.Millisecond)

		// 再注册一个服务节点触发命名空间推送
		_, err = registry.RegisterService(ctx, &pb.Service{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: "e2e-pay-service",
			ServiceType: "INTERNAL",
			Node: &pb.Node{
				NamespaceId:    ns,
				GroupName:      group,
				ServiceName:    "e2e-pay-service",
				IpAddress:      "127.0.0.1",
				PortNumber:     8082,
				InstanceStatus: "UP",
				HealthyStatus:  "HEALTHY",
				Ephemeral:      "Y",
				Weight:         100,
			},
		})
		if err != nil {
			t.Fatalf("RegisterService for push: %v", err)
		}

		select {
		case ev := <-eventCh:
			if ev.GetServiceName() == "" {
				t.Fatalf("推送事件缺少 serviceName: %+v", ev)
			}
			t.Logf("收到命名空间推送: type=%s service=%s", ev.GetEventType(), ev.GetServiceName())
		case recvErr := <-errCh:
			t.Fatalf("订阅接收失败: %v", recvErr)
		case <-time.After(10 * time.Second):
			t.Fatal("超时未收到命名空间服务变更推送")
		}
	})

	t.Run("config_save_watch", func(t *testing.T) {
		dataID := "application-e2e.yaml"
		watchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		stream, err := configClient.WatchConfig(watchCtx, &pb.WatchConfigRequest{
			NamespaceId:   ns,
			GroupName:     group,
			ConfigDataIds: []string{dataID},
		})
		if err != nil {
			t.Fatalf("WatchConfig: %v", err)
		}

		eventCh := make(chan *pb.ConfigChangeEvent, 8)
		errCh := make(chan error, 1)
		go func() {
			for {
				ev, recvErr := stream.Recv()
				if recvErr != nil {
					if recvErr == io.EOF || watchCtx.Err() != nil {
						return
					}
					errCh <- recvErr
					return
				}
				eventCh <- ev
			}
		}()

		time.Sleep(500 * time.Millisecond)

		saveResp, err := configClient.SaveConfig(ctx, &pb.ConfigData{
			NamespaceId:   ns,
			GroupName:     group,
			ConfigDataId:  dataID,
			ConfigContent: "server:\n  port: 8080\n",
			ContentType:   "yaml",
			ConfigDesc:    "e2e config",
			ChangedBy:     "e2e",
		})
		if err != nil || saveResp == nil || !saveResp.GetSuccess() {
			msg := ""
			if saveResp != nil {
				msg = saveResp.GetMessage()
			}
			t.Fatalf("SaveConfig: err=%v success=%v msg=%s", err, saveResp != nil && saveResp.GetSuccess(), msg)
		}

		// Watch 对不存在的配置会先推 CONFIG_DELETED 初始快照，需等到带内容的更新事件
		deadline := time.After(10 * time.Second)
		for {
			select {
			case ev := <-eventCh:
				if ev.GetConfig() != nil && ev.GetConfig().GetConfigDataId() == dataID &&
					ev.GetConfig().GetConfigContent() != "" {
					t.Logf("收到配置推送: type=%s dataId=%s", ev.GetEventType(), ev.GetConfig().GetConfigDataId())
					goto configPushOK
				}
				t.Logf("忽略配置事件: type=%s dataId=%s", ev.GetEventType(), ev.GetConfigDataId())
			case recvErr := <-errCh:
				t.Fatalf("配置监听接收失败: %v", recvErr)
			case <-deadline:
				t.Fatal("超时未收到配置变更推送")
			}
		}
	configPushOK:

		getResp, err := configClient.GetConfig(ctx, &pb.ConfigKey{
			NamespaceId:  ns,
			GroupName:    group,
			ConfigDataId: dataID,
		})
		if err != nil || !getResp.GetSuccess() || getResp.GetConfig() == nil {
			t.Fatalf("GetConfig: err=%v success=%v", err, getResp.GetSuccess())
		}
		if getResp.GetConfig().GetConfigContent() == "" {
			t.Fatal("配置内容为空")
		}
	})

	t.Run("unregister_node", func(t *testing.T) {
		getResp, err := registry.GetService(ctx, &pb.ServiceKey{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
		})
		if err != nil || !getResp.GetSuccess() || len(getResp.GetNodes()) == 0 {
			t.Fatalf("预置节点不存在，无法测注销: err=%v", err)
		}
		nodeID := getResp.GetNodes()[0].GetNodeId()

		unreg, err := registry.UnregisterNode(ctx, &pb.NodeKey{NodeId: nodeID})
		if err != nil || !unreg.GetSuccess() {
			t.Fatalf("UnregisterNode: err=%v success=%v msg=%s", err, unreg.GetSuccess(), unreg.GetMessage())
		}

		discover, err := registry.DiscoverNodes(ctx, &pb.DiscoverNodesRequest{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
			HealthyOnly: false,
		})
		if err != nil {
			t.Fatalf("DiscoverNodes after unregister: %v", err)
		}
		for _, n := range discover.GetNodes() {
			if n.GetNodeId() == nodeID {
				t.Fatalf("注销后仍能发现节点 %s", nodeID)
			}
		}
	})
}
