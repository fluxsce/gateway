<template>
  <RsDialog
    v-if="hasPermission"
    :open="props.visible ?? false"
    :title="props.dialogTitle ?? '导出'"
    layout="window"
    width="60%"
    :draggable="true"
    :fullscreenable="false"
    :modal="true"
    :show-overlay="true"
    :close-on-overlay-click="phase === 'idle'"
    class="g-export__dialog"
    @update:open="emit('update:visible', $event)"
  >
    <template #body>
      <div class="g-export__body">
        <div v-if="phase === 'idle'" class="g-export__status g-export__status--idle">
          <div class="g-export__icon-wrap">
            <GIcon size="32"><DownloadOutline /></GIcon>
          </div>
          <p class="g-export__status-title">准备导出</p>
          <p class="g-export__status-desc">
            点击「开始导出」生成并下载文件，大批量数据导出可能需要一些时间。
          </p>
        </div>

        <div v-else class="g-export__status" :class="`g-export__status--${phase}`">
          <div class="g-export__icon-wrap">
            <RsLoading v-if="phase === 'exporting'" size="lg" />
            <GIcon v-else-if="phase === 'done'" size="32"><CheckmarkCircleOutline /></GIcon>
            <GIcon v-else-if="phase === 'error'" size="32"><CloseCircleOutline /></GIcon>
          </div>
          <p class="g-export__status-title">{{ statusTitle }}</p>
          <p class="g-export__status-desc">{{ statusText }}</p>
        </div>

        <div v-if="phase === 'done' && resolvedFilename" class="g-export__file-info">
          <GIcon size="16" class="g-export__file-icon"><DocumentOutline /></GIcon>
          <span class="g-export__file-name" :title="resolvedFilename">{{ resolvedFilename }}</span>
          <span v-if="fileSize" class="g-export__file-size">{{ fileSize }}</span>
        </div>

        <div v-if="phase === 'exporting' || phase === 'done'" class="g-export__progress-wrap">
          <div class="g-export__progress-header">
            <span class="g-export__progress-label">{{ progressLabel }}</span>
            <span class="g-export__progress-pct">
              {{ phase === 'done' ? '100%' : (!generating && progress > 0) ? `${progress}%` : '' }}
            </span>
          </div>
          <div class="g-export__progress" role="progressbar">
            <div
              class="g-export__progress-bar"
              :class="{ 'is-processing': phase === 'exporting' }"
              :style="{ width: (phase === 'done' ? 100 : (!generating && progress > 0 ? progress : 0)) + '%' }"
            />
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="g-export__footer">
        <template v-if="phase === 'idle'">
          <RsButton size="sm" @click="emit('update:visible', false)">取消</RsButton>
          <RsButton size="sm" variant="primary" @click="startExport">开始导出</RsButton>
        </template>
        <template v-else-if="phase === 'error'">
          <RsButton size="sm" variant="primary" @click="startExport">重试</RsButton>
          <RsButton size="sm" @click="emit('update:visible', false)">关闭</RsButton>
        </template>
        <RsButton
          v-else
          size="sm"
          :disabled="phase === 'exporting'"
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
import { RsButton, RsDialog, RsLoading } from '@/ui'
import { CheckmarkCircleOutline, CloseCircleOutline, DocumentOutline, DownloadOutline } from '@vicons/ionicons5'
import type { AxiosProgressEvent, AxiosResponse } from 'axios'
import { computed, ref, watch } from 'vue'
import type { GExportEmits, GExportProps } from './types'

defineOptions({ name: 'GExport' })

const props = defineProps<GExportProps>()
const emit = defineEmits<GExportEmits>()

// ─── 状态 ──────────────────────────────────────────────────────────────────────

type Phase = 'idle' | 'exporting' | 'done' | 'error'

const phase = ref<Phase>('idle')
const progress = ref(0)
const generating = ref(true) // true=服务端生成中，false=传输中
const errorMsg = ref('')
const resolvedFilename = ref('')
const fileSize = ref('')

// ─── 权限 ──────────────────────────────────────────────────────────────────────

const hasPermission = computed(() => {
  if (!props.moduleId) return true
  return store.user.hasPermission(`${props.moduleId}:export`)
})

// ─── 弹窗打开时重置到 idle ─────────────────────────────────────────────────────

watch(() => props.visible, (val) => {
  if (val) {
    phase.value = 'idle'
    progress.value = 0
    generating.value = true
    errorMsg.value = ''
    resolvedFilename.value = ''
    fileSize.value = ''
  }
})

// ─── 状态文案 ─────────────────────────────────────────────────────────────────

const statusTitle = computed(() => {
  switch (phase.value) {
    case 'exporting': return '正在导出'
    case 'done':      return '导出成功'
    case 'error':     return '导出失败'
    default:          return ''
  }
})

const statusText = computed(() => {
  switch (phase.value) {
    case 'exporting': return generating.value ? '服务端生成文件中，请稍候…' : '文件传输中，请稍候…'
    case 'done':      return '文件已自动下载到本地'
    case 'error':     return errorMsg.value || '请稍后重试或联系管理员'
    default:          return ''
  }
})

const progressLabel = computed(() => {
  if (phase.value === 'done') return '下载完成'
  if (phase.value === 'exporting') return generating.value ? '生成中' : `下载中`
  return ''
})

// ─── 工具函数 ─────────────────────────────────────────────────────────────────

