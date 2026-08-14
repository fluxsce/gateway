<template>
  <div class="instance-name-selector">
    <RsInput
      v-bind="attrs"
      v-model="localValue"
      placeholder="请输入实例名称或点击选择"
      clearable
      size="sm"
      label-position="top"
      class="instance-name-selector__input"
      addon-after-icon="ellipsis"
      addon-after-icon-label="选择实例"
      @addon-after-click="handleSelectClick"
    />

    <ServiceCenterInstanceListModal
      v-model:visible="instanceSelectDialogVisible"
      v-model:model-value="localValue"
      title="选择服务中心实例"
      :width="1200"
      @select="handleInstanceSelect"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsInput } from '@/ui'
import { ref, useAttrs, watch } from 'vue'
import type { ServiceCenterInstance } from '../types'
import ServiceCenterInstanceListModal from './ServiceCenterInstanceListModal.vue'

defineOptions({
  name: 'ServiceCenterInstanceNameSelector',
  inheritAttrs: false,
})

const attrs = useAttrs()

interface Props {
  /** 实例名称值 */
  modelValue?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
})

interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'select', instance: ServiceCenterInstance): void
}

const emit = defineEmits<Emits>()

const instanceSelectDialogVisible = ref(false)
const localValue = ref(props.modelValue)

watch(
  () => props.modelValue,
  (newVal) => {
    localValue.value = newVal
  },
)

watch(localValue, (newVal) => {
  emit('update:modelValue', newVal)
})

/**
 * 处理选择按钮点击
 */
const handleSelectClick = () => {
  instanceSelectDialogVisible.value = true
}

/**
 * 处理实例选择
 */
const handleInstanceSelect = (instance: ServiceCenterInstance) => {
  if (!instance) return
  localValue.value = instance.instanceName
  emit('select', instance)
}
</script>

<style lang="scss" scoped>
.instance-name-selector {
  display: block;
  width: 100%;
  min-width: 0;
}

.instance-name-selector__input {
  width: 100%;
}
</style>
