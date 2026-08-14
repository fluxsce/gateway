<template>
  <RsDialog
    :open="showModal"
    :title="dialogTitle"
    layout="window"
    width="90%"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :show-overlay="true"
    :close-on-overlay-click="false"
    :show-footer="false"
    @update:open="handleUpdateVisible"
    @after-close="handleAfterLeave"
  >
    <template #body>
      <div class="alert-log-detail">
        <RsLoading :loading="loading" overlay block size="lg" />

        <div v-if="alertLog" class="alert-log-detail__container">
          <RsCard :title="t('dialog.basicInfo')" size="sm" variant="outlined" class="alert-log-detail__card">
            <RsDescriptions :columns="3" size="sm" bordered label-placement="left">
              <RsDescriptionsItem :label="t('columns.alertLogId')">
                {{ alertLog.alertLogId }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.alertLevel')">
                <RsTag :variant="getAlertLevelTagType(alertLog.alertLevel)" size="sm">
                  {{ getAlertLevelLabel(alertLog.alertLevel) }}
                </RsTag>
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.alertType')">
                {{ alertLog.alertType || '-' }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.channelName')">
                {{ alertLog.channelName || '-' }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.alertTimestamp')">
                {{ formatDate(alertLog.alertTimestamp || '') }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.sendStatus')">
                <RsTag :variant="getSendStatusTagType(alertLog.sendStatus)" size="sm">
                  {{ getSendStatusLabel(alertLog.sendStatus) }}
                </RsTag>
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.sendTime')">
                {{ alertLog.sendTime ? formatDate(alertLog.sendTime) : '-' }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.addTime')">
                {{ formatDate(alertLog.addTime) }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.addWho')">
                {{ alertLog.addWho }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.editTime')">
                {{ formatDate(alertLog.editTime || '') }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.editWho')">
                {{ alertLog.editWho }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.sendErrorMessage')" :span="3">
                {{ alertLog.sendErrorMessage || '-' }}
              </RsDescriptionsItem>
            </RsDescriptions>
          </RsCard>

          <RsCard v-if="alertLog.alertTitle" :title="t('dialog.alertTitle')" size="sm" variant="outlined" class="alert-log-detail__card">
            <div class="alert-log-detail__content">{{ alertLog.alertTitle }}</div>
          </RsCard>

          <RsCard v-if="alertLog.alertContent" :title="t('dialog.alertContent')" size="sm" variant="outlined" class="alert-log-detail__card">
            <div class="alert-log-detail__content">{{ alertLog.alertContent }}</div>
          </RsCard>

          <RsCard v-if="alertLog.alertTags" :title="t('dialog.alertTags')" size="sm" variant="outlined" class="alert-log-detail__card">
            <RsCodeBlock :code="formatJson(alertLog.alertTags)" lang="json" />
          </RsCard>
          <RsCard v-if="alertLog.alertExtra" :title="t('dialog.alertExtra')" size="sm" variant="outlined" class="alert-log-detail__card">
            <RsCodeBlock :code="formatJson(alertLog.alertExtra)" lang="json" />
          </RsCard>
          <RsCard v-if="alertLog.tableData" :title="t('dialog.tableData')" size="sm" variant="outlined" class="alert-log-detail__card">
            <RsCodeBlock :code="formatJson(alertLog.tableData)" lang="json" />
          </RsCard>
          <RsCard v-if="alertLog.sendResult" :title="t('dialog.sendResult')" size="sm" variant="outlined" class="alert-log-detail__card">
            <RsCodeBlock :code="formatJson(alertLog.sendResult)" lang="json" />
          </RsCard>
        </div>

        <RsEmpty v-else-if="!loading" :description="t('dialog.empty')" />
      </div>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import {
  RsCard,
  RsCodeBlock,
  RsDescriptions,
  RsDescriptionsItem,
  RsDialog,
  RsEmpty,
  RsLoading,
  RsTag,
  type RsTagVariant,
} from '@/ui'
import { formatDate, getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { computed, ref, watch } from 'vue'
import { getAlertLog } from '../api'
import type { AlertLevel, AlertLog, SendStatus } from '../types'

defineOptions({
  name: 'AlertLogDetailDialog',
})

interface Props {
  /** 是否显示弹窗 */
  visible: boolean
  /** 告警日志ID */
  alertLogId?: string
  /** 挂载目标 */
  to?: string | HTMLElement | false
}

interface Emits {
  (e: 'update:visible', value: boolean): void
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  alertLogId: '',
  to: undefined,
})

const emit = defineEmits<Emits>()

const message = useAppMessage()
const { t } = useModuleI18n('hub0082')

const loading = ref(false)
const alertLog = ref<AlertLog | null>(null)

const dialogTitle = computed(() => {
  return props.alertLogId
    ? t('dialog.titleWithId', { id: props.alertLogId })
    : t('dialog.title')
})

const showModal = computed({
  get() {
    return props.visible
  },
  set(value: boolean) {
    emit('update:visible', value)
  },
})

watch(
  () => props.visible,
  (val) => {
    if (val && props.alertLogId) {
      void loadAlertLog()
    } else if (!val) {
      alertLog.value = null
    }
  },
  { immediate: true },
)

watch(
  () => props.alertLogId,
  (val) => {
    if (val && props.visible) {
      void loadAlertLog()
    }
  },
)

/**
 * 加载预警日志详情。
 */
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
      message.error(getApiMessage(response, t('message.detailFailed')))
      alertLog.value = null
    }
  } catch (error: any) {
    console.error('加载预警日志详情失败:', error)
    message.error(error.message || t('message.loadDetailFailed'))
    alertLog.value = null
  } finally {
    loading.value = false
  }
}

/**
 * 处理弹窗可见性变化。
 * @param value - 是否打开
 */
const handleUpdateVisible = (value: boolean) => {
  showModal.value = value
}

/** 对话框关闭后清空详情数据 */
const handleAfterLeave = () => {
  alertLog.value = null
}

/**
 * 获取告警级别显示标签。
 * @param level - 告警级别
 */
const getAlertLevelLabel = (level?: AlertLevel | string | null) => {
  if (!level) return ''
  const levelMap: Record<string, string> = {
    INFO: t('level.info'),
    WARN: t('level.warn'),
    ERROR: t('level.error'),
    CRITICAL: t('level.critical'),
  }
  return levelMap[level] || String(level)
}

/**
 * 将告警级别映射为 RsTag variant。
 * @param level - 告警级别
 */
const getAlertLevelTagType = (level?: AlertLevel | string | null): RsTagVariant => {
  if (!level) return 'default'
  const levelMap: Record<string, RsTagVariant> = {
    INFO: 'info',
    WARN: 'warning',
    ERROR: 'danger',
    CRITICAL: 'danger',
  }
  return levelMap[level] || 'default'
}

/**
 * 获取发送状态显示标签。
 * @param status - 发送状态
 */
const getSendStatusLabel = (status?: SendStatus | string | null) => {
  if (!status) return ''
  const statusMap: Record<string, string> = {
    PENDING: t('sendStatus.pending'),
    SENDING: t('sendStatus.sending'),
    SUCCESS: t('sendStatus.success'),
    FAILED: t('sendStatus.failed'),
  }
  return statusMap[status] || String(status)
}

/**
 * 将发送状态映射为 RsTag variant。
 * @param status - 发送状态
 */
const getSendStatusTagType = (status?: SendStatus | string | null): RsTagVariant => {
  if (!status) return 'default'
  const statusMap: Record<string, RsTagVariant> = {
    PENDING: 'default',
    SENDING: 'info',
    SUCCESS: 'success',
    FAILED: 'danger',
  }
  return statusMap[status] || 'default'
}

/**
 * 格式化 JSON 字符串。
 * @param jsonStr - 原始 JSON 文本
 * @returns 美化后的 JSON；解析失败则原样返回
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

<style scoped lang="scss">
.alert-log-detail {
  position: relative;
  min-height: 8rem;
}

.alert-log-detail__container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.alert-log-detail__card {
  margin-bottom: 0;
}

.alert-log-detail__content {
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
}
</style>
