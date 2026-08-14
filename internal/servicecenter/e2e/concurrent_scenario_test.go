package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/testutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ScenarioReport 并发业务场景测试报告（写入 JSON，供评审展示）
type ScenarioReport struct {
	Title       string            `json:"title"`
	StartedAt   string            `json:"startedAt"`
	FinishedAt  string            `json:"finishedAt"`
	Endpoint    string            `json:"endpoint"`
	Namespace   string            `json:"namespace"`
	Group       string            `json:"group"`
	Scenarios   []ScenarioResult  `json:"scenarios"`
	Summary     ReportSummary     `json:"summary"`
	Environment map[string]string `json:"environment"`
}

// ScenarioResult 单场景结果
type ScenarioResult struct {
	Name       string  `json:"name"`
	Passed     bool    `json:"passed"`
	DurationMs int64   `json:"durationMs"`
	Metrics    Metrics `json:"metrics"`
	Error      string  `json:"error,omitempty"`
	Notes      string  `json:"notes,omitempty"`
}

// Metrics 关键指标
type Metrics struct {
	Workers         int     `json:"workers"`
	Services        int     `json:"services"`
	NodesPerService int     `json:"nodesPerService"`
	SuccessOps      int64   `json:"successOps"`
	FailedOps       int64   `json:"failedOps"`
	DiscoverHits    int64   `json:"discoverHits"`
	PushEvents      int64   `json:"pushEvents"`
	ConfigWatches   int     `json:"configWatches"`
	ConfigUpdates   int64   `json:"configUpdates"`
	OpsPerSecond    float64 `json:"opsPerSecond"`
}

// ReportSummary 汇总
type ReportSummary struct {
	TotalScenarios  int    `json:"totalScenarios"`
	Passed          int    `json:"passed"`
	Failed          int    `json:"failed"`
	TotalDurationMs int64  `json:"totalDurationMs"`
	Verdict         string `json:"verdict"`
}

// TestConcurrentBusinessScenarios 多服务 / 多线程注册发现 / 配置订阅真实压测场景。
func TestConcurrentBusinessScenarios(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.Start(ctx, testutil.Options{})
	if err != nil {
		t.Fatalf("start sqlite service center: %v", err)
	}
	t.Cleanup(func() { _ = env.Stop(context.Background()) })

	report := &ScenarioReport{
		Title:     "Service Center Concurrent Business Scenario Report",
		StartedAt: time.Now().Format(time.RFC3339),
		Endpoint:  env.Endpoint(),
		Namespace: env.NamespaceID,
		Group:     env.GroupName,
		Scenarios: make([]ScenarioResult, 0, 4),
		Environment: map[string]string{
			"storage": "sqlite-tempfile",
			"auth":    "disabled",
			"tls":     "disabled",
			"client":  "go-grpc",
		},
	}
	startAll := time.Now()

	run := func(name string, fn func(t *testing.T) ScenarioResult) {
		t.Run(name, func(t *testing.T) {
			res := fn(t)
			res.Name = name
			report.Scenarios = append(report.Scenarios, res)
			if !res.Passed {
				t.Errorf("%s failed: %s", name, res.Error)
			}
		})
	}

	run("multi_service_register_discover", func(t *testing.T) ScenarioResult {
		return scenarioMultiServiceRegisterDiscover(t, env)
	})
	run("multi_thread_register_discover", func(t *testing.T) ScenarioResult {
		return scenarioMultiThreadRegisterDiscover(t, env)
	})
	run("namespace_subscribe_under_churn", func(t *testing.T) ScenarioResult {
		return scenarioNamespaceSubscribeUnderChurn(t, env)
	})
	run("multi_config_watch_publish", func(t *testing.T) ScenarioResult {
		return scenarioMultiConfigWatchPublish(t, env)
	})

	report.FinishedAt = time.Now().Format(time.RFC3339)
	report.Summary.TotalDurationMs = time.Since(startAll).Milliseconds()
	report.Summary.TotalScenarios = len(report.Scenarios)
	for _, s := range report.Scenarios {
		if s.Passed {
			report.Summary.Passed++
		} else {
			report.Summary.Failed++
		}
	}
	if report.Summary.Failed == 0 {
		report.Summary.Verdict = "PASS"
	} else {
		report.Summary.Verdict = "FAIL"
	}

	outPath := writeReport(t, report)
	t.Logf("scenario report written: %s", outPath)
	t.Logf("VERDICT=%s passed=%d/%d durationMs=%d",
		report.Summary.Verdict, report.Summary.Passed, report.Summary.TotalScenarios, report.Summary.TotalDurationMs)
}

