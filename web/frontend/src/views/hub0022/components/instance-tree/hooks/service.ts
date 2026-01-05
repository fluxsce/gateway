/**
 * 网关实例树组件 Service
 * 处理数据加载和API调用
 */

import { isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { useMessage } from 'naive-ui'
import { queryAllGatewayInstances } from '../../../api'
import type { GatewayInstance } from '../types'
import type { GatewayInstanceTreeModel } from './model'

/**
 * 网关实例树 Service
 */
export function useGatewayInstanceTreeService(model: GatewayInstanceTreeModel) {
  const message = useMessage()

  const {
    loading,
    currentPage,
    pageSize,
    filterKeyword,
    setInstanceList,
    setTotalCount,
    setLoading,
    resetPage,
  } = model

  /**
   * 加载网关实例列表（后端分页）
   */
  async function loadGatewayInstances() {
    try {
      setLoading(true)
      const res = await queryAllGatewayInstances({
        activeFlag: 'Y', // 默认只加载启用的实例
        instanceName: filterKeyword.value.trim() || undefined, // 搜索关键词
        pageIndex: currentPage.value,
        pageSize: pageSize.value,
      })

      if (isApiSuccess(res)) {
        // 使用 parseJsonData 解析数组数据
        const instanceData = parseJsonData<GatewayInstance[]>(res, [])
        setInstanceList(Array.isArray(instanceData) ? instanceData : [])

        // 解析分页信息，设置总数
        try {
          const pageInfo = parsePageInfo(res)
          setTotalCount(pageInfo.totalCount || 0)
        } catch (error) {
          console.warn('解析分页信息失败:', error)
          setTotalCount(0)
        }
      } else {
        message.error(res.errMsg || '获取网关实例列表失败')
        setInstanceList([])
        setTotalCount(0)
      }
    } catch (error) {
      console.error('加载网关实例失败:', error)
      message.error('加载网关实例失败')
      setInstanceList([])
      setTotalCount(0)
    } finally {
      setLoading(false)
    }
  }

  return {
    loadGatewayInstances,
  }
}

/**
 * 网关实例树 Service 类型
 */
export type GatewayInstanceTreeService = ReturnType<typeof useGatewayInstanceTreeService>

