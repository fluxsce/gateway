/**
 * 全局状态管理
 * 简化的扁平化设计
 */
import { defineStore } from 'pinia'
import { getLocalStorage, setLocalStorage } from '../utils/storage'

export const useGlobalStore = defineStore('global', {
  state: (): GlobalState => ({
    // 应用信息
    appName: 'Gateway',
    appVersion: '',
    
    // 页面状态
    pageTitle: '',
    pageLoading: false,
    breadcrumbs: [],
    
    // UI状态
    showSidebar: getLocalStorage<boolean>('showSidebar', true),
  }),

  actions: {
    /**
     * 设置应用版本
     */
    setAppVersion(version: string) {
      this.appVersion = version
    },

    /**
     * 设置页面标题
     */
    setPageTitle(title: string) {
      this.pageTitle = title
      document.title = title ? `${title} - ${this.appName}` : this.appName
    },

    /**
     * 设置面包屑
     */
    setBreadcrumbs(breadcrumbs: BreadcrumbItem[]) {
      this.breadcrumbs = breadcrumbs
    },

    /**
     * 设置页面加载状态
     */
    setPageLoading(loading: boolean) {
      this.pageLoading = loading
    },

    /**
     * 切换侧边栏显示
     */
    toggleSidebar() {
      this.showSidebar = !this.showSidebar
      setLocalStorage('showSidebar', this.showSidebar)
    },
  },
})

// 面包屑项接口
export interface BreadcrumbItem {
  title: string
  path: string
}

// 全局状态定义
interface GlobalState {
  // 应用信息
  appName: string
  appVersion: string
  
  // 页面状态
  pageTitle: string
  pageLoading: boolean
  breadcrumbs: BreadcrumbItem[]
  
  // UI状态
  showSidebar: boolean
}