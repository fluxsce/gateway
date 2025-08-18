package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gateway/internal/registry/core"
)

// Client 注册中心客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
	tenantId   string
	userId     string
}

// NewClient 创建注册中心客户端
func NewClient(baseURL, tenantId, userId string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tenantId: tenantId,
		userId:   userId,
	}
}

// ================== 服务实例管理 ==================

// RegisterInstance 注册服务实例
func (c *Client) RegisterInstance(ctx context.Context, instance *core.ServiceInstance) error {
	// 设置租户ID和用户ID
	instance.TenantId = c.tenantId
	instance.AddWho = c.userId
	instance.EditWho = c.userId

	data, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("序列化实例数据失败: %w", err)
	}

	_, err = c.doRequest(ctx, "POST", "/api/v1/instances", data)
	return err
}

// DeregisterInstance 注销服务实例
func (c *Client) DeregisterInstance(ctx context.Context, instanceId string) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/instances/%s", instanceId), nil)
	return err
}

// GetInstance 获取服务实例
func (c *Client) GetInstance(ctx context.Context, instanceId string) (*core.ServiceInstance, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/instances/%s", instanceId), nil)
	if err != nil {
		return nil, err
	}

	var instance core.ServiceInstance
	if err := json.Unmarshal(data, &instance); err != nil {
		return nil, fmt.Errorf("解析实例数据失败: %w", err)
	}

	return &instance, nil
}

// Heartbeat 发送心跳
func (c *Client) Heartbeat(ctx context.Context, instanceId string) error {
	_, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/instances/%s/heartbeat", instanceId), nil)
	return err
}

// UpdateInstanceHealth 更新实例健康状态
func (c *Client) UpdateInstanceHealth(ctx context.Context, instanceId, healthStatus string) error {
	payload := map[string]string{"healthStatus": healthStatus}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化健康状态数据失败: %w", err)
	}

	_, err = c.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/instances/%s/health", instanceId), data)
	return err
}

// ================== 服务发现 ==================

// DiscoverInstances 发现服务实例
func (c *Client) DiscoverInstances(ctx context.Context, serviceName, groupName string, filters map[string]string) ([]*core.ServiceInstance, error) {
	url := fmt.Sprintf("/api/v1/services/%s/instances", serviceName)

	// 添加查询参数
	params := make(map[string]string)
	if groupName != "" {
		params["groupName"] = groupName
	}
	for k, v := range filters {
		params[k] = v
	}

	if len(params) > 0 {
		url += "?"
		first := true
		for k, v := range params {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%s", k, v)
			first = false
		}
	}

	data, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Instances []*core.ServiceInstance `json:"instances"`
		Count     int                     `json:"count"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析服务实例数据失败: %w", err)
	}

	return response.Instances, nil
}

// ListServices 列出服务
func (c *Client) ListServices(ctx context.Context, groupName string) ([]string, error) {
	url := "/api/v1/services"
	if groupName != "" {
		url += "?groupName=" + groupName
	}

	data, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Services []string `json:"services"`
		Count    int      `json:"count"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析服务列表数据失败: %w", err)
	}

	return response.Services, nil
}

// ================== 服务管理 ==================

// CreateService 创建服务
func (c *Client) CreateService(ctx context.Context, service *core.Service) error {
	// 设置租户ID和用户ID
	service.TenantId = c.tenantId
	service.AddWho = c.userId
	service.EditWho = c.userId

	data, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("序列化服务数据失败: %w", err)
	}

	_, err = c.doRequest(ctx, "POST", "/api/v1/services", data)
	return err
}

// GetService 获取服务
func (c *Client) GetService(ctx context.Context, serviceName string) (*core.Service, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/services/%s", serviceName), nil)
	if err != nil {
		return nil, err
	}

	var service core.Service
	if err := json.Unmarshal(data, &service); err != nil {
		return nil, fmt.Errorf("解析服务数据失败: %w", err)
	}

	return &service, nil
}

