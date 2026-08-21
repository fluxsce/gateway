package retention

import (
	"context"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/syssetting"
)

const deleteBatchPause = 50 * time.Millisecond

// TableJob 按 Dataset 声明硬删过期行。第一期动作只有 delete，软删和转冷以后换执行器即可。
type TableJob struct {
	db database.Database
	ds Dataset
}

// NewTableJob 创建按表删除的执行器。Cutoff 返回 skip 时该租户本轮不删。
func NewTableJob(db database.Database, ds Dataset) *TableJob {
	return &TableJob{db: db, ds: ds}
}

// NewJobs 按当前 Dataset 清单生成删除任务。
func NewJobs(db database.Database) []Job {
	datasets := Datasets()
	jobs := make([]Job, 0, len(datasets))
	for i := range datasets {
		jobs = append(jobs, NewTableJob(db, datasets[i]))
	}
	return jobs
}

// Name 返回任务名。
func (j *TableJob) Name() string {
	if j == nil {
		return ""
	}
	return j.ds.Name
}

// Run 按已知租户分批删除早于截止时间的行。单轮有条数上限，剩余下轮再删。
func (j *TableJob) Run(ctx context.Context) error {
	if j == nil || j.db == nil || j.ds.Cutoff == nil || j.ds.Table == "" || j.ds.TimeCol == "" {
		return nil
	}
	dbType := sqlutils.GetDatabaseType(j.db)
	for _, tenantId := range syssetting.KnownTenantIDs() {
		if err := ctx.Err(); err != nil {
			return err
		}
		before, ok := j.ds.Cutoff(tenantId)
		if !ok {
			continue
		}
		sql, args, err := limitedDelete(dbType, j.ds.Table, j.ds.TimeCol, tenantId, before)
		if err != nil {
			logger.Error("生命周期清理SQL无效", "job", j.ds.Name, "error", err)
			continue
		}
		total, truncated, err := j.deleteTenant(ctx, sql, args)
		if err != nil {
			logger.Error("生命周期清理失败", "job", j.ds.Name, "table", j.ds.Table, "tenantId", tenantId, "error", err)
			continue
		}
		if total > 0 {
			logger.Info("生命周期清理完成", "job", j.ds.Name, "tenantId", tenantId, "count", total, "before", before, "truncated", truncated)
		}
	}
	return nil
}

// deleteTenant 循环按批删除，删不满一批即结束；达到 maxDeleteBatches 时本轮停止。
func (j *TableJob) deleteTenant(ctx context.Context, sql string, args []interface{}) (int64, bool, error) {
	var total int64
	for i := 0; i < maxDeleteBatches; i++ {
		if err := ctx.Err(); err != nil {
			return total, false, err
		}
		affected, err := j.db.Exec(ctx, sql, args, true)
		if err != nil {
			return total, false, err
		}
		total += affected
		if affected < int64(deleteBatchSize) {
			return total, false, nil
		}
		// 批次之间让出锁，避免长时间占着表
		time.Sleep(deleteBatchPause)
	}
	return total, true, nil
}
