<template>
  <div class="cors-config-form-modal" id="cors-config-form-modal">
    <!-- CORS配置表单对话框（新增/编辑/查看共用） -->
    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="computedTitle"
      :width="props.width || 800"
      :to="props.to || '#cors-config-form-modal'"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditConfig || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
      @update:visible="handleFormDialogVisibleChange"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { computed, nextTick, ref, watch } from 'vue'
import { useCorsConfigPage } from './hooks'
import type { CorsConfigFormModalEmits, CorsConfigFormModalProps } from './hooks/types'

defineOptions({
  name: 'CorsConfigFormModal',
})

const props = withDefaults(defineProps<CorsConfigFormModalProps>(), {
  visible: false,
  width: 800,
  to: undefined,
  gatewayInstanceId: undefined,
  routeConfigId: undefined,
})

const emit = defineEmits<CorsConfigFormModalEmits>()

const gatewayInstanceId = ref(props.gatewayInstanceId)
const routeConfigId = ref(props.routeConfigId)
const moduleIdRef = ref(props.moduleId)

const page = useCorsConfigPage({
  gatewayInstanceId,
  routeConfigId,
  moduleId: moduleIdRef,
})

const { service, formDialogVisible, formDialogMode, currentEditConfig, openDialog, closeFormDialog, handleFormSubmit } = page

const computedTitle = computed(() => {
  if (props.title) return props.title
  if (formDialogMode.value === 'create') {
    return '新增CORS配置'
  }
  if (formDialogMode.value === 'edit') {
    return '编辑CORS配置'
  }
  return '查看CORS配置详情'
})

watch(
  () => props.visible,
  async (val) => {
    if (val) {
      gatewayInstanceId.value = props.gatewayInstanceId
      routeConfigId.value = props.routeConfigId
      moduleIdRef.value = props.moduleId

      if (formDialogVisible.value) {
        closeFormDialog()
        await nextTick()
      }
      await openDialog()
      return
    }
    if (formDialogVisible.value) {
      closeFormDialog()
    }
  },
)

const handleFormDialogVisibleChange = (value: boolean) => {
  if (!value) {
    emit('update:visible', false)
    emit('close')
    emit('refresh')
  }
}
</script>
