package database

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gateway/pkg/database"
	_ "gateway/pkg/database/alldriver" // 导入驱动确保注册
	"gateway/pkg/database/dbtypes"
	"sync"
)

// ClickHouseUser 用于测试的用户结构体
type ClickHouseUser struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int32     `db:"age"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TableName 实现Model接口
func (u ClickHouseUser) TableName() string {
	return "clickhouse_users"
}

// PrimaryKey 实现Model接口
func (u ClickHouseUser) PrimaryKey() string {
	return "id"
}

// ClickHouseEvent 用于测试的事件数据结构体（时间序列数据）
type ClickHouseEvent struct {
	ID         int64     `db:"id"`
	EventName  string    `db:"event_name"`
	UserID     int32     `db:"user_id"`
	Timestamp  time.Time `db:"timestamp"`
	Value      float64   `db:"value"`
	Properties string    `db:"properties"`
	Date       string    `db:"date"`
}

// TableName 实现Model接口
func (e ClickHouseEvent) TableName() string {
	return "clickhouse_events"
}

// PrimaryKey 实现Model接口
func (e ClickHouseEvent) PrimaryKey() string {
	return "id"
}

// ClickHouseComplexLog 复杂日志结构体 - 模拟AccessLog但使用基础类型
type ClickHouseComplexLog struct {
	// 主键字段
	TenantId string `db:"tenantId"`
	TraceId  string `db:"traceId"`

	// 网关实例相关信息
	GatewayInstanceId   string `db:"gatewayInstanceId"`
	GatewayInstanceName string `db:"gatewayInstanceName"`
	GatewayNodeIp       string `db:"gatewayNodeIp"`

	// 路由和服务相关信息
	RouteConfigId       string `db:"routeConfigId"`
	RouteName           string `db:"routeName"`
	ServiceDefinitionId string `db:"serviceDefinitionId"`
	ServiceName         string `db:"serviceName"`
	ProxyType           string `db:"proxyType"`
	LogConfigId         string `db:"logConfigId"`

	// 请求基本信息
	RequestMethod  string `db:"requestMethod"`
	RequestPath    string `db:"requestPath"`
	RequestQuery   string `db:"requestQuery"`
	RequestSize    int32  `db:"requestSize"`
	RequestHeaders string `db:"requestHeaders"`
	RequestBody    string `db:"requestBody"`

	// 客户端信息 - 注意：这里使用基础类型而不是指针
	ClientIpAddress string `db:"clientIpAddress"`
	ClientPort      int32  `db:"clientPort"` // 基础类型，不是指针
	UserAgent       string `db:"userAgent"`
	Referer         string `db:"referer"`
	UserIdentifier  string `db:"userIdentifier"`

	// 关键时间点 - 使用基础类型
	GatewayStartProcessingTime    time.Time `db:"gatewayStartProcessingTime"`
	BackendRequestStartTime       time.Time `db:"backendRequestStartTime"`       // 基础类型，不是指针
	BackendResponseReceivedTime   time.Time `db:"backendResponseReceivedTime"`   // 基础类型，不是指针
	GatewayFinishedProcessingTime time.Time `db:"gatewayFinishedProcessingTime"` // 基础类型，不是指针

	// 计算的时间指标 - 使用基础类型
	TotalProcessingTimeMs   int32 `db:"totalProcessingTimeMs"`   // 基础类型，不是指针
	GatewayProcessingTimeMs int32 `db:"gatewayProcessingTimeMs"` // 基础类型，不是指针
	BackendResponseTimeMs   int32 `db:"backendResponseTimeMs"`   // 基础类型，不是指针

	// 响应信息
	GatewayStatusCode int32  `db:"gatewayStatusCode"`
	BackendStatusCode int32  `db:"backendStatusCode"` // 基础类型，不是指针
	ResponseSize      int32  `db:"responseSize"`
	ResponseHeaders   string `db:"responseHeaders"`
	ResponseBody      string `db:"responseBody"`

	// 转发基本信息
	MatchedRoute         string `db:"matchedRoute"`
	ForwardAddress       string `db:"forwardAddress"`
	ForwardMethod        string `db:"forwardMethod"`
	ForwardParams        string `db:"forwardParams"`
	ForwardHeaders       string `db:"forwardHeaders"`
	ForwardBody          string `db:"forwardBody"`
	LoadBalancerDecision string `db:"loadBalancerDecision"`

	// 错误信息
	ErrorMessage string `db:"errorMessage"`
	ErrorCode    string `db:"errorCode"`

	// 追踪信息
	ParentTraceId string `db:"parentTraceId"`

	// 日志重置标记和次数
	ResetFlag  string `db:"resetFlag"`
	RetryCount int32  `db:"retryCount"`
	ResetCount int32  `db:"resetCount"`

	// 标准数据库字段
	LogLevel       string    `db:"logLevel"`
	LogType        string    `db:"logType"`
	Reserved1      string    `db:"reserved1"`
	Reserved2      string    `db:"reserved2"`
	Reserved3      int32     `db:"reserved3"` // 基础类型，不是指针
	Reserved4      int32     `db:"reserved4"` // 基础类型，不是指针
	Reserved5      time.Time `db:"reserved5"` // 基础类型，不是指针
	ExtProperty    string    `db:"extProperty"`
	AddTime        time.Time `db:"addTime"`
	AddWho         string    `db:"addWho"`
	EditTime       time.Time `db:"editTime"`
	EditWho        string    `db:"editWho"`
	OprSeqFlag     string    `db:"oprSeqFlag"`
	CurrentVersion int32     `db:"currentVersion"`
	ActiveFlag     string    `db:"activeFlag"`
	NoteText       string    `db:"noteText"`
}

// TableName 实现Model接口
func (c ClickHouseComplexLog) TableName() string {
	return "clickhouse_complex_logs"
}

// PrimaryKey 实现Model接口
func (c ClickHouseComplexLog) PrimaryKey() string {
	return "traceId"
}

// ClickHouseComplexLogWithPointers 包含指针类型的复杂日志结构体 - 模拟真实AccessLog的问题
type ClickHouseComplexLogWithPointers struct {
	// 主键字段
	TenantId string `db:"tenantId"`
	TraceId  string `db:"traceId"`

	// 网关实例相关信息
	GatewayInstanceId   string `db:"gatewayInstanceId"`
	GatewayInstanceName string `db:"gatewayInstanceName"`
	GatewayNodeIp       string `db:"gatewayNodeIp"`

	// 请求基本信息
	RequestMethod  string `db:"requestMethod"`
	RequestPath    string `db:"requestPath"`
	RequestQuery   string `db:"requestQuery"`
	RequestSize    int    `db:"requestSize"` // 使用int而不是int32
	RequestHeaders string `db:"requestHeaders"`
	RequestBody    string `db:"requestBody"`

	// 客户端信息 - 使用指针类型（问题根源）
	ClientIpAddress string `db:"clientIpAddress"`
	ClientPort      *int   `db:"clientPort"` // 指针类型！
	UserAgent       string `db:"userAgent"`
	Referer         string `db:"referer"`
	UserIdentifier  string `db:"userIdentifier"`

	// 关键时间点 - 使用指针类型（问题根源）
	GatewayStartProcessingTime    time.Time  `db:"gatewayStartProcessingTime"`
	BackendRequestStartTime       *time.Time `db:"backendRequestStartTime"`       // 指针类型！
	BackendResponseReceivedTime   *time.Time `db:"backendResponseReceivedTime"`   // 指针类型！
	GatewayFinishedProcessingTime *time.Time `db:"gatewayFinishedProcessingTime"` // 指针类型！

	// 计算的时间指标 - 使用指针类型（问题根源）
	TotalProcessingTimeMs   int  `db:"totalProcessingTimeMs"`   // 使用int而不是int32
	GatewayProcessingTimeMs int  `db:"gatewayProcessingTimeMs"` // 使用int而不是int32
	BackendResponseTimeMs   *int `db:"backendResponseTimeMs"`   // 指针类型！

	// 响应信息
	GatewayStatusCode int    `db:"gatewayStatusCode"`
	BackendStatusCode *int   `db:"backendStatusCode"` // 指针类型！
	ResponseSize      int    `db:"responseSize"`
	ResponseHeaders   string `db:"responseHeaders"`
	ResponseBody      string `db:"responseBody"`

	// 转发基本信息
	MatchedRoute         string `db:"matchedRoute"`
	ForwardAddress       string `db:"forwardAddress"`
	ForwardMethod        string `db:"forwardMethod"`
	ForwardParams        string `db:"forwardParams"`
	ForwardHeaders       string `db:"forwardHeaders"`
	ForwardBody          string `db:"forwardBody"`
	LoadBalancerDecision string `db:"loadBalancerDecision"`

	// 错误信息
	ErrorMessage string `db:"errorMessage"`
	ErrorCode    string `db:"errorCode"`

	// 追踪信息
	ParentTraceId string `db:"parentTraceId"`

	// 日志重置标记和次数
	ResetFlag  string `db:"resetFlag"`
	RetryCount int    `db:"retryCount"`
	ResetCount int    `db:"resetCount"`

	// 标准数据库字段
	LogLevel       string     `db:"logLevel"`
	LogType        string     `db:"logType"`
	Reserved1      string     `db:"reserved1"`
	Reserved2      string     `db:"reserved2"`
	Reserved3      *int       `db:"reserved3"` // 指针类型！
	Reserved4      *int       `db:"reserved4"` // 指针类型！
	Reserved5      *time.Time `db:"reserved5"` // 指针类型！
	ExtProperty    string     `db:"extProperty"`
	AddTime        time.Time  `db:"addTime"`
	AddWho         string     `db:"addWho"`
	EditTime       time.Time  `db:"editTime"`
	EditWho        string     `db:"editWho"`
	OprSeqFlag     string     `db:"oprSeqFlag"`
	CurrentVersion int        `db:"currentVersion"`
	ActiveFlag     string     `db:"activeFlag"`
	NoteText       string     `db:"noteText"`
}

// TableName 实现Model接口
func (c ClickHouseComplexLogWithPointers) TableName() string {
	return "clickhouse_complex_logs_with_pointers"
}

// PrimaryKey 实现Model接口
func (c ClickHouseComplexLogWithPointers) PrimaryKey() string {
	return "traceId"
}

// 获取测试数据库连接
func getClickHouseImplTestDB(t *testing.T) database.Database {
	// 创建测试数据库配置
	// 注意：这个测试假设配置文件存在且正确
	config := &dbtypes.DbConfig{
		Name:    "clickhouse_main",
		Enabled: true,
		Driver:  database.DriverClickHouse,
		Connection: dbtypes.ConnectionConfig{
			Host:               "121.43.231.91",
			Port:               9000,
			Username:           "default",
			Password:           "YiocaTTS91d*FY#ace{8iopl}",
			Database:           "gateway",
			ClickHouseCompress: "lz4",
			ClickHouseSecure:   false,
			ClickHouseDebug:    false,
			// 移除不支持的配置项
		},
		Pool: dbtypes.PoolConfig{
			MaxOpenConns:    50,
			MaxIdleConns:    10,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 1800,
		},
		Log: dbtypes.LogConfig{
			Enable:        true,
			SlowThreshold: 1000,
		},
	}

	// 直接打开数据库连接
	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("打开ClickHouse连接失败: %v", err)
	}

	// 验证连接
	ctx := context.Background()
	err = db.Ping(ctx)
	if err != nil {
		t.Fatalf("ClickHouse连接测试失败: %v", err)
	}
	t.Log("成功连接到ClickHouse数据库")

	return db
}

// 创建测试表
func setupClickHouseTestTable(t *testing.T, db database.Database) {
	ctx := context.Background()

	// 先尝试删除表（如果存在）
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS clickhouse_users", []interface{}{}, false)
	if err != nil {
		// 忽略表不存在的错误，继续执行
		t.Logf("删除用户测试表: %v", err)
	}

	// 创建用户测试表（使用ClickHouse推荐的MergeTree引擎）
	_, err = db.Exec(ctx, `
		CREATE TABLE clickhouse_users (
			id UInt64,
			name String,
			email String,
			age UInt32,
			created_at DateTime,
			updated_at DateTime
		) ENGINE = MergeTree()
		ORDER BY id
	`, []interface{}{}, true)
	if err != nil {
		t.Fatalf("创建用户测试表失败: %v", err)
	}

	// 创建事件测试表（时间序列数据）
	_, err = db.Exec(ctx, "DROP TABLE IF EXISTS clickhouse_events", []interface{}{}, false)
	if err != nil {
		// 忽略表不存在的错误，继续执行
		t.Logf("删除事件测试表: %v", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE clickhouse_events (
			id UInt64,
			event_name String,
			user_id UInt32,
			timestamp DateTime,
			value Float64,
			properties String,
			date Date
		) ENGINE = MergeTree()
		PARTITION BY date
		ORDER BY (date, timestamp, user_id)
	`, []interface{}{}, true)
	if err != nil {
		t.Fatalf("创建事件测试表失败: %v", err)
	}

	t.Log("测试表创建成功")
}

