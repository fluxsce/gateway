<template>
  <div class="custom-render-test-page">
    <div class="page-header">
      <h1>全局自定义渲染测试</h1>
      <p class="page-description">
        测试 $gRender：无需在模板中引入组件，TS 内直接 show(Component, props, options)，子组件通过 emit 关闭。
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>基础打开</h2>
        <p class="section-description">
          通过 $gRender.show(Component, props) 打开弹窗
        </p>
        <div class="button-group">
          <NButton type="primary" @click="openSimple">打开简单弹窗</NButton>
        </div>
      </section>

      <section class="test-section">
        <h2>带回调</h2>
        <p class="section-description">
          传入 options.onSuccess / onClose，在关闭或点击确定时触发
        </p>
        <div class="button-group">
          <NButton type="primary" @click="openWithCallback">打开弹窗（带 onSuccess）</NButton>
        </div>
      </section>

      <section class="test-section">
        <h2>自定义 props</h2>
        <p class="section-description">
          传入不同 title、content，验证 props 透传
        </p>
        <div class="button-group">
          <NButton @click="openWithProps">打开自定义标题与内容</NButton>
        </div>
      </section>

      <section class="test-section">
        <h2>程序关闭</h2>
        <p class="section-description">
          先打开弹窗，再通过 $gRender.close() 关闭
        </p>
        <div class="button-group">
          <NButton @click="openThenClose">打开后 2 秒自动关闭</NButton>
        </div>
      </section>

      <section class="test-section">
        <h2>GTextShow 弹窗类型测试</h2>
        <p class="section-description">
          同一份文本分别在 NModal（Demo）与 GModal（项目封装）中打开，验证复制/格式化/滚动表现
        </p>
        <div class="button-group">
          <NButton @click="openTextJsonInNModal">NModal - JSON</NButton>
          <NButton @click="openTextXmlInNModal">NModal - XML</NButton>
          <NButton @click="openTextLongInNModal">NModal - 长文本</NButton>
          <NButton type="primary" @click="openTextJsonInGModal">GModal - JSON</NButton>
          <NButton type="primary" @click="openTextXmlInGModal">GModal - XML</NButton>
          <NButton type="primary" @click="openTextLongInGModal">GModal - 长文本</NButton>
        </div>
      </section>

      <section class="test-section result-section">
        <h2>最近一次结果</h2>
        <div class="result-output">
          <pre>{{ lastResult }}</pre>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, getCurrentInstance } from 'vue'
import { NButton } from 'naive-ui'
import CustomRenderDemoDialog from './CustomRenderDemoDialog.vue'
import CustomRenderGModalTextDialog from './CustomRenderGModalTextDialog.vue'

defineOptions({ name: 'CustomRenderTest' })

const instance = getCurrentInstance()
const $gRender = instance?.appContext.config.globalProperties.$gRender ?? (typeof window !== 'undefined' ? (window as any).$gRender : null)

const lastResult = ref<string>('点击上方按钮触发弹窗，结果将显示在此处。')

function setResult(value: unknown) {
  lastResult.value =
    value === undefined || value === null
      ? '(无)'
      : typeof value === 'object'
        ? JSON.stringify(value, null, 2)
        : String(value)
}

function openSimple() {
  $gRender?.show(CustomRenderDemoDialog, {
    show: true,
    title: '简单弹窗',
    content: '仅打开，无回调。点击确定/取消会关闭。',
  })
}

function openWithCallback() {
  $gRender?.show(
    CustomRenderDemoDialog,
    {
      show: true,
      title: '带回调弹窗',
      content: '点击确定会触发 onSuccess，并在此处显示返回数据。',
    },
    {
      onSuccess: (data?: unknown) => setResult({ event: 'onSuccess', data }),
      onClose: () => setResult({ event: 'onClose' }),
    }
  )
}

function openWithProps() {
  $gRender?.show(CustomRenderDemoDialog, {
    show: true,
    title: '自定义标题',
    content: '这是一段自定义内容，用于验证 props 正确透传给子组件。',
  })
}

function openThenClose() {
  $gRender?.show(CustomRenderDemoDialog, {
    show: true,
    title: '自动关闭',
    content: '约 2 秒后将通过 $gRender.close() 关闭。',
  })
  setTimeout(() => {
    $gRender?.close()
    setResult({ event: 'closed by $gRender.close()', at: new Date().toISOString() })
  }, 2000)
}

const jsonContent = `{\"name\":\"Gateway\",\"version\":\"1.0.0\",\"features\":[\"routing\",\"auth\",\"rateLimit\"],\"config\":{\"port\":8080,\"timeout\":30}}`

const xmlContent = `<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<config>\n  <server port=\"8080\" timeout=\"30\"/>\n  <routes>\n    <route path=\"/api\" upstream=\"http://backend\"/>\n  </routes>\n</config>`

const longContent = `这是一段较长的文本，用于验证 GTextShow 在弹窗中的展示能力。\n\n- 支持复制\n- 支持自动检测格式（auto）\n- 支持格式化（JSON/XML）\n- 支持滚动与最大高度\n\nLorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\n${Array.from({ length: 20 }).map((_, i) => `line ${i + 1}: The quick brown fox jumps over the lazy dog.`).join('\\n')}`

function openTextJsonInNModal() {
  $gRender?.show(CustomRenderDemoDialog, { show: true, title: 'NModal - JSON', content: jsonContent })
}
function openTextXmlInNModal() {
  $gRender?.show(CustomRenderDemoDialog, { show: true, title: 'NModal - XML', content: xmlContent })
}
function openTextLongInNModal() {
  $gRender?.show(CustomRenderDemoDialog, { show: true, title: 'NModal - 长文本', content: longContent })
}

function openTextJsonInGModal() {
  $gRender?.show(CustomRenderGModalTextDialog, { show: true, title: 'GModal - JSON', content: jsonContent })
}
function openTextXmlInGModal() {
  $gRender?.show(CustomRenderGModalTextDialog, { show: true, title: 'GModal - XML', content: xmlContent })
}
function openTextLongInGModal() {
  $gRender?.show(CustomRenderGModalTextDialog, { show: true, title: 'GModal - 长文本', content: longContent, showLineNumbers: true })
}
</script>

<style scoped lang="scss">
.custom-render-test-page {
  padding: var(--g-padding-xxl, 24px);
  max-width: 900px;
  margin: 0 auto;
  background: var(--g-bg-primary);
  min-height: 100%;
}

.page-header {
  margin-bottom: var(--g-space-xl);
  padding-bottom: var(--g-space-lg);
  border-bottom: 1px solid var(--g-border-primary);

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
    line-height: 1.6;
  }
}

.test-sections {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-xl);
}

.test-section {
  padding: var(--g-padding-lg);
  background: var(--g-bg-secondary);
  border: 1px solid var(--g-border-primary);
  border-radius: 12px;

  h2 {
    font-size: var(--g-font-size-base);
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 var(--g-space-xs);
  }

  .section-description {
    font-size: var(--g-font-size-sm);
    color: var(--g-text-secondary);
    margin: 0 0 var(--g-space-md);
    line-height: 1.5;
  }

  .button-group {
    display: flex;
    flex-wrap: wrap;
    gap: var(--g-space-sm);
  }

  &.result-section .result-output {
    background: var(--g-bg-primary);
    border: 1px solid var(--g-border-primary);
    border-radius: 8px;
    padding: var(--g-padding-md);
    font-size: 12px;
    font-family: ui-monospace, monospace;
    white-space: pre-wrap;
    word-break: break-all;

    pre {
      margin: 0;
      color: var(--g-text-secondary);
    }
  }
}
</style>
