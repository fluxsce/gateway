/**
 * 告警渠道配置列表页面级 Hook
 * - 组合 useAlertConfigService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useGDialog } from '@/components/gdialog'
import { useMessage } from 'naive-ui'
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
  gridRef?: Ref<any> | any,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()
  const gDialog = useGDialog()

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
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要删除的配置')
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
   * 将API数据转换为表单数据格式（加载时：将 serverConfig 和 sendConfig JSON 字符串解析为点号分隔字段）
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
    }

    // 将 serverConfig 对象的嵌套属性展开为点号分隔的字段
    if (serverConfigObj && typeof serverConfigObj === 'object') {
      Object.keys(serverConfigObj).forEach((key) => {
        formData[`serverConfig.${key}`] = serverConfigObj[key]
      })
    }

    // 将 sendConfig 对象的嵌套属性展开为点号分隔的字段
    if (sendConfigObj && typeof sendConfigObj === 'object') {
      Object.keys(sendConfigObj).forEach((key) => {
        const value = sendConfigObj[key]
        // 处理数组字段（如 to, cc, bcc 等），转换为逗号分隔的字符串
        if (Array.isArray(value) && (key === 'to' || key === 'cc' || key === 'bcc' || key === 'at_users' || key === 'mentioned_list' || key === 'mentioned_mobile_list')) {
          formData[`sendConfig.${key}`] = value.join(', ')
        } else {
          formData[`sendConfig.${key}`] = value
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
   * 将表单数据转换为API数据格式（保存时：将 serverConfig.xxx 和 sendConfig.xxx 字段合并为 JSON 字符串）
   */
  const convertToApiData = (formData: Record<string, any>): Partial<AlertConfig> => {
    // 从点号分隔字段重新构建 serverConfig 对象
    const serverConfigObj: Record<string, any> = {}
    Object.keys(formData).forEach((key) => {
      if (key.startsWith('serverConfig.')) {
        const subKey = key.replace('serverConfig.', '')
        const value = formData[key]
        // 只添加非空值
        if (value !== undefined && value !== null && value !== '') {
          serverConfigObj[subKey] = value
        }
      }
    })

    // 从点号分隔字段重新构建 sendConfig 对象
    const sendConfigObj: Record<string, any> = {}
    Object.keys(formData).forEach((key) => {
      if (key.startsWith('sendConfig.')) {
        const subKey = key.replace('sendConfig.', '')
        const value = formData[key]
        // 只添加非空值
        if (value !== undefined && value !== null && value !== '') {
          // 处理字符串数组（如 to, cc, bcc 等用逗号分隔的字符串）
          if (typeof value === 'string' && (subKey === 'to' || subKey === 'cc' || subKey === 'bcc' || subKey === 'at_users' || subKey === 'mentioned_list' || subKey === 'mentioned_mobile_list')) {
            // 将逗号分隔的字符串转换为数组
            const arr = value.split(',').map(s => s.trim()).filter(s => s)
            if (arr.length > 0) {
              sendConfigObj[subKey] = arr
            }
          } else {
            sendConfigObj[subKey] = value
          }
        }
      }
    })

    // 构建 API 数据对象，排除点号分隔字段
    const apiData: Record<string, any> = {}
    Object.keys(formData).forEach((key) => {
      // 排除以 serverConfig. 和 sendConfig. 开头的点号分隔字段
      if (!key.startsWith('serverConfig.') && !key.startsWith('sendConfig.')) {
        apiData[key] = formData[key]
      }
    })

    return {
      ...apiData,
      // serverConfig 前端是点号分隔字段，后端需要 JSON 字符串
      serverConfig: Object.keys(serverConfigObj).length > 0
        ? JSON.stringify(serverConfigObj)
        : undefined,
      // sendConfig 前端是点号分隔字段，后端需要 JSON 字符串
      sendConfig: Object.keys(sendConfigObj).length > 0
        ? JSON.stringify(sendConfigObj)
        : undefined,
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
  const handleMenuClick = async ({ menu, row }: { menu: any; row: AlertConfig }) => {
    switch (menu.code) {
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
        console.warn('未知的菜单项:', menu.code)
    }
  }

  // ============= 单个操作处理 =============

  /**
   * 处理删除
   */
  const handleDelete = async (config: AlertConfig) => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      content: `确定要删除渠道配置"${config.channelName}"吗？此操作不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
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

    const confirmed = await gDialog.warning({
      title: '确认设置',
      content: `确定要将渠道"${config.channelName}"设置为默认渠道吗？`,
      positiveText: '确定',
      negativeText: '取消',
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

    const confirmed = await gDialog.warning({
      title: '确认重载',
      content: `确定要重载渠道"${config.channelName}"的配置吗？重载后将立即生效。`,
      positiveText: '重载',
      negativeText: '取消',
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

