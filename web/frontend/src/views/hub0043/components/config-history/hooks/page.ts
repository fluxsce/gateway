/**
 * Hub0043 配置历史管理页面级 Hook
 * - 组合 useConfigHistoryService（纯业务逻辑）
 * - 处理详情对话框、回滚对话框等页面交互
 */

import { useGDialog } from '@/components/gdialog'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { nextTick, ref, watch } from 'vue'
import type { ConfigHistory, RollbackRequest } from '../../../types'
import { useConfigHistoryService } from './service'

/**
 * 配置历史管理页面级 Hook
 * @param searchFormRef 搜索表单引用
 */
export function useConfigHistoryPage(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 业务服务（包含 model、查询等）
  const service = useConfigHistoryService(searchFormRef)

  // 视图状态（'list' | 'detail'）
  const currentView = ref<'list' | 'detail'>('list')
  const currentHistoryDetail = ref<ConfigHistory | null>(null)
  
  // 保存的搜索表单数据（用于返回列表时恢复）
  const savedSearchFormData = ref<Record<string, any> | null>(null)

  // 回滚对话框状态
  const rollbackDialogVisible = ref(false)
  const currentRollbackHistory = ref<ConfigHistory | null>(null)
  const submitting = ref(false)

  /**
   * 打开详情视图
   */
  const openDetailView = async (history: ConfigHistory) => {
    // 保存搜索表单数据
    const formData = searchFormRef?.value?.getFormData?.() || {}
    savedSearchFormData.value = { ...formData }
    
    const detail = await service.getHistoryDetail(String(history.configHistoryId))
    if (detail) {
      currentHistoryDetail.value = detail
      currentView.value = 'detail'
    } else {
      // 如果加载失败，保持在列表视图
      currentHistoryDetail.value = null
    }
  }

  /**
   * 返回列表视图
   */
  const handleBackToList = () => {
    currentView.value = 'list'
    currentHistoryDetail.value = null
  }
  
  // 监听视图切换，恢复搜索表单数据
  watch(
    () => currentView.value,
    (newView) => {
      if (newView === 'list' && savedSearchFormData.value) {
        // 使用双重 nextTick 确保搜索表单组件已完全渲染
        nextTick(() => {
          nextTick(() => {
            if (searchFormRef?.value?.setFormData) {
              searchFormRef.value.setFormData(savedSearchFormData.value!)
            }
          })
        })
      }
    }
  )

  /**
   * 打开回滚对话框
   */
  const openRollbackDialog = (history: ConfigHistory) => {
    currentRollbackHistory.value = history
    rollbackDialogVisible.value = true
  }

  /**
   * 关闭回滚对话框
   */
  const closeRollbackDialog = () => {
    rollbackDialogVisible.value = false
    currentRollbackHistory.value = null
  }

  /**
   * 确认回滚
   */
  const handleRollbackConfirm = async (rollbackData: RollbackRequest) => {
    if (!currentRollbackHistory.value) {
      return
    }

    const confirmed = await gDialog.warning({
      title: '确认回滚',
      subtitle: `确定要将配置回滚到版本 ${currentRollbackHistory.value.newVersion || currentRollbackHistory.value.configVersion} 吗？`,
      content: '此操作会创建新的配置版本',
      positiveText: '确定',
      negativeText: '取消',
      width: 500,
    })

    if (!confirmed) {
      return
    }

    submitting.value = true
    try {
      const success = await service.rollback(rollbackData)
      if (success) {
        closeRollbackDialog()
        await service.loadHistory()
      }
    } finally {
      submitting.value = false
    }
  }

  /**
   * 搜索处理
   */
  const handleSearch = () => {
    service.loadHistory()
  }

  /**
   * 工具栏按钮点击
   */
  const handleToolbarClick = (key: string) => {
    if (key === 'back') {
      // 返回配置列表事件由父组件处理
      // 这里可以通过emit或者回调函数来处理
    }
  }

  /**
   * 分页变化
   */
  const handlePageChange = ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    // 历史记录不使用分页
  }

  /**
   * 表格右键菜单点击
   */
  const handleMenuClick = async (params: { code: string; row?: any }) => {
    if (!params.row) {
      return
    }
    const row = params.row as ConfigHistory

    switch (params.code) {
      case 'view':
        await openDetailView(row)
        break
      case 'rollback':
        openRollbackDialog(row)
        break
    }
  }

  return {
    // 服务（包含 model 和所有业务方法）
    service,

    // 视图状态
    currentView,
    currentHistoryDetail,
    rollbackDialogVisible,
    currentRollbackHistory,
    submitting,

    // 视图方法
    openDetailView,
    handleBackToList,
    openRollbackDialog,
    closeRollbackDialog,
    handleRollbackConfirm,

    // 工具栏和菜单
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
    handlePageChange,
  }
}

export type ConfigHistoryPage = ReturnType<typeof useConfigHistoryPage>

