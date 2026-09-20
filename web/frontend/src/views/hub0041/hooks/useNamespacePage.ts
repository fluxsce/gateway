/**
 * 命名空间管理页面级 Hook
 * - 组合 useNamespaceService（纯业务逻辑）
 * - 处理新增对话框、工具栏、卡片右键菜单等页面交互
 */

import { useAppMessage } from '@/composables/useAppMessage'
import { isApiSuccess, parseJsonData } from '@/utils/format'
import type { Ref } from 'vue'
import { nextTick, ref, watch } from 'vue'
import type { CenterOverview } from '../../hub0040/types'
import * as namespaceApi from '../api'
import type { Namespace } from '../types'
import { canNamespaceAction } from './model'
import { useNamespaceService } from './useNamespaceService'

export interface NamespaceQueryScope {
  instanceName: string
  environment: string
}

function namespaceIdentity(namespace: Namespace) {
  return `${namespace.tenantId}:${namespace.namespaceId}`
}

/**
 * 命名空间管理页面级 Hook
 * @param searchFormRef 搜索表单引用
 * @param moduleId 自定义模块ID，用于支持同一页面多个实例（默认 'hub0041'）
 */
export function useNamespacePage(searchFormRef?: Ref<any> | any, moduleId?: string) {
  const message = useAppMessage()
  const service = useNamespaceService(searchFormRef, moduleId)

  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditNamespace = ref<Namespace | null>(null)
  const submitting = ref(false)
  const selectedNamespace = ref<Namespace | null>(null)
  const queryScope = ref<NamespaceQueryScope | null>(null)
  const overview = ref<CenterOverview | null>(null)
  const overviewOnline = ref(false)
  const overviewLoading = ref(false)

  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditNamespace.value = null
    formDialogVisible.value = true
  }

  const openEditDialog = async (namespace: Namespace) => {
    try {
      const detailNamespace = await service.getNamespaceDetail(namespace.namespaceId)

      if (!detailNamespace) {
        message.error('获取命名空间详情失败')
        return
      }

      formDialogMode.value = 'edit'
      currentEditNamespace.value = detailNamespace
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取命名空间详情失败')
    }
  }

  const openViewDialog = async (namespace: Namespace) => {
    try {
      const detailNamespace = await service.getNamespaceDetail(namespace.namespaceId)

      if (!detailNamespace) {
        message.error('获取命名空间详情失败')
        return
      }

      formDialogMode.value = 'view'
      currentEditNamespace.value = detailNamespace
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取命名空间详情失败')
    }
  }

  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditNamespace.value = null
  }

  const handleSubmit = async (formData: Namespace) => {
    submitting.value = true
    try {
      let success = false
      if (formDialogMode.value === 'create') {
        success = await service.addNamespace(formData)
      } else if (formDialogMode.value === 'edit') {
        success = await service.editNamespace({
          ...formData,
          namespaceId: currentEditNamespace.value!.namespaceId,
        })
      }

      if (success) {
        closeFormDialog()
        await refreshMonitor()
      }
    } finally {
      submitting.value = false
    }
  }

  const handleToolbarClick = async (key: string) => {
    if (!canNamespaceAction(key) && key !== 'search' && key !== 'reset') {
      message.warning('没有操作权限')
      return
    }
    switch (key) {
      case 'add':
        openAddDialog()
        break
      case 'edit': {
        if (!selectedNamespace.value) {
          message.warning('请先选择要编辑的命名空间')
          return
        }
        openEditDialog(selectedNamespace.value)
        break
      }
      case 'delete': {
        if (!selectedNamespace.value) {
          message.warning('请先选择要删除的命名空间')
          return
        }
        if (await service.deleteNamespace(selectedNamespace.value)) {
          await loadOverview(queryScope.value)
        }
        break
      }
    }
  }

  const handleMenuClick = async (params: { key: string; row?: any }) => {
    if (!params.row) {
      return
    }
    if (!canNamespaceAction(params.key)) {
      message.warning('没有操作权限')
      return
    }
    const row = params.row as Namespace
    switch (params.key) {
      case 'view':
        await openViewDialog(row)
        break
      case 'edit':
        await openEditDialog(row)
        break
      case 'delete':
        if (await service.deleteNamespace(row)) {
          await refreshMonitor()
        }
        break
    }
  }

  const selectNamespace = (namespace: Namespace) => {
    if (selectedNamespace.value && namespaceIdentity(selectedNamespace.value) === namespaceIdentity(namespace)) {
      selectedNamespace.value = null
      void loadOverview(queryScope.value)
      return
    }
    selectedNamespace.value = namespace
    void loadOverview(queryScope.value, namespace.namespaceId)
  }

  const focusNamespace = (namespace: Namespace) => {
    if (selectedNamespace.value && namespaceIdentity(selectedNamespace.value) === namespaceIdentity(namespace)) {
      return
    }
    selectedNamespace.value = namespace
  }

  const handleCardAction = async (key: string, namespace: Namespace) => {
    await nextTick()
    selectedNamespace.value = namespace
    void loadOverview(queryScope.value, namespace.namespaceId)
    await handleMenuClick({ key, row: namespace })
  }

  const readScope = (formData?: Record<string, any>): NamespaceQueryScope | null => {
    const data = formData || searchFormRef?.value?.getFormData?.() || {}
    const instanceName = String(data.instanceName || '').trim()
    if (!instanceName) return null
    return {
      instanceName,
      environment: String(data.environment || '').trim(),
    }
  }

  const loadOverview = async (scope: NamespaceQueryScope | null, namespaceId?: string) => {
    if (!scope?.instanceName) {
      overview.value = null
      overviewOnline.value = false
      return
    }
    overviewLoading.value = true
    try {
      const response = await namespaceApi.queryNamespaceOverview({
        instanceName: scope.instanceName,
        environment: scope.environment || undefined,
        namespaceId: namespaceId || undefined,
      })
      if (!isApiSuccess(response)) {
        overview.value = null
        overviewOnline.value = false
        return
      }
      const data = parseJsonData<(CenterOverview & { online?: boolean }) | null>(response, null)
      overview.value = data
      overviewOnline.value = Boolean(data?.online)
    } catch {
      overview.value = null
      overviewOnline.value = false
    } finally {
      overviewLoading.value = false
    }
  }

  watch(
    () => service.model.namespaceList.value,
    (list) => {
      const current = selectedNamespace.value
      if (!current) return
      const next = list.find((row) => namespaceIdentity(row) === namespaceIdentity(current))
      selectedNamespace.value = next || null
      if (!next && queryScope.value) {
        void loadOverview(queryScope.value)
      }
    },
  )

  const refreshMonitor = async () => {
    const scope = queryScope.value || readScope()
    await service.handleRefresh()
    await loadOverview(scope, selectedNamespace.value?.namespaceId)
  }

  const handleSearch = async (formData?: Record<string, any>) => {
    queryScope.value = readScope(formData)
    await service.handleSearch(formData)
    await loadOverview(queryScope.value, selectedNamespace.value?.namespaceId)
  }

  const handleReset = async () => {
    queryScope.value = null
    selectedNamespace.value = null
    overview.value = null
    overviewOnline.value = false
    await service.handleReset()
  }

  const handleFormSubmit = (formData?: Record<string, any>) => {
    if (formData) {
      handleSubmit(formData as any)
    }
  }

  return {
    service,

    formDialogVisible,
    formDialogMode,
    currentEditNamespace,
    submitting,
    selectedNamespace,
    queryScope,
    overview,
    overviewOnline,
    overviewLoading,

    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
    handleSubmit,
    handleFormSubmit,

    selectNamespace,
    focusNamespace,
    handleToolbarClick,
    handleMenuClick,
    handleCardAction,
    handleSearch,
    handleReset,
    refreshMonitor,
  }
}

export type NamespacePage = ReturnType<typeof useNamespacePage>