func scenarioMultiServiceRegisterDiscover(t *testing.T, env *testutil.Env) ScenarioResult {
	const serviceCount = 50
	const nodesPerService = 3
	start := time.Now()
	res := ScenarioResult{Metrics: Metrics{Services: serviceCount, NodesPerService: nodesPerService, Workers: 1}}

	conn := mustDial(t, env.Endpoint())
	defer conn.Close()
	registry := pb.NewServiceRegistryClient(conn)
	ctx := context.Background()

	var success, failed int64
	for i := 0; i < serviceCount; i++ {
		svc := fmt.Sprintf("msvc-%03d", i)
		for n := 0; n < nodesPerService; n++ {
			_, err := registry.RegisterService(ctx, &pb.Service{
				NamespaceId: env.NamespaceID,
				GroupName:   env.GroupName,
				ServiceName: svc,
				ServiceType: "INTERNAL",
				Node: &pb.Node{
					NamespaceId:    env.NamespaceID,
					GroupName:      env.GroupName,
					ServiceName:    svc,
					IpAddress:      fmt.Sprintf("10.0.%d.%d", i%250+1, n+1),
					PortNumber:     int32(20000 + i*10 + n),
					InstanceStatus: "UP",
					HealthyStatus:  "HEALTHY",
					Ephemeral:      "Y",
					Weight:         100,
				},
			})
			if err != nil {
				atomic.AddInt64(&failed, 1)
			} else {
				atomic.AddInt64(&success, 1)
			}
		}
	}

	var discoverHits int64
	for i := 0; i < serviceCount; i++ {
		svc := fmt.Sprintf("msvc-%03d", i)
		resp, err := registry.DiscoverNodes(ctx, &pb.DiscoverNodesRequest{
			NamespaceId: env.NamespaceID,
			GroupName:   env.GroupName,
			ServiceName: svc,
			HealthyOnly: true,
		})
		if err == nil && resp.GetSuccess() && len(resp.GetNodes()) == nodesPerService {
			discoverHits++
		} else {
			failed++
		}
	}

	elapsed := time.Since(start)
	res.DurationMs = elapsed.Milliseconds()
	res.Metrics.SuccessOps = success
	res.Metrics.FailedOps = failed
	res.Metrics.DiscoverHits = discoverHits
	res.Metrics.OpsPerSecond = float64(success) / elapsed.Seconds()
	res.Passed = failed == 0 && discoverHits == serviceCount
	res.Notes = fmt.Sprintf("registered %d nodes across %d services; discover matched %d/%d",
		success, serviceCount, discoverHits, serviceCount)
	if !res.Passed {
		res.Error = fmt.Sprintf("failedOps=%d discoverHits=%d/%d", failed, discoverHits, serviceCount)
	}
	return res
}

