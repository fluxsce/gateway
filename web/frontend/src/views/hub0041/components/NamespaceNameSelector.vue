<template>
  <div class="namespace-name-selector">
    <RsInput
      v-bind="attrs"
      v-model="localValue"
      placeholder="请输入命名空间ID或点击选择"
      :disabled="disabled"
      clearable
      size="sm"
      label-position="top"
      class="namespace-name-selector__input"
      addon-after-icon="ellipsis"
      addon-after-icon-label="选择命名空间"
      @addon-after-click="handleSelectClick"
    />

    <NamespaceListModal
      v-model:visible="namespaceSelectDialogVisible"
      v-model:model-value="localValue"
      title="选择命名空间"
      :width="1200"
      @select="handleNamespaceSelect"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsInput } from '@/ui'
import { ref, useAttrs, watch } from 'vue'
import type { Namespace } from '../types'
import NamespaceListModal from './NamespaceListModal.vue'

defineOptions({
  name: 'NamespaceNameSelector',
  inheritAttrs: false,
})

const attrs = useAttrs()

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

interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'select', namespace: Namespace): void
}

const emit = defineEmits<Emits>()

const namespaceSelectDialogVisible = ref(false)
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
  if (props.disabled) return
  namespaceSelectDialogVisible.value = true
}

/**
 * 处理命名空间选择
 */
const handleNamespaceSelect = (namespace: Namespace) => {
  if (!namespace) return
  localValue.value = namespace.namespaceId
  emit('select', namespace)
}
</script>

<style lang="scss" scoped>
.namespace-name-selector {
  display: block;
  width: 100%;
  min-width: 0;
}

.namespace-name-selector__input {
  width: 100%;
}
</style>
