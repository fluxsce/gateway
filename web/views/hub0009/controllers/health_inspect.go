package controllers

import (
	"context"
	"fmt"
	"slices"
	"strings"

	scriptmongo "gateway/internal/script/mongo"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/mongo/types"
	"gateway/web/views/hub0009/models"

	"go.mongodb.org/mongo-driver/bson"
)

var clickHouseLogTables = []string{"HUB_GW_ACCESS_LOG", "HUB_GW_BACKEND_TRACE_LOG"}

const defaultMongoIndexName = "_id_"

type storeMeta struct {
	rootEnabled bool
	connEnabled bool
	database    string
}

func (m storeMeta) enabled() bool {
	return m.rootEnabled && m.connEnabled
}

func (m storeMeta) skipMessage() string {
	if !m.rootEnabled {
		return "配置未启用"
	}
	if !m.connEnabled {
		return "连接未启用"
	}
	return "进程未加载连接"
}

func clickHouseMeta() storeMeta {
	return storeMeta{
		rootEnabled: true,
		connEnabled: config.GetBool("database.connections."+clickHouseConn+".enabled", false),
		database:    config.GetString("database.connections."+clickHouseConn+".connection.database", ""),
	}
}

func mongoMeta(name string) storeMeta {
	if name == "" || name == cacheDefaultAlias {
		name = config.GetString("mongo.default", "mongo_main")
	}
	return storeMeta{
		rootEnabled: config.GetBool("mongo.enabled", false),
		connEnabled: config.GetBool("mongo.connections."+name+".enabled", false),
		database:    config.GetString("mongo.connections."+name+".database", ""),
	}
}

func attachStore(item models.HealthItem, meta storeMeta, objects []models.HealthStoreObject) models.HealthItem {
	item.Enabled = meta.enabled()
	item.Database = meta.database
	item.Objects = objects
	return item
}

func plannedClickHouseObjects() []models.HealthStoreObject {
	objects := make([]models.HealthStoreObject, 0, len(clickHouseLogTables))
	for _, name := range clickHouseLogTables {
		objects = append(objects, models.HealthStoreObject{Name: name, Kind: "table"})
	}
	return objects
}

func plannedMongoObjects() []models.HealthStoreObject {
	grouped := scriptmongo.GetIndexCommandsByCollection()
	names := scriptmongo.GetCollectionNames()
	slices.Sort(names)
	objects := make([]models.HealthStoreObject, 0, len(names))
	for _, name := range names {
		obj := models.HealthStoreObject{
			Name: name,
			Kind: "collection",
			Indexes: []models.HealthStoreIndex{{
				Name: defaultMongoIndexName,
				Keys: "_id",
			}},
		}
		for _, cmd := range grouped[name] {
			obj.Indexes = append(obj.Indexes, models.HealthStoreIndex{
				Name: desiredIndexLabel(cmd),
				Keys: formatIndexKeys(cmd.IndexModel.Keys),
			})
		}
		objects = append(objects, obj)
	}
	return objects
}

func desiredIndexLabel(cmd scriptmongo.MongoIndexCommand) string {
	if cmd.IndexModel.Options != nil && cmd.IndexModel.Options.Name != "" {
		return cmd.IndexModel.Options.Name
	}
	return cmd.Description
}

func formatIndexKeys(keys bson.D) string {
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		name := key.Key
		if indexKeyDescending(key.Value) {
			name += " desc"
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, ", ")
}

func indexKeyDescending(value interface{}) bool {
	switch v := value.(type) {
	case int:
		return v < 0
	case int32:
		return v < 0
	case int64:
		return v < 0
	default:
		return false
	}
}

func inspectClickHouse(parent context.Context, db database.Database) []models.HealthStoreObject {
	objects := plannedClickHouseObjects()
	if db == nil {
		return objects
	}
	ctx, cancel := context.WithTimeout(parent, healthPingTimeout)
	defer cancel()
	for i, obj := range objects {
		objects[i].Exists = clickHouseTableExists(ctx, db, obj.Name)
	}
	return objects
}

func clickHouseTableExists(ctx context.Context, db database.Database, name string) bool {
	allowed := false
	for _, table := range clickHouseLogTables {
		if table == name {
			allowed = true
			break
		}
	}
	if !allowed {
		return false
	}
	var rows []struct {
		N uint8 `db:"n"`
	}
	err := db.Query(ctx, &rows, "SELECT 1 AS n FROM "+name+" LIMIT 0", nil, true)
	if err != nil {
		logger.Warn("系统健康读取 ClickHouse 表失败", "table", name, "error", err.Error())
		return false
	}
	return true
}

func inspectMongo(parent context.Context, name string, cli interface {
	DefaultDatabase() (types.MongoDatabase, error)
}) []models.HealthStoreObject {
	objects := plannedMongoObjects()
	if cli == nil {
		return objects
	}
	mdb, err := cli.DefaultDatabase()
	if err != nil || mdb == nil {
		logger.Warn("系统健康读取 Mongo 库失败", "name", name, "error", fmt.Sprintf("%v", err))
		return objects
	}
	ctx, cancel := context.WithTimeout(parent, healthPingTimeout)
	defer cancel()
	collNames, err := mdb.ListCollectionNames(ctx, nil)
	if err != nil {
		logger.Warn("系统健康列出 Mongo 集合失败", "name", name, "error", err.Error())
		return objects
	}
	exists := make(map[string]struct{}, len(collNames))
	for _, collName := range collNames {
		exists[collName] = struct{}{}
	}
	for i, obj := range objects {
		if _, ok := exists[obj.Name]; !ok {
			continue
		}
		objects[i].Exists = true
		present := listMongoIndexNames(ctx, mdb.Collection(obj.Name))
		for j, idx := range objects[i].Indexes {
			_, objects[i].Indexes[j].Present = present[idx.Name]
		}
	}
	return objects
}

func listMongoIndexNames(ctx context.Context, coll types.MongoCollection) map[string]struct{} {
	names := map[string]struct{}{}
	if coll == nil {
		return names
	}
	cursor, err := coll.ListIndexes(ctx)
	if err != nil {
		logger.Warn("系统健康列出 Mongo 索引失败", "error", err.Error())
		return names
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		indexName, _ := doc["name"].(string)
		if indexName != "" {
			names[indexName] = struct{}{}
		}
	}
	return names
}
