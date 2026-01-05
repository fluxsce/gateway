<template>
  <n-ellipsis v-bind="ellipsisProps">
    <slot>
      {{ text }}
    </slot>
  </n-ellipsis>
</template>

<script setup lang="ts">
import { NEllipsis } from 'naive-ui';
import { computed } from 'vue';
import type { GEllipsisProps } from './types';

defineOptions({
  name: 'GEllipsis'
})

const props = withDefaults(defineProps<GEllipsisProps>(), {
  text: ''
})

// 提取 ellipsis 相关的 props，排除 text
const ellipsisProps = computed(() => {
  const { text, ...rest } = props
  return rest
})
</script>

<style scoped lang="scss">
// 可以在这里添加自定义样式
</style>

