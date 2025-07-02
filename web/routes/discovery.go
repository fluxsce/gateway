package routes

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"

	"github.com/gin-gonic/gin"
)

// RouteInitFunc 路由注册函数类型
// 定义了模块路由注册函数的签名
type RouteInitFunc func(*gin.Engine, database.Database)

// ModuleInfo 模块信息
type ModuleInfo struct {
	Name     string        // 模块名称
	BasePath string        // API基础路径
	InitFunc RouteInitFunc // 路由初始化函数
}

// 预定义的路由注册函数映射表
// 在编译时通过各模块的init函数自动填充
var registeredModules = make(map[string]*ModuleInfo)

// RegisterModuleRoutes 注册特定模块的路由
// 这个函数在各模块的init函数中被调用
// 参数:
//   - moduleName: 模块名称
//   - initFunc: 路由注册函数
func RegisterModuleRoutes(moduleName string, initFunc RouteInitFunc) {
	// 根据模块名称推断API基础路径
	basePath := "/gohub/" + moduleName
	
	registeredModules[moduleName] = &ModuleInfo{
		Name:     moduleName,
		BasePath: basePath,
		InitFunc: initFunc,
	}
	
	logger.Info("已注册模块的路由初始化函数", "module", moduleName, "basePath", basePath)
}

// RouteDiscovery 路由发现器
// 负责管理已注册的模块路由
type RouteDiscovery struct {
	db database.Database // 数据库连接
}

// NewRouteDiscovery 创建路由发现器
// 参数:
//   - baseDir: 项目根目录路径（保留兼容性，但不再使用）
//   - db: 数据库连接实例
//
// 返回:
//   - *RouteDiscovery: 路由发现器实例
func NewRouteDiscovery(baseDir string, db database.Database) *RouteDiscovery {
	return &RouteDiscovery{
		db: db,
	}
}

// DiscoverModules 发现所有已注册的模块
// 返回编译时注册的所有模块
// 返回:
//   - []Module: 已注册的模块列表
func (rd *RouteDiscovery) DiscoverModules() []Module {
	modules := make([]Module, 0, len(registeredModules))

	// 遍历所有已注册的模块
	for _, moduleInfo := range registeredModules {
		// 创建标准模块实例
		module := &StandardModule{
			name:     moduleInfo.Name,
			basePath: moduleInfo.BasePath,
			initFunc: moduleInfo.InitFunc,
			db:       rd.db,
		}

		modules = append(modules, module)
		logger.Info("发现已注册模块", "name", moduleInfo.Name, "basePath", moduleInfo.BasePath)
	}

	logger.Info("模块发现完成", "总数", len(modules))
	return modules
}

// GetRegisteredModules 获取所有已注册的模块信息
// 返回:
//   - map[string]*ModuleInfo: 已注册的模块信息映射
func GetRegisteredModules() map[string]*ModuleInfo {
	// 返回副本以避免外部修改
	result := make(map[string]*ModuleInfo)
	for k, v := range registeredModules {
		result[k] = v
	}
	return result
}

// GetModuleInfo 获取指定模块的信息
// 参数:
//   - moduleName: 模块名称
// 返回:
//   - *ModuleInfo: 模块信息，如果不存在则返回nil
func GetModuleInfo(moduleName string) *ModuleInfo {
	return registeredModules[moduleName]
}

// StandardModule 标准模块实现
// 实现了Module接口，代表一个标准的模块
type StandardModule struct {
	name     string            // 模块名称
	basePath string            // API基础路径
	initFunc RouteInitFunc     // 路由初始化函数
	db       database.Database // 数据库连接
}

// Name 返回模块名称
// 返回:
//   - string: 模块名称
func (m *StandardModule) Name() string {
	return m.name
}

// BasePath 返回模块API基础路径
// 返回:
//   - string: 模块API基础路径
func (m *StandardModule) BasePath() string {
	return m.basePath
}

// RegisterRoutes 注册模块路由
// 使用预注册的路由初始化函数注册模块路由
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func (m *StandardModule) RegisterRoutes(router *gin.Engine, db database.Database) {
	if m.initFunc == nil {
		logger.Warn("模块没有路由初始化函数", "module", m.name)
		return
	}

	logger.Info("注册模块路由", "module", m.name, "basePath", m.basePath)
	m.initFunc(router, db)
	logger.Info("模块路由注册完成", "module", m.name)
}
