package loader

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"gohub/internal/gateway/config"
)

// DatabaseConfigLoader 数据库配置加载器
type DatabaseConfigLoader struct {
	factory *GatewayConfigFactory
	db      *sql.DB
	options *DatabaseOptions
}

// DatabaseOptions 数据库连接选项
type DatabaseOptions struct {
	Driver          string        `yaml:"driver" json:"driver"`                       // 数据库驱动 (mysql, postgresql, sqlite3)
	DSN             string        `yaml:"dsn" json:"dsn"`                             // 数据源名称
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`       // 最大打开连接数
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`       // 最大空闲连接数
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"` // 连接最大生存时间
	TableName       string        `yaml:"table_name" json:"table_name"`               // 配置表名
}

// GatewayConfigRecord 数据库中的配置记录
type GatewayConfigRecord struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	ConfigData  string    `db:"config_data" json:"config_data"` // JSON格式的配置数据
	Version     int       `db:"version" json:"version"`         // 配置版本
	Environment string    `db:"environment" json:"environment"` // 环境标识 (dev, test, prod)
	IsActive    bool      `db:"is_active" json:"is_active"`     // 是否激活
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	CreatedBy   string    `db:"created_by" json:"created_by"`
	UpdatedBy   string    `db:"updated_by" json:"updated_by"`
	Description string    `db:"description" json:"description"`
}

// NewDatabaseConfigLoader 创建数据库配置加载器
func NewDatabaseConfigLoader(options *DatabaseOptions) (*DatabaseConfigLoader, error) {
	if options == nil {
		options = &DatabaseOptions{
			Driver:          "sqlite3",
			DSN:             "./gateway_configs.db",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Hour,
			TableName:       "gateway_configs",
		}
	}

	loader := &DatabaseConfigLoader{
		factory: NewGatewayConfigFactory(ConfigSourceDB),
		options: options,
	}

	// 初始化数据库连接
	if err := loader.initDatabase(); err != nil {
		return nil, fmt.Errorf("初始化数据库连接失败: %w", err)
	}

	return loader, nil
}

// initDatabase 初始化数据库连接
func (d *DatabaseConfigLoader) initDatabase() error {
	var err error
	d.db, err = sql.Open(d.options.Driver, d.options.DSN)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 设置连接池参数
	d.db.SetMaxOpenConns(d.options.MaxOpenConns)
	d.db.SetMaxIdleConns(d.options.MaxIdleConns)
	d.db.SetConnMaxLifetime(d.options.ConnMaxLifetime)

	// 测试连接
	if err := d.db.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 创建配置表（如果不存在）
	if err := d.createConfigTable(); err != nil {
		return fmt.Errorf("创建配置表失败: %w", err)
	}

	return nil
}

