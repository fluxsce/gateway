<template>
  <GModal
    v-if="hasPermission"
    :visible="props.visible ?? false"
    :title="props.dialogTitle ?? '导入'"
    width="60%"
    preset="card"
    :mask="true"
    :mask-closable="phase === 'idle'"
    :draggable="true"
    :show-footer="true"
    :show-cancel="false"
    :show-confirm="false"
    :show-fullscreen-toggle="false"
    :block-scroll="false"
    :footer-toolbar="footerButtons"
    class="g-import__modal"
    @toolbar-click="handleToolbarClick"
    @update:visible="emit('update:visible', $event)"
  >
    <div class="g-import__body">

      <!-- idle：文件选择区 -->
      <div v-if="phase === 'idle'" class="g-import__status g-import__status--idle">
        <n-upload
          ref="uploadRef"
          :accept="props.accept ?? '.xlsx,.xls'"
          :show-file-list="false"
          :custom-request="() => {}"
          @change="handleUploadChange"
          class="g-import__upload"
        >
          <n-upload-dragger class="g-import__dragger" :class="{ 'g-import__dragger--has-file': !!selectedFile }">
            <div class="g-import__dragger-inner">
              <div class="g-import__icon-wrap">
                <n-icon size="32"><CloudUploadOutline /></n-icon>
              </div>
              <template v-if="selectedFile">
                <p class="g-import__status-title">{{ selectedFile.name }}</p>
                <p class="g-import__status-desc">{{ formatSize(selectedFile.size) }}，点击或拖拽可重新选择</p>
              </template>
              <template v-else>
                <p class="g-import__status-title">点击选择或拖拽文件到此处</p>
                <p class="g-import__status-desc">支持 .xlsx / .xls 格式，大小不超过 {{ maxSizeLabel }}</p>
              </template>
            </div>
          </n-upload-dragger>
        </n-upload>
      </div>

      <!-- uploading / done / error：执行状态 -->
      <div v-else class="g-import__status" :class="`g-import__status--${phase}`">
        <div class="g-import__icon-wrap">
          <n-spin v-if="phase === 'uploading'" size="large" />
          <n-icon v-else-if="phase === 'done'" size="32"><CheckmarkCircleOutline /></n-icon>
          <n-icon v-else-if="phase === 'error'" size="32"><CloseCircleOutline /></n-icon>
        </div>
        <p class="g-import__status-title">{{ statusTitle }}</p>
        <p class="g-import__status-desc">{{ statusText }}</p>
      </div>

      <!-- 进度区域（上传中或完成后显示） -->
      <div v-if="phase === 'uploading' || phase === 'done'" class="g-import__progress-wrap">
        <div class="g-import__progress-header">
          <span class="g-import__progress-label">{{ progressLabel }}</span>
          <span class="g-import__progress-pct">{{ phase === 'done' ? '100%' : progress > 0 ? `${progress}%` : '' }}</span>
        </div>
        <n-progress
          type="line"
          :percentage="phase === 'done' ? 100 : progress"
          :show-indicator="false"
          :height="8"
          :processing="phase === 'uploading' && progress < 100"
          :status="phase === 'done' ? 'success' : 'default'"
          :border-radius="4"
          class="g-import__progress"
        />
      </div>

    </div>
  </GModal>
</template>

<script setup lang="ts">
import service from '@/api/request'
import { GModal } from '@/components/gmodal'
import type { GModalToolbarButton } from '@/components/gmodal/types'
import { store } from '@/stores'
import { CheckmarkCircleOutline, CloseCircleOutline, CloudUploadOutline } from '@vicons/ionicons5'
import type { AxiosProgressEvent } from 'axios'
import { NIcon, NProgress, NSpin, NUpload, NUploadDragger } from 'naive-ui'
import { computed, ref, watch } from 'vue'
import type { GImportEmits, GImportProps } from './types'

defineOptions({ name: 'GImport' })

const props = defineProps<GImportProps>()
const emit = defineEmits<GImportEmits>()

// ─── 状态 ──────────────────────────────────────────────────────────────────────

type Phase = 'idle' | 'uploading' | 'done' | 'error'

const phase = ref<Phase>('idle')
const progress = ref(0)
const errorMsg = ref('')
const selectedFile = ref<File | null>(null)
const uploadRef = ref()

// ─── 权限 ──────────────────────────────────────────────────────────────────────

const hasPermission = computed(() => {
  if (!props.moduleId) return true
  return store.user.hasButton(`${props.moduleId}:import`)
})

// ─── 弹窗打开时重置到 idle ─────────────────────────────────────────────────────

const resetUpload = () => {
  selectedFile.value = null
  // 清空 NUpload 内部文件列表，避免 max 限制导致 dragger 被禁用
  uploadRef.value?.clear()
}

watch(() => props.visible, (val) => {
  if (val) {
    phase.value = 'idle'
    progress.value = 0
    errorMsg.value = ''
    resetUpload()
  }
})

// ─── 辅助计算 ─────────────────────────────────────────────────────────────────

const maxSizeLabel = computed(() => {
  const mb = (props.maxSize ?? 20 * 1024 * 1024) / 1024 / 1024
  return `${mb.toFixed(0)}MB`
})

const statusTitle = computed(() => {
  switch (phase.value) {
    case 'uploading': return '正在导入'
    case 'done':      return '导入成功'
    case 'error':     return '导入失败'
    default:          return ''
  }
})

