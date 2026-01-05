/**
 * Hub0023 网关日志管理页面级 Hook
 * - 组合 useGatewayLogService（纯业务逻辑）
 * - 处理详情对话框、工具栏、右键菜单等页面交互
 */

import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { GatewayLogListItem } from '../../../types'
import { useGatewayLogService } from './service'

/**
 * 网关日志管理页面级 Hook
 */
export function useGatewayLogPage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 业务服务（包含 model、查询等）
  const service = useGatewayLogService(searchFormRef)

  // 详情对话框状态
  const detailDialogVisible = ref(false)
  const selectedTraceId = ref('')

  /** 打开查看详情对话框 */
  const openViewDialog = (log: GatewayLogListItem) => {
    selectedTraceId.value = log.traceId
    detailDialogVisible.value = true
  }

  /** 关闭详情对话框 */
  const closeDetailDialog = () => {
    detailDialogVisible.value = false
    selectedTraceId.value = ''
  }


  /**
   * 处理搜索（接收 SearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  /**
   * 工具栏按钮点击处理
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'view': {
        // 查看当前高亮的行（点击选中的行）
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要查看的日志')
          return
        }
        openViewDialog(currentRow as GatewayLogListItem)
        break
      }

      case 'batchReset': {
        // 批量重置选中的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const selectedRows = gridRef.value.getCheckboxRecords() as GatewayLogListItem[]
        if (selectedRows.length === 0) {
          message.warning('请选择要重置的日志')
          return
        }
        await service.resetGatewayLogs(selectedRows, '批量重置', 'current_user')
        break
      }

      case 'export': {
        // 导出日志
        await service.exportGatewayLogs(formData)
        break
      }

      case 'search': {
        // search 操作由 @search 事件处理，这里不需要重复处理
        break
      }
    }
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = async ({ code, row }: { code: string; row?: GatewayLogListItem }) => {
    if (!row) return

    switch (code) {
      case 'view':
        openViewDialog(row)
        break

      case 'reset':
        await service.resetGatewayLogs([row], '手动重置', 'current_user')
        break
    }
  }

  return {
    // 业务服务（包含 model 与查询）
    service,

    // 详情对话框
    detailDialogVisible,
    selectedTraceId,
    openViewDialog,
    closeDetailDialog,

    // 事件处理器
    handleToolbarClick,
    handleMenuClick,
    handleSearch
  }
}

export type GatewayLogPage = ReturnType<typeof useGatewayLogPage>

