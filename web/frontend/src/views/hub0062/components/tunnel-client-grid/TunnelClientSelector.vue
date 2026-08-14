<template>
  <div class="tunnel-client-selector">
    <RsInput
      v-bind="attrs"
      v-model="localValue"
      placeholder="请输入客户端ID或点击选择"
      :disabled="disabled"
      clearable
      size="sm"
      label-position="top"
      class="tunnel-client-selector__input"
      addon-after-icon="ellipsis"
      addon-after-icon-label="选择客户端"
      @addon-after-click="handleSelectClick"
    />

    <TunnelClientListModal
      v-model:visible="clientSelectDialogVisible"
      @select="handleSelect"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsInput } from '@/ui'
import { ref, useAttrs, watch } from 'vue'
import type { TunnelClient } from '../../types'
import TunnelClientListModal from './TunnelClientListModal.vue'

defineOptions({
  name: 'TunnelClientSelector',
  inheritAttrs: false,
})

const attrs = useAttrs()

interface Props {
  /** 客户端ID值 */
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
  (e: 'select', client: TunnelClient): void
}

const emit = defineEmits<Emits>()

const clientSelectDialogVisible = ref(false)
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
  clientSelectDialogVisible.value = true
}

/**
 * 处理选择客户端
 */
const handleSelect = (client: TunnelClient) => {
  if (!client) return
  localValue.value = client.tunnelClientId
  emit('select', client)
}
</script>

<style lang="scss" scoped>
.tunnel-client-selector {
  display: block;
  width: 100%;
  min-width: 0;
}

.tunnel-client-selector__input {
  width: 100%;
}
</style>
