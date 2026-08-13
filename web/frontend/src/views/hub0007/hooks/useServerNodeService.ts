/**
 * 系统节点监控业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { JsonDataObj } from '@/types/api'
import { createBackendPaginationParams } from '@/utils/pagination'
import type { Ref } from 'vue'
import * as serverNodeApi from '../api'
import { useServerNodeModel } from './model'

/**
 * 系统节点服务 Hook
 */
export function useServerNodeService(searchFormRef?: Ref<any> | any) {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0007')

  const model = useServerNodeModel()

  const { loading, serverList, pageInfo, setServerList, updatePagination } = model

  /**
   * 加载系统节点列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadServerNodes = async (searchParams?: Record<string, any>) => {
    loading.value = true
    try {
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(
            Object.entries(finalSearchParams).filter(
              ([, value]) => value !== '' && value !== null && value !== undefined,
            ),
          )
        : {}

      const params = {
        ...effectiveSearchParams,
        activeFlag: 'Y',
        ...createBackendPaginationParams(pageInfo.value?.pageIndex, pageInfo.value?.pageSize),
      }

      const response: JsonDataObj = await serverNodeApi.queryServerInfos(params)

      if (response.oK) {
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const servers = Array.isArray(bizData) ? bizData : []
          setServerList(servers)
        }

        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || t('messages.queryListFailed'))
      }
    } catch (error) {
      console.error('加载系统节点列表失败:', error)
      message.error(t('messages.loadListFailed'))
    } finally {
      loading.value = false
    }
  }

  /**
   * 处理搜索（重置到第一页）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
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

  /**
   * 获取节点详情
   */
  const getServerDetail = async (metricServerId: string) => {
    try {
      const response: JsonDataObj = await serverNodeApi.getServerInfo(metricServerId)

      if (response.oK && response.bizData) {
        return JSON.parse(response.bizData)
      }

      message.error(response.errMsg || t('messages.getDetailFailed'))
      return null
    } catch (error) {
      console.error('获取节点详情失败:', error)
      message.error(t('messages.getDetailFailed'))
      return null
    }
  }

  return {
    model,
    loading,
    serverList,
    pageInfo,
    loadServerNodes,
    handleSearch,
    handlePageChange,
    getServerDetail,
  }
}

/**
 * ServerNodeService 类型定义
 */
export type ServerNodeService = ReturnType<typeof useServerNodeService>
