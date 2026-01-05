export default {
  title: '动态服务管理',
  subtitle: '管理隧道服务配置和状态',
  stats: {
    total: '总服务数',
    active: '活跃服务',
    inactive: '未激活服务',
    error: '错误服务',
    connections: '总连接数',
    traffic: '总流量'
  },
  table: {
    serviceName: '服务名称',
    serviceType: '服务类型',
    serviceStatus: '服务状态',
    localAddress: '本地地址',
    localPort: '本地端口',
    remotePort: '远程端口',
    connectionCount: '连接数',
    registeredTime: '注册时间'
  },
  filter: {
    status: '服务状态',
    type: '服务类型'
  },
  status: {
    active: '活跃',
    inactive: '未激活',
    error: '错误',
    offline: '离线'
  },
  dialog: {
    createTitle: '创建服务',
    editTitle: '编辑服务',
    tabs: {
      basic: '基础配置',
      advanced: '高级配置',
      note: '备注'
    }
  },
  form: {
    serviceName: '服务名称',
    serviceNameRequired: '请输入服务名称',
    serviceNamePlaceholder: '请输入服务名称',
    serviceDescription: '服务描述',
    serviceDescriptionPlaceholder: '请输入服务描述（可选）',
    tunnelClientId: '隧道客户端',
    tunnelClientIdRequired: '请选择隧道客户端',
    tunnelClientIdPlaceholder: '请选择要使用的隧道客户端',
    serviceType: '服务类型',
    serviceTypeRequired: '请选择服务类型',
    serviceTypePlaceholder: '请选择服务类型',
    localAddress: '本地地址',
    localAddressRequired: '请输入本地地址',
    localAddressPlaceholder: '请输入本地地址（如: 127.0.0.1）',
    localPort: '本地端口',
    localPortRequired: '请输入本地端口',
    localPortPlaceholder: '请输入本地端口（1-65535）',
    remotePort: '远程端口',
    remotePortPlaceholder: '请输入远程端口（可选，留空自动分配）',
    subDomain: '子域名',
    subDomainPlaceholder: '请输入子域名（用于HTTP/HTTPS服务）',
    useEncryption: '启用加密',
    useCompression: '启用压缩',
    maxConnections: '最大连接数',
    maxConnectionsRequired: '请输入最大连接数',
    maxConnectionsPlaceholder: '请输入最大连接数',
    bandwidthLimit: '带宽限制',
    bandwidthLimitPlaceholder: '请输入带宽限制（可选）',
    httpAuth: 'HTTP 基础认证',
    httpUser: 'HTTP 用户名',
    httpUserPlaceholder: '请输入 HTTP 认证用户名（可选）',
    httpPassword: 'HTTP 密码',
    httpPasswordPlaceholder: '请输入 HTTP 认证密码（可选）',
    activeFlag: '活动标记',
    noteText: '备注',
    noteTextPlaceholder: '请输入备注信息（可选）'
  },
  actions: {
    register: '注册',
    unregister: '注销'
  },
  message: {
    enableSuccess: '服务启用成功',
    enableFailed: '服务启用失败',
    disableSuccess: '服务禁用成功',
    disableFailed: '服务禁用失败',
    registerSuccess: '服务注册成功',
    registerFailed: '服务注册失败',
    unregisterSuccess: '服务注销成功',
    unregisterFailed: '服务注销失败'
  }
}

