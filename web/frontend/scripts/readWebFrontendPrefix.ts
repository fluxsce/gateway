/**
 * 从仓库 configs/web.yaml 读取控制台静态资源前缀。
 * Vite 与 VitePress 启动、打包时调用，避免与 .env 里的 VITE_BASE_URL 各写一份。
 */
import fs from 'node:fs'
import path from 'node:path'

/**
 * 读取 web.frontend.prefix，并规范成 Vite base 可用的路径。
 * 根路径返回 `/`；其余去掉尾斜杠，例如 `/gatewayweb`。
 * 文件缺失或没有该字段时回落 `/`，与 Go 侧 `web.frontend.prefix` 默认值一致。
 * @param frontendRoot - `web/frontend` 目录的绝对路径
 * @returns 控制台公共路径前缀
 */
export function readWebFrontendPrefix(frontendRoot: string): string {
  const yamlPath = path.resolve(frontendRoot, '../../configs/web.yaml')
  let text = ''
  try {
    text = fs.readFileSync(yamlPath, 'utf8')
  } catch (error) {
    console.warn(`[frontend-prefix] 无法读取 ${yamlPath}，前缀回落 /`, error)
    return '/'
  }
  const raw = readFrontendPrefix(text)
  if (raw == null) {
    console.warn(`[frontend-prefix] ${yamlPath} 未找到 web.frontend.prefix，前缀回落 /`)
    return '/'
  }
  return normalizePublicPrefix(raw)
}

/**
 * 只取 web 下 frontend.prefix，忽略 static.prefix 等其它 prefix。
 * @param text - web.yaml 原文
 * @returns 原始标量；未找到时返回 undefined
 */
function readFrontendPrefix(text: string): string | undefined {
  let webIndent = -1
  let frontendIndent = -1
  for (const rawLine of text.split(/\r?\n/)) {
    const trimmed = rawLine.trim()
    if (trimmed === '' || trimmed.startsWith('#')) continue
    const indent = rawLine.length - rawLine.trimStart().length
    const body = stripInlineComment(trimmed)

    if (webIndent >= 0 && indent <= webIndent) {
      webIndent = -1
      frontendIndent = -1
    }
    if (frontendIndent >= 0 && indent <= frontendIndent) {
      frontendIndent = -1
    }

    if (indent === 0 && /^web\s*:/.test(body)) {
      webIndent = indent
      frontendIndent = -1
      continue
    }
    if (webIndent >= 0 && indent > webIndent && /^frontend\s*:/.test(body)) {
      frontendIndent = indent
      continue
    }
    if (frontendIndent >= 0 && indent > frontendIndent && /^prefix\s*:/.test(body)) {
      return unquote(body.replace(/^prefix\s*:\s*/, '').trim())
    }
  }
  return undefined
}

/**
 * 去掉行尾注释。引号内的 # 保留。
 * @param body - 去掉缩进后的一行
 * @returns 不含行尾注释的内容
 */
function stripInlineComment(body: string): string {
  let quote: '"' | "'" | '' = ''
  for (let i = 0; i < body.length; i++) {
    const ch = body[i]
    if (quote) {
      if (ch === quote) quote = ''
      continue
    }
    if (ch === '"' || ch === "'") {
      quote = ch
      continue
    }
    if (ch === '#' && (i === 0 || /\s/.test(body[i - 1] ?? ''))) {
      return body.slice(0, i).trim()
    }
  }
  return body
}

/**
 * 去掉成对的单引号或双引号。
 * @param value - YAML 标量原文
 * @returns 引号内的文本；无引号时原样返回
 */
function unquote(value: string): string {
  if (
    (value.startsWith('"') && value.endsWith('"') && value.length >= 2) ||
    (value.startsWith("'") && value.endsWith("'") && value.length >= 2)
  ) {
    return value.slice(1, -1)
  }
  return value
}

/**
 * 规范公共路径：补前导斜杠，非根路径去掉尾斜杠。
 * @param raw - web.yaml 中的 prefix
 * @returns `/` 或 `/gatewayweb` 这种形式
 */
function normalizePublicPrefix(raw: string): string {
  let p = raw.trim()
  if (p === '' || p === '/') return '/'
  if (!p.startsWith('/')) p = `/${p}`
  return p.replace(/\/+$/, '')
}
