package store

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/database"
	"gateway/pkg/utils/random"
)

// ConfigStore 访问已发布配置及历史表。
type ConfigStore struct {
	db database.Database
}

// configRow 映射当前已发布配置行。
type configRow struct {
	ConfigDataId      string    `db:"configDataId"`
	TenantId          string    `db:"tenantId"`
	NamespaceId       string    `db:"namespaceId"`
	GroupName         string    `db:"groupName"`
	ConfigContent     string    `db:"configContent"`
	ContentType       string    `db:"contentType"`
	ConfigDescription string    `db:"configDescription"`
	Encrypted         string    `db:"encrypted"`
	Version           int64     `db:"version"`
	Md5Value          string    `db:"md5Value"`
	AddTime           time.Time `db:"addTime"`
	AddWho            string    `db:"addWho"`
	EditTime          time.Time `db:"editTime"`
	EditWho           string    `db:"editWho"`
	OprSeqFlag        string    `db:"oprSeqFlag"`
	CurrentVersion    int       `db:"currentVersion"`
	ActiveFlag        string    `db:"activeFlag"`
}

// historyRow 映射配置发布历史行。
type historyRow struct {
	ConfigHistoryId string    `db:"configHistoryId"`
	TenantId        string    `db:"tenantId"`
	ConfigDataId    string    `db:"configDataId"`
	NamespaceId     string    `db:"namespaceId"`
	GroupName       string    `db:"groupName"`
	ChangeType      string    `db:"changeType"`
	OldContent      string    `db:"oldContent"`
	NewContent      string    `db:"newContent"`
	OldVersion      int64     `db:"oldVersion"`
	NewVersion      int64     `db:"newVersion"`
	OldMd5Value     string    `db:"oldMd5Value"`
	NewMd5Value     string    `db:"newMd5Value"`
	ChangeReason    string    `db:"changeReason"`
	ChangedBy       string    `db:"changedBy"`
	ChangedAt       time.Time `db:"changedAt"`
	AddTime         time.Time `db:"addTime"`
	AddWho          string    `db:"addWho"`
	EditTime        time.Time `db:"editTime"`
	EditWho         string    `db:"editWho"`
	OprSeqFlag      string    `db:"oprSeqFlag"`
	CurrentVersion  int       `db:"currentVersion"`
	ActiveFlag      string    `db:"activeFlag"`
	NoteText        string    `db:"noteText"`
}

const historyChangeDraft = "DRAFT"

// Get 读取当前已发布配置，不存在时返回 nil, nil。
func (s *ConfigStore) Get(ctx context.Context, tenantID, namespaceID, group, dataID string) (*model.ConfigRelease, error) {
	query := `SELECT * FROM HUB_SERVICE_CONFIG_DATA WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ?`
	var row configRow
	err := s.db.QueryOne(ctx, &row, query, []interface{}{tenantID, namespaceID, group, dataID}, true)
	if err == database.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}
	return row.toRelease(), nil
}

// Count 返回命名空间下已发布配置条数。
func (s *ConfigStore) Count(ctx context.Context, tenantID, namespaceID string) (int, error) {
	query := `SELECT COUNT(*) AS n FROM HUB_SERVICE_CONFIG_DATA
		WHERE tenantId = ? AND namespaceId = ? AND activeFlag = 'Y'`
	var row countRow
	if err := s.db.QueryOne(ctx, &row, query, []interface{}{tenantID, namespaceID}, true); err != nil {
		if err == database.ErrRecordNotFound {
			return 0, nil
		}
		return 0, fmt.Errorf("统计配置失败: %w", err)
	}
	return int(row.N), nil
}

// CountByCenter 返回归属该中心实例的已发布配置条数。
func (s *ConfigStore) CountByCenter(ctx context.Context, tenantID, instanceName string) (int, error) {
	query := `SELECT COUNT(*) AS n FROM HUB_SERVICE_CONFIG_DATA d
		INNER JOIN HUB_SERVICE_NAMESPACE n ON d.tenantId = n.tenantId AND d.namespaceId = n.namespaceId
		WHERE d.tenantId = ? AND n.instanceName = ? AND d.activeFlag = 'Y' AND n.activeFlag = 'Y'`
	var row countRow
	if err := s.db.QueryOne(ctx, &row, query, []interface{}{tenantID, instanceName}, true); err != nil {
		if err == database.ErrRecordNotFound {
			return 0, nil
		}
		return 0, fmt.Errorf("统计配置失败: %w", err)
	}
	return int(row.N), nil
}

type countRow struct {
	N int64 `db:"n"`
}

