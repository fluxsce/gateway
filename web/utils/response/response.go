package response

import (
	"encoding/json"
	"gohub/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PageInfo 后端返回的分页信息对象结构
type PageInfo struct {
	// 基础数据，通常用于存储额外信息
	BaseData string `json:"baseData"`
	// 当前页面记录数量
	CurPageCount int `json:"curPageCount"`
	// 数据库服务ID
	DbsId string `json:"dbsId"`
	// 主键字段名
	MainKey string `json:"mainKey"`
	// 排序规则列表，通常为字符串形式的SQL排序表达式
	OrderByList string `json:"orderByList"`
	// 额外数据，存储非标准信息
	OtherData string `json:"otherData"`
	// 当前页码，从1开始
	PageIndex int `json:"pageIndex"`
	// 每页记录数
	PageSize int `json:"pageSize"`
	// 查询参数对象JSON字符串
	ParamObjectsJson string `json:"paramObjectsJson"`
	// 时间类型字段名列表
	TimeTypeFieldNames string `json:"timeTypeFieldNames"`
	// 记录总数
	TotalCount int `json:"totalCount"`
	// 总页数
	TotalPageIndex int `json:"totalPageIndex"`
}

// JsonData 定义了与后端交互的标准响应格式
type JsonData struct {
	// 操作是否成功（注意首字母o为小写，K为大写，Java命名习惯）
	OK bool `json:"oK"`
	// 状态标识，表示操作状态
	State bool `json:"state"`
	// 业务数据，通常是序列化后的JSON字符串
	BizData string `json:"bizData"`
	// 扩展对象，可以是任意类型
	ExtObj interface{} `json:"extObj"`
	// 分页查询数据，通常包含分页相关信息
	PageQueryData string `json:"pageQueryData"`
	// 消息标识，用于追踪和日志
	MessageId string `json:"messageId"`
	// 错误消息，当操作失败时提供错误详情
	ErrMsg string `json:"errMsg"`
	// 弹出消息，可用于前端展示
	PopMsg string `json:"popMsg"`
	// 扩展消息，提供额外信息
	ExtMsg string `json:"extMsg"`
	// 预留主键字段1，用于特定业务场景
	Pkey1 string `json:"pkey1"`
	// 预留主键字段2，用于特定业务场景
	Pkey2 string `json:"pkey2"`
	// 预留主键字段3，用于特定业务场景
	Pkey3 string `json:"pkey3"`
	// 预留主键字段4，用于特定业务场景
	Pkey4 string `json:"pkey4"`
	// 预留主键字段5，用于特定业务场景
	Pkey5 string `json:"pkey5"`
	// 预留主键字段6，用于特定业务场景
	Pkey6 string `json:"pkey6"`
}

// 创建成功的响应
func Success(bizData interface{}, messageId string) JsonData {
	// 将业务数据转换为JSON字符串
	var bizDataStr string
	switch v := bizData.(type) {
	case string:
		bizDataStr = v
	case nil:
		bizDataStr = ""
	default:
		// 使用json.Marshal转换为字符串
		jsonBytes, err := json.Marshal(bizData)
		if err != nil {
			// 使用自定义logger记录错误
			logger.Error("业务数据序列化失败", "error", err)
			bizDataStr = ""
		} else {
			bizDataStr = string(jsonBytes)
		}
	}

	return JsonData{
		OK:        true,
		State:     true,
		BizData:   bizDataStr,
		MessageId: messageId,
	}
}

// 创建失败的响应
func Error(errMsg string, messageId string) JsonData {
	return JsonData{
		OK:        false,
		State:     false,
		ErrMsg:    errMsg,
		MessageId: messageId,
	}
}

// 创建带分页的成功响应
func Page(bizData interface{}, pageInfo PageInfo, messageId string) JsonData {
	// 将业务数据转换为JSON字符串
	var bizDataStr string
	switch v := bizData.(type) {
	case string:
		bizDataStr = v
	case nil:
		bizDataStr = ""
	default:
		// 使用json.Marshal转换为字符串
		jsonBytes, err := json.Marshal(bizData)
		if err != nil {
			// 使用自定义logger记录错误
			logger.Error("业务数据序列化失败", "error", err)
			bizDataStr = ""
		} else {
			bizDataStr = string(jsonBytes)
		}
	}

	// 将分页信息转换为JSON字符串
	pageQueryDataStr := ""
	jsonBytes, err := json.Marshal(pageInfo)
	if err != nil {
		// 使用自定义logger记录错误
		logger.Error("分页信息序列化失败", "error", err)
	} else {
		pageQueryDataStr = string(jsonBytes)
	}

	return JsonData{
		OK:            true,
		State:         true,
		BizData:       bizDataStr,
		PageQueryData: pageQueryDataStr,
		MessageId:     messageId,
	}
}

// 创建一般的响应（既不是成功也不是失败的特殊情况）
func Json(data interface{}, messageId string) JsonData {
	// 将业务数据转换为JSON字符串
	var bizDataStr string
	switch v := data.(type) {
	case string:
		bizDataStr = v
	case nil:
		bizDataStr = ""
	default:
		// 使用json.Marshal转换为字符串
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			// 使用自定义logger记录错误
			logger.Error("业务数据序列化失败", "error", err)
			bizDataStr = ""
		} else {
			bizDataStr = string(jsonBytes)
		}
	}

	return JsonData{
		OK:        true,
		State:     true,
		BizData:   bizDataStr,
		MessageId: messageId,
	}
}

// 返回成功响应
func SuccessJSON(c *gin.Context, data interface{}, messageId string) {
	c.JSON(http.StatusOK, Success(data, messageId))
}

// 返回错误响应
func ErrorJSON(c *gin.Context, errMsg string, messageId string, status ...int) {
	httpStatus := http.StatusOK
	if len(status) > 0 {
		httpStatus = status[0]
	}
	c.JSON(httpStatus, Error(errMsg, messageId))
}

// 返回带分页的成功响应
func PageJSON(c *gin.Context, data interface{}, pageInfo PageInfo, messageId string) {
	c.JSON(http.StatusOK, Page(data, pageInfo, messageId))
}

// 创建新的分页信息对象
func NewPageInfo(pageIndex, pageSize, total int) PageInfo {
	return PageInfo{
		PageIndex:      pageIndex,
		PageSize:       pageSize,
		TotalCount:     total,
		TotalPageIndex: (total + pageSize - 1) / pageSize, // 计算总页数
		CurPageCount:   min(pageSize, total-(pageIndex-1)*pageSize),
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
