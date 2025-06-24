package executor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"
)

// LogExecutor 日志执行器
// 用于测试和调试，将消息输出到日志
type LogExecutor struct {
	*BaseExecutor
}

// NewLogExecutor 创建日志执行器实例
// 主要用于测试和调试，将任务执行信息输出到日志
// 返回:
//   *LogExecutor: 初始化的日志执行器实例
func NewLogExecutor() *LogExecutor {
	return &LogExecutor{
		BaseExecutor: NewBaseExecutor("LogExecutor", "Log message executor"),
	}
}

// Execute 执行日志输出任务
// 将任务参数格式化后输出到日志，主要用于测试和调试
// 参数:
//   ctx: 上下文，用于控制执行超时和取消
//   params: 任务参数，支持字符串或包含message字段的map
// 返回:
//   error: 始终返回nil，因为日志输出不会失败
func (e *LogExecutor) Execute(ctx context.Context, params interface{}) error {
	// 默认日志消息
	message := "Task executed"
	
	// 根据参数类型解析消息内容
	if params != nil {
		if msg, ok := params.(string); ok {
			// 参数是字符串类型，直接使用
			message = msg
		} else if msgMap, ok := params.(map[string]interface{}); ok {
			// 参数是map类型，尝试获取message字段
			if msg, exists := msgMap["message"]; exists {
				message = fmt.Sprintf("%v", msg)
			}
		}
	}
	
	// 输出格式化的日志信息
	log.Printf("[%s] %s", e.GetName(), message)
	return nil
}

// HTTPExecutor HTTP请求执行器
// 用于执行HTTP健康检查或API调用
type HTTPExecutor struct {
	*BaseExecutor
	client *http.Client
}

// NewHTTPExecutor 创建HTTP执行器实例
// 用于执行HTTP请求，支持健康检查、API调用等场景
// 内置30秒超时的HTTP客户端
// 返回:
//   *HTTPExecutor: 初始化的HTTP执行器实例
func NewHTTPExecutor() *HTTPExecutor {
	return &HTTPExecutor{
		BaseExecutor: NewBaseExecutor("HTTPExecutor", "HTTP request executor"),
		client: &http.Client{
			Timeout: 30 * time.Second,  // 默认30秒超时
		},
	}
}

// HTTPParams HTTP请求参数
type HTTPParams struct {
	URL    string            `json:"url"`
	Method string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

// Execute 执行HTTP请求任务
// 支持GET、POST等HTTP方法，可配置请求头，适用于健康检查和API调用
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   params: HTTP请求参数，包含url、method、headers等字段
// 返回:
//   error: 请求失败或状态码>=400时返回错误
func (e *HTTPExecutor) Execute(ctx context.Context, params interface{}) error {
	var httpParams HTTPParams
	
	if params == nil {
		return fmt.Errorf("HTTP parameters are required")
	}
	
	// 尝试解析参数
	if paramMap, ok := params.(map[string]interface{}); ok {
		if url, exists := paramMap["url"]; exists {
			httpParams.URL = fmt.Sprintf("%v", url)
		}
		if method, exists := paramMap["method"]; exists {
			httpParams.Method = fmt.Sprintf("%v", method)
		}
		if headers, exists := paramMap["headers"]; exists {
			if headersMap, ok := headers.(map[string]interface{}); ok {
				httpParams.Headers = make(map[string]string)
				for k, v := range headersMap {
					httpParams.Headers[k] = fmt.Sprintf("%v", v)
				}
			}
		}
	}
	
	if httpParams.URL == "" {
		return fmt.Errorf("URL is required")
	}
	
	if httpParams.Method == "" {
		httpParams.Method = "GET"
	}
	
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, httpParams.Method, httpParams.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// 设置请求头
	for k, v := range httpParams.Headers {
		req.Header.Set(k, v)
	}
	
	// 执行请求
	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}
	
	log.Printf("[%s] HTTP request to %s completed with status: %d", 
		e.GetName(), httpParams.URL, resp.StatusCode)
	
	return nil
}

// CommandExecutor 命令执行器
// 用于执行系统命令或脚本
type CommandExecutor struct {
	*BaseExecutor
}

// NewCommandExecutor 创建命令执行器实例
// 用于执行系统命令、脚本或外部程序，支持参数传递和工作目录设置
// 返回:
//   *CommandExecutor: 初始化的命令执行器实例
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{
		BaseExecutor: NewBaseExecutor("CommandExecutor", "System command executor"),
	}
}

// CommandParams 命令参数
type CommandParams struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Dir     string   `json:"dir"`
}

// Execute 执行系统命令任务
// 支持执行各种系统命令、脚本和外部程序，可指定参数和工作目录
// 参数:
//   ctx: 上下文，用于控制命令执行超时和取消
//   params: 命令参数，包含command、args、dir等字段
// 返回:
//   error: 命令执行失败时返回错误，包含输出信息
func (e *CommandExecutor) Execute(ctx context.Context, params interface{}) error {
	var cmdParams CommandParams
	
	if params == nil {
		return fmt.Errorf("command parameters are required")
	}
	
	// 尝试解析参数
	if paramMap, ok := params.(map[string]interface{}); ok {
		if command, exists := paramMap["command"]; exists {
			cmdParams.Command = fmt.Sprintf("%v", command)
		}
		if args, exists := paramMap["args"]; exists {
			if argsList, ok := args.([]interface{}); ok {
				for _, arg := range argsList {
					cmdParams.Args = append(cmdParams.Args, fmt.Sprintf("%v", arg))
				}
			}
		}
		if dir, exists := paramMap["dir"]; exists {
			cmdParams.Dir = fmt.Sprintf("%v", dir)
		}
	}
	
	if cmdParams.Command == "" {
		return fmt.Errorf("command is required")
	}
	
	// 创建命令
	cmd := exec.CommandContext(ctx, cmdParams.Command, cmdParams.Args...)
	if cmdParams.Dir != "" {
		cmd.Dir = cmdParams.Dir
	}
	
	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command execution failed: %w, output: %s", err, string(output))
	}
	
	log.Printf("[%s] Command '%s' executed successfully, output: %s", 
		e.GetName(), cmdParams.Command, string(output))
	
	return nil
} 