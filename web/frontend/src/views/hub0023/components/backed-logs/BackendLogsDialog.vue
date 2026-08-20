<template>
  <RsDialog
    :open="page.showModal.value"
    :title="page.dialogTitle.value"
    layout="window"
    width="90%"
    class="hub0023-backend-logs-dialog"
    :draggable="true"
    :fullscreenable="true"
    :show-overlay="true"
    :close-on-overlay-click="false"
    :show-footer="false"
    @update:open="(visible: boolean) => (page.showModal.value = visible)"
    @after-close="page.handleAfterLeave"
  >
    <template #body>
    <div class="backend-logs-body">
      <RsLoading :loading="page.loading.value" overlay block size="lg" />

      <div v-if="page.gatewayLogInfo.value" class="backend-logs-container">
        <RsTabs
          v-model="page.activeTab.value"
          :items="tabItems"
          variant="line"
          size="sm"
          borderless
        >
          <template #basic>
            <div class="trace-detail-container">
              <RsCard title="基本信息" size="sm" variant="outlined" class="detail-card">
                <RsDescriptions :columns="3" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem label="链路追踪ID">
                    <RsTag variant="info" size="sm">{{ page.gatewayLogInfo.value.traceId }}</RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="网关实例ID">
                    <RsTag variant="success" size="sm">{{ page.gatewayLogInfo.value.gatewayInstanceId }}</RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="租户ID">
                    <RsTag variant="warning" size="sm">{{ page.gatewayLogInfo.value.tenantId }}</RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="日志级别">
                    <RsTag :variant="page.getLogLevelType(page.gatewayLogInfo.value.logLevel)" size="sm">
                      {{ page.getLogLevelText(page.gatewayLogInfo.value.logLevel) }}
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="日志类型">
                    <RsTag :variant="page.getLogTypeColor(page.gatewayLogInfo.value.logType)" size="sm">
                      {{ page.getLogTypeText(page.gatewayLogInfo.value.logType) }}
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="记录时间">
                    <span>{{ page.formatDate(page.gatewayLogInfo.value.addTime) }}</span>
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard title="请求信息" size="sm" variant="outlined" class="detail-card">
                <RsDescriptions :columns="2" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem label="请求方法">
                    <RsTag :variant="page.getMethodColor(page.gatewayLogInfo.value.requestMethod)" size="sm">
                      {{ page.gatewayLogInfo.value.requestMethod }}
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="请求路径">
                    <span class="ellipsis-2">{{ page.gatewayLogInfo.value.requestPath }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="请求查询参数">
                    <span class="ellipsis-2">{{ page.gatewayLogInfo.value.requestQuery || '无' }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="客户端IP">
                    <RsTag variant="info" size="sm">{{ page.gatewayLogInfo.value.clientIpAddress }}</RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="请求大小">
                    <RsTag variant="info" size="sm">
                      {{ page.formatFileSize(page.gatewayLogInfo.value.requestSize || 0) }}
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="客户端端口">
                    <span>{{ page.gatewayLogInfo.value.clientPort || '无' }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="User-Agent">
                    <span class="ellipsis-2">{{ page.gatewayLogInfo.value.userAgent || '无' }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="Referer">
                    <span class="ellipsis-2">{{ page.gatewayLogInfo.value.referer || '无' }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="用户标识">
                    <span>{{ page.gatewayLogInfo.value.userIdentifier || '无' }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="父链路追踪ID">
                    <span>{{ page.gatewayLogInfo.value.parentTraceId || '无' }}</span>
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard title="响应信息" size="sm" variant="outlined" class="detail-card">
                <RsDescriptions :columns="3" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem label="网关状态码">
                    <RsTag :variant="page.getStatusCodeType(page.gatewayLogInfo.value.gatewayStatusCode)" size="sm">
                      {{ page.gatewayLogInfo.value.gatewayStatusCode }}
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="总处理时间">
                    <RsTag :variant="page.getResponseTimeType(page.gatewayLogInfo.value.totalProcessingTimeMs || 0)" size="sm">
                      {{ page.gatewayLogInfo.value.totalProcessingTimeMs || 0 }}ms
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="网关处理时间">
                    <RsTag :variant="page.getResponseTimeType(page.gatewayLogInfo.value.gatewayProcessingTimeMs || 0)" size="sm">
                      {{ page.gatewayLogInfo.value.gatewayProcessingTimeMs || 0 }}ms
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="响应大小">
                    <RsTag variant="info" size="sm">
                      {{ page.formatFileSize(page.gatewayLogInfo.value.responseSize || 0) }}
                    </RsTag>
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard title="时间跟踪" size="sm" variant="outlined" class="detail-card">
                <RsDescriptions :columns="2" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem label="网关开始处理">
                    <span>{{ page.formatDate(page.gatewayLogInfo.value.gatewayStartProcessingTime, 'YYYY-MM-DD HH:mm:ss.SSS') }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="网关完成处理">
                    <span>{{ page.gatewayLogInfo.value.gatewayFinishedProcessingTime ? page.formatDate(page.gatewayLogInfo.value.gatewayFinishedProcessingTime, 'YYYY-MM-DD HH:mm:ss.SSS') : '未完成' }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="重试次数">
                    <RsTag variant="warning" size="sm">{{ page.gatewayLogInfo.value.retryCount || 0 }}</RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="重置次数">
                    <RsTag variant="danger" size="sm">{{ page.gatewayLogInfo.value.resetCount || 0 }}</RsTag>
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard title="路由信息" size="sm" variant="outlined" class="detail-card">
                <RsDescriptions :columns="2" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem label="代理类型">
                    <RsTag :variant="page.getProxyTypeColor(page.gatewayLogInfo.value.proxyType)" size="sm">
                      {{ page.getProxyTypeText(page.gatewayLogInfo.value.proxyType) }}
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="匹配路由">
                    <span class="ellipsis-2">{{ page.gatewayLogInfo.value.matchedRoute || '无' }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="路由名称">
                    <span class="ellipsis-2">{{ page.gatewayLogInfo.value.routeName || '无' }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="网关节点IP">
                    <RsTag variant="info" size="sm">{{ page.gatewayLogInfo.value.gatewayNodeIp || '无' }}</RsTag>
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard
                v-if="page.hasGatewayLogError(page.gatewayLogInfo.value)"
                title="错误信息"
                size="sm"
                variant="outlined"
                class="detail-card"
              >
                <RsDescriptions :columns="1" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem v-if="page.gatewayLogInfo.value.errorCode" label="错误码">
                    <RsTag variant="danger" size="sm">{{ page.gatewayLogInfo.value.errorCode }}</RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem v-if="page.gatewayLogInfo.value.errorMessage" label="错误消息">
                    <div class="error-message">{{ page.gatewayLogInfo.value.errorMessage }}</div>
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard
                v-if="page.getGatewayLogNote(page.gatewayLogInfo.value)"
                title="备注信息"
                size="sm"
                variant="outlined"
                class="detail-card"
              >
                <RsDescriptions :columns="1" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem label="备注">
                    {{ page.getGatewayLogNote(page.gatewayLogInfo.value) }}
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard v-if="page.gatewayLogInfo.value.requestHeaders" title="请求头信息" size="sm" variant="outlined" class="detail-card">
                <RsCodeBlock :code="page.formatJsonData(page.gatewayLogInfo.value.requestHeaders)" lang="json" />
              </RsCard>
              <RsCard v-if="page.gatewayLogInfo.value.requestBody" title="请求体" size="sm" variant="outlined" class="detail-card">
                <RsCodeBlock :code="page.formatJsonData(page.gatewayLogInfo.value.requestBody)" lang="json" />
              </RsCard>
              <RsCard v-if="page.gatewayLogInfo.value.responseHeaders" title="响应头" size="sm" variant="outlined" class="detail-card">
                <RsCodeBlock :code="page.formatJsonData(page.gatewayLogInfo.value.responseHeaders)" lang="json" />
              </RsCard>
              <RsCard v-if="page.gatewayLogInfo.value.responseBody" title="响应体" size="sm" variant="outlined" class="detail-card">
                <RsCodeBlock :code="page.formatJsonData(page.gatewayLogInfo.value.responseBody)" lang="json" />
              </RsCard>
            </div>
          </template>

          <template
            v-for="(trace, index) in page.backendTraces.value"
            :key="trace.backendTraceId || index"
            #[`service-${index}`]
          >
            <div class="trace-detail-container">
              <RsCard title="基本信息" size="sm" variant="outlined" class="detail-card">
                <RsDescriptions :columns="3" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem label="后端追踪ID">
                    <RsTag variant="info" size="sm">{{ trace.backendTraceId }}</RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="服务定义ID">
                    <RsTag variant="success" size="sm">{{ trace.serviceDefinitionId || '无' }}</RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="服务名称">
                    <RsTag variant="warning" size="sm">{{ trace.serviceName || '无' }}</RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="追踪状态">
                    <RsTag :variant="page.getTraceStatusType(trace.traceStatus)" size="sm">
                      {{ page.getTraceStatusText(trace.traceStatus) }}
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="成功标记">
                    <RsTag :variant="trace.successFlag === 'Y' ? 'success' : 'danger'" size="sm">
                      {{ trace.successFlag === 'Y' ? '成功' : '失败' }}
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="重试次数">
                    <RsTag variant="warning" size="sm">{{ trace.retryCount || 0 }}</RsTag>
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard title="转发信息" size="sm" variant="outlined" class="detail-card">
                <RsDescriptions :columns="2" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem label="转发地址">
                    <span class="ellipsis-2">{{ trace.forwardAddress || '无' }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="转发方法">
                    <RsTag :variant="page.getMethodColor(trace.forwardMethod)" size="sm">
                      {{ trace.forwardMethod || '无' }}
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="转发路径">
                    <span class="ellipsis-2">{{ trace.forwardPath || '无' }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="转发查询参数">
                    <span class="ellipsis-2">{{ trace.forwardQuery || '无' }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="请求大小">
                    <RsTag variant="info" size="sm">
                      {{ page.formatFileSize(trace.requestSize || 0) }}
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="负载均衡策略">
                    <span>{{ trace.loadBalancerStrategy || '无' }}</span>
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard title="时间信息" size="sm" variant="outlined" class="detail-card">
                <RsDescriptions :columns="2" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem label="请求开始时间">
                    <span>{{
                      trace.requestStartTime
                        ? page.formatDate(trace.requestStartTime, 'YYYY-MM-DD HH:mm:ss.SSS')
                        : '无'
                    }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="响应接收时间">
                    <span>{{
                      trace.responseReceivedTime
                        ? page.formatDate(trace.responseReceivedTime, 'YYYY-MM-DD HH:mm:ss.SSS')
                        : '无'
                    }}</span>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="请求耗时">
                    <RsTag :variant="page.getResponseTimeType(trace.requestDurationMs || 0)" size="sm">
                      {{ trace.requestDurationMs || 0 }}ms
                    </RsTag>
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard title="响应信息" size="sm" variant="outlined" class="detail-card">
                <RsDescriptions :columns="3" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem label="状态码">
                    <RsTag :variant="page.getStatusCodeType(trace.statusCode || 0)" size="sm">
                      {{ trace.statusCode || '无' }}
                    </RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem label="响应大小">
                    <RsTag variant="info" size="sm">
                      {{ page.formatFileSize(trace.responseSize || 0) }}
                    </RsTag>
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard
                v-if="trace.errorCode || trace.errorMessage"
                title="错误信息"
                size="sm"
                variant="outlined"
                class="detail-card"
              >
                <RsDescriptions :columns="1" size="sm" bordered label-placement="left">
                  <RsDescriptionsItem v-if="trace.errorCode" label="错误码">
                    <RsTag variant="danger" size="sm">{{ trace.errorCode }}</RsTag>
                  </RsDescriptionsItem>
                  <RsDescriptionsItem v-if="trace.errorMessage" label="错误消息">
                    <div class="error-message">{{ trace.errorMessage }}</div>
                  </RsDescriptionsItem>
                </RsDescriptions>
              </RsCard>

              <RsCard v-if="trace.forwardHeaders" title="转发头信息" size="sm" variant="outlined" class="detail-card">
                <RsCodeBlock :code="page.formatJsonData(trace.forwardHeaders)" lang="json" />
              </RsCard>
              <RsCard v-if="trace.forwardBody" title="转发体" size="sm" variant="outlined" class="detail-card">
                <RsCodeBlock :code="page.formatJsonData(trace.forwardBody)" lang="json" />
              </RsCard>
              <RsCard v-if="trace.responseHeaders" title="响应头" size="sm" variant="outlined" class="detail-card">
                <RsCodeBlock :code="page.formatJsonData(trace.responseHeaders)" lang="json" />
              </RsCard>
              <RsCard v-if="trace.responseBody" title="响应体" size="sm" variant="outlined" class="detail-card">
                <RsCodeBlock :code="page.formatJsonData(trace.responseBody)" lang="json" />
              </RsCard>
              <RsCard v-if="trace.extProperty" title="扩展信息" size="sm" variant="outlined" class="detail-card">
                <RsCodeBlock :code="page.formatJsonData(trace.extProperty)" lang="json" />
              </RsCard>
            </div>
          </template>
        </RsTabs>
      </div>

      <RsEmpty v-else description="暂无日志数据" />
    </div>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import {
  RsCard,
  RsDescriptions,
  RsDescriptionsItem,
  RsDialog,
  RsEmpty,
  RsLoading,
  RsTabs,
  RsTag,
  type RsTabItem,
} from '@/ui'
import { RsCodeBlock } from '@/ui/code-block'
import { computed } from 'vue'
import { useBackendLogsPage } from './page'

interface Props {
  /** 是否显示弹窗 */
  visible: boolean
  /** 链路追踪ID */
  traceId?: string
  /** 网关实例 ID，与列表行一致，详情查询必填 */
  gatewayInstanceId?: string
}

interface Emits {
  (e: 'update:visible', value: boolean): void
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  traceId: '',
  gatewayInstanceId: '',
})

const emit = defineEmits<Emits>()

const page = useBackendLogsPage(props, emit)

/** 基础信息 + 各后端追踪日志 Tab */
const tabItems = computed<RsTabItem[]>(() => {
  const items: RsTabItem[] = [{ value: 'basic', label: '基础信息' }]
  page.backendTraces.value.forEach((trace, index) => {
    items.push({
      value: `service-${index}`,
      label: page.getServiceTabName(trace, index),
    })
  })
  return items
})
</script>

<style scoped>
.backend-logs-body {
  position: relative;
  min-height: 200px;
}

.trace-detail-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
  word-break: break-word;
}

.detail-card {
  margin-bottom: 0;
}

.ellipsis-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.error-message {
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--rs-danger);
  font-family: var(--rs-font-mono, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace);
}

.note-text {
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--rs-muted);
  font-style: italic;
}

</style>