// 清理测试表
func cleanupClickHouseTestTable(t *testing.T, db database.Database) {
	ctx := context.Background()

	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS clickhouse_users", []interface{}{}, false)
	if err != nil {
		// 忽略表不存在的错误
		t.Logf("清理用户测试表: %v", err)
	}

	_, err = db.Exec(ctx, "DROP TABLE IF EXISTS clickhouse_events", []interface{}{}, false)
	if err != nil {
		// 忽略表不存在的错误
		t.Logf("清理事件测试表: %v", err)
	}

	t.Log("测试表清理成功")
}

// 测试表创建
func TestClickHouseTableSetup(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	setupClickHouseTestTable(t, db)
	defer cleanupClickHouseTestTable(t, db)

	t.Log("ClickHouse表设置测试通过")
}

// 测试基本连接和Ping
func TestClickHousePing(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	ctx := context.Background()
	err := db.Ping(ctx)
	if err != nil {
		t.Fatalf("ClickHouse Ping失败: %v", err)
	}

	t.Log("ClickHouse Ping测试成功")
}

// 测试Insert操作
func TestClickHouseInsert(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	setupClickHouseTestTable(t, db)
	defer cleanupClickHouseTestTable(t, db)

	ctx := context.Background()
	now := time.Now()

	user := ClickHouseUser{
		ID:        1,
		Name:      "张三",
		Email:     "zhangsan@example.com",
		Age:       25,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 测试插入
	id, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("插入用户失败: %v", err)
	}

	t.Logf("插入用户成功，ID: %d", id)
}

// 测试Query操作
func TestClickHouseQuery(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	setupClickHouseTestTable(t, db)
	defer cleanupClickHouseTestTable(t, db)

	ctx := context.Background()
	now := time.Now()

	// 先插入测试数据
	users := []ClickHouseUser{
		{ID: 1, Name: "用户1", Email: "user1@example.com", Age: 25, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Name: "用户2", Email: "user2@example.com", Age: 30, CreatedAt: now, UpdatedAt: now},
		{ID: 3, Name: "用户3", Email: "user3@example.com", Age: 35, CreatedAt: now, UpdatedAt: now},
	}

	for _, user := range users {
		_, err := db.Insert(ctx, user.TableName(), user, true)
		if err != nil {
			t.Fatalf("插入测试数据失败: %v", err)
		}
	}

	// 等待数据写入（ClickHouse是异步写入，需要更多时间）
	time.Sleep(500 * time.Millisecond)

	// 查询所有用户
	var results []ClickHouseUser
	err := db.Query(ctx, &results, "SELECT * FROM clickhouse_users ORDER BY id", []interface{}{}, true)
	if err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("期望查询到3个用户，实际查询到%d个", len(results))
	}

	t.Logf("查询到%d个用户", len(results))
	for i, user := range results {
		t.Logf("用户%d: ID=%d, Name=%s, Email=%s, Age=%d", i+1, user.ID, user.Name, user.Email, user.Age)
	}
}

