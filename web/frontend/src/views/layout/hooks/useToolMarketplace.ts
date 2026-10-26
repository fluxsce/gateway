import {
  installTool,
  queryTools,
  saveToolConfig,
  uninstallTool,
  type ToolCatalogItem,
} from '@/views/hub0010/api'
import { openToolPage } from '@/views/hub0010/openToolPage'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { computed, ref } from 'vue'
import {
  ToolCategory,
  ToolConfigType,
  ToolPermissionLevel,
  ToolStatus,
  type Tool,
  type ToolCategoryItem,
  type ViewMode,
} from '../types/toolMarketplace'

const textKey: Record<string, string> = {
  'json-formatter': 'jsonFormatter',
  timestamp: 'timestamp',
  base64: 'base64',
  'jwt-decoder': 'jwtDecoder',
}

/**
 * 工具市场状态。目录和安装结果来自 hub0010，可见工具由角色里的工具资源码决定。
 */
export function useToolMarketplace() {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0010')

  const loading = ref(false)
  const error = ref<string | null>(null)
  const rawTools = ref<ToolCatalogItem[]>([])
  const selectedTool = ref<Tool | null>(null)

  const searchQuery = ref('')
  const activeCategory = ref<ToolCategory>(ToolCategory.ALL)
  const previewDrawerVisible = ref(false)
  const configDialogVisible = ref(false)
  const viewMode = ref<ViewMode>('grid')

  const canInstall = computed(() => store.user.hasPermission('hub0010:install'))
  const canUninstall = computed(() => store.user.hasPermission('hub0010:uninstall'))
  const canConfigure = computed(() => store.user.hasPermission('hub0010:config'))

  const categories = computed<ToolCategoryItem[]>(() => [
    { key: ToolCategory.ALL, label: t('categories.all') },
    { key: ToolCategory.UTILITY, label: t('categories.utility') },
    { key: ToolCategory.CHART, label: t('categories.chart') },
    { key: ToolCategory.FORM, label: t('categories.form') },
    { key: ToolCategory.TABLE, label: t('categories.table') },
    { key: ToolCategory.LAYOUT, label: t('categories.layout') },
    { key: ToolCategory.DATA, label: t('categories.data') },
    { key: ToolCategory.UI, label: t('categories.ui') },
    { key: ToolCategory.BUSINESS, label: t('categories.business') },
  ])

  const tools = computed<Tool[]>(() => rawTools.value.map(toTool))

  const filteredTools = computed(() => {
    let result = tools.value
    if (activeCategory.value !== ToolCategory.ALL) {
      result = result.filter((tool) => tool.category === activeCategory.value)
    }
    const keyword = searchQuery.value.trim().toLowerCase()
    if (!keyword) return result
    return result.filter(
      (tool) =>
        tool.name.toLowerCase().includes(keyword) ||
        tool.displayName.toLowerCase().includes(keyword) ||
        tool.description.toLowerCase().includes(keyword) ||
        tool.tags.some((tag) => tag.toLowerCase().includes(keyword)) ||
        tool.author.toLowerCase().includes(keyword),
    )
  })

  const emptyBecauseNoGrant = computed(
    () => !loading.value && rawTools.value.length === 0 && !searchQuery.value.trim(),
  )

  /**
   * 拉取目录。失败时清空列表，避免继续展示上一次的结果。
   */
  const loadToolList = async () => {
    loading.value = true
    error.value = null
    try {
      const result = await queryTools()
      if (!isApiSuccess(result)) {
        throw new Error(getApiMessage(result, t('message.loadFail', { error: '' })))
      }
      rawTools.value = parseJsonData<ToolCatalogItem[]>(result, [])
    } catch (err) {
      rawTools.value = []
      const text = err instanceof Error ? err.message : String(err)
      error.value = text
      message.error(t('message.loadFail', { error: text }))
    } finally {
      loading.value = false
    }
  }

  const setViewMode = (mode: ViewMode) => {
    viewMode.value = mode
  }

  const patchStatus = (toolId: string, status: ToolCatalogItem['status']) => {
    const index = rawTools.value.findIndex((item) => item.id === toolId)
    if (index < 0) return
    rawTools.value[index] = { ...rawTools.value[index], status }
  }

  const handleInstallTool = async (tool: Tool) => {
    patchStatus(tool.id, 'installing')
    try {
      const result = await installTool(tool.id, tool.defaultConfig)
      if (!isApiSuccess(result)) {
        throw new Error(getApiMessage(result))
      }
      message.success(t('message.installOk', { name: tool.displayName }))
      await loadToolList()
    } catch (err) {
      patchStatus(tool.id, 'available')
      const text = err instanceof Error ? err.message : String(err)
      message.error(t('message.installFail', { name: tool.displayName, error: text }))
    }
  }

  const handleUninstallTool = async (tool: Tool) => {
    try {
      const result = await uninstallTool(tool.id)
      if (!isApiSuccess(result)) {
        throw new Error(getApiMessage(result))
      }
      message.success(t('message.uninstallOk', { name: tool.displayName }))
      await loadToolList()
    } catch (err) {
      const text = err instanceof Error ? err.message : String(err)
      message.error(t('message.uninstallFail', { name: tool.displayName, error: text }))
    }
  }

  const handleConfigureTool = (tool: Tool) => {
    selectedTool.value = tool
    configDialogVisible.value = true
  }

  const handlePreviewTool = (tool: Tool) => {
    selectedTool.value = tool
    previewDrawerVisible.value = true
  }

  const handleOpenTool = (tool: Tool) => {
    previewDrawerVisible.value = false
    const opened = openToolPage(tool.id, tool.displayName)
    if (!opened) message.error(t('work.unavailable'))
    return opened
  }

  const handleSaveToolConfig = async (config: Record<string, string | number>) => {
    if (!selectedTool.value) return
    const current = selectedTool.value
    try {
      const result = await saveToolConfig(current.id, config)
      if (!isApiSuccess(result)) {
        throw new Error(getApiMessage(result))
      }
      message.success(t('message.configOk', { name: current.displayName }))
      configDialogVisible.value = false
      await loadToolList()
    } catch (err) {
      const text = err instanceof Error ? err.message : String(err)
      message.error(t('message.configFail', { error: text }))
    }
  }

  /**
   * 把目录行补成卡片使用的 Tool。展示名跟随当前语言。
   */
  function toTool(item: ToolCatalogItem): Tool {
    const key = textKey[item.id]
    const displayName = key ? t(`tools.${key}.name`) : item.name
    const description = key ? t(`tools.${key}.description`) : ''
    const status =
      item.status === 'installed'
        ? ToolStatus.INSTALLED
        : item.status === 'installing'
          ? ToolStatus.INSTALLING
          : ToolStatus.AVAILABLE
    return {
      id: item.id,
      name: item.name,
      displayName,
      description,
      version: item.version,
      author: item.author,
      category: item.category as ToolCategory,
      tags: item.tags ?? [],
      icon: item.icon,
      status,
      permissionLevel: ToolPermissionLevel.PUBLIC,
      installTime: item.installTime,
      configSchema: (item.configSchema ?? []).map((field) => ({
        key: field.key,
        label: t(`config.${field.key}`),
        type: field.type === 'select' ? ToolConfigType.SELECT : ToolConfigType.STRING,
        options: (field.options ?? []).map((option) => ({
          label: String(option.value),
          value: option.value,
        })),
      })),
      defaultConfig: item.defaultConfig ?? {},
      userConfig: item.userConfig,
      createTime: item.installTime ?? '',
      updateTime: item.installTime ?? '',
      createBy: '',
      updateBy: '',
      activeFlag: 'Y',
    }
  }

  return {
    filteredTools,
    categories,
    selectedTool,
    loading,
    error,
    emptyBecauseNoGrant,
    searchQuery,
    activeCategory,
    viewMode,
    previewDrawerVisible,
    configDialogVisible,
    canInstall,
    canUninstall,
    canConfigure,
    setViewMode,
    handleInstallTool,
    handleUninstallTool,
    handleConfigureTool,
    handlePreviewTool,
    handleOpenTool,
    handleSaveToolConfig,
    initializeMarketplace: loadToolList,
  }
}
