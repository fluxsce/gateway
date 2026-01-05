/**
 * Vite配置文件
 *
 * 本文件定义了Vite构建工具的配置，包括插件、构建选项、开发服务器设置等
 * Vite是一个面向现代浏览器的快速开发构建工具，利用浏览器原生ES模块导入特性
 */

import { fileURLToPath, URL } from 'node:url'
import path from 'path'
import fs from 'fs'

import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue' // Vue 3单文件组件支持
import vueDevTools from 'vite-plugin-vue-devtools' // Vue开发者工具增强插件
import { viteMockServe } from 'vite-plugin-mock' // Mock数据服务插件
import Components from 'unplugin-vue-components/vite' // 组件自动导入
import AutoImport from 'unplugin-auto-import/vite' // API自动导入
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers' // Naive UI组件解析器

/**
 * 安全解析语言文件内容
 * 使用正则替换注释和格式化为有效的JSON字符串
 *
 * 这个函数用于将语言文件的TypeScript对象转换为正确的JSON对象
 * 主要进行以下处理：
 * 1. 移除JS注释（包括单行和多行注释）
 * 2. 删除尾随逗号，使其符合JSON语法
 * 3. 为属性名添加双引号
 * 4. 将单引号替换为双引号
 *
 * @param content 语言文件内容字符串
 * @returns 解析后的对象
 */
function safeParseI18nContent(content: string) {
  try {
    // 移除注释
    const noComments = content.replace(/\/\*[\s\S]*?\*\/|\/\/.*/g, '')
    // 移除可能的尾随逗号以使其成为有效的JSON
    const validJson = noComments
      .replace(/,(\s*[}\]])/g, '$1') // 移除对象或数组结束前的逗号
      .replace(/(\w+):/g, '"$1":') // 将属性名转换为带引号的格式
      .replace(/'/g, '"') // 将单引号替换为双引号

    return JSON.parse(validJson)
  } catch (e) {
    console.error('Error parsing i18n content:', e)
    return {}
  }
}

