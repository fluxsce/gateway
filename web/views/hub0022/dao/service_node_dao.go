package dao

import (
	"context"
	"errors"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0022/models"
	"strings"
	"time"
)

// ServiceNodeDAO 服务节点数据访问对象
type ServiceNodeDAO struct {
	db database.Database
}

// NewServiceNodeDAO 创建服务节点DAO
func NewServiceNodeDAO(db database.Database) *ServiceNodeDAO {
	return &ServiceNodeDAO{
		db: db,
	}
}

// QueryServiceNodes 分页查询服务节点列表
func (dao *ServiceNodeDAO) QueryServiceNodes(ctx context.Context, tenantId string, page, pageSize int, filters map[string]interface{}) ([]*models.ServiceNodeModel, int, error) {
	// 构建查询条件
	whereClause := "WHERE tenantId = ?"
	params := []interface{}{tenantId}

	// 添加筛选条件
	for key, value := range filters {
		// 跳过nodeEnabled字段，数据库不再维护
		if key == "nodeEnabled" {
			continue
		}

		// 对于字符串类型的值，使用empty.IsNotEmpty判断
		if strValue, ok := value.(string); ok {
			if !empty.IsEmpty(strValue) {
				if key == "nodeHost" {
					// 支持模糊查询
					whereClause += fmt.Sprintf(" AND %s LIKE ?", key)
					params = append(params, "%"+strValue+"%")
				} else {
					whereClause += fmt.Sprintf(" AND %s = ?", key)
					params = append(params, strValue)
				}
			}
		} else if value != nil {
			// 非字符串类型的值，只要不为nil就添加条件
			whereClause += fmt.Sprintf(" AND %s = ?", key)
			params = append(params, value)
		}
	}

	// 构建基础查询语句
	baseQuery := fmt.Sprintf("SELECT * FROM HUB_GW_SERVICE_NODE %s ORDER BY addTime DESC", whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务节点总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.ServiceNodeModel{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	// 执行分页查询
	var nodes []*models.ServiceNodeModel
	err = dao.db.Query(ctx, &nodes, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务节点列表失败")
	}

	return nodes, total, nil
}

// GetServiceNodeById 根据ID获取服务节点
func (dao *ServiceNodeDAO) GetServiceNodeById(ctx context.Context, serviceNodeId, tenantId string) (*models.ServiceNodeModel, error) {
	if serviceNodeId == "" {
		return nil, errors.New("serviceNodeId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_SERVICE_NODE 
		WHERE serviceNodeId = ? AND tenantId = ?
	`

	var node models.ServiceNodeModel
	err := dao.db.QueryOne(ctx, &node, query, []interface{}{serviceNodeId, tenantId}, true)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil // 没有找到记录，返回nil
		}
		return nil, huberrors.WrapError(err, "获取服务节点失败")
	}

	return &node, nil
}

// GetServiceNodesByService 获取服务定义下的所有节点
func (dao *ServiceNodeDAO) GetServiceNodesByService(ctx context.Context, serviceDefinitionId, tenantId string) ([]*models.ServiceNodeModel, error) {
	if serviceDefinitionId == "" {
		return nil, errors.New("serviceDefinitionId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_SERVICE_NODE 
		WHERE serviceDefinitionId = ? AND tenantId = ?
		ORDER BY nodeWeight DESC, addTime ASC
	`

	var nodes []*models.ServiceNodeModel
	err := dao.db.Query(ctx, &nodes, query, []interface{}{serviceDefinitionId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取服务节点列表失败")
	}

	return nodes, nil
}

// CreateServiceNode 创建服务节点
func (dao *ServiceNodeDAO) CreateServiceNode(ctx context.Context, node *models.ServiceNodeModel, operatorId string) (string, error) {
	if node == nil {
		return "", errors.New("服务节点不能为空")
	}
	if node.ServiceDefinitionId == "" {
		return "", errors.New("serviceDefinitionId不能为空")
	}
	if node.NodeHost == "" {
		return "", errors.New("nodeHost不能为空")
	}
	if node.NodePort <= 0 {
		return "", errors.New("nodePort必须大于0")
	}

	// 自动生成服务节点ID（如果为空）
	if node.ServiceNodeId == "" {
		// 使用公共方法生成32位唯一字符串，前缀为"SN"
		node.ServiceNodeId = random.GenerateUniqueStringWithPrefix("SN", 32)
	}

	// 设置默认值
	if node.NodeId == "" {
		node.NodeId = fmt.Sprintf("node-%s", node.ServiceNodeId[:8])
	}
	if node.NodeProtocol == "" {
		node.NodeProtocol = "HTTP"
	}
	if node.NodeWeight <= 0 {
		node.NodeWeight = 100
	}
	if node.HealthStatus == "" {
		node.HealthStatus = "Y"
	}
	if node.NodeStatus == 0 {
		node.NodeStatus = 1 // 默认在线状态
	}
	if node.NodeUrl == "" {
		// 构建节点URL
		protocol := strings.ToLower(node.NodeProtocol)
		node.NodeUrl = fmt.Sprintf("%s://%s:%d", protocol, node.NodeHost, node.NodePort)
	}

	// 设置审计字段
	now := time.Now()
	node.AddTime = &now
	node.EditTime = &now
	node.AddWho = operatorId
	node.EditWho = operatorId
	// 生成oprSeqFlag - 使用32位唯一字符串
	node.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)
	node.CurrentVersion = 1
	node.ActiveFlag = "Y"

	// 构建SQL语句
	sql := `
		INSERT INTO HUB_GW_SERVICE_NODE (
			tenantId, serviceNodeId, serviceDefinitionId, nodeId,
			nodeUrl, nodeHost, nodePort, nodeProtocol, nodeWeight,
			healthStatus, nodeMetadata, nodeStatus, lastHealthCheckTime,
			healthCheckResult, reserved1, reserved2, reserved3, reserved4,
			reserved5, extProperty, addTime, addWho, editTime,
			editWho, oprSeqFlag, currentVersion, activeFlag, noteText
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
	`

	// 执行插入
	_, err := dao.db.Exec(ctx, sql, []interface{}{
		node.TenantId, node.ServiceNodeId, node.ServiceDefinitionId, node.NodeId,
		node.NodeUrl, node.NodeHost, node.NodePort, node.NodeProtocol, node.NodeWeight,
		node.HealthStatus, node.NodeMetadata, node.NodeStatus, node.LastHealthCheckTime,
		node.HealthCheckResult, node.Reserved1, node.Reserved2, node.Reserved3, node.Reserved4,
		node.Reserved5, node.ExtProperty, node.AddTime, node.AddWho, node.EditTime,
		node.EditWho, node.OprSeqFlag, node.CurrentVersion, node.ActiveFlag, node.NoteText,
	}, true)

	if err != nil {
		return "", huberrors.WrapError(err, "创建服务节点失败")
	}

	return node.ServiceNodeId, nil
}

// UpdateServiceNode 更新服务节点
func (dao *ServiceNodeDAO) UpdateServiceNode(ctx context.Context, node *models.ServiceNodeModel, operatorId string) error {
	if node == nil {
		return errors.New("服务节点不能为空")
	}
	if node.ServiceNodeId == "" {
		return errors.New("serviceNodeId不能为空")
	}

	// 获取当前节点信息，以便保留不可修改的字段
	currentNode, err := dao.GetServiceNodeById(ctx, node.ServiceNodeId, node.TenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取当前服务节点信息失败")
	}
	if currentNode == nil {
		return errors.New("服务节点不存在")
	}

	// 保留不可修改的字段
	node.TenantId = currentNode.TenantId
	node.ServiceNodeId = currentNode.ServiceNodeId
	node.ServiceDefinitionId = currentNode.ServiceDefinitionId // 服务定义ID不允许修改
	node.AddTime = currentNode.AddTime
	node.AddWho = currentNode.AddWho

	// 更新审计字段
	now := time.Now()
	node.EditTime = &now
	node.EditWho = operatorId
	// 生成oprSeqFlag - 使用32位唯一字符串
	node.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)
	node.CurrentVersion = currentNode.CurrentVersion + 1

	// 如果URL为空，但主机和端口已提供，则重新生成URL
	if (node.NodeUrl == "" || node.NodeUrl != currentNode.NodeUrl) && node.NodeHost != "" && node.NodePort > 0 {
		protocol := strings.ToLower(node.NodeProtocol)
		if protocol == "" {
			protocol = strings.ToLower(currentNode.NodeProtocol)
		}
		node.NodeUrl = fmt.Sprintf("%s://%s:%d", protocol, node.NodeHost, node.NodePort)
	}

	// 构建更新SQL
	sql := `
		UPDATE HUB_GW_SERVICE_NODE SET
			nodeId = ?, nodeUrl = ?,
			nodeHost = ?, nodePort = ?, nodeProtocol = ?, nodeWeight = ?,
			healthStatus = ?, nodeMetadata = ?, nodeStatus = ?,
			lastHealthCheckTime = ?, healthCheckResult = ?, reserved1 = ?, reserved2 = ?,
			reserved3 = ?, reserved4 = ?, reserved5 = ?, extProperty = ?,
			editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?,
			activeFlag = ?, noteText = ?
		WHERE serviceNodeId = ? AND tenantId = ? AND currentVersion = ?
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		node.NodeId, node.NodeUrl,
		node.NodeHost, node.NodePort, node.NodeProtocol, node.NodeWeight,
		node.HealthStatus, node.NodeMetadata, node.NodeStatus,
		node.LastHealthCheckTime, node.HealthCheckResult, node.Reserved1, node.Reserved2,
		node.Reserved3, node.Reserved4, node.Reserved5, node.ExtProperty,
		node.EditTime, node.EditWho, node.OprSeqFlag, node.CurrentVersion,
		node.ActiveFlag, node.NoteText,
		node.ServiceNodeId, node.TenantId, currentNode.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新服务节点失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("更新失败，可能是版本冲突或服务节点不存在")
	}

	return nil
}

// DeleteServiceNode 删除服务节点
func (dao *ServiceNodeDAO) DeleteServiceNode(ctx context.Context, serviceNodeId, tenantId, operatorId string) error {
	if serviceNodeId == "" {
		return errors.New("serviceNodeId不能为空")
	}

	// 检查服务节点是否存在
	existing, err := dao.GetServiceNodeById(ctx, serviceNodeId, tenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取服务节点失败")
	}
	if existing == nil {
		return errors.New("服务节点不存在")
	}

	// 执行物理删除
	sql := `DELETE FROM HUB_GW_SERVICE_NODE WHERE serviceNodeId = ? AND tenantId = ?`

	result, err := dao.db.Exec(ctx, sql, []interface{}{serviceNodeId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除服务节点失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("删除失败，服务节点不存在")
	}

	return nil
}

// UpdateNodeHealth 更新节点健康状态
func (dao *ServiceNodeDAO) UpdateNodeHealth(ctx context.Context, serviceNodeId, tenantId, healthStatus, healthCheckResult, operatorId string) error {
	if serviceNodeId == "" || healthStatus == "" {
		return errors.New("serviceNodeId和healthStatus不能为空")
	}

	// 获取当前节点信息
	currentNode, err := dao.GetServiceNodeById(ctx, serviceNodeId, tenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取服务节点信息失败")
	}
	if currentNode == nil {
		return errors.New("服务节点不存在")
	}

	// 更新审计字段
	now := time.Now()
	newVersion := currentNode.CurrentVersion + 1
	// 生成oprSeqFlag - 使用32位唯一字符串
	oprSeqFlag := random.GenerateUniqueStringWithPrefix("", 32)

	// 构建更新SQL
	sql := `
		UPDATE HUB_GW_SERVICE_NODE SET
			healthStatus = ?,
			healthCheckResult = ?,
			lastHealthCheckTime = ?,
			editTime = ?,
			editWho = ?,
			oprSeqFlag = ?,
			currentVersion = ?
		WHERE serviceNodeId = ? AND tenantId = ? AND currentVersion = ?
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		healthStatus,
		healthCheckResult,
		now,
		now,
		operatorId,
		oprSeqFlag,
		newVersion,
		serviceNodeId,
		tenantId,
		currentNode.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新节点健康状态失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("更新失败，可能是版本冲突或服务节点不存在")
	}

	return nil
}
