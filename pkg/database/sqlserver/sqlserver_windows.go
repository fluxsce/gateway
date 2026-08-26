//go:build windows

package sqlserver

import (
	_ "github.com/microsoft/go-mssqldb/integratedauth/winsspi"
)
