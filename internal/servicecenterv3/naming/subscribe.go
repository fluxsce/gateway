package naming

import (
	"sync"

	"gateway/internal/servicecenterv3/model"
)

// subscriber 管理单个中心实例上的命名订阅。投递在锁外阻塞，通道满则等待，不丢事件。
type subscriber struct {
	mu    sync.RWMutex
	items []*subItem
}

// subItem 是一条订阅：指定服务名集合，或 allInNS 覆盖命名空间。
type subItem struct {
	namespaceID  string
	group        string
	serviceNames map[string]struct{}
	allInNS      bool
	ch           chan<- model.NamingEvent
	connectionID string
}

// newSubscriber 构造空订阅表。
func newSubscriber() *subscriber {
	return &subscriber{}
}

// add 按服务名列表订阅；names 空时不按服务名过滤（该组下全部服务）。
func (s *subscriber) add(namespaceID, group string, names []string, ch chan<- model.NamingEvent, connectionID string) {
	set := map[string]struct{}{}
	for _, n := range names {
		if n != "" {
			set[n] = struct{}{}
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, &subItem{
		namespaceID:  namespaceID,
		group:        group,
		serviceNames: set,
		ch:           ch,
		connectionID: connectionID,
	})
}

// addNamespace 订阅命名空间（可选组）下全部服务。
func (s *subscriber) addNamespace(namespaceID, group string, ch chan<- model.NamingEvent, connectionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, &subItem{
		namespaceID:  namespaceID,
		group:        group,
		allInNS:      true,
		ch:           ch,
		connectionID: connectionID,
	})
}

// bindConnection 把已登记的 ch 与数据面连接绑定。
func (s *subscriber) bindConnection(ch chan<- model.NamingEvent, connectionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, it := range s.items {
		if it.ch == ch {
			it.connectionID = connectionID
		}
	}
}

// remove 取消该通道上的全部订阅。
func (s *subscriber) remove(ch chan<- model.NamingEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = compactSubItems(s.items, func(it *subItem) bool { return it.ch != ch })
}

// removeByConnection 按连接 ID 取消订阅，空 ID 为无操作。
func (s *subscriber) removeByConnection(connectionID string) {
	if connectionID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = compactSubItems(s.items, func(it *subItem) bool { return it.connectionID != connectionID })
}

// compactSubItems 原地压缩并把尾部指针置空，避免切片复用把已取消订阅的通道钉在堆上。
func compactSubItems(items []*subItem, keep func(*subItem) bool) []*subItem {
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

// removeServices 从该通道去掉指定服务名。names 空表示去掉该 ns/group 下非命名空间级的订阅。
func (s *subscriber) removeServices(ch chan<- model.NamingEvent, namespaceID, group string, names []string) {
	drop := map[string]struct{}{}
	for _, n := range names {
		if n != "" {
			drop[n] = struct{}{}
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = compactSubItems(s.items, func(it *subItem) bool {
		if it.ch != ch || it.allInNS {
			return true
		}
		if namespaceID != "" && it.namespaceID != namespaceID {
			return true
		}
		if group != "" && it.group != group {
			return true
		}
		if len(drop) == 0 {
			return false
		}
		if len(it.serviceNames) == 0 {
			return true
		}
		for n := range drop {
			delete(it.serviceNames, n)
		}
		return len(it.serviceNames) > 0
	})
}

type subMatch struct {
	connectionID string
	namespaceID  string
	group        string
	serviceName  string
	scope        string
}

func (it *subItem) matches(namespaceID, group, serviceName string) bool {
	if it.namespaceID != namespaceID {
		return false
	}
	if it.group != "" && group != "" && it.group != group {
		return false
	}
	if it.allInNS || len(it.serviceNames) == 0 {
		return true
	}
	_, ok := it.serviceNames[serviceName]
	return ok
}

func (it *subItem) scope() string {
	if it.allInNS {
		return "namespace"
	}
	if len(it.serviceNames) == 0 {
		return "group"
	}
	return "service"
}

func (s *subscriber) listMatching(namespaceID, group, serviceName string) []subMatch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := make(map[string]struct{}, len(s.items))
	out := make([]subMatch, 0)
	for _, it := range s.items {
		if !it.matches(namespaceID, group, serviceName) {
			continue
		}
		key := it.connectionID
		if key == "" {
			key = it.namespaceID + ":" + it.group + ":" + it.scope()
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, subMatch{
			connectionID: it.connectionID,
			namespaceID:  it.namespaceID,
			group:        it.group,
			serviceName:  serviceName,
			scope:        it.scope(),
		})
	}
	return out
}

func (s *subscriber) listByConnections(connectionIDs map[string]struct{}) []subMatch {
	if len(connectionIDs) == 0 {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := make(map[string]struct{}, len(s.items))
	out := make([]subMatch, 0)
	for _, it := range s.items {
		if _, ok := connectionIDs[it.connectionID]; !ok || it.connectionID == "" {
			continue
		}
		scope := it.scope()
		names := it.namedServices()
		if len(names) == 0 {
			key := it.connectionID + ":" + it.namespaceID + ":" + it.group + ":" + scope
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, subMatch{
				connectionID: it.connectionID,
				namespaceID:  it.namespaceID,
				group:        it.group,
				scope:        scope,
			})
			continue
		}
		for _, name := range names {
			key := it.connectionID + ":" + it.namespaceID + ":" + it.group + ":" + name
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, subMatch{
				connectionID: it.connectionID,
				namespaceID:  it.namespaceID,
				group:        it.group,
				serviceName:  name,
				scope:        scope,
			})
		}
	}
	return out
}

func (it *subItem) namedServices() []string {
	if it == nil || it.allInNS || len(it.serviceNames) == 0 {
		return nil
	}
	out := make([]string, 0, len(it.serviceNames))
	for name := range it.serviceNames {
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

// publish 向匹配的订阅投递。先在读锁下拷贝目标，再锁外阻塞写入，避免满队列丢事件，也避免持锁发送造成死锁。
func (s *subscriber) publish(ev model.NamingEvent) {
	s.mu.RLock()
	targets := make([]chan<- model.NamingEvent, 0, len(s.items))
	seen := make(map[chan<- model.NamingEvent]struct{}, len(s.items))
	for _, it := range s.items {
		if it.namespaceID != ev.NamespaceID {
			continue
		}
		if it.group != "" && ev.GroupName != "" && it.group != ev.GroupName {
			continue
		}
		if !it.allInNS && len(it.serviceNames) > 0 {
			if _, ok := it.serviceNames[ev.ServiceName]; !ok {
				continue
			}
		}
		if _, ok := seen[it.ch]; ok {
			continue
		}
		seen[it.ch] = struct{}{}
		targets = append(targets, it.ch)
	}
	s.mu.RUnlock()
	for _, ch := range targets {
		deliverNaming(ch, ev)
	}
}

// deliverNaming 阻塞写入；通道已关闭（连接已清理）时忽略 panic。
func deliverNaming(ch chan<- model.NamingEvent, ev model.NamingEvent) {
	defer func() { _ = recover() }()
	ch <- ev
}
