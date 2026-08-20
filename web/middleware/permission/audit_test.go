package permission

import (
	"testing"
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
	if IsWriteAuditButtonCode("hub0023:resetQuery") {
		t.Fatal("resetQuery should not be write-audited")
	}
	if IsWriteAuditButtonCode("hub0005:roleAuth") {
		t.Fatal("roleAuth should not be middleware-audited")
	}
	if FirstWriteAuditButtonCode([]string{"hub0020:securityConfig:add", "hub0020:securityConfig"}) != "hub0020:securityConfig:add" {
		t.Fatal("first write code should be :add")
	}
}

func TestAuditTargetFromGetter(t *testing.T) {
	got := AuditTargetFromGetter(func(key string) string {
		if key == "securityConfigId" {
			return "SC1"
		}
		if key == "configName" {
			return "default"
		}
		return ""
	})
	if got.TargetType != "SECURITY_CONFIG" || got.TargetId != "SC1" || got.TargetName != "default" {
		t.Fatalf("got %+v", got)
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
		"hub0002:query":            "",
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
	if got := ModuleCodeFromResourceOrPath("", "/gateway/hub0002/deleteUser"); got != "hub0002" {
		t.Fatalf("from path: got %q", got)
	}
}

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
