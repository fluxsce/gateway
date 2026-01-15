<template>
  <n-card
    :class="[
      'g-card',
      {
        'g-card--no-title': !showTitle,
        'g-card--hoverable': hoverable === true || hoverable === 'hover',
        'g-card--shadow-always': hoverable === 'always'
      },
      props.class
    ]"
    :title="showTitle ? (title || '') : undefined"
    :bordered="bordered"
    :size="size"
    :segmented="segmented"
    :embedded="embedded"
    :content-style="computedContentStyle"
    :header-style="headerStyle"
    :style="style"
  >
    <!-- 标题插槽 -->
    <template v-if="showTitle && $slots.header" #header>
      <slot name="header" />
    </template>

    <!-- 头部扩展插槽 -->
    <template v-if="showTitle && $slots['header-extra']" #header-extra>
      <slot name="header-extra" />
    </template>

    <!-- 操作插槽 -->
    <template v-if="$slots.action" #action>
      <slot name="action" />
    </template>

    <!-- 封面插槽 -->
    <template v-if="$slots.cover" #cover>
      <slot name="cover" />
    </template>

    <!-- 底部插槽 -->
    <template v-if="$slots.footer" #footer>
      <slot name="footer" />
    </template>

    <!-- 默认内容插槽 -->
    <slot />
  </n-card>
</template>

<script setup lang="ts">
import { NCard } from 'naive-ui';
import { computed } from 'vue';
import type { GCardEmits, GCardProps } from './types';

defineOptions({
  name: 'GCard'
})

const props = withDefaults(defineProps<GCardProps>(), {
  showTitle: false,
  bordered: false,
  hoverable: false,
  size: 'medium',
  embedded: false
})

defineEmits<GCardEmits>()

// 默认内容样式：无 padding
const computedContentStyle = computed(() => {
  if (props.contentStyle) {
    return props.contentStyle
  }
  return { padding: 0 }
})
</script>

<style scoped lang="scss">
.g-card {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;

  // 不显示标题时，移除头部区域（但如果有 header-extra 插槽，仍然显示）
  &.g-card--no-title {
    :deep(.n-card-header) {
      display: none;
    }
  }

  // 确保当 showTitle 为 true 时，header 区域可见
  &:not(.g-card--no-title) {
    :deep(.n-card-header) {
      display: flex !important;
    }
  }

  // hover 效果
  &.g-card--hoverable {
    transition: box-shadow 0.3s var(--n-bezier);
    cursor: default;

    &:hover {
      box-shadow: var(--n-box-shadow);
    }
  }

  // 总是显示阴影
  &.g-card--shadow-always {
    box-shadow: var(--n-box-shadow);
  }

  // 确保内容区域占满剩余空间
  :deep(.n-card__content) {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  // 确保底部区域在底部
  :deep(.n-card__footer) {
    margin-top: auto;
  }
}
</style>

