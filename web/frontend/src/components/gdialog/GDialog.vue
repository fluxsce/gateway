<template>
  <n-modal
    v-model:show="localShow"
    :mask-closable="props.maskClosable"
    :close-on-esc="props.closeOnEsc"
    :draggable="props.draggable"
    :transform-origin="'center'"
    :style="modalStyle"
    @update:show="handleUpdateShow"
    @after-enter="emit('after-enter')"
    @after-leave="emit('after-leave')"
  >
    <n-card
      :style="cardStyle"
      :bordered="false"
      size="huge"
      role="dialog"
      aria-modal="true"
      :class="['g-dialog', props.cardClass, { 'g-dialog--gradient-header': props.headerStyle === 'gradient' }]"
    >
      <!-- 头部区域 -->
      <template #header>
        <slot name="header" :title="props.title" :subtitle="props.subtitle" :icon="props.icon">
          <div class="g-dialog__header" :style="headerStyle">
            <div v-if="props.icon || $slots.icon" class="g-dialog__header-icon">
              <slot name="icon">
                <n-icon v-if="props.icon" :size="props.iconSize || 24">
                  <component :is="props.icon" />
                </n-icon>
              </slot>
            </div>
            <div class="g-dialog__header-content">
              <h3 v-if="props.title" class="g-dialog__title">{{ props.title }}</h3>
              <p v-if="props.subtitle && props.subtitlePosition !== 'footer'" class="g-dialog__subtitle">{{ props.subtitle }}</p>
            </div>
            <div class="g-dialog__header-extra">
              <slot name="header-extra">
                <n-button
                  v-if="props.closable"
                  quaternary
                  circle
                  class="g-dialog__close-btn"
                  @click="handleClose"
                >
                  <template #icon>
                    <n-icon :size="18">
                      <CloseOutline />
                    </n-icon>
                  </template>
                </n-button>
              </slot>
            </div>
          </div>
        </slot>
      </template>

      <!-- 内容区域 -->
      <div class="g-dialog__body">
        <n-scrollbar
          v-if="props.showScrollbar"
          :style="{ maxHeight: props.contentMaxHeight || '70vh' }"
          trigger="hover"
          :x-scrollable="false"
        >
          <div class="g-dialog__content-wrapper">
            <div class="g-dialog__content">
              <slot />
            </div>
          </div>
        </n-scrollbar>
        <div v-else class="g-dialog__content-wrapper" :style="{ maxHeight: props.contentMaxHeight }">
          <div class="g-dialog__content">
            <slot />
          </div>
        </div>
      </div>

      <!-- 底部操作区 -->
      <template #footer>
        <slot name="footer" :confirmLoading="props.confirmLoading" :onConfirm="handleConfirm" :onCancel="handleCancel">
          <div v-if="props.showFooter" class="g-dialog__footer">
            <div class="g-dialog__footer-inner">
              <div v-if="props.subtitle && props.subtitlePosition === 'footer'" class="g-dialog__footer-subtitle">
                {{ props.subtitle }}
              </div>
              <n-space :justify="props.footerButtonAlign || 'end'" :size="8" class="g-dialog__footer-buttons">
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
          </div>
        </slot>
      </template>
    </n-card>
  </n-modal>
</template>

<script setup lang="ts">
import { CloseOutline } from '@vicons/ionicons5'
import { NButton, NCard, NIcon, NModal, NScrollbar, NSpace } from 'naive-ui'
import { computed, ref, watch } from 'vue'
import type { GDialogEmits, GDialogProps } from './types'

defineOptions({
  name: 'GDialog'
})

const props = withDefaults(defineProps<GDialogProps>(), {
  show: false,
  width: 1000,
  maskClosable: false,
  closeOnEsc: false,
  closable: true,
  showScrollbar: true,
  showFooter: true,
  showCancel: true,
  showConfirm: true,
  cancelText: '取消',
  confirmText: '确定',
  confirmLoading: false,
  autoCloseOnConfirm: false,
  draggable: false,
  headerStyle: 'default',
  iconSize: 24,
  subtitlePosition: 'footer'
})

const emit = defineEmits<GDialogEmits>()

const localShow = ref(props.show)

// 监听外部 show 变化
watch(
  () => props.show,
  (newVal) => {
    localShow.value = newVal
  }
)

// 监听内部 localShow 变化，同步到外部
watch(localShow, (newVal) => {
  if (newVal !== props.show) {
    emit('update:show', newVal)
  }
})

const modalStyle = computed(() => {
  const baseStyle: Record<string, any> = {}
  
  if (props.style) {
    if (typeof props.style === 'object' && !Array.isArray(props.style)) {
      Object.assign(baseStyle, props.style)
    }
  }

  return baseStyle
})

const cardStyle = computed(() => {
  const width = typeof props.width === 'number' ? `${props.width}px` : props.width
  return {
    width,
    ...(props.style && typeof props.style === 'object' && !Array.isArray(props.style) ? props.style : {})
  }
})

const headerStyle = computed(() => {
  if (props.headerCustomStyle) {
    return props.headerCustomStyle
  }
  return {}
})

const handleUpdateShow = (value: boolean) => {
  localShow.value = value
  if (!value) {
    emit('close')
  }
}

const handleClose = () => {
  localShow.value = false
  emit('close')
}

const handleCancel = () => {
  emit('cancel')
  localShow.value = false
}

