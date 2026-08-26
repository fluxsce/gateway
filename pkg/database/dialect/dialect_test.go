package dialect

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"gateway/pkg/database/dbtypes"

	mysqldriver "github.com/go-sql-driver/mysql"
	mssql "github.com/microsoft/go-mssqldb"
)

func TestRewriteOraclePlaceholders(t *testing.T) {
	got := RewriteQuery("SELECT * FROM t WHERE a = ? AND b = ?", ColonNum)
	if got != "SELECT * FROM t WHERE a = :1 AND b = :2" {
		t.Fatalf("got %q", got)
	}
}

func TestRewritePostgresAndSQLServer(t *testing.T) {
	if got := RewriteQuery("SELECT * FROM t WHERE a = ?", DollarNum); got != "SELECT * FROM t WHERE a = $1" {
		t.Fatalf("postgres got %q", got)
	}
	if got := RewriteQuery("SELECT * FROM t WHERE a = ?", AtP); got != "SELECT * FROM t WHERE a = @p1" {
		t.Fatalf("sqlserver got %q", got)
	}
}

func TestGetAliases(t *testing.T) {
	d, err := Get(dbtypes.DriverMariaDB)
	if err != nil {
		t.Fatal(err)
	}
	if d.Name() != dbtypes.DriverMySQL {
		t.Fatalf("alias should resolve to mysql spec, got %s", d.Name())
	}
}

