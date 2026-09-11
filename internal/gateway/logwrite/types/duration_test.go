package types

import (
	"testing"
	"time"
)

func TestDurationMillis(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want int
	}{
		{0, 0},
		{-time.Millisecond, 0},
		{200 * time.Microsecond, 1},
		{499 * time.Microsecond, 1},
		{500 * time.Microsecond, 1},
		{time.Millisecond, 1},
		{1499 * time.Microsecond, 1},
		{1500 * time.Microsecond, 2},
		{7 * time.Millisecond, 7},
	}
	for _, tc := range cases {
		if got := DurationMillis(tc.d); got != tc.want {
			t.Fatalf("DurationMillis(%v) = %d, want %d", tc.d, got, tc.want)
		}
	}
}

func TestElapsedMillis(t *testing.T) {
	start := time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
	if ElapsedMillis(time.Time{}, start) != 0 {
		t.Fatal("zero start should be 0")
	}
	if ElapsedMillis(start, start.Add(200*time.Microsecond)) != 1 {
		t.Fatal("sub-millisecond elapsed should be 1")
	}
	if ElapsedMillis(start, start.Add(1500*time.Microsecond)) != 2 {
		t.Fatal("1.5ms should round to 2")
	}
}
