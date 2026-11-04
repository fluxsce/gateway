package utils

import "testing"

func TestCounterRate(t *testing.T) {
	if got := CounterRate(100, 40, 2); got != 30 {
		t.Fatalf("正常增量: got %v", got)
	}
	if got := CounterRate(10, 100, 2); got != 0 {
		t.Fatalf("计数器回绕应记 0: got %v", got)
	}
	if got := CounterRate(100, 40, 0); got != 0 {
		t.Fatalf("无效间隔应记 0: got %v", got)
	}
	if got := CounterRate(40, 40, 1); got != 0 {
		t.Fatalf("无增量应记 0: got %v", got)
	}
}
