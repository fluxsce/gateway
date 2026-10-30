package sqlutils

import (
	"strings"
	"testing"

	"gateway/pkg/database/dbtypes"
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

func TestBuildLimitedDeleteQuerySQLiteUsesCTE(t *testing.T) {
	sql, args, err := BuildLimitedDeleteQuery(DatabaseSQLite, "HUB_METRIC_TEMP_LOG", "tenantId = ? AND collectTime < ?", []interface{}{"t1", "ts"}, 2000)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, "WITH doomed AS") || !strings.Contains(sql, "DELETE FROM HUB_METRIC_TEMP_LOG") {
		t.Fatalf("sql=%q", sql)
	}
	if strings.Contains(sql, "SELECT rowid FROM (SELECT rowid FROM") {
		t.Fatalf("should not nest same-table rowid subquery, sql=%q", sql)
	}
	if len(args) != 3 || args[2] != 2000 {
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

func TestDatabaseTypeForVastbase(t *testing.T) {
	if databaseTypeFor(dbtypes.DriverVastbase, "PG") != DatabasePostgreSQL {
		t.Fatal("PG mode should use postgres dialect")
	}
	if databaseTypeFor(dbtypes.DriverVastbase, "A") != DatabaseVastbaseOracle {
		t.Fatal("A mode should use vastbase oracle dialect")
	}
	if databaseTypeFor(dbtypes.DriverVastbase, "") != DatabasePostgreSQL {
		t.Fatal("unknown mode should stay on postgres dialect")
	}
}

func TestLimitOffsetClauseVastbaseOracle(t *testing.T) {
	got := LimitOffsetClause(DatabaseVastbaseOracle, 20, 40)
	if got != "OFFSET 40 ROWS FETCH NEXT 20 ROWS ONLY" {
		t.Fatalf("got %q", got)
	}
	if mysql := LimitOffsetClause(DatabaseMySQL, 20, 0); mysql != "LIMIT 20" {
		t.Fatalf("mysql %q", mysql)
	}
}

func TestBuildLimitedDeleteQueryVastbaseOracle(t *testing.T) {
	sql, _, err := BuildLimitedDeleteQuery(DatabaseVastbaseOracle, "HUB_ALERT_LOG", "tenantId = ?", []interface{}{"t1"}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, "ROWNUM") || strings.Contains(sql, "ctid") {
		t.Fatalf("sql=%q", sql)
	}
}

func TestBuildLimitedDeleteQueryRejectsBadTable(t *testing.T) {
	if _, _, err := BuildLimitedDeleteQuery(DatabaseMySQL, "HUB_ALERT_LOG;DROP", "id = ?", []interface{}{1}, 10); err == nil {
		t.Fatal("expected invalid table")
	}
}
