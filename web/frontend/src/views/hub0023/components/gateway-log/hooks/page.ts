/**
 * Hub0023 网关日志管理页面级 Hook
 * - 组合 useGatewayLogService（纯业务逻辑）
 * - 处理详情对话框、工具栏、右键菜单等页面交互
 */

import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref, watch } from 'vue'
import type { GatewayLogListItem } from '../../../types'
import { useGatewayLogService } from './service'

/** 单次批量重发前端条数上限（超出部分不进入弹窗，避免 DOM 与请求压力过大） */
export const MAX_GATEWAY_LOG_RESEND_BATCH = 100

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
  /** 与详情查询一并传入后端的网关实例 ID（与列表行一致） */
  const selectedGatewayInstanceId = ref('')

  /** 请求重发弹窗：左侧 Trace 列表，右侧 GRestfulApi */
  const resendDialogVisible = ref(false)
  const resendLogs = ref<GatewayLogListItem[]>([])

  /** 打开查看详情对话框 */
  const openViewDialog = (log: GatewayLogListItem) => {
    selectedTraceId.value = log.traceId
    selectedGatewayInstanceId.value = String(log.gatewayInstanceId ?? '').trim()
    detailDialogVisible.value = true
  }

  /** 关闭详情对话框 */
  const closeDetailDialog = () => {
    detailDialogVisible.value = false
    selectedTraceId.value = ''
    selectedGatewayInstanceId.value = ''
  }

  /** 打开重发弹窗（传入列表行，至少含 traceId） */
  const openResendDialog = (logs: GatewayLogListItem[]) => {
    if (logs.length === 0) {
      message.warning('请选择要重发的日志')
      return
    }
    const total = logs.length
    const capped =
      total > MAX_GATEWAY_LOG_RESEND_BATCH ? logs.slice(0, MAX_GATEWAY_LOG_RESEND_BATCH) : [...logs]
    if (total > MAX_GATEWAY_LOG_RESEND_BATCH) {
      message.warning(
        `单次最多重发 ${MAX_GATEWAY_LOG_RESEND_BATCH} 条，当前已选 ${total} 条，仅前 ${MAX_GATEWAY_LOG_RESEND_BATCH} 条会进入重发`
      )
    }
    resendLogs.value = capped
    resendDialogVisible.value = true
  }

  const closeResendDialog = () => {
    resendDialogVisible.value = false
    resendLogs.value = []
  }

  watch(resendDialogVisible, (v) => {
    if (!v) {
      resendLogs.value = []
    }
  })


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
        // 查看：优先勾选行，无勾选时回退到当前高亮行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const selectedRow = gridRef.value.getSelectedOrCurrentRecord()
        if (!selectedRow) {
          message.warning('请先选择或点击要查看的日志')
          return
        }
        openViewDialog(selectedRow as GatewayLogListItem)
        break
      }

      case 'batchReset': {
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const selectedRows = gridRef.value.getCheckboxRecords() as GatewayLogListItem[]
        openResendDialog(selectedRows)
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
        openResendDialog([row])
        break
    }
  }

  return {
    // 业务服务（包含 model 与查询）
    service,

    // 详情对话框
    detailDialogVisible,
    selectedTraceId,
    selectedGatewayInstanceId,
    openViewDialog,
    closeDetailDialog,

    // 重发弹窗
    resendDialogVisible,
    resendLogs,
    openResendDialog,
    closeResendDialog,

    // 事件处理器
    handleToolbarClick,
    handleMenuClick,
    handleSearch
  }
}

export type GatewayLogPage = ReturnType<typeof useGatewayLogPage>

