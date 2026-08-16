<template>
  <div class="static-host-config-form-modal" id="static-host-config-form-modal">
    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="dialogTitle"
      :width="800"
      :to="to || '#static-host-config-form-modal'"
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
import { getStaticHostConfig, saveStaticHostConfig } from '../../api'
import { useStaticHostConfigModel, type StaticHostConfig } from './hooks/model'

defineOptions({
  name: 'StaticHostConfigFormModal',
})

const props = withDefaults(
  defineProps<{
    visible: boolean
    routeConfigId?: string
    to?: string
  }>(),
  {
    visible: false,
    routeConfigId: '',
    to: undefined,
  },
)

const emit = defineEmits<{
  'update:visible': [value: boolean]
  saved: []
}>()

const message = useAppMessage()
const model = useStaticHostConfigModel()
const formDialogVisible = ref(false)
const formDialogMode = ref<'create' | 'edit'>('create')
const currentConfig = ref<Partial<StaticHostConfig> | null>(null)
const saving = ref(false)

const dialogTitle = computed(() =>
  formDialogMode.value === 'edit' ? '编辑静态资源' : '新增静态资源',
)

const defaultConfig = (routeId: string): Partial<StaticHostConfig> => ({
  routeConfigId: routeId,
  rootDirectory: '',
  stripRoutePrefix: 'Y',
  spaFallback: 'N',
  indexFiles: 'index.html',
  cacheControlMaxAge: 3600,
  allowedExtensions: '',
  maxFileSizeBytes: 0,
  followSymlinks: 'N',
  enablePrecompress: 'Y',
  errorPage404: '',
  errorPage403: '',
})

function toFormIndexFiles(raw?: string): string {
  if (!raw) {
    return 'index.html'
  }
  const text = raw.trim()
  if (!text) {
    return 'index.html'
  }
  try {
    const parsed = JSON.parse(text) as unknown
    if (Array.isArray(parsed)) {
      return parsed.filter((item) => typeof item === 'string' && item.trim()).join(',')
    }
  } catch {
    // 兼容逗号分隔原文
  }
  return text
}

function toFormAllowedExtensions(raw?: string): string {
  if (!raw) {
    return ''
  }
  const text = raw.trim()
  if (!text) {
    return ''
  }
  try {
    const parsed = JSON.parse(text) as unknown
    if (Array.isArray(parsed)) {
      return parsed.filter((item) => typeof item === 'string' && item.trim()).join(',')
    }
  } catch {
    // 兼容逗号分隔原文
  }
  return text
}

const loadConfig = async () => {
  if (!props.routeConfigId) {
    return
  }
  const response = await getStaticHostConfig(props.routeConfigId)
  if (!isApiSuccess(response)) {
    message.error(getApiMessage(response, '获取静态资源配置失败'))
    return
  }
  const config = parseJsonData<StaticHostConfig | null>(response, null)
  if (config && config.staticHostConfigId) {
    currentConfig.value = {
      ...config,
      indexFiles: toFormIndexFiles(config.indexFiles),
      rewriteRules: config.rewriteRules || '',
      allowedExtensions: toFormAllowedExtensions(config.allowedExtensions),
    }
    formDialogMode.value = 'edit'
    return
  }
  currentConfig.value = defaultConfig(props.routeConfigId)
  formDialogMode.value = 'create'
}

const handleFormSubmit = async (formData?: Record<string, any>) => {
  if (!formData || !props.routeConfigId) {
    return
  }
  saving.value = true
  try {
    const response = await saveStaticHostConfig({
      ...formData,
      routeConfigId: props.routeConfigId,
      staticHostConfigId: currentConfig.value?.staticHostConfigId,
      indexFiles: typeof formData.indexFiles === 'string' ? formData.indexFiles : 'index.html',
      rewriteRules: typeof formData.rewriteRules === 'string' ? formData.rewriteRules : '',
    })
    if (isApiSuccess(response)) {
      message.success(getApiMessage(response, '静态资源配置保存成功'))
      emit('update:visible', false)
      emit('saved')
      return
    }
    message.error(getApiMessage(response, '静态资源配置保存失败'))
  } catch {
    message.error('静态资源配置保存失败')
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
