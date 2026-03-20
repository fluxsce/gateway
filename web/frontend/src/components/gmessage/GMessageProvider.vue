<template>
  <Teleport to="body">
    <div
      :class="['g-message-provider', `g-message-provider--${position}`, props.containerClass]"
      :style="containerStyle"
    >
      <TransitionGroup
        name="g-message-list"
        tag="div"
        class="g-message-provider__container"
      >
        <GMessage
          v-for="item in displayItems"
          :key="item.id"
          :type="item.type"
          :content="typeof item.content === 'string' ? item.content : ''"
          :closable="item.closable"
          :show="item.visible"
          :class-name="item.className"
          :style="item.style"
          @close="handleItemClose(item.id)"
          @after-close="handleItemAfterClose(item.id)"
        >
          <template v-if="typeof item.content !== 'string'">
            <component :is="item.content" />
          </template>
        </GMessage>
      </TransitionGroup>
    </div>
  </Teleport>
  <slot />
</template>

<script setup lang="ts">
/**
 * 某位置的消息 Provider：维护消息列表、定时关闭、最大条数，并向 utils 注册 message API，
 * 供 $gMessage.info/success/error/warning/loading/destroyAll 调用。
 */
import { shallowRef, computed, onMounted, onBeforeUnmount } from 'vue'
import type { Component, VNode } from 'vue'
import { setMessageProvider } from './utils'
import GMessage from './GMessage.vue'
import type {
  GMessageProviderProps,
  GMessageOptions,
  GMessageType,
  GMessageProviderApi,
} from './types'

defineOptions({
  name: 'GMessageProvider',
})

const props = withDefaults(defineProps<GMessageProviderProps>(), {
  position: 'top',
  duration: 3000,
  closable: true,
  max: 5,
})

interface MessageItem {
  id: string
  content: string | VNode | Component
  type: GMessageType
  duration: number
  closable: boolean
  className?: string
  style?: string | Record<string, string | number>
  position: string
  visible: boolean
  timer?: ReturnType<typeof setTimeout>
  /** 关闭动画结束后移除条目的定时器，卸载时需 clear 避免泄漏 */
  afterCloseTimer?: ReturnType<typeof setTimeout>
  onClose?: () => void
  onAfterClose?: () => void
}

const items = shallowRef<MessageItem[]>([])
let idCounter = 0

function generateId() {
  return `g-message-${Date.now()}-${++idCounter}`
}

function normalizeOptions(
  content: string | VNode | Component | GMessageOptions,
  options?: GMessageOptions
): GMessageOptions & { content: string | VNode | Component } {
  if (typeof content === 'string') {
    return { content, ...options }
  }
  if (content && typeof content === 'object' && ('render' in content || 'setup' in content || 'template' in content)) {
    return { content: content as VNode | Component, ...options }
  }
  const opts = content as GMessageOptions
  return { content: opts.content ?? '', ...opts, ...options }
}

const displayItems = computed(() => {
  const list = items.value.filter((i) => i.visible && i.position === props.position)
  if (list.length <= props.max) return list
  return list.slice(-props.max)
})

const containerStyle = computed((): string | Record<string, string | number> =>
  typeof props.containerStyle === 'string' ? props.containerStyle : props.containerStyle ?? {}
)

function createMessage(options: GMessageOptions & { content: string | VNode | Component }, defaultType: GMessageType): string {
  const opts = normalizeOptions(options.content ?? '', options)
  const id = generateId()
  const item: MessageItem = {
    id,
    content: opts.content ?? '',
    type: opts.type ?? defaultType,
    duration: opts.duration ?? props.duration,
    closable: opts.closable ?? props.closable,
    className: opts.className,
    style: opts.style,
    position: opts.position ?? props.position,
    visible: true,
    onClose: opts.onClose,
    onAfterClose: opts.onAfterClose,
  }
  items.value = [...items.value, item]

  const visibleCount = items.value.filter((i) => i.visible && i.position === props.position).length
  if (visibleCount > props.max) {
    const toClose = items.value.filter((i) => i.visible && i.position === props.position).slice(0, visibleCount - props.max)
    toClose.forEach((i) => handleItemClose(i.id))
  }

  const duration = item.duration ?? props.duration
  if (duration > 0) {
    item.timer = setTimeout(() => handleItemClose(id), duration)
  }
  return id
}

