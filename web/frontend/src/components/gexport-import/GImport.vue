<template>
  <RsDialog
    v-if="hasPermission"
    :open="props.visible ?? false"
    :title="props.dialogTitle ?? '导入'"
    layout="window"
    width="60%"
    :draggable="true"
    :fullscreenable="false"
    :modal="true"
    :show-overlay="true"
    :close-on-overlay-click="phase === 'idle'"
    class="g-import__dialog"
    @update:open="emit('update:visible', $event)"
  >
    <template #body>
      <div class="g-import__body">
        <div v-if="phase === 'idle'" class="g-import__status g-import__status--idle">
          <RsUpload
            v-model="uploadFiles"
            :accept="props.accept ?? '.xlsx,.xls'"
            :max-count="1"
            :max-size="props.maxSize ?? 20 * 1024 * 1024"
            hide-dropzone-when-full
            show-download
            label="点击选择或拖拽配置文件到此处"
            :hint="`支持 .xlsx / .xls，不超过 ${maxSizeLabel}`"
            @reject="handleReject"
          />
        </div>

        <div v-else class="g-import__status" :class="`g-import__status--${phase}`">
          <div class="g-import__icon-wrap">
            <RsLoading v-if="phase === 'uploading'" size="lg" />
            <GIcon v-else-if="phase === 'done'" size="32"><CheckmarkCircleOutline /></GIcon>
            <GIcon v-else-if="phase === 'error'" size="32"><CloseCircleOutline /></GIcon>
          </div>
          <p class="g-import__status-title">{{ statusTitle }}</p>
          <p class="g-import__status-desc">{{ statusText }}</p>
          <p v-if="selectedFile" class="g-import__file-name">
            {{ selectedFile.name }}（{{ formatSize(selectedFile.size) }}）
          </p>
        </div>

        <div v-if="phase === 'uploading' || phase === 'done'" class="g-import__progress-wrap">
          <div class="g-import__progress-header">
            <span class="g-import__progress-label">{{ progressLabel }}</span>
            <span class="g-import__progress-pct">
              {{ phase === 'done' ? '100%' : progress > 0 ? `${progress}%` : '' }}
            </span>
          </div>
          <div class="g-import__progress" role="progressbar">
            <div
              class="g-import__progress-bar"
              :class="{ 'is-processing': phase === 'uploading' && progress < 100 }"
              :style="{ width: (phase === 'done' ? 100 : progress) + '%' }"
            />
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="g-import__footer">
        <template v-if="phase === 'idle'">
          <RsButton size="sm" @click="emit('update:visible', false)">取消</RsButton>
          <RsButton
            size="sm"
            variant="primary"
            :disabled="!selectedFile"
            @click="startImport"
          >
            开始导入
          </RsButton>
        </template>
        <template v-else-if="phase === 'error'">
          <RsButton size="sm" variant="primary" @click="retryImport">重新选择</RsButton>
          <RsButton size="sm" @click="emit('update:visible', false)">关闭</RsButton>
        </template>
        <RsButton
          v-else
          size="sm"
          :disabled="phase === 'uploading'"
          @click="emit('update:visible', false)"
        >
          关闭
        </RsButton>
      </div>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import service from '@/api/request'
import GIcon from '@/components/gicon/GIcon.vue'
import { store } from '@/stores'
import { RsButton, RsDialog, RsLoading, RsUpload } from '@/ui'
import { CheckmarkCircleOutline, CloseCircleOutline } from '@vicons/ionicons5'
import type { AxiosProgressEvent } from 'axios'
import { computed, ref, watch } from 'vue'
import type { GImportEmits, GImportProps } from './types'

defineOptions({ name: 'GImport' })

const props = defineProps<GImportProps>()
const emit = defineEmits<GImportEmits>()

type Phase = 'idle' | 'uploading' | 'done' | 'error'

const phase = ref<Phase>('idle')
const progress = ref(0)
const errorMsg = ref('')
/** RsUpload 受控文件列表；选中后列表即预览 */
const uploadFiles = ref<File[]>([])

const selectedFile = computed(() => uploadFiles.value[0] ?? null)

const hasPermission = computed(() => {
  if (!props.moduleId) return true
  return store.user.hasPermission(`${props.moduleId}:import`)
})

const resetUpload = () => {
  uploadFiles.value = []
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      phase.value = 'idle'
      progress.value = 0
      errorMsg.value = ''
      resetUpload()
    }
  },
)

const maxSizeLabel = computed(() => {
  const mb = (props.maxSize ?? 20 * 1024 * 1024) / 1024 / 1024
  return `${mb.toFixed(0)}MB`
})

const statusTitle = computed(() => {
  switch (phase.value) {
    case 'uploading':
      return '正在导入'
    case 'done':
      return '导入成功'
    case 'error':
      return '导入失败'
    default:
      return ''
  }
})

const statusText = computed(() => {
  switch (phase.value) {
    case 'uploading':
      return '文件上传中，请稍候…'
    case 'done':
      return '数据已成功导入系统'
    case 'error':
      return errorMsg.value || '请稍后重试或联系管理员'
    default:
      return ''
  }
})

