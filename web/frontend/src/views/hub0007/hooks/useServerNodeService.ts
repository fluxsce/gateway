/**
 * 系统节点监控业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as serverNodeApi from '../api'
import { useServerNodeModel } from './model'

/**
 * 系统节点服务 Hook
 */
export function useServerNodeService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 初始化 Model
  const model = useServerNodeModel()

  const { loading, serverList, pageInfo, setServerList, updatePagination } = model

  // ============= 数据加载 =============

  /**
   * 加载系统节点列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadServerNodes = async (searchParams?: Record<string, any>) => {
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
        // 默认只查询活动状态的节点
        activeFlag: 'Y',
        // 分页参数
        ...createBackendPaginationParams(pageInfo.value?.pageIndex, pageInfo.value?.pageSize)
      }

      // 调用 API
      const response: JsonDataObj = await serverNodeApi.queryServerInfos(params)

      if (response.oK) {
        // 解析业务数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const servers = Array.isArray(bizData) ? bizData : []
          setServerList(servers)
        }

        // 解析分页信息
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询系统节点列表失败')
      }
    } catch (error) {
      console.error('加载系统节点列表失败:', error)
      message.error('加载系统节点列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 处理搜索（重置到第一页）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    // 重置分页到第一页
    if (pageInfo.value) {
      pageInfo.value.pageIndex = 1
    }
    await loadServerNodes(searchParams)
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async (newPageInfo: { pageIndex?: number; pageSize?: number }) => {
    if (pageInfo.value) {
      if (newPageInfo.pageIndex !== undefined) {
        pageInfo.value.pageIndex = newPageInfo.pageIndex
      }
      if (newPageInfo.pageSize !== undefined) {
        pageInfo.value.pageSize = newPageInfo.pageSize
      }
    }
    await loadServerNodes()
  }

  // ============= 查看详情 =============

  /**
   * 获取节点详情
   */
  const getServerDetail = async (metricServerId: string) => {
    try {
      const response: JsonDataObj = await serverNodeApi.getServerInfo(metricServerId)

      if (response.oK && response.bizData) {
        const serverInfo = JSON.parse(response.bizData)
        return serverInfo
      } else {
        message.error(response.errMsg || '获取节点详情失败')
        return null
      }
    } catch (error) {
      console.error('获取节点详情失败:', error)
      message.error('获取节点详情失败')
      return null
    }
  }

  // ============= 导出 =============

  return {
    model,
    loading,
    serverList,
    pageInfo,
    loadServerNodes,
    handleSearch,
    handlePageChange,
    getServerDetail
  }
}

/**
 * ServerNodeService 类型定义
 */
export type ServerNodeService = ReturnType<typeof useServerNodeService>

