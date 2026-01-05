/**
 * Mock数据模块入口文件
 * 用于在开发阶段模拟后端API响应数据
 * 通过环境变量VITE_USE_MOCK控制是否启用
 *
 * 本模块负责:
 * 1. 收集所有mock模块定义
 * 2. 导出生产环境配置函数
 * 3. 提供默认导出给vite-plugin-mock开发环境使用
 */

// vite-plugin-mock/client用于生产环境
import { createProdMockServer } from 'vite-plugin-mock/client'
import type { MockMethod } from 'vite-plugin-mock'

/**
 * 自动导入所有模块
 * glob导入pattern:
 * - 匹配'./modules/'目录下所有ts文件
 * - eager: true表示同步导入而非异步
 */
const modules = import.meta.glob<{ default: MockMethod[] }>('./modules/*.ts', { eager: true })

/**
 * 收集所有mock定义
 * 遍历所有模块文件并合并它们的默认导出
 * 排除以'_'开头的文件(约定这些是工具文件)
 */
const mockModules: MockMethod[] = []
Object.keys(modules).forEach((key) => {
  // 跳过以'_'开头的文件
  if (key.includes('/_')) {
    return
  }
  // 合并该模块的mock定义到全局数组
  mockModules.push(...modules[key].default)
})

/**
 * 在生产环境中设置mock
 * 仅当VITE_USE_MOCK为true时使用
 * 该函数在main.ts中被动态导入并调用
 */
export function setupProdMock(): void {
  if (import.meta.env.VITE_USE_MOCK === 'true') {
    createProdMockServer(mockModules)
  }
}

/**
 * 模块默认导出
 * 由vite-plugin-mock在开发环境自动使用
 * 不需要在代码中显式导入
 */
export default mockModules
