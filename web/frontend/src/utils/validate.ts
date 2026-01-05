/**
 * 验证工具类
 * 提供常用的数据验证函数，用于检查数据是否符合特定格式或规则
 */

/**
 * 验证手机号
 * 验证中国大陆手机号格式(1开头的11位数字)
 *
 * @param phone 手机号 - 待验证的手机号码
 * @returns true-有效的手机号，false-无效的手机号
 *
 * @example
 * isValidPhone('13800138000') // true
 * isValidPhone('1380013800') // false (长度不对)
 * isValidPhone('23800138000') // false (不是1开头)
 */
export const isValidPhone = (phone: string): boolean => {
  const reg = /^1[3-9]\d{9}$/
  return reg.test(phone)
}

/**
 * 验证邮箱
 * 验证邮箱格式，要求包含@符号和域名部分
 *
 * @param email 邮箱 - 待验证的邮箱地址
 * @returns true-有效的邮箱，false-无效的邮箱
 *
 * @example
 * isValidEmail('user@example.com') // true
 * isValidEmail('user@example') // false (域名格式不完整)
 * isValidEmail('user.example.com') // false (缺少@符号)
 */
export const isValidEmail = (email: string): boolean => {
  const reg = /^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$/
  return reg.test(email)
}

/**
 * 验证身份证号
 * 支持15位或18位身份证号码格式验证(简单格式验证，不包含校验码验证)
 *
 * @param idCard 身份证号 - 待验证的身份证号码
 * @returns true-有效的身份证号，false-无效的身份证号
 *
 * @example
 * isValidIdCard('11010119900101123X') // true
 * isValidIdCard('110101199001011') // true (15位)
 * isValidIdCard('1101011990010') // false (长度不对)
 */
export const isValidIdCard = (idCard: string): boolean => {
  const reg = /(^\d{15}$)|(^\d{18}$)|(^\d{17}(\d|X|x)$)/
  return reg.test(idCard)
}

/**
 * 验证URL
 * 验证是否为有效的URL地址，需要包含协议部分(http/https/ftp)
 *
 * @param url URL地址 - 待验证的网址
 * @returns true-有效的URL，false-无效的URL
 *
 * @example
 * isValidUrl('https://www.example.com') // true
 * isValidUrl('http://localhost:8080') // true
 * isValidUrl('www.example.com') // false (缺少协议)
 */
