<template>
  <!-- 扁平非弹窗的数据表单组件 -->
  <div class="g-data-form">
    <n-form
      ref="formRef"
      :model="formModel"
      :rules="formRules"
      label-placement="left"
      label-width="auto"
      size="small"
    >
      <!-- 多页签布局：如果有超过 1 个页签，则显示 NTabs，否则保持单页 -->
      <n-tabs
        v-if="tabs.length > 1"
        v-model:value="activeTab"
        type="line"
        size="small"
        class="g-data-form__tabs"
      >
        <n-tab-pane
          v-for="tab in tabs"
          :key="tab.key"
          :name="tab.key"
          :tab="tab.label"
          style="padding: var(--n-pane-padding-top) var(--n-pane-padding-right) var(--n-pane-padding-bottom) var(--n-pane-padding-left) !important;"
        >
          <n-grid :cols="24" :x-gap="16" :y-gap="8">
            <template v-for="field in getFieldsByTab(tab.key)" :key="field.field">
              <!-- fieldset 类型：使用 GFieldset 包裹子字段 -->
              <template v-if="field.type === 'fieldset' && (typeof field.show === 'function' ? field.show(formModel) : (field.show !== false))">
                <n-grid-item :span="24">
                  <GFieldset
                    :title="getFieldLabel(field)"
                    :title-strong="field.props?.titleStrong ?? false"
                    :title-size="field.props?.titleSize ?? 300"
                    :border-style="field.props?.borderStyle"
                    :selected="field.props?.selected ?? false"
                    :disabled="field.disabled ?? false"
                  >
                    <n-grid :cols="24" :x-gap="16" :y-gap="8">
                      <template v-for="child in field.children" :key="child.field">
                        <n-grid-item v-if="typeof child.show === 'function' ? child.show(formModel) : (child.show !== false)" :span="child.span ?? 12">
                          <n-form-item :path="child.field" style="width: 100%;">
                            <template #label>
                              <component :is="renderLabel(child)" />
                            </template>
                            <!-- 自定义渲染 -->
                            <component
                              v-if="child.type === 'custom' && child.render"
                              :is="child.render(formModel)"
                            />
                            <!-- 文件上传类型：使用 v-model:fileList -->
                            <component
                              v-else-if="(child.type as string) === 'file'"
                              :is="getFieldComponent(child)"
                              v-bind="getFieldComponentProps(child)"
                              v-model:file-list="formModel[child.field]"
                            />
                            <!-- 内置渲染 -->
                            <component
                              v-else
                              :is="getFieldComponent(child)"
                              v-bind="getFieldComponentProps(child)"
                              v-model:value="formModel[child.field]"
                              @update:value="(value: any) => handleFieldValueUpdate(child, value)"
                            />
                          </n-form-item>
                        </n-grid-item>
                      </template>
                    </n-grid>
                  </GFieldset>
                </n-grid-item>
              </template>

              <!-- 普通字段：直接渲染 -->
              <template v-else>
                <n-grid-item v-if="typeof field.show === 'function' ? field.show(formModel) : (field.show !== false)" :span="field.span ?? 12">
                  <n-form-item :path="field.field" style="width: 100%;">
                    <template #label>
                      <component :is="renderLabel(field)" />
                    </template>
                    <!-- 自定义渲染 -->
                    <component
                      v-if="field.type === 'custom' && field.render"
                      :is="field.render(formModel)"
                    />
                    <!-- 文件上传类型：使用 v-model:fileList -->
                    <component
                      v-else-if="(field.type as string) === 'file'"
                      :is="getFieldComponent(field)"
                      v-bind="getFieldComponentProps(field)"
                      v-model:file-list="formModel[field.field]"
                    />
                    <!-- 内置渲染 -->
                    <component
                      v-else
                      :is="getFieldComponent(field)"
                      v-bind="getFieldComponentProps(field)"
                      v-model:value="formModel[field.field]"
                      @update:value="(value: any) => handleFieldValueUpdate(field, value)"
                    />
                  </n-form-item>
                </n-grid-item>
              </template>
            </template>
          </n-grid>
        </n-tab-pane>
      </n-tabs>

      <!-- 单页布局：保持原有行为，方便兼容旧代码 -->
      <template v-else>
        <n-grid :cols="24" :x-gap="16" :y-gap="8">
          <template v-for="field in props.formFields" :key="field.field">
            <!-- fieldset 类型：使用 GFieldset 包裹子字段 -->
            <template v-if="field.type === 'fieldset' && (typeof field.show === 'function' ? field.show(formModel) : (field.show !== false))">
              <n-grid-item :span="24">
                <GFieldset
                  :title="getFieldLabel(field)"
                  :title-strong="field.props?.titleStrong ?? false"
                  :title-size="field.props?.titleSize ?? 'normal'"
                  :border-style="field.props?.borderStyle"
                  :selected="field.props?.selected ?? false"
                  :disabled="field.disabled ?? false"
                >
                  <n-grid :cols="24" :x-gap="16" :y-gap="8">
                    <template v-for="child in field.children" :key="child.field">
                      <n-grid-item v-if="typeof child.show === 'function' ? child.show(formModel) : (child.show !== false)" :span="child.span ?? 12">
                        <n-form-item :path="child.field">
                          <template #label>
                            <component :is="renderLabel(child)" />
                          </template>
                          <!-- 自定义渲染 -->
                          <component
                            v-if="child.type === 'custom' && child.render"
                            :is="child.render(formModel)"
                          />
                          <!-- 文件上传类型：使用 v-model:fileList -->
                          <component
                            v-else-if="(child.type as string) === 'file'"
                            :is="getFieldComponent(child)"
                            v-bind="getFieldComponentProps(child)"
                            v-model:file-list="formModel[child.field]"
                          />
                          <!-- 内置渲染 -->
                          <component
                            v-else
                            :is="getFieldComponent(child)"
                            v-bind="getFieldComponentProps(child)"
                            v-model:value="formModel[child.field]"
                            @update:value="(value: any) => handleFieldValueUpdate(child, value)"
                          />
                        </n-form-item>
                      </n-grid-item>
                    </template>
                  </n-grid>
                </GFieldset>
              </n-grid-item>
            </template>

            <!-- 普通字段：直接渲染 -->
            <template v-else>
              <n-grid-item v-if="typeof field.show === 'function' ? field.show(formModel) : (field.show !== false)" :span="field.span ?? 12">
                <n-form-item :path="field.field">
                  <template #label>
                    <component :is="renderLabel(field)" />
                  </template>
                  <!-- 自定义渲染 -->
                  <component
                    v-if="field.type === 'custom' && field.render"
                    :is="field.render(formModel)"
                  />
                  <!-- 文件上传类型：使用 v-model:fileList -->
                  <component
                    v-else-if="(field.type as string) === 'file'"
                    :is="getFieldComponent(field)"
                    v-bind="getFieldComponentProps(field)"
                    v-model:file-list="formModel[field.field]"
                  />
                  <!-- 内置渲染 -->
                  <component
                    v-else
                    :is="getFieldComponent(field)"
                    v-bind="getFieldComponentProps(field)"
                    v-model:value="formModel[field.field]"
                    @update:value="(value: any) => handleFieldValueUpdate(field, value)"
                  />
                </n-form-item>
              </n-grid-item>
            </template>
          </template>
        </n-grid>
      </template>
    </n-form>

    <!-- 底部操作区：可选显示提交按钮 -->
    <div v-if="showFooter" class="g-data-form__footer">
      <slot name="footer" :form-data="formModel" :form-ref="formRef" :on-submit="handleSubmit">
        <n-space justify="end" :size="8">
          <n-button
            v-if="showSubmit"
            type="primary"
            size="small"
            :loading="submitLoading"
            @click="handleSubmit"
          >
            {{ submitText }}
          </n-button>
        </n-space>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { GDate, GFieldset, GFileUpload, GTips } from '@/components'
