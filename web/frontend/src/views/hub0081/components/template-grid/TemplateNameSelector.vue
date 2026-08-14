<template>
  <div class="template-name-selector">
    <RsInput
      v-bind="attrs"
      :model-value="localValue"
      placeholder="请输入模板名称或点击选择"
      clearable
      size="sm"
      label-position="top"
      class="template-name-selector__input"
      addon-after-icon="ellipsis"
      addon-after-icon-label="选择模板"
      @update:model-value="handleInputChange"
      @addon-after-click="handleSelectClick"
    />

    <TemplateListModal
      v-model:visible="templateSelectDialogVisible"
      v-model:model-value="localValue"
      title="选择模板"
      :width="1200"
      :channel-type="channelType"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsInput } from '@/ui'
import { ref, useAttrs, watch } from 'vue'
import TemplateListModal from './TemplateListModal.vue'

defineOptions({
  name: 'TemplateNameSelector',
  inheritAttrs: false,
})

const attrs = useAttrs()

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

interface Emits {
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

const templateSelectDialogVisible = ref(false)
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
  templateSelectDialogVisible.value = true
}
</script>

<style lang="scss" scoped>
.template-name-selector {
  display: block;
  width: 100%;
  min-width: 0;
}

.template-name-selector__input {
  width: 100%;
}
</style>
