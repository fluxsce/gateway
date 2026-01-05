import { createPinia } from 'pinia'
import { createApp } from 'vue'
import App from './App.vue'
import { setupI18n } from './locales'
import { setupPlugins } from './plugins'
import router from './router'
import { initializeStores, setupStoreHelpers } from './stores'
// 由于在App.vue中已经初始化了主题，这里不需要再次导入和初始化
// import { initTheme } from './utils/theme'

// 配置 highlight.js 样式
import 'highlight.js/styles/atom-one-light.css'

// 配置 vxe-table
import VxeUIBase, { VxeUI } from 'vxe-pc-ui'
import 'vxe-pc-ui/lib/style.css'
import VxeUITable from 'vxe-table'
import 'vxe-table/lib/style.css'

// 配置 vxe-table 插件 - 使用 naive-ui 渲染插件
import VxeUIPluginRenderNaive from '@vxe-ui/plugin-render-naive'
import '@vxe-ui/plugin-render-naive/dist/style.css'

// 配置 vxe-table 右键菜单插件
import VxeUIPluginMenu from '@vxe-ui/plugin-menu'
import '@vxe-ui/plugin-menu/dist/style.css'

//全局样式
import './styles/index.scss'


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

// 异步初始化应用
async function initApp() {
  try {
    // 创建Vue应用实例
    const app = createApp(App)

    // 初始化Pinia状态管理
    const pinia = createPinia()
    app.use(pinia)

    // 初始化i18n，使用简化的方法
    const i18n = setupI18n()
    app.use(i18n)

    // 重要：先初始化stores，再添加路由
    // 初始化所有stores（不再负责多语言）
    await initializeStores()

    // 设置store辅助函数（模板中可通过$user、$app等访问）
    setupStoreHelpers(app)

    // 注册所有自定义插件（包括API工具）
    setupPlugins(app)

    // 主题在App.vue组件中初始化，避免重复初始化
    // initTheme()

    // 使用插件 - 结合 naive-ui 使用
    VxeUI.use(VxeUIPluginRenderNaive)
    
    // 使用右键菜单插件
    VxeUI.use(VxeUIPluginMenu)

    // 配置 vxe-table（必须在路由之前注册）
    app.use(VxeUIBase).use(VxeUITable)

    // 使用路由
    app.use(router)

    // 安装所有插件后再挂载应用
    app.mount('#app')

    console.log('应用初始化完成')

    return app
  } catch (error) {
    console.error('应用初始化过程中发生错误:', error)
    throw error
  }
}

// 启动应用
initApp().catch((err) => {
  console.error('应用初始化失败:', err)

  // 显示友好的错误信息到页面
  const rootEl = document.getElementById('app')
  if (rootEl) {
    rootEl.innerHTML = `
      <div style="padding: 20px; text-align: center; color: #666;">
        <h2>应用加载失败</h2>
        <p>请刷新页面或联系管理员</p>
        <p style="font-size: 12px; margin-top: 10px;">错误信息: ${err.message || '未知错误'}</p>
      </div>
    `
  }
})
