package db

import "strings"

// alreadyPresentMarkers 库里对象或数据已经在时，驱动返回的典型报错。
// 这类语句再执行也不会改变现状，启动时不应当成失败反复重跑。
var alreadyPresentMarkers = []string{
	"already exists",
	"duplicate entry",
	"duplicate column",
	"duplicate key",
	"multiple primary key",
	"unique constraint failed",
	"unique constraint violated",
	"ora-00955",
	"ora-01430",
	"ora-01408",
	"ora-00001",
	"ora-02260",
	"ora-02261",
	"ora-02264",
	"ora-02275",
	"there is already an object named",
	"violation of primary key",
	"violation of unique key",
	"cannot insert duplicate key",
	"column names in each table must be unique",
}

// isAlreadyPresentError 表、列、索引、约束或种子数据已经存在。
func isAlreadyPresentError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, marker := range alreadyPresentMarkers {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}
