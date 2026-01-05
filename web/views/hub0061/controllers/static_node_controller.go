package controllers

import (
	"strings"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0061/dao"
	"gateway/web/views/hub0061/models"

	"github.com/gin-gonic/gin"
)

// StaticNodeController 静态节点控制器
type StaticNodeController struct {
	db            database.Database
	staticNodeDAO *dao.StaticNodeDAO
}

// NewStaticNodeController 创建静态节点控制器实例
func NewStaticNodeController(db database.Database) *StaticNodeController {
	return &StaticNodeController{
		db:            db,
		staticNodeDAO: dao.NewStaticNodeDAO(db),
	}
}

// QueryStaticNodes 查询静态节点列表
// @Summary 查询静态节点列表
// @Description 分页查询静态节点列表
// @Tags 静态隧道管理
// @Produce json
// @Param pageIndex query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/staticNodes [get]
func (c *StaticNodeController) QueryStaticNodes(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 绑定查询条件
	var query models.StaticNodeQueryRequest
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定静态节点查询条件失败，使用默认条件", "error", err.Error())
	}
	query.PageIndex = page
	query.PageSize = pageSize

	// 调用DAO获取节点列表
	nodes, total, err := c.staticNodeDAO.QueryStaticNodes(&query)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取静态节点列表失败", err)
		response.ErrorJSON(ctx, "获取静态节点列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "tunnelStaticNodeId"

	response.PageJSON(ctx, nodes, pageInfo, constants.SD00002)
}

