package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	alertinit "gateway/internal/alert/init"
	"gateway/internal/alert/loader"
	"gateway/internal/alert/queue"
	alerttypes "gateway/internal/types/alerttypes"
	"gateway/pkg/alert"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// AlertManager 告警管理器
// 提供统一的告警系统管理接口，供其他模块调用
// 单例模式，通过 InitGlobalManager 初始化，GetGlobalManager 获取
type AlertManager struct {
	// 数据库连接
	db database.Database
	// 租户ID
	tenantId string
	// 初始化器
	initializer *alertinit.AlertInitializer
	// 全局队列
	alertQueue *queue.AlertQueue
	// 运行状态
	running bool
	mu      sync.RWMutex
}

// newAlertManager 创建告警管理器（私有方法，仅内部使用）
func newAlertManager(db database.Database, tenantId string) *AlertManager {
	return &AlertManager{
		db:          db,
		tenantId:    tenantId,
		initializer: alertinit.NewAlertInitializer(db).SetTenantId(tenantId),
		running:     false,
	}
}

// Start 启动告警管理器
// 初始化告警系统，加载配置，启动队列
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 启动错误
func (m *AlertManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("告警管理器已经在运行")
	}

	logger.InfoWithTrace(ctx, "启动告警管理器", "tenantId", m.tenantId)

	// 初始化告警系统
	alertQueue, success := m.initializer.Initialize(ctx)
	if !success {
		logger.WarnWithTrace(ctx, "告警系统初始化失败或部分失败")
		return fmt.Errorf("告警系统初始化失败")
	}

	// 保存队列引用
	m.alertQueue = alertQueue

	// 设置全局队列
	queue.SetGlobalQueue(alertQueue)

	m.running = true

	// 获取全局管理器检查状态
	channelMgr := alert.GetGlobalManager()
	logger.InfoWithTrace(ctx, "告警管理器启动成功",
		"channels", len(channelMgr.ListChannels()),
		"queue_size", alertQueue.GetQueueSize())

	return nil
}

// Stop 停止告警管理器
// 关闭队列和所有渠道
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 停止错误
func (m *AlertManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("告警管理器未运行")
	}

	logger.InfoWithTrace(ctx, "停止告警管理器")

	// 关闭告警系统
	if err := m.initializer.Shutdown(ctx, m.alertQueue); err != nil {
		logger.ErrorWithTrace(ctx, "关闭告警系统失败", "error", err)
		return err
	}

	m.running = false
	logger.InfoWithTrace(ctx, "告警管理器已停止")
	return nil
}