// List 列出已发布配置；group 空表示全部。
func (s *ConfigStore) List(ctx context.Context, tenantID, namespaceID, group string) ([]*model.ConfigRelease, error) {
	query := `SELECT * FROM HUB_SERVICE_CONFIG_DATA WHERE tenantId = ? AND namespaceId = ? AND activeFlag = 'Y'`
	args := []interface{}{tenantID, namespaceID}
	if group != "" {
		query += ` AND groupName = ?`
		args = append(args, group)
	}
	var rows []*configRow
	if err := s.db.Query(ctx, &rows, query, args, true); err != nil {
		return nil, fmt.Errorf("查询配置列表失败: %w", err)
	}
	out := make([]*model.ConfigRelease, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toRelease())
	}
	return out, nil
}

// SaveRelease 写当前发布并追加历史。old 可为 nil（首次发布）。
func (s *ConfigStore) SaveRelease(ctx context.Context, rel *model.ConfigRelease, old *model.ConfigRelease, changeType, operator string) error {
	row := fromRelease(rel, operator)
	var err error
	if old == nil {
		_, err = s.db.Insert(ctx, "HUB_SERVICE_CONFIG_DATA", row, true)
	} else {
		_, err = s.db.Update(ctx, "HUB_SERVICE_CONFIG_DATA", row,
			"tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ?",
			[]interface{}{rel.TenantID, rel.NamespaceID, rel.GroupName, rel.DataID}, true, true)
	}
	if err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	return s.insertHistory(ctx, rel, old, changeType, operator)
}

// Delete 删除当前已发布配置（历史由实现决定是否保留）。
func (s *ConfigStore) Delete(ctx context.Context, tenantID, namespaceID, group, dataID string) error {
	_, err := s.db.Delete(ctx, "HUB_SERVICE_CONFIG_DATA",
		"tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ?",
		[]interface{}{tenantID, namespaceID, group, dataID}, true)
	if err != nil {
		return fmt.Errorf("删除配置失败: %w", err)
	}
	_ = s.DeleteDraft(ctx, tenantID, namespaceID, group, dataID)
	return nil
}

