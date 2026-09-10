package retention

import (
	"context"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

const (
	leaseKey        = "cleanup"
	leaseTenantId   = "default"
	leaseActiveFlag = "Y"
)

// leaseSQL 抢锁只用到的 SQL 能力，便于单测只 mock Exec 与 QueryOne。
type leaseSQL interface {
	Exec(ctx context.Context, query string, args []interface{}, autoCommit bool) (int64, error)
	QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error
}

// leaseProbe 仅用于确认租约行是否已在，不读取 owner 做决策。
type leaseProbe struct {
	Owner string `db:"owner"`
}

// TryAcquireLease 尝试持有或续期清理租约。同一时刻通常只有一个节点返回 true。
// 已持有则续期；租约过期则抢占；尚无行则插入。
// 他节点持有未过期租约时直接放弃，不写库、不当错误。
func TryAcquireLease(ctx context.Context, db database.Database, owner string, ttl time.Duration) bool {
	if db == nil || owner == "" || ttl <= 0 {
		return false
	}
	return acquireLease(ctx, db, owner, ttl)
}

// acquireLease 先 UPDATE 续期或抢占；影响 0 行再查行是否已在。
// 已在表示他节点持有有效租约，本轮忽略。只有表空才 INSERT。
func acquireLease(ctx context.Context, db leaseSQL, owner string, ttl time.Duration) bool {
	now := time.Now()
	expire := now.Add(ttl)
	opr := random.GenerateUniqueStringWithPrefix("", 32)

	n, err := db.Exec(ctx, `
		UPDATE HUB_RETENTION_LEASE
		SET owner = ?, expireTime = ?, editTime = ?, editWho = ?,
		    oprSeqFlag = ?, currentVersion = currentVersion + 1
		WHERE tenantId = ? AND leaseKey = ? AND activeFlag = ?
		  AND (owner = ? OR expireTime < ?)
	`, []interface{}{owner, expire, now, owner, opr, leaseTenantId, leaseKey, leaseActiveFlag, owner, now}, true)
	if err != nil {
		logger.Warn("续期生命周期租约失败", "error", err.Error())
		return false
	}
	if n > 0 {
		return true
	}

	exists, err := leaseRowExists(ctx, db)
	if err != nil {
		logger.Warn("查询生命周期租约失败", "error", err.Error())
		return false
	}
	if exists {
		return false
	}

	_, err = db.Exec(ctx, `
		INSERT INTO HUB_RETENTION_LEASE (
			tenantId, leaseKey, owner, expireTime,
			addTime, addWho, editTime, editWho,
			oprSeqFlag, currentVersion, activeFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, []interface{}{
		leaseTenantId, leaseKey, owner, expire,
		now, owner, now, owner,
		opr, 1, leaseActiveFlag,
	}, true)
	if err != nil {
		// 两个空表节点同时 INSERT 会撞主键，视为没抢到
		if database.IsDuplicateKey(err) {
			return false
		}
		logger.Warn("写入生命周期租约失败", "error", err.Error())
		return false
	}
	return true
}

// leaseRowExists 在 UPDATE 未抢到后确认 default/cleanup 行是否已在。
// 查不到记录不当错误；其它查询失败原样返回。
func leaseRowExists(ctx context.Context, db leaseSQL) (bool, error) {
	var row leaseProbe
	err := db.QueryOne(ctx, &row, `
		SELECT owner FROM HUB_RETENTION_LEASE
		WHERE tenantId = ? AND leaseKey = ? AND activeFlag = ?
	`, []interface{}{leaseTenantId, leaseKey, leaseActiveFlag}, true)
	if err == nil {
		return true, nil
	}
	if database.IsRecordNotFound(err) {
		return false, nil
	}
	return false, err
}
