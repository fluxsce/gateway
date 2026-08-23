package controllers

import (
	"gateway/internal/gateway/handler/statichost"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/middleware/audit"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0021/dao"
	"gateway/web/views/hub0021/models"

	"github.com/gin-gonic/gin"
)

// StaticHostConfigController 路由级本机目录托管配置控制器。
type StaticHostConfigController struct {
	db  database.Database
	dao *dao.StaticHostConfigDAO
}

// NewStaticHostConfigController 创建静态托管配置控制器。
func NewStaticHostConfigController(db database.Database) *StaticHostConfigController {
	return &StaticHostConfigController{
		db:  db,
		dao: dao.NewStaticHostConfigDAO(db),
	}
}

// GetStaticHostConfigRequest 按路由查询静态托管配置。
type GetStaticHostConfigRequest struct {
	RouteConfigId string `json:"routeConfigId" form:"routeConfigId" binding:"required"`
}

// GetStaticHostConfig 按路由 ID 获取静态托管配置，无行时返回空对象。
func (c *StaticHostConfigController) GetStaticHostConfig(ctx *gin.Context) {
	var req GetStaticHostConfigRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}
	tenantId := request.GetTenantID(ctx)
	cfg, err := c.dao.GetByRouteConfigId(ctx, tenantId, req.RouteConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取静态托管配置失败", err)
		response.ErrorJSON(ctx, "获取静态托管配置失败: "+err.Error(), constants.ED00009)
		return
	}
	response.SuccessJSON(ctx, cfg, constants.SD00002)
}

// SaveStaticHostConfig 新增或更新路由静态托管配置。
func (c *StaticHostConfigController) SaveStaticHostConfig(ctx *gin.Context) {
	var req models.StaticHostConfig
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}
	if req.RouteConfigId == "" {
		response.ErrorJSON(ctx, "路由配置ID不能为空", constants.ED00007)
		return
	}
	if err := statichost.ValidateForSave(req.RootDirectory, req.IndexFiles, req.RewriteRules); err != nil {
		response.ErrorJSON(ctx, "静态托管配置无效: "+err.Error(), constants.ED00007)
		return
	}
	if err := statichost.ValidateSecurityOptions(req.AllowedExtensions, req.MaxFileSizeBytes); err != nil {
		response.ErrorJSON(ctx, "静态托管配置无效: "+err.Error(), constants.ED00007)
		return
	}
	if err := statichost.ValidateErrorPages(req.ErrorPage404, req.ErrorPage403); err != nil {
		response.ErrorJSON(ctx, "静态托管配置无效: "+err.Error(), constants.ED00007)
		return
	}
	if err := statichost.ValidateFallbackRoots(req.FallbackRoots); err != nil {
		response.ErrorJSON(ctx, "备用目录无效: "+err.Error(), constants.ED00007)
		return
	}
	if err := statichost.ValidateCacheControlByExt(req.CacheControlByExt); err != nil {
		response.ErrorJSON(ctx, "按类型缓存无效: "+err.Error(), constants.ED00007)
		return
	}
	if err := statichost.ValidateSecurityHeaders(req.SecurityHeaders); err != nil {
		response.ErrorJSON(ctx, "页面安全头无效: "+err.Error(), constants.ED00007)
		return
	}
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)
	req.TenantId = tenantId
	if err := c.dao.UpsertByRouteConfigId(ctx, &req, operatorId); err != nil {
		logger.ErrorWithTrace(ctx, "保存静态托管配置失败", err)
		response.ErrorJSON(ctx, "保存静态托管配置失败: "+err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionUpdate,
		ModuleCode:   "hub0021",
		TargetType:   "STATIC_HOST",
		TargetId:     req.RouteConfigId,
		ResourceCode: "hub0021:staticHostConfig",
	})
	saved, err := c.dao.GetByRouteConfigId(ctx, tenantId, req.RouteConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取保存后的静态托管配置失败", err)
		response.SuccessJSON(ctx, gin.H{"routeConfigId": req.RouteConfigId}, constants.SD00004)
		return
	}
	response.SuccessJSON(ctx, saved, constants.SD00004)
}
