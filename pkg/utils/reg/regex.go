package reg

import (
	"regexp"
	"strings"
)

// 常用正则表达式模式
const (
	// Email 邮箱地址
	PatternEmail = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`

	// URL URL地址
	PatternURL = `^https?://[^\s/$.?#].[^\s]*$`

	// IP地址（IPv4）
	PatternIPv4 = `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`

	// IP地址（IPv6，简化版本）
	PatternIPv6 = `^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$|^::1$|^::$`

	// 手机号（中国大陆）
	PatternPhoneCN = `^1[3-9]\d{9}$`

	// 身份证号（中国大陆18位）
	PatternIDCardCN = `^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`

	// 中文（使用 Unicode 属性类 \p{Han} 匹配中文字符）
	PatternChinese = `^\p{Han}+$`

	// 数字
	PatternNumber = `^\d+$`

	// 整数（包括负数）
	PatternInteger = `^-?\d+$`

	// 浮点数
	PatternFloat = `^-?\d+\.\d+$`

	// 字母和数字
	PatternAlphanumeric = `^[a-zA-Z0-9]+$`

	// 字母、数字和下划线
	PatternAlphanumericUnderscore = `^[a-zA-Z0-9_]+$`

	// 字母、数字、下划线和连字符
	PatternAlphanumericUnderscoreHyphen = `^[a-zA-Z0-9_\-]+$`

	// 日期格式 YYYY-MM-DD
	PatternDate = `^\d{4}-\d{2}-\d{2}$`

	// 时间格式 HH:MM:SS
	PatternTime = `^([01]\d|2[0-3]):[0-5]\d:[0-5]\d$`

	// 日期时间格式 YYYY-MM-DD HH:MM:SS
	PatternDateTime = `^\d{4}-\d{2}-\d{2} ([01]\d|2[0-3]):[0-5]\d:[0-5]\d$`

	// 端口号（1-65535）
	PatternPort = `^([1-9]\d{0,3}|[1-5]\d{4}|6[0-4]\d{3}|65[0-4]\d{2}|655[0-2]\d|6553[0-5])$`

	// 域名
	PatternDomain = `^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`

	// UUID
	PatternUUID = `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`

	// 十六进制字符串
	PatternHex = `^[0-9a-fA-F]+$`

	// Base64字符串
	PatternBase64 = `^[A-Za-z0-9+/]*={0,2}$`

	// 用户名（字母、数字、下划线，3-20个字符）
	PatternUsername = `^[a-zA-Z0-9_]{3,20}$`

	// 密码（至少8位，包含字母和数字）
	PatternPassword = `^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d@$!%*#?&]{8,}$`

	// 路径（Unix/Linux风格）
	PatternUnixPath = `^(/[^/ ]*)+/?$`

	// 路径（Windows风格）
	PatternWindowsPath = `^[a-zA-Z]:\\(?:[^\\/:*?"<>|\r\n]+\\)*[^\\/:*?"<>|\r\n]*$`
)

// ==================== 核心函数 ====================

// Compile 编译正则表达式（不缓存）
func Compile(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}

// Match 检查字符串是否匹配正则表达式
func Match(pattern, text string) (bool, error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(text), nil
}

// Find 查找第一个匹配的子字符串
func Find(pattern, text string) (string, error) {
	re, err := Compile(pattern)
	if err != nil {
		return "", err
	}
	return re.FindString(text), nil
}

// FindAll 查找所有匹配的子字符串
func FindAll(pattern, text string, n int) ([]string, error) {
	re, err := Compile(pattern)
	if err != nil {
		return nil, err
	}
	return re.FindAllString(text, n), nil
}

// Replace 替换匹配的字符串
func Replace(pattern, text, replacement string) (string, error) {
	re, err := Compile(pattern)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(text, replacement), nil
}

// ReplaceFunc 使用函数替换匹配的字符串
func ReplaceFunc(pattern, text string, replacer func(string) string) (string, error) {
	re, err := Compile(pattern)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllStringFunc(text, replacer), nil
}

// Split 使用正则表达式分割字符串
func Split(pattern, text string, n int) ([]string, error) {
	re, err := Compile(pattern)
	if err != nil {
		return nil, err
	}
	return re.Split(text, n), nil
}

// FindSubmatch 查找第一个匹配的子字符串及其子匹配
func FindSubmatch(pattern, text string) ([]string, error) {
	re, err := Compile(pattern)
	if err != nil {
		return nil, err
	}
	return re.FindStringSubmatch(text), nil
}

// FindAllSubmatch 查找所有匹配的子字符串及其子匹配
func FindAllSubmatch(pattern, text string, n int) ([][]string, error) {
	re, err := Compile(pattern)
	if err != nil {
		return nil, err
	}
	return re.FindAllStringSubmatch(text, n), nil
}