const statusText = computed(() => {
  switch (phase.value) {
    case 'uploading': return '文件上传中，请稍候…'
    case 'done':      return '数据已成功导入系统'
    case 'error':     return errorMsg.value || '请稍后重试或联系管理员'
    default:          return ''
  }
})

const progressLabel = computed(() =>
  phase.value === 'done' ? '上传完成' : progress.value > 0 ? '上传中' : '准备中'
)

// ─── Footer toolbar ───────────────────────────────────────────────────────────

const footerButtons = computed<GModalToolbarButton[]>(() => {
  if (phase.value === 'idle') {
    return [
      { key: 'cancel', label: '取消' },
      {
        key: 'start',
        label: '开始导入',
        buttonProps: { type: 'primary', disabled: !selectedFile.value },
      },
    ]
  }
  if (phase.value === 'error') {
    return [
      { key: 'retry', label: '重新选择', buttonProps: { type: 'primary' } },
      { key: 'close', label: '关闭' },
    ]
  }
  return [
    {
      key: 'close',
      label: '关闭',
      buttonProps: { disabled: phase.value === 'uploading' },
    },
  ]
})

const handleToolbarClick = (key: string) => {
  if (key === 'start') {
    startImport()
  } else if (key === 'retry') {
    phase.value = 'idle'
    resetUpload()
  } else if (key === 'cancel' || key === 'close') {
    emit('update:visible', false)
  }
}

// ─── 文件选择 ─────────────────────────────────────────────────────────────────

const handleUploadChange = (options: { file: any }) => {
  const raw: File | undefined = options.file?.file
  if (!raw) return

  // 先清空内部列表，让下一次选择不受 max 约束影响
  uploadRef.value?.clear()

  const maxSize = props.maxSize ?? 20 * 1024 * 1024
  if (raw.size > maxSize) {
    errorMsg.value = `文件大小不能超过 ${(maxSize / 1024 / 1024).toFixed(0)}MB`
    phase.value = 'error'
    return
  }
  selectedFile.value = raw
}

// ─── 上传逻辑 ─────────────────────────────────────────────────────────────────

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

// ─── 工具函数 ─────────────────────────────────────────────────────────────────

const formatSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}
</script>

<style lang="scss" scoped>
.g-import__modal :deep(.g-modal__body) {
  height: auto;
  min-height: unset;
  overflow-y: visible;
}

.g-import__body {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-lg);
  padding: var(--g-space-xl) var(--g-space-xl) var(--g-space-md);
}

/* ── idle：文件拖拽选择区 ── */
.g-import__upload {
  width: 100%;
}

.g-import__dragger {
  border-radius: var(--g-radius-lg) !important;
  background: var(--g-bg-secondary) !important;
  border: 1.5px dashed var(--g-border-primary) !important;
  transition: border-color var(--g-transition-base) var(--g-transition-ease),
              background var(--g-transition-base) var(--g-transition-ease) !important;

  &:hover {
    border-color: var(--g-primary) !important;
    background: color-mix(in srgb, var(--g-primary) 4%, transparent) !important;
  }

  &--has-file {
    border-color: var(--g-primary) !important;
    background: color-mix(in srgb, var(--g-primary) 6%, transparent) !important;

    .g-import__icon-wrap { color: var(--g-primary); }
  }
}

.g-import__dragger-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--g-space-sm);
  padding: var(--g-space-xl) var(--g-space-lg);
}

/* ── 执行状态区域 ── */
.g-import__status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--g-space-sm);
  padding: var(--g-space-lg);
  border-radius: var(--g-radius-lg);
  background: var(--g-bg-secondary);
  transition: background var(--g-transition-base) var(--g-transition-ease);

  &--uploading { background: var(--g-bg-secondary); }
  &--done      { background: color-mix(in srgb, var(--g-success) 8%, transparent); }
  &--error     { background: color-mix(in srgb, var(--g-error)   8%, transparent); }
}

.g-import__icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--g-bg-primary);
  box-shadow: var(--g-shadow-sm);
  color: var(--g-text-secondary);

  .g-import__status--idle      & { color: var(--g-primary); }
  .g-import__status--uploading & { color: var(--g-primary); }
  .g-import__status--done      & { color: var(--g-success); }
  .g-import__status--error     & { color: var(--g-error);   }
}

.g-import__status-title {
  margin: 0;
  font-size: var(--g-font-size-lg);
  font-weight: var(--g-font-weight-medium);
  color: var(--g-text-primary);
  word-break: break-all;
  text-align: center;

  .g-import__status--done  & { color: var(--g-success); }
  .g-import__status--error & { color: var(--g-error);   }
}

.g-import__status-desc {
  margin: 0;
  font-size: var(--g-font-size-sm);
  color: var(--g-text-tertiary);
  text-align: center;
  line-height: 1.6;
}

/* ── 进度区域 ── */
.g-import__progress-wrap {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-xs);
}

.g-import__progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.g-import__progress-label {
  font-size: var(--g-font-size-xs);
  color: var(--g-text-secondary);
}

.g-import__progress-pct {
  font-size: var(--g-font-size-xs);
  font-weight: var(--g-font-weight-medium);
  color: var(--g-text-secondary);
  min-width: 36px;
  text-align: right;
}

.g-import__progress {
  width: 100%;
}
</style>
