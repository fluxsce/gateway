package sqlutils

import (
	"strings"
	"testing"
)

func TestBuildLimitedDeleteQueryMySQL(t *testing.T) {
	sql, args, err := BuildLimitedDeleteQuery(DatabaseMySQL, "HUB_ALERT_LOG", "tenantId = ? AND alertTimestamp < ?", []interface{}{"t1", "ts"}, 2000)
	if err != nil {
		t.Fatal(err)
	}
	want := "DELETE FROM HUB_ALERT_LOG WHERE tenantId = ? AND alertTimestamp < ? LIMIT ?"
	if sql != want {
		t.Fatalf("sql=%q", sql)
	}
	if len(args) != 3 || args[0] != "t1" || args[2] != 2000 {
		t.Fatalf("args=%v", args)
	}
}

func TestBuildLimitedDeleteQueryOracleNoINList(t *testing.T) {
	sql, args, err := BuildLimitedDeleteQuery(DatabaseOracle, "HUB_ALERT_LOG", "tenantId = ? AND alertTimestamp < ?", []interface{}{"t1", "ts"}, 2000)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, "AND ROWNUM <= ?") || !strings.Contains(sql, "DELETE FROM HUB_ALERT_LOG") {
		t.Fatalf("sql=%q", sql)
	}
	if strings.Contains(sql, " IN (") {
		t.Fatalf("oracle should not use IN list, sql=%q", sql)
	}
	if len(args) != 3 || args[2] != 2000 {
		t.Fatalf("args=%v", args)
	}
}

func TestBuildLimitedDeleteQuerySQLServerTopFirst(t *testing.T) {
	_, args, err := BuildLimitedDeleteQuery(DatabaseSQLServer, "HUB_ALERT_LOG", "tenantId = ?", []interface{}{"t1"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if args[0] != 100 {
		t.Fatalf("TOP limit should be first, args=%v", args)
	}
}

func TestBuildLimitedDeleteQueryRejectsBadTable(t *testing.T) {
	if _, _, err := BuildLimitedDeleteQuery(DatabaseMySQL, "HUB_ALERT_LOG;DROP", "id = ?", []interface{}{1}, 10); err == nil {
		t.Fatal("expected invalid table")
	}
}
