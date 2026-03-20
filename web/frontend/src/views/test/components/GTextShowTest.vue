<template>
  <div class="gtext-show-test-page">
    <div class="page-header">
      <h1>GTextShow 文本展示测试</h1>
      <p class="page-description">
        多格式（JSON / XML / 纯文本）、自动检测、复制、格式化、行号
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>JSON（自动检测 + 格式化）</h2>
        <GTextShow
          :content="jsonContent"
          format="auto"
          :show-line-numbers="showLineNumbers"
          :show-copy-button="true"
          :auto-format="true"
          max-height="240px"
          @copy="onCopy"
        />
      </section>

      <section class="test-section">
        <h2>XML</h2>
        <GTextShow
          :content="xmlContent"
          format="xml"
          :show-line-numbers="showLineNumbers"
          max-height="200px"
        />
      </section>

      <section class="test-section">
        <h2>纯文本</h2>
        <GTextShow
          :content="plainContent"
          format="txt"
          :show-copy-button="true"
          max-height="120px"
        />
      </section>

      <section class="test-section">
        <h2>选项</h2>
        <div class="options">
          <NCheckbox v-model:checked="showLineNumbers">显示行号</NCheckbox>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { NCheckbox } from 'naive-ui'
import { GTextShow } from '@/components'
import { ref } from 'vue'

defineOptions({ name: 'GTextShowTest' })

const showLineNumbers = ref(false)

const jsonContent = `{"name":"Gateway","version":"1.0.0","features":["routing","auth","rateLimit"],"config":{"port":8080,"timeout":30}}`

const xmlContent = `<?xml version="1.0" encoding="UTF-8"?>
<config>
  <server port="8080" timeout="30"/>
  <routes>
    <route path="/api" upstream="http://backend"/>
  </routes>
</config>`

const plainContent = `这是一段纯文本示例。
可用于日志、说明等。
支持复制按钮与自动格式检测。`

function onCopy(value: string) {
  console.log('GTextShow copy, length:', value.length)
}
</script>

<style scoped lang="scss">
.gtext-show-test-page {
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

.options {
  display: flex;
  align-items: center;
  gap: var(--g-space-md);
}
</style>
