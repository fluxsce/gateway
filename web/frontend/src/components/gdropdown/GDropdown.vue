<template>
  <RsDropdown
    v-model="selected"
    class="g-dropdown"
    :items="(rsItems as any)"
    :disabled="props.disabled"
    :show-selected="false"
    placeholder=" "
    @select="handleSelect"
  >
    <template #trigger>
      <div ref="triggerRef" class="g-dropdown__trigger" :class="{ 'is-disabled': props.disabled }">
        <slot />
      </div>
    </template>
  </RsDropdown>
</template>

<script setup lang="ts">
import { RsDropdown, type RsDropdownItems } from '@/ui'
import { computed, ref } from 'vue'
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
const selected = ref<string>()
const triggerRef = ref<HTMLElement | null>(null)
const visible = ref(false)

const rsItems = computed(() =>
  (props.options || [])
    .filter((opt: any) => opt.type !== 'divider' && opt.key !== 'divider')
    .map((opt: any) => ({
      label: String(opt.label ?? opt.key ?? ''),
      value: String(opt.key ?? opt.value ?? ''),
      disabled: Boolean(opt.disabled),
    })),
)

function handleSelect(value: string) {
  const option = (props.options || []).find((opt: any) => String(opt.key ?? opt.value) === value)
  emit('select', value, option ?? { key: value, label: value })
}

defineExpose<GDropdownInstance>({
  close: () => {
    visible.value = false
  },
  open: () => {
    visible.value = true
  },
  get visible() {
    return visible.value
  },
})
</script>

<style scoped lang="scss">
.g-dropdown {
  display: inline-flex;
}

.g-dropdown__trigger {
  cursor: pointer;

  &.is-disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }
}
</style>
