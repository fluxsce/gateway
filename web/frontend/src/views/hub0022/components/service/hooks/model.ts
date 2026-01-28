/**
 * 服务定义列表管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField, DataFormTab } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import type { ServiceSelectionMetadata } from '@/views/hub0042/components'
import { ServiceSelector } from '@/views/hub0042/components'
import { AddOutline, SettingsOutline, TrashOutline } from '@vicons/ionicons5'
import { h, ref } from 'vue'
import type { ServiceDefinition } from '../types'
import { LoadBalanceStrategy, ServiceType } from '../types'

/**
 * 服务定义列表管理 Model
 */
export function useServiceDefinitionModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0022'
  
  /** 加载状态 */
  const loading = ref(false)

  /** 服务定义列表数据 */
  const serviceList = ref<ServiceDefinition[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'serviceName',
        label: '服务名称',
        type: 'input',
        placeholder: '请输入服务名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'serviceType',
        label: '服务类型',
        type: 'select',
        placeholder: '请选择服务类型',
        span: 6,
        clearable: true,
        options: [
          { label: '静态配置', value: ServiceType.STATIC },
          { label: '服务发现', value: ServiceType.DISCOVERY },
        ],
      },
      {
        field: 'loadBalanceStrategy',
        label: '负载均衡策略',
        type: 'select',
        placeholder: '请选择负载均衡策略',
        span: 6,
        clearable: true,
        options: [
          { label: '轮询算法', value: LoadBalanceStrategy.ROUND_ROBIN },
          { label: '随机算法', value: LoadBalanceStrategy.RANDOM },
          { label: 'IP哈希算法', value: LoadBalanceStrategy.IP_HASH },
          { label: '最少连接算法', value: LoadBalanceStrategy.LEAST_CONN },
          { label: '加权轮询算法', value: LoadBalanceStrategy.WEIGHTED_ROUND_ROBIN },
          { label: '一致性哈希算法', value: LoadBalanceStrategy.CONSISTENT_HASH },
        ],
      },
      {
        field: 'activeFlag',
        label: '状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        options: [
          { label: '全部', value: '' },
          { label: '启用', value: 'Y' },
          { label: '禁用', value: 'N' },
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增服务',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新增服务定义',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '批量删除选中的服务定义',
      },
      {
        key: 'manageNodes',
        label: '节点管理',
        icon: SettingsOutline,
        type: 'default',
        tooltip: '管理服务节点',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 服务表单配置（聚拢配置，类似 searchFormConfig） =============
  const serviceFormConfig = {
    tabs: [
      { 
        key: 'basic', 
        label: '基本信息',
        // 基本信息标签页始终显示
      },
      { 
        key: 'loadbalance', 
        label: '负载均衡',
        // 只在服务类型为静态配置时显示
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC,
      },
      { 
        key: 'health', 
        label: '健康检查',
        // 只在服务类型为静态配置时显示
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC,
      },
      { 
        key: 'discovery', 
        label: '服务发现',
        // 只在服务类型为服务发现时显示
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.DISCOVERY,
      },
      { 
        key: 'other', 
        label: '其他配置',
        // 其他配置标签页始终显示
      },
    ] as DataFormTab[],
    fields: [
      // ============= 主键字段（隐藏，但必须存在用于编辑） =============
      {
        field: 'serviceDefinitionId',
        label: '服务定义ID',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        primary: true, // 标记为主键字段，编辑模式下自动禁用
        show: false, // 隐藏字段，但必须存在用于更新
      },
      {
        field: 'tenantId',
        label: '租户ID',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        show: false, // 隐藏字段，但必须存在用于更新
      },
      {
        field: 'proxyConfigId',
        label: '代理配置ID',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        show: false, // 隐藏字段，通过 props 传入
      },
      // ============= 基本信息 Tab =============
      {
        field: 'serviceName',
        label: '服务名称',
        type: 'input',
        placeholder: '请输入服务名称',
        span: 12,
        tabKey: 'basic',
        required: true,
        tips: '服务的唯一标识名称，用于区分不同的后端服务，长度2-50个字符',
        rules: [
          { required: true, message: '请输入服务名称', trigger: ['blur', 'input'] },
          { min: 2, max: 50, message: '服务名称长度在2-50个字符', trigger: ['blur', 'input'] },
        ],
      },
      {
        field: 'serviceType',
        label: '服务类型',
        type: 'select',
        placeholder: '请选择服务类型',
        span: 12,
        tabKey: 'basic',
        required: true,
        defaultValue: ServiceType.STATIC,
        tips: '静态配置：手动配置服务节点地址；服务发现：从服务注册中心自动发现服务实例',
        options: [
          { label: '静态配置', value: ServiceType.STATIC },
          { label: '服务发现', value: ServiceType.DISCOVERY },
        ],
        rules: [
          {
            required: true,
            message: '请选择服务类型',
            trigger: ['blur', 'change'],
            validator: (_rule: any, value: any) => {
              if (value === null || value === undefined || value === '') {
                return new Error('请选择服务类型')
              }
              if (typeof value === 'number' && (value === 0 || value === 1)) {
                return true
              }
              return new Error('请选择有效的服务类型')
            },
          },
        ],
      },
      {
        field: 'serviceDesc',
        label: '服务描述',
        type: 'textarea',
        placeholder: '请输入服务描述',
        span: 24,
        tabKey: 'basic',
        tips: '服务的详细描述信息，用于说明服务的用途和功能',
        props: {
          rows: 3,
        },
      },

      // ============= 负载均衡 Tab =============
      {
        field: 'loadBalanceStrategy',
        label: '负载均衡策略',
        type: 'select',
        placeholder: '请选择负载均衡策略',
        span: 12,
        tabKey: 'loadbalance',
        required: true,
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC,
        defaultValue: LoadBalanceStrategy.ROUND_ROBIN,
        tips: '轮询：按顺序依次分发请求；随机：随机选择节点；IP哈希：根据客户端IP哈希选择节点；最少连接：选择连接数最少的节点；加权轮询：根据节点权重轮询；一致性哈希：保证相同请求路由到同一节点',
        options: [
          { label: '轮询算法', value: LoadBalanceStrategy.ROUND_ROBIN },
          { label: '随机算法', value: LoadBalanceStrategy.RANDOM },
          { label: 'IP哈希算法', value: LoadBalanceStrategy.IP_HASH },
          { label: '最少连接算法', value: LoadBalanceStrategy.LEAST_CONN },
          { label: '加权轮询算法', value: LoadBalanceStrategy.WEIGHTED_ROUND_ROBIN },
          { label: '一致性哈希算法', value: LoadBalanceStrategy.CONSISTENT_HASH },
        ],
        rules: [
          {
            required: true,
            message: '请选择负载均衡策略',
            trigger: ['blur', 'change'],
            validator: (_rule: any, value: any, formData: Record<string, any>) => {
              if (formData.serviceType === ServiceType.STATIC && (value === null || value === undefined || value === '')) {
                return new Error('请选择负载均衡策略')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'sessionAffinity',
        label: '会话亲和性',
        type: 'switch',
        span: 12,
        tabKey: 'loadbalance',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC,
        defaultValue: 'N',
        tips: '启用后，同一客户端的请求会尽量路由到同一个后端节点，保证会话状态的一致性',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'stickySession',
        label: '粘性会话',
        type: 'switch',
        span: 12,
        tabKey: 'loadbalance',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC,
        defaultValue: 'N',
        tips: '启用后，通过Cookie等方式强制将同一会话的请求路由到固定的后端节点',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'maxRetries',
        label: '最大重试次数',
        type: 'number',
        placeholder: '3',
        span: 12,
        tabKey: 'loadbalance',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC,
        defaultValue: 3,
        tips: '当请求失败时，最多重试的次数。范围0-10，0表示不重试',
        props: {
          min: 0,
          max: 10,
        },
        rules: [
          {
            required: true,
            message: '请输入最大重试次数',
            trigger: ['blur', 'change'],
            validator: (_rule: any, value: any) => {
              if (value === null || value === undefined || value === '') {
                return new Error('请输入最大重试次数')
              }
              const num = Number(value)
              if (isNaN(num) || num < 0 || num > 10) {
                return new Error('重试次数必须在0-10之间')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'retryTimeoutMs',
        label: '重试超时时间(ms)',
        type: 'number',
        placeholder: '5000',
        span: 12,
        tabKey: 'loadbalance',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC,
        defaultValue: 5000,
        tips: '每次重试请求的超时时间（毫秒）。范围100-60000ms',
        props: {
          min: 100,
          max: 60000,
        },
        rules: [
          {
            required: true,
            message: '请输入重试超时时间',
            trigger: ['blur', 'change'],
            validator: (_rule: any, value: any) => {
              if (value === null || value === undefined || value === '') {
                return new Error('请输入重试超时时间')
              }
              const num = Number(value)
              if (isNaN(num) || num < 100 || num > 60000) {
                return new Error('超时时间必须在100-60000毫秒之间')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'enableCircuitBreaker',
        label: '启用熔断器',
        type: 'switch',
        span: 12,
        tabKey: 'loadbalance',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC,
        defaultValue: 'N',
        tips: '启用后，当服务节点故障率过高时自动熔断，避免大量请求打到故障节点',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },

      // ============= 健康检查 Tab =============
      {
        field: 'healthCheckEnabled',
        label: '启用健康检查',
        type: 'switch',
        span: 24,
        tabKey: 'health',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC,
        defaultValue: 'N',
        tips: '启用后，网关会定期检查后端节点的健康状态，自动剔除不健康的节点',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'healthCheckPath',
        label: '检查路径',
        type: 'input',
        placeholder: '/health',
        span: 12,
        tabKey: 'health',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC && formData.healthCheckEnabled === 'Y',
        defaultValue: '/health',
        tips: '健康检查请求的URL路径，通常是后端服务提供的健康检查接口',
        rules: [
          {
            required: true,
            message: '请输入健康检查路径',
            trigger: ['blur', 'input'],
            validator: (_rule: any, value: any, formData: Record<string, any>) => {
              if (formData.healthCheckEnabled === 'Y' && (!value || value.trim() === '')) {
                return new Error('请输入健康检查路径')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'healthCheckMethod',
        label: '检查方法',
        type: 'select',
        placeholder: '请选择检查方法',
        span: 12,
        tabKey: 'health',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC && formData.healthCheckEnabled === 'Y',
        defaultValue: 'GET',
        tips: '健康检查使用的HTTP方法，通常使用GET或HEAD方法',
        options: [
          { label: 'GET', value: 'GET' },
          { label: 'POST', value: 'POST' },
          { label: 'HEAD', value: 'HEAD' },
        ],
        rules: [
          {
            required: true,
            message: '请选择健康检查方法',
            trigger: ['blur', 'change'],
            validator: (_rule: any, value: any, formData: Record<string, any>) => {
              if (formData.healthCheckEnabled === 'Y' && (!value || value === '')) {
                return new Error('请选择健康检查方法')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'healthCheckIntervalSeconds',
        label: '检查间隔(秒)',
        type: 'number',
        placeholder: '30',
        span: 12,
        tabKey: 'health',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC && formData.healthCheckEnabled === 'Y',
        defaultValue: 30,
        tips: '两次健康检查之间的时间间隔（秒）。范围1-300秒',
        props: {
          min: 1,
          max: 300,
        },
        rules: [
          {
            required: true,
            message: '请输入检查间隔',
            trigger: ['blur', 'change'],
            validator: (_rule: any, value: any, formData: Record<string, any>) => {
              if (formData.healthCheckEnabled === 'Y' && (value === null || value === undefined || value === '')) {
                return new Error('请输入检查间隔')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'healthCheckTimeoutMs',
        label: '检查超时(ms)',
        type: 'number',
        placeholder: '5000',
        span: 12,
        tabKey: 'health',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC && formData.healthCheckEnabled === 'Y',
        defaultValue: 5000,
        tips: '健康检查请求的超时时间（毫秒），超时则视为检查失败。范围100-30000ms',
        props: {
          min: 100,
          max: 30000,
        },
        rules: [
          {
            required: true,
            message: '请输入检查超时',
            trigger: ['blur', 'change'],
            validator: (_rule: any, value: any, formData: Record<string, any>) => {
              if (formData.healthCheckEnabled === 'Y' && (value === null || value === undefined || value === '')) {
                return new Error('请输入检查超时')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'healthyThreshold',
        label: '健康阈值',
        type: 'number',
        placeholder: '2',
        span: 12,
        tabKey: 'health',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC && formData.healthCheckEnabled === 'Y',
        defaultValue: 2,
        tips: '连续成功检查次数达到此值时，节点从 unhealthy 转为 healthy。范围1-10',
        props: {
          min: 1,
          max: 10,
        },
        rules: [
          {
            required: true,
            message: '请输入健康阈值',
            trigger: ['blur', 'change'],
            validator: (_rule: any, value: any, formData: Record<string, any>) => {
              if (formData.healthCheckEnabled === 'Y' && (value === null || value === undefined || value === '')) {
                return new Error('请输入健康阈值')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'unhealthyThreshold',
        label: '不健康阈值',
        type: 'number',
        placeholder: '3',
        span: 12,
        tabKey: 'health',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC && formData.healthCheckEnabled === 'Y',
        defaultValue: 3,
        tips: '连续失败检查次数达到此值时，节点从 healthy 转为 unhealthy。范围1-10',
        props: {
          min: 1,
          max: 10,
        },
        rules: [
          {
            required: true,
            message: '请输入不健康阈值',
            trigger: ['blur', 'change'],
            validator: (_rule: any, value: any, formData: Record<string, any>) => {
              if (formData.healthCheckEnabled === 'Y' && (value === null || value === undefined || value === '')) {
                return new Error('请输入不健康阈值')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'expectedStatusCodes',
        label: '期望状态码',
        type: 'input',
        placeholder: '200,201,204',
        span: 24,
        tabKey: 'health',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.STATIC && formData.healthCheckEnabled === 'Y',
        defaultValue: '200',
        tips: '健康检查视为成功的HTTP状态码，支持逗号分隔多个值（如：200,201,204）或JSON数组格式',
        rules: [
          {
            required: true,
            message: '请输入期望状态码',
            trigger: ['blur', 'input'],
            validator: (_rule: any, value: any, formData: Record<string, any>) => {
              if (formData.healthCheckEnabled === 'Y' && (!value || value.trim() === '')) {
                return new Error('请输入期望状态码')
              }
              return true
            },
          },
        ],
      },

      // ============= 服务发现 Tab =============
      {
        field: 'discoveryType',
        label: '发现类型',
        type: 'select',
        placeholder: '请选择服务发现类型',
        span: 12,
        tabKey: 'discovery',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.DISCOVERY,
        disabled: true,
        defaultValue: 'REGISTRY',
        tips: '服务发现类型，当前仅支持从服务注册中心发现服务实例',
        options: [
          { label: '服务注册', value: 'REGISTRY' },
        ],
      },
      {
        field: 'serviceMetadata',
        label: '服务元数据',
        type: 'textarea',
        placeholder: '{}',
        span: 24,
        tabKey: 'discovery',
        show: false, // 隐藏字段，通过 ServiceRegistrySelector 选择服务后自动填充
        props: {
          rows: 5,
        },
        tips: '服务元数据，包含从服务注册中心选择的服务信息（JSON格式）',
      },
      {
        field: 'serviceSelection',
        label: '注册服务',
        type: 'custom' as const,
        span: 24,
        tabKey: 'discovery',
        show: (formData: Record<string, any>) => formData.serviceType === ServiceType.DISCOVERY,
        tips: '从服务注册中心选择一个可用的服务，选择后会自动填充服务元数据',
        render: (formData: Record<string, any>, context?: {
          selectedService?: { value: ServiceSelectionMetadata | null }
          onServiceChange?: (metadata: ServiceSelectionMetadata | null) => void
          to?: string
        }) => {
          const selectedService = context?.selectedService?.value || null
          const onServiceChange = context?.onServiceChange
          const to = context?.to || 'body'
          
          return h(ServiceSelector, {
            modelValue: selectedService,
            to: to,
            'onUpdate:modelValue': (metadata: ServiceSelectionMetadata | null) => {
              onServiceChange?.(metadata)
            },
            onChange: (metadata: ServiceSelectionMetadata | null) => {
              onServiceChange?.(metadata)
            }
          })
        },
      },
      {
        field: 'discoveryConfig',
        label: '发现配置',
        type: 'textarea',
        placeholder: '{}',
        span: 24,
        tabKey: 'discovery',
        show: false, // 隐藏字段，用于存储服务发现相关配置
        props: {
          rows: 5,
        },
      },

      // ============= 其他配置 Tab =============
      {
        field: 'activeFlag',
        label: '启用状态',
        type: 'switch',
        span: 12,
        tabKey: 'other',
        defaultValue: 'Y',
        tips: '控制服务定义是否启用，禁用的服务不会被加载到网关',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'noteText',
        label: '备注信息',
        type: 'textarea',
        placeholder: '请输入备注信息',
        span: 24,
        tabKey: 'other',
        tips: '服务的额外备注信息，用于记录配置说明或其他相关信息',
        props: {
          rows: 4,
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

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'serviceDefinitionId',
        title: '服务定义ID',
        visible: false, // 隐藏主键字段，但保留在数据中以便编辑时使用
        width: 0,
      },
      {
        field: 'serviceName',
        title: '服务名称',
        sortable: true,
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
      },
      {
        field: 'serviceDesc',
        title: '服务描述',
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
      },
      {
        field: 'serviceType',
        title: '服务类型',
        align: 'center',
        slots: { default: 'serviceType' },
        width: 120,
      },
      {
        field: 'loadBalanceStrategy',
        title: '负载均衡策略',
        align: 'center',
        slots: { default: 'loadBalanceStrategy' },
        width: 150,
      },
      {
        field: 'nodeCount',
        title: '服务节点',
        align: 'center',
        formatter: () => '0', // 暂时返回0，后续接入真实API后修改
        width: 100,
      },
      {
        field: 'sessionAffinity',
        title: '会话亲和性',
        align: 'center',
        slots: { default: 'sessionAffinity' },
        width: 120,
      },
      {
        field: 'maxRetries',
        title: '最大重试',
        align: 'center',
        width: 100,
      },
      {
        field: 'enableCircuitBreaker',
        title: '熔断器',
        align: 'center',
        slots: { default: 'enableCircuitBreaker' },
        width: 100,
      },
      {
        field: 'healthCheckEnabled',
        title: '健康检查',
        align: 'center',
        slots: { default: 'healthCheckEnabled' },
        width: 120,
      },
      {
        field: 'healthCheckPath',
        title: '检查路径',
        align: 'center',
        showOverflow: 'tooltip',
        width: 150,
      },
      {
        field: 'activeFlag',
        title: '状态',
        align: 'center',
        slots: { default: 'activeFlag' },
        width: 100,
      },
      {
        field: 'addTime',
        title: '创建时间',
        sortable: true,
        align: 'center',
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
        width: 180,
      },
      {
        field: 'addWho',
        title: '创建人',
        align: 'center',
        showOverflow: true,
        width: 120,
      },
      {
        field: 'editTime',
        title: '修改时间',
        sortable: true,
        align: 'center',
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
        width: 180,
      },
      {
        field: 'editWho',
        title: '修改人',
        align: 'center',
        showOverflow: true,
        width: 120,
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
          code: 'edit',
          name: '编辑',
          prefixIcon: 'vxe-icon-edit',
        },
        {
          code: 'manageNodes',
          name: '节点管理',
          prefixIcon: 'vxe-icon-setting',
        },
        {
          code: 'delete',
          name: '删除',
          prefixIcon: 'vxe-icon-delete',
        },
      ],
    },
    height: '100%',
    // 设置行唯一键字段为主键字段
    rowId: 'serviceDefinitionId',
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
   * 设置服务定义列表
   */
  const setServiceList = (list: ServiceDefinition[]) => {
    serviceList.value = list
  }

  /**
   * 在列表中添加服务定义
   */
  const addServiceToList = (service: ServiceDefinition) => {
    serviceList.value.push(service)
  }

  /**
   * 更新列表中的服务定义
   */
  const updateServiceInList = (service: ServiceDefinition) => {
    const index = serviceList.value.findIndex(
      (item) => item.serviceDefinitionId === service.serviceDefinitionId
    )
    if (index >= 0) {
      // 使用 Object.assign 更新对象属性，保持 Vue 响应式
      Object.assign(serviceList.value[index], service)
    }
  }

  /**
   * 从列表中移除服务定义
   */
  const removeServiceFromList = (serviceDefinitionId: string) => {
    const index = serviceList.value.findIndex(
      (item) => item.serviceDefinitionId === serviceDefinitionId
    )
    if (index >= 0) {
      serviceList.value.splice(index, 1)
    }
  }

  /**
   * 从列表中批量移除服务定义
   */
  const removeServicesFromList = (serviceDefinitionIds: string[]) => {
    serviceList.value = serviceList.value.filter(
      (item) => !serviceDefinitionIds.includes(item.serviceDefinitionId)
    )
  }

  return {
    // 数据状态
    moduleId,
    loading,
    serviceList,
    pageInfo,

    // 配置
    searchFormConfig,
    serviceFormConfig,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setServiceList,
    addServiceToList,
    updateServiceInList,
    removeServiceFromList,
    removeServicesFromList,
  }
}

/**
 * 服务定义列表管理 Model 类型
 */
export type ServiceDefinitionModel = ReturnType<typeof useServiceDefinitionModel>