// GetStaticNode 获取静态节点详情
// @Summary 获取静态节点详情
// @Description 根据节点ID获取静态节点详细信息
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param request body object{tunnelStaticNodeId=string} true "节点ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/getStaticNode [post]
func (c *StaticNodeController) GetStaticNode(ctx *gin.Context) {
	// 从请求体中获取节点ID
	nodeId := request.GetParam(ctx, "tunnelStaticNodeId")
	if nodeId == "" {
		response.ErrorJSON(ctx, "节点ID不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取节点信息
	node, err := c.staticNodeDAO.GetStaticNode(ctx, nodeId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取静态节点详情失败", err)
		response.ErrorJSON(ctx, "获取静态节点详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if node == nil {
		response.ErrorJSON(ctx, "节点不存在", constants.ED00008)
		return
	}

	response.SuccessJSON(ctx, node, constants.SD00002)
}

// CreateStaticNode 创建静态节点
// @Summary 创建静态节点
// @Description 创建新的静态节点
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param node body types.TunnelStaticNode true "节点信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/staticNodes [post]
func (c *StaticNodeController) CreateStaticNode(ctx *gin.Context) {
	var req types.TunnelStaticNode
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 参数验证
	if strings.TrimSpace(req.NodeName) == "" {
		response.ErrorJSON(ctx, "节点名称不能为空", constants.ED00007)
		return
	}
	if strings.TrimSpace(req.TunnelStaticServerId) == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00007)
		return
	}
	if strings.TrimSpace(req.TargetAddress) == "" {
		response.ErrorJSON(ctx, "目标地址不能为空", constants.ED00007)
		return
	}
	if req.TargetPort <= 0 || req.TargetPort > 65535 {
		response.ErrorJSON(ctx, "目标端口必须在1-65535之间", constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := strings.TrimSpace(req.TenantId)
	if tenantId == "" {
		tenantId = request.GetTenantID(ctx)
	}
	req.TenantId = tenantId

	// 检查节点名称是否已存在（同一服务器下）
	exists, err := c.staticNodeDAO.CheckNodeNameExists(ctx, req.TunnelStaticServerId, req.NodeName, "")
	if err != nil {
		logger.ErrorWithTrace(ctx, "检查节点名称是否存在失败", err)
		response.ErrorJSON(ctx, "检查节点名称是否存在失败: "+err.Error(), constants.ED00003)
		return
	}
	if exists {
		response.ErrorJSON(ctx, "节点名称已存在", constants.ED00013)
		return
	}

	// 使用工具类获取操作人ID
	operatorId := request.GetOperatorID(ctx)

	// 生成ID和审计字段
	req.TunnelStaticNodeId = random.Generate32BitRandomString()
	req.AddWho = operatorId
	req.EditWho = operatorId
	req.OprSeqFlag = random.Generate32BitRandomString()

	// 调用DAO创建节点
	err = c.staticNodeDAO.CreateStaticNode(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建静态节点失败", err)
		response.ErrorJSON(ctx, "创建静态节点失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新创建的节点信息
	newNode, err := c.staticNodeDAO.GetStaticNode(ctx, req.TunnelStaticNodeId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的节点信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"tunnelStaticNodeId": req.TunnelStaticNodeId,
			"message":            "节点创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	response.SuccessJSON(ctx, newNode, constants.SD00003)
}

// UpdateStaticNode 更新静态节点
// @Summary 更新静态节点
// @Description 更新静态节点信息
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param node body types.TunnelStaticNode true "节点信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/staticNodes [put]
func (c *StaticNodeController) UpdateStaticNode(ctx *gin.Context) {
	var updateData types.TunnelStaticNode
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.TunnelStaticNodeId == "" {
		response.ErrorJSON(ctx, "节点ID不能为空", constants.ED00007)
		return
	}
	if strings.TrimSpace(updateData.NodeName) == "" {
		response.ErrorJSON(ctx, "节点名称不能为空", constants.ED00007)
		return
	}
	if strings.TrimSpace(updateData.TargetAddress) == "" {
		response.ErrorJSON(ctx, "目标地址不能为空", constants.ED00007)
		return
	}
	if updateData.TargetPort <= 0 || updateData.TargetPort > 65535 {
		response.ErrorJSON(ctx, "目标端口必须在1-65535之间", constants.ED00006)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 获取现有节点信息
	currentNode, err := c.staticNodeDAO.GetStaticNode(ctx, updateData.TunnelStaticNodeId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取节点信息失败", err)
		response.ErrorJSON(ctx, "获取节点信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentNode == nil {
		response.ErrorJSON(ctx, "节点不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段
	updateData.TenantId = currentNode.TenantId
	updateData.TunnelStaticServerId = currentNode.TunnelStaticServerId
	updateData.AddTime = currentNode.AddTime
	updateData.AddWho = currentNode.AddWho
	updateData.EditTime = time.Now()
	updateData.EditWho = operatorId

	// 调用DAO更新节点
	err = c.staticNodeDAO.UpdateStaticNode(ctx, &updateData)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新静态节点失败", err)
		response.ErrorJSON(ctx, "更新静态节点失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的节点信息
	updatedNode, err := c.staticNodeDAO.GetStaticNode(ctx, updateData.TunnelStaticNodeId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的节点信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	response.SuccessJSON(ctx, updatedNode, constants.SD00004)
}

// DeleteStaticNode 删除静态节点
// @Summary 删除静态节点
// @Description 删除静态节点
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param request body object{tunnelStaticNodeId=string} true "节点ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/deleteStaticNode [post]
func (c *StaticNodeController) DeleteStaticNode(ctx *gin.Context) {
	// 从请求体中获取节点ID
	var req struct {
		TunnelStaticNodeId string `json:"tunnelStaticNodeId" form:"tunnelStaticNodeId" query:"tunnelStaticNodeId" binding:"required"`
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	nodeId := req.TunnelStaticNodeId
	if nodeId == "" {
		response.ErrorJSON(ctx, "节点ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO删除节点
	err := c.staticNodeDAO.DeleteStaticNode(ctx, nodeId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除静态节点失败", err)
		response.ErrorJSON(ctx, "删除静态节点失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"tunnelStaticNodeId": nodeId,
	}, constants.SD00005)
}

// GetStaticNodeStats 获取节点统计信息
// @Summary 获取节点统计信息
// @Description 获取静态节点统计信息
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param request body object{tunnelStaticServerId=string} false "服务器ID（可选）"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/getStaticNodeStats [post]
func (c *StaticNodeController) GetStaticNodeStats(ctx *gin.Context) {
	serverId := request.GetParam(ctx, "tunnelStaticServerId")

	stats, err := c.staticNodeDAO.GetStaticNodeStats(ctx, serverId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取节点统计信息失败", err)
		response.ErrorJSON(ctx, "获取节点统计信息失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, stats, constants.SD00002)
}
