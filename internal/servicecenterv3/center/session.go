package center

import (
	"sync"
	"time"

	"gateway/internal/servicecenterv3/model"
)

// SessionHub 记录单个中心实例上的数据面连接。
type SessionHub struct {
	mu    sync.RWMutex
	conns map[string]*model.ConnectionInfo
}

// NewSessionHub 构造空会话表。
func NewSessionHub() *SessionHub {
	return &SessionHub{conns: map[string]*model.ConnectionInfo{}}
}

// Add 登记或覆盖一条连接。
func (h *SessionHub) Add(info *model.ConnectionInfo) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[info.ConnectionID] = info
}

// Touch 刷新最近活跃时间。
func (h *SessionHub) Touch(connectionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.conns[connectionID]; ok {
		c.LastActive = time.Now()
	}
}

// Remove 删除连接记录。
func (h *SessionHub) Remove(connectionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, connectionID)
}

// List 返回连接副本，顺序不稳定。
func (h *SessionHub) List() []model.ConnectionInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]model.ConnectionInfo, 0, len(h.conns))
	for _, c := range h.conns {
		out = append(out, *c)
	}
	return out
}

// Count 返回当前连接数。
func (h *SessionHub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns)
}
