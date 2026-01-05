import { ref, computed, reactive, watch, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import {
  ToolCategory,
  ToolStatus,
  type Tool,
  type ToolCategoryItem,
  type ToolSearchParams,
  type ViewMode,
  type ToolMarketplaceConfig,
} from '../types/toolMarketplace'

// 简单的本地存储hook

/**
 * 工具市场Hook
 * 管理工具市场的状态和交互逻辑
 */
export function useToolMarketplace() {
  const message = useMessage()

  // 基础状态
  const loading = ref(false)
  const error = ref<string | null>(null)
  const initialized = ref(false)

  // 工具数据
  const tools = ref<Tool[]>([])
  const selectedTool = ref<Tool | null>(null)

  // 界面状态
  const searchQuery = ref('')
  const activeCategory = ref<ToolCategory>(ToolCategory.ALL)
  const previewDrawerVisible = ref(false)
  const configDialogVisible = ref(false)

  // 视图模式
  const viewMode = ref<ViewMode>('grid')

  // 搜索参数
  const searchParams = reactive<ToolSearchParams>({
    keyword: '',
    category: ToolCategory.ALL,
  })

  // 工具分类配置
  const categories = ref<ToolCategoryItem[]>([
    { key: ToolCategory.ALL, label: '全部', icon: 'apps-outline' },
    { key: ToolCategory.UTILITY, label: '实用工具', icon: 'build-outline' },
    { key: ToolCategory.CHART, label: '图表工具', icon: 'bar-chart-outline' },
    { key: ToolCategory.FORM, label: '表单工具', icon: 'clipboard-outline' },
    { key: ToolCategory.TABLE, label: '表格工具', icon: 'grid-outline' },
    { key: ToolCategory.LAYOUT, label: '布局工具', icon: 'resize-outline' },
    { key: ToolCategory.DATA, label: '数据处理', icon: 'analytics-outline' },
    { key: ToolCategory.UI, label: 'UI组件', icon: 'color-palette-outline' },
    { key: ToolCategory.BUSINESS, label: '业务组件', icon: 'business-outline' },
  ])

  // 工具市场配置
  const config = reactive<ToolMarketplaceConfig>({
    enableAutoUpdate: true,
    checkUpdateInterval: 24,
    maxConcurrentInstalls: 3,
    enableUsageStats: true,
    defaultCategory: ToolCategory.ALL,
    allowedCategories: Object.values(ToolCategory),
    maxToolsPerPage: 20,
  })

  // 计算属性：过滤后的工具列表
  const filteredTools = computed(() => {
    let result = tools.value

    // 分类过滤
    if (searchParams.category && searchParams.category !== ToolCategory.ALL) {
      result = result.filter((tool) => tool.category === searchParams.category)
    }

    // 关键词搜索
    if (searchParams.keyword) {
      const keyword = searchParams.keyword.toLowerCase()
      result = result.filter(
        (tool) =>
          tool.name.toLowerCase().includes(keyword) ||
          tool.displayName.toLowerCase().includes(keyword) ||
          tool.description.toLowerCase().includes(keyword) ||
          tool.tags.some((tag) => tag.toLowerCase().includes(keyword)) ||
          tool.author.toLowerCase().includes(keyword),
      )
    }

    return result
  })

  // 计算属性：已安装工具数量
  const installedToolsCount = computed(
    () => tools.value.filter((tool) => tool.status === ToolStatus.INSTALLED).length,
  )

  // 计算属性：可用工具数量
  const availableToolsCount = computed(
    () => tools.value.filter((tool) => tool.status === ToolStatus.AVAILABLE).length,
  )

  // 监听搜索查询变化
  watch(
    searchQuery,
    (newQuery) => {
      searchParams.keyword = newQuery
    },
    { immediate: true },
  )

  // 监听活动分类变化
  watch(
    activeCategory,
    (newCategory) => {
      searchParams.category = newCategory
    },
    { immediate: true },
  )

  /**
   * 初始化工具市场
   */
  const initializeMarketplace = async () => {
    if (initialized.value) return

    try {
      loading.value = true
      error.value = null

      await loadToolList()
      initialized.value = true
    } catch (err) {
      error.value = err instanceof Error ? err.message : '初始化失败'
      console.error('初始化工具市场失败:', err)
    } finally {
      loading.value = false
    }
  }

  /**
   * 加载工具列表 - 模拟数据
   */
  const loadToolList = async () => {
    try {
      // 模拟API延迟
      await new Promise((resolve) => setTimeout(resolve, 300))

      // 模拟工具数据
      const mockTools: Tool[] = [
        {
          id: 'tool-1',
          name: 'json-formatter',
          displayName: 'JSON格式化工具',
          description: '美化和格式化JSON数据，支持语法高亮和折叠',
          version: '1.0.0',
          author: 'HubTools',
          category: ToolCategory.UTILITY,
          tags: ['JSON', '格式化', '美化'],
          icon: 'code-outline',
          status: ToolStatus.AVAILABLE,
          permissionLevel: 'public' as any,
          size: 50,
          createTime: '2024-01-01T00:00:00Z',
          updateTime: '2024-01-01T00:00:00Z',
          createBy: 'system',
          updateBy: 'system',
          activeFlag: 'Y',
        },
        {
          id: 'tool-2',
          name: 'color-picker',
          displayName: '颜色选择器',
          description: '强大的颜色选择工具，支持多种颜色格式',
          version: '2.1.0',
          author: 'HubTools',
          category: ToolCategory.UI,
          tags: ['颜色', '选择器', 'UI'],
          icon: 'color-palette-outline',
          status: ToolStatus.INSTALLED,
          permissionLevel: 'public' as any,
          size: 80,
          createTime: '2024-01-01T00:00:00Z',
          updateTime: '2024-01-01T00:00:00Z',
          createBy: 'system',
          updateBy: 'system',
          activeFlag: 'Y',
        },
      ]

      tools.value = mockTools
    } catch (err) {
      console.error('加载工具列表失败:', err)
      throw err
    }
  }

  /**
   * 设置视图模式
   */
  const setViewMode = (mode: ViewMode) => {
    viewMode.value = mode
  }

  /**
   * 处理搜索
   */
  const handleSearch = (query: string) => {
    searchQuery.value = query
  }

  /**
   * 处理分类变化
   */
  const handleCategoryChange = (category: ToolCategory) => {
    activeCategory.value = category
  }

  /**
   * 处理工具安装
   */
  const handleInstallTool = async (tool: Tool) => {
    try {
      // 更新状态为安装中
      const toolIndex = tools.value.findIndex((t) => t.id === tool.id)
      if (toolIndex >= 0) {
        tools.value[toolIndex].status = ToolStatus.INSTALLING
      }

      // 模拟安装过程
      await new Promise((resolve) => setTimeout(resolve, 1000))

      // 更新工具状态为已安装
      if (toolIndex >= 0) {
        tools.value[toolIndex].status = ToolStatus.INSTALLED
        tools.value[toolIndex].installTime = new Date().toISOString()
      }

      message.success(`${tool.displayName} 安装成功`)
    } catch (err) {
      // 恢复状态
      const toolIndex = tools.value.findIndex((t) => t.id === tool.id)
      if (toolIndex >= 0) {
        tools.value[toolIndex].status = ToolStatus.AVAILABLE
      }

      const errorMsg = err instanceof Error ? err.message : '安装失败'
      message.error(`${tool.displayName} 安装失败: ${errorMsg}`)
    }
  }

  /**
   * 处理工具卸载
   */
  const handleUninstallTool = async (tool: Tool) => {
    try {
      // 模拟卸载过程
      await new Promise((resolve) => setTimeout(resolve, 500))

      // 更新工具状态
      const toolIndex = tools.value.findIndex((t) => t.id === tool.id)
      if (toolIndex >= 0) {
        tools.value[toolIndex].status = ToolStatus.AVAILABLE
        tools.value[toolIndex].installTime = undefined
      }

      message.success(`${tool.displayName} 卸载成功`)
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '卸载失败'
      message.error(`${tool.displayName} 卸载失败: ${errorMsg}`)
    }
  }

  /**
   * 处理工具配置
   */
  const handleConfigureTool = (tool: Tool) => {
    selectedTool.value = tool
    configDialogVisible.value = true
  }

  /**
   * 处理工具预览
   */
  const handlePreviewTool = (tool: Tool) => {
    selectedTool.value = tool
    previewDrawerVisible.value = true
  }

  /**
   * 保存工具配置
   */
  const handleSaveToolConfig = async (config: Record<string, any>) => {
    if (!selectedTool.value) return

    try {
      // 模拟配置保存
      await new Promise((resolve) => setTimeout(resolve, 300))

      // 更新本地配置
      const toolIndex = tools.value.findIndex((t) => t.id === selectedTool.value!.id)
      if (toolIndex >= 0) {
        tools.value[toolIndex].userConfig = config
      }

      message.success(`${selectedTool.value.displayName} 配置保存成功`)
      configDialogVisible.value = false
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '配置保存失败'
      message.error(`配置保存失败: ${errorMsg}`)
    }
  }

  // 组件挂载时初始化
  onMounted(() => {
    initializeMarketplace()
  })

  return {
    // 响应式状态
    tools,
    filteredTools,
    categories,
    selectedTool,
    loading,
    error,
    initialized,

    // 界面状态
    searchQuery,
    activeCategory,
    viewMode,
    previewDrawerVisible,
    configDialogVisible,

    // 配置
    config,

    // 统计信息
    installedToolsCount,
    availableToolsCount,

    // 方法
    initializeMarketplace,
    setViewMode,
    handleSearch,
    handleCategoryChange,
    handleInstallTool,
    handleUninstallTool,
    handleConfigureTool,
    handlePreviewTool,
    handleSaveToolConfig,
  }
}