const handleConfirm = () => {
  emit('confirm')
  if (props.autoCloseOnConfirm) {
    localShow.value = false
  }
}
</script>

<style scoped lang="scss">
.g-dialog {
  border-radius: var(--g-radius-2xl, 16px);
  box-shadow: var(--g-dialog-shadow);
  background: var(--g-dialog-bg);
  /* 性能优化：启用硬件加速 */
  transform: translateZ(0);
  overflow: hidden;
}

/* 默认头部样式 */
.g-dialog__header {
  display: flex;
  align-items: center;
  gap: var(--g-space-md, 16px);
  height: var(--g-dialog-header-height);
  padding: 0 var(--g-space-lg, 24px);
  background: var(--g-dialog-bg);
  border-bottom: 1px solid var(--g-border-primary);
  box-sizing: border-box;
}

.g-dialog--gradient-header .g-dialog__header {
  background: var(--g-dialog-header-bg-gradient);
  border-bottom: none;
  margin: 0;
  border-radius: var(--g-radius-2xl, 16px) var(--g-radius-2xl, 16px) 0 0;
}

.g-dialog__header-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.g-dialog--gradient-header .g-dialog__header-icon {
  color: white;
}

.g-dialog__header-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: var(--g-space-xs, 4px);
}

.g-dialog__title {
  font-size: var(--g-font-size-lg, 16px);
  font-weight: 600;
  color: var(--g-text-primary);
  margin: 0;
  line-height: 1.5;
}

.g-dialog--gradient-header .g-dialog__title {
  color: white;
}

.g-dialog__subtitle {
  font-size: var(--g-font-size-sm, 13px);
  color: var(--g-text-secondary);
  margin: 0;
  line-height: 1.5;
}

.g-dialog--gradient-header .g-dialog__subtitle {
  color: rgba(255, 255, 255, 0.85);
}

.g-dialog__header-extra {
  display: flex;
  align-items: center;
  gap: var(--g-space-sm, 8px);
  flex-shrink: 0;
}

.g-dialog__close-btn {
  transition: all var(--g-transition-base, 200ms) var(--g-transition-ease);
  color: var(--g-text-secondary);
}

.g-dialog--gradient-header .g-dialog__close-btn {
  color: white;
}

.g-dialog__close-btn:hover {
  background-color: var(--g-hover-overlay);
}

.g-dialog--gradient-header .g-dialog__close-btn:hover {
  background-color: rgba(255, 255, 255, 0.15);
  color: white;
}

/* 内容区域 */
.g-dialog__body {
  background: var(--g-dialog-bg);
  overflow: hidden;
}

.g-dialog__content-wrapper {
  box-sizing: border-box;
}

.g-dialog__content {
  padding: var(--g-space-lg, 24px);
  min-height: 60px;
  box-sizing: border-box;
  color: var(--g-text-primary);
  line-height: 1.6;
  word-wrap: break-word;
}

/* 底部操作区 */
.g-dialog__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--g-space-md, 16px);
  min-height: var(--g-dialog-footer-height);
  height: var(--g-dialog-footer-height);
  padding: 0;
  border-top: 1px solid var(--g-border-primary);
  border-radius: 0 0 var(--g-radius-2xl, 16px) var(--g-radius-2xl, 16px);
  box-sizing: border-box;
}

.g-dialog__footer-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--g-space-md, 16px);
  width: 100%;
  height: 100%;
  padding: 0 var(--g-space-lg, 24px);
}

.g-dialog__footer-subtitle {
  flex: 1;
  font-size: var(--g-font-size-sm, 13px);
  color: var(--g-text-secondary);
  line-height: 1.5;
  margin: 0;
}

.g-dialog__footer-buttons {
  flex-shrink: 0;
  margin-left: auto;
}

/* 按钮样式：统一与 GModal 保持一致 */
.g-dialog__footer :deep(.n-button) {
  border-radius: var(--g-radius-md, 6px);
}

/* 确保按钮 hover 效果与 GModal 一致 */
.g-dialog__footer :deep(.n-button:not(.n-button--disabled)):hover {
  transition: all var(--g-transition-base, 200ms) var(--g-transition-ease);
}

/* 深度选择器：自定义卡片头部样式 */
.g-dialog :deep(.n-card-header) {
  padding: 0 !important;
  border-bottom: none;
  background: transparent;
}

.g-dialog :deep(.n-card-body) {
  padding: 0;
  background: transparent;
}

/* 覆盖 Naive UI 的 footer 样式 - 支持所有可能的类名变体 */
.g-dialog :deep(.n-card > .n-card__footer),
.g-dialog :deep(.n-card > .n-card-footer),
.g-dialog :deep(.n-card__footer),
.g-dialog :deep(.n-card-footer) {
  padding: 0 !important;
  border-top: none !important;
  background: transparent !important;
  min-height: var(--g-dialog-footer-height) !important;
  height: var(--g-dialog-footer-height) !important;
  max-height: var(--g-dialog-footer-height) !important;
  box-sizing: border-box !important;
  overflow: hidden !important;
  font-size: inherit !important;
  line-height: inherit !important;
}

.g-dialog--gradient-header :deep(.n-card-header) {
  background: transparent;
  padding: 0 !important;
}

/* 优化卡片样式 */
.g-dialog :deep(.n-card) {
  background: var(--g-dialog-bg);
}
</style>

