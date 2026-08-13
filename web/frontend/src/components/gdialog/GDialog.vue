<template>
  <!--
    @deprecated 请直接使用 RsDialog（@/ui）。本组件仅过渡兼容，后续将删除。
  -->
  <RsDialog
    :open="props.show"
    :title="dialogTitle"
    :description="headerDescription"
    layout="window"
    :width="dialogWidth"
    :draggable="props.draggable"
    :resizable="false"
    :fullscreenable="props.showFullscreenToggle"
    :show-overlay="true"
    :show-close="props.closable"
    :close-on-overlay-click="props.maskClosable"
    :class="['g-dialog', props.cardClass, { 'g-dialog--gradient-header': props.headerStyle === 'gradient' }]"
    :style="props.style"
    @update:open="handleUpdateShow"
  >
    <template #body>
      <div class="g-dialog__body">
        <div v-if="props.icon || $slots.icon" class="g-dialog__header-icon">
          <slot name="icon">
            <component :is="props.icon" v-if="props.icon" />
          </slot>
        </div>
        <p v-if="props.subtitle && props.subtitlePosition !== 'footer'" class="g-dialog__subtitle">
          {{ props.subtitle }}
        </p>
        <RsScrollbar v-if="props.showScrollbar" :style="{ maxHeight: props.contentMaxHeight || '70vh' }">
          <div class="g-dialog__content">
            <slot />
          </div>
        </RsScrollbar>
        <div v-else class="g-dialog__content" :style="{ maxHeight: props.contentMaxHeight }">
          <slot />
        </div>
      </div>
    </template>
    <template #footer>
      <slot
        name="footer"
        :confirm-loading="props.confirmLoading"
        :on-confirm="handleConfirm"
        :on-cancel="handleCancel"
      >
        <div
          v-if="props.showFooter"
          class="g-dialog__footer"
          :style="{ alignItems: footerAlign }"
        >
          <div v-if="props.subtitle && props.subtitlePosition === 'footer'" class="g-dialog__footer-subtitle">
            {{ props.subtitle }}
          </div>
          <div class="g-dialog__footer-buttons" :style="{ justifyContent: footerAlign }">
            <RsButton v-if="props.showCancel" size="sm" variant="default" @click="handleCancel">
              {{ props.cancelText }}
            </RsButton>
            <RsButton
              v-if="props.showConfirm"
              variant="primary"
              size="sm"
              :loading="props.confirmLoading"
              @click="handleConfirm"
            >
              {{ props.confirmText }}
            </RsButton>
          </div>
        </div>
      </slot>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
/**
 * @deprecated 请直接使用 `@/ui` 的 `RsDialog`。本组件仅作过渡兼容，后续将删除。
 */
import { RsButton, RsDialog, RsScrollbar } from '@/ui'
import { computed, onMounted } from 'vue'
import type { GDialogEmits, GDialogProps } from './types'

defineOptions({
  name: 'GDialog',
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
  subtitlePosition: 'footer',
  footerButtonAlign: 'end',
  showFullscreenToggle: true,
})

const emit = defineEmits<GDialogEmits>()

onMounted(() => {
  if (import.meta.env.DEV) {
    console.warn('[GDialog] 已弃用：请改用 RsDialog（@/ui）。')
  }
})

const dialogTitle = computed(() => props.title?.trim() || ' ')

const headerDescription = computed(() =>
  props.subtitlePosition === 'header' ? props.subtitle : undefined,
)

/** 将历史数值/百分比宽度映射到 RsDialog 的 sm | md | lg */
const dialogWidth = computed<'sm' | 'md' | 'lg'>(() => {
  const w = props.width
  if (w === 'sm' || w === 'md' || w === 'lg') return w
  const n = typeof w === 'number' ? w : Number.parseInt(String(w), 10)
  if (!Number.isFinite(n)) return 'md'
  if (n >= 900) return 'lg'
  if (n <= 480) return 'sm'
  return 'md'
})

const footerAlign = computed(() => props.footerButtonAlign ?? 'end')

function handleUpdateShow(value: boolean) {
  emit('update:show', value)
  if (!value) emit('close')
}

function handleCancel() {
  emit('cancel')
  emit('update:show', false)
}

function handleConfirm() {
  emit('confirm')
  if (props.autoCloseOnConfirm) {
    emit('update:show', false)
  }
}

function handleClose() {
  emit('close')
  emit('update:show', false)
}

defineExpose({ close: handleClose })
</script>

<style scoped lang="scss">
.g-dialog__body {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  min-height: 0;
}

.g-dialog__subtitle {
  margin: 0;
  color: var(--g-text-secondary, var(--rs-muted));
  font-size: var(--rs-font-size-sm);
}

.g-dialog__content {
  min-width: 0;
}

.g-dialog__footer {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
}

.g-dialog__footer-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.g-dialog__footer-subtitle {
  color: var(--g-text-secondary, var(--rs-muted));
  font-size: var(--rs-font-size-xs);
}
</style>
