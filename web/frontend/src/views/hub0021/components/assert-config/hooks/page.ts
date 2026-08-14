/**
 * 断言配置列表页面级 Hook
 * 管理对话框、事件处理等页面级交互逻辑
 */

import { rsConfirm } from '@/ui'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { useAppMessage } from '@/composables/useAppMessage'
import type { Ref } from 'vue'
import { ref } from 'vue'
import { getRouteAssertionById } from '../../../api'
import { useAssertConfigService } from './service'
import type { AssertConfig } from './types'

/**
 * 断言配置列表页面级 Hook
 * @param gridRef 表格引用（可选）
 * @param routeConfigId 路由配置ID（必填）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useAssertConfigPage(
  routeConfigId: Ref<string | undefined> | string,
  gridRef?: Ref<any> | any,
  searchFormRef?: Ref<any> | any
) {
  const message = useAppMessage()

  // 使用服务层
  const service = useAssertConfigService(routeConfigId, searchFormRef)

  // ============= 对话框状态 =============

  /** 表单对话框是否显示 */
  const formDialogVisible = ref(false)

  /** 表单对话框模式：create | edit | view */
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')

  /** 当前编辑的断言 */
  const currentEditAssert = ref<Partial<AssertConfig> | null>(null)

  // ============= 对话框操作 =============

  /**
   * 打开新增对话框
   */
  const openAddDialog = () => {
    currentEditAssert.value = {
      assertionOrder: 100,
      assertionType: 'HEADER' as any,
      assertionOperator: 'EQUAL' as any,
      caseSensitive: 'Y',
      isRequired: 'Y',
      activeFlag: 'Y',
    }
    formDialogMode.value = 'create'
    formDialogVisible.value = true
  }

  /**
   * 打开编辑对话框
   */
  const openEditDialog = async (assert: AssertConfig) => {
    try {
      // 根据断言ID查询最新的断言配置
      const response = await getRouteAssertionById(assert.routeAssertionId)
      
      if (isApiSuccess(response)) {
        const latestAssert = parseJsonData<AssertConfig>(response)
        if (latestAssert) {
          currentEditAssert.value = { ...latestAssert }
          formDialogMode.value = 'edit'
          formDialogVisible.value = true
          return
        }
      }
      
      // 如果查询失败，使用传入的数据作为后备
      message.warning(getApiMessage(response, '获取断言配置失败，使用缓存数据'))
      currentEditAssert.value = { ...assert }
      formDialogMode.value = 'edit'
      formDialogVisible.value = true
    } catch (error) {
      console.error('打开编辑对话框失败:', error)
      message.error('获取断言配置失败')
      // 出错时使用传入的数据
      currentEditAssert.value = { ...assert }
      formDialogMode.value = 'edit'
      formDialogVisible.value = true
    }
  }

  /**
   * 打开查看对话框
   */
  const openViewDialog = async (assert: AssertConfig) => {
    try {
      currentEditAssert.value = { ...assert }
      formDialogMode.value = 'view'
      formDialogVisible.value = true
    } catch (error) {
      console.error('打开查看对话框失败:', error)
    }
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditAssert.value = null
    formDialogMode.value = 'create'
  }

  // ============= 表单提交 =============

  /**
   * 处理表单提交
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) {
      return
    }
    try {
      const isEdit = formDialogMode.value === 'edit'
      const assertData = {
        ...formData,
        routeConfigId: typeof routeConfigId === 'string' ? routeConfigId : routeConfigId?.value,
      }

      let success = false
      if (isEdit) {
        success = await service.editAssert(assertData)
      } else {
        success = await service.addAssert(assertData)
      }

      if (success) {
        closeFormDialog()
      }
    } catch (error) {
      console.error('提交表单失败:', error)
    }
  }

  // ============= 搜索和工具栏 =============

  /**
   * 处理搜索
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  /**
   * 处理工具栏按钮点击
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break
      case 'delete':
        await handleBatchDelete()
        break
      case 'search':
        await handleSearch(formData)
        break
      case 'reset':
        // 重置搜索表单后重新加载
        await service.handleReset()
        break
      default:
        break
    }
  }

  // ============= 菜单操作 =============

  /**
   * 处理右键菜单点击
   */
  const handleMenuClick = async (params: { key: string; row?: any }) => {
    const { key, row } = params
    if (!row) {
      return
    }

    const assert = row as AssertConfig

    switch (key) {
      case 'view':
        await openViewDialog(assert)
        break
      case 'edit':
        await openEditDialog(assert)
        break
      case 'toggle-status':
        await handleToggleStatus(assert)
        break
      case 'delete':
        await handleDelete(assert)
        break
      default:
        break
    }
  }

  // ============= CRUD 操作 =============

  /**
   * 处理删除
   */
  const handleDelete = async (assert: AssertConfig) => {
    const confirmed = await rsConfirm.warning({
      title: '确认删除',
      description: `确定要删除断言"${assert.assertionName}"吗？`,
    })

    if (confirmed) {
      await service.deleteAssert(assert.routeAssertionId)
    }
  }

  /**
   * 处理批量删除
   * 使用 getActiveRows：优先勾选行，无勾选时回退当前高亮行
   */
  const handleBatchDelete = async () => {
    if (!gridRef?.value) {
      message.warning('无法获取表格引用')
      return
    }

    const selectedRows = (gridRef.value.getActiveRows?.() || []) as AssertConfig[]
    if (selectedRows.length === 0) {
      message.warning('请选择或点击要删除的断言')
      return
    }

    const confirmed = await rsConfirm.warning({
      title: selectedRows.length > 1 ? '确认批量删除' : '确认删除',
      description:
        selectedRows.length > 1
          ? `确定要删除选中的 ${selectedRows.length} 个断言吗？`
          : `确定要删除断言"${selectedRows[0].assertionName}"吗？`,
    })

    if (!confirmed) {
      return
    }

    const routeAssertionIds = selectedRows
      .map((row) => row.routeAssertionId)
      .filter((id): id is string => Boolean(id))

    if (routeAssertionIds.length === 0) {
      message.warning('断言配置ID无效')
      return
    }

    const success =
      routeAssertionIds.length === 1
        ? await service.deleteAssert(routeAssertionIds[0])
        : await service.deleteAsserts(routeAssertionIds)

    if (success) {
      gridRef.value.clearSelection?.()
    }
  }

  /**
   * 处理切换状态
   */
  const handleToggleStatus = async (assert: AssertConfig) => {
    await service.toggleAssertStatus(assert)
  }

  // ============= 分页 =============


  return {
    service,
    formDialogVisible,
    formDialogMode,
    currentEditAssert,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
    handleFormSubmit,
    handleSearch,
    handleToolbarClick,
    handleMenuClick,
    handleDelete,
    handleBatchDelete,
    handleToggleStatus,
  }
}

/**
 * 断言配置列表页面级 Hook 类型
 */
export type AssertConfigPage = ReturnType<typeof useAssertConfigPage>

