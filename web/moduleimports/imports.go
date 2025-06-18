// Package moduleimports 用于统一导入所有模块的routes包
// 这个包作为模块导入的集中点，简化了Web应用中对所有模块的导入管理
// 当项目中新增模块时，只需要在这里添加相应的导入语句即可

package moduleimports

import (
	// 导入所有模块的routes包，这样它们的init函数会被自动执行
	// 每个模块在导入时会通过init函数自动注册自己的路由

	// hub0002模块
	_ "gohub/web/views/hub0002/routes"
	// 添加其他模块的导入
	// _ "gohub/web/views/hub0001/routes"
	// _ "gohub/web/views/hub0003/routes"
	// _ "gohub/web/views/hub0004/routes"
	// 导入认证模块
	_ "gohub/web/views/hub0001/routes"
	// 导入网关管理模块
	_ "gohub/web/views/hub0020/routes"
)

// 这个包没有导出任何函数或变量
// 它的唯一作用是通过init函数在应用启动时自动注册所有模块
