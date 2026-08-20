<template>
  <div class="rs-search-form">
    <GToolbar
      v-if="showToolbar"
      :module-id="moduleId"
      :buttons="toolbarButtonsComputed"
      :align="toolbarAlign"
      :bordered="false"
      class="rs-search-form__toolbar"
      @button-click="handleToolbarClick"
    />

    <RsForm
      ref="formRef"
      :model="formData"
      :rules="formRules"
      :label-position="rsLabelPosition"
      :label-align="rsLabelAlign"
      :label-width="rsLabelWidth"
      :size="rsSize"
      gap="sm"
      max-width="full"
      class="rs-search-form__form"
    >
      <div
        class="rs-search-form__grid"
        :style="gridStyle"
      >
        <div
          v-for="field in visibleFields"
          :key="field.field"
          class="rs-search-form__cell"
          :style="{ gridColumn: `span ${field.span || defaultFieldSpan}` }"
        >
          <RsFormItem
            :name="field.field"
            :label="getFieldLabel(field)"
            :required="field.required"
            :label-position="rsLabelPosition"
            class="rs-search-form__item"
          >
            <template
              v-if="resolveTips(field)"
              #label
            >
              <span class="rs-search-form__label-text" :title="getFieldLabel(field)">
                {{ getFieldLabel(field) }}
                <component
                  :is="renderFieldTips(field)"
                />
              </span>
            </template>

            <RsInput
              v-if="field.type === 'input' || !field.type"
              :model-value="asStringValue(formData[field.field])"
              :placeholder="field.placeholder || `请输入${getFieldLabel(field)}`"
              :disabled="field.disabled"
              :clearable="field.clearable !== false"
              :size="rsSize"
              label-position="top"
              v-bind="omitControlProps(field.props)"
              @update:model-value="(v: string) => onFieldUpdate(field, v)"
            />

            <RsInputNumber
              v-else-if="field.type === 'number'"
              :model-value="asNumberValue(formData[field.field])"
              :placeholder="field.placeholder || `请输入${getFieldLabel(field)}`"
              :disabled="field.disabled"
              :size="rsSize"
              label-position="top"
              v-bind="omitControlProps(field.props)"
              @update:model-value="(v) => onFieldUpdate(field, v)"
            />

            <RsDatePicker
              v-else-if="isDateField(field.type)"
              :model-value="formData[field.field]"
              :placeholder="field.placeholder || `请选择${getFieldLabel(field)}`"
              :disabled="field.disabled"
              :range="isDateRangeField(field.type)"
              label-position="top"
              :size="rsSize"
              v-bind="omitControlProps(field.props)"
              value-format="string"
              :with-time="isDateTimeField(field.type)"
              :with-seconds="isDateTimeField(field.type)"
              @update:model-value="(v: unknown) => onFieldUpdate(field, v)"
            />

            <RsSelect
              v-else-if="field.type === 'select'"
              :model-value="formData[field.field]"
              :placeholder="field.placeholder || `请选择${getFieldLabel(field)}`"
              :options="normalizeSelectOptions(field.options)"
              :disabled="field.disabled"
              :clearable="field.clearable !== false"
              :size="rsSize"
              match-trigger-width
              v-bind="omitControlProps(field.props)"
              block
              @update:model-value="(v: unknown) => onFieldUpdate(field, v)"
            />

            <RsSwitch
              v-else-if="field.type === 'switch'"
              :model-value="formData[field.field]"
              :disabled="field.disabled"
              :size="rsSize"
              v-bind="omitControlProps(field.props)"
              @update:model-value="(v: unknown) => onFieldUpdate(field, v)"
            />

            <component
              :is="renderCustomField(field)"
              v-else-if="field.type === 'custom' && field.render"
            />
          </RsFormItem>
        </div>
      </div>
    </RsForm>
  </div>
</template>

<script setup lang="ts">
import type { ToolbarButton } from '@/components/toolbar'
import GToolbar from '@/components/toolbar/GToolbar.vue'
import {
  RsDatePicker,
  RsForm,
  RsFormItem,
  RsInput,
  RsInputNumber,
  RsSelect,
  RsSwitch,
  RsTooltip,
  setByNamePath,
  type RsFormNamePath,
  type RsFormRules,
  type RsFormValidationResult,
  type RsSelectOptions,
} from '@/ui'
import { OptionsOutline, RefreshOutline, SearchOutline } from '@vicons/ionicons5'
import { computed, h, onMounted, ref, type Component, type VNode } from 'vue'
import type {
  RsSearchField,
  RsSearchFieldType,
  RsSearchFormEmits,
  RsSearchFormExpose,
  RsSearchFormProps,
  RsSearchFormRenderContext,
} from './types'

defineOptions({
  name: 'RsSearchForm',
})

