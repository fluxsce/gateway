package dbloader

import (
	"context"
	"fmt"

	"gateway/internal/gateway/handler/statichost"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
)

// StaticHostConfigLoader 从数据库加载路由级本机目录托管配置。
type StaticHostConfigLoader struct {
	db       database.Database
	tenantId string
}

// NewStaticHostConfigLoader 创建静态托管配置加载器。
func NewStaticHostConfigLoader(db database.Database, tenantId string) *StaticHostConfigLoader {
	return &StaticHostConfigLoader{
		db:       db,
		tenantId: tenantId,
	}
}

// LoadRouteStaticHostConfig 按路由 ID 加载一条活动静态托管配置。
// 无行时返回 nil, nil，该路由继续走代理。
func (loader *StaticHostConfigLoader) LoadRouteStaticHostConfig(ctx context.Context, routeId string) (*statichost.StaticHostConfig, error) {
	if routeId == "" {
		return nil, nil
	}

	baseQuery := `
		SELECT tenantId, staticHostConfigId, routeConfigId, configName, rootDirectory,
		       stripRoutePrefix, indexFiles, rewriteRules, spaFallback, cacheControlMaxAge,
		       allowedExtensions, maxFileSizeBytes, followSymlinks, enablePrecompress,
		       errorPage404, errorPage403, configPriority
		FROM HUB_GW_STATIC_HOST_CONFIG
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
	`

	pagination := sqlutils.NewPaginationInfo(1, 1)
	dbType := sqlutils.GetDatabaseType(loader.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, fmt.Errorf("构建静态托管配置分页查询失败: %w", err)
	}

	allArgs := append([]interface{}{loader.tenantId, routeId}, paginationArgs...)
	var records []StaticHostConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询路由静态托管配置失败: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	return mapStaticHostRecord(records[0]), nil
}

// mapStaticHostRecord 把表记录映射为运行时静态托管配置。
func mapStaticHostRecord(record StaticHostConfigRecord) *statichost.StaticHostConfig {
	cfg := &statichost.StaticHostConfig{
		Enabled:            true,
		ID:                 record.StaticHostConfigId,
		Name:               record.ConfigName,
		RootDirectory:      record.RootDirectory,
		StripRoutePrefix:   record.StripRoutePrefix == "Y",
		IndexFiles:         parseIndexFiles(record.IndexFiles),
		RewriteRules:       parseRewriteRules(record.RewriteRules),
		SPAFallback:        record.SpaFallback == "Y",
		CacheControlMaxAge: record.CacheControlMaxAge,
		AllowedExtensions:  parseAllowedExtensions(record.AllowedExtensions),
		MaxFileSizeBytes:   record.MaxFileSizeBytes,
		FollowSymlinks:     record.FollowSymlinks == "Y",
		EnablePrecompress:  record.EnablePrecompress != "N",
		ErrorPage404:       stringValue(record.ErrorPage404),
		ErrorPage403:       stringValue(record.ErrorPage403),
	}
	return cfg
}

// parseIndexFiles 解析索引文件列表，兼容 JSON 数组与逗号分隔文本。
func parseIndexFiles(raw *string) []string {
	if raw == nil {
		return []string{"index.html"}
	}
	return statichost.ParseIndexFilesText(*raw)
}

// parseRewriteRules 解析静态托管路径重写规则。
func parseRewriteRules(raw *string) []statichost.RewriteRule {
	if raw == nil {
		return nil
	}
	return statichost.ParseRewriteRulesText(*raw)
}

func stringValue(raw *string) string {
	if raw == nil {
		return ""
	}
	return *raw
}

func parseAllowedExtensions(raw *string) []string {
	if raw == nil {
		return nil
	}
	return statichost.ParseAllowedExtensionsText(*raw)
}
