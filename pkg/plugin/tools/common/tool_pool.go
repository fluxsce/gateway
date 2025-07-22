package common

import (
	"fmt"
	"sync"
	"time"

	"gateway/pkg/plugin/tools/interfaces"
)

// ToolPool 工具池
// 简单的工具池实现，用于管理各种类型的工具实例
// 全局唯一实例，确保资源统一管理
type ToolPool struct {
	// 工具存储映射：key = toolID, value = Tool
	tools map[string]interfaces.Tool

	// 工具类型映射：key = toolID, value = toolType
	toolTypes map[string]string

	// 读写锁，保护并发访问
	mu sync.RWMutex
}

// 全局工具池实例
var (
	globalToolPool     *ToolPool
	globalToolPoolOnce sync.Once
)

// GetGlobalToolPool 获取全局工具池实例
// 使用单例模式确保全局唯一性
func GetGlobalToolPool() *ToolPool {
	globalToolPoolOnce.Do(func() {
		globalToolPool = NewToolPool()
	})
	return globalToolPool
}

// NewToolPool 创建新的工具池实例
func NewToolPool() *ToolPool {
	return &ToolPool{
		tools:     make(map[string]interfaces.Tool),
		toolTypes: make(map[string]string),
	}
}

// AddTool 添加工具到池中
func (p *ToolPool) AddTool(tool interfaces.Tool) error {
	if tool == nil {
		return fmt.Errorf("工具不能为nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	toolID := tool.GetID()
	toolType := tool.GetType()

	// 检查工具ID是否已存在
	if _, exists := p.tools[toolID]; exists {
		return fmt.Errorf("工具ID已存在: %s", toolID)
	}

	// 添加到主映射
	p.tools[toolID] = tool

	// 添加到类型映射
	p.toolTypes[toolID] = toolType

	return nil
}

// GetTool 根据ID获取工具
func (p *ToolPool) GetTool(toolID string) (interfaces.Tool, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	tool, exists := p.tools[toolID]
	return tool, exists
}

// GetSFTPTool 根据ID获取SFTP工具
func (p *ToolPool) GetSFTPTool(toolID string) (interfaces.SFTPTool, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	tool, exists := p.tools[toolID]
	if !exists {
		return nil, false
	}

	if sftpTool, ok := tool.(interfaces.SFTPTool); ok {
		return sftpTool, true
	}

	return nil, false
}

// GetSSHTool 根据ID获取SSH工具
func (p *ToolPool) GetSSHTool(toolID string) (interfaces.SSHTool, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	tool, exists := p.tools[toolID]
	if !exists {
		return nil, false
	}

	if sshTool, ok := tool.(interfaces.SSHTool); ok {
		return sshTool, true
	}

	return nil, false
}

// GetFTPTool 根据ID获取FTP工具
func (p *ToolPool) GetFTPTool(toolID string) (interfaces.FTPTool, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	tool, exists := p.tools[toolID]
	if !exists {
		return nil, false
	}

	if ftpTool, ok := tool.(interfaces.FTPTool); ok {
		return ftpTool, true
	}

	return nil, false
}

// GetToolsByType 根据类型获取工具列表
func (p *ToolPool) GetToolsByType(toolType string) []interfaces.Tool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var tools []interfaces.Tool
	for toolID, tType := range p.toolTypes {
		if tType == toolType {
			if tool, exists := p.tools[toolID]; exists {
				tools = append(tools, tool)
			}
		}
	}

	return tools
}

// GetSFTPTools 获取所有SFTP工具
func (p *ToolPool) GetSFTPTools() []interfaces.SFTPTool {
	tools := p.GetToolsByType("sftp")
	var sftpTools []interfaces.SFTPTool

	for _, tool := range tools {
		if sftpTool, ok := tool.(interfaces.SFTPTool); ok {
			sftpTools = append(sftpTools, sftpTool)
		}
	}

	return sftpTools
}

// GetSSHTools 获取所有SSH工具
func (p *ToolPool) GetSSHTools() []interfaces.SSHTool {
	tools := p.GetToolsByType("ssh")
	var sshTools []interfaces.SSHTool

	for _, tool := range tools {
		if sshTool, ok := tool.(interfaces.SSHTool); ok {
			sshTools = append(sshTools, sshTool)
		}
	}

	return sshTools
}

// GetFTPTools 获取所有FTP工具
func (p *ToolPool) GetFTPTools() []interfaces.FTPTool {
	tools := p.GetToolsByType("ftp")
	var ftpTools []interfaces.FTPTool

	for _, tool := range tools {
		if ftpTool, ok := tool.(interfaces.FTPTool); ok {
			ftpTools = append(ftpTools, ftpTool)
		}
	}

	return ftpTools
}

// RemoveTool 从池中移除工具
func (p *ToolPool) RemoveTool(toolID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	tool, exists := p.tools[toolID]
	if !exists {
		return fmt.Errorf("工具不存在: %s", toolID)
	}

	// 关闭工具
	if err := tool.Close(); err != nil {
		return fmt.Errorf("关闭工具失败: %w", err)
	}

	// 从主映射中移除
	delete(p.tools, toolID)

	// 从类型映射中移除
	delete(p.toolTypes, toolID)

	return nil
}

// GetAllTools 获取所有工具
func (p *ToolPool) GetAllTools() map[string]interfaces.Tool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 返回副本以避免并发修改
	result := make(map[string]interfaces.Tool)
	for id, tool := range p.tools {
		result[id] = tool
	}
	return result
}

