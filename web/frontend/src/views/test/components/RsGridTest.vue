<template>
  <div class="rs-grid-test-page">
    <div class="page-header">
      <h1>RsGrid 表格测试</h1>
      <p class="page-description">
        niuma-ui RsTable 封装：勾选 + 序号、工具栏、分页、排序筛选、右键菜单、暴露方法
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section test-section--fill">
        <h2>勾选 + 序号（默认业务形态）</h2>
        <p class="hint">
          勾选列与序号列应紧凑锁宽，中间保留竖线分隔，不应出现多余空白。
        </p>
        <div class="grid-host">
          <RsGrid
            ref="gridRef"
            module-id="test-rs-grid"
            :data="pagedData"
            :columns="columns"
            :loading="loading"
            selectable
            show-index
            :toolbar-config="toolbarConfig"
            :pagination-config="paginationConfig"
            :menu-config="menuConfig"
            height="100%"
            @toolbar-button-click="onToolbarClick"
            @selection-change="onSelectionChange"
            @row-click="onRowClick"
            @row-dblclick="onRowDblclick"
            @page-change="onPageChange"
            @refresh="onRefresh"
            @menu-click="onMenuClick"
          />
        </div>
        <pre class="result">{{ statusJson }}</pre>
      </section>

      <section class="test-section">
        <h2>仅序号 / 仅勾选 / 无前缀</h2>
        <div class="variant-grid">
          <div class="variant">
            <h3>仅序号</h3>
            <div class="variant-host">
              <RsGrid
                module-id="test-rs-grid"
                :data="sampleRows"
                :columns="miniColumns"
                :selectable="false"
                show-index
                height="100%"
              />
            </div>
          </div>
          <div class="variant">
            <h3>仅勾选</h3>
            <div class="variant-host">
              <RsGrid
                module-id="test-rs-grid"
                :data="sampleRows"
                :columns="miniColumns"
                selectable
                :show-index="false"
                height="100%"
              />
            </div>
          </div>
          <div class="variant">
            <h3>无前缀列</h3>
            <div class="variant-host">
              <RsGrid
                module-id="test-rs-grid"
                :data="sampleRows"
                :columns="miniColumns"
                :selectable="false"
                :show-index="false"
                height="100%"
              />
            </div>
          </div>
        </div>
      </section>

      <section class="test-section">
        <h2>暴露方法</h2>
        <div class="row">
          <RsButton size="sm" @click="reloadDemo">reloadData</RsButton>
          <RsButton size="sm" @click="clearDemo">clearData</RsButton>
          <RsButton size="sm" @click="selectFirst">勾选首行</RsButton>
          <RsButton size="sm" @click="clearSel">clearSelection</RsButton>
          <RsButton size="sm" variant="primary" @click="logActive">打印 active</RsButton>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ToolbarButton } from '@/components/toolbar'
import { RsGrid, type RsGridColumn, type RsGridExpose } from '@/components/rs-grid'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsButton } from '@/ui'
import { computed, ref } from 'vue'

defineOptions({ name: 'RsGridTest' })

const message = useAppMessage()
const gridRef = ref<RsGridExpose | null>(null)
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const selectedCount = ref(0)
const lastEvent = ref('—')

interface DemoRow {
  id: string
  userId: string
  userName: string
  status: string
  dept: string
}

const allRows = ref<DemoRow[]>(
  Array.from({ length: 37 }, (_, i) => ({
    id: `u-${i + 1}`,
    userId: `UID${String(i + 1).padStart(4, '0')}`,
    userName: `用户${i + 1}`,
    status: i % 3 === 0 ? '禁用' : '启用',
    dept: ['研发', '产品', '运营'][i % 3],
  })),
)

const sampleRows = computed(() => allRows.value.slice(0, 5))

const pagedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return allRows.value.slice(start, start + pageSize.value)
})

const columns: RsGridColumn<DemoRow>[] = [
  { key: 'userId', title: '用户ID', ellipsis: true, sortable: true, filterable: true },
  { key: 'userName', title: '用户名', ellipsis: true, sortable: true },
  { key: 'dept', title: '部门', width: 100 },
  { key: 'status', title: '状态', width: 88, align: 'center' },
]

const miniColumns: RsGridColumn<DemoRow>[] = [
  { key: 'userId', title: '用户ID' },
  { key: 'userName', title: '用户名' },
]

