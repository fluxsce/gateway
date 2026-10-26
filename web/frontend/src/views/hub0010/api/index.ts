/**
 * 工具市场 API。
 *
 * 路径前缀 /gateway/hub0010
 * - POST /queryTools 当前用户有权使用的工具及本人安装状态
 * - POST /installTool 安装到当前用户
 * - POST /uninstallTool 卸载当前用户的安装
 * - POST /saveToolConfig 保存已安装工具的个人配置
 */

import { createApi } from '@/api/request'
import { moduleApiPrefix } from '@/api/requestPath'
import type { JsonDataObj } from '@/types/api'

const toolApi = createApi(moduleApiPrefix('hub0010'))

/** 目录里一条工具的配置项。 */
export interface ToolConfigField {
  key: string
  type: string
  options?: Array<{ value: number | string }>
}

/** 查询接口返回的工具。名称和说明由前端按 id 翻译。 */
export interface ToolCatalogItem {
  id: string
  name: string
  category: string
  version: string
  author: string
  icon: string
  tags: string[]
  status: 'available' | 'installed' | 'installing'
  resourceCode: string
  configSchema: ToolConfigField[]
  defaultConfig: Record<string, unknown>
  userConfig?: Record<string, unknown>
  installTime?: string
}

/** 读取当前用户可见的工具目录。 */
export const queryTools = async (): Promise<JsonDataObj> => {
  return toolApi.post('/queryTools', {})
}

/** 安装工具。config 省略时由服务端写入默认配置。 */
export const installTool = async (
  toolId: string,
  config?: Record<string, unknown>,
): Promise<JsonDataObj> => {
  return toolApi.post('/installTool', { toolId, config })
}

/** 卸载当前用户已安装的工具。 */
export const uninstallTool = async (toolId: string): Promise<JsonDataObj> => {
  return toolApi.post('/uninstallTool', { toolId })
}

/** 保存已安装工具的个人配置。 */
export const saveToolConfig = async (
  toolId: string,
  config: Record<string, unknown>,
): Promise<JsonDataObj> => {
  return toolApi.post('/saveToolConfig', { toolId, config })
}
