package dao

import (
	"context"
	"fmt"
	"gateway/internal/metric_collect/types"
	"gateway/pkg/database"
	"time"
)

// ServerInfoDAO 服务器信息数据访问对象
type ServerInfoDAO struct {
	db database.Database
}

// NewServerInfoDAO 创建服务器信息DAO实例
func NewServerInfoDAO(db database.Database) *ServerInfoDAO {
	return &ServerInfoDAO{db: db}
}

// InsertServerInfo 插入服务器信息
func (dao *ServerInfoDAO) InsertServerInfo(ctx context.Context, serverInfo *types.ServerInfo) error {
	// 设置通用字段默认值
	now := time.Now()
	serverInfo.AddTime = now
	serverInfo.EditTime = now
	serverInfo.CurrentVersion = 1
	serverInfo.ActiveFlag = types.ActiveFlagYes

	// 验证必填字段
	if serverInfo.MetricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	if serverInfo.TenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if serverInfo.Hostname == "" {
		return fmt.Errorf("主机名不能为空")
	}
	if serverInfo.AddWho == "" {
		return fmt.Errorf("创建人ID不能为空")
	}
	if serverInfo.EditWho == "" {
		return fmt.Errorf("修改人ID不能为空")
	}
	if serverInfo.OprSeqFlag == "" {
		return fmt.Errorf("操作序列标识不能为空")
	}

	tableName := serverInfo.TableName()
	_, err := dao.db.Insert(ctx, tableName, serverInfo, true)
	if err != nil {
		return fmt.Errorf("插入服务器信息失败: %w", err)
	}
	return nil
}

// BatchInsertServerInfo 批量插入服务器信息
func (dao *ServerInfoDAO) BatchInsertServerInfo(ctx context.Context, serverInfos []*types.ServerInfo) error {
	if len(serverInfos) == 0 {
		return nil
	}

	// 设置通用字段默认值
	now := time.Now()
	for _, serverInfo := range serverInfos {
		serverInfo.AddTime = now
		serverInfo.EditTime = now
		serverInfo.CurrentVersion = 1
		serverInfo.ActiveFlag = types.ActiveFlagYes

		// 验证必填字段
		if serverInfo.MetricServerId == "" {
			return fmt.Errorf("服务器ID不能为空")
		}
		if serverInfo.TenantId == "" {
			return fmt.Errorf("租户ID不能为空")
		}
		if serverInfo.Hostname == "" {
			return fmt.Errorf("主机名不能为空")
		}
		if serverInfo.AddWho == "" {
			return fmt.Errorf("创建人ID不能为空")
		}
		if serverInfo.EditWho == "" {
			return fmt.Errorf("修改人ID不能为空")
		}
		if serverInfo.OprSeqFlag == "" {
			return fmt.Errorf("操作序列标识不能为空")
		}
	}

	tableName := serverInfos[0].TableName()

	// 转换为interface{}切片
	items := make([]interface{}, len(serverInfos))
	for i, serverInfo := range serverInfos {
		items[i] = serverInfo
	}

	_, err := dao.db.BatchInsert(ctx, tableName, items, true)
	if err != nil {
		return fmt.Errorf("批量插入服务器信息失败: %w", err)
	}
	return nil
}

// DeleteServerInfo 删除服务器信息（软删除）
func (dao *ServerInfoDAO) DeleteServerInfo(ctx context.Context, tenantId, metricServerId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}

	tableName := (&types.ServerInfo{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = ? WHERE tenantId = ? AND metricServerId = ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{types.ActiveFlagNo, tenantId, metricServerId}, true)
	if err != nil {
		return fmt.Errorf("删除服务器信息失败: %w", err)
	}
	return nil
}

// DeleteServerInfoByTime 根据时间范围删除服务器信息（物理删除）
func (dao *ServerInfoDAO) DeleteServerInfoByTime(ctx context.Context, tenantId string, beforeTime time.Time) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}

	tableName := (&types.ServerInfo{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND addTime < ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, beforeTime}, true)
	if err != nil {
		return fmt.Errorf("根据时间删除服务器信息失败: %w", err)
	}
	return nil
}

