// Package live 是服务注册发现的集群活视图。
// 节点心跳、Owner、健康状态以这里为准；HUB_SERVICE / HUB_SERVICE_NODE 仍是目录与审计。
//
// 目录按中心 → 命名空间 → 服务 → 节点，便于点开看：
//
//	scv3:g:{center}                         Hash  命名空间 → 1
//	scv3:ns:{center}:{ns}                   Hash  group|service → tenant
//	scv3:s:{tenant}:{ns}:{group}:{service}  Hash  nodeId → 1
//	scv3:n:{tenant}:{nodeId}                String 节点 JSON，临时节点带 TTL
//
// 一个服务通常几十个节点，先不拆 scv3:s；真到几百再按 nodeId 分片。
// 生命时钟只在 scv3:n 上 EXPIRE。集群下正文逐个 GET（MGET 会 CROSSSLOT）。
package live

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"gateway/internal/servicecenterv3/model"
	pkgcache "gateway/pkg/cache"
)

const (
	nodePrefix   = "scv3:n:"
	svcPrefix    = "scv3:s:"
	nsPrefix     = "scv3:ns:"
	centerPrefix = "scv3:g:"
	sep          = ":"
	serviceField = "|"
	present      = "1"
)

// Store 用 pkg/cache 存节点 JSON 与分层目录。后端是 Redis 时各网关共享同一份活视图。
type Store struct {
	c pkgcache.Cache
}

// New 包装缓存后端。c 为 nil 时返回 nil。
func New(c pkgcache.Cache) *Store {
	if c == nil {
		return nil
	}
	return &Store{c: c}
}

// Put 写入节点并挂到中心 / 命名空间 / 服务目录。ttl<=0 时节点键永不过期（持久节点）。
func (s *Store) Put(ctx context.Context, node *model.Node, ttl time.Duration) error {
	if s == nil || s.c == nil || node == nil || node.NodeID == "" {
		return nil
	}
	cp := cloneNode(node)
	raw, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = -1
	}
	if err := s.c.Set(ctx, nodeKey(cp.TenantID, cp.NodeID), raw, ttl); err != nil {
		return err
	}
	if err := s.c.HSet(ctx, serviceKey(cp.TenantID, cp.NamespaceID, cp.GroupName, cp.ServiceName), cp.NodeID, present); err != nil {
		return err
	}
	if cp.CenterInstanceName == "" || cp.NamespaceID == "" {
		return nil
	}
	if err := s.c.HSet(ctx, nsKey(cp.CenterInstanceName, cp.NamespaceID), serviceRef(cp.GroupName, cp.ServiceName), norm(cp.TenantID)); err != nil {
		return err
	}
	return s.c.HSet(ctx, centerKey(cp.CenterInstanceName), cp.NamespaceID, present)
}

// Touch 只续节点键 TTL，不写 JSON、不碰目录。键不存在返回 false。
func (s *Store) Touch(ctx context.Context, tenantID, nodeID string, ttl time.Duration) (bool, error) {
	if s == nil || s.c == nil || nodeID == "" {
		return true, nil
	}
	key := nodeKey(tenantID, nodeID)
	if ttl <= 0 {
		raw, err := s.c.Get(ctx, key)
		if err != nil {
			return false, err
		}
		return raw != nil, nil
	}
	return s.c.Expire(ctx, key, ttl)
}

// Get 按租户 + nodeID 读节点。键不存在返回 nil, nil。
func (s *Store) Get(ctx context.Context, tenantID, nodeID string) (*model.Node, error) {
	if s == nil || s.c == nil || nodeID == "" {
		return nil, nil
	}
	raw, err := s.c.Get(ctx, nodeKey(tenantID, nodeID))
	if err != nil || raw == nil {
		return nil, err
	}
	return decodeNode(raw)
}

// Delete 删节点正文，并从服务 Hash 摘掉；服务/命名空间空了则沿目录往上摘。
func (s *Store) Delete(ctx context.Context, node *model.Node) error {
	if s == nil || s.c == nil || node == nil || node.NodeID == "" {
		return nil
	}
	if err := s.c.Delete(ctx, nodeKey(node.TenantID, node.NodeID)); err != nil {
		return err
	}
	svc := serviceKey(node.TenantID, node.NamespaceID, node.GroupName, node.ServiceName)
	_, _ = s.c.HDel(ctx, svc, node.NodeID)
	s.pruneService(ctx, node)
	return nil
}

// ListByService 列出该服务下仍活着的节点；正文没了的 field 从服务 Hash 摘掉。
func (s *Store) ListByService(ctx context.Context, tenantID, namespaceID, group, service string) ([]*model.Node, error) {
	if s == nil || s.c == nil || service == "" {
		return nil, nil
	}
	fields, err := s.c.HGetAll(ctx, serviceKey(tenantID, namespaceID, group, service))
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(fields))
	for id := range fields {
		if id != "" {
			ids = append(ids, id)
		}
	}
	out, stale, err := s.loadNodes(ctx, tenantID, ids)
	if err != nil {
		return nil, err
	}
	if len(stale) > 0 {
		_, _ = s.c.HDel(ctx, serviceKey(tenantID, namespaceID, group, service), stale...)
	}
	return out, nil
}

