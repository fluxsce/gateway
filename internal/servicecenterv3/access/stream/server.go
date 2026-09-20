package stream

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/alert"
	"gateway/internal/servicecenterv3/infra/store"
	"gateway/internal/servicecenterv3/model"
	pb "gateway/internal/servicecenterv3/proto"
	"gateway/pkg/logger"
	"gateway/pkg/utils/cert"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// Server 在 CenterInstance 监听端口上只注册 ServiceCenterStream。
type Server struct {
	cfg     *model.CenterInstance
	handler *Handler
	store   *store.Store
	grpc    *grpc.Server
	ln      net.Listener
	running atomic.Bool
	wg      sync.WaitGroup
}

// compatStreamServiceName 与旧 Java Stream 客户端的 gRPC 路径一致。
// 生成代码因 protobuf package 改为 servicecenterv3.stream，必须额外注册旧名。
const compatStreamServiceName = "stream.ServiceCenterStream"

func registerStreamService(gs grpc.ServiceRegistrar, srv pb.ServiceCenterStreamServer) {
	pb.RegisterServiceCenterStreamServer(gs, srv)
	compat := pb.ServiceCenterStream_ServiceDesc
	compat.ServiceName = compatStreamServiceName
	gs.RegisterService(&compat, srv)
}

// NewServer 构造流服务器，Start 之前不占端口。
func NewServer(cfg *model.CenterInstance, naming contract.Naming, config contract.Config, session Session, st *store.Store) *Server {
	return &Server{
		cfg:     cfg,
		handler: NewHandler(cfg, naming, config, session),
		store:   st,
	}
}

// Start 监听并服务 Connect 流。重复调用返回错误。
func (s *Server) Start(ctx context.Context) error {
	if s.running.Load() {
		return fmt.Errorf("stream server already running")
	}
	opts := s.options()
	if model.IsY(s.cfg.EnableTLS) {
		tlsCfg, err := s.tlsConfig()
		if err != nil {
			return err
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsCfg)))
	}
	gs := grpc.NewServer(opts...)
	registerStreamService(gs, s.handler)
	if model.IsY(s.cfg.EnableReflection) {
		reflection.Register(gs)
	}
	addr := s.cfg.ListenEndpoint()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("监听 %s 失败: %w", addr, err)
	}
	s.grpc = gs
	s.ln = ln
	s.running.Store(true)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := gs.Serve(ln); err != nil && s.running.Load() {
			logger.Error("servicecenterv3 stream 异常停止", err, "instance", s.cfg.InstanceName)
			alert.StopAbnormal(s.cfg, err.Error())
		}
		s.running.Store(false)
	}()
	time.Sleep(50 * time.Millisecond)
	return nil
}

// Stop 优雅关闭 gRPC，超时后强制 Stop。未运行时为无操作。
func (s *Server) Stop() {
	if !s.running.Swap(false) {
		return
	}
	if s.handler != nil {
		s.handler.NotifyClose("server_shutdown", "服务端正在关闭", 5)
	}
	if s.grpc != nil {
		stopped := make(chan struct{})
		go func() {
			s.grpc.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			s.grpc.Stop()
		}
	}
	s.wg.Wait()
}

// options 按中心实例定义组装 gRPC 选项，含 keepalive、大小限制与鉴权拦截器。
func (s *Server) options() []grpc.ServerOption {
	cfg := s.cfg
	recv := cfg.MaxRecvMsgSize
	if recv <= 0 {
		recv = 16 * 1024 * 1024
	}
	send := cfg.MaxSendMsgSize
	if send <= 0 {
		send = 16 * 1024 * 1024
	}
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(recv),
		grpc.MaxSendMsgSize(send),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             time.Duration(nonzero(cfg.KeepAliveMinTime, 5)) * time.Second,
			PermitWithoutStream: model.IsY(cfg.PermitWithoutStream),
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:                  time.Duration(nonzero(cfg.KeepAliveTime, 30)) * time.Second,
			Timeout:               time.Duration(nonzero(cfg.KeepAliveTimeout, 10)) * time.Second,
			MaxConnectionIdle:     time.Duration(cfg.MaxConnectionIdle) * time.Second,
			MaxConnectionAge:      time.Duration(cfg.MaxConnectionAge) * time.Second,
			MaxConnectionAgeGrace: time.Duration(nonzero(cfg.MaxConnectionAgeGrace, 20)) * time.Second,
		}),
		grpc.ChainStreamInterceptor(chainInterceptors(cfg, s.store)),
	}
	if cfg.MaxConcurrentStreams > 0 {
		opts = append(opts, grpc.MaxConcurrentStreams(uint32(cfg.MaxConcurrentStreams)))
	}
	if cfg.ReadBufferSize > 0 {
		opts = append(opts, grpc.ReadBufferSize(cfg.ReadBufferSize))
	}
	if cfg.WriteBufferSize > 0 {
		opts = append(opts, grpc.WriteBufferSize(cfg.WriteBufferSize))
	}
	return opts
}

// tlsConfig 按 FILE/DATABASE 证书来源构造 TLS；EnableMTLS=Y 时要求客户端证书。
func (s *Server) tlsConfig() (*tls.Config, error) {
	cfg := s.cfg
	certCfg := &cert.CertConfig{KeyPassword: cfg.CertPassword}
	switch cfg.CertStorageType {
	case "FILE":
		certCfg.CertFile = cfg.CertFilePath
		certCfg.KeyFile = cfg.KeyFilePath
	case "DATABASE":
		certCfg.CertContent = cfg.CertContent
		certCfg.KeyContent = cfg.KeyContent
	default:
		return nil, fmt.Errorf("不支持的证书存储类型: %s", cfg.CertStorageType)
	}
	tlsCfg, err := cert.NewCertLoader(certCfg).CreateTLSConfig()
	if err != nil {
		return nil, err
	}
	if model.IsY(cfg.EnableMTLS) {
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		clientCAs, caErr := loadClientCAPool(cfg.CertChainContent)
		if caErr != nil {
			return nil, fmt.Errorf("enableMTLS=Y requires certChainContent: %w", caErr)
		}
		tlsCfg.ClientCAs = clientCAs
	}
	return tlsCfg, nil
}

// loadClientCAPool 从 certChainContent 加载 mTLS 客户端 CA：PEM 文本或 PEM 文件路径。
func loadClientCAPool(certChainContent string) (*x509.CertPool, error) {
	raw := strings.TrimSpace(certChainContent)
	if raw == "" {
		return nil, fmt.Errorf("certChainContent is empty")
	}
	pemBytes := []byte(raw)
	if !strings.Contains(raw, "BEGIN CERTIFICATE") {
		b, err := os.ReadFile(raw)
		if err != nil {
			return nil, fmt.Errorf("read CA file: %w", err)
		}
		pemBytes = b
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		return nil, fmt.Errorf("unable to parse client CA certificate")
	}
	return pool, nil
}

// nonzero 在 v<=0 时返回 fallback，避免把 0 秒 keepalive 交给 gRPC。
func nonzero(v, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}