// GetServerInfoById 根据ID获取服务器信息
func (dao *ServerInfoDAO) GetServerInfoById(ctx context.Context, tenantId, metricServerId string) (*types.ServerInfo, error) {
	if tenantId == "" {
		return nil, fmt.Errorf("租户ID不能为空")
	}
	if metricServerId == "" {
		return nil, fmt.Errorf("服务器ID不能为空")
	}

	tableName := (&types.ServerInfo{}).TableName()
	sql := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND metricServerId = ? AND activeFlag = ?", tableName)

	var serverInfo types.ServerInfo
	err := dao.db.QueryOne(ctx, &serverInfo, sql, []interface{}{tenantId, metricServerId, types.ActiveFlagYes}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询服务器信息失败: %w", err)
	}

	return &serverInfo, nil
}

// GetServerInfoByHostname 根据主机名获取服务器信息
func (dao *ServerInfoDAO) GetServerInfoByHostname(ctx context.Context, tenantId, hostname string) (*types.ServerInfo, error) {
	if tenantId == "" {
		return nil, fmt.Errorf("租户ID不能为空")
	}
	if hostname == "" {
		return nil, fmt.Errorf("主机名不能为空")
	}

	tableName := (&types.ServerInfo{}).TableName()
	sql := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND hostname = ? AND activeFlag = ?", tableName)

	var serverInfo types.ServerInfo
	err := dao.db.QueryOne(ctx, &serverInfo, sql, []interface{}{tenantId, hostname, types.ActiveFlagYes}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询服务器信息失败: %w", err)
	}

	return &serverInfo, nil
}

// GetServerInfoByHostnameAndNetwork 根据主机名、IP和MAC地址获取服务器信息
// 通过主机名、IP地址和MAC地址的组合来唯一确定一台服务器
func (dao *ServerInfoDAO) GetServerInfoByHostnameAndNetwork(ctx context.Context, tenantId, hostname, ipAddress, macAddress string) (*types.ServerInfo, error) {
	if tenantId == "" {
		return nil, fmt.Errorf("租户ID不能为空")
	}
	if hostname == "" {
		return nil, fmt.Errorf("主机名不能为空")
	}

	tableName := (&types.ServerInfo{}).TableName()

	// 构建查询条件
	sql := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND hostname = ? AND activeFlag = ?", tableName)
	params := []interface{}{tenantId, hostname, types.ActiveFlagYes}

	// 如果提供了IP地址，添加IP条件
	if ipAddress != "" {
		sql += " AND ipAddress = ?"
		params = append(params, ipAddress)
	}

	// 如果提供了MAC地址，添加MAC条件
	if macAddress != "" {
		sql += " AND macAddress = ?"
		params = append(params, macAddress)
	}

	var serverInfo types.ServerInfo
	err := dao.db.QueryOne(ctx, &serverInfo, sql, params, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询服务器信息失败: %w", err)
	}

	return &serverInfo, nil
}

// UpdateServerInfo 更新服务器信息
func (dao *ServerInfoDAO) UpdateServerInfo(ctx context.Context, serverInfo *types.ServerInfo) error {
	if serverInfo.TenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if serverInfo.MetricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}

	// 更新时间和版本
	now := time.Now()
	serverInfo.EditTime = now
	serverInfo.LastUpdateTime = now
	serverInfo.CurrentVersion++

	tableName := serverInfo.TableName()
	sql := fmt.Sprintf(`UPDATE %s SET 
		hostname = ?, osType = ?, osVersion = ?, kernelVersion = ?, architecture = ?, 
		bootTime = ?, ipAddress = ?, macAddress = ?, serverLocation = ?, serverType = ?, 
		lastUpdateTime = ?, editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?
		WHERE tenantId = ? AND metricServerId = ?`, tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		serverInfo.Hostname, serverInfo.OsType, serverInfo.OsVersion, serverInfo.KernelVersion, serverInfo.Architecture,
		serverInfo.BootTime, serverInfo.IpAddress, serverInfo.MacAddress, serverInfo.ServerLocation, serverInfo.ServerType,
		serverInfo.LastUpdateTime, serverInfo.EditTime, serverInfo.EditWho, serverInfo.OprSeqFlag, serverInfo.CurrentVersion,
		serverInfo.TenantId, serverInfo.MetricServerId,
	}, true)
	if err != nil {
		return fmt.Errorf("更新服务器信息失败: %w", err)
	}
	return nil
}
