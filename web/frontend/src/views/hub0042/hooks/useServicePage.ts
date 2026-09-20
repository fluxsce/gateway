/**
 * 服务监控页面级 Hook
 * - 组合 useServiceService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import type { RsSearchFormExpose } from '@/components/form/rs-search'
import type { RsGridExpose } from '@/components/rs-grid'
import { useAppMessage } from '@/composables/useAppMessage'
import { rsConfirm } from '@/ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { Namespace } from '../../hub0041/types'
import type { Service } from '../types'
import { canServiceAction } from './model'
import { useServiceService } from './useServiceService'

/**
 * 服务监控页面级 Hook
 */
export function useServicePage(
  gridRef?: Ref<RsGridExpose | null>,
  searchFormRef?: Ref<RsSearchFormExpose | null>,
) {
  const message = useAppMessage()
  // 业务服务（包含 model、增删改查等）
  const service = useServiceService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditService = ref<Service | null>(null)
  const submitting = ref(false)
  const selectedNamespace = ref<Namespace | null>(null)

  /**
   * 打开新增服务对话框
   * @param namespace 选中的命名空间（可选），用于预填充 namespaceId
   */
  const openAddDialog = (namespace?: { namespaceId: string } | null) => {
    formDialogMode.value = 'create'
    // 如果有选中的命名空间，预填充 namespaceId
    if (namespace?.namespaceId) {
      currentEditService.value = {
        namespaceId: namespace.namespaceId,
        groupName: 'DEFAULT_GROUP',
      } as Service
    } else {
      currentEditService.value = null
    }
    formDialogVisible.value = true
  }

  /**
   * 打开编辑服务对话框
   */
  const openEditDialog = async (serviceItem: Service) => {
    try {
      const detailService = await service.getServiceDetail(
        serviceItem.namespaceId,
        serviceItem.groupName,
        serviceItem.serviceName
      )
      
      if (!detailService) {
        message.error('获取服务详情失败')
        return
      }

      formDialogMode.value = 'edit'
      currentEditService.value = detailService
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取服务详情失败')
    }
  }

  /**
   * 打开查看服务对话框
   */
  const openViewDialog = async (serviceItem: Service) => {
    try {
      const detailService = await service.getServiceDetail(
        serviceItem.namespaceId,
        serviceItem.groupName,
        serviceItem.serviceName
      )
      
      if (!detailService) {
        message.error('获取服务详情失败')
        return
      }

      formDialogMode.value = 'view'
      currentEditService.value = detailService
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取服务详情失败')
    }
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditService.value = null
  }

  /**
   * 提交表单（新增/编辑）
   */
  const handleSubmit = async (formData: Service) => {
    submitting.value = true
    try {
      let success = false
      if (formDialogMode.value === 'create') {
        success = await service.addService(formData)
      } else if (formDialogMode.value === 'edit') {
        success = await service.editService({
          ...formData,
          namespaceId: currentEditService.value!.namespaceId,
          groupName: currentEditService.value!.groupName,
          serviceName: currentEditService.value!.serviceName,
        })
      }

      if (success) {
        closeFormDialog()
        await service.handleRefresh()
      }
    } finally {
      submitting.value = false
    }
  }

  /** 与 hub0082 一致：优先勾选，无勾选时取高亮行 */
  const getSelectedRows = (): Service[] => {
    return (gridRef?.value?.getActiveRows?.() || []) as Service[]
  }

  /**
   * 处理工具栏按钮点击
   * @param key 按钮key
   * @param namespace 选中的命名空间（可选），用于新增服务时预填充
   */
  const handleToolbarClick = async (key: string, namespace?: { namespaceId: string } | null) => {
    if (!canServiceAction(key) && key !== 'search' && key !== 'reset') {
      message.warning('没有操作权限')
      return
    }
    const scopedNamespace = namespace || selectedNamespace.value
    switch (key) {
      case 'add': {
        const namespaceId = scopedNamespace?.namespaceId || readNamespaceId()
        if (!namespaceId) {
          message.warning('请先选择命名空间')
          return
        }
        openAddDialog(scopedNamespace || { namespaceId })
        break
      }
      case 'edit':
        const selectedRows = getSelectedRows()
        if (selectedRows.length === 0) {
          message.warning('请先选择要编辑的服务')
          return
        }
        if (selectedRows.length > 1) {
          message.warning('只能编辑一个服务')
          return
        }
        await openEditDialog(selectedRows[0])
        break
      case 'delete': {
        const deleteRows = getSelectedRows()
        if (deleteRows.length === 0) {
          message.warning('请先选择要删除的服务')
          return
        }
        if (deleteRows.length === 1) {
          const success = await service.deleteService(deleteRows[0])
          if (success) {
            await service.handleRefresh()
          }
          return
        }
        await handleBatchDelete(deleteRows)
        break
      }
      case 'batchDelete':
        await handleBatchDeleteFromGrid()
        break
    }
  }

  const handleBatchDeleteFromGrid = async () => {
    if (!gridRef?.value) {
      message.warning('表格未就绪')
      return
    }
    const selectedRows = (gridRef.value.getActiveRows?.() || []) as Service[]
    if (selectedRows.length === 0) {
      message.warning('请先勾选要删除的服务')
      return
    }
    await handleBatchDelete(selectedRows)
  }

  /**
   * 批量删除服务
   */
  const handleBatchDelete = async (services: Service[]) => {
    if (!canServiceAction('batchDelete') && services.length > 1) {
      message.warning('没有批量删除权限')
      return
    }
    const confirmed = await rsConfirm.warning({
      title: '确认批量删除',
      subtitle: `将删除 ${services.length} 个服务`,
      description: '此操作不可恢复，请谨慎操作',
      confirmText: '确定删除',
      cancelText: '取消',
      width: 500,
    })

    if (!confirmed) {
      return
    }

    const result = await service.batchDeleteServices(services.map((item) => ({
      namespaceId: item.namespaceId,
      groupName: item.groupName,
      serviceName: item.serviceName,
      instanceName: item.instanceName,
    })))
    if (!result) {
      return
    }
    if (result.successCount > 0) {
      message.success(`成功删除 ${result.successCount} 个服务`)
      await service.handleRefresh()
    }
    if (result.failCount > 0) {
      message.warning(result.error
        ? `${result.failCount} 个服务删除失败：${result.error}`
        : `${result.failCount} 个服务删除失败`)
    }
  }

  /**
   * 处理表格右键菜单
   */
  const handleMenuClick = async ({ key, row }: { key: string; row?: Service }) => {
    switch (key) {
      case 'batchDelete':
        if (!canServiceAction('batchDelete')) {
          message.warning('没有操作权限')
          return
        }
        await handleBatchDeleteFromGrid()
        break
      case 'edit':
        if (!row) return
        if (!canServiceAction(key)) {
          message.warning('没有操作权限')
          return
        }
        await openEditDialog(row)
        break
      case 'delete':
        if (!row) return
        if (!canServiceAction(key)) {
          message.warning('没有操作权限')
          return
        }
        await service.deleteService(row)
        await service.handleRefresh()
        break
      default:
        break
    }
  }

  const readNamespaceId = (formData?: Record<string, any>) => {
    const data = formData || searchFormRef?.value?.getFormData?.() || {}
    return String(data.namespaceId || '').trim()
  }

  const handleSearch = async (formData?: Record<string, any>) => {
    const namespaceId = readNamespaceId(formData)
    if (!namespaceId) {
      selectedNamespace.value = null
      service.model.setServiceList([])
      return
    }
    selectedNamespace.value = { namespaceId } as Namespace
    await service.handleSearch(namespaceId, formData)
  }

  const handleReset = async () => {
    selectedNamespace.value = null
    await service.handleReset()
  }

  /**
   * 表单提交处理（适配 RsDataFormModal 的提交格式）
   */
  const handleFormSubmit = (formData?: Record<string, any>) => {
    if (formData) {
      handleSubmit(formData as any)
    }
  }

  /**
   * 服务表单提交（自动填充命名空间ID）
   */
  const handleServiceFormSubmit = (formData?: Record<string, any>, namespace?: { namespaceId: string } | null) => {
    if (formData) {
      const scoped = namespace || selectedNamespace.value
      if (scoped && !formData.namespaceId) {
        formData.namespaceId = scoped.namespaceId
      }
      handleFormSubmit(formData)
    }
  }

  return {
    // 服务（包含 model 和所有业务方法）
    service,

    // 对话框状态
    formDialogVisible,
    formDialogMode,
    currentEditService,
    submitting,
    selectedNamespace,

    // 对话框方法
    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
    handleSubmit,
    handleFormSubmit,
    handleServiceFormSubmit,

    // 工具栏和菜单
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
    handleReset,
  }
}

export type ServicePage = ReturnType<typeof useServicePage>

