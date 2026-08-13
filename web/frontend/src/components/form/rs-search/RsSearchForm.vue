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
          <!-- input / number / date：控件自带 label -->
          <template v-if="usesBuiltInLabel(field.type)">
            <RsInput
              v-if="field.type === 'input' || !field.type"
              v-model="formData[field.field]"
              :name="field.field"
              :label="getFieldLabel(field)"
              :hint="getStringHint(field)"
              :placeholder="field.placeholder || `请输入${getFieldLabel(field)}`"
              :disabled="field.disabled"
              :clearable="field.clearable !== false"
              :required="field.required"
              :size="rsSize"
              v-bind="field.props"
              @update:model-value="handleFieldChange(field.field, $event)"
            />

            <RsInputNumber
              v-else-if="field.type === 'number'"
              v-model="formData[field.field]"
              :name="field.field"
              :label="getFieldLabel(field)"
              :hint="getStringHint(field)"
              :placeholder="field.placeholder || `请输入${getFieldLabel(field)}`"
              :disabled="field.disabled"
              :required="field.required"
              :size="rsSize"
              v-bind="field.props"
              @update:model-value="handleFieldChange(field.field, $event)"
            />

            <RsDatePicker
              v-else-if="isDateField(field.type)"
              v-model="formData[field.field]"
              :name="field.field"
              :label="getFieldLabel(field)"
              :hint="getStringHint(field)"
              :placeholder="field.placeholder || `请选择${getFieldLabel(field)}`"
              :disabled="field.disabled"
              :required="field.required"
              :range="isDateRangeField(field.type)"
              :label-position="rsLabelPosition"
              :size="rsSize"
              v-bind="field.props"
              :with-time="isDateTimeField(field.type)"
              :with-seconds="isDateTimeField(field.type)"
              @update:model-value="handleFieldChange(field.field, $event)"
            />
          </template>

          <!-- select / switch / custom：RsLabel + 控件 -->
          <div
            v-else
            class="rs-search-form__field"
            :class="`rs-search-form__field--${labelPlacement}`"
            :style="externalFieldStyle"
          >
            <RsLabel
              class="rs-search-form__label"
              :required="field.required"
              :disabled="field.disabled"
              :nowrap="labelPlacement === 'left'"
            >
              <span class="rs-search-form__label-text" :title="getFieldLabel(field)">
                <template v-if="labelPlacement === 'left'">
                  {{ getFieldLabel(field) }}
                </template>
                <GEllipsis v-else :text="getFieldLabel(field)" />
                <component
                  :is="renderFieldTips(field)"
                  v-if="resolveTips(field)"
                />
              </span>
            </RsLabel>

            <div class="rs-search-form__control">
              <RsSelect
                v-if="field.type === 'select'"
                v-model="formData[field.field]"
                :name="field.field"
                :placeholder="field.placeholder || `请选择${getFieldLabel(field)}`"
                :options="normalizeSelectOptions(field.options)"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                :required="field.required"
                :size="rsSize"
                match-trigger-width
                v-bind="field.props"
                block
                @update:model-value="handleFieldChange(field.field, $event)"
              />

              <RsSwitch
                v-else-if="field.type === 'switch'"
                v-model="formData[field.field]"
                :disabled="field.disabled"
                :size="rsSize"
                v-bind="field.props"
                @update:model-value="handleFieldChange(field.field, $event)"
              />

              <component
                :is="field.render?.(formData)"
                v-else-if="field.type === 'custom' && field.render"
              />
            </div>
          </div>
        </div>
      </div>
    </RsForm>
  </div>
</template>

<script setup lang="ts">
import GEllipsis from '@/components/gellipsis/GEllipsis.vue'
import type { ToolbarButton } from '@/components/toolbar'
import GToolbar from '@/components/toolbar/GToolbar.vue'
import {
  RsDatePicker,
  RsForm,
  RsInput,
  RsInputNumber,
  RsLabel,
  RsSelect,
  RsSwitch,
  RsTooltip,
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
} from './types'

defineOptions({
  name: 'RsSearchForm',
})

const props = withDefaults(defineProps<RsSearchFormProps>(), {
  /* 与 RsForm 默认一致；过短会导致左侧中文标签挤压 */
  labelWidth: '6rem',
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
  // 供自带 label 的控件（含 DatePicker）对齐标签宽度
  '--rs-field-label-width': rsLabelWidth.value,
  '--rs-search-control-height': `var(--rs-control-height-${rsSize.value})`,
}))

/** 外部 RsLabel 布局的标签宽度 / 对齐（与 niuma-ui .rs-field--label-left 同语义） */
const externalFieldStyle = computed(() => {
  if (props.labelPlacement === 'left') {
    return {
      gridTemplateColumns: `minmax(${rsLabelWidth.value}, max-content) minmax(0, 1fr)`,
      '--rs-search-label-align': props.labelAlign === 'right' ? 'end' : 'start',
    }
  }
  return {}
})

