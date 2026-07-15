package dao

import "testing"

func TestResolveGracefulShutdownTimeoutMs(t *testing.T) {
	tests := []struct {
		name         string
		value        int
		currentValue int
		want         int
	}{
		{name: "explicit value", value: 45000, currentValue: 30000, want: 45000},
		{name: "preserve current value", currentValue: 25000, want: 25000},
		{name: "default missing values", want: defaultGracefulShutdownTimeoutMs},
		{name: "default invalid values", value: -1, currentValue: -1, want: defaultGracefulShutdownTimeoutMs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveGracefulShutdownTimeoutMs(tt.value, tt.currentValue); got != tt.want {
				t.Fatalf("resolveGracefulShutdownTimeoutMs() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestResolveCapacityLimit(t *testing.T) {
	if got := resolveCapacityLimit(2500, 1000, defaultMaxWorkers); got != 2500 {
		t.Fatalf("explicit capacity = %d, want 2500", got)
	}
	if got := resolveCapacityLimit(0, 1500, defaultMaxWorkers); got != 1500 {
		t.Fatalf("preserved capacity = %d, want 1500", got)
	}
	if got := resolveCapacityLimit(0, 0, defaultMaxWorkers); got != defaultMaxWorkers {
		t.Fatalf("default capacity = %d, want %d", got, defaultMaxWorkers)
	}
}
