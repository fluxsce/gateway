<template>
  <div
    class="rs-data-form__grid"
    :style="gridStyle"
  >
    <template
      v-for="field in fields"
      :key="field.field"
    >
      <!-- fieldset：分组容器（递归渲染 children） -->
      <div
        v-if="field.type === 'fieldset' && isFieldVisible(field)"
        class="rs-data-form__cell rs-data-form__cell--fieldset"
        style="grid-column: span 24"
      >
        <GFieldset
          :title="getFieldLabel(field)"
          :title-strong="field.props?.titleStrong ?? false"
          :title-size="field.props?.titleSize ?? 'normal'"
          :border-style="field.props?.borderStyle ?? 'dashed'"
          :selected="field.props?.selected ?? false"
          :disabled="field.disabled ?? false"
        >
          <RsDataFormFields
            :fields="field.children || []"
            :form-model="formModel"
            :mode="mode"
            :label-width="labelWidth"
            :label-placement="labelPlacement"
            :label-align="labelAlign"
            :size="size"
            :cols="cols"
            :x-gap="xGap"
            :y-gap="yGap"
            @field-update="(f, v) => emit('field-update', f, v)"
          />
        </GFieldset>
      </div>

      <!--
        普通字段：统一「外部 label + 控件」结构（对齐旧 GDataFormModal / Naive）。
        不再混用控件内置 label，避免 Input/Select 行高、间距不一致。
      -->
      <div
        v-else-if="isFieldVisible(field)"
        class="rs-data-form__cell"
        :style="{ gridColumn: `span ${field.span ?? 12}` }"
      >
        <div
          class="rs-data-form__field"
          :class="`rs-data-form__field--${labelPlacement}`"
          :style="externalFieldStyle"
        >
          <RsLabel
            class="rs-data-form__label"
            :for-id="controlId(field)"
            :required="field.required"
          >
            <span class="rs-data-form__label-text">
              <GEllipsis :text="getFieldLabel(field)" />
              <component
                :is="renderFieldTips(field)"
                v-if="resolveTips(field)"
              />
            </span>
          </RsLabel>

          <div class="rs-data-form__control">
            <RsInput
              v-if="field.type === 'input' || !field.type"
              :id="controlId(field)"
              :model-value="asStringValue(formModel[field.field])"
              :name="field.field"
              :placeholder="field.placeholder || `请输入${getFieldLabel(field)}`"
              :disabled="isFieldDisabled(field)"
              :clearable="field.clearable !== false"
              :required="field.required"
              :size="rsSize"
              v-bind="getInputExtraProps(field)"
              @update:model-value="(v: string) => emit('field-update', field, v)"
            />

            <RsInputNumber
              v-else-if="field.type === 'number'"
              :id="controlId(field)"
              :model-value="formModel[field.field]"
              :name="field.field"
              :placeholder="field.placeholder || `请输入${getFieldLabel(field)}`"
              :disabled="isFieldDisabled(field)"
              :required="field.required"
              :size="rsSize"
              v-bind="omitControlProps(field.props)"
              @update:model-value="(v) => onNumberUpdate(field, v)"
            />

            <RsDatePicker
              v-else-if="isDateField(field.type)"
              :id="controlId(field)"
              :name="field.field"
              :model-value="formModel[field.field]"
              :placeholder="field.placeholder || `请选择${getFieldLabel(field)}`"
              :disabled="isFieldDisabled(field)"
              :required="field.required"
              :range="isDateRangeField(field.type)"
              :size="rsSize"
              value-format="string"
              v-bind="omitControlProps(field.props)"
              :with-time="isDateTimeField(field.type)"
              :with-seconds="isDateTimeField(field.type)"
              @update:model-value="(v: any) => emit('field-update', field, v)"
            />

            <textarea
              v-else-if="field.type === 'textarea'"
              :id="controlId(field)"
              class="rs-data-form__textarea"
              :name="field.field"
              :value="asStringValue(formModel[field.field])"
              :placeholder="field.placeholder || `请输入${getFieldLabel(field)}`"
              :disabled="isFieldDisabled(field)"
              :rows="field.props?.rows ?? 3"
              v-bind="omitControlProps(field.props)"
              @input="onTextareaInput(field, $event)"
            />

            <RsSelect
              v-else-if="field.type === 'select'"
              :id="controlId(field)"
              :model-value="normalizeSelectModel(field)"
              :name="field.field"
              :placeholder="field.placeholder || `请选择${getFieldLabel(field)}`"
              :options="normalizeSelectOptions(field.options)"
              :disabled="isFieldDisabled(field)"
              :clearable="field.clearable !== false"
              :required="field.required"
              :size="rsSize"
              block
              match-trigger-width
              v-bind="omitControlProps(field.props)"
              @update:model-value="onSelectUpdate(field, $event)"
            />

            <RsSwitch
              v-else-if="field.type === 'switch'"
              :id="controlId(field)"
              :model-value="getSwitchBoolean(field)"
              :disabled="isFieldDisabled(field)"
              :size="rsSize"
              v-bind="omitSwitchProps(field.props)"
              @update:model-value="(v: boolean) => emit('field-update', field, fromSwitchBoolean(field, v))"
            />

            <RsUpload
              v-else-if="field.type === 'file'"
              :model-value="(formModel[field.field] as File[]) ?? []"
              :disabled="isFieldDisabled(field)"
              :accept="field.props?.config?.accept ?? field.props?.accept"
              :max-count="field.props?.config?.max ?? field.props?.maxCount ?? 1"
              :max-size="field.props?.config?.maxSize ?? field.props?.maxSize"
              :label="field.props?.config?.uploadText ?? field.props?.label"
              :hint="field.props?.config?.uploadDescription ?? field.props?.hint"
              :show-file-list="field.props?.config?.showFileList !== false"
              :show-download="field.props?.showDownload ?? false"
              :hide-dropzone-when-full="(field.props?.config?.max ?? field.props?.maxCount ?? 1) === 1"
              @update:model-value="(v: File[]) => emit('field-update', field, v)"
              @reject="(errors) => onFileReject(field, errors)"
            />

            <component
              :is="field.render?.(formModel)"
              v-else-if="field.type === 'custom' && field.render"
            />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import GEllipsis from '@/components/gellipsis/GEllipsis.vue'
