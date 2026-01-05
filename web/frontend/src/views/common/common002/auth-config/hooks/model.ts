/**
 * 认证配置 Model
 * 统一管理表单配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import { NDynamicTags, NInput, NInputNumber, NSelect, NSwitch } from 'naive-ui'
import { h, ref } from 'vue'

/**
 * 认证配置 Model
 */
export function useAuthConfigModel() {
  // ============= 数据状态 =============
  const moduleId = 'hubcommon002-auth'
  /** 加载状态 */
  const loading = ref(false)

  // ============= 表单配置 =============

  // 创建数组字段渲染函数
  const createArrayFieldRender = (field: string, placeholder: string) => {
    return (formData: Record<string, any>) => {
      const value = formData[field] || []

      return h(NDynamicTags, {
        value,
        'onUpdate:value': (newValue: string[]) => {
          formData[field] = newValue
        },
        inputProps: {
          placeholder,
        },
      })
    }
  }

  // 认证类型选项
  const authTypeOptions = [
    { label: 'JWT认证', value: 'JWT' },
    { label: 'API Key认证', value: 'API_KEY' },
    { label: 'OAuth2认证', value: 'OAUTH2' },
    { label: 'Basic认证', value: 'BASIC' },
  ]

  // 认证策略选项
  const authStrategyOptions = [
    { label: '必需认证', value: 'REQUIRED' },
    { label: '可选认证', value: 'OPTIONAL' },
    { label: '禁用认证', value: 'DISABLED' },
  ]

  // JWT算法选项
  const jwtAlgorithmOptions = [
    { label: 'HS256', value: 'HS256' },
    { label: 'HS384', value: 'HS384' },
    { label: 'HS512', value: 'HS512' },
    { label: 'RS256', value: 'RS256' },
    { label: 'RS384', value: 'RS384' },
    { label: 'RS512', value: 'RS512' },
  ]

  // API Key位置选项
  const apiKeyLocationOptions = [
    { label: 'Header', value: 'header' },
    { label: 'Query参数', value: 'query' },
  ]

  // 创建嵌套字段的 render 函数（用于访问 authConfig 对象的嵌套属性）
  // 使用点号分隔的字段名（如 authConfig.secret），更符合对象属性访问方式
  const createNestedFieldRender = (subField: string, fieldType: 'input' | 'select' | 'number' | 'switch' | 'tags', options?: any) => {
    return (formData: Record<string, any>) => {
      // 从点号分隔字段读取值（如 authConfig.secret）
      const dotKey = `authConfig.${subField}`
      const value = formData[dotKey]

      if (fieldType === 'input') {
        return h(NInput, {
          value: value || '',
          type: options?.type || 'text',
          placeholder: options?.placeholder || '',
          'onUpdate:value': (val: string) => {
            formData[dotKey] = val
          },
        })
      } else if (fieldType === 'select') {
        return h(NSelect, {
          value: value || options?.defaultValue || '',
          options: options?.options || [],
          placeholder: options?.placeholder || '',
          'onUpdate:value': (val: string) => {
            formData[dotKey] = val
          },
        })
      } else if (fieldType === 'number') {
        return h(NInputNumber, {
          value: value ?? options?.defaultValue ?? 0,
          min: options?.min,
          max: options?.max,
          placeholder: options?.placeholder || '',
          style: 'width: 100%;',
          'onUpdate:value': (val: number | null) => {
            formData[dotKey] = val ?? options?.defaultValue ?? 0
          },
        })
      } else if (fieldType === 'switch') {
        return h(NSwitch, {
          value: value ?? options?.defaultValue ?? false,
          'onUpdate:value': (val: boolean) => {
            formData[dotKey] = val
          },
        })
      } else if (fieldType === 'tags') {
        return h(NDynamicTags, {
          value: value || [],
          'onUpdate:value': (val: string[]) => {
            formData[dotKey] = val
          },
          inputProps: {
            placeholder: options?.placeholder || '',
          },
        })
      }

      // 默认返回一个空的 div（不应该到达这里）
      return h('div')
    }
  }

  // ============= 表单页签配置 =============
  const formTabs = [
    { key: 'basic', label: '基础设置' },
    { key: 'other', label: '其他' },
  ]

  /** 表单字段配置 */
  const formFields: DataFormField[] = [
    // ============= 主键字段（隐藏，但必须存在用于更新） =============
    {
      field: 'authConfigId',
      label: '认证配置ID',
      type: 'input',
      span: 8,
      show: false,
      tabKey: 'basic',
    },
    {
      field: 'tenantId',
      label: '租户ID',
      type: 'input',
      span: 8,
      show: false,
      tabKey: 'basic',
    },
    {
      field: 'gatewayInstanceId',
      label: '网关实例ID',
      type: 'input',
      span: 8,
      show: false,
      tabKey: 'basic',
    },
    {
      field: 'routeConfigId',
      label: '路由配置ID',
      type: 'input',
      span: 8,
      show: false,
      tabKey: 'basic',
    },
    // ============= 基础配置分组 =============
    {
      field: 'basicConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      tabKey: 'basic',
      children: [
        {
          field: 'authName',
          label: '认证配置名称',
          type: 'input',
          placeholder: '请输入认证配置名称',
          span: 12,
          required: true,
          defaultValue: '认证配置',
          tabKey: 'basic',
        },
        {
          field: 'authType',
          label: '认证类型',
          type: 'select',
          span: 12,
          required: true,
          defaultValue: 'JWT',
          tabKey: 'basic',
          props: {
            options: authTypeOptions,
            placeholder: '选择认证类型',
          },
          tips: '支持JWT、API Key、OAuth2、Basic认证',
        },
        {
          field: 'authStrategy',
          label: '认证策略',
          type: 'select',
          span: 12,
          required: true,
          defaultValue: 'REQUIRED',
          tabKey: 'basic',
          props: {
            options: authStrategyOptions,
            placeholder: '选择认证策略',
          },
          tips: '必需：必须认证才能访问；可选：可以认证但不强制；禁用：不进行认证',
        },
        {
          field: 'configPriority',
          label: '配置优先级',
          type: 'number',
          placeholder: '数值越小优先级越高',
          span: 12,
          defaultValue: 1,
          show: false, // 隐藏优先级字段
          tabKey: 'basic',
          props: {
            min: 0,
            max: 999,
          },
          tips: '数值越小优先级越高，相同条件下优先级高的配置会优先生效',
        },
        {
          field: 'failureStatusCode',
          label: '失败状态码',
          type: 'number',
          placeholder: '认证失败HTTP状态码',
          span: 12,
          defaultValue: 401,
          tabKey: 'basic',
          props: {
            min: 400,
            max: 599,
          },
          tips: '认证失败时返回的HTTP状态码',
        },
        {
          field: 'failureMessage',
          label: '失败提示消息',
          type: 'input',
          placeholder: '认证失败提示消息',
          span: 12,
          defaultValue: '认证失败',
          tabKey: 'basic',
          tips: '认证失败时返回的错误消息',
        },
      ],
    },
    // ============= 认证参数配置分组 =============
    {
      field: 'authConfigDetails',
      label: '',
      type: 'fieldset',
      span: 24,
      tabKey: 'basic',
      children: [
        // ============= JWT 认证配置字段 =============
        {
          field: 'authConfig.secret',
          label: 'JWT密钥',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'JWT',
          required: true,
          tips: 'JWT签名密钥，用于验证和生成Token',
          render: createNestedFieldRender('secret', 'input', { type: 'password', placeholder: 'JWT签名密钥' }),
          rules: [
            {
              validator: (rule: any, value: any) => {
                const formData = (rule as any).source || {}
                if (formData.authType === 'JWT') {
                  const secret = formData['authConfig.secret']
                  if (!secret || (typeof secret === 'string' && secret.trim() === '')) {
                    return new Error('JWT密钥不能为空')
                  }
                }
                return true
              },
              trigger: ['blur', 'change']
            }
          ],
        },
        {
          field: 'authConfig.algorithm',
          label: '签名算法',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'JWT',
          required: true,
          tips: 'JWT签名算法，如HS256、RS256等',
          render: createNestedFieldRender('algorithm', 'select', { options: jwtAlgorithmOptions, defaultValue: 'HS256', placeholder: '选择签名算法' }),
          rules: [
            {
              validator: (rule: any, value: any) => {
                const formData = (rule as any).source || {}
                if (formData.authType === 'JWT') {
                  const algorithm = formData['authConfig.algorithm']
                  if (!algorithm || (typeof algorithm === 'string' && algorithm.trim() === '')) {
                    return new Error('签名算法不能为空')
                  }
                }
                return true
              },
              trigger: ['blur', 'change']
            }
          ],
        },
        {
          field: 'authConfig.issuer',
          label: '签发者(Issuer)',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'JWT',
          tips: 'JWT签发者标识，可选',
          render: createNestedFieldRender('issuer', 'input', { placeholder: 'JWT签发者' }),
        },
        {
          field: 'authConfig.expiration',
          label: '过期时间(秒)',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'JWT',
          defaultValue: 3600,
          tips: 'Token过期时间，默认3600秒',
          render: createNestedFieldRender('expiration', 'number', { min: 60, max: 86400, defaultValue: 3600, placeholder: 'Token过期时间' }),
        },
        {
          field: 'authConfig.verifyExpiration',
          label: '验证过期时间',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'JWT',
          tips: '是否验证Token过期时间',
          render: createNestedFieldRender('verifyExpiration', 'switch', { defaultValue: true }),
        },
        {
          field: 'authConfig.verifyIssuer',
          label: '验证签发者',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'JWT',
          tips: '是否验证Token签发者',
          render: createNestedFieldRender('verifyIssuer', 'switch', { defaultValue: false }),
        },
        {
          field: 'authConfig.refreshWindow',
          label: '刷新窗口(秒)',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'JWT',
          defaultValue: 300,
          tips: '过期前多少秒可以刷新Token',
          render: createNestedFieldRender('refreshWindow', 'number', { min: 0, max: 3600, defaultValue: 300, placeholder: '过期前多少秒可刷新' }),
        },
        {
          field: 'authConfig.includeInResponse',
          label: '响应包含Token',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'JWT',
          tips: '是否在响应头中包含生成的Token',
          render: createNestedFieldRender('includeInResponse', 'switch', { defaultValue: false }),
        },
        {
          field: 'authConfig.responseHeaderName',
          label: '响应头名称',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'JWT',
          defaultValue: 'X-Auth-Token',
          tips: '响应头中Token的字段名',
          render: createNestedFieldRender('responseHeaderName', 'input', { placeholder: 'X-Auth-Token', defaultValue: 'X-Auth-Token' }),
        },
        // ============= API Key 认证配置字段 =============
        {
          field: 'authConfig.keyLocation',
          label: 'API Key位置',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'API_KEY',
          required: true,
          tips: 'API Key的获取位置，Header或Query参数',
          render: createNestedFieldRender('keyLocation', 'select', { options: apiKeyLocationOptions, defaultValue: 'header', placeholder: 'API Key获取位置' }),
          rules: [
            {
              validator: (rule: any, value: any) => {
                const formData = (rule as any).source || {}
                if (formData.authType === 'API_KEY') {
                  const keyLocation = formData['authConfig.keyLocation']
                  if (!keyLocation || (typeof keyLocation === 'string' && keyLocation.trim() === '')) {
                    return new Error('API Key位置不能为空')
                  }
                }
                return true
              },
              trigger: ['blur', 'change']
            }
          ],
        },
        {
          field: 'authConfig.keyName',
          label: '参数名称',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'API_KEY',
          required: true,
          tips: 'API Key的参数名称，如X-API-Key',
          render: createNestedFieldRender('keyName', 'input', { placeholder: 'api_key', defaultValue: 'X-API-Key' }),
          rules: [
            {
              validator: (rule: any, value: any) => {
                const formData = (rule as any).source || {}
                if (formData.authType === 'API_KEY') {
                  const keyName = formData['authConfig.keyName']
                  if (!keyName || (typeof keyName === 'string' && keyName.trim() === '')) {
                    return new Error('参数名称不能为空')
                  }
                }
                return true
              },
              trigger: ['blur', 'change']
            }
          ],
        },
        {
          field: 'authConfig.validKeys',
          label: '有效API Keys',
          type: 'custom',
          span: 24,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'API_KEY',
          required: true,
          tips: '允许访问的API Key列表，至少配置一个',
          render: createNestedFieldRender('validKeys', 'tags', { placeholder: '输入API Key' }),
          rules: [
            {
              validator: (rule: any, value: any) => {
                const formData = (rule as any).source || {}
                if (formData.authType === 'API_KEY') {
                  const validKeys = formData['authConfig.validKeys']
                  if (!validKeys || !Array.isArray(validKeys) || validKeys.length === 0) {
                    return new Error('至少需要配置一个有效API Key')
                  }
                }
                return true
              },
              trigger: ['blur', 'change']
            }
          ],
        },
        // ============= OAuth2 认证配置字段 =============
        {
          field: 'authConfig.tokenEndpoint',
          label: 'Token端点URL',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'OAUTH2',
          required: true,
          tips: 'OAuth2 Token获取端点URL',
          render: createNestedFieldRender('tokenEndpoint', 'input', { placeholder: 'https://auth.example.com/oauth/token' }),
          rules: [
            {
              validator: (rule: any, value: any) => {
                const formData = (rule as any).source || {}
                if (formData.authType === 'OAUTH2') {
                  const tokenEndpoint = formData['authConfig.tokenEndpoint']
                  if (!tokenEndpoint || (typeof tokenEndpoint === 'string' && tokenEndpoint.trim() === '')) {
                    return new Error('Token端点URL不能为空')
                  }
                }
                return true
              },
              trigger: ['blur', 'change']
            }
          ],
        },
        {
          field: 'authConfig.introspectEndpoint',
          label: '内省端点URL',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'OAUTH2',
          tips: 'OAuth2 Token内省端点URL，可选',
          render: createNestedFieldRender('introspectEndpoint', 'input', { placeholder: 'https://auth.example.com/oauth/introspect' }),
        },
        {
          field: 'authConfig.clientID',
          label: 'Client ID',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'OAUTH2',
          required: true,
          tips: 'OAuth2客户端ID',
          render: createNestedFieldRender('clientID', 'input', { placeholder: '客户端ID' }),
          rules: [
            {
              validator: (rule: any, value: any) => {
                const formData = (rule as any).source || {}
                if (formData.authType === 'OAUTH2') {
                  const clientID = formData['authConfig.clientID']
                  if (!clientID || (typeof clientID === 'string' && clientID.trim() === '')) {
                    return new Error('Client ID不能为空')
                  }
                }
                return true
              },
              trigger: ['blur', 'change']
            }
          ],
        },
        {
          field: 'authConfig.clientSecret',
          label: 'Client Secret',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'OAUTH2',
          required: true,
          tips: 'OAuth2客户端密钥',
          render: createNestedFieldRender('clientSecret', 'input', { type: 'password', placeholder: '客户端密钥' }),
          rules: [
            {
              validator: (rule: any, value: any) => {
                const formData = (rule as any).source || {}
                if (formData.authType === 'OAUTH2') {
                  const clientSecret = formData['authConfig.clientSecret']
                  if (!clientSecret || (typeof clientSecret === 'string' && clientSecret.trim() === '')) {
                    return new Error('Client Secret不能为空')
                  }
                }
                return true
              },
              trigger: ['blur', 'change']
            }
          ],
        },
        {
          field: 'authConfig.scope',
          label: 'Scope',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'OAUTH2',
          tips: 'OAuth2授权范围，多个用空格分隔',
          render: createNestedFieldRender('scope', 'input', { placeholder: 'read write' }),
        },
        // ============= Basic 认证配置字段 =============
        {
          field: 'authConfig.username',
          label: '用户名',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'BASIC',
          required: true,
          tips: 'Basic认证用户名',
          render: createNestedFieldRender('username', 'input', { placeholder: '认证用户名' }),
          rules: [
            {
              validator: (rule: any, value: any) => {
                const formData = (rule as any).source || {}
                if (formData.authType === 'BASIC') {
                  const username = formData['authConfig.username']
                  if (!username || (typeof username === 'string' && username.trim() === '')) {
                    return new Error('用户名不能为空')
                  }
                }
                return true
              },
              trigger: ['blur', 'change']
            }
          ],
        },
        {
          field: 'authConfig.password',
          label: '密码',
          type: 'custom',
          span: 12,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.authType === 'BASIC',
          required: true,
          tips: 'Basic认证密码',
          render: createNestedFieldRender('password', 'input', { type: 'password', placeholder: '认证密码' }),
          rules: [
            {
              validator: (rule: any, value: any) => {
                const formData = (rule as any).source || {}
                if (formData.authType === 'BASIC') {
                  const password = formData['authConfig.password']
                  if (!password || (typeof password === 'string' && password.trim() === '')) {
                    return new Error('密码不能为空')
                  }
                }
                return true
              },
              trigger: ['blur', 'change']
            }
          ],
        },
      ],
    },
    // ============= 豁免配置分组 =============
    {
      field: 'exemptConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      tabKey: 'basic',
      children: [
        {
          field: 'exemptPaths',
          label: '豁免路径',
          type: 'custom',
          span: 12,
          defaultValue: [],
          tabKey: 'basic',
          tips: '支持通配符，如：/public/*, /health, /api/v1/status',
          render: createArrayFieldRender('exemptPaths', '输入豁免路径'),
        },
        {
          field: 'exemptHeaders',
          label: '豁免请求头',
          type: 'custom',
          span: 12,
          defaultValue: [],
          tabKey: 'basic',
          tips: '包含指定请求头的请求将跳过认证',
          render: createArrayFieldRender('exemptHeaders', '输入豁免请求头'),
        },
      ],
    },
    // ============= 活动状态（放在最后） =============
    {
      field: 'activeFlag',
      label: '活动状态',
      type: 'switch',
      span: 12,
      defaultValue: 'N', // 默认关闭（非活动状态）
      tabKey: 'basic',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
    // ============= 其他字段（只读） =============
    {
      field: 'addTime',
      label: '创建时间',
      type: 'datetime',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'addWho',
      label: '创建人',
      type: 'input',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'editTime',
      label: '修改时间',
      type: 'datetime',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'editWho',
      label: '修改人',
      type: 'input',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
  ]

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,

    // 配置
    formTabs,
    formFields,
  }
}