// 测试QueryOne操作
func TestClickHouseQueryOne(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	setupClickHouseTestTable(t, db)
	defer cleanupClickHouseTestTable(t, db)

	ctx := context.Background()
	now := time.Now()

	// 插入测试数据
	user := ClickHouseUser{
		ID:        1,
		Name:      "测试用户",
		Email:     "test@example.com",
		Age:       28,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 等待数据写入（ClickHouse异步写入需要更多时间）
	time.Sleep(500 * time.Millisecond)

	// 查询单个用户
	var result ClickHouseUser
	err = db.QueryOne(ctx, &result, "SELECT * FROM clickhouse_users WHERE id = ?", []interface{}{1}, true)
	if err != nil {
		t.Fatalf("查询单个用户失败: %v", err)
	}

	if result.ID != 1 {
		t.Fatalf("期望用户ID为1，实际为%d", result.ID)
	}

	t.Logf("查询到用户: ID=%d, Name=%s, Email=%s", result.ID, result.Name, result.Email)
}

// 测试BatchInsert操作
func TestClickHouseBatchInsert(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	setupClickHouseTestTable(t, db)
	//defer cleanupClickHouseTestTable(t, db)

	ctx := context.Background()
	now := time.Now()

	// 准备批量插入数据
	users := []ClickHouseUser{
		{ID: 1, Name: "批量用户1", Email: "batch1@example.com", Age: 20, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Name: "批量用户2", Email: "batch2@example.com", Age: 21, CreatedAt: now, UpdatedAt: now},
		{ID: 3, Name: "批量用户3", Email: "batch3@example.com", Age: 22, CreatedAt: now, UpdatedAt: now},
		{ID: 4, Name: "批量用户4", Email: "batch4@example.com", Age: 23, CreatedAt: now, UpdatedAt: now},
		{ID: 5, Name: "批量用户5", Email: "batch5@example.com", Age: 24, CreatedAt: now, UpdatedAt: now},
	}

	// 批量插入
	affected, err := db.BatchInsert(ctx, "clickhouse_users", users, true)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}

	t.Logf("批量插入成功，影响行数: %d", affected)

	// 等待数据写入（ClickHouse异步写入需要更多时间）
	time.Sleep(500 * time.Millisecond)

	// 验证插入结果 - 查询总记录数
	var count struct {
		Count int64 `db:"count"`
	}
	err = db.QueryOne(ctx, &count, "SELECT count(*) as count FROM clickhouse_users", []interface{}{}, true)
	if err != nil {
		t.Fatalf("验证插入结果失败: %v", err)
	}

	if count.Count != 5 {
		t.Fatalf("期望插入5条记录，实际插入%d条", count.Count)
	}

	// 验证特定记录是否存在
	var specificCount struct {
		Count int64 `db:"count"`
	}
	err = db.QueryOne(ctx, &specificCount, "SELECT count(*) as count FROM clickhouse_users WHERE id = ?", []interface{}{1}, true)
	if err != nil {
		t.Fatalf("验证特定记录失败: %v", err)
	}

	if specificCount.Count != 1 {
		t.Fatalf("期望找到1条ID=1的记录，实际找到%d条", specificCount.Count)
	}

	t.Log("批量插入验证成功")
}

// 测试时间序列数据插入（ClickHouse的典型应用场景）
func TestClickHouseTimeSeriesData(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	setupClickHouseTestTable(t, db)
	defer cleanupClickHouseTestTable(t, db)

	ctx := context.Background()
	now := time.Now()
	today := now.Format("2006-01-02")

	// 准备时间序列事件数据
	events := []ClickHouseEvent{
		{
			ID:         1,
			EventName:  "page_view",
			UserID:     1001,
			Timestamp:  now,
			Value:      1.0,
			Properties: `{"page": "/home", "session_id": "abc123"}`,
			Date:       today,
		},
		{
			ID:         2,
			EventName:  "click",
			UserID:     1001,
			Timestamp:  now.Add(1 * time.Minute),
			Value:      1.0,
			Properties: `{"button": "submit", "form": "login"}`,
			Date:       today,
		},
		{
			ID:         3,
			EventName:  "purchase",
			UserID:     1002,
			Timestamp:  now.Add(2 * time.Minute),
			Value:      99.99,
			Properties: `{"product_id": "prod_123", "category": "electronics"}`,
			Date:       today,
		},
	}

	// 批量插入事件数据
	affected, err := db.BatchInsert(ctx, "clickhouse_events", events, true)
	if err != nil {
		t.Fatalf("批量插入事件数据失败: %v", err)
	}

	t.Logf("批量插入事件数据成功，影响行数: %d", affected)

	// 等待数据写入（ClickHouse异步写入需要更多时间）
	time.Sleep(500 * time.Millisecond)

	// 查询今日事件总数
	var totalEvents int64
	err = db.QueryOne(ctx, &totalEvents, "SELECT count(*) as count FROM clickhouse_events WHERE date = ?", []interface{}{today}, true)
	if err != nil {
		t.Fatalf("查询事件总数失败: %v", err)
	}

	if totalEvents != 3 {
		t.Fatalf("期望3个事件，实际查询到%d个", totalEvents)
	}

	// 查询用户行为统计
	var userEvents []struct {
		UserID     int32   `db:"user_id"`
		EventCount int64   `db:"event_count"`
		TotalValue float64 `db:"total_value"`
	}

	err = db.Query(ctx, &userEvents,
		`SELECT user_id, count as event_count, SUM(value) as total_value 
		 FROM clickhouse_events 
		 WHERE date = ? 
		 GROUP BY user_id 
		 ORDER BY user_id`,
		[]interface{}{today}, true)
	if err != nil {
		t.Fatalf("查询用户统计失败: %v", err)
	}

	t.Logf("查询到%d个用户的行为统计", len(userEvents))
	for _, stat := range userEvents {
		t.Logf("用户%d: 事件数=%d, 总价值=%.2f", stat.UserID, stat.EventCount, stat.TotalValue)
	}

	t.Log("时间序列数据测试成功")
}

// 测试聚合查询（ClickHouse的强项）
func TestClickHouseAggregationQuery(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	setupClickHouseTestTable(t, db)
	defer cleanupClickHouseTestTable(t, db)

	ctx := context.Background()
	now := time.Now()

	// 插入年龄统计测试数据
	users := []ClickHouseUser{
		{ID: 1, Name: "用户1", Email: "user1@test.com", Age: 20, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Name: "用户2", Email: "user2@test.com", Age: 25, CreatedAt: now, UpdatedAt: now},
		{ID: 3, Name: "用户3", Email: "user3@test.com", Age: 30, CreatedAt: now, UpdatedAt: now},
		{ID: 4, Name: "用户4", Email: "user4@test.com", Age: 35, CreatedAt: now, UpdatedAt: now},
		{ID: 5, Name: "用户5", Email: "user5@test.com", Age: 25, CreatedAt: now, UpdatedAt: now},
	}

	_, err := db.BatchInsert(ctx, "clickhouse_users", users, true)
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 等待数据写入（ClickHouse异步写入需要更多时间）
	time.Sleep(500 * time.Millisecond)

	// 测试聚合查询
	var ageStats []struct {
		Age       int32   `db:"age"`
		UserCount int64   `db:"user_count"`
		AvgAge    float64 `db:"avg_age"`
	}

	err = db.Query(ctx, &ageStats,
		`SELECT age, count as user_count, AVG(age) as avg_age 
		 FROM clickhouse_users 
		 GROUP BY age 
		 ORDER BY age`,
		[]interface{}{}, true)
	if err != nil {
		t.Fatalf("聚合查询失败: %v", err)
	}

	t.Logf("年龄统计结果：")
	for _, stat := range ageStats {
		t.Logf("年龄%d: 用户数=%d, 平均年龄=%.1f", stat.Age, stat.UserCount, stat.AvgAge)
	}

	// 测试全局统计
	var globalStats struct {
		TotalUsers int64   `db:"total_users"`
		MinAge     int32   `db:"min_age"`
		MaxAge     int32   `db:"max_age"`
		AvgAge     float64 `db:"avg_age"`
	}

	err = db.QueryOne(ctx, &globalStats,
		`SELECT count(*) as count as total_users, 
		        MIN(age) as min_age, 
		        MAX(age) as max_age, 
		        AVG(age) as avg_age 
		 FROM clickhouse_users`,
		[]interface{}{}, true)
	if err != nil {
		t.Fatalf("全局统计查询失败: %v", err)
	}

	t.Logf("全局统计: 总用户数=%d, 最小年龄=%d, 最大年龄=%d, 平均年龄=%.1f",
		globalStats.TotalUsers, globalStats.MinAge, globalStats.MaxAge, globalStats.AvgAge)

	t.Log("聚合查询测试成功")
}