import { GFieldset } from '@/components/gfieldset'
import { useAppMessage } from '@/composables/useAppMessage'
import {
  RsDatePicker,
  RsInput,
  RsInputNumber,
  RsLabel,
  RsSelect,
  RsSwitch,
  RsTooltip,
  RsUpload,
} from '@/ui'
import type { RsInputNumberValue, RsSelectOptions } from '@/ui'
import { computed, h, type Component, type VNode } from 'vue'
import type {
  RsDataFormField,
  RsDataFormFieldsEmits,
  RsDataFormFieldsProps,
  RsDataFormFieldType,
} from './types'

defineOptions({
  name: 'RsDataFormFields',
})

const props = withDefaults(defineProps<RsDataFormFieldsProps>(), {
  mode: 'create',
  labelWidth: 'auto',
  labelPlacement: 'top',
  labelAlign: 'left',
  size: 'small',
  cols: 24,
  xGap: 16,
  yGap: 8,
})

const emit = defineEmits<RsDataFormFieldsEmits>()
const message = useAppMessage()

const onFileReject = (
  field: RsDataFormField,
  errors: { reason: 'accept' | 'maxSize' | 'maxCount' }[],
) => {
  const first = errors[0]
  if (!first) return
  const max = field.props?.config?.max ?? field.props?.maxCount ?? 1
  const maxSize = field.props?.config?.maxSize ?? field.props?.maxSize ?? 0
  let text = `最多上传 ${max} 个文件`
  if (first.reason === 'accept') text = '文件类型不符合要求'
  else if (first.reason === 'maxSize') text = `文件大小不能超过 ${(maxSize / 1024 / 1024).toFixed(2)}MB`
  message.error(text)
}

/** 业务 size → niuma-ui 尺寸 */
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

const rsLabelWidth = computed(() =>
  typeof props.labelWidth === 'number' ? `${props.labelWidth}px` : String(props.labelWidth),
)

