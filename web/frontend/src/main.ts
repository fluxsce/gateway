import { removeBootSplash } from './bootSplash'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import { createApp, nextTick } from 'vue'
import App from './App.vue'
import { setupI18n } from './locales'
import { setupPlugins } from './plugins'
import router from './router'
import { initializeStores, setupStoreHelpers } from './stores'

import 'highlight.js/styles/atom-one-light.css'

// niuma-ui 基座样式 → 业务主题 → 品牌覆盖
import 'niuma-ui/styles.css'
import './styles/index.scss'
import './styles/rs-brand.css'

// 配置被动事件监听器以提高滚动性能（如无需要可注释掉）
// import { setupPassiveEvents } from './utils/passive-events'
// setupPassiveEvents({
//   enabled: true,
//   verbose: import.meta.env.DEV, // 开发环境显示日志
//   excludeSelectors: [
//     // 项目特定的排除选择器
//     '.custom-chart',
//     '.interactive-map'
//   ]
// })

/**
 * startApp 启动完整 Vue 应用（由 `boot.ts` 动态导入触发）。
 */
export async function startApp() {
  try {
    // 创建Vue应用实例
    const app = createApp(App)

    // 初始化Pinia状态管理
    const pinia = createPinia()
    pinia.use(piniaPluginPersistedstate)
    app.use(pinia)

    // 初始化i18n，使用简化的方法
    const i18n = setupI18n()
    app.use(i18n)

    // 重要：先初始化stores，再添加路由
    // 初始化所有stores（含主题同步 data-theme，与 XiRang 一致由 store 维护）
    await initializeStores()

    // 设置store辅助函数（模板中可通过$user、$app等访问）
    setupStoreHelpers(app)

    setupPlugins(app)

    app.use(router)

    // 等待路由完成初始导航（含异步路由组件 chunk），避免挂载后短暂空白
    await router.isReady()

    // 安装所有插件后再挂载应用
    app.mount('#app')
    await nextTick()
    removeBootSplash()

    console.log('应用初始化完成')

    return app
  } catch (error) {
    console.error('应用初始化过程中发生错误:', error)
    removeBootSplash()
    throw error
  }
}
