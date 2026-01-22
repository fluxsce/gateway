<template>
  <n-input-group>
    <n-input
      :value="modelValue"
      placeholder="请输入模板名称或点击选择"
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

  <!-- 模板选择对话框 -->
  <TemplateListModal
    v-model:visible="templateSelectDialogVisible"
    v-model:model-value="localValue"
    title="选择模板"
    :width="1200"
  />
</template>

<script lang="ts" setup>
import { EllipsisHorizontalOutline } from '@vicons/ionicons5'
import { NButton, NIcon, NInput, NInputGroup } from 'naive-ui'
import { ref, watch } from 'vue'
import TemplateListModal from './TemplateListModal.vue'

// 定义组件名称
defineOptions({
  name: 'TemplateNameSelector'
})

// ============= Props =============

interface Props {
  /** 模板名称值 */
  modelValue?: string
  /** 渠道类型（可选，用于过滤模板） */
  channelType?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  channelType: undefined,
})

// ============= Emits =============

interface Emits {
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

// ============= 弹窗状态 =============

const templateSelectDialogVisible = ref(false)
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
  templateSelectDialogVisible.value = true
}
</script>

<style lang="scss" scoped>
.n-input-group {
  width: 100%;
}
</style>

