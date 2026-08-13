<template>
  <div class="rs-button-test-page">
    <div class="page-header">
      <h1>RsButton 按钮测试</h1>
      <p class="page-description">
        覆盖 primary / secondary（default）/ ghost / danger / text / link，以及尺寸、边框、loading、禁用、图标与表单主次按钮组合。
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>主次变体（表单操作）</h2>
        <p class="hint">
          保存用 <code>primary</code>，重置/取消用 <code>secondary</code>（等价 default）。请确认 secondary 边框在浅/深色主题下都清晰可见。
        </p>
        <div class="row">
          <RsButton variant="primary" @click="onClick('保存')">保存</RsButton>
          <RsButton variant="secondary" @click="onClick('重置')">重置</RsButton>
        </div>
        <div class="row">
          <RsButton variant="primary" size="sm" @click="onClick('确认')">确认</RsButton>
          <RsButton variant="secondary" size="sm" @click="onClick('取消')">取消</RsButton>
        </div>
      </section>

      <section class="test-section">
        <h2>全部变体</h2>
        <div class="row">
          <RsButton variant="primary">primary</RsButton>
          <RsButton variant="secondary">secondary</RsButton>
          <RsButton variant="default">default</RsButton>
          <RsButton variant="ghost">ghost</RsButton>
          <RsButton variant="danger">danger</RsButton>
          <RsButton variant="text">text</RsButton>
          <RsButton variant="link">link</RsButton>
        </div>
      </section>

      <section class="test-section">
        <h2>边框对比（secondary）</h2>
        <p class="hint">
          左：默认有边框；右：显式 <code>:bordered="false"</code>。默认态应能清楚看到轮廓线。
        </p>
        <div class="compare-row">
          <div class="compare-item surface-panel">
            <span class="compare-label">bordered 默认</span>
            <RsButton variant="secondary">重置</RsButton>
          </div>
          <div class="compare-item surface-panel">
            <span class="compare-label">bordered=false</span>
            <RsButton variant="secondary" :bordered="false">重置</RsButton>
          </div>
          <div class="compare-item surface-panel">
            <span class="compare-label">ghost + bordered</span>
            <RsButton variant="ghost" :bordered="true">幽灵描边</RsButton>
          </div>
        </div>
      </section>

      <section class="test-section">
        <h2>尺寸</h2>
        <div class="row row--align">
          <RsButton variant="secondary" size="ssm">ssm</RsButton>
          <RsButton variant="secondary" size="sm">sm</RsButton>
          <RsButton variant="secondary" size="md">md</RsButton>
          <RsButton variant="secondary" size="lg">lg</RsButton>
          <RsButton variant="primary" size="ssm">ssm</RsButton>
          <RsButton variant="primary" size="sm">sm</RsButton>
          <RsButton variant="primary" size="md">md</RsButton>
          <RsButton variant="primary" size="lg">lg</RsButton>
        </div>
      </section>

      <section class="test-section">
        <h2>Loading / Disabled</h2>
        <div class="row">
          <RsButton variant="primary" :loading="loading" @click="toggleLoading">
            {{ loading ? '加载中…' : '切换 loading' }}
          </RsButton>
          <RsButton variant="secondary" :loading="loading">secondary loading</RsButton>
          <RsButton variant="primary" disabled>primary disabled</RsButton>
          <RsButton variant="secondary" disabled>secondary disabled</RsButton>
        </div>
      </section>

      <section class="test-section">
        <h2>图标</h2>
        <div class="row">
          <RsButton variant="primary" icon="plus">新建</RsButton>
          <RsButton variant="secondary" icon="refresh-cw">刷新</RsButton>
          <RsButton variant="ghost" icon="settings" icon-only tooltip="设置" />
          <RsButton variant="secondary" icon="folder" reveal-label>打开项目</RsButton>
        </div>
      </section>

      <section class="test-section">
        <h2>圆角</h2>
        <div class="row">
          <RsButton variant="secondary" radius="none">none</RsButton>
          <RsButton variant="secondary" radius="sm">sm</RsButton>
          <RsButton variant="secondary" radius="md">md</RsButton>
          <RsButton variant="secondary" radius="lg">lg</RsButton>
          <RsButton variant="secondary" radius="full">full（默认）</RsButton>
        </div>
      </section>
    </div>

    <p class="footer-hint">最近点击：{{ lastClick || '—' }}</p>
  </div>
</template>

<script setup lang="ts">
import { useAppMessage } from '@/composables/useAppMessage'
import { RsButton } from '@/ui'
import { ref } from 'vue'

defineOptions({ name: 'RsButtonTest' })

const message = useAppMessage()
const lastClick = ref('')
const loading = ref(false)

/**
 * 记录按钮点击并弹出 toast，便于确认交互。
 */
const onClick = (label: string) => {
  lastClick.value = label
  message.info(`点击：${label}`)
}

const toggleLoading = () => {
  loading.value = !loading.value
}
</script>

<style scoped lang="scss">
.rs-button-test-page {
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
  margin: 0 0 var(--g-space-md);
  font-size: var(--g-font-size-sm);
  color: var(--g-text-secondary);
  line-height: 1.5;

  code {
    font-size: 12px;
    padding: 0 4px;
    border-radius: 4px;
    background: var(--g-bg-tertiary, var(--g-bg-primary));
  }
}

.row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--g-space-sm);
  align-items: center;
}

.row--align {
  align-items: flex-end;
}

.compare-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: var(--g-space-md);
}

.compare-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--g-space-sm);
}

.compare-label {
  font-size: var(--g-font-size-xs);
  color: var(--g-text-tertiary);
}

.surface-panel {
  padding: var(--g-padding-md);
  background: var(--g-bg-primary);
  border: 1px dashed var(--g-border-primary);
  border-radius: var(--g-radius-md);
}

.footer-hint {
  margin-top: var(--g-space-xl);
  font-size: var(--g-font-size-sm);
  color: var(--g-text-secondary);
}
</style>
