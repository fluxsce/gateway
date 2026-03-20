<template>
  <n-select
    ref="selectRef"
    class="g-select"
    :value="props.value"
    :options="props.options"
    :placeholder="props.placeholder"
    :disabled="props.disabled"
    :clearable="props.clearable"
    :size="props.size"
    :multiple="props.multiple"
    :filterable="props.filterable"
    v-bind="$attrs"
    @update:value="handleUpdateValue"
  />
</template>

<script setup lang="ts">
import { NSelect } from 'naive-ui'
import { ref } from 'vue'
import type { GSelectEmits, GSelectProps, GSelectValue, GSelectInstance } from './types'

defineOptions({
  name: 'GSelect',
})

const props = withDefaults(defineProps<GSelectProps>(), {
  value: null,
  options: () => [],
  placeholder: '请选择',
  disabled: false,
  clearable: true,
  size: 'small',
  multiple: false,
  filterable: false,
})

const emit = defineEmits<GSelectEmits>()

const selectRef = ref<InstanceType<typeof NSelect> | null>(null)

function handleUpdateValue(value: GSelectValue) {
  emit('update:value', value)
}

function focus() {
  selectRef.value?.focus()
}

function blur() {
  selectRef.value?.blur()
}

defineExpose<GSelectInstance>({
  focus,
  blur,
})
</script>

<style scoped lang="scss">
.g-select {
  display: inline-flex;
  width: 100%;
  outline: none;
  border-radius: var(--g-radius-sm, 4px);
  color: var(--g-text-primary);
  transition:
    border-color var(--g-transition-base) var(--g-transition-ease),
    box-shadow var(--g-transition-base) var(--g-transition-ease),
    background-color var(--g-transition-base) var(--g-transition-ease);

  :deep(.n-base-selection) {
    border-radius: var(--g-radius-sm, 4px);
  }

  :deep(.n-base-selection--disabled) {
    cursor: not-allowed;
    opacity: 0.6;
    color: var(--g-text-disabled);
  }
}
</style>

<style lang="scss">
/* 下拉面板与 GDropdown 一致：边框、选项占满整行、左右内边距 12px */
.n-base-select-menu {
  border: 1px solid var(--g-border-primary);
}

.n-base-select-option::before {
  left: 0 !important;
  right: 0 !important;
}

.n-base-select-option {
  padding: 0 12px;
  box-sizing: border-box;
}
</style>
