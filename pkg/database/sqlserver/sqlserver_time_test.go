package sqlserver

import (
	"strings"
	"testing"
	"time"
)

func TestConvertTimeArgsKeepsLocalWallClock(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	tm := time.Date(2026, 8, 26, 23, 50, 26, 0, loc)
	got := convertTimeArgs([]interface{}{"default", tm, "sid"})
	s, ok := got[1].(string)
	if !ok {
		t.Fatalf("want string, got %T", got[1])
	}
	if !strings.HasPrefix(s, "2026-08-26T23:50:26") {
		t.Fatalf("want local 23:50, got %q", s)
	}
	if strings.Contains(s, "15:50") {
		t.Fatalf("must not convert to UTC: %q", s)
	}
}

func TestConvertTimeArgsNilSlice(t *testing.T) {
	if convertTimeArgs(nil) != nil {
		t.Fatal("nil in, nil out")
	}
}

func TestConvertTimeArgsNilPointer(t *testing.T) {
	var p *time.Time
	got := convertTimeArgs([]interface{}{p})
	if got[0] != nil {
		t.Fatalf("got %#v", got[0])
	}
}
