package mongo

import "testing"

func TestIsManagedMongoLogCollection(t *testing.T) {
	if !IsManagedMongoLogCollection("HUB_GW_ACCESS_LOG") || !IsManagedMongoLogCollection("HUB_GW_BACKEND_TRACE_LOG") {
		t.Fatal("系统日志集合应在白名单内")
	}
	if IsManagedMongoLogCollection("users") || IsManagedMongoLogCollection("HUB_GW_INSTANCE") {
		t.Fatal("其它集合不得进入索引对齐")
	}
	for _, cmd := range GetMongoInitCommands() {
		if !IsManagedMongoLogCollection(cmd.CollectionName) {
			t.Fatalf("命令列表出现非白名单集合 %s", cmd.CollectionName)
		}
	}
}

func TestKeepIndexNamesIncludesDefaultAndDeclared(t *testing.T) {
	cmds := GetMongoInitCommands()
	byColl := GetIndexCommandsByCollection()
	for coll, list := range byColl {
		keep := keepIndexNames(list)
		if _, ok := keep[defaultIndexName]; !ok {
			t.Fatalf("%s 未保留 %s", coll, defaultIndexName)
		}
		for _, cmd := range list {
			name := desiredIndexName(cmd)
			if _, ok := keep[name]; !ok {
				t.Fatalf("%s 未保留声明索引 %s", coll, name)
			}
		}
		if _, ok := keep["idx_hot_routes"]; ok {
			t.Fatalf("%s 不应保留已废弃的 idx_hot_routes", coll)
		}
	}
	if len(cmds) != 5 {
		t.Fatalf("最小集应为 5 条，得到 %d", len(cmds))
	}
	foundBackendPK := false
	for _, cmd := range cmds {
		if cmd.CollectionName != "HUB_GW_BACKEND_TRACE_LOG" || desiredIndexName(cmd) != "idx_tenant_trace" {
			continue
		}
		if len(cmd.IndexModel.Keys) != 3 ||
			cmd.IndexModel.Keys[0].Key != "tenantId" ||
			cmd.IndexModel.Keys[1].Key != "traceId" ||
			cmd.IndexModel.Keys[2].Key != "backendTraceId" {
			t.Fatalf("从表详情索引键应为 tenantId,traceId,backendTraceId，得到 %#v", cmd.IndexModel.Keys)
		}
		foundBackendPK = true
	}
	if !foundBackendPK {
		t.Fatal("缺少从表 idx_tenant_trace")
	}
}

func TestIsIndexPresentError(t *testing.T) {
	if isIndexPresentError(nil) {
		t.Fatal("nil 不应视为已存在")
	}
	if !isIndexAlreadyExistsError(errString("index already exists")) {
		t.Fatal("already exists 应识别")
	}
	if !isIndexConflictError(errString("IndexOptionsConflict")) {
		t.Fatal("IndexOptionsConflict 应识别")
	}
	if isIndexConflictError(errString("index already exists")) {
		t.Fatal("already exists 不应当成选项冲突去重建")
	}
}

type errString string

func (e errString) Error() string { return string(e) }
