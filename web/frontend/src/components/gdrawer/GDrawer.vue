<template>
  <RsDrawer
    :open="props.show"
    :title="props.title"
    :side="props.placement"
    :size="drawerSize"
    :width="drawerWidth"
    :show-overlay="props.mask"
    :show-close="props.closable"
    :close-on-overlay-click="props.maskClosable"
    class="g-drawer"
    @update:open="handleUpdateShow"
  >
    <div class="g-drawer__body" :class="props.bodyClass" :style="props.bodyStyle">
      <slot />
    </div>
    <template #footer>
      <slot name="footer">
        <div v-if="props.showFooter" class="g-drawer__footer">
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
  </RsDrawer>
</template>

<script setup lang="ts">
import { RsButton, RsDrawer } from '@/ui'
import { computed } from 'vue'
import type { GDrawerEmits, GDrawerProps } from './types'

defineOptions({
  name: 'GDrawer',
})

const props = withDefaults(defineProps<GDrawerProps>(), {
  show: false,
  title: '',
  width: 400,
  placement: 'right',
  mask: true,
  blockScroll: true,
  maskClosable: true,
  closable: true,
  showFooter: true,
  showCancel: true,
  showConfirm: true,
  cancelText: '取消',
  confirmText: '确定',
  confirmLoading: false,
  autoFocus: true,
  resizable: false,
})

const emit = defineEmits<GDrawerEmits>()

const drawerSize = computed(() => {
  const w = typeof props.width === 'number' ? props.width : parseInt(String(props.width), 10)
  if (!Number.isFinite(w)) return 'md'
  if (w >= 720) return 'lg'
  if (w <= 320) return 'sm'
  return 'md'
})

/** 透传精确宽度，覆盖 size 预设，与历史 GDrawer width 对齐 */
const drawerWidth = computed(() => {
  if (typeof props.width === 'number' && Number.isFinite(props.width) && props.width > 0) {
    return props.width
  }
  const trimmed = String(props.width ?? '').trim()
  return trimmed || undefined
})

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
}

defineExpose({
  close: () => emit('update:show', false),
})
</script>

<style scoped lang="scss">
.g-drawer__body {
  min-height: 0;
  height: 100%;
  overflow: auto;
}

.g-drawer__footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  width: 100%;
}
</style>
