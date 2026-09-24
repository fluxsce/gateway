package controllers

import (
	"errors"
	"strings"

	"encoding/json"

	"gateway/internal/servicecenterv3"
	"gateway/internal/servicecenterv3/catalog"
	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/middleware/audit"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0042/dao"
	"gateway/web/views/hub0042/models"

	"github.com/gin-gonic/gin"
)

var (
	errServiceKeyRequired       = errors.New("namespaceId和serviceName不能为空")
	errNamingInstanceUnresolved = errors.New("无法解析服务中心实例，运行时服务未删除")
	errNamingRuntimeUnavailable = errors.New("服务中心实例未就绪，运行时服务未删除")
)

// ServiceController 服务控制器
type ServiceController struct {
	db           database.Database
	serviceDAO   *dao.ServiceDAO
	namespaceDAO *catalog.NamespaceDAO
}

// NewServiceController 创建服务控制器
func NewServiceController(db database.Database) *ServiceController {
	return &ServiceController{
		db:           db,
		serviceDAO:   dao.NewServiceDAO(db),
		namespaceDAO: catalog.NewNamespaceDAO(db),
	}
}

// QueryServices 获取服务列表
// @Summary 获取服务列表
// @Description 分页获取服务列表，支持条件查询
// @Tags 服务监控
// @Produce json
// @Param pageIndex query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param serviceName query string false "服务名称（模糊查询）"
// @Param namespaceId query string true "命名空间ID"
// @Param groupName query string false "分组名称"
// @Param serviceType query string false "服务类型（INTERNAL, NACOS, CONSUL, EUREKA, ETCD, ZOOKEEPER）"
// @Param instanceName query string false "服务中心实例名称"
// @Param environment query string false "部署环境（DEVELOPMENT, STAGING, PRODUCTION）"
// @Param activeFlag query string false "活动状态（Y/N）"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/services [get]
func (c *ServiceController) QueryServices(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.ServiceQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定服务查询条件失败，使用默认条件", "error", err.Error())
	}
	if query.NamespaceId == "" {
		response.ErrorJSON(ctx, "请先选择命名空间", constants.ED00006)
		return
	}

	// 调用DAO获取服务列表
	services, total, err := c.serviceDAO.ListServices(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务列表失败", err)
		response.ErrorJSON(ctx, "获取服务列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建服务列表
	serviceList := make([]map[string]interface{}, 0, len(services))
	instanceName, _ := c.resolveCenter(ctx, tenantId, query.NamespaceId)

	for _, service := range services {
		serviceInfo := map[string]interface{}{
			"tenantId":           service.TenantId,
			"namespaceId":        service.NamespaceId,
			"groupName":          service.GroupName,
			"serviceName":        service.ServiceName,
			"instanceName":       instanceName,
			"serviceType":        service.ServiceType,
			"serviceVersion":     service.ServiceVersion,
			"serviceDescription": service.ServiceDescription,
			"protectThreshold":   service.ProtectThreshold,
			"activeFlag":         service.ActiveFlag,
			"addTime":            service.AddTime,
			"addWho":             service.AddWho,
			"editTime":           service.EditTime,
			"editWho":            service.EditWho,
			"currentVersion":     service.CurrentVersion,
			"noteText":           service.NoteText,
			"nodeCount":          0,
			"healthyNodeCount":   0,
			"unhealthyNodeCount": 0,
			"subscriberCount":    0,
			"subscriptionCount":  0,
		}

		c.attachServiceNodeStats(ctx, tenantId, service.NamespaceId, service.GroupName, service.ServiceName, serviceInfo)
		c.attachServiceSubscriberStats(ctx, tenantId, service.NamespaceId, service.GroupName, service.ServiceName, serviceInfo)
		serviceList = append(serviceList, serviceInfo)
	}

	// v3：仅在第一页并入运行时独有服务，避免每一页重复出现临时服务
	if query.NamespaceId != "" && page <= 1 {
		serviceList, _ = c.mergeRuntimeOnlyServices(ctx, tenantId, query.NamespaceId, query.GroupName, serviceList)
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "serviceName"

	// 使用统一的分页响应
	response.PageJSON(ctx, serviceList, pageInfo, constants.SD00002)
}

// GetService 获取单个服务详情
// @Summary 获取服务详情
// @Description 根据服务主键获取服务详细信息，包括节点列表
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param namespaceId query string true "命名空间ID"
// @Param groupName query string true "分组名称"
// @Param serviceName query string true "服务名称"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/services/detail [get]
func (c *ServiceController) GetService(ctx *gin.Context) {
	namespaceId := request.GetParam(ctx, "namespaceId")
	groupName := request.GetParam(ctx, "groupName")
	serviceName := request.GetParam(ctx, "serviceName")

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 验证必填字段
	if namespaceId == "" || groupName == "" || serviceName == "" {
		response.ErrorJSON(ctx, "namespaceId、groupName和serviceName不能为空", constants.ED00006)
		return
	}

	// 调用DAO获取服务详情
	service, err := c.serviceDAO.GetServiceById(ctx, tenantId, namespaceId, groupName, serviceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务详情失败", err)
		response.ErrorJSON(ctx, "获取服务详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if service == nil {
		if runtime := c.runtimeServiceDetail(ctx, tenantId, namespaceId, groupName, serviceName); runtime != nil {
			response.SuccessJSON(ctx, runtime, constants.SD00001)
			return
		}
		response.ErrorJSON(ctx, "服务不存在", constants.ED00008)
		return
	}

	// 构建服务详情响应
	serviceInfo := map[string]interface{}{
		"tenantId":              service.TenantId,
		"namespaceId":           service.NamespaceId,
		"groupName":             service.GroupName,
		"serviceName":           service.ServiceName,
		"serviceType":           service.ServiceType,
		"serviceVersion":        service.ServiceVersion,
		"serviceDescription":    service.ServiceDescription,
		"externalServiceConfig": service.ExternalServiceConfig,
		"metadataJson":          service.MetadataJson,
		"tagsJson":              service.TagsJson,
		"protectThreshold":      service.ProtectThreshold,
		"selectorJson":          service.SelectorJson,
		"activeFlag":            service.ActiveFlag,
		"addTime":               service.AddTime,
		"addWho":                service.AddWho,
		"editTime":              service.EditTime,
		"editWho":               service.EditWho,
		"currentVersion":        service.CurrentVersion,
		"noteText":              service.NoteText,
		"extProperty":           service.ExtProperty,
		"nodes":                 []interface{}{},
		"nodeCount":             0,
		"healthyNodeCount":      0,
		"unhealthyNodeCount":    0,
	}

	c.attachServiceNodes(ctx, tenantId, namespaceId, groupName, serviceName, serviceInfo)
	c.attachServiceSubscribers(ctx, tenantId, namespaceId, groupName, serviceName, serviceInfo)

	// 直接返回服务对象
	response.SuccessJSON(ctx, serviceInfo, constants.SD00001)
}

// AddService 创建服务
// @Summary 创建服务
// @Description 创建新的服务
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param service body catalog.Service true "服务信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/services [post]
func (c *ServiceController) AddService(ctx *gin.Context) {
	var req catalog.Service
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID，不使用前端传递的值（前置校验已保证非空）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if req.NamespaceId == "" {
		response.ErrorJSON(ctx, "命名空间ID不能为空", constants.ED00006)
		return
	}
	if req.ServiceName == "" {
		response.ErrorJSON(ctx, "服务名称不能为空", constants.ED00006)
		return
	}

	// 设置默认值
	if req.GroupName == "" {
		req.GroupName = "DEFAULT_GROUP"
	}
	if req.ServiceType == "" {
		req.ServiceType = "INTERNAL"
	}

	// 检查服务是否已存在
	existingService, err := c.serviceDAO.GetServiceById(ctx, tenantId, req.NamespaceId, req.GroupName, req.ServiceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "检查服务是否存在时出错", err)
		response.ErrorJSON(ctx, "检查服务是否存在失败: "+err.Error(), constants.ED00009)
		return
	}
	if existingService != nil {
		response.ErrorJSON(ctx, "服务已存在，服务名称: "+req.ServiceName, constants.ED00008)
		return
	}

	// 只设置租户ID，其他默认参数（新增人、新增时间等）由DAO处理
	req.TenantId = tenantId

	// 调用DAO添加服务
	err = c.serviceDAO.AddService(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建服务失败", err)
		response.ErrorJSON(ctx, "创建服务失败: "+err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionCreate,
		ModuleCode:   "hub0042",
		TargetType:   "SERVICE",
		TargetId:     req.ServiceName,
		TargetName:   req.ServiceName,
		ResourceCode: "hub0042:add",
		Detail:       "namespace=" + req.NamespaceId + " group=" + req.GroupName,
	})

	// 查询新添加的服务信息
	newService, err := c.serviceDAO.GetServiceById(ctx, tenantId, req.NamespaceId, req.GroupName, req.ServiceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的服务信息失败", err)
		// 即使查询失败，也返回成功但只带有基本信息
		response.SuccessJSON(ctx, gin.H{
			"namespaceId": req.NamespaceId,
			"groupName":   req.GroupName,
			"serviceName": req.ServiceName,
			"message":     "服务创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newService == nil {
		logger.ErrorWithTrace(ctx, "新创建的服务不存在", "serviceName", req.ServiceName)
		response.SuccessJSON(ctx, gin.H{
			"namespaceId": req.NamespaceId,
			"groupName":   req.GroupName,
			"serviceName": req.ServiceName,
			"message":     "服务创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务创建成功",
		"namespaceId", newService.NamespaceId,
		"groupName", newService.GroupName,
		"serviceName", newService.ServiceName,
		"tenantId", tenantId,
		"operatorId", operatorId)

	if err := c.syncNamingService(ctx, tenantId, operatorId, newService); err != nil {
		logger.ErrorWithTrace(ctx, "同步服务到运行时失败", err,
			"namespaceId", newService.NamespaceId,
			"groupName", newService.GroupName,
			"serviceName", newService.ServiceName)
		response.ErrorJSON(ctx, "服务已写入目录，但同步运行时失败: "+err.Error(), constants.ED00009)
		return
	}

	// 直接返回服务对象
	response.SuccessJSON(ctx, newService, constants.SD00003)
}

// EditService 更新服务
// @Summary 更新服务
// @Description 更新服务信息
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param service body catalog.Service true "服务信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/services [put]
func (c *ServiceController) EditService(ctx *gin.Context) {
	var req catalog.Service
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if req.NamespaceId == "" || req.GroupName == "" || req.ServiceName == "" {
		response.ErrorJSON(ctx, "namespaceId、groupName和serviceName不能为空", constants.ED00006)
		return
	}

	// 获取现有服务信息进行校验
	currentService, err := c.serviceDAO.GetServiceById(ctx, tenantId, req.NamespaceId, req.GroupName, req.ServiceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务信息失败", err)
		response.ErrorJSON(ctx, "获取服务信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentService == nil {
		response.ErrorJSON(ctx, "服务不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段，确保关键字段不被前端覆盖
	req.TenantId = currentService.TenantId
	req.NamespaceId = currentService.NamespaceId
	req.GroupName = currentService.GroupName
	req.ServiceName = currentService.ServiceName

	// 调用DAO更新服务（DAO会处理EditTime和EditWho）
	err = c.serviceDAO.UpdateService(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新服务失败", err)
		response.ErrorJSON(ctx, "更新服务失败: "+err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionUpdate,
		ModuleCode:   "hub0042",
		TargetType:   "SERVICE",
		TargetId:     req.ServiceName,
		TargetName:   req.ServiceName,
		ResourceCode: "hub0042:edit",
		Detail:       "namespace=" + req.NamespaceId + " group=" + req.GroupName,
	})

	// 查询更新后的服务信息
	updatedService, err := c.serviceDAO.GetServiceById(ctx, tenantId, req.NamespaceId, req.GroupName, req.ServiceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的服务信息失败", err)
		// 即使查询失败，也返回成功但只带有简单消息
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	if err := c.syncNamingService(ctx, tenantId, operatorId, updatedService); err != nil {
		logger.ErrorWithTrace(ctx, "同步服务到运行时失败", err,
			"namespaceId", req.NamespaceId,
			"groupName", req.GroupName,
			"serviceName", req.ServiceName)
		response.ErrorJSON(ctx, "目录已更新，但同步运行时失败: "+err.Error(), constants.ED00009)
		return
	}

	// 直接返回服务对象
	response.SuccessJSON(ctx, updatedService, constants.SD00004)
}

// DeleteService 删除服务
// @Summary 删除服务
// @Description 删除服务
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param namespaceId query string true "命名空间ID"
// @Param groupName query string true "分组名称"
// @Param serviceName query string true "服务名称"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/services [delete]
func (c *ServiceController) DeleteService(ctx *gin.Context) {
	namespaceId := request.GetParam(ctx, "namespaceId")
	groupName := request.GetParam(ctx, "groupName")
	serviceName := request.GetParam(ctx, "serviceName")

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if err := c.removeService(ctx, tenantId, operatorId, namespaceId, groupName, serviceName, ""); err != nil {
		if errors.Is(err, errServiceKeyRequired) {
			response.ErrorJSON(ctx, err.Error(), constants.ED00006)
			return
		}
		logger.ErrorWithTrace(ctx, "删除服务失败", err)
		response.ErrorJSON(ctx, "删除服务失败: "+err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionDelete,
		ModuleCode:   "hub0042",
		TargetType:   "SERVICE",
		TargetId:     serviceName,
		TargetName:   serviceName,
		ResourceCode: "hub0042:delete",
		Detail:       "namespace=" + namespaceId + " group=" + groupName,
	})

	response.SuccessJSON(ctx, gin.H{
		"namespaceId": namespaceId,
		"groupName":   groupName,
		"serviceName": serviceName,
		"message":     "服务删除成功",
	}, constants.SD00005)
}

// BatchDeleteServices 批量删除服务
func (c *ServiceController) BatchDeleteServices(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	var req struct {
		Services json.RawMessage `json:"services" form:"services"`
	}
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}
	raw := req.Services
	if len(strings.TrimSpace(string(raw))) == 0 {
		if s := strings.TrimSpace(request.GetParam(ctx, "services")); s != "" {
			raw = json.RawMessage(s)
		}
	}
	items, err := parseBatchDeleteServices(raw)
	if err != nil {
		response.ErrorJSON(ctx, "参数错误: 服务列表格式不正确", constants.ED00006)
		return
	}
	if len(items) == 0 {
		response.ErrorJSON(ctx, "请先选择要删除的服务", constants.ED00006)
		return
	}

	successCount := 0
	failCount := 0
	firstErr := ""
	deleted := make([]string, 0, len(items))
	for _, item := range items {
		if err := c.removeService(ctx, tenantId, operatorId, item.NamespaceId, item.GroupName, item.ServiceName, item.InstanceName); err != nil {
			logger.WarnWithTrace(ctx, "批量删除服务失败", "error", err.Error(),
				"namespaceId", item.NamespaceId, "groupName", item.GroupName, "serviceName", item.ServiceName)
			failCount++
			if firstErr == "" {
				firstErr = err.Error()
			}
			continue
		}
		successCount++
		deleted = append(deleted, item.ServiceName)
	}

	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionDelete,
		ModuleCode:   "hub0042",
		TargetType:   "SERVICE",
		TargetId:     strings.Join(deleted, ","),
		ResourceCode: "hub0042:batchDelete",
		Detail:       "batch",
	})

	response.SuccessJSON(ctx, gin.H{
		"successCount": successCount,
		"failCount":    failCount,
		"error":        firstErr,
		"message":      "批量删除完成",
	}, constants.SD00005)
}

type batchDeleteServiceItem struct {
	NamespaceId  string `json:"namespaceId"`
	GroupName    string `json:"groupName"`
	ServiceName  string `json:"serviceName"`
	InstanceName string `json:"instanceName"`
}

func parseBatchDeleteServices(raw json.RawMessage) ([]batchDeleteServiceItem, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return nil, nil
	}
	var items []batchDeleteServiceItem
	if err := json.Unmarshal([]byte(trimmed), &items); err == nil {
		return items, nil
	}
	var one batchDeleteServiceItem
	if err := json.Unmarshal([]byte(trimmed), &one); err == nil && (one.NamespaceId != "" || one.ServiceName != "") {
		return []batchDeleteServiceItem{one}, nil
	}
	var wrapped string
	if err := json.Unmarshal([]byte(trimmed), &wrapped); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(wrapped), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func normalizeServiceKey(namespaceId, groupName, serviceName string) (string, string, string, error) {
	namespaceId = strings.TrimSpace(namespaceId)
	groupName = strings.TrimSpace(groupName)
	serviceName = strings.TrimSpace(serviceName)
	if namespaceId == "" || serviceName == "" {
		return "", "", "", errServiceKeyRequired
	}
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}
	return namespaceId, groupName, serviceName, nil
}

func isServiceMissing(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, contract.ErrServiceNotFound) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "服务不存在") || strings.Contains(msg, "未找到要删除的服务")
}

func (c *ServiceController) removeService(ctx *gin.Context, tenantId, operatorId, namespaceId, groupName, serviceName, instanceName string) error {
	namespaceId, groupName, serviceName, err := normalizeServiceKey(namespaceId, groupName, serviceName)
	if err != nil {
		return err
	}

	// 先走 v3：清活视图、L1、集群事件，并删同一张 HUB_SERVICE。
	v3Err := c.deregisterNamingService(ctx, tenantId, operatorId, namespaceId, groupName, serviceName, instanceName)
	dbErr := c.serviceDAO.DeleteService(ctx, tenantId, namespaceId, groupName, serviceName, operatorId)

	if dbErr != nil && !isServiceMissing(dbErr) {
		return dbErr
	}
	if v3Err != nil && !isServiceMissing(v3Err) {
		return v3Err
	}
	return nil
}

func (c *ServiceController) deregisterNamingService(ctx *gin.Context, tenantId, operatorId, namespaceId, groupName, serviceName, hintInstanceName string) error {
	pool := servicecenterv3.GetPool()
	if pool == nil {
		return nil
	}
	instanceName, environment := c.resolveCenter(ctx, tenantId, namespaceId)
	if hint := strings.TrimSpace(hintInstanceName); hint != "" {
		instanceName = hint
	}
	if instanceName == "" {
		return errNamingInstanceUnresolved
	}
	naming, err := pool.NamingOf(instanceName, environment)
	if err != nil {
		return err
	}
	if naming == nil {
		return errNamingRuntimeUnavailable
	}
	return naming.DeregisterService(ctx.Request.Context(), namingCC(tenantId, instanceName, namespaceId, operatorId), groupName, serviceName)
}

func (c *ServiceController) syncNamingService(ctx *gin.Context, tenantId, operatorId string, svc *catalog.Service) error {
	if svc == nil {
		return nil
	}
	instanceName, environment := c.resolveCenter(ctx, tenantId, svc.NamespaceId)
	return upsertNamingService(ctx.Request.Context(), tenantId, instanceName, environment, svc.NamespaceId, operatorId, toNamingService(svc))
}

func toNamingService(svc *catalog.Service) *model.Service {
	if svc == nil {
		return nil
	}
	out := &model.Service{
		TenantID:         svc.TenantId,
		NamespaceID:      svc.NamespaceId,
		GroupName:        svc.GroupName,
		ServiceName:      svc.ServiceName,
		Version:          svc.ServiceVersion,
		Description:      svc.ServiceDescription,
		ProtectThreshold: svc.ProtectThreshold,
	}
	if svc.MetadataJson != "" {
		var meta map[string]string
		if err := json.Unmarshal([]byte(svc.MetadataJson), &meta); err == nil {
			out.Metadata = meta
		}
	}
	if svc.TagsJson != "" {
		var tags map[string]string
		if err := json.Unmarshal([]byte(svc.TagsJson), &tags); err == nil {
			out.Tags = tags
		}
	}
	return out
}

// EditNode 编辑节点
// @Summary 编辑节点
// @Description 更新服务节点信息（如IP、端口、权重、元数据等），直接操作缓存，不操作数据库
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param node body catalog.ServiceNode true "节点信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/nodes [put]
func (c *ServiceController) EditNode(ctx *gin.Context) {
	var req catalog.ServiceNode
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if req.NodeId == "" {
		response.ErrorJSON(ctx, "nodeId不能为空", constants.ED00006)
		return
	}

	if err := c.updateRuntimeNode(ctx, operatorId, &req); err != nil {
		logger.ErrorWithTrace(ctx, "更新节点失败", err, "nodeId", req.NodeId)
		response.ErrorJSON(ctx, "更新节点失败: "+err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionUpdate,
		ModuleCode:   "hub0042",
		TargetType:   "SERVICE_NODE",
		TargetId:     req.NodeId,
		TargetName:   req.ServiceName,
		ResourceCode: "hub0042:node:edit",
	})
	response.SuccessJSON(ctx, req, constants.SD00004)
}

// OfflineNode 下线节点
// @Summary 下线节点
// @Description 将服务节点下线（设置状态为DOWN），直接操作缓存，不操作数据库
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param nodeId query string true "节点ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/nodes/offline [post]
func (c *ServiceController) OfflineNode(ctx *gin.Context) {
	nodeId := request.GetParam(ctx, "nodeId")

	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if nodeId == "" {
		response.ErrorJSON(ctx, "nodeId不能为空", constants.ED00006)
		return
	}

	if err := updateNamingNode(ctx.Request.Context(), operatorId, &model.Node{
		NodeID: nodeId,
		Status: model.NodeDown,
	}); err != nil {
		logger.ErrorWithTrace(ctx, "下线节点失败", err, "nodeId", nodeId)
		response.ErrorJSON(ctx, "下线节点失败: "+err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionUpdate,
		ModuleCode:   "hub0042",
		TargetType:   "SERVICE_NODE",
		TargetId:     nodeId,
		ResourceCode: "hub0042:node:offline",
	})
	if pool := servicecenterv3.GetPool(); pool != nil {
		if current, _, ok := pool.FindNode(nodeId); ok {
			response.SuccessJSON(ctx, nodeToMap(current), constants.SD00004)
			return
		}
	}
	response.SuccessJSON(ctx, gin.H{"nodeId": nodeId, "instanceStatus": "DOWN"}, constants.SD00004)
}

func (c *ServiceController) resolveCenter(ctx *gin.Context, tenantId, namespaceId string) (instanceName, environment string) {
	if namespaceId == "" {
		return "", ""
	}
	ns, err := c.namespaceDAO.GetNamespace(ctx.Request.Context(), tenantId, namespaceId)
	if err != nil || ns == nil {
		return "", ""
	}
	return ns.InstanceName, ns.Environment
}

func (c *ServiceController) attachServiceNodeStats(ctx *gin.Context, tenantId, namespaceId, groupName, serviceName string, info map[string]interface{}) {
	instanceName, environment := c.resolveCenter(ctx, tenantId, namespaceId)
	if instanceName != "" && servicecenterv3.GetPool() != nil {
		nodes, err := listNamingNodes(ctx.Request.Context(), tenantId, instanceName, environment, namespaceId, groupName, serviceName)
		if err == nil {
			overlayNodeStats(info, nodes)
		}
	}
}

func (c *ServiceController) attachServiceSubscriberStats(ctx *gin.Context, tenantId, namespaceId, groupName, serviceName string, info map[string]interface{}) {
	instanceName, environment := c.resolveCenter(ctx, tenantId, namespaceId)
	if instanceName == "" {
		info["subscriberCount"] = 0
		info["subscriptionCount"] = 0
		return
	}
	info["subscriberCount"] = len(listServiceSubscribers(instanceName, environment, namespaceId, groupName, serviceName))
	info["subscriptionCount"] = len(listServiceSubscriptions(instanceName, environment, namespaceId, groupName, serviceName))
}

func (c *ServiceController) attachServiceSubscribers(ctx *gin.Context, tenantId, namespaceId, groupName, serviceName string, info map[string]interface{}) {
	instanceName, environment := c.resolveCenter(ctx, tenantId, namespaceId)
	consumers := listServiceSubscribers(instanceName, environment, namespaceId, groupName, serviceName)
	upstreams := listServiceSubscriptions(instanceName, environment, namespaceId, groupName, serviceName)
	info["subscriberCount"] = len(consumers)
	info["subscribers"] = subscribersToMaps(consumers)
	info["subscriptionCount"] = len(upstreams)
	info["subscriptions"] = subscribersToMaps(upstreams)
}

func (c *ServiceController) attachServiceNodes(ctx *gin.Context, tenantId, namespaceId, groupName, serviceName string, info map[string]interface{}) {
	instanceName, environment := c.resolveCenter(ctx, tenantId, namespaceId)
	if instanceName != "" && servicecenterv3.GetPool() != nil {
		nodes, err := listNamingNodes(ctx.Request.Context(), tenantId, instanceName, environment, namespaceId, groupName, serviceName)
		if err == nil {
			overlayNodeStats(info, nodes)
			info["nodes"] = nodesToMaps(nodes)
		}
	}
}

func (c *ServiceController) mergeRuntimeOnlyServices(ctx *gin.Context, tenantId, namespaceId, groupName string, current []map[string]interface{}) ([]map[string]interface{}, int) {
	instanceName, environment := c.resolveCenter(ctx, tenantId, namespaceId)
	if instanceName == "" || servicecenterv3.GetPool() == nil {
		return current, 0
	}
	runtime, err := listNamingServices(ctx.Request.Context(), tenantId, instanceName, environment, namespaceId, groupName)
	if err != nil || len(runtime) == 0 {
		return current, 0
	}
	seen := make(map[string]struct{}, len(current))
	for _, row := range current {
		key := fmtServiceKey(asString(row["namespaceId"]), asString(row["groupName"]), asString(row["serviceName"]))
		seen[key] = struct{}{}
	}
	extra := 0
	for _, svc := range runtime {
		key := fmtServiceKey(svc.NamespaceID, svc.GroupName, svc.ServiceName)
		if _, ok := seen[key]; ok {
			continue
		}
		if row := runtimeServiceToMap(svc); row != nil {
			c.attachServiceSubscriberStats(ctx, tenantId, svc.NamespaceID, svc.GroupName, svc.ServiceName, row)
			current = append(current, row)
			extra++
		}
	}
	return current, extra
}

func (c *ServiceController) runtimeServiceDetail(ctx *gin.Context, tenantId, namespaceId, groupName, serviceName string) map[string]interface{} {
	instanceName, environment := c.resolveCenter(ctx, tenantId, namespaceId)
	if instanceName == "" || servicecenterv3.GetPool() == nil {
		return nil
	}
	svc, err := getNamingService(ctx.Request.Context(), tenantId, instanceName, environment, namespaceId, groupName, serviceName)
	if err != nil || svc == nil {
		return nil
	}
	info := runtimeServiceToMap(svc)
	info["nodes"] = nodesToMaps(svc.Nodes)
	c.attachServiceSubscribers(ctx, tenantId, namespaceId, groupName, serviceName, info)
	return info
}

func (c *ServiceController) updateRuntimeNode(ctx *gin.Context, operatorId string, req *catalog.ServiceNode) error {
	inst := &model.Node{
		NodeID:        req.NodeId,
		IP:            req.IpAddress,
		Port:          req.PortNumber,
		Weight:        req.Weight,
		Status:        req.InstanceStatus,
		HealthyStatus: req.HealthyStatus,
	}
	if req.MetadataJson != "" {
		var meta map[string]string
		if err := json.Unmarshal([]byte(req.MetadataJson), &meta); err == nil {
			inst.Metadata = meta
		}
	}
	if err := updateNamingNode(ctx.Request.Context(), operatorId, inst); err != nil {
		return err
	}
	if pool := servicecenterv3.GetPool(); pool != nil {
		if current, _, ok := pool.FindNode(req.NodeId); ok {
			req.NamespaceId = current.NamespaceID
			req.GroupName = current.GroupName
			req.ServiceName = current.ServiceName
			req.IpAddress = current.IP
			req.PortNumber = current.Port
			req.Weight = current.Weight
			req.InstanceStatus = current.Status
			req.HealthyStatus = current.HealthyStatus
			req.Ephemeral = model.YN(current.Ephemeral)
		}
	}
	return nil
}

func fmtServiceKey(namespaceId, groupName, serviceName string) string {
	return namespaceId + "/" + groupName + "/" + serviceName
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