func scenarioMultiThreadRegisterDiscover(t *testing.T, env *testutil.Env) ScenarioResult {
	const workers = 20
	const servicesPerWorker = 5
	const nodesPerService = 2
	start := time.Now()
	res := ScenarioResult{Metrics: Metrics{
		Workers:         workers,
		Services:        workers * servicesPerWorker,
		NodesPerService: nodesPerService,
	}}

	var success, failed, discoverHits int64
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		workerID := w
		go func() {
			defer wg.Done()
			conn, err := grpc.NewClient(env.Endpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				atomic.AddInt64(&failed, 1)
				return
			}
			defer conn.Close()
			registry := pb.NewServiceRegistryClient(conn)
			ctx := context.Background()

			for s := 0; s < servicesPerWorker; s++ {
				svc := fmt.Sprintf("thr-%02d-svc-%02d", workerID, s)
				for n := 0; n < nodesPerService; n++ {
					_, err := registry.RegisterService(ctx, &pb.Service{
						NamespaceId: env.NamespaceID,
						GroupName:   env.GroupName,
						ServiceName: svc,
						ServiceType: "INTERNAL",
						Node: &pb.Node{
							NamespaceId:    env.NamespaceID,
							GroupName:      env.GroupName,
							ServiceName:    svc,
							IpAddress:      fmt.Sprintf("11.%d.%d.%d", workerID+1, s+1, n+1),
							PortNumber:     int32(30000 + workerID*100 + s*10 + n),
							InstanceStatus: "UP",
							HealthyStatus:  "HEALTHY",
							Ephemeral:      "Y",
							Weight:         100,
						},
					})
					if err != nil {
						atomic.AddInt64(&failed, 1)
					} else {
						atomic.AddInt64(&success, 1)
					}
				}
				resp, err := registry.DiscoverNodes(ctx, &pb.DiscoverNodesRequest{
					NamespaceId: env.NamespaceID,
					GroupName:   env.GroupName,
					ServiceName: svc,
					HealthyOnly: true,
				})
				if err == nil && resp.GetSuccess() && len(resp.GetNodes()) >= nodesPerService {
					atomic.AddInt64(&discoverHits, 1)
				} else {
					atomic.AddInt64(&failed, 1)
				}
			}
		}()
	}
	wg.Wait()

	elapsed := time.Since(start)
	res.DurationMs = elapsed.Milliseconds()
	res.Metrics.SuccessOps = success
	res.Metrics.FailedOps = failed
	res.Metrics.DiscoverHits = discoverHits
	res.Metrics.OpsPerSecond = float64(success) / elapsed.Seconds()
	expectedDiscover := int64(workers * servicesPerWorker)
	res.Passed = failed == 0 && discoverHits == expectedDiscover
	res.Notes = fmt.Sprintf("%d workers x %d services x %d nodes; discover %d/%d",
		workers, servicesPerWorker, nodesPerService, discoverHits, expectedDiscover)
	if !res.Passed {
		res.Error = fmt.Sprintf("failedOps=%d discoverHits=%d/%d", failed, discoverHits, expectedDiscover)
	}
	return res
}

func scenarioNamespaceSubscribeUnderChurn(t *testing.T, env *testutil.Env) ScenarioResult {
	const churnServices = 30
	start := time.Now()
	res := ScenarioResult{Metrics: Metrics{Services: churnServices, Workers: 1}}

	conn := mustDial(t, env.Endpoint())
	defer conn.Close()
	registry := pb.NewServiceRegistryClient(conn)

	subCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	stream, err := registry.SubscribeNamespace(subCtx, &pb.SubscribeNamespaceRequest{
		NamespaceId: env.NamespaceID,
		GroupName:   env.GroupName,
	})
	if err != nil {
		res.Error = err.Error()
		return res
	}

	var pushEvents int64
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, recvErr := stream.Recv()
			if recvErr != nil {
				return
			}
			atomic.AddInt64(&pushEvents, 1)
		}
	}()
	time.Sleep(400 * time.Millisecond)

	var success, failed int64
	ctx := context.Background()
	for i := 0; i < churnServices; i++ {
		svc := fmt.Sprintf("churn-%03d", i)
		_, err := registry.RegisterService(ctx, &pb.Service{
			NamespaceId: env.NamespaceID,
			GroupName:   env.GroupName,
			ServiceName: svc,
			ServiceType: "INTERNAL",
			Node: &pb.Node{
				NamespaceId:    env.NamespaceID,
				GroupName:      env.GroupName,
				ServiceName:    svc,
				IpAddress:      "127.0.0.1",
				PortNumber:     int32(40000 + i),
				InstanceStatus: "UP",
				HealthyStatus:  "HEALTHY",
				Ephemeral:      "Y",
				Weight:         100,
			},
		})
		if err != nil {
			failed++
		} else {
			success++
		}
	}

	// 等待推送收敛
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) && atomic.LoadInt64(&pushEvents) < churnServices {
		time.Sleep(50 * time.Millisecond)
	}
	cancel()
	<-done

	elapsed := time.Since(start)
	res.DurationMs = elapsed.Milliseconds()
	res.Metrics.SuccessOps = success
	res.Metrics.FailedOps = failed
	res.Metrics.PushEvents = atomic.LoadInt64(&pushEvents)
	res.Metrics.OpsPerSecond = float64(success) / elapsed.Seconds()
	// 允许少量推送丢失（channel 非阻塞丢弃），要求至少收到 80%
	minPush := int64(float64(churnServices) * 0.8)
	res.Passed = failed == 0 && res.Metrics.PushEvents >= minPush
	res.Notes = fmt.Sprintf("registered %d churn services; push events=%d (min=%d)",
		success, res.Metrics.PushEvents, minPush)
	if !res.Passed {
		res.Error = fmt.Sprintf("pushEvents=%d min=%d failed=%d", res.Metrics.PushEvents, minPush, failed)
	}
	return res
}

