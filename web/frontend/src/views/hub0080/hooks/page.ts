/**
 * 告警渠道配置列表页面级 Hook
 * - 组合 useAlertConfigService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { takeNamedObject } from '@/components/form/rs-data'
import type { RsSearchFormExpose } from '@/components/form/rs-search'
import type { RsGridExpose } from '@/components/rs-grid'
import { useAppMessage } from '@/composables/useAppMessage'
import { rsConfirm } from '@/ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { AlertConfig } from '../types'
import { useAlertConfigService } from './service'

/**
 * 告警渠道配置列表页面级 Hook
 * @param gridRef Grid 组件引用（可选）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useAlertConfigPage(
  gridRef?: Ref<RsGridExpose | null>,
  searchFormRef?: Ref<RsSearchFormExpose | null>,
) {
  const message = useAppMessage()
  // 业务服务（包含 model、增删改查等）
  const service = useAlertConfigService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditConfig = ref<AlertConfig | null>(null)

  // ============= 搜索和分页 =============

  /**
   * 处理搜索
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    service.model.resetPagination()
    await service.loadConfigList(searchParams)
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadConfigList()
  }

  // ============= 工具栏按钮处理 =============

  /**
   * 处理工具栏按钮点击
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break

      case 'delete': {
        // 删除当前高亮的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getActiveRow()
        if (!currentRow) {
          message.warning('请先选择或点击要删除的配置')
          return
        }
        await handleDelete(currentRow as AlertConfig)
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
   * 打开新增对话框
   */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditConfig.value = null
    formDialogVisible.value = true
  }

  /**
   * 将API数据转换为表单数据格式（加载时：将 serverConfig 和 sendConfig JSON 字符串解析为嵌套对象）
   */
  const convertToFormData = (config: AlertConfig): any => {
    // 安全解析 JSON 字符串
    const parseJson = (value: any): any => {
      if (typeof value === 'string' && value) {
        try {
          return JSON.parse(value)
        } catch {
          return value
        }
      }
      return value
    }

    // 解析 serverConfig JSON 字符串为对象
    const serverConfigObj = parseJson(config.serverConfig) || {}
    // 解析 sendConfig JSON 字符串为对象
    const sendConfigObj = parseJson(config.sendConfig) || {}

    // 构建表单数据对象
    const formData: any = {
      ...config,
      serverConfig: serverConfigObj && typeof serverConfigObj === 'object' ? { ...serverConfigObj } : {},
      sendConfig: {},
    }

    if (sendConfigObj && typeof sendConfigObj === 'object') {
      Object.keys(sendConfigObj).forEach((key) => {
        const value = sendConfigObj[key]
        if (Array.isArray(value) && (key === 'to' || key === 'cc' || key === 'bcc' || key === 'at_users' || key === 'mentioned_list' || key === 'mentioned_mobile_list')) {
          formData.sendConfig[key] = value.join(', ')
        } else {
          formData.sendConfig[key] = value
        }
      })
    }

    return formData
  }

  /**
   * 打开编辑对话框
   */
  const openEditDialog = async (config: AlertConfig) => {
    if (!config.channelName) {
      message.warning('渠道名称不能为空')
      return
    }

    try {
      // 获取最新数据
      const latestConfig = await service.getConfigDetail(config.channelName)
      if (latestConfig) {
        formDialogMode.value = 'edit'
        currentEditConfig.value = latestConfig
        // 转换为表单数据格式
        const formData = convertToFormData(latestConfig)
        currentEditConfig.value = formData as any
        formDialogVisible.value = true
      } else {
        // 降级：使用传入的数据
        formDialogMode.value = 'edit'
        currentEditConfig.value = convertToFormData(config) as any
        formDialogVisible.value = true
      }
    } catch (error) {
      // 降级：使用传入的数据
      formDialogMode.value = 'edit'
      currentEditConfig.value = convertToFormData(config) as any
      formDialogVisible.value = true
    }
  }

  /**
   * 打开查看对话框
   */
  const openViewDialog = async (config: AlertConfig) => {
    if (!config.channelName) {
      message.warning('渠道名称不能为空')
      return
    }

    try {
      // 获取最新数据
      const latestConfig = await service.getConfigDetail(config.channelName)
      if (latestConfig) {
        formDialogMode.value = 'view'
        // 转换为表单数据格式
        const formData = convertToFormData(latestConfig)
        currentEditConfig.value = formData as any
        formDialogVisible.value = true
      } else {
        // 降级：使用传入的数据
        formDialogMode.value = 'view'
        currentEditConfig.value = convertToFormData(config) as any
        formDialogVisible.value = true
      }
    } catch (error) {
      // 降级：使用传入的数据
      formDialogMode.value = 'view'
      currentEditConfig.value = convertToFormData(config) as any
      formDialogVisible.value = true
    }
  }

  /**
   * 打开复制对话框（基于现有配置创建新配置）
   */
  const openCopyDialog = async (config: AlertConfig) => {
    if (!config.channelName) {
      message.warning('渠道名称不能为空')
      return
    }

    try {
      // 获取最新数据
      const latestConfig = await service.getConfigDetail(config.channelName)
      const sourceConfig = latestConfig || config

      // 转换为表单数据格式
      const formData = convertToFormData(sourceConfig)

      // 清空主键和系统字段，准备创建新配置
      formData.channelName = '' // 清空渠道名称，用户必须输入新名称
      formData.tenantId = formData.tenantId || '' // 保留租户ID（如果需要）
      
      // 清空系统字段（这些字段会在创建时自动生成）
      delete formData.addTime
      delete formData.addWho
      delete formData.editTime
      delete formData.editWho
      delete formData.oprSeqFlag
      delete formData.currentVersion
      
      // 重置统计字段
      formData.totalSentCount = 0
      formData.successCount = 0
      formData.failureCount = 0
      formData.lastSendTime = null
      formData.lastSuccessTime = null
      formData.lastFailureTime = null
      formData.lastErrorMessage = null

      // 设置为创建模式
      formDialogMode.value = 'create'
      currentEditConfig.value = formData as any
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取配置详情失败，无法复制')
      console.error('复制配置失败:', error)
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
   * 将表单数据转换为API数据格式（保存时：将 serverConfig / sendConfig 嵌套对象合并为 JSON 字符串）
   */
  const convertToApiData = (formData: Record<string, any>): Partial<AlertConfig> => {
    // 从嵌套对象重新构建 serverConfig / sendConfig
    const serverConfigObj = takeNamedObject(formData, 'serverConfig')
    const sendConfigObj = takeNamedObject(formData, 'sendConfig')
    const listKeys = new Set(['to', 'cc', 'bcc', 'at_users', 'mentioned_list', 'mentioned_mobile_list'])
    Object.keys(serverConfigObj).forEach((key) => {
      const value = serverConfigObj[key]
      if (value === undefined || value === null || value === '') delete serverConfigObj[key]
    })
    Object.keys(sendConfigObj).forEach((key) => {
      const value = sendConfigObj[key]
      if (value === undefined || value === null || value === '') {
        delete sendConfigObj[key]
        return
      }
      if (typeof value === 'string' && listKeys.has(key)) {
        const arr = value.split(',').map((s) => s.trim()).filter((s) => s)
        if (arr.length > 0) sendConfigObj[key] = arr
        else delete sendConfigObj[key]
      }
    })

    const apiData: Record<string, any> = {}
    Object.keys(formData).forEach((key) => {
      if (
        key === 'serverConfig' ||
        key === 'sendConfig' ||
        key.startsWith('serverConfig.') ||
        key.startsWith('sendConfig.')
      ) {
        return
      }
      apiData[key] = formData[key]
    })

    return {
      ...apiData,
      serverConfig: Object.keys(serverConfigObj).length > 0 ? JSON.stringify(serverConfigObj) : undefined,
      sendConfig: Object.keys(sendConfigObj).length > 0 ? JSON.stringify(sendConfigObj) : undefined,
    } as Partial<AlertConfig>
  }

  /**
   * 处理表单提交
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      closeFormDialog()
      return
    }

    try {
      // 将表单数据转换为API数据格式
      const submitData = convertToApiData(formData)

      let success = false
      if (formDialogMode.value === 'create') {
        success = await service.addConfig(submitData)
      } else if (formDialogMode.value === 'edit' && currentEditConfig.value) {
        success = await service.editConfig(
          currentEditConfig.value.channelName,
          submitData
        )
      }

      if (success) {
        closeFormDialog()
      }
    } catch (error: any) {
      console.error('提交配置失败:', error)
      message.error(error.message || '提交失败，请重试')
    }
  }

  // ============= 右键菜单处理 =============

  /**
   * 处理右键菜单点击
   */
  const handleMenuClick = async ({ key, row }: { key: string; row?: AlertConfig }) => {
    if (!row) return
    switch (key) {
      case 'view':
        await openViewDialog(row)
        break
      case 'edit':
        await openEditDialog(row)
        break
      case 'copy':
        await openCopyDialog(row)
        break
      case 'reload':
        await handleReloadChannel(row)
        break
      case 'setDefault':
        await handleSetDefault(row)
        break
      case 'test':
        await handleTestChannel(row)
        break
      case 'delete':
        await handleDelete(row)
        break
      default:
        console.warn('未知的菜单项:', key)
    }
  }

  // ============= 单个操作处理 =============

  /**
   * 处理删除
   */
  const handleDelete = async (config: AlertConfig) => {
    const confirmed = await rsConfirm.warning({
      title: '确认删除',
      description: `确定要删除渠道配置"${config.channelName}"吗？此操作不可恢复。`,
      confirmText: '删除',
      cancelText: '取消',
    })

    if (!confirmed) {
      return
    }

    // 注意：后端暂未提供删除接口，这里先提示
    message.warning('删除功能暂未实现')
  }

  /**
   * 处理设置默认渠道
   */
  const handleSetDefault = async (config: AlertConfig) => {
    if (config.defaultFlag === 'Y') {
      message.info('该渠道已经是默认渠道')
      return
    }

    const confirmed = await rsConfirm.warning({
      title: '确认设置',
      description: `确定要将渠道"${config.channelName}"设置为默认渠道吗？`,
      confirmText: '确定',
      cancelText: '取消',
    })

    if (!confirmed) {
      return
    }

    await service.setDefault(config.channelName)
  }

  /**
   * 处理切换状态
   */
  const handleToggleStatus = async (config: AlertConfig) => {
    await service.toggleConfigStatus(config)
  }

  // 测试弹窗状态
  const testModalVisible = ref(false)
  const currentTestConfig = ref<AlertConfig | null>(null)

  /**
   * 处理测试渠道
   */
  const handleTestChannel = async (config: AlertConfig) => {
    if (!config.channelName) {
      message.warning('渠道名称不能为空')
      return
    }

    if (config.activeFlag !== 'Y') {
      message.warning('该渠道未启用，无法测试')
      return
    }

    // 打开测试弹窗
    currentTestConfig.value = config
    testModalVisible.value = true
  }

  /**
   * 处理重载渠道配置
   */
  const handleReloadChannel = async (config: AlertConfig) => {
    if (!config.channelName) {
      message.warning('渠道名称不能为空')
      return
    }

    const confirmed = await rsConfirm.warning({
      title: '确认重载',
      description: `确定要重载渠道"${config.channelName}"的配置吗？重载后将立即生效。`,
      confirmText: '重载',
      cancelText: '取消',
    })
    if (!confirmed) return

    const ok = await service.reloadChannel(config.channelName)
    if (ok) {
      // 轻量刷新列表，确保状态展示一致
      await service.loadConfigList()
    }
  }

  /**
   * 关闭测试弹窗
   */
  const closeTestModal = () => {
    testModalVisible.value = false
    currentTestConfig.value = null
  }

  return {
    // 服务
    service,

    // 对话框状态
    formDialogVisible,
    formDialogMode,
    currentEditConfig,

    // 测试弹窗状态
    testModalVisible,
    currentTestConfig,

    // 方法
    handleSearch,
    handlePageChange,
    handleToolbarClick,
    handleMenuClick,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    openCopyDialog,
    closeFormDialog,
    handleFormSubmit,
    handleDelete,
    handleSetDefault,
    handleToggleStatus,
    handleTestChannel,
    handleReloadChannel,
    closeTestModal,
  }
}