import type { FormInst, FormRules } from 'naive-ui'
import {
    NButton,
    NForm,
    NFormItem,
    NGrid,
    NGridItem,
    NInput,
    NInputNumber,
    NSelect,
    NSpace,
    NSwitch,
    NTabPane,
    NTabs
} from 'naive-ui'
import type { Component, VNode } from 'vue'
import { computed, h, ref, watch } from 'vue'
import type { DataFormField, DataFormTab } from './types'

defineOptions({
  name: 'GDataForm'
})

// ============= Props =============

interface Props {
  /**
   * 当前业务模式：
   * - 'create'：新增
   * - 'edit'：编辑
   * - 'view'：查看详情（通常会禁用表单，仅展示）
   * @default 'create'
   */
  mode?: 'create' | 'edit' | 'view'

  /**
   * 表单字段配置列表
   * 支持嵌套结构，type 为 'fieldset' 的字段可以包含 children
   */
  formFields?: DataFormField[]

  /**
   * 表单页签配置
   * - 如果不配置，则不显示页签，所有字段在同一页
   * - 如果配置，则会根据 DataFormField.tabKey 将字段分配到对应页签
   */
  formTabs?: DataFormTab[]

  /**
   * 初始表单数据（用于编辑模式，传入要编辑的数据）
   */
  initialData?: Record<string, any>