const props = withDefaults(defineProps<RsSearchFormProps>(), {
  fields: () => [],
  /* 固定标签列宽；栅格内各列控件起点对齐。过短会省略，业务可按最长标签上调 */
  labelWidth: '7rem',
  labelPlacement: 'left',
  labelAlign: 'right',
  size: 'small',
  inline: false,
  cols: 24,
  xGap: 12,
  yGap: 10,
  showSearchButton: true,
  showResetButton: true,
  searchButtonText: '查询',
  resetButtonText: '重置',
  resetButtonKey: 'reset',
  moreButtonText: '更多条件',
  toolbarAlign: 'right',
  showToolbar: true,
})

const emit = defineEmits<RsSearchFormEmits>()

/** RsForm 实例暴露的方法 */
type RsSearchFormRef = {
  validate: () => Promise<RsFormValidationResult>
  clearValidation: (names?: string | string[]) => void
  resetFields: (names?: string | string[]) => void
}

const formRef = ref<RsSearchFormRef | null>(null)
const formData = ref<Record<string, any>>({})
const showMoreFields = ref(false)

/** 切换更多条件显示/隐藏 */
const toggleMoreFields = () => {
  showMoreFields.value = !showMoreFields.value
}

/** 将业务 size 映射为 niuma-ui 尺寸 */
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

