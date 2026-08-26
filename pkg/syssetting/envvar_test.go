package syssetting

import (
	"strings"
	"testing"
)

func TestValidateEnvVarName(t *testing.T) {
	if err := ValidateEnvVarName("GATEWAY_INTERNAL"); err != nil {
		t.Fatalf("valid name rejected: %v", err)
	}
	if err := ValidateEnvVarName("1abc"); err == nil {
		t.Fatal("name starting with digit should fail")
	}
	if err := ValidateEnvVarName("has-dash"); err == nil {
		t.Fatal("name with dash should fail")
	}
	if err := ValidateEnvVarName(""); err == nil {
		t.Fatal("empty name should fail")
	}
}

func TestExpandEnvVars(t *testing.T) {
	PutEnvVars("t-expand", EnvVarsSettings{Items: []EnvVar{
		{Name: "TOKEN", Value: "secret-token"},
		{Name: "PREFIX", Value: "/api"},
	}})

	got := ExpandEnvVars("t-expand", "Bearer ${TOKEN}")
	if got != "Bearer secret-token" {
		t.Fatalf("got %q", got)
	}
	got = ExpandEnvVars("t-expand", "${PREFIX}/v1")
	if got != "/api/v1" {
		t.Fatalf("got %q", got)
	}
	got = ExpandEnvVars("t-expand", "${MISSING:fallback}")
	if got != "fallback" {
		t.Fatalf("default should apply, got %q", got)
	}
	got = ExpandEnvVars("t-expand", "${MISSING}")
	if got != "" {
		t.Fatalf("missing without default should be empty, got %q", got)
	}
	got = ExpandEnvVars("t-expand", "plain")
	if got != "plain" {
		t.Fatalf("plain text changed: %q", got)
	}
}

func TestMaskEnvVarsHidesSecret(t *testing.T) {
	views := MaskEnvVars(EnvVarsSettings{Items: []EnvVar{
		{Name: "OPEN", Value: "visible", Secret: false},
		{Name: "HIDE", Value: "cipher", Secret: true},
	}})
	if len(views) != 2 {
		t.Fatalf("len=%d", len(views))
	}
	if views[0].Value != "visible" || views[0].HasValue != true {
		t.Fatalf("open var: %+v", views[0])
	}
	if views[1].Value != envVarSecretMask || !views[1].HasValue || views[1].Secret != true {
		t.Fatalf("secret var should be masked: %+v", views[1])
	}
}

func TestValidateEnvVarsDuplicate(t *testing.T) {
	err := ValidateEnvVars(EnvVarsSettings{Items: []EnvVar{
		{Name: "A"}, {Name: "A"},
	}})
	if err == nil || !strings.Contains(err.Error(), "重复") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}
