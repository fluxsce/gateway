<template>
  <div
    :class="[
      'g-icon',
      { 'g-icon--disabled': disabled, 'g-icon--spin': spin },
      props.class,
    ]"
    :style="computedStyle"
    role="img"
    @click="handleClick"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
  >
    <component
      :is="iconComponent"
      v-if="iconComponent"
      :width="computedSize"
      :height="computedSize"
      :style="{ color: computedColor, width: `${computedSize}px`, height: `${computedSize}px` }"
    />
    <slot v-else />
  </div>
</template>

<script setup lang="ts">
import type { Component, CSSProperties } from 'vue'
import { computed, markRaw, ref } from 'vue'
import { getIcon, getIconSync, IconLibrary } from '@/utils/icon'
import type { GIconColor, GIconEmits, GIconProps } from './types'
import { G_ICON_SIZE_MAP } from './types'

defineOptions({
  name: 'GIcon',
  inheritAttrs: false,
})

const props = withDefaults(defineProps<GIconProps>(), {
  size: 'medium',
  disabled: false,
  spin: false,
  spinSpeed: 2,
})

const emit = defineEmits<GIconEmits>()

const rootEl = ref<HTMLElement | null>(null)
const iconLoadTrigger = ref(0)

const resolvedLibrary = computed(() => {
  if (props.library === 'antd') return IconLibrary.ANTD
  return IconLibrary.IONICONS5
})

const iconComponent = computed<Component | undefined>(() => {
  void iconLoadTrigger.value
  const icon = props.icon
  if (!icon) return undefined

  if (typeof icon === 'string') {
    const cached = getIconSync(icon, resolvedLibrary.value)
    if (cached) return markRaw(cached)
    getIcon(icon, resolvedLibrary.value)
      .then((component) => {
        if (component) iconLoadTrigger.value++
      })
      .catch(() => {})
    return undefined
  }

  return markRaw(icon as Component)
})

const computedSize = computed(() => {
  if (typeof props.size === 'number') return props.size
  if (typeof props.size === 'string' && /^\d+$/.test(props.size)) return Number(props.size)
  return G_ICON_SIZE_MAP[props.size] ?? G_ICON_SIZE_MAP.medium
})

const computedColor = computed(() => {
  if (!props.color) return undefined
  const presets: Record<string, string> = {
    primary: 'var(--g-primary)',
    success: 'var(--g-success, #18a058)',
    warning: 'var(--g-warning, #f0a020)',
    error: 'var(--g-error, #d03050)',
    info: 'var(--g-info, #2080f0)',
  }
  return presets[props.color as GIconColor] ?? props.color
})

const computedStyle = computed((): CSSProperties => {
  const style: CSSProperties = {
    width: `${computedSize.value}px`,
    height: `${computedSize.value}px`,
    color: computedColor.value,
  }
  if (props.style) {
    if (typeof props.style === 'string') return props.style as unknown as CSSProperties
    Object.assign(style, props.style)
  }
  if (props.spin) {
    style.animation = `g-icon-spin ${props.spinSpeed}s linear infinite`
  }
  if (props.disabled) {
    style.opacity = '0.4'
    style.cursor = 'not-allowed'
  }
  return style
})

function handleClick(e: MouseEvent) {
  if (!props.disabled) emit('click', e)
}

function handleMouseEnter(e: MouseEvent) {
  emit('mouseenter', e)
}

function handleMouseLeave(e: MouseEvent) {
  emit('mouseleave', e)
}

defineExpose<{ $el: HTMLElement | null }>({
  get $el() {
    return rootEl.value
  },
})
</script>

<style scoped lang="scss">
.g-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: opacity 0.2s ease;

  :deep(svg) {
    display: block;
    width: 100%;
    height: 100%;
  }

  &--disabled {
    pointer-events: none;
  }

  &--spin {
    display: inline-block;
  }
}

@keyframes g-icon-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
