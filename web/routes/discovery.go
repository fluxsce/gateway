package routes

import (
	"fmt"
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/gin-gonic/gin"
)

// RouteDiscovery 路由发现器
// 负责扫描项目目录结构并自动发现可用的模块和路由
// 通过约定优于配置的方式，减少手动注册路由的工作量
type RouteDiscovery struct {
	baseDir string            // 项目根目录
	db      database.Database // 数据库连接
}

// NewRouteDiscovery 创建路由发现器
// 参数:
//   - baseDir: 项目根目录路径
//   - db: 数据库连接实例
//
// 返回:
//   - *RouteDiscovery: 路由发现器实例
func NewRouteDiscovery(baseDir string, db database.Database) *RouteDiscovery {
	return &RouteDiscovery{
		baseDir: baseDir,
		db:      db,
	}
}

// DiscoverModules 发现所有模块
// 通过扫描views目录下的以"hub"开头的子目录来自动发现可用模块
// 每个子目录被视为一个独立模块，如果该目录下存在routes子目录，则认为是一个有效模块
// 返回:
//   - []Module: 发现的模块列表
func (rd *RouteDiscovery) DiscoverModules() []Module {
	modules := make([]Module, 0)

	// 查找views目录下的所有子目录
	viewsDir := filepath.Join(rd.baseDir, "views")
	items, err := os.ReadDir(viewsDir)
	if err != nil {
		logger.Error("无法读取views目录", err)
		return modules
	}

	// 遍历views目录下的所有项
	for _, item := range items {
		// 只处理目录且名称以"hub"开头的项
		if item.IsDir() && strings.HasPrefix(item.Name(), "hub") {
			moduleName := item.Name()

			// 创建模块实例
			module := &StandardModule{
				name:     moduleName,                                    // 模块名称
				basePath: "/api/" + moduleName,                          // API基础路径
				routeDir: filepath.Join(viewsDir, moduleName, "routes"), // 路由目录
				db:       rd.db,                                         // 数据库连接
			}

			// 检查模块是否有routes目录，只有存在routes目录的才是有效模块
			if _, err := os.Stat(module.routeDir); !os.IsNotExist(err) {
				// 尝试自动发现并注册路由函数
				if initFunc := rd.discoverModuleInitFunc(moduleName, module.routeDir); initFunc != nil {
					RegisterModuleRoutes(moduleName, initFunc)
				}

				modules = append(modules, module)
				logger.Info("发现模块", "name", moduleName, "path", module.routeDir)
			}
		}
	}

	return modules
}

// discoverModuleInitFunc 发现模块的路由初始化函数
// 尝试通过以下方式找到模块的路由初始化函数：
// 1. 尝试导入模块的routes包
// 2. 查找并调用RegisterRoutesFunc函数
// 参数:
//   - moduleName: 模块名称
//   - routeDir: 路由目录路径
//
// 返回:
//   - RouteInitFunc: 找到的路由初始化函数，如果未找到则返回nil
func (rd *RouteDiscovery) discoverModuleInitFunc(moduleName string, routeDir string) RouteInitFunc {
	// 由于Go的编译机制限制，我们无法在运行时动态加载Go代码
	// 以下是通过反射尝试在已编译的程序中找到对应的函数

	// 假设所有模块都按照约定导入，使用预定义的包路径
	// 例如：web/views/hub0001/routes, web/views/hub0002/routes 等
	// 我们可以尝试使用反射来查找这些包中的RegisterRoutesFunc函数

	// 1. 通过Module包名查找到对应的路由包
	// 2. 在路由包中查找RegisterRoutesFunc函数
	// 3. 调用该函数获取路由初始化函数

	// 注意：这里只是通过约定查找，真正的动态加载需要使用插件机制

	// 尝试查找已注册的路由初始化函数
	if initFunc := getRouteInitFunc(moduleName); initFunc != nil {
		return initFunc
	}

	// 搜索是否有同名module_routes.go文件
	moduleRoutesFile := filepath.Join(routeDir, "module_routes.go")
	if _, err := os.Stat(moduleRoutesFile); os.IsNotExist(err) {
		logger.Debug("模块没有module_routes.go文件", "module", moduleName, "path", moduleRoutesFile)
		return nil
	}

	logger.Info("找到模块路由文件", "module", moduleName, "path", moduleRoutesFile)

	// 由于Go的限制，我们无法在运行时动态加载Go代码
	// 这里我们需要预先了解所有可能的模块包路径并导入它们
	// 在初始化阶段将自动调用init函数注册路由

	// 使用一个约定，所有模块初始化函数都会自动注册
	// 在每个模块的init函数中调用RegisterModuleRoutes

	// 在实际使用时，可以通过确保所有模块都被导入来解决这个问题
	// 比如在main.go中添加 _ "gohub/web/views/hub0001/routes" 等导入语句

	return nil
}

