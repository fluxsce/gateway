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

// dropMissingIndexMarkers MySQL 1091：DROP INDEX 时索引已经不存在。
// 错误原文形如 Can't DROP '...'; check that column/key exists。
// 删列、删约束会带同一错误号，不能只凭 1091 跳过，必须同时是删索引语句。
var dropMissingIndexMarkers = []string{
	"can't drop",
	"check that column/key exists",
}

// isAlreadyPresentError 表、列、索引、约束或种子数据已经存在。
func isAlreadyPresentError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return containsAny(msg, alreadyPresentMarkers)
}

// isDropMissingIndexError 本条是删索引，且 MySQL 返回 1091（索引已不存在）。
// 初始化执行到这条语句时跳过，不记失败、不再重跑。
func isDropMissingIndexError(stmt string, err error) bool {
	if err == nil || !isDropIndexStatement(stmt) {
		return false
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "1091") {
		return false
	}
	return containsAny(msg, dropMissingIndexMarkers)
}

func containsAny(msg string, markers []string) bool {
	for _, marker := range markers {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// isDropIndexStatement 语句是在删索引。MySQL 里 DROP KEY 与 DROP INDEX 同义。
func isDropIndexStatement(stmt string) bool {
	s := strings.ToLower(strings.Join(strings.Fields(stmt), " "))
	s = strings.TrimSuffix(s, ";")
	return strings.Contains(s, " drop index ") || strings.HasPrefix(s, "drop index ") ||
		strings.Contains(s, " drop key ") || strings.HasPrefix(s, "drop key ")
}
