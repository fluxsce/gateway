package dialect

import (
	"errors"
	"strings"
)

// ErrorClass 驱动错误归类。sqlbase 再映射成 database.Err*，方言包不依赖 database。
type ErrorClass int

const (
	// ClassUnknown 未识别，原样返回。
	ClassUnknown ErrorClass = iota
	// ClassDuplicateKey 唯一约束冲突。
	ClassDuplicateKey
	// ClassConnection 连不上或连接已断。
	ClassConnection
	// ClassInvalidQuery SQL 无法执行。
	ClassInvalidQuery
)

// ClassifyByMessage 按错误文案归类，给没有原生错误码的驱动用。
func ClassifyByMessage(err error) ErrorClass {
	if err == nil {
		return ClassUnknown
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "duplicate entry"),
		strings.Contains(msg, "duplicate key"),
		strings.Contains(msg, "unique constraint"),
		strings.Contains(msg, "violation of unique key"),
		strings.Contains(msg, "violation of primary key"),
		strings.Contains(msg, "ora-00001"):
		return ClassDuplicateKey
	case strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "invalid connection"),
		strings.Contains(msg, "broken pipe"),
		strings.Contains(msg, "unable to open database"),
		strings.Contains(msg, "unable to open tcp"),
		strings.Contains(msg, "login failed"),
		strings.Contains(msg, "can not connect"):
		return ClassConnection
	case strings.Contains(msg, "syntax error"),
		strings.Contains(msg, "sql syntax"),
		strings.Contains(msg, "incorrect syntax"),
		strings.Contains(msg, "ora-00900"):
		return ClassInvalidQuery
	default:
		return ClassUnknown
	}
}

// UnwrapClass 顺着 error 链找第一个非 Unknown 分类。
func UnwrapClass(err error, classify func(error) ErrorClass) ErrorClass {
	if classify == nil {
		classify = ClassifyByMessage
	}
	for err != nil {
		if c := classify(err); c != ClassUnknown {
			return c
		}
		err = errors.Unwrap(err)
	}
	return ClassUnknown
}