function parseFilename(disposition: string | null | undefined): string {
  if (!disposition) return ''
  const utf8Match = disposition.match(/filename\*\s*=\s*UTF-8''([^;]+)/i)
  if (utf8Match) {
    try { return decodeURIComponent(utf8Match[1].trim()) } catch { /* fall through */ }
  }
  const plain = disposition.match(/filename\s*=\s*"?([^";]+)"?/i)
  if (plain) {
    try { return decodeURIComponent(plain[1].trim()) } catch { return plain[1].trim() }
  }
  return ''
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}

// ─── 导出逻辑 ─────────────────────────────────────────────────────────────────

const startExport = async () => {
  phase.value = 'exporting'
  progress.value = 0
  generating.value = true
  errorMsg.value = ''
  resolvedFilename.value = ''
  fileSize.value = ''

  try {
    const response = await service.request({
      method: 'POST',
      url: props.url,
      data: props.params ?? {},
      responseType: 'blob',
      timeout: props.timeout ?? 0,
      showLoading: false,
      headers: { 'Content-Type': 'application/json' },
      onDownloadProgress: (evt: AxiosProgressEvent) => {
        // 第一次收到进度事件说明响应体已开始到达
        generating.value = false
        if (evt.total && evt.total > 0) {
          progress.value = Math.round((evt.loaded / evt.total) * 100)
        }
      },
    } as any) as AxiosResponse<Blob>

    const blob: Blob = response.data ?? (response as unknown as Blob)
    if (!blob || blob.size === 0) throw new Error('服务端返回空文件')

    // 后端出错时可能返回 JSON 错误体（Content-Type: application/json），而非 xlsx
    const contentType = String(response.headers?.['content-type'] ?? '')
    if (contentType.includes('application/json')) {
      const text = await blob.text()
      let errMsg = '导出失败'
      try {
        const json = JSON.parse(text)
        errMsg = json?.errMsg || json?.msg || json?.message || json?.error || errMsg
      } catch { /* 非 JSON，使用原始文本 */ }
      throw new Error(errMsg)
    }

    const disposition = response.headers?.['content-disposition']
    let filename = parseFilename(disposition)
    if (!filename) {
      filename = props.filename ?? 'export'
      if (!filename.includes('.')) {
        const ext = blob.type.includes('spreadsheetml')
          ? '.xlsx'
          : blob.type.includes('csv')
            ? '.csv'
            : ''
        filename += ext
      }
    }

    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = filename
    document.body.appendChild(anchor)
    anchor.click()
    document.body.removeChild(anchor)
    URL.revokeObjectURL(url)

    resolvedFilename.value = filename
    fileSize.value = formatSize(blob.size)
    progress.value = 100
    phase.value = 'done'
    emit('success')
  } catch (err: any) {
    const error = err instanceof Error ? err : new Error(err?.message ?? '导出失败')
    errorMsg.value = error.message
    phase.value = 'error'
    emit('error', error)
  }
}
</script>

<style lang="scss" scoped>
.g-export__footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  width: 100%;
}

.g-export__body {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-lg);
  padding: var(--g-space-xl) var(--g-space-xl) var(--g-space-md);
}

/* ── 状态区域 ── */
.g-export__status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--g-space-sm);
  padding: var(--g-space-lg);
  border-radius: var(--g-radius-lg);
  background: var(--g-bg-secondary);
  transition: background var(--g-transition-base) var(--g-transition-ease);

  &--done  { background: color-mix(in srgb, var(--g-success) 8%, transparent); }
  &--error { background: color-mix(in srgb, var(--g-error)   8%, transparent); }
}

.g-export__icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--g-bg-primary);
  box-shadow: var(--g-shadow-sm);

  .g-export__status--idle      & { color: var(--g-primary); }
  .g-export__status--exporting & { color: var(--g-primary); }
  .g-export__status--done      & { color: var(--g-success); }
  .g-export__status--error     & { color: var(--g-error);   }
}

.g-export__status-title {
  margin: 0;
  font-size: var(--g-font-size-lg);
  font-weight: var(--g-font-weight-medium);
  color: var(--g-text-primary);

  .g-export__status--done  & { color: var(--g-success); }
  .g-export__status--error & { color: var(--g-error);   }
}

.g-export__status-desc {
  margin: 0;
  font-size: var(--g-font-size-sm);
  color: var(--g-text-tertiary);
  text-align: center;
  line-height: 1.6;
  max-width: 360px;
}

/* ── 文件信息条 ── */
.g-export__file-info {
  display: flex;
  align-items: center;
  gap: var(--g-space-xs);
  padding: var(--g-space-sm) var(--g-space-md);
  border-radius: var(--g-radius-md);
  background: var(--g-bg-tertiary);
  border: 1px solid var(--g-border-primary);
}

.g-export__file-icon {
  flex-shrink: 0;
  color: var(--g-text-secondary);
}

.g-export__file-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--g-font-size-sm);
  color: var(--g-text-primary);
  font-weight: var(--g-font-weight-medium);
}

.g-export__file-size {
  flex-shrink: 0;
  font-size: var(--g-font-size-xs);
  color: var(--g-text-tertiary);
}

/* ── 进度区域 ── */
.g-export__progress-wrap {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-xs);
}

.g-export__progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.g-export__progress-label {
  font-size: var(--g-font-size-xs);
  color: var(--g-text-secondary);
}

.g-export__progress-pct {
  font-size: var(--g-font-size-xs);
  font-weight: var(--g-font-weight-medium);
  color: var(--g-text-secondary);
  min-width: 36px;
  text-align: right;
}

.g-export__progress {
  width: 100%;
}
</style>