const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${props.cols}, minmax(0, 1fr))`,
  columnGap: `${props.xGap}px`,
  rowGap: `${props.yGap}px`,
  '--rs-field-label-width': rsLabelWidth.value,
  '--rs-search-control-height': `var(--rs-control-height-${rsSize.value})`,
}))

const isDateField = (type?: RsSearchFieldType) =>
  type === 'date' || type === 'daterange' || type === 'datetime' || type === 'datetimerange'

const isDateRangeField = (type?: RsSearchFieldType) =>
  type === 'daterange' || type === 'datetimerange'

const isDateTimeField = (type?: RsSearchFieldType) =>
  type === 'datetime' || type === 'datetimerange'

/** 控件不再画 label；业务 props 不得覆盖 Form.Item 契约 */
const omitControlProps = (extra?: Record<string, unknown>) => {
  if (!extra) return {}
  const {
    label: _label,
    required: _required,
    name: _name,
    labelPosition: _labelPosition,
    valueFormat: _valueFormat,
    ...rest
  } = extra
  return rest
}

const asStringValue = (value: unknown): string => {
  if (value == null) return ''
  if (typeof value === 'string') return value
  return String(value)
}

const asNumberValue = (value: unknown): number | null => {
  if (value == null || value === '') return null
  const num = typeof value === 'number' ? value : Number(value)
  return Number.isNaN(num) ? null : num
}

/**
 * 将业务 options 交给 RsSelect；value 保持 string | number。
 */
const normalizeSelectOptions = (
  options?: RsSearchField['options'],
): RsSelectOptions => {
  if (!options?.length) return []
  return options.map((opt) => ({
    label: opt.label,
    value: opt.value,
    disabled: opt.disabled,
  }))
}

/** 从字段配置收集 RsForm rules */
const formRules = computed<RsFormRules>(() => {
  const rules: RsFormRules = {}
  const allFields = [...props.fields, ...(props.moreFields || [])]
  for (const field of allFields) {
    if (field.rules) {
      rules[field.field] = field.rules
    } else if (field.required) {
      const label =
        typeof field.label === 'function' ? field.label(formData.value) : field.label
      rules[field.field] = {
        required: true,
        message: `请输入${label}`,
      }
    }
  }
  return rules
})

const emptyRange = (): { start: string; end: string } => ({ start: '', end: '' })

/** 初始化表单数据（只使用字段的 defaultValue；range 只认 { start, end }） */
const initFormData = () => {
  const data: Record<string, any> = {}
  const applyDefaults = (fields: RsSearchField[]) => {
    fields.forEach((field) => {
      if (field.defaultValue !== undefined) {
        data[field.field] = field.defaultValue
        return
      }
      switch (field.type) {
        case 'number':
          data[field.field] = null
          break
        case 'switch':
          data[field.field] = false
          break
        case 'date':
        case 'datetime':
          data[field.field] = ''
          break
        case 'daterange':
        case 'datetimerange':
          data[field.field] = emptyRange()
          break
        case 'select':
          data[field.field] = ''
          break
        default:
          data[field.field] = ''
      }
    })
  }
  applyDefaults(props.fields)
  if (props.moreFields) {
    applyDefaults(props.moreFields)
  }
  formData.value = data
}

/** 当前可见字段（基础 + 展开后的更多条件） */
const visibleFields = computed(() => {
  const basic = props.fields.filter((field) => field.show !== false)
  if (!showMoreFields.value || !props.moreFields?.length) {
    return basic
  }
  return [...basic, ...props.moreFields.filter((field) => field.show !== false)]
})

const hasMoreFields = computed(() => (props.moreFields || []).length > 0)

/** 默认字段占位：一行 4 个 */
const defaultFieldSpan = computed(() => Math.floor(props.cols / 4))

const getFieldLabel = (field: RsSearchField): string => {
  return typeof field.label === 'function' ? field.label(formData.value) : field.label
}

const resolveTips = (
  field: RsSearchField,
): string | Component | VNode | null => {
  if (!field.tips) return null
  if (typeof field.tips === 'function') {
    const tipsFactory = field.tips as (
      data: Record<string, any>,
    ) => string | Component | VNode
    return tipsFactory(formData.value)
  }
  return field.tips
}

const renderFieldTips = (field: RsSearchField) => {
  const tips = resolveTips(field)
  if (tips == null) return () => null
  if (typeof tips === 'string') {
    return () =>
      h('span', { class: 'rs-search-form__label-tips' }, [
        h(RsTooltip, { content: tips, icon: true }),
      ])
  }
  return () => h('span', { class: 'rs-search-form__label-tips' }, [tips as VNode])
}

const onFieldUpdate = (field: RsSearchField, value: unknown) => {
  formData.value[field.field] = value
  emit('field-change', field.field, value)
}

const setFieldValue = (name: RsFormNamePath, value: unknown) => {
  setByNamePath(formData.value, name, value)
}

const customRenderContext = (field: RsSearchField): RsSearchFormRenderContext => ({
  value: formData.value[field.field],
  onUpdate: (value) => onFieldUpdate(field, value),
  setFieldValue,
})

const renderCustomField = (field: RsSearchField) =>
  field.render?.(formData.value, customRenderContext(field))

const toolbarButtonsComputed = computed<ToolbarButton[]>(() => {
  const buttons: ToolbarButton[] = []

  if (props.toolbarButtons && props.toolbarButtons.length > 0) {
    buttons.push(...props.toolbarButtons)
  }

  // 查询/重置走 Toolbar 权限：`{moduleId}:search` / `{moduleId}:{resetButtonKey}`
  if (props.showSearchButton) {
    const hasSearchButton = buttons.some((btn) => btn.key === 'search')
    if (!hasSearchButton) {
      buttons.push({
        key: 'search',
        label: props.searchButtonText,
        icon: SearchOutline,
        type: 'primary',
        onClick: handleSearch,
      })
    }
  }

  const resetButtonKey = props.resetButtonKey || 'reset'
  if (props.showResetButton) {
    const hasResetButton = buttons.some((btn) => btn.key === resetButtonKey)
    if (!hasResetButton) {
      buttons.push({
        key: resetButtonKey,
        label: props.resetButtonText,
        icon: RefreshOutline,
        onClick: handleReset,
      })
    }
  }

  if (hasMoreFields.value) {
    buttons.push({
      key: 'more',
      label: props.moreButtonText,
      icon: OptionsOutline,
      onClick: toggleMoreFields,
    })
  }

  return buttons
})

const handleSearch = async () => {
  try {
    const result = await formRef.value?.validate()
    if (result && !result.valid) {
      return
    }
    emit('search', { ...formData.value })
  } catch (error) {
    console.error('表单验证失败:', error)
  }
}

const handleReset = () => {
  initFormData()
  formRef.value?.clearValidation()
  emit('reset')
}

  const handleToolbarClick = (key: string) => {
  const matched = toolbarButtonsComputed.value.find((btn) => btn.key === key)
  // 已有 onClick 的按钮（search / 重置 / more）不再重复触发 toolbar-click
  if (matched?.onClick) {
    return
  }
  emit('toolbar-click', key, { ...formData.value })
}

const getFormRef = () => formRef.value
const getFormData = () => formData.value
const setFormData = (data: Record<string, any>) => {
  formData.value = { ...formData.value, ...data }
}
const resetForm = handleReset
const validate = async () => {
  const result = await formRef.value?.validate()
  if (result && !result.valid) {
    throw new Error('表单验证失败')
  }
}
const submit = handleSearch

initFormData()
onMounted(() => {
  // 字段异步注入时再兜底一次
  if (!Object.keys(formData.value).length) {
    initFormData()
  }
})

defineExpose<RsSearchFormExpose>({
  getFormRef,
  getFormData,
  setFormData,
  resetForm,
  validate,
  submit,
  toggleMoreFields,
})
</script>

<style lang="scss" scoped>
.rs-search-form {
  width: 100%;
  background-color: var(--rs-surface, var(--g-bg-primary));
  display: flex;
  flex-direction: column;
  box-sizing: border-box;

  &__toolbar {
    flex-shrink: 0;
    padding: 0 var(--rs-space-md, 12px);
  }

  &__form {
    width: 100%;
    flex: 1;
  }

  &__grid {
    width: 100%;
    display: grid;
    padding: var(--rs-space-sm, 8px) var(--rs-space-md, 12px);
    box-sizing: border-box;
    align-items: start;
  }

  &__cell {
    min-width: 0;
  }

  &__item {
    width: 100%;
    min-width: 0;
  }

  &__label-text {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    max-width: 100%;
    min-width: 0;
  }

  &__label-tips {
    display: inline-flex;
    align-items: center;
    flex-shrink: 0;
  }
}
</style>
