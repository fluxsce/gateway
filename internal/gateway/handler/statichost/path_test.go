package statichost

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStripRouteLookupPath(t *testing.T) {
	got := stripRouteLookupPath("/app/index.html", "/app", true)
	if got != "/index.html" {
		t.Fatalf("剥离前缀后应为 /index.html，实际 %s", got)
	}
	got = stripRouteLookupPath("/application/index.html", "/app", true)
	if got != "/application/index.html" {
		t.Fatalf("不同路径段不应剥离，实际 %s", got)
	}
}

func TestRoutePathPrefixFromRegex(t *testing.T) {
	got := routePathPrefix(`^/datahub01webVue/(d10|d12|d13)`)
	if got != "/datahub01webVue" {
		t.Fatalf("正则字面前缀应为 /datahub01webVue，实际 %s", got)
	}
	got = stripRouteLookupPath("/datahub01webVue/d10app/user/1", `^/datahub01webVue/(d10|d12)`, true)
	if got != "/d10app/user/1" {
		t.Fatalf("正则路由剥前缀后应为 /d10app/user/1，实际 %s", got)
	}
}

func TestStripKeepsDottedFilenameAfterLibPrefix(t *testing.T) {
	route := `^/lib/sce-vcom-dialogs(?:\.[\w-]+)?\.js$`
	got := stripRouteLookupPath("/lib/sce-vcom-dialogs.sec.js", route, true)
	if got != "/sce-vcom-dialogs.sec.js" {
		t.Fatalf("带点文件名应只剥 /lib，实际 %s", got)
	}
	got = stripRouteLookupPath("/lib/sce-vcom-dialogs.sec.js", "/lib/sce-vcom-dialogs.sec.js", true)
	if got != "/sce-vcom-dialogs.sec.js" {
		t.Fatalf("精确文件路由应保留文件名，实际 %s", got)
	}
	got = stripRouteLookupPath("/lib/sce-vcom-dialogs.js", route, true)
	if got != "/sce-vcom-dialogs.js" {
		t.Fatalf("无中间段时也应剥 /lib，实际 %s", got)
	}
}

func TestResolveLookupPathStripAndRewrite(t *testing.T) {
	got := resolveLookupPath("/app/index.html", "/app", true, nil)
	if got != "/index.html" {
		t.Fatalf("剥离前缀后应为 /index.html，实际 %s", got)
	}

	// /app 不能误剥 /application
	got = resolveLookupPath("/application/index.html", "/app", true, nil)
	if got != "/application/index.html" {
		t.Fatalf("不同路径段不应剥离，实际 %s", got)
	}

	rules, err := compileRewriteRules([]RewriteRule{
		{Mode: RewriteModePrefix, From: "/old", To: "/new"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got = resolveLookupPath("/app/old/a.js", "/app", true, rules)
	if got != "/new/a.js" {
		t.Fatalf("剥离后再按前缀重写应为 /new/a.js，实际 %s", got)
	}
}

func TestJoinUnderRootRejectsEscape(t *testing.T) {
	root := t.TempDir()
	// URL 中的 .. 会先被 Clean 掉，/../secret.txt 变成根下的 secret.txt，不能逃出目录。
	cleaned, err := joinUnderRoot(root, "/../secret.txt")
	if err != nil {
		t.Fatalf("Clean 后的相对路径不应报错: %v", err)
	}
	if !isUnderRoot(root, cleaned) {
		t.Fatalf("Clean 后的路径必须仍在根内: %s", cleaned)
	}

	inside, err := joinUnderRoot(root, "/css/app.css")
	if err != nil {
		t.Fatalf("合法相对路径不应失败: %v", err)
	}
	want := filepath.Join(root, "css", "app.css")
	if inside != want {
		t.Fatalf("拼接结果 %s，期望 %s", inside, want)
	}

	outside := filepath.Join(filepath.Dir(root), "outside.txt")
	if isUnderRoot(root, outside) {
		t.Fatal("根目录的兄弟路径不能算在根内")
	}
}

func TestHasHiddenPathComponent(t *testing.T) {
	if !hasHiddenPathComponent("/.git/config") {
		t.Fatal("应拒绝 .git")
	}
	if !hasHiddenPathComponent("/assets/.env") {
		t.Fatal("应拒绝隐藏文件")
	}
	if hasHiddenPathComponent("/.well-known/acme-challenge/x") {
		t.Fatal(".well-known 应放行")
	}
	if hasHiddenPathComponent("/index.html") {
		t.Fatal("普通文件不应视为隐藏路径")
	}
}

func TestHasFileExtension(t *testing.T) {
	if !hasFileExtension("/assets/app.js") {
		t.Fatal(".js 应视为带扩展名")
	}
	if hasFileExtension("/user/profile") {
		t.Fatal("SPA 路径不应视为带扩展名")
	}
}

func TestSpaFallbackLookups(t *testing.T) {
	got := spaFallbackLookups("/d10app/user/1", []string{"index.html"}, true)
	if len(got) != 2 || got[0] != "/d10app/index.html" || got[1] != "/index.html" {
		t.Fatalf("子应用 SPA 回退路径不正确: %+v", got)
	}
	got = spaFallbackLookups("/user/1", []string{"index.html"}, false)
	if len(got) != 1 || got[0] != "/index.html" {
		t.Fatalf("普通 SPA 只回退根索引: %+v", got)
	}
}

func TestIsUnderRoot(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if !isUnderRoot(root, child) {
		t.Fatal("子目录应判定在根内")
	}
	if isUnderRoot(root, root+"-outside") {
		t.Fatal("前缀相似的外部目录不能算在根内")
	}
}
