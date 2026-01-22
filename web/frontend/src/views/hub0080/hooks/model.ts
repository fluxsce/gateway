/**
 * 告警渠道配置列表 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { AddOutline, TrashOutline } from '@vicons/ionicons5'
import { h, ref } from 'vue'
import { TemplateNameSelector } from '../../hub0081/components/template-grid'
import type {
  AlertConfig,
  ChannelType
} from '../types'
import {
  ACTIVE_FLAG_OPTIONS,
  ASYNC_SEND_FLAG_OPTIONS,
  CHANNEL_TYPE_OPTIONS,
  DEFAULT_FLAG_OPTIONS,
  MESSAGE_CONTENT_FORMAT_OPTIONS,
} from '../types'

/**
 * 告警渠道配置列表 Model
 */
export function useAlertConfigModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0080:alert-config'
  
  /** 加载状态 */
  const loading = ref(false)

  /** 告警渠道配置列表数据 */
  const configList = ref<AlertConfig[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'channelName',
        label: '渠道名称',
        type: 'input',
        placeholder: '请输入渠道名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'channelType',
        label: '渠道类型',
        type: 'select',
        placeholder: '请选择渠道类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...CHANNEL_TYPE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
        ],
      },
      {
        field: 'activeFlag',
        label: '启用状态',
        type: 'select',
        placeholder: '请选择启用状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...ACTIVE_FLAG_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
        ],
      },
      {
        field: 'defaultFlag',
        label: '默认渠道',
        type: 'select',
        placeholder: '请选择是否默认',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...DEFAULT_FLAG_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增渠道',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新增告警渠道配置',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '批量删除选中的渠道',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 获取渠道类型显示标签 */
  const getChannelTypeLabel = (channelType: ChannelType) => {
    const option = CHANNEL_TYPE_OPTIONS.find(opt => opt.value === channelType)
    return option?.label || channelType
  }

  /** 获取渠道类型标签颜色 */
  const getChannelTypeTagType = (channelType: ChannelType): "default" | "success" | "error" | "warning" | "primary" | "info" => {
    const typeMap: Record<ChannelType, "default" | "success" | "error" | "warning" | "primary" | "info"> = {
      email: 'primary',
      qq: 'info',
      wechat_work: 'success',
      dingtalk: 'warning',
      webhook: 'default',
      sms: 'error',
    }
    return typeMap[channelType] || 'default'
  }

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'channelName',
        title: '渠道名称',
        align: 'center',
        showOverflow: 'tooltip',
        width: 160,
      },
      {
        field: 'channelType',
        title: '渠道类型',
        align: 'center',
        width: 120,
        slots: { default: 'channelType' },
      },
      {
        field: 'channelDesc',
        title: '渠道描述',
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
      },
      {
        field: 'priorityLevel',
        title: '优先级',
        align: 'center',
        width: 100,
      },
      {
        field: 'defaultFlag',
        title: '默认渠道',
        align: 'center',
        width: 100,
        slots: { default: 'defaultFlag' },
      },
      {
        field: 'activeFlag',
        title: '启用状态',
        align: 'center',
        width: 100,
        slots: { default: 'activeFlag' },
      },
      {
        field: 'totalSentCount',
        title: '总发送数',
        align: 'center',
        width: 100,
      },
      {
        field: 'successCount',
        title: '成功数',
        align: 'center',
        width: 100,
      },
      {
        field: 'failureCount',
        title: '失败数',
        align: 'center',
        width: 100,
      },
      {
        field: 'lastSendTime',
        title: '最后发送时间',
        align: 'center',
        formatter: ({ row }) => formatDate(row.lastSendTime),
        width: 160,
      },
      {
        field: 'addTime',
        title: '创建时间',
        align: 'center',
        formatter: ({ row }) => formatDate(row.addTime),
        width: 160,
      },
      {
        field: 'editTime',
        title: '修改时间',
        align: 'center',
        formatter: ({ row }) => formatDate(row.editTime),
        width: 160,
      },
    ],
    showCheckbox: true,
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      showCopyRow: true,
      customMenus: [
        {
          code: 'view',
          name: '查看详情',
          prefixIcon: 'vxe-icon-eye-fill',
        },
        {
          code: 'edit',
          name: '编辑',
          prefixIcon: 'vxe-icon-edit',
        },
        {
          code: 'copy',
          name: '复制',
          prefixIcon: 'vxe-icon-copy',
        },
        {
          code: 'reload',
          name: '重载配置',
          prefixIcon: 'vxe-icon-refresh',
        },
        {
          code: 'setDefault',
          name: '设为默认',
          prefixIcon: 'vxe-icon-star',
        },
        {
          code: 'test',
          name: '预警测试',
          prefixIcon: 'vxe-icon-check',
        },
        {
          code: 'delete',
          name: '删除',
          prefixIcon: 'vxe-icon-delete',
        },
      ],
    },
  }

  // ============= 状态更新方法 =============

  /**
   * 设置配置列表
   */
  function setConfigList(list: AlertConfig[]) {
    configList.value = list
  }

  /**
   * 设置加载状态
   */
  function setLoading(value: boolean) {
    loading.value = value
  }

  /**
   * 重置分页信息
   */
  function resetPagination() {
    pageInfo.value = undefined
  }

  /**
   * 更新分页信息（接收后端 PageInfoObj）
   */
  function updatePagination(newPageInfo: Partial<PageInfoObj>) {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  /**
   * 添加配置到列表
   */
  function addConfigToList(config: AlertConfig) {
    configList.value.push(config)
  }

  /**
   * 更新列表中的配置
   */
  function updateConfigInList(channelName: string, tenantId: string | undefined, updatedConfig: Partial<AlertConfig>) {
    const index = configList.value.findIndex(
      (c) => c.channelName === channelName && (!tenantId || c.tenantId === tenantId)
    )
    if (index !== -1) {
      Object.assign(configList.value[index], updatedConfig)
    }
  }

  /**
   * 从列表中移除配置
   */
  function removeConfigFromList(channelName: string) {
    const index = configList.value.findIndex((c) => c.channelName === channelName)
    if (index >= 0) {
      configList.value.splice(index, 1)
    }
  }

  /**
   * 从列表中批量移除配置
   */
  function removeConfigsFromList(channelNames: string[]) {
    configList.value = configList.value.filter((c) => !channelNames.includes(c.channelName))
  }

  // ============= 表单配置 =============

  /** 表单页签配置 */
  const formTabs = [
    {
      key: 'basic',
      label: '基本信息',
    },
    {
      key: 'config',
      label: '渠道配置',
    },
    {
      key: 'other',
      label: '其他信息',
    },
  ]

  /** 配置表单字段（用于 GdataFormModal） */
  const formFields: DataFormField[] = [
    // 主键字段（隐藏，但必须存在用于编辑）
    {
      field: 'channelName',
      label: '渠道名称',
      type: 'input' as const,
      span: 12,
      primary: true,
      tabKey: 'basic',
      required: true,
      rules: [
        { required: true, message: '请输入渠道名称', trigger: ['blur', 'input'] },
        { max: 32, message: '渠道名称不能超过32个字符', trigger: ['blur', 'input'] },
        {
          pattern: /^[a-zA-Z0-9_]+$/,
          message: '渠道名称只能包含英文字母、数字和下划线',
          trigger: ['blur', 'input'],
        },
      ],
    },
    {
      field: 'tenantId',
      label: '租户ID',
      type: 'input' as const,
      span: 12,
      show: false,
    },
    // 基本信息
    {
      field: 'channelType',
      label: '渠道类型',
      type: 'select' as const,
      placeholder: '请选择渠道类型',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 'email',
      options: CHANNEL_TYPE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
      rules: [
        { required: true, message: '请选择渠道类型', trigger: ['blur', 'change'] },
      ],
    },
    {
      field: 'channelDesc',
      label: '渠道描述',
      type: 'input' as const,
      placeholder: '请输入渠道描述',
      span: 12,
      tabKey: 'basic',
      props: {
        type: 'textarea',
        rows: 2,
        maxlength: 500,
        showCount: true,
      },
    },
    {
      field: 'priorityLevel',
      label: '优先级',
      type: 'number' as const,
      placeholder: '请输入优先级（1-10，数字越小优先级越高）',
      span: 12,
      tabKey: 'basic',
      defaultValue: 5,
      tips: '优先级范围：1-10，数字越小优先级越高',
      props: {
        min: 1,
        max: 10,
        precision: 0,
      },
      rules: [
        { required: true, type: 'number', message: '请输入优先级', trigger: ['blur', 'change'] },
      ],
    },
    {
      field: 'defaultFlag',
      label: '默认渠道',
      type: 'select' as const,
      placeholder: '请选择是否设为默认渠道',
      span: 12,
      tabKey: 'basic',
      defaultValue: 'N',
      options: DEFAULT_FLAG_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
      tips: '设为默认后，发送告警时如果不指定渠道，将使用此渠道',
    },
    {
      field: 'activeFlag',
      label: '启用状态',
      type: 'switch' as const,
      span: 12,
      tabKey: 'basic',
      defaultValue: 'Y',
      tips: '禁用后此渠道将不会被使用',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
    {
      field: 'defaultTemplateName',
      label: '默认模板名称',
      type: 'custom' as const,
      span: 12,
      tabKey: 'basic',
      placeholder: '请输入模板名称或点击选择',
      render: (formData: Record<string, any>) => {
        return h(TemplateNameSelector, {
          modelValue: formData.defaultTemplateName || '',
          'onUpdate:modelValue': (value: string) => {
            formData.defaultTemplateName = value
          },
          channelType: formData.channelType,
        })
      },
    },
    // 渠道配置页签 - 使用 fieldset 分组，根据渠道类型动态显示
    {
      field: 'serverConfigGroup',
      label: '服务器配置',
      type: 'fieldset' as const,
      span: 24,
      tabKey: 'config',
      show: (formData) => formData.channelType === 'email',
      children: [
        {
          field: 'serverConfig.smtp_host',
          label: 'SMTP服务器地址',
          type: 'input' as const,
          placeholder: '请输入SMTP服务器地址，如：smtp.example.com',
          span: 12,
          required: true,
          tips: 'SMTP服务器地址，如 smtp.gmail.com',
          rules: [
            { required: true, message: '请输入SMTP服务器地址', trigger: ['blur', 'input'] },
          ],
        },
        {
          field: 'serverConfig.smtp_port',
          label: 'SMTP端口',
          type: 'number' as const,
          placeholder: '请输入SMTP端口',
          span: 12,
          required: true,
          defaultValue: 587,
          tips: 'SMTP端口，常用端口：587(TLS), 465(SSL), 25(非加密)',
          props: {
            min: 1,
            max: 65535,
            precision: 0,
          },
          rules: [
            { required: true, type: 'number', message: '请输入SMTP端口', trigger: ['blur', 'change'] },
          ],
        },
        {
          field: 'serverConfig.username',
          label: '用户名',
          type: 'input' as const,
          placeholder: '请输入SMTP用户名',
          span: 12,
          required: true,
          tips: 'SMTP认证用户名，通常是邮箱地址',
          rules: [
            { required: true, message: '请输入用户名', trigger: ['blur', 'input'] },
          ],
        },
        {
          field: 'serverConfig.password',
          label: '密码',
          type: 'input' as const,
          placeholder: '请输入SMTP密码',
          span: 12,
          required: true,
          tips: 'SMTP认证密码或授权码',
          props: {
            type: 'password',
            showPasswordOn: 'click',
          },
          rules: [
            { required: true, message: '请输入密码', trigger: ['blur', 'input'] },
          ],
        },
        {
          field: 'serverConfig.from',
          label: '发件人地址',
          type: 'input' as const,
          placeholder: '请输入发件人邮箱地址',
          span: 12,
          required: true,
          tips: '默认发件人邮箱地址',
          rules: [
            { required: true, message: '请输入发件人地址', trigger: ['blur', 'input'] },
            { type: 'email', message: '请输入有效的邮箱地址', trigger: ['blur', 'input'] },
          ],
        },
        {
          field: 'serverConfig.from_name',
          label: '发件人名称',
          type: 'input' as const,
          placeholder: '请输入发件人名称（可选）',
          span: 12,
          tips: '发件人显示名称，如：系统告警',
        },
        {
          field: 'serverConfig.use_tls',
          label: '使用TLS',
          type: 'switch' as const,
          span: 12,
          defaultValue: true,
          tips: '是否使用TLS加密连接',
          props: {
            checkedValue: true,
            uncheckedValue: false,
          },
        },
        {
          field: 'serverConfig.skip_verify',
          label: '跳过证书验证',
          type: 'switch' as const,
          span: 12,
          defaultValue: false,
          tips: '是否跳过TLS证书验证（不推荐，仅用于测试）',
          props: {
            checkedValue: true,
            uncheckedValue: false,
          },
        },
        {
          field: 'serverConfig.timeout',
          label: '超时时间（秒）',
          type: 'number' as const,
          placeholder: '请输入超时时间',
          span: 12,
          defaultValue: 30,
          tips: 'SMTP连接超时时间',
          props: {
            min: 1,
            max: 300,
            precision: 0,
          },
        },
      ],
    },
    {
      field: 'qqServerConfigGroup',
      label: '服务器配置',
      type: 'fieldset' as const,
      span: 24,
      tabKey: 'config',
      show: (formData) => formData.channelType === 'qq',
      children: [
        {
          field: 'serverConfig.webhook_url',
          label: 'Webhook地址',
          type: 'input' as const,
          placeholder: '请输入QQ机器人Webhook地址',
          span: 24,
          required: true,
          tips: 'QQ机器人Webhook地址，从QQ群机器人配置中获取',
          rules: [
            { required: true, message: '请输入Webhook地址', trigger: ['blur', 'input'] },
            { type: 'url', message: '请输入有效的URL地址', trigger: ['blur', 'input'] },
          ],
        },
        {
          field: 'serverConfig.secret',
          label: '签名密钥',
          type: 'input' as const,
          placeholder: '请输入签名密钥（可选）',
          span: 12,
          tips: 'QQ机器人签名密钥（如果配置了签名验证）',
          props: {
            type: 'password',
            showPasswordOn: 'click',
          },
        },
        {
          field: 'serverConfig.timeout',
          label: '超时时间（秒）',
          type: 'number' as const,
          placeholder: '请输入超时时间',
          span: 12,
          defaultValue: 30,
          tips: 'HTTP请求超时时间',
          props: {
            min: 1,
            max: 300,
            precision: 0,
          },
        },
      ],
    },
    {
      field: 'wechatWorkServerConfigGroup',
      label: '服务器配置',
      type: 'fieldset' as const,
      span: 24,
      tabKey: 'config',
      show: (formData) => formData.channelType === 'wechat_work',
      children: [
        {
          field: 'serverConfig.webhook_url',
          label: 'Webhook地址',
          type: 'input' as const,
          placeholder: '请输入企业微信机器人Webhook地址',
          span: 24,
          required: true,
          tips: '企业微信机器人Webhook地址，从企业微信群机器人配置中获取',
          rules: [
            { required: true, message: '请输入Webhook地址', trigger: ['blur', 'input'] },
            { type: 'url', message: '请输入有效的URL地址', trigger: ['blur', 'input'] },
          ],
        },
        {
          field: 'serverConfig.secret',
          label: '签名密钥',
          type: 'input' as const,
          placeholder: '请输入签名密钥（可选）',
          span: 12,
          tips: '企业微信机器人签名密钥（如果配置了签名验证）',
          props: {
            type: 'password',
            showPasswordOn: 'click',
          },
        },
        {
          field: 'serverConfig.message_type',
          label: '消息类型',
          type: 'select' as const,
          placeholder: '请选择消息类型',
          span: 12,
          defaultValue: 'markdown',
          tips: '消息类型：text（文本）或 markdown（Markdown格式）',
          options: [
            { label: '文本', value: 'text' },
            { label: 'Markdown', value: 'markdown' },
          ],
        },
        {
          field: 'serverConfig.timeout',
          label: '超时时间（秒）',
          type: 'number' as const,
          placeholder: '请输入超时时间',
          span: 12,
          defaultValue: 30,
          tips: 'HTTP请求超时时间',
          props: {
            min: 1,
            max: 300,
            precision: 0,
          },
        },
      ],
    },
    {
      field: 'sendConfigGroup',
      label: '发送配置',
      type: 'fieldset' as const,
      span: 24,
      tabKey: 'config',
      show: (formData) => formData.channelType === 'email',
      children: [
        {
          field: 'sendConfig.to',
          label: '收件人列表',
          type: 'input' as const,
          placeholder: '请输入收件人邮箱，多个用逗号分隔',
          span: 24,
          required: true,
          tips: '收件人邮箱地址列表，多个用逗号分隔，如：user1@example.com,user2@example.com',
          rules: [
            { required: true, message: '请输入收件人列表', trigger: ['blur', 'input'] },
          ],
        },
        {
          field: 'sendConfig.cc',
          label: '抄送列表',
          type: 'input' as const,
          placeholder: '请输入抄送邮箱，多个用逗号分隔（可选）',
          span: 12,
          tips: '抄送邮箱地址列表，多个用逗号分隔',
        },
        {
          field: 'sendConfig.bcc',
          label: '密送列表',
          type: 'input' as const,
          placeholder: '请输入密送邮箱，多个用逗号分隔（可选）',
          span: 12,
          tips: '密送邮箱地址列表，多个用逗号分隔',
        },
      ],
    },
    {
      field: 'qqSendConfigGroup',
      label: '发送配置',
      type: 'fieldset' as const,
      span: 24,
      tabKey: 'config',
      show: (formData) => formData.channelType === 'qq',
      children: [
        {
          field: 'sendConfig.at_all',
          label: '@所有人',
          type: 'switch' as const,
          span: 12,
          defaultValue: false,
          tips: '是否@所有人',
          props: {
            checkedValue: true,
            uncheckedValue: false,
          },
        },
        {
          field: 'sendConfig.at_users',
          label: '@指定用户',
          type: 'input' as const,
          placeholder: '请输入QQ号，多个用逗号分隔（可选）',
          span: 12,
          tips: '要@的QQ号列表，多个用逗号分隔',
        },
      ],
    },
    {
      field: 'wechatWorkSendConfigGroup',
      label: '发送配置',
      type: 'fieldset' as const,
      span: 24,
      tabKey: 'config',
      show: (formData) => formData.channelType === 'wechat_work',
      children: [
        {
          field: 'sendConfig.mentioned_list',
          label: '@成员列表（userid）',
          type: 'input' as const,
          placeholder: '请输入成员userid，多个用逗号分隔（可选）',
          span: 12,
          tips: '要@的成员userid列表，多个用逗号分隔',
        },
        {
          field: 'sendConfig.mentioned_mobile_list',
          label: '@成员手机号列表',
          type: 'input' as const,
          placeholder: '请输入成员手机号，多个用逗号分隔（可选）',
          span: 12,
          tips: '要@的成员手机号列表，多个用逗号分隔',
        },
      ],
    },
    {
      field: 'commonConfigGroup',
      label: '通用配置',
      type: 'fieldset' as const,
      span: 24,
      tabKey: 'config',
      children: [
        {
          field: 'messageContentFormat',
          label: '消息内容格式',
          type: 'select' as const,
          placeholder: '请选择消息内容格式',
          span: 12,
          defaultValue: 'html',
          options: MESSAGE_CONTENT_FORMAT_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
          tips: '消息内容的格式类型',
        },
        {
          field: 'timeoutSeconds',
          label: '超时时间（秒）',
          type: 'number' as const,
          placeholder: '请输入超时时间',
          span: 12,
          defaultValue: 30,
          props: {
            min: 1,
            max: 300,
            precision: 0,
          },
        },
        {
          field: 'retryCount',
          label: '重试次数',
          type: 'number' as const,
          placeholder: '请输入重试次数',
          span: 12,
          defaultValue: 3,
          props: {
            min: 0,
            max: 10,
            precision: 0,
          },
        },
        {
          field: 'retryIntervalSecs',
          label: '重试间隔（秒）',
          type: 'number' as const,
          placeholder: '请输入重试间隔',
          span: 12,
          defaultValue: 5,
          props: {
            min: 1,
            max: 60,
            precision: 0,
          },
        },
        {
          field: 'asyncSendFlag',
          label: '异步发送',
          type: 'select' as const,
          placeholder: '请选择是否异步发送',
          span: 12,
          defaultValue: 'Y',
          options: ASYNC_SEND_FLAG_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
          tips: '启用异步发送可提高性能，但可能延迟消息到达',
        },
      ],
    },
    // 其他信息
    {
      field: 'noteText',
      label: '备注信息',
      type: 'input' as const,
      placeholder: '请输入备注信息',
      span: 24,
      tabKey: 'other',
      props: {
        type: 'textarea',
        rows: 3,
        maxlength: 500,
        showCount: true,
      },
    },
    {
      field: 'addTime',
      label: '创建时间',
      type: 'datetime' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'addWho',
      label: '创建人',
      type: 'input' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'editTime',
      label: '修改时间',
      type: 'datetime' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'editWho',
      label: '修改人',
      type: 'input' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
  ]

  return {
    // 状态
    moduleId,
    loading,
    configList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,
    formFields,
    formTabs,

    // 工具函数
    getChannelTypeLabel,
    getChannelTypeTagType,

    // 方法
    setConfigList,
    setLoading,
    resetPagination,
    updatePagination,
    addConfigToList,
    updateConfigInList,
    removeConfigFromList,
    removeConfigsFromList,
  }
}

/**
 * 告警渠道配置列表 Model 类型
 */
export type AlertConfigModel = ReturnType<typeof useAlertConfigModel>

