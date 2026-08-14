#!/usr/bin/env node
/**
 * niuma-ui@1.1.1 用 import.meta.glob('../../node_modules/@lucide/vue/...') 收集图标。
 * pnpm 把 @lucide/vue 放在与 niuma-ui 同级的虚拟 store，该 glob 匹配 0 文件，打包后 RsIcon 全空。
 * 在 niuma-ui 包内补上指向真实 @lucide/vue 的符号链接，让旧版 glob 能扫到图标。
 */
'use strict'

const { createRequire } = require('module')
const { dirname, join } = require('path')
const fs = require('fs')

const frontendRoot = process.cwd()
const hostRequire = createRequire(join(frontendRoot, 'package.json'))

let niumaRoot
try {
  niumaRoot = dirname(hostRequire.resolve('niuma-ui/package.json'))
} catch {
  console.warn('[ensure-niuma-lucide] niuma-ui not installed, skip')
  process.exit(0)
}

const niumaRequire = createRequire(join(niumaRoot, 'package.json'))
let lucideRoot
try {
  lucideRoot = dirname(niumaRequire.resolve('@lucide/vue/package.json'))
} catch {
  console.error('[ensure-niuma-lucide] @lucide/vue not resolvable from niuma-ui')
  process.exit(1)
}

const dest = join(niumaRoot, 'node_modules', '@lucide', 'vue')
const iconsDir = join(dest, 'dist', 'esm', 'icons')
if (fs.existsSync(iconsDir)) {
  console.log('[ensure-niuma-lucide] lucide icons already at', iconsDir)
  process.exit(0)
}

fs.mkdirSync(dirname(dest), { recursive: true })
if (fs.existsSync(dest)) {
  fs.rmSync(dest, { recursive: true, force: true })
}
fs.symlinkSync(lucideRoot, dest, 'junction')
console.log('[ensure-niuma-lucide] linked', lucideRoot, '->', dest)
