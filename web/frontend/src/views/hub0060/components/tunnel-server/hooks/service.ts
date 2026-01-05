/**
 * 隧道服务器管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { PlayCircleOutline, RefreshCircleOutline, StopCircleOutline, WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as tunnelServerApi from '../../../api'
import type { TunnelServer, TunnelServerForm } from '../../../types'
import { useTunnelServerModel } from './model'

/**
 * 隧道服务器服务 Hook（纯业务逻辑）
 */
export function useTunnelServerService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model
  const model = useTunnelServerModel()

  const {
    loading,
    tunnelServerList,
    pageInfo,
    setTunnelServerList,
    updatePagination,
    addTunnelServerToList,
    updateTunnelServerInList,
    removeTunnelServerFromList
  } = model

  // ============= 数据加载 =============

  /**
   * 加载隧道服务器列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadTunnelServers = async (searchParams?: Record<string, any>) => {
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
      const response: JsonDataObj = await tunnelServerApi.queryTunnelServers(params)

      // 使用标准的 JsonDataObj 格式
      if (response.oK) {
        // 解析业务数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const servers = Array.isArray(bizData) ? bizData : []
          setTunnelServerList(servers)
        }

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询隧道服务器列表失败')
      }
    } catch (error) {
      console.error('加载隧道服务器列表失败:', error)
      message.error('加载隧道服务器列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索隧道服务器
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    // loadTunnelServers 会自动从 searchFormRef 获取查询条件
    await loadTunnelServers(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadTunnelServers()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    // loadTunnelServers 会自动从 searchFormRef 获取查询条件
    await loadTunnelServers()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadTunnelServers()
  }

  // ============= 隧道服务器操作 =============

  /**
   * 创建隧道服务器
   */
  const createTunnelServer = async (serverData: Partial<TunnelServer>): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await tunnelServerApi.createTunnelServer(serverData as TunnelServerForm)

      if (response.oK && response.state) {
        message.success(response.popMsg || '创建隧道服务器成功')
        
        // 如果返回了新增的服务器数据，添加到列表
        if (response.bizData) {
          const newServer = JSON.parse(response.bizData)
          addTunnelServerToList(newServer)
        } else {
          // 否则重新加载列表
          await loadTunnelServers()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '创建隧道服务器失败')
        return false
      }
    } catch (error) {
      console.error('创建隧道服务器失败:', error)
      message.error('创建隧道服务器失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新隧道服务器
   */
  const updateTunnelServer = async (serverData: TunnelServer): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await tunnelServerApi.updateTunnelServer({
        ...serverData,
        tunnelServerId: serverData.tunnelServerId
      } as TunnelServerForm & { tunnelServerId: string })

      if (response.oK && response.state) {
        message.success(response.popMsg || '更新隧道服务器成功')
        
        // 更新列表中的服务器数据
        if (response.bizData) {
          const updatedServer = JSON.parse(response.bizData)
          updateTunnelServerInList(updatedServer.tunnelServerId, updatedServer.tenantId, updatedServer)
        } else {
          updateTunnelServerInList(serverData.tunnelServerId, serverData.tenantId, serverData)
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '更新隧道服务器失败')
        return false
      }
    } catch (error) {
      console.error('更新隧道服务器失败:', error)
      message.error('更新隧道服务器失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除隧道服务器
   */
  const deleteTunnelServer = async (server: TunnelServer): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除隧道服务器 "${server.serverName}" 吗？`,
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
      const response: JsonDataObj = await tunnelServerApi.deleteTunnelServer(server.tunnelServerId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '删除隧道服务器成功')
        removeTunnelServerFromList(server.tunnelServerId, server.tenantId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (tunnelServerList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadTunnelServers()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '删除隧道服务器失败')
        return false
      }
    } catch (error) {
      console.error('删除隧道服务器失败:', error)
      message.error('删除隧道服务器失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 启动隧道服务器
   */
  const startTunnelServer = async (server: TunnelServer): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认启动',
      subtitle: '启动后将开始监听客户端连接',
      content: `确定要启动隧道服务器 "${server.serverName}" 吗？`,
      icon: PlayCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定启动',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await tunnelServerApi.startTunnelServer(server.tunnelServerId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '启动隧道服务器成功')
        await loadTunnelServers()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '启动隧道服务器失败')
        return false
      }
    } catch (error) {
      console.error('启动隧道服务器失败:', error)
      message.error('启动隧道服务器失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 停止隧道服务器
   */
  const stopTunnelServer = async (server: TunnelServer): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认停止',
      subtitle: '停止后将断开所有客户端连接',
      content: `确定要停止隧道服务器 "${server.serverName}" 吗？`,
      icon: StopCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定停止',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await tunnelServerApi.stopTunnelServer(server.tunnelServerId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '停止隧道服务器成功')
        await loadTunnelServers()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '停止隧道服务器失败')
        return false
      }
    } catch (error) {
      console.error('停止隧道服务器失败:', error)
      message.error('停止隧道服务器失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 重启隧道服务器
   */
  const restartTunnelServer = async (server: TunnelServer): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认重启',
      subtitle: '重启将先停止后启动服务器',
      content: `确定要重启隧道服务器 "${server.serverName}" 吗？`,
      icon: RefreshCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定重启',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await tunnelServerApi.restartTunnelServer(server.tunnelServerId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '重启隧道服务器成功')
        await loadTunnelServers()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '重启隧道服务器失败')
        return false
      }
    } catch (error) {
      console.error('重启隧道服务器失败:', error)
      message.error('重启隧道服务器失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 查看隧道服务器详情
   */
  const viewTunnelServer = async (server: TunnelServer) => {
    try {
      const response: JsonDataObj = await tunnelServerApi.getTunnelServerDetail(server.tunnelServerId)
      
      if (response.oK) {
        const serverInfo = JSON.parse(response.bizData)
        return serverInfo
      } else {
        message.error(response.errMsg || '获取隧道服务器详情失败')
        return null
      }
    } catch (error) {
      console.error('获取隧道服务器详情失败:', error)
      message.error('获取隧道服务器详情失败')
      return null
    }
  }

  return {
    // Model 实例（包含 paginationConfig 和 menuConfig）
    model,
    
    // 数据加载
    loadTunnelServers,
    
    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
    
    // 隧道服务器操作
    createTunnelServer,
    updateTunnelServer,
    deleteTunnelServer,
    startTunnelServer,
    stopTunnelServer,
    restartTunnelServer,
    viewTunnelServer
  }
}

/**
 * 服务返回类型
 */
export type TunnelServerService = ReturnType<typeof useTunnelServerService>

