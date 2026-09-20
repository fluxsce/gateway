package config

import (
	"sync"

	"gateway/internal/servicecenterv3/model"
)

// watcher 管理单个中心实例上的配置 Watch。投递在锁外阻塞，通道满则等待，不丢事件。
type watcher struct {
	mu    sync.RWMutex
	items []*watchItem
}

// watchItem 是一条 Watch：dataIDs 空表示组内全部配置。
type watchItem struct {
	namespaceID  string
	group        string
	dataIDs      map[string]struct{}
	ch           chan<- model.ConfigEvent
	connectionID string
}

// newWatcher 构造空 Watch 表。
func newWatcher() *watcher {
	return &watcher{}
}

// add 登记 Watch；dataIDs 空表示该组下全部已发布配置。
func (w *watcher) add(namespaceID, group string, dataIDs []string, ch chan<- model.ConfigEvent, connectionID string) {
	set := map[string]struct{}{}
	for _, id := range dataIDs {
		if id != "" {
			set[id] = struct{}{}
		}
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.items = append(w.items, &watchItem{
		namespaceID:  namespaceID,
		group:        group,
		dataIDs:      set,
		ch:           ch,
		connectionID: connectionID,
	})
}

// bindConnection 把已登记的 ch 与数据面连接绑定。
func (w *watcher) bindConnection(ch chan<- model.ConfigEvent, connectionID string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, it := range w.items {
		if it.ch == ch {
			it.connectionID = connectionID
		}
	}
}

// remove 取消该通道上的全部 Watch。
func (w *watcher) remove(ch chan<- model.ConfigEvent) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.items = compactWatchItems(w.items, func(it *watchItem) bool { return it.ch != ch })
}

// removeByConnection 按连接 ID 取消 Watch，空 ID 为无操作。
func (w *watcher) removeByConnection(connectionID string) {
	if connectionID == "" {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.items = compactWatchItems(w.items, func(it *watchItem) bool { return it.connectionID != connectionID })
}

// compactWatchItems 原地压缩并把尾部指针置空，避免切片复用把已取消 Watch 的通道钉在堆上。
func compactWatchItems(items []*watchItem, keep func(*watchItem) bool) []*watchItem {
	n := 0
	for _, it := range items {
		if keep(it) {
			items[n] = it
			n++
		}
	}
	for i := n; i < len(items); i++ {
		items[i] = nil
	}
	return items[:n]
}

// removeDataIDs 从该通道去掉指定 dataID。ids 空表示去掉该 ns/group 下按 ID 订阅的条目。
func (w *watcher) removeDataIDs(ch chan<- model.ConfigEvent, namespaceID, group string, ids []string) {
	drop := map[string]struct{}{}
	for _, id := range ids {
		if id != "" {
			drop[id] = struct{}{}
		}
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.items = compactWatchItems(w.items, func(it *watchItem) bool {
		if it.ch != ch {
			return true
		}
		if namespaceID != "" && it.namespaceID != namespaceID {
			return true
		}
		if group != "" && it.group != group {
			return true
		}
		if len(it.dataIDs) == 0 {
			return true
		}
		if len(drop) == 0 {
			return false
		}
		for id := range drop {
			delete(it.dataIDs, id)
		}
		return len(it.dataIDs) > 0
	})
}

// publish 向匹配的 Watch 投递。先在读锁下拷贝目标，再锁外阻塞写入，不丢事件。
func (w *watcher) publish(ev model.ConfigEvent) {
	w.mu.RLock()
	targets := make([]chan<- model.ConfigEvent, 0, len(w.items))
	seen := make(map[chan<- model.ConfigEvent]struct{}, len(w.items))
	for _, it := range w.items {
		if it.namespaceID != ev.NamespaceID {
			continue
		}
		if it.group != "" && ev.GroupName != "" && it.group != ev.GroupName {
			continue
		}
		if len(it.dataIDs) > 0 {
			if _, ok := it.dataIDs[ev.DataID]; !ok {
				continue
			}
		}
		if _, ok := seen[it.ch]; ok {
			continue
		}
		seen[it.ch] = struct{}{}
		targets = append(targets, it.ch)
	}
	w.mu.RUnlock()
	for _, ch := range targets {
		deliverConfig(ch, ev)
	}
}

// deliverConfig 阻塞写入；通道已关闭时忽略 panic。
func deliverConfig(ch chan<- model.ConfigEvent, ev model.ConfigEvent) {
	defer func() { _ = recover() }()
	ch <- ev
}
