<template>
  <RsDialog
    :open="visible"
    :title="computedTitle || ' '"
    layout="window"
    :width="width"
    :draggable="draggable"
    :fullscreenable="showFullscreenToggle"
    :modal="modal"
    :show-overlay="mask"
    :show-close="closable"
    :close-on-overlay-click="maskClosable"
    :teleport-to="to"
    class="rs-data-form-modal"
    @update:open="handleUpdateVisible"
    @after-open="emit('after-enter')"
    @after-close="() => emit('after-leave')"
  >
    <template #body>
      <div class="rs-data-form-modal__body">
        <slot
          :mode="mode"
          :form-data="formDataProxy"
          :form-ref="formRef"
        >
          <RsDataForm
            ref="formRef"
            class="rs-data-form-modal__form"
            :mode="mode"
            :form-fields="formFields"
            :form-tabs="formTabs"
            :initial-data="initialData"
            :label-placement="labelPlacement"
            :label-align="labelAlign"
            :label-width="labelWidth"
            :size="size"
            @update:model-value="(data) => (formDataProxy = data)"
          />
        </slot>
      </div>
    </template>

    <template
      v-if="showFooter"
      #footer
    >
      <slot
        name="footer"
        :mode="mode"
        :confirm-loading="confirmLoading"
        :on-confirm="handleConfirm"
        :on-cancel="handleCancel"
      >
        <div class="rs-data-form-modal__footer">
          <RsButton
            v-if="showCancel"
            size="sm"
            @click="handleCancel"
          >
            {{ resolvedCancelText }}
          </RsButton>
          <RsButton
            v-if="showConfirm && mode !== 'view'"
            variant="primary"
            size="sm"
            :loading="confirmLoading"
            @click="handleConfirm"
          >
            {{ resolvedConfirmText }}
          </RsButton>
        </div>
      </slot>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsButton, RsDialog } from '@/ui'
import { computed, ref } from 'vue'
import RsDataForm from './RsDataForm.vue'
import type {
  RsDataFormExpose,
  RsDataFormModalEmits,
  RsDataFormModalProps,
} from './types'

defineOptions({
  name: 'RsDataFormModal',
})

const props = withDefaults(defineProps<RsDataFormModalProps>(), {
  visible: false,
  title: '',
  width: '720px',
  mode: 'create',
  formFields: () => [],
  formTabs: () => [],
  autoCloseOnConfirm: true,
  autoResetOnClose: false,
  showFooter: true,
  showCancel: true,
  showConfirm: true,
  confirmLoading: false,
  mask: false,
  maskClosable: false,
  /** 默认非模态：无遮罩时可操作背后页面 / 其它模块 */
  modal: false,
  closable: true,
  draggable: true,
  showFullscreenToggle: true,
  labelPlacement: 'top',
  labelAlign: 'left',
  labelWidth: 'auto',
  size: 'small',
  // to 默认 undefined：不强制全局 body；业务可传选择器或 false 禁用 Teleport
  to: undefined,
})

const emit = defineEmits<RsDataFormModalEmits>()

const { t } = useModuleI18n('common')

const formRef = ref<RsDataFormExpose | null>(null)
const formDataProxy = ref<Record<string, any>>({})

const resolvedCancelText = computed(() => props.cancelText ?? t('cancel'))
const resolvedConfirmText = computed(() => props.confirmText ?? t('save'))

const computedTitle = computed(() => {
  if (props.title) return props.title
  switch (props.mode) {
    case 'create':
      return t('formModal.create')
    case 'edit':
      return t('formModal.edit')
    case 'view':
      return t('formModal.view')
    default:
      return ''
  }
})

const handleUpdateVisible = (open: boolean) => {
  emit('update:visible', open)
  if (!open) {
    emit('close')
    if (props.autoResetOnClose) {
      formRef.value?.reset()
      emit('reset')
    }
  }
}

const handleCancel = () => {
  emit('cancel')
  handleUpdateVisible(false)
}

const handleConfirm = async () => {
  if (props.mode === 'view') {
    handleUpdateVisible(false)
    return
  }

  const ok = await formRef.value?.validate()
  if (!ok) return

  const formData = formRef.value?.getFormData()
  emit('submit', formData)
  emit('confirm', formData)

  if (props.autoCloseOnConfirm) {
    handleUpdateVisible(false)
  }
}
</script>

<style scoped lang="scss">
/*
 * 弹窗外壳样式由 RsDialog + rs-brand token 负责。
 * 这里只处理内容区滚动与页脚按钮排布，避免与 Dialog 的 padding 叠出「空心」感。
 */
.rs-data-form-modal__body {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-height: 0;
  max-height: min(70vh, 640px);
  overflow: auto;
}

.rs-data-form-modal__form {
  width: 100%;
  min-width: 0;
}

.rs-data-form-modal__footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--rs-space-sm, 8px);
  width: 100%;
}
</style>
