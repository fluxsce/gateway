package service

import (
	"os"
	"time"

	"gateway/internal/cluster/types"
	"gateway/pkg/logger"
)

const (
	nodeHeartbeatInterval = 15 * time.Second
	nodeStaleAfter        = 90 * time.Second
)

// nodeLoop 登记本节点并按心跳刷新。过期节点从集群表和采集节点表去掉。
func (s *ClusterServiceImpl) nodeLoop() {
	defer s.wg.Done()

	if s.touchNode() {
		s.expireNodes()
	}

	ticker := time.NewTicker(nodeHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if s.touchNode() {
				s.expireNodes()
			}
		}
	}
}

// touchNode 登记本节点。失败时返回 false，调用方不要接着清理采集节点。
func (s *ClusterServiceImpl) touchNode() bool {
	now := time.Now()
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}
	node := &types.ClusterNode{
		NodeId:         s.nodeId,
		TenantId:       s.tenantId,
		NodeIp:         s.nodeIp,
		Hostname:       hostname,
		StartedTime:    now,
		LastSeenTime:   now,
		AddTime:        now,
		AddWho:         s.nodeId,
		EditTime:       now,
		EditWho:        s.nodeId,
		OprSeqFlag:     now.Format("20060102150405"),
		CurrentVersion: 1,
		ActiveFlag:     "Y",
	}
	if err := s.nodeDAO.Upsert(s.ctx, node); err != nil {
		logger.Warn("登记集群节点失败", "error", err, "nodeId", s.nodeId)
		return false
	}
	return true
}

func (s *ClusterServiceImpl) expireNodes() {
	seenBefore := time.Now().Add(-nodeStaleAfter)
	if err := s.nodeDAO.Expire(s.ctx, s.tenantId, seenBefore); err != nil {
		logger.Warn("清理过期集群节点失败", "error", err, "nodeId", s.nodeId)
	}
}
