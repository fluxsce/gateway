package store

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/database"
	"gateway/pkg/utils/random"
)

// CenterStore 访问 HUB_SERVICE 中心实例定义表。
type CenterStore struct {
	db database.Database
}

// centerRow 映射 HUB_SERVICE 中心实例行，不对外暴露。
type centerRow struct {
	TenantID              string    `db:"tenantId"`
	InstanceName          string    `db:"instanceName"`
	Environment           string    `db:"environment"`
	ServerType            string    `db:"serverType"`
	ListenAddress         string    `db:"listenAddress"`
	ListenPort            int       `db:"listenPort"`
	MaxRecvMsgSize        int       `db:"maxRecvMsgSize"`
	MaxSendMsgSize        int       `db:"maxSendMsgSize"`
	KeepAliveTime         int       `db:"keepAliveTime"`
	KeepAliveTimeout      int       `db:"keepAliveTimeout"`
	KeepAliveMinTime      int       `db:"keepAliveMinTime"`
	PermitWithoutStream   string    `db:"permitWithoutStream"`
	MaxConnectionIdle     int       `db:"maxConnectionIdle"`
	MaxConnectionAge      int       `db:"maxConnectionAge"`
	MaxConnectionAgeGrace int       `db:"maxConnectionAgeGrace"`
	EnableReflection      string    `db:"enableReflection"`
	EnableTLS             string    `db:"enableTLS"`
	CertStorageType       string    `db:"certStorageType"`
	CertFilePath          string    `db:"certFilePath"`
	KeyFilePath           string    `db:"keyFilePath"`
	CertContent           string    `db:"certContent"`
	KeyContent            string    `db:"keyContent"`
	CertChainContent      string    `db:"certChainContent"`
	CertPassword          string    `db:"certPassword"`
	EnableMTLS            string    `db:"enableMTLS"`
	MaxConcurrentStreams  int       `db:"maxConcurrentStreams"`
	ReadBufferSize        int       `db:"readBufferSize"`
	WriteBufferSize       int       `db:"writeBufferSize"`
	HealthCheckInterval   int       `db:"healthCheckInterval"`
	HealthCheckTimeout    int       `db:"healthCheckTimeout"`
	InstanceStatus        string    `db:"instanceStatus"`
	StatusMessage         string    `db:"statusMessage"`
	EnableAuth            string    `db:"enableAuth"`
	IPWhitelist           string    `db:"ipWhitelist"`
	IPBlacklist           string    `db:"ipBlacklist"`
	ExtProperty           string    `db:"extProperty"`
	AddTime               time.Time `db:"addTime"`
	AddWho                string    `db:"addWho"`
	EditTime              time.Time `db:"editTime"`
	EditWho               string    `db:"editWho"`
	OprSeqFlag            string    `db:"oprSeqFlag"`
	CurrentVersion        int       `db:"currentVersion"`
	ActiveFlag            string    `db:"activeFlag"`
}

// Get 按租户、实例名、环境读取定义，不存在时返回 nil, nil。
// environment 为空时：仅一行则返回该行；多行则报错，避免 DEV/PROD 同名互相覆盖。
func (s *CenterStore) Get(ctx context.Context, tenantID, instanceName, environment string) (*model.CenterInstance, error) {
	if environment != "" {
		query := `SELECT * FROM HUB_SERVICE_INSTANCE WHERE tenantId = ? AND instanceName = ? AND environment = ? AND activeFlag = 'Y'`
		var row centerRow
		err := s.db.QueryOne(ctx, &row, query, []interface{}{tenantID, instanceName, environment}, true)
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("查询中心实例失败: %w", err)
		}
		return row.toModel(), nil
	}
	query := `SELECT * FROM HUB_SERVICE_INSTANCE WHERE tenantId = ? AND instanceName = ? AND activeFlag = 'Y' ORDER BY environment`
	var rows []*centerRow
	if err := s.db.Query(ctx, &rows, query, []interface{}{tenantID, instanceName}, true); err != nil {
		return nil, fmt.Errorf("查询中心实例失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	if len(rows) > 1 {
		return nil, fmt.Errorf("实例 %s 存在多个环境，必须指定 environment", instanceName)
	}
	return rows[0].toModel(), nil
}

