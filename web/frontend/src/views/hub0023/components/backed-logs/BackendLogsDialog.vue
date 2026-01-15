<template>
  <GModal
    v-model:visible="page.showModal.value"
    :title="page.dialogTitle.value"
    :width="'90%'"
    :style="{ maxWidth: '1200px' }"
    preset="dialog"
    :mask-closable="false"
    :closable="true"
    :draggable="true"
    :showConfirm="false"
    @after-leave="page.handleAfterLeave"
  >

    <n-spin :show="page.loading.value">
      <div v-if="page.gatewayLogInfo.value" class="backend-logs-container">
        <n-tabs v-model:value="page.activeTab.value"  size="small">
          <!-- 基础信息 tab（外部请求信息） -->
          <n-tab-pane name="basic" tab="基础信息">
            <div class="trace-detail-container">
              <!-- 基本信息 -->
              <n-card title="基本信息" size="small" class="detail-card">
                <n-descriptions :column="3" size="small" bordered>
                  <n-descriptions-item label="链路追踪ID">
                    <n-tag type="info" size="small">{{ page.gatewayLogInfo.value.traceId }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="网关实例ID">
                    <n-tag type="success" size="small">{{ page.gatewayLogInfo.value.gatewayInstanceId }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="租户ID">
                    <n-tag type="warning" size="small">{{ page.gatewayLogInfo.value.tenantId }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="日志级别">
                    <n-tag :type="page.getLogLevelType(page.gatewayLogInfo.value.logLevel) as any" size="small">
                      {{ page.getLogLevelText(page.gatewayLogInfo.value.logLevel) }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="日志类型">
                    <n-tag :type="page.getLogTypeColor(page.gatewayLogInfo.value.logType) as any" size="small">
                      {{ page.getLogTypeText(page.gatewayLogInfo.value.logType) }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="记录时间">
                    <span>{{ page.formatDate(page.gatewayLogInfo.value.addTime) }}</span>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>

              <!-- 请求信息 -->
              <n-card title="请求信息" size="small" class="detail-card">
                <n-descriptions :column="2" size="small" bordered>
                  <n-descriptions-item label="请求方法">
                    <n-tag :type="page.getMethodColor(page.gatewayLogInfo.value.requestMethod) as any" size="small">
                      {{ page.gatewayLogInfo.value.requestMethod }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="请求路径">
                    <n-ellipsis :line-clamp="2">{{ page.gatewayLogInfo.value.requestPath }}</n-ellipsis>
                  </n-descriptions-item>
                  <n-descriptions-item label="请求查询参数">
                    <n-ellipsis :line-clamp="2">{{ page.gatewayLogInfo.value.requestQuery || '无' }}</n-ellipsis>
                  </n-descriptions-item>
                  <n-descriptions-item label="客户端IP">
                    <n-tag type="info" size="small">{{ page.gatewayLogInfo.value.clientIpAddress }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="请求大小">
                    <n-tag type="info" size="small">
                      {{ page.formatFileSize(page.gatewayLogInfo.value.requestSize || 0) }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="客户端端口">
                    <span>{{ page.gatewayLogInfo.value.clientPort || '无' }}</span>
                  </n-descriptions-item>
                  <n-descriptions-item label="User-Agent">
                    <n-ellipsis :line-clamp="2">{{ page.gatewayLogInfo.value.userAgent || '无' }}</n-ellipsis>
                  </n-descriptions-item>
                  <n-descriptions-item label="Referer">
                    <n-ellipsis :line-clamp="2">{{ page.gatewayLogInfo.value.referer || '无' }}</n-ellipsis>
                  </n-descriptions-item>
                  <n-descriptions-item label="用户标识">
                    <span>{{ page.gatewayLogInfo.value.userIdentifier || '无' }}</span>
                  </n-descriptions-item>
                  <n-descriptions-item label="父链路追踪ID">
                    <span>{{ page.gatewayLogInfo.value.parentTraceId || '无' }}</span>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>

              <!-- 响应信息 -->
              <n-card title="响应信息" size="small" class="detail-card">
                <n-descriptions :column="3" size="small" bordered>
                  <n-descriptions-item label="网关状态码">
                    <n-tag :type="page.getStatusCodeType(page.gatewayLogInfo.value.gatewayStatusCode) as any" size="small">
                      {{ page.gatewayLogInfo.value.gatewayStatusCode }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="总处理时间">
                    <n-tag :type="page.getResponseTimeType(page.gatewayLogInfo.value.totalProcessingTimeMs || 0) as any" size="small">
                      {{ page.gatewayLogInfo.value.totalProcessingTimeMs || 0 }}ms
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="网关处理时间">
                    <n-tag :type="page.getResponseTimeType(page.gatewayLogInfo.value.gatewayProcessingTimeMs || 0) as any" size="small">
                      {{ page.gatewayLogInfo.value.gatewayProcessingTimeMs || 0 }}ms
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="响应大小">
                    <n-tag type="info" size="small">
                      {{ page.formatFileSize(page.gatewayLogInfo.value.responseSize || 0) }}
                    </n-tag>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>

              <!-- 时间跟踪 -->
              <n-card title="时间跟踪" size="small" class="detail-card">
                <n-descriptions :column="2" size="small" bordered>
                  <n-descriptions-item label="网关开始处理">
                    <span>{{ page.formatDate(page.gatewayLogInfo.value.gatewayStartProcessingTime, 'YYYY-MM-DD HH:mm:ss.SSS') }}</span>
                  </n-descriptions-item>
                  <n-descriptions-item label="网关完成处理">
                    <span>{{ page.gatewayLogInfo.value.gatewayFinishedProcessingTime ? page.formatDate(page.gatewayLogInfo.value.gatewayFinishedProcessingTime, 'YYYY-MM-DD HH:mm:ss.SSS') : '未完成' }}</span>
                  </n-descriptions-item>
                  <n-descriptions-item label="重试次数">
                    <n-tag type="warning" size="small">{{ page.gatewayLogInfo.value.retryCount || 0 }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="重置次数">
                    <n-tag type="error" size="small">{{ page.gatewayLogInfo.value.resetCount || 0 }}</n-tag>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>

              <!-- 路由信息 -->
              <n-card title="路由信息" size="small" class="detail-card">
                <n-descriptions :column="2" size="small" bordered>
                  <n-descriptions-item label="代理类型">
                    <n-tag :type="page.getProxyTypeColor(page.gatewayLogInfo.value.proxyType) as any" size="small">
                      {{ page.getProxyTypeText(page.gatewayLogInfo.value.proxyType) }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="匹配路由">
                    <n-ellipsis :line-clamp="2">{{ page.gatewayLogInfo.value.matchedRoute || '无' }}</n-ellipsis>
                  </n-descriptions-item>
                  <n-descriptions-item label="路由名称">
                    <n-ellipsis :line-clamp="2">{{ page.gatewayLogInfo.value.routeName || '无' }}</n-ellipsis>
                  </n-descriptions-item>
                  <n-descriptions-item label="网关节点IP">
                    <n-tag type="info" size="small">{{ page.gatewayLogInfo.value.gatewayNodeIp || '无' }}</n-tag>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>

              <!-- 错误信息 -->
              <n-card v-if="page.gatewayLogInfo.value.errorCode || page.gatewayLogInfo.value.errorMessage" title="错误信息" size="small" class="detail-card">
                <n-descriptions :column="1" size="small" bordered>
                  <n-descriptions-item v-if="page.gatewayLogInfo.value.errorCode" label="错误码">
                    <n-tag type="error" size="small">{{ page.gatewayLogInfo.value.errorCode }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item v-if="page.gatewayLogInfo.value.errorMessage" label="错误消息">
                    <div class="error-message">{{ page.gatewayLogInfo.value.errorMessage }}</div>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>

              <!-- 请求头信息 -->
              <n-card v-if="page.gatewayLogInfo.value.requestHeaders" title="请求头信息" size="small" class="detail-card">
                <GTextShow :content="page.gatewayLogInfo.value.requestHeaders" format="auto" :auto-format="true" :max-height="300" />
              </n-card>

              <!-- 请求体 -->
              <n-card v-if="page.gatewayLogInfo.value.requestBody" title="请求体" size="small" class="detail-card">
                <GTextShow :content="page.gatewayLogInfo.value.requestBody" format="auto" :auto-format="true" :max-height="300" />
              </n-card>

              <!-- 响应头 -->
              <n-card v-if="page.gatewayLogInfo.value.responseHeaders" title="响应头" size="small" class="detail-card">
                <GTextShow :content="page.gatewayLogInfo.value.responseHeaders" format="auto" :auto-format="true" :max-height="300" />
              </n-card>

              <!-- 响应体 -->
              <n-card v-if="page.gatewayLogInfo.value.responseBody" title="响应体" size="small" class="detail-card">
                <GTextShow :content="page.gatewayLogInfo.value.responseBody" format="auto" :auto-format="true" :max-height="300" />
              </n-card>
            </div>
          </n-tab-pane>

          <!-- 后端服务追踪日志 tabs -->
          <n-tab-pane
            v-for="(trace, index) in page.backendTraces.value"
            :key="trace.backendTraceId || index"
            :name="`service-${index}`"
            :tab="page.getServiceTabName(trace, index)"
          >
            <div class="trace-detail-container">
              <!-- 基本信息 -->
              <n-card title="基本信息" size="small" class="detail-card">
                <n-descriptions :column="3" size="small" bordered>
                  <n-descriptions-item label="后端追踪ID">
                    <n-tag type="info" size="small">{{ trace.backendTraceId }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="服务定义ID">
                    <n-tag type="success" size="small">{{ trace.serviceDefinitionId || '无' }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="服务名称">
                    <n-tag type="warning" size="small">{{ trace.serviceName || '无' }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="追踪状态">
                    <n-tag :type="page.getTraceStatusType(trace.traceStatus) as any" size="small">
                      {{ page.getTraceStatusText(trace.traceStatus) }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="成功标记">
                    <n-tag :type="trace.successFlag === 'Y' ? 'success' : 'error'" size="small">
                      {{ trace.successFlag === 'Y' ? '成功' : '失败' }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="重试次数">
                    <n-tag type="warning" size="small">{{ trace.retryCount || 0 }}</n-tag>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>

              <!-- 转发信息 -->
              <n-card title="转发信息" size="small" class="detail-card">
                <n-descriptions :column="2" size="small" bordered>
                  <n-descriptions-item label="转发地址">
                    <n-ellipsis :line-clamp="2">{{ trace.forwardAddress || '无' }}</n-ellipsis>
                  </n-descriptions-item>
                  <n-descriptions-item label="转发方法">
                    <n-tag :type="page.getMethodColor(trace.forwardMethod) as any" size="small">
                      {{ trace.forwardMethod || '无' }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="转发路径">
                    <n-ellipsis :line-clamp="2">{{ trace.forwardPath || '无' }}</n-ellipsis>
                  </n-descriptions-item>
                  <n-descriptions-item label="转发查询参数">
                    <n-ellipsis :line-clamp="2">{{ trace.forwardQuery || '无' }}</n-ellipsis>
                  </n-descriptions-item>
                  <n-descriptions-item label="请求大小">
                    <n-tag type="info" size="small">
                      {{ page.formatFileSize(trace.requestSize || 0) }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="负载均衡策略">
                    <span>{{ trace.loadBalancerStrategy || '无' }}</span>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>

              <!-- 时间信息 -->
              <n-card title="时间信息" size="small" class="detail-card">
                <n-descriptions :column="2" size="small" bordered>
                  <n-descriptions-item label="请求开始时间">
                    <span>{{
                      trace.requestStartTime
                        ? page.formatDate(trace.requestStartTime, 'YYYY-MM-DD HH:mm:ss.SSS')
                        : '无'
                    }}</span>
                  </n-descriptions-item>
                  <n-descriptions-item label="响应接收时间">
                    <span>{{
                      trace.responseReceivedTime
                        ? page.formatDate(trace.responseReceivedTime, 'YYYY-MM-DD HH:mm:ss.SSS')
                        : '无'
                    }}</span>
                  </n-descriptions-item>
                  <n-descriptions-item label="请求耗时">
                    <n-tag :type="page.getResponseTimeType(trace.requestDurationMs || 0) as any" size="small">
                      {{ trace.requestDurationMs || 0 }}ms
                    </n-tag>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>

              <!-- 响应信息 -->
              <n-card title="响应信息" size="small" class="detail-card">
                <n-descriptions :column="3" size="small" bordered>
                  <n-descriptions-item label="状态码">
                    <n-tag :type="page.getStatusCodeType(trace.statusCode || 0) as any" size="small">
                      {{ trace.statusCode || '无' }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="响应大小">
                    <n-tag type="info" size="small">
                      {{ page.formatFileSize(trace.responseSize || 0) }}
                    </n-tag>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>

              <!-- 错误信息 -->
              <n-card v-if="trace.errorCode || trace.errorMessage" title="错误信息" size="small" class="detail-card">
                <n-descriptions :column="1" size="small" bordered>
                  <n-descriptions-item v-if="trace.errorCode" label="错误码">
                    <n-tag type="error" size="small">{{ trace.errorCode }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item v-if="trace.errorMessage" label="错误消息">
                    <div class="error-message">{{ trace.errorMessage }}</div>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>

              <!-- 转发头信息 -->
              <n-card v-if="trace.forwardHeaders" title="转发头信息" size="small" class="detail-card">
                <GTextShow :content="trace.forwardHeaders" format="auto" :auto-format="true" :max-height="300" />
              </n-card>

              <!-- 转发体 -->
              <n-card v-if="trace.forwardBody" title="转发体" size="small" class="detail-card">
                <GTextShow :content="trace.forwardBody" format="auto" :auto-format="true" :max-height="300" />
              </n-card>

              <!-- 响应头 -->
              <n-card v-if="trace.responseHeaders" title="响应头" size="small" class="detail-card">
                <GTextShow :content="trace.responseHeaders" format="auto" :auto-format="true" :max-height="300" />
              </n-card>

              <!-- 响应体 -->
              <n-card v-if="trace.responseBody" title="响应体" size="small" class="detail-card">
                <GTextShow :content="trace.responseBody" format="auto" :auto-format="true" :max-height="300" />
              </n-card>

              <!-- 扩展信息 -->
              <n-card v-if="trace.extProperty" title="扩展信息" size="small" class="detail-card">
                <GTextShow :content="trace.extProperty" format="auto" :auto-format="true" :max-height="300" />
              </n-card>
            </div>
          </n-tab-pane>
        </n-tabs>
      </div>

      <n-empty v-else description="暂无日志数据" />
    </n-spin>
  </GModal>
</template>

<script setup lang="ts">
import GModal from '@/components/gmodal/GModal.vue'
import GTextShow from '@/components/gtext-show/GTextShow.vue'
import { NCard, NDescriptions, NDescriptionsItem, NEllipsis, NEmpty, NSpin, NTabPane, NTabs, NTag } from 'naive-ui'
import { useBackendLogsPage } from './page'

interface Props {
  /** 是否显示弹窗 */
  visible: boolean
  /** 链路追踪ID */
  traceId?: string
}

interface Emits {
  (e: 'update:visible', value: boolean): void
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  traceId: '',
})

const emit = defineEmits<Emits>()

// 使用页面级 Hook
const page = useBackendLogsPage(props, emit)
</script>

<style scoped>


.trace-detail-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-card {
  margin-bottom: 0;
}

.detail-card :deep(.n-card__content) {
  padding: 12px;
}

.error-message {
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--n-error-color);
  font-family: var(--n-font-family-mono);
}

.note-text {
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--n-text-color-2);
  font-style: italic;
}

.trace-detail-container :deep(.n-descriptions-item-content) {
  word-break: break-word;
}
</style>

