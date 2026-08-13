<template>
  <div class="rs-data-form">
    <RsForm
      ref="formRef"
      :rules="formRules"
      :label-position="rsLabelPosition"
      :label-align="rsLabelAlign"
      :label-width="rsLabelWidth"
      :size="rsSize"
      gap="sm"
      max-width="full"
      class="rs-data-form__form"
    >
      <div
        v-if="tabs.length > 1"
        class="rs-data-form__tabs-wrap"
      >
        <!--
          panelless：只用导航栏，内容区自行按 activeTab 切换。
          避免 TabsContent 未隐藏时多页字段叠在同一视口（分组「漂」到右侧）。
        -->
        <RsTabs
          v-model="activeTab"
          :items="rsTabItems"
          variant="line"
          size="sm"
          panelless
          class="rs-data-form__tabs"
        />
        <div
          v-for="tab in tabs"
          v-show="activeTab === tab.key"
          :key="tab.key"
          class="rs-data-form__tab-panel"
          role="tabpanel"
        >
          <RsDataFormFields
            :fields="getFieldsByTab(tab.key)"
            :form-model="formModel"
            :mode="mode"
            :label-width="labelWidth"
            :label-placement="labelPlacement"
            :label-align="labelAlign"
            :size="size"
            :cols="cols"
            :x-gap="xGap"
            :y-gap="yGap"
            @field-update="handleFieldValueUpdate"
          />
        </div>
      </div>

      <RsDataFormFields
        v-else
        :fields="formFields"
        :form-model="formModel"
        :mode="mode"
        :label-width="labelWidth"
        :label-placement="labelPlacement"
        :label-align="labelAlign"
        :size="size"
        :cols="cols"
        :x-gap="xGap"
        :y-gap="yGap"
        @field-update="handleFieldValueUpdate"
      />
    </RsForm>

    <div
      v-if="showFooter"
      class="rs-data-form__footer"
    >
      <slot name="footer" :form-data="formModel" :submit="handleSubmit">
        <div class="rs-data-form__footer-actions">
          <RsButton
            v-if="showSubmit && mode !== 'view'"
            variant="primary"
            size="sm"
            :loading="submitLoading"
            @click="handleSubmit"
          >
            {{ submitText }}
          </RsButton>
        </div>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  RsButton,
  RsForm,
  RsTabs,
  type RsFormRules,
  type RsFormValidationResult,
} from '@/ui'
import { computed, ref, watch } from 'vue'
import RsDataFormFields from './RsDataFormFields.vue'
import type {
  RsDataFormEmits,
  RsDataFormExpose,
  RsDataFormField,
  RsDataFormProps,
  RsDataFormTab,
} from './types'

defineOptions({
  name: 'RsDataForm',
})

const props = withDefaults(defineProps<RsDataFormProps>(), {
  mode: 'create',
  formFields: () => [],
  formTabs: () => [],
  initialData: () => ({}),
  injectMode: true,
  showFooter: false,
  showSubmit: false,
  submitText: '保存',
  submitLoading: false,
  labelWidth: 'auto',
  labelPlacement: 'top',
  labelAlign: 'left',
  size: 'small',
  cols: 24,
  xGap: 16,
  yGap: 8,
})

const emit = defineEmits<RsDataFormEmits>()

type RsFormRef = {
  validate: () => Promise<RsFormValidationResult>
  clearValidation: (names?: string | string[]) => void
  resetFields: (names?: string | string[]) => void
}

const formRef = ref<RsFormRef | null>(null)
const formModel = ref<Record<string, any>>({})

const rsSize = computed(() => {
  switch (props.size) {
    case 'small':
      return 'sm' as const
    case 'large':
      return 'lg' as const
    default:
      return 'md' as const
  }
})

const rsLabelPosition = computed(() => props.labelPlacement)
const rsLabelAlign = computed(() =>
  props.labelAlign === 'right' ? ('end' as const) : ('start' as const),
)
const rsLabelWidth = computed(() =>
  typeof props.labelWidth === 'number' ? `${props.labelWidth}px` : String(props.labelWidth),
)

const tabs = computed<RsDataFormTab[]>(() => {
  let tabsList: RsDataFormTab[] = []

  if (props.formTabs?.length) {
    tabsList = props.formTabs
  } else {
    const tabKeys = Array.from(
      new Set(
        (props.formFields || [])
          .map((field) => field.tabKey)
          .filter((key): key is string => !!key),
      ),
    )
    tabsList =
      tabKeys.length > 0
        ? tabKeys.map((key) => ({ key, label: key }))
        : [{ key: 'default', label: '表单' }]
  }

  return tabsList.filter((tab) => {
    if (typeof tab.show === 'function') return tab.show(formModel.value)
    return tab.show !== false
  })
})

