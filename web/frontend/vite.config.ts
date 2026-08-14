/**
 * Vite 配置：插件、路径别名、开发代理与生产分包。
 */

import fs from 'fs'
import { createRequire } from 'node:module'
import { dirname } from 'node:path'
import { fileURLToPath, URL } from 'node:url'
import path from 'path'

import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { defineConfig, loadEnv } from 'vite'
import { viteMockServe } from 'vite-plugin-mock'
import vueDevTools from 'vite-plugin-vue-devtools'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const require = createRequire(import.meta.url)

/** link / file 安装的 niuma-ui 根目录，供 Vite server.fs.allow 与 HMR */
let niumaUiRoot = path.resolve(__dirname, '../../../../shangijan/niuma-ui')
try {
  niumaUiRoot = dirname(require.resolve('niuma-ui/package.json'))
} catch {
  // 保持上面的相对路径回退
}

/**
 * 将语言文件中的 TypeScript 对象字面量转成可 JSON.parse 的字符串。
 * 处理步骤：去掉注释、去掉尾随逗号、给属性名加双引号、单引号改双引号。
 * @param content 语言文件原文
 * @returns 解析后的对象；失败时返回空对象
 */
function safeParseI18nContent(content: string) {
  try {
    const noComments = content.replace(/\/\*[\s\S]*?\*\/|\/\/.*/g, '')
    const validJson = noComments
      .replace(/,(\s*[}\]])/g, '$1')
      .replace(/(\w+):/g, '"$1":')
      .replace(/'/g, '"')

    return JSON.parse(validJson)
  } catch (e) {
    console.error('Error parsing i18n content:', e)
    return {}
  }
}

/**
 * 开发环境下将帮助手册路径代理到独立运行的 VitePress（与 getDocsSitePath() 规则一致）。
 * 另开终端执行 `npm run docs:dev`（默认 5274 端口，与 package.json 的 dev:docs 保持一致），否则 iframe 会 404。
 *
 * 生产发版：`npm run build` 中的 `build-only` 在 `vite build` 之后执行 `vitepress build docs`。
 * 文档输出到 `dist/docs/`（见 `docs/.vitepress/config.ts` 的 `outDir`），与主应用一并部署即可。
 */
function resolveDocsDevProxy(env: Record<string, string>) {
  const target = (env.VITE_DOCS_DEV_TARGET || 'http://127.0.0.1:5274').replace(/\/+$/, '')
  let base = (env.VITE_BASE_URL || '').trim()
  if (base === '/' || base === '') {
    return {
      '^/docs': { target, changeOrigin: true, ws: true },
    }
  }
  if (!base.startsWith('/')) base = `/${base}`
  base = base.replace(/\/+$/, '')
  const prefix = `${base}/docs`
  return {
    [`^${prefix}`]: { target, changeOrigin: true, ws: true },
  }
}

/**
 * 将 node_modules 按依赖族拆成独立 chunk，减轻首包 JS 体积并利于缓存。
 * 顺序靠前的规则优先匹配。
 */
function resolveManualChunk(id: string): string | undefined {
  if (!id.includes('node_modules')) return undefined
  // Rollup 在 Windows 下 id 可能含反斜杠，统一后再匹配
  const m = id.replace(/\\/g, '/')
  if (m.includes('/niuma-ui') || m.includes('/reka-ui') || m.includes('/@lucide/') || m.includes('/vue-sonner'))
    return 'niuma-ui'
  if (m.includes('/echarts') || m.includes('/zrender')) return 'echarts'
  if (m.includes('/@antv')) return 'antv'
  if (m.includes('/@codemirror') || m.includes('/codemirror/')) return 'codemirror'
  if (m.includes('/@tiptap') || m.includes('/prosemirror')) return 'tiptap'
  if (m.includes('/highlight.js')) return 'highlight'
  if (m.includes('/@intlify')) return 'vue-i18n'
  if (m.includes('/vue-i18n')) return 'vue-i18n'
  if (m.includes('/vue-router')) return 'vue-router'
  if (m.includes('/pinia')) return 'pinia'
  if (m.includes('/axios')) return 'axios'
  if (m.includes('/@vicons')) return 'vicons'
  if (m.includes('/cron-parser')) return 'cron-parser'
  if (m.includes('/async-validator')) return 'async-validator'
  if (m.includes('/node_modules/@vue/') || m.includes('/node_modules/vue/')) return 'vue'
  return 'vendor'
}

