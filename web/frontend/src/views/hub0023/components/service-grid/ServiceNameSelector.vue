<template>
  <n-input-group>
    <n-input
      :value="modelValue"
      placeholder="请输入服务名称或点击选择"
      clearable
      size="small"
      @update:value="handleInputChange"
    />
    <n-button type="primary" size="small" @click="handleSelectClick">
      <template #icon>
        <n-icon><EllipsisHorizontalOutline /></n-icon>
      </template>
    </n-button>
  </n-input-group>

  <!-- 服务选择对话框 -->
  <ServiceListModal
    v-model:visible="serviceSelectDialogVisible"
    v-model:model-value="localValue"
    title="选择服务"
    :width="1200"
    :gateway-instance-id="gatewayInstanceId"
  />
</template>

<script lang="ts" setup>
import { EllipsisHorizontalOutline } from '@vicons/ionicons5'
import { NButton, NIcon, NInput, NInputGroup } from 'naive-ui'
import { ref, watch } from 'vue'
import ServiceListModal from './ServiceListModal.vue'

// 定义组件名称
defineOptions({
  name: 'ServiceNameSelector'
})

// ============= Props =============

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

// ============= Emits =============

interface Emits {
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

// ============= 弹窗状态 =============

const serviceSelectDialogVisible = ref(false)
const localValue = ref(props.modelValue)

// 监听 props.modelValue 变化，同步到本地状态
watch(() => props.modelValue, (newVal) => {
  localValue.value = newVal
})

// 监听本地值变化，同步到父组件
watch(localValue, (newVal) => {
  emit('update:modelValue', newVal)
})

// ============= 事件处理 =============

/**
 * 处理输入框值变化
 */
const handleInputChange = (value: string) => {
  localValue.value = value
}

/**
 * 处理选择按钮点击
 */
const handleSelectClick = () => {
  serviceSelectDialogVisible.value = true
}
</script>

<style lang="scss" scoped>
.n-input-group {
  width: 100%;
}
</style>

