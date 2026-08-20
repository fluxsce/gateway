/**
 * 用户相关API的Mock数据
 * 模拟用户登录、注册、获取用户信息等接口
 * 注意：字段命名符合数据库命名规范，布尔型使用Flag后缀，状态使用适当的值类型表示
 */
import type { MockMethod } from 'vite-plugin-mock'
import type { JsonDataObj } from '@/types/api'
import { moduleApiPrefix, requestPathHelper } from '../../api/requestPath'
import Mock from 'mockjs'

/**
 * 请求处理函数参数接口
 */
interface RequestParams {
  url: string
  body: Record<string, any>
  query: Record<string, string>
  headers: Record<string, string>
  method: string
}

// 模拟用户数据
const users = [
  {
    userId: '1',
    userName: 'admin',
    password: '123456',
    realName: '管理员',
    avatar: 'https://avatars.githubusercontent.com/u/10000000',
    email: 'admin@example.com',
    mobile: '13800138000',
    deptId: '1',
    deptName: '系统管理部',
    roles: ['admin'],
    tenantAdminFlag: 'Y',
    deptAdminFlag: 'Y',
    activeFlag: 'Y',
    lastLoginTime: '2023-01-01 00:00:00',
    addTime: '2023-01-01 00:00:00',
    editTime: '2023-01-01 00:00:00',
    statusFlag: 'Y',
  },
  {
    userId: '2',
    userName: 'user',
    password: '123456',
    realName: '普通用户',
    avatar: 'https://avatars.githubusercontent.com/u/20000000',
    email: 'user@example.com',
    mobile: '13900139000',
    deptId: '2',
    deptName: '研发部',
    roles: ['user'],
    tenantAdminFlag: 'N',
    deptAdminFlag: 'N',
    activeFlag: 'Y',
    lastLoginTime: '2023-01-02 00:00:00',
    addTime: '2023-01-02 00:00:00',
    editTime: '2023-01-02 00:00:00',
    statusFlag: 'Y',
  },
]

// 存储生成的验证码数据
const captchaStore = new Map<string, string>()

/**
 * 创建JsonDataObj格式的响应数据
 */
function createJsonDataResponse<T>(data: T, success = true, message = ''): JsonDataObj {
  if (success) {
    return {
      oK: true,
      state: true,
      bizData: JSON.stringify(data),
      extObj: null,
      pageQueryData: '',
      messageId: '',
      errMsg: '',
      popMsg: message || '操作成功',
      extMsg: '',
      pkey1: '',
      pkey2: '',
      pkey3: '',
      pkey4: '',
      pkey5: '',
      pkey6: '',
    }
  } else {
    return {
      oK: false,
      state: false,
      bizData: '',
      extObj: null,
      pageQueryData: '',
      messageId: '',
      errMsg: message || '操作失败',
      popMsg: message || '操作失败',
      extMsg: '',
      pkey1: '',
      pkey2: '',
      pkey3: '',
      pkey4: '',
      pkey5: '',
      pkey6: '',
    }
  }
}

export default [
  // 验证码生成接口
  {
    url: requestPathHelper.join(moduleApiPrefix('user'), 'captcha'),
    method: 'post',
    response: () => {
      // 生成随机6位验证码（仅 mock 服务端持有，不返回给前端）
      const captcha = Mock.Random.string('0123456789', 6)
      // 生成验证码ID
      const captchaId = `captcha-${Date.now()}-${Mock.Random.string('abcdef0123456789', 8)}`

      // 存储验证码
      captchaStore.set(captchaId, captcha)

      // 30秒后自动删除验证码
      setTimeout(() => {
        captchaStore.delete(captchaId)
      }, 30000)

      // 1x1 PNG，不含答案文本
      const image =
        'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=='

      return createJsonDataResponse(
        {
          captchaId,
          image,
          expireAt: Date.now() + 120000,
        },
        true,
        '获取验证码成功',
      )
    },
  },

  // 用户登录
  {
    url: requestPathHelper.join(moduleApiPrefix('user'), 'login'),
    method: 'post',
    response: ({ body }: Pick<RequestParams, 'body'>) => {
      const { userId, password, captchaCode, captchaId } = body

      if (!captchaId || !captchaCode) {
        return createJsonDataResponse(null, false, '验证码不能为空')
      }

      const storedCaptcha = captchaStore.get(captchaId)
      if (!storedCaptcha) {
        return createJsonDataResponse(null, false, '验证码已过期')
      }

      if (storedCaptcha !== captchaCode) {
        return createJsonDataResponse(null, false, '验证码错误')
      }

      captchaStore.delete(captchaId)

      // 支持通过userId或userName查找用户
      const user = users.find(
        (item) =>
          (item.userId === userId || item.userName === userId) && item.password === password,
      )

      if (user) {
        // 成功登录，直接返回用户对象（不包含密码）
        return createJsonDataResponse(
          {
            ...user,
            password: undefined, // 不返回密码
          },
          true,
          '登录成功',
        )
      }

      return createJsonDataResponse(null, false, '用户名或密码错误')
    },
  },
] as MockMethod[]
