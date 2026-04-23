<template>
  <div class="route-view-loading-mask-test">
    <div class="header">
      <div class="title">RouteViewLoadingMask 效果预览</div>
      <div class="actions">
        <NButton size="small" @click="toggle">
          {{ showMask ? '关闭遮罩' : '显示遮罩' }}
        </NButton>
        <NButton size="small" secondary @click="simulateBusy">
          模拟 1.2s 加载
        </NButton>
      </div>
    </div>

    <div class="hint">
      遮罩使用 <code>position: absolute</code> 覆盖容器区域；这里用一个相对定位容器模拟
      <code>MainLayoutContent</code> 的内容区。
    </div>

    <div class="preview">
      <div class="preview-shell">
        <RouteViewLoadingMask v-if="showMask" />

        <div class="preview-content">
          <div class="block" v-for="i in 18" :key="i">
            <div class="block-title">内容块 {{ i }}</div>
            <div class="block-desc">
              用于观察遮罩透明度、模糊效果、滚动时是否裁剪，以及动画是否平滑。
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import RouteViewLoadingMask from '@/components/RouteViewLoadingMask.vue'
import { NButton } from 'naive-ui'
import { ref } from 'vue'

defineOptions({ name: 'RouteViewLoadingMaskTest' })

const showMask = ref(false)
let busyTimer: number | null = null

function toggle() {
  showMask.value = !showMask.value
}

function simulateBusy() {
  if (busyTimer != null) {
    window.clearTimeout(busyTimer)
    busyTimer = null
  }
  showMask.value = true
  busyTimer = window.setTimeout(() => {
    showMask.value = false
    busyTimer = null
  }, 1200)
}
</script>

<style scoped lang="scss">
.route-view-loading-mask-test {
  padding: var(--g-padding-xxl, 24px);
  max-width: 1100px;
  margin: 0 auto;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.title {
  font-size: 18px;
  font-weight: 700;
  color: var(--g-text-primary);
}

.actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.hint {
  margin-top: 10px;
  color: var(--g-text-secondary);
  font-size: 13px;
  line-height: 1.7;
}

.hint code {
  padding: 0 6px;
  border-radius: 6px;
  background: color-mix(in srgb, var(--g-bg-secondary) 75%, transparent);
  border: 1px solid var(--g-border-primary);
}

.preview {
  margin-top: 16px;
}

.preview-shell {
  position: relative;
  height: 520px;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid var(--g-border-primary);
  background: var(--g-bg-secondary);
}

.preview-content {
  height: 100%;
  overflow: auto;
  padding: 14px;
  box-sizing: border-box;
}

.block {
  padding: 12px 12px 10px;
  border-radius: 10px;
  border: 1px solid var(--g-border-primary);
  background: var(--g-bg-primary);
}

.block + .block {
  margin-top: 10px;
}

.block-title {
  font-weight: 600;
  color: var(--g-text-primary);
}

.block-desc {
  margin-top: 6px;
  font-size: 12px;
  color: var(--g-text-tertiary);
  line-height: 1.6;
}
</style>

