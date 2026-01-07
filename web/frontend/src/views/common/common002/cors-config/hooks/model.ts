/**
 * CORS配置 Model
 * 统一管理表单配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import { NDynamicTags } from 'naive-ui'
import { h, ref } from 'vue'

/**
 * CORS配置 Model
 * @param moduleId 模块ID（用于权限控制，必填，由父组件统一传入）
 */
export function useCorsConfigModel(moduleId: string) {
  // ============= 数据状态 =============
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

  // ============= 表单页签配置 =============
  const formTabs = [
    { key: 'basic', label: '基础设置' },
    { key: 'other', label: '其他' },
  ]

  /** 表单字段配置 */
  const formFields: DataFormField[] = [
    // ============= 主键字段（隐藏，但必须存在用于更新） =============
    {
      field: 'corsConfigId',
      label: 'CORS配置ID',
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
    {
      field: 'securityConfigId',
      label: '安全配置ID',
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
          field: 'configName',
          label: '配置名称',
          type: 'input',
          placeholder: '请输入CORS配置名称',
          span: 12,
          required: true,
          defaultValue: 'CORS跨域配置',
          tabKey: 'basic',
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
          field: 'allowCredentials',
          label: '允许携带凭证',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tabKey: 'basic',
          tips: '是否允许跨域请求携带Cookie、Authorization等凭证信息',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'maxAgeSeconds',
          label: '预检缓存时间',
          type: 'number',
          placeholder: '预检请求缓存时间(秒)',
          span: 12,
          defaultValue: 86400,
          tabKey: 'basic',
          props: {
            min: 0,
            max: 86400,
          },
          tips: 'OPTIONS预检请求的缓存时间，0表示不缓存，最大86400秒(24小时)',
        },
      ],
    },
    // ============= 允许的源配置分组 =============
    {
      field: 'allowOriginsConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      tabKey: 'basic',
      children: [
        {
          field: 'allowOrigins',
          label: '允许的源',
          type: 'custom',
          span: 24,
          defaultValue: [],
          required: true,
          tabKey: 'basic',
          tips: '支持具体域名(https://example.com)、通配符(*)，建议避免使用*以提高安全性',
          render: createArrayFieldRender('allowOrigins', '输入允许的源，如：https://example.com'),
          rules: [
            {
              required: true,
              message: '请至少配置一个允许的源',
              trigger: 'blur',
              validator: (_rule: any, value: string[]) => {
                if (!value || value.length === 0) {
                  return new Error('请至少配置一个允许的源')
                }
                return true
              },
            },
          ],
        },
      ],
    },
    // ============= HTTP方法配置分组 =============
    {
      field: 'allowMethodsConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      tabKey: 'basic',
      children: [
        {
          field: 'allowMethods',
          label: '允许的HTTP方法',
          type: 'custom',
          span: 24,
          defaultValue: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
          required: true,
          tabKey: 'basic',
          tips: '常用方法：GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD',
          render: createArrayFieldRender('allowMethods', '输入HTTP方法，如：GET, POST'),
          rules: [
            {
              required: true,
              message: '请至少配置一个允许的HTTP方法',
              trigger: 'blur',
              validator: (_rule: any, value: string[]) => {
                if (!value || value.length === 0) {
                  return new Error('请至少配置一个允许的HTTP方法')
                }
                return true
              },
            },
          ],
        },
      ],
    },
    // ============= 请求头配置分组 =============
    {
      field: 'headersConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      tabKey: 'basic',
      children: [
        {
          field: 'allowHeaders',
          label: '允许的请求头',
          type: 'custom',
          span: 12,
          defaultValue: ['Content-Type', 'Authorization', 'X-Requested-With'],
          tabKey: 'basic',
          tips: '如：Content-Type, Authorization, X-Requested-With',
          render: createArrayFieldRender('allowHeaders', '输入允许的请求头'),
        },
        {
          field: 'exposeHeaders',
          label: '暴露的响应头',
          type: 'custom',
          span: 12,
          defaultValue: [],
          tabKey: 'basic',
          tips: '浏览器可以访问的自定义响应头',
          render: createArrayFieldRender('exposeHeaders', '输入暴露的响应头'),
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
    {
      field: 'oprSeqFlag',
      label: '操作序列标识',
      type: 'input',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'currentVersion',
      label: '当前版本号',
      type: 'number',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'noteText',
      label: '备注',
      type: 'textarea',
      span: 24,
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

