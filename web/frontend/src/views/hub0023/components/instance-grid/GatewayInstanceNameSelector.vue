<template>
  <RsInput
    v-bind="attrs"
    :model-value="localValue"
    placeholder="请输入实例名称或点击选择"
    clearable
    size="sm"
    label-position="top"
    class="instance-name-selector"
    addon-after-icon="ellipsis"
    addon-after-icon-label="选择实例"
    @update:model-value="handleInputChange"
    @addon-after-click="handleSelectClick"
  />

  <GatewayInstanceListModal
    v-model:visible="instanceSelectDialogVisible"
    v-model:model-value="localValue"
    title="选择网关实例"
    :width="1200"
    @select="onModalInstanceSelect"
  />
</template>

<script lang="ts" setup>
import { RsInput } from '@/ui'
import type { GatewayInstance } from '@/views/hub0020/types'
import { ref, useAttrs, watch } from 'vue'
import GatewayInstanceListModal from './GatewayInstanceListModal.vue'

defineOptions({
  name: 'GatewayInstanceNameSelector',
  inheritAttrs: false,
})

const attrs = useAttrs()

interface Props {
  /** 实例名称值 */
  modelValue?: string
  /** 网关实例 ID（与弹窗选择同步；手动输入时不保证有值） */
  gatewayInstanceId?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  gatewayInstanceId: '',
})

interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'update:gatewayInstanceId', value: string): void
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

/** 弹窗表格选中行时同步实例 ID */
function onModalInstanceSelect(instance: GatewayInstance) {
  emit('update:gatewayInstanceId', instance.gatewayInstanceId || '')
}

/**
 * 处理输入框值变化
 */
const handleInputChange = (value: string) => {
  localValue.value = value
  emit('update:gatewayInstanceId', '')
}

/**
 * 处理选择按钮点击
 */
const handleSelectClick = () => {
  instanceSelectDialogVisible.value = true
}
</script>

<style lang="scss" scoped>
.instance-name-selector {
  width: 100%;
}
</style>
