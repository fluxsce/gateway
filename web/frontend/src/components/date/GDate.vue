<template>
  <n-date-picker
    v-bind="$attrs"
    :value="internalValue"
    @update:value="handleValueChange"
  />
</template>

<script setup lang="ts">
import { NDatePicker } from 'naive-ui';
import { computed, watch } from 'vue';
import type { GDateEmits, GDateProps } from './types';

const props = withDefaults(defineProps<GDateProps>(), {
  value: null,
  outputFormat: 'iso'
})

const emit = defineEmits<GDateEmits>()

/**
 * 将 ISO 字符串或时间戳转换为 NDatePicker 需要的格式（时间戳）
 * 对于日期范围类型，返回 [number, number] | null
 * 对于单个日期类型，返回 number | null
 */
const parseDateValue = (
  val: string | number | number[] | null | undefined
): number | [number, number] | null => {
  if (val === null || val === undefined) {
    return null
  }

  // 如果是数组（日期范围），处理为 [number, number] 元组
  if (Array.isArray(val)) {
    const timestamps = val
      .map((item) => {
        if (typeof item === 'string') {
          const timestamp = new Date(item).getTime()
          // 检查是否为有效日期（避免无效日期如 "0001-01-01T00:00:00Z"）
          return isNaN(timestamp) || timestamp < 0 ? null : timestamp
        }
        return typeof item === 'number' ? item : null
      })
      .filter((item): item is number => item !== null)

    // 日期范围需要正好 2 个元素
    if (timestamps.length === 2) {
      return [timestamps[0], timestamps[1]] as [number, number]
    }
    // 如果只有 1 个元素，返回单个数字（某些情况下可能是单个日期）
    if (timestamps.length === 1) {
      return timestamps[0]
    }
    return null
  }

  // 如果是字符串，转换为时间戳
  if (typeof val === 'string') {
    // 处理空字符串
    if (val.trim() === '') {
      return null
    }
    const timestamp = new Date(val).getTime()
    // 检查是否为有效日期（避免无效日期如 "0001-01-01T00:00:00Z"）
    if (isNaN(timestamp) || timestamp < 0) {
      return null
    }
    return timestamp
  }

  // 如果是数字，直接返回（已经是时间戳）
  if (typeof val === 'number') {
    return val
  }

  return null
}

/**
 * 内部值（转换为时间戳格式，供 NDatePicker 使用）
 */
const internalValue = computed(() => {
  return parseDateValue(props.value) as number | [number, number] | null
})

/**
 * 将时间戳转换为 ISO 字符串
 */
const timestampToISO = (timestamp: number): string => {
  const date = new Date(timestamp)
  // 检查是否为有效日期
  if (isNaN(date.getTime()) || date.getTime() < 0) {
    return ''
  }
  return date.toISOString()
}

/**
 * 处理值变化，根据 outputFormat 转换格式后透传给父组件
 */
const handleValueChange = (value: number | [number, number] | null) => {
  if (value === null) {
    emit('update:value', null)
    return
  }

  if (props.outputFormat === 'iso') {
    // 输出 ISO 字符串格式
    if (Array.isArray(value)) {
      // 日期范围：转换为 ISO 字符串数组
      emit('update:value', [
        timestampToISO(value[0]),
        timestampToISO(value[1])
      ] as string[])
    } else {
      // 单个日期：转换为 ISO 字符串
      emit('update:value', timestampToISO(value))
    }
  } else {
    // 输出时间戳格式（保持原样）
    emit('update:value', value)
  }
}

// 当外部 value 变化时，确保内部值同步
watch(
  () => props.value,
  () => {
    // 这里不需要额外处理，computed 会自动更新
  },
  { deep: true }
)
</script>