// IsRunning 检查管理器是否运行中
func (m *AlertManager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// SendAlert 发送告警（异步）
// 推送告警消息到队列，由工作协程异步处理
// 参数:
//   - ctx: 上下文
//   - msg: 告警消息
//
// 返回:
//   - error: 发送错误
func (m *AlertManager) SendAlert(ctx context.Context, msg *queue.AlertMessage) error {
	if !m.IsRunning() {
		return fmt.Errorf("告警管理器未运行")
	}

	if m.alertQueue == nil {
		return fmt.Errorf("告警队列未初始化")
	}

	return m.alertQueue.Push(ctx, msg)
}

// SendAlertSync 同步发送告警
// 直接通过全局管理器发送，立即返回结果
// 参数:
//   - ctx: 上下文
//   - channelId: 渠道ID
//   - message: 告警消息
//   - options: 发送选项
//
// 返回:
//   - *alert.SendResult: 发送结果
func (m *AlertManager) SendAlertSync(ctx context.Context, channelId string, message *alert.Message, options *alert.SendOptions) *alert.SendResult {
	if !m.IsRunning() {
		return &alert.SendResult{
			Success: false,
			Error:   fmt.Errorf("告警管理器未运行"),
		}
	}

	channelMgr := alert.GetGlobalManager()
	return channelMgr.Send(ctx, channelId, message, options)
}

// SendToDefault 发送到默认渠道
// 参数:
//   - ctx: 上下文
//   - message: 告警消息
//   - options: 发送选项
//
// 返回:
//   - *alert.SendResult: 发送结果
func (m *AlertManager) SendToDefault(ctx context.Context, message *alert.Message, options *alert.SendOptions) *alert.SendResult {
	if !m.IsRunning() {
		return &alert.SendResult{
			Success: false,
			Error:   fmt.Errorf("告警管理器未运行"),
		}
	}

	channelMgr := alert.GetGlobalManager()
	return channelMgr.SendToDefault(ctx, message, options)
}

// SendToMultiple 发送到多个渠道
// 参数:
//   - ctx: 上下文
//   - channelIds: 渠道ID列表
//   - message: 告警消息
//   - options: 发送选项
//
// 返回:
//   - map[string]*alert.SendResult: 各渠道发送结果
func (m *AlertManager) SendToMultiple(ctx context.Context, channelIds []string, message *alert.Message, options *alert.SendOptions) map[string]*alert.SendResult {
	if !m.IsRunning() {
		result := make(map[string]*alert.SendResult)
		for _, id := range channelIds {
			result[id] = &alert.SendResult{
				Success: false,
				Error:   fmt.Errorf("告警管理器未运行"),
			}
		}
		return result
	}

	channelMgr := alert.GetGlobalManager()
	return channelMgr.SendToMultiple(ctx, channelIds, message, options)
}

// SendToAll 发送到所有启用的渠道
// 参数:
//   - ctx: 上下文
//   - message: 告警消息
//   - options: 发送选项
//
// 返回:
//   - map[string]*alert.SendResult: 各渠道发送结果
func (m *AlertManager) SendToAll(ctx context.Context, message *alert.Message, options *alert.SendOptions) map[string]*alert.SendResult {
	if !m.IsRunning() {
		return make(map[string]*alert.SendResult)
	}

	channelMgr := alert.GetGlobalManager()
	return channelMgr.SendToAll(ctx, message, options)
}

// ReloadConfig 重新加载配置（热重载）
// 从数据库重新加载渠道配置并更新
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 重载错误
func (m *AlertManager) ReloadConfig(ctx context.Context) error {
	if !m.IsRunning() {
		return fmt.Errorf("告警管理器未运行")
	}

	logger.InfoWithTrace(ctx, "重新加载告警配置")

	if err := m.initializer.ReloadChannels(ctx); err != nil {
		logger.ErrorWithTrace(ctx, "重新加载配置失败", "error", err)
		return err
	}

	logger.InfoWithTrace(ctx, "告警配置重新加载成功")
	return nil
}

// GetChannelList 获取渠道列表
// 返回:
//   - []string: 渠道名称列表
func (m *AlertManager) GetChannelList() []string {
	if !m.IsRunning() {
		return []string{}
	}

	channelMgr := alert.GetGlobalManager()
	return channelMgr.ListChannels()
}

// GetChannelStats 获取所有渠道统计信息
// 返回:
//   - map[string]map[string]interface{}: 各渠道统计信息
func (m *AlertManager) GetChannelStats() map[string]map[string]interface{} {
	if !m.IsRunning() {
		return make(map[string]map[string]interface{})
	}

	channelMgr := alert.GetGlobalManager()
	return channelMgr.Stats()
}

// HealthCheck 对所有渠道进行健康检查
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - map[string]error: 各渠道健康检查结果
func (m *AlertManager) HealthCheck(ctx context.Context) map[string]error {
	if !m.IsRunning() {
		return make(map[string]error)
	}

	channelMgr := alert.GetGlobalManager()
	return channelMgr.HealthCheck(ctx)
}

// GetQueueSize 获取队列大小
// 返回:
//   - int: 队列中待处理的消息数
func (m *AlertManager) GetQueueSize() int {
	if !m.IsRunning() || m.alertQueue == nil {
		return 0
	}

	return m.alertQueue.GetQueueSize()
}

// HasChannel 检查渠道是否存在
// 参数:
//   - channelId: 渠道ID
//
// 返回:
//   - bool: 是否存在
func (m *AlertManager) HasChannel(channelId string) bool {
	if !m.IsRunning() {
		return false
	}

	channelMgr := alert.GetGlobalManager()
	return channelMgr.HasChannel(channelId)
}

// EnableChannel 启用渠道
// 参数:
//   - channelId: 渠道ID
//
// 返回:
//   - error: 操作错误
func (m *AlertManager) EnableChannel(channelId string) error {
	if !m.IsRunning() {
		return fmt.Errorf("告警管理器未运行")
	}

	channelMgr := alert.GetGlobalManager()
	return channelMgr.EnableChannel(channelId)
}

// DisableChannel 禁用渠道
// 参数:
//   - channelId: 渠道ID
//
// 返回:
//   - error: 操作错误
func (m *AlertManager) DisableChannel(channelId string) error {
	if !m.IsRunning() {
		return fmt.Errorf("告警管理器未运行")
	}

	channelMgr := alert.GetGlobalManager()
	return channelMgr.DisableChannel(channelId)
}

// SetDefaultChannel 设置默认渠道
// 参数:
//   - channelId: 渠道ID
//
// 返回:
//   - error: 操作错误
func (m *AlertManager) SetDefaultChannel(channelId string) error {
	if !m.IsRunning() {
		return fmt.Errorf("告警管理器未运行")
	}

	channelMgr := alert.GetGlobalManager()
	return channelMgr.SetDefaultChannel(channelId)
}

// LoadTemplate 加载告警模板
// 参数:
//   - ctx: 上下文
//   - templateId: 模板ID
//
// 返回:
//   - *alerttypes.AlertTemplate: 模板配置
//   - error: 加载错误
func (m *AlertManager) LoadTemplate(ctx context.Context, templateId string) (*alerttypes.AlertTemplate, error) {
	if !m.IsRunning() {
		return nil, fmt.Errorf("告警管理器未运行")
	}

	configLoader := loader.NewAlertConfigLoader(m.db, m.tenantId)
	template, err := configLoader.LoadTemplate(ctx, templateId)
	if err != nil {
		return nil, fmt.Errorf("加载模板失败: %w", err)
	}

	return template, nil
}

// LoadTemplates 加载所有告警模板
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []*alerttypes.AlertTemplate: 模板列表
//   - error: 加载错误
func (m *AlertManager) LoadTemplates(ctx context.Context) ([]*alerttypes.AlertTemplate, error) {
	if !m.IsRunning() {
		return nil, fmt.Errorf("告警管理器未运行")
	}

	configLoader := loader.NewAlertConfigLoader(m.db, m.tenantId)
	templates, err := configLoader.LoadTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("加载模板列表失败: %w", err)
	}

	return templates, nil
}

