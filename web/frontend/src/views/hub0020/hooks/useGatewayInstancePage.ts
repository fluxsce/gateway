/**
 * 网关实例管理页面级 Hook
 * - 组合 useGatewayInstanceService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useGDialog } from '@/components/gdialog'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { PlayCircleOutline, RefreshOutline, StopCircleOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import * as gatewayApi from '../api'
import type { GatewayInstance } from '../types'
import { useGatewayInstanceService } from './useGatewayInstanceService'

/**
 * 网关实例管理页面级 Hook
 */
export function useGatewayInstancePage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

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

      // 将证书和私钥内容转换为文件列表格式（用于回显）
      const formData: any = { ...detailInstance }
      
      // 如果有证书内容，转换为文件列表格式
      if (detailInstance.certContent) {
        formData.certFileList = [{
          id: 'cert-file',
          name: detailInstance.certFilePath || 'certificate.pem',
          content: detailInstance.certContent,
          status: 'finished',
        }]
      } else {
        formData.certFileList = []
      }

      // 如果有私钥内容，转换为文件列表格式
      if (detailInstance.keyContent) {
        formData.keyFileList = [{
          id: 'key-file',
          name: detailInstance.keyFilePath || 'private-key.pem',
          content: detailInstance.keyContent,
          status: 'finished',
        }]
      } else {
        formData.keyFileList = []
      }

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

      // 将证书和私钥内容转换为文件列表格式（用于回显）
      const formData: any = { ...detailInstance }
      
      // 如果有证书内容，转换为文件列表格式
      if (detailInstance.certContent) {
        formData.certFileList = [{
          id: 'cert-file',
          name: detailInstance.certFilePath || 'certificate.pem',
          content: detailInstance.certContent,
          status: 'finished',
        }]
      } else {
        formData.certFileList = []
      }

      // 如果有私钥内容，转换为文件列表格式
      if (detailInstance.keyContent) {
        formData.keyFileList = [{
          id: 'key-file',
          name: detailInstance.keyFilePath || 'private-key.pem',
          content: detailInstance.keyContent,
          status: 'finished',
        }]
      } else {
        formData.keyFileList = []
      }

      formDialogMode.value = 'view'
      currentEditInstance.value = formData
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取实例详情失败')
    }
  }

  /**
   * 处理搜索（接收 SearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  /**
   * 提交表单（新增/编辑共用，由 GdataFormModal 收集表单数据后回调）
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      return
    }

    submitting.value = true
    try {
      // 处理文件内容：将文件列表中的 content 和 name 提取到对应的字段
      const processedData = { ...formData }
      
      // 处理证书文件：从 certFileList 中提取 content 和 name
      if (processedData.certFileList && Array.isArray(processedData.certFileList) && processedData.certFileList.length > 0) {
        const certFile = processedData.certFileList[0]
        if (certFile) {
          if (certFile.content) {
            processedData.certContent = certFile.content
          }
          // 提取文件名到 certFilePath（如果用户上传了新文件，使用新文件名）
          if (certFile.name) {
            processedData.certFilePath = certFile.name
          }
        }
      } else if (formDialogMode.value === 'edit' && currentEditInstance.value) {
        // 编辑模式下，如果用户没有上传新文件，保留原有的文件名
        if (currentEditInstance.value.certFilePath && !processedData.certFilePath) {
          processedData.certFilePath = currentEditInstance.value.certFilePath
        }
      }
      // 删除 certFileList，不提交给后端
      delete processedData.certFileList

      // 处理私钥文件：从 keyFileList 中提取 content 和 name
      if (processedData.keyFileList && Array.isArray(processedData.keyFileList) && processedData.keyFileList.length > 0) {
        const keyFile = processedData.keyFileList[0]
        if (keyFile) {
          if (keyFile.content) {
            processedData.keyContent = keyFile.content
          }
          // 提取文件名到 keyFilePath（如果用户上传了新文件，使用新文件名）
          if (keyFile.name) {
            processedData.keyFilePath = keyFile.name
          }
        }
      } else if (formDialogMode.value === 'edit' && currentEditInstance.value) {
        // 编辑模式下，如果用户没有上传新文件，保留原有的文件名
        if (currentEditInstance.value.keyFilePath && !processedData.keyFilePath) {
          processedData.keyFilePath = currentEditInstance.value.keyFilePath
        }
      }
      // 删除 keyFileList，不提交给后端
      delete processedData.keyFileList


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
        // 编辑当前高亮的行（点击选中的行）
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要编辑的实例')
          return
        }
        await openEditDialog(currentRow as GatewayInstance)
        break
      }

      case 'delete': {
        // 删除当前高亮的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要删除的实例')
          return
        }
        await service.deleteInstance(currentRow as GatewayInstance)
        break
      }

      case 'search': {
        // 如果传递了表单数据，直接使用它进行查询
        // formData 参数在 SearchForm 的 handleToolbarClick 中传递
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
   * 将 extProperty JSON 字符串展开为扁平字段（用于表单回填）
   */
  const flattenExtProperty = (data: Record<string, any>) => {
    if (!data.extProperty || typeof data.extProperty !== 'string') {
      return
    }
    try {
      const extObj = JSON.parse(data.extProperty || '{}')
      if (extObj && typeof extObj === 'object') {
        // 将 extProperty 对象的属性展开为 extProperty.xxx 字段
        Object.keys(extObj).forEach((key) => {
          const value = extObj[key]
          // alertStatusCodes 特殊处理：确保是字符串数组
          if (key === 'alertStatusCodes') {
            if (Array.isArray(value)) {
              data[`extProperty.${key}`] = value.map((v: any) => String(v))
            } else if (typeof value === 'string') {
              data[`extProperty.${key}`] = value.split(',').map((s: string) => s.trim()).filter(Boolean)
            }
          } else {
            data[`extProperty.${key}`] = value
          }
        })
      }
    } catch {
      // ignore
    }
  }

  /**
   * 将扁平字段打包回 extProperty JSON 字符串（用于提交）
   */
  const unflattenExtProperty = (data: Record<string, any>) => {
    const extPropertyObj: Record<string, any> = {}
    
    // 收集所有 extProperty.xxx 字段
    Object.keys(data).forEach((key) => {
      if (key.startsWith('extProperty.')) {
        const subKey = key.substring('extProperty.'.length)
        extPropertyObj[subKey] = data[key]
        delete data[key] // 删除扁平字段
      }
    })
    
    // 转换为 JSON 字符串
    if (Object.keys(extPropertyObj).length === 0) {
      data.extProperty = ''
    } else {
      try {
        data.extProperty = JSON.stringify(extPropertyObj)
      } catch {
        data.extProperty = ''
      }
    }
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

        // 展开 extProperty 为扁平字段
        flattenExtProperty(logConfig)
        
        currentLogConfig.value = logConfig
        logConfigDialogVisible.value = true
      } else {
        const errorMsg = getApiMessage(response, '获取日志配置失败')
        message.error(errorMsg)
      }
    } catch (error) {
      message.error('获取日志配置失败')
      console.error('Error fetching log config:', error)
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

      // 将扁平字段打包回 extProperty JSON 字符串
      unflattenExtProperty(processedData)

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
      console.error('Error saving log config:', error)
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
    const confirmed = await gDialog.warning({
      title: '确认启动',
      subtitle: '启动后将开始处理请求',
      content: `确定要启动实例"${instance.instanceName}"吗？`,
      icon: PlayCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定启动',
      negativeText: '取消',
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
    const confirmed = await gDialog.warning({
      title: '确认停止',
      subtitle: '停止后将无法处理请求',
      content: `确定要停止实例"${instance.instanceName}"吗？`,
      icon: StopCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定停止',
      negativeText: '取消',
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
    const confirmed = await gDialog.warning({
      title: '确认网关重载',
      subtitle: '重载将重新加载配置',
      content: `确定要对实例"${instance.instanceName}"执行网关重载操作吗？`,
      icon: RefreshOutline,
      headerStyle: 'gradient',
      positiveText: '确定重载',
      negativeText: '取消',
      width: 500
    })
    
    if (confirmed) {
      service.reloadInstance(instance)
    }
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = async ({ code, row }: { code: string; row?: GatewayInstance }) => {
    if (!row) return

    switch (code) {
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

