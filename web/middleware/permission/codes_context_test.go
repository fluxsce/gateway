package permission

import (
	"context"
	"errors"
	"testing"
)

func TestWithUserResourceCodesBagIdempotent(t *testing.T) {
	ctx := WithUserResourceCodesBag(context.Background())
	again := WithUserResourceCodesBag(ctx)
	first, ok1 := userResourceCodesBagFrom(ctx)
	second, ok2 := userResourceCodesBagFrom(again)
	if !ok1 || !ok2 || first != second {
		t.Fatal("repeated bind should reuse the same bag")
	}
}

func TestUserResourceCodesBagLoadsOnce(t *testing.T) {
	bag := &userResourceCodesBag{}
	calls := 0
	load := func() (map[string]struct{}, error) {
		calls++
		return map[string]struct{}{"hub0002": {}}, nil
	}

	first, err := bag.getOrLoad("u1", "t1", load)
	if err != nil {
		t.Fatal(err)
	}
	second, err := bag.getOrLoad("u1", "t1", load)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 load, got %d", calls)
	}
	if _, ok := first["hub0002"]; !ok {
		t.Fatal("missing code from first load")
	}
	if _, ok := second["hub0002"]; !ok {
		t.Fatal("missing code from second load")
	}
}

func TestUserResourceCodesBagCachesError(t *testing.T) {
	bag := &userResourceCodesBag{}
	want := errors.New("db down")
	calls := 0
	load := func() (map[string]struct{}, error) {
		calls++
		return nil, want
	}

	_, err1 := bag.getOrLoad("u1", "t1", load)
	_, err2 := bag.getOrLoad("u1", "t1", load)
	if !errors.Is(err1, want) || !errors.Is(err2, want) {
		t.Fatalf("want cached error, got %v / %v", err1, err2)
	}
	if calls != 1 {
		t.Fatalf("failed load should still run once, got %d", calls)
	}
}

func TestUserResourceCodesBagSkipsOnIdentityMismatch(t *testing.T) {
	bag := &userResourceCodesBag{}
	calls := 0
	loadFor := func(code string) func() (map[string]struct{}, error) {
		return func() (map[string]struct{}, error) {
			calls++
			return map[string]struct{}{code: {}}, nil
		}
	}

	first, err := bag.getOrLoad("u1", "t1", loadFor("hub0002"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := bag.getOrLoad("u2", "t1", loadFor("hub0005"))
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("identity mismatch should reload, got %d", calls)
	}
	if _, ok := first["hub0002"]; !ok {
		t.Fatal("first identity should keep own codes")
	}
	if _, ok := second["hub0005"]; !ok {
		t.Fatal("second identity should not reuse first codes")
	}
}