const toolbarButtons: ToolbarButton[] = [
  { key: 'add', label: '新增', icon: 'AddOutline', type: 'primary' },
  { key: 'edit', label: '编辑', icon: 'CreateOutline' },
  { key: 'delete', label: '删除', icon: 'TrashOutline', type: 'error' },
]

const toolbarConfig = {
  show: true,
  showRefresh: true,
  showFullscreen: true,
  buttons: toolbarButtons,
  align: 'right' as const,
}

const paginationConfig = computed(() => ({
  show: true,
  currentPage: currentPage.value,
  pageSize: pageSize.value,
  total: allRows.value.length,
  pageSizes: [10, 20, 50],
  align: 'right' as const,
}))

const menuConfig = {
  enabled: true,
  items: [
    { key: 'view', label: '查看' },
    { key: 'copy', label: '复制 ID' },
    { key: 'sep', separator: true, label: '' },
    { key: 'delete', label: '删除', danger: true },
  ],
}

const statusJson = computed(() =>
  JSON.stringify(
    {
      page: currentPage.value,
      pageSize: pageSize.value,
      selectedCount: selectedCount.value,
      lastEvent: lastEvent.value,
    },
    null,
    2,
  ),
)

function onToolbarClick(key: string) {
  lastEvent.value = `toolbar:${key}`
  message.info(`工具栏：${key}`)
}

function onSelectionChange(rows: DemoRow[]) {
  selectedCount.value = rows.length
  lastEvent.value = `selection:${rows.length}`
}

function onRowClick(params: { row: DemoRow }) {
  lastEvent.value = `row-click:${params.row.userId}`
}

function onRowDblclick(params: { row: DemoRow }) {
  lastEvent.value = `row-dblclick:${params.row.userId}`
  message.info(`双击 ${params.row.userName}`)
}

function onPageChange(params: { currentPage: number; pageSize: number }) {
  currentPage.value = params.currentPage
  pageSize.value = params.pageSize
  lastEvent.value = `page:${params.currentPage}/${params.pageSize}`
}

function onRefresh() {
  loading.value = true
  lastEvent.value = 'refresh'
  window.setTimeout(() => {
    loading.value = false
    message.success('已刷新')
  }, 400)
}

function onMenuClick(params: { key: string; row?: DemoRow }) {
  lastEvent.value = `menu:${params.key}:${params.row?.userId ?? '-'}`
  message.info(`菜单 ${params.key}`)
}

async function reloadDemo() {
  await gridRef.value?.reloadData(allRows.value.slice(0, 8))
  lastEvent.value = 'reloadData'
}

async function clearDemo() {
  await gridRef.value?.clearData()
  lastEvent.value = 'clearData'
}

function selectFirst() {
  const row = pagedData.value[0]
  if (!row) return
  gridRef.value?.setSelectedRows([row], true)
}

function clearSel() {
  gridRef.value?.clearSelection()
}

function logActive() {
  const active = gridRef.value?.getActiveRows() ?? []
  lastEvent.value = `active:${active.map((r) => r.userId).join(',') || 'none'}`
  message.info(lastEvent.value)
}
</script>

<style scoped lang="scss">
.rs-grid-test-page {
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

  &--fill {
    display: flex;
    flex-direction: column;
    min-height: 480px;
  }
}

.hint {
  margin: 0 0 var(--g-space-sm);
  font-size: var(--g-font-size-xs);
  color: var(--g-text-tertiary);
}

.grid-host {
  flex: 1;
  min-height: 360px;
  border: 1px solid var(--g-border-primary);
  border-radius: var(--g-radius-md);
  overflow: hidden;
  background: var(--g-bg-primary);
}

.variant-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--g-space-md);
}

.variant {
  min-width: 0;

  h3 {
    margin: 0 0 var(--g-space-sm);
    font-size: var(--g-font-size-sm);
    color: var(--g-text-secondary);
  }
}

.variant-host {
  height: 220px;
  min-height: 0;
  overflow: hidden;
  border: 1px solid var(--g-border-primary);
  border-radius: var(--g-radius-md);
  background: var(--g-bg-primary);
}

.row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--g-space-sm);
}

.result {
  margin: var(--g-space-md) 0 0;
  padding: var(--g-padding-sm);
  background: var(--g-bg-tertiary);
  border-radius: var(--g-radius-md);
  font-size: var(--g-font-size-xs);
  color: var(--g-text-secondary);
  overflow: auto;
}
</style>