// StandardModule 标准模块实现
// 实现了Module接口，代表一个标准的模块
type StandardModule struct {
	name     string            // 模块名称
	basePath string            // API基础路径
	routeDir string            // 路由目录
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
// 根据预定义的路由注册函数或约定式路由自动注册模块路由
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func (m *StandardModule) RegisterRoutes(router *gin.Engine, db database.Database) {
	// 检查路由目录是否存在
	if _, err := os.Stat(m.routeDir); os.IsNotExist(err) {
		logger.Warn("模块的routes目录不存在", "module", m.name)
		return
	}

	// 查找并执行特定模块的Init函数或RegisterXXXRoutes函数
	// 由于Go的限制，我们无法直接动态加载Go源代码
	// 我们需要在编译时已经知道所有可能的模块
	// 这里提供几种可能的解决方案：

	// 方案1: 使用预定义的路由注册表 (最简单的方法)
	// 尝试从预定义的路由注册函数表中获取当前模块的注册函数
	if initFunc := getRouteInitFunc(m.name); initFunc != nil {
		logger.Info("使用预定义函数注册模块路由", "module", m.name)
		initFunc(router, db)
		return
	}

	// 方案2: 约定式路由 (基于目录结构和命名约定自动生成路由)
	// 如果没有找到预定义的注册函数，使用约定式路由
	logger.Info("使用约定式路由注册模块", "module", m.name)

	// 创建模块API路由组
	apiGroup := router.Group(m.basePath)

	// 扫描控制器目录，为每个控制器生成路由
	controllersDir := filepath.Join(filepath.Dir(m.routeDir), "controllers")
	if _, err := os.Stat(controllersDir); !os.IsNotExist(err) {
		registerControllersByConvention(apiGroup, controllersDir, db)
	}

	logger.Info("模块路由注册完成", "module", m.name)
}

// registerControllersByConvention 基于约定注册控制器路由
// 根据控制器文件名称自动生成路由
// 约定: 控制器文件名必须以_controller.go结尾
// 参数:
//   - group: 路由组
//   - controllersDir: 控制器目录路径
//   - db: 数据库连接
func registerControllersByConvention(group *gin.RouterGroup, controllersDir string, db database.Database) {
	// 读取控制器目录下的所有Go文件
	files, err := os.ReadDir(controllersDir)
	if err != nil {
		logger.Error("无法读取控制器目录", err, "path", controllersDir)
		return
	}

	// 遍历所有文件
	for _, file := range files {
		// 只处理非目录且文件名以_controller.go结尾的文件
		if !file.IsDir() && strings.HasSuffix(file.Name(), "_controller.go") {
			// 从文件名获取资源名称
			resource := strings.TrimSuffix(file.Name(), "_controller.go")
			logger.Debug("发现控制器", "resource", resource)

			// 在实际项目中，这里应该查找控制器并注册路由
			// 由于我们无法动态加载Go代码，这里只是一个示例
			logger.Debug("为资源创建路由组", "resource", resource, "path", "/"+resource)
		}
	}
}

// RouteInitFunc 路由注册函数类型
// 定义了模块路由注册函数的签名
type RouteInitFunc func(*gin.Engine, database.Database)

// 预定义的路由注册函数映射表
// 在实际项目中，这可以通过代码生成工具自动生成
// 键: 模块名称
// 值: 对应的路由注册函数
var routeInitFuncs = map[string]RouteInitFunc{
	// 在此处添加已知模块的路由注册函数
	// 例如: "hub0001": hub0001routes.Init,
	//       "hub0002": hub0002routes.Init,
}

// getRouteInitFunc 获取指定模块的路由注册函数
// 参数:
//   - moduleName: 模块名称
//
// 返回:
//   - RouteInitFunc: 路由注册函数，如果不存在则返回nil
func getRouteInitFunc(moduleName string) RouteInitFunc {
	return routeInitFuncs[moduleName]
}

// RegisterModuleRoutes 注册特定模块的路由
// 这个函数是一个辅助函数，用于在routeInitFuncs映射表中注册函数
// 参数:
//   - moduleName: 模块名称
//   - initFunc: 路由注册函数
func RegisterModuleRoutes(moduleName string, initFunc RouteInitFunc) {
	routeInitFuncs[moduleName] = initFunc
	logger.Info("已注册模块的路由初始化函数", "module", moduleName)
}

// LoadPlugin 加载插件
// 通过Go的plugin机制动态加载模块插件
// 参数:
//   - pluginPath: 插件文件路径
//
// 返回:
//   - Module: 加载的模块实例
//   - error: 错误信息
func LoadPlugin(pluginPath string) (Module, error) {
	// 打开插件
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("无法加载插件 %s: %v", pluginPath, err)
	}

	// 查找插件中的NewModule函数
	newModuleSym, err := p.Lookup("NewModule")
	if err != nil {
		return nil, fmt.Errorf("插件中未找到 NewModule 函数: %v", err)
	}

	// 将符号转换为函数
	newModule, ok := newModuleSym.(func() Module)
	if !ok {
		return nil, fmt.Errorf("插件中的 NewModule 不是预期的函数类型")
	}

	// 调用函数获取模块
	return newModule(), nil
}
