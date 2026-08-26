<template>
  <div class="env-vars">
    <RsAlert type="info" class="hint">{{
      t('envVars.hint', { syntax: '${NAME}', example: '${GATEWAY_INTERNAL}' })
    }}</RsAlert>
    <div class="env-vars__toolbar">
      <RsButton variant="primary" :disabled="!canEdit" @click="openCreate">
        {{ t('envVars.add') }}
      </RsButton>
    </div>
    <RsTable
      class="env-vars__table"
      :columns="columns"
      :data="envVars.items"
      row-key="name"
      size="sm"
      height="100%"
    >
      <template #empty>{{ t('envVars.empty') }}</template>
    </RsTable>

    <RsDialog
      :open="dialogOpen"
      :title="editing ? t('envVars.editTitle') : t('envVars.addTitle')"
      layout="window"
      :width="480"
      :show-overlay="true"
      :close-on-overlay-click="false"
      @update:open="onDialogOpen"
    >
      <template #body>
        <RsForm
          ref="formRef"
          :model="form"
          :rules="formRules"
          label-position="left"
          label-width="7rem"
          gap="md"
        >
          <RsInput
            v-model="form.name"
            name="name"
            :label="t('envVars.name')"
            :placeholder="t('envVars.namePlaceholder')"
            :maxlength="64"
          />
          <RsFormItem :label="t('envVars.secret')">
            <RsSwitch v-model="form.secret" />
          </RsFormItem>
          <RsInput
            v-model="form.value"
            name="value"
            :label="t('envVars.value')"
            :type="form.secret ? 'password' : 'text'"
            :placeholder="valuePlaceholder"
            :visibility-toggle="form.secret"
          />
          <RsInput
            v-model="form.note"
            name="note"
            :label="t('envVars.note')"
            :placeholder="t('envVars.notePlaceholder')"
            :maxlength="256"
          />
        </RsForm>
      </template>
      <template #footer>
        <div class="env-vars__footer">
          <RsButton variant="secondary" @click="dialogOpen = false">{{ t('envVars.cancel') }}</RsButton>
          <RsButton variant="primary" :loading="savingEnvVar" :disabled="!canEdit" @click="submit">
            {{ t('common.save') }}
          </RsButton>
        </div>
      </template>
    </RsDialog>
  </div>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import {
  rsConfirm,
  RsAlert,
  RsButton,
  RsDialog,
  RsForm,
  RsFormItem,
  RsInput,
  RsSwitch,
  RsTable,
  RsTag,
  type RsFormValidationResult,
  type RsFormRules,
  type RsTableColumn,
} from '@/ui'
import { computed, h, reactive, ref } from 'vue'
import type { EnvVarItem, EnvVarsSettings } from '../types'

defineOptions({ name: 'EnvVarsPanel' })

const props = defineProps<{
  envVars: EnvVarsSettings
  canEdit: boolean
  savingEnvVar: boolean
}>()

const emit = defineEmits<{
  save: [payload: { name: string; originalName?: string; value: string; secret: boolean; note: string }]
  remove: [name: string]
}>()

const { t } = useModuleI18n('hub0009')
const dialogOpen = ref(false)
const editing = ref(false)
const formRef = ref<{ validate?: () => Promise<RsFormValidationResult> } | null>(null)
const form = reactive({
  originalName: '',
  name: '',
  value: '',
  secret: false,
  note: '',
})

const valuePlaceholder = computed(() => {
  if (editing.value && form.secret) {
    return t('envVars.valueKeepPlaceholder')
  }
  return t('envVars.valuePlaceholder')
})

const formRules: RsFormRules = {
  name: [
    { required: true, message: t('envVars.nameRequired'), trigger: ['blur', 'change'] },
    {
      pattern: /^[A-Za-z_][A-Za-z0-9_]*$/,
      message: t('envVars.namePattern'),
      trigger: ['blur', 'change'],
    },
  ],
}

const columns = computed<RsTableColumn<EnvVarItem>[]>(() => [
  { title: t('envVars.name'), key: 'name' },
  {
    title: t('envVars.secret'),
    key: 'secret',
    width: 88,
    render: (row) =>
      h(
        RsTag,
        { variant: row.secret ? 'warning' : 'default', size: 'sm' },
        { default: () => (row.secret ? t('envVars.secretYes') : t('envVars.secretNo')) },
      ),
  },
  {
    title: t('envVars.value'),
    key: 'value',
    render: (row) => (row.secret ? (row.hasValue ? '********' : '') : row.value || ''),
  },
  { title: t('envVars.note'), key: 'note' },
  {
    title: t('envVars.actions'),
    key: 'actions',
    width: 148,
    render: (row) =>
      h('div', { class: 'env-vars__actions' }, [
        h(
          RsButton,
          {
            variant: 'ghost',
            size: 'sm',
            disabled: !props.canEdit,
            onClick: () => openEdit(row),
          },
          { default: () => t('envVars.edit') },
        ),
        h(
          RsButton,
          {
            variant: 'ghost',
            size: 'sm',
            disabled: !props.canEdit,
            onClick: () => confirmRemove(row),
          },
          { default: () => t('envVars.delete') },
        ),
      ]),
  },
])

const resetForm = () => {
  form.originalName = ''
  form.name = ''
  form.value = ''
  form.secret = false
  form.note = ''
}

const openCreate = () => {
  editing.value = false
  resetForm()
  dialogOpen.value = true
}

const openEdit = (row: EnvVarItem) => {
  editing.value = true
  form.originalName = row.name
  form.name = row.name
  form.value = row.secret ? '' : row.value || ''
  form.secret = row.secret
  form.note = row.note || ''
  dialogOpen.value = true
}

const onDialogOpen = (open: boolean) => {
  dialogOpen.value = open
  if (!open) {
    resetForm()
    editing.value = false
  }
}

const submit = async () => {
  if (formRef.value?.validate) {
    const result = await formRef.value.validate()
    if (!result?.valid) {
      return
    }
  }
  emit('save', {
    name: form.name.trim(),
    originalName: editing.value ? form.originalName : undefined,
    value: form.value,
    secret: form.secret,
    note: form.note.trim(),
  })
}

const confirmRemove = async (row: EnvVarItem) => {
  const ok = await rsConfirm.warning({
    title: t('envVars.deleteTitle'),
    description: t('envVars.deleteConfirm', { name: row.name }),
    confirmText: t('envVars.delete'),
    cancelText: t('envVars.cancel'),
  })
  if (ok) {
    emit('remove', row.name)
  }
}

defineExpose({
  closeDialog: () => {
    dialogOpen.value = false
  },
})
</script>

<style lang="scss" scoped>
.env-vars {
  display: flex;
  flex-direction: column;
  min-height: 0;
  flex: 1 1 auto;
}

.env-vars__toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.env-vars__table {
  flex: 1 1 auto;
  min-height: 0;
}

.env-vars__footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.env-vars :deep(.env-vars__actions) {
  display: flex;
  gap: 4px;
}

.hint {
  margin-bottom: 16px;
  flex-shrink: 0;
}
</style>
