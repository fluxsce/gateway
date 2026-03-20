<template>
  <div class="gdropdown-test-page">
    <div class="page-header">
      <h1>GDropdown 下拉菜单测试</h1>
      <p class="page-description">
        基于 Naive NDropdown：支持 click/hover、disabled、不同 placement、以及触发器宽度同步（minWidth）
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>基础（click）</h2>
        <div class="demo-row">
          <GDropdown :options="options" @select="onSelect">
            <NButton size="small" secondary>
              点击打开
            </NButton>
          </GDropdown>
        </div>
      </section>

      <section class="test-section">
        <h2>Hover + 延迟</h2>
        <div class="demo-row">
          <GDropdown :options="options" trigger="hover" :delay="120" @select="onSelect">
            <div class="hover-trigger">
              悬停打开（120ms）
            </div>
          </GDropdown>
        </div>
      </section>

      <section class="test-section">
        <h2>Placement</h2>
        <div class="demo-row wrap">
          <GDropdown :options="options" placement="bottom-start" @select="onSelect">
            <NButton size="small">bottom-start</NButton>
          </GDropdown>
          <GDropdown :options="options" placement="bottom-end" @select="onSelect">
            <NButton size="small">bottom-end</NButton>
          </GDropdown>
          <GDropdown :options="options" placement="top-start" @select="onSelect">
            <NButton size="small">top-start</NButton>
          </GDropdown>
          <GDropdown :options="options" placement="top-end" @select="onSelect">
            <NButton size="small">top-end</NButton>
          </GDropdown>
        </div>
      </section>

      <section class="test-section">
        <h2>Disabled</h2>
        <div class="demo-row">
          <GDropdown :options="options" disabled @select="onSelect">
            <NButton size="small" tertiary disabled>
              禁用（不可打开）
            </NButton>
          </GDropdown>
        </div>
      </section>

      <section class="test-section">
        <h2>触发器宽度同步（minWidth）</h2>
        <p class="section-description">
          打开时会读取 trigger 宽度，并设置 dropdown menu 的 minWidth。
        </p>
        <div class="demo-row wrap">
          <GDropdown :options="options" @select="onSelect">
            <NButton size="small" class="w-160">
              宽按钮（160px）
            </NButton>
          </GDropdown>
          <GDropdown :options="options" @select="onSelect">
            <NButton size="small" class="w-260">
              更宽按钮（260px）
            </NButton>
          </GDropdown>
        </div>
      </section>

      <section class="test-section result-section">
        <h2>最近一次选择</h2>
        <div class="result-output">
          <pre>{{ lastSelect }}</pre>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { GDropdown } from '@/components'
import { NButton } from 'naive-ui'
import type { DropdownOption } from 'naive-ui'
import { ref } from 'vue'

defineOptions({ name: 'GDropdownTest' })

const lastSelect = ref('（尚未选择）')

const options: DropdownOption[] = [
  { label: '服务列表', key: 'service-list' },
  { label: '刷新', key: 'refresh' },
  { type: 'divider', key: 'd1' },
  { label: '删除（disabled）', key: 'delete', disabled: true },
]

function onSelect(key: string | number, option: DropdownOption) {
  lastSelect.value = JSON.stringify({ key, label: option.label }, null, 2)
}
</script>

<style scoped lang="scss">
.gdropdown-test-page {
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

.section-description {
  font-size: var(--g-font-size-sm);
  color: var(--g-text-secondary);
  margin: -6px 0 var(--g-space-md);
}

.demo-row {
  display: flex;
  gap: var(--g-space-md);
  align-items: center;
}

.demo-row.wrap {
  flex-wrap: wrap;
}

.hover-trigger {
  display: inline-flex;
  align-items: center;
  padding: 8px 10px;
  border-radius: var(--g-radius-md);
  border: 1px dashed var(--g-border-secondary);
  color: var(--g-text-primary);
  background: var(--g-bg-tertiary);
}

.w-160 {
  width: 160px;
}

.w-260 {
  width: 260px;
}

.result-output {
  background: var(--g-bg-tertiary);
  border-radius: var(--g-radius-md);
  border: 1px solid var(--g-border-primary);
  padding: var(--g-padding-md);
  color: var(--g-text-secondary);
  font-size: var(--g-font-size-sm);
  overflow: auto;
}
</style>

