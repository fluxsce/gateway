<template>
  <div class="rs-search-form-test-page">
    <div class="page-header">
      <h1>RsSearchForm 搜索表单测试</h1>
      <p class="page-description">
        niuma-ui 实现：字段栅格、更多条件、内置查询/重置工具栏、校验、label 布局
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>基础查询（含更多条件）</h2>
        <RsSearchForm
          ref="searchRef"
          module-id="test-rs-search"
          :fields="basicFields"
          :more-fields="moreFields"
          label-placement="left"
          label-align="right"
          label-width="5.5rem"
          @search="onSearch"
          @toolbar-click="onToolbarClick"
          @field-change="onFieldChange"
        />
        <pre class="result">{{ lastSearchJson }}</pre>
      </section>

      <section class="test-section">
        <h2>顶部标签 + 自定义工具栏按钮</h2>
        <RsSearchForm
          module-id="test-rs-search"
          :fields="topLabelFields"
          label-placement="top"
          :toolbar-buttons="extraToolbarButtons"
          search-button-text="立即查询"
          @search="onSearch"
          @toolbar-click="onToolbarClick"
        />
      </section>

      <section class="test-section">
        <h2>暴露方法</h2>
        <div class="row">
          <RsButton size="sm" @click="fillDemo">写入示例数据</RsButton>
          <RsButton size="sm" @click="resetForm">重置</RsButton>
          <RsButton size="sm" variant="primary" @click="validateForm">校验</RsButton>
        </div>
        <p class="hint">最近字段变更：{{ lastFieldChange || '—' }}</p>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ToolbarButton } from '@/components/toolbar'
import { RsSearchForm, type RsSearchField, type RsSearchFormExpose } from '@/components/form/rs-search'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsButton } from '@/ui'
import { computed, ref } from 'vue'

defineOptions({ name: 'RsSearchFormTest' })

const message = useAppMessage()
const searchRef = ref<RsSearchFormExpose | null>(null)
const lastSearch = ref<Record<string, any> | null>(null)
const lastFieldChange = ref('')

const lastSearchJson = computed(() =>
  lastSearch.value ? JSON.stringify(lastSearch.value, null, 2) : '（尚未查询）',
)

const basicFields: RsSearchField[] = [
  { field: 'keyword', label: '关键字', type: 'input', placeholder: '用户名/姓名', span: 6, tips: '支持模糊匹配' },
  {
    field: 'status',
    label: '状态',
    type: 'select',
    span: 6,
    options: [
      { label: '全部', value: '' },
      { label: '启用', value: 1 },
      { label: '禁用', value: 0 },
    ],
  },
  { field: 'age', label: '年龄', type: 'number', span: 6 },
  { field: 'enabled', label: '仅启用', type: 'switch', span: 6, defaultValue: false },
]

const moreFields: RsSearchField[] = [
  { field: 'createdAt', label: '创建日期', type: 'daterange', span: 12 },
  { field: 'expireAt', label: '过期时间', type: 'datetime', span: 12 },
  { field: 'deptId', label: '部门ID', type: 'input', span: 6, required: true },
]

const topLabelFields: RsSearchField[] = [
  { field: 'name', label: '名称', type: 'input', span: 8, required: true },
  {
    field: 'type',
    label: '类型',
    type: 'select',
    span: 8,
    options: [
      { label: 'A', value: 'a' },
      { label: 'B', value: 'b' },
    ],
  },
  { field: 'day', label: '日期', type: 'date', span: 8 },
]

const extraToolbarButtons: ToolbarButton[] = [
  { key: 'export', label: '导出', icon: 'DownloadOutline' },
]

const onSearch = (data: Record<string, any>) => {
  lastSearch.value = data
  message.success('已触发查询')
}

const onToolbarClick = (key: string, formData?: Record<string, any>) => {
  message.info(`工具栏：${key}`)
  if (key === 'export') {
    lastSearch.value = formData || searchRef.value?.getFormData() || null
  }
}

const onFieldChange = (field: string, value: any) => {
  lastFieldChange.value = `${field} = ${JSON.stringify(value)}`
}

const fillDemo = () => {
  searchRef.value?.setFormData({
    keyword: 'demo',
    status: '1',
    age: 18,
    enabled: true,
    deptId: 'D001',
  })
  message.info('已写入示例数据（第一块表单）')
}

const resetForm = () => {
  searchRef.value?.resetForm()
  message.info('已重置')
}

const validateForm = async () => {
  try {
    await searchRef.value?.validate()
    message.success('校验通过')
  } catch {
    message.error('校验未通过（请展开更多条件并填写部门ID）')
  }
}
</script>

<style scoped lang="scss">
.rs-search-form-test-page {
  padding: var(--g-padding-lg);
  max-width: 1100px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: var(--g-space-xl);

  h1 {
    font-size: 24px;
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 var(--g-space-sm);
  }

  .page-description {
    font-size: var(--g-font-size-sm);
    color: var(--g-text-secondary);
    margin: 0;
  }
}

.test-sections {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-xl);
}

.test-section {
  padding: var(--g-padding-md);
  background: var(--g-bg-secondary);
  border: 1px solid var(--g-border-primary);
  border-radius: var(--g-radius-lg);

  h2 {
    font-size: var(--g-font-size-lg);
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 var(--g-space-md);
  }
}

.result {
  margin: var(--g-space-md) 0 0;
  padding: var(--g-padding-sm);
  background: var(--g-bg-tertiary, #f5f5f5);
  border-radius: var(--g-radius-md);
  font-size: 12px;
  color: var(--g-text-secondary);
  overflow: auto;
}

.row {
  display: flex;
  gap: var(--g-space-sm);
  flex-wrap: wrap;
}

.hint {
  margin: var(--g-space-sm) 0 0;
  font-size: var(--g-font-size-sm);
  color: var(--g-text-secondary);
}
</style>
