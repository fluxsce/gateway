<template>
  <RsButton
    class="toolbar-btn-face"
    variant="text"
    :tone="tone"
    :size="size"
    radius="sm"
    :disabled="disabled"
    :loading="loading"
    @click="emit('click')"
  >
    <GIcon
      v-if="icon && !loading"
      class="toolbar-btn-face__icon"
      :icon="icon"
      :size="14"
    />
    <span
      v-if="label"
      class="toolbar-btn-face__label"
    >{{ label }}</span>
  </RsButton>
</template>

<script setup lang="ts">
import { GIcon } from '@/components/gicon'
import { RsButton } from '@/ui'
import type { RsButtonTone } from 'niuma-ui'
import type { Component } from 'vue'
import { computed } from 'vue'
import type { ToolbarButtonType } from './types'

defineOptions({ name: 'ToolbarBtnFace' })

const props = withDefaults(
  defineProps<{
    label?: string
    icon?: Component | string
    /** 工具栏按钮 type → RsButton tone */
    buttonType?: ToolbarButtonType
    disabled?: boolean
    loading?: boolean
    size?: 'ssm' | 'sm' | 'md' | 'lg'
  }>(),
  {
    buttonType: 'default',
    size: 'sm',
  },
)

const emit = defineEmits<{
  click: []
}>()

const tone = computed<RsButtonTone>(() => {
  switch (props.buttonType) {
    case 'primary':
      return 'primary'
    case 'error':
      return 'danger'
    case 'warning':
      return 'warning'
    case 'success':
      return 'success'
    case 'info':
      return 'info'
    default:
      return 'neutral'
  }
})
</script>

<style scoped lang="scss">
/* 仅补充工具栏场景的图标尺寸，颜色/形态交给 RsButton text+tone */
.toolbar-btn-face__icon {
  display: inline-flex;
  flex-shrink: 0;
  width: 14px;
  height: 14px;
  color: inherit;
}

.toolbar-btn-face__label {
  display: inline-flex;
  align-items: center;
  line-height: 1;
  white-space: nowrap;
}
</style>
