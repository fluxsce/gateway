<template>
  <n-input-group>
    <n-input
      :value="modelValue"
      placeholder="请输入告警渠道名称或点击选择"
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

  <!-- 告警渠道选择对话框 -->
  <AlertChannelListModal
    v-model:visible="channelSelectDialogVisible"
    v-model:model-value="localValue"
    title="选择告警渠道"
    :width="1200"
  />
</template>

<script lang="ts" setup>
import { EllipsisHorizontalOutline } from '@vicons/ionicons5'
import { NButton, NIcon, NInput, NInputGroup } from 'naive-ui'
import { ref, watch } from 'vue'
import AlertChannelListModal from './AlertChannelListModal.vue'

// 定义组件名称
defineOptions({
  name: 'AlertChannelNameSelector'
})

// ============= Props =============

interface Props {
  /** 告警渠道名称值 */
  modelValue?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
})

// ============= Emits =============

interface Emits {
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

// ============= 弹窗状态 =============

const channelSelectDialogVisible = ref(false)
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
  channelSelectDialogVisible.value = true
}
</script>

<style lang="scss" scoped>
.n-input-group {
  width: 100%;
}
</style>

