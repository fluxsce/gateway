import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { defineConfig } from 'vitepress'
import { loadEnv } from 'vite'

const vitepressDir = path.dirname(fileURLToPath(import.meta.url))
/** 前端包根目录（含 .env*），避免 VitePress 解析配置时 cwd 非 web/frontend 导致 base 错误 */
const frontendRoot = path.resolve(vitepressDir, '../..')

/**
 * 文档站点 base：与主应用 `.env` 中 `VITE_BASE_URL` 对齐，并固定挂在 `/docs/` 子路径下，
 * 避免与主应用根路径冲突（例如主应用为 `/gatewayweb/` 时，文档为 `/gatewayweb/docs/`）。
 */
function resolveDocsBase(mode: string): string {
  const env = loadEnv(mode, frontendRoot, '')
  let p = (env.VITE_BASE_URL ?? '').trim()
  if (p === '/' || p === '') {
    return '/docs/'
  }
  if (!p.startsWith('/')) {
    p = `/${p}`
  }
  p = p.replace(/\/+$/, '')
  return `${p}/docs/`
}

const vitepressEnvMode = process.env.NODE_ENV === 'production' ? 'production' : 'development'

/**
 * 模块侧栏分组：与 `src/router/layoutRouteRegistry.ts` 中 `GATEWAY_LAYOUT_ROUTE_TREE`
 * 的分组与顺序一致；展示文案使用路由 `meta.title`（不含 moduleName）。
 */
const moduleSidebarGroups = [
  {
    text: '入门与总览',
    items: [
      { text: '控制台导航与菜单', link: '/modules/navigation' },
      { text: '系统监控', link: '/modules/hub0000' },
    ],
  },
  {
    text: '系统设置',
    collapsed: false,
    items: [
      { text: '用户管理', link: '/modules/hub0002' },
      { text: '角色管理', link: '/modules/hub0005' },
      { text: '权限资源管理', link: '/modules/hub0006' },
      { text: '系统节点监控', link: '/modules/hub0007' },
      { text: '集群节点事件', link: '/modules/hub0008' },
    ],
  },
  {
    text: '网关管理',
    collapsed: false,
    items: [
      { text: '实例管理', link: '/modules/hub0020' },
      { text: '代理管理', link: '/modules/hub0022' },
      { text: '路由管理', link: '/modules/hub0021' },
      { text: '网关日志管理', link: '/modules/hub0023' },
    ],
  },
  {
    text: '隧道管理',
    collapsed: false,
    items: [
      { text: '隧道服务器', link: '/modules/hub0060' },
      { text: '静态映射', link: '/modules/hub0061' },
      { text: '隧道客户端', link: '/modules/hub0062' },
    ],
  },
  {
    text: '服务治理',
    collapsed: false,
    items: [
      { text: '服务中心实例管理', link: '/modules/hub0040' },
      { text: '命名空间管理', link: '/modules/hub0041' },
      { text: '服务列表', link: '/modules/hub0042' },
      { text: '配置中心', link: '/modules/hub0043' },
    ],
  },
  {
    text: '预警管理',
    collapsed: false,
    items: [
      { text: '预警服务配置', link: '/modules/hub0080' },
      { text: '预警模板管理', link: '/modules/hub0081' },
      { text: '预警日志管理', link: '/modules/hub0082' },
    ],
  },
]

export default defineConfig({
  base: resolveDocsBase(vitepressEnvMode),
  /**
   * 构建产物写入主应用 dist，与 `npm run build` 中 `vite build` 输出同目录，
   * 部署后 iframe / 新窗口访问的 `/…/docs/` 与 `getDocsSitePath()` 一致。
   */
  outDir: path.join(frontendRoot, 'dist', 'docs'),
  title: 'FLUX Gateway 软件使用说明',
  description: 'Gateway Web 控制台操作说明（与控制台菜单分组一致）',
  lang: 'zh-CN',
  appearance: true,
  lastUpdated: true,
  head: [
    ['meta', { name: 'theme-color', content: '#0f172a' }],
    ['meta', { name: 'keywords', content: 'FLUX Gateway, API 网关, 控制台, 文档' }],
  ],
  markdown: {
    lineNumbers: true,
  },
  themeConfig: {
    siteTitle: 'FLUX Gateway',
    outline: {
      label: '本页目录',
      level: [2, 3],
    },
    docFooter: {
      prev: '上一篇',
      next: '下一篇',
    },
    nav: [
      { text: '首页', link: '/' },
      { text: '模块说明', link: '/modules/navigation' },
    ],
    sidebar: {
      '/modules/': moduleSidebarGroups,
    },
    search: {
      provider: 'local',
      options: {
        translations: {
          button: {
            buttonText: '搜索文档',
            buttonAriaLabel: '搜索文档',
          },
          modal: {
            noResultsText: '没有找到相关结果',
            resetButtonTitle: '清除查询条件',
            backButtonTitle: '关闭搜索',
            displayDetails: '查看详情',
            footer: {
              selectText: '跳转',
              selectKeyAriaLabel: '按回车跳转',
              navigateText: '切换',
              navigateUpKeyAriaLabel: '上一个结果',
              navigateDownKeyAriaLabel: '下一个结果',
              closeText: '关闭',
              closeKeyAriaLabel: '按 Esc 关闭',
            },
          },
        },
      },
    },
    socialLinks: [{ icon: 'github', link: 'https://github.com/fluxsce/gateway' }],
    footer: {
      message: 'FLUX Gateway · 企业级 API 网关',
      copyright: '文档与控制台菜单对齐，具体以实际部署版本为准',
    },
  },
})