// UpdateService 更新服务
func (c *Client) UpdateService(ctx context.Context, serviceName string, service *core.Service) error {
	// 设置用户ID
	service.EditWho = c.userId

	data, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("序列化服务数据失败: %w", err)
	}

	_, err = c.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/services/%s", serviceName), data)
	return err
}

// DeleteService 删除服务
func (c *Client) DeleteService(ctx context.Context, serviceName string) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/services/%s", serviceName), nil)
	return err
}

// ================== 服务分组管理 ==================

// CreateGroup 创建服务分组
func (c *Client) CreateGroup(ctx context.Context, group *core.ServiceGroup) error {
	// 设置租户ID和用户ID
	group.TenantId = c.tenantId
	group.AddWho = c.userId
	group.EditWho = c.userId

	data, err := json.Marshal(group)
	if err != nil {
		return fmt.Errorf("序列化分组数据失败: %w", err)
	}

	_, err = c.doRequest(ctx, "POST", "/api/v1/groups", data)
	return err
}

// GetGroup 获取服务分组
func (c *Client) GetGroup(ctx context.Context, groupName string) (*core.ServiceGroup, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/groups/%s", groupName), nil)
	if err != nil {
		return nil, err
	}

	var group core.ServiceGroup
	if err := json.Unmarshal(data, &group); err != nil {
		return nil, fmt.Errorf("解析分组数据失败: %w", err)
	}

	return &group, nil
}

// ListGroups 列出服务分组
func (c *Client) ListGroups(ctx context.Context) ([]*core.ServiceGroup, error) {
	data, err := c.doRequest(ctx, "GET", "/api/v1/groups", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Groups []*core.ServiceGroup `json:"groups"`
		Count  int                  `json:"count"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析分组列表数据失败: %w", err)
	}

	return response.Groups, nil
}

// ================== 统计信息 ==================

// GetStats 获取统计信息
func (c *Client) GetStats(ctx context.Context) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "/api/v1/stats", nil)
	if err != nil {
		return nil, err
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("解析统计数据失败: %w", err)
	}

	return stats, nil
}

// GetHealthStats 获取健康检查统计
func (c *Client) GetHealthStats(ctx context.Context) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "/api/v1/stats/health", nil)
	if err != nil {
		return nil, err
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("解析健康统计数据失败: %w", err)
	}

	return stats, nil
}

// ================== 健康检查 ==================

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "/health", nil)
	if err != nil {
		return nil, err
	}

	var health map[string]interface{}
	if err := json.Unmarshal(data, &health); err != nil {
		return nil, fmt.Errorf("解析健康检查数据失败: %w", err)
	}

	return health, nil
}

// ReadinessCheck 就绪检查
func (c *Client) ReadinessCheck(ctx context.Context) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "/ready", nil)
	if err != nil {
		return nil, err
	}

	var ready map[string]interface{}
	if err := json.Unmarshal(data, &ready); err != nil {
		return nil, fmt.Errorf("解析就绪检查数据失败: %w", err)
	}

	return ready, nil
}

// ================== 内部方法 ==================

// doRequest 执行HTTP请求
func (c *Client) doRequest(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Tenant-Id", c.tenantId)
	req.Header.Set("X-User-Id", c.userId)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			if msg, ok := errorResp["error"].(string); ok {
				return nil, fmt.Errorf("API错误 (%d): %s", resp.StatusCode, msg)
			}
		}
		return nil, fmt.Errorf("HTTP错误: %d %s", resp.StatusCode, resp.Status)
	}

	return respBody, nil
}

// SetTimeout 设置请求超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// SetTenantId 设置租户ID
func (c *Client) SetTenantId(tenantId string) {
	c.tenantId = tenantId
}

// SetUserId 设置用户ID
func (c *Client) SetUserId(userId string) {
	c.userId = userId
}
