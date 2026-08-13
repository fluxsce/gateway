<template>
  <RsDialog
    :open="props.visible"
    :title="dialogTitle"
    layout="window"
    :width="props.width"
    :teleport-to="props.to"
    :draggable="props.draggable"
    :resizable="props.resizable"
    :fullscreenable="props.showFullscreenToggle"
    :show-overlay="props.mask"
    :show-close="props.closable"
    :close-on-overlay-click="props.maskClosable"
    :modal="props.trapFocus !== false"
    :class="modalRootClassMerged"
    :data-g-modal="modalInstanceId"
    @update:open="handleUpdateVisible"
  >
    <template #title>
      <span class="g-modal__title">
        <span v-if="resolvedHeaderIcon" class="g-modal__title-icon" aria-hidden="true">
          <component :is="resolvedHeaderIcon" v-if="isHeaderIconComponent" />
          <RsIcon v-else :name="String(resolvedHeaderIcon)" size="sm" />
        </span>
        <span class="g-modal__title-text">{{ dialogTitle }}</span>
      </span>
    </template>
    <template #body>
      <div class="g-modal__body">
        <slot />
      </div>
    </template>
    <template #footer>
      <slot name="footer">
        <div v-if="props.showFooter" class="g-modal__footer">
          <RsButton
            v-for="btn in props.footerToolbar"
            :key="btn.key"
            size="sm"
            v-bind="btn.buttonProps"
            @click="handleToolbarClick(btn.key)"
          >
            {{ btn.label }}
          </RsButton>
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
      </slot>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import { RsButton, RsDialog, RsIcon } from '@/ui'
import { computed, useAttrs, useSlots } from 'vue'
import { randomUUID } from '@/utils/uuid'
import type { GModalEmits, GModalProps } from './types'

defineOptions({
  /** @deprecated 请改用 RsDialog（@/ui），本组件将逐步淘汰 */
  name: 'GModal',
  inheritAttrs: false,
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
  // 与 RsDialog / RsDataFormModal 一致：默认不因外部点击关闭，避免右键菜单选中后弹窗立刻被关掉
  maskClosable: false,
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
const attrs = useAttrs()
const modalInstanceId = randomUUID()

const dialogTitle = computed(() => props.title || ' ')

/** dialog 预设默认带信息图标；可用 headerIcon 覆盖，传 null 关闭 */
const resolvedHeaderIcon = computed(() => {
  if (props.headerIcon === null) return null
  if (props.headerIcon) return props.headerIcon
  if (props.preset === 'dialog') return 'info'
  return null
})

const isHeaderIconComponent = computed(() => {
  const icon = resolvedHeaderIcon.value
  return typeof icon === 'object' && icon !== null
})

const modalRootClass = computed(() => ({
  'g-modal': true,
  'g-modal--no-footer': !props.showFooter && !slots.footer,
}))

const modalRootClassMerged = computed(() => {
  const extra = attrs.class
  if (!extra) return modalRootClass.value
  return [modalRootClass.value, extra]
})

function handleUpdateVisible(value: boolean) {
  emit('update:visible', value)
  if (!value) emit('close')
}

function handleCancel() {
  emit('cancel')
  emit('update:visible', false)
}

function handleConfirm() {
  emit('confirm')
}

function handleClose() {
  emit('close')
  emit('update:visible', false)
}

function handleToolbarClick(key: string) {
  emit('toolbar-click', key)
}

defineExpose({
  close: handleClose,
})
</script>

<style scoped lang="scss">
.g-modal__title {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.g-modal__title-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 22px;
  height: 22px;
  border-radius: 999px;
  color: #fff;
  background: var(--rs-primary, var(--g-primary));
}

.g-modal__title-icon :deep(svg) {
  width: 14px;
  height: 14px;
}

.g-modal__title-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.g-modal__body {
  min-height: 0;
  flex: 1 1 auto;
  overflow: auto;
}

.g-modal__footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
}
</style>
