package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"

	"gateway/internal/gateway/config"
)

func TestRepeatedReloadReleasesAllGenerations(t *testing.T) {
	baselineGoroutines := runtime.NumGoroutine()
	cfg := config.DefaultGatewayConfig
	cfg.InstanceID = "generation-resource-stability-test"
	cfg.Base.Listen = "127.0.0.1:0"
	cfg.Base.GracefulShutdownTimeout = 200 * time.Millisecond

	gateway, err := NewGatewayFactory().CreateGateway(&cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := gateway.Start(); err != nil {
		t.Fatal(err)
	}

	const reloadCount = 30
	retired := make([]*gatewayGeneration, 0, reloadCount)
	for i := 0; i < reloadCount; i++ {
		retired = append(retired, gateway.currentGeneration.Load())
		next := *gateway.GetConfig()
		next.Base.IdleTimeout = time.Duration(100+i) * time.Millisecond
		if err := gateway.Reload(&next); err != nil {
			_ = gateway.Stop()
			t.Fatalf("reload %d failed: %v", i, err)
		}
	}
	gateway.generationWG.Wait()

	for i, generation := range retired {
		select {
		case <-generation.serveDone:
		default:
			_ = gateway.Stop()
			t.Fatalf("retired generation %d still has a running HTTP server", i)
		}
		select {
		case <-generation.listener.done:
		default:
			_ = gateway.Stop()
			t.Fatalf("retired generation %d still accepts connections", i)
		}
	}

	dispatcher := gateway.dispatcher
	if err := gateway.Stop(); err != nil {
		t.Fatal(err)
	}
	if gateway.currentGeneration.Load() != nil || gateway.dispatcher != nil {
		t.Fatal("Stop() retained generation or listener resources")
	}
	if dispatcher.activeConnections.Load() != 0 || gateway.requestLimiter.activeCount() != 0 {
		t.Fatalf("capacity counters did not return to zero: connections=%d requests=%d",
			dispatcher.activeConnections.Load(), gateway.requestLimiter.activeCount())
	}

	// 日志写入器会异步刷新旧实例，给这些有界后台任务留出收敛时间。
	deadline := time.Now().Add(3 * time.Second)
	for {
		runtime.GC()
		if runtime.NumGoroutine() <= baselineGoroutines+12 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("goroutines did not converge after repeated reloads: baseline=%d current=%d",
				baselineGoroutines, runtime.NumGoroutine())
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestConcurrentKeepAliveTrafficDuringRepeatedReload(t *testing.T) {
	base, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	dispatcher := newListenerDispatcher(base)
	gateway := &Gateway{dispatcher: dispatcher}
	initial := newTestGeneration("traffic-generation-0", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	initial.listener = dispatcher.newGenerationListener()
	gateway.startGenerationServer(initial)
	if err := gateway.waitGenerationReady(initial); err != nil {
		t.Fatal(err)
	}
	dispatcher.switchTo(initial.listener)
	gateway.currentGeneration.Store(initial)
	dispatcher.start()

	transport := &http.Transport{
		MaxIdleConns:        64,
		MaxIdleConnsPerHost: 64,
	}
	client := &http.Client{Transport: transport}
	trafficCtx, cancelTraffic := context.WithCancel(context.Background())
	requestErrors := make(chan error, 1)
	var requestWG sync.WaitGroup
	for i := 0; i < 8; i++ {
		requestWG.Add(1)
		go func() {
			defer requestWG.Done()
			runReloadTraffic(trafficCtx, client, "http://"+base.Addr().String(), requestErrors)
		}()
	}

	for i := 1; i <= 20; i++ {
		next := newTestGeneration("traffic-generation", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		if err := gateway.activateGeneration(next); err != nil {
			cancelTraffic()
			requestWG.Wait()
			t.Fatal(err)
		}
	}
	gateway.generationWG.Wait()
	cancelTraffic()
	requestWG.Wait()
	transport.CloseIdleConnections()

	select {
	case err := <-requestErrors:
		t.Fatalf("traffic failed during reload: %v", err)
	default:
	}

	if err := dispatcher.Close(); err != nil && err != net.ErrClosed {
		t.Fatal(err)
	}
	if current := gateway.currentGeneration.Load(); current != nil {
		if err := gateway.drainGeneration(current); err != nil {
			t.Fatal(err)
		}
	}
	gateway.wg.Wait()
}

func runReloadTraffic(ctx context.Context, client *http.Client, address string, requestErrors chan<- error) {
	consecutiveErrors := 0
	for ctx.Err() == nil {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
		if err != nil {
			reportTrafficError(requestErrors, err)
			return
		}
		response, err := client.Do(request)
		if err != nil {
			if ctx.Err() == nil {
				consecutiveErrors++
				if consecutiveErrors < 3 {
					continue
				}
				reportTrafficError(requestErrors, err)
			}
			return
		}
		consecutiveErrors = 0
		_ = response.Body.Close()
		if response.StatusCode != http.StatusNoContent {
			reportTrafficError(requestErrors, fmt.Errorf("unexpected HTTP status: %d", response.StatusCode))
			return
		}
	}
}

func reportTrafficError(requestErrors chan<- error, err error) {
	select {
	case requestErrors <- err:
	default:
	}
}

func BenchmarkGenerationAcquireRelease(b *testing.B) {
	generation := &gatewayGeneration{}
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if !generation.acquire() {
				b.Error("active generation rejected a request")
				return
			}
			generation.release()
		}
	})
}

func BenchmarkRequestAdmissionLimiter(b *testing.B) {
	var limiter requestAdmissionLimiter
	limiter.setLimit(0)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if !limiter.tryAcquire() {
				b.Error("unlimited request admission rejected a request")
				return
			}
			limiter.release()
		}
	})
}

