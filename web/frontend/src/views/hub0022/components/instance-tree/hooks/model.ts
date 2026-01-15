/**
 * 网关实例树组件 Model
 * 统一管理数据状态和计算属性
 */

import type { DataFormField, DataFormTab } from '@/components/form/data/types'
import type { ContextMenuConfig } from '@/components/gmenu'
import { AddOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { computed, h, ref } from 'vue'
import type { GatewayInstance, InstanceTreeOption } from '../types'
import { ProxyType as ProxyTypeEnum } from '../types'
/**
 * 网关实例树 Model
 */
export function useGatewayInstanceTreeModel() {
  // ============= 数据状态 =============

  /** 模块ID */
  const moduleId = 'hub0022'

  /** 加载状态 */
  const loading = ref(false)

  /** 网关实例列表数据 */
  const instanceList = ref<GatewayInstance[]>([])

  /** 分页状态 */
  const currentPage = ref(1)
  const pageSize = ref(20) // 默认每页20条
  
  /** 数据总数（后端分页） */
  const totalCount = ref(0)

  /** 过滤关键词 */
  const filterKeyword = ref('')

  // ============= 计算属性 =============

  /**
   * 将实例列表转换为树形结构（后端分页，直接使用 instanceList）
   */
  const treeData = computed<InstanceTreeOption[]>(() => {
    return instanceList.value.map(instance => ({
      key: instance.gatewayInstanceId,
      label: getInstanceLabel(instance),
      instance: instance,
    }))
  })

  // ============= 辅助方法 =============

  /**
   * 获取实例标签
   */
  function getInstanceLabel(instance: GatewayInstance): string {
    const port = instance.tlsEnabled === 'Y' ? instance.httpsPort : instance.httpPort
    return `${instance.instanceName || '未命名'} (${instance.bindAddress || '-'}:${port || '-'})`
  }

  // ============= 状态更新方法 =============

  /**
   * 设置实例列表
   */
  function setInstanceList(list: GatewayInstance[]) {
    instanceList.value = list
  }

  /**
   * 设置加载状态
   */
  function setLoading(value: boolean) {
    loading.value = value
  }

  /**
   * 设置当前页
   */
  function setCurrentPage(page: number) {
    currentPage.value = page
  }

  /**
   * 设置每页大小
   */
  function setPageSize(size: number) {
    pageSize.value = size
  }

  /**
   * 设置过滤关键词
   */
  function setFilterKeyword(keyword: string) {
    filterKeyword.value = keyword
  }

  /**
   * 设置数据总数
   */
  function setTotalCount(count: number) {
    totalCount.value = count
  }

  /**
   * 重置分页到第一页
   */
  function resetPage() {
    currentPage.value = 1
  }

  /**
   * 清空搜索关键词
   */
  function clearFilter() {
    filterKeyword.value = ''
  }

  // ============= 右键菜单配置 =============

  /**
   * 树节点右键菜单配置
   */
  const treeMenuConfig: ContextMenuConfig = {
    enabled: true,
    showCopyNode: true,
    customMenus: [
      {
        code: 'addProxy',
        name: '代理配置',
        prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(AddOutline) }),
      },
    ],
  }

  // ============= 代理配置表单配置 =============

  /**
   * 代理配置表单字段配置
   * 用于 GdataFormModal 组件
   */
  const proxyFormConfig = {
    tabs: [
      { 
        key: 'basic', 
        label: '基本信息',
        // 基本信息标签页始终显示
      },
      { 
        key: 'http', 
        label: 'HTTP配置',
        // 只在代理类型为 HTTP 时显示
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
      },
      { 
        key: 'websocket', 
        label: 'WebSocket配置',
        // 只在代理类型为 WebSocket 时显示
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.WEBSOCKET,
      },
      { 
        key: 'tcp', 
        label: 'TCP配置',
        // 只在代理类型为 TCP 时显示
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.TCP,
      },
      { 
        key: 'udp', 
        label: 'UDP配置',
        // 只在代理类型为 UDP 时显示
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.UDP,
      },
      { 
        key: 'custom', 
        label: '其它',
        // 其它标签页始终显示
      },
    ] as DataFormTab[],
    fields: [
      // ============= 主键字段（隐藏，但必须存在用于编辑） =============
      {
        field: 'proxyConfigId',
        label: '代理配置ID',
        type: 'input' as const,
        span: 12,
        tabKey: 'basic',
        primary: true,
        show: false,
      },
      {
        field: 'gatewayInstanceId',
        label: '网关实例ID',
        type: 'input' as const,
        span: 12,
        tabKey: 'basic',
        show: false,
      },
      {
        field: 'tenantId',
        label: '租户ID',
        type: 'input' as const,
        span: 12,
        tabKey: 'basic',
        show: false,
      },
      // ============= 基本信息 Tab =============
      {
        field: 'proxyName',
        label: '代理名称',
        type: 'input' as const,
        placeholder: '请输入代理名称',
        span: 12,
        tabKey: 'basic',
        required: true,
        tips: '代理配置的唯一标识名称，用于区分不同的代理配置',
        rules: [
          { required: true, message: '请输入代理名称', trigger: ['blur', 'input'] },
          { max: 50, message: '代理名称不能超过50个字符', trigger: ['blur', 'input'] },
        ],
      },
      {
        field: 'proxyType',
        label: '代理类型',
        type: 'select' as const,
        placeholder: '请选择代理类型',
        span: 12,
        tabKey: 'basic',
        required: true,
        defaultValue: ProxyTypeEnum.HTTP,
        tips: '选择代理协议类型：HTTP(支持HTTP/HTTPS协议)、WebSocket(支持WebSocket协议)、TCP(支持TCP协议)、UDP(支持UDP协议)',
        options: [
          { label: 'HTTP', value: ProxyTypeEnum.HTTP },
          { label: 'WebSocket', value: ProxyTypeEnum.WEBSOCKET },
          { label: 'TCP', value: ProxyTypeEnum.TCP },
          { label: 'UDP', value: ProxyTypeEnum.UDP },
        ],
        rules: [
          { required: true, message: '请选择代理类型', trigger: ['blur', 'change'] },
        ],
      },
      {
        field: 'configPriority',
        label: '配置优先级',
        type: 'number' as const,
        placeholder: '请输入配置优先级',
        span: 12,
        tabKey: 'basic',
        required: true,
        defaultValue: 100,
        tips: '配置优先级，数值越小优先级越高。当存在多个代理配置时，优先级高的配置会优先匹配',
        props: {
          min: 0,
          max: 999,
        },
        rules: [
          {
            required: true,
            message: '请输入优先级',
            trigger: ['blur', 'change'],
            validator: (_rule: any, value: any) => {
              if (value === null || value === undefined || value === '') {
                return new Error('请输入优先级')
              }
              const num = Number(value)
              if (isNaN(num)) {
                return new Error('优先级必须是数字')
              }
              if (num < 0 || num > 999) {
                return new Error('优先级必须是0-999之间的数字')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'activeFlag',
        label: '状态',
        type: 'switch' as const,
        span: 12,
        tabKey: 'basic',
        defaultValue: 'Y',
        tips: '代理配置的启用状态。启用后配置才会生效，禁用后代理将不会处理请求',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'noteText',
        label: '备注',
        type: 'textarea' as const,
        placeholder: '请输入备注信息',
        span: 24,
        tabKey: 'basic',
        tips: '代理配置的备注说明，用于记录配置的用途、注意事项等信息',
        props: {
          rows: 3,
        },
      },
      // ============= HTTP配置 Tab =============
      // 1. 协议版本
      {
        field: 'proxyConfig.httpVersion',
        label: 'HTTP版本',
        type: 'select' as const,
        placeholder: '请选择HTTP版本',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: '1.1',
        tips: 'HTTP协议版本，HTTP/1.0 或 HTTP/1.1。HTTP/1.1 支持连接复用、管道化等特性，性能更好',
        options: [
          { label: 'HTTP/1.0', value: '1.0' },
          { label: 'HTTP/1.1', value: '1.1' },
        ],
      },
      // 2. 连接相关配置
      {
        field: 'proxyConfig.connectTimeout',
        label: '连接超时（秒）',
        type: 'number' as const,
        placeholder: '请输入连接超时',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: 30,
        tips: '建立TCP连接的超时时间，超过此时间未建立连接则请求失败',
        props: {
          min: 1,
        },
      },
      {
        field: 'proxyConfig.keepAlive',
        label: '保持连接',
        type: 'switch' as const,
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: true,
        tips: '是否启用HTTP Keep-Alive连接复用。启用后可以复用TCP连接，提高性能，减少连接建立开销',
      },
      {
        field: 'proxyConfig.maxIdleConns',
        label: '最大空闲连接数',
        type: 'number' as const,
        placeholder: '请输入最大空闲连接数',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: 100,
        tips: '连接池中保持的最大空闲连接数。空闲连接可以被复用，提高性能，但会占用内存资源',
        props: {
          min: 0,
        },
      },
      {
        field: 'proxyConfig.idleConnTimeout',
        label: '空闲连接超时（秒）',
        type: 'number' as const,
        placeholder: '请输入空闲连接超时',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        required: true,
        defaultValue: 90,
        tips: '连接池中空闲连接的最大保持时间，超过此时间的空闲连接会被关闭释放资源',
        props: {
          min: 1,
          max: 7200,
        },
        rules: [
          {
            required: true,
            message: '请输入空闲连接超时时间',
            trigger: ['blur', 'change'],
            type: 'number',
            validator: (_rule: any, value: any) => {
              if (value === null || value === undefined) return new Error('请输入空闲连接超时时间')
              if (value < 1) return new Error('空闲连接超时时间必须大于0秒')
              if (value > 7200) return new Error('空闲连接超时时间不能超过7200秒')
              return true
            },
          },
        ],
      },
      // 3. 超时相关配置
      {
        field: 'proxyConfig.timeout',
        label: '总超时时间（秒）',
        type: 'number' as const,
        placeholder: '请输入超时时间',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        required: true,
        defaultValue: 60,
        tips: '请求的总超时时间（包括连接、发送、读取），超过此时间未完成则请求失败',
        props: {
          min: 1,
          max: 3600,
        },
        rules: [
          {
            required: true,
            message: '请输入超时时间',
            trigger: ['blur', 'change'],
            type: 'number',
            validator: (_rule: any, value: any) => {
              if (value === null || value === undefined) return new Error('请输入超时时间')
              if (value < 1) return new Error('超时时间必须大于0秒')
              if (value > 3600) return new Error('超时时间不能超过3600秒')
              return true
            },
          },
        ],
      },
      {
        field: 'proxyConfig.sendTimeout',
        label: '发送超时（秒）',
        type: 'number' as const,
        placeholder: '请输入发送超时',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: 60,
        tips: '发送请求数据的超时时间，超过此时间未发送完成则请求失败',
        props: {
          min: 1,
        },
      },
      {
        field: 'proxyConfig.readTimeout',
        label: '读取超时（秒）',
        type: 'number' as const,
        placeholder: '请输入读取超时',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: 60,
        tips: '读取响应数据的超时时间，超过此时间未读取完成则请求失败',
        props: {
          min: 1,
        },
      },
      // 4. 重试相关配置
      {
        field: 'proxyConfig.retryCount',
        label: '重试次数',
        type: 'number' as const,
        placeholder: '请输入重试次数',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: 0,
        tips: '请求失败时的自动重试次数。设置为0表示不重试',
        props: {
          min: 0,
        },
      },
      {
        field: 'proxyConfig.retryTimeout',
        label: '重试超时（秒）',
        type: 'number' as const,
        placeholder: '请输入重试超时',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        required: true,
        defaultValue: 30,
        tips: '每次重试的超时时间，如果单次重试超过此时间则重试失败，继续下一次重试',
        props: {
          min: 1,
          max: 300,
        },
        rules: [
          {
            required: true,
            message: '请输入重试超时时间',
            trigger: ['blur', 'change'],
            type: 'number',
            validator: (_rule: any, value: any) => {
              if (value === null || value === undefined) return new Error('请输入重试超时时间')
              if (value < 1) return new Error('重试超时时间必须大于0秒')
              if (value > 300) return new Error('重试超时时间不能超过300秒')
              return true
            },
          },
        ],
      },
      // 5. 缓冲相关配置
      {
        field: 'proxyConfig.proxyBuffering',
        label: '代理缓冲',
        type: 'switch' as const,
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: true,
        tips: '是否启用代理缓冲。启用后会在代理层缓冲请求和响应，可以优化传输性能',
      },
      {
        field: 'proxyConfig.bufferSize',
        label: '缓冲区大小（字节）',
        type: 'number' as const,
        placeholder: '请输入缓冲区大小',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: 4096,
        tips: '默认缓冲区大小，用于临时存储请求和响应数据',
        props: {
          min: 0,
        },
      },
      {
        field: 'proxyConfig.maxBufferSize',
        label: '最大缓冲区大小（字节）',
        type: 'number' as const,
        placeholder: '请输入最大缓冲区大小',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: 65536,
        tips: '缓冲区的最大大小限制，防止缓冲区无限增长导致内存溢出',
        props: {
          min: 0,
        },
      },
      {
        field: 'proxyConfig.copyResponseBody',
        label: '复制响应体',
        type: 'switch' as const,
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: false,
        tips: '是否复制完整的响应体到内存。启用后可以多次读取响应，但会占用更多内存',
      },
      // 6. 请求头相关配置
      {
        field: 'proxyConfig.followRedirects',
        label: '跟随重定向',
        type: 'switch' as const,
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: false,
        tips: '是否自动跟随HTTP重定向响应（3xx状态码）。启用后会自动跳转到重定向地址',
      },
      {
        field: 'proxyConfig.preserveHost',
        label: '保留原始Host',
        type: 'switch' as const,
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: false,
        tips: '是否保留客户端原始Host请求头。启用后目标服务器会看到原始Host，而不是代理服务器的地址',
      },
      {
        field: 'proxyConfig.addXForwardedFor',
        label: '添加X-Forwarded-For',
        type: 'switch' as const,
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: true,
        tips: '是否添加X-Forwarded-For请求头，用于标识客户端的真实IP地址，方便目标服务器获取客户端信息',
      },
      {
        field: 'proxyConfig.addXRealIP',
        label: '添加X-Real-IP',
        type: 'switch' as const,
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: true,
        tips: '是否添加X-Real-IP请求头，用于标识客户端的真实IP地址，是X-Forwarded-For的简化版本',
      },
      {
        field: 'proxyConfig.addXForwardedProto',
        label: '添加X-Forwarded-Proto',
        type: 'switch' as const,
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: true,
        tips: '是否添加X-Forwarded-Proto请求头，用于标识客户端使用的协议（http或https）',
      },
      // 7. TLS/SSL相关配置
      {
        field: 'proxyConfig.tlsInsecureSkipVerify',
        label: '跳过TLS证书验证',
        type: 'switch' as const,
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: false,
        tips: '是否跳过TLS证书验证（仅用于测试环境）。生产环境建议关闭以确保安全性',
      },
      {
        field: 'proxyConfig.tlsMinVersion',
        label: '最小TLS版本',
        type: 'select' as const,
        placeholder: '请选择最小TLS版本',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: '1.2',
        tips: '支持的最小TLS协议版本。低于此版本的连接将被拒绝，建议使用TLS 1.2或更高版本',
        options: [
          { label: 'TLS 1.0', value: '1.0' },
          { label: 'TLS 1.1', value: '1.1' },
          { label: 'TLS 1.2', value: '1.2' },
          { label: 'TLS 1.3', value: '1.3' },
        ],
      },
      {
        field: 'proxyConfig.tlsMaxVersion',
        label: '最大TLS版本',
        type: 'select' as const,
        placeholder: '请选择最大TLS版本',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        defaultValue: '1.3',
        tips: '支持的最大TLS协议版本。高于此版本的连接将降级到此版本',
        options: [
          { label: 'TLS 1.0', value: '1.0' },
          { label: 'TLS 1.1', value: '1.1' },
          { label: 'TLS 1.2', value: '1.2' },
          { label: 'TLS 1.3', value: '1.3' },
        ],
      },
      {
        field: 'proxyConfig.tlsServerName',
        label: 'TLS服务器名称',
        type: 'input' as const,
        placeholder: '请输入TLS服务器名称（可选）',
        span: 12,
        tabKey: 'http',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.HTTP,
        tips: 'TLS SNI（Server Name Indication）服务器名称，用于指定要连接的服务器主机名，通常与证书域名匹配',
      },
      // 注意：setHeaders、passHeaders、hideHeaders 等复杂字段需要使用自定义组件或特殊处理
      // 这里暂时不包含，可以在后续扩展
      // ============= WebSocket配置 Tab =============
      {
        field: 'proxyConfig.pingInterval',
        label: 'Ping间隔（秒）',
        type: 'number' as const,
        placeholder: '请输入Ping间隔',
        span: 12,
        tabKey: 'websocket',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.WEBSOCKET,
        defaultValue: 30,
        tips: 'WebSocket连接心跳检测的Ping消息发送间隔，用于保持连接活跃并检测连接状态',
        props: {
          min: 1,
        },
      },
      {
        field: 'proxyConfig.pongTimeout',
        label: 'Pong超时（秒）',
        type: 'number' as const,
        placeholder: '请输入Pong超时',
        span: 12,
        tabKey: 'websocket',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.WEBSOCKET,
        defaultValue: 10,
        tips: '发送Ping消息后等待Pong响应的超时时间，超过此时间未收到Pong则视为连接断开',
        props: {
          min: 1,
        },
      },
      {
        field: 'proxyConfig.maxMessageSize',
        label: '最大消息大小（字节）',
        type: 'number' as const,
        placeholder: '请输入最大消息大小',
        span: 12,
        tabKey: 'websocket',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.WEBSOCKET,
        defaultValue: 32768,
        tips: 'WebSocket单条消息的最大大小限制，超过此大小的消息将被拒绝，防止内存溢出',
        props: {
          min: 0,
        },
      },
      {
        field: 'proxyConfig.enableCompression',
        label: '启用压缩',
        type: 'switch' as const,
        span: 12,
        tabKey: 'websocket',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.WEBSOCKET,
        defaultValue: false,
        tips: '是否启用WebSocket消息压缩（permessage-deflate）。启用后可以减少网络传输量，但会增加CPU开销',
      },
      // ============= TCP配置 Tab =============
      {
        field: 'proxyConfig.connectTimeout',
        label: '连接超时（秒）',
        type: 'number' as const,
        placeholder: '请输入连接超时',
        span: 12,
        tabKey: 'tcp',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.TCP,
        defaultValue: 5,
        tips: '建立TCP连接的超时时间，超过此时间未建立连接则连接失败',
        props: {
          min: 1,
        },
      },
      {
        field: 'proxyConfig.keepAlive',
        label: '保持连接',
        type: 'switch' as const,
        span: 12,
        tabKey: 'tcp',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.TCP,
        defaultValue: true,
        tips: '是否启用TCP Keep-Alive机制。启用后会定期发送探测包保持TCP连接活跃，及时发现断开的连接',
      },
      // ============= UDP配置 Tab =============
      {
        field: 'proxyConfig.bufferSize',
        label: '缓冲区大小（字节）',
        type: 'number' as const,
        placeholder: '请输入缓冲区大小',
        span: 12,
        tabKey: 'udp',
        show: (formData: Record<string, any>) => formData.proxyType === ProxyTypeEnum.UDP,
        defaultValue: 4096,
        tips: 'UDP数据包接收缓冲区大小，影响可以接收的最大UDP数据包大小',
        props: {
          min: 1,
        },
      },
      // ============= 其它 Tab =============
      {
        field: 'customConfig',
        label: '自定义配置',
        type: 'textarea' as const,
        placeholder: '{}',
        span: 24,
        tabKey: 'custom',
        tips: '自定义配置项（JSON格式），用于存储额外的配置参数，可根据实际需求扩展使用',
        props: {
          rows: 5,
        },
      },
      {
        field: 'addTime',
        label: '创建时间',
        type: 'datetime' as const,
        span: 12,
        tabKey: 'custom',
        disabled: true,
      },
      {
        field: 'addWho',
        label: '创建人',
        type: 'input' as const,
        span: 12,
        tabKey: 'custom',
        disabled: true,
      },
      {
        field: 'editTime',
        label: '修改时间',
        type: 'datetime' as const,
        span: 12,
        tabKey: 'custom',
        disabled: true,
      },
      {
        field: 'editWho',
        label: '修改人',
        type: 'input' as const,
        span: 12,
        tabKey: 'custom',
        disabled: true,
      },
    ] as DataFormField[],
  }

  return {
    // 状态
    moduleId,
    loading,
    instanceList,
    currentPage,
    pageSize,
    totalCount,
    filterKeyword,

    // 计算属性
    treeData,

    // 配置
    treeMenuConfig,
    proxyFormConfig,

    // 方法
    getInstanceLabel,
    setInstanceList,
    setLoading,
    setCurrentPage,
    setPageSize,
    setTotalCount,
    setFilterKeyword,
    resetPage,
    clearFilter,
  }
}

/**
 * 网关实例树 Model 类型
 */
export type GatewayInstanceTreeModel = ReturnType<typeof useGatewayInstanceTreeModel>