// ListByTenant 列出租户下已启用的中心实例定义。
func (s *CenterStore) ListByTenant(ctx context.Context, tenantID string) ([]*model.CenterInstance, error) {
	query := `SELECT * FROM HUB_SERVICE_INSTANCE WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY environment, instanceName`
	var rows []*centerRow
	if err := s.db.Query(ctx, &rows, query, []interface{}{tenantID}, true); err != nil {
		return nil, fmt.Errorf("查询中心实例列表失败: %w", err)
	}
	return rowsToCenters(rows), nil
}

// ListActive 列出全部已启用的中心实例，不按租户过滤。
func (s *CenterStore) ListActive(ctx context.Context) ([]*model.CenterInstance, error) {
	query := `SELECT * FROM HUB_SERVICE_INSTANCE WHERE activeFlag = 'Y' ORDER BY tenantId, environment, instanceName`
	var rows []*centerRow
	if err := s.db.Query(ctx, &rows, query, nil, true); err != nil {
		return nil, fmt.Errorf("查询中心实例列表失败: %w", err)
	}
	return rowsToCenters(rows), nil
}

func rowsToCenters(rows []*centerRow) []*model.CenterInstance {
	out := make([]*model.CenterInstance, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toModel())
	}
	return out
}

// Create 插入中心实例定义。
func (s *CenterStore) Create(ctx context.Context, inst *model.CenterInstance, operator string) error {
	now := time.Now()
	row := fromCenterModel(inst)
	row.AddTime = now
	row.EditTime = now
	row.AddWho = operator
	row.EditWho = operator
	row.OprSeqFlag = random.Generate32BitRandomString()
	row.CurrentVersion = 1
	if row.ActiveFlag == "" {
		row.ActiveFlag = "Y"
	}
	if row.ServerType == "" {
		row.ServerType = "GRPC"
	}
	_, err := s.db.Insert(ctx, "HUB_SERVICE_INSTANCE", row, true)
	if err != nil {
		return fmt.Errorf("创建中心实例失败: %w", err)
	}
	return nil
}

// Update 更新中心实例定义。
func (s *CenterStore) Update(ctx context.Context, inst *model.CenterInstance, operator string) error {
	row := fromCenterModel(inst)
	row.EditTime = time.Now()
	row.EditWho = operator
	row.OprSeqFlag = random.Generate32BitRandomString()
	where := "tenantId = ? AND instanceName = ? AND environment = ?"
	_, err := s.db.Update(ctx, "HUB_SERVICE_INSTANCE", row, where,
		[]interface{}{inst.TenantID, inst.InstanceName, inst.Environment}, true, true)
	if err != nil {
		return fmt.Errorf("更新中心实例失败: %w", err)
	}
	return nil
}

// Delete 按联合键删除定义。
func (s *CenterStore) Delete(ctx context.Context, tenantID, instanceName, environment string) error {
	_, err := s.db.Delete(ctx, "HUB_SERVICE_INSTANCE",
		"tenantId = ? AND instanceName = ? AND environment = ?",
		[]interface{}{tenantID, instanceName, environment}, true)
	if err != nil {
		return fmt.Errorf("删除中心实例失败: %w", err)
	}
	return nil
}

// UpdateStatus 只更新运行状态与说明。
// 进程退出时调用方应跳过。这一行按租户、实例名、环境共用，没有 Pod 身份，退出回写会盖住正在滚动启动的副本。
func (s *CenterStore) UpdateStatus(ctx context.Context, tenantID, instanceName, environment, status, message string) error {
	query := `UPDATE HUB_SERVICE_INSTANCE SET instanceStatus = ?, statusMessage = ?, lastStatusTime = ?
		WHERE tenantId = ? AND instanceName = ? AND environment = ?`
	_, err := s.db.Exec(ctx, query, []interface{}{status, message, time.Now(), tenantID, instanceName, environment}, true)
	if err != nil {
		return fmt.Errorf("更新中心实例状态失败: %w", err)
	}
	return nil
}

