package dialect

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"

	"gateway/pkg/database/dbtypes"
	huberrors "gateway/pkg/utils/huberrors"

	mssql "github.com/microsoft/go-mssqldb"
)

func init() {
	Register(&Spec{
		name:       dbtypes.DriverSQLServer,
		aliases:    []string{"mssql"},
		openDriver: "sqlserver",
		style:      AtP,
		timeFn:     "GETDATE()",
		paginate: func(baseQuery string, page, pageSize, offset int) (string, []interface{}, error) {
			return paginateOffsetFetch(baseQuery, "ORDER BY (SELECT NULL)", pageSize, offset)
		},
		limitedDel: limitedDeleteSQLServer,
		generate:   generateSQLServer,
		validate:   validateSQLServerDSN,
		classify:   classifySQLServer,
	})
}

func classifySQLServer(err error) ErrorClass {
	var se mssql.Error
	if errors.As(err, &se) {
		switch se.Number {
		case 2627, 2601:
			return ClassDuplicateKey
		case 102, 156, 207, 208:
			return ClassInvalidQuery
		}
	}
	return ClassifyByMessage(err)
}

func validateSQLServerDSN(dsn string) error {
	lower := strings.ToLower(dsn)
	if strings.HasPrefix(lower, "sqlserver://") || strings.Contains(lower, "server=") {
		return nil
	}
	return huberrors.NewError("SQL Server DSN格式不正确，应以sqlserver://开头或使用 ADO 连接串")
}

func generateSQLServer(config *dbtypes.DbConfig) (string, error) {
	if config.Connection.Host == "" {
		return "", huberrors.NewError("SQL Server数据库需要host参数")
	}
	if config.Connection.Database == "" {
		return "", huberrors.NewError("SQL Server数据库需要database参数")
	}

	c := config.Connection
	instance := strings.TrimSpace(c.SQLServerInstance)

	u := &url.URL{Scheme: "sqlserver"}
	if strings.TrimSpace(c.SQLServerAuthenticator) == "" && c.Username != "" {
		if c.Password != "" {
			u.User = url.UserPassword(c.Username, c.Password)
		} else {
			u.User = url.User(c.Username)
		}
	}

	// 命名实例写在 URL path：port=0 走 SQL Browser；port>0 直连该端口。默认实例无 path，端口缺省 1433。
	switch {
	case instance != "" && c.Port <= 0:
		u.Host = c.Host
		u.Path = "/" + instance
	case instance != "":
		u.Host = net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
		u.Path = "/" + instance
	default:
		port := c.Port
		if port <= 0 {
			port = 1433
		}
		u.Host = net.JoinHostPort(c.Host, strconv.Itoa(port))
	}

	q := url.Values{}
	q.Set("database", c.Database)

	encrypt := strings.TrimSpace(c.SQLServerEncrypt)
	if encrypt == "" {
		encrypt = "disable"
	}
	encrypt = strings.ToLower(encrypt)
	switch encrypt {
	case "disable", "false", "true", "strict":
	default:
		return "", huberrors.NewError("不支持的 SQL Server encrypt: %s，支持: disable,false,true,strict", encrypt)
	}
	q.Set("encrypt", encrypt)

	if c.SQLServerTrustServerCertificate {
		q.Set("TrustServerCertificate", "true")
	}
	if c.SQLServerConnectionTimeout > 0 {
		q.Set("connection timeout", strconv.Itoa(c.SQLServerConnectionTimeout))
	}
	if c.SQLServerDialTimeout > 0 {
		q.Set("dial timeout", strconv.Itoa(c.SQLServerDialTimeout))
	}
	if c.SQLServerAppName != "" {
		q.Set("app name", c.SQLServerAppName)
	}
	if c.SQLServerWorkstationID != "" {
		q.Set("workstation id", c.SQLServerWorkstationID)
	}
	if c.SQLServerKeepAlive > 0 {
		q.Set("keepalive", strconv.Itoa(c.SQLServerKeepAlive))
	}
	if auth := strings.TrimSpace(c.SQLServerAuthenticator); auth != "" {
		q.Set("authenticator", strings.ToLower(auth))
	}
	if intent := strings.TrimSpace(c.SQLServerApplicationIntent); intent != "" {
		q.Set("ApplicationIntent", intent)
	}
	if c.SQLServerMultiSubnetFailover {
		q.Set("MultiSubnetFailover", "true")
	}
	if c.SQLServerFailoverPartner != "" {
		q.Set("failoverpartner", c.SQLServerFailoverPartner)
	}
	if c.SQLServerHostNameInCertificate != "" {
		q.Set("hostNameInCertificate", c.SQLServerHostNameInCertificate)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}
