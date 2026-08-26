package controllers

import (
	"encoding/json"
	"strings"

	"gateway/pkg/logger"
	"gateway/pkg/syssetting"
	"gateway/web/middleware/audit"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0009/dao"
	"gateway/web/views/hub0009/models"

	"github.com/gin-gonic/gin"
)

const envVarSecretMask = "********"

// SaveEnvVar 新增或更新一条全局环境变量。密文变量加密入库，编辑时值留空表示保持原值。
func (c *SettingController) SaveEnvVar(ctx *gin.Context) {
	var req models.SaveEnvVarRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "租户不能为空", constants.ED00006)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.OriginalName = strings.TrimSpace(req.OriginalName)
	req.Note = strings.TrimSpace(req.Note)
	if err := syssetting.ValidateEnvVarName(req.Name); err != nil {
		response.ErrorJSON(ctx, err.Error(), constants.ED00006)
		return
	}
	if req.OriginalName != "" {
		if err := syssetting.ValidateEnvVarName(req.OriginalName); err != nil {
			response.ErrorJSON(ctx, err.Error(), constants.ED00006)
			return
		}
	}

	settings, version, err := c.loadEnvVars(ctx, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询环境变量失败", err)
		response.ErrorJSON(ctx, "查询环境变量失败: "+err.Error(), constants.ED00009)
		return
	}
	if req.CurrentVersion != version {
		response.ErrorJSON(ctx, dao.ErrVersionConflict.Error(), constants.ED00006)
		return
	}

	lookupName := req.OriginalName
	if lookupName == "" {
		lookupName = req.Name
	}
	existing, existIdx := findEnvVar(settings.Items, lookupName)
	renaming := existing != nil && req.Name != lookupName
	if renaming {
		if _, otherIdx := findEnvVar(settings.Items, req.Name); otherIdx >= 0 {
			response.ErrorJSON(ctx, "环境变量名已存在: "+req.Name, constants.ED00006)
			return
		}
	}
	if existing == nil {
		if _, otherIdx := findEnvVar(settings.Items, req.Name); otherIdx >= 0 {
			response.ErrorJSON(ctx, "环境变量名已存在: "+req.Name, constants.ED00006)
			return
		}
	}

	storedValue, err := resolveEnvVarStoredValue(existing, req)
	if err != nil {
		response.ErrorJSON(ctx, err.Error(), constants.ED00006)
		return
	}

	item := syssetting.EnvVar{
		Name:   req.Name,
		Value:  storedValue,
		Secret: req.Secret,
		Note:   req.Note,
	}
	if existIdx >= 0 {
		settings.Items[existIdx] = item
	} else {
		settings.Items = append(settings.Items, item)
	}

	if err := syssetting.ValidateEnvVars(settings); err != nil {
		response.ErrorJSON(ctx, err.Error(), constants.ED00006)
		return
	}

	newVersion, err := c.persistEnvVars(ctx, tenantId, operatorId, settings, version)
	if err != nil {
		if err == dao.ErrVersionConflict {
			response.ErrorJSON(ctx, err.Error(), constants.ED00006)
			return
		}
		logger.ErrorWithTrace(ctx, "保存环境变量失败", err)
		response.ErrorJSON(ctx, "保存环境变量失败: "+err.Error(), constants.ED00009)
		return
	}

	action := audit.AuditActionUpdate
	if existing == nil {
		action = audit.AuditActionCreate
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       action,
		ModuleCode:   "hub0009",
		TargetType:   "ENV_VAR",
		TargetId:     req.Name,
		TargetName:   req.Name,
		ResourceCode: "hub0009:edit",
		Detail: audit.SanitizeAuditDetail(map[string]interface{}{
			"name":      req.Name,
			"encrypted": req.Secret,
			"version":   newVersion,
		}),
	})

	response.SuccessJSON(ctx, gin.H{
		"groupCode":      syssetting.GroupEnvVars,
		"currentVersion": newVersion,
		"items":          toEnvVarViews(settings),
	}, constants.SD00004)
}

