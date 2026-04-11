/**
 * 工具函数集合
 * 作为统一导出的入口文件，集中管理所有工具函数
 * 使用时只需从 @/utils 导入，无需关心具体实现文件
 * 例如: import { formatDate, isValidEmail } from '@/utils'
 */

// 导出存储相关工具函数
export * from './storage'

// 导出格式化相关工具函数
export * from './format'

// 导出验证相关工具函数
export * from './validate'

// 导出日志相关工具函数
export * from './logger'

// 导出图标相关工具函数
export * from './icon'
// 渲染图标 VNode 已迁移至 GIcon，此处兼容旧导入
export { renderIconVNode } from '@/components/gicon'

// 导出剪贴板相关工具函数
export * from './clipboard'

// DOM 唯一 id 生成（如弹窗挂载容器）
export { createDomId } from './messageUtil'

