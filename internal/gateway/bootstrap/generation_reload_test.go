package bootstrap

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gateway/internal/gateway/config"
)

func TestGenerationSwitchKeepsInflightRequestOnOldGeneration(t *testing.T) {
	base, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	dispatcher := newListenerDispatcher(base)
	dispatcher.start()

	gateway := &Gateway{dispatcher: dispatcher}
	oldStarted := make(chan struct{})
	releaseOld := make(chan struct{})
	old := newTestGeneration("old", func(w http.ResponseWriter, _ *http.Request) {
		close(oldStarted)
		<-releaseOld
		_, _ = w.Write([]byte("old"))
	})
	old.listener = dispatcher.newGenerationListener()
	gateway.startGenerationServer(old)
	dispatcher.switchTo(old.listener)
	gateway.currentGeneration.Store(old)

	oldResult := make(chan string, 1)
	go func() {
		oldResult <- requestBody(t, base.Addr().String())
	}()
	<-oldStarted

	next := newTestGeneration("new", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("new"))
	})
	if err := gateway.activateGeneration(next); err != nil {
		t.Fatal(err)
	}

	if got := requestBody(t, base.Addr().String()); got != "new" {
		t.Fatalf("new connection response = %q, want new", got)
	}
	close(releaseOld)
	if got := <-oldResult; got != "old" {
		t.Fatalf("in-flight response = %q, want old", got)
	}
	gateway.generationWG.Wait()

	if err := dispatcher.Close(); err != nil && err != net.ErrClosed {
		t.Fatal(err)
	}
	if err := gateway.drainGeneration(next); err != nil {
		t.Fatal(err)
	}
	gateway.wg.Wait()
}

