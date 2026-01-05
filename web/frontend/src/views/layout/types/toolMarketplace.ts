// 工具市场相关类型定义

// 工具状态枚举
export enum ToolStatus {
  AVAILABLE = 'available', // 可用
  INSTALLED = 'installed', // 已安装
  UPDATING = 'updating', // 更新中
  INSTALLING = 'installing', // 安装中
  ERROR = 'error', // 错误状态
}

// 工具分类枚举
export enum ToolCategory {
  ALL = 'all', // 全部
  UTILITY = 'utility', // 实用工具
  CHART = 'chart', // 图表工具
  FORM = 'form', // 表单工具
  TABLE = 'table', // 表格工具
  LAYOUT = 'layout', // 布局工具
  DATA = 'data', // 数据处理
  UI = 'ui', // UI组件
  BUSINESS = 'business', // 业务组件
}

// 工具权限级别
export enum ToolPermissionLevel {
  PUBLIC = 'public', // 公开
  TENANT = 'tenant', // 租户级别
  USER = 'user', // 用户级别
  ADMIN = 'admin', // 管理员级别
}

// 工具配置类型
export enum ToolConfigType {
  BOOLEAN = 'boolean',
  STRING = 'string',
  NUMBER = 'number',
  SELECT = 'select',
  MULTI_SELECT = 'multi_select',
  COLOR = 'color',
  JSON = 'json',
}

// 工具配置项接口
export interface ToolConfigItem {
  key: string // 配置键
  label: string // 显示名称
  type: ToolConfigType // 配置类型
  defaultValue?: any // 默认值
  required?: boolean // 是否必填
  description?: string // 描述
  options?: Array<{ label: string; value: any }> // 选项（用于select类型）
  validation?: {
    min?: number
    max?: number
    pattern?: string
    message?: string
  }
}

// 工具接口
export interface Tool {
  id: string // 工具唯一标识
  name: string // 工具名称
  displayName: string // 显示名称
  description: string // 工具描述
  version: string // 版本号
  author: string // 作者
  category: ToolCategory // 分类
  tags: string[] // 标签
  icon?: string // 图标
  screenshot?: string[] // 截图
  status: ToolStatus // 状态
  permissionLevel: ToolPermissionLevel // 权限级别

  // 安装信息
  installTime?: string // 安装时间
  lastUpdateTime?: string // 最后更新时间
  usageCount?: number // 使用次数

  // 技术信息
  dependencies?: string[] // 依赖
  size?: number // 大小（KB）
  documentation?: string // 文档链接
  repository?: string // 仓库地址

  // 配置信息
  configSchema?: ToolConfigItem[] // 配置模式
  defaultConfig?: Record<string, any> // 默认配置
  userConfig?: Record<string, any> // 用户配置

  // 组件信息
  componentPath?: string // 组件路径
  componentName?: string // 组件名称
  props?: Record<string, any> // 组件属性

  // 元数据
  metadata?: Record<string, any> // 扩展元数据
  createTime: string // 创建时间
  updateTime: string // 更新时间
  createBy: string // 创建人
  updateBy: string // 更新人
  activeFlag: 'Y' | 'N' // 激活状态
}

// 工具分类接口
export interface ToolCategoryItem {
  key: ToolCategory // 分类键
  label: string // 显示标签
  icon?: string // 图标
  count?: number // 工具数量
  description?: string // 描述
}

// 工具搜索条件
export interface ToolSearchParams {
  keyword?: string // 关键词
  category?: ToolCategory // 分类
  status?: ToolStatus // 状态
  author?: string // 作者
  tags?: string[] // 标签
  permissionLevel?: ToolPermissionLevel // 权限级别
  installed?: boolean // 是否已安装
}

// 工具安装参数
export interface ToolInstallParams {
  toolId: string // 工具ID
  config?: Record<string, any> // 配置参数
  autoStart?: boolean // 是否自动启动
}

// 工具卸载参数
export interface ToolUninstallParams {
  toolId: string // 工具ID
  removeConfig?: boolean // 是否移除配置
  removeData?: boolean // 是否移除数据
}

// 工具配置参数
export interface ToolConfigParams {
  toolId: string // 工具ID
  config: Record<string, any> // 配置数据
}

// 工具操作结果
export interface ToolOperationResult {
  success: boolean // 是否成功
  message?: string // 消息
  data?: any // 返回数据
  error?: string // 错误信息
}

// 工具使用统计
export interface ToolUsageStats {
  toolId: string // 工具ID
  usageCount: number // 使用次数
  lastUsedTime: string // 最后使用时间
  avgUsageTime?: number // 平均使用时长
  errorCount?: number // 错误次数
  rating?: number // 评分
}

// 工具市场配置
export interface ToolMarketplaceConfig {
  enableAutoUpdate: boolean // 是否启用自动更新
  checkUpdateInterval: number // 检查更新间隔（小时）
  maxConcurrentInstalls: number // 最大并发安装数
  enableUsageStats: boolean // 是否启用使用统计
  defaultCategory: ToolCategory // 默认分类
  allowedCategories: ToolCategory[] // 允许的分类
  maxToolsPerPage: number // 每页最大工具数
}

// 工具市场状态
export interface ToolMarketplaceState {
  tools: Tool[] // 工具列表
  categories: ToolCategoryItem[] // 分类列表
  searchParams: ToolSearchParams // 搜索参数
  selectedTool: Tool | null // 选中的工具
  loading: boolean // 加载状态
  error: string | null // 错误信息
  config: ToolMarketplaceConfig // 配置
}

// 视图模式
export type ViewMode = 'grid' | 'list'

// 工具市场事件
export interface ToolMarketplaceEvents {
  onToolInstall: (tool: Tool) => void
  onToolUninstall: (tool: Tool) => void
  onToolConfigure: (tool: Tool) => void
  onToolPreview: (tool: Tool) => void
  onToolUpdate: (tool: Tool) => void
  onCategoryChange: (category: ToolCategory) => void
  onSearchChange: (params: ToolSearchParams) => void
}

// API响应类型
export interface ToolApiResponse<T = any> {
  oK: boolean
  errCode?: string
  errMsg?: string
  popMsg?: string
  bizData?: T
}

// 工具列表查询参数
export interface ToolQueryParams {
  tenantId?: string
  category?: ToolCategory
  status?: ToolStatus
  keyword?: string
  installed?: boolean
  pageNo?: number
  pageSize?: number
}

// 分页结果
export interface PageResult<T> {
  list: T[]
  total: number
  pageNo: number
  pageSize: number
  totalPages: number
}

// 工具详情
export interface ToolDetail extends Tool {
  fullDescription?: string // 完整描述
  changelog?: string // 更新日志
  reviews?: ToolReview[] // 评价列表
  similarTools?: Tool[] // 相似工具
}

// 工具评价
export interface ToolReview {
  id: string
  toolId: string
  userId: string
  userName: string
  rating: number // 1-5星
  comment: string
  createTime: string
  helpful?: number // 有用数
}

// 工具快捷操作
export interface ToolQuickAction {
  key: string
  label: string
  icon?: string
  action: (tool: Tool) => void
  visible?: (tool: Tool) => boolean
  disabled?: (tool: Tool) => boolean
}
