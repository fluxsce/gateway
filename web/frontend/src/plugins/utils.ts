/**
 * 工具函数插件
 * 将utils包中的工具函数注册为全局可用的插件
 */
import type { App } from 'vue'
import * as storageUtils from '@/utils/storage'
import * as formatUtils from '@/utils/format'
import * as validateUtils from '@/utils/validate'

// 所有工具函数的集合
const utils = {
  // 存储相关工具
  storage: {
    ...storageUtils
  },
  // 格式化相关工具
  format: {
    ...formatUtils
  },
  // 验证相关工具
  validate: {
    ...validateUtils
  }
}

/**
 * 工具函数插件
 * 注册全局工具函数
 */
export default {
  install(app: App) {
    // 注册为全局属性，可通过this.$utils访问
    app.config.globalProperties.$utils = utils
    
    // 提供依赖注入方式访问
    app.provide('$utils', utils)
  }
} 