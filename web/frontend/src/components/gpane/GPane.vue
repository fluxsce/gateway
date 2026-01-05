<template>
  <div class="g-pane" :class="`g-pane--${direction}`">
    <!-- noResize 模式：使用 flex 布局，面板根据内容自适应 -->
    <div v-if="noResize" class="g-pane__flex-container" :class="`g-pane__flex-container--${direction}`">
      <div class="g-pane__flex-pane g-pane__flex-pane--1" :class="pane1Class" :style="pane1Style">
        <slot name="1">
          <slot name="pane1" />
        </slot>
      </div>
      <div class="g-pane__flex-pane g-pane__flex-pane--2" :class="pane2Class" :style="pane2Style">
        <slot name="2">
          <slot name="pane2" />
        </slot>
      </div>
    </div>

    <!-- 正常模式：使用 n-split 进行可拖拽分割 -->
    <n-split
      v-else
      :direction="direction"
      :default-size="defaultSize"
      :size="size"
      :min="min"
      :max="max"
      :resize-trigger-size="resizeTriggerSize"
      :disabled="disabled"
      :pane1-class="pane1Class"
      :pane1-style="pane1Style"
      :pane2-class="pane2Class"
      :pane2-style="pane2Style"
      @update:size="handleUpdateSize"
      :on-drag-start="handleDragStart"
      :on-drag-end="handleDragEnd"
    >
      <!-- 上/左 面板：兼容 NSplit 的 #1 插槽，同时支持自定义 pane1 插槽 -->
      <template #1>
        <slot name="1">
          <slot name="pane1" />
        </slot>
      </template>

      <!-- 下/右 面板：兼容 NSplit 的 #2 插槽，同时支持自定义 pane2 插槽 -->
      <template #2>
        <slot name="2">
          <slot name="pane2" />
        </slot>
      </template>
    </n-split>
  </div>
</template>

<script setup lang="ts">
import { NSplit } from 'naive-ui';
import type { GPaneEmits, GPaneProps } from './types';

defineOptions({
  name: 'GPane'
})

const props = withDefaults(defineProps<GPaneProps>(), {
  direction: 'vertical',
  defaultSize: 0.3,
  min: 0,
  max: 1,
  disabled: false,
  noResize: false,
  resizeTriggerSize: 2
})

const emit = defineEmits<GPaneEmits>()

const handleUpdateSize = (size: number | string) => {
  emit('update:size', size)
}

const handleDragStart = (e: Event) => {
  emit('drag-start', e)
}

const handleDragEnd = (e: Event) => {
  emit('drag-end', e)
}
</script>

<style scoped lang="scss">
.g-pane {
  width: 100%;
  height: 100%;

  &--vertical {
    display: flex;
    flex-direction: column;
  }

  &--horizontal {
    display: flex;
    flex-direction: row;
  }

  :deep(.n-split) {
    width: 100%;
    height: 100%;
  }

  // 可以按需在这里通过 :deep(.n-split__resize-trigger-wrapper) 自定义分割条样式

  /* noResize 模式：使用 flex 布局，面板根据内容自适应 */
  .g-pane__flex-container {
    width: 100%;
    height: 100%;
    display: flex;

    /* 垂直方向：上下分割 */
    &--vertical {
      flex-direction: column;

      .g-pane__flex-pane--1 {
        flex: 0 0 auto;
        width: 100%;
        min-height: 0;
      }

      .g-pane__flex-pane--2 {
        flex: 1 1 auto;
        width: 100%;
        min-height: 0;
        overflow: hidden;
      }
    }

    /* 水平方向：左右分割 */
    &--horizontal {
      flex-direction: row;

      .g-pane__flex-pane--1 {
        flex: 0 0 auto;
        height: 100%;
        min-width: 0;
      }

      .g-pane__flex-pane--2 {
        flex: 1 1 auto;
        height: 100%;
        min-width: 0;
        overflow: hidden;
      }
    }
  }

  /* 面板通用样式 */
  .g-pane__flex-pane {
    display: flex;
    flex-direction: column;
  }
}
</style>


