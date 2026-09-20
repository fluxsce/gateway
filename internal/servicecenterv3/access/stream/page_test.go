package stream

import "testing"

func TestPageOf(t *testing.T) {
	all := []int{1, 2, 3, 4, 5}
	out, next, total := pageOf(all, 2, "")
	if total != 5 || next != "2" || len(out) != 2 || out[0] != 1 {
		t.Fatalf("first page: out=%v next=%s total=%d", out, next, total)
	}
	out, next, total = pageOf(all, 2, next)
	if total != 5 || next != "4" || len(out) != 2 || out[0] != 3 {
		t.Fatalf("second page: out=%v next=%s total=%d", out, next, total)
	}
	out, next, _ = pageOf(all, 2, next)
	if next != "" || len(out) != 1 || out[0] != 5 {
		t.Fatalf("last page: out=%v next=%s", out, next)
	}
	allOut, next, total := pageOf(all, 0, "")
	if next != "" || total != 5 || len(allOut) != 5 {
		t.Fatalf("no page: out=%v next=%s total=%d", allOut, next, total)
	}
}