export default defineConfig(({ command, mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    /** 根据环境变量设置 base，生产环境可为空，开发环境可设子路径 */
    base: env.VITE_BASE_URL || '/',

    plugins: [
      vue(),
      // niuma-ui/styles.css 使用 @import "tailwindcss"（v4），需 Vite 插件解析
      tailwindcss(),
      // 仅开发服务注册 DevTools，避免生产构建携带调试相关逻辑
      ...(command === 'serve' ? [vueDevTools()] : []),

      /**
       * 组件自动导入（本地 G* 等；UI 基座请从 `@/ui` / `niuma-ui` 显式导入）。
       */
      Components({
        dts: 'src/types/components.d.ts',
      }),

      /**
       * API 自动导入：可直接使用 Vue、Vue Router、Pinia 等 API，无需手动 import。
       */
      AutoImport({
        imports: [
          'vue',
          'vue-router',
          'pinia',
          'vue-i18n',
          {
            '@/composables/useAppMessage': ['useAppMessage'],
          },
          {
            '@/stores/auth': ['useAuthStore'],
            '@/stores/user': ['useUserStore'],
            '@/stores/global': ['useGlobalStore'],
            '@/stores/locale': ['useLocaleStore'],
          },
          {
            '@/hooks/useModuleI18n': ['useModuleI18n'],
          },
          {
            '@/api/request': ['get', 'post', 'put', 'del'],
          },
        ],
        eslintrc: {
          enabled: true,
        },
        dts: 'src/types/auto-imports.d.ts',
      }),

      /**
       * 开发环境虚拟路由，用于预览语言资源，无需重启服务。
       * 访问方式：/@i18n/[locale]/[moduleName]，例如 /@i18n/zh-CN/common。
       */
      process.env.NODE_ENV === 'development'
        ? {
            name: 'i18n-resource-routes',
            configureServer(server) {
              server.middlewares.use((req, res, next) => {
                if (req.url?.startsWith('/@i18n/')) {
                  const locale = req.url.split('/')[2]
                  const moduleName = req.url.split('/')[3]?.split('.')[0]

                  try {
                    let content = {}

                    if (moduleName) {
                      const localePath = locale === 'zh-CN' ? 'zh-Cn' : locale
                      const filePath = path.resolve(
                        __dirname,
                        `./src/locales/${localePath}/${moduleName}.ts`,
                      )

                      if (fs.existsSync(filePath)) {
                        const fileContent = fs.readFileSync(filePath, 'utf-8')
                        const match = fileContent.match(/export\s+default\s+(\{[\s\S]*\})/m)
                        if (match && match[1]) {
                          try {
                            content = safeParseI18nContent(match[1])
                          } catch (e) {
                            console.error(`Error parsing i18n module: ${moduleName}`, e)
                          }
                        }
                      }
                    } else {
                      const filePath = path.resolve(__dirname, `./src/locales/${locale}.ts`)

                      if (fs.existsSync(filePath)) {
                        const fileContent = fs.readFileSync(filePath, 'utf-8')
                        const match = fileContent.match(/export\s+default\s+(\{[\s\S]*\})/m)
                        if (match && match[1]) {
                          try {
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
       * Mock 数据服务，通过环境变量 VITE_USE_MOCK 控制是否启用。
       */
      viteMockServe({
        enable: env.VITE_USE_MOCK === 'true',
        mockPath: 'src/mock/modules',
        logger: true,
      }),
    ],

    resolve: {
      alias: {
        /** `@` 指向 src，例如 `import X from '@/components/X.vue'` */
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
      /**
       * 确保 CodeMirror 相关包使用单一实例，避免动态导入出现多实例。
       */
      dedupe: [
        '@codemirror/state',
        '@codemirror/view',
        '@codemirror/language',
        '@codemirror/lang-javascript',
        '@codemirror/lang-json',
        '@codemirror/lang-html',
        '@codemirror/lang-css',
        '@codemirror/lang-xml',
        '@codemirror/lang-sql',
        '@codemirror/lang-yaml',
        '@codemirror/lang-markdown',
        '@codemirror/lang-python',
        '@codemirror/lang-java',
        '@codemirror/lang-go',
        '@codemirror/lang-rust',
      ],
    },

    server: {
      fs: {
        allow: ['.', ...(niumaUiRoot ? [niumaUiRoot] : [])],
      },
      /**
       * 开发：帮助手册（VitePress）由另一进程提供，经代理挂到与主应用相同的 `/gatewayweb/docs/` 下，
       * 便于 MainLayoutHeader iframe 同源加载。见 `.env.development` 的 `VITE_DOCS_DEV_TARGET`。
       */
      ...(command === 'serve' ? { proxy: resolveDocsDevProxy(env) } : {}),
    },

    build: {
      /** niuma-ui / echarts / vicons 等 chunk 常超 500kB，提高阈值避免误报 */
      chunkSizeWarningLimit: 1200,
      /**
       * 按依赖族拆分 node_modules，缩小首屏入口 chunk、提升缓存命中率。
       */
      rollupOptions: {
        output: {
          manualChunks: resolveManualChunk,
        },
      },
    },
  }
})