// ==================== 常用验证函数 ====================

// IsEmail 验证是否为邮箱地址
func IsEmail(email string) bool {
	matched, _ := Match(PatternEmail, email)
	return matched
}

// IsURL 验证是否为URL地址
func IsURL(url string) bool {
	matched, _ := Match(PatternURL, url)
	return matched
}

// IsIPv4 验证是否为IPv4地址
func IsIPv4(ip string) bool {
	matched, _ := Match(PatternIPv4, ip)
	return matched
}

// IsIPv6 验证是否为IPv6地址（简化版本）
func IsIPv6(ip string) bool {
	matched, _ := Match(PatternIPv6, ip)
	return matched
}

// IsIP 验证是否为IP地址（IPv4或IPv6）
func IsIP(ip string) bool {
	return IsIPv4(ip) || IsIPv6(ip)
}

// IsPhoneCN 验证是否为中国大陆手机号
func IsPhoneCN(phone string) bool {
	matched, _ := Match(PatternPhoneCN, phone)
	return matched
}

// IsIDCardCN 验证是否为中国大陆身份证号
func IsIDCardCN(idCard string) bool {
	matched, _ := Match(PatternIDCardCN, idCard)
	return matched
}

// IsChinese 验证是否全部为中文
func IsChinese(text string) bool {
	matched, _ := Match(PatternChinese, text)
	return matched
}

// IsNumber 验证是否为数字
func IsNumber(text string) bool {
	matched, _ := Match(PatternNumber, text)
	return matched
}

// IsInteger 验证是否为整数（包括负数）
func IsInteger(text string) bool {
	matched, _ := Match(PatternInteger, text)
	return matched
}

// IsFloat 验证是否为浮点数
func IsFloat(text string) bool {
	matched, _ := Match(PatternFloat, text)
	return matched
}

// IsAlphanumeric 验证是否为字母和数字
func IsAlphanumeric(text string) bool {
	matched, _ := Match(PatternAlphanumeric, text)
	return matched
}

// IsAlphanumericUnderscore 验证是否为字母、数字和下划线
func IsAlphanumericUnderscore(text string) bool {
	matched, _ := Match(PatternAlphanumericUnderscore, text)
	return matched
}

// IsAlphanumericUnderscoreHyphen 验证是否为字母、数字、下划线和连字符
func IsAlphanumericUnderscoreHyphen(text string) bool {
	matched, _ := Match(PatternAlphanumericUnderscoreHyphen, text)
	return matched
}

// IsDate 验证是否为日期格式 YYYY-MM-DD
func IsDate(date string) bool {
	matched, _ := Match(PatternDate, date)
	return matched
}

// IsTime 验证是否为时间格式 HH:MM:SS
func IsTime(time string) bool {
	matched, _ := Match(PatternTime, time)
	return matched
}

// IsDateTime 验证是否为日期时间格式 YYYY-MM-DD HH:MM:SS
func IsDateTime(dateTime string) bool {
	matched, _ := Match(PatternDateTime, dateTime)
	return matched
}

// IsPort 验证是否为端口号（1-65535）
func IsPort(port string) bool {
	matched, _ := Match(PatternPort, port)
	return matched
}

// IsDomain 验证是否为域名
func IsDomain(domain string) bool {
	matched, _ := Match(PatternDomain, domain)
	return matched
}

// IsUUID 验证是否为UUID
func IsUUID(uuid string) bool {
	matched, _ := Match(PatternUUID, uuid)
	return matched
}

// IsHex 验证是否为十六进制字符串
func IsHex(text string) bool {
	matched, _ := Match(PatternHex, text)
	return matched
}

// IsBase64 验证是否为Base64字符串
func IsBase64(text string) bool {
	matched, _ := Match(PatternBase64, text)
	return matched
}

// IsUsername 验证是否为用户名（字母、数字、下划线，3-20个字符）
func IsUsername(username string) bool {
	matched, _ := Match(PatternUsername, username)
	return matched
}

// IsPassword 验证是否为密码（至少8位，包含字母和数字）
func IsPassword(password string) bool {
	matched, _ := Match(PatternPassword, password)
	return matched
}

// IsUnixPath 验证是否为Unix/Linux路径
func IsUnixPath(path string) bool {
	matched, _ := Match(PatternUnixPath, path)
	return matched
}

// IsWindowsPath 验证是否为Windows路径
func IsWindowsPath(path string) bool {
	matched, _ := Match(PatternWindowsPath, path)
	return matched
}

// ==================== 提取函数 ====================

