package contract

import "testing"

func TestCallContextRequire(t *testing.T) {
	if err := (CallContext{}).RequireCenter(); err != ErrInvalidContext {
		t.Fatalf("empty context should fail, got %v", err)
	}
	cc := CallContext{TenantID: "t1", CenterInstanceName: "dev"}
	if err := cc.RequireCenter(); err != nil {
		t.Fatal(err)
	}
	if err := cc.RequireNamespace(); err != ErrInvalidContext {
		t.Fatalf("missing namespace should fail, got %v", err)
	}
	cc.NamespaceID = "ns"
	if err := cc.RequireNamespace(); err != nil {
		t.Fatal(err)
	}
	if !cc.AllowsNamespace("any") {
		t.Fatal("empty allow list should pass")
	}
	cc.AllowedNamespaces = []string{"ns-a", "ns-b"}
	if cc.AllowsNamespace("ns") {
		t.Fatal("unlisted namespace should be rejected")
	}
	if !cc.AllowsNamespace("ns-a") {
		t.Fatal("listed namespace should pass")
	}
}
