<template>
  <fieldset
    class="g-fieldset"
    :class="{
      [`g-fieldset--border-${borderStyle}`]: borderStyle,
      'g-fieldset--selected': selected,
      'g-fieldset--disabled': disabled
    }"
    :disabled="disabled"
  >
    <!-- 标题区域（使用 legend 元素保持语义化） -->
    <legend
      v-if="title || $slots.title"
      class="g-fieldset__legend"
    >
      <div class="g-fieldset__title-wrapper">
        <slot name="title">
          <span
            v-if="title"
            class="g-fieldset__title"
            :class="{
              'g-fieldset__title--strong': titleStrong,
              [`g-fieldset__title--${titleSize}`]: typeof titleSize === 'string' && titleSize
            }"
            :style="typeof titleSize === 'number' ? { fontWeight: titleSize } : undefined"
          >
            {{ title }}
          </span>
        </slot>
        <slot name="title-extra" />
      </div>
    </legend>

    <!-- 内容区域 -->
    <div class="g-fieldset__content">
      <slot />
    </div>
  </fieldset>
</template>

<script setup lang="ts">
import { provide, watch } from 'vue';
import type { GFieldsetEmits, GFieldsetProps } from './types';

defineOptions({
  name: 'GFieldset'
})

const props = withDefaults(defineProps<GFieldsetProps>(), {
  title: '',
  titleStrong: false,
  titleSize: 200,
  borderStyle: 'dashed',
  selected: false,
  disabled: false
})

const emit = defineEmits<GFieldsetEmits>()

// 向下提供 disabled 状态，供子组件使用
provide('fieldsetDisabled', () => props.disabled)

// 监听 disabled 变化，同步到 provide
watch(
  () => props.disabled,
  () => {
    // provide 的值会自动更新，这里不需要额外操作
  }
)
</script>

<style scoped lang="scss">
.g-fieldset {
  position: relative;
  margin: 0;
  padding: var(--g-space-sm, 8px);
  border: 1px dashed var(--g-border-primary, #e0e0e6);
  border-radius: var(--g-border-radius, 4px);
  background-color: var(--g-bg-color, #ffffff);
  transition: all 0.2s ease;

  // 实线边框
  &.g-fieldset--border-solid {
    border: 1px solid var(--g-border-primary, #e0e0e6);
  }

  // 虚线边框
  &.g-fieldset--border-dashed {
    border: 1px dashed var(--g-border-primary, #e0e0e6);
  }

  // 无边框
  &.g-fieldset--border-none {
    border: none;
    padding: 0;
  }

  // 选中状态（高亮显示）
  &.g-fieldset--selected {
    border-color: var(--g-primary-color, #2080f0);
    background-color: var(--g-primary-color-light, rgba(32, 128, 240, 0.05));
    box-shadow: 0 0 0 2px rgba(32, 128, 240, 0.1);
  }

  // 禁用状态样式
  &.g-fieldset--disabled {
    opacity: 0.6;
    pointer-events: none;
    background-color: var(--g-bg-color-disabled, #f5f5f5);

    // 确保所有子表单控件都被禁用
    :deep(input),
    :deep(select),
    :deep(textarea),
    :deep(button),
    :deep(.n-input),
    :deep(.n-select),
    :deep(.n-textarea),
    :deep(.n-button),
    :deep(.n-switch),
    :deep(.n-checkbox),
    :deep(.n-radio) {
      cursor: not-allowed;
      opacity: 0.6;
    }
  }
}

.g-fieldset__legend {
  position: relative;
  display: flex;
  align-items: center;
  width: auto;
  max-width: 100%;
  margin: 0;
  
  padding: 0 var(--g-space-xs, 8px);
  border: none;
  font-size: inherit;
  font-weight: inherit;
  color: inherit;
  cursor: default;
  float: none; // 防止 legend 导致边框断开

  // 禁用状态下不可点击
  .g-fieldset--disabled & {
    cursor: not-allowed;
    opacity: 0.6;
  }
}

.g-fieldset__title-wrapper {
  display: flex;
  align-items: center;
  gap: var(--g-space-xs, 8px);
  flex: 1;
}

.g-fieldset__title {
  font-size: 14px;
  font-weight: normal; // 默认不加粗
  color: var(--g-text-color, #333);

  // 小号标题
  &.g-fieldset__title--small {
    font-size: 12px;
    font-weight: normal;
  }

  // 正常大小标题（默认）
  &.g-fieldset__title--normal {
    font-size: 14px;
    font-weight: normal;
  }

  // 大号标题
  &.g-fieldset__title--large {
    font-size: 16px;
    font-weight: 600;
  }

  // 加粗样式（覆盖 titleSize 的 font-weight）
  &.g-fieldset__title--strong {
    font-weight: 600;
  }

  .g-fieldset--selected & {
    color: var(--g-primary-color, #2080f0);
  }

  .g-fieldset--disabled & {
    color: var(--g-text-color-disabled, #bbb);
  }
}

</style>

