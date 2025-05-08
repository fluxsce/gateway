package routes

import (
	"gohub/pkg/database"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

// Module 模块接口，每个模块需要实现此接口
type Module interface {
	// Name 返回模块名称
	Name() string

	// BasePath 返回模块API基础路径
	BasePath() string

	// RegisterRoutes 注册模块路由
	RegisterRoutes(router *gin.Engine, db database.Database)
}

// ModuleManager 模块管理器
type ModuleManager struct {
	modules []Module
	db      database.Database
}

// NewModuleManager 创建模块管理器
func NewModuleManager(db database.Database) *ModuleManager {
	return &ModuleManager{
		modules: make([]Module, 0),
		db:      db,
	}
}

// Register 注册模块
func (m *ModuleManager) Register(module Module) {
	m.modules = append(m.modules, module)
}

// RegisterAll 注册所有模块的路由
func (m *ModuleManager) RegisterAll(router *gin.Engine) {
	for _, module := range m.modules {
		module.RegisterRoutes(router, m.db)
	}
}

// GetCallerModule 获取调用者的模块名
// 用于自动推断模块名称，便于统一命名
func GetCallerModule() string {
	_, file, _, _ := runtime.Caller(1)
	dir := filepath.Dir(file)
	parts := strings.Split(dir, string(filepath.Separator))

	// 查找views目录后的第一个目录作为模块名
	for i, part := range parts {
		if part == "views" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	// 如果找不到，返回文件所在的最后一个目录
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "unknown"
}
