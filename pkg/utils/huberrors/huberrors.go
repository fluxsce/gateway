package huberrors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// HubError 带有位置和堆栈信息的错误类型
type HubError struct {
	Message  string  // 错误消息
	File     string  // 发生错误的文件
	Line     int     // 发生错误的行号
	Function string  // 发生错误的函数
	Err      error   // 原始错误
	Stack    []Frame // 错误发生时的调用栈
}

// Frame 代表调用栈中的一帧
type Frame struct {
	File     string // 文件路径
	Line     int    // 行号
	Function string // 函数名
}

// Error 实现error接口，返回格式化的错误信息
func (e *HubError) Error() string {
	var result strings.Builder

	// 添加当前错误信息和位置
	if e.File != "" {
		result.WriteString(fmt.Sprintf("%s (at %s:%d in %s)", e.Message, e.File, e.Line, e.Function))
	} else {
		result.WriteString(e.Message)
	}

	// 添加原始错误信息
	if e.Err != nil {
		result.WriteString(fmt.Sprintf("\n原因: %v", e.Err))
	}

	return result.String()
}

// Unwrap 返回原始错误，支持errors.Is/As()
func (e *HubError) Unwrap() error {
	return e.Err
}

// FullError 返回包含完整调用栈的错误信息
func (e *HubError) FullError() string {
	var result strings.Builder

	// 添加当前错误信息
	result.WriteString(fmt.Sprintf("错误: %s\n", e.Message))

	// 添加错误位置
	if e.File != "" {
		result.WriteString(fmt.Sprintf("位置: %s:%d (%s)\n", e.File, e.Line, e.Function))
	}

	// 添加调用栈信息
	if len(e.Stack) > 0 {
		result.WriteString("调用栈:\n")
		for i, frame := range e.Stack {
			result.WriteString(fmt.Sprintf("  %d: %s:%d (%s)\n", i+1, frame.File, frame.Line, frame.Function))
		}
	}

	// 递归添加原因链
	if e.Err != nil {
		result.WriteString("\n错误原因:\n")

		// 如果嵌套的也是HubError，获取其完整错误信息
		if hubErr, ok := e.Err.(*HubError); ok {
			// 缩进嵌套错误信息
			nestedErr := hubErr.FullError()
			lines := strings.Split(nestedErr, "\n")
			for _, line := range lines {
				result.WriteString("  " + line + "\n")
			}
		} else {
			// 普通错误直接显示
			result.WriteString("  " + e.Err.Error() + "\n")
		}
	}

	return result.String()
}

// 收集当前调用栈
func captureStack(skip int) []Frame {
	const depth = 32
	var pcs [depth]uintptr

	// 跳过此函数和上层调用者
	n := runtime.Callers(skip, pcs[:])
	frames := make([]Frame, 0, n)

	for _, pc := range pcs[0:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc)
		frames = append(frames, Frame{
			File:     file,
			Line:     line,
			Function: fn.Name(),
		})
	}

	return frames
}

// NewError 创建一个带有位置信息和调用栈的错误
// 此函数会自动获取调用者的文件名、行号和函数名
func NewError(msg string, args ...interface{}) error {
	// 获取调用者信息
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	formattedMsg := fmt.Sprintf(msg, args...)

	return &HubError{
		Message:  formattedMsg,
		File:     file,
		Line:     line,
		Function: fn.Name(),
		Stack:    captureStack(2), // 跳过NewError自身
	}
}

// WrapError 包装现有错误并添加位置信息和调用栈
func WrapError(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	// 获取调用者信息
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	formattedMsg := fmt.Sprintf(msg, args...)

	return &HubError{
		Message:  formattedMsg,
		File:     file,
		Line:     line,
		Function: fn.Name(),
		Err:      err,
		Stack:    captureStack(2), // 跳过WrapError自身
	}
}

// ErrorStack 为任何错误提取完整的调用栈信息
// 适用于普通error或HubError类型
func ErrorStack(err error) string {
	if err == nil {
		return ""
	}

	// 如果是HubError，直接使用其FullError方法
	if hubErr, ok := err.(*HubError); ok {
		return hubErr.FullError()
	}

	// 如果是普通错误，将其包装为HubError
	wrapped := WrapError(err, "错误")
	if hubErr, ok := wrapped.(*HubError); ok {
		return hubErr.FullError()
	}

	// 兜底返回
	return err.Error()
}

// Location 返回错误发生的位置信息
// 返回:
//   - file: 文件路径
//   - line: 行号
//   - function: 函数名
func (e *HubError) Location() (file string, line int, function string) {
	return e.File, e.Line, e.Function
}

// GetRootCause 获取最底层原始错误
// 即持续调用Unwrap直到找到最深层的错误
func GetRootCause(err error) error {
	for err != nil {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
	return err
}
