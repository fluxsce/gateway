<template>
  <n-modal
    :class="{ 'g-modal--fullscreen': isFullscreen }"
    :show="props.visible"
    :preset="props.preset"
    :style="modalStyle"
    :draggable="props.draggable"
    :unstable-show-mask="props.mask"
    :mask-closable="props.maskClosable"
    :closable="props.closable"
    :auto-focus="props.autoFocus"
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

    <!-- 头部插槽：完全自定义头部，支持标题 -->
    <template #header>
      <div class="g-modal__header">
        <div class="g-modal__header-main">
          <!-- 如果传入了自定义 header 插槽，则使用自定义内容；否则使用默认的 title -->
          <slot name="header">
            <span v-if="props.title" class="g-modal__title">{{ props.title }}</span>
          </slot>
        </div>
      </div>
    </template>

    <!-- 右上角区域：自定义 close 插槽，让全屏按钮和关闭按钮挨在一起显示 -->
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

    <!-- 主体内容：固定在头部和底部之间，仅内容区域滚动 -->
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
</template>

<script setup lang="ts">
import { CloseOutline, ContractOutline, ExpandOutline } from '@vicons/ionicons5'
import { NButton, NIcon, NModal, NSpace } from 'naive-ui'
import { computed, ref } from 'vue'
import type { GModalEmits, GModalProps } from './types'

defineOptions({
  name: 'GModal'
})

const props = withDefaults(defineProps<GModalProps>(), {
  visible: false,
  title: '',
  // 默认宽度占 80%
  width: '60%',
  preset: 'dialog',
  // 默认不显示遮罩层，允许背景页面可见/可移动
  mask: false,
  // 默认不阻止背景滚动
  blockScroll: false,
  // 默认支持拖拽
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
  segmented: false,
  bordered: false,
  showFullscreenToggle: true
})

const emit = defineEmits<GModalEmits>()

const isFullscreen = ref(false)

const modalStyle = computed(() => {
  // 全屏时忽略外部传入的 width，强制占满视口宽度
  const baseStyle = isFullscreen.value
    ? {
        width: '100vw',
        maxWidth: '100vw',
        // 高度由 JS 直接控制，避免仅靠类名时计算不准确
        height: '100vh',
        maxHeight: '100vh',
      }
    : {
        width: typeof props.width === 'number' ? `${props.width}px` : props.width
      }

  if (!props.style) return baseStyle

  // 合并外部传入的 style
  if (Array.isArray(props.style)) {
    return [baseStyle, ...props.style]
  }

  return [baseStyle, props.style]
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

/**
 * 动画进入完成后的处理
 */
const handleAfterEnter = () => {
  emit('after-enter')
}

/**
 * 动画离开完成后的处理
 * 外部可以通过监听此事件来处理内容销毁等逻辑
 */
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
</script>

<style scoped lang="scss">
.g-modal__header {
  display: flex;
  align-items: center;
  height: var(--g-modal-header-height);
  /* 去掉左侧内边距，让标题更靠近左侧图标，仅保留右侧间距 */
  padding: 0 var(--g-space-sm) 0 0;
  border-bottom: 1px solid var(--g-border-primary);
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
  // padding: var(--g-space-sm);
  /* 默认情况下，弹窗高度限制为 80vh，头部和底部固定，只有内容区域滚动 */
  height: calc(
    80vh - var(--g-modal-header-height) - var(--g-modal-footer-height) - 2 * var(--g-space-sm)
  );
  overflow-y: auto;
}

.g-modal__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  height: var(--g-modal-footer-height);
  padding: 0 var(--g-space-sm);
  border-top: 1px solid var(--g-border-primary);
}

/* 全屏样式：通过类控制，而不是在 JS 中计算样式
   - 宽度占满视口（不留左右空白）（具体数值在 baseStyle 中处理）
   - 顶部从全局 Header 下方开始，对齐主内容区域
*/
.g-modal--fullscreen {
  /* 让弹窗从顶部开始而不是垂直居中，并在纵向上拉伸 */
  align-self: stretch;
  transform: none !important;
}

/* 内部卡片 / 对话框宽度也拉满 */
.g-modal--fullscreen :deep(.n-card),
.g-modal--fullscreen :deep(.n-dialog) {
  width: 100% !important;
  max-width: 100% !important;
  border-radius: 0;
}

.g-modal--fullscreen .g-modal__body {
  /* 全屏时，减去全局 Header、高度和底部高度，只让内容区域滚动 */
  max-height: calc(
    100vh - var(--g-header-height) - var(--g-modal-header-height) - var(--g-modal-footer-height) - 2 * var(
        --g-space-sm
      )
  );
}
</style>


