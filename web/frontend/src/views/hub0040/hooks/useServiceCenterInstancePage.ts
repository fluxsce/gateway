/**
 * 服务中心实例管理页面级 Hook
 * - 组合 useServiceCenterInstanceService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useAppMessage } from '@/composables/useAppMessage'
import { rsConfirm } from '@/ui'
import { flattenExtProperty, unflattenExtProperty } from '@/utils/format'
import { consumeTextFileField, filesFromTextContent } from '@/utils/uploadFile'
import { PlayCircleOutline, RefreshOutline, StopCircleOutline } from '@vicons/ionicons5'
import type { Ref } from 'vue'
import { ref, watch } from 'vue'
import type { CenterAuthToken, CenterConnection, CenterIssuedAuthToken, CenterOverview, ServiceCenterInstance } from '../types'
import { canInstanceAction } from './model'
import { useServiceCenterInstanceService } from './useServiceCenterInstanceService'

function parseJsonList(raw: unknown): string[] {
  if (Array.isArray(raw)) return raw.map(String).filter(Boolean)
  if (typeof raw !== 'string' || !raw.trim()) return []
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed.map(String).filter(Boolean) : []
  } catch {
    return raw.split(',').map((s) => s.trim()).filter(Boolean)
  }
}

function accessConfigError(data: Record<string, any>): string | null {
  if (data.enableMTLS === 'Y' && data.enableTLS !== 'Y') {
    return '启用双向 TLS 前请先启用 TLS'
  }
  if (data.enableMTLS === 'Y' && !String(data.certChainContent || '').trim()) {
    return '启用双向 TLS 时必须上传客户端 CA 证书'
  }
  return null
}

function hydrateInstanceForm(detail: ServiceCenterInstance): Record<string, any> {
  const formData: Record<string, any> = { ...detail }
  formData.certFileList = filesFromTextContent(
    detail.certFilePath || 'certificate.pem',
    detail.certContent || '',
  )
  formData.keyFileList = filesFromTextContent(
    detail.keyFilePath || 'private-key.pem',
    detail.keyContent || '',
  )
  formData.certChainFileList = filesFromTextContent(
    'client-ca.pem',
    detail.certChainContent || '',
  )
  formData.ipWhitelist = parseJsonList(formData.ipWhitelist)
  formData.ipBlacklist = parseJsonList(formData.ipBlacklist)
  flattenExtProperty(formData)
  return formData
}

/**
 * 服务中心实例管理页面级 Hook
 */
function instanceIdentity(row: ServiceCenterInstance) {
  return `${row.tenantId}\0${row.instanceName}\0${row.environment}`
}

