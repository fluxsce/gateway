package statichost

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	// RewriteModePrefix 按路径段前缀替换，默认模式。
	RewriteModePrefix = "prefix"
	// RewriteModeExact 仅当路径完全一致时替换。
	RewriteModeExact = "exact"
	// RewriteModeRegex 正则替换，目标支持 $1 捕获组。
	RewriteModeRegex = "regex"
)

// RewriteRule 静态托管文件查找重写规则。
// 在剥离路由前缀之后、读文件之前按顺序匹配，命中第一条后停止。
// 不修改 Request.URL；访问日志与后续过滤器仍看到请求原路径（若过滤器未改过）。
type RewriteRule struct {
	// From 是匹配源：前缀、精确路径或正则。
	From string `json:"from" yaml:"from" mapstructure:"from"`
	// To 是替换目标；正则模式可使用 $1、$2。
	To string `json:"to" yaml:"to" mapstructure:"to"`
	// Mode 为 prefix、exact 或 regex，空值按 prefix 处理。
	Mode string `json:"mode,omitempty" yaml:"mode,omitempty" mapstructure:"mode,omitempty"`
}

// CompiledRewriteRule 加载期编译后的重写规则，请求路径只读使用。
type CompiledRewriteRule struct {
	Mode string
	From string
	To   string
	Re   *regexp.Regexp
}

// applyRewriteRules 按顺序应用未编译规则，供单测使用。
func applyRewriteRules(lookupPath string, rules []RewriteRule) string {
	compiled, err := compileRewriteRules(rules)
	if err != nil {
		return cleanURLPath(lookupPath)
	}
	return applyCompiledRewriteRules(lookupPath, compiled)
}

// applyCompiledRewriteRules 按顺序应用已编译规则，第一条命中后返回。
func applyCompiledRewriteRules(lookupPath string, rules []CompiledRewriteRule) string {
	current := cleanURLPath(lookupPath)
	for _, rule := range rules {
		next, matched := applyCompiledRewriteRule(current, rule)
		if matched {
			return next
		}
	}
	return current
}

func compileRewriteRules(rules []RewriteRule) ([]CompiledRewriteRule, error) {
	normalized := normalizeRewriteRules(rules)
	if len(normalized) == 0 {
		return nil, nil
	}
	result := make([]CompiledRewriteRule, 0, len(normalized))
	for i, rule := range normalized {
		compiled := CompiledRewriteRule{Mode: rule.Mode, From: rule.From, To: rule.To}
		if rule.Mode == RewriteModeRegex {
			re, err := regexp.Compile(rule.From)
			if err != nil {
				return nil, fmt.Errorf("rewrite rule %d: invalid regex %q: %w", i+1, rule.From, err)
			}
			compiled.Re = re
		}
		result = append(result, compiled)
	}
	return result, nil
}

// applyRewriteRule 应用单条规则；未命中时返回原路径和 false。
func applyRewriteRule(lookupPath string, rule RewriteRule) (string, bool) {
	from := strings.TrimSpace(rule.From)
	to := strings.TrimSpace(rule.To)
	if from == "" {
		return lookupPath, false
	}
	mode := strings.ToLower(strings.TrimSpace(rule.Mode))
	if mode == "" {
		mode = RewriteModePrefix
	}

	switch mode {
	case RewriteModeExact:
		if cleanURLPath(lookupPath) != cleanURLPath(from) {
			return lookupPath, false
		}
		return cleanURLPath(to), true
	case RewriteModeRegex:
		compiled, err := compileRewriteRules([]RewriteRule{rule})
		if err != nil || len(compiled) == 0 {
			return lookupPath, false
		}
		return applyCompiledRewriteRule(lookupPath, compiled[0])
	default:
		return applyPrefixRewrite(lookupPath, from, to)
	}
}

// applyCompiledRewriteRule 应用单条已编译规则；未命中时返回原路径和 false。
func applyCompiledRewriteRule(lookupPath string, rule CompiledRewriteRule) (string, bool) {
	switch rule.Mode {
	case RewriteModeExact:
		if cleanURLPath(lookupPath) != cleanURLPath(rule.From) {
			return lookupPath, false
		}
		return cleanURLPath(rule.To), true
	case RewriteModeRegex:
		if rule.Re == nil || !rule.Re.MatchString(lookupPath) {
			return lookupPath, false
		}
		return cleanURLPath(rule.Re.ReplaceAllString(lookupPath, rule.To)), true
	default:
		return applyPrefixRewrite(lookupPath, rule.From, rule.To)
	}
}

// applyPrefixRewrite 按路径段边界替换前缀，避免 /old 误伤 /older。
func applyPrefixRewrite(lookupPath, from, to string) (string, bool) {
	from = cleanURLPath(from)
	if from == "/" {
		// 根前缀会匹配所有路径，相当于给查找路径加上目标前缀。
		remaining := strings.TrimPrefix(cleanURLPath(lookupPath), "/")
		joined := strings.TrimSuffix(cleanURLPath(to), "/")
		if remaining == "" {
			return cleanURLPath(joined), true
		}
		if joined == "/" || joined == "" {
			return cleanURLPath("/" + remaining), true
		}
		return cleanURLPath(joined + "/" + remaining), true
	}
	base := strings.TrimSuffix(from, "/")
	if !hasPathPrefix(base, lookupPath) {
		return lookupPath, false
	}
	remaining := strings.TrimPrefix(lookupPath, base)
	if remaining == "" {
		remaining = "/"
	}
	if !strings.HasPrefix(remaining, "/") {
		remaining = "/" + remaining
	}
	target := strings.TrimSuffix(cleanURLPath(to), "/")
	if target == "" {
		target = "/"
	}
	if remaining == "/" {
		return target, true
	}
	if target == "/" {
		return remaining, true
	}
	return cleanURLPath(target + remaining), true
}