const activeTab = ref<string>(tabs.value[0]?.key || 'default')

const rsTabItems = computed(() =>
  tabs.value.map((tab) => ({ value: tab.key, label: tab.label })),
)

watch(
  tabs,
  (newTabs) => {
    if (newTabs.length > 0 && !newTabs.some((tab) => tab.key === activeTab.value)) {
      activeTab.value = newTabs[0]?.key || 'default'
    }
  },
  { immediate: false },
)

const getFieldsByTab = (tabKey: string): RsDataFormField[] => {
  const filterFields = (fields: RsDataFormField[], parentTabKey?: string): RsDataFormField[] =>
    fields
      .filter((field) => {
        const fieldTabKey = field.tabKey || parentTabKey || tabs.value[0]?.key || 'default'
        return fieldTabKey === tabKey
      })
      .map((field) => {
        if (field.type === 'fieldset' && field.children) {
          const currentTabKey = field.tabKey || parentTabKey || tabs.value[0]?.key || 'default'
          return {
            ...field,
            children: filterFields(field.children, currentTabKey),
          }
        }
        return field
      })

  return filterFields(props.formFields)
}

const getFieldLabel = (field: RsDataFormField): string =>
  typeof field.label === 'function' ? field.label(formModel.value) : field.label

/** 将 ISO / 其它可解析日期规范为 RsDatePicker 的 string 格式 */
const normalizeDateInput = (value: unknown, withTime: boolean): string => {
  if (value == null || value === '') return ''
  if (typeof value === 'number') {
    const d = new Date(value)
    if (Number.isNaN(d.getTime())) return ''
    return formatLocalDate(d, withTime)
  }
  if (typeof value !== 'string') return ''
  // 已是 YYYY-MM-DD[ HH:mm:ss] 则原样返回
  if (/^\d{4}-\d{2}-\d{2}( \d{2}:\d{2}(:\d{2})?)?$/.test(value)) return value
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  return formatLocalDate(d, withTime)
}

