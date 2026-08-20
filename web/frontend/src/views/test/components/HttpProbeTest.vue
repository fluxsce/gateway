<!--
  探测监控列表同款接口 POST /gateway/hub0000/server/query。
  左侧是浏览器 fetch 的原始 HTTP（状态码 + 正文），右侧是 createApi 解包后的 JsonDataObj。
-->
<template>
  <div class="http-probe-page">
    <div class="page-header">
      <h1>HTTP 探测：监控服务器列表</h1>
      <p>
        与系统监控相同：<code>POST /gateway/hub0000/server/query</code>。需已登录。对比「后端原文」和「前端拦截器处理后」。
      </p>
    </div>

    <div class="toolbar">
      <RsButton variant="primary" :loading="loading" @click="runProbe">发送请求</RsButton>
    </div>

    <div class="panels">
      <section class="panel">
        <h2>1. fetch 原始响应（绕过 axios 拦截器）</h2>
        <pre>{{ rawDump }}</pre>
      </section>
      <section class="panel">
        <h2>2. createApi 业务层结果</h2>
        <pre>{{ apiDump }}</pre>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getHttpErrorMessage, isHttpCancelledError, isHttpTransportError } from '@/api/requestError'
import { createApi } from '@/api/request'
import { isApiSuccess, getApiMessage } from '@/utils/format'
import { RsButton } from '@/ui'
import { ref } from 'vue'

defineOptions({ name: 'HttpProbeTest' })

const hub0000Api = createApi('/gateway/hub0000')
const loading = ref(false)
const rawDump = ref('点击「发送请求」')
const apiDump = ref('点击「发送请求」')

const queryBody = {
  pageNum: 1,
  pageSize: 10,
}

/**
 * 把对象格式化成可读 JSON。
 * @param value - 任意值
 * @returns 字符串
 */
function pretty(value: unknown): string {
  return JSON.stringify(value, null, 2)
}

/**
 * 用 fetch 直打接口，看后端 HTTP 状态和正文，不被 axios 改写。
 */
async function probeRawFetch(): Promise<void> {
  const params = new URLSearchParams()
  params.set('pageNum', String(queryBody.pageNum))
  params.set('pageSize', String(queryBody.pageSize))

  const response = await fetch('/gateway/hub0000/server/query', {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
    },
    body: params.toString(),
  })

  const text = await response.text()
  let parsed: unknown = text
  try {
    parsed = JSON.parse(text)
  } catch {
    parsed = text
  }

  rawDump.value = pretty({
    httpStatus: response.status,
    httpStatusText: response.statusText,
    contentType: response.headers.get('content-type'),
    body: parsed,
  })
}

/**
 * 走和监控页相同的 createApi，看拦截器之后页面拿到什么。
 */
async function probeCreateApi(): Promise<void> {
  try {
    const result = await hub0000Api.post('/server/query', queryBody)
    apiDump.value = pretty({
      layer: 'JsonDataObj（resolve，不是 throw）',
      isApiSuccess: isApiSuccess(result),
      message: getApiMessage(result, ''),
      jsonData: result,
    })
  } catch (error) {
    if (isHttpCancelledError(error)) {
      apiDump.value = pretty({ layer: 'cancelled' })
      return
    }
    apiDump.value = pretty({
      layer: 'throw（传输失败）',
      isHttpTransportError: isHttpTransportError(error),
      message: getHttpErrorMessage(error, '未知错误'),
      error:
        error instanceof Error
          ? { name: error.name, message: error.message }
          : String(error),
    })
  }
}

/**
 * 同时打两路，便于对照。
 */
async function runProbe(): Promise<void> {
  loading.value = true
  rawDump.value = '请求中…'
  apiDump.value = '请求中…'
  try {
    await Promise.all([probeRawFetch(), probeCreateApi()])
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.http-probe-page {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 16px;

  h1 {
    margin: 0 0 8px;
    font-size: 22px;
  }

  p {
    margin: 0;
    color: var(--g-text-secondary);
    line-height: 1.5;
  }

  code {
    font-size: 13px;
  }
}

.toolbar {
  margin-bottom: 16px;
}

.panels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.panel {
  border: 1px solid var(--g-border-primary);
  border-radius: 8px;
  padding: 12px;
  background: var(--g-bg-secondary);
  min-width: 0;

  h2 {
    margin: 0 0 8px;
    font-size: 14px;
  }

  pre {
    margin: 0;
    max-height: 70vh;
    overflow: auto;
    font-size: 12px;
    line-height: 1.45;
    white-space: pre-wrap;
    word-break: break-word;
  }
}

@media (max-width: 900px) {
  .panels {
    grid-template-columns: 1fr;
  }
}
</style>
