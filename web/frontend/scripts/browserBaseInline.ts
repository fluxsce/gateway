/**
 * 生产 index.html 页头脚本。
 * 在模块脚本之前根据地址栏插入 <base>，让相对资源落到 /test/gatewayweb 或 /pro/gatewayweb。
 * 判断规则与 src/utils/browserBase.ts 的 detectBrowserBase 保持一致。
 */

/**
 * 生成控制台页头脚本。
 * @param mount - web.frontend.prefix，例如 /gatewayweb
 * @returns 可插入 head 开头的 script 标签
 */
export function browserBaseBootScript(mount: string): string {
  const marker = JSON.stringify(mount)
  return `<script>(function(){var mount=${marker};var path=location.pathname||"/";var base=mount;var from=0;while(from<path.length){var i=path.indexOf(mount,from);if(i<0)break;var next=path.charAt(i+mount.length);if(next===""||next==="/"){base=path.slice(0,i+mount.length);break;}from=i+mount.length;}if(base.charAt(base.length-1)!=="/")base+="/";document.write("<base href=\\""+base+"\\">");window.__GW_BROWSER_BASE__=base.slice(0,-1);})();</script>`
}

/**
 * 生成帮助手册页头脚本。文档的 base 是控制台路径再加 /docs/。
 * @param mount - web.frontend.prefix
 * @returns 可插入 head 开头的 script 标签
 */
export function docsBaseBootScript(mount: string): string {
  const marker = JSON.stringify(mount)
  return `<script>(function(){var mount=${marker};var path=location.pathname||"/";var base=mount;var from=0;while(from<path.length){var i=path.indexOf(mount,from);if(i<0)break;var next=path.charAt(i+mount.length);if(next===""||next==="/"){base=path.slice(0,i+mount.length);break;}from=i+mount.length;}var docs=(base==="/"?"/docs/":base.replace(/\\/$/,"")+"/docs/");window.__GW_DOCS_BASE__=docs;document.write("<base href=\\""+docs+"\\">");})();</script>`
}

/**
 * 手册站点数据在 body 末尾才赋值。模块脚本执行前把 base 改成地址栏上的手册路径。
 * @returns 放在 body 结束前的 script 标签
 */
export function docsSiteDataPatchScript(): string {
  return `<script>(function(){var docs=window.__GW_DOCS_BASE__;if(typeof docs==="string"&&window.__VP_SITE_DATA__){window.__VP_SITE_DATA__.base=docs;}})();</script>`
}

/**
 * 只改顶层字符串字面量。包在 JSON.parse 里的站点数据保持原样，避免把 JSON 截断。
 * 图片等更长路径改成相对手册根目录，服务端渲染不用 window。
 * 刚好等于手册 base 的字符串只在浏览器包里改成 window.__GW_DOCS_BASE__。
 * @param code - 打包后的脚本
 * @param docsBase - 打包时的手册 base，例如 /gatewayweb/docs/
 * @param useRuntimeBase - 浏览器包为 true，服务端包为 false
 * @returns 改写后的脚本
 */
export function rewriteDocsBaseLiterals(code: string, docsBase: string, useRuntimeBase: boolean): string {
  if (docsBase === '') return code
  let out = ''
  let index = 0
  while (index < code.length) {
    const quote = code[index]
    if (quote !== '"' && quote !== "'" && quote !== '`') {
      out += quote
      index++
      continue
    }
    if (index > 0 && code[index - 1] === '\\') {
      out += quote
      index++
      continue
    }
    const end = endOfJsString(code, index)
    const literal = code.slice(index + 1, end)
    if (quote !== '`' && literal.startsWith(docsBase)) {
      const rest = literal.slice(docsBase.length)
      if (rest === '') {
        if (useRuntimeBase) {
          const previous = out.length > 0 ? out[out.length - 1] : ''
          const expression = 'window.__GW_DOCS_BASE__'
          out += /[A-Za-z0-9_$]/.test(previous) ? ` ${expression}` : expression
        } else {
          out += code.slice(index, end + 1)
        }
      } else {
        out += JSON.stringify(rest)
      }
    } else {
      out += code.slice(index, end + 1)
    }
    index = Math.min(end + 1, code.length)
  }
  if (!useRuntimeBase) return out
  return relativizeEmbeddedDocsHrefs(replaceExactDocsBase(out, docsBase), docsBase)
}

/**
 * 把刚好等于手册 base 的脚本字符串改成运行时路径。
 * 前面是反斜杠的不算，避免改到 JSON 里的转义引号。
 * @param code - 脚本
 * @param docsBase - 打包时的手册 base
 * @returns 改写后的脚本
 */
function replaceExactDocsBase(code: string, docsBase: string): string {
  const escaped = docsBase.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const exact = new RegExp(`(?<!\\\\)(["'])${escaped}\\1`, 'g')
  return code.replace(exact, (match, _quote: string, offset: number, source: string) => {
    const previous = offset > 0 ? source[offset - 1] : ''
    if (previous === ':') return match
    const expression = 'window.__GW_DOCS_BASE__'
    return /[A-Za-z0-9_$]/.test(previous) ? ` ${expression}` : expression
  })
}

/**
 * 手册正文里写死的绝对链接改成相对手册根目录，交给页头 base 解析。
 * @param code - 脚本
 * @param docsBase - 打包时的手册 base
 * @returns 改写后的脚本
 */
function relativizeEmbeddedDocsHrefs(code: string, docsBase: string): string {
  const pairs = [
    [`="${docsBase}`, '="./'],
    [`='${docsBase}`, "='./"],
    [`=\\"${docsBase}`, '=\\"./'],
    [`=\\'${docsBase}`, "=\\'./"],
  ]
  let next = code
  for (const [from, to] of pairs) {
    next = next.split(from).join(to)
  }
  return next
}

/**
 * 找到字符串结束引号。未闭合时停在末尾。
 * @param code - 脚本
 * @param start - 起始引号下标
 * @returns 结束引号下标；未闭合时为最后一个字符下标
 */
function endOfJsString(code: string, start: number): number {
  const quote = code[start]
  for (let index = start + 1; index < code.length; index++) {
    if (code[index] === '\\') {
      index++
      continue
    }
    if (code[index] === quote) return index
  }
  return code.length - 1
}
