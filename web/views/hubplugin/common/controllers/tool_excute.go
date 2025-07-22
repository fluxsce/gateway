package controllers

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/timerinit/sftp"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"

	"github.com/gin-gonic/gin"
)

// ToolExecuteController 工具执行测试控制器
type ToolExecuteController struct {
	db database.Database
}

// NewToolExecuteController 创建工具执行测试控制器
func NewToolExecuteController(db database.Database) *ToolExecuteController {
	return &ToolExecuteController{
		db: db,
	}
}

// TestToolExecution 测试工具执行
// @Summary 测试工具执行
// @Description 根据工具配置创建执行器并进行测试
// @Tags 工具执行测试
// @Accept json
// @Produce json
// @Param data body object true "测试参数"
// @Success 200 {object} response.Response
// @Router /api/tool/test-execute [post]
func (c *ToolExecuteController) TestToolExecution(ctx *gin.Context) {
	var params struct {
		ToolConfigId string `json:"toolConfigId" binding:"required"`
		ToolType     string `json:"toolType" binding:"required"`
	}

	if err := request.BindSafely(ctx, &params); err != nil {
		logger.Error("工具执行测试参数解析失败", "error", err, "params", params)
		response.ErrorJSON(ctx, "参数解析失败: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		logger.Error("工具执行测试无法获取租户信息", "context", ctx.Request.Header)
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 创建测试上下文
	testCtx, cancel := context.WithTimeout(ctx.Request.Context(), 30*time.Second)
	defer cancel()

	logger.Info("开始工具执行测试", "tenantId", tenantId, "toolConfigId", params.ToolConfigId, "toolType", params.ToolType)

	// 根据工具类型创建执行器并测试
	result := c.testByToolType(testCtx, tenantId, params.ToolConfigId, params.ToolType)

	// 检查测试结果
	if success, ok := result["success"].(bool); ok && success {
		logger.Info("工具执行测试成功", "tenantId", tenantId, "toolConfigId", params.ToolConfigId, "toolType", params.ToolType)
		response.SuccessJSON(ctx, result, constants.SD00001)
	} else {
		// 安全地获取错误消息
		var errorMessage string
		if msg, exists := result["message"]; exists {
			switch v := msg.(type) {
			case string:
				errorMessage = v
			case error:
				errorMessage = v.Error()
			default:
				errorMessage = fmt.Sprintf("%v", v)
			}
		} else {
			errorMessage = "测试失败，未知错误"
		}

		// 获取详细错误信息
		var detailError string
		if errDetail, exists := result["error"]; exists {
			switch v := errDetail.(type) {
			case string:
				detailError = v
			case error:
				detailError = v.Error()
			default:
				detailError = fmt.Sprintf("%v", v)
			}
		}

		// 记录详细的错误日志
		logger.Error("工具执行测试失败",
			"tenantId", tenantId,
			"toolConfigId", params.ToolConfigId,
			"toolType", params.ToolType,
			"message", errorMessage,
			"detail", detailError,
			"result", result)

		// 构建完整的错误消息
		fullErrorMessage := errorMessage
		if detailError != "" && detailError != errorMessage {
			fullErrorMessage = fmt.Sprintf("%s: %s", errorMessage, detailError)
		}

		response.ErrorJSON(ctx, fullErrorMessage, constants.ED00009)
	}
}

// testByToolType 根据工具类型进行测试
func (c *ToolExecuteController) testByToolType(ctx context.Context, tenantId, toolConfigId, toolType string) map[string]interface{} {
	switch toolType {
	case "SFTP_TRANSFER":
		return c.testSFTP(ctx, tenantId, toolConfigId)
	default:
		errMsg := fmt.Sprintf("不支持的工具类型: %s", toolType)
		logger.Error("不支持的工具类型", "toolType", toolType, "tenantId", tenantId, "toolConfigId", toolConfigId)
		return map[string]interface{}{
			"success": false,
			"message": errMsg,
		}
	}
}

// testSFTP 测试SFTP工具
func (c *ToolExecuteController) testSFTP(ctx context.Context, tenantId, toolConfigId string) map[string]interface{} {
	// 使用静态方法创建SFTP执行器
	executor, err := sftp.CreateSFTPExecutorStatic(ctx, c.db, tenantId, toolConfigId)
	if err != nil {
		logger.Error("创建SFTP执行器失败", "error", err, "tenantId", tenantId, "toolConfigId", toolConfigId)
		return map[string]interface{}{
			"success": false,
			"message": "创建SFTP执行器失败",
			"error":   err.Error(),
			"stage":   "executor_creation",
		}
	}

	// 确保执行器关闭
	defer func() {
		if err := executor.Close(); err != nil {
			logger.Error("关闭SFTP执行器失败", "error", err, "tenantId", tenantId, "toolConfigId", toolConfigId)
		}
	}()

	// 获取SFTP客户端
	sftpClient := executor.GetSFTPClient()
	if sftpClient == nil {
		logger.Error("获取SFTP客户端失败", "tenantId", tenantId, "toolConfigId", toolConfigId)
		return map[string]interface{}{
			"success": false,
			"message": "获取SFTP客户端失败",
			"error":   "SFTP客户端为空",
			"stage":   "client_retrieval",
		}
	}

	// 测试连接
	if !sftpClient.IsConnected() {
		if err := sftpClient.Connect(ctx); err != nil {
			logger.Error("SFTP连接失败", "error", err, "tenantId", tenantId, "toolConfigId", toolConfigId)
			return map[string]interface{}{
				"success": false,
				"message": "SFTP连接失败",
				"error":   err.Error(),
				"stage":   "connection",
			}
		}
	}

	logger.Info("SFTP连接成功", "tenantId", tenantId, "toolConfigId", toolConfigId)

	// 测试基本操作（列出根目录）
	files, err := sftpClient.ListDirectory(ctx, "/")
	if err != nil {
		logger.Error("SFTP操作测试失败", "error", err, "tenantId", tenantId, "toolConfigId", toolConfigId)
		return map[string]interface{}{
			"success": false,
			"message": "SFTP操作测试失败",
			"error":   err.Error(),
			"stage":   "operation",
		}
	}

	logger.Info("SFTP操作测试成功", "tenantId", tenantId, "toolConfigId", toolConfigId, "fileCount", len(files))

	return map[string]interface{}{
		"success":   true,
		"message":   "SFTP连接和操作测试成功",
		"data":      map[string]interface{}{"fileCount": len(files)},
		"stage":     "completed",
		"timestamp": time.Now().Unix(),
	}
}
