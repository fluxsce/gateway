<template>
  <n-modal
    :data-g-modal="modalInstanceId"
    :class="modalRootClass"
    :show="props.visible"
    :preset="props.preset"
    :style="modalStyle"
    :draggable="props.draggable"
    :show-mask="props.mask"
    :mask-closable="props.maskClosable"
    :closable="props.closable"
    :auto-focus="props.autoFocus"
    :trap-focus="props.trapFocus"
    :block-scroll="props.blockScroll"
    :to="props.to === false ? undefined : props.to"
    :segmented="props.segmented"
    :bordered="props.bordered"
    @update:show="handleUpdateVisible"
    @after-enter="handleAfterEnter"
    @after-leave="handleAfterLeave"
  >
    <!-- 图标插槽：用于替换 dialog 的默认图标 -->
    <template #icon>
      <n-icon v-if="props.headerIcon">
        <component :is="props.headerIcon" />
      </n-icon>
      <slot v-else name="headerIcon"></slot>
    </template>

    <template #header>
      <div class="g-modal__header">
        <div class="g-modal__header-main">
          <slot name="header">
            <span v-if="props.title" class="g-modal__title">{{ props.title }}</span>
          </slot>
        </div>
      </div>
    </template>

    <template #close>
      <div class="g-modal__close">
        <n-icon
          v-if="props.showFullscreenToggle"
          size="18"
          class="g-modal__icon g-modal__icon--fullscreen"
          @click.stop="toggleFullscreen"
        >
          <component :is="isFullscreen ? ContractOutline : ExpandOutline" />
        </n-icon>
        <n-icon
          size="18"
          class="g-modal__icon g-modal__icon--close"
          @click.stop="handleClose"
        >
          <CloseOutline />
        </n-icon>
      </div>
    </template>

    <div class="g-modal__body">
      <slot />
    </div>

    <!-- 底部操作区，支持关闭 / 确认按钮，或完全自定义
         注意：Naive UI 对不同 preset 使用的插槽不同：
         - preset="card" 使用 footer 插槽
         - preset="dialog" 使用 action 插槽
         这里避免在带有插槽名的 <template> 上使用 v-if，防止编译器报错，
         而是在内部元素上控制显示。
    -->
    <!-- dialog 预设：使用 action 插槽 -->
    <template #action>
      <slot name="footer">
        <div
          v-if="props.showFooter && props.preset === 'dialog'"
          class="g-modal__footer"
        >
          <n-space justify="end" :size="8">
            <n-button
              v-for="btn in props.footerToolbar"
              :key="btn.key"
              size="small"
              v-bind="btn.buttonProps"
              @click="handleToolbarClick(btn.key)"
            >
              {{ btn.label }}
            </n-button>
            <n-button v-if="props.showCancel" size="small" @click="handleCancel">
              {{ props.cancelText }}
            </n-button>
            <n-button
              v-if="props.showConfirm"
              type="primary"
              size="small"
              :loading="props.confirmLoading"
              @click="handleConfirm"
            >
              {{ props.confirmText }}
            </n-button>
          </n-space>
        </div>
      </slot>
    </template>

    <!-- card 等其他预设：使用 footer 插槽 -->
    <template #footer>
      <slot name="footer">
        <div
          v-if="props.showFooter && props.preset !== 'dialog'"
          class="g-modal__footer"
        >
          <n-space justify="end" :size="8">
            <n-button
              v-for="btn in props.footerToolbar"
              :key="btn.key"
              size="small"
              v-bind="btn.buttonProps"
              @click="handleToolbarClick(btn.key)"
            >
              {{ btn.label }}
            </n-button>
            <n-button v-if="props.showCancel" size="small" @click="handleCancel">
              {{ props.cancelText }}
            </n-button>
            <n-button
              v-if="props.showConfirm"
              type="primary"
              size="small"
              :loading="props.confirmLoading"
              @click="handleConfirm"
            >
              {{ props.confirmText }}
            </n-button>
          </n-space>
        </div>
      </slot>
    </template>
  </n-modal>

  <!--
    Teleport（Vue 内置）：将子节点渲染到 DOM 中其它位置（此处为 body），逻辑仍属本组件。
    缩放手柄挂到 body 的原因：与 Naive NModal 一样脱离局部 overflow/transform，避免被裁剪；
    手柄使用 fixed 与弹层叠放，拖拽时命中区域不被中间层挡住。
  -->
  <Teleport to="body">
    <template v-if="props.visible && props.resizable && !isFullscreen">
      <div
        class="g-modal-resize-handle"
        role="presentation"
        tabindex="-1"
        :style="resizeHandleStyles.e"
        @mousedown="(ev) => startResize(ev, 'e')"
      />
      <div
        class="g-modal-resize-handle"
        role="presentation"
        tabindex="-1"
        :style="resizeHandleStyles.s"
        @mousedown="(ev) => startResize(ev, 's')"
      />
      <div
        class="g-modal-resize-handle"
        role="presentation"
        tabindex="-1"
        :style="resizeHandleStyles.se"
        @mousedown="(ev) => startResize(ev, 'se')"
      />
    </template>
  </Teleport>
