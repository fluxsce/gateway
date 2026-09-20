package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/database"
	"gateway/pkg/utils/random"
)

// ServiceStore 访问服务定义表。
type ServiceStore struct {
	db database.Database
}

// serviceRow 映射服务定义行。
type serviceRow struct {
	TenantId           string    `db:"tenantId"`
	NamespaceId        string    `db:"namespaceId"`
	GroupName          string    `db:"groupName"`
	ServiceName        string    `db:"serviceName"`
	ServiceType        string    `db:"serviceType"`
	ServiceVersion     string    `db:"serviceVersion"`
	ServiceDescription string    `db:"serviceDescription"`
	MetadataJson       string    `db:"metadataJson"`
	TagsJson           string    `db:"tagsJson"`
	ProtectThreshold   float64   `db:"protectThreshold"`
	AddTime            time.Time `db:"addTime"`
	AddWho             string    `db:"addWho"`
	EditTime           time.Time `db:"editTime"`
	EditWho            string    `db:"editWho"`
	OprSeqFlag         string    `db:"oprSeqFlag"`
	CurrentVersion     int       `db:"currentVersion"`
	ActiveFlag         string    `db:"activeFlag"`
}

// Get 读取服务定义，不存在时返回 nil, nil。
func (s *ServiceStore) Get(ctx context.Context, tenantID, namespaceID, group, name string) (*model.Service, error) {
	query := `SELECT * FROM HUB_SERVICE WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND serviceName = ?`
	var row serviceRow
	err := s.db.QueryOne(ctx, &row, query, []interface{}{tenantID, namespaceID, group, name}, true)
	if err == database.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询服务失败: %w", err)
	}
	return row.toModel(), nil
}

// ListByNamespace 列出命名空间下服务；group 空表示全部组。
func (s *ServiceStore) ListByNamespace(ctx context.Context, tenantID, namespaceID, group string) ([]*model.Service, error) {
	query := `SELECT * FROM HUB_SERVICE WHERE tenantId = ? AND namespaceId = ? AND activeFlag = 'Y'`
	args := []interface{}{tenantID, namespaceID}
	if group != "" {
		query += ` AND groupName = ?`
		args = append(args, group)
	}
	var rows []*serviceRow
	if err := s.db.Query(ctx, &rows, query, args, true); err != nil {
		return nil, fmt.Errorf("查询服务列表失败: %w", err)
	}
	out := make([]*model.Service, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toModel())
	}
	return out, nil
}

// Upsert 插入或更新服务定义。
func (s *ServiceStore) Upsert(ctx context.Context, svc *model.Service, operator string) error {
	existing, err := s.Get(ctx, svc.TenantID, svc.NamespaceID, svc.GroupName, svc.ServiceName)
	if err != nil {
		return err
	}
	row := fromServiceModel(svc, operator)
	if row.ServiceType == "" && existing != nil && existing.ServiceType != "" {
		row.ServiceType = existing.ServiceType
	}
	if existing == nil {
		_, err = s.db.Insert(ctx, "HUB_SERVICE", row, true)
		if err != nil {
			return fmt.Errorf("创建服务失败: %w", err)
		}
		return nil
	}
	_, err = s.db.Update(ctx, "HUB_SERVICE", row,
		"tenantId = ? AND namespaceId = ? AND groupName = ? AND serviceName = ?",
		[]interface{}{svc.TenantID, svc.NamespaceID, svc.GroupName, svc.ServiceName}, true, true)
	if err != nil {
		return fmt.Errorf("更新服务失败: %w", err)
	}
	return nil
}

// Delete 删除服务定义。
func (s *ServiceStore) Delete(ctx context.Context, tenantID, namespaceID, group, name string) error {
	_, err := s.db.Delete(ctx, "HUB_SERVICE",
		"tenantId = ? AND namespaceId = ? AND groupName = ? AND serviceName = ?",
		[]interface{}{tenantID, namespaceID, group, name}, true)
	if err != nil {
		return fmt.Errorf("删除服务失败: %w", err)
	}
	return nil
}

// toModel 把库表行转为领域对象。
func (r *serviceRow) toModel() *model.Service {
	meta := decodeMap(r.MetadataJson)
	return &model.Service{
		TenantID:         r.TenantId,
		NamespaceID:      r.NamespaceId,
		GroupName:        r.GroupName,
		ServiceName:      r.ServiceName,
		Version:          r.ServiceVersion,
		Description:      r.ServiceDescription,
		Metadata:         meta,
		Tags:             decodeMap(r.TagsJson),
		ProtectThreshold: r.ProtectThreshold,
		ServiceType:      r.ServiceType,
		AutoCreated:      model.IsY(meta[model.MetaAutoCreated]),
	}
}

// fromServiceModel 把领域对象转为库表行。
func fromServiceModel(svc *model.Service, operator string) *serviceRow {
	now := time.Now()
	if operator == "" {
		operator = "system"
	}
	meta := map[string]string{}
	for k, v := range svc.Metadata {
		meta[k] = v
	}
	if svc.AutoCreated {
		meta[model.MetaAutoCreated] = model.FlagY
	} else {
		delete(meta, model.MetaAutoCreated)
	}
	return &serviceRow{
		TenantId:           svc.TenantID,
		NamespaceId:        svc.NamespaceID,
		GroupName:          svc.GroupName,
		ServiceName:        svc.ServiceName,
		ServiceType:        serviceTypeOf(svc),
		ServiceVersion:     svc.Version,
		ServiceDescription: svc.Description,
		MetadataJson:       encodeMap(meta),
		TagsJson:           encodeMap(svc.Tags),
		ProtectThreshold:   svc.ProtectThreshold,
		AddTime:            now,
		AddWho:             operator,
		EditTime:           now,
		EditWho:            operator,
		OprSeqFlag:         random.Generate32BitRandomString(),
		CurrentVersion:     1,
		ActiveFlag:         model.FlagY,
	}
}

// decodeMap 解析 JSON 对象字符串，非法输入返回空 map。
func decodeMap(raw string) map[string]string {
	if raw == "" {
		return map[string]string{}
	}
	out := map[string]string{}
	_ = json.Unmarshal([]byte(raw), &out)
	if out == nil {
		return map[string]string{}
	}
	return out
}

// encodeMap 把 map 编码为 JSON 对象字符串。
func encodeMap(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func serviceTypeOf(svc *model.Service) string {
	if svc != nil && svc.ServiceType != "" {
		return svc.ServiceType
	}
	return model.ServiceTypeInternal
}