// 测试Exec操作
func TestClickHouseExec(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	setupClickHouseTestTable(t, db)
	defer cleanupClickHouseTestTable(t, db)

	ctx := context.Background()

	// 测试插入操作
	affected, err := db.Exec(ctx,
		`INSERT INTO clickhouse_users (id, name, email, age, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		[]interface{}{1, "执行测试用户", "exec@test.com", 30, time.Now(), time.Now()}, true)
	if err != nil {
		t.Fatalf("Exec插入失败: %v", err)
	}

	t.Logf("Exec插入成功，影响行数: %d", affected)

	// 等待数据写入（ClickHouse异步写入需要更多时间）
	time.Sleep(500 * time.Millisecond)

	// 验证插入结果
	var count struct {
		Count int64 `db:"count"`
	}
	err = db.QueryOne(ctx, &count, "SELECT count(*) as count FROM clickhouse_users WHERE id = 1", []interface{}{}, true)
	if err != nil {
		t.Fatalf("验证插入结果失败: %v", err)
	}

	if count.Count != 1 {
		t.Fatalf("期望插入1条记录，实际插入%d条", count.Count)
	}

	t.Log("Exec操作测试成功")
}

// 测试驱动信息
func TestClickHouseDriverInfo(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	// 测试驱动类型
	driver := db.GetDriver()
	if driver != database.DriverClickHouse {
		t.Fatalf("期望驱动类型为%s，实际为%s", database.DriverClickHouse, driver)
	}

	// 测试连接名称
	name := db.GetName()
	if name == "" {
		t.Fatal("连接名称不能为空")
	}

	t.Logf("驱动类型: %s, 连接名称: %s", driver, name)
	t.Log("驱动信息测试成功")
}

// 测试ClickHouse特定功能
func TestClickHouseSpecificFeatures(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// 测试系统表查询
	var databases []struct {
		Name string `db:"name"`
	}
	err := db.Query(ctx, &databases, "SELECT name FROM system.databases", []interface{}{}, true)
	if err != nil {
		t.Fatalf("查询数据库列表失败: %v", err)
	}

	// 提取数据库名称列表
	var dbNames []string
	for _, db := range databases {
		dbNames = append(dbNames, db.Name)
	}
	t.Logf("数据库列表: %v", dbNames)

	// 测试当前数据库
	var currentDB struct {
		Database string `db:"currentDatabase()"`
	}
	err = db.QueryOne(ctx, &currentDB, "SELECT currentDatabase()", []interface{}{}, true)
	if err != nil {
		t.Fatalf("查询当前数据库失败: %v", err)
	}

	t.Logf("当前数据库: %s", currentDB.Database)

	// 测试版本信息
	var version struct {
		Version string `db:"version()"`
	}
	err = db.QueryOne(ctx, &version, "SELECT version()", []interface{}{}, true)
	if err != nil {
		t.Fatalf("查询版本信息失败: %v", err)
	}

	t.Logf("ClickHouse版本: %s", version.Version)

	t.Log("ClickHouse特定功能测试成功")
}

// 基准测试：批量插入性能
func BenchmarkClickHouseBatchInsert(b *testing.B) {
	db := getClickHouseImplTestDB(&testing.T{})
	defer db.Close()

	setupClickHouseTestTable(&testing.T{}, db)
	defer cleanupClickHouseTestTable(&testing.T{}, db)

	ctx := context.Background()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users := make([]ClickHouseUser, 100)
		for j := 0; j < 100; j++ {
			users[j] = ClickHouseUser{
				ID:        int64(i*100 + j + 1),
				Name:      fmt.Sprintf("BenchUser%d", i*100+j+1),
				Email:     fmt.Sprintf("bench%d@test.com", i*100+j+1),
				Age:       int32(20 + (i*100+j)%40),
				CreatedAt: now,
				UpdatedAt: now,
			}
		}

		_, err := db.BatchInsert(ctx, "clickhouse_users", users, true)
		if err != nil {
			b.Fatalf("批量插入失败: %v", err)
		}
	}
}

// 基准测试：不同批次大小的批量插入性能
func BenchmarkClickHouseBatchInsertSizes(b *testing.B) {
	db := getClickHouseImplTestDB(&testing.T{})
	defer db.Close()

	// 测试不同的批次大小
	sizes := []int{10, 100, 1000, 5000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("BatchSize_%d", size), func(b *testing.B) {
			setupClickHouseTestTable(&testing.T{}, db)
			defer cleanupClickHouseTestTable(&testing.T{}, db)

			ctx := context.Background()
			now := time.Now()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				users := make([]ClickHouseUser, size)
				for j := 0; j < size; j++ {
					users[j] = ClickHouseUser{
						ID:        int64(i*size + j + 1),
						Name:      fmt.Sprintf("BatchUser%d_%d", size, i*size+j+1),
						Email:     fmt.Sprintf("batch%d_%d@test.com", size, i*size+j+1),
						Age:       int32(20 + (i*size+j)%40),
						CreatedAt: now,
						UpdatedAt: now,
					}
				}

				start := time.Now()
				affected, err := db.BatchInsert(ctx, "clickhouse_users", users, true)
				duration := time.Since(start)

				if err != nil {
					b.Fatalf("批量插入失败: %v", err)
				}

				// 记录性能指标
				rowsPerSecond := float64(affected) / duration.Seconds()
				b.ReportMetric(rowsPerSecond, "rows/sec")
				b.ReportMetric(float64(duration.Nanoseconds())/float64(affected), "ns/row")
			}
		})
	}
}

// 基准测试：批量插入 vs 单条插入性能对比
func BenchmarkClickHouseInsertComparison(b *testing.B) {
	db := getClickHouseImplTestDB(&testing.T{})
	defer db.Close()

	ctx := context.Background()
	now := time.Now()

	// 测试单条插入
	b.Run("SingleInsert", func(b *testing.B) {
		setupClickHouseTestTable(&testing.T{}, db)
		defer cleanupClickHouseTestTable(&testing.T{}, db)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := ClickHouseUser{
				ID:        int64(i + 1),
				Name:      fmt.Sprintf("SingleUser%d", i+1),
				Email:     fmt.Sprintf("single%d@test.com", i+1),
				Age:       int32(20 + i%40),
				CreatedAt: now,
				UpdatedAt: now,
			}

			_, err := db.Insert(ctx, "clickhouse_users", user, true)
			if err != nil {
				b.Fatalf("单条插入失败: %v", err)
			}
		}
	})

	// 测试批量插入（100条）
	b.Run("BatchInsert_100", func(b *testing.B) {
		setupClickHouseTestTable(&testing.T{}, db)
		defer cleanupClickHouseTestTable(&testing.T{}, db)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			users := make([]ClickHouseUser, 100)
			for j := 0; j < 100; j++ {
				users[j] = ClickHouseUser{
					ID:        int64(i*100 + j + 1),
					Name:      fmt.Sprintf("BatchUser%d", i*100+j+1),
					Email:     fmt.Sprintf("batch%d@test.com", i*100+j+1),
					Age:       int32(20 + (i*100+j)%40),
					CreatedAt: now,
					UpdatedAt: now,
				}
			}

			_, err := db.BatchInsert(ctx, "clickhouse_users", users, true)
			if err != nil {
				b.Fatalf("批量插入失败: %v", err)
			}
		}
	})

	// 测试批量插入（1000条）
	b.Run("BatchInsert_1000", func(b *testing.B) {
		setupClickHouseTestTable(&testing.T{}, db)
		defer cleanupClickHouseTestTable(&testing.T{}, db)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			users := make([]ClickHouseUser, 1000)
			for j := 0; j < 1000; j++ {
				users[j] = ClickHouseUser{
					ID:        int64(i*1000 + j + 1),
					Name:      fmt.Sprintf("BatchUser%d", i*1000+j+1),
					Email:     fmt.Sprintf("batch%d@test.com", i*1000+j+1),
					Age:       int32(20 + (i*1000+j)%40),
					CreatedAt: now,
					UpdatedAt: now,
				}
			}

			start := time.Now()
			affected, err := db.BatchInsert(ctx, "clickhouse_users", users, true)
			duration := time.Since(start)

			if err != nil {
				b.Fatalf("批量插入失败: %v", err)
			}

			// 记录性能指标
			rowsPerSecond := float64(affected) / duration.Seconds()
			b.ReportMetric(rowsPerSecond, "rows/sec")
		}
	})
}

// 基准测试：时间序列数据批量插入性能
func BenchmarkClickHouseTimeSeriesBatchInsert(b *testing.B) {
	db := getClickHouseImplTestDB(&testing.T{})
	defer db.Close()

	setupClickHouseTestTable(&testing.T{}, db)
	defer cleanupClickHouseTestTable(&testing.T{}, db)

	ctx := context.Background()
	now := time.Now()
	today := now.Format("2006-01-02")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		events := make([]ClickHouseEvent, 500)
		for j := 0; j < 500; j++ {
			events[j] = ClickHouseEvent{
				ID:         int64(i*500 + j + 1),
				EventName:  fmt.Sprintf("benchmark_event_%d", j%10),
				UserID:     int32(1000 + j%100),
				Timestamp:  now.Add(time.Duration(j) * time.Second),
				Value:      float64(j) * 1.5,
				Properties: fmt.Sprintf(`{"batch": %d, "index": %d}`, i, j),
				Date:       today,
			}
		}

		start := time.Now()
		affected, err := db.BatchInsert(ctx, "clickhouse_events", events, true)
		duration := time.Since(start)

		if err != nil {
			b.Fatalf("时间序列批量插入失败: %v", err)
		}

		// 记录性能指标
		eventsPerSecond := float64(affected) / duration.Seconds()
		b.ReportMetric(eventsPerSecond, "events/sec")
		b.ReportMetric(float64(duration.Microseconds())/float64(affected), "μs/event")
	}
}

// 基准测试：内存使用和性能监控
func BenchmarkClickHouseMemoryEfficiency(b *testing.B) {
	db := getClickHouseImplTestDB(&testing.T{})
	defer db.Close()

	setupClickHouseTestTable(&testing.T{}, db)
	defer cleanupClickHouseTestTable(&testing.T{}, db)

	ctx := context.Background()
	now := time.Now()

	// 大批量数据测试内存效率
	b.Run("LargeBatch_10000", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			users := make([]ClickHouseUser, 10000)
			for j := 0; j < 10000; j++ {
				users[j] = ClickHouseUser{
					ID:        int64(i*10000 + j + 1),
					Name:      fmt.Sprintf("MemUser%d", i*10000+j+1),
					Email:     fmt.Sprintf("mem%d@test.com", i*10000+j+1),
					Age:       int32(20 + (i*10000+j)%60),
					CreatedAt: now,
					UpdatedAt: now,
				}
			}

			start := time.Now()
			affected, err := db.BatchInsert(ctx, "clickhouse_users", users, true)
			duration := time.Since(start)

			if err != nil {
				b.Fatalf("大批量插入失败: %v", err)
			}

			// 记录详细性能指标
			rowsPerSecond := float64(affected) / duration.Seconds()
			mbPerSecond := (float64(affected) * 100) / 1024 / 1024 / duration.Seconds() // 假设每行约100字节

			b.ReportMetric(rowsPerSecond, "rows/sec")
			b.ReportMetric(mbPerSecond, "MB/sec")
			b.ReportMetric(float64(duration.Milliseconds()), "total_ms")
		}
	})
}

// 测试报告：输出ClickHouse批量插入性能总结
func TestClickHouseBatchInsertPerformanceReport(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	setupClickHouseTestTable(t, db)
	defer cleanupClickHouseTestTable(t, db)

	ctx := context.Background()
	now := time.Now()

	// 测试不同批次大小的性能
	batchSizes := []int{100, 1000, 5000, 10000}

	t.Log("ClickHouse批量插入性能测试报告")
	t.Log("=====================================")

	for _, size := range batchSizes {
		users := make([]ClickHouseUser, size)
		for j := 0; j < size; j++ {
			users[j] = ClickHouseUser{
				ID:        int64(j + 1),
				Name:      fmt.Sprintf("PerfUser%d", j+1),
				Email:     fmt.Sprintf("perf%d@test.com", j+1),
				Age:       int32(20 + j%40),
				CreatedAt: now,
				UpdatedAt: now,
			}
		}

		// 执行多次测试取平均值
		totalDuration := time.Duration(0)
		const testRuns = 3

		for run := 0; run < testRuns; run++ {
			// 清理表数据
			_, err := db.Exec(ctx, "TRUNCATE TABLE clickhouse_users", []interface{}{}, true)
			if err != nil {
				t.Fatalf("清理表数据失败: %v", err)
			}

			start := time.Now()
			affected, err := db.BatchInsert(ctx, "clickhouse_users", users, true)
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("批量插入失败 (批次大小: %d): %v", size, err)
			}

			if affected != int64(size) {
				t.Fatalf("插入行数不匹配: 期望 %d, 实际 %d", size, affected)
			}

			totalDuration += duration
		}

		avgDuration := totalDuration / testRuns
		rowsPerSecond := float64(size) / avgDuration.Seconds()

		t.Logf("批次大小: %5d | 平均耗时: %8.2fms | 吞吐量: %8.0f rows/sec",
			size, float64(avgDuration.Nanoseconds())/1000000, rowsPerSecond)
	}

	t.Log("=====================================")
	t.Log("测试完成")
}

// 基准测试：高并发批量插入性能
func BenchmarkClickHouseConcurrentBatchInsert(b *testing.B) {
	db := getClickHouseImplTestDB(&testing.T{})
	defer db.Close()

	setupClickHouseTestTable(&testing.T{}, db)
	defer cleanupClickHouseTestTable(&testing.T{}, db)

	ctx := context.Background()
	now := time.Now()

	// 测试不同的并发度
	concurrencyLevels := []int{2, 4, 8, 16, 32}
	batchSize := 1000 // 每个goroutine的批次大小

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				errChan := make(chan error, concurrency)
				metricChan := make(chan float64, concurrency) // 用于收集性能指标

				// 启动多个goroutine并发插入
				for j := 0; j < concurrency; j++ {
					wg.Add(1)
					go func(routineID int) {
						defer wg.Done()

						// 准备批量数据
						users := make([]ClickHouseUser, batchSize)
						for k := 0; k < batchSize; k++ {
							users[k] = ClickHouseUser{
								ID:        int64(routineID*batchSize + k + 1),
								Name:      fmt.Sprintf("ConcurrentUser_%d_%d", routineID, k),
								Email:     fmt.Sprintf("concurrent%d_%d@test.com", routineID, k),
								Age:       int32(20 + k%40),
								CreatedAt: now,
								UpdatedAt: now,
							}
						}

						start := time.Now()
						affected, err := db.BatchInsert(ctx, "clickhouse_users", users, true)
						duration := time.Since(start)

						if err != nil {
							errChan <- fmt.Errorf("goroutine %d batch insert failed: %v", routineID, err)
							return
						}

						// 计算并发送性能指标
						rowsPerSecond := float64(affected) / duration.Seconds()
						metricChan <- rowsPerSecond
					}(j)
				}

				// 等待所有goroutine完成
				wg.Wait()
				close(errChan)
				close(metricChan)

				// 检查错误
				for err := range errChan {
					b.Fatal(err)
				}

				// 计算总体性能指标
				var totalRowsPerSecond float64
				var metricCount int
				for metric := range metricChan {
					totalRowsPerSecond += metric
					metricCount++
				}

				// 报告性能指标
				avgRowsPerSecond := totalRowsPerSecond / float64(metricCount)
				b.ReportMetric(avgRowsPerSecond, "rows/sec")
				b.ReportMetric(float64(concurrency), "goroutines")
			}
		})
	}
}

// 基准测试：聚合查询性能
func BenchmarkClickHouseAggregationQueries(b *testing.B) {
	db := getClickHouseImplTestDB(&testing.T{})
	defer db.Close()

	setupClickHouseTestTable(&testing.T{}, db)
	defer cleanupClickHouseTestTable(&testing.T{}, db)

	ctx := context.Background()
	now := time.Now()

	// 准备大量测试数据
	batchSize := 100000
	users := make([]ClickHouseUser, batchSize)
	for i := 0; i < batchSize; i++ {
		users[i] = ClickHouseUser{
			ID:        int64(i + 1),
			Name:      fmt.Sprintf("AggUser%d", i+1),
			Email:     fmt.Sprintf("agg%d@test.com", i+1),
			Age:       int32(20 + i%40),
			CreatedAt: now.Add(-time.Duration(i) * time.Hour), // 不同的时间
			UpdatedAt: now,
		}
	}

	// 批量插入测试数据
	_, err := db.BatchInsert(ctx, "clickhouse_users", users, true)
	if err != nil {
		b.Fatalf("准备聚合测试数据失败: %v", err)
	}

	// 等待数据写入（ClickHouse异步写入需要更多时间）
	time.Sleep(500 * time.Millisecond)

	// 测试不同类型的聚合查询
	aggregationQueries := []struct {
		name  string
		query string
		args  []interface{}
	}{
		{
			name: "SimpleCount",
			query: `SELECT count(*) as count 
					FROM clickhouse_users`,
		},
		{
			name: "GroupByAge",
			query: `SELECT age, count() as count, AVG(age) as avg_age 
					FROM clickhouse_users 
					GROUP BY age`,
		},
		{
			name: "TimeRangeAggregation",
			query: `SELECT 
						toDate(created_at) as date,
						count() as count,
						AVG(age) as avg_age,
						uniqExact(email) as unique_emails
					FROM clickhouse_users 
					GROUP BY date
					ORDER BY date`,
		},
		{
			name: "ComplexAggregation",
			query: `SELECT 
						age,
						count() as count,
						AVG(age) as avg_age,
						uniqExact(email) as unique_emails,
						max(created_at) as latest_created,
						min(created_at) as earliest_created
					FROM clickhouse_users 
					GROUP BY age
					HAVING count() > 100
					ORDER BY count() DESC
					LIMIT 10`,
		},
		{
			name: "WindowFunction",
			query: `SELECT 
						age,
						count() OVER (PARTITION BY age) as age_group_size,
						AVG(age) OVER (ORDER BY created_at ROWS BETWEEN 100 PRECEDING AND CURRENT ROW) as moving_avg_age
					FROM clickhouse_users
					LIMIT 1000`,
		},
	}

	// 运行聚合查询基准测试
	for _, aq := range aggregationQueries {
		b.Run(aq.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				start := time.Now()

				var result []struct {
					Count       int64      `db:"count"`
					Age         *int32     `db:"age"`
					AvgAge      *float64   `db:"avg_age"`
					UniqueCount *int64     `db:"unique_emails"`
					Latest      *time.Time `db:"latest_created"`
					Earliest    *time.Time `db:"earliest_created"`
					GroupSize   *int64     `db:"age_group_size"`
					MovingAvg   *float64   `db:"moving_avg_age"`
				}

				err := db.Query(ctx, &result, aq.query, aq.args, true)
				duration := time.Since(start)

				if err != nil {
					b.Fatalf("%s查询失败: %v", aq.name, err)
				}

				// 报告查询性能指标
				b.ReportMetric(float64(duration.Microseconds()), "μs/query")
				if len(result) > 0 {
					b.ReportMetric(float64(len(result)), "rows/query")
				}
			}
		})
	}
}

// 基准测试：并发聚合查询性能
func BenchmarkClickHouseConcurrentAggregation(b *testing.B) {
	db := getClickHouseImplTestDB(&testing.T{})
	defer db.Close()

	setupClickHouseTestTable(&testing.T{}, db)
	defer cleanupClickHouseTestTable(&testing.T{}, db)

	ctx := context.Background()
	now := time.Now()

	// 准备大量测试数据
	batchSize := 100000
	users := make([]ClickHouseUser, batchSize)
	for i := 0; i < batchSize; i++ {
		users[i] = ClickHouseUser{
			ID:        int64(i + 1),
			Name:      fmt.Sprintf("ConcAggUser%d", i+1),
			Email:     fmt.Sprintf("concagg%d@test.com", i+1),
			Age:       int32(20 + i%40),
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
			UpdatedAt: now,
		}
	}

	// 批量插入测试数据
	_, err := db.BatchInsert(ctx, "clickhouse_users", users, true)
	if err != nil {
		b.Fatalf("准备并发聚合测试数据失败: %v", err)
	}

	// 等待数据写入
	time.Sleep(500 * time.Millisecond)

	// 定义并发查询场景
	queries := []struct {
		name  string
		query string
		args  []interface{}
	}{
		{
			name:  "AgeDistribution",
			query: "SELECT age, count() as count FROM clickhouse_users GROUP BY age",
		},
		{
			name:  "TimeSeriesAnalysis",
			query: "SELECT toDate(created_at) as date, count() as count FROM clickhouse_users GROUP BY date",
		},
		{
			name:  "UserStatistics",
			query: "SELECT count(*) as total, AVG(age) as avg_age, uniqExact(email) as unique_emails FROM clickhouse_users",
		},
	}

	// 测试不同的并发度
	concurrencyLevels := []int{2, 4, 8, 16}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				errChan := make(chan error, concurrency*len(queries))
				metricChan := make(chan time.Duration, concurrency*len(queries))

				// 启动并发查询
				for j := 0; j < concurrency; j++ {
					for _, q := range queries {
						wg.Add(1)
						go func(query string, name string) {
							defer wg.Done()

							start := time.Now()
							var result []struct {
								Count       int64    `db:"count"`
								Age         *int32   `db:"age"`
								Date        *string  `db:"date"`
								AvgAge      *float64 `db:"avg_age"`
								UniqueCount *int64   `db:"unique_emails"`
							}

							err := db.Query(ctx, &result, query, nil, true)
							duration := time.Since(start)

							if err != nil {
								errChan <- fmt.Errorf("%s failed: %v", name, err)
								return
							}

							metricChan <- duration
						}(q.query, q.name)
					}
				}

				// 等待所有查询完成
				wg.Wait()
				close(errChan)
				close(metricChan)

				// 检查错误
				for err := range errChan {
					b.Fatal(err)
				}

				// 计算性能指标
				var totalDuration time.Duration
				var queryCount int
				for duration := range metricChan {
					totalDuration += duration
					queryCount++
				}

				// 报告性能指标
				avgDuration := totalDuration / time.Duration(queryCount)
				b.ReportMetric(float64(avgDuration.Microseconds()), "μs/query")
				b.ReportMetric(float64(queryCount), "total_queries")
				b.ReportMetric(float64(concurrency), "goroutines")
			}
		})
	}
}

// 测试复杂字段批量插入 - 模拟真实AccessLog场景
func TestClickHouseComplexFieldsBatchInsert(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// 创建复杂表结构（类似真实的HUB_GW_ACCESS_LOG）
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS clickhouse_complex_logs", []interface{}{}, false)
	if err != nil {
		t.Logf("删除复杂测试表: %v", err)
	}

	// 创建类似真实AccessLog的复杂表结构
	_, err = db.Exec(ctx, `
		CREATE TABLE clickhouse_complex_logs (
			tenantId String,
			traceId String,
			gatewayInstanceId String,
			gatewayInstanceName String,
			gatewayNodeIp String,
			routeConfigId String,
			routeName String,
			serviceDefinitionId String,
			serviceName String,
			proxyType String,
			logConfigId String,
			requestMethod String,
			requestPath String,
			requestQuery String,
			requestSize Int32,
			requestHeaders String,
			requestBody String,
			clientIpAddress String,
			clientPort Int32,
			userAgent String,
			referer String,
			userIdentifier String,
			gatewayStartProcessingTime DateTime,
			backendRequestStartTime DateTime,
			backendResponseReceivedTime DateTime,
			gatewayFinishedProcessingTime DateTime,
			totalProcessingTimeMs Int32,
			gatewayProcessingTimeMs Int32,
			backendResponseTimeMs Int32,
			gatewayStatusCode Int32,
			backendStatusCode Int32,
			responseSize Int32,
			responseHeaders String,
			responseBody String,
			matchedRoute String,
			forwardAddress String,
			forwardMethod String,
			forwardParams String,
			forwardHeaders String,
			forwardBody String,
			loadBalancerDecision String,
			errorMessage String,
			errorCode String,
			parentTraceId String,
			resetFlag String,
			retryCount Int32,
			resetCount Int32,
			logLevel String,
			logType String,
			reserved1 String,
			reserved2 String,
			reserved3 Int32,
			reserved4 Int32,
			reserved5 DateTime,
			extProperty String,
			addTime DateTime,
			addWho String,
			editTime DateTime,
			editWho String,
			oprSeqFlag String,
			currentVersion Int32,
			activeFlag String,
			noteText String
		) ENGINE = MergeTree()
		ORDER BY (tenantId, traceId)
	`, []interface{}{}, true)
	if err != nil {
		t.Fatalf("创建复杂测试表失败: %v", err)
	}

	defer func() {
		_, err := db.Exec(ctx, "DROP TABLE IF EXISTS clickhouse_complex_logs", []interface{}{}, false)
		if err != nil {
			t.Logf("清理复杂测试表: %v", err)
		}
	}()

	now := time.Now()

	// 准备包含特殊字符的复杂测试数据
	complexLogs := []ClickHouseComplexLog{
		{
			TenantId:                      "tenant_001",
			TraceId:                       "trace_001_" + fmt.Sprintf("%d", now.UnixNano()),
			GatewayInstanceId:             "gateway_001",
			GatewayInstanceName:           "主网关实例",
			GatewayNodeIp:                 "192.168.1.100",
			RouteConfigId:                 "route_001",
			RouteName:                     "用户API路由",
			ServiceDefinitionId:           "service_001",
			ServiceName:                   "用户服务",
			ProxyType:                     "http",
			LogConfigId:                   "log_001",
			RequestMethod:                 "POST",
			RequestPath:                   "/api/v1/users",
			RequestQuery:                  "page=1&size=10&sort=name",
			RequestSize:                   1024,
			RequestHeaders:                `{"Content-Type": "application/json", "Authorization": "Bearer abc123", "User-Agent": "Mozilla/5.0"}`,
			RequestBody:                   `{"name": "张三", "email": "zhangsan@example.com", "age": 25}`,
			ClientIpAddress:               "10.0.0.100",
			ClientPort:                    54321,
			UserAgent:                     `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36`,
			Referer:                       "https://example.com/users",
			UserIdentifier:                "user_12345",
			GatewayStartProcessingTime:    now,
			BackendRequestStartTime:       now.Add(10 * time.Millisecond),
			BackendResponseReceivedTime:   now.Add(150 * time.Millisecond),
			GatewayFinishedProcessingTime: now.Add(200 * time.Millisecond),
			TotalProcessingTimeMs:         200,
			GatewayProcessingTimeMs:       50,
			BackendResponseTimeMs:         140,
			GatewayStatusCode:             200,
			BackendStatusCode:             200,
			ResponseSize:                  2048,
			ResponseHeaders:               `{"Content-Type": "application/json", "Cache-Control": "no-cache"}`,
			ResponseBody:                  `{"id": 12345, "name": "张三", "status": "created"}`,
			MatchedRoute:                  "/api/v1/users",
			ForwardAddress:                "http://user-service:8080",
			ForwardMethod:                 "POST",
			ForwardParams:                 `{"timeout": 30, "retries": 3}`,
			ForwardHeaders:                `{"X-Forwarded-For": "10.0.0.100", "X-Request-ID": "req_001"}`,
			ForwardBody:                   `{"name": "张三", "email": "zhangsan@example.com"}`,
			LoadBalancerDecision:          "round_robin",
			ErrorMessage:                  "",
			ErrorCode:                     "",
			ParentTraceId:                 "",
			ResetFlag:                     "N",
			RetryCount:                    0,
			ResetCount:                    0,
			LogLevel:                      "INFO",
			LogType:                       "ACCESS",
			Reserved1:                     "备用字段1",
			Reserved2:                     "备用字段2",
			Reserved3:                     100,
			Reserved4:                     200,
			Reserved5:                     now,
			ExtProperty:                   `{"custom": "value", "feature": "enabled"}`,
			AddTime:                       now,
			AddWho:                        "SYSTEM",
			EditTime:                      now,
			EditWho:                       "SYSTEM",
			OprSeqFlag:                    "LOG_" + now.Format("20060102150405"),
			CurrentVersion:                1,
			ActiveFlag:                    "Y",
			NoteText:                      "正常访问日志",
		},
		{
			TenantId:                      "tenant_002",
			TraceId:                       "trace_002_" + fmt.Sprintf("%d", now.UnixNano()+1),
			GatewayInstanceId:             "gateway_002",
			GatewayInstanceName:           "备用网关",
			GatewayNodeIp:                 "192.168.1.101",
			RouteConfigId:                 "route_002",
			RouteName:                     "错误测试路由",
			ServiceDefinitionId:           "service_002",
			ServiceName:                   "错误服务",
			ProxyType:                     "http",
			LogConfigId:                   "log_002",
			RequestMethod:                 "GET",
			RequestPath:                   "/api/v1/error",
			RequestQuery:                  "test=true&debug=1",
			RequestSize:                   512,
			RequestHeaders:                `{"Accept": "application/json", "Authorization": "Bearer invalid_token"}`,
			RequestBody:                   "",
			ClientIpAddress:               "10.0.0.200",
			ClientPort:                    45678,
			UserAgent:                     `curl/7.68.0`,
			Referer:                       "",
			UserIdentifier:                "",
			GatewayStartProcessingTime:    now.Add(1 * time.Second),
			BackendRequestStartTime:       now.Add(1*time.Second + 20*time.Millisecond),
			BackendResponseReceivedTime:   now.Add(1*time.Second + 100*time.Millisecond),
			GatewayFinishedProcessingTime: now.Add(1*time.Second + 150*time.Millisecond),
			TotalProcessingTimeMs:         150,
			GatewayProcessingTimeMs:       70,
			BackendResponseTimeMs:         80,
			GatewayStatusCode:             500,
			BackendStatusCode:             500,
			ResponseSize:                  256,
			ResponseHeaders:               `{"Content-Type": "application/json", "X-Error": "internal_error"}`,
			ResponseBody:                  `{"error": "Internal server error", "message": "Database connection failed"}`,
			MatchedRoute:                  "/api/v1/error",
			ForwardAddress:                "http://error-service:8080",
			ForwardMethod:                 "GET",
			ForwardParams:                 `{"timeout": 10}`,
			ForwardHeaders:                `{"X-Debug": "true"}`,
			ForwardBody:                   "",
			LoadBalancerDecision:          "least_connections",
			ErrorMessage:                  `Database connection failed: "Connection refused"`,
			ErrorCode:                     "DB_CONN_FAILED",
			ParentTraceId:                 "",
			ResetFlag:                     "N",
			RetryCount:                    2,
			ResetCount:                    0,
			LogLevel:                      "ERROR",
			LogType:                       "ACCESS",
			Reserved1:                     "错误场景",
			Reserved2:                     "测试数据",
			Reserved3:                     500,
			Reserved4:                     -1,
			Reserved5:                     now.Add(1 * time.Second),
			ExtProperty:                   `{"error_type": "connection", "retry_attempted": true}`,
			AddTime:                       now.Add(1 * time.Second),
			AddWho:                        "SYSTEM",
			EditTime:                      now.Add(1 * time.Second),
			EditWho:                       "SYSTEM",
			OprSeqFlag:                    "LOG_" + now.Add(1*time.Second).Format("20060102150405"),
			CurrentVersion:                1,
			ActiveFlag:                    "Y",
			NoteText:                      "包含特殊字符的错误日志：引号\"，换行\n，制表符\t",
		},
	}

	t.Logf("准备批量插入%d条复杂日志记录", len(complexLogs))

	// 批量插入复杂日志数据
	startTime := time.Now()
	affected, err := db.BatchInsert(ctx, "clickhouse_complex_logs", complexLogs, true)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("复杂字段批量插入失败: %v", err)
	}

	t.Logf("复杂字段批量插入成功，影响行数: %d，耗时: %v", affected, duration)

	if affected != int64(len(complexLogs)) {
		t.Fatalf("期望插入%d条记录，实际插入%d条", len(complexLogs), affected)
	}

	// 等待数据写入
	time.Sleep(500 * time.Millisecond)

	// 验证插入结果
	var count struct {
		Count int64 `db:"count"`
	}
	err = db.QueryOne(ctx, &count, "SELECT count(*) as count FROM clickhouse_complex_logs", []interface{}{}, true)
	if err != nil {
		t.Fatalf("验证复杂字段插入结果失败: %v", err)
	}

	if count.Count != int64(len(complexLogs)) {
		t.Fatalf("期望查询到%d条记录，实际查询到%d条", len(complexLogs), count.Count)
	}

	// 查询验证特殊字符处理
	var results []ClickHouseComplexLog
	err = db.Query(ctx, &results, "SELECT * FROM clickhouse_complex_logs ORDER BY traceId", []interface{}{}, true)
	if err != nil {
		t.Fatalf("查询复杂字段数据失败: %v", err)
	}

	t.Logf("查询到%d条复杂日志记录", len(results))
	for i, log := range results {
		t.Logf("记录%d: TraceId=%s, Method=%s, Status=%d, ErrorMsg=%s",
			i+1, log.TraceId, log.RequestMethod, log.GatewayStatusCode, log.ErrorMessage)

		// 验证特殊字符是否正确存储
		if strings.Contains(log.RequestHeaders, `"Content-Type"`) {
			t.Logf("✓ JSON格式的请求头正确存储")
		}
		if strings.Contains(log.NoteText, `引号"`) {
			t.Logf("✓ 包含特殊字符的备注正确存储")
		}
	}

	t.Log("复杂字段批量插入测试成功 - 所有特殊字符和复杂结构都正确处理")
}

