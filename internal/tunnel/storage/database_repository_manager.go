// Package storage 实现隧道管理系统的数据存储
// 基于Database接口的Repository模式实现
package storage

import (
	"gateway/pkg/database"
)

// DatabaseRepositoryManager 基于数据库的存储管理器实现
// 提供统一的数据库访问接口，管理所有隧道相关的Repository
type DatabaseRepositoryManager struct {
	db                   database.Database
	tunnelServerRepo     TunnelServerRepository
	tunnelServerNodeRepo TunnelServerNodeRepository
	tunnelClientRepo     TunnelClientRepository
	tunnelServiceRepo    TunnelServiceRepository
	tunnelSessionRepo    TunnelSessionRepository
	tunnelConnectionRepo TunnelConnectionRepository
}

// NewDatabaseRepositoryManager 创建新的数据库存储管理器
//
// 参数:
//   - db: 数据库连接接口
//
// 返回:
//   - RepositoryManager: 存储管理器接口实例
func NewDatabaseRepositoryManager(db database.Database) RepositoryManager {
	manager := &DatabaseRepositoryManager{
		db: db,
	}

	// 初始化各个Repository
	manager.tunnelServerRepo = NewTunnelServerRepository(db)
	manager.tunnelServerNodeRepo = NewTunnelServerNodeRepository(db)
	manager.tunnelClientRepo = NewTunnelClientRepository(db)
	manager.tunnelServiceRepo = NewTunnelServiceRepository(db)
	manager.tunnelSessionRepo = NewTunnelSessionRepository(db)
	manager.tunnelConnectionRepo = NewTunnelConnectionRepository(db)

	return manager
}

// GetTunnelServerRepository 获取隧道服务器存储接口
func (m *DatabaseRepositoryManager) GetTunnelServerRepository() TunnelServerRepository {
	return m.tunnelServerRepo
}

// GetTunnelServerNodeRepository 获取隧道服务器节点存储接口
func (m *DatabaseRepositoryManager) GetTunnelServerNodeRepository() TunnelServerNodeRepository {
	return m.tunnelServerNodeRepo
}

// GetTunnelClientRepository 获取隧道客户端存储接口
func (m *DatabaseRepositoryManager) GetTunnelClientRepository() TunnelClientRepository {
	return m.tunnelClientRepo
}

// GetTunnelServiceRepository 获取隧道服务存储接口
func (m *DatabaseRepositoryManager) GetTunnelServiceRepository() TunnelServiceRepository {
	return m.tunnelServiceRepo
}

// GetTunnelSessionRepository 获取隧道会话存储接口
func (m *DatabaseRepositoryManager) GetTunnelSessionRepository() TunnelSessionRepository {
	return m.tunnelSessionRepo
}

// GetTunnelConnectionRepository 获取隧道连接存储接口
func (m *DatabaseRepositoryManager) GetTunnelConnectionRepository() TunnelConnectionRepository {
	return m.tunnelConnectionRepo
}

// Close 关闭存储管理器
// 释放数据库连接等资源
func (m *DatabaseRepositoryManager) Close() error {
	if closer, ok := m.db.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}