// GetToolCount 获取工具总数
func (p *ToolPool) GetToolCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.tools)
}

// GetToolCountByType 获取指定类型的工具数量
func (p *ToolPool) GetToolCountByType(toolType string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := 0
	for _, tType := range p.toolTypes {
		if tType == toolType {
			count++
		}
	}
	return count
}

// CloseAllTools 关闭所有工具
func (p *ToolPool) CloseAllTools() []error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errors []error
	for toolID, tool := range p.tools {
		if err := tool.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭工具 %s 失败: %w", toolID, err))
		}
	}

	// 清空映射
	p.tools = make(map[string]interfaces.Tool)
	p.toolTypes = make(map[string]string)

	return errors
}

// CloseToolsByType 关闭指定类型的所有工具
func (p *ToolPool) CloseToolsByType(toolType string) []error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errors []error
	var toRemove []string

	// 找到需要关闭的工具
	for toolID, tType := range p.toolTypes {
		if tType == toolType {
			toRemove = append(toRemove, toolID)
		}
	}

	// 关闭并移除工具
	for _, toolID := range toRemove {
		if tool, exists := p.tools[toolID]; exists {
			if err := tool.Close(); err != nil {
				errors = append(errors, fmt.Errorf("关闭工具 %s 失败: %w", toolID, err))
			}
			delete(p.tools, toolID)
			delete(p.toolTypes, toolID)
		}
	}

	return errors
}

// GetStats 获取工具池统计信息
func (p *ToolPool) GetStats() *ToolPoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := &ToolPoolStats{
		TotalTools:  len(p.tools),
		ActiveTools: 0,
		ToolsByType: make(map[string]int),
		LastUpdated: time.Now(),
	}

	// 统计活跃工具和类型分布
	for toolID, tool := range p.tools {
		if tool.IsActive() {
			stats.ActiveTools++
		}

		if toolType, exists := p.toolTypes[toolID]; exists {
			stats.ToolsByType[toolType]++
		}
	}

	return stats
}

// ToolPoolStats 工具池统计信息
type ToolPoolStats struct {
	// 总工具数量
	TotalTools int `json:"total_tools"`

	// 活跃工具数量
	ActiveTools int `json:"active_tools"`

	// 按类型分组的工具数量
	ToolsByType map[string]int `json:"tools_by_type"`

	// 最后更新时间
	LastUpdated time.Time `json:"last_updated"`
}

// String 返回统计信息的字符串表示
func (s *ToolPoolStats) String() string {
	return fmt.Sprintf("总工具: %d, 活跃: %d, 类型分布: %v",
		s.TotalTools, s.ActiveTools, s.ToolsByType)
}
