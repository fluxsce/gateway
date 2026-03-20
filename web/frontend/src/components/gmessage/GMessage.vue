<template>
  <Transition name="g-message-fade" appear>
    <div
      v-show="visible"
      :class="['g-message', `g-message--${type}`, props.className]"
      :style="computedStyle"
    >
      <GIcon
        v-if="showIcon && computedIcon"
        :icon="computedIcon"
        class="g-message__icon"
        :class="`g-message__icon--${type}`"
        size="small"
      />
      <div class="g-message__content">
        <slot>{{ typeof content === 'string' ? content : '' }}</slot>
      </div>
      <GIcon
        v-if="closable"
        icon="CloseOutline"
        class="g-message__close"
        size="small"
        @click="handleClose"
      />
    </div>
  </Transition>
</template>

<script setup lang="ts">
/**
 * 单条消息展示组件（由 GMessageProvider 按条渲染）
 * 支持 type、content、closable、showIcon、自定义 icon/className/style，以及 close / after-close 事件
 */
import { GIcon } from '@/components/gicon'
import {
  CheckmarkCircleOutline,
  CloseCircleOutline,
  InformationCircleOutline,
  WarningOutline,
} from '@vicons/ionicons5'
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import type { Component } from 'vue'
import type { GMessageType } from './types'

defineOptions({
  name: 'GMessage',
})

const props = withDefaults(
  defineProps<{
    type?: GMessageType
    content?: string
    closable?: boolean
    showIcon?: boolean
    show?: boolean
    icon?: Component
    className?: string
    style?: string | Record<string, string | number>
  }>(),
  {
    type: 'info',
    content: '',
    closable: false,
    showIcon: true,
    show: true,
  }
)

const emit = defineEmits<{ close: []; 'after-close': [] }>()

const visible = ref(props.show)
let afterCloseTimer: ReturnType<typeof setTimeout> | undefined

const defaultIcons: Record<GMessageType, Component> = {
  info: InformationCircleOutline,
  success: CheckmarkCircleOutline,
  warning: WarningOutline,
  error: CloseCircleOutline,
  loading: InformationCircleOutline, // 可选后续用 Loading 图标
}

const computedIcon = computed(() => (props.showIcon && props.icon ? props.icon : defaultIcons[props.type]))

const computedStyle = computed((): string | Record<string, string | number> =>
  typeof props.style === 'string' ? props.style : props.style ?? {}
)

function handleClose() {
  visible.value = false
  emit('close')
  if (afterCloseTimer !== undefined) {
    clearTimeout(afterCloseTimer)
    afterCloseTimer = undefined
  }
  afterCloseTimer = setTimeout(() => {
    emit('after-close')
    afterCloseTimer = undefined
  }, 300)
}

watch(
  () => props.show,
  (val) => {
    if (val === false && visible.value) handleClose()
    else visible.value = val
  }
)

onBeforeUnmount(() => {
  if (afterCloseTimer !== undefined) {
    clearTimeout(afterCloseTimer)
    afterCloseTimer = undefined
  }
})

defineExpose({
  close: handleClose,
  show: () => { visible.value = true },
})
</script>

<style scoped lang="scss">
.g-message {
  display: flex;
  align-items: flex-start;
  gap: var(--g-space-sm);
  padding: var(--g-padding-sm) var(--g-padding-md);
  border-radius: var(--g-radius-md);
  font-size: 14px;
  line-height: 1.5;
  min-width: 300px;
  max-width: 500px;
  background: var(--g-bg-secondary);
  border: 1px solid var(--g-border-primary);
  box-shadow: var(--g-shadow-sm);
  transition: all var(--g-transition-base) var(--g-transition-ease);

  &__icon {
    flex-shrink: 0;

    &--info {
      color: var(--g-info);
    }
    &--success {
      color: var(--g-success);
    }
    &--warning {
      color: var(--g-warning);
    }
    &--error {
      color: var(--g-error);
    }
    &--loading {
      color: var(--g-primary);
    }
  }

  &__content {
    flex: 1;
    min-width: 0;
    color: var(--g-text-primary);
  }

  &__close {
    flex-shrink: 0;
    cursor: pointer;
    color: var(--g-text-tertiary);
    &:hover {
      color: var(--g-text-primary);
    }
  }
}

.g-message-fade-enter-active,
.g-message-fade-leave-active {
  transition: opacity var(--g-transition-base) var(--g-transition-ease),
    transform var(--g-transition-base) var(--g-transition-ease);
}
.g-message-fade-enter-from,
.g-message-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