const usesBuiltInLabel = (type?: RsSearchFieldType) =>
  !type ||
  type === 'input' ||
  type === 'number' ||
  type === 'date' ||
  type === 'daterange' ||
  type === 'datetime' ||
  type === 'datetimerange'

const isDateField = (type?: RsSearchFieldType) =>
  type === 'date' || type === 'daterange' || type === 'datetime' || type === 'datetimerange'

const isDateRangeField = (type?: RsSearchFieldType) =>
  type === 'daterange' || type === 'datetimerange'

const isDateTimeField = (type?: RsSearchFieldType) =>
  type === 'datetime' || type === 'datetimerange'

/**
 * 将业务 options 规范化为 RsSelect 所需的 string value。
 */
const normalizeSelectOptions = (
  options?: RsSearchField['options'],
): RsSelectOptions => {
  if (!options?.length) return []
  return options.map((opt) => ({
    label: opt.label,
    value: String(opt.value),
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

/** 初始化表单数据（只使用字段的 defaultValue） */
const initFormData = () => {
  const data: Record<string, any> = {}
  const applyDefaults = (fields: RsSearchField[]) => {
    fields.forEach((field) => {
      if (field.defaultValue !== undefined) {
        // select 的 value 统一为 string，与 RsSelect 对齐
        data[field.field] =
          field.type === 'select' && field.defaultValue !== null
            ? String(field.defaultValue)
            : field.defaultValue
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
          data[field.field] = { start: '', end: '' }
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

/**
 * 解析 tips（支持函数）。
 * tips 联合类型含 Vue Component 构造函数，`typeof === 'function'` 无法收窄为可调用工厂，
 * 因此对「按 formData 生成 tips」的分支显式断言后再调用。
 */
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

/** 内置 label 控件的字符串 tip → hint */
const getStringHint = (field: RsSearchField): string | undefined => {
  const tips = resolveTips(field)
  return typeof tips === 'string' ? tips : undefined
}

/** select/switch/custom：字符串 tips 用 RsTooltip icon，其它直接渲染 */
const renderFieldTips = (field: RsSearchField) => {
  const tips = resolveTips(field)
  if (tips == null) return () => null
  if (typeof tips === 'string') {
    return () =>
      h('span', { class: 'rs-search-form__label-tips' }, [
        h(RsTooltip, { content: tips, icon: true })
      ])
  }
  return () => h('span', { class: 'rs-search-form__label-tips' }, [tips as any])
}

const toolbarButtonsComputed = computed<ToolbarButton[]>(() => {
  const buttons: ToolbarButton[] = []

  if (props.toolbarButtons && props.toolbarButtons.length > 0) {
    buttons.push(...props.toolbarButtons)
  }

  // 查询/重置通过 onClick 绑定，保证权限拦截下仍可点击（ToolbarButton 对 search/reset/more 放行）
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

  if (props.showResetButton) {
    const hasResetButton = buttons.some((btn) => btn.key === 'reset')
    if (!hasResetButton) {
      buttons.push({
        key: 'reset',
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

const handleFieldChange = (field: string, value: any) => {
  emit('field-change', field, value)
}

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
}

const handleToolbarClick = (key: string) => {
  const button = toolbarButtonsComputed.value.find((btn) => btn.key === key)
  // 已有 onClick 的按钮（search/reset/more）不再重复触发 toolbar-click
  if (button?.onClick) {
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
  /* 内容面：走已桥接的 --rs-surface（勿用 --rs-bg 页面底，否则与 MainLayout 糊成一片） */
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
    /* 表单字段区上下边距一致（勿用根 gap，否则只垫在工具栏下方） */
    padding: var(--rs-space-sm, 8px) var(--rs-space-md, 12px);
    box-sizing: border-box;
  }

  &__cell {
    min-width: 0;
  }

  &__field {
    width: 100%;
    min-width: 0;

    &--left {
      display: grid;
      align-items: center;
      column-gap: var(--rs-space-sm, 12px);

      .rs-search-form__label {
        margin: 0;
        justify-self: var(--rs-search-label-align, end);
        text-align: var(--rs-search-label-align, end);
      }
    }

    &--top {
      display: flex;
      flex-direction: column;
      align-items: stretch;
      /* 与 .rs-field gap 一致；标签外观由全局 --rs-label-* token 统一 */
      gap: var(--rs-space-xs, 4px);
      width: 100%;

      .rs-search-form__label {
        margin: 0;
      }
    }
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

  &__control {
    width: 100%;
    min-width: 0;
    /*
     * 用 grid 拉伸子项：flex 下 Select（inline-flex）会按内容收缩，
     * 导致顶栏 label 时「类型」比同列 Input/DatePicker 窄、整行视觉不齐。
     */
    display: grid;
    align-items: center;
    justify-items: stretch;
    min-height: var(--rs-search-control-height, var(--rs-control-height-sm));
  }
}
</style>