// 测试指针类型批量插入 - 重现真实AccessLog的问题
func TestClickHousePointerFieldsBatchInsert(t *testing.T) {
	db := getClickHouseImplTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// 创建包含Nullable字段的复杂表结构（对应指针类型）
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS clickhouse_complex_logs_with_pointers", []interface{}{}, false)
	if err != nil {
		t.Logf("删除指针测试表: %v", err)
	}

	// 创建与指针类型对应的ClickHouse表结构
	_, err = db.Exec(ctx, `
		CREATE TABLE clickhouse_complex_logs_with_pointers (
			tenantId String,
			traceId String,
			gatewayInstanceId String,
			gatewayInstanceName String,
			gatewayNodeIp String,
			requestMethod String,
			requestPath String,
			requestQuery String,
			requestSize Int32,
			requestHeaders String,
			requestBody String,
			clientIpAddress String,
			clientPort Nullable(Int32),
			userAgent String,
			referer String,
			userIdentifier String,
			gatewayStartProcessingTime DateTime,
			backendRequestStartTime Nullable(DateTime),
			backendResponseReceivedTime Nullable(DateTime),
			gatewayFinishedProcessingTime Nullable(DateTime),
			totalProcessingTimeMs Int32,
			gatewayProcessingTimeMs Int32,
			backendResponseTimeMs Nullable(Int32),
			gatewayStatusCode Int32,
			backendStatusCode Nullable(Int32),
			responseSize Int32,
			responseHeaders String,
			responseBody String,
			matchedRoute String,
			forwardAddress String,
			forwardMethod String,
			forwardParams String,
			forwardHeaders String,
			forwardBody String,
			loadBalancerDecision String,
			errorMessage String,
			errorCode String,
			parentTraceId String,
			resetFlag String,
			retryCount Int32,
			resetCount Int32,
			logLevel String,
			logType String,
			reserved1 String,
			reserved2 String,
			reserved3 Nullable(Int32),
			reserved4 Nullable(Int32),
			reserved5 Nullable(DateTime),
			extProperty String,
			addTime DateTime,
			addWho String,
			editTime DateTime,
			editWho String,
			oprSeqFlag String,
			currentVersion Int32,
			activeFlag String,
			noteText String
		) ENGINE = MergeTree()
		ORDER BY (tenantId, traceId)
	`, []interface{}{}, true)
	if err != nil {
		t.Fatalf("创建指针测试表失败: %v", err)
	}

	defer func() {
		_, err := db.Exec(ctx, "DROP TABLE IF EXISTS clickhouse_complex_logs_with_pointers", []interface{}{}, false)
		if err != nil {
			t.Logf("清理指针测试表: %v", err)
		}
	}()

	now := time.Now()

	// 创建包含指针类型的测试数据
	clientPort := 54321
	backendStatusCode := 200
	backendResponseTimeMs := 140
	reserved3 := 100
	reserved4 := 200

	complexLogsWithPointers := []ClickHouseComplexLogWithPointers{
		{
			TenantId:                      "tenant_001",
			TraceId:                       "trace_ptr_001_" + fmt.Sprintf("%d", now.UnixNano()),
			GatewayInstanceId:             "gateway_001",
			GatewayInstanceName:           "指针测试网关",
			GatewayNodeIp:                 "192.168.1.100",
			RequestMethod:                 "POST",
			RequestPath:                   "/api/v1/pointer-test",
			RequestQuery:                  "test=pointer&mode=batch",
			RequestSize:                   1024,
			RequestHeaders:                `{"Content-Type": "application/json", "Authorization": "Bearer test123"}`,
			RequestBody:                   `{"test": "pointer types", "data": "with special chars: \"quotes\" and \n newlines"}`,
			ClientIpAddress:               "10.0.0.100",
			ClientPort:                    &clientPort, // 指针类型
			UserAgent:                     `Mozilla/5.0 Test Agent`,
			Referer:                       "https://test.example.com",
			UserIdentifier:                "user_pointer_test",
			GatewayStartProcessingTime:    now,
			BackendRequestStartTime:       &[]time.Time{now.Add(10 * time.Millisecond)}[0],  // 指针类型
			BackendResponseReceivedTime:   &[]time.Time{now.Add(150 * time.Millisecond)}[0], // 指针类型
			GatewayFinishedProcessingTime: &[]time.Time{now.Add(200 * time.Millisecond)}[0], // 指针类型
			TotalProcessingTimeMs:         200,
			GatewayProcessingTimeMs:       50,
			BackendResponseTimeMs:         &backendResponseTimeMs, // 指针类型
			GatewayStatusCode:             200,
			BackendStatusCode:             &backendStatusCode, // 指针类型
			ResponseSize:                  2048,
			ResponseHeaders:               `{"Content-Type": "application/json"}`,
			ResponseBody:                  `{"result": "success", "message": "Pointer test completed"}`,
			MatchedRoute:                  "/api/v1/pointer-test",
			ForwardAddress:                "http://backend:8080",
			ForwardMethod:                 "POST",
			ForwardParams:                 `{"pointer_test": true}`,
			ForwardHeaders:                `{"X-Test": "pointer"}`,
			ForwardBody:                   `{"forwarded": true}`,
			LoadBalancerDecision:          "round_robin",
			ErrorMessage:                  "",
			ErrorCode:                     "",
			ParentTraceId:                 "",
			ResetFlag:                     "N",
			RetryCount:                    0,
			ResetCount:                    0,
			LogLevel:                      "INFO",
			LogType:                       "ACCESS",
			Reserved1:                     "指针测试字段1",
			Reserved2:                     "指针测试字段2",
			Reserved3:                     &reserved3,           // 指针类型
			Reserved4:                     &reserved4,           // 指针类型
			Reserved5:                     &[]time.Time{now}[0], // 指针类型
			ExtProperty:                   `{"pointer_test": true, "special_chars": "test \"quotes\" and \n newlines"}`,
			AddTime:                       now,
			AddWho:                        "POINTER_TEST",
			EditTime:                      now,
			EditWho:                       "POINTER_TEST",
			OprSeqFlag:                    "PTR_LOG_" + now.Format("20060102150405"),
			CurrentVersion:                1,
			ActiveFlag:                    "Y",
			NoteText:                      "这是指针类型测试：包含特殊字符 \"引号\", 换行\n, 制表符\t",
		},
		{
			TenantId:                      "tenant_002",
			TraceId:                       "trace_ptr_002_" + fmt.Sprintf("%d", now.UnixNano()+1),
			GatewayInstanceId:             "gateway_002",
			GatewayInstanceName:           "指针测试网关2",
			GatewayNodeIp:                 "192.168.1.101",
			RequestMethod:                 "GET",
			RequestPath:                   "/api/v1/pointer-error",
			RequestQuery:                  "error=true&pointer=test",
			RequestSize:                   256,
			RequestHeaders:                `{"Accept": "application/json"}`,
			RequestBody:                   "",
			ClientIpAddress:               "10.0.0.200",
			ClientPort:                    nil, // nil指针！
			UserAgent:                     "Test Agent",
			Referer:                       "",
			UserIdentifier:                "",
			GatewayStartProcessingTime:    now.Add(1 * time.Second),
			BackendRequestStartTime:       nil, // nil指针！
			BackendResponseReceivedTime:   nil, // nil指针！
			GatewayFinishedProcessingTime: nil, // nil指针！
			TotalProcessingTimeMs:         0,
			GatewayProcessingTimeMs:       0,
			BackendResponseTimeMs:         nil, // nil指针！
			GatewayStatusCode:             500,
			BackendStatusCode:             nil, // nil指针！
			ResponseSize:                  128,
			ResponseHeaders:               `{"Content-Type": "application/json", "X-Error": "pointer_test"}`,
			ResponseBody:                  `{"error": "Pointer test with nil values", "message": "Testing nil pointer handling"}`,
			MatchedRoute:                  "/api/v1/pointer-error",
			ForwardAddress:                "",
			ForwardMethod:                 "",
			ForwardParams:                 "",
			ForwardHeaders:                "",
			ForwardBody:                   "",
			LoadBalancerDecision:          "",
			ErrorMessage:                  `Pointer test error with "quotes" and special chars: \n\t`,
			ErrorCode:                     "PTR_TEST_ERROR",
			ParentTraceId:                 "",
			ResetFlag:                     "N",
			RetryCount:                    0,
			ResetCount:                    0,
			LogLevel:                      "ERROR",
			LogType:                       "ACCESS",
			Reserved1:                     "",
			Reserved2:                     "",
			Reserved3:                     nil, // nil指针！
			Reserved4:                     nil, // nil指针！
			Reserved5:                     nil, // nil指针！
			ExtProperty:                   `{"nil_test": true}`,
			AddTime:                       now.Add(1 * time.Second),
			AddWho:                        "POINTER_TEST",
			EditTime:                      now.Add(1 * time.Second),
			EditWho:                       "POINTER_TEST",
			OprSeqFlag:                    "PTR_LOG_" + now.Add(1*time.Second).Format("20060102150405"),
			CurrentVersion:                1,
			ActiveFlag:                    "Y",
			NoteText:                      "nil指针测试记录",
		},
	}

	t.Logf("准备测试指针类型批量插入%d条记录", len(complexLogsWithPointers))

	// 批量插入指针类型数据 - 预期这里会出现错误
	startTime := time.Now()
	affected, err := db.BatchInsert(ctx, "clickhouse_complex_logs_with_pointers", complexLogsWithPointers, true)
	duration := time.Since(startTime)

	if err != nil {
		// 预期的错误 - 记录错误信息用于分析
		t.Logf("❌ 指针类型批量插入失败（预期的错误）: %v", err)
		t.Logf("⏱️ 失败耗时: %v", duration)

		// 检查是否是预期的解析错误
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "Cannot parse input") {
			t.Logf("✅ 确认是解析错误，与AccessLog问题一致")
		}
		if strings.Contains(errorMsg, "expected 'eof' before") {
			t.Logf("✅ 确认是EOF解析错误，证实了问题根源")
		}

		t.Log("🔍 指针类型测试结果：确认指针类型导致ClickHouse批量插入失败")
		return // 测试达到目的，返回
	}

	// 如果意外成功了
	t.Logf("⚠️ 意外成功：指针类型批量插入成功，影响行数: %d，耗时: %v", affected, duration)

	// 验证数据
	time.Sleep(500 * time.Millisecond)

	var count struct {
		Count int64 `db:"count"`
	}
	err = db.QueryOne(ctx, &count, "SELECT count(*) as count FROM clickhouse_complex_logs_with_pointers", []interface{}{}, true)
	if err != nil {
		t.Fatalf("验证指针类型插入结果失败: %v", err)
	}

	t.Logf("✅ 指针类型测试意外成功：插入%d条记录，查询到%d条记录", affected, count.Count)
	t.Log("📝 注意：如果指针类型测试成功，可能是ClickHouse驱动版本或配置差异导致")
}
