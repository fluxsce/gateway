/**
 * boot 为极小的首屏启动器：只负责尽快展示静态 HTML，再异步加载真正的应用入口 `main.ts`。
 *
 * 目的：`index.html` 首条 `<script type="module">` 不再直接指向 `main.ts`，避免浏览器在首屏就解析/下载
 * 巨型依赖图（Vue/Naive 等），从观感上减少“长时间空白等待”。
 */

function removeBootSplash() {
  document.getElementById('app-boot-splash')?.remove()
}

function showFatal(message: string) {
  removeBootSplash()
  const root = document.getElementById('app')
  if (!root) return
  root.innerHTML = `
      <div style="padding: 20px; text-align: center; color: #666;">
        <h2>应用加载失败</h2>
        <p>请刷新页面或联系管理员</p>
        <p style="font-size: 12px; margin-top: 10px;">错误信息: ${message}</p>
      </div>
    `
}

/**
 * scheduleAppLoad 在下一帧开始加载应用入口，给浏览器一次 paint 静态 HTML 的机会。
 */
function scheduleAppLoad() {
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      void loadApp()
    })
  })
}

async function loadApp() {
  try {
    const mod = await import('./main')
    await mod.startApp()
  } catch (e) {
    console.error('应用启动失败:', e)
    const msg = e instanceof Error ? e.message : String(e)
    showFatal(msg || '未知错误')
  }
}

scheduleAppLoad()
