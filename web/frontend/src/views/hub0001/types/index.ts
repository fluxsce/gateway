/**
 * hub0001模块类型定义
 */

/**
 * 登录表单数据接口
 */
export interface LoginFormData {
  /** 用户ID */
  userId: string
  /** 密码 */
  password: string
  /** 验证码 */
  captchaCode: string
  /** 验证码ID */
  captchaId?: string
  /** 记住登录 */
  rememberMe: boolean
}

/**
 * 手机验证码登录表单数据接口
 */
export interface PhoneLoginFormData {
  /** 手机号码 */
  phone: string
  /** 验证码 */
  code: string
  /** 记住登录 */
  rememberMe: boolean
}

/**
 * 登录响应数据接口
 */
export interface LoginResponseData {
  /** 访问令牌 */
  token: string
  /** 刷新令牌 */
  refreshToken: string
  /** 用户信息 */
  user: any
  /** 权限列表 */
  permissions: string[]
}
