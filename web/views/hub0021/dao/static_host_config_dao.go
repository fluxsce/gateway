package dao

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gateway/internal/gateway/handler/statichost"
	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0021/models"
)

// StaticHostConfigDAO 本机目录托管配置数据访问对象。
type StaticHostConfigDAO struct {
	db database.Database
}

// NewStaticHostConfigDAO 创建静态托管配置 DAO。
func NewStaticHostConfigDAO(db database.Database) *StaticHostConfigDAO {
	return &StaticHostConfigDAO{db: db}
}

// GetByRouteConfigId 按路由 ID 取一条静态托管配置，含已停用行以便重新启用。
func (dao *StaticHostConfigDAO) GetByRouteConfigId(ctx context.Context, tenantId, routeConfigId string) (*models.StaticHostConfig, error) {
	if routeConfigId == "" {
		return nil, errors.New("routeConfigId不能为空")
	}
	query := `
		SELECT * FROM HUB_GW_STATIC_HOST_CONFIG
		WHERE tenantId = ? AND routeConfigId = ?
		ORDER BY activeFlag DESC, configPriority ASC
	`
	var config models.StaticHostConfig
	err := dao.db.QueryOne(ctx, &config, query, []interface{}{tenantId, routeConfigId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询路由静态托管配置失败")
	}
	return &config, nil
}

// UpsertByRouteConfigId 按路由 ID 新增或更新静态托管配置，并置为活动。
func (dao *StaticHostConfigDAO) UpsertByRouteConfigId(ctx context.Context, config *models.StaticHostConfig, operatorId string) error {
	if config == nil {
		return errors.New("静态托管配置不能为空")
	}
	if config.RouteConfigId == "" {
		return errors.New("routeConfigId不能为空")
	}
	if strings.TrimSpace(config.RootDirectory) == "" {
		return errors.New("rootDirectory不能为空")
	}

	existing, err := dao.GetByRouteConfigId(ctx, config.TenantId, config.RouteConfigId)
	if err != nil {
		return err
	}

	now := time.Now()
	applyStaticHostDefaults(config)
	config.ActiveFlag = "Y"
	config.EditTime = now
	config.EditWho = operatorId
	config.OprSeqFlag = random.Generate32BitRandomString()

	if existing == nil {
		if config.StaticHostConfigId == "" {
			config.StaticHostConfigId = random.GenerateUniqueStringWithPrefix("SH", 32)
		}
		config.AddTime = now
		config.AddWho = operatorId
		config.CurrentVersion = 1
		_, err = dao.db.Insert(ctx, "HUB_GW_STATIC_HOST_CONFIG", config, true)
		if err != nil {
			return huberrors.WrapError(err, "添加静态托管配置失败")
		}
		return nil
	}

	config.StaticHostConfigId = existing.StaticHostConfigId
	config.AddTime = existing.AddTime
	config.AddWho = existing.AddWho
	config.CurrentVersion = existing.CurrentVersion + 1
	sql := `
		UPDATE HUB_GW_STATIC_HOST_CONFIG SET
			configName = ?, rootDirectory = ?, stripRoutePrefix = ?, indexFiles = ?,
			rewriteRules = ?, spaFallback = ?, cacheControlMaxAge = ?, allowedExtensions = ?,
			maxFileSizeBytes = ?, followSymlinks = ?, enablePrecompress = ?,
			errorPage404 = ?, errorPage403 = ?, configPriority = ?, noteText = ?,
			editTime = ?, editWho = ?, currentVersion = ?, oprSeqFlag = ?, activeFlag = ?
		WHERE tenantId = ? AND staticHostConfigId = ? AND currentVersion = ?
	`
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		config.ConfigName, config.RootDirectory, config.StripRoutePrefix, config.IndexFiles,
		config.RewriteRules, config.SpaFallback, config.CacheControlMaxAge, config.AllowedExtensions,
		config.MaxFileSizeBytes, config.FollowSymlinks, config.EnablePrecompress,
		config.ErrorPage404, config.ErrorPage403, config.ConfigPriority, config.NoteText,
		config.EditTime, config.EditWho, config.CurrentVersion, config.OprSeqFlag, config.ActiveFlag,
		config.TenantId, config.StaticHostConfigId, existing.CurrentVersion,
	}, true)
	if err != nil {
		return huberrors.WrapError(err, "更新静态托管配置失败")
	}
	if result == 0 {
		return errors.New("静态托管配置已被其他用户修改，请刷新后重试")
	}
	return nil
}

