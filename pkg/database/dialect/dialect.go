// Package dialect 描述各数据库在 SQL、占位符和 DSN 上的差异。
// 新增关系库时优先实现 Dialect 并 Register，避免再改 sqlutils / dsn 的中心 switch。
package dialect

import (
	"fmt"
	"sync"

	"gateway/pkg/database/dbtypes"
)

// Dialect 一种数据库的方言策略。
// sqlutils 拼 SQL、dsn 生成连接串、sqlbase 执行时都通过本接口取差异，而不是按驱动名写 switch。
type Dialect interface {
	// Name 返回规范驱动名，与 dbtypes.Driver* 一致。
	Name() string
	// Aliases 返回可共用本方言的驱动别名，例如 mariadb、tidb。
	Aliases() []string
	// OpenDriver 返回 database/sql.Open 的驱动名，例如 mysql、sqlite3、godror。
	OpenDriver() string
	// FallbackDSN 在配置未提供 DSN 且生成器也无法给出时使用，例如 SQLite 的 :memory:。
	FallbackDSN() string
	// RewriteQuery 把 SQL 里的 ? 改成该库驱动认识的占位符。问号风格原样返回。
	RewriteQuery(query string) string
	// Placeholder 返回第 index 个占位符（从 1 起），用于拼 IN 列表等。
	Placeholder(index int) string
	// BuildPagination 给基础查询加上分页，参数仍用 ?，执行前由 RewriteQuery 转换。
	BuildPagination(baseQuery string, page, pageSize, offset int) (string, []interface{}, error)
	// BuildLimitedDelete 生成限制条数的删除语句，参数顺序与历史实现一致。
	BuildLimitedDelete(table, where string, whereArgs []interface{}, limit int) (string, []interface{}, error)
	// CurrentTimeFunction 返回该库的当前时间函数，例如 NOW()、SYSDATE。
	CurrentTimeFunction() (string, error)
	// GenerateDSN 按连接配置生成 DSN。调用方已有 DSN 时不应再调本方法。
	GenerateDSN(cfg *dbtypes.DbConfig) (string, error)
	// ValidateDSN 检查 DSN 外观是否像该库。
	ValidateDSN(dsn string) error
	// UpdateSQL 拼 UPDATE/ALTER UPDATE。where 为空则不加 WHERE。
	UpdateSQL(table, setClause, where string) string
	// DeleteSQL 拼 DELETE/ALTER DELETE。where 为空则不加 WHERE。
	DeleteSQL(table, where string) string
	// BatchDeleteByKeysSQL 按主键 IN 列表批量删除。
	BatchDeleteByKeysSQL(table, keyField string, keyCount int) string
	// ClassifyError 把驱动错误归成可稳定处理的类别。
	ClassifyError(err error) ErrorClass
}

var (
	regMu    sync.RWMutex
	registry = make(map[string]Dialect)
)

// Register 注册方言。同名后注册覆盖先注册。别名一并登记。
func Register(d Dialect) {
	if d == nil || d.Name() == "" {
		return
	}
	regMu.Lock()
	defer regMu.Unlock()
	registry[d.Name()] = d
	for _, alias := range d.Aliases() {
		if alias != "" {
			registry[alias] = d
		}
	}
}

// Get 按驱动名或别名取方言。
func Get(name string) (Dialect, error) {
	regMu.RLock()
	d, ok := registry[name]
	regMu.RUnlock()
	if !ok || d == nil {
		return nil, fmt.Errorf("unsupported database type: %s", name)
	}
	return d, nil
}

// MustGet 取方言，未注册则 panic。仅用于已知已注册的驱动初始化。
func MustGet(name string) Dialect {
	d, err := Get(name)
	if err != nil {
		panic(err)
	}
	return d
}
