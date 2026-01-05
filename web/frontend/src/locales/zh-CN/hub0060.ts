/**
 * hub0060 隧道服务器管理模块中文语言包
 */
export default {
  // 页面标题和描述
  pageTitle: '隧道服务器管理',
  pageDescription: '管理FRP隧道服务器配置，包括控制端口、认证配置、网络设置等',

  // 统计信息
  stats: {
    totalServers: '总服务器数',
    runningServers: '运行中',
    stoppedServers: '已停止',
    errorServers: '错误状态',
    totalClients: '总客户端数',
    totalConnections: '总连接数',
    clientUsage: '使用率',
    connectionLoad: '负载',
    healthyServers: '健康服务器',
    stoppedServersStat: '停止服务器',
    abnormalServers: '异常服务器',
    healthOverview: '健康状态概览'
  },

  // 表格列标题
  table: {
    serverName: '服务器名称',
    controlAddress: '控制地址',
    status: '状态',
    maxClients: '客户端限制',
    auth: '认证',
    tls: 'TLS',
    heartbeat: '心跳配置',
    startTime: '启动时间',
    createTime: '创建时间',
    actions: '操作',
    running: '运行中',
    stopped: '已停止',
    error: '错误',
    enabled: '启用',
    disabled: '禁用'
  },

  // 搜索和过滤
  search: {
    serverName: '搜索服务器名称',
    serverStatus: '服务器状态',
    controlAddress: '控制地址',
    controlPort: '控制端口',
    search: '搜索',
    reset: '重置'
  },

  // 按钮操作
  actions: {
    refresh: '刷新',
    create: '新增服务器',
    batchDelete: '批量删除',
    start: '启动',
    stop: '停止',
    restart: '重启',
    test: '测试',
    edit: '编辑',
    delete: '删除',
    startServer: '启动服务器',
    stopServer: '停止服务器',
    restartServer: '重启服务器',
    testConnection: '测试连接',
    editServer: '编辑服务器',
    deleteServer: '删除服务器'
  },

  // 对话框
  dialog: {
    create: '创建隧道服务器',
    edit: '编辑隧道服务器',
    basicConfig: '基础配置',
    networkConfig: '网络配置',
    authConfig: '认证配置',
    advancedConfig: '高级配置',
    
    // 表单字段
    form: {
      serverName: '服务器名称',
      serverNamePlaceholder: '请输入服务器名称',
      serverDescription: '服务器描述',
      serverDescriptionPlaceholder: '请输入服务器描述',
      controlAddress: '控制地址',
      controlAddressPlaceholder: '请输入控制地址，如: 0.0.0.0',
      controlPort: '控制端口',
      controlPortPlaceholder: '请输入控制端口',
      dashboardPort: '管理面板端口',
      dashboardPortPlaceholder: '请输入管理面板端口',
      httpPort: 'HTTP端口',
      httpPortPlaceholder: '虚拟主机HTTP端口',
      httpsPort: 'HTTPS端口',
      httpsPortPlaceholder: '虚拟主机HTTPS端口',
      maxClients: '最大客户端数',
      maxClientsPlaceholder: '最大客户端连接数',
      maxPortsPerClient: '每客户端最大端口',
      maxPortsPerClientPlaceholder: '每个客户端最大端口数',
      allowPorts: '允许的端口范围',
      allowPortsPlaceholder: '如: 10000-20000,30000-40000',
      enableTokenAuth: '启用Token认证',
      authToken: '认证Token',
      authTokenPlaceholder: '请输入或生成认证Token',
      generateToken: '生成Token',
      enableTls: '启用TLS',
      tlsCertFile: 'TLS证书文件',
      tlsCertFilePlaceholder: '请输入TLS证书文件路径',
      tlsKeyFile: 'TLS私钥文件',
      tlsKeyFilePlaceholder: '请输入TLS私钥文件路径',
      heartbeatInterval: '心跳间隔(秒)',
      heartbeatIntervalPlaceholder: '心跳间隔时间',
      heartbeatTimeout: '心跳超时(秒)',
      heartbeatTimeoutPlaceholder: '心跳超时时间',
      logLevel: '日志级别',
      logLevelPlaceholder: '请选择日志级别',
      noteText: '备注信息',
      noteTextPlaceholder: '请输入备注信息'
    },

    // 按钮
    cancel: '取消',
    createBtn: '创建',
    update: '更新'
  },

  // 确认对话框
  confirm: {
    batchDelete: '确认批量删除',
    batchDeleteContent: '确认删除选中的 {count} 个隧道服务器吗？',
    delete: '确认删除这个隧道服务器吗？',
    confirmDelete: '确认删除',
    cancel: '取消'
  },

  // 消息提示
  messages: {
    createSuccess: '隧道服务器创建成功',
    updateSuccess: '隧道服务器更新成功',
    deleteSuccess: '隧道服务器删除成功',
    batchDeleteSuccess: '成功删除 {count} 个隧道服务器',
    startSuccess: '隧道服务器启动成功',
    stopSuccess: '隧道服务器停止成功',
    restartSuccess: '隧道服务器重启成功',
    testSuccess: '连接测试成功',
    testSuccessWithLatency: '连接测试成功 (延迟: {latency}ms)',
    testFailed: '连接测试失败：无法连接到服务器',
    generateTokenSuccess: '认证令牌生成成功',
    refreshSuccess: '数据刷新成功',
    createFailed: '创建隧道服务器失败',
    updateFailed: '更新隧道服务器失败',
    deleteFailed: '删除隧道服务器失败',
    batchDeleteFailed: '批量删除失败',
    startFailed: '启动失败',
    stopFailed: '停止失败',
    restartFailed: '重启失败',
    generateTokenFailed: '生成认证令牌失败',
    getListFailed: '获取隧道服务器列表失败',
    getStatsFailed: '获取统计信息失败'
  },

  // 表单验证
  validation: {
    serverNameRequired: '请输入服务器名称',
    serverNameLength: '服务器名称长度应在2-100字符之间',
    controlAddressRequired: '请输入控制地址',
    controlPortRequired: '请输入控制端口',
    controlPortRange: '端口范围应在1-65535之间',
    maxClientsRequired: '请输入最大客户端数',
    maxClientsRange: '最大客户端数应在1-10000之间',
    heartbeatIntervalRequired: '请输入心跳间隔',
    heartbeatIntervalRange: '心跳间隔应在10-300秒之间',
    heartbeatTimeoutRequired: '请输入心跳超时',
    heartbeatTimeoutRange: '心跳超时应在30-600秒之间',
    authTokenRequired: '启用Token认证时必须提供认证Token',
    tlsCertFileRequired: '启用TLS时必须提供证书文件路径',
    tlsKeyFileRequired: '启用TLS时必须提供私钥文件路径'
  },

  // 状态选项
  options: {
    logLevel: {
      debug: 'Debug',
      info: 'Info',
      warn: 'Warning',
      error: 'Error'
    },
    status: {
      running: '运行中',
      stopped: '已停止',
      error: '错误'
    },
    auth: {
      enabled: '启用',
      disabled: '禁用'
    }
  },

  // 分页
  pagination: {
    total: '共 {total} 条记录'
  }
}