func TestCreateGenerationServerKeepsTimeoutsImmutable(t *testing.T) {
	cfg := &config.GatewayConfig{
		Base: config.BaseConfig{
			Listen:           "127.0.0.1:0",
			ReadTimeout:      2 * time.Second,
			WriteTimeout:     3 * time.Second,
			IdleTimeout:      4 * time.Second,
			MaxHeaderBytes:   4096,
			KeepAliveEnabled: true,
		},
	}
	gateway := &Gateway{}
	generation := &gatewayGeneration{config: cfg}
	server, err := NewGatewayFactory().createGenerationServer(gateway, generation, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if server.ReadTimeout != 2*time.Second ||
		server.WriteTimeout != 3*time.Second ||
		server.IdleTimeout != 4*time.Second ||
		server.MaxHeaderBytes != 4096 {
		t.Fatalf("unexpected server timeouts: read=%s write=%s idle=%s maxHeader=%d",
			server.ReadTimeout, server.WriteTimeout, server.IdleTimeout, server.MaxHeaderBytes)
	}
}

func TestReloadValidationKeepsCurrentGeneration(t *testing.T) {
	oldConfig := &config.GatewayConfig{Base: config.BaseConfig{Listen: "127.0.0.1:8080"}}
	old := &gatewayGeneration{config: oldConfig}
	gateway := &Gateway{gatewayConfig: oldConfig}
	gateway.currentGeneration.Store(old)

	err := NewGatewayFactory().ReloadGateway(gateway, &config.GatewayConfig{
		Base: config.BaseConfig{Listen: "127.0.0.1:9090"},
	})
	if err == nil {
		t.Fatal("ReloadGateway() error = nil, want listen change rejection")
	}
	if gateway.currentGeneration.Load() != old {
		t.Fatal("failed reload changed current generation")
	}
}

func TestCreateGenerationServerUsesIndependentTLSConfigs(t *testing.T) {
	certOne, keyOne := writeTestCertificate(t, "generation-one")
	certTwo, keyTwo := writeTestCertificate(t, "generation-two")
	factory := NewGatewayFactory()
	gateway := &Gateway{}

	createServer := func(certFile, keyFile string) *http.Server {
		cfg := &config.GatewayConfig{Base: config.BaseConfig{
			Listen:      "127.0.0.1:0",
			EnableHTTPS: true,
			CertFile:    certFile,
			KeyFile:     keyFile,
		}}
		generation := &gatewayGeneration{config: cfg}
		server, err := factory.createGenerationServer(gateway, generation, cfg)
		if err != nil {
			t.Fatal(err)
		}
		return server
	}

	serverOne := createServer(certOne, keyOne)
	serverTwo := createServer(certTwo, keyTwo)
	firstOne := serverOne.TLSConfig.Certificates[0].Certificate[0]
	firstTwo := serverTwo.TLSConfig.Certificates[0].Certificate[0]
	if string(firstOne) == string(firstTwo) {
		t.Fatal("TLS generations unexpectedly share the same certificate")
	}
}

func TestGatewayReloadPublishesNewServerTimeoutWithoutRebinding(t *testing.T) {
	cfg := config.DefaultGatewayConfig
	cfg.InstanceID = "generation-reload-test"
	cfg.Base.Listen = "127.0.0.1:0"
	cfg.Base.GracefulShutdownTimeout = time.Second
	gateway, err := NewGatewayFactory().CreateGateway(&cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := gateway.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = gateway.Stop()
	})

	oldGeneration := gateway.currentGeneration.Load()
	oldAddress := gateway.dispatcher.base.Addr().String()
	nextConfig := cfg
	nextConfig.Base.ReadTimeout = 175 * time.Millisecond
	if err := gateway.Reload(&nextConfig); err != nil {
		t.Fatal(err)
	}

	nextGeneration := gateway.currentGeneration.Load()
	if nextGeneration == oldGeneration {
		t.Fatal("Reload() did not publish a new generation")
	}
	if nextGeneration.server.ReadTimeout != 175*time.Millisecond {
		t.Fatalf("new ReadTimeout = %s, want 175ms", nextGeneration.server.ReadTimeout)
	}
	if gateway.dispatcher.base.Addr().String() != oldAddress {
		t.Fatal("Reload() rebound the operating-system listener")
	}
	for _, idleTimeout := range []time.Duration{225 * time.Millisecond, 325 * time.Millisecond} {
		currentConfig := *gateway.GetConfig()
		currentConfig.Base.IdleTimeout = idleTimeout
		if err := gateway.Reload(&currentConfig); err != nil {
			t.Fatal(err)
		}
		if gateway.currentGeneration.Load().server.IdleTimeout != idleTimeout {
			t.Fatalf("consecutive reload IdleTimeout = %s, want %s",
				gateway.currentGeneration.Load().server.IdleTimeout, idleTimeout)
		}
	}
	gateway.generationWG.Wait()
}

func TestGatewayReloadAndStopAreSerialized(t *testing.T) {
	cfg := config.DefaultGatewayConfig
	cfg.InstanceID = "generation-concurrency-test"
	cfg.Base.Listen = "127.0.0.1:0"
	cfg.Base.GracefulShutdownTimeout = 200 * time.Millisecond
	gateway, err := NewGatewayFactory().CreateGateway(&cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := gateway.Start(); err != nil {
		t.Fatal(err)
	}

	nextConfig := cfg
	nextConfig.Base.IdleTimeout = 250 * time.Millisecond
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	var reloadErr, stopErr error
	go func() {
		defer wg.Done()
		<-start
		reloadErr = gateway.Reload(&nextConfig)
	}()
	go func() {
		defer wg.Done()
		<-start
		stopErr = gateway.Stop()
	}()
	close(start)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("concurrent Reload and Stop did not finish")
	}
	if stopErr != nil {
		t.Fatalf("Stop() error = %v", stopErr)
	}
	if reloadErr != nil && !strings.Contains(reloadErr.Error(), "网关未运行") {
		t.Fatalf("Reload() unexpected error = %v", reloadErr)
	}
}

