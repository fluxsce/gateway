<template>
  <GModal
    :visible="modalVisible"
    :title="'预警测试 - ' + (currentConfig?.channelName || '')"
    :width="800"
    :to="props.to"
    :show-footer="true"
    :show-cancel="true"
    :show-confirm="true"
    cancel-text="取消"
    confirm-text="发送测试"
    :confirm-loading="loading"
    @update:visible="handleUpdateVisible"
    @confirm="handleConfirm"
    @cancel="handleCancel"
  >
    <div class="alert-test-modal" id="alert-test-modal">
      <n-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-placement="left"
        label-width="100"
        require-mark-placement="right-hanging"
      >
        <n-form-item label="渠道名称" path="channelName">
          <n-input
            :value="currentConfig?.channelName || ''"
            disabled
            placeholder="渠道名称"
          />
        </n-form-item>

        <n-form-item label="渠道类型" path="channelType">
          <n-input
            :value="getChannelTypeLabel(currentConfig?.channelType)"
            disabled
            placeholder="渠道类型"
          />
        </n-form-item>

        <n-form-item label="测试主题" path="title">
          <n-input
            v-model:value="formData.title"
            placeholder="请输入测试消息主题"
            clearable
          />
        </n-form-item>

        <n-form-item label="测试内容" path="content">
          <div class="alert-test-modal__codemirror-wrapper">
            <GCodeMirror
              v-model="formData.content"
              language="plaintext"
              :height="300"
              :min-height="200"
              :max-height="500"
              :line-numbers="true"
              :line-wrapping="true"
              placeholder="请输入测试消息内容"
            />
          </div>
        </n-form-item>
      </n-form>
    </div>
  </GModal>
</template>

<script setup lang="ts">
import { GCodeMirror } from '@/components'
import { GModal } from '@/components/gmodal'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { NForm, NFormItem, NInput, useMessage } from 'naive-ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { testAlertChannel } from '../api'
import type { AlertConfig } from '../types'
import { CHANNEL_TYPE_OPTIONS } from '../types'

// 定义组件名称
defineOptions({
  name: 'AlertTestModal'
})

// ============= Props =============

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

// ============= Emits =============

export interface AlertTestModalEmits {
  /** v-model:visible 更新事件 */
  (event: 'update:visible', value: boolean): void
  /** 关闭事件 */
  (event: 'close'): void
  /** 测试成功事件 */
  (event: 'success', result: any): void
}

const emit = defineEmits<AlertTestModalEmits>()

// ============= Refs =============

const formRef = ref()
const message = useMessage()

// ============= 状态 =============

const modalVisible = ref(props.visible)
const currentConfig = ref<AlertConfig | null>(props.config)
const loading = ref(false)

// 表单数据
const formData = ref({
  title: '告警渠道测试',
  content: `这是一条测试告警消息，用于验证告警渠道配置是否正确。

测试时间：${new Date().toLocaleString('zh-CN')}

您可以在此编辑测试消息内容，然后点击"发送测试"按钮进行测试。`,
})

// 表单验证规则
const formRules = {
  title: [
    { required: true, message: '请输入测试消息主题', trigger: ['blur', 'input'] },
    { max: 200, message: '主题长度不能超过200个字符', trigger: ['blur', 'input'] },
  ],
  content: [
    { required: true, message: '请输入测试消息内容', trigger: ['blur', 'change'] },
  ],
}

// ============= 监听器 =============

// 监听 props.visible 变化
const stopVisibleWatch = watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
  if (newVal) {
    // 打开时重置表单
    resetForm()
  }
})

// 监听 props.config 变化
const stopConfigWatch = watch(() => props.config, (newVal) => {
  currentConfig.value = newVal
  if (newVal) {
    // 更新默认内容
    formData.value.content = `这是一条测试告警消息，用于验证告警渠道配置是否正确。

测试时间：${new Date().toLocaleString('zh-CN')}
渠道名称：${newVal.channelName}
渠道类型：${getChannelTypeLabel(newVal.channelType)}

您可以在此编辑测试消息内容，然后点击"发送测试"按钮进行测试。`
  }
})

// ============= 资源清理 =============

onBeforeUnmount(() => {
  stopVisibleWatch()
  stopConfigWatch()
})

// ============= 工具函数 =============

/**
 * 获取渠道类型标签
 */
const getChannelTypeLabel = (channelType?: string) => {
  if (!channelType) return ''
  const option = CHANNEL_TYPE_OPTIONS.find(opt => opt.value === channelType)
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
  formRef.value?.restoreValidation()
}

// ============= 事件处理 =============

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

  // 验证表单
  try {
    await formRef.value?.validate()
  } catch (error) {
    return
  }

  try {
    loading.value = true

    const response = await testAlertChannel(
      currentConfig.value.channelName,
      formData.value.title,
      formData.value.content
    )

    if (isApiSuccess(response)) {
      const result = parseJsonData<any>(response)
      const successMsg = result?.message || '测试消息发送成功'
      message.success(successMsg)
      
      // 触发成功事件
      emit('success', result)
      
      // 不自动关闭弹窗，让用户手动关闭
    } else {
      message.error(getApiMessage(response, '测试消息发送失败'))
    }
  } catch (error: any) {
    console.error('测试告警渠道失败:', error)
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
  padding: var(--g-space-md, 16px);

  &__codemirror-wrapper {
    width: 100%;
    border: 1px solid var(--g-border-primary, #e0e0e6);
    border-radius: var(--g-border-radius, 4px);
    overflow: hidden;
  }
}

:deep(.n-form-item-label) {
  font-weight: 500;
}
</style>