  /**
   * 是否显示底部操作区
   * @default false
   */
  showFooter?: boolean

  /**
   * 是否显示提交按钮
   * @default false
   */
  showSubmit?: boolean

  /**
   * 提交按钮文本
   * @default '保存'
   */
  submitText?: string

  /**
   * 提交按钮加载状态
   * @default false
   */
  submitLoading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  mode: 'create',
  formFields: () => [] as DataFormField[],
  formTabs: () => [] as DataFormTab[],
  initialData: () => ({}),
  showFooter: false,
  showSubmit: false,
  submitText: '保存',
  submitLoading: false,
})

// ============= Emits =============

interface Emits {
  /**
   * 表单提交事件（通常在点击提交按钮时触发）
   */
  (event: 'submit', formData?: Record<string, any>): void

  /**
   * 表单数据变化事件
   */
  (event: 'update:modelValue', formData: Record<string, any>): void

  /**
   * 需要重置内部表单时触发
   */
  (event: 'reset'): void
}

const emit = defineEmits<Emits>()

// ============= 表单引用与数据模型 =============

const formRef = ref<FormInst | null>(null)
const formModel = ref<Record<string, any>>({})

// ============= 页签相关 =============

// 真实使用的页签列表：
// 1. 如果显式传入 formTabs，则优先使用
// 2. 否则根据字段的 tabKey 去重生成
// 3. 如果仍然为空，则退化为单页签（不在视图中展示）
// 4. 根据 show 属性过滤页签（支持函数动态计算）
const tabs = computed<DataFormTab[]>(() => {
  let tabsList: DataFormTab[] = []

  if (props.formTabs && props.formTabs.length > 0) {
    tabsList = props.formTabs
  } else {
    const tabKeys = Array.from(
      new Set(
        (props.formFields || [])
          .map((field) => field.tabKey)
          .filter((key): key is string => !!key)
      )
    )

    if (tabKeys.length > 0) {
      tabsList = tabKeys.map((key) => ({ key, label: key }))
    } else {
      // 默认单页签，不显示标签，仅用于统一处理逻辑
      tabsList = [{ key: 'default', label: '表单' }]
    }
  }

  // 根据 show 属性过滤页签
  return tabsList.filter((tab) => {
    if (typeof tab.show === 'function') {
      return tab.show(formModel.value)
    }
    return tab.show !== false
  })
})