/** CSS Grid 布局变量 */
const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${props.cols}, minmax(0, 1fr))`,
  columnGap: `${props.xGap}px`,
  rowGap: `${props.yGap}px`,
}))

/** 左标签时固定标签列宽，保证同列输入框左缘对齐 */
const externalFieldStyle = computed(() => {
  if (props.labelPlacement === 'left') {
    const width = rsLabelWidth.value === 'auto' ? '96px' : rsLabelWidth.value
    return {
      gridTemplateColumns: `${width} minmax(0, 1fr)`,
      '--rs-data-label-align': props.labelAlign === 'right' ? 'end' : 'start',
    }
  }
  return {}
})

const controlId = (field: RsDataFormField) => `rs-data-field-${field.field}`

const isDateField = (type?: RsDataFormFieldType) =>
  type === 'date' || type === 'daterange' || type === 'datetime' || type === 'datetimerange'

const isDateRangeField = (type?: RsDataFormFieldType) =>
  type === 'daterange' || type === 'datetimerange'

const isDateTimeField = (type?: RsDataFormFieldType) =>
  type === 'datetime' || type === 'datetimerange'

const asStringValue = (value: unknown): string => {
  if (value === null || value === undefined) return ''
  return String(value)
}

const getFieldLabel = (field: RsDataFormField): string =>
  typeof field.label === 'function' ? field.label(props.formModel) : field.label

const isFieldVisible = (field: RsDataFormField): boolean => {
  if (typeof field.show === 'function') return field.show(props.formModel)
  return field.show !== false
}

/**
 * 控件禁用：view 全禁；edit 时主键禁；其余取字段 disabled。
 * 仅用于输入控件，不要传给 label / fieldset（避免整块界面变灰）。
 */
const isFieldDisabled = (field: RsDataFormField): boolean => {
  if (props.mode === 'view') return true
  if (props.mode === 'edit' && field.primary === true) return true
  return field.disabled ?? false
}

/**
 * 解析 tips（支持函数）。
 * tips 联合类型含 Vue Component 构造函数，`typeof === 'function'` 无法收窄为可调用工厂。
 */
const resolveTips = (field: RsDataFormField): string | Component | VNode | null => {
  if (!field.tips) return null
  if (typeof field.tips === 'function') {
    const tipsFactory = field.tips as (
      data: Record<string, any>,
    ) => string | Component | VNode
    return tipsFactory(props.formModel)
  }
  return field.tips
}

/** 字符串 tips 用 RsTooltip icon 挂在标签旁（不占控件下方行，避免栅格行高错乱） */
const renderFieldTips = (field: RsDataFormField) => {
  const tips = resolveTips(field)
  if (tips == null) return () => null
  if (typeof tips === 'string') {
    return () =>
      h('span', { class: 'rs-data-form__label-tips' }, [
        h(RsTooltip, { content: tips, icon: true }),
      ])
  }
  return () => h('span', { class: 'rs-data-form__label-tips' }, [tips as any])
}

/**
 * 过滤不应直接透传给控件的业务属性
 * （switch 映射值、file 专用 props、fieldset 标题 props、旧 Naive 密码配置等）
 */
const omitControlProps = (raw?: Record<string, any>): Record<string, any> => {
  if (!raw) return {}
  const {
    onUpdateValue,
    checkedValue,
    uncheckedValue,
    showPasswordOn,
    config,
    title,
    titleIcon,
    titleIconColor,
    showDownload,
    downloadText,
    callbacks,
    titleStrong,
    titleSize,
    borderStyle,
    selected,
    rows,
    ...rest
  } = raw
  return rest
}

/** 兼容旧 Naive 密码显隐配置 → RsInput visibilityToggle */
const getInputExtraProps = (field: RsDataFormField): Record<string, any> => {
  const extra = omitControlProps(field.props)
  if (field.props?.type === 'password' || field.props?.showPasswordOn) {
    return {
      ...extra,
      type: 'password' as const,
      visibilityToggle: extra.visibilityToggle ?? true,
    }
  }
  return extra
}

const omitSwitchProps = (raw?: Record<string, any>) => omitControlProps(raw)

/** options.value 统一为 string，匹配 RsSelect */
const normalizeSelectOptions = (options?: RsDataFormField['options']): RsSelectOptions => {
  if (!options?.length) return []
  return options.map((opt) => ({
    label: opt.label,
    value: String(opt.value),
    disabled: opt.disabled,
  }))
}

const normalizeSelectModel = (field: RsDataFormField): string => {
  const value = props.formModel[field.field]
  if (value === null || value === undefined) return ''
  return String(value)
}

/** 将 RsSelect 的 string 值还原为配置中的原始类型 */
const coerceSelectValue = (field: RsDataFormField, value: string): string | number => {
  if (value === '' || value == null) return value
  const matched = field.options?.find((opt) => String(opt.value) === String(value))
  return matched ? matched.value : value
}

/** RsSelect 可能回传 string | string[]（多选场景），单选取首项 */
const onSelectUpdate = (field: RsDataFormField, value: string | string[]) => {
  const next = Array.isArray(value) ? (value[0] ?? '') : value
  emit('field-update', field, coerceSelectValue(field, next))
}

const onTextareaInput = (field: RsDataFormField, event: Event) => {
  const target = event.target as HTMLTextAreaElement | null
  emit('field-update', field, target?.value ?? '')
}

const onNumberUpdate = (field: RsDataFormField, value: RsInputNumberValue) => {
  emit('field-update', field, value)
}

/** switch 显示值：支持 props.checkedValue / uncheckedValue 映射 */
const getSwitchBoolean = (field: RsDataFormField): boolean => {
  const checked = field.props?.checkedValue ?? true
  return props.formModel[field.field] === checked
}

const fromSwitchBoolean = (field: RsDataFormField, value: boolean) => {
  const checked = field.props?.checkedValue ?? true
  const unchecked = field.props?.uncheckedValue ?? false
  return value ? checked : unchecked
}
</script>

<style scoped lang="scss">
.rs-data-form__grid {
  display: grid;
  width: 100%;
  align-items: start;
}

.rs-data-form__cell {
  min-width: 0;
}

.rs-data-form__cell--fieldset {
  margin-top: 2px;
}

.rs-data-form__field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  width: 100%;
  min-width: 0;
}

.rs-data-form__field--left {
  display: grid;
  align-items: center;
  column-gap: 0.5rem;
}

.rs-data-form__label {
  margin: 0;
  min-width: 0;
  /* 对齐旧 GDataFormModal 标签观感 */
  --rs-label-font-size: var(--rs-font-size-sm, 13px);
  --rs-label-font-weight: 400;
  --rs-label-color: var(--g-text-secondary, var(--rs-muted));
  --rs-label-line-height: 1.4;
}

.rs-data-form__field--left .rs-data-form__label {
  justify-content: var(--rs-data-label-align, start);
  text-align: var(--rs-data-label-align, start);
}

.rs-data-form__label-text {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 100%;
  min-width: 0;
}

.rs-data-form__label-tips {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
}

.rs-data-form__control {
  min-width: 0;
  width: 100%;
  display: grid;
  align-items: center;
  justify-items: stretch;
}

.rs-data-form__textarea {
  box-sizing: border-box;
  width: 100%;
  min-height: 4.5rem;
  padding: 0.4rem 0.6rem;
  border: 1px solid var(--rs-input-border, var(--rs-border, #d0d5dd));
  border-radius: var(--rs-radius-sm, 6px);
  background: var(--rs-input-bg, var(--rs-surface, #fff));
  color: var(--rs-text, inherit);
  font: inherit;
  font-size: var(--rs-font-size-sm, 13px);
  line-height: 1.5;
  resize: vertical;
  outline: none;
  transition: border-color var(--rs-transition-fast, 0.15s ease);

  &:hover:not(:disabled) {
    border-color: var(--rs-input-border-hover, var(--rs-border));
  }

  &:focus:not(:disabled) {
    border-color: var(--rs-focus-border, var(--rs-primary));
  }

  &:disabled {
    cursor: not-allowed;
    opacity: 0.55;
    background: var(--rs-surface-hover, #f2f4f7);
  }

  &::placeholder {
    color: var(--rs-placeholder, #c0c4cc);
  }
}
</style>
