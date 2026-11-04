package cpu

import (
	"testing"

	"gateway/pkg/metric/types"

	gopsutilcpu "github.com/shirou/gopsutil/v4/cpu"
)

func TestApplyCPUPercentsUsesIntervalDelta(t *testing.T) {
	prev := gopsutilcpu.TimesStat{
		User: 100, Nice: 10, System: 50, Idle: 800,
		Iowait: 20, Irq: 5, Softirq: 5, Steal: 10,
	}
	cur := prev
	cur.User = 140
	cur.System = 70
	cur.Idle = 840

	metrics := &types.CPUMetrics{}
	applyCPUPercents(metrics, prev, cur)

	if metrics.UserPercent != 40 || metrics.SystemPercent != 20 || metrics.IdlePercent != 40 {
		t.Fatalf("分项应是差值占比: user=%v system=%v idle=%v", metrics.UserPercent, metrics.SystemPercent, metrics.IdlePercent)
	}
	if metrics.UsagePercent != 60 {
		t.Fatalf("总使用率应排除 idle 与 iowait: got %v", metrics.UsagePercent)
	}
	if metrics.IOWaitPercent != 0 || metrics.IrqPercent != 0 || metrics.SoftIrqPercent != 0 {
		t.Fatalf("没有增量的分项应为 0: io=%v irq=%v soft=%v", metrics.IOWaitPercent, metrics.IrqPercent, metrics.SoftIrqPercent)
	}
}

func TestApplyCPUPercentsIncludesNiceAndKeepsStealInUsage(t *testing.T) {
	prev := gopsutilcpu.TimesStat{}
	cur := gopsutilcpu.TimesStat{User: 10, Nice: 10, System: 10, Idle: 50, Steal: 20}

	metrics := &types.CPUMetrics{}
	applyCPUPercents(metrics, prev, cur)

	if metrics.UserPercent != 20 {
		t.Fatalf("nice 应计入用户态: got %v", metrics.UserPercent)
	}
	if metrics.UsagePercent != 50 {
		t.Fatalf("steal 应计入忙占比: got %v", metrics.UsagePercent)
	}
}

func TestApplyCPUPercentsIgnoresCounterReset(t *testing.T) {
	prev := gopsutilcpu.TimesStat{User: 100, Idle: 100}
	cur := gopsutilcpu.TimesStat{User: 1, Idle: 1}

	metrics := &types.CPUMetrics{UsagePercent: 12}
	applyCPUPercents(metrics, prev, cur)
	if metrics.UsagePercent != 12 {
		t.Fatalf("计数回绕时不应改写已有占比: got %v", metrics.UsagePercent)
	}
}
