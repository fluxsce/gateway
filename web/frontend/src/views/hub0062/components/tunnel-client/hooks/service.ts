/**
 * 隧道客户端管理服务层
 * 处理业务逻辑和API调用
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { PlayCircleOutline, StopCircleOutline, WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as tunnelClientApi from '../../../api'
import type { TunnelClient } from '../../../types'
import { useTunnelClientModel } from './model'

/**
 * 隧道客户端管理服务
 */
export function useTunnelClientService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()
  const model = useTunnelClientModel()

  const {
    loading,
    clientList,
    pageInfo,
    updatePagination,
  } = model

  /**
   * 加载客户端列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  async function loadClientList(searchParams?: Record<string, any>) {
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
      const paginationParams = createBackendPaginationParams(
        pageInfo.value?.pageIndex,
        pageInfo.value?.pageSize
      )
      
      const params = {
        // 查询条件
        ...effectiveSearchParams,
        // 分页参数（使用 pageIndex 与后端保持一致）
        pageIndex: paginationParams.pageIndex,
        pageSize: paginationParams.pageSize
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await tunnelClientApi.queryTunnelClients(params)

      // 使用标准的 JsonDataObj 格式
      if (response.oK) {
        // 解析业务数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const clients = Array.isArray(bizData) ? bizData : []
          clientList.value = clients
        }

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询客户端列表失败')
      }
    } catch (error) {
      console.error('加载客户端列表失败:', error)
      message.error('加载客户端列表失败')
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取客户端详情
   */
  async function getClientDetail(tunnelClientId: string): Promise<TunnelClient | null> {
    try {
      const response: JsonDataObj = await tunnelClientApi.getTunnelClient(tunnelClientId)
      
      if (response.oK) {
        if (response.bizData) {
          const clientInfo = JSON.parse(response.bizData)
          return clientInfo
        }
        return null
      } else {
        message.error(response.errMsg || '获取客户端详情失败')
        return null
      }
    } catch (error) {
      console.error('获取客户端详情失败:', error)
      message.error('获取客户端详情失败')
      return null
    }
  }

  /**
   * 新增客户端
   */
  async function addClient(data: Partial<TunnelClient>): Promise<boolean> {
    loading.value = true
    try {
      const response: JsonDataObj = await tunnelClientApi.createTunnelClient(data as any)

      if (response.oK && response.state) {
        message.success(response.popMsg || '创建客户端成功')
        await loadClientList()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '创建客户端失败')
        return false
      }
    } catch (error) {
      console.error('创建客户端失败:', error)
      message.error('创建客户端失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 编辑客户端
   */
  async function editClient(tunnelClientId: string, data: Partial<TunnelClient>): Promise<boolean> {
    loading.value = true
    try {
      const response: JsonDataObj = await tunnelClientApi.updateTunnelClient({
        ...data,
        tunnelClientId,
      } as any)

      if (response.oK && response.state) {
        message.success(response.popMsg || '更新客户端成功')
        await loadClientList()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '更新客户端失败')
        return false
      }
    } catch (error) {
      console.error('更新客户端失败:', error)
      message.error('更新客户端失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除客户端
   */
  async function deleteClient(client: TunnelClient): Promise<boolean> {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除客户端 "${client.clientName}" 吗？`,
      icon: WarningOutline,
      headerStyle: 'gradient',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await tunnelClientApi.deleteTunnelClient(client.tunnelClientId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '删除客户端成功')
        await loadClientList()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '删除客户端失败')
        return false
      }
    } catch (error) {
      console.error('删除客户端失败:', error)
      message.error('删除客户端失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 批量删除客户端
   */
  async function deleteClients(tunnelClientIds: string[]): Promise<boolean> {
    loading.value = true
    try {
      let successCount = 0
      for (const id of tunnelClientIds) {
        const response: JsonDataObj = await tunnelClientApi.deleteTunnelClient(id)
        if (response.oK && response.state) {
          successCount++
        }
      }

      if (successCount > 0) {
        message.success(`成功删除 ${successCount} 个客户端`)
        await loadClientList()
        return true
      } else {
        message.error('删除客户端失败')
        return false
      }
    } catch (error) {
      console.error('批量删除客户端失败:', error)
      message.error('批量删除客户端失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 连接客户端
   */
  async function connectClient(client: TunnelClient): Promise<boolean> {
    const confirmed = await gDialog.warning({
      title: '确认连接',
      subtitle: '连接后将开始与服务器建立连接',
      content: `确定要连接客户端 "${client.clientName}" 吗？`,
      icon: PlayCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定连接',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await tunnelClientApi.startClient(client.tunnelClientId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '连接客户端成功')
        await loadClientList()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '连接客户端失败')
        return false
      }
    } catch (error) {
      console.error('连接客户端失败:', error)
      message.error('连接客户端失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 断开客户端连接
   */
  async function disconnectClient(client: TunnelClient): Promise<boolean> {
    const confirmed = await gDialog.warning({
      title: '确认断开',
      subtitle: '断开后将停止与服务器的连接',
      content: `确定要断开客户端 "${client.clientName}" 的连接吗？`,
      icon: StopCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定断开',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await tunnelClientApi.stopClient(client.tunnelClientId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '断开连接成功')
        await loadClientList()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '断开连接失败')
        return false
      }
    } catch (error) {
      console.error('断开连接失败:', error)
      message.error('断开连接失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 切换客户端状态
   */
  async function toggleClientStatus(client: TunnelClient): Promise<boolean> {
    const newStatus = client.activeFlag === 'Y' ? 'N' : 'Y'
    return await editClient(client.tunnelClientId, {
      activeFlag: newStatus,
    })
  }

  return {
    // Model 实例（包含 paginationConfig 和 menuConfig）
    model,
    
    // 数据加载
    loadClientList,
    
    // 客户端操作
    getClientDetail,
    addClient,
    editClient,
    deleteClient,
    deleteClients,
    connectClient,
    disconnectClient,
    toggleClientStatus,
  }
}

/**
 * 隧道客户端管理服务类型
 */
export type TunnelClientService = ReturnType<typeof useTunnelClientService>

