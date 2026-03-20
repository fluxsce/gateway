<template>
  <n-dropdown
    :show="visible"
    :options="props.options"
    :placement="props.placement"
    :trigger="props.trigger"
    :disabled="props.disabled"
    :show-arrow="false"
    :size="props.size"
    :menu-props="triggerWidth > 0 ? getMenuProps() : undefined"
    class="g-dropdown"
    @select="handleSelect"
    @update:show="visible = $event"
  >
    <div ref="triggerRef" class="g-dropdown__trigger" :class="{ 'is-disabled': props.disabled }">
      <slot />
    </div>
  </n-dropdown>
</template>

<script setup lang="ts">
import type { DropdownMenuProps, DropdownOption } from 'naive-ui'
import { NDropdown } from 'naive-ui'
import { nextTick, ref, watch } from 'vue'
import type { GDropdownEmits, GDropdownInstance, GDropdownProps } from './types'

defineOptions({
  name: 'GDropdown',
})

const props = withDefaults(defineProps<GDropdownProps>(), {
  options: () => [],
  placement: 'bottom-start',
  disabled: false,
  trigger: 'click',
  showArrow: false,
  delay: 0,
  size: 'small',
})

const emit = defineEmits<GDropdownEmits>()

const visible = ref(false)
const triggerRef = ref<HTMLElement | null>(null)
const triggerWidth = ref(0)

watch(visible, (show) => {
  if (show) {
    nextTick(() => {
      triggerWidth.value = triggerRef.value?.offsetWidth ?? 0
    })
  }
})

function getMenuProps(): DropdownMenuProps {
  const w = triggerWidth.value
  return () => (w > 0 ? { style: { minWidth: `${w}px` } } : {}) as ReturnType<DropdownMenuProps>
}

function handleSelect(key: string | number, option: DropdownOption) {
  emit('select', key, option)
}

function close() {
  visible.value = false
}

defineExpose<GDropdownInstance>({
  close,
  open: () => { visible.value = true },
  get visible() { return visible.value },
})
</script>

<style scoped lang="scss">
.g-dropdown {
  display: inline-flex;
}

.g-dropdown__trigger {
  cursor: pointer;
  outline: none;
  //padding: 2px 4px;
  border-radius: var(--g-radius-sm, 4px);
  color: var(--g-text-primary);
  transition:
    border-color var(--g-transition-base) var(--g-transition-ease),
    box-shadow var(--g-transition-base) var(--g-transition-ease),
    background-color var(--g-transition-base) var(--g-transition-ease);

  &:hover:not(.is-disabled) {
    background-color: var(--g-hover-overlay);
  }

  &:focus-visible {
    box-shadow: 0 0 0 2px var(--g-primary-light);
  }

  &.is-disabled {
    cursor: not-allowed;
    opacity: 0.6;
    color: var(--g-text-disabled);
  }
}
</style>

<style lang="scss">
/* 下拉面板与 select 一致：边框、选项占满整行无左右留白、左右内边距 12px（与 xirang-dropdown 一致） */
.n-dropdown-menu {
  border: 1px solid var(--g-border-primary);
}

.n-dropdown-option-body::before {
  left: 0 !important;
  right: 0 !important;
}

.n-dropdown-option-body {
  padding-left: 12px;
  padding-right: 12px;
  box-sizing: border-box;
}
</style>