// DeactivateByRouteConfigId 将路由下活动的静态托管置为停用。
// 路由改为服务代理时调用，避免数据面仍优先出文件。
func (dao *StaticHostConfigDAO) DeactivateByRouteConfigId(ctx context.Context, tenantId, routeConfigId, operatorId string) error {
	if routeConfigId == "" {
		return nil
	}
	sql := `
		UPDATE HUB_GW_STATIC_HOST_CONFIG
		SET activeFlag = 'N', editTime = ?, editWho = ?, oprSeqFlag = ?
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
	`
	_, err := dao.db.Exec(ctx, sql, []interface{}{
		time.Now(), operatorId, random.Generate32BitRandomString(), tenantId, routeConfigId,
	}, true)
	if err != nil {
		return huberrors.WrapError(err, "停用路由静态托管配置失败")
	}
	return nil
}

// DeleteByRouteConfigId 删除路由时同步删除静态托管配置。
func (dao *StaticHostConfigDAO) DeleteByRouteConfigId(ctx context.Context, tenantId, routeConfigId string) error {
	if routeConfigId == "" {
		return nil
	}
	sql := `DELETE FROM HUB_GW_STATIC_HOST_CONFIG WHERE tenantId = ? AND routeConfigId = ?`
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, routeConfigId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除静态托管配置失败")
	}
	return nil
}

func applyStaticHostDefaults(config *models.StaticHostConfig) {
	if config.ConfigName == "" {
		config.ConfigName = "static-" + config.RouteConfigId
	}
	if config.StripRoutePrefix == "" {
		config.StripRoutePrefix = "Y"
	}
	if config.SpaFallback == "" {
		config.SpaFallback = "N"
	}
	if config.FollowSymlinks == "" {
		config.FollowSymlinks = "N"
	}
	if config.EnablePrecompress == "" {
		config.EnablePrecompress = "Y"
	}
	if config.CacheControlMaxAge < 0 {
		config.CacheControlMaxAge = 0
	}
	if config.MaxFileSizeBytes < 0 {
		config.MaxFileSizeBytes = 0
	}
	config.IndexFiles = normalizeIndexFiles(config.IndexFiles)
	config.RewriteRules = statichost.EncodeRewriteRulesJSON(statichost.ParseRewriteRulesText(config.RewriteRules))
	config.AllowedExtensions = encodeAllowedExtensions(config.AllowedExtensions)
	config.ErrorPage404 = strings.TrimSpace(config.ErrorPage404)
	config.ErrorPage403 = strings.TrimSpace(config.ErrorPage403)
}

func encodeAllowedExtensions(raw string) string {
	items := statichost.ParseAllowedExtensionsText(raw)
	if len(items) == 0 {
		return ""
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func normalizeIndexFiles(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return `["index.html"]`
	}
	var fromJSON []string
	if err := json.Unmarshal([]byte(text), &fromJSON); err == nil && len(fromJSON) > 0 {
		encoded, err := json.Marshal(fromJSON)
		if err == nil {
			return string(encoded)
		}
	}
	parts := strings.Split(text, ",")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		cleaned = append(cleaned, name)
	}
	if len(cleaned) == 0 {
		return `["index.html"]`
	}
	encoded, err := json.Marshal(cleaned)
	if err != nil {
		return `["index.html"]`
	}
	return string(encoded)
}
