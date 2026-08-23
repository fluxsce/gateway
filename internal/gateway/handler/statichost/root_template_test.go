package statichost

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandRootTokenFromFirstSegment(t *testing.T) {
	t.Parallel()
	tmpl := "/var/www/sites/app-{v1,v10,v2}/dist"
	got, err := expandRootToken(tmpl, "/v10ui/user/1")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean("/var/www/sites/app-v10/dist")
	if got != want {
		t.Fatalf("expanded = %s, want %s", got, want)
	}
	if _, err := expandRootToken(tmpl, "/v99/index.html"); err == nil {
		t.Fatal("不在允许名单中的第一段应失败")
	}
	if _, err := expandRootToken("/{v1}/dist", []string{"/v1"}[0]); err == nil {
		t.Fatal("没有固定父目录的模板应失败")
	}
}

func TestExpandRootTokenExactRequiresFullSegment(t *testing.T) {
	t.Parallel()
	tmpl := "/var/www/sites/app-{d10,d12}/dist"
	if _, err := expandRootTokenMatch(tmpl, "/d10app/user/1", true); err == nil {
		t.Fatal("精确匹配时 d10app 不应命中 d10")
	}
	got, err := expandRootTokenMatch(tmpl, "/d10/user/1", true)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean("/var/www/sites/app-d10/dist")
	if got != want {
		t.Fatalf("expanded = %s, want %s", got, want)
	}
}

func TestValidateForSaveAllowsRootToken(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	tmpl := filepath.ToSlash(parent) + "/app-{v1,v2}/dist"
	if err := ValidateForSave(tmpl, "index.html", ""); err != nil {
		t.Fatalf("带允许名单的根目录应允许保存: %v", err)
	}
}

func TestBindRootStaysUnderBase(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	if err := os.MkdirAll(filepath.Join(parent, "app-v1", "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	tmpl := filepath.ToSlash(parent) + "/app-{v1,v2}/dist"
	snap, err := Compile(&StaticHostConfig{Enabled: true, RootDirectory: tmpl})
	if err != nil {
		t.Fatal(err)
	}
	if !snap.RootHasPlaceholders {
		t.Fatal("应识别根目录占位符")
	}
	bound, err := snap.bindRoot("/v1/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if !isUnderRoot(parent, bound.RootAbs) {
		t.Fatalf("展开根 %s 应在 %s 下", bound.RootAbs, parent)
	}
	if filepath.Base(filepath.Dir(bound.RootAbs)) != "app-v1" {
		t.Fatalf("展开根不正确: %s", bound.RootAbs)
	}
}
