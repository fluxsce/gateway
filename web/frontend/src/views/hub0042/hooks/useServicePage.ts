/**
 * 服务监控页面级 Hook
 * - 组合 useServiceService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useGDialog } from '@/components/gdialog'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { Service } from '../types'
import { useServiceService } from './useServiceService'

/**
 * 服务监控页面级 Hook
 */
export function useServicePage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 业务服务（包含 model、增删改查等）
  const service = useServiceService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditService = ref<Service | null>(null)
  const submitting = ref(false)

  /**
   * 打开新增服务对话框
   */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditService.value = null
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

  /**
   * 处理工具栏按钮点击
   */
  const handleToolbarClick = (key: string) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break
      case 'edit':
        const selectedRows = gridRef?.value?.getCheckboxRecords() || []
        if (selectedRows.length === 0) {
          message.warning('请先选择要编辑的服务')
          return
        }
        if (selectedRows.length > 1) {
          message.warning('只能编辑一个服务')
          return
        }
        openEditDialog(selectedRows[0])
        break
      case 'delete':
        const deleteRows = gridRef?.value?.getCheckboxRecords() || []
        if (deleteRows.length === 0) {
          message.warning('请先选择要删除的服务')
          return
        }
        handleBatchDelete(deleteRows)
        break
    }
  }

  /**
   * 批量删除服务
   */
  const handleBatchDelete = async (services: Service[]) => {
    const confirmed = await gDialog.warning({
      title: '确认批量删除',
      subtitle: `将删除 ${services.length} 个服务`,
      content: '此操作不可恢复，请谨慎操作',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return
    }

    let successCount = 0
    for (const serviceItem of services) {
      const success = await service.deleteService(serviceItem)
      if (success) {
        successCount++
      }
    }

    if (successCount > 0) {
      message.success(`成功删除 ${successCount} 个服务`)
      await service.handleRefresh()
    }
  }

  /**
   * 处理表格右键菜单（适配 GGrid 的事件格式）
   */
  const handleMenuClick = async (params: { code: string; row?: any }) => {
    if (!params.row) {
      return
    }
    const row = params.row as Service
    switch (params.code) {
      case 'view':
        // view 操作现在由父组件处理，通过 openServiceDetail
        // 这里可以触发一个事件或者直接调用父组件的方法
        break
      case 'edit':
        await openEditDialog(row)
        break
      case 'delete':
        await service.deleteService(row)
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
  const handleFormSubmit = (formData?: Record<string, any>) => {
    if (formData) {
      handleSubmit(formData as any)
    }
  }

  /**
   * 服务表单提交（自动填充命名空间ID）
   * @param selectedNamespace 选中的命名空间（可选）
   */
  const handleServiceFormSubmit = (formData?: Record<string, any>, selectedNamespace?: { namespaceId: string } | null) => {
    if (formData) {
      // 如果选中了命名空间，自动填充命名空间ID
      if (selectedNamespace && !formData.namespaceId) {
        formData.namespaceId = selectedNamespace.namespaceId
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
  }
}

export type ServicePage = ReturnType<typeof useServicePage>

