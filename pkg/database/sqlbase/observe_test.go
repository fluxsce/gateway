package sqlbase

import (
	"database/sql"
	"testing"

	"gateway/pkg/database"
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/database/dialect"

	mysqldriver "github.com/go-sql-driver/mysql"
)

func TestWrapErrDuplicateKey(t *testing.T) {
	d := New(dialect.MustGet(dbtypes.DriverMySQL), Hooks{})
	err := d.WrapErr(&mysqldriver.MySQLError{Number: 1062, Message: "Duplicate entry 'a' for key 'PRIMARY'"})
	if !database.IsDuplicateKey(err) {
		t.Fatalf("got %v", err)
	}
}

func TestWrapErrRecordNotFoundStaysSentinel(t *testing.T) {
	d := New(dialect.MustGet(dbtypes.DriverSQLite), Hooks{})
	if got := d.WrapErr(database.ErrRecordNotFound); got != database.ErrRecordNotFound {
		t.Fatalf("got %v", got)
	}
	if got := d.WrapErr(sql.ErrNoRows); got != database.ErrRecordNotFound {
		t.Fatalf("no rows should become sentinel, got %v", got)
	}
}
