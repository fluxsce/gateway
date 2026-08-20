/**
 * 移除 index.html 静态首屏 Loading。
 */
export function removeBootSplash() {
  document.getElementById('app-boot-splash')?.remove()
  document.documentElement.classList.remove('app-booting')
}
