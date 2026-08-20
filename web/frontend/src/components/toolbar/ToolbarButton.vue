<template>
  <div class="toolbar-button-wrapper">
    <slot name="prefix" />

    <component
      v-if="button.render"
      :is="button.render()"
    />

    <RsDropdown
      v-else-if="button.dropdown && button.dropdownOptions"
      :items="dropdownItems"
      :show-selected="false"
      placeholder=" "
      @select="handleDropdownSelect"
    >
      <template #trigger>
        <ToolbarBtnFace
          :label="button.label"
          :icon="button.icon"
          :button-type="button.type"
          :disabled="isButtonDisabled"
          :loading="button.loading"
        />
      </template>
    </RsDropdown>

    <RsTooltip
      v-else
      :content="button.tooltip"
      :disabled="!button.tooltip"
      side="top"
    >
      <ToolbarBtnFace
        :label="button.label"
        :icon="button.icon"
        :button-type="button.type"
        :disabled="isButtonDisabled"
        :loading="button.loading"
        @click="handleClick"
      />
    </RsTooltip>

    <slot name="suffix" />
  </div>
</template>

<script setup lang="ts">
import { store } from '@/stores'
import { RsDropdown, RsTooltip, type RsDropdownItems } from '@/ui'
import { computed } from 'vue'
import ToolbarBtnFace from './ToolbarBtnFace.vue'
import type { ToolbarButton } from './types'

defineOptions({ name: 'ToolbarButton' })

interface Props {
  button: ToolbarButton
  moduleId?: string
}

interface Emits {
  (event: 'click', key: string): void
  (event: 'dropdown-select', buttonKey: string, optionKey: string): void
}

defineSlots<{
  prefix?: () => unknown
  suffix?: () => unknown
}>()

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const getButtonPermissionCode = computed(() => {
  if (!props.moduleId) return props.button.key
  return `${props.moduleId}:${props.button.key}`
})

const isButtonDisabled = computed(() => {
  if (props.button.disabled) return true
  // more 展开条件、以及 skipPermission 的纯 UI 动作不走权限码
  if (props.button.key === 'more' || props.button.skipPermission) return false
  if (!props.moduleId) return false
  const code = getButtonPermissionCode.value
  if (!code) return false
  return !store.user.hasPermission(code)
})

const dropdownItems = computed<RsDropdownItems>(() =>
  (props.button.dropdownOptions || []).map((opt) => ({
    label: String(opt.label ?? opt.key ?? ''),
    value: String(opt.key ?? ''),
    disabled: Boolean(opt.disabled),
  })),
)

function handleClick() {
  if (isButtonDisabled.value || props.button.loading) return
  // onClick 由 GToolbar 统一执行，避免重复触发
  emit('click', props.button.key)
}

function handleDropdownSelect(optionKey: string) {
  emit('dropdown-select', props.button.key, optionKey)
}
</script>

<style scoped lang="scss">
.toolbar-button-wrapper {
  display: inline-flex;
  align-items: center;
}
</style>
