/**
 * Hub0043 配置管理页面级 Hook
 * - 组合 useConfigService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useGDialog } from '@/components/gdialog'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { nextTick, ref, watch } from 'vue'
import type { Config } from '../../../types'
import { useConfigService } from './service'

/**
 * 配置管理页面级 Hook
 * @param gridRef 表格引用
 * @param searchFormRef 搜索表单引用
 * @param formRef 表单引用（可选，用于视图切换模式）
 * @param options 配置选项
 */
export function useConfigPage(
  gridRef?: Ref<any> | any,
  searchFormRef?: Ref<any> | any,
  formRef?: Ref<any> | any,
  options?: {
    /** 是否使用视图切换模式（true: 表单视图，false: 对话框模式） */
    useViewMode?: boolean
    /** 视图切换回调 */
    onViewChange?: (view: 'list' | 'form') => void
    /** 历史事件回调 */
    onHistoryClick?: (config: Config) => void
  }
) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 业务服务（包含 model、增删改查等）
  const service = useConfigService(searchFormRef)

  // 配置选项
  const useViewMode = options?.useViewMode ?? false
  const onViewChange = options?.onViewChange
  const onHistoryClick = options?.onHistoryClick

  // 视图状态（视图切换模式使用）
  const currentView = ref<'list' | 'form'>('list')
  const savedSearchFormData = ref<Record<string, any> | null>(null)

  // 表单对话框状态（对话框模式使用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditConfig = ref<Config | null>(null)
  const submitting = ref(false)

  /**
   * 打开新增配置对话框或切换到表单视图
   */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditConfig.value = null
    
    if (useViewMode) {
      // 视图切换模式：保存搜索表单数据并切换到表单视图
      const formData = searchFormRef?.value?.getFormData?.() || {}
      savedSearchFormData.value = { ...formData }
      const namespaceId = formData.namespaceId
      
      currentView.value = 'form'
      onViewChange?.('form')
      
      // 如果有命名空间，在表单初始化后自动填充
      // 需要等待视图切换和组件渲染完成
      if (namespaceId) {
        nextTick(() => {
          // 再次等待确保 GDataForm 组件完全渲染
          nextTick(() => {
            if (formRef?.value?.setFormData) {
              const currentFormData = formRef.value.getFormData() || {}
              formRef.value.setFormData({
                ...currentFormData,
                namespaceId: namespaceId,
              })
            }
          })
        })
      }
    } else {
      // 对话框模式：显示对话框
      formDialogVisible.value = true
    }
  }

  /**
   * 打开编辑配置对话框或切换到表单视图
   */
  const openEditDialog = async (config: Config) => {
    try {
      const detailConfig = await service.getConfigDetail(
        config.namespaceId,
        config.groupName || 'DEFAULT_GROUP',
        config.configDataId
      )
      
      if (!detailConfig) {
        message.error('获取配置详情失败')
        return
      }

      formDialogMode.value = 'edit'
      currentEditConfig.value = detailConfig
      
      if (useViewMode) {
        // 视图切换模式：保存搜索表单数据并切换到表单视图
        const formData = searchFormRef?.value?.getFormData?.() || {}
        savedSearchFormData.value = { ...formData }
        currentView.value = 'form'
        onViewChange?.('form')
      } else {
        // 对话框模式：显示对话框
        formDialogVisible.value = true
      }
    } catch (error) {
      message.error('获取配置详情失败')
    }
  }

  /**
   * 打开查看配置对话框或切换到表单视图
   */
  const openViewDialog = async (config: Config) => {
    try {
      const detailConfig = await service.getConfigDetail(
        config.namespaceId,
        config.groupName || 'DEFAULT_GROUP',
        config.configDataId
      )
      
      if (!detailConfig) {
        message.error('获取配置详情失败')
        return
      }

      formDialogMode.value = 'view'
      currentEditConfig.value = detailConfig
      
      if (useViewMode) {
        // 视图切换模式：保存搜索表单数据并切换到表单视图
        const formData = searchFormRef?.value?.getFormData?.() || {}
        savedSearchFormData.value = { ...formData }
        currentView.value = 'form'
        onViewChange?.('form')
      } else {
        // 对话框模式：显示对话框
        formDialogVisible.value = true
      }
    } catch (error) {
      message.error('获取配置详情失败')
    }
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditConfig.value = null
  }

  /**
   * 提交表单（新增/编辑）
   */
  const handleSubmit = async (formData: Config): Promise<boolean> => {
    submitting.value = true
    try {
      let success = false
      if (formDialogMode.value === 'create') {
        success = await service.addConfig({
          ...formData,
          namespaceId: formData.namespaceId,
          configDataId: formData.configDataId,
          configContent: formData.configContent,
        })
      } else if (formDialogMode.value === 'edit') {
        success = await service.editConfig({
          ...formData,
          namespaceId: currentEditConfig.value!.namespaceId,
          groupName: currentEditConfig.value!.groupName || 'DEFAULT_GROUP',
          configDataId: currentEditConfig.value!.configDataId,
          configContent: formData.configContent,
        })
      }

      if (success) {
        if (useViewMode) {
          // 视图切换模式：返回列表视图
          handleBackToList()
        } else {
          // 对话框模式：关闭对话框
          closeFormDialog()
        }
      }
      return success
    } finally {
      submitting.value = false
    }
  }

  /**
   * 返回列表视图（视图切换模式使用）
   */
  const handleBackToList = () => {
    currentView.value = 'list'
    onViewChange?.('list')
    // 清空当前编辑的配置
    if (currentEditConfig.value) {
      currentEditConfig.value = null
    }
  }

  // 监听视图切换，在切换到列表视图时恢复搜索表单数据（视图切换模式）
  if (useViewMode) {
    watch(
      () => currentView.value,
      (newView) => {
        if (newView === 'list' && savedSearchFormData.value) {
          // 等待视图切换和组件渲染完成
          nextTick(() => {
            if (searchFormRef?.value?.setFormData) {
              // 再次等待确保 SearchForm 组件完全渲染
              nextTick(() => {
                searchFormRef.value.setFormData(savedSearchFormData.value!)
              })
            }
          })
        }
      },
      { immediate: false }
    )
  }

  /**
   * 检查命名空间是否为空
   */
  const checkNamespace = (): boolean => {
    const formData = searchFormRef?.value?.getFormData?.() || {}
    if (!formData.namespaceId) {
      message.warning('请先选择命名空间')
      return false
    }
    return true
  }

  /**
   * 处理工具栏按钮点击
   */
  const handleToolbarClick = (key: string) => {
    // 所有操作都需要验证命名空间
    if (!checkNamespace()) {
      return
    }

    switch (key) {
      case 'add':
        openAddDialog()
        break
      case 'edit':
        // 优先使用当前行记录（通过点击行选中的记录）
        const currentRecord = gridRef?.value?.getCurrentRecord?.()
        if (currentRecord) {
          openEditDialog(currentRecord)
          break
        }
        // 如果没有当前行记录，则使用复选框选中的记录
        const selectedRows = gridRef?.value?.getCheckboxRecords() || []
        if (selectedRows.length === 0) {
          message.warning('请先选择要编辑的配置')
          return
        }
        if (selectedRows.length > 1) {
          message.warning('只能编辑一个配置')
          return
        }
        openEditDialog(selectedRows[0])
        break
      case 'delete': {
        // 优先使用当前行记录（通过点击行选中的记录）
        const currentDeleteRecord = gridRef?.value?.getCurrentRecord?.()
        if (currentDeleteRecord) {
          // 如果有当前行记录，直接删除当前行
          handleBatchDelete([currentDeleteRecord])
          break
        }
        
        // 如果没有当前行记录，则使用复选框选中的记录
        const checkboxRows = gridRef?.value?.getCheckboxRecords() || []
        if (checkboxRows.length === 0) {
          message.warning('请先选择要删除的配置')
          return
        }
        
        handleBatchDelete(checkboxRows)
        break
      }
    }
  }

  /**
   * 批量删除配置
   */
  const handleBatchDelete = async (configs: Config[]) => {
    // 生成配置 key 列表
    const configKeys = configs.map(config => config.configDataId).join('\n')
    const confirmed = await gDialog.warning({
      title: '确认批量删除',
      subtitle: `将删除 ${configs.length} 个配置`,
      content: `此操作不可恢复，请谨慎操作\n\n配置ID列表:\n${configKeys}`,
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return
    }

    let successCount = 0
    let failCount = 0
    for (const config of configs) {
      // 批量删除时跳过确认对话框和单个成功/失败提示
      const success = await service.deleteConfig(config, true)
      if (success) {
        successCount++
      } else {
        failCount++
      }
    }

    // 统一显示批量删除结果
    if (successCount > 0 && failCount === 0) {
      message.success(`成功删除 ${successCount} 个配置`)
      await service.handleRefresh()
    } else if (successCount > 0 && failCount > 0) {
      message.warning(`成功删除 ${successCount} 个配置，${failCount} 个配置删除失败`)
      await service.handleRefresh()
    } else if (failCount > 0) {
      message.error(`删除失败，共 ${failCount} 个配置删除失败`)
    }
  }

  /**
   * 处理表格右键菜单
   */
  const handleMenuClick = async (params: { code: string; row?: any }) => {
    if (!params.row) {
      return
    }
    
    // 所有操作都需要验证命名空间（除了 history，因为它只是触发事件）
    if (params.code !== 'history' && !checkNamespace()) {
      return
    }
    
    const row = params.row as Config
    switch (params.code) {
      case 'view':
        await openViewDialog(row)
        break
      case 'edit':
        await openEditDialog(row)
        break
      case 'history':
        // 触发历史事件回调
        onHistoryClick?.(row)
        break
      case 'delete':
        await service.deleteConfig(row)
        await service.handleRefresh()
        break
    }
  }

  /**
   * 搜索处理
   */
  const handleSearch = () => {
    service.handleSearch()
  }

  /**
   * 表单提交处理（适配 GdataFormModal 的提交格式）
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) {
      return
    }
    
    // 验证命名空间
    if (!formData.namespaceId) {
      message.warning('请先选择命名空间')
      return
    }
    
    await handleSubmit(formData as any)
  }

  return {
    // 服务（包含 model 和所有业务方法）
    service,

    // 视图状态（视图切换模式）
    currentView,

    // 对话框状态（对话框模式）
    formDialogVisible,
    formDialogMode,
    currentEditConfig,
    submitting,

    // 对话框方法
    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
    handleSubmit,
    handleFormSubmit,

    // 视图切换方法
    handleBackToList,

    // 工具栏和菜单
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
  }
}

export type ConfigPage = ReturnType<typeof useConfigPage>