// GetDraft 读取未发布草稿，不存在时返回 nil, nil。
func (s *ConfigStore) GetDraft(ctx context.Context, tenantID, namespaceID, group, dataID string) (*model.ConfigDraft, error) {
	query := `SELECT * FROM HUB_SERVICE_CONFIG_HISTORY
		WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ?
		AND changeType = ? AND activeFlag = 'Y'
		ORDER BY editTime DESC`
	var rows []*historyRow
	if err := s.db.Query(ctx, &rows, query, []interface{}{tenantID, namespaceID, group, dataID, historyChangeDraft}, true); err != nil {
		return nil, fmt.Errorf("查询配置草稿失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	r := rows[0]
	return &model.ConfigDraft{
		TenantID:    r.TenantId,
		NamespaceID: r.NamespaceId,
		GroupName:   r.GroupName,
		DataID:      r.ConfigDataId,
		Content:     r.NewContent,
		ContentType: r.NoteText,
		Description: r.ChangeReason,
	}, nil
}

// SaveDraft 写入或覆盖未发布草稿，不改已发布行、不通知订阅方。
func (s *ConfigStore) SaveDraft(ctx context.Context, draft *model.ConfigDraft, operator string) error {
	if draft == nil || draft.DataID == "" {
		return fmt.Errorf("草稿无效")
	}
	if operator == "" {
		operator = "system"
	}
	if err := s.DeleteDraft(ctx, draft.TenantID, draft.NamespaceID, draft.GroupName, draft.DataID); err != nil {
		return err
	}
	now := time.Now()
	sum := md5.Sum([]byte(draft.Content))
	row := &historyRow{
		ConfigHistoryId: random.Generate32BitRandomString(),
		TenantId:        draft.TenantID,
		ConfigDataId:    draft.DataID,
		NamespaceId:     draft.NamespaceID,
		GroupName:       draft.GroupName,
		ChangeType:      historyChangeDraft,
		NewContent:      draft.Content,
		NewVersion:      0,
		NewMd5Value:     hex.EncodeToString(sum[:]),
		ChangeReason:    draft.Description,
		ChangedBy:       operator,
		ChangedAt:       now,
		AddTime:         now,
		AddWho:          operator,
		EditTime:        now,
		EditWho:         operator,
		OprSeqFlag:      random.Generate32BitRandomString(),
		CurrentVersion:  1,
		ActiveFlag:      model.FlagY,
		NoteText:        draft.ContentType,
	}
	if _, err := s.db.Insert(ctx, "HUB_SERVICE_CONFIG_HISTORY", row, true); err != nil {
		return fmt.Errorf("保存配置草稿失败: %w", err)
	}
	return nil
}

// DeleteDraft 删除未发布草稿。
func (s *ConfigStore) DeleteDraft(ctx context.Context, tenantID, namespaceID, group, dataID string) error {
	_, err := s.db.Delete(ctx, "HUB_SERVICE_CONFIG_HISTORY",
		"tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ? AND changeType = ?",
		[]interface{}{tenantID, namespaceID, group, dataID, historyChangeDraft}, true)
	if err != nil {
		return fmt.Errorf("删除配置草稿失败: %w", err)
	}
	return nil
}

// ListHistory 按版本倒序读取发布历史。
func (s *ConfigStore) ListHistory(ctx context.Context, tenantID, namespaceID, group, dataID string, limit int) ([]*model.ConfigRelease, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `SELECT * FROM HUB_SERVICE_CONFIG_HISTORY
		WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ?
		AND changeType <> '` + historyChangeDraft + `'
		ORDER BY newVersion DESC`
	var rows []*historyRow
	if err := s.db.Query(ctx, &rows, query, []interface{}{tenantID, namespaceID, group, dataID}, true); err != nil {
		return nil, fmt.Errorf("查询配置历史失败: %w", err)
	}
	if len(rows) > limit {
		rows = rows[:limit]
	}
	out := make([]*model.ConfigRelease, 0, len(rows))
	for _, r := range rows {
		out = append(out, &model.ConfigRelease{
			TenantID:    r.TenantId,
			NamespaceID: r.NamespaceId,
			GroupName:   r.GroupName,
			DataID:      r.ConfigDataId,
			Content:     r.NewContent,
			Version:     r.NewVersion,
			MD5:         r.NewMd5Value,
			Reason:      r.ChangeReason,
			PublishedBy: r.ChangedBy,
			PublishedAt: r.ChangedAt,
		})
	}
	return out, nil
}

// insertHistory 追加一条发布历史，old 为变更前内容。
func (s *ConfigStore) insertHistory(ctx context.Context, rel *model.ConfigRelease, old *model.ConfigRelease, changeType, operator string) error {
	now := time.Now()
	if operator == "" {
		operator = "system"
	}
	h := &historyRow{
		ConfigHistoryId: random.Generate32BitRandomString(),
		TenantId:        rel.TenantID,
		ConfigDataId:    rel.DataID,
		NamespaceId:     rel.NamespaceID,
		GroupName:       rel.GroupName,
		ChangeType:      changeType,
		NewContent:      rel.Content,
		NewVersion:      rel.Version,
		NewMd5Value:     rel.MD5,
		ChangeReason:    rel.Reason,
		ChangedBy:       operator,
		ChangedAt:       now,
		AddTime:         now,
		AddWho:          operator,
		EditTime:        now,
		EditWho:         operator,
		OprSeqFlag:      random.Generate32BitRandomString(),
		CurrentVersion:  1,
		ActiveFlag:      model.FlagY,
	}
	if old != nil {
		h.OldContent = old.Content
		h.OldVersion = old.Version
		h.OldMd5Value = old.MD5
	}
	_, err := s.db.Insert(ctx, "HUB_SERVICE_CONFIG_HISTORY", h, true)
	if err != nil {
		return fmt.Errorf("写入配置历史失败: %w", err)
	}
	return nil
}

// toRelease 把当前配置行转为已发布快照。
func (r *configRow) toRelease() *model.ConfigRelease {
	return &model.ConfigRelease{
		TenantID:    r.TenantId,
		NamespaceID: r.NamespaceId,
		GroupName:   r.GroupName,
		DataID:      r.ConfigDataId,
		Content:     r.ConfigContent,
		ContentType: r.ContentType,
		Description: r.ConfigDescription,
		Version:     r.Version,
		MD5:         r.Md5Value,
		PublishedAt: r.EditTime,
		PublishedBy: r.EditWho,
	}
}

// fromRelease 把已发布快照转为当前配置行。
func fromRelease(rel *model.ConfigRelease, operator string) *configRow {
	now := time.Now()
	if operator == "" {
		operator = "system"
	}
	if rel.MD5 == "" {
		sum := md5.Sum([]byte(rel.Content))
		rel.MD5 = hex.EncodeToString(sum[:])
	}
	return &configRow{
		ConfigDataId:      rel.DataID,
		TenantId:          rel.TenantID,
		NamespaceId:       rel.NamespaceID,
		GroupName:         rel.GroupName,
		ConfigContent:     rel.Content,
		ContentType:       rel.ContentType,
		ConfigDescription: rel.Description,
		Encrypted:         model.FlagN,
		Version:           rel.Version,
		Md5Value:          rel.MD5,
		AddTime:           now,
		AddWho:            operator,
		EditTime:          now,
		EditWho:           operator,
		OprSeqFlag:        random.Generate32BitRandomString(),
		CurrentVersion:    1,
		ActiveFlag:        model.FlagY,
	}
}
