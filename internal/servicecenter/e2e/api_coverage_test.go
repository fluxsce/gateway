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

// TestAPICoverageGaps_SQLite 补齐此前 UNTESTED 的 Classic RPC：
// RegisterNode、UnregisterService、SubscribeServices、
// ListConfigs / DeleteConfig / GetConfigHistory / RollbackConfig。
func TestAPICoverageGaps_SQLite(t *testing.T) {
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
	ns := env.NamespaceID
	group := env.GroupName

	t.Run("register_node_standalone", func(t *testing.T) {
		const serviceName = "e2e-register-node-svc"
		regSvc, err := registry.RegisterService(ctx, &pb.Service{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
			ServiceType: "INTERNAL",
		})
		if err != nil || !regSvc.GetSuccess() {
			t.Fatalf("RegisterService(无节点): err=%v success=%v msg=%s",
				err, regSvc.GetSuccess(), regSvc.GetMessage())
		}

		regNode, err := registry.RegisterNode(ctx, &pb.Node{
			NamespaceId:    ns,
			GroupName:      group,
			ServiceName:    serviceName,
			IpAddress:      "127.0.0.1",
			PortNumber:     18081,
			InstanceStatus: "UP",
			HealthyStatus:  "HEALTHY",
			Ephemeral:      "Y",
			Weight:         100,
		})
		if err != nil || !regNode.GetSuccess() || regNode.GetNodeId() == "" {
			t.Fatalf("RegisterNode: err=%v success=%v msg=%s nodeId=%s",
				err, regNode.GetSuccess(), regNode.GetMessage(), regNode.GetNodeId())
		}

		discover, err := registry.DiscoverNodes(ctx, &pb.DiscoverNodesRequest{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
			HealthyOnly: true,
		})
		if err != nil || !discover.GetSuccess() || len(discover.GetNodes()) != 1 {
			t.Fatalf("DiscoverNodes after RegisterNode: err=%v success=%v nodes=%d",
				err, discover.GetSuccess(), len(discover.GetNodes()))
		}
		if discover.GetNodes()[0].GetNodeId() != regNode.GetNodeId() {
			t.Fatalf("nodeId mismatch: discover=%s register=%s",
				discover.GetNodes()[0].GetNodeId(), regNode.GetNodeId())
		}
		t.Logf("RegisterNode ok nodeId=%s", regNode.GetNodeId())
	})

	t.Run("unregister_service_whole", func(t *testing.T) {
		const serviceName = "e2e-unreg-svc-whole"
		var nodeIDs []string
		for port := int32(18101); port <= 18102; port++ {
			reg, err := registry.RegisterService(ctx, &pb.Service{
				NamespaceId: ns,
				GroupName:   group,
				ServiceName: serviceName,
				ServiceType: "INTERNAL",
				Node: &pb.Node{
					NamespaceId:    ns,
					GroupName:      group,
					ServiceName:    serviceName,
					IpAddress:      "127.0.0.1",
					PortNumber:     port,
					InstanceStatus: "UP",
					HealthyStatus:  "HEALTHY",
					Ephemeral:      "Y",
					Weight:         100,
				},
			})
			if err != nil || !reg.GetSuccess() {
				t.Fatalf("RegisterService: err=%v success=%v", err, reg.GetSuccess())
			}
			nodeIDs = append(nodeIDs, reg.GetNodeId())
		}

		unreg, err := registry.UnregisterService(ctx, &pb.ServiceKey{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
		})
		if err != nil || !unreg.GetSuccess() {
			t.Fatalf("UnregisterService(整服务): err=%v success=%v msg=%s",
				err, unreg.GetSuccess(), unreg.GetMessage())
		}

		discover, err := registry.DiscoverNodes(ctx, &pb.DiscoverNodesRequest{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
			HealthyOnly: false,
		})
		if err != nil {
			t.Fatalf("DiscoverNodes: %v", err)
		}
		if len(discover.GetNodes()) != 0 {
			t.Fatalf("整服务注销后仍有节点: count=%d ids=%v", len(discover.GetNodes()), nodeIDs)
		}
		t.Logf("UnregisterService whole ok removedNodes=%d", len(nodeIDs))
	})

	t.Run("unregister_service_with_node_id", func(t *testing.T) {
		const serviceName = "e2e-unreg-svc-node"
		var nodeIDs []string
		for port := int32(18111); port <= 18112; port++ {
			reg, err := registry.RegisterService(ctx, &pb.Service{
				NamespaceId: ns,
				GroupName:   group,
				ServiceName: serviceName,
				ServiceType: "INTERNAL",
				Node: &pb.Node{
					NamespaceId:    ns,
					GroupName:      group,
					ServiceName:    serviceName,
					IpAddress:      "127.0.0.1",
					PortNumber:     port,
					InstanceStatus: "UP",
					HealthyStatus:  "HEALTHY",
					Ephemeral:      "Y",
					Weight:         100,
				},
			})
			if err != nil || !reg.GetSuccess() {
				t.Fatalf("RegisterService: err=%v success=%v", err, reg.GetSuccess())
			}
			nodeIDs = append(nodeIDs, reg.GetNodeId())
		}

		unreg, err := registry.UnregisterService(ctx, &pb.ServiceKey{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
			NodeId:      nodeIDs[0],
		})
		if err != nil || !unreg.GetSuccess() {
			t.Fatalf("UnregisterService(带 nodeId): err=%v success=%v msg=%s",
				err, unreg.GetSuccess(), unreg.GetMessage())
		}

		discover, err := registry.DiscoverNodes(ctx, &pb.DiscoverNodesRequest{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
			HealthyOnly: false,
		})
		if err != nil || !discover.GetSuccess() {
			t.Fatalf("DiscoverNodes: err=%v success=%v", err, discover.GetSuccess())
		}
		if len(discover.GetNodes()) != 1 {
			t.Fatalf("期望剩余 1 个节点, got=%d", len(discover.GetNodes()))
		}
		if discover.GetNodes()[0].GetNodeId() != nodeIDs[1] {
			t.Fatalf("剩余节点应为 %s, got %s", nodeIDs[1], discover.GetNodes()[0].GetNodeId())
		}
		t.Logf("UnregisterService with nodeId ok kept=%s removed=%s", nodeIDs[1], nodeIDs[0])
	})

	t.Run("subscribe_services_push", func(t *testing.T) {
		const serviceName = "e2e-sub-services"
		subCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		stream, err := registry.SubscribeServices(subCtx, &pb.SubscribeServicesRequest{
			NamespaceId:  ns,
			GroupName:    group,
			ServiceNames: []string{serviceName},
		})
		if err != nil {
			t.Fatalf("SubscribeServices: %v", err)
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

		time.Sleep(500 * time.Millisecond)

		_, err = registry.RegisterService(ctx, &pb.Service{
			NamespaceId: ns,
			GroupName:   group,
			ServiceName: serviceName,
			ServiceType: "INTERNAL",
			Node: &pb.Node{
				NamespaceId:    ns,
				GroupName:      group,
				ServiceName:    serviceName,
				IpAddress:      "127.0.0.1",
				PortNumber:     18201,
				InstanceStatus: "UP",
				HealthyStatus:  "HEALTHY",
				Ephemeral:      "Y",
				Weight:         100,
			},
		})
		if err != nil {
			t.Fatalf("RegisterService for SubscribeServices: %v", err)
		}

		deadline := time.After(10 * time.Second)
		for {
			select {
			case ev := <-eventCh:
				if ev.GetServiceName() != serviceName {
					t.Logf("忽略非目标服务事件: %s", ev.GetServiceName())
					continue
				}
				// 订阅瞬间可能先推 SERVICE_NOT_FOUND 快照，等到注册后的 NODE_ADDED/更新事件
				if ev.GetEventType() == "SERVICE_NOT_FOUND" {
					t.Logf("忽略初始快照: type=%s", ev.GetEventType())
					continue
				}
				t.Logf("SubscribeServices push ok type=%s", ev.GetEventType())
				goto subServicesOK
			case recvErr := <-errCh:
				t.Fatalf("SubscribeServices 接收失败: %v", recvErr)
			case <-deadline:
				t.Fatal("超时未收到 SubscribeServices 推送")
			}
		}
	subServicesOK:
	})

	t.Run("config_list_history_rollback_delete", func(t *testing.T) {
		dataID := "e2e-config-lifecycle.yaml"
		v1 := "app:\n  version: v1\n"
		v2 := "app:\n  version: v2\n"

		save1, err := configClient.SaveConfig(ctx, &pb.ConfigData{
			NamespaceId:   ns,
			GroupName:     group,
			ConfigDataId:  dataID,
			ConfigContent: v1,
			ContentType:   "yaml",
			ConfigDesc:    "lifecycle v1",
			ChangedBy:     "e2e",
		})
		if err != nil || !save1.GetSuccess() {
			t.Fatalf("SaveConfig v1: err=%v success=%v msg=%s", err, save1.GetSuccess(), save1.GetMessage())
		}

		save2, err := configClient.SaveConfig(ctx, &pb.ConfigData{
			NamespaceId:   ns,
			GroupName:     group,
			ConfigDataId:  dataID,
			ConfigContent: v2,
			ContentType:   "yaml",
			ConfigDesc:    "lifecycle v2",
			ChangedBy:     "e2e",
		})
		if err != nil || !save2.GetSuccess() {
			t.Fatalf("SaveConfig v2: err=%v success=%v msg=%s", err, save2.GetSuccess(), save2.GetMessage())
		}

		list, err := configClient.ListConfigs(ctx, &pb.ListConfigsRequest{
			NamespaceId: ns,
			GroupName:   group,
		})
		if err != nil || !list.GetSuccess() {
			t.Fatalf("ListConfigs: err=%v success=%v msg=%s", err, list.GetSuccess(), list.GetMessage())
		}
		found := false
		for _, c := range list.GetConfigs() {
			if c.GetConfigDataId() == dataID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("ListConfigs 未包含 %s, count=%d", dataID, len(list.GetConfigs()))
		}

		hist, err := configClient.GetConfigHistory(ctx, &pb.GetConfigHistoryRequest{
			NamespaceId:  ns,
			GroupName:    group,
			ConfigDataId: dataID,
			Limit:        10,
		})
		if err != nil || !hist.GetSuccess() {
			t.Fatalf("GetConfigHistory: err=%v success=%v msg=%s", err, hist.GetSuccess(), hist.GetMessage())
		}
		if len(hist.GetHistory()) < 2 {
			t.Fatalf("期望至少 2 条历史, got=%d", len(hist.GetHistory()))
		}

		var targetVersion int64
		for _, h := range hist.GetHistory() {
			if h.GetConfigContent() == v1 {
				targetVersion = h.GetConfigVersion()
				break
			}
		}
		if targetVersion == 0 {
			// 历史按 newVersion 倒序，取较小版本作为 v1
			targetVersion = hist.GetHistory()[len(hist.GetHistory())-1].GetConfigVersion()
		}

		rb, err := configClient.RollbackConfig(ctx, &pb.RollbackConfigRequest{
			NamespaceId:   ns,
			GroupName:     group,
			ConfigDataId:  dataID,
			TargetVersion: targetVersion,
			ChangedBy:     "e2e",
			ChangeReason:  "e2e rollback",
		})
		if err != nil || !rb.GetSuccess() {
			t.Fatalf("RollbackConfig: err=%v success=%v msg=%s", err, rb.GetSuccess(), rb.GetMessage())
		}

		get, err := configClient.GetConfig(ctx, &pb.ConfigKey{
			NamespaceId:  ns,
			GroupName:    group,
			ConfigDataId: dataID,
		})
		if err != nil || !get.GetSuccess() || get.GetConfig() == nil {
			t.Fatalf("GetConfig after rollback: err=%v success=%v", err, get.GetSuccess())
		}
		if get.GetConfig().GetConfigContent() != v1 {
			t.Fatalf("回滚后内容应为 v1, got=%q", get.GetConfig().GetConfigContent())
		}

		del, err := configClient.DeleteConfig(ctx, &pb.ConfigKey{
			NamespaceId:  ns,
			GroupName:    group,
			ConfigDataId: dataID,
		})
		if err != nil || !del.GetSuccess() {
			t.Fatalf("DeleteConfig: err=%v success=%v msg=%s", err, del.GetSuccess(), del.GetMessage())
		}

		listAfter, err := configClient.ListConfigs(ctx, &pb.ListConfigsRequest{
			NamespaceId: ns,
			GroupName:   group,
		})
		if err != nil || !listAfter.GetSuccess() {
			t.Fatalf("ListConfigs after delete: err=%v success=%v", err, listAfter.GetSuccess())
		}
		for _, c := range listAfter.GetConfigs() {
			if c.GetConfigDataId() == dataID {
				t.Fatalf("删除后 ListConfigs 仍包含 %s", dataID)
			}
		}
		t.Logf("config lifecycle ok history=%d rollbackTo=%d listBefore=%d",
			len(hist.GetHistory()), targetVersion, len(list.GetConfigs()))
	})
}
