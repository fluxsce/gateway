/**
 * hub0001模块API接口
 */
import { request } from '@/api/request'
import { moduleApiPrefix, requestPathHelper } from '@/api/requestPath'
import type { JsonDataObj } from '@/types/api'
import type { LoginFormData } from '../types'

const userApiPrefix = moduleApiPrefix('user')

/**
 * hub0001模块API
 */
export const hub0001Api = {
  /**
   * 用户登录
   * @param data 登录表单数据
   * @returns 返回登录响应数据，包含token和用户信息
   */
  login(data: LoginFormData): Promise<JsonDataObj> {
    return request({
      url: requestPathHelper.join(userApiPrefix, 'login'),
      method: 'POST',
      data,
    })
  },

  /**
   * 获取验证码
   * @returns 返回签名票 captchaId 与 PNG Data URI image，不含答案
   */
  getCaptcha(): Promise<JsonDataObj> {
    return request({
      url: requestPathHelper.join(userApiPrefix, 'captcha'),
      method: 'POST',
      params: {
        t: new Date().getTime(),
      },
    })
  },

  /**
   * 获取系统版本信息
   * @returns 返回系统版本信息，包含版本号和应用名称
   */
  getVersion(): Promise<JsonDataObj> {
    return request({
      url: requestPathHelper.join(userApiPrefix, 'version'),
      method: 'GET',
    })
  },

  /**
   * 当前登录用户修改密码。身份由 session 决定，不必传 userId。
   */
  changePassword(data: { oldPassword: string; newPassword: string }): Promise<JsonDataObj> {
    return request({
      url: requestPathHelper.join(userApiPrefix, 'password'),
      method: 'PUT',
      data,
    })
  },

  /**
   * 读取当前登录用户资料。身份由 session 决定，不必传 userId。
   */
  getProfile(): Promise<JsonDataObj> {
    return request({
      url: requestPathHelper.join(userApiPrefix, 'profile'),
      method: 'GET',
    })
  },

  /**
   * 更新当前登录用户资料。身份由 session 决定，请求里的 userId 会被忽略。
   */
  updateProfile(data: {
    realName: string
    email?: string
    mobile?: string
    gender?: number
    avatar?: string
  }): Promise<JsonDataObj> {
    return request({
      url: requestPathHelper.join(userApiPrefix, 'profile'),
      method: 'PUT',
      data,
    })
  },
}