func TestGatewayCanStartAfterStopWithFreshGeneration(t *testing.T) {
	cfg := config.DefaultGatewayConfig
	cfg.InstanceID = "generation-restart-test"
	cfg.Base.Listen = "127.0.0.1:0"
	cfg.Base.GracefulShutdownTimeout = time.Second
	gateway, err := NewGatewayFactory().CreateGateway(&cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := gateway.Start(); err != nil {
		t.Fatal(err)
	}
	first := gateway.currentGeneration.Load()
	if err := gateway.Stop(); err != nil {
		t.Fatal(err)
	}
	if gateway.currentGeneration.Load() != nil {
		t.Fatal("Stop() retained a closed generation")
	}

	if err := gateway.Start(); err != nil {
		t.Fatal(err)
	}
	second := gateway.currentGeneration.Load()
	if second == nil || second == first {
		t.Fatal("Start() reused the stopped generation")
	}
	if err := gateway.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestReloadRejectsProtocolSwitchBeforePublishing(t *testing.T) {
	cfg := config.DefaultGatewayConfig
	cfg.InstanceID = "generation-protocol-test"
	cfg.Base.Listen = "127.0.0.1:0"
	gateway, err := NewGatewayFactory().CreateGateway(&cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := gateway.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = gateway.Stop()
	})

	current := gateway.currentGeneration.Load()
	next := cfg
	next.Base.EnableHTTPS = !cfg.Base.EnableHTTPS
	if err := gateway.Reload(&next); err == nil {
		t.Fatal("Reload() allowed an HTTP/HTTPS protocol switch")
	}
	if gateway.currentGeneration.Load() != current {
		t.Fatal("failed protocol reload changed the current generation")
	}
}

func TestReloadLogFailureKeepsCurrentGeneration(t *testing.T) {
	cfg := config.DefaultGatewayConfig
	cfg.InstanceID = "generation-log-rollback-test"
	cfg.Base.Listen = "127.0.0.1:0"
	gateway, err := NewGatewayFactory().CreateGateway(&cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := gateway.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = gateway.Stop()
	})

	current := gateway.currentGeneration.Load()
	next := cfg
	next.Log.OutputTargets = "INVALID"
	if err := gateway.Reload(&next); err == nil {
		t.Fatal("Reload() accepted an invalid log configuration")
	}
	if gateway.currentGeneration.Load() != current {
		t.Fatal("log update failure changed the current generation")
	}
}

func TestReloadUpdatesCapacityLimitsWithoutClosingActiveResources(t *testing.T) {
	cfg := config.DefaultGatewayConfig
	cfg.InstanceID = "generation-capacity-reload-test"
	cfg.Base.Listen = "127.0.0.1:0"
	cfg.Base.MaxConnections = 2
	cfg.Base.MaxWorkers = 1
	gateway, err := NewGatewayFactory().CreateGateway(&cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := gateway.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = gateway.Stop()
	})

	if got := gateway.dispatcher.maxConnections.Load(); got != 2 {
		t.Fatalf("initial max connections = %d, want 2", got)
	}
	if got := gateway.requestLimiter.limit.Load(); got != 1 {
		t.Fatalf("initial max workers = %d, want 1", got)
	}

	next := cfg
	next.Base.MaxConnections = 4
	next.Base.MaxWorkers = 3
	if err := gateway.Reload(&next); err != nil {
		t.Fatal(err)
	}
	if got := gateway.dispatcher.maxConnections.Load(); got != 4 {
		t.Fatalf("reloaded max connections = %d, want 4", got)
	}
	if got := gateway.requestLimiter.limit.Load(); got != 3 {
		t.Fatalf("reloaded max workers = %d, want 3", got)
	}
}

func newTestGeneration(name string, handler http.HandlerFunc) *gatewayGeneration {
	cfg := &config.GatewayConfig{
		InstanceID: name,
		Base: config.BaseConfig{
			GracefulShutdownTimeout: time.Second,
		},
	}
	return &gatewayGeneration{
		config:    cfg,
		server:    &http.Server{Handler: handler},
		serveDone: make(chan struct{}),
	}
}

func requestBody(t *testing.T, address string) string {
	t.Helper()
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	resp, err := client.Get("http://" + address)
	if err != nil {
		t.Errorf("request failed: %v", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("read response failed: %v", err)
		return ""
	}
	return string(body)
}

func writeTestCertificate(t *testing.T, commonName string) (string, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	return certPath, keyPath
}