// ParseRewriteRulesText 解析 JSON 数组或逐行文本为重写规则。
// 行格式：`prefix /old /new`、`exact /a /b`、`regex ^/v/(.*)$ /$1`，或 `/old => /new`。
func ParseRewriteRulesText(raw string) []RewriteRule {
	rules, _ := parseRewriteRulesText(raw, false)
	return rules
}

// ParseRewriteRulesTextStrict 解析重写规则，非法 JSON、无法识别的行或非法正则均返回错误。
func ParseRewriteRulesTextStrict(raw string) ([]RewriteRule, error) {
	return parseRewriteRulesText(raw, true)
}

func parseRewriteRulesText(raw string, strict bool) ([]RewriteRule, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil, nil
	}
	var fromJSON []RewriteRule
	if err := json.Unmarshal([]byte(text), &fromJSON); err == nil {
		return normalizeRewriteRules(fromJSON), nil
	} else if strict && (strings.HasPrefix(text, "[") || strings.HasPrefix(text, "{")) {
		return nil, fmt.Errorf("rewrite rules json: %w", err)
	}
	lines := strings.Split(text, "\n")
	rules := make([]RewriteRule, 0, len(lines))
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		rule, ok := parseRewriteRuleLine(line)
		if !ok {
			if strict {
				return nil, fmt.Errorf("rewrite rule line %d is invalid", i+1)
			}
			continue
		}
		rules = append(rules, rule)
	}
	return normalizeRewriteRules(rules), nil
}

// FormatRewriteRulesText 把规则格式化为管理端可编辑的逐行文本。
func FormatRewriteRulesText(rules []RewriteRule) string {
	rules = normalizeRewriteRules(rules)
	if len(rules) == 0 {
		return ""
	}
	lines := make([]string, 0, len(rules))
	for _, rule := range rules {
		lines = append(lines, rule.Mode+" "+rule.From+" "+rule.To)
	}
	return strings.Join(lines, "\n")
}

// EncodeRewriteRulesJSON 把规则编码为入库 JSON，空列表写空字符串。
func EncodeRewriteRulesJSON(rules []RewriteRule) string {
	rules = normalizeRewriteRules(rules)
	if len(rules) == 0 {
		return ""
	}
	encoded, err := json.Marshal(rules)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func parseRewriteRuleLine(line string) (RewriteRule, bool) {
	text := strings.TrimSpace(line)
	if text == "" || strings.HasPrefix(text, "#") {
		return RewriteRule{}, false
	}
	if strings.Contains(text, "=>") {
		parts := strings.SplitN(text, "=>", 2)
		left := strings.TrimSpace(parts[0])
		to := ""
		if len(parts) > 1 {
			to = strings.TrimSpace(parts[1])
		}
		fields := strings.Fields(left)
		if len(fields) == 0 {
			return RewriteRule{}, false
		}
		mode := RewriteModePrefix
		from := fields[0]
		if isRewriteMode(fields[0]) {
			if len(fields) < 2 {
				return RewriteRule{}, false
			}
			mode = strings.ToLower(fields[0])
			from = strings.Join(fields[1:], " ")
		}
		return RewriteRule{Mode: mode, From: from, To: to}, true
	}
	fields := strings.Fields(text)
	if len(fields) < 2 {
		return RewriteRule{}, false
	}
	if isRewriteMode(fields[0]) {
		if len(fields) < 3 {
			return RewriteRule{}, false
		}
		return RewriteRule{
			Mode: strings.ToLower(fields[0]),
			From: fields[1],
			To:   strings.Join(fields[2:], " "),
		}, true
	}
	return RewriteRule{
		Mode: RewriteModePrefix,
		From: fields[0],
		To:   strings.Join(fields[1:], " "),
	}, true
}

func isRewriteMode(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case RewriteModePrefix, RewriteModeExact, RewriteModeRegex:
		return true
	default:
		return false
	}
}

// normalizeRewriteRules 去掉空规则并规范化模式名。
func normalizeRewriteRules(rules []RewriteRule) []RewriteRule {
	if len(rules) == 0 {
		return nil
	}
	result := make([]RewriteRule, 0, len(rules))
	for _, rule := range rules {
		from := strings.TrimSpace(rule.From)
		to := strings.TrimSpace(rule.To)
		if from == "" {
			continue
		}
		mode := strings.ToLower(strings.TrimSpace(rule.Mode))
		switch mode {
		case RewriteModeExact, RewriteModeRegex, RewriteModePrefix:
		case "":
			mode = RewriteModePrefix
		default:
			mode = RewriteModePrefix
		}
		result = append(result, RewriteRule{From: from, To: to, Mode: mode})
	}
	return result
}
