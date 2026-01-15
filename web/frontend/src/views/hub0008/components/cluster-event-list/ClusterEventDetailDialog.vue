<template>
  <GDialog
    v-model:show="localShow"
    :width="1200"
    title="事件详情"
    :show-footer="true"
    :show-cancel="false"
    confirm-text="关闭"
    :mask-closable="true"
    :close-on-esc="true"
    @close="handleClose"
    @cancel="handleClose"
    @confirm="handleClose"
  >
    <div class="cluster-event-detail">
      <!-- 基本信息 -->
      <n-descriptions
        :column="2"
        bordered
        label-placement="left"
        label-style="width: 120px; font-weight: 500;"
        class="cluster-event-detail__info"
      >
        <n-descriptions-item label="事件ID">
          {{ event?.eventId || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="事件类型">
          <n-tag type="primary" size="small">
            {{ event?.eventType || '-' }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="事件动作">
          <n-tag
            :type="
              event?.eventAction === 'START'
                ? 'success'
                : event?.eventAction === 'STOP'
                  ? 'error'
                  : event?.eventAction === 'RELOAD'
                    ? 'warning'
                    : event?.eventAction === 'RESTART'
                      ? 'info'
                      : event?.eventAction === 'CREATE'
                        ? 'success'
                        : event?.eventAction === 'UPDATE'
                          ? 'info'
                          : event?.eventAction === 'DELETE'
                            ? 'error'
                            : event?.eventAction === 'REFRESH' || event?.eventAction === 'INVALIDATE'
                              ? 'warning'
                              : 'default'
            "
            size="small"
          >
            {{ event?.eventAction || '-' }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="发布节点ID">
          {{ event?.sourceNodeId || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="发布节点IP">
          {{ event?.sourceNodeIp || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="事件时间">
          {{ formatDateString(event?.eventTime) || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="过期时间">
          {{ formatDateString(event?.expireTime) || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="活动状态">
          <n-tag :type="event?.activeFlag === 'Y' ? 'success' : 'error'" size="small">
            {{ event?.activeFlag === 'Y' ? '活动' : '非活动' }}
          </n-tag>
        </n-descriptions-item>
      </n-descriptions>

      <!-- 事件负载 -->
      <div class="cluster-event-detail__payload">
        <div class="cluster-event-detail__payload-title">事件负载（JSON）</div>
        <GTextShow
          :content="event?.eventPayload || ''"
          format="json"
          :show-line-numbers="true"
          :show-copy-button="true"
          :auto-format="true"
          :max-height="500"
        />
      </div>
    </div>
  </GDialog>
</template>

<script setup lang="ts">
import { GDialog } from '@/components/gdialog'
import { GTextShow } from '@/components/gtext-show'
import { formatDate } from '@/utils/format'
import { NDescriptions, NDescriptionsItem, NTag } from 'naive-ui'
import { ref, watch } from 'vue'
import type { ClusterEvent } from '../../types'

defineOptions({
  name: 'ClusterEventDetailDialog'
})

interface Props {
  show?: boolean
  event?: ClusterEvent | null
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  event: null
})

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'close'): void
}>()

const localShow = ref(props.show)

// 监听外部 show 变化
watch(
  () => props.show,
  (newVal) => {
    localShow.value = newVal
  }
)

// 监听内部 localShow 变化，同步到外部
watch(localShow, (newVal) => {
  if (newVal !== props.show) {
    emit('update:show', newVal)
  }
})

/**
 * 格式化日期
 */
const formatDateString = (dateStr?: string | null): string => {
  if (!dateStr) return '-'
  return formatDate(dateStr, 'YYYY-MM-DD HH:mm:ss')
}

/**
 * 处理关闭
 */
const handleClose = () => {
  localShow.value = false
  emit('close')
}
</script>

<style scoped lang="scss">
.cluster-event-detail {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-lg, 24px);

  &__info {
    width: 100%;
  }

  &__payload {
    display: flex;
    flex-direction: column;
    gap: var(--g-space-sm, 8px);

    &-title {
      font-size: var(--g-font-size-base, 14px);
      font-weight: 600;
      color: var(--g-text-primary);
    }
  }
}
</style>

