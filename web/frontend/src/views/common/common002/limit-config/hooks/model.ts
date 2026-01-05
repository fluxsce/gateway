/**
 * 限流配置 Model
 * 统一管理表单配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import { NAlert, NText } from 'naive-ui'
import { h, ref } from 'vue'

/**
 * 限流配置 Model
 */
export function useRateLimitConfigModel() {
  // ============= 数据状态 =============
  const moduleId = 'hubcommon002-rate-limit'
  /** 加载状态 */
  const loading = ref(false)

  // ============= 表单配置 =============

  // 限流键策略选项
  const keyStrategyOptions = [
    { label: 'IP策略', value: 'ip' },
    { label: '用户策略', value: 'user' },
    { label: '路径策略', value: 'path' },
    { label: '服务策略', value: 'service' },
    { label: '路由策略', value: 'route' },
  ]

  // 限流算法选项
  const algorithmOptions = [
    { label: '令牌桶算法', value: 'token-bucket' },
    { label: '漏桶算法', value: 'leaky-bucket' },
    { label: '滑动窗口算法', value: 'sliding-window' },
    { label: '固定窗口算法', value: 'fixed-window' },
    { label: '无限流', value: 'none' },
  ]

  // 创建算法说明的渲染函数
  const createAlgorithmDescriptionRender = () => {
    return (formData: Record<string, any>) => {
      const algorithm = formData.algorithm

      if (algorithm === 'token-bucket') {
        return h(NAlert, { type: 'info' }, {
          default: () => h(NText, null, {
            default: () => [
              h('strong', null, '令牌桶算法：'),
              '以固定速率向桶中放入令牌，请求需要消耗令牌才能通过。支持突发流量，桶满时多余令牌会被丢弃。',
              h('br'),
              '• 限流速率：每秒向桶中填充的令牌数（令牌/秒）',
              h('br'),
              '• 突发容量：桶的最大容量（最大令牌数），决定了最大突发请求数',
            ],
          }),
        })
      } else if (algorithm === 'leaky-bucket') {
        return h(NAlert, { type: 'info' }, {
          default: () => h(NText, null, {
            default: () => [
              h('strong', null, '漏桶算法：'),
              '请求先进入桶中，然后以固定速率从桶中流出。能够平滑突发流量，但不支持突发处理。',
              h('br'),
              '• 限流速率：每秒从桶中漏出的请求数（处理速率）',
              h('br'),
              '• 突发容量：桶的最大容量（最大可容纳的请求数），超出时请求会被拒绝',
            ],
          }),
        })
      } else if (algorithm === 'sliding-window') {
        return h(NAlert, { type: 'info' }, {
          default: () => h(NText, null, {
            default: () => [
              h('strong', null, '滑动窗口算法：'),
              '在固定时间窗口内限制请求数量。窗口会随时间滑动，提供更精确的限流控制，避免边界突刺问题。',
              h('br'),
              '• 限流速率：时间窗口内允许的最大请求数',
              h('br'),
              '• 时间窗口：统计请求数量的时间范围（秒）',
            ],
          }),
        })
      } else if (algorithm === 'fixed-window') {
        return h(NAlert, { type: 'info' }, {
          default: () => h(NText, null, {
            default: () => [
              h('strong', null, '固定窗口算法：'),
              '在固定时间窗口内限制请求数量。窗口到期后重置计数器，可能出现边界突发（临界问题）。',
              h('br'),
              '• 限流速率：时间窗口内允许的最大请求数',
              h('br'),
              '• 时间窗口：统计请求数量的时间范围（秒）',
            ],
          }),
        })
      } else if (algorithm === 'none') {
        return h(NAlert, { type: 'warning' }, {
          default: () => h(NText, null, {
            default: () => [
              h('strong', null, '无限流：'),
              '不进行任何限流控制，所有请求都会通过。通常用于临时关闭限流或测试环境。',
            ],
          }),
        })
      }

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
      field: 'rateLimitConfigId',
      label: '限流配置ID',
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
          field: 'limitName',
          label: '限流规则名称',
          type: 'input',
          placeholder: '请输入限流规则名称',
          span: 12,
          required: true,
          defaultValue: '限流规则',
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
          field: 'keyStrategy',
          label: '限流键策略',
          type: 'select',
          span: 12,
          required: true,
          defaultValue: 'ip',
          tabKey: 'basic',
          props: {
            options: keyStrategyOptions,
            placeholder: '选择限流键策略',
          },
          tips: 'IP策略：按IP地址限流；用户策略：按用户ID限流；路径策略：按请求路径限流；服务策略：按服务名限流；路由策略：按路由配置限流',
        },
        {
          field: 'algorithm',
          label: '限流算法',
          type: 'select',
          span: 12,
          required: true,
          defaultValue: 'token-bucket',
          tabKey: 'basic',
          props: {
            options: algorithmOptions,
            placeholder: '选择限流算法',
          },
          tips: '支持令牌桶、漏桶、滑动窗口、固定窗口等限流算法',
        },
      ],
    },
    // ============= 限流参数配置分组 =============
    {
      field: 'limitParamsConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      tabKey: 'basic',
      children: [
        {
          field: 'limitRate',
          label: (formData: Record<string, any>) => {
            const algorithm = formData?.algorithm
            if (algorithm === 'token-bucket') {
              return '填充速率(令牌/秒)'
            } else if (algorithm === 'leaky-bucket') {
              return '漏出速率(请求/秒)'
            } else if (algorithm === 'sliding-window' || algorithm === 'fixed-window') {
              return '窗口内最大请求数'
            }
            return '限流速率(次/秒)'
          },
          type: 'number',
          placeholder: '每秒允许的请求数',
          span: 12,
          required: true,
          defaultValue: 100,
          tabKey: 'basic',
          props: {
            min: 1,
            max: 10000,
          },
          tips: (formData: Record<string, any>) => {
            const algorithm = formData?.algorithm
            if (algorithm === 'token-bucket') {
              return '每秒向桶中填充的令牌数（令牌/秒）。每个请求需要消耗一个令牌才能通过。'
            } else if (algorithm === 'leaky-bucket') {
              return '每秒从桶中漏出的请求数（处理速率）。无论输入如何，输出速率都严格限制为此值。'
            } else if (algorithm === 'sliding-window' || algorithm === 'fixed-window') {
              return '时间窗口内允许的最大请求数。超过此数量的请求将被拒绝。'
            }
            return '每秒允许通过的请求数量'
          },
        },
        {
          field: 'burstCapacity',
          label: (formData: Record<string, any>) => {
            const algorithm = formData?.algorithm
            if (algorithm === 'token-bucket') {
              return '桶容量(令牌)'
            } else if (algorithm === 'leaky-bucket') {
              return '桶容量(请求)'
            }
            return '突发容量'
          },
          type: 'number',
          placeholder: '允许的突发请求数',
          span: 12,
          required: true,
          defaultValue: 200,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => {
            const algorithm = formData?.algorithm
            return algorithm === 'token-bucket' || algorithm === 'leaky-bucket'
          },
          props: {
            min: 0,
            max: 100000,
          },
          tips: (formData: Record<string, any>) => {
            const algorithm = formData?.algorithm
            if (algorithm === 'token-bucket') {
              return '桶的最大容量（最大令牌数）。桶满时可以处理此数量的突发请求。'
            } else if (algorithm === 'leaky-bucket') {
              return '桶的最大容量（最大可容纳的请求数）。超出此容量时，新请求会被拒绝。'
            }
            return '允许的突发请求数量，用于令牌桶和漏桶算法'
          },
        },
        {
          field: 'timeWindowSeconds',
          label: '时间窗口(秒)',
          type: 'number',
          placeholder: '限流时间窗口',
          span: 12,
          required: true,
          defaultValue: 1,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => {
            const algorithm = formData?.algorithm
            return algorithm === 'sliding-window' || algorithm === 'fixed-window'
          },
          props: {
            min: 1,
            max: 3600,
          },
          tips: (formData: Record<string, any>) => {
            const algorithm = formData?.algorithm
            if (algorithm === 'sliding-window') {
              return '滑动时间窗口的大小（秒）。窗口会随时间滑动，提供更精确的限流控制。'
            } else if (algorithm === 'fixed-window') {
              return '固定时间窗口的大小（秒）。窗口到期后重置计数器，可能出现边界突发。'
            }
            return '限流时间窗口，用于滑动窗口和固定窗口算法'
          },
        },
        {
          field: 'rejectionStatusCode',
          label: '拒绝状态码',
          type: 'number',
          placeholder: '限流时返回的HTTP状态码',
          span: 12,
          required: true,
          defaultValue: 429,
          tabKey: 'basic',
          props: {
            min: 400,
            max: 599,
          },
          tips: '限流时返回的HTTP状态码，通常为429（Too Many Requests）',
        },
        {
          field: 'rejectionMessage',
          label: '拒绝提示消息',
          type: 'input',
          placeholder: '限流时的提示消息',
          span: 24,
          required: true,
          defaultValue: '请求过于频繁，请稍后再试',
          tabKey: 'basic',
          tips: '限流时返回给客户端的错误消息',
        },
      ],
    },
    // ============= 算法说明分组 =============
    {
      field: 'algorithmDescription',
      label: '',
      type: 'fieldset',
      span: 24,
      tabKey: 'basic',
      children: [
        {
          field: 'algorithmDesc',
          label: '算法说明',
          type: 'custom',
          span: 24,
          tabKey: 'basic',
          show: (formData: Record<string, any>) => formData.algorithm && formData.algorithm !== 'none',
          render: createAlgorithmDescriptionRender(),
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
      tabKey: 'other',
      placeholder: '请输入备注信息',
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

