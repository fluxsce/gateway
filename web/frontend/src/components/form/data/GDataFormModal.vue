<template>
  <!-- 基于通用 GModal 的数据编辑弹窗封装 -->
  <GModal
    v-bind="props"
    @update:visible="handleUpdateVisible"
    @cancel="handleCancel"
    @confirm="handleConfirm"
    @after-enter="emit('after-enter')"
    @after-leave="emit('after-leave')"
  >
    <!-- 头部：可根据 mode 设置默认标题，也支持自定义 header 插槽 -->
    <template #header>
      <slot name="header" :mode="props.mode">
        <div class="g-data-modal__header">
          <span class="g-modal__title">
            {{ computedTitle }}
          </span>
        </div>
      </slot>
    </template>

    <!-- 默认内容区域，基于 formFields 自动渲染表单；也支持自定义插槽覆盖 -->
    <div class="g-data-modal__body">
      <slot :mode="props.mode" :form-data="formModel" :form-ref="formRef">
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
            class="g-data-modal__tabs"
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
                  <template v-if="field.type === 'fieldset'">
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
                <template v-if="field.type === 'fieldset'">
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
      </slot>
    </div>

    <!-- 底部操作区：默认提供取消/确定按钮，也允许完全自定义 -->
    <template v-if="props.showFooter" #footer>
      <slot name="footer" :mode="props.mode" :confirmLoading="props.confirmLoading" :onConfirm="handleConfirm" :onCancel="handleCancel">
        <div class="g-data-modal__footer">
          <n-space justify="end" :size="8">
            <n-button v-if="props.showCancel" size="small" @click="handleCancel">
              {{ props.cancelText }}
            </n-button>
            <n-button
              v-if="props.showConfirm && props.mode !== 'view'"
              type="primary"
              size="small"
              :loading="props.confirmLoading"
              @click="handleConfirm"
            >
              {{ props.confirmText }}
            </n-button>
          </n-space>
        </div>
      </slot>
    </template>
  </GModal>
</template>

<script setup lang="ts">
import { GDate, GFieldset, GFileUpload, GModal, GTips } from '@/components'
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
import { computed, h, ref, watch } from 'vue'
import type { DataFormField, DataFormTab, DataModalEmits, DataModalProps } from './types'

defineOptions({
  name: 'GdataFormModal'
})

const props = withDefaults(defineProps<DataModalProps>(), {
  mode: 'create',
  visible: false,
  title: '',
  width: '60%',
  preset: 'dialog',
  maskClosable: false,
  closable: true,
  showFooter: true,
  showCancel: true,
  showConfirm: true,
  cancelText: '取消',
  confirmText: '保存',
  confirmLoading: false,
  autoCloseOnConfirm: true,
  autoResetOnClose: false,
  autoFocus: true,
  segmented: false,
  bordered: false,
  // 这里必须给 showFullscreenToggle 一个默认值，否则通过 v-bind="props"
  // 传给 GModal 时会变成 undefined，从而覆盖 GModal 内部的默认值
  showFullscreenToggle: true,
  formFields: () => [] as DataFormField[],
  formTabs: () => [] as DataFormTab[]
})

const emit = defineEmits<DataModalEmits>()

// 表单引用与数据模型
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
      tabsList = [{ key: 'default', label: props.title || '表单' }]
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

// 默认标题：支持通过 title 覆盖，否则根据 mode 显示"新增/编辑/查看"
const computedTitle = computed(() => {
  if (props.title) return props.title
  switch (props.mode) {
    case 'create':
      return '新增'
    case 'edit':
      return '编辑'
    case 'view':
      return '查看详情'
    default:
      return ''
  }
})

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
}

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

// 打开弹窗时初始化表单
// 注意：只在 visible 变为 true 时初始化，避免 initialData 变化时重新初始化表单（会丢失用户已编辑的内容）
watch(
  () => props.visible,
  (val) => {
    if (val) {
      // 弹窗打开时初始化表单数据
      initFormModel()
    }
  }
)

const handleUpdateVisible = (value: boolean) => {
  emit('update:visible', value)
  if (!value && props.autoResetOnClose) {
    emit('reset')
    initFormModel()
  }
  if (!value) {
    emit('close')
  }
}

const handleCancel = () => {
  emit('cancel')
  emit('update:visible', false)
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
}

// 提交前进行表单校验，并将表单数据透传给调用方
const handleConfirm = async () => {
  if (formRef.value) {
    try {
      await formRef.value.validate()
    } catch {
      // 校验失败，不提交
      return
    }
  }

  // 确保获取最新的表单数据
  // 使用 toRaw 确保获取的是当前最新的值，避免响应式代理导致的旧值问题
  const formData = { ...formModel.value }
  
  // GDate 组件已经自动处理了时间戳到 ISO 字符串的转换，直接提交即可
  emit('submit', formData)
  emit('confirm')
  if (props.autoCloseOnConfirm) {
    emit('update:visible', false)
  }
}
</script>

<style scoped lang="scss">
.g-data-modal__header {
  display: flex;
  align-items: center;
  height: var(--g-modal-header-height);
  /* 头部样式由 GModal 统一控制，这里不再增加额外内边距和边框，避免与左侧图标拉开过大间距 */
  padding: 0;
}

.g-data-modal__body {
  padding: var(--g-space-md);
}

.g-data-modal__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  height: var(--g-modal-footer-height);
  padding: 0 var(--g-space-md);
  border-top: 1px solid var(--g-border-primary);
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


