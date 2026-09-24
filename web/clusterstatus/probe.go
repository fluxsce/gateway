package clusterstatus

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"gateway/pkg/config"
)

const (
	probeStatusOK   = "ok"
	probeStatusFail = "fail"
	probeStatusSkip = "skip"
	probeTimeout    = 2 * time.Second
	probeParallel   = 8
)

// ListenTarget 一次探测的监听口。各模块从自己的表填这些字段，探测不再关心表结构。
type ListenTarget struct {
	ID          string
	Name        string
	BindAddress string
	Scheme      string
	Port        int
	HealthPath  string
}

// probeListeners 从当前进程连每个存活节点上的监听口。
// 实例表的 healthStatus 是多进程共用一行，后写的会盖住先写的，这里不用它。
func probeListeners(ctx context.Context, nodes []ClusterNodeRow, targets []ListenTarget) []ListenProbe {
	if len(nodes) == 0 || len(targets) == 0 {
		return []ListenProbe{}
	}
	jobs := make([]ListenProbe, 0, len(nodes)*len(targets))
	jobTarget := make([]int, 0, len(nodes)*len(targets))
	for _, node := range nodes {
		for ti, target := range targets {
			jobs = append(jobs, planProbe(node, target))
			jobTarget = append(jobTarget, ti)
		}
	}
	transport := &http.Transport{
		Proxy:                 nil,
		DisableKeepAlives:     true,
		DialContext:           (&net.Dialer{Timeout: probeTimeout}).DialContext,
		TLSHandshakeTimeout:   probeTimeout,
		ResponseHeaderTimeout: probeTimeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{
		Timeout:   probeTimeout,
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	sem := make(chan struct{}, probeParallel)
	done := make(chan struct{}, len(jobs))
	for i := range jobs {
		if jobs[i].Status == probeStatusSkip {
			done <- struct{}{}
			continue
		}
		go func(i int) {
			defer func() { done <- struct{}{} }()
			select {
			case <-ctx.Done():
				jobs[i].Status = probeStatusFail
				jobs[i].Message = "探测已取消"
				return
			case sem <- struct{}{}:
			}
			defer func() { <-sem }()
			status, message := dialListener(ctx, client, jobs[i], targets[jobTarget[i]].HealthPath)
			jobs[i].Status = status
			jobs[i].Message = message
		}(i)
	}
	for range jobs {
		<-done
	}
	return jobs
}

func planProbe(node ClusterNodeRow, target ListenTarget) ListenProbe {
	item := ListenProbe{
		NodeId:     node.NodeId,
		NodeIp:     node.NodeIp,
		Hostname:   node.Hostname,
		TargetId:   target.ID,
		TargetName: target.Name,
		Scheme:     target.Scheme,
		Port:       target.Port,
		Status:     probeStatusFail,
	}
	host, skip := probeHost(target.BindAddress, node.NodeIp)
	if skip != "" {
		item.Status = probeStatusSkip
		item.Message = skip
		return item
	}
	if target.Port <= 0 {
		item.Status = probeStatusSkip
		item.Message = "没有监听端口"
		return item
	}
	item.NodeIp = host
	return item
}

func probeHost(bind, nodeIP string) (string, string) {
	bind = strings.TrimSpace(bind)
	nodeIP = strings.TrimSpace(nodeIP)
	switch strings.ToLower(bind) {
	case "", "0.0.0.0", "::", "[::]", "*":
		if nodeIP == "" {
			return "", "节点没有 IP"
		}
		return nodeIP, ""
	case "127.0.0.1", "localhost", "::1", "[::1]":
		return "", "绑定在本机回环，别的机器探不到这个端口"
	}
	if nodeIP != "" && bind == nodeIP {
		return nodeIP, ""
	}
	return "", "绑定地址不是这个节点"
}

func listenHealthPath() string {
	path := strings.TrimSpace(config.GetString("app.gateway.health.path", "/_gw/health"))
	if path == "" || path == "/" || path == "-" || strings.EqualFold(path, "off") {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func dialListener(ctx context.Context, client *http.Client, item ListenProbe, path string) (string, string) {
	host := item.NodeIp
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		host = "[" + host + "]"
	}
	addr := fmt.Sprintf("%s:%d", host, item.Port)
	if path == "" {
		dialer := net.Dialer{Timeout: probeTimeout}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return probeStatusFail, "端口连不上"
		}
		_ = conn.Close()
		return probeStatusOK, "端口可连接"
	}
	url := fmt.Sprintf("%s://%s%s", item.Scheme, addr, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return probeStatusFail, "无法发起探测"
	}
	resp, err := client.Do(req)
	if err != nil {
		return probeStatusFail, "端口连不上"
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 256))
	if resp.StatusCode != http.StatusOK {
		return probeStatusFail, fmt.Sprintf("探活返回 %d", resp.StatusCode)
	}
	return probeStatusOK, "探活正常"
}