// 当前激活的页签（内部状态）
const activeTab = ref<string>(tabs.value[0]?.key || 'default')

// 当 tabs 变化时，如果当前 activeTab 不在可见页签列表中，则切换到第一个可见页签
watch(
  tabs,
  (newTabs) => {
    if (newTabs.length > 0 && !newTabs.some((tab) => tab.key === activeTab.value)) {
      activeTab.value = newTabs[0]?.key || 'default'
    }
  },
  { immediate: false }
)

// 获取字段所属页签 key（未设置时归入第一个）
const getFieldTabKey = (field: DataFormField) => {
  return field.tabKey || tabs.value[0]?.key || 'default'
}

// 按页签过滤字段（递归处理嵌套结构）
const getFieldsByTab = (tabKey: string): DataFormField[] => {
  const filterFields = (fields: DataFormField[], parentTabKey?: string): DataFormField[] => {
    return fields
      .filter((field) => {
        // 如果字段有 tabKey，使用字段的 tabKey；否则使用父级的 tabKey
        const fieldTabKey = field.tabKey || parentTabKey || tabs.value[0]?.key || 'default'
        return fieldTabKey === tabKey
      })
      .map((field) => {
        // 如果是 fieldset 类型且有 children，递归处理 children
        // children 继承父级的 tabKey（如果 children 没有自己的 tabKey）
        if (field.type === 'fieldset' && field.children) {
          const currentTabKey = field.tabKey || parentTabKey || tabs.value[0]?.key || 'default'
          return {
            ...field,
            children: filterFields(field.children, currentTabKey)
          }
        }
        return field
      })
  }
  return filterFields(props.formFields)
}

// ============= 表单初始化 =============

// 递归处理字段的默认值初始化
const processFieldDefaultValue = (field: DataFormField, model: Record<string, any>, initialData: Record<string, any>) => {
  const key = field.field
  
  // 跳过 fieldset 类型本身（它只是容器，不存储数据）
  if (field.type === 'fieldset') {
    // 递归处理 children
    if (field.children) {
      field.children.forEach((child) => {
        processFieldDefaultValue(child, model, initialData)
      })
    }
    return
  }
  
  // 优先使用 initialData 中的值（编辑模式）
  // 只有当值不为 undefined 时才使用，否则使用 defaultValue
  if (Object.prototype.hasOwnProperty.call(initialData, key) && initialData[key] !== undefined) {
    model[key] = initialData[key]
  } else if (Object.prototype.hasOwnProperty.call(field, 'defaultValue')) {
    model[key] = field.defaultValue
  } else {
    // 针对不同类型设置更合理的默认值，避免组件类型校验告警
    switch (field.type) {
      case 'date':
      case 'datetime':
      case 'daterange':
      case 'datetimerange':
        // Naive UI NDatePicker 的 value 类型为 number | null | number[]
        model[key] = null
        break
      case 'number':
        model[key] = null
        break
      case 'switch':
        // 开关类如果没有 defaultValue，则默认 false
        model[key] = false
        break
      case 'file':
        // 文件上传类型默认使用空数组
        model[key] = []
        break
      default:
        // 其余输入类默认使用空字符串
        model[key] = model[key] ?? ''
    }
  }
}

// 根据 formFields 初始化表单数据
const initFormModel = () => {
  const model: Record<string, any> = {}
  
  // 如果传入了初始数据（编辑模式），优先使用初始数据
  // 使用深拷贝避免响应式引用问题，确保初始数据不会影响后续的表单编辑
  const initialData = props.initialData ? JSON.parse(JSON.stringify(props.initialData)) : {}
  
  // 递归处理所有字段（包括 fieldset 的 children）
  props.formFields.forEach((field) => {
    processFieldDefaultValue(field, model, initialData)
  })
  
  formModel.value = model
  formRef.value?.restoreValidation()
  
  // 触发更新事件
  emit('update:modelValue', formModel.value)
}

// ============= 表单验证 =============