export function useServiceCenterInstancePage(searchFormRef?: Ref<any> | any) {
  const message = useAppMessage()
  // 业务服务（包含 model、增删改查等）
  const service = useServiceCenterInstanceService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditInstance = ref<ServiceCenterInstance | null>(null)
  const submitting = ref(false)
  const runtimeDrawerVisible = ref(false)
  const runtimeLoading = ref(false)
  const runtimeInstance = ref<ServiceCenterInstance | null>(null)
  const runtimeOverview = ref<CenterOverview | null>(null)
  const runtimeConnections = ref<CenterConnection[]>([])
  const tokenDrawerVisible = ref(false)
  const tokenLoading = ref(false)
  const tokenIssuing = ref(false)
  const tokenInstance = ref<ServiceCenterInstance | null>(null)
  const tokenList = ref<CenterAuthToken[]>([])
  const issuedToken = ref<CenterIssuedAuthToken | null>(null)
  const selectedInstance = ref<ServiceCenterInstance | null>(null)

  watch(
    () => service.model.instanceList.value,
    (list) => {
      const current = selectedInstance.value
      if (!current) return
      const next = list.find((row) => instanceIdentity(row) === instanceIdentity(current))
      selectedInstance.value = next || null
    },
  )

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
  const openEditDialog = async (instance: ServiceCenterInstance) => {
    try {
      // 获取完整详情
      const detailInstance = await service.getInstanceDetail(
        instance.instanceName,
        instance.environment
      )
      
      if (!detailInstance) {
        message.error('获取实例详情失败')
        return
      }

      formDialogMode.value = 'edit'
      currentEditInstance.value = hydrateInstanceForm(detailInstance) as ServiceCenterInstance
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
  const openViewDialog = async (instance: ServiceCenterInstance) => {
    try {
      // 从后端获取最新数据，确保显示的是最新状态
      const detailInstance = await service.getInstanceDetail(
        instance.instanceName,
        instance.environment
      )
      
      if (!detailInstance) {
        message.error('获取实例详情失败')
        return
      }

      formDialogMode.value = 'view'
      currentEditInstance.value = hydrateInstanceForm(detailInstance) as ServiceCenterInstance
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

      await consumeTextFileField(processedData, 'certFileList', 'certContent', 'certFilePath', {
        fallbackPath: editFallback?.certFilePath,
      })
      await consumeTextFileField(processedData, 'keyFileList', 'keyContent', 'keyFilePath', {
        fallbackPath: editFallback?.keyFilePath,
      })
      await consumeTextFileField(processedData, 'certChainFileList', 'certChainContent', '_certChainPath')
      delete processedData._certChainPath

      const accessError = accessConfigError(processedData)
      if (accessError) {
        message.error(accessError)
        return
      }

      // 处理 IP 白名单和黑名单（数组转 JSON 字符串）
      if (Array.isArray(processedData.ipWhitelist)) {
        processedData.ipWhitelist = processedData.ipWhitelist.length > 0
          ? JSON.stringify(processedData.ipWhitelist)
          : ''
      }
      if (Array.isArray(processedData.ipBlacklist)) {
        processedData.ipBlacklist = processedData.ipBlacklist.length > 0
          ? JSON.stringify(processedData.ipBlacklist)
          : ''
      }

      if (typeof processedData.environment === 'string') {
        processedData.environment = processedData.environment.trim()
      }

      // 将 extProperty 嵌套对象打包回 JSON 字符串
      unflattenExtProperty(processedData)

      if (formDialogMode.value === 'create') {
        // 新增模式
        const success = await service.addInstance(processedData as ServiceCenterInstance)
        if (success) {
          closeFormDialog()
        }
      } else if (formDialogMode.value === 'edit') {
        // 编辑模式
        if (!currentEditInstance.value) return
        const success = await service.editInstance({
          ...processedData,
          instanceName: currentEditInstance.value.instanceName,
          environment: currentEditInstance.value.environment,
        } as Partial<ServiceCenterInstance> & { instanceName: string; environment: string })
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
    if (!canInstanceAction(key) && key !== 'search' && key !== 'reset') {
      message.warning('没有操作权限')
      return
    }
    switch (key) {
      case 'add':
        // 直接打开新增对话框
        openAddDialog()
        break

      case 'edit': {
        if (!selectedInstance.value) {
          message.warning('请先选择要编辑的实例')
          return
        }
        await openEditDialog(selectedInstance.value)
        break
      }

      case 'delete': {
        if (!selectedInstance.value) {
          message.warning('请先选择要删除的实例')
          return
        }
        await service.deleteInstance(selectedInstance.value)
        break
      }

      case 'search': {
        // 如果传递了表单数据，直接使用它进行查询
        // formData 参数在 RsSearchForm 的 handleToolbarClick 中传递
        await service.handleSearch(formData)
        break
      }
    }
  }

  /**
   * 处理启动实例
   */
  const handleStartInstance = async (instance: ServiceCenterInstance) => {
    const confirmed = await rsConfirm.warning({
      title: '确认启动',
      subtitle: '启动后将开始处理请求',
      description: `确定要启动实例"${instance.instanceName}" (${instance.environment}) 吗？`,
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
  const handleStopInstance = async (instance: ServiceCenterInstance) => {
    const confirmed = await rsConfirm.warning({
      title: '确认停止',
      subtitle: '停止后将无法处理请求',
      description: `确定要停止实例"${instance.instanceName}" (${instance.environment}) 吗？`,
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
   * 处理配置重载
   */
  const handleReloadInstance = async (instance: ServiceCenterInstance) => {
    const confirmed = await rsConfirm.warning({
      title: '确认重载配置',
      subtitle: '重载将重新加载配置',
      description: `确定要对实例"${instance.instanceName}" (${instance.environment}) 执行配置重载操作吗？`,
      icon: RefreshOutline,
      confirmText: '确定重载',
      cancelText: '取消',
      width: 500
    })
    
    if (confirmed) {
      service.reloadInstance(instance)
    }
  }

  const selectInstance = (instance: ServiceCenterInstance) => {
    selectedInstance.value = instance
  }

  /**
   * 卡片菜单 / 底部按钮
   */
  const handleCardAction = async (key: string, instance: ServiceCenterInstance) => {
    selectedInstance.value = instance
    await handleMenuClick({ key, row: instance })
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = async ({ key, row }: { key: string; row?: ServiceCenterInstance }) => {
    if (!row) return
    if (!canInstanceAction(key)) {
      message.warning('没有操作权限')
      return
    }

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

      case 'reload':
        handleReloadInstance(row)
        break

      case 'runtime':
        await handleOpenRuntime(row)
        break

      case 'token':
        await handleOpenToken(row)
        break
    }
  }

  const handleOpenRuntime = async (instance: ServiceCenterInstance) => {
    runtimeInstance.value = instance
    runtimeDrawerVisible.value = true
    runtimeLoading.value = true
    try {
      const [overview, connections] = await Promise.all([
        service.getRuntimeOverview(instance.instanceName, instance.environment),
        service.listRuntimeConnections(instance.instanceName, instance.environment),
      ])
      runtimeOverview.value = overview
      runtimeConnections.value = connections
    } finally {
      runtimeLoading.value = false
    }
  }

  const handleOpenToken = async (instance: ServiceCenterInstance) => {
    tokenInstance.value = instance
    issuedToken.value = null
    tokenDrawerVisible.value = true
    tokenLoading.value = true
    try {
      tokenList.value = await service.listAuthTokens(instance.instanceName, instance.environment)
    } finally {
      tokenLoading.value = false
    }
  }

  const handleIssueToken = async (payload: { tokenName: string; expireDays: number }) => {
    const instance = tokenInstance.value
    if (!instance) return
    if (!canInstanceAction('edit')) {
      message.warning('没有颁发权限')
      return
    }
    tokenIssuing.value = true
    try {
      const issued = await service.issueAuthToken(instance, payload.tokenName, payload.expireDays)
      if (!issued) return
      issuedToken.value = issued
      tokenList.value = await service.listAuthTokens(instance.instanceName, instance.environment)
    } finally {
      tokenIssuing.value = false
    }
  }

  const handleRevokeToken = async (token: CenterAuthToken) => {
    const instance = tokenInstance.value
    if (!instance) return
    if (!canInstanceAction('edit')) {
      message.warning('没有吊销权限')
      return
    }
    const confirmed = await rsConfirm.warning({
      title: '确认吊销',
      subtitle: '吊销后接入方将无法再用此令牌',
      description: `确定要吊销「${token.tokenName || token.tokenPreview}」吗？`,
      confirmText: '确定吊销',
      cancelText: '取消',
      width: 480,
    })
    if (!confirmed) return
    const ok = await service.revokeAuthToken(instance, token.tokenId)
    if (!ok) return
    if (issuedToken.value?.tokenId === token.tokenId) {
      issuedToken.value = null
    }
    tokenList.value = await service.listAuthTokens(instance.instanceName, instance.environment)
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

    // 事件处理器
    handleToolbarClick,
    handleMenuClick,
    handleCardAction,
    handleSearch,
    handleStartInstance,
    handleStopInstance,
    handleReloadInstance,
    handleOpenRuntime,
    selectedInstance,
    selectInstance,
    runtimeDrawerVisible,
    runtimeLoading,
    runtimeInstance,
    runtimeOverview,
    runtimeConnections,
    tokenDrawerVisible,
    tokenLoading,
    tokenIssuing,
    tokenInstance,
    tokenList,
    issuedToken,
    handleIssueToken,
    handleRevokeToken,
  }
}

export type ServiceCenterInstancePage = ReturnType<typeof useServiceCenterInstancePage>

