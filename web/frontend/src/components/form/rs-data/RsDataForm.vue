<template>
  <div class="rs-data-form">
    <RsForm
      ref="formRef"
      :model="formModel"
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
            :set-field-value="setFieldValue"
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
        :set-field-value="setFieldValue"
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
  type RsFormNamePath,
  type RsFormRules,
  type RsFormValidationResult,
} from '@/ui'
import { formatDate } from '@/utils/format'
import { computed, ref, watch } from 'vue'
import RsDataFormFields from './RsDataFormFields.vue'
import { getByNamePath, readInitialField, seedNamedObjects, setByNamePath, toRfc3339SubmitValue } from './form-model'
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

/** 将 ISO / 其它可解析日期交给 DatePicker：RFC3339 原样保留（valueFormat=iso），墙钟字符串原样保留 */
const normalizeDateInput = (value: unknown, withTime: boolean): string => {
  if (value == null || value === '') return ''
  if (typeof value === 'string') {
    const trimmed = value.trim()
    if (/^\d{4}-\d{2}-\d{2}T/.test(trimmed)) return trimmed
    if (/^\d{4}-\d{2}-\d{2}( \d{2}:\d{2}(:\d{2})?)?$/.test(trimmed)) return trimmed
    return formatDate(value, withTime ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD')
  }
  if (typeof value === 'number') {
    return formatDate(value, withTime ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD')
  }
  return ''
}

/** select 在表单模型中保持 options 原始类型；RsSelect 只在控件边界转 string */
const coerceSelectModelValue = (field: RsDataFormField, value: unknown): unknown => {
  if (value === '' || value == null) return value
  const matched = field.options?.find((opt) => String(opt.value) === String(value))
  return matched ? matched.value : value
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
  const initial = readInitialField(initialData, key)
  if (initial.found && initial.value !== undefined) {
    let value: unknown = initial.value
    if (field.type === 'select') {
      value = coerceSelectModelValue(field, value)
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
    setByNamePath(model, key, value)
    return
  }

  if (Object.prototype.hasOwnProperty.call(field, 'defaultValue')) {
    setByNamePath(
      model,
      key,
      field.type === 'select'
        ? coerceSelectModelValue(field, field.defaultValue)
        : field.defaultValue,
    )
    return
  }

  switch (field.type) {
    case 'date':
    case 'datetime':
      setByNamePath(model, key, '')
      break
    case 'daterange':
    case 'datetimerange':
      setByNamePath(model, key, { start: '', end: '' })
      break
    case 'number':
      setByNamePath(model, key, null)
      break
    case 'switch':
      setByNamePath(model, key, field.props?.uncheckedValue ?? false)
      break
    case 'file':
      setByNamePath(model, key, [])
      break
    case 'select':
      setByNamePath(model, key, '')
      break
    default:
      setByNamePath(model, key, '')
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
    const fromRaw = readInitialField(rawInitial, key)
    if (fromRaw.found) {
      setByNamePath(initialData, key, fromRaw.value)
    }
  })

  seedNamedObjects(model, initialData)
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
    'custom',
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

/** 提交前还原 select 类型，并把 date/datetime 转为 RFC3339（Go time.Time） */
const normalizeSubmitData = (data: Record<string, any>) => {
  const result = { ...data }
  const walk = (fields: RsDataFormField[]) => {
    fields.forEach((field) => {
      if (field.type === 'fieldset') {
        if (field.children) walk(field.children)
        return
      }
      if (field.type === 'select') {
        const raw = getByNamePath(result, field.field)
        if (raw === '' || raw == null) return
        const matched = field.options?.find((opt) => String(opt.value) === String(raw))
        if (matched) setByNamePath(result, field.field, matched.value)
        return
      }
      if (field.type === 'date' || field.type === 'datetime') {
        setByNamePath(
          result,
          field.field,
          toRfc3339SubmitValue(getByNamePath(result, field.field), field.type === 'datetime'),
        )
        return
      }
      if (field.type === 'daterange' || field.type === 'datetimerange') {
        const raw = getByNamePath(result, field.field) as { start?: unknown; end?: unknown } | unknown[]
        const withTime = field.type === 'datetimerange'
        if (Array.isArray(raw)) {
          setByNamePath(result, field.field, {
            start: toRfc3339SubmitValue(raw[0], withTime),
            end: toRfc3339SubmitValue(raw[1], withTime),
          })
          return
        }
        if (raw && typeof raw === 'object') {
          setByNamePath(result, field.field, {
            start: toRfc3339SubmitValue((raw as { start?: unknown }).start, withTime),
            end: toRfc3339SubmitValue((raw as { end?: unknown }).end, withTime),
          })
        }
      }
    })
  }
  walk(props.formFields)
  Object.keys(result).forEach((key) => {
    if (key.startsWith('_')) delete result[key]
  })
  return result
}

const handleFieldValueUpdate = (field: RsDataFormField, value: any) => {
  const next = field.type === 'select' ? coerceSelectModelValue(field, value) : value
  setByNamePath(formModel.value, field.field, next)
  if (typeof field.props?.onUpdateValue === 'function') {
    field.props.onUpdateValue(next, formModel.value)
  }
  emit('update:modelValue', { ...formModel.value })
}

const handleSubmit = async () => {
  const ok = await validate()
  if (!ok) return
  emit('submit', normalizeSubmitData(formModel.value))
}

/**
 * 仅在 mode 或 initialData「内容」变化时重置表单。
 * 业务侧常见 :initial-data="getXxx()" / 每次渲染 new 对象，再叠加 confirmLoading
 * 重渲染时，旧的 deep watch 会把用户正在编辑的值刷回旧 initialData（保存闪回）。
 * File 无法稳定序列化，签名里忽略；文件字段仍由 initFormModel 从原始引用回填。
 */
const initialDataSignature = (data: Record<string, any> | undefined): string => {
  try {
    return JSON.stringify(data ?? {}, (_key, value) => {
      if (value instanceof File) return undefined
      if (Array.isArray(value) && value.some((item) => item instanceof File)) {
        return value.map((item) => (item instanceof File ? { name: item.name, size: item.size } : item))
      }
      return value
    })
  } catch {
    return ''
  }
}

const lastInitSignature = ref('')

watch(
  () => `${props.mode}::${initialDataSignature(props.initialData)}`,
  (signature) => {
    if (signature === lastInitSignature.value) return
    lastInitSignature.value = signature
    initFormModel()
  },
  { immediate: true },
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

const getFieldValue = (name: RsFormNamePath): unknown => getByNamePath(formModel.value, name)

const setFieldValue = (name: RsFormNamePath, value: unknown): void => {
  setByNamePath(formModel.value, name, value)
  emit('update:modelValue', { ...formModel.value })
}

const getFieldsValue = (): Record<string, any> => {
  const result = { ...formModel.value }
  Object.keys(result).forEach((key) => {
    if (key.startsWith('_')) delete result[key]
  })
  return result
}

const setFieldsValue = (values: Record<string, unknown>): void => {
  Object.entries(values).forEach(([name, value]) => {
    setByNamePath(formModel.value, name, value)
  })
  emit('update:modelValue', { ...formModel.value })
}

defineExpose<RsDataFormExpose>({
  validate,
  reset,
  getFormData,
  setFormData,
  getFieldValue,
  setFieldValue,
  getFieldsValue,
  setFieldsValue,
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
