import { describe, expect, it } from 'vitest'
import {
  browserBaseBootScript,
  docsBaseBootScript,
  rewriteDocsBaseLiterals,
} from '../../../scripts/browserBaseInline'
import { detectBrowserBase } from '../browserBase'

/**
 * 在给定地址栏下执行页头脚本。
 * @param script - 含 script 标签的页头脚本
 * @param pathname - location.pathname
 * @returns 脚本写入的 window 与 document.write 结果
 */
function runBootScript(script: string, pathname: string): { base?: string; docs?: string; written: string } {
  const body = script.replace(/^<script>/, '').replace(/<\/script>$/, '')
  const window: { __GW_BROWSER_BASE__?: string; __GW_DOCS_BASE__?: string } = {}
  let written = ''
  const document = {
    write(html: string) {
      written += html
    },
  }
  const location = { pathname }
  const run = new Function('window', 'document', 'location', body)
  run(window, document, location)
  return { base: window.__GW_BROWSER_BASE__, docs: window.__GW_DOCS_BASE__, written }
}

describe('detectBrowserBase', () => {
  it('keeps the outer pro or test segment on one domain', () => {
    expect(detectBrowserBase('/gatewayweb', '/test/gatewayweb/login')).toBe('/test/gatewayweb')
    expect(detectBrowserBase('/gatewayweb', '/pro/gatewayweb')).toBe('/pro/gatewayweb')
    expect(detectBrowserBase('/gatewayweb', '/gatewayweb/hub0002')).toBe('/gatewayweb')
  })

  it('falls back to the mount prefix when the address has no console path', () => {
    expect(detectBrowserBase('/gatewayweb', '/')).toBe('/gatewayweb')
  })

  it('boot script writes the outer segment into base and window', () => {
    const testPage = runBootScript(browserBaseBootScript('/gatewayweb'), '/test/gatewayweb/login')
    expect(testPage.base).toBe('/test/gatewayweb')
    expect(testPage.written).toBe('<base href="/test/gatewayweb/">')

    const proPage = runBootScript(browserBaseBootScript('/gatewayweb'), '/pro/gatewayweb')
    expect(proPage.base).toBe('/pro/gatewayweb')
    expect(proPage.written).toBe('<base href="/pro/gatewayweb/">')

    const docsPage = runBootScript(docsBaseBootScript('/gatewayweb'), '/test/gatewayweb/docs/modules/hub0002.html')
    expect(docsPage.docs).toBe('/test/gatewayweb/docs/')
    expect(docsPage.written).toBe('<base href="/test/gatewayweb/docs/">')
  })
})

describe('rewriteDocsBaseLiterals', () => {
  const docsBase = '/gatewayweb/docs/'

  it('rewrites script strings and leaves JSON.parse payloads intact', () => {
    expect(rewriteDocsBaseLiterals('const n="/gatewayweb/docs/";', docsBase, true)).toBe(
      'const n=window.__GW_DOCS_BASE__;',
    )
    expect(rewriteDocsBaseLiterals('const s="/gatewayweb/docs/assets/a.png"', docsBase, true)).toBe(
      'const s="assets/a.png"',
    )
    expect(rewriteDocsBaseLiterals('return"/gatewayweb/docs/"+e', docsBase, true)).toBe(
      'return window.__GW_DOCS_BASE__+e',
    )
    expect(rewriteDocsBaseLiterals('const n="/gatewayweb/docs/";', docsBase, false)).toBe(
      'const n="/gatewayweb/docs/";',
    )
    const siteData = 'JSON.parse(\'{"base":"/gatewayweb/docs/"}\')'
    expect(rewriteDocsBaseLiterals(siteData, docsBase, true)).toBe(siteData)
    const escaped = 'JSON.parse("{\\"base\\":\\"/gatewayweb/docs/\\"}")'
    expect(rewriteDocsBaseLiterals(escaped, docsBase, true)).toBe(escaped)
    expect(rewriteDocsBaseLiterals('href=\\"/gatewayweb/docs/modules/hub0000.html\\"', docsBase, true)).toBe(
      'href=\\"./modules/hub0000.html\\"',
    )
  })
})
