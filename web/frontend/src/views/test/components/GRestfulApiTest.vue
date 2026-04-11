<template>
  <div class="grestful-api-test-page">
    <div class="page-header">
      <h1>GRestfulApi REST 调试</h1>
      <p class="page-description">
        类 Postman：HTTP 方法、URL、Query、请求头、请求体（Raw / urlencoded）、响应状态与正文；实际出站由网关
        <code>/gateway/hubplugin/http/execute</code> 代发，不依赖浏览器直连目标的 CORS。
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>交互演示</h2>
        <p class="section-description">
          默认示例 URL 指向 httpbin（需服务端能访问外网）。可改为同域或内网可达地址。
        </p>
        <g-restful-api
          initial-url="https://httpbin.org/get"
          initial-method="GET"
        />
      </section>

      <section class="test-section">
        <h2>初始化 Props</h2>
        <p class="section-description">
          验证挂载时 <code>initialUrl</code>、<code>initialMethod</code>、<code>initialHeadersJson</code>、
          <code>initialRawBody</code>、<code>initialBodyProcessType</code> 是否写入界面。组件仅在挂载时读取这些
          Props；点击「重新挂载」递增 <code>:key</code> 以重复验证。
        </p>
        <div class="init-toolbar">
          <n-button size="small" type="primary" @click="initRemountKey += 1">
            重新挂载（key={{ initRemountKey }}）
          </n-button>
        </div>
        <g-restful-api
          :key="initRemountKey"
          class="init-grestful"
          initial-url="https://httpbin.org/post"
          initial-method="POST"
          :initial-headers-json="initHeadersJson"
          :initial-raw-body="initRawBody"
          initial-body-process-type="json"
        />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { GRestfulApi } from '@/components'
import { NButton } from 'naive-ui'
import { ref } from 'vue'

defineOptions({ name: 'GRestfulApiTest' })

/** 用于强制销毁并重建 GRestfulApi，复测挂载期初始化逻辑 */
const initRemountKey = ref(0)

/** 与网关日志重发等场景一致：JSON 对象字符串 */
const initHeadersJson = JSON.stringify({
  'X-Init-Test': 'grc',
  Accept: 'application/json',
})

const initRawBody = JSON.stringify({ hello: 'from-init', n: 1 }, null, 2)
</script>

<style scoped lang="scss">
.grestful-api-test-page {
  padding: var(--g-padding-xxl, 24px);
  max-width: 1200px;
  margin: 0 auto;
  min-height: 100%;
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
    line-height: 1.5;

    code {
      font-size: 12px;
      padding: 0 4px;
      border-radius: 4px;
      background: var(--g-bg-tertiary, rgba(0, 0, 0, 0.06));
    }
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
  line-height: 1.5;

  code {
    font-size: 12px;
    padding: 0 4px;
    border-radius: 4px;
    background: var(--g-bg-tertiary, rgba(0, 0, 0, 0.06));
  }
}

.init-toolbar {
  margin-bottom: var(--g-space-md);
}

.init-grestful {
  margin-top: var(--g-space-sm);
}
</style>