const progressLabel = computed(() => {
  if (phase.value === 'done') return '上传完成'
  if (progress.value > 0) return '上传中'
  return '准备中'
})

const retryImport = () => {
  phase.value = 'idle'
  resetUpload()
}

const handleReject = (errors: { reason: string }[]) => {
  const first = errors[0]
  if (!first) return
  if (first.reason === 'accept') {
    errorMsg.value = '仅支持 .xlsx / .xls 格式'
  } else if (first.reason === 'maxSize') {
    errorMsg.value = `文件大小不能超过 ${maxSizeLabel.value}`
  } else {
    errorMsg.value = '文件不符合导入要求'
  }
  phase.value = 'error'
}

const startImport = async () => {
  if (!selectedFile.value) return

  phase.value = 'uploading'
  progress.value = 0
  errorMsg.value = ''

  try {
    const formData = new FormData()
    formData.append(props.fieldName ?? 'file', selectedFile.value)
    if (props.params) {
      for (const [k, v] of Object.entries(props.params)) {
        formData.append(k, String(v))
      }
    }

    const result = await service.request({
      method: 'POST',
      url: props.url,
      data: formData,
      headers: { 'Content-Type': 'multipart/form-data' },
      showLoading: false,
      onUploadProgress: (evt: AxiosProgressEvent) => {
        if (evt.total && evt.total > 0) {
          progress.value = Math.round((evt.loaded / evt.total) * 100)
        }
      },
    } as any)

    progress.value = 100
    phase.value = 'done'
    emit('success', result)
  } catch (err: any) {
    const error = err instanceof Error ? err : new Error(err?.message ?? '导入失败')
    errorMsg.value = error.message
    phase.value = 'error'
    emit('error', error)
  }
}

const formatSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}
</script>

<style lang="scss" scoped>
.g-import__footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  width: 100%;
}

.g-import__body {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-lg, 16px);
  padding: var(--g-space-xl, 24px) var(--g-space-xl, 24px) var(--g-space-md, 12px);
}

.g-import__status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--g-space-sm, 8px);
  padding: var(--g-space-lg, 16px);
  border-radius: var(--g-radius-lg, 12px);
  background: var(--g-bg-secondary, var(--rs-surface));
  transition: background var(--g-transition-base, 0.2s) ease;

  &--idle {
    align-items: stretch;
    padding: 0;
    background: transparent;
  }

  &--uploading {
    background: var(--g-bg-secondary, var(--rs-surface));
  }

  &--done {
    background: color-mix(in srgb, var(--g-success, #18a058) 8%, transparent);
  }

  &--error {
    background: color-mix(in srgb, var(--g-error, #d03050) 8%, transparent);
  }
}

.g-import__icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--g-bg-primary, var(--rs-bg));
  box-shadow: var(--g-shadow-sm, var(--rs-shadow-sm));
  color: var(--g-text-secondary, var(--rs-muted));

  .g-import__status--uploading & {
    color: var(--g-primary, var(--rs-primary));
  }

  .g-import__status--done & {
    color: var(--g-success, #18a058);
  }

  .g-import__status--error & {
    color: var(--g-error, #d03050);
  }
}

.g-import__status-title {
  margin: 0;
  font-size: var(--g-font-size-lg, 16px);
  font-weight: 500;
  color: var(--g-text-primary, var(--rs-text));
  word-break: break-all;
  text-align: center;

  .g-import__status--done & {
    color: var(--g-success, #18a058);
  }

  .g-import__status--error & {
    color: var(--g-error, #d03050);
  }
}

.g-import__status-desc {
  margin: 0;
  font-size: var(--g-font-size-sm, 13px);
  color: var(--g-text-tertiary, var(--rs-muted));
  text-align: center;
  line-height: 1.6;
}

.g-import__file-name {
  margin: 0;
  font-size: var(--g-font-size-xs, 12px);
  color: var(--g-text-secondary, var(--rs-muted));
  text-align: center;
  word-break: break-all;
}

.g-import__progress-wrap {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-xs, 4px);
}

.g-import__progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.g-import__progress-label,
.g-import__progress-pct {
  font-size: var(--g-font-size-xs, 12px);
  color: var(--g-text-secondary, var(--rs-muted));
}

.g-import__progress-pct {
  font-weight: 500;
  min-width: 36px;
  text-align: right;
}

.g-import__progress {
  width: 100%;
  height: 6px;
  border-radius: 999px;
  background: var(--rs-border-subtle, #e5e7eb);
  overflow: hidden;
}

.g-import__progress-bar {
  height: 100%;
  border-radius: inherit;
  background: var(--rs-primary, #2563eb);
  transition: width 0.2s ease;

  &.is-processing {
    background-image: linear-gradient(
      90deg,
      color-mix(in srgb, var(--rs-primary, #2563eb) 70%, transparent),
      var(--rs-primary, #2563eb),
      color-mix(in srgb, var(--rs-primary, #2563eb) 70%, transparent)
    );
    background-size: 200% 100%;
  }
}
</style>
