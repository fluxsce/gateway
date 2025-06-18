package request

import (
	"fmt"
	"gohub/pkg/logger"
	"gohub/web/utils/auth"
	"gohub/web/utils/constants"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
	"time"

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
	pageStr := c.DefaultQuery("page", constants.DefaultPage)
	pageSizeStr := c.DefaultQuery("pageSize", constants.DefaultPageSize)

	// 转换参数
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err = strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < constants.MinPageSize {
		pageSize = constants.MinPageSize
	}

	// 限制最大分页大小
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
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

// BindSafely 安全绑定函数，处理常见的类型转换错误
func BindSafely(c *gin.Context, obj interface{}) error {
	contentType := c.GetHeader("Content-Type")

	// 记录请求内容类型，便于调试
	logger.Debug("请求安全绑定", "content-type", contentType, "method", c.Request.Method)

	// 先尝试标准绑定
	err := Bind(c, obj)
	if err != nil {
		logger.Warn("绑定出错，尝试自定义处理", "error", err.Error())

		// 不管错误类型，尝试使用Map先接收数据
		var rawData map[string]interface{} = make(map[string]interface{})
		if contentType == "" || strings.Contains(contentType, "application/json") {
			bindErr := c.ShouldBindJSON(&rawData)
			if bindErr == nil {
				// 使用反射处理转换
				cleanAndSetValue(obj, rawData)
				return nil
			} else {
				logger.Warn("JSON数据绑定到map失败", "error", bindErr.Error())
			}
		} else if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
			strings.Contains(contentType, "multipart/form-data") {
			formData := make(map[string][]string)
			if c.Request.Form != nil {
				formData = c.Request.Form
			} else if c.Request.PostForm != nil {
				formData = c.Request.PostForm
			} else {
				c.Request.ParseMultipartForm(32 << 20) // 32MB
				if c.Request.Form != nil {
					formData = c.Request.Form
				}
			}

			// 将表单数据转换为map
			for k, v := range formData {
				if len(v) > 0 {
					rawData[k] = v[0]
				}
			}

			// 处理转换
			cleanAndSetValue(obj, rawData)
			return nil
		}

		// 如果自定义处理失败，返回原始错误
		return err
	}

	return nil
}

// cleanAndSetValue 清理并设置值，处理整数类型转换问题
func cleanAndSetValue(obj interface{}, data map[string]interface{}) {
	val := reflect.ValueOf(obj)

	// 确保是指针类型
	if val.Kind() != reflect.Ptr {
		return
	}

	// 获取指针指向的值
	val = val.Elem()

	// 确保是结构体
	if val.Kind() != reflect.Struct {
		return
	}

	// 遍历结构体字段
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 如果字段不可设置，跳过
		if !field.CanSet() {
			continue
		}

		// 获取json标签名
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			continue
		}

		// 分割标签，获取字段名
		parts := strings.Split(jsonTag, ",")
		jsonName := parts[0]

		// 查找map中是否有对应的值
		if value, ok := data[jsonName]; ok {
			processField(field, value)
		} else {
			// 检查form或query标签
			formTag := fieldType.Tag.Get("form")
			if formTag != "" {
				parts = strings.Split(formTag, ",")
				formName := parts[0]

				if value, ok := data[formName]; ok {
					processField(field, value)
				}
			}
		}
	}
}

// processField 处理各种类型的字段
func processField(field reflect.Value, value interface{}) {
	switch field.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		setIntValue(field, value)
	case reflect.Float64, reflect.Float32:
		setFloatValue(field, value)
	case reflect.Bool:
		setBoolValue(field, value)
	case reflect.String:
		setStringValue(field, value)
	case reflect.Struct:
		// 处理特殊结构体类型，如time.Time
		if field.Type().String() == "time.Time" {
			setTimeValue(field, value)
		}
	case reflect.Ptr:
		// 处理指针类型
		if field.Type().Elem().String() == "time.Time" {
			setTimePtrValue(field, value)
		}
	}
}

// setIntValue 设置整数值
func setIntValue(field reflect.Value, value interface{}) {
	// 如果是字符串，尝试转换
	if strValue, ok := value.(string); ok {
		// 去除空格
		strValue = strings.TrimSpace(strValue)
		if strValue == "" {
			// 如果是空字符串，不修改值
			return
		}

		// 尝试转换为整数
		if intValue, err := strconv.ParseInt(strValue, 10, 64); err == nil {
			field.SetInt(intValue)
		}
	} else if floatValue, ok := value.(float64); ok {
		// JSON解析数字默认为float64
		field.SetInt(int64(floatValue))
	} else if intValue, ok := value.(int); ok {
		field.SetInt(int64(intValue))
	} else if intValue, ok := value.(int64); ok {
		field.SetInt(intValue)
	}
}

// setFloatValue 设置浮点值
func setFloatValue(field reflect.Value, value interface{}) {
	// 如果是字符串，尝试转换
	if strValue, ok := value.(string); ok {
		// 去除空格
		strValue = strings.TrimSpace(strValue)
		if strValue == "" {
			// 如果是空字符串，不修改值
			return
		}

		// 尝试转换为浮点数
		if floatValue, err := strconv.ParseFloat(strValue, 64); err == nil {
			field.SetFloat(floatValue)
		}
	} else if floatValue, ok := value.(float64); ok {
		field.SetFloat(floatValue)
	} else if intValue, ok := value.(int); ok {
		field.SetFloat(float64(intValue))
	}
}

