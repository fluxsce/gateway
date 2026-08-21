package controllers

import (
	"encoding/json"

	clusterPublish "gateway/internal/cluster/publish"
	"gateway/pkg/database"
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

// SettingController 环境设置控制器，按分组读写租户级策略。
type SettingController struct {
	db  database.Database
	dao *dao.SettingDAO
	pub *clusterPublish.EnvSettingEventPublisher
}

// NewSettingController 创建环境设置控制器。
func NewSettingController(db database.Database) *SettingController {
	return &SettingController{
		db:  db,
		dao: dao.NewSettingDAO(db),
		pub: clusterPublish.NewEnvSettingEventPublisher(),
	}
}

// GetEnvSettings 返回当前租户的归档策略、归档任务与 Web 超时，未落库时用默认值、版本为 0。
func (c *SettingController) GetEnvSettings(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "租户不能为空", constants.ED00006)
		return
	}

	rows, err := c.dao.ListByTenant(ctx, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询环境设置失败", err)
		response.ErrorJSON(ctx, "查询环境设置失败: "+err.Error(), constants.ED00009)
		return
	}

	resp := models.EnvSettingsResponse{
		Retention:    retentionView(syssetting.DefaultRetention(), 0),
		RetentionJob: retentionJobView(syssetting.DefaultRetentionJob(), 0),
		WebTimeout:   webTimeoutView(syssetting.DefaultWebTimeout(), 0),
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		switch row.GroupCode {
		case syssetting.GroupRetention:
			resp.Retention = retentionView(syssetting.ParseRetention(row.Content), row.CurrentVersion)
		case syssetting.GroupRetentionJob:
			resp.RetentionJob = retentionJobView(syssetting.ParseRetentionJob(row.Content), row.CurrentVersion)
		case syssetting.GroupWebTimeout:
			resp.WebTimeout = webTimeoutView(syssetting.ParseWebTimeout(row.Content), row.CurrentVersion)
		}
	}
	response.SuccessJSON(ctx, resp, constants.SD00002)
}

// SaveEnvSetting 保存单个分组，带乐观锁。
func (c *SettingController) SaveEnvSetting(ctx *gin.Context) {
	var req models.SaveSettingRequest
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

	content, err := encodeGroup(req)
	if err != nil {
		response.ErrorJSON(ctx, err.Error(), constants.ED00006)
		return
	}

	version, err := c.dao.Upsert(ctx, tenantId, req.GroupCode, content, operatorId, req.CurrentVersion)
	if err != nil {
		if err == dao.ErrVersionConflict {
			response.ErrorJSON(ctx, err.Error(), constants.ED00006)
			return
		}
		logger.ErrorWithTrace(ctx, "保存环境设置失败", err)
		response.ErrorJSON(ctx, "保存环境设置失败: "+err.Error(), constants.ED00009)
		return
	}

	applyCache(tenantId, req.GroupCode, content)
	if err := c.pub.PublishReload(ctx.Request.Context(), tenantId, req.GroupCode, operatorId); err != nil {
		logger.WarnWithTrace(ctx, "发布环境设置集群事件失败", "error", err.Error())
	}

	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionUpdate,
		ModuleCode:   "hub0009",
		TargetType:   "ENV_SETTING",
		TargetId:     req.GroupCode,
		TargetName:   groupDisplayName(req.GroupCode),
		ResourceCode: "hub0009:edit",
		Detail:       audit.SanitizeAuditDetail(map[string]interface{}{"groupCode": req.GroupCode, "version": version}),
	})

	response.SuccessJSON(ctx, gin.H{
		"groupCode":      req.GroupCode,
		"currentVersion": version,
	}, constants.SD00004)
}

func encodeGroup(req models.SaveSettingRequest) (string, error) {
	switch req.GroupCode {
	case syssetting.GroupRetention:
		v := syssetting.RetentionSettings{
			AuditLogDays:          req.AuditLogDays,
			TaskLogDays:           req.TaskLogDays,
			AlertLogDays:          req.AlertLogDays,
			ClusterEventDays:      req.ClusterEventDays,
			MetricsDays:           req.MetricsDays,
			GatewayLogDefaultDays: req.GatewayLogDefaultDays,
		}
		if err := syssetting.ValidateRetention(v); err != nil {
			return "", err
		}
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case syssetting.GroupRetentionJob:
		v := syssetting.RetentionJobSettings{
			Enabled:         req.Enabled,
			IntervalMinutes: req.IntervalMinutes,
			StartTime:       req.StartTime,
		}
		if err := syssetting.ValidateRetentionJob(v); err != nil {
			return "", err
		}
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case syssetting.GroupWebTimeout:
		v := syssetting.WebTimeoutSettings{
			RequestTimeoutSeconds: req.RequestTimeoutSeconds,
			SessionExpireHours:    req.SessionExpireHours,
		}
		if err := syssetting.ValidateWebTimeout(v); err != nil {
			return "", err
		}
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	default:
		return "", errUnknownGroup(req.GroupCode)
	}
}

func applyCache(tenantId, groupCode, content string) {
	syssetting.ApplyGroup(tenantId, groupCode, content)
}

func retentionView(v syssetting.RetentionSettings, version int) models.RetentionView {
	return models.RetentionView{
		AuditLogDays:          v.AuditLogDays,
		TaskLogDays:           v.TaskLogDays,
		AlertLogDays:          v.AlertLogDays,
		ClusterEventDays:      v.ClusterEventDays,
		MetricsDays:           v.MetricsDays,
		GatewayLogDefaultDays: v.GatewayLogDefaultDays,
		CurrentVersion:        version,
	}
}

func retentionJobView(v syssetting.RetentionJobSettings, version int) models.RetentionJobView {
	return models.RetentionJobView{
		Enabled:         v.Enabled,
		IntervalMinutes: v.IntervalMinutes,
		StartTime:       v.StartTime,
		CurrentVersion:  version,
	}
}

func webTimeoutView(v syssetting.WebTimeoutSettings, version int) models.WebTimeoutView {
	return models.WebTimeoutView{
		RequestTimeoutSeconds: v.RequestTimeoutSeconds,
		SessionExpireHours:    v.SessionExpireHours,
		CurrentVersion:        version,
	}
}

func groupDisplayName(groupCode string) string {
	switch groupCode {
	case syssetting.GroupRetention:
		return "归档策略"
	case syssetting.GroupRetentionJob:
		return "归档任务"
	case syssetting.GroupWebTimeout:
		return "Web访问超时"
	default:
		return groupCode
	}
}

type unknownGroupError string

func (e unknownGroupError) Error() string {
	return "不支持的设置分组: " + string(e)
}

func errUnknownGroup(code string) error {
	return unknownGroupError(code)
}
