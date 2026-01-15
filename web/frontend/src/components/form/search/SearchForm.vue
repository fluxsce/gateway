<template>
  <div class="search-form">
    <!-- 工具栏 -->
    <g-toolbar
      v-if="showToolbar"
      :module-id="moduleId"
      :buttons="toolbarButtonsComputed"
      :align="toolbarAlign"
      :bordered="false"
      class="search-form__toolbar"
      @button-click="handleToolbarClick"
    >
    </g-toolbar>

    <!-- 搜索表单 -->
    <n-form
      ref="formRef"
      :model="formData"

      :label-placement="labelPlacement"
      :label-align="labelAlign"
      :size="size"
      :inline="inline"
      class="search-form__form"
    >
      <n-grid :cols="cols" :x-gap="xGap" :y-gap="yGap" class="search-form__grid">
        <!-- 基础查询字段 -->
        <template v-for="field in basicFields" :key="field.field">
          <n-gi :span="field.span || defaultFieldSpan">
            <n-form-item
              :path="field.field"
              :rule="field.rules"
              :required="field.required"
            >
              <template #label>
                <component :is="renderLabel(field)" />
              </template>
              <!-- 输入框 -->
              <n-input
                v-if="field.type === 'input' || !field.type"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请输入${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 下拉选择 -->
              <n-select
                v-else-if="field.type === 'select'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请选择${field.label}`"
                :options="field.options || []"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 日期选择 -->
              <n-date-picker
                v-else-if="field.type === 'date'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请选择${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                type="date"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 日期范围选择 -->
              <n-date-picker
                v-else-if="field.type === 'daterange'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请选择${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                type="daterange"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 日期时间选择 -->
              <n-date-picker
                v-else-if="field.type === 'datetime'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请选择${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                type="datetime"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 日期时间范围选择 -->
              <n-date-picker
                v-else-if="field.type === 'datetimerange'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请选择${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                type="datetimerange"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 数字输入 -->
              <n-input-number
                v-else-if="field.type === 'number'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请输入${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 开关 -->
              <n-switch
                v-else-if="field.type === 'switch'"
                v-model:value="formData[field.field]"
                :disabled="field.disabled"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 自定义渲染 -->
              <component
                v-else-if="field.type === 'custom' && field.render"
                :is="field.render(formData)"
              />
            </n-form-item>
          </n-gi>
        </template>
        
        <!-- 更多查询条件字段 -->
        <template v-if="showMoreFields">
          <template v-for="field in moreFields" :key="field.field">
            <n-gi :span="field.span || defaultFieldSpan">
            <n-form-item
              :path="field.field"
              :rule="field.rules"
              :required="field.required"
            >
              <template #label>
                <component :is="renderLabel(field)" />
              </template>
              <!-- 输入框 -->
              <n-input
                v-if="field.type === 'input' || !field.type"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请输入${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 下拉选择 -->
              <n-select
                v-else-if="field.type === 'select'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请选择${field.label}`"
                :options="field.options || []"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 日期选择 -->
              <n-date-picker
                v-else-if="field.type === 'date'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请选择${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                type="date"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 日期范围选择 -->
              <n-date-picker
                v-else-if="field.type === 'daterange'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请选择${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                type="daterange"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 日期时间选择 -->
              <n-date-picker
                v-else-if="field.type === 'datetime'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请选择${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                type="datetime"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 日期时间范围选择 -->
              <n-date-picker
                v-else-if="field.type === 'datetimerange'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请选择${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                type="datetimerange"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 数字输入 -->
              <n-input-number
                v-else-if="field.type === 'number'"
                v-model:value="formData[field.field]"
                :placeholder="field.placeholder || `请输入${field.label}`"
                :disabled="field.disabled"
                :clearable="field.clearable !== false"
                size="small"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 开关 -->
              <n-switch
                v-else-if="field.type === 'switch'"
                v-model:value="formData[field.field]"
                :disabled="field.disabled"
                v-bind="field.props"
                @update:value="handleFieldChange(field.field, $event)"
              />

              <!-- 自定义渲染 -->
              <component
                v-else-if="field.type === 'custom' && field.render"
                :is="field.render(formData)"
              />
            </n-form-item>
            </n-gi>
          </template>
        </template>
      </n-grid>
    </n-form>
  </div>
</template>

<script setup lang="ts">
import { GTips } from '@/components'
import GEllipsis from '@/components/gellipsis/GEllipsis.vue'
import type { ToolbarButton } from '@/components/toolbar'
import GToolbar from '@/components/toolbar/GToolbar.vue'
import { OptionsOutline, RefreshOutline, SearchOutline } from '@vicons/ionicons5'
import type { FormInst } from 'naive-ui'
import {
  NDatePicker,
  NForm,
  NFormItem,
  NGi,
  NGrid,
  NInput,
  NInputNumber,
  NSelect,
  NSwitch
} from 'naive-ui'
import { computed, h, onMounted, ref } from 'vue'
import type {
  SearchFormEmits,
  SearchFormExpose,
  SearchFormProps
} from './types'

// 定义组件名称
defineOptions({
  name: 'SearchForm'
})

// Props
const props = withDefaults(defineProps<SearchFormProps>(), {
  labelWidth: 'auto',
  labelPlacement: 'left',
  labelAlign: 'left',
  size: 'small',
  inline: false,
  cols: 24,
  xGap: 8,
  yGap: 8,
  showSearchButton: true,
  showResetButton: true,
  searchButtonText: '查询',
  resetButtonText: '重置',
  moreButtonText: '更多条件',
  toolbarAlign: 'right',
  showToolbar: true
})

// Emits
const emit = defineEmits<SearchFormEmits>()

// 表单引用
const formRef = ref<FormInst>()

// 表单数据
const formData = ref<Record<string, any>>({})

// 是否显示更多条件
const showMoreFields = ref(false)

// 切换更多条件显示/隐藏
const toggleMoreFields = () => {
  showMoreFields.value = !showMoreFields.value
}

// 初始化表单数据（只使用字段的 defaultValue）
const initFormData = () => {
  const data: Record<string, any> = {}
  // 处理基础字段
  props.fields.forEach(field => {
    if (field.defaultValue !== undefined) {
      data[field.field] = field.defaultValue
    } else {
      // 根据字段类型设置默认值
      switch (field.type) {
        case 'number':
          data[field.field] = null
          break
        case 'switch':
          data[field.field] = false
          break
        default:
          data[field.field] = ''
      }
    }
  })
  // 处理更多条件字段
  if (props.moreFields) {
    props.moreFields.forEach(field => {
      if (field.defaultValue !== undefined) {
        data[field.field] = field.defaultValue
      } else {
        // 根据字段类型设置默认值
        switch (field.type) {
          case 'number':
            data[field.field] = null
            break
          case 'switch':
            data[field.field] = false
            break
          default:
            data[field.field] = ''
        }
      }
    })
  }
  formData.value = data
}

// 基础查询字段
const basicFields = computed(() => {
  return props.fields.filter(field => field.show !== false)
})

// 更多查询条件字段
const moreFields = computed(() => {
  return props.moreFields || []
})

// 是否有更多条件字段
const hasMoreFields = computed(() => {
  return moreFields.value.length > 0
})

// 默认字段占位
const defaultFieldSpan = computed(() => {
  return Math.floor(props.cols / 4) // 默认一行4个字段
})

// 获取字段 label（支持函数类型）
const getFieldLabel = (field: any): string => {
  return typeof field.label === 'function' ? field.label(formData.value) : field.label
}

// 渲染带 tips 的 label
const renderLabel = (field: any) => {
  const labelText = getFieldLabel(field)

  if (!field.tips) {
    return () => h(GEllipsis, { text: labelText })
  }

  return () =>
    h('span', { class: 'search-form-label-with-tips' }, [
      h(GEllipsis, { text: labelText }),
      h('span', { class: 'search-form-label-tips' }, [
        (() => {
          if (!field.tips) return null
          let tipsValue: string | Component | VNode
          if (typeof field.tips === 'function') {
            tipsValue = (field.tips as (formData: Record<string, any>) => string | Component | VNode)(
              formData.value
            )
          } else {
            tipsValue = field.tips
          }
          return typeof tipsValue === 'string' ? h(GTips, { content: tipsValue }) : (tipsValue as any)
        })()
      ])
    ])
}

// 计算工具栏按钮
const toolbarButtonsComputed = computed<ToolbarButton[]>(() => {
  const buttons: ToolbarButton[] = []
  
  // 如果提供了自定义按钮，先添加自定义按钮
  if (props.toolbarButtons && props.toolbarButtons.length > 0) {
    buttons.push(...props.toolbarButtons)
  }

  // 根据配置决定是否添加默认的查询和重置按钮
  // 即使有自定义按钮，只要 showSearchButton/showResetButton 为 true，也会添加默认按钮
  if (props.showSearchButton) {
    // 检查是否已经存在 search 按钮，避免重复
    const hasSearchButton = buttons.some(btn => btn.key === 'search')
    if (!hasSearchButton) {
      buttons.push({
        key: 'search',
        label: props.searchButtonText,
        icon: SearchOutline,
        type: 'primary',
        onClick: handleSearch
      })
    }
  }

  if (props.showResetButton) {
    // 检查是否已经存在 reset 按钮，避免重复
    const hasResetButton = buttons.some(btn => btn.key === 'reset')
    if (!hasResetButton) {
      buttons.push({
        key: 'reset',
        label: props.resetButtonText,
        icon: RefreshOutline,
        onClick: handleReset
      })
    }
  }

  // 如果有更多条件字段，在最后追加"更多条件"按钮
  if (hasMoreFields.value) {
    buttons.push({
      key: 'more',
      label: props.moreButtonText,
      icon: OptionsOutline,
      onClick: toggleMoreFields
    })
  }

  return buttons
})

// 处理字段值变化
const handleFieldChange = (field: string, value: any) => {
  emit('field-change', field, value)
}

// 处理查询
const handleSearch = async () => {
  try {
    await formRef.value?.validate()
    // 触发搜索事件，传递表单数据
    emit('search', { ...formData.value })
  } catch (error) {
    console.error('表单验证失败:', error)
  }
}

// 处理重置（只重置表单数据，不发送事件）
const handleReset = () => {
  initFormData()
}

// 处理工具栏按钮点击
const handleToolbarClick = (key: string) => {
  // 查找按钮配置，检查是否有自定义 onClick
  const button = toolbarButtonsComputed.value.find(btn => btn.key === key)
  
  // 如果按钮有自定义的 onClick（如 search、reset、more 等默认按钮），
  // 说明已经通过 onClick 处理了，不需要再触发 toolbar-click 事件
  // 这样可以避免重复触发事件（比如 search 按钮会同时触发 @search 和 @toolbar-click）
  if (button?.onClick) {
    // 有自定义 onClick 的按钮不触发 toolbar-click，避免重复
    return
  }
  
  // 对于没有自定义 onClick 的按钮（通常是自定义工具栏按钮），触发 toolbar-click 事件
  emit('toolbar-click', key)
}

// 暴露的方法
const getFormRef = () => formRef.value
const getFormData = () => formData.value
const setFormData = (data: Record<string, any>) => {
  formData.value = { ...formData.value, ...data }
}
const resetForm = handleReset
const validate = async () => {
  if (formRef.value) {
    await formRef.value.validate()
  }
}
const submit = handleSearch

// 组件挂载时初始化
onMounted(() => {
  initFormData()
})

// 暴露方法
defineExpose<SearchFormExpose>({
  getFormRef,
  getFormData,
  setFormData,
  resetForm,
  validate,
  submit,
  toggleMoreFields
})
</script>

<style lang="scss" scoped>
.search-form {
  width: 100%;
  background-color: var(--g-bg-primary);
  display: flex;
  flex-direction: column;
  gap: var(--g-space-xs);
  &__toolbar {
    flex-shrink: 0;
  }

  &__form {
    width: 100%;
    flex: 1;
  }

  &__grid {
    width: 100%;
  }

  // 表单项样式优化
  :deep(.n-form-item) {
    margin-bottom: 0;

    .n-form-item-blank {
      width: 100%;
    }

    // Label 容器允许收缩，以便 GEllipsis 组件正常工作
    .n-form-item-label {
      min-width: 0; // 允许收缩
      max-width: 100%;
    }
  }
  .search-form__grid {
    padding: 0px 10px;
  }

  // 带 tips 的 label 样式
  .search-form-label-with-tips {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .search-form-label-tips {
    display: inline-flex;
    align-items: center;
    margin-left: 4px;
  }
}
</style>

