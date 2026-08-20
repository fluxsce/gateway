/**
 * 审计日志页面级 Hook
 * - 组合 useAuditLogService
 * - 处理查看对话框、搜索、分页、右键菜单
 */

import type { RsSearchFormExpose } from '@/components/form/rs-search'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { Ref } from 'vue'
import { nextTick, onMounted, ref } from 'vue'
import type { AuthAuditLog } from '../types'
import { useAuditLogService } from './service'

/**
 * 审计日志页面级 Hook
 * @param searchFormRef - 搜索表单引用（可选）
 */
export function useAuditLogPage(searchFormRef?: Ref<RsSearchFormExpose | null>) {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0004')
  const service = useAuditLogService(searchFormRef)

  const viewDialogVisible = ref(false)
  const selectedAuditId = ref('')
  const exportVisible = ref(false)
  const exportParams = ref<Record<string, any>>({})

  /**
   * 处理搜索。
   * @param searchParams - 查询条件；不传则从表单读取
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    service.model.resetPagination()
    await service.loadLogList(searchParams)
  }

  /**
   * 处理分页变化。
   * @param currentPage - 当前页码
   * @param pageSize - 每页条数
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadLogList()
  }

  /**
   * 打开查看对话框。
   * @param log - 当前行日志
   */
  const openViewDialog = (log: AuthAuditLog) => {
    if (!log.auditId) {
      message.warning(t('message.auditIdRequired'))
      return
    }
    selectedAuditId.value = log.auditId
    viewDialogVisible.value = true
  }

  /**
   * 处理右键菜单点击。
   * @param key - 菜单项 key
   * @param row - 当前行
   */
  const handleMenuClick = ({ key, row }: { key: string; row?: AuthAuditLog }) => {
    if (!row) return
    if (key === 'view') {
      openViewDialog(row)
    }
  }

  /**
   * 处理工具栏点击。
   * @param key - 按钮 key
   */
  const handleToolbarClick = (key: string) => {
    if (key === 'export') {
      exportParams.value = service.buildExportParams()
      exportVisible.value = true
    }
  }

  onMounted(() => {
    void nextTick().then(() => handleSearch())
  })

  return {
    service,
    viewDialogVisible,
    selectedAuditId,
    handleSearch,
    handlePageChange,
    handleMenuClick,
    handleToolbarClick,
    openViewDialog,
    exportVisible,
    exportParams,
  }
}
