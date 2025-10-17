// Package moduleimports 用于统一导入所有模块的routes包
// 这个包作为模块导入的集中点，简化了Web应用中对所有模块的导入管理
// 当项目中新增模块时，只需要在这里添加相应的导入语句即可

package moduleimports

import (
	// 导入所有模块的routes包，这样它们的init函数会被自动执行
	// 每个模块在导入时会通过init函数自动注册自己的路由

	// hub002模块 - 安全配置管理模块（位于common目录）
	_ "gateway/web/views/hubcommon002/routes"
	// 添加其他模块的导入
	// _ "gateway/web/views/hub0001/routes"

	// _ "gateway/web/views/hub0004/routes"
	// 导入指标查询模块
	_ "gateway/web/views/hub0000/routes"
	// 导入认证模块
	_ "gateway/web/views/hub0001/routes"
	// 导入用户管理模块
	_ "gateway/web/views/hub0002/routes"
	// 导入定时任务管理模块
	_ "gateway/web/views/hub0003/routes"
	// 导入网关管理模块
	_ "gateway/web/views/hub0020/routes"
	// 导入路由管理模块
	_ "gateway/web/views/hub0021/routes"
	// 导入代理管理模块
	_ "gateway/web/views/hub0022/routes"
	// 导入网关日志管理模块
	_ "gateway/web/views/hub0023/routes"
	// 导入服务分组命名空间管理模块
	_ "gateway/web/views/hub0040/routes"
	// 导入服务注册列表管理模块
	_ "gateway/web/views/hub0041/routes"
	// 导入JVM监控查询模块
	_ "gateway/web/views/hub0042/routes"
	// 导入隧道服务器管理模块
	_ "gateway/web/views/hub0060/routes"
	//导入插件管理模块
	_ "gateway/web/views/hubplugin/routes"
	// 导入SDK服务模块
	_ "gateway/web/sdkservice/regist/routes"
	//导入应用上报资源模块
	_ "gateway/web/sdkservice/reporter/routes"
)

// 这个包没有导出任何函数或变量
// 它的唯一作用是通过init函数在应用启动时自动注册所有模块
