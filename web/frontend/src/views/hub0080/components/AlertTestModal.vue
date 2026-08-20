<template>
  <RsDialog
    :open="modalVisible"
    :title="'预警测试 - ' + (currentConfig?.channelName || '')"
    layout="window"
    :width="800"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :show-overlay="true"
    :close-on-overlay-click="false"
    @update:open="handleUpdateVisible"
  >
    <template #body>
      <div class="alert-test-modal" id="alert-test-modal">
        <RsForm
          ref="formRef"
          :model="formData"
          :rules="formRules"
          label-position="left"
          label-width="6.25rem"
          gap="md"
        >
          <RsInput
            :model-value="currentConfig?.channelName || ''"
            name="channelName"
            label="渠道名称"
            disabled
            placeholder="渠道名称"
          />

          <RsInput
            :model-value="getChannelTypeLabel(currentConfig?.channelType)"
            name="channelType"
            label="渠道类型"
            disabled
            placeholder="渠道类型"
          />

          <RsInput
            v-model="formData.title"
            name="title"
            label="测试主题"
            placeholder="请输入测试消息主题"
            clearable
          />

          <RsFormItem v-model="formData.content" name="content" label="测试内容" required>
            <textarea
              v-model="formData.content"
              class="alert-test-modal__textarea"
              rows="12"
              placeholder="请输入测试消息内容"
            />
          </RsFormItem>
        </RsForm>
      </div>
    </template>

    <template #footer>
      <div class="alert-test-modal__footer">
        <RsButton variant="secondary" @click="handleCancel">取消</RsButton>
        <RsButton variant="primary" :loading="loading" @click="handleConfirm">发送测试</RsButton>
      </div>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import { useAppMessage } from '@/composables/useAppMessage'
import {
  RsButton,
  RsDialog,
  RsForm,
  RsFormItem,
  RsInput,
  type RsFormRules,
  type RsFormValidationResult,
} from '@/ui'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { onBeforeUnmount, ref, watch } from 'vue'
import { testAlertChannel } from '../api'
import type { AlertConfig } from '../types'
import { CHANNEL_TYPE_OPTIONS } from '../types'

defineOptions({
  name: 'AlertTestModal',
})

export interface AlertTestModalProps {
  /** 是否显示模态框 */
  visible?: boolean
  /** 当前配置 */
  config?: AlertConfig | null
  /** 挂载目标 */
  to?: string | HTMLElement | false
}

const props = withDefaults(defineProps<AlertTestModalProps>(), {
  visible: false,
  config: null,
  to: undefined,
})

export interface AlertTestModalEmits {
  /** v-model:visible 更新事件 */
  (event: 'update:visible', value: boolean): void
  /** 关闭事件 */
  (event: 'close'): void
  /** 测试成功事件 */
  (event: 'success', result: any): void
}

const emit = defineEmits<AlertTestModalEmits>()

/** RsForm 暴露的校验方法 */
interface RsFormExpose {
  validate: () => Promise<RsFormValidationResult>
}

const formRef = ref<RsFormExpose | null>(null)
const message = useAppMessage()

const modalVisible = ref(props.visible)
const currentConfig = ref<AlertConfig | null>(props.config)
const loading = ref(false)

const formData = ref({
  title: '告警渠道测试',
  content: `这是一条测试告警消息，用于验证告警渠道配置是否正确。

测试时间：${new Date().toLocaleString('zh-CN')}

您可以在此编辑测试消息内容，然后点击"发送测试"按钮进行测试。`,
})

const formRules: RsFormRules = {
  title: [
    { required: true, message: '请输入测试消息主题' },
    { max: 200, message: '主题长度不能超过200个字符' },
  ],
  content: [
    { required: true, message: '请输入测试消息内容' },
  ],
}

const stopVisibleWatch = watch(
  () => props.visible,
  (newVal) => {
    modalVisible.value = newVal
    if (newVal) {
      resetForm()
    }
  },
)

const stopConfigWatch = watch(
  () => props.config,
  (newVal) => {
    currentConfig.value = newVal
    if (newVal) {
      formData.value.content = `这是一条测试告警消息，用于验证告警渠道配置是否正确。

测试时间：${new Date().toLocaleString('zh-CN')}
渠道名称：${newVal.channelName}
渠道类型：${getChannelTypeLabel(newVal.channelType)}

您可以在此编辑测试消息内容，然后点击"发送测试"按钮进行测试。`
    }
  },
)

onBeforeUnmount(() => {
  stopVisibleWatch()
  stopConfigWatch()
})

/**
 * 获取渠道类型标签
 */
const getChannelTypeLabel = (channelType?: string) => {
  if (!channelType) return ''
  const option = CHANNEL_TYPE_OPTIONS.find((opt) => opt.value === channelType)
  return option?.label || channelType
}

/**
 * 重置表单
 */
const resetForm = () => {
  formData.value = {
    title: '告警渠道测试',
    content: `这是一条测试告警消息，用于验证告警渠道配置是否正确。

测试时间：${new Date().toLocaleString('zh-CN')}
${currentConfig.value ? `渠道名称：${currentConfig.value.channelName}\n渠道类型：${getChannelTypeLabel(currentConfig.value.channelType)}\n` : ''}
您可以在此编辑测试消息内容，然后点击"发送测试"按钮进行测试。`,
  }
}

/**
 * 处理模态框可见性变化
 */
const handleUpdateVisible = (value: boolean) => {
  modalVisible.value = value
  emit('update:visible', value)
  if (!value) {
    emit('close')
  }
}

/**
 * 处理确认（发送测试）
 */
const handleConfirm = async () => {
  if (!currentConfig.value?.channelName) {
    message.warning('渠道配置不存在')
    return
  }

  const result = await formRef.value?.validate()
  if (!result?.valid) {
    return
  }

  try {
    loading.value = true

    const response = await testAlertChannel(
      currentConfig.value.channelName,
      formData.value.title,
      formData.value.content,
    )

    if (isApiSuccess(response)) {
      const testResult = parseJsonData<any>(response)
      const successMsg = testResult?.message || '测试消息发送成功'
      message.success(successMsg)
      emit('success', testResult)
    } else {
      message.error(getApiMessage(response, '测试消息发送失败'))
    }
  } catch (error: any) {
    message.error(error.message || '测试告警渠道失败')
  } finally {
    loading.value = false
  }
}

/**
 * 处理取消
 */
const handleCancel = () => {
  handleUpdateVisible(false)
}
</script>

<style scoped lang="scss">
.alert-test-modal {
  padding: var(--rs-space-md, 16px);
}

.alert-test-modal__textarea {
  box-sizing: border-box;
  width: 100%;
  min-height: 12rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--rs-input-border, var(--rs-border, #d0d5dd));
  border-radius: var(--rs-radius-sm, 6px);
  background: var(--rs-input-bg, var(--rs-surface, #fff));
  color: var(--rs-text, inherit);
  font: inherit;
  resize: vertical;
}

.alert-test-modal__footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
}
</style>
