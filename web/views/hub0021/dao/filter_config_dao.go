package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0021/models"
	"strings"
	"time"
)

// FilterConfigDAO 过滤器配置数据访问对象
type FilterConfigDAO struct {
	db database.Database
}

// NewFilterConfigDAO 创建过滤器配置DAO
func NewFilterConfigDAO(db database.Database) *FilterConfigDAO {
	return &FilterConfigDAO{
		db: db,
	}
}

// generateFilterConfigId 生成过滤器配置ID
// 格式：FC + YYYYMMDD + HHMMSS + 4位随机数
// 示例：FC20240615143022A1B2
func (dao *FilterConfigDAO) generateFilterConfigId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")

	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)

	return fmt.Sprintf("FC%s%s", timeStr, randomStr)
}

// isFilterConfigIdExists 检查过滤器配置ID是否已存在
func (dao *FilterConfigDAO) isFilterConfigIdExists(ctx context.Context, filterConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_FILTER_CONFIG WHERE filterConfigId = ?`

	var result struct {
		Count int `db:"count"`
	}

	err := dao.db.QueryOne(ctx, &result, query, []interface{}{filterConfigId}, true)
	if err != nil {
		return false, err
	}

	return result.Count > 0, nil
}

// generateUniqueFilterConfigId 生成唯一的过滤器配置ID
// 如果生成的ID已存在，会重新生成直到找到唯一的ID（最多尝试10次）
func (dao *FilterConfigDAO) generateUniqueFilterConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10

	for attempt := 0; attempt < maxAttempts; attempt++ {
		filterConfigId := dao.generateFilterConfigId()

		exists, err := dao.isFilterConfigIdExists(ctx, filterConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查过滤器配置ID是否存在失败")
		}

		if !exists {
			return filterConfigId, nil
		}

		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}

	return "", errors.New("生成唯一过滤器配置ID失败，已达到最大尝试次数")
}

// AddFilterConfig 添加过滤器配置
// 参数:
//   - ctx: 上下文对象
//   - filterConfig: 过滤器配置信息
//   - operatorId: 操作人ID
//
// 返回:
//   - filterConfigId: 新创建的过滤器配置ID
//   - err: 可能的错误
func (dao *FilterConfigDAO) AddFilterConfig(ctx context.Context, filterConfig *models.FilterConfig, operatorId string) (string, error) {
	// 验证租户ID
	if filterConfig.TenantId == "" {
		return "", errors.New("租户ID不能为空")
	}

	// 验证必填字段
	if filterConfig.FilterName == "" {
		return "", errors.New("过滤器名称不能为空")
	}
	if filterConfig.FilterType == "" {
		return "", errors.New("过滤器类型不能为空")
	}
	if filterConfig.FilterAction == "" {
		return "", errors.New("过滤器执行时机不能为空")
	}

	// 验证过滤器类型和执行时机的有效性
	if !models.IsValidFilterType(filterConfig.FilterType) {
		return "", errors.New("无效的过滤器类型: " + filterConfig.FilterType)
	}
	if !models.IsValidFilterAction(filterConfig.FilterAction) {
		return "", errors.New("无效的过滤器执行时机: " + filterConfig.FilterAction)
	}

	// 验证实例级或路由级配置（二选一）
	if filterConfig.GatewayInstanceId == "" && filterConfig.RouteConfigId == "" {
		return "", errors.New("必须指定网关实例ID或路由配置ID")
	}
	if filterConfig.GatewayInstanceId != "" && filterConfig.RouteConfigId != "" {
		return "", errors.New("不能同时指定网关实例ID和路由配置ID")
	}

	// 自动生成过滤器配置ID（如果为空）
	if filterConfig.FilterConfigId == "" {
		generatedId, err := dao.generateUniqueFilterConfigId(ctx)
		if err != nil {
			return "", huberrors.WrapError(err, "生成过滤器配置ID失败")
		}
		filterConfig.FilterConfigId = generatedId
	} else {
		// 如果提供了ID，检查是否已存在
		exists, err := dao.isFilterConfigIdExists(ctx, filterConfig.FilterConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查过滤器配置ID是否存在失败")
		}
		if exists {
			return "", errors.New("过滤器配置ID已存在")
		}
	}

	// 设置一些自动填充的字段
	now := time.Now()
	filterConfig.AddTime = now
	filterConfig.AddWho = operatorId
	filterConfig.EditTime = now
	filterConfig.EditWho = operatorId
	filterConfig.OprSeqFlag = filterConfig.FilterConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	filterConfig.CurrentVersion = 1
	filterConfig.ActiveFlag = "Y"

	// 设置默认值
	if filterConfig.FilterOrder == 0 {
		filterConfig.FilterOrder = 100 // 默认执行顺序
	}
	if filterConfig.FilterConfig == "" {
		filterConfig.FilterConfig = "{}" // 默认空配置
	}

	// 验证FilterConfig是否为有效的JSON
	var configTest interface{}
	if err := json.Unmarshal([]byte(filterConfig.FilterConfig), &configTest); err != nil {
		return "", errors.New("过滤器配置不是有效的JSON格式: " + err.Error())
	}

	// 使用数据库接口的Insert方法插入记录
	_, err := dao.db.Insert(ctx, "HUB_GW_FILTER_CONFIG", filterConfig, true)

	if err != nil {
		// 检查是否是过滤器名重复错误
		if dao.isDuplicateFilterNameError(err) {
			return "", huberrors.WrapError(err, "过滤器名已存在")
		}
		return "", huberrors.WrapError(err, "添加过滤器配置失败")
	}

	return filterConfig.FilterConfigId, nil
}

// GetFilterConfigById 根据过滤器配置ID获取过滤器配置信息
func (dao *FilterConfigDAO) GetFilterConfigById(ctx context.Context, filterConfigId, tenantId string) (*models.FilterConfig, error) {
	if filterConfigId == "" || tenantId == "" {
		return nil, errors.New("filterConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_FILTER_CONFIG 
		WHERE filterConfigId = ? AND tenantId = ?
	`

	var filterConfig models.FilterConfig
	err := dao.db.QueryOne(ctx, &filterConfig, query, []interface{}{filterConfigId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询过滤器配置失败")
	}

	return &filterConfig, nil
}

// UpdateFilterConfig 更新过滤器配置
func (dao *FilterConfigDAO) UpdateFilterConfig(ctx context.Context, filterConfig *models.FilterConfig, operatorId string) error {
	if filterConfig.FilterConfigId == "" || filterConfig.TenantId == "" {
		return errors.New("filterConfigId和tenantId不能为空")
	}

	// 验证必填字段
	if filterConfig.FilterName == "" {
		return errors.New("过滤器名称不能为空")
	}
	if filterConfig.FilterType == "" {
		return errors.New("过滤器类型不能为空")
	}
	if filterConfig.FilterAction == "" {
		return errors.New("过滤器执行时机不能为空")
	}

	// 验证过滤器类型和执行时机的有效性
	if !models.IsValidFilterType(filterConfig.FilterType) {
		return errors.New("无效的过滤器类型: " + filterConfig.FilterType)
	}
	if !models.IsValidFilterAction(filterConfig.FilterAction) {
		return errors.New("无效的过滤器执行时机: " + filterConfig.FilterAction)
	}

	// 验证实例级或路由级配置（二选一）
	if filterConfig.GatewayInstanceId == "" && filterConfig.RouteConfigId == "" {
		return errors.New("必须指定网关实例ID或路由配置ID")
	}
	if filterConfig.GatewayInstanceId != "" && filterConfig.RouteConfigId != "" {
		return errors.New("不能同时指定网关实例ID和路由配置ID")
	}

	// 验证FilterConfig是否为有效的JSON
	if filterConfig.FilterConfig != "" {
		var configTest interface{}
		if err := json.Unmarshal([]byte(filterConfig.FilterConfig), &configTest); err != nil {
			return errors.New("过滤器配置不是有效的JSON格式: " + err.Error())
		}
	}

	// 更新时间和操作人
	filterConfig.EditTime = time.Now()
	filterConfig.EditWho = operatorId

	// 检查过滤器配置是否存在
	existing, err := dao.GetFilterConfigById(ctx, filterConfig.FilterConfigId, filterConfig.TenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取现有过滤器配置失败")
	}
	if existing == nil {
		return errors.New("过滤器配置不存在")
	}

	// 保留不可修改的字段
	filterConfig.TenantId = existing.TenantId
	filterConfig.FilterConfigId = existing.FilterConfigId
	filterConfig.AddTime = existing.AddTime
	filterConfig.AddWho = existing.AddWho
	filterConfig.OprSeqFlag = existing.OprSeqFlag
	filterConfig.CurrentVersion = existing.CurrentVersion + 1

	// 构建更新SQL
	sql := `
		UPDATE HUB_GW_FILTER_CONFIG SET
			gatewayInstanceId = ?, routeConfigId = ?, filterName = ?, filterType = ?, filterAction = ?,
			filterOrder = ?, filterConfig = ?, filterDesc = ?, configId = ?,
			reserved1 = ?, reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?,
			extProperty = ?, editTime = ?, editWho = ?, currentVersion = ?, activeFlag = ?, noteText = ?
		WHERE filterConfigId = ? AND tenantId = ? AND currentVersion = ?
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		filterConfig.GatewayInstanceId, filterConfig.RouteConfigId, filterConfig.FilterName,
		filterConfig.FilterType, filterConfig.FilterAction,
		filterConfig.FilterOrder, filterConfig.FilterConfig, filterConfig.FilterDesc, filterConfig.ConfigId,
		filterConfig.Reserved1, filterConfig.Reserved2, filterConfig.Reserved3, filterConfig.Reserved4, filterConfig.Reserved5,
		filterConfig.ExtProperty, filterConfig.EditTime, filterConfig.EditWho, filterConfig.CurrentVersion,
		filterConfig.ActiveFlag, filterConfig.NoteText,
		filterConfig.FilterConfigId, filterConfig.TenantId, existing.CurrentVersion,
	}, true)

	if err != nil {
		if dao.isDuplicateFilterNameError(err) {
			return huberrors.WrapError(err, "过滤器名已存在")
		}
		return huberrors.WrapError(err, "更新过滤器配置失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("过滤器配置数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteFilterConfig 删除过滤器配置
func (dao *FilterConfigDAO) DeleteFilterConfig(ctx context.Context, filterConfigId, tenantId, operatorId string) error {
	if filterConfigId == "" || tenantId == "" {
		return errors.New("filterConfigId和tenantId不能为空")
	}

	// 先检查记录是否存在
	existingConfig, err := dao.GetFilterConfigById(ctx, filterConfigId, tenantId)
	if err != nil {
		return huberrors.WrapError(err, "查询过滤器配置失败")
	}
	if existingConfig == nil {
		return errors.New("过滤器配置不存在")
	}

	// 执行实际删除
	sql := `DELETE FROM HUB_GW_FILTER_CONFIG WHERE filterConfigId = ? AND tenantId = ?`

	result, err := dao.db.Exec(ctx, sql, []interface{}{filterConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除过滤器配置失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("过滤器配置不存在或已被删除")
	}

	return nil
}

// ListFilterConfigs 获取过滤器配置列表
func (dao *FilterConfigDAO) ListFilterConfigs(ctx context.Context, tenantId string, gatewayInstanceId string, routeConfigId string, activeFlag string, page, pageSize int) ([]*models.FilterConfig, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建基础查询条件
	whereConditions := []string{"tenantId = ?"}
	args := []interface{}{tenantId}

	// 添加网关实例ID条件
	if gatewayInstanceId != "" {
		whereConditions = append(whereConditions, "gatewayInstanceId = ?")
		args = append(args, gatewayInstanceId)
	}

	// 添加路由配置ID条件
	if routeConfigId != "" {
		whereConditions = append(whereConditions, "routeConfigId = ?")
		args = append(args, routeConfigId)
	}

	// 添加activeFlag条件（如果指定了activeFlag参数）
	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 构建基础查询语句
	baseQuery := fmt.Sprintf("SELECT * FROM HUB_GW_FILTER_CONFIG WHERE %s ORDER BY filterAction ASC, filterOrder ASC, addTime DESC", whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, args, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询过滤器配置总数失败")
	}

	// 如果没有数据，直接返回
	if countResult.Count == 0 {
		return []*models.FilterConfig{}, 0, nil
	}

	// 创建分页信息
	paginationInfo := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, paginationInfo)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(args, paginationArgs...)

	// 执行分页查询
	var filterConfigs []*models.FilterConfig
	err = dao.db.Query(ctx, &filterConfigs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询过滤器配置列表失败")
	}

	return filterConfigs, countResult.Count, nil
}

// GetFilterConfigsByGatewayInstance 根据网关实例ID获取过滤器配置列表
func (dao *FilterConfigDAO) GetFilterConfigsByGatewayInstance(ctx context.Context, gatewayInstanceId, tenantId string) ([]*models.FilterConfig, error) {
	if gatewayInstanceId == "" || tenantId == "" {
		return nil, errors.New("gatewayInstanceId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_FILTER_CONFIG 
		WHERE gatewayInstanceId = ? AND tenantId = ?
		ORDER BY filterAction ASC, filterOrder ASC, addTime DESC
	`

	var filterConfigs []*models.FilterConfig
	err := dao.db.Query(ctx, &filterConfigs, query, []interface{}{gatewayInstanceId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询网关实例过滤器配置失败")
	}

	return filterConfigs, nil
}

// GetFilterConfigsByRoute 根据路由配置ID获取过滤器配置列表
func (dao *FilterConfigDAO) GetFilterConfigsByRoute(ctx context.Context, routeConfigId, tenantId string, activeFlag string) ([]*models.FilterConfig, error) {
	if routeConfigId == "" || tenantId == "" {
		return nil, errors.New("routeConfigId和tenantId不能为空")
	}

	// 构建查询条件
	whereConditions := []string{"routeConfigId = ?", "tenantId = ?"}
	args := []interface{}{routeConfigId, tenantId}

	// 添加activeFlag条件（如果指定了activeFlag参数）
	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT * FROM HUB_GW_FILTER_CONFIG 
		WHERE %s
		ORDER BY filterAction ASC, filterOrder ASC, addTime DESC
	`, whereClause)

	var filterConfigs []*models.FilterConfig
	err := dao.db.Query(ctx, &filterConfigs, query, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询路由过滤器配置失败")
	}

	return filterConfigs, nil
}

// GetFilterConfigsByType 根据过滤器类型获取过滤器配置列表
func (dao *FilterConfigDAO) GetFilterConfigsByType(ctx context.Context, filterType, tenantId string, gatewayInstanceId, routeConfigId string, activeFlag string) ([]*models.FilterConfig, error) {
	if filterType == "" || tenantId == "" {
		return nil, errors.New("filterType和tenantId不能为空")
	}

	// 验证过滤器类型的有效性
	if !models.IsValidFilterType(filterType) {
		return nil, errors.New("无效的过滤器类型: " + filterType)
	}

	// 构建查询条件
	whereConditions := []string{"tenantId = ?", "filterType = ?"}
	args := []interface{}{tenantId, filterType}

	// 添加activeFlag条件（如果指定了activeFlag参数）
	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	// 添加可选的网关实例ID或路由配置ID条件
	if gatewayInstanceId != "" {
		whereConditions = append(whereConditions, "gatewayInstanceId = ?")
		args = append(args, gatewayInstanceId)
	}
	if routeConfigId != "" {
		whereConditions = append(whereConditions, "routeConfigId = ?")
		args = append(args, routeConfigId)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT * FROM HUB_GW_FILTER_CONFIG 
		WHERE %s
		ORDER BY filterAction ASC, filterOrder ASC, addTime DESC
	`, whereClause)

	var filterConfigs []*models.FilterConfig
	err := dao.db.Query(ctx, &filterConfigs, query, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询指定类型过滤器配置失败")
	}

	return filterConfigs, nil
}

// GetFilterConfigsByAction 根据执行时机获取过滤器配置列表
func (dao *FilterConfigDAO) GetFilterConfigsByAction(ctx context.Context, filterAction, tenantId string, gatewayInstanceId, routeConfigId string, activeFlag string) ([]*models.FilterConfig, error) {
	if filterAction == "" || tenantId == "" {
		return nil, errors.New("filterAction和tenantId不能为空")
	}

	// 验证过滤器执行时机的有效性
	if !models.IsValidFilterAction(filterAction) {
		return nil, errors.New("无效的过滤器执行时机: " + filterAction)
	}

	// 构建查询条件
	whereConditions := []string{"tenantId = ?", "filterAction = ?"}
	args := []interface{}{tenantId, filterAction}

	// 添加activeFlag条件（如果指定了activeFlag参数）
	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	// 添加可选的网关实例ID或路由配置ID条件
	if gatewayInstanceId != "" {
		whereConditions = append(whereConditions, "gatewayInstanceId = ?")
		args = append(args, gatewayInstanceId)
	}
	if routeConfigId != "" {
		whereConditions = append(whereConditions, "routeConfigId = ?")
		args = append(args, routeConfigId)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT * FROM HUB_GW_FILTER_CONFIG 
		WHERE %s
		ORDER BY filterOrder ASC, addTime DESC
	`, whereClause)

	var filterConfigs []*models.FilterConfig
	err := dao.db.Query(ctx, &filterConfigs, query, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询指定执行时机过滤器配置失败")
	}

	return filterConfigs, nil
}

// GetFilterExecutionChain 获取过滤器执行链（按执行时机和顺序排序）
func (dao *FilterConfigDAO) GetFilterExecutionChain(ctx context.Context, tenantId string, gatewayInstanceId, routeConfigId string, activeFlag string) ([]*models.FilterConfig, error) {
	if tenantId == "" {
		return nil, errors.New("tenantId不能为空")
	}

	// 构建查询条件
	whereConditions := []string{"tenantId = ?"}
	args := []interface{}{tenantId}

	// 添加activeFlag条件（如果指定了activeFlag参数）
	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	// 添加可选的网关实例ID或路由配置ID条件
	if gatewayInstanceId != "" {
		whereConditions = append(whereConditions, "gatewayInstanceId = ?")
		args = append(args, gatewayInstanceId)
	}
	if routeConfigId != "" {
		whereConditions = append(whereConditions, "routeConfigId = ?")
		args = append(args, routeConfigId)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 按执行时机和执行顺序排序，确保执行链的正确顺序
	query := fmt.Sprintf(`
		SELECT * FROM HUB_GW_FILTER_CONFIG 
		WHERE %s
		ORDER BY 
			CASE filterAction 
				WHEN 'pre-routing' THEN 1 
				WHEN 'post-routing' THEN 2 
				WHEN 'pre-response' THEN 3 
				ELSE 4 
			END ASC,
			filterOrder ASC, 
			addTime ASC
	`, whereClause)

	var filterConfigs []*models.FilterConfig
	err := dao.db.Query(ctx, &filterConfigs, query, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询过滤器执行链失败")
	}

	return filterConfigs, nil
}

// UpdateFilterOrder 更新过滤器执行顺序
func (dao *FilterConfigDAO) UpdateFilterOrder(ctx context.Context, filterConfigId, tenantId string, newOrder int, operatorId string) error {
	if filterConfigId == "" || tenantId == "" {
		return errors.New("filterConfigId和tenantId不能为空")
	}

	// 更新过滤器执行顺序
	sql := `
		UPDATE HUB_GW_FILTER_CONFIG SET
			filterOrder = ?, editTime = ?, editWho = ?
		WHERE filterConfigId = ? AND tenantId = ?
	`

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		newOrder, time.Now(), operatorId,
		filterConfigId, tenantId,
	}, true)
	if err != nil {
		return huberrors.WrapError(err, "更新过滤器执行顺序失败")
	}

	return nil
}

// EnableFilterConfig 启用过滤器配置
func (dao *FilterConfigDAO) EnableFilterConfig(ctx context.Context, filterConfigId, tenantId, operatorId string) error {
	return dao.updateFilterConfigStatus(ctx, filterConfigId, tenantId, "Y", operatorId)
}

// DisableFilterConfig 禁用过滤器配置
func (dao *FilterConfigDAO) DisableFilterConfig(ctx context.Context, filterConfigId, tenantId, operatorId string) error {
	return dao.updateFilterConfigStatus(ctx, filterConfigId, tenantId, "N", operatorId)
}

// updateFilterConfigStatus 更新过滤器配置状态
func (dao *FilterConfigDAO) updateFilterConfigStatus(ctx context.Context, filterConfigId, tenantId, status, operatorId string) error {
	if filterConfigId == "" || tenantId == "" {
		return errors.New("filterConfigId和tenantId不能为空")
	}

	// 更新过滤器配置状态
	sql := `
		UPDATE HUB_GW_FILTER_CONFIG SET
			activeFlag = ?, editTime = ?, editWho = ?
		WHERE filterConfigId = ? AND tenantId = ?
	`

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		status, time.Now(), operatorId,
		filterConfigId, tenantId,
	}, true)
	if err != nil {
		return huberrors.WrapError(err, "更新过滤器配置状态失败")
	}

	return nil
}

// BatchUpdateFilterConfigs 批量更新过滤器配置
func (dao *FilterConfigDAO) BatchUpdateFilterConfigs(ctx context.Context, filterConfigIds []string, tenantId string, updates map[string]interface{}, operatorId string) error {
	if len(filterConfigIds) == 0 || tenantId == "" {
		return errors.New("filterConfigIds和tenantId不能为空")
	}

	// 构建批量更新的SQL语句
	var setParts []string
	var args []interface{}

	// 添加编辑时间和编辑人
	updates["editTime"] = time.Now()
	updates["editWho"] = operatorId

	// 构建SET子句
	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}

	// 构建WHERE条件
	placeholders := strings.Repeat("?,", len(filterConfigIds))
	placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号

	// 添加WHERE参数
	for _, id := range filterConfigIds {
		args = append(args, id)
	}
	args = append(args, tenantId)

	// 构建完整的SQL语句
	sql := fmt.Sprintf(`
		UPDATE HUB_GW_FILTER_CONFIG SET %s
		WHERE filterConfigId IN (%s) AND tenantId = ?
	`, strings.Join(setParts, ", "), placeholders)

	_, err := dao.db.Exec(ctx, sql, args, true)
	if err != nil {
		return huberrors.WrapError(err, "批量更新过滤器配置失败")
	}

	return nil
}

// BatchDeleteFilterConfigs 批量删除过滤器配置
func (dao *FilterConfigDAO) BatchDeleteFilterConfigs(ctx context.Context, filterConfigIds []string, tenantId, operatorId string) error {
	if len(filterConfigIds) == 0 || tenantId == "" {
		return errors.New("filterConfigIds和tenantId不能为空")
	}

	// 构建IN子句的占位符
	placeholders := strings.Repeat("?,", len(filterConfigIds))
	placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号

	// 构建参数列表
	args := make([]interface{}, 0, len(filterConfigIds)+1)
	for _, id := range filterConfigIds {
		args = append(args, id)
	}
	args = append(args, tenantId)

	// 执行批量删除
	sql := fmt.Sprintf(`DELETE FROM HUB_GW_FILTER_CONFIG WHERE filterConfigId IN (%s) AND tenantId = ?`, placeholders)

	_, err := dao.db.Exec(ctx, sql, args, true)
	if err != nil {
		return huberrors.WrapError(err, "批量删除过滤器配置失败")
	}

	return nil
}

// isDuplicateFilterNameError 检查是否是过滤器名重复错误
func (dao *FilterConfigDAO) isDuplicateFilterNameError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate") && strings.Contains(errStr, "filterName")
}
