<template>
  <RsInput
    v-bind="attrs"
    v-model="localValue"
    placeholder="请输入服务名称或点击选择"
    clearable
    size="sm"
    label-position="top"
    class="service-name-selector"
    addon-after-icon="ellipsis"
    addon-after-icon-label="选择服务"
    @addon-after-click="handleSelectClick"
  />

  <ServiceListModal
    v-model:visible="serviceSelectDialogVisible"
    v-model:model-value="localValue"
    title="选择服务"
    :width="1200"
    :gateway-instance-id="gatewayInstanceId"
  />
</template>

<script lang="ts" setup>
import { RsInput } from '@/ui'
import { ref, useAttrs, watch } from 'vue'
import ServiceListModal from './ServiceListModal.vue'

defineOptions({
  name: 'ServiceNameSelector',
  inheritAttrs: false,
})

const attrs = useAttrs()

interface Props {
  /** 服务名称值 */
  modelValue?: string
  /** 网关实例ID（可选，用于过滤服务） */
  gatewayInstanceId?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  gatewayInstanceId: undefined,
})

interface Emits {
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

const serviceSelectDialogVisible = ref(false)
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
  serviceSelectDialogVisible.value = true
}
</script>

<style lang="scss" scoped>
.service-name-selector {
  width: 100%;
}
</style>
