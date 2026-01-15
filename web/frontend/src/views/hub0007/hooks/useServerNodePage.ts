/**
 * 系统节点监控页面 Hook
 * 整合服务层与 UI 交互逻辑
 */

import type { Ref } from 'vue'
import { useServerNodeService } from './useServerNodeService'

/**
 * 系统节点管理页面 Hook
 */
export function useServerNodePage(gridRef: Ref, searchFormRef: Ref) {
  // 创建服务实例（传入 searchFormRef，让服务层可以自动获取搜索条件）
  const service = useServerNodeService(searchFormRef)

  const { model } = service

  /**
   * 处理搜索
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    await service.handleSearch(searchParams)
  }

  /**
   * 处理工具栏按钮点击
   */
  const handleToolbarClick = (buttonKey: string) => {
    console.log('Toolbar button clicked:', buttonKey)
    // 预留工具栏按钮处理逻辑
  }

  /**
   * 处理表格菜单点击
   */
  const handleMenuClick = async ({ code, row }: { code: string; row?: any }) => {
    if (!row) return
    
    switch (code) {
      case 'view':
        await handleViewDetail(row)
        break
      default:
        console.log('Unknown menu key:', code)
    }
  }

  /**
   * 查看节点详情
   */
  const handleViewDetail = async (row: any) => {
    console.log('查看节点详情:', row)
    const detail = await service.getServerDetail(row.metricServerId)
    if (detail) {
      // TODO: 显示详情对话框
      console.log('节点详情:', detail)
    }
  }

  return {
    model,
    service,
    handleToolbarClick,
    handleMenuClick,
    handleSearch
  }
}

/**
 * ServerNodePage 类型定义
 */
export type ServerNodePage = ReturnType<typeof useServerNodePage>
