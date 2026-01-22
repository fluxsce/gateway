package types

import (
	"context"
)

// AlertService 告警服务接口
// 核心职责：发送告警并写入日志
//
// 说明：
// - 此接口只提供告警发送功能
// - 配置查询、日志查询、模板查询等功能应通过 DAO 层直接使用
// - 查询方式：
//   - 配置查询：dao.NewConfigDAO(db).GetConfig() / ListConfigs()
//   - 日志查询：dao.NewLogDAO(db).GetLog()
//   - 模板查询：dao.NewTemplateDAO(db).GetTemplate()
type AlertService interface {
	// Start 启动告警服务
	// 启动后台 worker 线程：日志写入、发送处理、清理
	Start(ctx context.Context) error

	// Stop 停止告警服务
	// 优雅停止所有后台 worker 线程
	Stop(ctx context.Context) error

	// SendAlert 发送告警
	// 功能流程：
	//   1. 异步写入队列（非阻塞）
	//   2. 后台 worker 从队列读取并写入数据库（状态PENDING）
	//   3. 后台 worker 轮询数据库获取 PENDING 日志并发送
	//   4. 更新 HUB_ALERT_CONFIG 表的统计信息
	//
	// 参数:
	//   ctx: 上下文对象
	//   level: 告警级别（INFO/WARN/ERROR/CRITICAL）
	//   alertType: 告警类型，业务自定义类型标识
	//   title: 告警标题
	//   content: 告警内容
	//   channelName: 渠道名称，为空则使用默认渠道
	//   tags: 告警标签（可选）
	//   extra: 额外数据（可选）
	//   tableData: 表格数据（可选）
	// 返回:
	//   string: 告警日志ID（可用于后续查询）
	//   error: 错误信息
	SendAlert(ctx context.Context, level, alertType, title, content, channelName string, tags map[string]string, extra map[string]interface{}, tableData map[string]interface{}) (string, error)
}