// 递归处理字段验证规则（包括 fieldset 的 children）
const processFieldRules = (field: DataFormField, rules: FormRules) => {
  // 跳过 fieldset 类型本身（它只是容器，不存储数据）
  if (field.type === 'fieldset') {
    // 递归处理 children
    if (field.children) {
      field.children.forEach((child) => {
        processFieldRules(child, rules)
      })
    }
    return
  }

  // 处理字段的验证规则
  if (field.rules) {
    rules[field.field] = field.rules
  } else if (field.required) {
    const isSelectLike =
      field.type === 'select' ||
      field.type === 'date' ||
      field.type === 'datetime' ||
      field.type === 'daterange' ||
      field.type === 'datetimerange' ||
      field.type === 'switch'
    
    // 对于数字类型字段，需要特殊处理验证规则
    if (field.type === 'number') {
      rules[field.field] = [
        {
          required: true,
          type: 'number',
          message: `请输入${getFieldLabel(field)}`,
          trigger: 'blur',
          validator: (_rule: any, value: any) => {
            const label = getFieldLabel(field)
            // 允许 0 作为有效值，只检查是否为 null 或 undefined
            if (value === null || value === undefined || value === '') {
              return new Error(`请输入${label}`)
            }
            // 检查是否为有效数字
            const num = typeof value === 'number' ? value : Number(value)
            if (isNaN(num)) {
              return new Error(`${label}必须是数字`)
            }
            return true
          }
        }
      ]
    } else {
      rules[field.field] = [
        {
          required: true,
          message: `请输入${getFieldLabel(field)}`,
          trigger: isSelectLike ? 'change' : 'blur'
        }
      ]
    }
  }
}

// 根据字段配置生成表单校验规则
const formRules = computed<FormRules>(() => {
  const rules: FormRules = {}
  props.formFields.forEach((field) => {
    processFieldRules(field, rules)
  })
  return rules
})

// ============= 字段渲染 =============

// 获取字段 label（支持函数类型）
const getFieldLabel = (field: DataFormField): string => {
  return typeof field.label === 'function' 
    ? field.label(formModel.value) 
    : field.label
}

// 渲染带 tips 的 label
const renderLabel = (field: DataFormField) => {
  const labelText = getFieldLabel(field)
  
  if (!field.tips) {
    return () => labelText
  }
  
  return () => h('span', { class: 'g-form-label-with-tips' }, [
    labelText,
    h('span', { class: 'g-form-label-tips' }, [
      (() => {
        if (!field.tips) return null
        let tipsValue: string | Component | VNode
        if (typeof field.tips === 'function') {
          tipsValue = (field.tips as (formData: Record<string, any>) => string | Component | VNode)(formModel.value)
        } else {
          tipsValue = field.tips
        }
        return typeof tipsValue === 'string'
          ? h(GTips, { content: tipsValue })
          : (tipsValue as any)
      })()
    ])
  ])
}

// 根据字段类型选择对应组件
const getFieldComponent = (field: DataFormField) => {
  switch (field.type) {
    case 'select':
      return NSelect
    case 'number':
      return NInputNumber
    case 'switch':
      return NSwitch
    case 'date':
    case 'datetime':
    case 'daterange':
    case 'datetimerange':
      // 使用 GDate 组件，自动处理字符串和时间戳的转换
      return GDate
    case 'file':
      return GFileUpload
    case 'textarea':
      return NInput
    case 'input':
    default:
      return NInput
  }
}

