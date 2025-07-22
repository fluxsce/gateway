package pprof

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"time"

	"gateway/pkg/logger"
)

// Manager pprof管理器
type Manager struct {
	config   *Config
	server   *http.Server
	analyzer *Analyzer
	running  bool
	mu       sync.RWMutex
	stopCh   chan struct{}
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewManager 创建pprof管理器
func NewManager(config *Config) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		config:   config,
		analyzer: NewAnalyzer(config),
		stopCh:   make(chan struct{}),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start 启动pprof服务
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("pprof服务已经在运行")
	}

	if !m.config.Enabled {
		logger.Info("pprof服务未启用")
		return nil
	}

	// 创建HTTP服务器
	mux := http.NewServeMux()

	// 注册pprof路由
	m.registerPprofRoutes(mux)

	// 注册自定义路由
	m.registerCustomRoutes(mux)

	m.server = &http.Server{
		Addr:         m.config.Listen,
		Handler:      mux,
		ReadTimeout:  m.config.ReadTimeout,
		WriteTimeout: m.config.WriteTimeout,
	}

	// 启动HTTP服务器
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		logger.Info("启动pprof服务", "listen", m.config.Listen)

		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("pprof服务启动失败", "error", err)
		}
	}()

	// 启动自动分析器
	if m.config.AutoAnalysis.Enabled {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.runAutoAnalysis()
		}()
	}

	m.running = true

	logger.Info("pprof服务启动成功",
		"service", m.config.ServiceName,
		"listen", m.config.Listen,
		"web_ui", fmt.Sprintf("http://localhost%s/debug/pprof/", m.config.Listen),
		"auto_analysis", m.config.AutoAnalysis.Enabled,
	)

	return nil
}

// Stop 停止pprof服务
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	logger.Info("正在停止pprof服务...")

	// 发送停止信号
	close(m.stopCh)
	m.cancel()

	// 关闭HTTP服务器
	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := m.server.Shutdown(ctx); err != nil {
			logger.Error("关闭pprof服务器失败", "error", err)
		}
	}

	// 等待所有goroutine结束
	m.wg.Wait()

	m.running = false
	logger.Info("pprof服务已停止")

	return nil
}

// IsRunning 检查服务是否运行
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// registerPprofRoutes 注册pprof路由
func (m *Manager) registerPprofRoutes(mux *http.ServeMux) {
	// 如果启用了认证，添加认证中间件
	if m.config.EnableAuth {
		mux.HandleFunc("/debug/pprof/", m.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.DefaultServeMux.ServeHTTP(w, r)
		})))
	} else {
		mux.HandleFunc("/debug/pprof/", func(w http.ResponseWriter, r *http.Request) {
			http.DefaultServeMux.ServeHTTP(w, r)
		})
	}
}

// registerCustomRoutes 注册自定义路由
func (m *Manager) registerCustomRoutes(mux *http.ServeMux) {
	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 服务信息
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		info := map[string]interface{}{
			"service":       m.config.ServiceName,
			"listen":        m.config.Listen,
			"auto_analysis": m.config.AutoAnalysis.Enabled,
			"running":       m.running,
			"pprof_enabled": m.config.Enabled,
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"service": "%s",
			"listen": "%s",
			"auto_analysis": %t,
			"running": %t,
			"pprof_enabled": %t
		}`, info["service"], info["listen"], info["auto_analysis"], info["running"], info["pprof_enabled"])
	})

	// 手动触发分析
	mux.HandleFunc("/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		go func() {
			if err := m.analyzer.RunAnalysis(); err != nil {
				logger.Error("手动分析失败", "error", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("分析任务已启动"))
	})
}

// authMiddleware 认证中间件
func (m *Manager) authMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			token = r.URL.Query().Get("token")
		}

		if token != m.config.AuthToken {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}

		next.ServeHTTP(w, r)
	}
}

// runAutoAnalysis 运行自动分析
func (m *Manager) runAutoAnalysis() {
	ticker := time.NewTicker(m.config.AutoAnalysis.Interval)
	defer ticker.Stop()

	logger.Info("启动自动分析器",
		"interval", m.config.AutoAnalysis.Interval,
		"output_dir", m.config.AutoAnalysis.OutputDir,
	)

	for {
		select {
		case <-ticker.C:
			if err := m.analyzer.RunAnalysis(); err != nil {
				logger.Error("自动分析失败", "error", err)
			}
		case <-m.stopCh:
			logger.Info("自动分析器已停止")
			return
		case <-m.ctx.Done():
			logger.Info("自动分析器已取消")
			return
		}
	}
}

// GetStatus 获取服务状态
func (m *Manager) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"running":       m.running,
		"service_name":  m.config.ServiceName,
		"listen":        m.config.Listen,
		"auto_analysis": m.config.AutoAnalysis.Enabled,
		"auth_enabled":  m.config.EnableAuth,
		"output_dir":    m.config.AutoAnalysis.OutputDir,
	}
}

// UpdateConfig 更新配置
func (m *Manager) UpdateConfig(newConfig *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("无法在运行时更新配置，请先停止服务")
	}

	m.config = newConfig
	m.analyzer = NewAnalyzer(newConfig)

	logger.Info("pprof配置已更新")
	return nil
}

// CreateOutputDir 创建输出目录
func (m *Manager) CreateOutputDir() error {
	if m.config.AutoAnalysis.OutputDir == "" {
		return nil
	}

	if err := os.MkdirAll(m.config.AutoAnalysis.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	return nil
}
