/**
 * 服务定义列表页面级 Hook
 * - 组合 useServiceDefinitionService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { getApiMessage, isApiSuccess } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { onMounted, ref } from 'vue'
import { getServiceDefinition } from '../../../api'
import type { ServiceDefinition } from '../types'
import { useServiceDefinitionService } from './service'

/**
 * 服务定义列表页面级 Hook
 */
export function useServiceDefinitionPage(
  gatewayInstanceId?: string,
  searchFormRef?: Ref<any> | any,
  gridRef?: Ref<any> | any
) {
  const message = useMessage()

  // 服务发现选择相关状态（需要在 serviceResult 之前定义，以便传递给 model）
  const selectedService = ref<any | null>(null) // 当前选择的服务
  const serviceSelectionVisible = ref(false) // 服务选择器弹窗显示状态

  // 业务服务（包含 model、增删改查等）
  // gatewayInstanceId 直接作为 proxyConfigId 使用
  const serviceResult = useServiceDefinitionService(gatewayInstanceId, searchFormRef)

  // 对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditService = ref<ServiceDefinition | null>(null)
  
  const showNodeDialog = ref(false)
  const currentServiceId = ref<string>('')

  // ============= JSON 字段转换方法 =============

  /**
   * 将配置对象转换为表单数据格式（查询时：将 JSON 字符串字段解析为点号分隔字段）
   * 注意：serviceMetadata 保持为 JSON 字符串格式，不展开
   */
  const convertToFormData = (service: ServiceDefinition): any => {
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

    // 构建表单数据对象
    const formData: any = {
      ...service,
    }

    // 将 JSON 字符串字段解析为对象，然后展开为点号分隔字段
    // serviceMetadata 保持为 JSON 字符串格式，不展开
    const jsonFieldsToExpand = ['discoveryConfig', 'healthCheckHeaders', 'loadBalancerConfig']
    
    jsonFieldsToExpand.forEach((fieldName) => {
      const jsonValue = parseJson((service as any)[fieldName])
      if (jsonValue && typeof jsonValue === 'object' && !Array.isArray(jsonValue)) {
        // 将对象的属性展开为点号分隔字段（如 discoveryConfig.xxx）
        Object.keys(jsonValue).forEach((key) => {
          formData[`${fieldName}.${key}`] = jsonValue[key]
        })
        // 移除原始字段（已展开为点号分隔字段）
        delete formData[fieldName]
      }
    })

    // serviceMetadata 保持为 JSON 字符串格式，如果不存在则设置为 undefined
    if (formData.serviceMetadata && typeof formData.serviceMetadata === 'string') {
      // 保持为字符串，不展开
    } else if (formData.serviceMetadata && typeof formData.serviceMetadata === 'object') {
      // 如果是对象，转换为 JSON 字符串
      formData.serviceMetadata = JSON.stringify(formData.serviceMetadata)
    } else {
      formData.serviceMetadata = undefined
    }

    return formData
  }

  /**
   * 将表单数据转换为API数据格式（保存时：将点号分隔字段合并为 JSON 字符串）
   * 注意：serviceMetadata 保持为 JSON 字符串格式，直接使用
   */
  const convertToApiData = (formData: Record<string, any>): any => {
    // JSON 字段列表（需要展开的字段）
    const jsonFieldsToExpand = ['discoveryConfig', 'healthCheckHeaders', 'loadBalancerConfig']

    // 构建 API 数据对象
    const apiData: Record<string, any> = {}

    // 将点号分隔字段合并为 JSON 字符串
    jsonFieldsToExpand.forEach((fieldName) => {
      const jsonObj: Record<string, any> = {}
      
      // 遍历所有以 fieldName. 开头的字段
      Object.keys(formData).forEach((key) => {
        if (key.startsWith(`${fieldName}.`)) {
          const subKey = key.replace(`${fieldName}.`, '')
          jsonObj[subKey] = formData[key]
        }
      })

      // 如果有子字段，则转换为 JSON 字符串
      if (Object.keys(jsonObj).length > 0) {
        apiData[fieldName] = JSON.stringify(jsonObj)
      }
    })

    // 处理 serviceMetadata：保持为 JSON 字符串格式
    if (formData.serviceMetadata) {
      if (typeof formData.serviceMetadata === 'string') {
        // 如果是字符串，验证是否为有效 JSON，然后直接使用
        try {
          JSON.parse(formData.serviceMetadata)
          apiData.serviceMetadata = formData.serviceMetadata
        } catch {
          // JSON 格式无效，使用空对象
          apiData.serviceMetadata = '{}'
        }
      } else if (typeof formData.serviceMetadata === 'object') {
        // 如果是对象，转换为 JSON 字符串
        apiData.serviceMetadata = JSON.stringify(formData.serviceMetadata)
      } else {
        apiData.serviceMetadata = '{}'
      }
    } else {
      apiData.serviceMetadata = undefined
    }

    // 排除点号分隔字段，保留其他字段
    Object.keys(formData).forEach((key) => {
      // 排除以 JSON 字段名开头的点号分隔字段（这些字段已经合并到对应的 JSON 字段中）
      // 排除 serviceMetadata（已单独处理）
      if (!jsonFieldsToExpand.some((fieldName) => key.startsWith(`${fieldName}.`)) && key !== 'serviceMetadata') {
        apiData[key] = formData[key]
      }
    })

    return apiData
  }

  // ============= 统一校验 =============

  /**
   * 校验是否已选择网关实例
   * @param showMessage 是否显示提示消息
   * @returns 是否已选择实例
   */
  const validateInstanceSelected = (showMessage = true): boolean => {
    if (!gatewayInstanceId) {
      if (showMessage) {
        message.warning('请先选择网关实例')
      }
      return false
    }
    return true
  }

  // ============= 对话框操作 =============

  /**
   * 打开新增对话框
   */
  const openAddDialog = () => {
    if (!validateInstanceSelected()) {
      return
    }
    formDialogMode.value = 'create'
    currentEditService.value = null
    selectedService.value = null // 重置服务选择
    formDialogVisible.value = true
  }

  /**
   * 打开编辑对话框
   */
  const openEditDialog = async (service: ServiceDefinition) => {
    if (!validateInstanceSelected()) {
      return
    }
    try {
      // 获取完整详情
      const response = await getServiceDefinition(service.serviceDefinitionId, 'default')
      
      if (isApiSuccess(response)) {
        const detailService = JSON.parse(response.bizData) as ServiceDefinition
        
        // 使用 convertToFormData 将 JSON 字段展开为点号分隔字段
        const formData = convertToFormData(detailService)
        
        formDialogMode.value = 'edit'
        currentEditService.value = formData
        formDialogVisible.value = true
      } else {
        message.error(getApiMessage(response, '获取服务定义详情失败'))
      }
    } catch (error) {
      message.error('获取服务定义详情失败')
    }
  }

  /**
   * 打开复制对话框
   */
  const openCopyDialog = async (service: ServiceDefinition) => {
    if (!validateInstanceSelected()) {
      return
    }
    try {
      // 获取完整详情
      const response = await getServiceDefinition(service.serviceDefinitionId, 'default')
      
      if (isApiSuccess(response)) {
        const detailService = JSON.parse(response.bizData) as ServiceDefinition
        
        // 使用 convertToFormData 将 JSON 字段展开为点号分隔字段
        const formData = convertToFormData(detailService)
        
        // 复制数据，修改服务名称，清除ID
        const copiedService = {
          ...formData,
          serviceDefinitionId: undefined, // 清除ID，作为新记录
          serviceName: `${detailService.serviceName}_copy`,
          activeFlag: 'Y', // 新记录默认启用
        }
        
        formDialogMode.value = 'create' // 复制操作视为新增
        currentEditService.value = copiedService
        formDialogVisible.value = true
      } else {
        message.error(getApiMessage(response, '获取服务定义详情失败'))
      }
    } catch (error) {
      message.error('获取服务定义详情失败')
    }
  }

  /**
   * 打开节点管理对话框
   */
  const openNodeDialog = (service: ServiceDefinition) => {
    if (!validateInstanceSelected()) {
      return
    }
    // 设置服务定义ID，然后打开对话框
    currentServiceId.value = service.serviceDefinitionId
    showNodeDialog.value = true
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditService.value = null
    selectedService.value = null // 重置服务选择
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

    // 校验是否已选择实例
    if (!validateInstanceSelected()) {
      return false
    }

    // 如果是服务发现类型，确保 serviceMetadata 已设置
    if (formData.serviceType === 1) {
      if (selectedService.value) {
        // 更新服务元数据
        updateServiceMetadata(formData)
      } else if (!formData.serviceMetadata) {
        message.warning('请选择注册服务')
        return false
      }
    }

    try {
      // 使用 convertToApiData 将点号分隔字段合并为 JSON 字符串
      const apiData = convertToApiData(formData)
      
      // 准备提交数据
      const processedData: Partial<ServiceDefinition> & { tenantId: string } = {
        ...apiData,
        tenantId: 'default',
        proxyConfigId: gatewayInstanceId || formData.proxyConfigId,
      }

      let success = false
      if (formDialogMode.value === 'create') {
        success = await serviceResult.addService(processedData as any)
      } else if (formDialogMode.value === 'edit' && currentEditService.value) {
        success = await serviceResult.editService(
          currentEditService.value.serviceDefinitionId,
          processedData as any
        )
      }

      if (success) {
        closeFormDialog()
      }
    } catch (error) {
      message.error('操作失败')
    }
  }

  /**
   * 处理服务选择（从 ServiceRegistrySelector 回调）
   */
  const handleServiceSelect = (service: any) => {
    selectedService.value = service
    console.log('已选择服务:', service.serviceName)
  }

  /**
   * 更新服务元数据到表单数据
   */
  const updateServiceMetadata = (formData?: Record<string, any>) => {
    if (selectedService.value) {
      // 直接保存扁平的服务数据，不嵌套selectedService层
      const metadata = {
        serviceName: selectedService.value.serviceName,
        serviceGroupId: selectedService.value.serviceGroupId,
        groupName: selectedService.value.groupName,
        protocolType: selectedService.value.protocolType,
        contextPath: selectedService.value.contextPath,
        loadBalanceStrategy: selectedService.value.loadBalanceStrategy,
        healthCheckUrl: selectedService.value.healthCheckUrl,
        instances: selectedService.value.instances || []
      }
      // 转换为JSON字符串存储
      const metadataStr = JSON.stringify(metadata)
      if (formData) {
        formData.serviceMetadata = metadataStr
      }
      console.log('已更新服务元数据:', metadata)
      return metadataStr
    } else {
      if (formData) {
        formData.serviceMetadata = undefined
      }
      return undefined
    }
  }

  /**
   * 打开服务选择弹窗
   */
  const openServiceSelection = () => {
    serviceSelectionVisible.value = true
    console.log('打开服务选择弹窗')
  }

  /**
   * 获取负载均衡策略显示标签
   */
  const getLoadBalanceStrategyLabel = (strategy: string): string => {
    const strategyMap: Record<string, string> = {
      'ROUND_ROBIN': '轮询算法',
      'RANDOM': '随机算法',
      'IP_HASH': 'IP哈希算法',
      'LEAST_CONN': '最少连接算法',
      'LEAST_CONNECTIONS': '最少连接算法',
      'WEIGHTED_ROUND_ROBIN': '加权轮询算法',
      'CONSISTENT_HASH': '一致性哈希算法'
    }
    return strategyMap[strategy] || strategy
  }

  /**
   * 处理删除
   */
  const handleDelete = async (service: ServiceDefinition) => {
    if (!validateInstanceSelected()) {
      return
    }
    await serviceResult.deleteService(service.serviceDefinitionId)
  }

  /**
   * 处理批量删除
   */
  const handleBatchDelete = async () => {
    if (!validateInstanceSelected()) {
      return
    }
    if (!gridRef?.value) {
      message.warning('Grid 引用未设置')
      return
    }
    // 获取选中的记录
    const selectedRecords = gridRef.value.getCheckboxRecords?.() || gridRef.value.getSelectRecords?.() || []
    if (selectedRecords.length === 0) {
      message.warning('请选择要删除的服务定义')
      return
    }
    const serviceDefinitionIds = selectedRecords.map((record: ServiceDefinition) => record.serviceDefinitionId)
    await serviceResult.batchDeleteServices(serviceDefinitionIds)
  }

  /**
   * 处理搜索（合并 proxyConfigId，即 gatewayInstanceId）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    if (!validateInstanceSelected()) {
      return
    }
    // 合并 proxyConfigId（即 gatewayInstanceId）到查询参数中
    const searchParams = formData
      ? {
          ...formData,
          ...(gatewayInstanceId ? { proxyConfigId: gatewayInstanceId } : {}),
        }
      : gatewayInstanceId
        ? { proxyConfigId: gatewayInstanceId }
        : undefined
    await serviceResult.handleSearch(searchParams)
  }

  /**
   * 工具栏按钮点击处理
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break

      case 'delete': {
        // 批量删除选中的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        // 获取选中的记录
        const selectedRecords = gridRef.value.getCheckboxRecords?.() || gridRef.value.getSelectRecords?.() || []
        if (selectedRecords.length === 0) {
          message.warning('请选择要删除的服务定义')
          return
        }
        const serviceDefinitionIds = selectedRecords.map((record: ServiceDefinition) => record.serviceDefinitionId)
        await serviceResult.batchDeleteServices(serviceDefinitionIds)
        break
      }

      case 'manageNodes': {
        // 节点管理：需要选中单个服务定义
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        // 优先使用当前高亮的行
        const currentRow = gridRef.value.getCurrentRecord?.()
        if (currentRow) {
          openNodeDialog(currentRow as ServiceDefinition)
          return
        }
        // 如果没有当前行，尝试获取选中的记录
        const selectedRecords = gridRef.value.getCheckboxRecords?.() || gridRef.value.getSelectRecords?.() || []
        if (selectedRecords.length === 0) {
          message.warning('请先点击选择要管理的服务定义')
          return
        }
        if (selectedRecords.length > 1) {
          message.warning('请选择单个服务定义进行节点管理')
          return
        }
        openNodeDialog(selectedRecords[0] as ServiceDefinition)
        break
      }

      case 'search': {
        // 如果传递了表单数据，直接使用它进行查询
        // formData 参数在 SearchForm 的 handleToolbarClick 中传递
        await handleSearch(formData)
        break
      }
    }
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = ({ code, row }: { code: string; row?: ServiceDefinition }) => {
    if (!row) return

    switch (code) {
      case 'edit':
        openEditDialog(row)
        break
      case 'manageNodes':
        openNodeDialog(row)
        break
      case 'copy':
        openCopyDialog(row)
        break
      case 'delete':
        handleDelete(row)
        break
    }
  }

  /**
   * 刷新数据
   */
  const refresh = () => {
    serviceResult.loadServiceList()
  }

  /**
   * 加载服务列表（考虑 proxyConfigId，即 gatewayInstanceId）
   */
  const loadServiceList = async () => {
    if (gatewayInstanceId) {
      await serviceResult.loadServiceList({ proxyConfigId: gatewayInstanceId })
    } else {
      await serviceResult.loadServiceList()
    }
  }

  // ============= 生命周期 =============

  // 如果提供了 gatewayInstanceId，加载服务定义
  if (gatewayInstanceId) {
    onMounted(() => {
      serviceResult.loadServiceList({ proxyConfigId: gatewayInstanceId })
    })
  }

  // 重新设置 model 的 serviceFormConfig，注入服务选择相关方法
  // 由于 model 已经在 service 中初始化，我们需要动态更新字段配置
  const updatedFields = serviceResult.model.serviceFormConfig.fields.map((field: any) => {
    if (field.field === 'serviceSelection' && field.render) {
      const originalRender = field.render
      return {
        ...field,
        render: (formData: Record<string, any>) => {
          return originalRender(formData, {
            selectedService,
            openServiceSelection,
            getLoadBalanceStrategyLabel,
          })
        },
      }
    }
    return field
  })

  return {
    // 业务服务（包含 model 与增删改查）
    service: {
      ...serviceResult,
      // 覆盖 model，注入服务选择相关方法
      model: {
        ...serviceResult.model,
        serviceFormConfig: {
          ...serviceResult.model.serviceFormConfig,
          fields: updatedFields,
        },
      },
    },

    // 对话框状态
    formDialogVisible,
    formDialogMode,
    currentEditService,
    showNodeDialog,
    currentServiceId,

    // 服务发现选择相关状态
    selectedService,
    serviceSelectionVisible,

    // 方法
    openAddDialog,
    openEditDialog,
    openCopyDialog,
    openNodeDialog,
    closeFormDialog,
    handleFormSubmit,
    handleDelete,
    handleBatchDelete,
    handleSearch,
    handleToolbarClick,
    handleMenuClick,
    refresh,
    loadServiceList,
    handleServiceSelect,
    updateServiceMetadata,
    openServiceSelection,
    getLoadBalanceStrategyLabel,
  }
}

/**
 * 服务定义列表页面级 Hook 类型
 */
export type ServiceDefinitionPage = ReturnType<typeof useServiceDefinitionPage>

