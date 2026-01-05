/**
 * API请求插件
 * 将API请求工具注册为全局属性和Vue插件
 */
import type { App } from 'vue'
import service, { get, post, put, del } from '@/api/request'

/**
 * API工具对象
 * 内部使用，不导出
 */
const api = {
  get,
  post,
  put,
  delete: del,
  service
}

/**
 * API插件
 * 用于注册全局API工具
 */
export default {
  install(app: App) {
    // 注册为全局属性，可通过this.$api访问
    app.config.globalProperties.$api = api
    
    // 同时提供注入方式访问，可通过inject('$api')获取
    app.provide('$api', api)
  }
} 