// GetManagerInfo 获取管理器信息
// 返回:
//   - map[string]interface{}: 管理器信息
func (m *AlertManager) GetManagerInfo() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := map[string]interface{}{
		"running":    m.running,
		"tenant_id":  m.tenantId,
		"queue_size": 0,
		"channels":   []string{},
	}

	if m.running {
		if m.alertQueue != nil {
			info["queue_size"] = m.alertQueue.GetQueueSize()
		}
		channelMgr := alert.GetGlobalManager()
		info["channels"] = channelMgr.ListChannels()
	}

	return info
}

// 全局管理器实例（单例）
var (
	globalAlertManager   *AlertManager
	globalAlertManagerMu sync.RWMutex
)

// InitGlobalManager 初始化全局告警管理器（单例）
// 只能初始化一次，重复调用将返回错误
// 参数:
//   - ctx: 上下文
//   - db: 数据库连接
//   - tenantId: 租户ID
//
// 返回:
//   - error: 初始化错误
func InitGlobalManager(ctx context.Context, db database.Database, tenantId string) error {
	globalAlertManagerMu.Lock()
	defer globalAlertManagerMu.Unlock()

	// 检查是否已初始化
	if globalAlertManager != nil {
		return fmt.Errorf("全局告警管理器已初始化")
	}

	// 创建管理器实例
	mgr := newAlertManager(db, tenantId)

	// 启动管理器
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("启动告警管理器失败: %w", err)
	}

	globalAlertManager = mgr
	logger.InfoWithTrace(ctx, "全局告警管理器初始化成功", "tenantId", tenantId)
	return nil
}

// GetGlobalManager 获取全局告警管理器（单例）
// 返回:
//   - *AlertManager: 全局管理器实例，如果未初始化则返回 nil
func GetGlobalManager() *AlertManager {
	globalAlertManagerMu.RLock()
	defer globalAlertManagerMu.RUnlock()
	return globalAlertManager
}

// StopGlobalManager 停止全局告警管理器
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 停止错误
func StopGlobalManager(ctx context.Context) error {
	globalAlertManagerMu.Lock()
	defer globalAlertManagerMu.Unlock()

	if globalAlertManager == nil {
		return fmt.Errorf("全局告警管理器未初始化")
	}

	if err := globalAlertManager.Stop(ctx); err != nil {
		return err
	}

	// 清空全局实例
	globalAlertManager = nil
	logger.InfoWithTrace(ctx, "全局告警管理器已停止")
	return nil
}

// QuickSend 快速发送告警（便捷方法）
// 参数:
//   - ctx: 上下文
//   - channelIds: 渠道ID（逗号分隔）
//   - title: 标题
//   - content: 内容
//   - alertType: 告警类型
//   - severity: 严重级别
//
// 返回:
//   - error: 发送错误
func QuickSend(ctx context.Context, channelIds, title, content, alertType, severity string) error {
	mgr := GetGlobalManager()
	if mgr == nil {
		return fmt.Errorf("全局告警管理器未初始化")
	}

	msg := &queue.AlertMessage{
		ChannelIds:    channelIds,
		Title:         title,
		Content:       content,
		AlertType:     alertType,
		SeverityLevel: severity,
		TriggerSource: "api",
		ReceivedTime:  time.Now(),
	}

	return mgr.SendAlert(ctx, msg)
}
