<template>
  <n-input-group>
    <n-input
      :value="modelValue"
      placeholder="请输入命名空间ID或点击选择"
      :disabled="disabled"
      clearable
      size="small"
      @update:value="handleInputChange"
    />
    <n-button type="primary" size="small" :disabled="disabled" @click="handleSelectClick">
      <template #icon>
        <n-icon><EllipsisHorizontalOutline /></n-icon>
      </template>
    </n-button>
  </n-input-group>

  <!-- 命名空间选择对话框 -->
  <NamespaceListModal
    v-model:visible="namespaceSelectDialogVisible"
    v-model:model-value="localValue"
    title="选择命名空间"
    :width="1200"
    @select="handleNamespaceSelect"
  />
</template>

<script lang="ts" setup>
import { EllipsisHorizontalOutline } from '@vicons/ionicons5'
import { NButton, NIcon, NInput, NInputGroup } from 'naive-ui'
import { ref, watch } from 'vue'
import type { Namespace } from '../types'
import NamespaceListModal from './NamespaceListModal.vue'

// 定义组件名称
defineOptions({
  name: 'NamespaceNameSelector'
})

// ============= Props =============

interface Props {
  /** 命名空间ID值 */
  modelValue?: string
  /** 是否禁用 */
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  disabled: false,
})

// ============= Emits =============

interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'select', namespace: Namespace): void
}

const emit = defineEmits<Emits>()

// ============= 弹窗状态 =============

const namespaceSelectDialogVisible = ref(false)
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
  namespaceSelectDialogVisible.value = true
}

/**
 * 处理命名空间选择
 */
const handleNamespaceSelect = (namespace: Namespace) => {
  if (namespace) {
    localValue.value = namespace.namespaceId
    emit('select', namespace)
  }
}
</script>

<style lang="scss" scoped>
.n-input-group {
  width: 100%;
}
</style>

