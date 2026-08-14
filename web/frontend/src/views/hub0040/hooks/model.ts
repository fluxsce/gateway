/**
 * 服务中心实例管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsDataFormField, RsDataFormRenderContext } from '@/components/form/rs-data'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { RsDynamicTags, RsTag, getByNamePath, setByNamePath, type RsTagVariant } from '@/ui'
import { formatDate } from '@/utils/format'
import { AlertChannelNameSelector } from '@/views/hub0080/components'
import { h, ref } from 'vue'
import type { ServiceCenterInstance } from '../types/index'

/**
 * 服务中心实例表格配置（对齐 RsGrid Props 子集）。
 */
export interface ServiceCenterInstanceGridConfig {
  columns: RsGridColumn<ServiceCenterInstance>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * 获取实例状态标签变体
 */
function getInstanceStatusVariant(status: string): RsTagVariant {
  const statusMap: Record<string, RsTagVariant> = {
    RUNNING: 'success',
    STOPPED: 'default',
    STARTING: 'info',
    STOPPING: 'warning',
    ERROR: 'danger',
  }
  return statusMap[status] || 'default'
}

/**
 * 获取实例状态展示文案
 */
function getInstanceStatusText(status: string): string {
  const statusMap: Record<string, string> = {
    RUNNING: '运行中',
    STOPPED: '停止',
    STARTING: '启动中',
    STOPPING: '停止中',
    ERROR: '异常',
  }
  return statusMap[status] || status
}

/**
 * 服务中心实例管理 Model
 */
export function useServiceCenterInstanceModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0040'
  /** 加载状态 */
  const loading = ref(false)

