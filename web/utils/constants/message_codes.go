package constants

// 消息代码常量，用于标准化API响应中的messageId字段
// ED: Error代码 - 错误消息，格式ED00001开头
// SD: Success代码 - 成功消息，格式SD00001开头
// ND: Notice代码 - 提示消息，格式ND00001开头

// 通用错误代码
const (
	// 系统错误
	ED00001 = "ED00001" // 系统错误
	ED00002 = "ED00002" // 内部错误
	ED00003 = "ED00003" // 数据库错误
	ED00004 = "ED00004" // 网络错误
	ED00005 = "ED00005" // 无效的请求
	ED00006 = "ED00006" // 无效的参数
	ED00007 = "ED00007" // 缺少参数
	ED00008 = "ED00008" // 数据未找到
	ED00009 = "ED00009" // 操作失败
	ED00010 = "ED00010" // 权限拒绝
	ED00011 = "ED00011" // 未认证
	ED00012 = "ED00012" // 未授权
	ED00013 = "ED00013" // 记录已经存在
	ED00014 = "ED00014" // 验证失败
	ED00015 = "ED00015" // 业务约束错误
)

// 认证相关错误代码
const (
	ED00101 = "ED00101" // 登录失败
	ED00102 = "ED00102" // 用户不存在
	ED00103 = "ED00103" // 凭证无效
	ED00104 = "ED00104" // 用户已禁用
	ED00105 = "ED00105" // 用户已过期
	ED00106 = "ED00106" // 令牌无效
	ED00107 = "ED00107" // 令牌已过期
	ED00108 = "ED00108" // 刷新令牌失败
	ED00109 = "ED00109" // 密码错误
	ED00110 = "ED00110" // 密码修改失败
	ED00111 = "ED00111" // 验证码不存在或已过期
	ED00112 = "ED00112" // 验证码错误
	ED00113 = "ED00113" // 短信发送失败
	ED00114 = "ED00114" // Session不存在或已过期
	ED00115 = "ED00115" // Session已过期
)

// 通用成功代码
const (
	SD00001 = "SD00001" // 操作成功
	SD00002 = "SD00002" // 查询成功
	SD00003 = "SD00003" // 创建成功
	SD00004 = "SD00004" // 更新成功
	SD00005 = "SD00005" // 删除成功
)

// 认证相关成功代码
const (
	SD00101 = "SD00101" // 登录成功
	SD00102 = "SD00102" // 获取用户信息成功
	SD00103 = "SD00103" // 刷新令牌成功
	SD00104 = "SD00104" // 登出成功
	SD00105 = "SD00105" // 密码修改成功
	SD00106 = "SD00106" // 验证码生成成功
	SD00107 = "SD00107" // 验证码验证成功
)

// 提示代码
const (
	ND00001 = "ND00001" // 密码即将过期
	ND00002 = "ND00002" // 账号即将过期
	ND00003 = "ND00003" // 系统维护中
	ND00004 = "ND00004" // 版本更新
	ND00005 = "ND00005" // 需要数据同步
)

// HUB Session和Cookie相关常量
const (
	// Cookie名称常量
	HUB_SESSION_COOKIE = "HUB_LG" // Session ID的Cookie名称

	// Session配置常量
	HUB_SESSION_DOMAIN   = ""    // Cookie域名，空表示当前域名
	HUB_SESSION_PATH     = "/"   // Cookie路径
	HUB_SESSION_SECURE   = false // 是否仅HTTPS，生产环境应设为true
	HUB_SESSION_HTTPONLY = true  // 是否仅HTTP访问，防止XSS
	HUB_SESSION_SAMESITE = "Lax" // SameSite策略

	// Session超时时间配置（小时）
	HUB_SESSION_EXPIRE_HOURS = 12 // Session默认过期时间：12小时
)