// 组装组件属性
const getFieldComponentProps = (field: DataFormField) => {
  // 统一透传 disabled 等通用属性，再合并用户自定义 props
  // 在 view 模式下，自动禁用所有字段（除非字段本身已明确设置 disabled: false）
  // 在 edit 模式下，如果字段标记为 primary（主键），则自动禁用，防止修改
  const common = {
    disabled:
      props.mode === 'view'
        ? true
        : props.mode === 'edit' && field.primary === true
          ? true
          : field.disabled ?? false,
    ...(field.props ?? {})
  }
  switch (field.type) {
    case 'textarea':
      return {
        placeholder: field.placeholder,
        type: 'textarea',
        ...common
      }
    case 'select':
      return {
        placeholder: field.placeholder,
        options: field.options,
        clearable: field.clearable ?? true,
        ...common
      }
    case 'number':
      return {
        placeholder: field.placeholder,
        ...common
      }
    case 'switch':
      return {
        ...common
      }
    case 'date':
    case 'datetime':
    case 'daterange':
    case 'datetimerange':
      return {
        placeholder: field.placeholder,
        type: field.type,
        ...common
      }
    case 'file':
      // file 类型使用 GFileUpload 组件，callbacks 可直接在 props 中传入
      return {
        ...common,
        config: field.props?.config,
        title: field.props?.title,
        titleIcon: field.props?.titleIcon,
        titleIconColor: field.props?.titleIconColor,
        showDownload: field.props?.showDownload ?? false,
        downloadText: field.props?.downloadText,
        callbacks: field.props?.callbacks,
      }
    case 'input':
    default:
      return {
        placeholder: field.placeholder,
        clearable: field.clearable ?? true,
        ...common
      }
  }
}

// 处理字段值更新（支持字段的 onUpdateValue 回调）
const handleFieldValueUpdate = (field: DataFormField, value: any) => {
  // 更新表单模型的值
  formModel.value[field.field] = value
  
  // 如果字段配置了 onUpdateValue 回调，执行它
  if (field.props?.onUpdateValue && typeof field.props.onUpdateValue === 'function') {
    field.props.onUpdateValue(value, formModel.value)
  }
  
  // 触发更新事件
  emit('update:modelValue', formModel.value)
}

// ============= 表单提交 =============

// 提交前进行表单校验，并将表单数据透传给调用方
const handleSubmit = async () => {
  if (formRef.value) {
    try {
      await formRef.value.validate()
    } catch {
      // 校验失败，不提交
      return
    }
  }

  // 确保获取最新的表单数据
  const formData = { ...formModel.value }
  
  // GDate 组件已经自动处理了时间戳到 ISO 字符串的转换，直接提交即可
  emit('submit', formData)
}

// ============= 监听 initialData 变化 =============

// 监听 initialData 变化，重新初始化表单（用于编辑模式数据更新）
watch(
  () => props.initialData,
  () => {
    initFormModel()
  },
  { deep: true, immediate: true }
)

// ============= 暴露方法 =============

/**
 * 验证表单
 */
const validate = async () => {
  if (formRef.value) {
    return await formRef.value.validate()
  }
}

/**
 * 重置表单
 */
const reset = () => {
  initFormModel()
  emit('reset')
}

/**
 * 获取表单数据
 */
const getFormData = () => {
  return { ...formModel.value }
}

/**
 * 设置表单数据
 */
const setFormData = (data: Record<string, any>) => {
  formModel.value = { ...data }
  emit('update:modelValue', formModel.value)
}

defineExpose({
  validate,
  reset,
  getFormData,
  setFormData,
  formRef,
})
</script>

<style scoped lang="scss">
.g-data-form {
  width: 100%;
}

.g-data-form__tabs {
  width: 100%;
}

.g-data-form__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: var(--g-space-md) 0;
  border-top: 1px solid var(--g-border-primary);
  margin-top: var(--g-space-md);
}

/* 确保 label 宽度为 100%，统一所有标签页的 label 宽度 */
:deep(.n-form-item-label) {
  width: 100% !important;
}

:deep(.n-form-item-label--right-mark) {
  width: 100% !important;
}

:deep(.n-input-number) {
  width: 100% !important;
}

/* 带 tips 的 label 样式 */
.g-form-label-with-tips {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.g-form-label-tips {
  display: inline-flex;
  align-items: center;
  margin-left: 4px;
}
</style>

