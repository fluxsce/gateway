package retention

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gateway/pkg/database"
)

// leaseSQLFake 只实现抢锁用到的 Exec / QueryOne。
type leaseSQLFake struct {
	updateN   int64
	updateErr error
	exists    bool
	queryErr  error
	insertErr error
	queries   int
	inserts   int
}

func (f *leaseSQLFake) Exec(_ context.Context, query string, _ []interface{}, _ bool) (int64, error) {
	if strings.Contains(query, "UPDATE HUB_RETENTION_LEASE") {
		return f.updateN, f.updateErr
	}
	f.inserts++
	if f.insertErr != nil {
		return 0, f.insertErr
	}
	return 1, nil
}

func (f *leaseSQLFake) QueryOne(_ context.Context, _ interface{}, _ string, _ []interface{}, _ bool) error {
	f.queries++
	if f.queryErr != nil {
		return f.queryErr
	}
	if !f.exists {
		return database.ErrRecordNotFound
	}
	return nil
}

func TestTryAcquireLeaseRejectsEmpty(t *testing.T) {
	ctx := context.Background()
	if TryAcquireLease(ctx, nil, "n1", time.Minute) {
		t.Fatal("nil db")
	}
}

func TestAcquireLeaseRenewsWithoutQuery(t *testing.T) {
	db := &leaseSQLFake{updateN: 1}
	if !acquireLease(context.Background(), db, "n1", time.Minute) {
		t.Fatal("expected hold")
	}
	if db.queries != 0 || db.inserts != 0 {
		t.Fatalf("queries=%d inserts=%d", db.queries, db.inserts)
	}
}

func TestAcquireLeaseIgnoresHeldByOther(t *testing.T) {
	db := &leaseSQLFake{exists: true}
	if acquireLease(context.Background(), db, "n2", time.Minute) {
		t.Fatal("should ignore when another node holds")
	}
	if db.queries != 1 {
		t.Fatalf("queries=%d", db.queries)
	}
	if db.inserts != 0 {
		t.Fatalf("must not insert when row exists, inserts=%d", db.inserts)
	}
}

func TestAcquireLeaseInsertsWhenMissing(t *testing.T) {
	db := &leaseSQLFake{}
	if !acquireLease(context.Background(), db, "n1", time.Minute) {
		t.Fatal("expected insert")
	}
	if db.queries != 1 || db.inserts != 1 {
		t.Fatalf("queries=%d inserts=%d", db.queries, db.inserts)
	}
}

func TestAcquireLeaseDuplicateInsertIsMiss(t *testing.T) {
	db := &leaseSQLFake{insertErr: database.ErrDuplicateKey}
	if acquireLease(context.Background(), db, "n1", time.Minute) {
		t.Fatal("duplicate insert should be a miss")
	}
	if db.inserts != 1 {
		t.Fatalf("inserts=%d", db.inserts)
	}
}

func TestAcquireLeaseUpdateError(t *testing.T) {
	db := &leaseSQLFake{updateErr: errors.New("db down")}
	if acquireLease(context.Background(), db, "n1", time.Minute) {
		t.Fatal("update error")
	}
	if db.queries != 0 || db.inserts != 0 {
		t.Fatalf("queries=%d inserts=%d", db.queries, db.inserts)
	}
}

func TestAcquireLeaseQueryError(t *testing.T) {
	db := &leaseSQLFake{queryErr: errors.New("query down")}
	if acquireLease(context.Background(), db, "n1", time.Minute) {
		t.Fatal("query error")
	}
	if db.inserts != 0 {
		t.Fatalf("inserts=%d", db.inserts)
	}
}
