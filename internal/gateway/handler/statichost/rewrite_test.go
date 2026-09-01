package statichost

import "testing"

func TestApplyRewriteRulesFirstMatchWins(t *testing.T) {
	rules := []RewriteRule{
		{Mode: RewriteModePrefix, From: "/old", To: "/new"},
		{Mode: RewriteModePrefix, From: "/old", To: "/other"},
	}
	got := applyRewriteRules("/old/a.js", rules)
	if got != "/new/a.js" {
		t.Fatalf("应命中第一条规则，实际 %s", got)
	}
}

func TestApplyPrefixRewriteSegmentBoundary(t *testing.T) {
	got, ok := applyPrefixRewrite("/older/a.js", "/old", "/new")
	if ok {
		t.Fatalf("不同路径段不应命中，实际 %s", got)
	}

	got, ok = applyPrefixRewrite("/old/a.js", "/old", "/new")
	if !ok || got != "/new/a.js" {
		t.Fatalf("前缀重写应为 /new/a.js，实际 %s ok=%v", got, ok)
	}

	got, ok = applyPrefixRewrite("/old", "/old", "/new")
	if !ok || got != "/new" {
		t.Fatalf("精确前缀替换应为 /new，实际 %s ok=%v", got, ok)
	}
}

func TestApplyExactAndRegexRewrite(t *testing.T) {
	got, ok := applyRewriteRule("/favicon.ico", RewriteRule{
		Mode: RewriteModeExact,
		From: "/favicon.ico",
		To:   "/images/favicon.ico",
	})
	if !ok || got != "/images/favicon.ico" {
		t.Fatalf("精确重写失败: %s ok=%v", got, ok)
	}

	got, ok = applyRewriteRule("/A05SysBizWebVue/js/app.js", RewriteRule{
		Mode: RewriteModeExact,
		From: "A05SysBizWebVue",
		To:   "A05logWebVue",
	})
	if !ok || got != "/A05logWebVue/js/app.js" {
		t.Fatalf("精确应按字面字符串替换，实际 %s ok=%v", got, ok)
	}

	got, ok = applyRewriteRule("/static/A05SysBizWebVue/css/app.css", RewriteRule{
		Mode: RewriteModeExact,
		From: "A05SysBizWebVue",
		To:   "A05logWebVue",
	})
	if !ok || got != "/static/A05logWebVue/css/app.css" {
		t.Fatalf("中间路径段也应按字符替换，实际 %s ok=%v", got, ok)
	}

	got, ok = applyRewriteRule("/A05SysBizWebVueExtra/js/app.js", RewriteRule{
		Mode: RewriteModeExact,
		From: "A05SysBizWebVue",
		To:   "A05logWebVue",
	})
	if !ok || got != "/A05logWebVueExtra/js/app.js" {
		t.Fatalf("字面匹配会替换更长名字中的子串，实际 %s ok=%v", got, ok)
	}

	got, ok = applyRewriteRule("/other/js/app.js", RewriteRule{
		Mode: RewriteModeExact,
		From: "A05SysBizWebVue",
		To:   "A05logWebVue",
	})
	if ok {
		t.Fatalf("未出现的字符串不应改写，实际 %s", got)
	}

	got, ok = applyRewriteRule("/v2/css/app.css", RewriteRule{
		Mode: RewriteModeRegex,
		From: `^/v\d+/(.*)$`,
		To:   "/$1",
	})
	if !ok || got != "/css/app.css" {
		t.Fatalf("正则重写失败: %s ok=%v", got, ok)
	}

	got, ok = applyRewriteRule("/css/app.css", RewriteRule{
		Mode: RewriteModeRegex,
		From: `^/v\d+/(.*)$`,
		To:   "/$1",
	})
	if ok {
		t.Fatalf("未匹配的正则不应改写，实际 %s", got)
	}
}

func TestParseRewriteRulesTextLineFormats(t *testing.T) {
	got := ParseRewriteRulesText("prefix /old /new\n# skip\nexact /a /b\nregex ^/v/(.*)$ /$1\n/legacy => /current")
	if len(got) != 4 {
		t.Fatalf("应解析出 4 条规则，实际 %+v", got)
	}
	if got[0].To != "/new" || got[1].Mode != RewriteModeExact || got[2].To != "/$1" || got[3].From != "/legacy" {
		t.Fatalf("行格式解析不正确: %+v", got)
	}
}

func TestParseRewriteRulesTextStrictRejectsInvalid(t *testing.T) {
	if _, err := ParseRewriteRulesTextStrict("[not-json"); err == nil {
		t.Fatal("损坏的 JSON 应失败")
	}
	if _, err := ParseRewriteRulesTextStrict("regex-only-one-field"); err == nil {
		t.Fatal("无法识别的行应失败")
	}
	if _, err := compileRewriteRules([]RewriteRule{{Mode: RewriteModeRegex, From: "(unclosed", To: "/$1"}}); err == nil {
		t.Fatal("非法正则应在编译期失败")
	}
}

func TestNormalizeRewriteRulesDropsEmpty(t *testing.T) {
	got := normalizeRewriteRules([]RewriteRule{
		{From: "", To: "/x"},
		{From: "/a", To: "/b", Mode: "PREFIX"},
	})
	if len(got) != 1 || got[0].Mode != RewriteModePrefix {
		t.Fatalf("应只保留有效规则并规范化模式: %+v", got)
	}
}