const formatLocalDate = (d: Date, withTime: boolean) => {
  const pad = (n: number) => String(n).padStart(2, '0')
  const date = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
  if (!withTime) return date
  return `${date} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const processFieldDefaultValue = (
  field: RsDataFormField,
  model: Record<string, any>,
  initialData: Record<string, any>,
) => {
  if (field.type === 'fieldset') {
    field.children?.forEach((child) => processFieldDefaultValue(child, model, initialData))
    return
  }

  const key = field.field
  if (Object.prototype.hasOwnProperty.call(initialData, key) && initialData[key] !== undefined) {
    let value = initialData[key]
    // select 与 RsSelect 对齐：内部展示用 string，提交时再还原类型
    if (field.type === 'select' && value !== null && value !== '') {
      value = String(value)
    }
    if (field.type === 'date' || field.type === 'datetime') {
      value =
        value === null || value === undefined
          ? ''
          : normalizeDateInput(value, field.type === 'datetime')
    }
    if (field.type === 'daterange' || field.type === 'datetimerange') {
      const withTime = field.type === 'datetimerange'
      if (value === null || value === undefined) {
        value = { start: '', end: '' }
      } else if (Array.isArray(value)) {
        value = {
          start: normalizeDateInput(value[0], withTime),
          end: normalizeDateInput(value[1], withTime),
        }
      } else if (typeof value === 'object') {
        value = {
          start: normalizeDateInput((value as any).start, withTime),
          end: normalizeDateInput((value as any).end, withTime),
        }
      }
    }
    model[key] = value
    return
  }

  if (Object.prototype.hasOwnProperty.call(field, 'defaultValue')) {
    model[key] =
      field.type === 'select' && field.defaultValue != null
        ? String(field.defaultValue)
        : field.defaultValue
    return
  }

  switch (field.type) {
    case 'date':
    case 'datetime':
      model[key] = ''
      break
    case 'daterange':
    case 'datetimerange':
      model[key] = { start: '', end: '' }
      break
    case 'number':
      model[key] = null
      break
    case 'switch':
      model[key] = field.props?.uncheckedValue ?? false
      break
    case 'file':
      model[key] = []
      break
    case 'select':
      model[key] = ''
      break
    default:
      model[key] = ''
  }
}

/** 收集 type=file 字段名，JSON 深拷贝会丢掉 File，需从原始 initialData 回填 */
const collectFileFieldKeys = (fields: RsDataFormField[]): string[] => {
  const keys: string[] = []
  const walk = (list: RsDataFormField[]) => {
    list.forEach((field) => {
      if (field.type === 'fieldset') {
        if (field.children) walk(field.children)
        return
      }
      if (field.type === 'file') keys.push(field.field)
    })
  }
  walk(fields)
  return keys
}

const initFormModel = () => {
  const model: Record<string, any> = {}
  const rawInitial = props.initialData ?? {}
  const initialData = props.initialData
    ? JSON.parse(
        JSON.stringify(rawInitial, (_key, value) => {
          // File / File[] 无法 JSON 序列化，占位后由下方按字段恢复
          if (value instanceof File) return null
          if (Array.isArray(value) && value.some((item) => item instanceof File)) return []
          return value
        }),
      )
    : {}

  collectFileFieldKeys(props.formFields).forEach((key) => {
    if (Object.prototype.hasOwnProperty.call(rawInitial, key)) {
      initialData[key] = rawInitial[key]
    }
  })

  props.formFields.forEach((field) => processFieldDefaultValue(field, model, initialData))
  if (props.injectMode) {
    model._mode = props.mode
  }
  formModel.value = model
  formRef.value?.clearValidation()
  emit('update:modelValue', { ...formModel.value })
}

const processFieldRules = (field: RsDataFormField, rules: RsFormRules) => {
  if (field.type === 'fieldset') {
    field.children?.forEach((child) => processFieldRules(child, rules))
    return
  }

  if (field.rules) {
    rules[field.field] = field.rules
    return
  }

  if (!field.required) return

  const label = getFieldLabel(field)
  if (field.type === 'number') {
    rules[field.field] = {
      required: true,
      message: `请输入${label}`,
      validator: (value: any) => {
        if (value === null || value === undefined || value === '') {
          return `请输入${label}`
        }
        const num = typeof value === 'number' ? value : Number(value)
        if (Number.isNaN(num)) return `${label}必须是数字`
        return true
      },
    }
    return
  }

  const pickTypes = new Set([
    'select',
    'date',
    'datetime',
    'daterange',
    'datetimerange',
    'switch',
  ])
  const prefix = pickTypes.has(field.type || '') ? '请选择' : '请输入'
  rules[field.field] = {
    required: true,
    message: `${prefix}${label}`,
  }
}

const formRules = computed<RsFormRules>(() => {
  const rules: RsFormRules = {}
  props.formFields.forEach((field) => processFieldRules(field, rules))
  return rules
})

/** 提交前把 select 的 string 还原为 options 原始类型 */
const normalizeSubmitData = (data: Record<string, any>) => {
  const result = { ...data }
  const walk = (fields: RsDataFormField[]) => {
    fields.forEach((field) => {
      if (field.type === 'fieldset') {
        if (field.children) walk(field.children)
        return
      }
      if (field.type === 'select') {
        const raw = result[field.field]
        if (raw === '' || raw == null) return
        const matched = field.options?.find((opt) => String(opt.value) === String(raw))
        if (matched) result[field.field] = matched.value
      }
    })
  }
  walk(props.formFields)
  delete result._mode
  return result
}

const handleFieldValueUpdate = (field: RsDataFormField, value: any) => {
  formModel.value[field.field] = value
  if (typeof field.props?.onUpdateValue === 'function') {
    field.props.onUpdateValue(value, formModel.value)
  }
  emit('update:modelValue', { ...formModel.value })
}

const handleSubmit = async () => {
  const ok = await validate()
  if (!ok) return
  emit('submit', normalizeSubmitData(formModel.value))
}

watch(
  () => [props.initialData, props.mode] as const,
  () => {
    initFormModel()
  },
  { deep: true, immediate: true },
)

const validate = async (): Promise<boolean> => {
  const result = await formRef.value?.validate()
  return !result || result.valid
}

const reset = () => {
  initFormModel()
  emit('reset')
}

const getFormData = () => normalizeSubmitData(formModel.value)

const setFormData = (data: Record<string, any>) => {
  formModel.value = {
    ...data,
    ...(props.injectMode ? { _mode: props.mode } : {}),
  }
  emit('update:modelValue', { ...formModel.value })
}

defineExpose<RsDataFormExpose>({
  validate,
  reset,
  getFormData,
  setFormData,
  formRef,
})
</script>

<style scoped lang="scss">
.rs-data-form {
  width: 100%;
}

.rs-data-form__tabs-wrap {
  width: 100%;
}

.rs-data-form__tabs {
  width: 100%;
}

.rs-data-form__tab-panel {
  padding-top: var(--rs-space-md, 12px);
  box-sizing: border-box;
}

.rs-data-form__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: var(--g-space-md, 12px) 0;
  border-top: 1px solid var(--g-border-primary, var(--rs-border));
  margin-top: var(--g-space-md, 12px);
}

.rs-data-form__footer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
