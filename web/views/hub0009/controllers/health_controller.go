package controllers

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"gateway/pkg/cache"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	mongofactory "gateway/pkg/mongo/factory"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0009/models"

	"github.com/gin-gonic/gin"
)

const (
	healthStatusOK       = "ok"
	healthStatusError    = "error"
	healthStatusSkipped  = "skipped"
	healthStatusDegraded = "degraded"

	healthKindProcess    = "process"
	healthKindSQL        = "sql"
	healthKindClickHouse = "clickhouse"
	healthKindMongo      = "mongo"
	healthKindCache      = "cache"

	healthPingTimeout = 3 * time.Second
	clickHouseConn    = "clickhouse_main"
	cacheDefaultAlias = "default"
)

var healthKindOrder = map[string]int{
	healthKindProcess:    0,
	healthKindSQL:        1,
	healthKindClickHouse: 2,
	healthKindMongo:      3,
	healthKindCache:      4,
}

// GetSystemHealth 探测本进程已装配的平台库、ClickHouse、Mongo 与缓存是否可达。
// 未配置的依赖记为 skipped，不把整体状态打成 degraded。
func (c *SettingController) GetSystemHealth(ctx *gin.Context) {
	if request.GetTenantID(ctx) == "" {
		response.ErrorJSON(ctx, "租户不能为空", constants.ED00006)
		return
	}

	items := make([]models.HealthItem, 0, 8)
	items = append(items, models.HealthItem{
		Name:    "controlPlane",
		Kind:    healthKindProcess,
		Driver:  "web",
		Status:  healthStatusOK,
		Enabled: true,
	})
	items = append(items, collectSQLHealth(ctx.Request.Context())...)
	items = append(items, collectMongoHealth(ctx.Request.Context())...)
	items = append(items, collectCacheHealth(ctx.Request.Context())...)
	sortHealthItems(items)

	response.SuccessJSON(ctx, models.HealthResponse{
		Status:    overallHealthStatus(items),
		CheckedAt: time.Now().Unix(),
		Items:     items,
	}, constants.SD00002)
}

// collectSQLHealth 探测默认库与 ClickHouse。两边同名时只报一条。
func collectSQLHealth(parent context.Context) []models.HealthItem {
	items := make([]models.HealthItem, 0, 2)
	defaultName := config.GetString("database.default", "")
	if db := database.GetDefaultConnection(); db != nil {
		name := db.GetName()
		if name == "" {
			name = defaultName
		}
		items = append(items, pingHealth(parent, name, healthKindSQL, db.GetDriver(), db.Ping))
	} else {
		items = append(items, skippedHealth(defaultName, healthKindSQL, "", "平台库未初始化"))
	}

	if clickHouseConn == defaultName {
		return items
	}
	meta := clickHouseMeta()
	if db := database.GetConnection(clickHouseConn); db != nil {
		item := pingHealth(parent, clickHouseConn, healthKindClickHouse, db.GetDriver(), db.Ping)
		items = append(items, attachStore(item, meta, inspectClickHouse(parent, db)))
		return items
	}
	items = append(items, attachStore(
		skippedHealth(clickHouseConn, healthKindClickHouse, "clickhouse", meta.skipMessage()),
		meta,
		plannedClickHouseObjects(),
	))
	return items
}

