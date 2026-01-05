package controllers

import (
	"strings"
	"time"

	"gateway/internal/tunnel/static"
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

// StaticServerController 静态服务器控制器
type StaticServerController struct {
	db              database.Database
	staticServerDAO *dao.StaticServerDAO
}

// NewStaticServerController 创建静态服务器控制器实例
func NewStaticServerController(db database.Database) *StaticServerController {
	return &StaticServerController{
		db:              db,
		staticServerDAO: dao.NewStaticServerDAO(db),
	}
}

// QueryStaticServers 查询静态服务器列表
// @Summary 查询静态服务器列表
// @Description 分页查询静态服务器列表
// @Tags 静态隧道管理
// @Produce json
// @Param pageIndex query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/staticServers [get]
func (c *StaticServerController) QueryStaticServers(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 绑定查询条件
	var query models.StaticServerQueryRequest
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定静态服务器查询条件失败，使用默认条件", "error", err.Error())
	}
	query.PageIndex = page
	query.PageSize = pageSize

	// 调用DAO获取服务器列表
	servers, total, err := c.staticServerDAO.QueryStaticServers(&query)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取静态服务器列表失败", err)
		response.ErrorJSON(ctx, "获取静态服务器列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "tunnelStaticServerId"

	response.PageJSON(ctx, servers, pageInfo, constants.SD00002)
}

// GetStaticServer 获取静态服务器详情
// @Summary 获取静态服务器详情
// @Description 根据服务器ID获取静态服务器详细信息
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param request body object{tunnelStaticServerId=string} true "服务器ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/getStaticServer [post]
func (c *StaticServerController) GetStaticServer(ctx *gin.Context) {
	// 从请求体中获取服务器ID
	serverId := request.GetParam(ctx, "tunnelStaticServerId")
	if serverId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取服务器信息
	server, err := c.staticServerDAO.GetStaticServer(ctx, serverId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取静态服务器详情失败", err)
		response.ErrorJSON(ctx, "获取静态服务器详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if server == nil {
		response.ErrorJSON(ctx, "服务器不存在", constants.ED00008)
		return
	}

	response.SuccessJSON(ctx, server, constants.SD00002)
}

// CreateStaticServer 创建静态服务器
// @Summary 创建静态服务器
// @Description 创建新的静态服务器
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param server body types.TunnelStaticServer true "服务器信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/staticServers [post]
func (c *StaticServerController) CreateStaticServer(ctx *gin.Context) {
	var req types.TunnelStaticServer
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 参数验证
	if strings.TrimSpace(req.ServerName) == "" {
		response.ErrorJSON(ctx, "服务器名称不能为空", constants.ED00007)
		return
	}
	if req.ListenPort <= 0 || req.ListenPort > 65535 {
		response.ErrorJSON(ctx, "监听端口必须在1-65535之间", constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := strings.TrimSpace(req.TenantId)
	if tenantId == "" {
		tenantId = request.GetTenantID(ctx)
	}
	req.TenantId = tenantId

	// 检查服务器名称是否已存在
	exists, err := c.staticServerDAO.CheckServerNameExists(ctx, req.ServerName, "")
	if err != nil {
		logger.ErrorWithTrace(ctx, "检查服务器名称是否存在失败", err)
		response.ErrorJSON(ctx, "检查服务器名称是否存在失败: "+err.Error(), constants.ED00003)
		return
	}
	if exists {
		response.ErrorJSON(ctx, "服务器名称已存在", constants.ED00013)
		return
	}

	// 使用工具类获取操作人ID
	operatorId := request.GetOperatorID(ctx)

	// 生成ID和审计字段
	req.TunnelStaticServerId = random.Generate32BitRandomString()
	req.AddWho = operatorId
	req.EditWho = operatorId
	req.OprSeqFlag = random.Generate32BitRandomString()

	// 调用DAO创建服务器
	err = c.staticServerDAO.CreateStaticServer(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建静态服务器失败", err)
		response.ErrorJSON(ctx, "创建静态服务器失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新创建的服务器信息
	newServer, err := c.staticServerDAO.GetStaticServer(ctx, req.TunnelStaticServerId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的服务器信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"tunnelStaticServerId": req.TunnelStaticServerId,
			"message":              "服务器创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	response.SuccessJSON(ctx, newServer, constants.SD00003)
}

// UpdateStaticServer 更新静态服务器
// @Summary 更新静态服务器
// @Description 更新静态服务器信息
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param server body types.TunnelStaticServer true "服务器信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/staticServers [put]
func (c *StaticServerController) UpdateStaticServer(ctx *gin.Context) {
	var updateData types.TunnelStaticServer
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.TunnelStaticServerId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00007)
		return
	}
	if strings.TrimSpace(updateData.ServerName) == "" {
		response.ErrorJSON(ctx, "服务器名称不能为空", constants.ED00007)
		return
	}
	if updateData.ListenPort <= 0 || updateData.ListenPort > 65535 {
		response.ErrorJSON(ctx, "监听端口必须在1-65535之间", constants.ED00006)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 获取现有服务器信息
	currentServer, err := c.staticServerDAO.GetStaticServer(ctx, updateData.TunnelStaticServerId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务器信息失败", err)
		response.ErrorJSON(ctx, "获取服务器信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentServer == nil {
		response.ErrorJSON(ctx, "服务器不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段
	updateData.TenantId = currentServer.TenantId
	updateData.AddTime = currentServer.AddTime
	updateData.AddWho = currentServer.AddWho
	updateData.EditTime = time.Now()
	updateData.EditWho = operatorId

	// 调用DAO更新服务器
	err = c.staticServerDAO.UpdateStaticServer(ctx, &updateData)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新静态服务器失败", err)
		response.ErrorJSON(ctx, "更新静态服务器失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的服务器信息
	updatedServer, err := c.staticServerDAO.GetStaticServer(ctx, updateData.TunnelStaticServerId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的服务器信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	response.SuccessJSON(ctx, updatedServer, constants.SD00004)
}

// DeleteStaticServer 删除静态服务器
// @Summary 删除静态服务器
// @Description 删除静态服务器
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param request body object{tunnelStaticServerId=string} true "服务器ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/deleteStaticServer [post]
func (c *StaticServerController) DeleteStaticServer(ctx *gin.Context) {
	// 从请求体中获取服务器ID
	var req struct {
		TunnelStaticServerId string `json:"tunnelStaticServerId" form:"tunnelStaticServerId" query:"tunnelStaticServerId" binding:"required"`
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	serverId := req.TunnelStaticServerId
	if serverId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO删除服务器
	err := c.staticServerDAO.DeleteStaticServer(ctx, serverId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除静态服务器失败", err)
		response.ErrorJSON(ctx, "删除静态服务器失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"tunnelStaticServerId": serverId,
	}, constants.SD00005)
}

// GetStaticServerStats 获取服务器统计信息
// @Summary 获取服务器统计信息
// @Description 获取静态服务器统计信息
// @Tags 静态隧道管理
// @Produce json
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/getStaticServerStats [post]
func (c *StaticServerController) GetStaticServerStats(ctx *gin.Context) {
	stats, err := c.staticServerDAO.GetStaticServerStats(ctx)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务器统计信息失败", err)
		response.ErrorJSON(ctx, "获取服务器统计信息失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, stats, constants.SD00002)
}

// CheckPortConflict 检查端口冲突
// @Summary 检查端口冲突
// @Description 检查监听端口是否冲突
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param request body object{listenAddress=string,listenPort=int,serverType=string,excludeId=string} true "端口检查参数"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/checkPortConflict [post]
func (c *StaticServerController) CheckPortConflict(ctx *gin.Context) {
	listenAddress := request.GetParam(ctx, "listenAddress")
	listenPort := request.GetParamInt(ctx, "listenPort", 0)
	serverType := request.GetParam(ctx, "serverType")
	excludeId := request.GetParam(ctx, "excludeId")

	if listenAddress == "" {
		response.ErrorJSON(ctx, "监听地址不能为空", constants.ED00006)
		return
	}
	if listenPort <= 0 || listenPort > 65535 {
		response.ErrorJSON(ctx, "监听端口必须在1-65535之间", constants.ED00006)
		return
	}
	if serverType == "" {
		response.ErrorJSON(ctx, "服务器类型不能为空", constants.ED00006)
		return
	}

	conflict, err := c.staticServerDAO.CheckPortConflict(ctx, listenAddress, listenPort, serverType, excludeId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "检查端口冲突失败", err)
		response.ErrorJSON(ctx, "检查端口冲突失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{"conflict": conflict}, constants.SD00002)
}

// StartStaticServer 启动静态服务器
// @Summary 启动静态服务器
// @Description 启动指定的静态服务器
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param request body object{tunnelStaticServerId=string} true "服务器ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/startStaticServer [post]
func (c *StaticServerController) StartStaticServer(ctx *gin.Context) {
	serverId := request.GetParam(ctx, "tunnelStaticServerId")
	if serverId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00006)
		return
	}

	// 获取静态代理管理器
	manager := static.GetStaticProxyManager()
	if manager == nil {
		response.ErrorJSON(ctx, "静态代理管理器未初始化", constants.ED00009)
		return
	}

	// 启动服务器
	if err := manager.Start(ctx.Request.Context(), serverId); err != nil {
		logger.ErrorWithTrace(ctx, "启动静态服务器失败", err)
		response.ErrorJSON(ctx, "启动失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新服务器信息返回前端
	tenantId := request.GetTenantID(ctx)
	server, err := c.staticServerDAO.GetStaticServer(ctx, serverId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取服务器信息失败", "serverId", serverId, "error", err)
		response.SuccessJSON(ctx, gin.H{
			"tunnelStaticServerId": serverId,
			"message":              "启动成功",
		}, constants.SD00001)
		return
	}

	response.SuccessJSON(ctx, server, constants.SD00001)
}

// StopStaticServer 停止静态服务器
// @Summary 停止静态服务器
// @Description 停止指定的静态服务器
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param request body object{tunnelStaticServerId=string} true "服务器ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/stopStaticServer [post]
func (c *StaticServerController) StopStaticServer(ctx *gin.Context) {
	serverId := request.GetParam(ctx, "tunnelStaticServerId")
	if serverId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00006)
		return
	}

	// 获取静态代理管理器
	manager := static.GetStaticProxyManager()
	if manager == nil {
		response.ErrorJSON(ctx, "静态代理管理器未初始化", constants.ED00009)
		return
	}

	// 停止服务器
	if err := manager.Stop(ctx.Request.Context(), serverId); err != nil {
		logger.ErrorWithTrace(ctx, "停止静态服务器失败", err)
		response.ErrorJSON(ctx, "停止失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新服务器信息返回前端
	tenantId := request.GetTenantID(ctx)
	server, err := c.staticServerDAO.GetStaticServer(ctx, serverId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取服务器信息失败", "serverId", serverId, "error", err)
		response.SuccessJSON(ctx, gin.H{
			"tunnelStaticServerId": serverId,
			"message":              "停止成功",
		}, constants.SD00001)
		return
	}

	response.SuccessJSON(ctx, server, constants.SD00001)
}

// ReloadStaticServer 重载静态服务器配置
// @Summary 重载静态服务器配置
// @Description 重载指定静态服务器的配置
// @Tags 静态隧道管理
// @Accept json
// @Produce json
// @Param request body object{tunnelStaticServerId=string} true "服务器ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0061/reloadStaticServer [post]
func (c *StaticServerController) ReloadStaticServer(ctx *gin.Context) {
	serverId := request.GetParam(ctx, "tunnelStaticServerId")
	if serverId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00006)
		return
	}

	// 获取静态代理管理器
	manager := static.GetStaticProxyManager()
	if manager == nil {
		response.ErrorJSON(ctx, "静态代理管理器未初始化", constants.ED00009)
		return
	}

	// 重载配置
	if err := manager.Reload(ctx.Request.Context(), serverId); err != nil {
		logger.ErrorWithTrace(ctx, "重载静态服务器配置失败", err)
		response.ErrorJSON(ctx, "重载失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新服务器信息返回前端
	tenantId := request.GetTenantID(ctx)
	server, err := c.staticServerDAO.GetStaticServer(ctx, serverId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取服务器信息失败", "serverId", serverId, "error", err)
		response.SuccessJSON(ctx, gin.H{
			"tunnelStaticServerId": serverId,
			"message":              "重载成功",
		}, constants.SD00001)
		return
	}

	response.SuccessJSON(ctx, server, constants.SD00001)
}
