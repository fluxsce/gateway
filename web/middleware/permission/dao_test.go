package permission

import (
	"testing"
)

func TestBareModuleCodes(t *testing.T) {
	got := BareModuleCodes([]string{"hub0061", "hub0002", "hub0002:search"})
	has061 := false
	has002 := false
	for _, code := range got {
		if code == "hub0061" {
			has061 = true
		}
		if code == "hub0002" {
			has002 = true
		}
	}
	if !has061 {
		t.Fatalf("hub0061 is bare, got %v", got)
	}
	if has002 {
		t.Fatalf("hub0002 has a child button, should not be bare: %v", got)
	}
}
