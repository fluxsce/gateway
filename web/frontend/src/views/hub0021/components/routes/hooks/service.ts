/**
 * 路由配置列表业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess } from '@/utils/format'
import { WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { addRouteConfig, deleteRouteConfig, editRouteConfig, queryRouteConfigs } from '../../../api'
import type { RouteConfig } from '../types'
import { useRouteConfigModel } from './model'

/**
 * 路由配置列表服务 Hook（纯业务逻辑）
 * @param gatewayInstanceId 网关实例ID
 */
export function useRouteConfigService(gatewayInstanceId?: string, searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model
  const model = useRouteConfigModel()

  const {
    loading,
    routeList,
    pageInfo,
    setRouteList,
    updatePagination,
    addRouteToList,
    updateRouteInList,
    removeRouteFromList,
    removeRoutesFromList,
  } = model

  // ============= 数据加载 =============

  /**
   * 加载路由配置列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadRouteList = async (searchParams?: Record<string, any>) => {
    loading.value = true
    try {
      // 如果没有传入查询参数，从搜索表单获取
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      // 过滤掉空字符串、null 和 undefined 的查询条件
      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(
            Object.entries(finalSearchParams).filter(
              ([, value]) => value !== '' && value !== null && value !== undefined
            )
          )
        : {}

      // 构建请求参数：合并查询条件和分页参数
      const params = {
        // 查询条件
        ...effectiveSearchParams,
        // 如果 searchParams 中没有 gatewayInstanceId，且构造函数参数有，则使用构造函数参数的
        ...(effectiveSearchParams.gatewayInstanceId === undefined && gatewayInstanceId ? { gatewayInstanceId } : {}),
        // 分页参数
        ...createBackendPaginationParams(pageInfo.value?.pageIndex, pageInfo.value?.pageSize),
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await queryRouteConfigs(params as any)

      if (response.oK) {
        // 解析业务数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const routes = Array.isArray(bizData) ? bizData : []
          setRouteList(routes)
        }

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询路由配置列表失败')
      }
    } catch (error) {
      message.error('加载路由配置列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索路由配置
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    // 重置到第一页
    updatePagination({ pageIndex: 1, pageSize: pageInfo.value?.pageSize || 20 })
    await loadRouteList(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    updatePagination({ pageIndex: 1, pageSize: pageInfo.value?.pageSize || 20 })
    await loadRouteList()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    await loadRouteList()
  }

  // ============= 数据操作 =============

  /**
   * 删除路由配置
   */
  const deleteRoute = async (routeConfigId: string): Promise<boolean> => {
    // 查找要删除的路由配置信息，用于确认对话框
    const routeToDelete = routeList.value.find((route) => route.routeConfigId === routeConfigId)
    if (!routeToDelete) {
      message.error('未找到要删除的路由配置')
      return false
    }

    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除路由配置 "${routeToDelete.routeName || routeToDelete.routeConfigId}" 吗？`,
      icon: WarningOutline,
      headerStyle: 'gradient',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500,
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response = await deleteRouteConfig(routeConfigId)
      if (response.oK) {
        message.success(getApiMessage(response, '删除路由配置成功'))
        removeRouteFromList(routeConfigId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (routeList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadRouteList()
        } else {
          await loadRouteList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '删除路由配置失败'))
        return false
      }
    } catch (error) {
      message.error('删除路由配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 批量删除路由配置
   */
  const batchDeleteRoutes = async (routeConfigIds: string[]) => {
    const confirmed = await gDialog.warning({
      title: '确认批量删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除选中的 ${routeConfigIds.length} 个路由配置吗？`,
      icon: WarningOutline,
      headerStyle: 'gradient',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return
    }

    loading.value = true
    try {
      let successCount = 0
      let failCount = 0

      // 逐个删除
      for (const id of routeConfigIds) {
        try {
          const response = await deleteRouteConfig(id)
          if (response.oK) {
            successCount++
          } else {
            failCount++
          }
        } catch {
          failCount++
        }
      }

      // 显示结果
      if (successCount > 0) {
        message.success(`成功删除 ${successCount} 个路由配置${failCount > 0 ? `，失败 ${failCount} 个` : ''}`)
        removeRoutesFromList(routeConfigIds.slice(0, successCount))
        
        // 重新加载列表
        await loadRouteList()
      } else {
        message.error(`删除失败，共 ${failCount} 个`)
      }
    } catch (error) {
      message.error('批量删除失败')
    } finally {
      loading.value = false
    }
  }

  /**
   * 切换路由配置状态
   */
  const toggleRouteStatus = async (route: RouteConfig) => {
    const newStatus = route.activeFlag === 'Y' ? 'N' : 'Y'
    const action = newStatus === 'Y' ? '启用' : '禁用'

    try {
      const response = await editRouteConfig({
        routeConfigId: route.routeConfigId,
        activeFlag: newStatus,
      } as any)

      if (response.oK) {
        message.success(`${action}成功`)
        updateRouteInList({ ...route, activeFlag: newStatus })
        await loadRouteList()
      } else {
        message.error(response.errMsg || `${action}失败`)
      }
    } catch (error) {
      message.error(`${action}失败`)
    }
  }

  /**
   * 批量切换路由配置状态
   */
  const batchToggleRouteStatus = async (routeConfigIds: string[], status: 'Y' | 'N') => {
    const action = status === 'Y' ? '启用' : '禁用'

    try {
      const updatePromises = routeConfigIds.map((id) => {
        const route = routeList.value.find((r) => r.routeConfigId === id)
        if (!route) return Promise.resolve({ oK: false })
        return editRouteConfig({
          routeConfigId: id,
          activeFlag: status,
        } as any)
      })
      const responses = await Promise.all(updatePromises)

      const successCount = responses.filter((response: any) => response.oK).length
      if (successCount === routeConfigIds.length) {
        message.success(`成功${action} ${successCount} 个路由配置`)
      } else {
        message.warning(`${action}了 ${successCount}/${routeConfigIds.length} 个路由配置`)
      }

      await loadRouteList()
    } catch (error) {
      message.error(`批量${action}失败`)
    }
  }

  /**
   * 添加路由配置
   */
  const addRoute = async (routeData: Partial<RouteConfig>): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await addRouteConfig(routeData as any)

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '路由配置创建成功')
        message.success(successMsg)

        // 如果返回了新增的路由数据，添加到列表
        if (response.bizData) {
          const newRoute = JSON.parse(response.bizData)
          addRouteToList(newRoute)
        } else {
          // 否则重新加载列表
          await loadRouteList()
        }

        return true
      } else {
        const errorMsg = getApiMessage(response, '新增路由配置失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('新增路由配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 编辑路由配置
   */
  const editRoute = async (
    routeConfigId: string,
    routeData: Partial<RouteConfig>
  ): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await editRouteConfig({
        ...routeData,
        routeConfigId,
      } as any)

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '路由配置更新成功')
        message.success(successMsg)

        // 如果返回了更新的路由数据，更新列表中的对应项
        if (response.bizData) {
          const updatedRoute = JSON.parse(response.bizData)
          updateRouteInList(updatedRoute)
        } else {
          // 否则重新加载列表
          await loadRouteList()
        }

        return true
      } else {
        const errorMsg = getApiMessage(response, '更新路由配置失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('更新路由配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    // Model 实例（包含所有状态和配置）
    model,

    // 数据加载
    loadRouteList,

    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,

    // 路由配置操作
    addRoute,
    editRoute,
    deleteRoute,
    batchDeleteRoutes,
    toggleRouteStatus,
    batchToggleRouteStatus,
  }
}

/**
 * 路由配置列表服务类型
 */
export type RouteConfigService = ReturnType<typeof useRouteConfigService>

