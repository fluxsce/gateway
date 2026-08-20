/**
 * 预警日志页面级 Hook
 * - 组合 useAlertLogService（纯业务逻辑）
 * - 处理查看对话框、工具栏、右键菜单等页面交互
 */

import type { RsSearchFormExpose } from '@/components/form/rs-search'
import type { RsGridExpose } from '@/components/rs-grid'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { rsConfirm } from '@/ui'
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
  gridRef?: Ref<RsGridExpose | null>,
  searchFormRef?: Ref<RsSearchFormExpose | null>,
) {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0082')
  const service = useAlertLogService(searchFormRef)

  const viewDialogVisible = ref(false)
  const selectedAlertLogId = ref<string>('')

  /**
   * 处理搜索
   * @param searchParams - 查询条件；不传则从表单读取
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    service.model.resetPagination()
    await service.loadLogList(searchParams)
  }

  /**
   * 处理分页变化
   * @param currentPage - 当前页码
   * @param pageSize - 每页条数
   */
  const handlePageChange = async ({
    currentPage,
    pageSize,
  }: {
    currentPage: number
    pageSize: number
  }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadLogList()
  }

  /**
   * 处理工具栏按钮点击
   * @param key - 按钮 key
   * @param formData - 表单数据（search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'delete': {
        if (!gridRef?.value) {
          message.warning(t('message.gridRefMissing'))
          return
        }
        const selectedRows = (gridRef.value.getActiveRows?.() || []) as AlertLog[]
        if (selectedRows.length === 0) {
          message.warning(t('message.selectToDelete'))
          return
        }
        const alertLogIds = selectedRows.map((row) => row.alertLogId)
        await handleBatchDelete(alertLogIds)
        break
      }

      case 'search': {
        await handleSearch(formData)
        break
      }

      default:
        break
    }
  }

  /**
   * 打开查看对话框
   * @param log - 当前行日志
   */
  const openViewDialog = async (log: AlertLog) => {
    if (!log.alertLogId) {
      message.warning(t('message.alertLogIdRequired'))
      return
    }

    selectedAlertLogId.value = log.alertLogId
    viewDialogVisible.value = true
  }

  /**
   * 处理右键菜单点击
   * @param key - 菜单项 key
   * @param row - 当前行
   */
  const handleMenuClick = async ({ key, row }: { key: string; row?: AlertLog }) => {
    if (!row) return
    switch (key) {
      case 'view':
        await openViewDialog(row)
        break

      case 'delete':
        await handleDelete(row)
        break

      default:
        break
    }
  }

  /**
   * 处理单条删除
   * @param log - 待删除日志
   */
  const handleDelete = async (log: AlertLog) => {
    if (!log.alertLogId) {
      message.warning(t('message.alertLogIdRequired'))
      return
    }

    const confirmed = await rsConfirm.warning({
      title: t('confirm.deleteTitle'),
      description: t('confirm.deleteContent', { id: log.alertLogId }),
      confirmText: t('confirm.confirmText'),
      cancelText: t('confirm.cancelText'),
    })
    if (!confirmed) return

    await service.deleteLog(log.alertLogId)
  }

  /**
   * 处理批量删除
   * @param alertLogIds - 日志 ID 数组
   */
  const handleBatchDelete = async (alertLogIds: string[]) => {
    if (alertLogIds.length === 0) {
      message.warning(t('message.selectToDelete'))
      return
    }

    const confirmed = await rsConfirm.warning({
      title: t('confirm.batchDeleteTitle'),
      description: t('confirm.batchDeleteContent', { count: alertLogIds.length }),
      confirmText: t('confirm.confirmText'),
      cancelText: t('confirm.cancelText'),
    })
    if (!confirmed) return

    await service.batchDeleteLogs(alertLogIds)
  }

  return {
    service,
    viewDialogVisible,
    selectedAlertLogId,
    handleSearch,
    handlePageChange,
    handleToolbarClick,
    handleMenuClick,
    openViewDialog,
    handleDelete,
    handleBatchDelete,
  }
}
