/**
 * 插件管理
 * 集中注册所有自定义插件
 */
import type { App } from 'vue'
import ApiPlugin from './api'
import UtilsPlugin from './utils'

// 创建全局引用，以便非Vue组件访问
import service, { get, post, put, del } from '@/api/request'
export const $api = { get, post, put, delete: del, service }

// 插件列表 - 内部使用，不导出
const plugins = [
  ApiPlugin,
  UtilsPlugin,
  // 在这里添加其他插件
]

/**
 * 注册所有插件
 * 在应用主入口调用此函数注册全部插件
 * 
 * @param app Vue应用实例
 */
export function setupPlugins(app: App) {
  plugins.forEach(plugin => {
    app.use(plugin)
  })
  
  // 调试信息
  if (import.meta.env.DEV) {
    console.log('所有插件已注册')
  }
  
  // 添加到全局变量，便于非Vue组件文件使用
  window.$api = $api
}
