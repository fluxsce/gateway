package model

import "testing"

func TestParseExtPropertyMixedTypes(t *testing.T) {
	ext := `{"channelName":"ops","alertEnabled":"Y","nodeEvictionThreshold":8,"namespaceIds":["a","b"]}`
	m := ParseExtProperty(ext)
	if ExtString(m, "channelName") != "ops" {
		t.Fatalf("channel: %q", ExtString(m, "channelName"))
	}
	ac := ParseAlertConfig(ext)
	if !ac.AlertEnabled || ac.NodeEvictionThreshold != 8 {
		t.Fatalf("alert parse: %+v", ac)
	}
	if ExtYN(m, "alertEnabled", false) != true {
		t.Fatal("alertEnabled must be Y")
	}
	if ExtYN(map[string]interface{}{"alertEnabled": "true"}, "alertEnabled", false) {
		t.Fatal("true/1/yes must not count as Y")
	}
	if ExtYN(map[string]interface{}{"alertEnabled": true}, "alertEnabled", false) {
		t.Fatal("JSON bool must not count as Y")
	}
	if ExtYN(map[string]interface{}{"alertEnabled": "N"}, "alertEnabled", true) {
		t.Fatal("N must be false")
	}
	if YN(true) != FlagY || YN(false) != FlagN || !IsY(FlagY) || IsY(FlagN) {
		t.Fatal("YN/IsY must only use Y and N")
	}
	scope := ParseTokenScope(`{"instanceName":"sc-prod","environment":"PRODUCTION","namespaceIds":"ns1,ns2"}`)
	if scope.InstanceName != "sc-prod" || scope.Environment != "PRODUCTION" || len(scope.NamespaceIDs) != 2 {
		t.Fatalf("token scope: %+v", scope)
	}
}

func TestCenterRuntimeKeyAndFingerprint(t *testing.T) {
	a := &CenterInstance{InstanceName: "sc", Environment: "DEV", ListenPort: 1}
	b := &CenterInstance{InstanceName: "sc", Environment: "PROD", ListenPort: 1}
	if a.RuntimeKey() == b.RuntimeKey() {
		t.Fatal("dev/prod must not share runtime key")
	}
	if a.RuntimeKey() != "sc:DEV" {
		t.Fatalf("runtime key separator want sc:DEV, got %q", a.RuntimeKey())
	}
	c := *a
	c.EnableAuth = FlagY
	if a.ListenFingerprint() == c.ListenFingerprint() {
		t.Fatal("auth change must change listen fingerprint")
	}
}