// DeleteEnvVar 删除一条全局环境变量。
func (c *SettingController) DeleteEnvVar(ctx *gin.Context) {
	var req models.DeleteEnvVarRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "租户不能为空", constants.ED00006)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if err := syssetting.ValidateEnvVarName(req.Name); err != nil {
		response.ErrorJSON(ctx, err.Error(), constants.ED00006)
		return
	}

	settings, version, err := c.loadEnvVars(ctx, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询环境变量失败", err)
		response.ErrorJSON(ctx, "查询环境变量失败: "+err.Error(), constants.ED00009)
		return
	}
	if req.CurrentVersion != version {
		response.ErrorJSON(ctx, dao.ErrVersionConflict.Error(), constants.ED00006)
		return
	}

	_, idx := findEnvVar(settings.Items, req.Name)
	if idx < 0 {
		response.ErrorJSON(ctx, "环境变量不存在: "+req.Name, constants.ED00006)
		return
	}
	settings.Items = append(settings.Items[:idx], settings.Items[idx+1:]...)

	newVersion, err := c.persistEnvVars(ctx, tenantId, operatorId, settings, version)
	if err != nil {
		if err == dao.ErrVersionConflict {
			response.ErrorJSON(ctx, err.Error(), constants.ED00006)
			return
		}
		logger.ErrorWithTrace(ctx, "删除环境变量失败", err)
		response.ErrorJSON(ctx, "删除环境变量失败: "+err.Error(), constants.ED00009)
		return
	}

	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionDelete,
		ModuleCode:   "hub0009",
		TargetType:   "ENV_VAR",
		TargetId:     req.Name,
		TargetName:   req.Name,
		ResourceCode: "hub0009:edit",
		Detail:       audit.SanitizeAuditDetail(map[string]interface{}{"name": req.Name, "version": newVersion}),
	})

	response.SuccessJSON(ctx, gin.H{
		"groupCode":      syssetting.GroupEnvVars,
		"currentVersion": newVersion,
		"items":          toEnvVarViews(settings),
	}, constants.SD00004)
}

func (c *SettingController) loadEnvVars(ctx *gin.Context, tenantId string) (syssetting.EnvVarsSettings, int, error) {
	row, err := c.dao.Get(ctx, tenantId, syssetting.GroupEnvVars)
	if err != nil {
		return syssetting.DefaultEnvVars(), 0, err
	}
	if row == nil {
		return syssetting.DefaultEnvVars(), 0, nil
	}
	return syssetting.ParseEnvVars(row.Content), row.CurrentVersion, nil
}

func (c *SettingController) persistEnvVars(ctx *gin.Context, tenantId, operatorId string, settings syssetting.EnvVarsSettings, expectVersion int) (int, error) {
	body, err := json.Marshal(settings)
	if err != nil {
		return 0, err
	}
	content := string(body)
	version, err := c.dao.Upsert(ctx, tenantId, syssetting.GroupEnvVars, content, operatorId, expectVersion)
	if err != nil {
		return 0, err
	}
	applyCache(tenantId, syssetting.GroupEnvVars, content)
	if pubErr := c.pub.PublishReload(ctx.Request.Context(), tenantId, syssetting.GroupEnvVars, operatorId); pubErr != nil {
		logger.WarnWithTrace(ctx, "发布环境设置集群事件失败", "error", pubErr.Error())
	}
	return version, nil
}

func envVarsViewFromRow(row *models.SysSetting) models.EnvVarsView {
	if row == nil {
		return models.EnvVarsView{Items: []models.EnvVarItemView{}, CurrentVersion: 0}
	}
	settings := syssetting.ParseEnvVars(row.Content)
	return models.EnvVarsView{
		Items:          toEnvVarViews(settings),
		CurrentVersion: row.CurrentVersion,
	}
}

func toEnvVarViews(settings syssetting.EnvVarsSettings) []models.EnvVarItemView {
	masked := syssetting.MaskEnvVars(settings)
	views := make([]models.EnvVarItemView, 0, len(masked))
	for _, item := range masked {
		views = append(views, models.EnvVarItemView{
			Name:     item.Name,
			Value:    item.Value,
			Secret:   item.Secret,
			HasValue: item.HasValue,
			Note:     item.Note,
		})
	}
	return views
}

func findEnvVar(items []syssetting.EnvVar, name string) (*syssetting.EnvVar, int) {
	for i := range items {
		if items[i].Name == name {
			return &items[i], i
		}
	}
	return nil, -1
}

func isMaskedOrEmptyValue(value string) bool {
	v := strings.TrimSpace(value)
	return v == "" || v == envVarSecretMask
}

func resolveEnvVarStoredValue(existing *syssetting.EnvVar, req models.SaveEnvVarRequest) (string, error) {
	keepOld := existing != nil && isMaskedOrEmptyValue(req.Value)
	if keepOld {
		if req.Secret {
			if existing.Value == "" {
				return "", errEnvVarSecretRequired()
			}
			if existing.Secret {
				return existing.Value, nil
			}
			return syssetting.EncodeEnvVarValue(syssetting.DecodeEnvVarValue(existing.Value), true)
		}
		if existing.Secret {
			return "", errEnvVarPlainRequired()
		}
		return existing.Value, nil
	}

	if req.Secret && isMaskedOrEmptyValue(req.Value) {
		return "", errEnvVarSecretRequired()
	}
	return syssetting.EncodeEnvVarValue(req.Value, req.Secret)
}

type envVarValueError string

func (e envVarValueError) Error() string { return string(e) }

func errEnvVarSecretRequired() error {
	return envVarValueError("密文变量值不能为空")
}

func errEnvVarPlainRequired() error {
	return envVarValueError("取消密文存储时请重新填写变量值")
}
