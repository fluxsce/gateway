package process

import "testing"

func TestMachineCPUPercent(t *testing.T) {
	if got := machineCPUPercent(0.5, 1, 4); got != 12.5 {
		t.Fatalf("半核在 4 核机器上应为 12.5: got %v", got)
	}
	if got := machineCPUPercent(2, 1, 1); got != 100 {
		t.Fatalf("超过整机应限制为 100: got %v", got)
	}
	if got := machineCPUPercent(-1, 1, 2); got != 0 {
		t.Fatalf("负增量应为 0: got %v", got)
	}
	if got := machineCPUPercent(1, 0, 2); got != 0 {
		t.Fatalf("无效墙钟应为 0: got %v", got)
	}
}
