package retention

import (
	"fmt"
	"unicode"

	"gateway/pkg/database/sqlutils"
)

const (
	deleteBatchSize  = 2000
	maxDeleteBatches = 50
)

// limitedDelete 按库类型生成限制条数的 DELETE，走 sqlutils.BuildLimitedDeleteQuery。
// 不用主键 IN 列表，避免 Oracle 1000 项上限和占位符膨胀。
func limitedDelete(dbType sqlutils.DatabaseType, table, timeCol string, tenantId string, before interface{}) (string, []interface{}, error) {
	if !validTimeCol(timeCol) {
		return "", nil, fmt.Errorf("invalid time column")
	}
	where := "tenantId = ? AND " + timeCol + " < ?"
	return sqlutils.BuildLimitedDeleteQuery(dbType, table, where, []interface{}{tenantId, before}, deleteBatchSize)
}

func validTimeCol(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
