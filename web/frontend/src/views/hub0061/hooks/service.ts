/**
 * 静态服务管理服务层 Hook
 * 纯业务逻辑：数据获取、增删改查等操作
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { PlayCircleOutline, RefreshCircleOutline, StopCircleOutline, WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import {
  createStaticServer,
  deleteStaticServer,
  getStaticServer,
  queryStaticServers,
  reloadStaticServer,
  startStaticServer,
  stopStaticServer,
  updateStaticServer
} from '../api'
import type { TunnelStaticServer } from '../types'
import { useStaticServerModel } from './model'

/**
 * 静态服务管理服务层 Hook
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useStaticServerService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 使用 model
  const model = useStaticServerModel()
  
  // 解构 model 中的列表操作方法
  const {
    addServerToList,
    updateServerInList,
    removeServerFromList,
    removeServersFromList,
  } = model

  /**
   * 加载服务列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadServerList = async (searchParams?: Record<string, any>) => {
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

      // 构建请求参数：合并查询条件和分页参数
      const queryParams: any = {
        // 查询条件
        ...effectiveSearchParams,
        // 分页参数
        ...createBackendPaginationParams(
          model.pageInfo.value?.pageIndex,
          model.pageInfo.value?.pageSize
        ),
      }

      const response = await queryStaticServers(queryParams)

      if (isApiSuccess(response)) {
        // 解析业务数据
        const servers = parseJsonData<TunnelStaticServer[]>(response, []) || []
        model.setServerList(servers)

        // 解析分页信息
        const backendPageInfo = parsePageInfo(response)
        if (backendPageInfo && Object.keys(backendPageInfo).length > 0) {
          model.updatePagination(backendPageInfo)
        } else {
          model.resetPagination()
        }
      } else {
        message.error(getApiMessage(response, '加载服务列表失败'))
        model.setServerList([])
        model.resetPagination()
      }
    } catch (error: any) {
      console.error('加载服务列表失败:', error)
      message.error(error.message || '加载服务列表失败')
      model.setServerList([])
      model.resetPagination()
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 获取服务详情
   */
  const getServerDetail = async (tunnelStaticServerId: string): Promise<TunnelStaticServer | null> => {
    try {
      const response = await getStaticServer(tunnelStaticServerId)
      if (isApiSuccess(response)) {
        return parseJsonData<TunnelStaticServer>(response)
      } else {
        message.error(getApiMessage(response, '获取服务详情失败'))
        return null
      }
    } catch (error: any) {
      console.error('获取服务详情失败:', error)
      message.error(error.message || '获取服务详情失败')
      return null
    }
  }

  /**
   * 新增服务
   */
  const addServer = async (serverData: Partial<TunnelStaticServer>): Promise<boolean> => {
    try {
      model.setLoading(true)

      const submitData = {
        ...serverData,
        tenantId: 'default',
      }

      const response = await createStaticServer(submitData)
      if (isApiSuccess(response)) {
        message.success('新增服务成功')
        
        // 如果返回了新增的服务数据，添加到列表
        const newServer = parseJsonData<TunnelStaticServer | null>(response, null)
        if (newServer) {
          addServerToList(newServer)
        } else {
          // 否则重新加载列表
          await loadServerList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '新增服务失败'))
        return false
      }
    } catch (error: any) {
      console.error('新增服务失败:', error)
      message.error(error.message || '新增服务失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 编辑服务
   */
  const editServer = async (tunnelStaticServerId: string, serverData: Partial<TunnelStaticServer>): Promise<boolean> => {
    try {
      model.setLoading(true)

      const submitData = {
        ...serverData,
        tunnelStaticServerId,
        tenantId: 'default',
      }

      const response = await updateStaticServer(submitData as any)
      if (isApiSuccess(response)) {
        message.success('编辑服务成功')
        
        // 更新列表中的服务数据
        const updatedServer = parseJsonData<TunnelStaticServer | null>(response, null)
        if (updatedServer) {
          updateServerInList(updatedServer.tunnelStaticServerId, updatedServer.tenantId, updatedServer)
        } else {
          // 如果后端没有返回数据，重新加载列表以确保数据一致性
          await loadServerList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '编辑服务失败'))
        return false
      }
    } catch (error: any) {
      console.error('编辑服务失败:', error)
      message.error(error.message || '编辑服务失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 删除服务
   */
  const deleteServer = async (server: TunnelStaticServer): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除静态服务 "${server.serverName}" 吗？`,
      icon: WarningOutline,
      headerStyle: 'gradient',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    try {
      model.setLoading(true)

      const response = await deleteStaticServer(server.tunnelStaticServerId)
      if (isApiSuccess(response)) {
        message.success('删除服务成功')
        removeServerFromList(server.tunnelStaticServerId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (model.serverList.value.length === 0 && model.pageInfo.value && model.pageInfo.value.pageIndex > 1) {
          model.updatePagination({ pageIndex: model.pageInfo.value.pageIndex - 1 })
          await loadServerList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '删除服务失败'))
        return false
      }
    } catch (error: any) {
      console.error('删除服务失败:', error)
      message.error(error.message || '删除服务失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 批量删除服务
   */
  const deleteServers = async (tunnelStaticServerIds: string[]): Promise<boolean> => {
    try {
      model.setLoading(true)

      // 逐个删除
      const results = await Promise.all(
        tunnelStaticServerIds.map(id => deleteStaticServer(id))
      )

      const successCount = results.filter(r => isApiSuccess(r)).length
      if (successCount === tunnelStaticServerIds.length) {
        message.success(`成功删除 ${successCount} 个服务`)
        removeServersFromList(tunnelStaticServerIds)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (model.serverList.value.length === 0 && model.pageInfo.value && model.pageInfo.value.pageIndex > 1) {
          model.updatePagination({ pageIndex: model.pageInfo.value.pageIndex - 1 })
          await loadServerList()
        }
        
        return true
      } else {
        message.warning(`部分删除成功：${successCount}/${tunnelStaticServerIds.length}`)
        // 部分删除成功时，重新加载列表以确保数据一致性
        await loadServerList()
        return false
      }
    } catch (error: any) {
      console.error('批量删除服务失败:', error)
      message.error(error.message || '批量删除服务失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 切换服务状态
   */
  const toggleServerStatus = async (server: TunnelStaticServer): Promise<boolean> => {
    const newStatus = server.activeFlag === 'Y' ? 'N' : 'Y'
    return await editServer(server.tunnelStaticServerId, {
      ...server,
      activeFlag: newStatus,
    })
  }

  /**
   * 启动服务
   */
  const startServer = async (server: TunnelStaticServer): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认启动',
      subtitle: '启动后将开始监听客户端连接',
      content: `确定要启动静态服务 "${server.serverName}" 吗？`,
      icon: PlayCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定启动',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    try {
      model.setLoading(true)

      const response = await startStaticServer(server.tunnelStaticServerId)
      if (isApiSuccess(response)) {
        message.success('服务启动成功')
        
        // 更新列表中的服务数据
        const updatedServer = parseJsonData<TunnelStaticServer | null>(response, null)
        if (updatedServer) {
          updateServerInList(updatedServer.tunnelStaticServerId, updatedServer.tenantId, updatedServer)
        } else {
          await loadServerList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '服务启动失败'))
        return false
      }
    } catch (error: any) {
      console.error('服务启动失败:', error)
      message.error(error.message || '服务启动失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 停止服务
   */
  const stopServer = async (server: TunnelStaticServer): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认停止',
      subtitle: '停止后将断开所有客户端连接',
      content: `确定要停止静态服务 "${server.serverName}" 吗？`,
      icon: StopCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定停止',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    try {
      model.setLoading(true)

      const response = await stopStaticServer(server.tunnelStaticServerId)
      if (isApiSuccess(response)) {
        message.success('服务停止成功')
        
        // 更新列表中的服务数据
        const updatedServer = parseJsonData<TunnelStaticServer | null>(response, null)
        if (updatedServer) {
          updateServerInList(updatedServer.tunnelStaticServerId, updatedServer.tenantId, updatedServer)
        } else {
          await loadServerList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '服务停止失败'))
        return false
      }
    } catch (error: any) {
      console.error('服务停止失败:', error)
      message.error(error.message || '服务停止失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 重载服务配置
   */
  const reloadServer = async (server: TunnelStaticServer): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认重载',
      subtitle: '重载配置将应用最新的服务配置',
      content: `确定要重载静态服务 "${server.serverName}" 的配置吗？`,
      icon: RefreshCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定重载',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    try {
      model.setLoading(true)

      const response = await reloadStaticServer(server.tunnelStaticServerId)
      if (isApiSuccess(response)) {
        message.success('服务配置重载成功')
        
        // 更新列表中的服务数据
        const updatedServer = parseJsonData<TunnelStaticServer | null>(response, null)
        if (updatedServer) {
          updateServerInList(updatedServer.tunnelStaticServerId, updatedServer.tenantId, updatedServer)
        } else {
          await loadServerList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '服务配置重载失败'))
        return false
      }
    } catch (error: any) {
      console.error('服务配置重载失败:', error)
      message.error(error.message || '服务配置重载失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  return {
    model,
    loadServerList,
    getServerDetail,
    addServer,
    editServer,
    deleteServer,
    deleteServers,
    toggleServerStatus,
    startServer,
    stopServer,
    reloadServer,
  }
}

/**
 * 静态服务管理 Service 类型
 */
export type StaticServerService = ReturnType<typeof useStaticServerService>

