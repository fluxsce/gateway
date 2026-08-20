/**
 * hub0001模块API接口
 */
import { request } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { LoginFormData } from '../types'

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
      url: '/gateway/user/login',
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
      url: '/gateway/user/captcha',
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
      url: '/gateway/user/version',
      method: 'GET',
    })
  },
}
