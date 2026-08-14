import { getApiMessage, isApiSuccess } from '@/utils/format'
import { useAppMessage } from '@/composables/useAppMessage'
import type { RsSelectOption } from '@/ui'
import { computed, ref, watch } from 'vue'
import { addRouteConfig, editRouteConfig, queryServiceDefinitions } from '../api'
import type { RouteConfig } from '../types'
import { useRouteForm } from './useRouteForm'

export interface UseRouteConfigDialogOptions {
  onSuccess?: (route?: RouteConfig) => void
}

/**
 * 路由配置对话框Hook
 * 专注于路由的基本创建和编辑功能
 */
export function useRouteConfigDialog(options: UseRouteConfigDialogOptions = {}) {
  const message = useAppMessage()

  // 使用其他hooks
  const {
    formRef,
    formData,
    formRules,
    isEditMode,
    editingRouteId,
    httpMethodOptions,
    matchTypeOptions,
    getPathExample,
    getMatchTypeDescription,
    resetForm,
    fillFormData,
    validateForm,
    getFormData,
    handleMatchTypeChange,
    handlePathInput,
  } = useRouteForm()

  // 独立的提交状态管理
  const dialogSubmitting = ref(false)

  // 对话框状态
  const visible = ref(false)
  const editingRoute = ref<RouteConfig | null>(null)
  const gatewayInstanceId = ref('')

  // 元数据列表
  const metadataList = ref<{ key: string; value: string }[]>([])

  // 服务定义相关状态
  const serviceDefinitionOptions = ref<RsSelectOption[]>([])
  const loadingServiceDefinitions = ref(false)

  // Switch组件的计算属性
  const activeSwitch = computed({
    get: () => formData.activeFlag,
    set: (value: 'Y' | 'N') => {
      formData.activeFlag = value
    },
  })

  /**
   * 加载服务定义列表
   */
  const loadServiceDefinitions = async (instanceId: string) => {
    if (!instanceId) {
      serviceDefinitionOptions.value = []
      return
    }

    try {
      loadingServiceDefinitions.value = true
      const response = await queryServiceDefinitions({
        gatewayInstanceId: instanceId,
        pageIndex: 1,
        pageSize: 1000, // 加载所有服务定义
      })

      if (isApiSuccess(response)) {
        const pageData = JSON.parse(response.bizData)
        const serviceDefinitions = pageData?.list || pageData || []
        serviceDefinitionOptions.value = serviceDefinitions.map((service: any) => ({
          label: `${service.serviceName} (${service.serviceDefinitionId})`,
          value: service.serviceDefinitionId,
          // 可以保存更多信息用于显示
          disabled: service.activeFlag !== 'Y',
        }))
      } else {
        serviceDefinitionOptions.value = []
        console.warn('获取服务定义列表失败:', getApiMessage(response, '获取服务定义列表失败'))
      }
    } catch (error) {
      console.error('加载服务定义列表失败:', error)
      serviceDefinitionOptions.value = []
      message.error('加载服务定义列表失败')
    } finally {
      loadingServiceDefinitions.value = false
    }
  }

  /**
   * 处理服务定义选择变化
   */
  const handleServiceDefinitionChange = (value: string | null) => {
    // 可以在这里添加其他逻辑，比如根据选择的服务定义更新其他字段
    console.log('选择的服务定义ID:', value)
  }

  /**
   * 创建元数据项
   */
  const createMetadataItem = () => ({
    key: '',
    value: '',
  })

  /**
   * 打开对话框
   */
  const openDialog = (route?: RouteConfig, instanceId?: string) => {
    visible.value = true
    editingRoute.value = route || null
    gatewayInstanceId.value = instanceId || ''

    if (route) {
      fillFormData(route)
      // 填充元数据
      const metadata = route.routeMetadata || {}
      metadataList.value = Object.entries(metadata).map(([key, value]) => ({
        key,
        value: String(value),
      }))
    } else {
      resetForm()
      metadataList.value = []
    }

    // 加载服务定义列表
    if (instanceId) {
      loadServiceDefinitions(instanceId)
    }

    // 确保在下一个tick清除验证状态，让表单重新验证
    setTimeout(() => {
      formRef.value?.clearValidation()
    }, 100)
  }

  /**
   * 关闭对话框
   */
  const closeDialog = () => {
    visible.value = false
    editingRoute.value = null
    gatewayInstanceId.value = ''
    resetForm()
    metadataList.value = []
    serviceDefinitionOptions.value = []
  }

  /**
   * 处理提交
   */
  const handleSubmit = async () => {
    try {
      // 先验证表单
      const isValid = await validateForm()
      if (!isValid) {
        message.warning('请检查表单输入')
        return
      }

      // 处理元数据
      const metadata: Record<string, any> = {}
      metadataList.value.forEach((item) => {
        if (item.key.trim() && item.value.trim()) {
          metadata[item.key.trim()] = item.value.trim()
        }
      })
      formData.routeMetadata = metadata

      const formDataToSubmit = getFormData()

      dialogSubmitting.value = true

      if (isEditMode.value) {
        // 编辑路由
        const editData = {
          ...formDataToSubmit,
          routeConfigId: editingRouteId.value,
          gatewayInstanceId: gatewayInstanceId.value,
        }

        console.log('Updating route with data:', editData)
        const response = await editRouteConfig(editData)

        if (response.oK) {
          message.success('路由更新成功')

          // 解析返回的更新后路由数据
          let updatedRoute = null
          try {
            if (response.bizData) {
              const routeData = JSON.parse(response.bizData)
              // 后端返回的是单个对象，不是数组
              if (routeData && typeof routeData === 'object') {
                updatedRoute = routeData
                console.log('解析到更新后的路由:', updatedRoute)
              }
            }
          } catch (parseError) {
            console.error('解析更新后的路由数据失败:', parseError, 'bizData:', response.bizData)
          }

          closeDialog()
          // 传递更新后的路由数据
          options.onSuccess?.(updatedRoute)
        } else {
          message.error(response.errMsg || response.popMsg || '路由更新失败')
        }
      } else {
        // 创建路由
        const createData = {
          ...formDataToSubmit,
          gatewayInstanceId: gatewayInstanceId.value,
        }

        console.log('Creating route with data:', createData)
        const response = await addRouteConfig(createData)

        if (response.oK) {
          message.success('路由创建成功！您可以通过"路由配置管理"功能配置高级选项')

          // 解析返回的路由数据
          let newRoute = null
          try {
            if (response.bizData) {
              const routeData = JSON.parse(response.bizData)
              // 后端返回的是单个对象，不是数组
              if (routeData && typeof routeData === 'object') {
                newRoute = routeData
                console.log('解析到新创建的路由:', newRoute)
              } else {
                console.warn('返回的路由数据格式不正确:', routeData)
              }
            } else {
              console.warn('后端未返回路由数据')
            }
          } catch (parseError) {
            console.error('解析返回的路由数据失败:', parseError, 'bizData:', response.bizData)
          }

          closeDialog()
          // 传递新创建的路由数据
          options.onSuccess?.(newRoute)
        } else {
          message.error(response.errMsg || response.popMsg || '路由创建失败')
        }
      }
    } catch (error: any) {
      console.error('路由操作失败:', error)
      message.error(error.message || '操作失败，请重试')
    } finally {
      dialogSubmitting.value = false
    }
  }

  // 监听对话框可见性变化
  watch(visible, (newVisible) => {
    if (!newVisible) {
      // 对话框关闭时清理状态
      setTimeout(() => {
        editingRoute.value = null
        gatewayInstanceId.value = ''
      }, 200)
    }
  })

  return {
    // 对话框状态
    visible,
    editingRoute,
    gatewayInstanceId,

    // 表单相关
    formRef,
    formData,
    formRules,
    isEditMode,
    editingRouteId,

    // 选项数据
    httpMethodOptions,
    matchTypeOptions,

    // 计算属性
    getPathExample,
    getMatchTypeDescription,
    activeSwitch,

    // 元数据
    metadataList,
    createMetadataItem,

    // 服务定义相关
    serviceDefinitionOptions,
    loadingServiceDefinitions,
    handleServiceDefinitionChange,

    // 状态
    submitting: dialogSubmitting,

    // 方法
    openDialog,
    closeDialog,
    handleSubmit,
    handleMatchTypeChange,
    handlePathInput,
  }
}
