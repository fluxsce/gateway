<template>
  <n-tooltip
    :trigger="trigger"
    :placement="placement"
    :show-arrow="showArrow"
    :style="{ maxWidth: `${maxWidth}px` }"
    :delay="delay"
    :duration="duration"
  >
    <template #trigger>
      <n-icon
        :size="iconSize"
        :class="['g-tips__icon', `g-tips__icon--${iconSize}`]"
        :style="iconStyle"
      >
        <InformationCircleOutline />
      </n-icon>
    </template>
    <template #default>
      <div class="g-tips__content">
        <slot>
          {{ content }}
        </slot>
      </div>
    </template>
  </n-tooltip>
</template>

<script setup lang="ts">
import { InformationCircleOutline } from '@vicons/ionicons5';
import { NIcon, NTooltip } from 'naive-ui';
import type { GTipsProps } from './types';

defineOptions({
  name: 'GTips'
})

const props = withDefaults(defineProps<GTipsProps>(), {
  content: '',
  trigger: 'hover',
  placement: 'top',
  showArrow: true,
  maxWidth: 320,
  delay: 0,
  duration: 100,
  iconSize: 16,
  iconStyle: () => ({})
})
</script>

<style scoped lang="scss">
.g-tips {
  &__icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--g-text-color-secondary, rgba(0, 0, 0, 0.45));
    cursor: help;
    transition: color 0.2s ease;

    &:hover {
      color: var(--g-primary-color, #2080f0);
    }

    &--12 {
      font-size: 12px;
    }

    &--14 {
      font-size: 14px;
    }

    &--16 {
      font-size: 16px;
    }

    &--18 {
      font-size: 18px;
    }

    &--20 {
      font-size: 20px;
    }
  }

  &__content {
    line-height: 1.5;
    word-wrap: break-word;
    white-space: pre-wrap;
  }
}
</style>

