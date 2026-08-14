<template>
  <div class="alert-channel-name-selector">
    <RsInput
      v-bind="attrs"
      :model-value="localValue"
      placeholder="请输入告警渠道名称或点击选择"
      clearable
      size="sm"
      label-position="top"
      class="alert-channel-name-selector__input"
      addon-after-icon="ellipsis"
      addon-after-icon-label="选择告警渠道"
      @update:model-value="handleInputChange"
      @addon-after-click="handleSelectClick"
    />

    <AlertChannelListModal
      v-model:visible="channelSelectDialogVisible"
      v-model:model-value="localValue"
      title="选择告警渠道"
      :width="1200"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsInput } from '@/ui'
import { ref, useAttrs, watch } from 'vue'
import AlertChannelListModal from './AlertChannelListModal.vue'

defineOptions({
  name: 'AlertChannelNameSelector',
  inheritAttrs: false,
})

const attrs = useAttrs()

interface Props {
  /** 告警渠道名称值 */
  modelValue?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
})

interface Emits {
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

const channelSelectDialogVisible = ref(false)
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
.alert-channel-name-selector {
  display: block;
  width: 100%;
  min-width: 0;
}

.alert-channel-name-selector__input {
  width: 100%;
}
</style>