</template>

<script setup lang="ts">
import { CloseOutline, ContractOutline, ExpandOutline } from '@vicons/ionicons5'
import { NButton, NIcon, NModal, NSpace } from 'naive-ui'
import { computed, ref, toRef, useSlots, watch } from 'vue'
import type { GModalEmits, GModalProps } from './types'
import { useGModalResize } from './useGModalResize'

defineOptions({
  name: 'GModal',
})

const props = withDefaults(defineProps<GModalProps>(), {
  visible: false,
  title: '',
  width: '60%',
  height: undefined,
  resizable: false,
  resizeMinWidth: 320,
  resizeMinHeight: 200,
  preset: 'dialog',
  mask: false,
  blockScroll: false,
  draggable: true,
  maskClosable: true,
  closable: true,
  showFooter: true,
  showCancel: true,
  showConfirm: true,
  cancelText: '取消',
  confirmText: '确定',
  confirmLoading: false,
  autoFocus: true,
  trapFocus: false,
  segmented: false,
  bordered: false,
  showFullscreenToggle: true,
})

const emit = defineEmits<GModalEmits>()
const slots = useSlots()

/**
 * 每实例唯一 id，写入根节点 `data-g-modal`。
 * - useGModalResize 用 `document.querySelector('[data-g-modal="…"]')` 定位当前弹层；多实例并存时避免命中其它对话框。
 * - 样式中 `[data-g-modal]` 只表示「带标记的 GModal」，不依赖属性值。
 */
const modalInstanceId =
  typeof crypto !== 'undefined' && crypto.randomUUID
    ? crypto.randomUUID()
    : `g-modal-${Math.random().toString(36).slice(2)}`

const isFullscreen = ref(false)

/** 边框拖拽后的像素尺寸；百分比等非像素 width/height 在未拖拽前为 null */
const panelPixelWidth = ref<number | null>(null)
const panelPixelHeight = ref<number | null>(null)

function toCssSize(v: number | string | undefined): string | undefined {
  if (v === undefined || v === '') {
    return undefined
  }
  return typeof v === 'number' ? `${v}px` : String(v)
}

/** 仅从数字或 px 字符串解析像素，用于与拖拽统一为像素 */
function parseCssSizePx(v: number | string | undefined): number | null {
  if (v === undefined || v === null) {
    return null
  }
  if (typeof v === 'number' && !Number.isNaN(v)) {
    return v
  }
  const m = /^(\d+(?:\.\d+)?)px$/i.exec(String(v).trim())
  return m ? parseFloat(m[1]) : null
}

function viewportMaxW(): number {
  return typeof window !== 'undefined' ? Math.max(200, window.innerWidth - 24) : 1600
}

function viewportMaxH(): number {
  return typeof window !== 'undefined' ? Math.max(200, window.innerHeight - 24) : 1200
}

function parseMaxDim(v: number | string | undefined, viewportCap: number): number {
  if (v === undefined) {
    return viewportCap
  }
  if (typeof v === 'number' && !Number.isNaN(v)) {
    return Math.min(viewportCap, v)
  }
  const px = /^(\d+(?:\.\d+)?)px$/i.exec(String(v).trim())
  if (px) {
    return Math.min(viewportCap, parseFloat(px[1]))
  }
  return viewportCap
}

watch(
  () => props.visible,
  (v) => {
    if (v) {
      panelPixelWidth.value = parseCssSizePx(props.width)
      panelPixelHeight.value = parseCssSizePx(props.height)
    }
  }
)

const modalRootClass = computed(() => ({
  'g-modal--fullscreen': isFullscreen.value,
  /* 无内置底部且未提供 footer 插槽时隐藏 dialog 的 action 占位 */
  'g-modal--no-footer': !props.showFooter && !slots.footer,
}))

