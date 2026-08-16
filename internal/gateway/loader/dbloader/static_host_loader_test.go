package dbloader

import "testing"

func TestParseIndexFiles(t *testing.T) {
	jsonRaw := `["index.html","index.htm"]`
	got := parseIndexFiles(&jsonRaw)
	if len(got) != 2 || got[0] != "index.html" || got[1] != "index.htm" {
		t.Fatalf("JSON 解析结果 %+v", got)
	}

	csv := "index.html, default.html"
	got = parseIndexFiles(&csv)
	if len(got) != 2 || got[1] != "default.html" {
		t.Fatalf("逗号分隔解析结果 %+v", got)
	}

	unsafe := "index.html, ../secret"
	got = parseIndexFiles(&unsafe)
	if len(got) != 1 || got[0] != "index.html" {
		t.Fatalf("应丢弃带路径分隔的索引名，实际 %+v", got)
	}

	got = parseIndexFiles(nil)
	if len(got) != 1 || got[0] != "index.html" {
		t.Fatalf("空值应回退 index.html，实际 %+v", got)
	}
}

func TestParseRewriteRules(t *testing.T) {
	jsonRaw := `[{"mode":"prefix","from":"/old","to":"/new"}]`
	got := parseRewriteRules(&jsonRaw)
	if len(got) != 1 || got[0].From != "/old" || got[0].To != "/new" {
		t.Fatalf("JSON 规则解析失败: %+v", got)
	}

	lines := "prefix /old /new\nexact /favicon.ico /images/favicon.ico\n/legacy => /current"
	got = parseRewriteRules(&lines)
	if len(got) != 3 || got[1].Mode != "exact" || got[2].To != "/current" {
		t.Fatalf("逐行规则解析失败: %+v", got)
	}
}
