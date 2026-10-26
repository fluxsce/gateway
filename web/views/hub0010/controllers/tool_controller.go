package controllers

import (
	"encoding/json"
	"errors"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/middleware"
	"gateway/web/middleware/audit"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0010/catalog"
	"gateway/web/views/hub0010/dao"
	"gateway/web/views/hub0010/models"

	"github.com/gin-gonic/gin"
)

// ToolController 工具市场：目录按权限过滤，安装状态按用户保存。
type ToolController struct {
	dao *dao.InstallDAO
}

// NewToolController 创建工具市场控制器。
func NewToolController(db database.Database) *ToolController {
	return &ToolController{dao: dao.NewInstallDAO(db)}
}

// toolView 返回给前端的一条工具。文案由前端按 id 翻译。
type toolView struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	Category      string                `json:"category"`
	Version       string                `json:"version"`
	Author        string                `json:"author"`
	Icon          string                `json:"icon"`
	Tags          []string              `json:"tags"`
	Status        string                `json:"status"`
	ResourceCode  string                `json:"resourceCode"`
	ConfigSchema  []catalog.ConfigField `json:"configSchema"`
	DefaultConfig map[string]any        `json:"defaultConfig"`
	UserConfig    map[string]any        `json:"userConfig,omitempty"`
	InstallTime   string                `json:"installTime,omitempty"`
}

// QueryTools 返回当前用户有权使用的工具，并带上本人的安装状态。
func (c *ToolController) QueryTools(ctx *gin.Context) {
	user := middleware.GetUserContext(ctx)
	if user == nil {
		response.ErrorJSON(ctx, "请先登录", constants.ED00011)
		return
	}
	installs, err := c.dao.ListActive(ctx, user.TenantId, user.UserId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询工具安装记录失败", err)
		response.ErrorJSON(ctx, "查询工具失败", constants.ED00009)
		return
	}
	byID := map[string]*models.ToolInstall{}
	for _, row := range installs {
		if row != nil {
			byID[row.ToolId] = row
		}
	}
	list := make([]toolView, 0)
	for _, tool := range catalog.All() {
		allowed, err := c.allowTool(ctx, tool.ID)
		if err != nil {
			logger.ErrorWithTrace(ctx, "检查工具权限失败", err)
			response.ErrorJSON(ctx, "权限检查失败", constants.ED00001)
			return
		}
		if !allowed {
			continue
		}
		list = append(list, toView(tool, byID[tool.ID]))
	}
	response.SuccessJSON(ctx, list, constants.SD00002)
}

// InstallTool 把工具装到当前用户。需要安装按钮，以及该工具的资源码。
func (c *ToolController) InstallTool(ctx *gin.Context) {
	c.writeInstall(ctx, true, audit.AuditActionCreate, "hub0010:install", constants.SD00003)
}

// UninstallTool 卸载当前用户的工具，保留行并把活动标记置为 N。
func (c *ToolController) UninstallTool(ctx *gin.Context) {
	c.writeInstall(ctx, false, audit.AuditActionDelete, "hub0010:uninstall", constants.SD00001)
}

// SaveToolConfig 保存已安装工具的个人配置。
func (c *ToolController) SaveToolConfig(ctx *gin.Context) {
	user, tool, cfg, ok := c.prepareWrite(ctx)
	if !ok {
		return
	}
	row, err := c.dao.Get(ctx, user.TenantId, user.UserId, tool.ID)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询工具安装记录失败", err)
		response.ErrorJSON(ctx, "保存配置失败", constants.ED00009)
		return
	}
	if row == nil || row.ActiveFlag != "Y" {
		response.ErrorJSON(ctx, "请先安装该工具", constants.ED00006)
		return
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		response.ErrorJSON(ctx, "配置无法保存", constants.ED00006)
		return
	}
	if err := c.dao.Save(ctx, user.TenantId, user.UserId, tool.ID, string(raw), request.GetOperatorID(ctx), "Y"); err != nil {
		logger.ErrorWithTrace(ctx, "保存工具配置失败", err)
		response.ErrorJSON(ctx, "保存配置失败", constants.ED00009)
		return
	}
	c.audit(ctx, audit.AuditActionUpdate, "hub0010:config", tool.ID)
	response.SuccessJSON(ctx, gin.H{"toolId": tool.ID, "config": cfg}, constants.SD00004)
}