export const isValidUrl = (url: string): boolean => {
  const reg =
    /^(https?|ftp):\/\/([a-zA-Z0-9.-]+(:[a-zA-Z0-9.&%$-]+)*@)*((25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])){3}|([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.(com|edu|gov|int|mil|net|org|biz|arpa|info|name|pro|aero|coop|museum|[a-zA-Z]{2}))(:[0-9]+)*(\/($|[a-zA-Z0-9.,?'\\+&%$#=~_-]+))*$/
  return reg.test(url)
}

/**
 * 验证密码强度
 * 要求密码包含大小写字母和数字，长度8-16位
 *
 * @param password 密码 - 待验证的密码
 * @returns true-符合强密码规则，false-不符合强密码规则
 *
 * @example
 * isStrongPassword('Abc12345') // true
 * isStrongPassword('abc12345') // false (缺少大写字母)
 * isStrongPassword('ABC12345') // false (缺少小写字母)
 * isStrongPassword('Abcdefgh') // false (缺少数字)
 */
export const isStrongPassword = (password: string): boolean => {
  const reg = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,16}$/
  return reg.test(password)
}

/**
 * 检查是否为空值
 * 检查值是否为null、undefined、空字符串、空数组或空对象
 *
 * @param value 要检查的值 - 任意类型的值
 * @returns true-是空值，false-不是空值
 *
 * @example
 * isEmpty(null) // true
 * isEmpty('') // true
 * isEmpty([]) // true
 * isEmpty({}) // true
 * isEmpty('hello') // false
 * isEmpty([1, 2]) // false
 */
export const isEmpty = (value: unknown): boolean => {
  if (value === null || value === undefined) {
    return true
  }
  if (typeof value === 'string' && value.trim() === '') {
    return true
  }
  if (Array.isArray(value) && value.length === 0) {
    return true
  }
  if (typeof value === 'object' && Object.keys(value).length === 0) {
    return true
  }
  return false
}

/**
 * 验证IPv4地址
 * 验证是否为有效的IPv4地址格式(0.0.0.0 - 255.255.255.255)
 *
 * @param ip IPv4地址 - 待验证的IP地址
 * @returns true-有效的IPv4地址，false-无效的IPv4地址
 *
 * @example
 * isValidIPv4('192.168.1.1') // true
 * isValidIPv4('255.255.255.255') // true
 * isValidIPv4('0.0.0.0') // true
 * isValidIPv4('192.168.1.256') // false (超出范围)
 * isValidIPv4('192.168.1') // false (格式不完整)
 */
export const isValidIPv4 = (ip: string): boolean => {
  if (!ip || typeof ip !== 'string') return false

  const parts = ip.split('.')
  if (parts.length !== 4) return false

  return parts.every((part) => {
    // 检查是否为纯数字
    if (!/^\d+$/.test(part)) return false

    const num = parseInt(part, 10)

    // 检查范围 0-255
    if (num < 0 || num > 255) return false

    // 检查前导零(除了单独的'0')
    if (part.length > 1 && part[0] === '0') return false

    return true
  })
}

/**
 * 验证IPv6地址
 * 验证是否为有效的IPv6地址格式
 *
 * @param ip IPv6地址 - 待验证的IPv6地址
 * @returns true-有效的IPv6地址，false-无效的IPv6地址
 *
 * @example
 * isValidIPv6('2001:0db8:85a3:0000:0000:8a2e:0370:7334') // true
 * isValidIPv6('2001:db8:85a3::8a2e:370:7334') // true (压缩格式)
 * isValidIPv6('::1') // true (本地回环)
 * isValidIPv6('::') // true (全零地址)
 * isValidIPv6('2001:0db8:85a3::8a2e::7334') // false (多个::)
 */
export const isValidIPv6 = (ip: string): boolean => {
  if (!ip || typeof ip !== 'string') return false

  // IPv6正则表达式
  const ipv6Regex =
    /^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$/

  return ipv6Regex.test(ip)
}

/**
 * 验证IP地址(IPv4或IPv6)
 * 验证是否为有效的IP地址，支持IPv4和IPv6格式
 *
 * @param ip IP地址 - 待验证的IP地址
 * @returns true-有效的IP地址，false-无效的IP地址
 *
 * @example
 * isValidIP('192.168.1.1') // true (IPv4)
 * isValidIP('2001:db8::1') // true (IPv6)
 * isValidIP('invalid-ip') // false
 */
export const isValidIP = (ip: string): boolean => {
  return isValidIPv4(ip) || isValidIPv6(ip)
}

/**
 * 验证CIDR网段格式
 * 验证是否为有效的CIDR格式(IP/前缀长度)，支持IPv4和IPv6
 *
 * @param cidr CIDR网段 - 待验证的CIDR格式字符串
 * @returns true-有效的CIDR格式，false-无效的CIDR格式
 *
 * @example
 * isValidCIDR('192.168.1.0/24') // true (IPv4)
 * isValidCIDR('10.0.0.0/8') // true (IPv4)
 * isValidCIDR('2001:db8::/32') // true (IPv6)
 * isValidCIDR('192.168.1.0/33') // false (IPv4前缀长度超出范围)
 * isValidCIDR('192.168.1.0') // false (缺少前缀长度)
 */
export const isValidCIDR = (cidr: string): boolean => {
  if (!cidr || typeof cidr !== 'string') return false

  const parts = cidr.split('/')
  if (parts.length !== 2) return false

  const [ip, prefixStr] = parts
  const prefix = parseInt(prefixStr, 10)

  // 检查前缀长度是否为有效数字
  if (isNaN(prefix) || prefixStr !== prefix.toString()) return false

  // 验证IPv4 CIDR
  if (isValidIPv4(ip)) {
    return prefix >= 0 && prefix <= 32
  }

  // 验证IPv6 CIDR
  if (isValidIPv6(ip)) {
    return prefix >= 0 && prefix <= 128
  }

  return false
}

/**
 * 验证IP地址列表
 * 验证IP地址数组中的每个IP是否都有效
 *
 * @param ips IP地址数组 - 待验证的IP地址列表
 * @returns { valid: boolean, invalidIps: string[] } - 验证结果和无效的IP列表
 *
 * @example
 * validateIPList(['192.168.1.1', '10.0.0.1']) // { valid: true, invalidIps: [] }
 * validateIPList(['192.168.1.1', 'invalid-ip']) // { valid: false, invalidIps: ['invalid-ip'] }
 */
export const validateIPList = (ips: string[]): { valid: boolean; invalidIps: string[] } => {
  const invalidIps = ips.filter((ip) => !isValidIP(ip.trim()))
  return {
    valid: invalidIps.length === 0,
    invalidIps,
  }
}

/**
 * 验证CIDR网段列表
 * 验证CIDR数组中的每个网段是否都有效
 *
 * @param cidrs CIDR网段数组 - 待验证的CIDR列表
 * @returns { valid: boolean, invalidCidrs: string[] } - 验证结果和无效的CIDR列表
 *
 * @example
 * validateCIDRList(['192.168.1.0/24', '10.0.0.0/8']) // { valid: true, invalidCidrs: [] }
 * validateCIDRList(['192.168.1.0/24', '192.168.1.0/33']) // { valid: false, invalidCidrs: ['192.168.1.0/33'] }
 */
export const validateCIDRList = (cidrs: string[]): { valid: boolean; invalidCidrs: string[] } => {
  const invalidCidrs = cidrs.filter((cidr) => !isValidCIDR(cidr.trim()))
  return {
    valid: invalidCidrs.length === 0,
    invalidCidrs,
  }
}

/**
 * 验证正则表达式
 * 验证字符串是否为有效的正则表达式
 *
 * @param pattern 正则表达式字符串 - 待验证的正则表达式
 * @returns true-有效的正则表达式，false-无效的正则表达式
 *
 * @example
 * isValidRegex('Mozilla/.*Chrome.*') // true
 * isValidRegex('^test$') // true
 * isValidRegex('[') // false (无效的正则表达式)
 */
export const isValidRegex = (pattern: string): boolean => {
  if (!pattern || typeof pattern !== 'string') return false
  try {
    new RegExp(pattern)
    return true
  } catch {
    return false
  }
}

/**
 * 验证正则表达式列表
 * 验证正则表达式数组中的每个模式是否都有效
 *
 * @param patterns 正则表达式数组 - 待验证的正则表达式列表
 * @returns { valid: boolean, invalidPatterns: string[] } - 验证结果和无效的正则表达式列表
 *
 * @example
 * validateRegexList(['Mozilla/.*Chrome.*', '^test$']) // { valid: true, invalidPatterns: [] }
 * validateRegexList(['Mozilla/.*Chrome.*', '[']) // { valid: false, invalidPatterns: ['['] }
 */
export const validateRegexList = (patterns: string[]): { valid: boolean; invalidPatterns: string[] } => {
  const invalidPatterns = patterns.filter((pattern) => !isValidRegex(pattern.trim()))
  return {
    valid: invalidPatterns.length === 0,
    invalidPatterns,
  }
}
