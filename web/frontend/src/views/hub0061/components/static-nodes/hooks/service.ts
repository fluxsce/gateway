/**
 * 静态节点列表服务层 Hook
 * 纯业务逻辑：数据获取、增删改查等操作
 */

import { createBackendPaginationParams } from '@/components/gpage'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import {
    createStaticNode,
    deleteStaticNode,
    getStaticNode,
    queryStaticNodes,
    updateStaticNode,
} from '../../../api'
import { useStaticNodeModel } from './model'
import type { TunnelStaticNode } from './types'

/**
 * 静态节点列表服务层 Hook
 * @param tunnelStaticServerId 静态服务器ID（必需，用于关联节点）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useStaticNodeService(
  tunnelStaticServerId: Ref<string> | string,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()

  // 获取 tunnelStaticServerId 的实际值
  const getServerId = () => {
    return typeof tunnelStaticServerId === 'string' ? tunnelStaticServerId : tunnelStaticServerId?.value
  }

  // 使用 model
  const model = useStaticNodeModel()
  
  // 解构 model 中的列表操作方法
  const {
    addNodeToList,
    updateNodeInList,
    removeNodeFromList,
    removeNodesFromList,
  } = model

  /**
   * 加载节点列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadNodeList = async (searchParams?: Record<string, any>) => {
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

      // 获取 tunnelStaticServerId
      const serverId = getServerId()

      // 构建请求参数：合并查询条件、服务器ID和分页参数
      const queryParams: any = {
        // 查询条件
        ...effectiveSearchParams,
        // 服务器ID（必需）
        tunnelStaticServerId: serverId,
        // 分页参数
        ...createBackendPaginationParams(
          model.pageInfo.value?.pageIndex,
          model.pageInfo.value?.pageSize
        ),
      }

      const response = await queryStaticNodes(queryParams)

      if (isApiSuccess(response)) {
        // 解析业务数据
        const nodes = parseJsonData<TunnelStaticNode[]>(response, []) || []
        model.setNodeList(nodes)

        // 解析分页信息
        const backendPageInfo = parsePageInfo(response)
        if (backendPageInfo && Object.keys(backendPageInfo).length > 0) {
          model.updatePagination(backendPageInfo)
        } else {
          model.resetPagination()
        }
      } else {
        message.error(getApiMessage(response, '加载节点列表失败'))
        model.setNodeList([])
        model.resetPagination()
      }
    } catch (error: any) {
      console.error('加载节点列表失败:', error)
      message.error(error.message || '加载节点列表失败')
      model.setNodeList([])
      model.resetPagination()
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 获取节点详情
   */
  const getNodeDetail = async (tunnelStaticNodeId: string): Promise<TunnelStaticNode | null> => {
    try {
      const response = await getStaticNode(tunnelStaticNodeId)
      if (isApiSuccess(response)) {
        return parseJsonData<TunnelStaticNode>(response)
      } else {
        message.error(getApiMessage(response, '获取节点详情失败'))
        return null
      }
    } catch (error: any) {
      console.error('获取节点详情失败:', error)
      message.error(error.message || '获取节点详情失败')
      return null
    }
  }

  /**
   * 新增节点
   */
  const addNode = async (nodeData: Partial<TunnelStaticNode>): Promise<boolean> => {
    try {
      model.setLoading(true)

      const submitData = {
        ...nodeData,
        tenantId: 'default',
        tunnelStaticServerId: getServerId(),
      }

      const response = await createStaticNode(submitData)
      if (isApiSuccess(response)) {
        message.success('新增节点成功')
        
        // 如果返回了新增的节点数据，添加到列表
        const newNode = parseJsonData<TunnelStaticNode | null>(response, null)
        if (newNode) {
          addNodeToList(newNode)
        } else {
          // 否则重新加载列表
          await loadNodeList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '新增节点失败'))
        return false
      }
    } catch (error: any) {
      console.error('新增节点失败:', error)
      message.error(error.message || '新增节点失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 编辑节点
   */
  const editNode = async (tunnelStaticNodeId: string, nodeData: Partial<TunnelStaticNode>): Promise<boolean> => {
    try {
      model.setLoading(true)

      const submitData = {
        ...nodeData,
        tunnelStaticNodeId,
        tenantId: 'default',
      }

      const response = await updateStaticNode(submitData)
      if (isApiSuccess(response)) {
        message.success('编辑节点成功')
        
        // 更新列表中的节点数据：必须以返回的 response.bizData 为准
        const updatedNode = parseJsonData<TunnelStaticNode | null>(response, null)
        if (updatedNode) {
          updateNodeInList(updatedNode.tunnelStaticNodeId, updatedNode.tenantId, updatedNode)
        } else {
          // 如果后端没有返回数据，重新加载列表以确保数据一致性
          await loadNodeList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '编辑节点失败'))
        return false
      }
    } catch (error: any) {
      console.error('编辑节点失败:', error)
      message.error(error.message || '编辑节点失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 删除节点
   */
  const deleteNode = async (tunnelStaticNodeId: string): Promise<boolean> => {
    try {
      model.setLoading(true)

      const response = await deleteStaticNode(tunnelStaticNodeId)
      if (isApiSuccess(response)) {
        message.success('删除节点成功')
        removeNodeFromList(tunnelStaticNodeId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (model.nodeList.value.length === 0 && model.pageInfo.value && model.pageInfo.value.pageIndex > 1) {
          model.updatePagination({ pageIndex: model.pageInfo.value.pageIndex - 1 })
          await loadNodeList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '删除节点失败'))
        return false
      }
    } catch (error: any) {
      console.error('删除节点失败:', error)
      message.error(error.message || '删除节点失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 批量删除节点
   */
  const deleteNodes = async (tunnelStaticNodeIds: string[]): Promise<boolean> => {
    try {
      model.setLoading(true)

      // 逐个删除
      const results = await Promise.all(
        tunnelStaticNodeIds.map(id => deleteStaticNode(id))
      )

      const successCount = results.filter(r => isApiSuccess(r)).length
      if (successCount === tunnelStaticNodeIds.length) {
        message.success(`成功删除 ${successCount} 个节点`)
        removeNodesFromList(tunnelStaticNodeIds)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (model.nodeList.value.length === 0 && model.pageInfo.value && model.pageInfo.value.pageIndex > 1) {
          model.updatePagination({ pageIndex: model.pageInfo.value.pageIndex - 1 })
          await loadNodeList()
        }
        
        return true
      } else {
        message.warning(`部分删除成功：${successCount}/${tunnelStaticNodeIds.length}`)
        // 部分删除成功时，重新加载列表以确保数据一致性
        await loadNodeList()
        return false
      }
    } catch (error: any) {
      console.error('批量删除节点失败:', error)
      message.error(error.message || '批量删除节点失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 切换节点状态
   */
  const toggleNodeStatus = async (node: TunnelStaticNode): Promise<boolean> => {
    const newStatus = node.activeFlag === 'Y' ? 'N' : 'Y'
    return await editNode(node.tunnelStaticNodeId, {
      ...node,
      activeFlag: newStatus,
    })
  }

  return {
    model,
    loadNodeList,
    getNodeDetail,
    addNode,
    editNode,
    deleteNode,
    deleteNodes,
    toggleNodeStatus,
  }
}

