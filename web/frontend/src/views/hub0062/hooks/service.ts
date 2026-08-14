/**
 * 隧道客户端管理服务层
 * 处理业务逻辑和API调用
 */

import { useAppMessage } from '@/composables/useAppMessage'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import type { Ref } from 'vue'
import * as tunnelClientApi from '../api'
import type { TunnelClient, TunnelClientQueryParams } from '../types'
import { useTunnelClientModel } from './model'

/**
 * 隧道客户端管理服务
 */
export function useTunnelClientService(searchFormRef?: Ref<any> | any) {
  const message = useAppMessage()
  const model = useTunnelClientModel()

  /**
   * 加载客户端列表
   */
  async function loadClientList(searchParams?: Record<string, any>) {
    model.loading.value = true
    try {
      // 构建查询参数
      const params: TunnelClientQueryParams = {
        clientName: searchParams?.clientName || '',
        serverAddress: searchParams?.serverAddress || '',
        connectionStatus: searchParams?.connectionStatus || undefined,
        activeFlag: searchParams?.activeFlag || undefined,
        keyword: searchParams?.keyword || '',
        pageIndex: model.pageInfo.value?.pageIndex || 1,
        pageSize: model.pageInfo.value?.pageSize || 20,
      }

      const response = await tunnelClientApi.queryTunnelClients(params)

      if (isApiSuccess(response)) {
        // 解析客户端列表
        const clients = parseJsonData<TunnelClient[]>(response, []) || []
        model.clientList.value = clients

        // 解析分页信息
        const pageInfoData = parsePageInfo(response)
        if (pageInfoData) {
          model.pageInfo.value = pageInfoData
        }
      } else {
        message.error(getApiMessage(response, '加载客户端列表失败'))
      }
    } catch (error: any) {
      console.error('加载客户端列表失败:', error)
      message.error(error.message || '加载客户端列表失败')
    } finally {
      model.loading.value = false
    }
  }

  /**
   * 获取客户端详情
   */
  async function getClientDetail(tunnelClientId: string): Promise<TunnelClient | null> {
    try {
      const response = await tunnelClientApi.getTunnelClient(tunnelClientId)
      if (isApiSuccess(response)) {
        return parseJsonData<TunnelClient>(response)
      } else {
        message.error(getApiMessage(response, '获取客户端详情失败'))
        return null
      }
    } catch (error: any) {
      console.error('获取客户端详情失败:', error)
      message.error(error.message || '获取客户端详情失败')
      return null
    }
  }

  /**
   * 新增客户端
   */
  async function addClient(data: Partial<TunnelClient>): Promise<boolean> {
    model.loading.value = true
    try {
      const response = await tunnelClientApi.createTunnelClient(data as any)
      if (isApiSuccess(response)) {
        message.success('新增客户端成功')
        await loadClientList()
        return true
      } else {
        message.error(getApiMessage(response, '新增客户端失败'))
        return false
      }
    } catch (error: any) {
      console.error('新增客户端失败:', error)
      message.error(error.message || '新增客户端失败')
      return false
    } finally {
      model.loading.value = false
    }
  }

  /**
   * 编辑客户端
   */
  async function editClient(tunnelClientId: string, data: Partial<TunnelClient>): Promise<boolean> {
    model.loading.value = true
    try {
      const response = await tunnelClientApi.updateTunnelClient({
        ...data,
        tunnelClientId,
      } as any)
      if (isApiSuccess(response)) {
        message.success('编辑客户端成功')
        await loadClientList()
        return true
      } else {
        message.error(getApiMessage(response, '编辑客户端失败'))
        return false
      }
    } catch (error: any) {
      console.error('编辑客户端失败:', error)
      message.error(error.message || '编辑客户端失败')
      return false
    } finally {
      model.loading.value = false
    }
  }

  /**
   * 删除客户端
   */
  async function deleteClient(tunnelClientId: string): Promise<boolean> {
    model.loading.value = true
    try {
      const response = await tunnelClientApi.deleteTunnelClient(tunnelClientId)
      if (isApiSuccess(response)) {
        message.success('删除客户端成功')
        await loadClientList()
        return true
      } else {
        message.error(getApiMessage(response, '删除客户端失败'))
        return false
      }
    } catch (error: any) {
      console.error('删除客户端失败:', error)
      message.error(error.message || '删除客户端失败')
      return false
    } finally {
      model.loading.value = false
    }
  }

  /**
   * 批量删除客户端
   */
  async function deleteClients(tunnelClientIds: string[]): Promise<boolean> {
    model.loading.value = true
    try {
      let successCount = 0
      for (const id of tunnelClientIds) {
        const response = await tunnelClientApi.deleteTunnelClient(id)
        if (isApiSuccess(response)) {
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
    } catch (error: any) {
      console.error('批量删除客户端失败:', error)
      message.error(error.message || '批量删除客户端失败')
      return false
    } finally {
      model.loading.value = false
    }
  }

  /**
   * 连接客户端
   */
  async function connectClient(tunnelClientId: string): Promise<boolean> {
    model.loading.value = true
    try {
      const response = await tunnelClientApi.startClient(tunnelClientId)
      if (isApiSuccess(response)) {
        message.success('连接客户端成功')
        await loadClientList()
        return true
      } else {
        message.error(getApiMessage(response, '连接客户端失败'))
        return false
      }
    } catch (error: any) {
      console.error('连接客户端失败:', error)
      message.error(error.message || '连接客户端失败')
      return false
    } finally {
      model.loading.value = false
    }
  }

  /**
   * 断开客户端连接
   */
  async function disconnectClient(tunnelClientId: string): Promise<boolean> {
    model.loading.value = true
    try {
      const response = await tunnelClientApi.stopClient(tunnelClientId)
      if (isApiSuccess(response)) {
        message.success('断开连接成功')
        await loadClientList()
        return true
      } else {
        message.error(getApiMessage(response, '断开连接失败'))
        return false
      }
    } catch (error: any) {
      console.error('断开连接失败:', error)
      message.error(error.message || '断开连接失败')
      return false
    } finally {
      model.loading.value = false
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
    // Model
    model,

    // 方法
    loadClientList,
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