// setBoolValue 设置布尔值
func setBoolValue(field reflect.Value, value interface{}) {
	// 如果是字符串，尝试转换
	if strValue, ok := value.(string); ok {
		// 去除空格
		strValue = strings.TrimSpace(strValue)
		if strValue == "" {
			// 如果是空字符串，不修改值
			return
		}

		// 处理常见的布尔值表示
		switch strings.ToLower(strValue) {
		case "1", "true", "t", "yes", "y", "on":
			field.SetBool(true)
		case "0", "false", "f", "no", "n", "off":
			field.SetBool(false)
		}
	} else if boolValue, ok := value.(bool); ok {
		field.SetBool(boolValue)
	} else if intValue, ok := value.(int); ok {
		field.SetBool(intValue != 0)
	} else if floatValue, ok := value.(float64); ok {
		field.SetBool(floatValue != 0)
	}
}

// setStringValue 设置字符串值
func setStringValue(field reflect.Value, value interface{}) {
	// 转换为字符串
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case int:
		strValue = strconv.Itoa(v)
	case float64:
		strValue = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		strValue = strconv.FormatBool(v)
	default:
		// 尝试使用fmt.Sprint
		strValue = fmt.Sprint(value)
	}

	field.SetString(strValue)
}

// setTimeValue 设置时间值
func setTimeValue(field reflect.Value, value interface{}) {
	// 支持的时间格式
	timeFormats := []string{
		"2006-01-02 15:04:05",  // 标准日期时间格式
		"2006-01-02T15:04:05Z", // ISO时间格式
		"2006-01-02T15:04:05",  // ISO时间格式（无时区）
		"2006/01/02 15:04:05",  // 斜杠分隔日期时间
		"2006-01-02",           // 仅日期
		"15:04:05",             // 仅时间
		time.RFC3339,           // RFC3339格式
		time.RFC3339Nano,       // RFC3339纳秒格式
		time.RFC1123,           // RFC1123格式
		time.RFC1123Z,          // RFC1123带时区格式
	}

	// 如果已经是time.Time类型，则直接使用
	if timeValue, ok := value.(time.Time); ok {
		field.Set(reflect.ValueOf(timeValue))
		return
	}

	// 如果是数字（Unix时间戳），转换为time.Time
	if floatValue, ok := value.(float64); ok {
		sec := int64(floatValue)
		nsec := int64((floatValue - float64(sec)) * 1e9)
		t := time.Unix(sec, nsec)
		field.Set(reflect.ValueOf(t))
		return
	}

	// 如果是字符串，尝试使用不同格式解析
	if strValue, ok := value.(string); ok {
		// 去除空格
		strValue = strings.TrimSpace(strValue)
		if strValue == "" {
			// 如果是空字符串，不修改值
			return
		}

		// 尝试各种时间格式
		for _, format := range timeFormats {
			if t, err := time.Parse(format, strValue); err == nil {
				// 如果解析成功，设置字段值
				field.Set(reflect.ValueOf(t))
				return
			}
		}

		// 尝试解析Unix时间戳（以秒为单位）
		if timestamp, err := strconv.ParseInt(strValue, 10, 64); err == nil {
			t := time.Unix(timestamp, 0)
			field.Set(reflect.ValueOf(t))
			return
		}
	}
}

// setTimePtrValue 设置时间指针值
func setTimePtrValue(field reflect.Value, value interface{}) {
	// 如果是空值，则设置为nil
	if value == nil {
		field.Set(reflect.Zero(field.Type()))
		return
	}

	// 字符串类型处理
	if strValue, ok := value.(string); ok {
		strValue = strings.TrimSpace(strValue)
		if strValue == "" || strValue == "null" || strValue == "NULL" {
			field.Set(reflect.Zero(field.Type()))
			return
		}
	}

	// 创建新的时间指针
	if field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}

	// 获取指针内部的时间值
	timeValue := field.Elem()

	// 使用和setTimeValue相同的时间格式列表
	timeFormats := []string{
		"2006-01-02 15:04:05",  // 标准日期时间格式
		"2006-01-02T15:04:05Z", // ISO时间格式
		"2006-01-02T15:04:05",  // ISO时间格式（无时区）
		"2006/01/02 15:04:05",  // 斜杠分隔日期时间
		"2006-01-02",           // 仅日期
		"15:04:05",             // 仅时间
		time.RFC3339,           // RFC3339格式
		time.RFC3339Nano,       // RFC3339纳秒格式
		time.RFC1123,           // RFC1123格式
		time.RFC1123Z,          // RFC1123带时区格式
	}

	// 如果已经是time.Time类型，则直接使用
	if t, ok := value.(time.Time); ok {
		timeValue.Set(reflect.ValueOf(t))
		return
	}

	// 如果是数字（Unix时间戳），转换为time.Time
	if floatValue, ok := value.(float64); ok {
		sec := int64(floatValue)
		nsec := int64((floatValue - float64(sec)) * 1e9)
		t := time.Unix(sec, nsec)
		timeValue.Set(reflect.ValueOf(t))
		return
	}

	// 如果是字符串，尝试使用不同格式解析
	if strValue, ok := value.(string); ok {
		// 已经处理过空字符串情况

		// 尝试各种时间格式
		for _, format := range timeFormats {
			if t, err := time.Parse(format, strValue); err == nil {
				// 如果解析成功，设置字段值
				timeValue.Set(reflect.ValueOf(t))
				return
			}
		}

		// 尝试解析Unix时间戳（以秒为单位）
		if timestamp, err := strconv.ParseInt(strValue, 10, 64); err == nil {
			t := time.Unix(timestamp, 0)
			timeValue.Set(reflect.ValueOf(t))
			return
		}
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
