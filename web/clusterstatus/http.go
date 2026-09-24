package clusterstatus

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"

	"github.com/gin-gonic/gin"
)

// TargetLoader 从当前请求解析本模块要探测的监听口。
// 各模块自己读自己的主键和表，不要把字段名写进共用处理函数。
type TargetLoader func(ctx *gin.Context, tenantId string) ([]ListenTarget, error)

// Register 在某个模块路由组上挂拓扑接口。buttonCode 只校验这一模块的按钮。
func Register(group *gin.RouterGroup, db database.Database, buttonCode string, load TargetLoader) {
	group.POST("/queryClusterTopology", routes.RequireButton(buttonCode), Handle(db, load))
}

// GatewayInstanceLoader 网关实例页的参数解析：请求体里的 gatewayInstanceId。
func GatewayInstanceLoader(db database.Database) TargetLoader {
	return func(ctx *gin.Context, tenantId string) ([]ListenTarget, error) {
		gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
		if gatewayInstanceId == "" {
			return nil, errNeedInstance
		}
		return LoadGatewayInstance(ctx, db, tenantId, gatewayInstanceId)
	}
}

// Handle 查询集群节点并探测调用方给出的监听口。
func Handle(db database.Database, load TargetLoader) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantId := request.GetTenantID(ctx)
		targets, err := load(ctx, tenantId)
		if err != nil {
			response.ErrorJSON(ctx, err.Error(), constants.ED00009)
			return
		}
		view, err := Query(ctx, db, tenantId, targets)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询集群健康与节点失败", err)
			response.ErrorJSON(ctx, "查询集群健康与节点失败: "+err.Error(), constants.ED00009)
			return
		}
		response.SuccessJSON(ctx, view, constants.SD00002)
	}
}
