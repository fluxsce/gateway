/**
 * hub0001(登录)模块的中文本地化
 */
export default {
  login: {
    title: '账户登录',
    subtitle: '请使用您的账号密码登录系统',
    loginTitle: '账户登录',
    loginSubtitle: '请使用您的账号密码登录系统',
    rememberMe: '记住我',
    forgotPassword: '忘记密码？',
    loginButton: '登录',
    loginSuccess: '登录成功',
    loginFailed: '登录失败',
    networkError: '网络错误，请稍后重试',
    otherLoginMethods: '其他登录方式',
    testAccount: '测试账号：admin / 123456',
    logout: '退出登录',
    logoutSuccess: '您已成功退出登录',

    // 新增登录标签页文本
    accountLogin: '账号密码登录',
    phoneLogin: '手机验证码登录',

    // 用户ID和密码
    userId: '用户ID',
    username: '用户名',
    usernamePlaceholder: '请输入用户名',
    password: '密码',
    passwordPlaceholder: '请输入密码',

    // 验证码相关
    captcha: '验证码',
    captchaPlaceholder: '请输入图形验证码',

    // 手机验证码登录
    phoneNumber: '手机号码',
    phonePlaceholder: '请输入手机号码',
    verificationCode: '验证码',
    verificationCodePlaceholder: '请输入6位验证码',
    sendCode: '发送验证码',

    // 第三方登录
    orLoginWith: '或使用以下方式登录',

    // 服务条款
    termsText: '登录即表示您同意我们的<a href="#">服务条款</a>和<a href="#">隐私政策</a>',

    codeSent: '验证码已发送，请注意查收',
    codeSendFailed: '验证码发送失败，请稍后重试',
    wechatRedirect: '正在跳转到微信登录...',

    // 欢迎内容
    welcomeTitle: '欢迎使用FLUX Datahub 网关平台',
    welcomeSubtitle: '高效、安全、便捷的一体化管理平台',
    featureSecurityTitle: '安全可靠',
    featureSecurityDesc: '采用先进的加密技术，确保您的数据安全',
    featureAnalyticsTitle: '数据分析',
    featureAnalyticsDesc: '强大的分析工具，帮助您做出明智决策',
    featureCollaborationTitle: '协作共享',
    featureCollaborationDesc: '便捷的团队协作功能，提高工作效率',

    // 微信登录相关
    scanQrCode: '请扫描二维码登录',
  },
  welcome: {
    title: '欢迎使用FLUX Datahub 网关平台',
    subtitle: '高效、安全的数据管理解决方案',
    features: {
      secure: '安全可靠',
      dataAnalysis: '数据分析',
      multiTenant: '多租户支持',
    },
    copyright: '© {year} Web Hub - 版权所有',
  },
  forgotPassword: {
    title: '重置密码',
    subtitle: '请输入您的邮箱地址，我们将发送密码重置链接',
    emailPlaceholder: '您的电子邮箱',
    submit: '发送重置链接',
    backToLogin: '返回登录',
    checkEmail: '请检查您的邮箱获取重置指引',
    emailNotFound: '邮箱地址不存在',
    resetSuccess: '您的密码已成功重置',
    resetFailed: '密码重置失败',
  },
  validation: {
    userIdRequired: '请输入用户ID',
    userIdLength: '用户ID长度为3-20个字符',
    usernameRequired: '请输入用户名',
    usernameLength: '用户名长度为3-20个字符',
    passwordRequired: '请输入密码',
    passwordLength: '密码长度为6-32个字符',
    captchaRequired: '请输入验证码',
    captchaLength: '验证码长度为4个字符',
    phoneRequired: '请输入手机号码',
    phoneFormat: '请输入有效的手机号码',
    codeRequired: '请输入验证码',
    codeLength: '验证码长度为6位',
  },
  captcha: {
    title: '验证码',
    refresh: '点击刷新',
    placeholder: '请输入验证码',
  },
  social: {
    wechat: '微信登录',
    qq: 'QQ登录',
    github: 'GitHub登录',
  },
  dashboard: {
    title: '仪表盘',
    welcome: '欢迎回来，{name}',
    todayStats: '今日统计',
    logins: '登录次数',
    failedAttempts: '失败尝试',
    activeUsers: '活跃用户',
    avgSessionTime: '平均会话时长',
    recentActivity: '最近活动',
    activityTime: '时间',
    activityAction: '操作',
    activityIP: 'IP地址',
    activityLocation: '位置',
    viewAll: '查看全部',
  },
  logs: {
    title: '登录日志',
    subtitle: '查看所有登录活动',
    filters: {
      dateRange: '日期范围',
      searchPlaceholder: '搜索用户名、IP等',
      status: '状态',
      user: '用户',
    },
    table: {
      time: '时间',
      user: '用户',
      ip: 'IP地址',
      device: '设备',
      browser: '浏览器',
      status: '状态',
      location: '位置',
      actions: '操作',
    },
    status: {
      success: '成功',
      failed: '失败',
      locked: '锁定',
      suspicious: '可疑',
    },
    viewDetails: '查看详情',
    detailsTitle: '登录详情',
    noData: '没有找到登录记录',
    pagination: {
      prev: '上一页',
      next: '下一页',
      total: '共 {total} 条记录',
    },
  },
}
