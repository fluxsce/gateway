<template>
  <div class="toolbar-button-wrapper">
    <slot name="prefix" />
    
    <!-- 自定义渲染 -->
    <component v-if="button.render" :is="button.render()" />

    <!-- 下拉菜单按钮 -->
    <n-dropdown
      v-else-if="button.dropdown && button.dropdownOptions"
      :options="dropdownMenuOptions"
      @select="handleDropdownSelect"
      trigger="click"
    >
      <n-button
        :type="button.type || 'default'"
        :size="button.size || 'small'"
        :disabled="isButtonDisabled"
        :loading="button.loading"
        quaternary
      >
      <template v-if="button.icon" #icon>
            <n-icon>
              <component :is="iconComponent" />
            </n-icon>
          </template>
        {{ button.label }}
      </n-button>
    </n-dropdown>

    <!-- 普通按钮 -->
    <n-tooltip v-else :disabled="!button.tooltip" trigger="hover">
      <template #trigger>
        <n-button
          :type="button.type || 'default'"
          :size="button.size || 'small'"
          :disabled="isButtonDisabled"
          :loading="button.loading"
          quaternary
          @click="handleClick"
        >
          <template v-if="button.icon" #icon>
            <n-icon>
              <component :is="iconComponent" />
            </n-icon>
          </template>
          {{ button.label }}
        </n-button>
      </template>
      {{ button.tooltip }}
    </n-tooltip>

    <slot name="suffix" />
  </div>
</template>

<script setup lang="ts">
import { store } from '@/stores'
import { renderIcon, renderIconVNode } from '@/utils'
import type { DropdownOption } from 'naive-ui'
import { NButton, NDropdown, NIcon, NTooltip } from 'naive-ui'
import type { Component } from 'vue'
import { computed, toValue } from 'vue'
import type { ToolbarButton } from './types'

interface Props {
  button: ToolbarButton
  moduleId?: string
}

interface Emits {
  (event: 'click', key: string): void
  (event: 'dropdown-select', buttonKey: string, optionKey: string): void
}

defineSlots<{
  prefix?: () => any
  suffix?: () => any
}>()

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 处理图标组件（如果 icon 是字符串，使用 renderIcon；否则直接使用组件）
// 注意：renderIcon 返回的是 ref，使用 toValue 确保响应式更新
const iconComponent = computed<Component | null>(() => {
  if (!props.button.icon) return null
  if (typeof props.button.icon === 'string') {
    const iconRef = renderIcon(props.button.icon)
    // 使用 toValue 确保响应式更新（当 ref 的值变化时，computed 会重新计算）
    return toValue(iconRef)
  }
  // 如果已经是组件，直接返回
  return props.button.icon as Component
})

// 获取按钮权限编码：使用 moduleId:key 作为权限编码
const getButtonPermissionCode = computed(() => {
  if (props.moduleId) {
    return `${props.moduleId}:${props.button.key}`
  }
  // 如果没有 moduleId，返回空（表示不需要权限检查）
  return ''
})

// 检查按钮权限
const hasButtonPermission = computed(() => {
  const permissionCode = getButtonPermissionCode.value
  // 如果没有权限编码（没有 moduleId），默认允许
  if (!permissionCode) {
    return true
  }
  return store.user.hasButton(permissionCode)
})

// 按钮是否禁用（包含权限检查）
const isButtonDisabled = computed(() => {
  return props.button.disabled || !hasButtonPermission.value
})

// 转换下拉菜单选项格式
const dropdownMenuOptions = computed<DropdownOption[]>(() => {
  if (!props.button.dropdownOptions) return []
  
  return props.button.dropdownOptions.map(option => {
    // 检查选项权限：使用 moduleId:option.key 作为权限编码
    let optionDisabled = option.disabled
    if (props.moduleId) {
      const optionPermissionCode = `${props.moduleId}:${option.key}`
      if (!store.user.hasButton(optionPermissionCode)) {
        optionDisabled = true
      }
    }
    
    const menuOption: DropdownOption = {
      key: option.key,
      label: option.label,
      disabled: optionDisabled
    }
    
    // 添加图标（使用 renderIconVNode）
    if (option.icon) {
      menuOption.icon = renderIconVNode(option.icon, NIcon)
    }
    
    // 添加分割线
    if (option.divider) {
      menuOption.type = 'divider'
    }
    
    return menuOption
  })
})

// 处理按钮点击
const handleClick = () => {
  if (!isButtonDisabled.value && !props.button.loading) {
    emit('click', props.button.key)
  }
}

// 处理下拉菜单选择
const handleDropdownSelect = (key: string) => {
  emit('dropdown-select', props.button.key, key)
}
</script>

<style lang="scss" scoped>
.toolbar-button-wrapper {
  display: inline-flex;
  align-items: center;
  gap: var(--g-space-xs);
}
</style>

