package routes

import "testing"

func TestModuleAPIPrefix(t *testing.T) {
	got := ModuleAPIPrefix("hub0007")
	want := APIRoot + "/hub0007"
	if got != want {
		t.Fatalf("ModuleAPIPrefix: got %q, want %q", got, want)
	}
	if got != "/gateway/hub0007" {
		t.Fatalf("default APIRoot prefix changed: got %q", got)
	}
}
