/**
 * 预警日志页面级 Hook
 * - 组合 useAlertLogService（纯业务逻辑）
 * - 处理查看对话框、工具栏、右键菜单等页面交互
 */

import { useGDialog } from '@/components/gdialog'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { AlertLog } from '../types'
import { useAlertLogService } from './service'

/**
 * 预警日志页面级 Hook
 * @param gridRef Grid 组件引用（可选）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useAlertLogPage(
  gridRef?: Ref<any> | any,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 业务服务（包含 model、增删改查等）
  const service = useAlertLogService(searchFormRef)

  // 查看对话框状态
  const viewDialogVisible = ref(false)
  const selectedAlertLogId = ref<string>('')

  // ============= 搜索和分页 =============

  /**
   * 处理搜索
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    service.model.resetPagination()
    await service.loadLogList(searchParams)
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadLogList()
  }

  // ============= 工具栏按钮处理 =============

  /**
   * 处理工具栏按钮点击
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'delete': {
        // 批量删除选中的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const selectedRows = gridRef.value.getCheckboxRecords() || []
        if (selectedRows.length === 0) {
          message.warning('请先选择要删除的日志')
          return
        }
        const alertLogIds = selectedRows.map((row: AlertLog) => row.alertLogId)
        await handleBatchDelete(alertLogIds)
        break
      }

      case 'search': {
        // 如果传递了表单数据，直接使用它进行查询
        await handleSearch(formData)
        break
      }

      default:
        console.warn('未知的工具栏按钮:', key)
    }
  }

  // ============= 对话框处理 =============

  /**
   * 打开查看对话框
   */
  const openViewDialog = async (log: AlertLog) => {
    if (!log.alertLogId) {
      message.warning('日志ID不能为空')
      return
    }

    selectedAlertLogId.value = log.alertLogId
    viewDialogVisible.value = true
  }

  // ============= 右键菜单处理 =============

  /**
   * 处理右键菜单点击
   */
  const handleMenuClick = async ({ menu, row }: { menu: any; row: AlertLog }) => {
    switch (menu.code) {
      case 'view':
        await openViewDialog(row)
        break

      case 'delete':
        await handleDelete(row)
        break

      default:
        console.warn('未知的菜单项:', menu.code)
    }
  }

  // ============= 删除处理 =============

  /**
   * 处理删除
   */
  const handleDelete = async (log: AlertLog) => {
    if (!log.alertLogId) {
      message.warning('日志ID不能为空')
      return
    }

    const confirmed = await gDialog.warning({
      title: '确认删除',
      content: `确定要删除日志"${log.alertLogId}"吗？`,
      positiveText: '删除',
      negativeText: '取消',
    })
    if (!confirmed) return

    await service.deleteLog(log.alertLogId)
  }

  /**
   * 处理批量删除
   */
  const handleBatchDelete = async (alertLogIds: string[]) => {
    if (alertLogIds.length === 0) {
      message.warning('请选择要删除的日志')
      return
    }

    const confirmed = await gDialog.warning({
      title: '确认批量删除',
      content: `确定要删除选中的 ${alertLogIds.length} 条日志吗？`,
      positiveText: '删除',
      negativeText: '取消',
    })
    if (!confirmed) return

    await service.batchDeleteLogs(alertLogIds)
  }

  return {
    // 服务
    service,

    // 对话框状态
    viewDialogVisible,
    selectedAlertLogId,

    // 方法
    handleSearch,
    handlePageChange,
    handleToolbarClick,
    handleMenuClick,
    openViewDialog,
    handleDelete,
    handleBatchDelete,
  }
}

