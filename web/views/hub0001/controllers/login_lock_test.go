package controllers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

type memFailStore struct {
	mu   sync.Mutex
	data map[string]string
}

func newMemFailStore() *memFailStore {
	return &memFailStore{data: make(map[string]string)}
}

func (m *memFailStore) GetString(_ context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data[key], nil
}

func (m *memFailStore) SetString(_ context.Context, key string, value string, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *memFailStore) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func TestLoginLockService_noLockBeforeLimit(t *testing.T) {
	lock := &LoginLockService{store: newMemFailStore(), ttl: time.Minute}
	ctx := context.Background()
	for i := 0; i < 4; i++ {
		if remaining := lock.RecordFailure(ctx, "admin"); remaining > 0 {
			t.Fatalf("fail %d should not cool down, remaining=%v", i+1, remaining)
		}
	}
	locked, remaining := lock.Check(ctx, "admin")
	if locked || remaining > 0 {
		t.Fatalf("after 4 fails locked=%v remaining=%v", locked, remaining)
	}
}

func TestLoginLockService_progressiveCooldownAndClear(t *testing.T) {
	store := newMemFailStore()
	lock := &LoginLockService{store: store, ttl: time.Minute}
	ctx := context.Background()

	var remaining time.Duration
	for i := 0; i < 5; i++ {
		remaining = lock.RecordFailure(ctx, "admin")
	}
	if remaining < 25*time.Second || remaining > 30*time.Second {
		t.Fatalf("5th fail remaining=%v, want ~30s", remaining)
	}
	locked, checkRemain := lock.Check(ctx, "admin")
	if !locked || checkRemain < 20*time.Second {
		t.Fatalf("after 5 fails locked=%v remaining=%v", locked, checkRemain)
	}

	raw, _ := store.GetString(ctx, loginFailPrefix+"admin")
	var rec loginFailRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		t.Fatal(err)
	}
	rec.LockedUntil = time.Now().Add(-time.Second).Unix()
	payload, _ := json.Marshal(rec)
	_ = store.SetString(ctx, loginFailPrefix+"admin", string(payload), time.Minute)

	remaining = lock.RecordFailure(ctx, "admin")
	if remaining < 50*time.Second || remaining > time.Minute {
		t.Fatalf("6th fail remaining=%v, want ~60s", remaining)
	}

	lock.Clear(ctx, "admin")
	locked, remaining = lock.Check(ctx, "admin")
	if locked || remaining > 0 {
		t.Fatalf("after clear locked=%v remaining=%v", locked, remaining)
	}
}

func TestLoginLockService_nilStoreDoesNotPanic(t *testing.T) {
	lock := &LoginLockService{}
	ctx := context.Background()
	locked, remaining := lock.Check(ctx, "anyone")
	if locked || remaining > 0 {
		t.Fatalf("nil store locked=%v remaining=%v", locked, remaining)
	}
	if got := lock.RecordFailure(ctx, "anyone"); got != 0 {
		t.Fatalf("nil store RecordFailure=%v", got)
	}
	lock.Clear(ctx, "anyone")
}

func TestRemainSeconds(t *testing.T) {
	if RemainSeconds(0) != 0 {
		t.Fatal("zero duration should be 0")
	}
	if got := RemainSeconds(1500 * time.Millisecond); got != 2 {
		t.Fatalf("ceil 1.5s got %d", got)
	}
}
