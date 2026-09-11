package types

import (
	"testing"
	"time"
)

func TestCalculateProcessingTimeRoundsSubMillisecond(t *testing.T) {
	start := time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
	log := &AccessLog{
		GatewayStartProcessingTime:    start,
		GatewayFinishedProcessingTime: start.Add(400 * time.Microsecond),
	}
	log.CalculateProcessingTime()
	if log.TotalProcessingTimeMs != 1 || log.GatewayProcessingTimeMs != 1 {
		t.Fatalf("total=%d gateway=%d, want 1/1", log.TotalProcessingTimeMs, log.GatewayProcessingTimeMs)
	}
}

func TestCalculateProcessingTimeExceptRetryWait(t *testing.T) {
	start := time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
	log := &AccessLog{
		GatewayStartProcessingTime:    start,
		GatewayFinishedProcessingTime: start.Add(80 * time.Millisecond),
		BackendResponseTimeMs:         20,
	}
	log.CalculateProcessingTimeExcept(50)
	if log.TotalProcessingTimeMs != 80 {
		t.Fatalf("总时间必须含重试等待 total=%d, want 80", log.TotalProcessingTimeMs)
	}
	if log.GatewayProcessingTimeMs != 10 {
		t.Fatalf("gateway=%d, want 10 (80-20-50)", log.GatewayProcessingTimeMs)
	}
	if log.TotalProcessingTimeMs != log.GatewayProcessingTimeMs+log.BackendResponseTimeMs+50 {
		t.Fatalf("对账失败: total(%d) != gateway(%d)+backend(%d)+retry(50)",
			log.TotalProcessingTimeMs, log.GatewayProcessingTimeMs, log.BackendResponseTimeMs)
	}
}

func TestCalculateProcessingTimeExceptNegativeClampsZero(t *testing.T) {
	start := time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
	log := &AccessLog{
		GatewayStartProcessingTime:    start,
		GatewayFinishedProcessingTime: start.Add(10 * time.Millisecond),
		BackendResponseTimeMs:         8,
	}
	log.CalculateProcessingTimeExcept(20)
	if log.GatewayProcessingTimeMs != 0 {
		t.Fatalf("gateway=%d, want 0", log.GatewayProcessingTimeMs)
	}
}