  /** 服务中心实例列表数据 */
  const instanceList = ref<ServiceCenterInstance[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'instanceName',
        label: '实例名称',
        type: 'input',
        placeholder: '请输入实例名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'serverType',
        label: '服务器类型',
        type: 'select',
        placeholder: '请选择类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: 'gRPC', value: 'GRPC' },
          { label: 'HTTP', value: 'HTTP' },
        ],
      },
      {
        field: 'instanceStatus',
        label: '实例状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '停止', value: 'STOPPED' },
          { label: '启动中', value: 'STARTING' },
          { label: '运行中', value: 'RUNNING' },
          { label: '停止中', value: 'STOPPING' },
          { label: '异常', value: 'ERROR' },
        ],
      },
      {
        field: 'activeFlag',
        label: '活动状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '活动', value: 'Y' },
          { label: '非活动', value: 'N' },
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新建实例',
        icon: 'AddOutline',
        type: 'primary',
        tooltip: '新建服务中心实例',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: 'CreateOutline',
        tooltip: '编辑选中的实例',
      },
      {
        key: 'delete',
        label: '删除',
        icon: 'TrashOutline',
        type: 'error',
        tooltip: '删除选中的实例',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 实例表单配置（聚拢配置，类似 searchFormConfig） =============
  const instanceFormConfig = {
    tabs: [
      { key: 'basic', label: '基本信息' },
      { key: 'grpc', label: 'gRPC配置' },
      { key: 'tls', label: 'TLS配置' },
      { key: 'performance', label: '性能配置' },
      { key: 'health', label: '健康检查' },
      { key: 'access', label: '访问控制' },
      { key: 'alert', label: '告警配置' },
      { key: 'other', label: '其它' },
    ],
    fields: [
    // ============= 基本信息 Tab =============
    {
      field: 'tenantId',
      label: '租户ID',
      type: 'input',
      span: 12,
      tabKey: 'basic',
      primary: true,
      show: false, // 隐藏字段，从上下文自动获取
      disabled: true,
    },
    {
      field: 'instanceName',
      label: '实例名称',
      type: 'input',
      placeholder: '请输入实例名称',
      span: 12,
      tabKey: 'basic',
      primary: true,
      required: true,
    },
    {
      field: 'environment',
      label: '部署环境',
      type: 'select',
      placeholder: '请选择部署环境',
      span: 12,
      tabKey: 'basic',
      primary: true,
      required: true,
      options: [
        { label: '开发环境', value: 'DEVELOPMENT' },
        { label: '预发布环境', value: 'STAGING' },
        { label: '生产环境', value: 'PRODUCTION' },
      ],
    },
    {
      field: 'serverType',
      label: '服务器类型',
      type: 'select',
      placeholder: '请选择服务器类型',
      span: 12,
      tabKey: 'basic',
      defaultValue: 'GRPC',
      options: [
        { label: 'gRPC', value: 'GRPC' },
        { label: 'HTTP', value: 'HTTP' },
      ],
    },
    {
      field: 'listenAddress',
      label: '监听地址',
      type: 'input',
      placeholder: '如: 0.0.0.0',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: '0.0.0.0',
      tips: '服务器绑定的网络地址，0.0.0.0表示监听所有网络接口',
    },
    {
      field: 'listenPort',
      label: '监听端口',
      type: 'number',
      placeholder: '如: 12004',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 12004,
      props: {
        min: 1,
        max: 65535,
      },
    },
    {
      field: 'activeFlag',
      label: '活动状态',
      type: 'switch',
      span: 12,
      tabKey: 'basic',
      defaultValue: 'Y',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
    {
      field: 'noteText',
      label: '备注',
      type: 'textarea',
      placeholder: '请输入备注信息',
      span: 24,
      tabKey: 'basic',
      props: {
        rows: 3,
      },
    },

    // ============= gRPC配置 Tab =============
    {
      field: 'grpc-config-group',
      label: 'gRPC 消息大小配置',
      type: 'fieldset',
      tabKey: 'grpc',
      children: [
        {
          field: 'maxRecvMsgSize',
          label: '最大接收消息大小(字节)',
          type: 'number',
          placeholder: '16777216',
          span: 12,
          defaultValue: 16777216,
          tips: '单个 gRPC 消息的最大接收大小，默认16MB',
          props: {
            min: 1024,
            max: 104857600, // 100MB
          },
        },
        {
          field: 'maxSendMsgSize',
          label: '最大发送消息大小(字节)',
          type: 'number',
          placeholder: '16777216',
          span: 12,
          defaultValue: 16777216,
          tips: '单个 gRPC 消息的最大发送大小，默认16MB',
          props: {
            min: 1024,
            max: 104857600, // 100MB
          },
        },
      ],
    },
    {
      field: 'keepalive-config-group',
      label: 'gRPC Keep-Alive 配置',
      type: 'fieldset',
      tabKey: 'grpc',
      children: [
        {
          field: 'keepAliveTime',
          label: 'Keep-alive 发送间隔(秒)',
          type: 'number',
          placeholder: '30',
          span: 12,
          defaultValue: 30,
          tips: '服务器发送 Keep-alive ping 的间隔时间',
          props: {
            min: 1,
            max: 300,
          },
        },
        {
          field: 'keepAliveTimeout',
          label: 'Keep-alive 超时时间(秒)',
          type: 'number',
          placeholder: '10',
          span: 12,
          defaultValue: 10,
          tips: 'Keep-alive ping 的超时时间',
          props: {
            min: 1,
            max: 60,
          },
        },
        {
          field: 'keepAliveMinTime',
          label: '客户端最小 Keep-alive 间隔(秒)',
          type: 'number',
          placeholder: '15',
          span: 12,
          defaultValue: 15,
          tips: '客户端允许的最小 Keep-alive ping 间隔',
          props: {
            min: 1,
            max: 300,
          },
        },
        {
          field: 'permitWithoutStream',
          label: '允许无活跃流时发送 Keep-alive',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '是否允许在没有活跃流的情况下发送 Keep-alive ping',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
      ],
    },
    {
      field: 'connection-config-group',
      label: 'gRPC 连接管理配置',
      type: 'fieldset',
      tabKey: 'grpc',
      children: [
        {
          field: 'maxConnectionIdle',
          label: '最大连接空闲时间(秒)',
          type: 'number',
          placeholder: '0',
          span: 12,
          defaultValue: 0,
          tips: '连接的最大空闲时间，0表示无限制',
          props: {
            min: 0,
          },
        },
        {
          field: 'maxConnectionAge',
          label: '最大连接存活时间(秒)',
          type: 'number',
          placeholder: '0',
          span: 12,
          defaultValue: 0,
          tips: '连接的最大存活时间，0表示无限制',
          props: {
            min: 0,
          },
        },
        {
          field: 'maxConnectionAgeGrace',
          label: '连接关闭宽限期(秒)',
          type: 'number',
          placeholder: '20',
          span: 12,
          defaultValue: 20,
          tips: '连接关闭前的宽限期，允许正在处理的请求完成',
          props: {
            min: 0,
            max: 300,
          },
        },
        {
          field: 'enableReflection',
          label: '启用 gRPC 反射',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '启用 gRPC 反射服务，用于 grpcurl 等工具调试',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
      ],
    },

    // ============= TLS配置 Tab =============
    {
      field: 'tls-config-group',
      label: 'TLS安全配置',
      type: 'fieldset',
      tabKey: 'tls',
      children: [
        {
          field: 'certStorageType',
          label: '证书存储类型',
          type: 'select',
          span: 12,
          defaultValue: 'DATABASE',
          show: false, // 隐藏字段，默认存储到数据库
          options: [
            { label: '文件存储', value: 'FILE' },
            { label: '数据库存储', value: 'DATABASE' },
          ],
        },
        {
          field: 'enableTLS',
          label: '启用TLS',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '启用TLS加密传输，保护数据传输安全',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'enableMTLS',
          label: '启用双向TLS认证',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '启用双向TLS认证（mTLS），要求客户端也提供证书',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'certPassword',
          label: '证书密码',
          type: 'input',
          placeholder: '请输入证书密码(可选)',
          span: 12,
          tips: '如果私钥文件已加密，需要提供密码进行解密',
          props: {
            type: 'password',
            showPasswordOn: 'click',
          },
        },
        {
          field: 'certFileList',
          label: '证书文件',
          type: 'file',
          span: 24,
          props: {
            showDownload: true,
            config: {
              accept: '.crt,.pem,.cer',
              max: 1,
              maxSize: 10 * 1024 * 1024, // 10MB
              uploadText: '点击或拖拽上传证书',
              uploadDescription: '支持 .crt, .pem, .cer',
            },
          },
        },
        {
          field: 'keyFileList',
          label: '私钥文件',
          type: 'file',
          span: 24,
          props: {
            showDownload: true,
            config: {
              accept: '.key,.pem',
              max: 1,
              maxSize: 10 * 1024 * 1024, // 10MB
              uploadText: '点击或拖拽上传私钥',
              uploadDescription: '支持 .key, .pem',
            },
          },
        },
      ],
    },

    // ============= 性能配置 Tab =============
    {
      field: 'performance-config-group',
      label: '性能调优配置',
      type: 'fieldset',
      tabKey: 'performance',
      children: [
        {
          field: 'maxConcurrentStreams',
          label: '最大并发流数量',
          type: 'number',
          placeholder: '250',
          span: 12,
          defaultValue: 250,
          tips: '单个连接的最大并发流数量，0表示无限制',
          props: {
            min: 0,
            max: 10000,
          },
        },
        {
          field: 'readBufferSize',
          label: '读缓冲区大小(字节)',
          type: 'number',
          placeholder: '32768',
          span: 12,
          defaultValue: 32768,
          tips: '读取数据的缓冲区大小，默认32KB',
          props: {
            min: 1024,
            max: 1048576, // 1MB
          },
        },
        {
          field: 'writeBufferSize',
          label: '写缓冲区大小(字节)',
          type: 'number',
          placeholder: '32768',
          span: 12,
          defaultValue: 32768,
          tips: '写入数据的缓冲区大小，默认32KB',
          props: {
            min: 1024,
            max: 1048576, // 1MB
          },
        },
      ],
    },

    // ============= 健康检查 Tab =============
    {
      field: 'health-config-group',
      label: '健康检查配置',
      type: 'fieldset',
      tabKey: 'health',
      children: [
        {
          field: 'healthCheckInterval',
          label: '健康检查间隔(秒)',
          type: 'number',
          placeholder: '60',
          span: 12,
          defaultValue: 60,
          tips: '健康检查的执行间隔，0表示禁用健康检查。注意：客户端的心跳时间应小于此间隔，建议心跳时间为间隔的1/2到2/3',
          props: {
            min: 0,
            max: 3600,
          },
        },
        {
          field: 'healthCheckTimeout',
          label: '健康检查超时时间(秒)',
          type: 'number',
          placeholder: '5',
          span: 12,
          defaultValue: 5,
          tips: '健康检查的超时时间',
          props: {
            min: 1,
            max: 60,
          },
        },
      ],
    },

    // ============= 访问控制 Tab =============
    {
      field: 'access-config-group',
      label: '访问控制配置',
      type: 'fieldset',
      tabKey: 'access',
      children: [
        {
          field: 'enableAuth',
          label: '启用认证',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '启用认证后，客户端需要提供有效的认证信息才能访问',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'ipWhitelist',
          label: 'IP 白名单',
          type: 'custom',
          span: 24,
          defaultValue: [],
          tips: '允许访问的IP地址或CIDR网段，留空表示不限制',
          render: (formData: Record<string, any>, ctx?: RsDataFormRenderContext) => {
            let value = ctx ? ctx.value : formData.ipWhitelist || []
            if (typeof value === 'string') {
              try {
                value = JSON.parse(value)
              } catch {
                value = value.split(',').map((s: string) => s.trim()).filter(Boolean)
              }
            }
            return h(RsDynamicTags, {
              modelValue: Array.isArray(value) ? value : [],
              'onUpdate:modelValue': (newValue: string[]) => {
                const next = newValue.length > 0 ? JSON.stringify(newValue) : ''
                if (ctx?.onUpdate) ctx.onUpdate(next)
                else formData.ipWhitelist = next
              },
              placeholder: '添加IP地址或CIDR网段，如: 192.168.1.0/24',
            })
          },
        },
        {
          field: 'ipBlacklist',
          label: 'IP 黑名单',
          type: 'custom',
          span: 24,
          defaultValue: [],
          tips: '禁止访问的IP地址或CIDR网段',
          render: (formData: Record<string, any>, ctx?: RsDataFormRenderContext) => {
            let value = ctx ? ctx.value : formData.ipBlacklist || []
            if (typeof value === 'string') {
              try {
                value = JSON.parse(value)
              } catch {
                value = value.split(',').map((s: string) => s.trim()).filter(Boolean)
              }
            }
            return h(RsDynamicTags, {
              modelValue: Array.isArray(value) ? value : [],
              'onUpdate:modelValue': (newValue: string[]) => {
                const next = newValue.length > 0 ? JSON.stringify(newValue) : ''
                if (ctx?.onUpdate) ctx.onUpdate(next)
                else formData.ipBlacklist = next
              },
              placeholder: '添加IP地址或CIDR网段，如: 192.168.1.100',
            })
          },
        },
      ],
    },

    // ============= 告警配置 Tab =============
    {
      field: 'alert-basic-group',
      label: '告警基础配置',
      type: 'fieldset',
      tabKey: 'alert',
      children: [
        {
          field: 'extProperty.alertEnabled',
          label: '启用告警',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '启用后，服务中心将根据以下配置发送告警通知',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'extProperty.channelName',
          label: '告警渠道名称',
          type: 'custom',
          span: 12,
          placeholder: '请输入告警渠道名称或点击选择',
          tips: '不填写则使用默认告警渠道',
          render: (formData: Record<string, any>, ctx?: RsDataFormRenderContext) => {
            const raw = ctx ? ctx.value : getByNamePath(formData, 'extProperty.channelName')
            let channelName = ''
            if (typeof raw === 'string') channelName = raw
            else if (raw != null) channelName = String(raw)
            return h(AlertChannelNameSelector, {
              modelValue: channelName,
              'onUpdate:modelValue': (value: string) => {
                if (ctx?.onUpdate) ctx.onUpdate(value)
                else setByNamePath(formData, 'extProperty.channelName', value)
              },
            })
          },
        },
      ],
    },
    {
      field: 'alert-critical-group',
      label: '关键告警配置（默认启用）',
      type: 'fieldset',
      tabKey: 'alert',
      children: [
        {
          field: 'extProperty.alertOnStartFailure',
          label: '服务启动失败告警',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '当服务中心启动失败时发送告警',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'extProperty.alertOnStopAbnormal',
          label: '服务异常停止告警',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '当服务中心异常停止时发送告警',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'extProperty.alertOnHealthCheckFail',
          label: '健康检查失败告警',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '当服务健康检查失败时发送告警',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'extProperty.alertOnSyncFailure',
          label: '缓存同步失败告警',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '当缓存同步失败时发送告警',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'extProperty.alertOnConfigChange',
          label: '配置变更告警',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '当配置发生变更（新增/修改/删除/回滚）时发送告警',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
      ],
    },
    {
      field: 'alert-node-group',
      label: '节点告警配置',
      type: 'fieldset',
      tabKey: 'alert',
      children: [
        {
          field: 'extProperty.alertOnNodeEviction',
          label: '节点驱逐告警',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '当单次驱逐节点数量超过阈值时发送告警',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'extProperty.nodeEvictionThreshold',
          label: '节点驱逐阈值',
          type: 'number',
          placeholder: '5',
          span: 12,
          defaultValue: 5,
          tips: '单次健康检查驱逐节点数量达到此阈值时触发告警',
          props: {
            min: 1,
            max: 1000,
          },
        },
      ],
    },
    {
      field: 'alert-operation-group',
      label: '运维操作告警配置（默认关闭，高频操作）',
      type: 'fieldset',
      tabKey: 'alert',
      children: [
        {
          field: 'extProperty.alertOnNodeRegister',
          label: '节点注册告警',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '当节点注册时发送告警（高频操作，慎重开启）',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'extProperty.alertOnNodeUnregister',
          label: '节点注销告警',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '当节点注销时发送告警（高频操作，慎重开启）',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'extProperty.alertOnSubscribeNotify',
          label: '订阅通知告警',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '当服务订阅变更时发送告警（高频操作，慎重开启）',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'extProperty.alertOnConnectionLost',
          label: '连接断开告警',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '当客户端连接断开时发送告警（高频操作，慎重开启）',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
      ],
    },

    // ============= 其它 Tab =============
    {
      field: 'addTime',
      label: '创建时间',
      type: 'datetime',
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'addWho',
      label: '创建人',
      type: 'input',
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'editTime',
      label: '修改时间',
      type: 'datetime',
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'editWho',
      label: '修改人',
      type: 'input',
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
  ] as RsDataFormField[],
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 RsGrid Props 结构，排除响应式数据） */
  const gridConfig: ServiceCenterInstanceGridConfig = {
    columns: [
      {
        key: 'instanceName',
        title: '实例名称',
        sortable: true,
        align: 'center',
        ellipsis: true,
      },
      {
        key: 'environment',
        title: '部署环境',
        sortable: true,
        align: 'center',
        ellipsis: true,
        formatter: (value) => {
          const envMap: Record<string, string> = {
            DEVELOPMENT: '开发环境',
            STAGING: '预发布环境',
            PRODUCTION: '生产环境',
          }
          return envMap[String(value || '')] || String(value || '')
        },
      },
      {
        key: 'serverType',
        title: '服务器类型',
        align: 'center',
        ellipsis: true,
        formatter: (value) => (value === 'GRPC' ? 'gRPC' : String(value || '')),
      },
      {
        key: 'listenAddress',
        title: '监听地址',
        align: 'center',
        ellipsis: true,
      },
      {
        key: 'listenPort',
        title: '监听端口',
        align: 'center',
      },
      {
        key: 'instanceStatus',
        title: '实例状态',
        align: 'center',
        width: 110,
        render: (row) =>
          h(
            RsTag,
            {
              variant: getInstanceStatusVariant(row.instanceStatus),
              size: 'sm',
            },
            () => getInstanceStatusText(row.instanceStatus),
          ),
      },
      {
        key: 'isRunning',
        title: '运行状态',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.instanceStatus === 'RUNNING' ? 'success' : 'default',
              size: 'sm',
            },
            () => (row.instanceStatus === 'RUNNING' ? '运行中' : '已停止'),
          ),
      },
      {
        key: 'enableTLS',
        title: 'TLS',
        align: 'center',
        width: 90,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.enableTLS === 'Y' ? 'success' : 'default',
              size: 'sm',
            },
            () => (row.enableTLS === 'Y' ? '启用' : '禁用'),
          ),
      },
      {
        key: 'enableAuth',
        title: '认证',
        align: 'center',
        width: 90,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.enableAuth === 'Y' ? 'warning' : 'default',
              size: 'sm',
            },
            () => (row.enableAuth === 'Y' ? '启用' : '禁用'),
          ),
      },
      {
        key: 'activeFlag',
        title: '活动状态',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.activeFlag === 'Y' ? 'success' : 'default',
              size: 'sm',
            },
            () => (row.activeFlag === 'Y' ? '活动' : '非活动'),
          ),
      },
      {
        key: 'statusMessage',
        title: '状态消息',
        align: 'left',
        ellipsis: true,
        width: 200,
        formatter: (value) => (value ? String(value) : '-'),
      },
      {
        key: 'lastStatusTime',
        title: '最后状态变更时间',
        sortable: true,
        align: 'center',
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-',
      },
      {
        key: 'lastHealthCheckTime',
        title: '最后健康检查时间',
        sortable: true,
        align: 'center',
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-',
      },
      {
        key: 'addTime',
        title: '创建时间',
        sortable: true,
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        key: 'addWho',
        title: '创建人',
        ellipsis: true,
      },
      {
        key: 'editTime',
        title: '修改时间',
        sortable: true,
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        key: 'editWho',
        title: '修改人',
        ellipsis: true,
      },
    ],
    selectable: true,
    rowKey: 'oprSeqFlag',
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      items: [
        { key: 'view', label: '查看详情', icon: 'eye' },
        { key: 'edit', label: '编辑', icon: 'pencil' },
        { key: 'start', label: '启动', icon: 'play' },
        { key: 'stop', label: '停止', icon: 'square' },
        { key: 'reload', label: '重载配置', icon: 'refresh-cw' },
        { key: 'delete', label: '删除', icon: 'trash-2', danger: true },
      ],
    },
    height: '100%',
  }

  // ============= 辅助方法 =============

  /**
   * 重置分页
   */
  const resetPagination = () => {
    pageInfo.value = undefined
  }

  /**
   * 更新分页信息（接收后端 PageInfoObj）
   */
  const updatePagination = (newPageInfo: Partial<PageInfoObj>) => {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  /**
   * 设置实例列表
   */
  const setInstanceList = (list: ServiceCenterInstance[]) => {
    instanceList.value = list
  }

  /**
   * 清空实例列表
   */
  const clearInstanceList = () => {
    instanceList.value = []
  }

  /**
   * 添加实例到列表
   */
  const addInstanceToList = (instance: ServiceCenterInstance) => {
    instanceList.value.unshift(instance)
  }

  /**
   * 更新列表中的实例
   */
  const updateInstanceInList = (
    instanceName: string,
    environment: string,
    tenantId: string,
    updatedInstance: Partial<ServiceCenterInstance>
  ) => {
    const index = instanceList.value.findIndex(
      (i) => i.instanceName === instanceName && i.environment === environment && i.tenantId === tenantId
    )
    if (index !== -1) {
      Object.assign(instanceList.value[index], updatedInstance)
    }
  }

  /**
   * 从列表中删除实例
   */
  const removeInstanceFromList = (
    instanceName: string,
    environment: string,
    tenantId: string
  ) => {
    const index = instanceList.value.findIndex(
      (i) => i.instanceName === instanceName && i.environment === environment && i.tenantId === tenantId
    )
    if (index !== -1) {
      instanceList.value.splice(index, 1)
    }
  }

  /**
   * 批量删除实例
   */
  const removeInstancesFromList = (instances: ServiceCenterInstance[]) => {
    instances.forEach((instance) => {
      removeInstanceFromList(instance.instanceName, instance.environment, instance.tenantId)
    })
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    instanceList,
    pageInfo,

    // 配置
    searchFormConfig,
    instanceFormConfig,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setInstanceList,
    clearInstanceList,
    addInstanceToList,
    updateInstanceInList,
    removeInstanceFromList,
    removeInstancesFromList,
  }
}

/**
 * Model 返回类型
 */
export type ServiceCenterInstanceModel = ReturnType<typeof useServiceCenterInstanceModel>

