/**
 * 网关实例管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import {
  AddOutline,
  CreateOutline,
  DocumentOutline,
  GlobeOutline,
  KeyOutline,
  TrashOutline
} from '@vicons/ionicons5'
import { NDynamicTags, NIcon } from 'naive-ui'
import { h, ref } from 'vue'
import { AlertChannelNameSelector } from '../../hub0080/components'
import type { GatewayInstance } from '../types/index'

/**
 * 网关实例管理 Model
 */
export function useGatewayInstanceModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0020'
  /** 加载状态 */
  const loading = ref(false)

  /** 网关实例列表数据 */
  const instanceList = ref<GatewayInstance[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
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
        field: 'healthStatus',
        label: '健康状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '健康', value: 'Y' },
          { label: '不健康', value: 'N' },
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
        icon: AddOutline,
        type: 'primary',
        tooltip: '新建网关实例',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: CreateOutline,
        tooltip: '编辑选中的实例',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '删除选中的实例',
      }
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 实例表单配置（聚拢配置，类似 searchFormConfig） =============
  const instanceFormConfig = {
    tabs: [
      { key: 'basic', label: '基本信息' },
      { key: 'tls', label: 'TLS配置' },
      { key: 'performance', label: '性能配置' },
      { key: 'other', label: '其它' },
    ],
    fields: [
    // ============= 基本信息 Tab =============
    {
      field: 'instanceName',
      label: '实例名称',
      type: 'input',
      placeholder: '请输入实例名称',
      span: 12,
      tabKey: 'basic',
      required: true,
    },
    {
      field: 'bindAddress',
      label: '绑定地址',
      type: 'input',
      placeholder: '如: 0.0.0.0',
      span: 12,
      tabKey: 'basic',
      defaultValue: '0.0.0.0',
      tips: '服务器绑定的网络地址，0.0.0.0表示监听所有网络接口',
    },
    {
      field: 'httpPort',
      label: 'HTTP端口',
      type: 'number',
      placeholder: '如: 8080',
      span: 12,
      tabKey: 'basic',
      defaultValue: 8080,
      props: {
        min: 1,
        max: 65535,
      },
    },
    {
      field: 'httpsPort',
      label: 'HTTPS端口',
      type: 'number',
      placeholder: '如: 8443',
      span: 12,
      tabKey: 'basic',
      defaultValue: 8443,
      tips: 'HTTPS服务监听的端口号，范围1-65535，需要启用TLS后生效',
      props: {
        min: 1,
        max: 65535,
      },
    },
    {
      field: 'activeFlag',
      label: '自动启动',
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
      field: 'instanceDesc',
      label: '实例描述',
      type: 'textarea',
      placeholder: '请输入实例描述',
      span: 24,
      tabKey: 'basic',
      props: {
        rows: 3,
      },
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
          field: 'tlsEnabled',
          label: '启用TLS',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '启用HTTPS/TLS加密传输，保护数据传输安全',
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
          field: 'tlsVersion',
          label: 'TLS版本',
          type: 'select',
          placeholder: '选择支持的TLS版本',
          span: 12,
          defaultValue: ['1.2'],
          tips: '选择支持的TLS协议版本，建议至少启用TLS 1.2，TLS 1.0和1.1已不安全',
          props: {
            multiple: true,
            maxTagCount: 2,
          },
          options: [
            { label: 'TLS 1.0', value: '1.0' },
            { label: 'TLS 1.1', value: '1.1' },
            { label: 'TLS 1.2', value: '1.2' },
            { label: 'TLS 1.3', value: '1.3' },
          ],
        },
        {
          field: 'enableHttp2',
          label: '启用HTTP/2',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '启用HTTP/2协议，支持多路复用、头部压缩等特性，提升性能（需要TLS支持）',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'tlsCipherSuites',
          label: 'TLS密码套件',
          type: 'input',
          placeholder: '多个密码套件用逗号分隔，如: TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256，留空使用默认配置',
          span: 24,
          tips: '指定允许的TLS密码套件，留空使用系统默认的安全配置，建议使用强加密套件',
          clearable: true,
        },
        {
          field: 'certFileList',
          label: '证书文件',
          type: 'file',
          span: 12,
          props: {
            title: '证书文件',
            titleIcon: DocumentOutline,
            titleIconColor: '#18a058',
            showDownload: true,
            config: {
              accept: '.crt,.pem,.cer',
              max: 1,
              maxSize: 10 * 1024 * 1024, // 10MB
              mode: 'text',
              uploadText: '点击或拖拽上传证书',
              uploadDescription: '支持 .crt, .pem, .cer',
            },
          },
        },
        {
          field: 'keyFileList',
          label: '私钥文件',
          type: 'file',
          span: 12,
          props: {
            title: '私钥文件',
            titleIcon: KeyOutline,
            titleIconColor: '#f0a020',
            showDownload: true,
            config: {
              accept: '.key,.pem',
              max: 1,
              maxSize: 10 * 1024 * 1024, // 10MB
              mode: 'text',
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
      label: '性能配置',
      type: 'fieldset',
      tabKey: 'performance',
      children: [
        {
          field: 'maxConnections',
          label: '最大连接数',
          type: 'number',
          placeholder: '10000',
          span: 12,
          defaultValue: 10000,
          tips: '同时处理的最大并发连接数，超过此值的新连接将被拒绝',
          props: {
            min: 100,
            max: 100000,
          },
        },
        {
          field: 'maxWorkers',
          label: '最大工作协程数',
          type: 'number',
          placeholder: '1000',
          span: 12,
          defaultValue: 1000,
          tips: '用于处理请求的最大工作协程数，影响并发处理能力',
          props: {
            min: 100,
            max: 10000,
          },
        },
        {
          field: 'readTimeoutMs',
          label: '读取超时(ms)',
          type: 'number',
          placeholder: '30000',
          span: 12,
          defaultValue: 30000,
          tips: '从客户端读取请求头的最大时间，超时则关闭连接',
          props: {
            min: 1000,
            max: 300000,
          },
        },
        {
          field: 'writeTimeoutMs',
          label: '写入超时(ms)',
          type: 'number',
          placeholder: '30000',
          span: 12,
          defaultValue: 30000,
          tips: '向客户端写入响应的最大时间，超时则关闭连接',
          props: {
            min: 1000,
            max: 300000,
          },
        },
        {
          field: 'idleTimeoutMs',
          label: '空闲超时(ms)',
          type: 'number',
          placeholder: '60000',
          span: 12,
          defaultValue: 60000,
          tips: '保持连接空闲的最大时间，超时则关闭连接（用于HTTP Keep-Alive）',
          props: {
            min: 1000,
            max: 600000,
          },
        },
        {
          field: 'maxHeaderBytes',
          label: '最大请求头大小(字节)',
          type: 'number',
          placeholder: '1048576',
          span: 12,
          defaultValue: 1048576,
          tips: '限制单个请求头的最大字节数，防止恶意请求头过大，对应 http.Server.MaxHeaderBytes',
          props: {
            min: 1024,
            max: 10485760,
          },
        },
      ],
    },
    {
      field: 'keepalive-config-group',
      label: 'Keep-Alive配置',
      type: 'fieldset',
      tabKey: 'performance',
      children: [
        {
          field: 'keepAliveEnabled',
          label: 'HTTP Keep-Alive',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '启用HTTP Keep-Alive，允许在同一个TCP连接上发送多个HTTP请求，提高性能',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'tcpKeepAliveEnabled',
          label: 'TCP Keep-Alive',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '启用TCP Keep-Alive，定期发送探测包检测连接是否存活，自动清理僵尸连接',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
      ],
    },

    // ============= 其它 Tab =============
    {
      field: 'noteText',
      label: '备注',
      type: 'textarea',
      placeholder: '请输入备注信息',
      span: 24,
      tabKey: 'other',
      props: {
        rows: 3,
      },
    },
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
  ] as DataFormField[],
  }

  // ============= 日志配置表单配置（多标签页管理） =============
  const logConfigFormConfig = {
    tabs: [
      { key: 'basic', label: '基础配置' },
      { key: 'alert', label: '预警配置' },
      { key: 'other', label: '其它' },
    ],
    fields: [
      // ============= 主键字段（隐藏，但必须存在用于更新） =============
      {
        field: 'logConfigId',
        label: '日志配置ID',
        type: 'input',
        span: 12,
        show: false, // 隐藏字段，但必须存在用于更新
      },
      {
        field: 'tenantId',
        label: '租户ID',
        type: 'input',
        span: 12,
        show: false, // 隐藏字段，但必须存在用于更新
      },
      // ============= 基础配置 Tab =============
      {
        field: 'configName',
        label: '日志配置名称',
        type: 'input',
        placeholder: '请输入日志配置名称',
        span: 12,
        tabKey: 'basic',
        required: true,
      },
      {
        field: 'logFormat',
        label: '日志格式',
        type: 'select',
        placeholder: '选择日志格式',
        span: 12,
        tabKey: 'basic',
        defaultValue: 'JSON',
        options: [
          { label: 'JSON格式', value: 'JSON' },
          { label: '文本格式', value: 'TEXT' },
          { label: 'CSV格式', value: 'CSV' },
        ],
      },
      {
        field: 'configDesc',
        label: '日志配置描述',
        type: 'textarea',
        placeholder: '请输入日志配置描述',
        span: 24,
        tabKey: 'basic',
        props: {
          rows: 2,
        },
      },

      // ============= 内容控制 =============
      {
        field: 'content-control-group',
        label: '日志内容控制',
        type: 'fieldset',
        tabKey: 'basic',
        props: {
          titleSize: 300,
        },
        children: [
          {
            field: 'recordRequestBody',
            label: '记录请求体',
            type: 'switch',
            span: 12,
            defaultValue: 'N',
            props: {
              checkedValue: 'Y',
              uncheckedValue: 'N',
            },
          },
          {
            field: 'recordResponseBody',
            label: '记录响应体',
            type: 'switch',
            span: 12,
            defaultValue: 'N',
            props: {
              checkedValue: 'Y',
              uncheckedValue: 'N',
            },
          },
          {
            field: 'recordHeaders',
            label: '记录请求头',
            type: 'switch',
            span: 12,
            defaultValue: 'Y',
            props: {
              checkedValue: 'Y',
              uncheckedValue: 'N',
            },
          },
          {
            field: 'maxBodySizeBytes',
            label: '最大报文大小(字节)',
            type: 'number',
            placeholder: '1048576',
            span: 12,
            defaultValue: 1048576,
            props: {
              min: 0,
            },
          },
        ],
      },

      // ============= 输出目标 =============
      {
        field: 'output-target-group',
        label: '日志输出目标',
        type: 'fieldset',
        tabKey: 'basic',
        props: {
          titleSize: 300,
        },
        children: [
          {
            field: 'outputTargets',
            label: '输出目标',
            type: 'select',
            placeholder: '选择日志输出目标',
            span: 24, // 单列布局，与原对话框一致
            defaultValue: 'DATABASE',
            options: [
              { label: '文件输出', value: 'FILE' },
              { label: '数据库输出', value: 'DATABASE' },
              { label: 'MongoDB输出', value: 'MONGODB' },
              { label: 'Elasticsearch输出', value: 'ELASTICSEARCH' },
              { label: 'ClickHouse输出', value: 'CLICKHOUSE' },
            ],
          },
          {
            field: 'fileConfig.filePath',
            label: '文件路径',
            type: 'input',
            placeholder: './logs/gateway.log',
            span: 12,
            defaultValue: './logs/gateway.log',
            // 动态控制显示：只有当 outputTargets 包含 'FILE' 时才显示
            show: (formData: Record<string, any>) => {
              const outputTargets = formData.outputTargets || ''
              return typeof outputTargets === 'string' && outputTargets.includes('FILE')
            },
          },
          {
            field: 'logRetentionDays',
            label: '文件保留天数',
            type: 'number',
            placeholder: '30',
            span: 12,
            defaultValue: 30,
            props: {
              min: 1,
            },
            // 动态控制显示：只有当 outputTargets 包含 'FILE' 时才显示
            show: (formData: Record<string, any>) => {
              const outputTargets = formData.outputTargets || ''
              return typeof outputTargets === 'string' && outputTargets.includes('FILE')
            },
          },
        ],
      },

      // ============= 异步处理 =============
      {
        field: 'async-processing-group',
        label: '异步和批量处理',
        type: 'fieldset',
        tabKey: 'basic',
        props: {
          titleSize: 300,
        },
        children: [
          {
            field: 'enableAsyncLogging',
            label: '启用异步日志',
            type: 'switch',
            span: 12,
            defaultValue: 'Y',
            props: {
              checkedValue: 'Y',
              uncheckedValue: 'N',
            },
          },
          {
            field: 'asyncQueueSize',
            label: '异步队列大小',
            type: 'number',
            placeholder: '1000',
            span: 12,
            defaultValue: 1000,
            props: {
              min: 1,
            },
          },
          {
            field: 'batchSize',
            label: '批处理大小',
            type: 'number',
            placeholder: '100',
            span: 12,
            defaultValue: 100,
            props: {
              min: 1,
            },
          },
          {
            field: 'asyncFlushIntervalMs',
            label: '异步刷新间隔(ms)',
            type: 'number',
            placeholder: '5000',
            span: 12,
            defaultValue: 5000,
            props: {
              min: 100,
              max: 60000,
            },
          },
          {
            field: 'enableBatchProcessing',
            label: '启用批量处理',
            type: 'switch',
            span: 12,
            defaultValue: 'Y',
            props: {
              checkedValue: 'Y',
              uncheckedValue: 'N',
            },
          },
          {
            field: 'batchTimeoutMs',
            label: '批处理超时(ms)',
            type: 'number',
            placeholder: '1000',
            span: 12,
            defaultValue: 1000,
            props: {
              min: 1000,
              max: 300000,
            },
          },
        ],
      },

      // ============= 敏感数据 =============
      {
        field: 'sensitive-data-group',
        label: '敏感数据处理',
        type: 'fieldset',
        tabKey: 'basic',
        props: {
          titleSize: 300,
        },
        children: [
          {
            field: 'enableSensitiveDataMasking',
            label: '启用敏感数据脱敏',
            type: 'switch',
            span: 12,
            defaultValue: 'N',
            props: {
              checkedValue: 'Y',
              uncheckedValue: 'N',
            },
          },
          {
            field: 'maskingPattern',
            label: '脱敏模式',
            type: 'select',
            placeholder: '选择脱敏模式',
            span: 12,
            defaultValue: '****',
            options: [
              { label: '替换为*', value: '****' },
              { label: '替换为[MASKED]', value: '[MASKED]' },
              { label: '保留前后字符', value: 'KEEP_EDGE' },
              { label: '自定义规则', value: 'CUSTOM' },
            ],
          },
          {
            field: 'sensitiveFields',
            label: '敏感字段',
            type: 'custom',
            span: 24,
            defaultValue: [],
            render: (formData: Record<string, any>) => {
              return h(NDynamicTags, {
                value: formData.sensitiveFields || [],
                'onUpdate:value': (value: string[]) => {
                  formData.sensitiveFields = value
                },
                placeholder: '添加敏感字段',
              })
            },
          },
        ],
      },

      // ============= 预警配置 Tab =============
      {
        field: 'alert-config-group',
        label: '告警配置',
        type: 'fieldset',
        tabKey: 'alert',
        props: {
          titleSize: 300,
        },
        children: [
          {
            field: 'extProperty.alertEnabled',
            label: '开启告警',
            type: 'switch',
            span: 12,
            defaultValue: 'N',
            tips: '开启后才会对 404 / 超时 等事件触发告警',
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
            render: (formData: Record<string, any>) => {
              return h(AlertChannelNameSelector, {
                // 扁平字段（参考 common002/auth-config）：直接读写 formData['extProperty.xxx.yyy']
                modelValue: formData['extProperty.channelName'] || '',
                'onUpdate:modelValue': (value: string) => {
                  formData['extProperty.channelName'] = value
                },
              })
            },
          },
          {
            field: 'extProperty.alertStatusCodes',
            label: '状态码告警',
            type: 'custom',
            span: 24,
            defaultValue: ['502'],
            tips: '选择需要告警的HTTP状态码，多个状态码用逗号分隔',
            render: (formData: Record<string, any>) => {
              // 将数组转换为字符串数组（如果后端返回的是数字数组）
              const value = formData['extProperty.alertStatusCodes'] || []
              const strArray = Array.isArray(value) 
                ? value.map(v => String(v))
                : (typeof value === 'string' ? value.split(',').map(s => s.trim()).filter(Boolean) : [])
              
              return h(NDynamicTags, {
                value: strArray,
                'onUpdate:value': (value: string[]) => {
                  formData['extProperty.alertStatusCodes'] = value
                },
                placeholder: '输入状态码，如: 502, 503, 504',
              })
            },
          },
          {
            field: 'extProperty.alertOnTimeout',
            label: '超时告警',
            type: 'switch',
            span: 8,
            defaultValue: 'Y',
            props: {
              checkedValue: 'Y',
              uncheckedValue: 'N',
            },
          },
          {
            field: 'extProperty.timeoutThresholdMs',
            label: '超时阈值(ms)',
            type: 'number',
            span: 8,
            defaultValue: 120000,
            tips: '当总耗时 >= 阈值（毫秒）时触发超时告警',
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
    ] as DataFormField[],
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'gatewayInstanceId',
        title: '实例ID',
        sortable: true,
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'instanceName',
        title: '实例名称',
        sortable: true,
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'instanceDesc',
        title: '实例描述',
        align: 'center',
        showOverflow: 'tooltip',
      },
      {
        field: 'bindAddress',
        title: '绑定地址',
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'httpPort',
        title: 'HTTP端口',
        align: 'center',
      },
      {
        field: 'httpsPort',
        title: 'HTTPS端口',
        align: 'center',
      },
      {
        field: 'tlsEnabled',
        title: 'TLS',
        align: 'center',
        // 使用插槽方式支持自定义渲染（包括样式、标签等）
        // 在 GatewayInstanceManager.vue 中使用 <template #tlsEnabled="{ row }"> 来定义渲染内容
        slots: { default: 'tlsEnabled' },
        // 备选方案1：使用 formatter（简单文本格式化）
        // formatter: ({ cellValue }) => {
        //   return cellValue === 'Y' ? '启用' : '禁用'
        // },
        // 备选方案2：使用 cellRender（vxe-table 内置渲染器）
        // cellRender: {
        //   name: 'VxeTag',
        //   props: ({ row }: any) => ({
        //     type: row.tlsEnabled === 'Y' ? 'success' : 'default',
        //     content: row.tlsEnabled === 'Y' ? '启用' : '禁用',
        //   }),
        // },
      },
      {
        field: 'maxConnections',
        title: '最大连接数',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue ? cellValue.toLocaleString() : '0'
        },
      },
      {
        field: 'healthStatus',
        title: '健康状态',
        align: 'center',
        // 使用插槽方式支持自定义渲染（包括样式、标签等）
        slots: { default: 'healthStatus' },
        // 备选方案1：使用 formatter（简单文本格式化）
        // formatter: ({ cellValue }) => {
        //   return cellValue === 'Y' ? '健康' : '不健康'
        // },
        // 备选方案2：使用 cellRender（vxe-table 内置渲染器）
        // cellRender: {
        //   name: 'VxeTag',
        //   props: ({ row }: any) => ({
        //     type: row.healthStatus === 'Y' ? 'success' : 'error',
        //     content: row.healthStatus === 'Y' ? '健康' : '不健康',
        //   }),
        // },
      },
      {
        field: 'activeFlag',
        title: '活动状态',
        align: 'center',
        // 使用插槽方式支持自定义渲染（包括样式、标签等）
        slots: { default: 'activeFlag' },
        // 备选方案1：使用 formatter（简单文本格式化）
        // formatter: ({ cellValue }) => {
        //   return cellValue === 'Y' ? '活动' : '非活动'
        // },
        // 备选方案2：使用 cellRender（vxe-table 内置渲染器）
        // cellRender: {
        //   name: 'VxeTag',
        //   props: ({ row }: any) => ({
        //     type: row.activeFlag === 'Y' ? 'success' : 'default',
        //     content: row.activeFlag === 'Y' ? '活动' : '非活动',
        //   }),
        // },
      },
      {
        field: 'addTime',
        title: '创建时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'addWho',
        title: '创建人',
        showOverflow: true,
      },
      {
        field: 'editTime',
        title: '修改时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'editWho',
        title: '修改人',
        showOverflow: true,
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
      showCopyCell: true,
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
          code: 'delete',
          name: '删除',
          prefixIcon: 'vxe-icon-delete',
        },
        {
          code: 'start',
          name: '启动',
          // 注意：vxe-table 可能不支持 play/pause 相关图标
          // 如果图标不显示，可以尝试以下方案：
          // 1. 移除 prefixIcon 属性，只显示文字
          // 2. 使用其他可用的 vxe-icon 图标（如 vxe-icon-caret-right）
          // 3. 后续可以通过插槽方式使用自定义图标组件
          prefixIcon: 'vxe-icon-caret-right', // 右箭头图标，表示启动
        },
        {
          code: 'stop',
          name: '停止',
          // 注意：vxe-table 可能不支持 play/pause 相关图标
          // 如果图标不显示，可以尝试以下方案：
          // 1. 移除 prefixIcon 属性，只显示文字
          // 2. 使用其他可用的 vxe-icon 图标（如 vxe-icon-square）
          // 3. 后续可以通过插槽方式使用自定义图标组件
          prefixIcon: 'vxe-icon-square', // 方块图标，表示停止
        },
        {
          code: 'globalConfig',
          name: '全局配置',
          prefixIcon: 'vxe-icon-setting',
          children: [
            {
              code: 'ipAccessControl',
              name: 'IP访问控制',
              prefixIcon: 'vxe-icon-lock',
            },
            {
              code: 'userAgentAccessControl',
              name: 'User-Agent访问控制',
              prefixIcon: 'vxe-icon-user',
            },
            {
              code: 'apiAccessControl',
              name: 'API访问控制',
              // vxe-table 菜单只支持内置图标类名，不支持自定义扩展
              // 可选的内置图标：vxe-icon-code（代码）、vxe-icon-link（链接）、vxe-icon-setting（设置）等
              // 如果不需要图标，可以移除 prefixIcon 属性
              prefixIcon: 'vxe-icon-link', // 使用代码图标表示 API 接口
            },
            {
              code: 'domainAccessControl',
              name: '域名访问控制',
              // prefixIcon 支持 VNode 类型，可以使用函数返回自定义图标
              prefixIcon: () => h(NIcon, { size: 12 }, { default: () => h(GlobeOutline) }),
            },
            {
              code: 'corsConfig',
              name: '跨域配置',
              prefixIcon: 'vxe-icon-link',
            },
            {
              code: 'authConfig',
              name: '认证配置',
              prefixIcon: 'vxe-icon-setting',
            },
            {
              code: 'rateLimitConfig',
              name: '限流配置',
              prefixIcon: 'vxe-icon-setting',
            },
          ],
        },
        {
          code: 'logConfig',
          name: '日志配置',
          prefixIcon: 'vxe-icon-setting',
        },
        {
          code: 'reload',
          name: '网关重载',
          prefixIcon: 'vxe-icon-refresh',
        },
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
  const setInstanceList = (list: GatewayInstance[]) => {
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
  const addInstanceToList = (instance: GatewayInstance) => {
    instanceList.value.unshift(instance)
  }

  /**
   * 更新列表中的实例
   */
  const updateInstanceInList = (gatewayInstanceId: string, tenantId: string, updatedInstance: Partial<GatewayInstance>) => {
    const index = instanceList.value.findIndex(
      (i) => i.gatewayInstanceId === gatewayInstanceId && i.tenantId === tenantId
    )
    if (index !== -1) {
      Object.assign(instanceList.value[index], updatedInstance)
    }
  }

  /**
   * 从列表中删除实例
   */
  const removeInstanceFromList = (gatewayInstanceId: string, tenantId: string) => {
    const index = instanceList.value.findIndex(
      (i) => i.gatewayInstanceId === gatewayInstanceId && i.tenantId === tenantId
    )
    if (index !== -1) {
      instanceList.value.splice(index, 1)
    }
  }

  /**
   * 批量删除实例
   */
  const removeInstancesFromList = (instances: GatewayInstance[]) => {
    instances.forEach((instance) => {
      removeInstanceFromList(instance.gatewayInstanceId, instance.tenantId)
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
    logConfigFormConfig,
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
export type GatewayInstanceModel = ReturnType<typeof useGatewayInstanceModel>