// https://vite.dev/config/
export default defineConfig(({ command, mode }) => {
  // 根据当前工作目录中的 `mode` 加载 .env 文件
  // 设置第三个参数为 '' 来加载所有环境变量，而不管是否有 `VITE_` 前缀
  const env = loadEnv(mode, process.cwd(), '')

  return {
    /**
     * 基础路径配置
     * 根据环境变量设置baseurl，生产环境为空，开发环境可以设置子路径
     */
    base: env.VITE_BASE_URL || '/',

    /**
     * 插件配置
     * 扩展Vite的功能和集成第三方工具
     */
    plugins: [
      vue(), // 提供Vue 3单文件组件支持
      vueDevTools(), // 增强Vue开发者工具，提供更多调试功能

      /**
       * 组件自动导入配置
       * 使用此插件后，不需要手动import组件，直接在模板中使用即可
       * 例如：<NButton>按钮</NButton> 无需 import { NButton } from 'naive-ui'
       */
      Components({
        resolvers: [NaiveUiResolver()], // 支持Naive UI组件自动导入
        dts: 'src/types/components.d.ts', // 生成类型声明文件，用于TypeScript支持
      }),

      /**
       * API自动导入配置
       * 可以直接使用Vue、Vue Router、Pinia等API，无需手动导入
       * 例如：可以直接使用ref, reactive, computed，无需import { ref } from 'vue'
       */
      AutoImport({
        imports: [
          'vue', // 自动导入Vue Composition API (ref, reactive, computed, watch等)
          'vue-router', // 自动导入Vue Router API (useRouter, useRoute等)
          'pinia', // 自动导入Pinia API (defineStore, storeToRefs等)
          'vue-i18n', // 自动导入Vue I18n API (useI18n等)
          {
            // 自动导入Naive UI组合式API
            'naive-ui': ['useDialog', 'useMessage', 'useNotification', 'useLoadingBar'],
          },
          {
            // 自动导入本地store
            '@/stores/auth': ['useAuthStore'],
            '@/stores/user': ['useUserStore'],
            '@/stores/global': ['useGlobalStore'],
            '@/stores/locale': ['useLocaleStore'],
          },
          {
            // 自动导入自定义hooks
            '@/hooks/useModuleI18n': ['useModuleI18n'],
          },
          {
            // 自动导入API请求
            '@/api/request': ['get', 'post', 'put', 'del'],
          },
        ],
        eslintrc: {
          enabled: true, // 生成ESLint配置，避免未导入的变量报错
        },
        dts: 'src/types/auto-imports.d.ts', // 生成类型声明文件，提供TypeScript支持
      }),

      /**
       * i18n资源路由映射 - 仅在开发环境使用
       *
       * 这部分创建一个虚拟路由，可用于动态访问和预览语言资源文件
       * 使开发者能够在不重启服务的情况下检查翻译内容
       *
       * 访问方式：/@i18n/[locale]/[moduleName]
       * 例如：/@i18n/zh-CN/common 将返回中文下common模块的翻译
       */
      process.env.NODE_ENV === 'development'
        ? {
            name: 'i18n-resource-routes',
            configureServer(server) {
              // 添加虚拟路由，方便在开发阶段查看所有语言资源
              server.middlewares.use((req, res, next) => {
                if (req.url?.startsWith('/@i18n/')) {
                  const locale = req.url.split('/')[2]
                  const moduleName = req.url.split('/')[3]?.split('.')[0]

                  try {
                    // 构建I18n资源映射响应
                    let content = {}

                    // 处理模块特定请求
                    if (moduleName) {
                      const localePath = locale === 'zh-CN' ? 'zh-Cn' : locale
                      const filePath = path.resolve(
                        __dirname,
                        `./src/locales/${localePath}/${moduleName}.ts`,
                      )

                      if (fs.existsSync(filePath)) {
                        const fileContent = fs.readFileSync(filePath, 'utf-8')
                        // 简单解析导出内容 - 生产环境应该使用更健壮的方法
                        const match = fileContent.match(/export\s+default\s+(\{[\s\S]*\})/m)
                        if (match && match[1]) {
                          try {
                            // 使用更安全的方法解析内容，避免eval
                            content = safeParseI18nContent(match[1])
                          } catch (e) {
                            console.error(`Error parsing i18n module: ${moduleName}`, e)
                          }
                        }
                      }
                    }
                    // 获取整个语言包
                    else {
                      const filePath = path.resolve(__dirname, `./src/locales/${locale}.ts`)

                      if (fs.existsSync(filePath)) {
                        const fileContent = fs.readFileSync(filePath, 'utf-8')
                        const match = fileContent.match(/export\s+default\s+(\{[\s\S]*\})/m)
                        if (match && match[1]) {
                          try {
                            // 使用更安全的方法解析内容，避免eval
                            content = safeParseI18nContent(match[1])
                          } catch (e) {
                            console.error(`Error parsing i18n locale: ${locale}`, e)
                          }
                        }
                      }
                    }

                    res.writeHead(200, { 'Content-Type': 'application/json' })
                    res.end(JSON.stringify(content))
                    return
                  } catch (error) {
                    console.error('Error serving i18n resource:', error)
                  }
                }

                next()
              })
            },
          }
        : null,

      /**
       * Mock数据服务配置
       * 用于模拟后端API，在前后端分离开发中非常有用
       * 通过环境变量VITE_USE_MOCK控制是否启用
       */
      viteMockServe({
        // 是否启用
        enable: env.VITE_USE_MOCK === 'true',
        // mock文件存放目录
        mockPath: 'src/mock/modules',
        // 开发环境配置
        logger: true, // 在控制台输出请求日志
      }),
    ],

    /**
     * 路径解析配置
     * 设置路径别名，简化导入语句
     */
    resolve: {
      alias: {
        /**
         * @别名指向src目录
         * 使用示例: import Component from '@/components/Component.vue'
         * 而不是: import Component from '../../components/Component.vue'
         */
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },

    /**
     * 开发服务器配置
     * 如需自定义端口、代理等，可在此配置
     */
    server: {
      // port: 3000, // 自定义端口号
      // open: true, // 自动打开浏览器
      /**
       * API代理配置示例（已注释）
       * 用于开发环境下请求后端API避免跨域问题
       */
      // proxy: {
      //   '/api': {
      //     target: 'http://localhost:8080',
      //     changeOrigin: true,
      //     rewrite: (path) => path.replace(/^\/api/, '')
      //   }
      // }
    },

    /**
     * 构建选项配置
     * 定制项目构建输出
     */
    build: {
      // outDir: 'dist', // 输出目录
      // assetsDir: 'assets', // 静态资源目录
      // minify: 'terser', // 使用terser进行代码压缩
      // terserOptions: { // terser压缩选项
      //   compress: {
      //     drop_console: true, // 移除console
      //     drop_debugger: true // 移除debugger
      //   }
      // },
      /**
       * 分块策略（已注释）
       * 配置代码拆分方式，优化加载性能
       */
      // rollupOptions: {
      //   output: {
      //     manualChunks: {
      //       vendor: ['vue', 'vue-router', 'pinia'],
      //       ui: ['naive-ui']
      //     }
      //   }
      // }
    },
  }
})
