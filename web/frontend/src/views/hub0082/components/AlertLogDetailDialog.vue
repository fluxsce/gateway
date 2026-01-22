<template>
  <GModal
    v-model:visible="showModal"
    :title="dialogTitle"
    :width="'90%'"
    :style="{ maxWidth: '1200px' }"
    preset="dialog"
    :mask-closable="false"
    :closable="true"
    :draggable="true"
    :show-footer="false"
    @after-leave="handleAfterLeave"
  >
    <n-spin :show="loading">
      <div v-if="alertLog" class="alert-log-detail-container">
        <!-- 基本信息 -->
        <n-card title="基本信息" size="small" class="detail-card">
          <n-descriptions :column="3" size="small" bordered>
            <n-descriptions-item label="日志ID">
              <n-ellipsis :tooltip="false">{{ alertLog.alertLogId }}</n-ellipsis>
            </n-descriptions-item>
            <n-descriptions-item label="告警级别">
              <n-tag :type="getAlertLevelTagType(alertLog.alertLevel)" size="small">
                {{ getAlertLevelLabel(alertLog.alertLevel) }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="告警类型">
              {{ alertLog.alertType || '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="渠道名称">
              {{ alertLog.channelName || '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="告警时间">
              {{ formatDate(alertLog.alertTimestamp || '') }}
            </n-descriptions-item>
            <n-descriptions-item label="发送状态">
              <n-tag :type="getSendStatusTagType(alertLog.sendStatus)" size="small">
                {{ getSendStatusLabel(alertLog.sendStatus) }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="发送时间">
              {{ alertLog.sendTime ? formatDate(alertLog.sendTime) : '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="创建时间">
              {{ formatDate(alertLog.addTime) }}
            </n-descriptions-item>
            <n-descriptions-item label="创建人">
              {{ alertLog.addWho }}
            </n-descriptions-item>
            <n-descriptions-item label="修改时间">
              {{ formatDate(alertLog.editTime || '') }}
            </n-descriptions-item>
            <n-descriptions-item label="修改人">
              {{ alertLog.editWho }}
            </n-descriptions-item>
            <n-descriptions-item label="错误信息" :span="3">
              {{ alertLog.sendErrorMessage || '-' }}
            </n-descriptions-item>
          </n-descriptions>
        </n-card>

        <!-- 告警标题 -->
        <n-card v-if="alertLog.alertTitle" title="告警标题" size="small" class="detail-card">
          <div class="content-text">{{ alertLog.alertTitle }}</div>
        </n-card>

        <!-- 告警内容 -->
        <n-card v-if="alertLog.alertContent" title="告警内容" size="small" class="detail-card">
          <div class="content-text">{{ alertLog.alertContent }}</div>
        </n-card>

        <!-- JSON 数据展示 -->
        <n-card v-if="hasJsonData" title="扩展数据" size="small" class="detail-card">
          <n-collapse>
            <n-collapse-item v-if="alertLog.alertTags" title="告警标签 (JSON)" name="tags">
              <GTextShow :content="formatJson(alertLog.alertTags)" format="json" :auto-format="true" :max-height="300" />
            </n-collapse-item>
            <n-collapse-item v-if="alertLog.alertExtra" title="额外数据 (JSON)" name="extra">
              <GTextShow :content="formatJson(alertLog.alertExtra)" format="json" :auto-format="true" :max-height="300" />
            </n-collapse-item>
            <n-collapse-item v-if="alertLog.tableData" title="表格数据 (JSON)" name="table">
              <GTextShow :content="formatJson(alertLog.tableData)" format="json" :auto-format="true" :max-height="300" />
            </n-collapse-item>
            <n-collapse-item v-if="alertLog.sendResult" title="发送结果 (JSON)" name="result">
              <GTextShow :content="formatJson(alertLog.sendResult)" format="json" :auto-format="true" :max-height="300" />
            </n-collapse-item>
          </n-collapse>
        </n-card>
      </div>

      <n-empty v-else description="暂无日志数据" />
    </n-spin>
  </GModal>
</template>

<script setup lang="ts">
import { GModal } from '@/components/gmodal'
import { GTextShow } from '@/components/gtext-show'
import { formatDate, getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { NCard, NCollapse, NCollapseItem, NDescriptions, NDescriptionsItem, NEllipsis, NEmpty, NSpin, NTag, useMessage } from 'naive-ui'
import { computed, ref, watch } from 'vue'
import { getAlertLog } from '../api'
import type { AlertLevel, AlertLog, SendStatus } from '../types'
import { ALERT_LEVEL_OPTIONS, SEND_STATUS_OPTIONS } from '../types'

interface Props {
  /** 是否显示弹窗 */
  visible: boolean
  /** 告警日志ID */
  alertLogId?: string
}

interface Emits {
  (e: 'update:visible', value: boolean): void
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  alertLogId: '',
})

const emit = defineEmits<Emits>()

const message = useMessage()

// 状态管理
const loading = ref(false)
const alertLog = ref<AlertLog | null>(null)

// 计算属性
const dialogTitle = computed(() => {
  return props.alertLogId ? `预警日志详情 - ${props.alertLogId}` : '预警日志详情'
})

const showModal = computed({
  get() {
    return props.visible
  },
  set(value: boolean) {
    emit('update:visible', value)
  },
})

const hasJsonData = computed(() => {
  return !!(alertLog.value?.alertTags || alertLog.value?.alertExtra || alertLog.value?.tableData || alertLog.value?.sendResult)
})

// 监听 visible 变化，加载数据
watch(
  () => props.visible,
  (val) => {
    if (val && props.alertLogId) {
      loadAlertLog()
    } else if (!val) {
      // 关闭时清空数据
      alertLog.value = null
    }
  },
  { immediate: true }
)

// 监听 alertLogId 变化
watch(
  () => props.alertLogId,
  (val) => {
    if (val && props.visible) {
      loadAlertLog()
    }
  }
)

// 加载预警日志详情
const loadAlertLog = async () => {
  if (!props.alertLogId) {
    return
  }

  try {
    loading.value = true
    const response = await getAlertLog(props.alertLogId)
    if (isApiSuccess(response)) {
      alertLog.value = parseJsonData<AlertLog>(response)
    } else {
      message.error(getApiMessage(response, '获取预警日志详情失败'))
      alertLog.value = null
    }
  } catch (error: any) {
    console.error('加载预警日志详情失败:', error)
    message.error(error.message || '加载预警日志详情失败')
    alertLog.value = null
  } finally {
    loading.value = false
  }
}

// 对话框关闭后处理
const handleAfterLeave = () => {
  alertLog.value = null
}

// 工具函数
const getAlertLevelLabel = (level?: AlertLevel | string | null) => {
  if (!level) return ''
  const option = ALERT_LEVEL_OPTIONS.find(opt => opt.value === level)
  return option?.label || String(level)
}

const getAlertLevelTagType = (level?: AlertLevel | string | null): 'default' | 'success' | 'error' | 'warning' | 'primary' | 'info' => {
  if (!level) return 'default'
  const levelMap: Record<string, 'default' | 'success' | 'error' | 'warning' | 'primary' | 'info'> = {
    INFO: 'info',
    WARN: 'warning',
    ERROR: 'error',
    CRITICAL: 'error',
  }
  return levelMap[level] || 'default'
}

const getSendStatusLabel = (status?: SendStatus | string | null) => {
  if (!status) return ''
  const option = SEND_STATUS_OPTIONS.find(opt => opt.value === status)
  return option?.label || String(status)
}

const getSendStatusTagType = (status?: SendStatus | string | null): 'default' | 'success' | 'error' | 'warning' | 'primary' | 'info' => {
  if (!status) return 'default'
  const statusMap: Record<string, 'default' | 'success' | 'error' | 'warning' | 'primary' | 'info'> = {
    PENDING: 'default',
    SENDING: 'info',
    SUCCESS: 'success',
    FAILED: 'error',
  }
  return statusMap[status] || 'default'
}

/**
 * 格式化 JSON 字符串
 */
const formatJson = (jsonStr: string | null | undefined): string => {
  if (!jsonStr) return ''
  try {
    const obj = JSON.parse(jsonStr)
    return JSON.stringify(obj, null, 2)
  } catch {
    return jsonStr
  }
}
</script>

<style scoped>
.alert-log-detail-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-card {
  margin-bottom: 8px;
}

.content-text {
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
}
</style>

