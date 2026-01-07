<template>
  <div class="cors-config-form-modal" id="cors-config-form-modal">
    <!-- CORS配置表单对话框（新增/编辑/查看共用） -->
    <GdataFormModal
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
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import { computed, nextTick, ref, watch } from 'vue'
import { useCorsConfigPage } from './hooks'
import type { CorsConfigFormModalEmits, CorsConfigFormModalProps } from './hooks/types'

// 定义组件名称
defineOptions({
  name: 'CorsConfigFormModal'
})

// ============= Props =============

const props = withDefaults(defineProps<CorsConfigFormModalProps>(), {
  visible: false,
  width: 800,
  to: undefined,
  gatewayInstanceId: undefined,
  routeConfigId: undefined,
})

// ============= Emits =============

const emit = defineEmits<CorsConfigFormModalEmits>()

// ============= 页面逻辑 =============

// ============= 响应式 Props =============

// 创建响应式的 ref，用于传递给 Hook（不需要 watch，因为对话框打开时这些值已确定）
const gatewayInstanceId = ref<string | undefined>(props.gatewayInstanceId)
const routeConfigId = ref<string | undefined>(props.routeConfigId)
const moduleIdRef = ref<string>(props.moduleId)

// 使用页面级 Hook（传递响应式的 ref，包括 moduleId）
const page = useCorsConfigPage({
  gatewayInstanceId,
  routeConfigId,
  moduleId: moduleIdRef,
})

const { service, formDialogVisible, formDialogMode, currentEditConfig, openDialog, closeFormDialog, handleFormSubmit } = page

// ============= 计算属性 =============

// 计算标题（响应 formDialogMode 的变化）
const computedTitle = computed(() => {
  // 如果传入了自定义标题，优先使用
  if (props.title) return props.title
  // 根据模式动态生成标题
  if (formDialogMode.value === 'create') {
    return '新增CORS配置'
  } else if (formDialogMode.value === 'edit') {
    return '编辑CORS配置'
  } else {
    return '查看CORS配置详情'
  }
})

// ============= 模态框状态管理 =============

// 监听 props.visible 变化，同步到表单对话框
watch(
  () => props.visible,
  async (val) => {
    if (val) {
      // 打开模态框时，同步最新的 props 值到响应式 ref
      gatewayInstanceId.value = props.gatewayInstanceId
      routeConfigId.value = props.routeConfigId
      moduleIdRef.value = props.moduleId
      
      // 如果表单对话框已经打开，先关闭再打开，确保状态重置
      if (formDialogVisible.value) {
        closeFormDialog()
        // 使用 nextTick 确保关闭后再打开
        await nextTick()
      }
      // 打开对话框（Hook 内部会使用响应式的 gatewayInstanceId、routeConfigId 和 moduleId）
      await openDialog()
    } else {
      // 关闭模态框时，关闭表单对话框
      if (formDialogVisible.value) {
        closeFormDialog()
      }
    }
  }
)

// 处理表单对话框可见性变化
const handleFormDialogVisibleChange = (value: boolean) => {
  if (!value) {
    // 表单对话框关闭时，触发事件
    emit('update:visible', false)
    emit('close')
    emit('refresh')
  }
}
</script>

<style scoped lang="scss">
.cors-config-form-modal {
  min-height: 200px;
}
</style>

