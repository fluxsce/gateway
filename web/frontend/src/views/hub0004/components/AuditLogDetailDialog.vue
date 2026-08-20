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
      <div class="audit-log-detail">
        <RsLoading :loading="loading" overlay block size="lg" />

        <div v-if="auditLog" class="audit-log-detail__container">
          <RsCard :title="t('dialog.basicInfo')" size="sm" variant="outlined" class="audit-log-detail__card">
            <RsDescriptions :columns="3" size="sm" bordered label-placement="left">
              <RsDescriptionsItem :label="t('columns.auditId')">
                {{ auditLog.auditId }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.addTime')">
                {{ formatDate(auditLog.addTime || '') }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.action')">
                <RsTag :variant="getActionTagType(auditLog.action)" size="sm">
                  {{ getActionLabel(auditLog.action) }}
                </RsTag>
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.result')">
                <RsTag :variant="getResultTagType(auditLog.result)" size="sm">
                  {{ getResultLabel(auditLog.result) }}
                </RsTag>
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.userName')">
                {{ auditLog.userName || '-' }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.userId')">
                {{ auditLog.userId || '-' }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.moduleCode')">
                {{ getModuleLabel(auditLog.moduleCode) || '-' }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.resourceCode')">
                {{ auditLog.resourceCode || '-' }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.addWho')">
                {{ auditLog.addWho || '-' }}
              </RsDescriptionsItem>
            </RsDescriptions>
          </RsCard>

          <RsCard :title="t('dialog.targetInfo')" size="sm" variant="outlined" class="audit-log-detail__card">
            <RsDescriptions :columns="3" size="sm" bordered label-placement="left">
              <RsDescriptionsItem :label="t('columns.targetType')">
                {{ getTargetTypeLabel(auditLog.targetType) || '-' }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.targetName')">
                {{ auditLog.targetName || '-' }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.targetId')" :span="1">
                {{ auditLog.targetId || '-' }}
              </RsDescriptionsItem>
            </RsDescriptions>
          </RsCard>

          <RsCard :title="t('dialog.requestInfo')" size="sm" variant="outlined" class="audit-log-detail__card">
            <RsDescriptions :columns="3" size="sm" bordered label-placement="left">
              <RsDescriptionsItem :label="t('columns.requestMethod')">
                {{ auditLog.requestMethod || '-' }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.clientIP')">
                {{ auditLog.clientIP || '-' }}
              </RsDescriptionsItem>
              <RsDescriptionsItem :label="t('columns.requestPath')" :span="1">
                {{ auditLog.requestPath || '-' }}
              </RsDescriptionsItem>
            </RsDescriptions>
          </RsCard>

          <RsCard v-if="auditLog.detail" :title="t('dialog.detail')" size="sm" variant="outlined" class="audit-log-detail__card">
            <RsCodeBlock v-if="isJsonDetail" :code="formattedDetail" lang="json" />
            <div v-else class="audit-log-detail__content">{{ auditLog.detail }}</div>
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
  RsDescriptions,
  RsDescriptionsItem,
  RsDialog,
  RsEmpty,
  RsLoading,
  RsTag,
  type RsTagVariant,
} from '@/ui'
import { RsCodeBlock } from '@/ui/code-block'
import { formatDate, getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { computed, ref, watch } from 'vue'
import { getAuditLog } from '../api'
import type { AuditAction, AuditResult, AuthAuditLog } from '../types'

defineOptions({
  name: 'AuditLogDetailDialog',
})

interface Props {
  /** 是否显示弹窗 */
  visible: boolean
  /** 审计记录 ID */
  auditId?: string
  /** 挂载目标 */
  to?: string | HTMLElement | false
}

interface Emits {
  (e: 'update:visible', value: boolean): void
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  auditId: '',
  to: undefined,
})

const emit = defineEmits<Emits>()

const message = useAppMessage()
const { t } = useModuleI18n('hub0004')

const loading = ref(false)
const auditLog = ref<AuthAuditLog | null>(null)

const dialogTitle = computed(() => {
  return props.auditId ? t('dialog.titleWithId', { id: props.auditId }) : t('dialog.title')
})

const showModal = computed({
  get() {
    return props.visible
  },
  set(value: boolean) {
    emit('update:visible', value)
  },
})

const formattedDetail = computed(() => {
  const raw = auditLog.value?.detail
  if (!raw) return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
})

const isJsonDetail = computed(() => {
  const raw = auditLog.value?.detail
  if (!raw) return false
  try {
    JSON.parse(raw)
    return true
  } catch {
    return false
  }
})

watch(
  () => props.visible,
  (val) => {
    if (val && props.auditId) {
      void loadAuditLog()
    } else if (!val) {
      auditLog.value = null
    }
  },
  { immediate: true },
)

watch(
  () => props.auditId,
  (val) => {
    if (val && props.visible) {
      void loadAuditLog()
    }
  },
)

/**
 * 加载审计日志详情。
 */
const loadAuditLog = async () => {
  if (!props.auditId) {
    return
  }

  try {
    loading.value = true
    const response = await getAuditLog(props.auditId)
    if (isApiSuccess(response)) {
      auditLog.value = parseJsonData<AuthAuditLog>(response)
    } else {
      message.error(getApiMessage(response, t('message.detailFailed')))
      auditLog.value = null
    }
  } catch (error: any) {
    message.error(error.message || t('message.loadDetailFailed'))
    auditLog.value = null
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
  auditLog.value = null
}

/**
 * 获取动作显示文案。
 * @param action - 审计动作
 */
const getActionLabel = (action?: AuditAction | string | null) => {
  if (!action) return ''
  const map: Record<string, string> = {
    CREATE: t('action.create'),
    UPDATE: t('action.update'),
    DELETE: t('action.delete'),
    ROLLBACK: t('action.rollback'),
    GRANT: t('action.grant'),
  }
  return map[action] || String(action)
}

/**
 * 将动作映射为 RsTag variant。
 * @param action - 审计动作
 */
const getActionTagType = (action?: AuditAction | string | null): RsTagVariant => {
  if (!action) return 'default'
  const map: Record<string, RsTagVariant> = {
    CREATE: 'success',
    UPDATE: 'info',
    DELETE: 'danger',
    ROLLBACK: 'warning',
    GRANT: 'warning',
  }
  return map[action] || 'default'
}

/**
 * 获取结果显示文案。
 * @param result - 审计结果
 */
const getResultLabel = (result?: AuditResult | string | null) => {
  if (!result) return ''
  if (result === 'Y') return t('result.success')
  if (result === 'N') return t('result.fail')
  return String(result)
}

/**
 * 将结果映射为 RsTag variant。
 * @param result - 审计结果
 */
const getResultTagType = (result?: AuditResult | string | null): RsTagVariant => {
  if (result === 'Y') return 'success'
  if (result === 'N') return 'danger'
  return 'default'
}

/**
 * 获取模块显示文案。
 * @param code - 模块编码
 */
const getModuleLabel = (code?: string | null) => {
  if (!code) return ''
  const key = `module.${code}`
  const label = t(key)
  if (label && label !== key) {
    return `${code} ${label}`
  }
  return String(code)
}

/**
 * 获取目标类型显示文案。
 * @param type - 目标类型
 */
const getTargetTypeLabel = (type?: string | null) => {
  if (!type) return ''
  const key = `targetType.${type}`
  const label = t(key)
  if (label && label !== key) return label
  return String(type)
}
</script>

<style scoped lang="scss">
.audit-log-detail {
  position: relative;
  min-height: 8rem;
}

.audit-log-detail__container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.audit-log-detail__card {
  margin-bottom: 0;
}

.audit-log-detail__content {
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
}
</style>
