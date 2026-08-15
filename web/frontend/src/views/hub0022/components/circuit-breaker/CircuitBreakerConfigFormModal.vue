<template>
  <div class="circuit-breaker-config-form-modal" id="circuit-breaker-config-form-modal">
    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="dialogTitle"
      :width="800"
      :to="to || '#circuit-breaker-config-form-modal'"
      :form-fields="model.formFields"
      :initial-data="currentConfig || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="saving"
      @submit="handleFormSubmit"
      @update:visible="handleVisibleChange"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { useAppMessage } from '@/composables/useAppMessage'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { computed, ref, watch } from 'vue'
import { getCircuitBreakerConfig, saveCircuitBreakerConfig } from '../../api'
import type { CircuitBreakerConfig } from '../service/types'
import { useCircuitBreakerConfigModel } from './hooks/model'

defineOptions({
  name: 'CircuitBreakerConfigFormModal',
})

const props = withDefaults(
  defineProps<{
    visible: boolean
    targetServiceId?: string
    serviceName?: string
    to?: string
  }>(),
  {
    visible: false,
    targetServiceId: '',
    serviceName: '',
    to: undefined,
  },
)

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const message = useAppMessage()
const model = useCircuitBreakerConfigModel()
const formDialogVisible = ref(false)
const formDialogMode = ref<'create' | 'edit'>('create')
const currentConfig = ref<Partial<CircuitBreakerConfig> | null>(null)
const saving = ref(false)

const dialogTitle = computed(() => {
  const name = props.serviceName ? ` - ${props.serviceName}` : ''
  return formDialogMode.value === 'edit' ? `编辑熔断配置${name}` : `新增熔断配置${name}`
})

const defaultConfig = (serviceId: string): Partial<CircuitBreakerConfig> => ({
  targetServiceId: serviceId,
  errorRatePercent: 50,
  minimumRequests: 10,
  halfOpenMaxRequests: 3,
  openTimeoutSeconds: 60,
  slowCallThreshold: 60000,
  slowCallRatePercent: 50,
  windowSizeSeconds: 60,
})

const loadConfig = async () => {
  if (!props.targetServiceId) {
    return
  }
  const response = await getCircuitBreakerConfig(props.targetServiceId)
  if (!isApiSuccess(response)) {
    message.error(getApiMessage(response, '获取熔断配置失败'))
    return
  }
  const config = parseJsonData<CircuitBreakerConfig | null>(response, null)
  if (config && config.circuitBreakerConfigId) {
    currentConfig.value = config
    formDialogMode.value = 'edit'
    return
  }
  currentConfig.value = defaultConfig(props.targetServiceId)
  formDialogMode.value = 'create'
}

const handleFormSubmit = async (formData?: Record<string, any>) => {
  if (!formData || !props.targetServiceId) {
    return
  }
  saving.value = true
  try {
    const response = await saveCircuitBreakerConfig({
      ...formData,
      targetServiceId: props.targetServiceId,
      circuitBreakerConfigId: currentConfig.value?.circuitBreakerConfigId,
    })
    if (isApiSuccess(response)) {
      message.success(getApiMessage(response, '熔断配置保存成功'))
      emit('update:visible', false)
      return
    }
    message.error(getApiMessage(response, '熔断配置保存失败'))
  } catch {
    message.error('熔断配置保存失败')
  } finally {
    saving.value = false
  }
}

const handleVisibleChange = (value: boolean) => {
  formDialogVisible.value = value
  emit('update:visible', value)
  if (!value) {
    currentConfig.value = null
  }
}

watch(
  () => props.visible,
  async (visible) => {
    if (visible) {
      await loadConfig()
      formDialogVisible.value = true
      return
    }
    formDialogVisible.value = false
  },
)
</script>
