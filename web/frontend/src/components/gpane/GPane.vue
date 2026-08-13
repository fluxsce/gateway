<template>
  <!--
    @deprecated 请直接使用 @/ui 的 RsSplitPane。本组件仅过渡兼容，后续将删除。
  -->
  <div class="g-pane" :class="`g-pane--${direction}`">
    <!-- noResize：flex 布局，上/左面板随内容自适应，下/右占满剩余 -->
    <div
      v-if="noResize"
      class="g-pane__flex-container"
      :class="`g-pane__flex-container--${direction}`"
    >
      <div
        class="g-pane__flex-pane g-pane__flex-pane--1"
        :class="pane1Class"
        :style="computedPane1Style"
      >
        <slot name="1">
          <slot name="pane1" />
        </slot>
      </div>
      <div
        v-show="pane2Visible"
        class="g-pane__flex-pane g-pane__flex-pane--2"
        :class="pane2Class"
        :style="computedPane2Style"
      >
        <slot name="2">
          <slot name="pane2" />
        </slot>
      </div>
    </div>

    <!-- 可拖拽：RsSplitPane -->
    <RsSplitPane
      v-else
      class="g-pane__split"
      :orientation="direction"
      :panes="splitPanes"
      :disabled="disabled"
      with-handle
      v-model:sizes="splitSizes"
      @resize="handleResize"
      @resize-end="handleResizeEnd"
    >
      <template #pane1>
        <div class="g-pane__slot" :class="pane1Class" :style="pane1Style">
          <slot name="1">
            <slot name="pane1" />
          </slot>
        </div>
      </template>
      <template #pane2>
        <div class="g-pane__slot" :class="pane2Class" :style="pane2Style">
          <slot name="2">
            <slot name="pane2" />
          </slot>
        </div>
      </template>
    </RsSplitPane>
  </div>
</template>

<script setup lang="ts">
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import type { CSSProperties } from 'vue'
import { computed, ref, watch } from 'vue'
import type { GPaneEmits, GPaneExpose, GPaneProps } from './types'

defineOptions({
  name: 'GPane',
})

/**
 * @deprecated 请直接使用 `@/ui` 的 `RsSplitPane`。本组件仅作过渡兼容，后续将删除。
 */

const props = withDefaults(defineProps<GPaneProps>(), {
  direction: 'vertical',
  min: 0,
  max: 1,
  disabled: false,
  noResize: false,
  resizeTriggerSize: 2,
})

const emit = defineEmits<GPaneEmits>()

/** 面板二可见性（兼容旧 API） */
const pane2Visible = ref(true)

/** 受控 / 内部尺寸（0~1 分数，兼容旧 v-model:size） */
const currentSize = ref<number | string | undefined>(props.size ?? props.defaultSize)

watch(
  () => props.size,
  (value) => {
    if (value !== undefined) currentSize.value = value
  },
)

/**
 * 将 GPane 尺寸（0~1 / 百分比 / px）转为 RsSplitPane 百分比。
 * px 无法精确换算时回退到 fallback。
 */
function toPercent(value: number | string | undefined, fallback = 50): number {
  if (value == null) return fallback
  if (typeof value === 'number') {
    if (!Number.isFinite(value)) return fallback
    return Math.max(0, Math.min(100, value <= 1 ? value * 100 : value))
  }
  const trimmed = value.trim()
  if (trimmed.endsWith('%')) {
    const percent = Number.parseFloat(trimmed)
    return Number.isFinite(percent) ? Math.max(0, Math.min(100, percent)) : fallback
  }
  if (trimmed.endsWith('px')) {
    return fallback
  }
  const num = Number.parseFloat(trimmed)
  if (!Number.isFinite(num)) return fallback
  return Math.max(0, Math.min(100, num <= 1 ? num * 100 : num))
}

const pane1Percent = computed(() => {
  if (!pane2Visible.value) return 100
  return toPercent(currentSize.value ?? props.defaultSize, 30)
})

const pane1MinPercent = computed(() => toPercent(props.min, 0))
const pane1MaxPercent = computed(() => toPercent(props.max, 100))

const splitPanes = computed<RsSplitPaneItem[]>(() => {
  const size1 = pane1Percent.value
  return [
    {
      key: 'pane1',
      size: size1,
      min: pane1MinPercent.value,
      max: pane1MaxPercent.value,
    },
    {
      key: 'pane2',
      size: Math.max(0, 100 - size1),
      collapsible: true,
      collapsedSize: 0,
    },
  ]
})

const splitSizes = ref<number[]>([pane1Percent.value, Math.max(0, 100 - pane1Percent.value)])

watch(pane1Percent, (size1) => {
  splitSizes.value = [size1, Math.max(0, 100 - size1)]
})

watch(pane2Visible, (visible) => {
  if (!visible) {
    splitSizes.value = [100, 0]
    currentSize.value = 1
  } else if (props.size === undefined) {
    currentSize.value = props.defaultSize
  }
})

const handleResize = (sizes: number[]) => {
  if (!pane2Visible.value) return
  const fraction = (sizes[0] ?? 50) / 100
  currentSize.value = fraction
  emit('update:size', fraction)
}

const handleResizeEnd = (sizes: number[]) => {
  handleResize(sizes)
  emit('drag-end', new Event('resize-end'))
}

/** noResize + 指定 defaultSize 时按比例分配；未指定则 pane1 自适应内容 */
const computedPane1Style = computed(() => {
  const baseStyle: CSSProperties =
    typeof props.pane1Style === 'string' ? {} : props.pane1Style || {}
  if (!props.noResize || props.defaultSize === undefined) {
    return props.pane1Style as CSSProperties | string || {}
  }
  const size = pane1Percent.value / 100
  return { ...baseStyle, flex: `${size} ${size} 0` }
})

const computedPane2Style = computed(() => {
  const baseStyle: CSSProperties =
    typeof props.pane2Style === 'string' ? {} : props.pane2Style || {}
  if (!props.noResize || props.defaultSize === undefined) {
    return props.pane2Style as CSSProperties | string || {}
  }
  const remaining = 1 - pane1Percent.value / 100
  return { ...baseStyle, flex: `${remaining} ${remaining} 0` }
})

const setPane2Visible = (visible: boolean) => {
  pane2Visible.value = visible
}

const getPane2Visible = () => pane2Visible.value

const togglePane2Visible = () => {
  setPane2Visible(!pane2Visible.value)
}

const setSize = (size: number | string) => {
  currentSize.value = size
  emit('update:size', size)
}

const getSize = () => currentSize.value ?? props.defaultSize

defineExpose<GPaneExpose>({
  setPane2Visible,
  getPane2Visible,
  togglePane2Visible,
  setSize,
  getSize,
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

  .g-pane__split {
    width: 100%;
    height: 100%;
  }

  .g-pane__slot {
    width: 100%;
    height: 100%;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .g-pane__flex-container {
    width: 100%;
    height: 100%;
    display: flex;

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

  .g-pane__flex-pane {
    display: flex;
    flex-direction: column;
  }
}
</style>
