<template>
  <div class="gtoolbar-test-page">
    <div class="page-header">
      <h1>GToolbar 工具栏测试</h1>
      <p class="page-description">
        测试 GToolbar / ToolbarButton：扁平按钮、分组、尾部按钮、tooltip、dropdown、权限 key、对齐方式
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>扁平按钮（默认）</h2>
        <GToolbar
          module-id="test-toolbar"
          :buttons="flatButtons"
          bordered
          @button-click="onButtonClick"
        />
        <p class="hint">最近点击：{{ lastClick || '—' }}</p>
      </section>

      <section class="test-section">
        <h2>右对齐 + 尾部按钮</h2>
        <GToolbar
          module-id="test-toolbar"
          title="用户列表"
          :buttons="endButtons"
          align="space-between"
          bordered
          @button-click="onButtonClick"
        />
      </section>

      <section class="test-section">
        <h2>分组按钮</h2>
        <GToolbar
          module-id="test-toolbar"
          :groups="toolbarGroups"
          bordered
          @button-click="onButtonClick"
          @dropdown-select="onDropdownSelect"
        />
        <p class="hint">下拉选择：{{ lastDropdown || '—' }}</p>
      </section>

      <section class="test-section">
        <h2>加载 / 禁用</h2>
        <div class="row">
          <RsButton size="sm" @click="toggleLoading">切换 loading</RsButton>
          <RsButton size="sm" @click="toggleDisabled">切换 disabled</RsButton>
        </div>
        <GToolbar
          module-id="test-toolbar"
          :buttons="stateButtons"
          bordered
          @button-click="onButtonClick"
        />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import GToolbar from '@/components/toolbar/GToolbar.vue'
import type { ToolbarButton, ToolbarGroup } from '@/components/toolbar'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsButton } from '@/ui'
import { computed, ref } from 'vue'

defineOptions({ name: 'GToolbarTest' })

const message = useAppMessage()
const lastClick = ref('')
const lastDropdown = ref('')
const loading = ref(false)
const disabled = ref(false)

const flatButtons: ToolbarButton[] = [
  { key: 'add', label: '新增', icon: 'AddOutline', type: 'primary' },
  { key: 'edit', label: '编辑', icon: 'CreateOutline' },
  { key: 'delete', label: '删除', icon: 'TrashOutline', type: 'error', tooltip: '删除选中行' },
  { key: 'refresh', label: '刷新', icon: 'RefreshOutline' },
]

const endButtons: ToolbarButton[] = [
  { key: 'export', label: '导出', icon: 'DownloadOutline' },
  { key: 'import', label: '导入', icon: 'CloudUploadOutline', atEnd: true },
  { key: 'setting', label: '设置', icon: 'SettingsOutline', atEnd: true },
]

const toolbarGroups: ToolbarGroup[] = [
  {
    key: 'crud',
    title: '编辑',
    buttons: [
      { key: 'g-add', label: '新增', icon: 'AddOutline', type: 'primary' },
      { key: 'g-edit', label: '编辑', icon: 'CreateOutline' },
    ],
  },
  {
    key: 'more',
    title: '更多',
    divider: true,
    buttons: [
      {
        key: 'more-actions',
        label: '更多操作',
        icon: 'EllipsisHorizontalOutline',
        dropdown: true,
        dropdownOptions: [
          { key: 'copy', label: '复制', icon: 'CopyOutline' },
          { key: 'archive', label: '归档', icon: 'ArchiveOutline' },
          { key: 'divider-1', label: '', divider: true },
          { key: 'danger', label: '危险操作', icon: 'WarningOutline' },
        ],
      },
    ],
  },
]

const stateButtons = computed<ToolbarButton[]>(() => [
  {
    key: 'save',
    label: '保存',
    icon: 'SaveOutline',
    type: 'primary',
    loading: loading.value,
    disabled: disabled.value,
  },
  {
    key: 'cancel',
    label: '取消',
    icon: 'CloseOutline',
    disabled: disabled.value,
  },
])

const onButtonClick = (key: string) => {
  lastClick.value = key
  message.info(`按钮点击：${key}`)
}

const onDropdownSelect = (buttonKey: string, optionKey: string) => {
  lastDropdown.value = `${buttonKey} → ${optionKey}`
  message.success(`下拉选择：${optionKey}`)
}

const toggleLoading = () => {
  loading.value = !loading.value
}

const toggleDisabled = () => {
  disabled.value = !disabled.value
}
</script>

<style scoped lang="scss">
.gtoolbar-test-page {
  padding: var(--g-padding-lg);
  max-width: 960px;
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

.hint {
  margin: var(--g-space-sm) 0 0;
  font-size: var(--g-font-size-sm);
  color: var(--g-text-secondary);
}

.row {
  display: flex;
  gap: var(--g-space-sm);
  margin-bottom: var(--g-space-md);
}
</style>
