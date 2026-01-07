package reg

import (
	"regexp"
	"testing"

	"gateway/pkg/utils/reg"

	"github.com/stretchr/testify/assert"
)

// ==================== 验证函数测试 ====================

func TestIsEmail(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有效邮箱1", "user@example.com", true},
		{"有效邮箱2", "test.email+tag@example.co.uk", true},
		{"无效邮箱1", "invalid.email", false},
		{"无效邮箱2", "@example.com", false},
		{"无效邮箱3", "user@", false},
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsEmail(tc.input)
			assert.Equal(t, tc.expected, result, "IsEmail(%q) = %v, expected %v", tc.input, result, tc.expected)
		})
	}
}

func TestIsURL(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有效URL1", "http://example.com", true},
		{"有效URL2", "https://www.example.com/path?query=1", true},
		{"无效URL1", "not-a-url", false},
		{"无效URL2", "ftp://example.com", false}, // 只支持 http/https
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsURL(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsIPv4(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有效IPv4-1", "192.168.1.1", true},
		{"有效IPv4-2", "10.0.0.1", true},
		{"有效IPv4-3", "255.255.255.255", true},
		{"无效IPv4-1", "256.1.1.1", false},
		{"无效IPv4-2", "192.168.1", false},
		{"无效IPv4-3", "192.168.1.1.1", false},
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsIPv4(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsPhoneCN(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有效手机号1", "13800138000", true},
		{"有效手机号2", "15912345678", true},
		{"无效手机号1", "12345678901", false}, // 不是1开头
		{"无效手机号2", "1380013800", false},  // 长度不对
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsPhoneCN(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsNumber(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"纯数字", "123456", true},
		{"单个数字", "5", true},
		{"包含负号", "-123", false}, // IsNumber 不支持负数
		{"包含小数点", "123.45", false},
		{"包含字母", "123abc", false},
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsNumber(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsInteger(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"正整数", "123", true},
		{"负整数", "-123", true},
		{"零", "0", true},
		{"包含小数点", "123.45", false},
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsInteger(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsFloat(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"正浮点数", "123.45", true},
		{"负浮点数", "-123.45", true},
		{"整数", "123", false}, // IsFloat 要求必须有小数点
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsFloat(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsChinese(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"纯中文", "你好世界", true},
		{"包含数字", "你好123", false},
		{"包含英文", "hello", false},
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsChinese(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsDate(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有效日期1", "2024-01-01", true},
		{"有效日期2", "2024-12-31", true},
		{"无效日期1", "2024/01/01", false}, // 格式不对
		{"无效日期2", "24-01-01", false},   // 年份不对
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsDate(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsTime(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有效时间1", "12:30:45", true},
		{"有效时间2", "23:59:59", true},
		{"无效时间1", "24:00:00", false}, // 小时超出范围
		{"无效时间2", "12:60:00", false}, // 分钟超出范围
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsTime(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsPort(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有效端口1", "80", true},
		{"有效端口2", "8080", true},
		{"有效端口3", "65535", true},
		{"无效端口1", "0", false},     // 端口不能为0
		{"无效端口2", "65536", false}, // 超出范围
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsPort(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUUID(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有效UUID1", "550e8400-e29b-41d4-a716-446655440000", true},
		{"有效UUID2", "00000000-0000-0000-0000-000000000000", true},
		{"无效UUID1", "550e8400-e29b-41d4-a716", false}, // 格式不完整
		{"无效UUID2", "not-a-uuid", false},
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsUUID(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUsername(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有效用户名1", "user123", true},
		{"有效用户名2", "test_user", true},
		{"有效用户名3", "abc", true},        // 最小长度3
		{"无效用户名1", "ab", false},        // 长度不足
		{"无效用户名2", "user-name", false}, // 包含连字符
		{"空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := reg.IsUsername(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ==================== 提取函数测试 ====================

func TestExtractEmail(t *testing.T) {
	text := "联系我：user1@example.com 或 user2@test.com"
	emails := reg.ExtractEmail(text)
	assert.Equal(t, 2, len(emails))
	assert.Contains(t, emails, "user1@example.com")
	assert.Contains(t, emails, "user2@test.com")
}

func TestExtractIPv4(t *testing.T) {
	text := "服务器地址：192.168.1.1 和 10.0.0.1"
	ips := reg.ExtractIPv4(text)
	assert.Equal(t, 2, len(ips))
	assert.Contains(t, ips, "192.168.1.1")
	assert.Contains(t, ips, "10.0.0.1")
}

func TestExtractPhoneCN(t *testing.T) {
	text := "联系电话：13800138000 或 15912345678"
	phones := reg.ExtractPhoneCN(text)
	assert.Equal(t, 2, len(phones))
	assert.Contains(t, phones, "13800138000")
	assert.Contains(t, phones, "15912345678")
}

func TestExtractNumber(t *testing.T) {
	text := "价格是123元，数量是456"
	numbers := reg.ExtractNumber(text)
	assert.GreaterOrEqual(t, len(numbers), 2)
	assert.Contains(t, numbers, "123")
	assert.Contains(t, numbers, "456")
}

func TestExtractChinese(t *testing.T) {
	text := "这是中文测试123abc"
	chinese := reg.ExtractChinese(text)
	assert.Greater(t, len(chinese), 0)
}

// ==================== 清理和格式化函数测试 ====================

func TestRemoveWhitespace(t *testing.T) {
	input := "hello   world\t\n"
	result := reg.RemoveWhitespace(input)
	assert.Equal(t, "helloworld", result)
}

func TestNormalizeWhitespace(t *testing.T) {
	input := "hello   world\t\n  test"
	result := reg.NormalizeWhitespace(input)
	assert.Equal(t, "hello world test", result)
}

func TestRemoveSpecialChars(t *testing.T) {
	input := "hello@world#123 测试"
	result := reg.RemoveSpecialChars(input)
	assert.Equal(t, "helloworld123 测试", result)
}

func TestRemoveHTMLTags(t *testing.T) {
	input := "<p>Hello</p><br/>World"
	result := reg.RemoveHTMLTags(input)
	assert.Equal(t, "HelloWorld", result)
}

func TestEscapeRegex(t *testing.T) {
	input := "test.()[]{}*+?^$|"
	result := reg.EscapeRegex(input)
	// 验证转义后的字符串可以安全用作正则表达式
	matched, err := reg.Match(result, input)
	assert.NoError(t, err)
	assert.True(t, matched)
}

// ==================== 核心函数测试 ====================

func TestCompile(t *testing.T) {
	// 测试编译有效正则
	re, err := reg.Compile(`\d+`)
	assert.NoError(t, err)
	assert.NotNil(t, re)

	// 测试编译无效正则
	_, err = reg.Compile(`[invalid`)
	assert.Error(t, err)
}

func TestRegexUtil_Match(t *testing.T) {
	matched, err := reg.Match(`^\d+$`, "123")
	assert.NoError(t, err)
	assert.True(t, matched)

	matched, err = reg.Match(`^\d+$`, "abc")
	assert.NoError(t, err)
	assert.False(t, matched)

	// 测试无效正则
	_, err = reg.Match(`[invalid`, "test")
	assert.Error(t, err)
}

func TestRegexUtil_Find(t *testing.T) {
	result, err := reg.Find(`\d+`, "价格是123元")
	assert.NoError(t, err)
	assert.Equal(t, "123", result)

	result, err = reg.Find(`\d+`, "没有数字")
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestRegexUtil_FindAll(t *testing.T) {
	results, err := reg.FindAll(`\d+`, "价格123数量456", -1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Contains(t, results, "123")
	assert.Contains(t, results, "456")
}

func TestRegexUtil_Replace(t *testing.T) {
	result, err := reg.Replace(`\d+`, "价格123", "XXX")
	assert.NoError(t, err)
	assert.Equal(t, "价格XXX", result)
}

func TestRegexUtil_ReplaceFunc(t *testing.T) {
	result, err := reg.ReplaceFunc(`\d+`, "价格123", func(s string) string {
		return "[" + s + "]"
	})
	assert.NoError(t, err)
	assert.Equal(t, "价格[123]", result)
}

func TestRegexUtil_Split(t *testing.T) {
	results, err := reg.Split(`\s+`, "hello   world   test", -1)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(results))
	assert.Equal(t, "hello", results[0])
	assert.Equal(t, "world", results[1])
	assert.Equal(t, "test", results[2])
}

func TestRegexUtil_FindSubmatch(t *testing.T) {
	results, err := reg.FindSubmatch(`(\d{4})-(\d{2})-(\d{2})`, "日期：2024-01-01")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 4) // 完整匹配 + 3个分组
	assert.Equal(t, "2024-01-01", results[0])
	assert.Equal(t, "2024", results[1])
	assert.Equal(t, "01", results[2])
	assert.Equal(t, "01", results[3])
}

// ==================== 便捷函数测试 ====================

func TestMatch(t *testing.T) {
	matched, err := reg.Match(`^\d+$`, "123")
	assert.NoError(t, err)
	assert.True(t, matched)

	matched, err = reg.Match(`[invalid`, "test")
	assert.Error(t, err)
	assert.False(t, matched)
}

func TestFind(t *testing.T) {
	result, err := reg.Find(`\d+`, "价格123")
	assert.NoError(t, err)
	assert.Equal(t, "123", result)
}

func TestReplace(t *testing.T) {
	result, err := reg.Replace(`\d+`, "价格123", "XXX")
	assert.NoError(t, err)
	assert.Equal(t, "价格XXX", result)
}

// ==================== 高级功能测试 ====================

func TestValidateAndExtract(t *testing.T) {
	valid, matches, err := reg.ValidateAndExtract(`\d+`, "价格123数量456")
	assert.NoError(t, err)
	assert.True(t, valid)
	assert.GreaterOrEqual(t, len(matches), 2)
}

func TestExtractGroups(t *testing.T) {
	pattern := `(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})`
	text := "日期：2024-01-01"

	groups, err := reg.ExtractGroups(pattern, text)
	assert.NoError(t, err)
	assert.NotNil(t, groups)
	assert.Equal(t, "2024", groups["year"])
	assert.Equal(t, "01", groups["month"])
	assert.Equal(t, "01", groups["day"])
}

// ==================== 边界情况测试 ====================

func TestEmptyString(t *testing.T) {
	// 测试空字符串的各种验证
	assert.False(t, reg.IsEmail(""))
	assert.False(t, reg.IsURL(""))
	assert.False(t, reg.IsIPv4(""))
	assert.False(t, reg.IsNumber(""))
}

func TestInvalidPattern(t *testing.T) {
	// 测试无效正则表达式不会导致panic
	_, err := reg.Match(`[invalid`, "test")
	assert.Error(t, err)

	_, err = reg.Find(`[invalid`, "test")
	assert.Error(t, err)
}

func TestConcurrentAccess(t *testing.T) {
	// 测试并发访问静态函数
	// 使用 goroutine 并发编译和匹配
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			matched, err := reg.Match(`\d+`, "123")
			assert.NoError(t, err)
			assert.True(t, matched)
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// ==================== warehouseId 字段模式匹配测试 ====================

func TestMatchWarehouseIdPattern(t *testing.T) {
	// pattern: 包含多个可能的 warehouseId 值（用 | 分隔）
	pattern := `"warehouseId":"9211_ZA12|9231_SA12|9231_SA13|9231_SA14|9231_SA15|9251_MX12|9251_MX13|9251_MX14|9251_MX15|9203_MY99|9203_MY11|9203_MY12|9242_ID07|9242_ID08|9242_ID09"`

	// text: 实际的 JSON 报文
	testCases := []struct {
		name        string
		text        string
		shouldMatch bool
		description string
	}{
		{
			name:        "匹配第一个值",
			text:        `{"warehouseId":"9211_ZA12"}`,
			shouldMatch: true,
			description: "应该匹配到 pattern 中的第一个值 9211_ZA12",
		},
		{
			name:        "匹配中间的值",
			text:        `{"warehouseId":"9231_SA13"}`,
			shouldMatch: true,
			description: "应该匹配到 pattern 中的中间值 9231_SA13",
		},
		{
			name:        "匹配最后一个值",
			text:        `{"warehouseId":"9242_ID09"}`,
			shouldMatch: true,
			description: "应该匹配到 pattern 中的最后一个值 9242_ID09",
		},
		{
			name:        "匹配多个值（用|分隔）",
			text:        `{"warehouseId":"9211_ZA12|9231_SA12"}`,
			shouldMatch: true,
			description: "应该匹配到 pattern 中的多个值组合",
		},
		{
			name:        "不匹配-值不在pattern中",
			text:        `{"warehouseId":"9999_XX99"}`,
			shouldMatch: false,
			description: "不应该匹配到不在 pattern 中的值",
		},
		{
			name:        "不匹配-字段名不同",
			text:        `{"otherField":"9211_ZA12"}`,
			shouldMatch: false,
			description: "不应该匹配到不同的字段名",
		},
		{
			name:        "匹配-完整JSON包含其他字段",
			text:        `{"warehouseId":"9211_ZA12","otherField":"value"}`,
			shouldMatch: true,
			description: "应该在包含其他字段的完整 JSON 中匹配",
		},
		{
			name:        "匹配-多个warehouseId值",
			text:        `{"warehouseId":"9211_ZA12|9231_SA12|9231_SA13|9242_ID07"}`,
			shouldMatch: true,
			description: "应该匹配到多个 warehouseId 值的组合",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matched, err := reg.Match(pattern, tc.text)
			assert.NoError(t, err)
			assert.Equal(t, tc.shouldMatch, matched, tc.description)
		})
	}
}

func TestMatchWarehouseIdPatternWithLiteralPipe(t *testing.T) {
	// 测试 pattern 中的 | 作为字面量匹配（需要转义）
	// 如果 pattern 中的 | 是字面量分隔符，需要使用 \| 转义
	// 但实际使用中，pattern 中的 | 通常是正则表达式的或运算符

	// 测试：如果 pattern 中的 | 是字面量，需要转义
	// 但根据用户需求，pattern 中的 | 应该是正则表达式的或运算符
	// 所以这个测试主要用于验证字面量匹配的场景

	// 使用 EscapeRegex 转义整个值部分
	valuePart := "9211_ZA12|9231_SA12|9231_SA13"
	escapedValue := reg.EscapeRegex(valuePart)
	pattern := `"warehouseId":"` + escapedValue + `"`

	// 测试匹配单个值（转义后的 pattern 应该匹配字面量）
	text := `{"xxxdata":{"header":[{"warehouseId":"13231312","customerId":"JETOUR","orderType":"SO","soReference1":"280705621","soReference2":"1147360","soReference3":"9203","soReference4":"MY11","orderTime":"2026-01-05 10:00:12","consigneeId":"7002947","consigneeName":"BH Autotech Sdn Bhd","hedi02":"B02","hedi03":"","udf01":"","udf02":"7002947","hedi05":"","consigneeAddress1":"No: 1 & 3, Jalan Basco Kepayang1, Basco Avenue @ KepayangIpoh","consigneeCountry":"MY","consigneeProvince":"PER","consigneeCity":"PERAK","consigneeDistrict":"Perak","consigneeStreet":"NO: 1 & 3, JALAN BASCO KE","consigneeMail":"","consigneeTel1":"","consigneeTel2":"","consigneeZip":"31400","hedi06":"","carrierId":"","carrierName":"","noteText":"WTY;The system automatically generates processing orders","details":[{"referenceNo":"000010","sku":"422015329AAJF","qtyOrdered":1.0,"lotAtt05":"9203","lotAtt06":"MY11","dedi04":"EA","lotAtt09":""}]}]}}`
	matched, err := regexp.MatchString(pattern, text)
	assert.NoError(t, err)
	assert.True(t, matched, "转义后的 pattern 应该能匹配完全相同的值")

	// 测试不匹配不同的值
	text2 := `{"warehouseId":"9211_ZA12"}`
	matched, err = reg.Match(pattern, text2)
	assert.NoError(t, err)
	assert.False(t, matched, "转义后的 pattern 不应该匹配部分值")
}

func TestMatchWarehouseIdPatternAsRegexOr(t *testing.T) {
	// 如果 pattern 中的 | 是正则表达式的或运算符
	// 那么 pattern 应该能匹配任意一个值
	pattern := `"warehouseId":"(9211_ZA12|9231_SA12|9231_SA13|9231_SA14|9231_SA15|9251_MX12|9251_MX13|9251_MX14|9251_MX15|9203_MY99|9203_MY11|9203_MY12|9242_ID07|9242_ID08|9242_ID09)"`

	testCases := []struct {
		name        string
		text        string
		shouldMatch bool
	}{
		{
			name:        "匹配单个值",
			text:        `{"warehouseId":"9211_ZA12"}`,
			shouldMatch: true,
		},
		{
			name:        "匹配另一个值",
			text:        `{"warehouseId":"9242_ID09"}`,
			shouldMatch: true,
		},
		{
			name:        "不匹配不在列表中的值",
			text:        `{"warehouseId":"9999_XX99"}`,
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matched, err := reg.Match(pattern, tc.text)
			assert.NoError(t, err)
			assert.Equal(t, tc.shouldMatch, matched,
				"pattern 作为正则表达式或运算符应该%s匹配",
				map[bool]string{true: "", false: "不"}[tc.shouldMatch])
		})
	}
}

func TestFindWarehouseIdFromPattern(t *testing.T) {
	// 测试从 JSON 文本中提取 warehouseId 值
	pattern := `"warehouseId":"([^"]*)"`

	testCases := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "提取单个值",
			text:     `{"warehouseId":"9211_ZA12"}`,
			expected: "9211_ZA12",
		},
		{
			name:     "提取多个值",
			text:     `{"warehouseId":"9211_ZA12|9231_SA12|9231_SA13"}`,
			expected: "9211_ZA12|9231_SA12|9231_SA13",
		},
		{
			name:     "提取完整值列表",
			text:     `{"warehouseId":"9211_ZA12|9231_SA12|9231_SA13|9231_SA14|9231_SA15|9251_MX12|9251_MX13|9251_MX14|9251_MX15|9203_MY99|9203_MY11|9203_MY12|9242_ID07|9242_ID08|9242_ID09"}`,
			expected: "9211_ZA12|9231_SA12|9231_SA13|9231_SA14|9231_SA15|9251_MX12|9251_MX13|9251_MX14|9251_MX15|9203_MY99|9203_MY11|9203_MY12|9242_ID07|9242_ID08|9242_ID09",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches, err := reg.FindSubmatch(pattern, tc.text)
			assert.NoError(t, err)
			assert.NotNil(t, matches)
			assert.GreaterOrEqual(t, len(matches), 2, "应该至少有两个匹配组")

			if len(matches) >= 2 {
				value := matches[1] // 第一个分组是字段值
				assert.Equal(t, tc.expected, value, "提取的值应该匹配")
			}
		})
	}
}
