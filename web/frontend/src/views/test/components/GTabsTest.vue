<template>
  <div class="gtabs-test-page">
    <div class="page-header">
      <h1>GTabs 标签页测试</h1>
      <p class="page-description">
        多标签、拖拽排序、关闭、右键菜单、溢出下拉；类型 line / card
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>Line 类型（默认）</h2>
        <div class="tabs-demo">
          <GTabs
            v-model:tabs="tabsLine"
            v-model:active-tab-id="activeLine"
            type="line"
            :draggable="true"
            :closable="true"
            :context-menu="true"
            @change="onChange"
            @close="onClose"
            @sort="onSort"
            @context-menu="onContextMenu"
          />
          <div class="tab-panel">
            当前激活: <strong>{{ activeLine || '—' }}</strong>
          </div>
        </div>
      </section>

      <section class="test-section">
        <h2>Card 类型</h2>
        <div class="tabs-demo">
          <GTabs
            v-model:tabs="tabsCard"
            v-model:active-tab-id="activeCard"
            type="card"
            :draggable="true"
            :closable="true"
            :context-menu="true"
          />
          <div class="tab-panel">
            当前激活: <strong>{{ activeCard || '—' }}</strong>
          </div>
        </div>
      </section>

      <section class="test-section">
        <h2>操作</h2>
        <div class="button-group">
          <NButton size="small" @click="addTab('line')">Line 加一页</NButton>
          <NButton size="small" @click="addTab('card')">Card 加一页</NButton>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { NButton } from 'naive-ui'
import { GTabs } from '@/components'
import type { GTabsTabItem } from '@/components/gtabs/types'
import { ref } from 'vue'

defineOptions({ name: 'GTabsTest' })

const tabsLine = ref<GTabsTabItem[]>([
  { tabId: 'line-1', title: '首页', fixed: true, icon: 'HomeOutline' },
  { tabId: 'line-2', title: '配置', icon: 'SettingsOutline' },
  { tabId: 'line-3', title: '日志', icon: 'DocumentTextOutline' },
  { tabId: 'line-4', title: '监控', icon: 'StatsChartOutline' },
  { tabId: 'line-5', title: '用户管理', closable: true },
])
const activeLine = ref('line-1')

const tabsCard = ref<GTabsTabItem[]>([
  { tabId: 'card-1', title: '概览', fixed: true },
  { tabId: 'card-2', title: '服务列表' },
  { tabId: 'card-3', title: '路由配置' },
])
const activeCard = ref('card-1')

let lineSeq = 6
let cardSeq = 4

function addTab(type: 'line' | 'card') {
  if (type === 'line') {
    const id = `line-${lineSeq++}`
    tabsLine.value = [...tabsLine.value, { tabId: id, title: `新标签 ${id}` }]
    activeLine.value = id
  } else {
    const id = `card-${cardSeq++}`
    tabsCard.value = [...tabsCard.value, { tabId: id, title: `新标签 ${id}` }]
    activeCard.value = id
  }
}

function onChange(tabId: string) {
  console.log('GTabs change:', tabId)
}

function onClose(tabId: string) {
  console.log('GTabs close:', tabId)
}

function onSort(tabs: GTabsTabItem[]) {
  console.log('GTabs sort:', tabs.map((t) => t.tabId))
}

function onContextMenu(action: string, tabId: string) {
  console.log('GTabs context-menu:', action, tabId)
}
</script>

<style scoped lang="scss">
.gtabs-test-page {
  padding: var(--g-padding-lg);
  max-width: 900px;
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

.tabs-demo {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-sm);
}

.tab-panel {
  padding: var(--g-space-md);
  background: var(--g-bg-tertiary);
  border-radius: var(--g-radius-md);
  font-size: var(--g-font-size-sm);
  color: var(--g-text-secondary);
}

.button-group {
  display: flex;
  gap: var(--g-space-sm);
  flex-wrap: wrap;
}
</style>
