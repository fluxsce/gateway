// Package static 实现健康检查功能
// 负责定期检查后端节点的健康状态
package static

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// startHealthCheck 启动健康检查
// 根据配置的间隔定期检查所有节点的健康状态
//
// 参数:
//   - ctx: 代理内部 context，用于控制健康检查生命周期
func (p *staticProxy) startHealthCheck(ctx context.Context) {
	// 如果未配置健康检查类型，则不启动
	if p.server.HealthCheckType == nil || *p.server.HealthCheckType == "" {
		return
	}

	// 获取健康检查间隔
	interval := 30 * time.Second
	if p.server.HealthCheckInterval != nil {
		interval = time.Duration(*p.server.HealthCheckInterval) * time.Second
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.performHealthCheck()
			}
		}
	}()
}

// performHealthCheck 执行健康检查
// 检查所有节点的健康状态并更新数据库
func (p *staticProxy) performHealthCheck() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.server.Nodes == nil {
		return
	}

	// 为数据库操作创建独立的 context（带超时，不依赖代理生命周期）
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, node := range p.server.Nodes {
		healthy := p.checkNodeHealth(node)

		// 更新健康检查时间
		now := time.Now()
		node.LastHealthCheck = &now

		// 更新节点状态
		status := "healthy"
		if !healthy {
			status = "unhealthy"
			node.NodeStatus = "inactive"
			node.FailureCount++
		} else {
			node.NodeStatus = "active"
		}
		node.HealthCheckStatus = &status

		// 更新数据库（使用独立的 dbCtx）
		if err := p.staticNodeRepo.Update(dbCtx, node); err != nil {
			logger.Error("Failed to update node health check in database", map[string]interface{}{
				"nodeID": node.TunnelStaticNodeId,
				"error":  err.Error(),
			})
		}
	}
}

// checkNodeHealth 检查节点健康状态
// 根据配置的健康检查类型执行相应的检查
func (p *staticProxy) checkNodeHealth(node *types.TunnelStaticNode) bool {
	// 获取健康检查类型
	checkType := "tcp"
	if p.server.HealthCheckType != nil {
		checkType = *p.server.HealthCheckType
	}

	// 获取健康检查超时时间
	timeout := 5 * time.Second
	if p.server.HealthCheckTimeout != nil {
		timeout = time.Duration(*p.server.HealthCheckTimeout) * time.Second
	}

	// 根据类型执行健康检查
	switch checkType {
	case "tcp":
		return p.checkTCPHealth(node, timeout)
	case "http", "https":
		return p.checkHTTPHealth(node, timeout, checkType)
	default:
		// 默认使用TCP检查
		return p.checkTCPHealth(node, timeout)
	}
}

// checkTCPHealth TCP健康检查
// 尝试建立TCP连接到目标节点
func (p *staticProxy) checkTCPHealth(node *types.TunnelStaticNode, timeout time.Duration) bool {
	address := net.JoinHostPort(node.TargetAddress, fmt.Sprintf("%d", node.TargetPort))
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// checkHTTPHealth HTTP健康检查
// 发送HTTP请求到健康检查URL并检查响应状态码
func (p *staticProxy) checkHTTPHealth(node *types.TunnelStaticNode, timeout time.Duration, scheme string) bool {
	// 构建健康检查URL
	url := p.buildHealthCheckURL(node, scheme)
	if url == "" {
		// 如果没有配置健康检查URL，回退到TCP检查
		return p.checkTCPHealth(node, timeout)
	}

	// 创建带超时的独立 context（不依赖外部 context）
	reqCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: func(dialCtx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{
					Timeout: timeout,
				}
				return dialer.DialContext(dialCtx, network, addr)
			},
			// 禁用连接复用，确保每次检查都使用新连接
			DisableKeepAlives: true,
		},
	}

	// 创建请求（使用独立的 context）
	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		logger.Error("Failed to create health check request", map[string]interface{}{
			"nodeID": node.TunnelStaticNodeId,
			"url":    url,
			"error":  err.Error(),
		})
		return false
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Gateway-Health-Check/1.0")
	req.Header.Set("Connection", "close")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() {
		// 确保响应体被完全读取并关闭，避免连接泄露
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	// 检查状态码：2xx 和 3xx 都认为是健康的
	statusCode := resp.StatusCode
	return statusCode >= 200 && statusCode < 400
}

// buildHealthCheckURL 构建健康检查URL
// 如果配置了健康检查URL，使用配置的URL；否则使用节点的目标地址和端口
func (p *staticProxy) buildHealthCheckURL(node *types.TunnelStaticNode, scheme string) string {
	// 如果服务器配置了健康检查URL，使用配置的URL
	if p.server.HealthCheckUrl != nil && *p.server.HealthCheckUrl != "" {
		url := *p.server.HealthCheckUrl
		// 如果URL不包含协议，添加scheme
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = scheme + "://" + url
		}
		return url
	}

	// 否则使用节点的目标地址和端口构建URL
	// 默认使用根路径
	host := node.TargetAddress
	port := node.TargetPort

	// 如果端口是标准端口，可以省略
	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		return fmt.Sprintf("%s://%s/", scheme, host)
	}

	return fmt.Sprintf("%s://%s:%d/", scheme, host, port)
}
