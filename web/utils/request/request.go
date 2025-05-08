package request

import (
	"gohub/pkg/logger"
	"gohub/web/utils/auth"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// GetUserID 获取当前登录用户ID
func GetUserID(c *gin.Context) string {
	uc := auth.GetUserContext(c)
	if uc != nil {
		return uc.UserId
	}
	return ""
}

// GetTenantID 获取租户ID
func GetTenantID(c *gin.Context) string {
	uc := auth.GetUserContext(c)
	if uc != nil {
		return uc.TenantId
	}
	return "default"
}

// GetOperatorID 获取操作人ID
func GetOperatorID(c *gin.Context) string {
	uc := auth.GetUserContext(c)
	if uc != nil {
		return uc.UserId
	}
	return "system"
}

// GetUserName 获取当前登录用户名
func GetUserName(c *gin.Context) string {
	uc := auth.GetUserContext(c)
	if uc != nil {
		return uc.UserName
	}
	return ""
}

// GetUserContext 获取完整的用户上下文
func GetUserContext(c *gin.Context) *auth.UserContext {
	return auth.GetUserContext(c)
}

// GetPaginationParams 获取分页参数
func GetPaginationParams(c *gin.Context) (page, pageSize int) {
	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	// 转换参数
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err = strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	return page, pageSize
}

// GetQueryInt 获取请求中的整数参数
func GetQueryInt(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetQueryBool 获取请求中的布尔参数
func GetQueryBool(c *gin.Context, key string, defaultValue bool) bool {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	// 处理常见的布尔值表示
	switch valueStr {
	case "1", "true", "True", "TRUE", "yes", "Yes", "YES", "y", "Y":
		return true
	case "0", "false", "False", "FALSE", "no", "No", "NO", "n", "N":
		return false
	default:
		return defaultValue
	}
}

// GetParamID 获取路径参数中的ID
func GetParamID(c *gin.Context, key string) string {
	return c.Param(key)
}

// BindJSON 绑定JSON请求体并处理错误
func BindJSON(c *gin.Context, obj interface{}) error {
	return c.ShouldBindJSON(obj)
}

// BindForm 绑定表单数据(multipart/form-data)到结构体
func BindForm(c *gin.Context, obj interface{}) error {
	return c.ShouldBindWith(obj, binding.Form)
}

// BindQuery 绑定Query参数到结构体
func BindQuery(c *gin.Context, obj interface{}) error {
	return c.ShouldBindQuery(obj)
}

// BindFormPost 绑定表单提交数据(application/x-www-form-urlencoded)到结构体
func BindFormPost(c *gin.Context, obj interface{}) error {
	return c.ShouldBindWith(obj, binding.FormPost)
}

// BindUri 绑定URI参数到结构体
func BindUri(c *gin.Context, obj interface{}) error {
	return c.ShouldBindUri(obj)
}

// Bind 根据Content-Type自动选择绑定方法
// 支持:
// - application/json
// - application/x-www-form-urlencoded
// - multipart/form-data
func Bind(c *gin.Context, obj interface{}) error {
	contentType := c.GetHeader("Content-Type")

	// 记录请求内容类型，便于调试
	logger.Debug("请求绑定", "content-type", contentType, "method", c.Request.Method)

	// 根据Content-Type选择不同的绑定方法
	if strings.Contains(contentType, "application/json") {
		return c.ShouldBindJSON(obj)
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return c.ShouldBindWith(obj, binding.FormPost)
	} else if strings.Contains(contentType, "multipart/form-data") {
		return c.ShouldBindWith(obj, binding.Form)
	} else {
		// 默认尝试所有绑定方法
		return c.ShouldBind(obj)
	}
}

// GetClientInfo 获取客户端信息
func GetClientInfo(c *gin.Context) map[string]string {
	return map[string]string{
		"ip":         c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
		"referer":    c.GetHeader("Referer"),
	}
}

// GetFormFile 获取上传的文件
func GetFormFile(c *gin.Context, name string) (*multipart.FileHeader, error) {
	return c.FormFile(name)
}

// SaveUploadedFile 保存上传的文件
func SaveUploadedFile(c *gin.Context, fileHeader *multipart.FileHeader, dst string) error {
	return c.SaveUploadedFile(fileHeader, dst)
}
