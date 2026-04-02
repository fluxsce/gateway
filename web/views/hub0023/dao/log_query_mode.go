package dao

import (
	"context"
	"strings"

	"gateway/internal/gateway/loader/dbloader"
	"gateway/internal/gateway/logwrite"
	gwtypes "gateway/internal/gateway/logwrite/types"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
)

// ResolveGatewayLogQueryType 根据租户与网关实例解析网关日志查询应使用的存储类型。
// gatewayInstanceId 为空时，取该租户下按创建时间倒序的第一条实例（与实例列表默认排序一致），
// 并读取其关联日志配置的输出目标。
// 返回值与 app.gateway.log_query_type 一致：mongo、clickhouse、database。
func ResolveGatewayLogQueryType(ctx context.Context, db database.Database, tenantID, gatewayInstanceID string) string {
	fallback := normalizeLogQueryType(config.GetString("app.gateway.log_query_type", "database"))
	if tenantID == "" {
		return fallback
	}
	logConfigID, ok := resolveLogConfigIDForGatewayLogQuery(ctx, db, tenantID, gatewayInstanceID)
	if !ok || logConfigID == "" {
		return fallback
	}
	cfg := loadLogConfigViaLoader(ctx, db, tenantID, logConfigID)
	t := mapLogConfigToQueryType(ctx, cfg)
	if t == "" {
		return fallback
	}
	return t
}

// resolveLogConfigIDForGatewayLogQuery 解析用于日志查询的日志配置 ID。
// 第二个返回值为 false 表示未找到实例或查询失败，调用方应使用全局回退配置。
func resolveLogConfigIDForGatewayLogQuery(ctx context.Context, db database.Database, tenantID, gatewayInstanceID string) (string, bool) {
	if gatewayInstanceID != "" {
		var row struct {
			LogConfigID string `db:"logConfigId"`
		}
		q := `SELECT logConfigId FROM HUB_GW_INSTANCE WHERE tenantId = ? AND gatewayInstanceId = ?`
		err := db.QueryOne(ctx, &row, q, []interface{}{tenantID, gatewayInstanceID}, true)
		if err == database.ErrRecordNotFound {
			return "", false
		}
		if err != nil {
			return "", false
		}
		return row.LogConfigID, true
	}

	dbType := sqlutils.GetDatabaseType(db)
	pagination := sqlutils.NewPaginationInfo(1, 1)
	baseQuery := `
		SELECT logConfigId FROM HUB_GW_INSTANCE
		WHERE tenantId = ?
		ORDER BY addTime DESC
	`
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return "", false
	}
	params := []interface{}{tenantID}
	params = append(params, paginationArgs...)

	var row struct {
		LogConfigID string `db:"logConfigId"`
	}
	err = db.QueryOne(ctx, &row, paginatedQuery, params, true)
	if err == database.ErrRecordNotFound {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return row.LogConfigID, true
}

// loadLogConfigViaLoader 通过 dbloader.LogConfigLoader 加载完整日志配置，
// 与网关运行时加载 HUB_GW_LOG_CONFIG 的查询与转换逻辑保持一致。
func loadLogConfigViaLoader(ctx context.Context, db database.Database, tenantID, logConfigID string) *gwtypes.LogConfig {
	loader := dbloader.NewLogConfigLoader(db, tenantID)
	cfg, err := loader.LoadLogConfig(ctx, logConfigID)
	if err != nil {
		logger.WarnWithTrace(ctx, "加载日志配置失败，网关日志查询模式将使用全局回退", "logConfigId", logConfigID, "error", err)
		return nil
	}
	return cfg
}

// mapLogConfigToQueryType 使用 gwtypes.LogConfig.GetOutputTargets() 解析 outputTargets（与网关一致），
// 再结合 logwrite.IsTargetSupported 与 ValidateConfig「仅允许单一输出目标」的语义映射到查询后端。
// 多个逗号分隔目标时返回空并回退（ValidateConfig 同样不允许）；单一目标为 MONGODB、CLICKHOUSE、DATABASE 时分别对应 mongo、clickhouse、database。
func mapLogConfigToQueryType(ctx context.Context, cfg *gwtypes.LogConfig) string {
	if cfg == nil {
		return ""
	}
	targets := cfg.GetOutputTargets()
	if len(targets) > 1 {
		logger.WarnWithTrace(ctx, "日志配置 outputTargets 含多个目标，与网关 ValidateConfig 仅允许单目标不一致，网关日志查询使用全局回退", "outputTargets", cfg.OutputTargets)
		return ""
	}
	if len(targets) == 0 {
		return ""
	}
	target := targets[0]
	if !logwrite.IsTargetSupported(target) {
		logger.WarnWithTrace(ctx, "日志配置输出目标不受网关支持，网关日志查询使用全局回退", "target", string(target))
		return ""
	}
	switch target {
	case gwtypes.LogOutputMongoDB:
		return "mongo"
	case gwtypes.LogOutputClickHouse:
		return "clickhouse"
	case gwtypes.LogOutputDatabase:
		return "database"
	default:
		// CONSOLE、FILE、ELASTICSEARCH 等：无对应 hub0023 列表查询实现或与 ValidateConfig 中子配置校验无关的查询路由
		return ""
	}
}

// normalizeLogQueryType 将配置中的日志查询类型规范为 mongo、clickhouse、database。
func normalizeLogQueryType(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "mongo", "mongodb":
		return "mongo"
	case "clickhouse":
		return "clickhouse"
	default:
		return "database"
	}
}
