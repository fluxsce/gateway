/**
 * 过滤器配置列表服务层 Hook
 * 纯业务逻辑：数据获取、增删改查等操作
 */

import { createBackendPaginationParams } from '@/components/gpage'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import {
  addFilterConfig,
  deleteFilterConfig,
  editFilterConfig,
  getFilterConfig,
  queryFilterConfigs,
} from '../../../api'
import { useFilterConfigModel } from './model'
import type { FilterConfig } from './types'

/**
 * 过滤器配置列表服务层 Hook
 * @param moduleId 模块ID（用于权限控制，必填）
 * @param gatewayInstanceId 网关实例ID（可选，用于全局过滤器）
 * @param routeConfigId 路由配置ID（可选，用于路由过滤器）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useFilterConfigService(
  moduleId: string,
  gatewayInstanceId?: Ref<string | undefined> | string,
  routeConfigId?: Ref<string | undefined> | string,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()

  // 获取 gatewayInstanceId 和 routeConfigId 的实际值
  const getGatewayInstanceId = () => {
    return typeof gatewayInstanceId === 'string' ? gatewayInstanceId : gatewayInstanceId?.value
  }

  const getRouteConfigId = () => {
    return typeof routeConfigId === 'string' ? routeConfigId : routeConfigId?.value
  }

  // 使用 model（传入模块ID）
  const model = useFilterConfigModel(moduleId)
  
  // 解构 model 中的列表操作方法
  const {
    addFilterToList,
    updateFilterInList,
    removeFilterFromList,
    removeFiltersFromList,
  } = model

  /**
   * 加载过滤器列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadFilterList = async (searchParams?: Record<string, any>) => {
    try {
      model.setLoading(true)

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

      // 获取 gatewayInstanceId 和 routeConfigId
      const instanceId = getGatewayInstanceId()
      const routeId = getRouteConfigId()

      // 构建请求参数：合并查询条件、实例/路由ID和分页参数
      // 统一使用 queryFilterConfigs 方法，因为它支持分页
      const queryParams: any = {
        // 查询条件
        ...effectiveSearchParams,
        // 实例或路由ID（如果存在）
        ...(instanceId ? { gatewayInstanceId: instanceId } : {}),
        ...(routeId ? { routeConfigId: routeId } : {}),
        // 分页参数
        ...createBackendPaginationParams(
          model.pageInfo.value?.pageIndex,
          model.pageInfo.value?.pageSize
        ),
      }

      // 统一使用 queryFilterConfigs 方法（支持分页）
      const response = await queryFilterConfigs(queryParams)

      if (isApiSuccess(response)) {
        // 解析业务数据
        const filters = parseJsonData<FilterConfig[]>(response, []) || []
        
        // 按 filterOrder 排序
        if (Array.isArray(filters) && filters.length > 0) {
          filters.sort((a: FilterConfig, b: FilterConfig) => (a.filterOrder || 0) - (b.filterOrder || 0))
        }
        
        model.setFilterList(filters)

        // 解析分页信息
        const backendPageInfo = parsePageInfo(response)
        if (backendPageInfo && Object.keys(backendPageInfo).length > 0) {
          model.updatePagination(backendPageInfo)
        } else {
          model.resetPagination()
        }
      } else {
        message.error(getApiMessage(response, '加载过滤器列表失败'))
        model.setFilterList([])
        model.resetPagination()
      }
    } catch (error: any) {
      console.error('加载过滤器列表失败:', error)
      message.error(error.message || '加载过滤器列表失败')
      model.setFilterList([])
      model.resetPagination()
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 获取过滤器详情
   */
  const getFilterDetail = async (filterConfigId: string): Promise<FilterConfig | null> => {
    try {
      const response = await getFilterConfig(filterConfigId)
      if (isApiSuccess(response)) {
        // GetFilterConfig 返回的是单条数据，不是数组
        return parseJsonData<FilterConfig>(response)
      } else {
        message.error(getApiMessage(response, '获取过滤器详情失败'))
        return null
      }
    } catch (error: any) {
      console.error('获取过滤器详情失败:', error)
      message.error(error.message || '获取过滤器详情失败')
      return null
    }
  }

  /**
   * 新增过滤器
   */
  const addFilter = async (filterData: Partial<FilterConfig>): Promise<boolean> => {
    try {
      model.setLoading(true)

      const submitData = {
        ...filterData,
        tenantId: 'default',
        gatewayInstanceId: getGatewayInstanceId(),
        routeConfigId: getRouteConfigId(),
      }

      const response = await addFilterConfig(submitData)
      if (isApiSuccess(response)) {
        message.success('新增过滤器成功')
        
        // 如果返回了新增的过滤器数据，添加到列表
        const newFilter = parseJsonData<FilterConfig | null>(response, null)
        if (newFilter) {
          addFilterToList(newFilter)
        } else {
          // 否则重新加载列表
          await loadFilterList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '新增过滤器失败'))
        return false
      }
    } catch (error: any) {
      console.error('新增过滤器失败:', error)
      message.error(error.message || '新增过滤器失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 编辑过滤器
   */
  const editFilter = async (filterConfigId: string, filterData: Partial<FilterConfig>): Promise<boolean> => {
    try {
      model.setLoading(true)

      const submitData = {
        ...filterData,
        filterConfigId,
        tenantId: 'default',
      }

      const response = await editFilterConfig(submitData)
      if (isApiSuccess(response)) {
        message.success('编辑过滤器成功')
        
        // 更新列表中的过滤器数据：必须以返回的 response.bizData 为准
        const updatedFilter = parseJsonData<FilterConfig | null>(response, null)
        if (updatedFilter) {
          updateFilterInList(updatedFilter.filterConfigId, updatedFilter.tenantId, updatedFilter)
        } else {
          // 如果后端没有返回数据，重新加载列表以确保数据一致性
          await loadFilterList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '编辑过滤器失败'))
        return false
      }
    } catch (error: any) {
      console.error('编辑过滤器失败:', error)
      message.error(error.message || '编辑过滤器失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 删除过滤器
   */
  const deleteFilter = async (filterConfigId: string): Promise<boolean> => {
    try {
      model.setLoading(true)

      const response = await deleteFilterConfig(filterConfigId)
      if (isApiSuccess(response)) {
        message.success('删除过滤器成功')
        removeFilterFromList(filterConfigId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (model.filterList.value.length === 0 && model.pageInfo.value && model.pageInfo.value.pageIndex > 1) {
          model.updatePagination({ pageIndex: model.pageInfo.value.pageIndex - 1 })
          await loadFilterList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '删除过滤器失败'))
        return false
      }
    } catch (error: any) {
      console.error('删除过滤器失败:', error)
      message.error(error.message || '删除过滤器失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 批量删除过滤器
   */
  const deleteFilters = async (filterConfigIds: string[]): Promise<boolean> => {
    try {
      model.setLoading(true)

      // 逐个删除
      const results = await Promise.all(
        filterConfigIds.map(id => deleteFilterConfig(id))
      )

      const successCount = results.filter(r => isApiSuccess(r)).length
      if (successCount === filterConfigIds.length) {
        message.success(`成功删除 ${successCount} 个过滤器`)
        removeFiltersFromList(filterConfigIds)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (model.filterList.value.length === 0 && model.pageInfo.value && model.pageInfo.value.pageIndex > 1) {
          model.updatePagination({ pageIndex: model.pageInfo.value.pageIndex - 1 })
          await loadFilterList()
        }
        
        return true
      } else {
        message.warning(`部分删除成功：${successCount}/${filterConfigIds.length}`)
        // 部分删除成功时，重新加载列表以确保数据一致性
        await loadFilterList()
        return false
      }
    } catch (error: any) {
      console.error('批量删除过滤器失败:', error)
      message.error(error.message || '批量删除过滤器失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 切换过滤器状态
   */
  const toggleFilterStatus = async (filter: FilterConfig): Promise<boolean> => {
    const newStatus = filter.activeFlag === 'Y' ? 'N' : 'Y'
    return await editFilter(filter.filterConfigId, {
      ...filter,
      activeFlag: newStatus,
    })
  }

  return {
    model,
    loadFilterList,
    getFilterDetail,
    addFilter,
    editFilter,
    deleteFilter,
    deleteFilters,
    toggleFilterStatus,
  }
}