func TestLimitedDeleteDelegates(t *testing.T) {
	d := MustGet(dbtypes.DriverMySQL)
	sql, args, err := d.BuildLimitedDelete("HUB_ALERT_LOG", "tenantId = ?", []interface{}{"t1"}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if sql != "DELETE FROM HUB_ALERT_LOG WHERE tenantId = ? LIMIT ?" {
		t.Fatalf("sql=%q", sql)
	}
	if len(args) != 2 || args[1] != 10 {
		t.Fatalf("args=%v", args)
	}
}

func TestClickHouseMutationSQL(t *testing.T) {
	d := MustGet(dbtypes.DriverClickHouse)
	upd := d.UpdateSQL("t", "c = ?", "id = ?")
	if !strings.HasPrefix(upd, "ALTER TABLE t UPDATE") {
		t.Fatalf("update=%q", upd)
	}
	del := d.DeleteSQL("t", "id = ?")
	if !strings.HasPrefix(del, "ALTER TABLE t DELETE") {
		t.Fatalf("delete=%q", del)
	}
}

func TestMongoNotADialect(t *testing.T) {
	if _, err := Get(dbtypes.DriverMongoDB); err == nil {
		t.Fatal("mongodb must not register as SQL dialect")
	}
}

func TestClassifyDuplicateKey(t *testing.T) {
	if ClassifyByMessage(errors.New("UNIQUE constraint failed: users.name")) != ClassDuplicateKey {
		t.Fatal("sqlite unique")
	}
	if ClassifyByMessage(errors.New("duplicate key value violates unique constraint")) != ClassDuplicateKey {
		t.Fatal("postgres duplicate")
	}
	if ClassifyByMessage(errors.New("ORA-00001: unique constraint violated")) != ClassDuplicateKey {
		t.Fatal("oracle unique")
	}
	if ClassifyByMessage(errors.New("Violation of UNIQUE KEY constraint")) != ClassDuplicateKey {
		t.Fatal("sqlserver unique")
	}
	ss := MustGet(dbtypes.DriverSQLServer)
	if ss.ClassifyError(mssql.Error{Number: 2627, Message: "Violation of PRIMARY KEY constraint"}) != ClassDuplicateKey {
		t.Fatal("sqlserver 2627")
	}
	d := MustGet(dbtypes.DriverMySQL)
	me := &mysqldriver.MySQLError{Number: 1062, Message: "Duplicate entry"}
	if d.ClassifyError(me) != ClassDuplicateKey {
		t.Fatal("mysql 1062")
	}
}

func TestGenerateSQLServerDSN(t *testing.T) {
	d := MustGet(dbtypes.DriverSQLServer)
	got, err := d.GenerateDSN(&dbtypes.DbConfig{
		Driver: dbtypes.DriverSQLServer,
		Connection: dbtypes.ConnectionConfig{
			Host:     "localhost",
			Port:     1433,
			Username: "sa",
			Password: "Secret",
			Database: "gateway",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "sqlserver://") {
		t.Fatalf("dsn=%q", got)
	}
	if !strings.Contains(got, "database=gateway") {
		t.Fatalf("missing database: %q", got)
	}
	if !strings.Contains(got, "encrypt=disable") {
		t.Fatalf("missing encrypt: %q", got)
	}
	alias, err := Get("mssql")
	if err != nil || alias.Name() != dbtypes.DriverSQLServer {
		t.Fatalf("mssql alias: %v %v", alias, err)
	}
}

func TestGenerateSQLServerDSNScenarios(t *testing.T) {
	d := MustGet(dbtypes.DriverSQLServer)

	t.Run("named instance uses SQL Browser path", func(t *testing.T) {
		got, err := d.GenerateDSN(&dbtypes.DbConfig{Connection: dbtypes.ConnectionConfig{
			Host:              "dbhost",
			Username:          "sa",
			Password:          "Secret",
			Database:          "gateway",
			SQLServerInstance: "SQLEXPRESS",
		}})
		if err != nil {
			t.Fatal(err)
		}
		u, err := url.Parse(got)
		if err != nil {
			t.Fatal(err)
		}
		if u.Host != "dbhost" || u.Path != "/SQLEXPRESS" {
			t.Fatalf("browser dsn host/path = %q %q", u.Host, u.Path)
		}
	})

	t.Run("named instance with static port", func(t *testing.T) {
		got, err := d.GenerateDSN(&dbtypes.DbConfig{Connection: dbtypes.ConnectionConfig{
			Host:              "dbhost",
			Port:              1500,
			Username:          "sa",
			Password:          "Secret",
			Database:          "gateway",
			SQLServerInstance: "SQLEXPRESS",
		}})
		if err != nil {
			t.Fatal(err)
		}
		u, err := url.Parse(got)
		if err != nil {
			t.Fatal(err)
		}
		if u.Host != "dbhost:1500" || u.Path != "/SQLEXPRESS" {
			t.Fatalf("static port dsn host/path = %q %q", u.Host, u.Path)
		}
	})

	t.Run("azure sql encrypt and certificate host", func(t *testing.T) {
		got, err := d.GenerateDSN(&dbtypes.DbConfig{Connection: dbtypes.ConnectionConfig{
			Host:                           "myserver.database.windows.net",
			Port:                           1433,
			Username:                       "app",
			Password:                       "Secret",
			Database:                       "gateway",
			SQLServerEncrypt:               "true",
			SQLServerHostNameInCertificate: "*.database.windows.net",
		}})
		if err != nil {
			t.Fatal(err)
		}
		u, err := url.Parse(got)
		if err != nil {
			t.Fatal(err)
		}
		q := u.Query()
		if q.Get("encrypt") != "true" {
			t.Fatalf("encrypt=%q", q.Get("encrypt"))
		}
		if q.Get("hostNameInCertificate") != "*.database.windows.net" {
			t.Fatalf("cert host=%q", q.Get("hostNameInCertificate"))
		}
	})

	t.Run("windows integrated auth omits user", func(t *testing.T) {
		got, err := d.GenerateDSN(&dbtypes.DbConfig{Connection: dbtypes.ConnectionConfig{
			Host:                            "dbhost",
			Database:                        "gateway",
			SQLServerAuthenticator:          "winsspi",
			SQLServerEncrypt:                "true",
			SQLServerTrustServerCertificate: true,
		}})
		if err != nil {
			t.Fatal(err)
		}
		u, err := url.Parse(got)
		if err != nil {
			t.Fatal(err)
		}
		if u.User != nil {
			t.Fatalf("expected no user, got %q", u.User.String())
		}
		if u.Query().Get("authenticator") != "winsspi" {
			t.Fatalf("authenticator=%q", u.Query().Get("authenticator"))
		}
	})

	t.Run("always on read-only listener", func(t *testing.T) {
		got, err := d.GenerateDSN(&dbtypes.DbConfig{Connection: dbtypes.ConnectionConfig{
			Host:                         "ag-listener",
			Username:                     "sa",
			Password:                     "Secret",
			Database:                     "gateway",
			SQLServerApplicationIntent:   "ReadOnly",
			SQLServerMultiSubnetFailover: true,
			SQLServerEncrypt:             "true",
		}})
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.Parse(got)
		if err != nil {
			t.Fatal(err)
		}
		vals := q.Query()
		if vals.Get("ApplicationIntent") != "ReadOnly" {
			t.Fatalf("intent=%q", vals.Get("ApplicationIntent"))
		}
		if vals.Get("MultiSubnetFailover") != "true" {
			t.Fatalf("msf=%q", vals.Get("MultiSubnetFailover"))
		}
	})
}

func TestSQLServerLimitedDeleteArgOrder(t *testing.T) {
	d := MustGet(dbtypes.DriverSQLServer)
	sql, args, err := d.BuildLimitedDelete("HUB_ALERT_LOG", "tenantId = ?", []interface{}{"t1"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if sql != "DELETE TOP (?) FROM HUB_ALERT_LOG WHERE tenantId = ?" {
		t.Fatalf("sql=%q", sql)
	}
	if len(args) != 2 || args[0] != 100 || args[1] != "t1" {
		t.Fatalf("args=%v", args)
	}
}
