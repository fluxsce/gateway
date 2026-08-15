package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0022/dao"
	"gateway/web/views/hub0022/models"

	"github.com/gin-gonic/gin"
)

// CircuitBreakerConfigController 服务级熔断配置控制器。
type CircuitBreakerConfigController struct {
	db  database.Database
	dao *dao.CircuitBreakerConfigDAO
}

// NewCircuitBreakerConfigController 创建熔断配置控制器。
func NewCircuitBreakerConfigController(db database.Database) *CircuitBreakerConfigController {
	return &CircuitBreakerConfigController{
		db:  db,
		dao: dao.NewCircuitBreakerConfigDAO(db),
	}
}

// GetCircuitBreakerConfigRequest 按服务查询熔断配置。
type GetCircuitBreakerConfigRequest struct {
	TargetServiceId string `json:"targetServiceId" form:"targetServiceId" binding:"required"`
}

// GetCircuitBreakerConfig 按服务 ID 获取熔断配置，无行时返回空对象。
func (c *CircuitBreakerConfigController) GetCircuitBreakerConfig(ctx *gin.Context) {
	var req GetCircuitBreakerConfigRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}
	tenantId := request.GetTenantID(ctx)
	cfg, err := c.dao.GetByTargetServiceId(ctx, tenantId, req.TargetServiceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取熔断配置失败", err)
		response.ErrorJSON(ctx, "获取熔断配置失败: "+err.Error(), constants.ED00009)
		return
	}
	response.SuccessJSON(ctx, cfg, constants.SD00002)
}

// SaveCircuitBreakerConfig 新增或更新服务熔断阈值。
func (c *CircuitBreakerConfigController) SaveCircuitBreakerConfig(ctx *gin.Context) {
	var req models.CircuitBreakerConfig
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}
	if req.TargetServiceId == "" {
		response.ErrorJSON(ctx, "目标服务ID不能为空", constants.ED00007)
		return
	}
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)
	req.TenantId = tenantId
	if err := c.dao.UpsertByTargetServiceId(ctx, &req, operatorId); err != nil {
		logger.ErrorWithTrace(ctx, "保存熔断配置失败", err)
		response.ErrorJSON(ctx, "保存熔断配置失败: "+err.Error(), constants.ED00009)
		return
	}
	saved, err := c.dao.GetByTargetServiceId(ctx, tenantId, req.TargetServiceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取保存后的熔断配置失败", err)
		response.SuccessJSON(ctx, gin.H{"targetServiceId": req.TargetServiceId}, constants.SD00004)
		return
	}
	response.SuccessJSON(ctx, saved, constants.SD00004)
}
