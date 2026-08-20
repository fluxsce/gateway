package audit

import (
	"strings"
	"testing"

	"gateway/web/utils/constants"
)

func TestIsSensitiveButtonCode(t *testing.T) {
	if !IsSensitiveButtonCode("hub0002:delete") {
		t.Fatal("delete should be sensitive")
	}
	if IsSensitiveButtonCode("hub0002:add") {
		t.Fatal("add should not be sensitive")
	}
}

func TestIsWriteAuditButtonCode(t *testing.T) {
	if !IsWriteAuditButtonCode("hub0002:add") {
		t.Fatal("add should be write-audited")
	}
	if !IsWriteAuditButtonCode("hub0023:reset") {
		t.Fatal("hub0023:reset (resend) should be write-audited")
	}
	if !IsWriteAuditButtonCode("hub0022:addProxy") {
		t.Fatal("addProxy should be write-audited")
	}
	if !IsWriteAuditButtonCode("hub0022:circuitBreaker") {
		t.Fatal("circuitBreaker should be write-audited")
	}
	if !IsWriteAuditButtonCode("hub0021:routerConfig") {
		t.Fatal("routerConfig should be write-audited")
	}
	if !IsWriteAuditButtonCode("hub0020:export") {
		t.Fatal("export should be write-audited")
	}
	if IsWriteAuditButtonCode("hub0023:resetQuery") {
		t.Fatal("resetQuery should not be write-audited")
	}
	if IsWriteAuditButtonCode("hub0004:reset") {
		t.Fatal("form reset should not be write-audited")
	}
	if IsWriteAuditButtonCode("hub0002:search") {
		t.Fatal("search should not be write-audited")
	}
	if IsWriteAuditButtonCode("hub0002:view") {
		t.Fatal("view should not be write-audited")
	}
	if IsWriteAuditButtonCode("hub0005:roleAuth") {
		t.Fatal("roleAuth should not be middleware-audited")
	}
	if FirstWriteAuditButtonCode([]string{"hub0020:securityConfig:add", "hub0020:securityConfig"}) != "hub0020:securityConfig:add" {
		t.Fatal("first write code should be :add")
	}
}

func TestAuditActionFromResourceCode(t *testing.T) {
	cases := map[string]string{
		"hub0002:delete":           AuditActionDelete,
		"hub0002:add":              AuditActionCreate,
		"hub0080:create":           AuditActionCreate,
		"hub0002:edit":             AuditActionUpdate,
		"hub0002:update":           AuditActionUpdate,
		"hub0002:resetPassword":    AuditActionUpdate,
		"hub0002:roleAuth":         AuditActionGrant,
		"hub0043:history:rollback": AuditActionRollback,
		"hub0023:reset":            AuditActionUpdate,
		"hub0022:addProxy":         AuditActionUpdate,
		"hub0022:circuitBreaker":   AuditActionUpdate,
		"hub0021:routerConfig":     AuditActionUpdate,
		"hub0020:export":           AuditActionExport,
		"hub0080:test":             AuditActionUpdate,
		"hub0002:query":            "",
		"hub0002:search":           "",
		"hub0002:view":             "",
		"hub0004:reset":            "",
	}
	for code, want := range cases {
		if got := AuditActionFromResourceCode(code); got != want {
			t.Fatalf("%s: want %q, got %q", code, want, got)
		}
	}
}

func TestModuleCodeFromResourceOrPath(t *testing.T) {
	if got := ModuleCodeFromResourceOrPath("hub0021:delete", ""); got != "hub0021" {
		t.Fatalf("from resource: got %q", got)
	}
	if got := ModuleCodeFromResourceOrPath("hub0005", ""); got != "hub0005" {
		t.Fatalf("module resource: got %q", got)
	}
	if got := ModuleCodeFromResourceOrPath("", constants.APIRoot+"/hub0002/deleteUser"); got != "hub0002" {
		t.Fatalf("from path: got %q", got)
	}
}

func TestSanitizeAuditDetail(t *testing.T) {
	got := SanitizeAuditDetail(map[string]interface{}{
		"userName":  "alice",
		"password":  "secret",
		"captchaId": "ticket",
		"content":   "huge",
	})
	if strings.Contains(got, "secret") || strings.Contains(got, "ticket") || strings.Contains(got, "huge") {
		t.Fatalf("sensitive or omitted fields leaked: %s", got)
	}
	if !strings.Contains(got, "alice") {
		t.Fatalf("expected userName kept: %s", got)
	}
}
