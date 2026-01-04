/**
 * 环境声明文件
 * 用于定义Vite项目的全局类型声明和模块声明
 */

/// <reference types="vite/client" />  // 引用Vite客户端类型定义，提供import.meta.env等Vite特有API的类型支持

/**
 * Vue单文件组件(.vue)的类型声明
 * 使TypeScript能够识别.vue文件作为Vue组件模块
 * 这允许在TypeScript中直接导入Vue单文件组件
 */
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  // DefineComponent是Vue 3中定义组件类型的工具类型
  // 这里将组件定义为接受任意props和emits的组件类型
  const component: DefineComponent<{}, {}, any>
  export default component
}