// createConfigTable 创建配置表
func (d *DatabaseConfigLoader) createConfigTable() error {
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			config_data TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			environment VARCHAR(50) NOT NULL DEFAULT 'default',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(255) DEFAULT 'system',
			updated_by VARCHAR(255) DEFAULT 'system',
			description TEXT
		)
	`, d.options.TableName)

	_, err := d.db.Exec(createTableSQL)
	return err
}

// LoadConfig 从数据库加载配置
func (d *DatabaseConfigLoader) LoadConfig(configID string) (*config.GatewayConfig, error) {
	if configID == "" {
		return &config.DefaultGatewayConfig, nil
	}

	record, err := d.getConfigRecord(configID)
	if err != nil {
		return nil, fmt.Errorf("获取数据库配置记录失败: %w", err)
	}

	if record == nil {
		return &config.DefaultGatewayConfig, nil
	}

	// 解析JSON配置数据
	cfg := &config.GatewayConfig{}
	if err := json.Unmarshal([]byte(record.ConfigData), cfg); err != nil {
		return nil, fmt.Errorf("解析数据库配置数据失败: %w", err)
	}

	// 合并默认配置
	d.factory.mergeDefaultConfig(cfg)

	return cfg, nil
}

// getConfigRecord 获取配置记录
func (d *DatabaseConfigLoader) getConfigRecord(configID string) (*GatewayConfigRecord, error) {
	querySQL := fmt.Sprintf(`
		SELECT id, name, config_data, version, environment, is_active, 
		       created_at, updated_at, created_by, updated_by, description
		FROM %s 
		WHERE id = ? AND is_active = TRUE
		ORDER BY version DESC
		LIMIT 1
	`, d.options.TableName)

	row := d.db.QueryRow(querySQL, configID)

	record := &GatewayConfigRecord{}
	err := row.Scan(
		&record.ID,
		&record.Name,
		&record.ConfigData,
		&record.Version,
		&record.Environment,
		&record.IsActive,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedBy,
		&record.UpdatedBy,
		&record.Description,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("查询配置记录失败: %w", err)
	}

	return record, nil
}

// SaveConfig 保存配置到数据库
func (d *DatabaseConfigLoader) SaveConfig(record *GatewayConfigRecord) error {
	// 检查配置是否已存在
	existing, err := d.getConfigRecord(record.ID)
	if err != nil {
		return fmt.Errorf("检查配置记录失败: %w", err)
	}

	if existing != nil {
		// 更新现有配置
		return d.updateConfig(record)
	} else {
		// 插入新配置
		return d.insertConfig(record)
	}
}

// insertConfig 插入新配置
func (d *DatabaseConfigLoader) insertConfig(record *GatewayConfigRecord) error {
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (
			id, name, config_data, version, environment, is_active,
			created_at, updated_at, created_by, updated_by, description
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, d.options.TableName)

	now := time.Now()
	_, err := d.db.Exec(
		insertSQL,
		record.ID,
		record.Name,
		record.ConfigData,
		record.Version,
		record.Environment,
		record.IsActive,
		now,
		now,
		record.CreatedBy,
		record.UpdatedBy,
		record.Description,
	)

	return err
}

// updateConfig 更新配置
func (d *DatabaseConfigLoader) updateConfig(record *GatewayConfigRecord) error {
	updateSQL := fmt.Sprintf(`
		UPDATE %s SET 
			name = ?, config_data = ?, version = version + 1, 
			environment = ?, is_active = ?, updated_at = ?, 
			updated_by = ?, description = ?
		WHERE id = ?
	`, d.options.TableName)

	_, err := d.db.Exec(
		updateSQL,
		record.Name,
		record.ConfigData,
		record.Environment,
		record.IsActive,
		time.Now(),
		record.UpdatedBy,
		record.Description,
		record.ID,
	)

	return err
}

// ListConfigs 列出所有配置
func (d *DatabaseConfigLoader) ListConfigs(environment string) ([]*GatewayConfigRecord, error) {
	var querySQL string
	var args []interface{}

	if environment != "" {
		querySQL = fmt.Sprintf(`
			SELECT id, name, config_data, version, environment, is_active,
			       created_at, updated_at, created_by, updated_by, description
			FROM %s 
			WHERE environment = ? AND is_active = TRUE
			ORDER BY updated_at DESC
		`, d.options.TableName)
		args = []interface{}{environment}
	} else {
		querySQL = fmt.Sprintf(`
			SELECT id, name, config_data, version, environment, is_active,
			       created_at, updated_at, created_by, updated_by, description
			FROM %s 
			WHERE is_active = TRUE
			ORDER BY updated_at DESC
		`, d.options.TableName)
	}

	rows, err := d.db.Query(querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("查询配置列表失败: %w", err)
	}
	defer rows.Close()

	var records []*GatewayConfigRecord
	for rows.Next() {
		record := &GatewayConfigRecord{}
		err := rows.Scan(
			&record.ID,
			&record.Name,
			&record.ConfigData,
			&record.Version,
			&record.Environment,
			&record.IsActive,
			&record.CreatedAt,
			&record.UpdatedAt,
			&record.CreatedBy,
			&record.UpdatedBy,
			&record.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描配置记录失败: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// DeleteConfig 删除配置（软删除）
func (d *DatabaseConfigLoader) DeleteConfig(configID string, deletedBy string) error {
	updateSQL := fmt.Sprintf(`
		UPDATE %s SET 
			is_active = FALSE, updated_at = ?, updated_by = ?
		WHERE id = ?
	`, d.options.TableName)

	_, err := d.db.Exec(updateSQL, time.Now(), deletedBy, configID)
	return err
}

// ValidateConfig 验证数据库配置
func (d *DatabaseConfigLoader) ValidateConfig(configID string) error {
	if configID == "" {
		return fmt.Errorf("配置ID不能为空")
	}

	record, err := d.getConfigRecord(configID)
	if err != nil {
		return fmt.Errorf("获取数据库配置失败: %w", err)
	}

	if record == nil {
		return fmt.Errorf("配置ID %s 不存在", configID)
	}

	// 尝试解析配置数据
	cfg := &config.GatewayConfig{}
	if err := json.Unmarshal([]byte(record.ConfigData), cfg); err != nil {
		return fmt.Errorf("配置数据格式验证失败: %w", err)
	}

	return nil
}

// ExportConfigToJSON 将配置导出为JSON格式
func (d *DatabaseConfigLoader) ExportConfigToJSON(configID string, indent bool) (string, error) {
	cfg, err := d.LoadConfig(configID)
	if err != nil {
		return "", fmt.Errorf("加载配置失败: %w", err)
	}

	var data []byte
	if indent {
		data, err = json.MarshalIndent(cfg, "", "  ")
	} else {
		data, err = json.Marshal(cfg)
	}

	if err != nil {
		return "", fmt.Errorf("导出JSON配置失败: %w", err)
	}

	return string(data), nil
}

// ImportConfigFromJSON 从JSON导入配置
func (d *DatabaseConfigLoader) ImportConfigFromJSON(configID, name, jsonData, environment, operator string) error {
	// 验证JSON格式
	cfg := &config.GatewayConfig{}
	if err := json.Unmarshal([]byte(jsonData), cfg); err != nil {
		return fmt.Errorf("JSON格式验证失败: %w", err)
	}

	record := &GatewayConfigRecord{
		ID:          configID,
		Name:        name,
		ConfigData:  jsonData,
		Version:     1,
		Environment: environment,
		IsActive:    true,
		CreatedBy:   operator,
		UpdatedBy:   operator,
		Description: fmt.Sprintf("从JSON导入的配置 - %s", time.Now().Format("2006-01-02 15:04:05")),
	}

	return d.SaveConfig(record)
}

// Close 关闭数据库连接
func (d *DatabaseConfigLoader) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// GetDatabase 获取数据库连接
func (d *DatabaseConfigLoader) GetDatabase() *sql.DB {
	return d.db
}

// GetDatabaseOptions 获取数据库选项
func (d *DatabaseConfigLoader) GetDatabaseOptions() *DatabaseOptions {
	return d.options
}

// ReloadConfig 重新加载数据库配置
func (d *DatabaseConfigLoader) ReloadConfig(configID string) (*config.GatewayConfig, error) {
	return d.LoadConfig(configID)
}
