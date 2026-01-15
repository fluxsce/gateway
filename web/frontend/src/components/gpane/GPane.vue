<template>
  <div class="g-pane" :class="`g-pane--${direction}`">
    <!-- noResize 模式：使用 flex 布局，面板根据内容自适应 -->
    <div v-if="noResize" class="g-pane__flex-container" :class="`g-pane__flex-container--${direction}`">
      <div class="g-pane__flex-pane g-pane__flex-pane--1" :class="pane1Class" :style="computedPane1Style">
        <slot name="1">
          <slot name="pane1" />
        </slot>
      </div>
      <div class="g-pane__flex-pane g-pane__flex-pane--2" :class="pane2Class" :style="computedPane2Style">
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
      :size="currentSize"
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
import type { CSSProperties } from 'vue';
import { computed, ref } from 'vue';
import type { GPaneEmits, GPaneExpose, GPaneProps } from './types';

defineOptions({
  name: 'GPane'
})

const props = withDefaults(defineProps<GPaneProps>(), {
  direction: 'vertical',
  min: 0,
  max: 1,
  disabled: false,
  noResize: false,
  resizeTriggerSize: 2
})

const emit = defineEmits<GPaneEmits>()

// 内部状态：面板二的可见性
const pane2Visible = ref<boolean>(true)

// 内部状态：当前面板尺寸
const currentSize = ref<number | string | undefined>(props.size)

const handleUpdateSize = (size: number | string) => {
  // 只有在 pane2Visible 为 true 时才更新 currentSize
  if (pane2Visible.value) {
    currentSize.value = size
    emit('update:size', size)
  }
}

const handleDragStart = (e: Event) => {
  emit('drag-start', e)
}

const handleDragEnd = (e: Event) => {
  emit('drag-end', e)
}


// 将 defaultSize 转换为百分比数值 (0-1)
const normalizedSize = computed(() => {
  // 如果 pane2 不可见，返回 1（左侧占满）
  if (!pane2Visible.value) {
    return 1
  }
  
  const sizeToNormalize = props.defaultSize
  if (typeof sizeToNormalize === 'number') {
    return Math.max(0, Math.min(1, sizeToNormalize))
  }
  if (typeof sizeToNormalize === 'string') {
    // 处理百分比字符串，如 "80%" -> 0.8
    if (sizeToNormalize.endsWith('%')) {
      const percent = parseFloat(sizeToNormalize) / 100
      return Math.max(0, Math.min(1, percent))
    }
    // 如果是数字字符串，尝试解析
    const num = parseFloat(sizeToNormalize)
    if (!isNaN(num)) {
      return Math.max(0, Math.min(1, num <= 1 ? num : num / 100))
    }
  }
  return 0.5 // 默认 50%
})

// 计算 pane1 和 pane2 的 flex 值
const computedPane1Style = computed(() => {
  if (!props.noResize || props.defaultSize === undefined) {
    return props.pane1Style as CSSProperties | string || {}
  }
  
  const baseStyle: CSSProperties = typeof props.pane1Style === 'string' 
    ? {} 
    : (props.pane1Style || {})
  
  const size = normalizedSize.value
  return {
    ...baseStyle,
    flex: `${size} ${size} 0`
  } as CSSProperties
})

const computedPane2Style = computed(() => {
  if (!props.noResize || props.defaultSize === undefined) {
    return props.pane2Style as CSSProperties | string || {}
  }
  
  const baseStyle: CSSProperties = typeof props.pane2Style === 'string' 
    ? {} 
    : (props.pane2Style || {})
  
  const size = normalizedSize.value
  const remainingSize = 1 - size
  return {
    ...baseStyle,
    flex: `${remainingSize} ${remainingSize} 0`
  } as CSSProperties
})

// ============= 暴露的方法 =============

/**
 * 设置面板二（下/右）的可见性
 */
const setPane2Visible = (visible: boolean) => {
  pane2Visible.value = visible
  // 根据可见性设置 currentSize
  if (visible) {
    // 恢复显示，如果 props.size 不存在，设置为 undefined 以使用 defaultSize
    if (props.size === undefined) {
      currentSize.value = props.defaultSize
    }
  } else {
    // 隐藏面板2，设置为 1（左侧占满）
    currentSize.value = 1
  }
}

/**
 * 获取面板二（下/右）的可见性
 */
const getPane2Visible = () => {
  return pane2Visible.value
}

/**
 * 切换面板二（下/右）的可见性
 */
const togglePane2Visible = () => {
  setPane2Visible(!pane2Visible.value)
}

/**
 * 设置面板尺寸
 */
const setSize = (size: number | string) => {
  currentSize.value = size
  emit('update:size', size)
}

/**
 * 获取当前面板尺寸
 */
const getSize = () => {
  return currentSize.value ?? props.defaultSize
}

// 暴露方法
defineExpose<GPaneExpose>({
  setPane2Visible,
  getPane2Visible,
  togglePane2Visible,
  setSize,
  getSize
})
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