// ExtractEmail 从文本中提取所有邮箱地址
func ExtractEmail(text string) []string {
	// 移除 ^ 和 $ 锚点，以便在文本中查找匹配项
	pattern := `[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`
	emails, _ := FindAll(pattern, text, -1)
	return emails
}

// ExtractURL 从文本中提取所有URL地址
func ExtractURL(text string) []string {
	// 移除 ^ 和 $ 锚点，以便在文本中查找匹配项
	pattern := `https?://[^\s/$.?#].[^\s]*`
	urls, _ := FindAll(pattern, text, -1)
	return urls
}

// ExtractIPv4 从文本中提取所有IPv4地址
func ExtractIPv4(text string) []string {
	// 移除 ^ 和 $ 锚点，以便在文本中查找匹配项
	pattern := `((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`
	ips, _ := FindAll(pattern, text, -1)
	return ips
}

// ExtractPhoneCN 从文本中提取所有中国大陆手机号
func ExtractPhoneCN(text string) []string {
	// 移除 ^ 和 $ 锚点，以便在文本中查找匹配项
	pattern := `1[3-9]\d{9}`
	phones, _ := FindAll(pattern, text, -1)
	return phones
}

// ExtractNumber 从文本中提取所有数字
func ExtractNumber(text string) []string {
	// 移除 ^ 和 $ 锚点，以便在文本中查找匹配项
	pattern := `\d+`
	numbers, _ := FindAll(pattern, text, -1)
	return numbers
}

// ExtractChinese 从文本中提取所有中文字符串
func ExtractChinese(text string) []string {
	// 匹配连续的中文字符（使用 Unicode 属性类 \p{Han}）
	pattern := `\p{Han}+`
	chinese, _ := FindAll(pattern, text, -1)
	return chinese
}

// ==================== 清理和格式化函数 ====================

// RemoveWhitespace 移除所有空白字符
func RemoveWhitespace(text string) string {
	pattern := `\s+`
	result, _ := Replace(pattern, text, "")
	return result
}

// NormalizeWhitespace 规范化空白字符（多个空白字符替换为单个空格）
func NormalizeWhitespace(text string) string {
	pattern := `\s+`
	result, _ := Replace(pattern, text, " ")
	return strings.TrimSpace(result)
}

// RemoveSpecialChars 移除特殊字符（只保留字母、数字、中文、空格）
func RemoveSpecialChars(text string) string {
	// 使用 \p{Han} 匹配中文字符，或者使用 Unicode 范围 [\u4e00-\u9fa5]
	// 在 Go 的 regexp 中，需要使用 \p{Han} 来匹配中文字符
	pattern := `[^a-zA-Z0-9\p{Han}\s]`
	result, err := Replace(pattern, text, "")
	if err != nil {
		// 如果正则表达式编译失败，回退到不使用 Unicode 属性类的方式
		// 使用 Unicode 范围匹配中文
		pattern = `[^a-zA-Z0-9\s]`
		// 手动过滤中文字符
		var builder strings.Builder
		for _, r := range text {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') || r == ' ' || r == '\t' || r == '\n' ||
				(r >= 0x4e00 && r <= 0x9fa5) {
				builder.WriteRune(r)
			}
		}
		return builder.String()
	}
	return result
}

// RemoveHTMLTags 移除HTML标签
func RemoveHTMLTags(text string) string {
	pattern := `<[^>]*>`
	result, _ := Replace(pattern, text, "")
	return result
}

// RemoveEmoji 移除Emoji表情符号
func RemoveEmoji(text string) string {
	// 匹配大部分Emoji字符范围
	pattern := `[\x{1F300}-\x{1F9FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F600}-\x{1F64F}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]`
	result, _ := Replace(pattern, text, "")
	return result
}

// EscapeRegex 转义正则表达式特殊字符
func EscapeRegex(text string) string {
	specialChars := `\.+*?()|[]{}^$`
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, string(char), "\\"+string(char))
	}
	return result
}

// ==================== 验证和提取组合函数 ====================

// ValidateAndExtract 验证格式并提取匹配的部分
func ValidateAndExtract(pattern, text string) (bool, []string, error) {
	matched, err := Match(pattern, text)
	if err != nil {
		return false, nil, err
	}

	if !matched {
		return false, nil, nil
	}

	matches, err := FindAll(pattern, text, -1)
	if err != nil {
		return false, nil, err
	}

	return true, matches, nil
}

// ExtractGroups 提取正则表达式分组
func ExtractGroups(pattern, text string) (map[string]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	matches := re.FindStringSubmatch(text)
	if len(matches) == 0 {
		return nil, nil
	}

	groupNames := re.SubexpNames()
	result := make(map[string]string)

	for i, name := range groupNames {
		if i > 0 && name != "" && i < len(matches) {
			result[name] = matches[i]
		}
	}

	return result, nil
}
