<template>
  <GdataFormModal
    v-model:visible="visible"
    mode="create"
    title="配置回滚"
    :width="500"
    :mask-closable="false"
    :form-fields="rollbackFormFields"
    :initial-data="initialFormData"
    :confirm-loading="submitting"
    confirm-text="确定"
    @submit="handleSubmit"
    @cancel="handleCancel"
  />
</template>

<script lang="ts" setup>
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import type { DataFormField } from '@/components/form/data/types'
import { computed } from 'vue'
import type { ConfigHistory, RollbackRequest } from '../../types'

// 定义组件名称
defineOptions({
  name: 'RollbackDialog'
})

// ============= Props & Emits =============
interface Props {
  /** 是否显示对话框 */
  visible?: boolean
  /** 当前要回滚的历史记录 */
  history?: ConfigHistory | null
  /** 是否正在提交 */
  submitting?: boolean
}

interface Emits {
  /** 更新显示状态 */
  (e: 'update:visible', visible: boolean): void
  /** 确认回滚 */
  (e: 'confirm', data: RollbackRequest): void
  /** 取消 */
  (e: 'cancel'): void
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  history: null,
  submitting: false,
})

const emit = defineEmits<Emits>()

// ============= 计算属性 =============
const visible = computed({
  get: () => props.visible,
  set: (val: boolean) => emit('update:visible', val)
})

const currentHistory = computed(() => props.history)

// 表单字段配置
const rollbackFormFields = computed<DataFormField[]>(() => [
  {
    field: 'version',
    label: '目标版本',
    type: 'input',
    span: 24,
    disabled: true,
    defaultValue: String(currentHistory.value?.newVersion || currentHistory.value?.configVersion || '-'),
  },
  {
    field: 'changeReason',
    label: '变更原因',
    type: 'textarea',
    placeholder: '请输入变更原因（可选）',
    span: 24,
    props: {
      rows: 3,
    },
  },
])

// 初始表单数据（包含版本信息，版本号转换为字符串）
const initialFormData = computed(() => ({
  version: String(currentHistory.value?.newVersion || currentHistory.value?.configVersion || '-'),
  changeReason: '',
}))

// ============= 方法 =============
/**
 * 处理表单提交
 */
const handleSubmit = (formData?: Record<string, any>) => {
  if (!currentHistory.value) {
    return
  }

  const rollbackData: RollbackRequest = {
    configHistoryId: String(currentHistory.value.configHistoryId),
    changeReason: formData?.changeReason || '',
  }

  emit('confirm', rollbackData)
}

/**
 * 取消
 */
const handleCancel = () => {
  emit('cancel')
  visible.value = false
}
</script>

<style scoped>
/* 组件样式 */
</style>