// collectMongoHealth 探测已加载的 Mongo 连接；未启用时记一条 skipped。
func collectMongoHealth(parent context.Context) []models.HealthItem {
	names := mongofactory.ListConnections()
	if len(names) == 0 {
		name := config.GetString("mongo.default", "mongo_main")
		meta := mongoMeta(name)
		return []models.HealthItem{attachStore(
			skippedHealth(name, healthKindMongo, "mongo", meta.skipMessage()),
			meta,
			plannedMongoObjects(),
		)}
	}
	items := make([]models.HealthItem, 0, len(names))
	for _, name := range names {
		if name == cacheDefaultAlias {
			continue
		}
		meta := mongoMeta(name)
		cli, err := mongofactory.GetConnection(name)
		if err != nil || cli == nil {
			if err != nil {
				logger.Warn("系统健康探测失败", "name", name, "kind", healthKindMongo, "error", err.Error())
			}
			items = append(items, attachStore(models.HealthItem{
				Name:    name,
				Kind:    healthKindMongo,
				Driver:  "mongo",
				Status:  healthStatusError,
				Message: "连接不可用",
			}, meta, plannedMongoObjects()))
			continue
		}
		item := pingHealth(parent, name, healthKindMongo, "mongo", cli.Ping)
		items = append(items, attachStore(item, meta, inspectMongo(parent, name, cli)))
	}
	if len(items) > 0 {
		return items
	}
	name := config.GetString("mongo.default", "mongo_main")
	meta := mongoMeta(name)
	return []models.HealthItem{attachStore(
		skippedHealth(name, healthKindMongo, "mongo", meta.skipMessage()),
		meta,
		plannedMongoObjects(),
	)}
}

// collectCacheHealth 探测已加载的缓存。跳过 default 别名，避免与真实连接名重复。
func collectCacheHealth(parent context.Context) []models.HealthItem {
	names := cache.GetGlobalManager().ListCaches()
	items := make([]models.HealthItem, 0, len(names))
	for _, name := range names {
		if name == cacheDefaultAlias {
			continue
		}
		inst := cache.GetCache(name)
		if inst == nil {
			items = append(items, skippedHealth(name, healthKindCache, "", "连接不可用"))
			continue
		}
		items = append(items, pingHealth(parent, name, healthKindCache, inst.GetCacheType(), inst.Ping))
	}
	if len(items) > 0 {
		return items
	}
	if inst := cache.GetDefaultCache(); inst != nil {
		return []models.HealthItem{pingHealth(parent, cacheDefaultAlias, healthKindCache, inst.GetCacheType(), inst.Ping)}
	}
	name := config.GetString("cache.default", cacheDefaultAlias)
	return []models.HealthItem{skippedHealth(name, healthKindCache, "", "未配置")}
}

// pingHealth 对单个依赖做带超时的 Ping，记录耗时；失败只回错误文案。
func pingHealth(parent context.Context, name, kind, driver string, ping func(context.Context) error) models.HealthItem {
	item := models.HealthItem{Name: name, Kind: kind, Driver: driver, Status: healthStatusOK, Enabled: true}
	ctx, cancel := context.WithTimeout(parent, healthPingTimeout)
	defer cancel()
	start := time.Now()
	err := ping(ctx)
	item.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		logger.Warn("系统健康探测失败", "name", name, "kind", kind, "error", err.Error())
		item.Status = healthStatusError
		item.Message = sanitizeHealthError(err)
	}
	return item
}

// sanitizeHealthError 只回分类文案，避免把主机、账号或连接串送到浏览器。
func sanitizeHealthError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "探测超时"
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded"):
		return "探测超时"
	case strings.Contains(msg, "auth") || strings.Contains(msg, "password") ||
		strings.Contains(msg, "access denied") || strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "noauth"):
		return "认证失败"
	default:
		return "连接失败"
	}
}

// skippedHealth 返回未装配依赖的占位结果。
func skippedHealth(name, kind, driver, message string) models.HealthItem {
	return models.HealthItem{
		Name:    name,
		Kind:    kind,
		Driver:  driver,
		Status:  healthStatusSkipped,
		Message: message,
	}
}

// overallHealthStatus 已配置组件全部可达为 ok，任一失败为 degraded。
func overallHealthStatus(items []models.HealthItem) string {
	for _, item := range items {
		if item.Status == healthStatusError {
			return healthStatusDegraded
		}
	}
	return healthStatusOK
}

// sortHealthItems 按组件类型再按连接名排序，保证页上顺序稳定。
func sortHealthItems(items []models.HealthItem) {
	slices.SortFunc(items, func(a, b models.HealthItem) int {
		ao, bo := healthKindOrder[a.Kind], healthKindOrder[b.Kind]
		if ao != bo {
			return ao - bo
		}
		if a.Name < b.Name {
			return -1
		}
		if a.Name > b.Name {
			return 1
		}
		return 0
	})
}