const modalStyle = computed(() => {
  if (isFullscreen.value) {
    /* 必须用视口单位：父链为 NScrollbar → .n-scrollbar-content，height:100% 无确定参照会导致 flex+height:0 把 body 算成 0 */
    const base = {
      width: '100vw',
      maxWidth: '100vw',
      height: '100vh',
      maxHeight: '100vh',
      boxSizing: 'border-box',
      overflow: 'hidden',
    } as Record<string, string>
    if (!props.style) {
      return base
    }
    return Array.isArray(props.style) ? [base, ...props.style] : [base, props.style]
  }

  const base: Record<string, string> = {
    display: 'flex',
    flexDirection: 'column',
    boxSizing: 'border-box',
    /* max-height 仅在 overflow 非 visible 时才会裁剪；否则子项仍会画出白底外 */
    overflow: 'hidden',
  }

  const pw = panelPixelWidth.value
  const ph = panelPixelHeight.value
  if (pw != null) {
    base.width = `${pw}px`
    base.maxWidth = `${parseMaxDim(props.resizeMaxWidth, viewportMaxW())}px`
  } else {
    const w = toCssSize(props.width)
    if (w) {
      base.width = w
    }
    base.maxWidth = `${parseMaxDim(props.resizeMaxWidth, viewportMaxW())}px`
  }

  if (ph != null) {
    base.height = `${ph}px`
    base.maxHeight = `${parseMaxDim(props.resizeMaxHeight, viewportMaxH())}px`
  } else {
    const h = toCssSize(props.height)
    if (h) {
      base.height = h
      base.maxHeight = h
    } else {
      /* 仅 max-height 时列 flex 无法得到确定剩余高度，body 的 overflow:auto 不生效；固定高度后由内部 flex+height:0 吃掉剩余空间并滚动 */
      base.height = '80vh'
      base.maxHeight = '80vh'
    }
  }

  const shellHasExplicitHeight =
    ph != null ||
    (props.height !== undefined &&
      props.height !== null &&
      String(props.height).trim() !== '')
  const noFooterBar = !props.showFooter && !slots.footer
  base['--g-modal-body-max-height'] = shellHasExplicitHeight
    ? '100%'
    : noFooterBar
      ? 'calc(80vh - 4.5rem)'
      : 'calc(80vh - 6rem)'

  if (!props.style) {
    return base
  }
  return Array.isArray(props.style) ? [base, ...props.style] : [base, props.style]
})

const { handleStyles: resizeHandleStyles, startResize, syncHandlePositions } = useGModalResize({
  instanceId: modalInstanceId,
  visible: toRef(props, 'visible'),
  resizable: toRef(props, 'resizable'),
  isFullscreen,
  getMinWidth: () => props.resizeMinWidth ?? 320,
  getMinHeight: () => props.resizeMinHeight ?? 200,
  getMaxWidth: () => parseMaxDim(props.resizeMaxWidth, viewportMaxW()),
  getMaxHeight: () => parseMaxDim(props.resizeMaxHeight, viewportMaxH()),
  panelPixelWidth,
  panelPixelHeight,
  onResizeEnd: () => {
    const w = panelPixelWidth.value
    const h = panelPixelHeight.value
    if (w != null && h != null) {
      emit('resize', { width: w, height: h })
    }
  },
})

const toggleFullscreen = () => {
  isFullscreen.value = !isFullscreen.value
}

const handleClose = () => {
  emit('update:visible', false)
  isFullscreen.value = false
  emit('close')
}

const handleUpdateVisible = (value: boolean) => {
  emit('update:visible', value)
  if (!value) {
    isFullscreen.value = false
    emit('close')
  }
}

const handleAfterEnter = () => {
  if (props.resizable && !isFullscreen.value) {
    syncHandlePositions()
  }
  emit('after-enter')
}

const handleAfterLeave = () => {
  emit('after-leave')
}

const handleCancel = () => {
  emit('cancel')
  emit('update:visible', false)
  isFullscreen.value = false
}

const handleConfirm = () => {
  emit('confirm')
}

const handleToolbarClick = (key: string) => {
  emit('toolbar-click', key)
}
</script>

<style scoped lang="scss">
.g-modal__header {
  display: flex;
  align-items: center;
  height: var(--g-modal-header-height);
  padding: 0 var(--g-space-sm) 0 0;
  border-bottom: 1px solid var(--g-border-primary);
  flex-shrink: 0;
}