// ListByCenter 沿中心 → 命名空间 → 服务走目录，列出仍活着的节点。
func (s *Store) ListByCenter(ctx context.Context, centerName string) ([]*model.Node, error) {
	if s == nil || s.c == nil || centerName == "" {
		return nil, nil
	}
	namespaces, err := s.c.HGetAll(ctx, centerKey(centerName))
	if err != nil {
		return nil, err
	}
	out := make([]*model.Node, 0)
	emptyNS := make([]string, 0)
	for nsID := range namespaces {
		if nsID == "" {
			continue
		}
		services, err := s.c.HGetAll(ctx, nsKey(centerName, nsID))
		if err != nil {
			return nil, err
		}
		if len(services) == 0 {
			emptyNS = append(emptyNS, nsID)
			continue
		}
		emptySvc := make([]string, 0)
		for ref, tenant := range services {
			group, name, ok := splitServiceRef(ref)
			if !ok {
				emptySvc = append(emptySvc, ref)
				continue
			}
			if tenant == "_" {
				tenant = ""
			}
			nodes, err := s.ListByService(ctx, tenant, nsID, group, name)
			if err != nil {
				return nil, err
			}
			if len(nodes) == 0 {
				emptySvc = append(emptySvc, ref)
				continue
			}
			out = append(out, nodes...)
		}
		if len(emptySvc) > 0 {
			_, _ = s.c.HDel(ctx, nsKey(centerName, nsID), emptySvc...)
		}
		left, _ := s.c.HGetAll(ctx, nsKey(centerName, nsID))
		if len(left) == 0 {
			emptyNS = append(emptyNS, nsID)
		}
	}
	if len(emptyNS) > 0 {
		_, _ = s.c.HDel(ctx, centerKey(centerName), emptyNS...)
	}
	return out, nil
}

func (s *Store) pruneService(ctx context.Context, node *model.Node) {
	if node == nil {
		return
	}
	left, err := s.c.HGetAll(ctx, serviceKey(node.TenantID, node.NamespaceID, node.GroupName, node.ServiceName))
	if err != nil || len(left) > 0 {
		return
	}
	_ = s.c.Delete(ctx, serviceKey(node.TenantID, node.NamespaceID, node.GroupName, node.ServiceName))
	if node.CenterInstanceName == "" || node.NamespaceID == "" {
		return
	}
	ns := nsKey(node.CenterInstanceName, node.NamespaceID)
	_, _ = s.c.HDel(ctx, ns, serviceRef(node.GroupName, node.ServiceName))
	remain, _ := s.c.HGetAll(ctx, ns)
	if len(remain) > 0 {
		return
	}
	_ = s.c.Delete(ctx, ns)
	_, _ = s.c.HDel(ctx, centerKey(node.CenterInstanceName), node.NamespaceID)
}

func (s *Store) loadNodes(ctx context.Context, tenantID string, ids []string) ([]*model.Node, []string, error) {
	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, nodeKey(tenantID, id))
	}
	got, err := s.fetchNodes(ctx, keys)
	if err != nil {
		return nil, nil, err
	}
	out := make([]*model.Node, 0, len(got))
	stale := make([]string, 0)
	for _, id := range ids {
		n := got[nodeKey(tenantID, id)]
		if n == nil {
			stale = append(stale, id)
			continue
		}
		out = append(out, n)
	}
	return out, stale, nil
}

// fetchNodes 先试 MGET（单机/同 slot）；集群 CROSSSLOT 时改逐个 GET。
func (s *Store) fetchNodes(ctx context.Context, keys []string) (map[string]*model.Node, error) {
	out := make(map[string]*model.Node, len(keys))
	if len(keys) == 0 {
		return out, nil
	}
	raws, err := s.c.MGet(ctx, keys)
	if err != nil {
		return s.getNodes(ctx, keys)
	}
	for _, key := range keys {
		raw, ok := raws[key]
		if !ok || raw == nil {
			continue
		}
		n, decErr := decodeNode(raw)
		if decErr != nil || n == nil {
			continue
		}
		out[key] = n
	}
	return out, nil
}

func (s *Store) getNodes(ctx context.Context, keys []string) (map[string]*model.Node, error) {
	out := make(map[string]*model.Node, len(keys))
	for _, key := range keys {
		raw, err := s.c.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		if raw == nil {
			continue
		}
		n, err := decodeNode(raw)
		if err != nil || n == nil {
			continue
		}
		out[key] = n
	}
	return out, nil
}

func decodeNode(raw []byte) (*model.Node, error) {
	var node model.Node
	if err := json.Unmarshal(raw, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

func nodeKey(tenantID, nodeID string) string {
	return nodePrefix + norm(tenantID) + sep + nodeID
}

func serviceKey(tenantID, namespaceID, group, service string) string {
	return svcPrefix + norm(tenantID) + sep + namespaceID + sep + group + sep + service
}

func nsKey(centerName, namespaceID string) string {
	return nsPrefix + centerName + sep + namespaceID
}

func centerKey(centerName string) string {
	return centerPrefix + centerName
}

func serviceRef(group, service string) string {
	return group + serviceField + service
}

func splitServiceRef(ref string) (group, service string, ok bool) {
	group, service, found := strings.Cut(ref, serviceField)
	if !found || service == "" {
		return "", "", false
	}
	return group, service, true
}

func norm(id string) string {
	if id == "" {
		return "_"
	}
	return id
}

func cloneNode(n *model.Node) *model.Node {
	if n == nil {
		return nil
	}
	cp := *n
	if n.Metadata != nil {
		cp.Metadata = make(map[string]string, len(n.Metadata))
		for k, v := range n.Metadata {
			cp.Metadata[k] = v
		}
	}
	return &cp
}
