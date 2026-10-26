package catalog

import "testing"

func TestResourceCode(t *testing.T) {
	if got := ResourceCode("json-formatter"); got != "hub0010:tool:json-formatter" {
		t.Fatalf("resource code = %s", got)
	}
}

func TestNormalizeConfigIndent(t *testing.T) {
	tool, ok := Find("json-formatter")
	if !ok {
		t.Fatal("missing json-formatter")
	}
	got, err := NormalizeConfig(tool, map[string]any{"indent": float64(4)})
	if err != nil {
		t.Fatal(err)
	}
	if got["indent"] != 4 {
		t.Fatalf("indent = %#v", got["indent"])
	}
	if _, err := NormalizeConfig(tool, map[string]any{"indent": float64(3)}); err == nil {
		t.Fatal("expected reject indent 3")
	}
}

func TestNormalizeConfigRejectsUnknownToolFields(t *testing.T) {
	tool, ok := Find("timestamp")
	if !ok {
		t.Fatal("missing timestamp")
	}
	if _, err := NormalizeConfig(tool, map[string]any{"indent": 2}); err == nil {
		t.Fatal("expected reject extra config")
	}
	got, err := NormalizeConfig(tool, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("config = %#v", got)
	}
}
