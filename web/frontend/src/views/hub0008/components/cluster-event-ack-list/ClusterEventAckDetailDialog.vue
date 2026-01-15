<template>
  <GDialog
    v-model:show="localShow"
    :width="1200"
    title="事件确认详情"
    :show-footer="true"
    :show-cancel="false"
    confirm-text="关闭"
    :mask-closable="true"
    :close-on-esc="true"
    @close="handleClose"
    @cancel="handleClose"
    @confirm="handleClose"
  >
    <div class="cluster-event-ack-detail">
      <!-- 基本信息 -->
      <n-descriptions
        :column="2"
        bordered
        label-placement="left"
        label-style="width: 120px; font-weight: 500;"
        class="cluster-event-ack-detail__info"
      >
        <n-descriptions-item label="确认ID">
          {{ ack?.ackId || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="事件ID">
          {{ ack?.eventId || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="处理节点ID">
          {{ ack?.nodeId || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="处理节点IP">
          {{ ack?.nodeIp || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="确认状态">
          <n-tag
            :type="
              ack?.ackStatus === 'PENDING'
                ? 'warning'
                : ack?.ackStatus === 'SUCCESS'
                  ? 'success'
                  : ack?.ackStatus === 'FAILED'
                    ? 'error'
                    : 'default'
            "
            size="small"
          >
            {{
              ack?.ackStatus === 'PENDING'
                ? '待处理'
                : ack?.ackStatus === 'SUCCESS'
                  ? '成功'
                  : ack?.ackStatus === 'FAILED'
                    ? '失败'
                    : ack?.ackStatus === 'SKIPPED'
                      ? '跳过'
                      : ack?.ackStatus || '-'
            }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="重试次数">
          {{ ack?.retryCount ?? '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="处理时间">
          {{ formatDateString(ack?.processTime) || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="活动状态">
          <n-tag :type="ack?.activeFlag === 'Y' ? 'success' : 'error'" size="small">
            {{ ack?.activeFlag === 'Y' ? '活动' : '非活动' }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="创建时间">
          {{ formatDateString(ack?.addTime) || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="创建人">
          {{ ack?.addWho || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="修改时间">
          {{ formatDateString(ack?.editTime) || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="修改人">
          {{ ack?.editWho || '-' }}
        </n-descriptions-item>
      </n-descriptions>

      <!-- 结果信息 -->
      <div class="cluster-event-ack-detail__result">
        <div class="cluster-event-ack-detail__result-title">结果信息</div>
        <div class="cluster-event-ack-detail__result-content">
          {{ ack?.resultMessage || '-' }}
        </div>
      </div>

      <!-- 备注信息 -->
      <div v-if="ack?.noteText" class="cluster-event-ack-detail__note">
        <div class="cluster-event-ack-detail__note-title">备注信息</div>
        <div class="cluster-event-ack-detail__note-content">
          {{ ack.noteText }}
        </div>
      </div>

      <!-- 扩展属性 -->
      <div v-if="ack?.extProperty" class="cluster-event-ack-detail__ext">
        <div class="cluster-event-ack-detail__ext-title">扩展属性（JSON）</div>
        <GTextShow
          :content="ack.extProperty"
          format="json"
          :show-line-numbers="true"
          :show-copy-button="true"
          :auto-format="true"
          :max-height="300"
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
import type { ClusterEventAck } from '../../types'

defineOptions({
  name: 'ClusterEventAckDetailDialog'
})

interface Props {
  show?: boolean
  ack?: ClusterEventAck | null
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  ack: null
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
.cluster-event-ack-detail {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-lg, 24px);

  &__info {
    width: 100%;
  }

  &__result,
  &__note,
  &__ext {
    display: flex;
    flex-direction: column;
    gap: var(--g-space-sm, 8px);

    &-title {
      font-size: var(--g-font-size-base, 14px);
      font-weight: 600;
      color: var(--g-text-primary);
    }

    &-content {
      font-size: var(--g-font-size-base, 14px);
      color: var(--g-text-primary);
      line-height: 1.6;
      word-break: break-word;
      padding: var(--g-space-sm, 8px);
      background-color: var(--g-bg-tertiary, #f5f5f5);
      border-radius: var(--g-radius-sm, 4px);
    }
  }
}
</style>