// toModel 把库表行转为领域对象。
func (r *centerRow) toModel() *model.CenterInstance {
	return &model.CenterInstance{
		TenantID:              r.TenantID,
		InstanceName:          r.InstanceName,
		Environment:           r.Environment,
		ListenAddress:         r.ListenAddress,
		ListenPort:            r.ListenPort,
		MaxRecvMsgSize:        r.MaxRecvMsgSize,
		MaxSendMsgSize:        r.MaxSendMsgSize,
		KeepAliveTime:         r.KeepAliveTime,
		KeepAliveTimeout:      r.KeepAliveTimeout,
		KeepAliveMinTime:      r.KeepAliveMinTime,
		PermitWithoutStream:   r.PermitWithoutStream,
		MaxConnectionIdle:     r.MaxConnectionIdle,
		MaxConnectionAge:      r.MaxConnectionAge,
		MaxConnectionAgeGrace: r.MaxConnectionAgeGrace,
		EnableReflection:      r.EnableReflection,
		EnableTLS:             r.EnableTLS,
		CertStorageType:       r.CertStorageType,
		CertFilePath:          r.CertFilePath,
		KeyFilePath:           r.KeyFilePath,
		CertContent:           r.CertContent,
		KeyContent:            r.KeyContent,
		CertChainContent:      r.CertChainContent,
		CertPassword:          r.CertPassword,
		EnableMTLS:            r.EnableMTLS,
		MaxConcurrentStreams:  r.MaxConcurrentStreams,
		ReadBufferSize:        r.ReadBufferSize,
		WriteBufferSize:       r.WriteBufferSize,
		HealthCheckInterval:   r.HealthCheckInterval,
		HealthCheckTimeout:    r.HealthCheckTimeout,
		EnableAuth:            r.EnableAuth,
		IPWhitelist:           r.IPWhitelist,
		IPBlacklist:           r.IPBlacklist,
		ExtProperty:           r.ExtProperty,
		Status:                r.InstanceStatus,
		StatusMessage:         r.StatusMessage,
		ActiveFlag:            r.ActiveFlag,
	}
}

// fromCenterModel 把领域对象转为库表行。
func fromCenterModel(inst *model.CenterInstance) *centerRow {
	return &centerRow{
		TenantID:              inst.TenantID,
		InstanceName:          inst.InstanceName,
		Environment:           inst.Environment,
		ServerType:            "GRPC",
		ListenAddress:         inst.ListenAddress,
		ListenPort:            inst.ListenPort,
		MaxRecvMsgSize:        inst.MaxRecvMsgSize,
		MaxSendMsgSize:        inst.MaxSendMsgSize,
		KeepAliveTime:         inst.KeepAliveTime,
		KeepAliveTimeout:      inst.KeepAliveTimeout,
		KeepAliveMinTime:      inst.KeepAliveMinTime,
		PermitWithoutStream:   inst.PermitWithoutStream,
		MaxConnectionIdle:     inst.MaxConnectionIdle,
		MaxConnectionAge:      inst.MaxConnectionAge,
		MaxConnectionAgeGrace: inst.MaxConnectionAgeGrace,
		EnableReflection:      inst.EnableReflection,
		EnableTLS:             inst.EnableTLS,
		CertStorageType:       inst.CertStorageType,
		CertFilePath:          inst.CertFilePath,
		KeyFilePath:           inst.KeyFilePath,
		CertContent:           inst.CertContent,
		KeyContent:            inst.KeyContent,
		CertChainContent:      inst.CertChainContent,
		CertPassword:          inst.CertPassword,
		EnableMTLS:            inst.EnableMTLS,
		MaxConcurrentStreams:  inst.MaxConcurrentStreams,
		ReadBufferSize:        inst.ReadBufferSize,
		WriteBufferSize:       inst.WriteBufferSize,
		HealthCheckInterval:   inst.HealthCheckInterval,
		HealthCheckTimeout:    inst.HealthCheckTimeout,
		InstanceStatus:        inst.Status,
		StatusMessage:         inst.StatusMessage,
		EnableAuth:            inst.EnableAuth,
		IPWhitelist:           inst.IPWhitelist,
		IPBlacklist:           inst.IPBlacklist,
		ExtProperty:           inst.ExtProperty,
		ActiveFlag:            inst.ActiveFlag,
	}
}
