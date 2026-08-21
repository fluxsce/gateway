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

// TryAcquireLease 尝试持有或续期清理租约。同一时刻通常只有一个节点返回 true。
// 已持有则续期；租约过期则抢占；尚无行则插入。
func TryAcquireLease(ctx context.Context, db database.Database, owner string, ttl time.Duration) bool {
	if db == nil || owner == "" || ttl <= 0 {
		return false
	}
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
		return false
	}
	return true
}
