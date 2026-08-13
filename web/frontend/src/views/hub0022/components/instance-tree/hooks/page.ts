/**
 * 网关实例树组件 Page Hook
 * 处理页面交互、事件处理
 */

import { useAppMessage } from '@/composables/useAppMessage'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import type { RsTreeNode } from '@/ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { addProxyConfig, editProxyConfig, getProxyConfigByInstance } from '../../../api'
import type { GatewayInstance, InstanceTreeNode, ProxyConfig } from '../types'
import { useGatewayInstanceTreeModel } from './model'
import { useGatewayInstanceTreeService } from './service'

/**
 * 网关实例树 Page Hook
 */
export function useGatewayInstanceTreePage() {
  const message = useAppMessage()

  // 初始化 Model
  const model = useGatewayInstanceTreeModel()

  // 初始化 Service
  const service = useGatewayInstanceTreeService(model)

  // ============= 代理配置对话框状态 =============

  /** 代理配置对话框显示状态 */
  const proxyFormDialogVisible = ref(false)
  /** 代理配置对话框模式 */
  const proxyFormDialogMode = ref<'create' | 'edit' | 'view'>('create')
  /** 当前编辑的代理配置 */
  const currentEditProxy = ref<ProxyConfig | null>(null)
  /** 提交状态 */
  const proxySubmitting = ref(false)
  /** 当前选中的网关实例ID（用于新增代理配置时使用） */
  const currentGatewayInstanceId = ref<string>('')
  /** 右键菜单关联的树节点 */
  const contextNode = ref<InstanceTreeNode | null>(null)

  // ============= 监听器 =============

  // 监听过滤关键词变化，重置到第一页并重新加载数据
  const stopFilterKeywordWatch = watch(model.filterKeyword, () => {
    model.resetPage()
    service.loadGatewayInstances()
  })

  // 组件卸载时清理监听器
  onBeforeUnmount(() => {
    stopFilterKeywordWatch()
  })

  // ============= JSON 字段转换方法 =============

  /**
   * 将配置对象转换为表单数据格式（查询时：将 JSON 字符串字段解析为点号分隔字段）
   */
  const convertProxyToFormData = (proxy: ProxyConfig): any => {
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
      ...proxy,
    }

    // 解析 proxyConfig（JSON 字符串 -> 对象 -> 点号分隔字段）
    const proxyConfigObj = parseJson(proxy.proxyConfig)
    if (proxyConfigObj && typeof proxyConfigObj === 'object' && !Array.isArray(proxyConfigObj)) {
      // 将对象的属性展开为点号分隔字段（如 proxyConfig.timeout）
      Object.keys(proxyConfigObj).forEach((key) => {
        formData[`proxyConfig.${key}`] = proxyConfigObj[key]
      })
      // 移除原始字段（已展开为点号分隔字段）
      delete formData.proxyConfig
    }

    // 解析 customConfig（JSON 字符串 -> 格式化的 JSON 字符串，用于 textarea 显示）
    const customConfigObj = parseJson(proxy.customConfig)
    if (customConfigObj && typeof customConfigObj === 'object') {
      // textarea 需要字符串，所以格式化为 JSON 字符串（带缩进，方便编辑）
      try {
        formData.customConfig = JSON.stringify(customConfigObj, null, 2)
      } catch {
        formData.customConfig = '{}'
      }
    } else {
      formData.customConfig = '{}'
    }

    return formData
  }

  /**
   * 将表单数据转换为API数据格式（保存时：将点号分隔字段合并为 JSON 字符串）
   */
  const convertProxyToApiData = (formData: Record<string, any>): any => {
    // 构建 API 数据对象
    const apiData: Record<string, any> = {}

    // 复制基本字段
    Object.keys(formData).forEach((key) => {
      if (!key.startsWith('proxyConfig.')) {
        apiData[key] = formData[key]
      }
    })

    // 将点号分隔字段合并为 JSON 字符串
    const proxyConfigObj: Record<string, any> = {}
    Object.keys(formData).forEach((key) => {
      if (key.startsWith('proxyConfig.')) {
        const subKey = key.replace('proxyConfig.', '')
        proxyConfigObj[subKey] = formData[key]
      }
    })

    // 将 proxyConfig 对象转换为 JSON 字符串
    if (Object.keys(proxyConfigObj).length > 0) {
      apiData.proxyConfig = JSON.stringify(proxyConfigObj)
    } else {
      apiData.proxyConfig = '{}'
    }

    // 将 customConfig 转换为 JSON 字符串
    // formData.customConfig 可能是字符串（来自 textarea）或对象
    if (formData.customConfig) {
      if (typeof formData.customConfig === 'string') {
        // 如果是字符串，验证是否为有效 JSON，然后直接使用
        try {
          // 验证 JSON 格式
          JSON.parse(formData.customConfig)
          apiData.customConfig = formData.customConfig
        } catch {
          // JSON 格式无效，使用空对象
          apiData.customConfig = '{}'
        }
      } else if (typeof formData.customConfig === 'object') {
        // 如果是对象，转换为 JSON 字符串
        apiData.customConfig = JSON.stringify(formData.customConfig)
      } else {
        apiData.customConfig = '{}'
      }
    } else {
      apiData.customConfig = '{}'
    }

    return apiData
  }

  // ============= 代理配置对话框操作 =============

  /**
   * 加载代理配置（根据 gatewayInstanceId 查询，取第一个配置）
   * 参考认证配置逻辑：如果配置存在则进入编辑模式，不存在则进入新增模式
   */
  const loadProxyConfig = async () => {
    const gatewayInstanceId = currentGatewayInstanceId.value

    if (!gatewayInstanceId) {
      return
    }

    try {
      const response = await getProxyConfigByInstance(gatewayInstanceId)

      if (isApiSuccess(response)) {
        // 后端现在返回单个 ProxyConfig 对象或 null
        const proxyConfig = parseJsonData<ProxyConfig | null>(response, null)

        if (proxyConfig) {
          // 如果找到配置，进入编辑模式
          currentEditProxy.value = proxyConfig
          proxyFormDialogMode.value = 'edit'
        } else {
          // 如果没有配置，进入新增模式
          currentEditProxy.value = null
          proxyFormDialogMode.value = 'create'
        }
      } else {
        // API 调用失败，默认进入新增模式
        currentEditProxy.value = null
        proxyFormDialogMode.value = 'create'
      }
    } catch {
      // 异常情况，默认进入新增模式
      currentEditProxy.value = null
      proxyFormDialogMode.value = 'create'
    }
  }

  /**
   * 打开代理配置对话框
   * 参考认证配置逻辑：先加载配置，如果存在则编辑，不存在则新增
   */
  const openProxyDialog = async (instance: GatewayInstance) => {
    currentGatewayInstanceId.value = instance.gatewayInstanceId
    // 先加载配置数据，然后再打开对话框
    await loadProxyConfig()
    // 数据加载完成后再打开对话框
    proxyFormDialogVisible.value = true
  }

  /**
   * 获取代理配置表单初始数据
   */
  const getProxyFormInitialData = (): any => {
    if (proxyFormDialogMode.value === 'create') {
      // 新增模式：使用默认值
      return {
        gatewayInstanceId: currentGatewayInstanceId.value,
        proxyType: 'http',
        configPriority: 100,
        activeFlag: 'Y',
      }
    } else if (currentEditProxy.value) {
      // 编辑/查看模式：使用转换后的数据
      return convertProxyToFormData(currentEditProxy.value)
    }
    return undefined
  }

  /**
   * 提交代理配置表单
   */
  const handleProxyFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return false

    // 查看模式下不执行提交
    if (proxyFormDialogMode.value === 'view') {
      return false
    }

    try {
      proxySubmitting.value = true

      // 使用 convertProxyToApiData 将点号分隔字段合并为 JSON 字符串
      const apiData = convertProxyToApiData(formData)

      // 准备提交数据
      const processedData: Partial<ProxyConfig> & { tenantId: string } = {
        ...apiData,
        tenantId: 'default',
        gatewayInstanceId: currentGatewayInstanceId.value || formData.gatewayInstanceId,
      }

      let success = false
      if (proxyFormDialogMode.value === 'create') {
        const response = await addProxyConfig(processedData as any)
        if (isApiSuccess(response)) {
          message.success(getApiMessage(response, '创建代理配置成功'))
          success = true
        } else {
          message.error(getApiMessage(response, '创建代理配置失败'))
        }
      } else if (proxyFormDialogMode.value === 'edit' && currentEditProxy.value) {
        const response = await editProxyConfig({
          ...processedData,
          proxyConfigId: currentEditProxy.value.proxyConfigId,
        } as any)
        if (isApiSuccess(response)) {
          message.success(getApiMessage(response, '更新代理配置成功'))
          success = true
        } else {
          message.error(getApiMessage(response, '更新代理配置失败'))
        }
      }

      if (success) {
        // 保存成功后，重新加载配置数据（因为新增后应该进入编辑模式）
        await loadProxyConfig()
        // 保持对话框打开，让用户继续编辑
      }
      return success
    } catch {
      message.error('操作失败')
      return false
    } finally {
      proxySubmitting.value = false
    }
  }

  // ============= 事件处理 =============

  /**
   * 记录右键菜单关联节点
   */
  function setContextNode(node: RsTreeNode) {
    contextNode.value = node as InstanceTreeNode
  }

  /**
   * 处理树节点点击选择
   */
  function handleNodeClick(
    node: RsTreeNode,
    key: string,
    emit: (e: 'select', instanceId: string, instance: GatewayInstance) => void,
  ) {
    const instanceNode = node as InstanceTreeNode
    if (instanceNode.instance) {
      emit('select', key, instanceNode.instance)
    }
  }

  /**
   * 处理分页变化（后端分页，需要重新加载数据）
   */
  function handlePageChange({ currentPage: page, pageSize: size }: { currentPage: number; pageSize: number }) {
    model.setCurrentPage(page)
    model.setPageSize(size)
    // 后端分页，需要重新加载数据
    service.loadGatewayInstances()
  }

  /**
   * 处理刷新
   */
  function handleRefresh() {
    service.loadGatewayInstances()
  }

  /**
   * 处理右键菜单点击
   */
  async function handleContextMenuSelect(key: string) {
    const instance = contextNode.value?.instance
    if (!instance) return

    switch (key) {
      case 'addProxy':
        await openProxyDialog(instance)
        break
      default:
        break
    }
  }

  return {
    // Model
    model,

    // Service
    service,

    // 代理配置对话框状态
    proxyFormDialogVisible,
    proxyFormDialogMode,
    currentEditProxy,
    proxySubmitting,
    getProxyFormInitialData,

    // 事件处理
    setContextNode,
    handleNodeClick,
    handlePageChange,
    handleRefresh,
    handleContextMenuSelect,
    handleProxyFormSubmit,
  }
}

/**
 * 网关实例树 Page 类型
 */
export type GatewayInstanceTreePage = ReturnType<typeof useGatewayInstanceTreePage>
