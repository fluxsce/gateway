/**
 * 网关实例管理页面级 Hook
 * - 组合 useGatewayInstanceService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { rsConfirm } from '@/ui'
import { useAppMessage } from '@/composables/useAppMessage'
import { flattenExtProperty, getApiMessage, isApiSuccess, parseJsonData, parseJsonObjectColumn, stringifyJsonObjectColumn, unflattenExtProperty } from '@/utils/format'
import { consumeTextFileField, filesFromTextContent } from '@/utils/uploadFile'
import { PlayCircleOutline, RefreshOutline, StopCircleOutline } from '@vicons/ionicons5'
import type { Ref } from 'vue'
import { ref } from 'vue'
import * as gatewayApi from '../api'
import type { GatewayInstance } from '../types'
import { useGatewayInstanceService } from './useGatewayInstanceService'

/**
 * 网关实例管理页面级 Hook
 */
export function useGatewayInstancePage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useAppMessage()
// 业务服务（包含 model、增删改查等）
  const service = useGatewayInstanceService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditInstance = ref<GatewayInstance | null>(null)
  const submitting = ref(false)

  // 配置管理对话框状态
  const configDialogVisible = ref(false)
  const currentConfigInstanceId = ref<string>('')

  // 日志配置对话框状态
  const logConfigDialogVisible = ref(false)
  const logConfigDialogMode = ref<'edit' | 'view'>('edit')
  const currentLogConfig = ref<Record<string, any> | null>(null)
  const logConfigSubmitting = ref(false)

  // IP访问控制对话框状态
  const ipAccessControlDialogVisible = ref(false)
  const ipAccessControlSecurityConfigId = ref<string>('')

  /**
   * 打开IP访问控制对话框
   */
  const openIpAccessControlDialog = async (instance: GatewayInstance) => {
    // 使用 gatewayInstanceId 作为 securityConfigId
    // 因为 IP 访问配置的 securityConfigId 字段可以关联到网关实例ID
    ipAccessControlSecurityConfigId.value = instance.gatewayInstanceId
    ipAccessControlDialogVisible.value = true
  }

  // User-Agent访问控制对话框状态
  const userAgentAccessControlDialogVisible = ref(false)
  const userAgentAccessControlSecurityConfigId = ref<string>('')

  /**
   * 打开User-Agent访问控制对话框
   */
  const openUserAgentAccessControlDialog = async (instance: GatewayInstance) => {
    // 使用 gatewayInstanceId 作为 securityConfigId
    // 因为 User-Agent 访问配置的 securityConfigId 字段可以关联到网关实例ID
    userAgentAccessControlSecurityConfigId.value = instance.gatewayInstanceId
    userAgentAccessControlDialogVisible.value = true
  }

  // API访问控制对话框状态
  const apiAccessControlDialogVisible = ref(false)
  const apiAccessControlSecurityConfigId = ref<string>('')

  /**
   * 打开API访问控制对话框
   */
  const openApiAccessControlDialog = async (instance: GatewayInstance) => {
    // 使用 gatewayInstanceId 作为 securityConfigId
    // 因为 API 访问配置的 securityConfigId 字段可以关联到网关实例ID
    apiAccessControlSecurityConfigId.value = instance.gatewayInstanceId
    apiAccessControlDialogVisible.value = true
  }

  // 域名访问控制对话框状态
  const domainAccessControlDialogVisible = ref(false)
  const domainAccessControlSecurityConfigId = ref<string>('')

  /**
   * 打开域名访问控制对话框
   */
  const openDomainAccessControlDialog = async (instance: GatewayInstance) => {
    // 使用 gatewayInstanceId 作为 securityConfigId
    // 因为域名访问配置的 securityConfigId 字段可以关联到网关实例ID
    domainAccessControlSecurityConfigId.value = instance.gatewayInstanceId
    domainAccessControlDialogVisible.value = true
  }

  // 跨域配置对话框状态
  const corsConfigDialogVisible = ref(false)
  const corsConfigGatewayInstanceId = ref<string>('')

  /**
   * 打开跨域配置对话框
   */
  const openCorsConfigDialog = async (instance: GatewayInstance) => {
    corsConfigGatewayInstanceId.value = instance.gatewayInstanceId
    corsConfigDialogVisible.value = true
  }

  // 认证配置对话框状态
  const authConfigDialogVisible = ref(false)
  const authConfigGatewayInstanceId = ref<string>('')

  /**
   * 打开认证配置对话框
   */
  const openAuthConfigDialog = async (instance: GatewayInstance) => {
    authConfigGatewayInstanceId.value = instance.gatewayInstanceId
    authConfigDialogVisible.value = true
  }

  // 限流配置对话框状态
  const rateLimitConfigDialogVisible = ref(false)
  const rateLimitConfigGatewayInstanceId = ref<string>('')

  // 导出/导入
  const exportVisible = ref(false)
  const exportInstanceId = ref<string>('')
  const importVisible = ref(false)

  /**
   * 打开限流配置对话框
   */
  const openRateLimitConfigDialog = async (instance: GatewayInstance) => {
    rateLimitConfigGatewayInstanceId.value = instance.gatewayInstanceId
    rateLimitConfigDialogVisible.value = true
  }

  /**
   * 打开新增实例对话框
   */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditInstance.value = null
    formDialogVisible.value = true
  }

  /**
   * 打开编辑实例对话框
   */
  const openEditDialog = async (instance: GatewayInstance) => {
    try {
      // 获取完整详情
      const detailInstance = await service.getInstanceDetail(
        instance.gatewayInstanceId,
        instance.tenantId
      )
      
      if (!detailInstance) {
        message.error('获取实例详情失败')
        return
      }

      // 数据库 certContent/keyContent → File[]，供 RsUpload 回显
      const formData: any = { ...detailInstance }
      formData.certFileList = filesFromTextContent(
        detailInstance.certFilePath || 'certificate.pem',
        detailInstance.certContent || '',
      )
      formData.keyFileList = filesFromTextContent(
        detailInstance.keyFilePath || 'private-key.pem',
        detailInstance.keyContent || '',
      )

      formDialogMode.value = 'edit'
      currentEditInstance.value = formData
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取实例详情失败')
    }
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditInstance.value = null
  }
  
  /**
   * 打开查看详情对话框
   * 从后端获取最新数据，确保显示的是最新状态
   */
  const openViewDialog = async (instance: GatewayInstance) => {
    try {
      // 从后端获取最新数据，确保显示的是最新状态
      const detailInstance = await service.getInstanceDetail(
        instance.gatewayInstanceId,
        instance.tenantId
      )
      
      if (!detailInstance) {
        message.error('获取实例详情失败')
        return
      }

      const formData: any = { ...detailInstance }
      formData.certFileList = filesFromTextContent(
        detailInstance.certFilePath || 'certificate.pem',
        detailInstance.certContent || '',
      )
      formData.keyFileList = filesFromTextContent(
        detailInstance.keyFilePath || 'private-key.pem',
        detailInstance.keyContent || '',
      )

      formDialogMode.value = 'view'
      currentEditInstance.value = formData
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取实例详情失败')
    }
  }

  /**
   * 处理搜索（接收 RsSearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  /**
   * 提交表单（新增/编辑共用，由 RsDataFormModal 收集表单数据后回调）
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      return
    }

    submitting.value = true
    try {
      const processedData = { ...formData }
      const editFallback = formDialogMode.value === 'edit' ? currentEditInstance.value : null

      // File[] → certContent / keyContent（与库字段一致）；临时列表不提交
      await consumeTextFileField(processedData, 'certFileList', 'certContent', 'certFilePath', {
        fallbackPath: editFallback?.certFilePath,
      })
      await consumeTextFileField(processedData, 'keyFileList', 'keyContent', 'keyFilePath', {
        fallbackPath: editFallback?.keyFilePath,
      })

      if (formDialogMode.value === 'create') {
        // 新增模式
        const success = await service.addInstance(processedData as GatewayInstance)
        if (success) {
          closeFormDialog()
        }
      } else if (formDialogMode.value === 'edit') {
        // 编辑模式
        if (!currentEditInstance.value) return
        const success = await service.editInstance({
          ...processedData,
          gatewayInstanceId: currentEditInstance.value.gatewayInstanceId,
        } as Partial<GatewayInstance> & { gatewayInstanceId: string })
        if (success) {
          closeFormDialog()
        }
      }
    } catch (error) {
      message.error('读取证书/私钥文件失败')
    } finally {
      submitting.value = false
    }
  }

  /**
   * 工具栏按钮点击处理
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'add':
        // 直接打开新增对话框
        openAddDialog()
        break

      case 'edit': {
        // 优先复选框勾选行，否则当前高亮行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getSelectedOrCurrentRecord()
        if (!currentRow) {
          message.warning('请先勾选或点击选择要编辑的实例')
          return
        }
        await openEditDialog(currentRow as GatewayInstance)
        break
      }

      case 'delete': {
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getSelectedOrCurrentRecord()
        if (!currentRow) {
          message.warning('请先勾选或点击选择要删除的实例')
          return
        }
        await service.deleteInstance(currentRow as GatewayInstance)
        break
      }

      case 'export': {
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const exportRow = gridRef.value.getSelectedOrCurrentRecord()
        if (!exportRow) {
          message.warning('请先勾选或点击选择要导出配置的实例')
          return
        }
        exportInstanceId.value = (exportRow as GatewayInstance).gatewayInstanceId
        exportVisible.value = true
        break
      }

      case 'import':
        importVisible.value = true
        break

      case 'search': {
        // 如果传递了表单数据，直接使用它进行查询
        // formData 参数在 RsSearchForm 的 handleToolbarClick 中传递
        await service.handleSearch(formData)
        break
      }
    }
  }

  /**
   * 打开配置管理对话框
   */
  const openConfigDialog = (instance: GatewayInstance) => {
    currentConfigInstanceId.value = instance.gatewayInstanceId
    configDialogVisible.value = true
  }

  /**
   * 打开日志配置对话框
   * 实例存在的情况下编辑日志配置，所以一定是编辑模式
   */
  const openLogConfigDialog = async (instance: GatewayInstance) => {
    logConfigDialogMode.value = 'edit'
    
    try {
      // 通过 instance 的 logConfigId 获取日志配置详情
      if (!instance.logConfigId) {
        message.warning('该实例尚未配置日志，请先创建日志配置')
        return
      }

      const response: any = await gatewayApi.getLogConfig(
        instance.logConfigId
      )
      
      if (isApiSuccess(response)) {
        const logConfig = parseJsonData<Record<string, any>>(response, {})
        
        // 处理 sensitiveFields：如果是字符串，需要解析为数组
        if (logConfig.sensitiveFields) {
          if (typeof logConfig.sensitiveFields === 'string') {
            try {
              logConfig.sensitiveFields = JSON.parse(logConfig.sensitiveFields)
            } catch {
              logConfig.sensitiveFields = logConfig.sensitiveFields.split(',').map((s: string) => s.trim()).filter(Boolean)
            }
          }
        } else {
          logConfig.sensitiveFields = []
        }

        // 解析 JSON 列为嵌套对象，供 NamePath 字段读写
        flattenExtProperty(logConfig)
        parseJsonObjectColumn(logConfig, 'fileConfig')
        
        currentLogConfig.value = logConfig
        logConfigDialogVisible.value = true
      } else {
        const errorMsg = getApiMessage(response, '获取日志配置失败')
        message.error(errorMsg)
      }
    } catch (error) {
      message.error('获取日志配置失败')
    }
  }

  /**
   * 关闭日志配置对话框
   */
  const closeLogConfigDialog = () => {
    logConfigDialogVisible.value = false
    currentLogConfig.value = null
  }

  /**
   * 处理日志配置提交
   */
  const handleLogConfigSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (logConfigDialogMode.value === 'view') {
      return
    }

    logConfigSubmitting.value = true
    try {
      // formData 中应该包含 logConfigId
      if (!formData.logConfigId) {
        message.error('日志配置信息不完整')
        return
      }

      // 处理 sensitiveFields：如果是数组，转换为字符串格式（JSON 字符串）
      const processedData = { ...formData }
      if (processedData.sensitiveFields && Array.isArray(processedData.sensitiveFields)) {
        processedData.sensitiveFields = JSON.stringify(processedData.sensitiveFields)
      }

      // 将嵌套对象打包回 JSON 字符串列
      unflattenExtProperty(processedData)
      stringifyJsonObjectColumn(processedData, 'fileConfig')

      const response: any = await gatewayApi.editLogConfig(processedData)
      
      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '日志配置保存成功')
        message.success(successMsg)
        closeLogConfigDialog()
      } else {
        const errorMsg = getApiMessage(response, '保存日志配置失败')
        message.error(errorMsg)
      }
    } catch (error) {
      message.error('保存日志配置失败')
    } finally {
      logConfigSubmitting.value = false
    }
  }

  /**
   * 关闭配置管理对话框
   */
  const closeConfigDialog = () => {
    configDialogVisible.value = false
    currentConfigInstanceId.value = ''
  }

  /**
   * 处理启动实例
   */
  const handleStartInstance = async (instance: GatewayInstance) => {
    const confirmed = await rsConfirm.warning({
      title: '确认启动',
      subtitle: '启动后将开始处理请求',
      description: `确定要启动实例"${instance.instanceName}"吗？`,
      icon: PlayCircleOutline,
      confirmText: '确定启动',
      cancelText: '取消',
      width: 500
    })
    
    if (confirmed) {
      service.startInstance(instance)
    }
  }

  /**
   * 处理停止实例
   */
  const handleStopInstance = async (instance: GatewayInstance) => {
    const confirmed = await rsConfirm.warning({
      title: '确认停止',
      subtitle: '停止后将无法处理请求',
      description: `确定要停止实例"${instance.instanceName}"吗？`,
      icon: StopCircleOutline,
      confirmText: '确定停止',
      cancelText: '取消',
      width: 500
    })
    
    if (confirmed) {
      service.stopInstance(instance)
    }
  }

  /**
   * 处理网关重载
   */
  const handleReloadInstance = async (instance: GatewayInstance) => {
    const confirmed = await rsConfirm.warning({
      title: '确认网关重载',
      subtitle: '重载将重新加载配置',
      description: `确定要对实例"${instance.instanceName}"执行网关重载操作吗？`,
      icon: RefreshOutline,
      confirmText: '确定重载',
      cancelText: '取消',
      width: 500
    })
    
    if (confirmed) {
      service.reloadInstance(instance)
    }
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = async ({ key, row }: { key: string; row?: GatewayInstance }) => {
    // 导入不依赖行数据，空白区域右键也可触发
    if (key === 'import') {
      importVisible.value = true
      return
    }

    if (!row) return

    switch (key) {
      case 'view':
        await openViewDialog(row)
        break

      case 'edit':
        await openEditDialog(row)
        break

      case 'delete':
        await service.deleteInstance(row)
        break

      case 'start':
        handleStartInstance(row)
        break

      case 'stop':
        handleStopInstance(row)
        break

      case 'ipAccessControl':
        await openIpAccessControlDialog(row)
        break

      case 'userAgentAccessControl':
        await openUserAgentAccessControlDialog(row)
        break

      case 'apiAccessControl':
        await openApiAccessControlDialog(row)
        break

      case 'domainAccessControl':
        await openDomainAccessControlDialog(row)
        break

      case 'corsConfig':
        await openCorsConfigDialog(row)
        break

      case 'authConfig':
        await openAuthConfigDialog(row)
        break

      case 'rateLimitConfig':
        await openRateLimitConfigDialog(row)
        break

      case 'logConfig':
        openLogConfigDialog(row)
        break

      case 'reload':
        handleReloadInstance(row)
        break

      case 'export':
        exportInstanceId.value = row.gatewayInstanceId
        exportVisible.value = true
        break

    }
  }

  return {
    // 业务服务（包含 model 与增删改查）
    service,

    // 表单对话框（新增/编辑/查看共用）
    formDialogVisible,
    formDialogMode,
    currentEditInstance,
    submitting,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    handleFormSubmit,

    // 配置管理对话框
    configDialogVisible,
    currentConfigInstanceId,
    openConfigDialog,
    closeConfigDialog,

    // 日志配置对话框
    logConfigDialogVisible,
    logConfigDialogMode,
    currentLogConfig,
    logConfigSubmitting,
    openLogConfigDialog,
    closeLogConfigDialog,
    handleLogConfigSubmit,

    // IP访问控制对话框
    ipAccessControlDialogVisible,
    ipAccessControlSecurityConfigId,
    openIpAccessControlDialog,

    // User-Agent访问控制对话框
    userAgentAccessControlDialogVisible,
    userAgentAccessControlSecurityConfigId,
    openUserAgentAccessControlDialog,

    // API访问控制对话框
    apiAccessControlDialogVisible,
    apiAccessControlSecurityConfigId,
    openApiAccessControlDialog,

    // 域名访问控制对话框
    domainAccessControlDialogVisible,
    domainAccessControlSecurityConfigId,
    openDomainAccessControlDialog,

    // 跨域配置对话框
    corsConfigDialogVisible,
    corsConfigGatewayInstanceId,
    openCorsConfigDialog,

    // 认证配置对话框
    authConfigDialogVisible,
    authConfigGatewayInstanceId,
    openAuthConfigDialog,

    // 限流配置对话框
    rateLimitConfigDialogVisible,
    rateLimitConfigGatewayInstanceId,
    openRateLimitConfigDialog,

    // 导出/导入
    exportVisible,
    exportInstanceId,
    importVisible,

    // 事件处理器
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
    handleStartInstance,
    handleStopInstance,
    handleReloadInstance,
  }
}

export type GatewayInstancePage = ReturnType<typeof useGatewayInstancePage>