func scenarioMultiConfigWatchPublish(t *testing.T, env *testutil.Env) ScenarioResult {
	const watchers = 10
	const configs = 20
	start := time.Now()
	res := ScenarioResult{Metrics: Metrics{Workers: watchers, ConfigWatches: watchers, Services: configs}}

	var updates atomic.Int64
	var failed atomic.Int64
	var wg sync.WaitGroup

	// 每个 watcher 监听全部 configs（通过多个 dataId）
	dataIDs := make([]string, configs)
	for i := 0; i < configs; i++ {
		dataIDs[i] = fmt.Sprintf("cfg-%03d.yaml", i)
	}

	watchCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	wg.Add(watchers)
	for w := 0; w < watchers; w++ {
		go func() {
			defer wg.Done()
			conn, err := grpc.NewClient(env.Endpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				failed.Add(1)
				return
			}
			defer conn.Close()
			cfgClient := pb.NewConfigCenterClient(conn)
			stream, err := cfgClient.WatchConfig(watchCtx, &pb.WatchConfigRequest{
				NamespaceId:   env.NamespaceID,
				GroupName:     env.GroupName,
				ConfigDataIds: dataIDs,
			})
			if err != nil {
				failed.Add(1)
				return
			}
			for {
				ev, recvErr := stream.Recv()
				if recvErr != nil {
					if recvErr == io.EOF || watchCtx.Err() != nil {
						return
					}
					failed.Add(1)
					return
				}
				// 忽略初始 CONFIG_DELETED，统计带内容的更新
				if ev.GetConfig() != nil && ev.GetConfig().GetConfigContent() != "" {
					updates.Add(1)
				}
			}
		}()
	}

	time.Sleep(600 * time.Millisecond)

	pubConn := mustDial(t, env.Endpoint())
	defer pubConn.Close()
	publisher := pb.NewConfigCenterClient(pubConn)
	ctx := context.Background()
	var publishOK int64
	for i := 0; i < configs; i++ {
		resp, err := publisher.SaveConfig(ctx, &pb.ConfigData{
			NamespaceId:   env.NamespaceID,
			GroupName:     env.GroupName,
			ConfigDataId:  dataIDs[i],
			ConfigContent: fmt.Sprintf("key: value-%d\n", i),
			ContentType:   "yaml",
			ChangedBy:     "e2e-load",
		})
		if err != nil || resp == nil || !resp.GetSuccess() {
			failed.Add(1)
		} else {
			publishOK++
		}
	}

	// 每个 watcher 对每个 config 至少应收到 1 次带内容更新 => watchers * configs
	expectedMin := int64(watchers * configs)
	deadline := time.Now().Add(12 * time.Second)
	for time.Now().Before(deadline) && updates.Load() < expectedMin {
		time.Sleep(50 * time.Millisecond)
	}
	cancel()
	wg.Wait()

	elapsed := time.Since(start)
	res.DurationMs = elapsed.Milliseconds()
	res.Metrics.SuccessOps = publishOK
	res.Metrics.FailedOps = failed.Load()
	res.Metrics.ConfigUpdates = updates.Load()
	res.Metrics.OpsPerSecond = float64(publishOK) / elapsed.Seconds()
	// 允许少量丢失，要求 >= 90%
	minUpdates := int64(float64(expectedMin) * 0.9)
	res.Passed = failed.Load() == 0 && publishOK == configs && updates.Load() >= minUpdates
	res.Notes = fmt.Sprintf("%d watchers x %d configs; publishOK=%d updateEvents=%d (min=%d)",
		watchers, configs, publishOK, updates.Load(), minUpdates)
	if !res.Passed {
		res.Error = fmt.Sprintf("updates=%d min=%d failed=%d publishOK=%d",
			updates.Load(), minUpdates, failed.Load(), publishOK)
	}
	return res
}

func mustDial(t *testing.T, endpoint string) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial %s: %v", endpoint, err)
	}
	return conn
}

func writeReport(t *testing.T, report *ScenarioReport) string {
	t.Helper()
	dir := filepath.Join("testdata", "reports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir report dir: %v", err)
	}
	name := fmt.Sprintf("concurrent-scenario-%s.json", time.Now().Format("20060102-150405"))
	path := filepath.Join(dir, name)
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	// 同时写一份固定名，方便画布/脚本读取最新结果
	latest := filepath.Join(dir, "latest-concurrent-scenario.json")
	_ = os.WriteFile(latest, raw, 0o644)
	return path
}
