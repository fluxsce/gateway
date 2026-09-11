package controllers

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"gateway/pkg/logger"
	"gateway/web/middleware"
	"gateway/web/middleware/audit"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0001/models"
	hubdao "gateway/web/views/hub0002/dao"
	hubmodels "gateway/web/views/hub0002/models"

	"github.com/gin-gonic/gin"
)

var (
	errProfileRealNameRequired = errors.New("真实姓名不能为空")
	errProfileRealNameLength   = errors.New("真实姓名长度为2到50个字符")
	errProfileEmailInvalid     = errors.New("邮箱格式不正确")
	errProfileMobileInvalid    = errors.New("手机号格式不正确")
	errProfileGenderInvalid    = errors.New("性别取值无效")
)

// GetProfile 返回当前登录用户的资料。身份只取 session，忽略请求中的 userId。
func (c *AuthController) GetProfile(ctx *gin.Context) {
	userContext := middleware.GetUserContext(ctx)
	if userContext == nil {
		response.ErrorJSON(ctx, "未获取到用户信息，请重新登录", constants.ED00011, http.StatusUnauthorized)
		return
	}

	user, err := c.userDAO.GetUserById(ctx, userContext.UserId, userContext.TenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取个人资料失败", err)
		response.ErrorJSON(ctx, "获取个人资料失败", constants.ED00009)
		return
	}
	if user == nil {
		response.ErrorJSON(ctx, "用户不存在", constants.ED00102)
		return
	}

	response.SuccessJSON(ctx, ownProfileToMap(user), constants.SD00102)
}

// UpdateProfile 只改本人资料字段。请求里的 userId、租户、管理员标记一律忽略。
func (c *AuthController) UpdateProfile(ctx *gin.Context) {
	userContext := middleware.GetUserContext(ctx)
	if userContext == nil {
		response.ErrorJSON(ctx, "未获取到用户信息，请重新登录", constants.ED00011, http.StatusUnauthorized)
		return
	}

	var req models.ProfileUpdateRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}
	if err := validateOwnProfile(req); err != nil {
		response.ErrorJSON(ctx, err.Error(), constants.ED00007)
		return
	}

	realName := strings.TrimSpace(req.RealName)
	email := strings.TrimSpace(req.Email)
	mobile := strings.TrimSpace(req.Mobile)
	avatar := strings.TrimSpace(req.Avatar)

	err := c.userDAO.UpdateOwnProfile(ctx, userContext.UserId, userContext.TenantId, userContext.UserId, realName, email, mobile, avatar, req.Gender)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新个人资料失败", err, "userId", userContext.UserId)
		if errors.Is(err, hubdao.ErrUserNotFound) {
			response.ErrorJSON(ctx, err.Error(), constants.ED00102)
			return
		}
		response.ErrorJSON(ctx, "更新个人资料失败: "+err.Error(), constants.ED00009)
		return
	}

	if patchErr := c.sessionManager.UpdateSessionProfile(ctx, userContext.SessionId, realName, email, mobile, avatar); patchErr != nil {
		logger.WarnWithTrace(ctx, "个人资料已保存，刷新会话展示字段失败", "error", patchErr)
	}

	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionUpdate,
		ModuleCode:   "hub0001",
		TargetType:   "USER",
		TargetId:     userContext.UserId,
		TargetName:   realName,
		ResourceCode: "hub0001:profile",
		Detail:       "updateOwnProfile",
	})

	user, err := c.userDAO.GetUserById(ctx, userContext.UserId, userContext.TenantId)
	if err != nil || user == nil {
		response.SuccessJSON(ctx, gin.H{"message": "更新成功"}, constants.SD00004)
		return
	}
	response.SuccessJSON(ctx, ownProfileToMap(user), constants.SD00004)
}

func validateOwnProfile(req models.ProfileUpdateRequest) error {
	realName := strings.TrimSpace(req.RealName)
	if realName == "" {
		return errProfileRealNameRequired
	}
	nameLen := utf8.RuneCountInString(realName)
	if nameLen < 2 || nameLen > 50 {
		return errProfileRealNameLength
	}
	email := strings.TrimSpace(req.Email)
	if email != "" && !strings.Contains(email, "@") {
		return errProfileEmailInvalid
	}
	mobile := strings.TrimSpace(req.Mobile)
	if mobile != "" && !ownProfileMobileOK(mobile) {
		return errProfileMobileInvalid
	}
	if req.Gender < 0 || req.Gender > 2 {
		return errProfileGenderInvalid
	}
	return nil
}

func ownProfileMobileOK(mobile string) bool {
	if len(mobile) != 11 || mobile[0] != '1' {
		return false
	}
	if mobile[1] < '3' || mobile[1] > '9' {
		return false
	}
	for i := 2; i < 11; i++ {
		if mobile[i] < '0' || mobile[i] > '9' {
			return false
		}
	}
	return true
}

func ownProfileToMap(user *hubmodels.User) gin.H {
	return gin.H{
		"userId":          user.UserId,
		"userName":        user.UserName,
		"realName":        user.RealName,
		"tenantId":        user.TenantId,
		"deptId":          user.DeptId,
		"email":           user.Email,
		"mobile":          user.Mobile,
		"gender":          user.Gender,
		"avatar":          user.Avatar,
		"statusFlag":      user.StatusFlag,
		"deptAdminFlag":   user.DeptAdminFlag,
		"tenantAdminFlag": user.TenantAdminFlag,
		"userExpireDate":  user.UserExpireDate,
		"activeFlag":      user.ActiveFlag,
	}
}