func (c *ToolController) writeInstall(ctx *gin.Context, install bool, action, resourceCode, messageID string) {
	user, tool, cfg, ok := c.prepareWrite(ctx)
	if !ok {
		return
	}
	active := "N"
	stored := "{}"
	if install {
		active = "Y"
		raw, err := json.Marshal(cfg)
		if err != nil {
			response.ErrorJSON(ctx, "配置无法保存", constants.ED00006)
			return
		}
		stored = string(raw)
	} else {
		row, err := c.dao.Get(ctx, user.TenantId, user.UserId, tool.ID)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询工具安装记录失败", err)
			response.ErrorJSON(ctx, "卸载工具失败", constants.ED00009)
			return
		}
		if row != nil && row.ConfigJson != "" {
			stored = row.ConfigJson
		}
	}
	if err := c.dao.Save(ctx, user.TenantId, user.UserId, tool.ID, stored, request.GetOperatorID(ctx), active); err != nil {
		logger.ErrorWithTrace(ctx, "保存工具安装状态失败", err)
		response.ErrorJSON(ctx, "保存工具状态失败", constants.ED00009)
		return
	}
	c.audit(ctx, action, resourceCode, tool.ID)
	response.SuccessJSON(ctx, gin.H{"toolId": tool.ID, "installed": install}, messageID)
}

func (c *ToolController) prepareWrite(ctx *gin.Context) (*middlewareUser, catalog.Tool, map[string]any, bool) {
	user := middleware.GetUserContext(ctx)
	if user == nil || user.TenantId == "" || user.UserId == "" {
		response.ErrorJSON(ctx, "请先登录", constants.ED00011)
		return nil, catalog.Tool{}, nil, false
	}
	var req models.ToolActionRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return nil, catalog.Tool{}, nil, false
	}
	tool, found := catalog.Find(req.ToolId)
	if !found {
		response.ErrorJSON(ctx, "工具不存在", constants.ED00006)
		return nil, catalog.Tool{}, nil, false
	}
	allowed, err := c.allowTool(ctx, tool.ID)
	if err != nil {
		logger.ErrorWithTrace(ctx, "检查工具权限失败", err)
		response.ErrorJSON(ctx, "权限检查失败", constants.ED00001)
		return nil, catalog.Tool{}, nil, false
	}
	if !allowed {
		response.ErrorJSON(ctx, "没有使用该工具的权限", constants.ED00010)
		return nil, catalog.Tool{}, nil, false
	}
	cfg, err := catalog.NormalizeConfig(tool, req.Config)
	if err != nil {
		response.ErrorJSON(ctx, err.Error(), constants.ED00006)
		return nil, catalog.Tool{}, nil, false
	}
	return &middlewareUser{TenantId: user.TenantId, UserId: user.UserId}, tool, cfg, true
}

// middlewareUser 只保留写库需要的身份，避免控制器依赖会话结构的其它字段。
type middlewareUser struct {
	TenantId string
	UserId   string
}

func (c *ToolController) allowTool(ctx *gin.Context, toolID string) (bool, error) {
	user := middleware.GetUserContext(ctx)
	if user == nil {
		return false, errors.New("未找到用户")
	}
	if user.TenantAdminFlag == "Y" {
		return true, nil
	}
	svc := middleware.GetPermissionService()
	if svc == nil {
		return false, errors.New("权限服务未初始化")
	}
	return svc.HasAnyButtonAccess(ctx.Request.Context(), user.UserId, user.TenantId, []string{catalog.ResourceCode(toolID)})
}

func (c *ToolController) audit(ctx *gin.Context, action, resourceCode, toolID string) {
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       action,
		ModuleCode:   catalog.ModuleCode,
		TargetType:   "TOOL",
		TargetId:     toolID,
		TargetName:   toolID,
		ResourceCode: resourceCode,
		Detail:       audit.SanitizeAuditDetail(map[string]interface{}{"toolId": toolID}),
	})
}

func toView(tool catalog.Tool, row *models.ToolInstall) toolView {
	view := toolView{
		ID:            tool.ID,
		Name:          tool.ID,
		Category:      tool.Category,
		Version:       tool.Version,
		Author:        tool.Author,
		Icon:          tool.Icon,
		Tags:          tool.Tags,
		Status:        "available",
		ResourceCode:  catalog.ResourceCode(tool.ID),
		ConfigSchema:  tool.ConfigSchema,
		DefaultConfig: tool.DefaultConfig,
	}
	if view.ConfigSchema == nil {
		view.ConfigSchema = []catalog.ConfigField{}
	}
	if view.Tags == nil {
		view.Tags = []string{}
	}
	if row == nil || row.ActiveFlag != "Y" {
		return view
	}
	view.Status = "installed"
	view.InstallTime = row.InstallTime.Format("2006-01-02T15:04:05Z07:00")
	if row.ConfigJson != "" {
		var cfg map[string]any
		if err := json.Unmarshal([]byte(row.ConfigJson), &cfg); err == nil {
			view.UserConfig = cfg
		}
	}
	return view
}