function handleItemClose(id: string) {
  const item = items.value.find((i) => i.id === id)
  if (!item) return
  if (item.timer !== undefined) {
    clearTimeout(item.timer)
    item.timer = undefined
  }
  if (item.onClose) {
    try {
      item.onClose()
    } catch (_) {}
  }
  item.visible = false
  item.afterCloseTimer = setTimeout(() => {
    item.afterCloseTimer = undefined
    handleItemAfterClose(id)
  }, 300)
}

function handleItemAfterClose(id: string) {
  const item = items.value.find((i) => i.id === id)
  if (!item) return
  if (item.afterCloseTimer !== undefined) {
    clearTimeout(item.afterCloseTimer)
    item.afterCloseTimer = undefined
  }
  if (item.onAfterClose) {
    try {
      item.onAfterClose()
    } catch (_) {}
  }
  items.value = items.value.filter((i) => i.id !== id)
}

function destroyAll() {
  items.value.forEach((item) => {
    if (item.timer !== undefined) clearTimeout(item.timer)
    if (item.afterCloseTimer !== undefined) clearTimeout(item.afterCloseTimer)
    if (item.onClose) {
      try {
        item.onClose()
      } catch (_) {}
    }
  })
  items.value = []
}

function defaultCall(content: string | GMessageOptions, opts?: GMessageOptions) {
  createMessage(normalizeOptions(content ?? '', opts), 'info')
}

const messageApi = Object.assign(defaultCall, {
  success: (c: string | GMessageOptions, o?: GMessageOptions) =>
    createMessage(normalizeOptions(c, o), 'success'),
  error: (c: string | GMessageOptions, o?: GMessageOptions) =>
    createMessage(normalizeOptions(c, o), 'error'),
  warning: (c: string | GMessageOptions, o?: GMessageOptions) =>
    createMessage(normalizeOptions(c, o), 'warning'),
  info: (c: string | GMessageOptions, o?: GMessageOptions) =>
    createMessage(normalizeOptions(c, o), 'info'),
  loading: (c: string | GMessageOptions, o?: GMessageOptions) =>
    createMessage(normalizeOptions(c, o), 'loading'),
  destroyAll,
}) as GMessageProviderApi

onMounted(() => {
  setMessageProvider(props.position, { message: messageApi })
})

onBeforeUnmount(() => {
  items.value.forEach((item) => {
    if (item.timer !== undefined) clearTimeout(item.timer)
    if (item.afterCloseTimer !== undefined) clearTimeout(item.afterCloseTimer)
  })
  items.value = []
  setMessageProvider(props.position, null)
})

defineExpose({
  message: messageApi,
  destroyAll,
})
</script>

<style scoped lang="scss">
.g-message-provider {
  position: fixed;
  z-index: 6000;
  pointer-events: none;

  &--top {
    top: 20px;
    left: 50%;
    transform: translateX(-50%);
  }
  &--top-left {
    top: 20px;
    left: 20px;
  }
  &--top-right {
    top: 20px;
    right: 20px;
  }
  &--bottom {
    bottom: 20px;
    left: 50%;
    transform: translateX(-50%);
  }
  &--bottom-left {
    bottom: 20px;
    left: 20px;
  }
  &--bottom-right {
    bottom: 20px;
    right: 20px;
  }

  &__container {
    display: flex;
    flex-direction: column;
    gap: var(--g-space-sm);
    pointer-events: auto;
    align-items: center;
    min-height: 0;
  }
}

.g-message-list-move,
.g-message-list-enter-active,
.g-message-list-leave-active {
  transition: all var(--g-transition-base) var(--g-transition-ease);
}
.g-message-list-enter-from,
.g-message-list-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
.g-message-list-leave-active {
  position: absolute;
}
</style>
