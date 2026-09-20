package servicecenterv3

import (
	"context"
	"fmt"
	"sync"

	"gateway/internal/servicecenterv3/access/inproc"
	"gateway/internal/servicecenterv3/center"
	"gateway/internal/servicecenterv3/infra/store"
	"gateway/internal/servicecenterv3/naming"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

const (
	// EngineLegacy 仅保留字符串，本版本不再启动旧包。
	EngineLegacy = "legacy"
	// EngineV3 走本包，是当前唯一启动路径。
	EngineV3 = "servicecenterv3"
)

var (
	mu          sync.RWMutex
	defaultPool *center.Pool
	inprocAdp   *inproc.Adapter
)

// Init 用共享数据库构造进程级实例池，可重复调用并覆盖单例。
// db 不可为 nil。返回的 Pool 实现 contract.Admin，并发安全。
func Init(_ context.Context, db database.Database) (*center.Pool, error) {
	if db == nil {
		return nil, fmt.Errorf("servicecenterv3: 数据库实例不能为空")
	}
	pool := center.NewPool(store.New(db))
	mu.Lock()
	defaultPool = pool
	inprocAdp = inproc.New(pool)
	mu.Unlock()
	logger.Info("servicecenterv3 已初始化", "version", Version, "apiLevel", APILevel)
	return pool, nil
}

// StartAll 建池后立即返回，库中已启用实例在后台监听，不阻塞应用启动。
func StartAll(ctx context.Context, db database.Database) error {
	if _, err := Init(ctx, db); err != nil {
		return err
	}
	naming.LocalGatewayID = config.GetNodeId()
	pool := GetPool()
	go func() {
		if err := pool.LoadAndStart(ctx); err != nil {
			logger.Error("后台启动 servicecenterv3 实例失败", "error", err)
			return
		}
		logger.Info("servicecenterv3 实例启动完成")
	}()
	return nil
}

// StopAll 停止进程内全部中心实例。未 Init 时为无操作。
func StopAll(ctx context.Context) error {
	pool := GetPool()
	if pool == nil {
		return nil
	}
	return pool.StopAll(ctx)
}

// GetPool 返回 Init 设置的进程级池，未初始化时为 nil。
func GetPool() *center.Pool {
	mu.RLock()
	defer mu.RUnlock()
	return defaultPool
}

// InProcess 返回同进程 Naming 适配器，未初始化时为 nil。
func InProcess() *inproc.Adapter {
	mu.RLock()
	defer mu.RUnlock()
	return inprocAdp
}

// HasRunning 报告是否至少有一个中心实例在监听。
func HasRunning() bool {
	pool := GetPool()
	return pool != nil && pool.HasRunning()
}