.g-modal__header-main {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.g-modal__header-icon {
  color: var(--n-color-primary);
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.g-modal__title {
  font-size: var(--g-font-size-lg);
  font-weight: 500;
  color: var(--g-text-primary);
}

.g-modal__close {
  display: inline-flex;
  align-items: center;
  gap: var(--g-space-xs);
}

.g-modal__icon {
  cursor: pointer;
  color: var(--g-text-secondary);
  transition:
    color var(--g-transition-base) var(--g-transition-ease),
    background-color var(--g-transition-base) var(--g-transition-ease);
}

.g-modal__icon:hover {
  color: var(--g-text-primary);
  background-color: var(--g-hover-overlay);
}

.g-modal__body {
  min-width: 0;
  overflow-x: hidden;
}

.g-modal__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  height: var(--g-modal-footer-height);
  padding: 0 var(--g-space-sm);
  border-top: 1px solid var(--g-border-primary);
  flex-shrink: 0;
  box-sizing: border-box;
}

/* 透明命中区，不画灰条；光标由 useGModalResize 内联 cursor（ew/ns/nwse-resize）在边缘悬停时体现 */
.g-modal-resize-handle {
  touch-action: none;
  outline: none;
  pointer-events: auto;
  background: transparent;
}

/* Naive 将 NModal 的 class / data 合并到 NDialog 根节点，与 n-dialog 同元素 */
:deep(.n-dialog.n-modal),
:deep(.n-card.n-modal) {
  display: flex !important;
  flex-direction: column !important;
  min-width: 0 !important;
  min-height: 0 !important;
  overflow: hidden !important;
}

:deep(.n-dialog.n-modal:not(.g-modal--fullscreen) .n-dialog__title) {
  flex-shrink: 0;
}

:deep(.n-dialog.g-modal--no-footer .n-dialog__action) {
  display: none !important;
  height: 0 !important;
  min-height: 0 !important;
  padding: 0 !important;
  margin: 0 !important;
  border: none !important;
  overflow: hidden !important;
}

/* 全屏：覆盖主题 446px，纵向撑满 */
:deep(.n-dialog.g-modal--fullscreen),
:deep(.n-card.g-modal--fullscreen) {
  width: 100vw !important;
  max-width: 100vw !important;
  min-width: 0 !important;
  height: 100vh !important;
  max-height: 100vh !important;
  margin: 0 !important;
  border-radius: 0;
  align-self: stretch !important;
  transform: none !important;
  box-sizing: border-box;
  display: flex !important;
  flex-direction: column !important;
  overflow: hidden !important;
}

:deep(.n-dialog.g-modal--fullscreen .n-dialog__title),
:deep(.n-card.g-modal--fullscreen .n-card-header) {
  flex-shrink: 0;
}

:deep(.n-dialog.g-modal--fullscreen .n-dialog__content) {
  /* flex-basis 0% + min-height:0 即可在列 flex 中分到剩余高度；勿对子项设 height:0，否则部分环境下会固定为 0 */
  flex: 1 1 0% !important;
  min-height: 0 !important;
  width: 100%;
  max-width: none;
  box-sizing: border-box;
  display: flex !important;
  flex-direction: column !important;
  overflow: hidden !important;
}

:deep(.n-dialog.g-modal--fullscreen .g-modal__body),
:deep(.n-card.g-modal--fullscreen .g-modal__body) {
  flex: 1 1 0% !important;
  min-height: 0 !important;
  overflow-y: auto !important;
  width: 100%;
  box-sizing: border-box;
}

:deep(.n-card.g-modal--fullscreen .n-card__content) {
  flex: 1 1 0% !important;
  min-height: 0 !important;
  display: flex !important;
  flex-direction: column !important;
  overflow: hidden !important;
}
</style>

<style lang="scss">
/* Teleport 后 Naive 根节点与 scoped :deep 选择器偶发不匹配，用 data-g-modal + 非 scoped 保证 content/body 限高与滚动 */
.n-dialog[data-g-modal]:not(.g-modal--fullscreen) .n-dialog__content,
.n-card[data-g-modal]:not(.g-modal--fullscreen) .n-card__content {
  flex: 1 1 0% !important;
  min-height: 0 !important;
  height: 0 !important;
  min-width: 0 !important;
  display: flex !important;
  flex-direction: column !important;
  overflow: hidden !important;
  text-align: start;
}

.n-dialog[data-g-modal]:not(.g-modal--fullscreen) .g-modal__body,
.n-card[data-g-modal]:not(.g-modal--fullscreen) .g-modal__body {
  flex: 1 1 0% !important;
  min-height: 0 !important;
  max-height: var(--g-modal-body-max-height, calc(80vh - 6rem)) !important;
  overflow-y: auto !important;
  overflow-x: hidden !important;
  box-sizing: border-box !important;
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
}

/* 全屏：Teleport 后必须用本段保证 n-dialog 为列 flex + 100vh（scoped :deep 可能未命中），否则子项 flex 分高为 0 */
.n-dialog[data-g-modal].g-modal--fullscreen,
.n-card[data-g-modal].g-modal--fullscreen {
  display: flex !important;
  flex-direction: column !important;
  width: 100vw !important;
  max-width: 100vw !important;
  min-width: 0 !important;
  height: 100vh !important;
  max-height: 100vh !important;
  min-height: 0 !important;
  margin: 0 !important;
  box-sizing: border-box !important;
  overflow: hidden !important;
}

.n-dialog[data-g-modal].g-modal--fullscreen .n-dialog__title,
.n-card[data-g-modal].g-modal--fullscreen .n-card-header {
  flex-shrink: 0 !important;
}

.n-dialog[data-g-modal].g-modal--fullscreen .n-dialog__content,
.n-card[data-g-modal].g-modal--fullscreen .n-card__content {
  flex: 1 1 0% !important;
  min-height: 0 !important;
  min-width: 0 !important;
  width: 100% !important;
  max-width: none;
  box-sizing: border-box !important;
  display: flex !important;
  flex-direction: column !important;
  overflow: hidden !important;
}

.n-dialog[data-g-modal].g-modal--fullscreen .g-modal__body,
.n-card[data-g-modal].g-modal--fullscreen .g-modal__body {
  flex: 1 1 0% !important;
  min-height: 0 !important;
  overflow-y: auto !important;
  overflow-x: hidden !important;
  width: 100% !important;
  box-sizing: border-box !important;
  overscroll-behavior: contain;
}

/* 全屏：抵消 Naive modal 的 min-height:100% 与外层 NScrollbar，避免外层滚动与 body 双滚动条；body 仅由 flex 占满并滚动 */
.n-modal-body-wrapper:has(.n-dialog[data-g-modal].g-modal--fullscreen),
.n-modal-body-wrapper:has(.n-card[data-g-modal].g-modal--fullscreen) {
  overflow: hidden !important;
}

.n-modal-body-wrapper:has(.n-dialog[data-g-modal].g-modal--fullscreen) .n-modal-scroll-content,
.n-modal-body-wrapper:has(.n-card[data-g-modal].g-modal--fullscreen) .n-modal-scroll-content {
  min-height: 0 !important;
  height: 100% !important;
  max-height: 100% !important;
  width: 100% !important;
  overflow: hidden !important;
  display: flex !important;
  align-items: stretch !important;
  justify-content: flex-start !important;
  box-sizing: border-box !important;
}

.n-modal-body-wrapper:has(.n-dialog[data-g-modal].g-modal--fullscreen) .n-scrollbar,
.n-modal-body-wrapper:has(.n-card[data-g-modal].g-modal--fullscreen) .n-scrollbar {
  height: 100% !important;
  max-height: 100% !important;
  min-height: 0 !important;
}

.n-modal-body-wrapper:has(.n-dialog[data-g-modal].g-modal--fullscreen) .n-scrollbar-container,
.n-modal-body-wrapper:has(.n-card[data-g-modal].g-modal--fullscreen) .n-scrollbar-container {
  overflow: hidden !important;
  max-height: 100% !important;
}

/* 让 NScrollbar 内层与视口同高，避免 100vh 的 dialog 在「未设高的」scrollbar-content 里参与 flex 时分高为 0 */
.n-modal-body-wrapper:has(.n-dialog[data-g-modal].g-modal--fullscreen) .n-scrollbar-content,
.n-modal-body-wrapper:has(.n-card[data-g-modal].g-modal--fullscreen) .n-scrollbar-content {
  min-height: 0 !important;
  height: 100% !important;
  max-height: 100% !important;
  box-sizing: border-box !important;
  display: flex !important;
  flex-direction: column !important;
}
</style>
