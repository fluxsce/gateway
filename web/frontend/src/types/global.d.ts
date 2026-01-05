/**
 * 全局类型声明文件
 * 增强内置对象和提供全局类型
 */
/// <reference types="./components.d.ts" />
import { store } from '@/stores'
import * as storageUtils from '@/utils/storage'
import * as formatUtils from '@/utils/format'
import * as validateUtils from '@/utils/validate'

// API接口类型
interface ApiInterface {
  get: <T = any>(url: string, params?: any, config?: any) => Promise<T>
  post: <T = any>(url: string, data?: any, params?: any, config?: any) => Promise<T>
  put: <T = any>(url: string, data?: any, config?: any) => Promise<T>
  delete: <T = any>(url: string, params?: any, config?: any) => Promise<T>
  service: any
}

// 工具函数接口类型
interface UtilsInterface {
  storage: typeof storageUtils
  format: typeof formatUtils
  validate: typeof validateUtils
}

// 扩展 Vue 全局属性类型
declare module '@vue/runtime-core' {
  export interface ComponentCustomProperties {
    /**
     * 用户信息全局访问对象
     * 在模板中可以通过 $user 直接访问用户信息
     */
    $user: {
      /**
       * 用户显示名称
       */
      readonly displayName: string

      /**
       * 用户头像
       */
      readonly avatar: string

      /**
       * 是否管理员
       */
      readonly isAdmin: boolean

      /**
       * 是否已登录
       */
      readonly isLoggedIn: boolean

      /**
       * 检查用户是否拥有指定权限
       */
      hasPermission(permCode: string): boolean

      /**
       * 检查用户是否拥有指定角色
       */
      hasRole(roleCode: string): boolean
    }

    /**
     * 应用信息全局访问对象
     * 在模板中可以通过 $app 直接访问应用信息
     */
    $app: {
      /**
       * 应用名称
       */
      readonly name: string

      /**
       * 应用版本
       */
      readonly version: string

      /**
       * 未读通知数量
       */
      readonly notificationCount: number
    }

    /**
     * 全局状态管理访问对象
     * 在模板和脚本中可以通过 $store 直接访问所有状态管理
     */
    $store: typeof store

    /**
     * API请求工具
     * 全局注册的API工具，用于发起HTTP请求
     */
    $api: ApiInterface

    /**
     * 工具函数集合
     * 全局注册的工具函数，包含存储、格式化和验证等常用功能
     */
    $utils: UtilsInterface

    /**
     * highlight.js 实例
     * 用于代码语法高亮
     */
    $hljs: any
  }
}

// 扩展window全局对象类型
declare global {
  interface Window {
    $api: ApiInterface
    hljs: any // highlight.js 全局对象
    // 可以添加其他全局属性
  }
}
