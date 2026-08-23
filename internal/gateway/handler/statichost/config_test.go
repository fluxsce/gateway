package statichost

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompileBuildsReadOnlySnapshot(t *testing.T) {
	root := t.TempDir()
	cfg := &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		RewriteRules: []RewriteRule{
			{Mode: RewriteModeRegex, From: `^/v\d+/(.*)$`, To: "/$1"},
		},
	}
	snap, err := Compile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !snap.IsActive() || snap.RootAbs == "" || len(snap.Rules) != 1 || snap.Rules[0].Re == nil {
		t.Fatalf("快照未编译完整: %+v", snap)
	}
	if cfg.RewriteRules[0].From != `^/v\d+/(.*)$` {
		t.Fatal("Compile 不得改写调用方规则切片语义")
	}
}

func TestCompileRejectsInvalidRegex(t *testing.T) {
	_, err := Compile(&StaticHostConfig{
		Enabled:       true,
		RootDirectory: t.TempDir(),
		RewriteRules: []RewriteRule{
			{Mode: RewriteModeRegex, From: "(unclosed", To: "/$1"},
		},
	})
	if err == nil {
		t.Fatal("非法正则应在 Compile 失败")
	}
}

func TestValidateForSave(t *testing.T) {
	root := t.TempDir()
	if err := ValidateForSave(root, "index.html", "prefix /old /new"); err != nil {
		t.Fatalf("合法配置应通过: %v", err)
	}
	if err := ValidateForSave("", "index.html", ""); err == nil {
		t.Fatal("空根目录应失败")
	}
	file := filepath.Join(root, "not-dir.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateForSave(file, "index.html", ""); err == nil {
		t.Fatal("根目录指向文件应失败")
	}
	if err := ValidateForSave(root, "../secret", ""); err == nil {
		t.Fatal("非法索引文件应失败")
	}
	if err := ValidateForSave(root, "index.html", "regex (unclosed /$1"); err == nil {
		t.Fatal("非法正则应失败")
	}
	missing := filepath.Join(root, "not-created-yet")
	if err := ValidateForSave(missing, "index.html", ""); err != nil {
		t.Fatalf("尚不存在的目录应允许先保存: %v", err)
	}
}

func TestParseAllowedExtensionsAndAcceptEncoding(t *testing.T) {
	got := ParseAllowedExtensionsText(".JS, css, /bad")
	if len(got) != 2 || got[0] != ".js" || got[1] != ".css" {
		t.Fatalf("扩展名规范化失败: %+v", got)
	}
	if !acceptsEncoding("gzip, deflate, br", "br") {
		t.Fatal("应接受 br")
	}
	if acceptsEncoding("gzip;q=0", "gzip") {
		t.Fatal("q=0 不应接受")
	}
}

func TestIsHashedAsset(t *testing.T) {
	if !isHashedAsset("app.a1b2c3d4e5.js") {
		t.Fatal("webpack 哈希文件名应识别")
	}
	if !isHashedAsset("index-B3k9xY2a.css") {
		t.Fatal("Vite 哈希文件名应识别")
	}
	if isHashedAsset("app.js") {
		t.Fatal("无哈希的 js 不应识别")
	}
	if isHashedAsset("jquery-3.7.1.min.js") {
		t.Fatal("版本号文件名不应当成内容哈希")
	}
}

func TestParseSecurityHeadersText(t *testing.T) {
	items, err := ParseSecurityHeadersText("X-Frame-Options: SAMEORIGIN\nReferrer-Policy: no-referrer")
	if err != nil || len(items) != 2 {
		t.Fatalf("合法安全头应通过: %v %+v", err, items)
	}
	if _, err := ParseSecurityHeadersText("Location: /evil"); err == nil {
		t.Fatal("非白名单头应拒绝")
	}
	if _, err := ParseSecurityHeadersText("Cache-Control: no-store"); err == nil {
		t.Fatal("缓存头应拒绝")
	}
	if _, err := ParseSecurityHeadersText("X-Frame-Options: SAMEORIGIN\r\nX-Injected: 1"); err == nil {
		t.Fatal("含换行的安全头应拒绝")
	}
	if _, err := ParseSecurityHeadersText("X-Frame-Options: SAMEORIGIN\x01"); err == nil {
		t.Fatal("含控制字符的安全头应拒绝")
	}
}

func TestParseCacheControlByExtText(t *testing.T) {
	got, err := ParseCacheControlByExtText(".js=86400\n.css = 3600")
	if err != nil || got[".js"] != 86400 || got[".css"] != 3600 {
		t.Fatalf("按扩展名缓存解析失败: %v %+v", err, got)
	}
	if _, err := ParseCacheControlByExtText(".js=999999999"); err == nil {
		t.Fatal("超大秒数应拒绝")
	}
	if _, err := ParseCacheControlByExtText(".html=60"); err == nil {
		t.Fatal("HTML 缓存覆盖应拒绝")
	}
}

func TestParseFallbackRootsText(t *testing.T) {
	got, err := ParseFallbackRootsText("/var/www/a\n/var/www/b")
	if err != nil || len(got) != 2 {
		t.Fatalf("备用目录解析失败: %v %+v", err, got)
	}
	if _, err := ParseFallbackRootsText("/a\n/b\n/c\n/d"); err == nil {
		t.Fatal("超过 3 个备用目录应拒绝")
	}
	if _, err := ParseFallbackRootsText("/var/www/app-{v1,v2}"); err == nil {
		t.Fatal("备用目录不得使用占位符")
	}
}
