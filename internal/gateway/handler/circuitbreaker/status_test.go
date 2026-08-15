package circuitbreaker

import (
	"errors"
	"testing"
)

func TestIsFailureStatus(t *testing.T) {
	cases := []struct {
		code int
		want bool
	}{
		{200, false},
		{404, false},
		{429, false},
		{499, false},
		{500, true},
		{502, true},
		{503, true},
		{599, true},
		{0, false},
	}
	for _, tc := range cases {
		if got := IsFailureStatus(tc.code); got != tc.want {
			t.Fatalf("IsFailureStatus(%d)=%v, want %v", tc.code, got, tc.want)
		}
	}
}

func TestAttemptFailed(t *testing.T) {
	if AttemptFailed(nil, 200) {
		t.Fatal("2xx 不应记失败")
	}
	if AttemptFailed(nil, 404) {
		t.Fatal("4xx 不应记失败")
	}
	if !AttemptFailed(nil, 500) {
		t.Fatal("5xx 应记失败")
	}
	if !AttemptFailed(errors.New("timeout"), 200) {
		t.Fatal("传输错误应记失败")
	}
	if AttemptError(nil, 500) == nil {
		t.Fatal("5xx 应带熔断错误原因")
	}
}
