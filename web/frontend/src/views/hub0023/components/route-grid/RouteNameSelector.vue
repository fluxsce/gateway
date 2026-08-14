<template>
  <RsInput
    v-bind="attrs"
    v-model="localValue"
    placeholder="请输入路由名称或点击选择"
    clearable
    size="sm"
    label-position="top"
    class="route-name-selector"
    addon-after-icon="ellipsis"
    addon-after-icon-label="选择路由"
    @addon-after-click="handleSelectClick"
  />

  <RouteListModal
    v-model:visible="routeSelectDialogVisible"
    v-model:model-value="localValue"
    title="选择路由"
    :width="1200"
    :gateway-instance-id="gatewayInstanceId"
  />
</template>

<script lang="ts" setup>
import { RsInput } from '@/ui'
import { ref, useAttrs, watch } from 'vue'
import RouteListModal from './RouteListModal.vue'

defineOptions({
  name: 'RouteNameSelector',
  inheritAttrs: false,
})

const attrs = useAttrs()

interface Props {
  /** 路由名称值 */
  modelValue?: string
  /** 网关实例ID（可选，用于过滤路由） */
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

const routeSelectDialogVisible = ref(false)
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
  routeSelectDialogVisible.value = true
}
</script>

<style lang="scss" scoped>
.route-name-selector {
  width: 100%;
}
</style>