func BenchmarkListenerDispatcherKeepAlive(b *testing.B) {
	benchmarkListenerDispatcher(b)
}

func BenchmarkGatewayReload(b *testing.B) {
	cfg := config.DefaultGatewayConfig
	cfg.InstanceID = "generation-reload-benchmark"
	cfg.Base.Listen = "127.0.0.1:0"
	cfg.Base.GracefulShutdownTimeout = time.Second
	gateway, err := NewGatewayFactory().CreateGateway(&cfg, "")
	if err != nil {
		b.Fatal(err)
	}
	if err := gateway.Start(); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		_ = gateway.Stop()
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		next := *gateway.GetConfig()
		next.Base.IdleTimeout = time.Duration(100+i%100) * time.Millisecond
		if err := gateway.Reload(&next); err != nil {
			b.Fatal(err)
		}
		gateway.generationWG.Wait()
	}
}

func BenchmarkVirtualListenerDelivery(b *testing.B) {
	listener := &virtualListener{
		addr:    &net.TCPAddr{},
		conns:   make(chan net.Conn),
		done:    make(chan struct{}),
		readyCh: make(chan struct{}),
	}
	dispatcherDone := make(chan struct{})
	acceptDone := make(chan struct{})
	go func() {
		defer close(acceptDone)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	b.Cleanup(func() {
		close(dispatcherDone)
		_ = listener.Close()
		<-acceptDone
	})

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client, server := net.Pipe()
			if !listener.deliver(server, dispatcherDone) {
				b.Error("virtual listener rejected a connection")
				_ = client.Close()
				return
			}
			_ = client.Close()
		}
	})
}

func benchmarkListenerDispatcher(b *testing.B) {
	base, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	dispatcher := newListenerDispatcher(base)
	gateway := &Gateway{dispatcher: dispatcher}
	generation := newTestGeneration("generation-dispatch-benchmark", func(w http.ResponseWriter, _ *http.Request) {
		current := gateway.currentGeneration.Load()
		if !generationAcquire(current) {
			http.Error(w, "generation draining", http.StatusServiceUnavailable)
			return
		}
		current.release()
		w.WriteHeader(http.StatusNoContent)
	})
	generation.listener = dispatcher.newGenerationListener()
	gateway.currentGeneration.Store(generation)
	gateway.startGenerationServer(generation)
	if err := gateway.waitGenerationReady(generation); err != nil {
		b.Fatal(err)
	}
	dispatcher.switchTo(generation.listener)
	dispatcher.start()

	transport := &http.Transport{
		MaxIdleConns:        256,
		MaxIdleConnsPerHost: 256,
	}
	client := &http.Client{Transport: transport}
	address := "http://" + base.Addr().String()

	b.Cleanup(func() {
		transport.CloseIdleConnections()
		_ = dispatcher.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = generation.server.Shutdown(ctx)
		gateway.wg.Wait()
	})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, requestErr := client.Get(address)
			if requestErr != nil {
				b.Error(requestErr)
				return
			}
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusNoContent {
				b.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
				return
			}
		}
	})
}

func generationAcquire(generation *gatewayGeneration) bool {
	return generation != nil && generation.acquire()
}
