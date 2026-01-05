<template>
  <n-card v-if="showCard" size="small" class="g-file-upload-card">
    <template #header v-if="title || $slots.title">
      <n-space align="center" justify="space-between">
        <n-space align="center">
          <n-icon v-if="titleIcon" :size="16" :color="titleIconColor">
            <component :is="titleIcon" />
          </n-icon>
          <span>{{ title }}</span>
        </n-space>
        <n-button
          v-if="showDownload && hasFile"
          text
          size="small"
          @click="handleDownload"
          :title="downloadText || '下载文件'"
        >
          <template #icon>
            <n-icon size="16">
              <DownloadOutline />
            </n-icon>
          </template>
        </n-button>
      </n-space>
    </template>

    <n-upload
      :file-list="internalFileList"
      :max="config.max ?? 1"
      :accept="config.accept"
      :show-file-list="config.showFileList !== false"
      :disabled="disabled"
      :custom-request="handleCustomRequest"
      @change="handleFileChange"
      @remove="handleFileRemove"
      class="g-file-upload"
    >
      <n-upload-dragger v-if="config.draggable !== false">
        <div class="upload-content">
          <n-icon size="32" depth="3">
            <DocumentOutline />
          </n-icon>
          <n-text depth="3">{{ config.uploadText || '点击或拖拽上传文件' }}</n-text>
          <n-text v-if="config.uploadDescription" depth="3" style="font-size: 12px;">
            {{ config.uploadDescription }}
          </n-text>
        </div>
      </n-upload-dragger>
    </n-upload>
  </n-card>

  <n-upload
    v-else
    :file-list="internalFileList"
    :max="config.max ?? 1"
    :accept="config.accept"
    :show-file-list="config.showFileList !== false"
    :disabled="disabled"
    :custom-request="handleCustomRequest"
    @change="handleFileChange"
    @remove="handleFileRemove"
    class="g-file-upload"
  >
    <n-upload-dragger v-if="config.draggable !== false">
      <div class="upload-content">
        <n-icon size="32" depth="3">
          <DocumentOutline />
        </n-icon>
        <n-text depth="3">{{ config.uploadText || '点击或拖拽上传文件' }}</n-text>
        <n-text v-if="config.uploadDescription" depth="3" style="font-size: 12px;">
          {{ config.uploadDescription }}
        </n-text>
      </div>
    </n-upload-dragger>
  </n-upload>
</template>

<script setup lang="ts">
import { DocumentOutline, DownloadOutline } from '@vicons/ionicons5'
import { NButton, NCard, NIcon, NSpace, NText, NUpload, NUploadDragger, useMessage } from 'naive-ui'
import { computed, ref, watch } from 'vue'
import type { FileInfo, FileUploadConfig, GFileUploadEmits, GFileUploadProps } from './types'

defineOptions({
  name: 'GFileUpload'
})

const props = withDefaults(defineProps<GFileUploadProps>(), {
  fileList: () => [],
  config: () => ({
    mode: 'text',
    max: 1,
    maxSize: 10 * 1024 * 1024, // 10MB
    showFileList: true,
    draggable: true,
  } as FileUploadConfig),
  disabled: false,
  showDownload: false,
  downloadText: '下载',
})

const emit = defineEmits<GFileUploadEmits>()

const message = useMessage()

// 内部文件列表
const internalFileList = ref<any[]>([])

// 是否显示卡片包装
const showCard = computed(() => !!props.title || !!props.titleIcon)

// 是否有文件
const hasFile = computed(() => {
  return internalFileList.value.length > 0 && props.fileList && props.fileList.length > 0
})

// 读取文件为文本
const readFileAsText = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => resolve(e.target?.result as string)
    reader.onerror = reject
    reader.readAsText(file)
  })
}

// 读取文件为Base64
const readFileAsBase64 = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const result = e.target?.result as string
      // 移除 data:xxx;base64, 前缀
      const base64 = result.split(',')[1] || result
      resolve(base64)
    }
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

// 处理自定义上传请求
const handleCustomRequest = () => {
  // 不进行实际的上传，只读取文件内容
}

// 处理文件变化
const handleFileChange = async (options: { file: any; fileList: any[] }) => {
  const { file, fileList } = options

  if (file.file) {
    try {
      // 文件大小检查
      if (props.config.maxSize && file.file.size > props.config.maxSize) {
        message.error(`文件大小不能超过 ${(props.config.maxSize / 1024 / 1024).toFixed(2)}MB`)
        return
      }

      const mode = props.config.mode || 'text'
      let content: string | undefined
      let base64: string | undefined

      // 根据模式读取文件
      if (mode === 'text') {
        content = await readFileAsText(file.file)
      } else if (mode === 'base64') {
        base64 = await readFileAsBase64(file.file)
      }

      // 创建文件信息
      const fileInfo: FileInfo = {
        id: file.id || `file-${Date.now()}`,
        name: file.file.name,
        size: file.file.size,
        type: file.file.type,
        status: 'finished',
        content,
        base64,
        file: mode === 'binary' ? file.file : undefined,
      }

      // 更新文件列表
      const newFileList = [fileInfo]
      internalFileList.value = fileList
      emit('update:fileList', newFileList)
      emit('change', fileInfo)
      
      // 调用外部传入的 onChange 回调（如果存在）
      props.callbacks?.onChange?.(fileInfo)
      
      message.success('文件上传成功')
    } catch (error) {
      message.error('文件读取失败')
      emit('error', error as Error)
      
      // 调用外部传入的 onError 回调（如果存在）
      props.callbacks?.onError?.(error as Error)
      
      console.error('Error reading file:', error)
    }
  }
}

// 处理文件移除
const handleFileRemove = () => {
  internalFileList.value = []
  emit('update:fileList', [])
  emit('remove')
  
  // 调用外部传入的 onRemove 回调（如果存在）
  props.callbacks?.onRemove?.()
  
  message.info('已移除文件')
}

// 处理文件下载
const handleDownload = () => {
  if (!props.fileList || props.fileList.length === 0) {
    message.warning('没有可下载的文件')
    return
  }

  const fileInfo = props.fileList[0]
  emit('download', fileInfo)

  // 调用外部传入的 onDownload 回调（如果存在）
  if (props.callbacks?.onDownload) {
    props.callbacks.onDownload(fileInfo)
    return // 如果外部有处理，就不执行默认逻辑
  }

  // 如果没有监听 download 事件，执行默认下载逻辑
  if (fileInfo.content) {
    const filename = fileInfo.name || 'file.txt'
    const blob = new Blob([fileInfo.content], { type: fileInfo.type || 'text/plain' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    message.success('文件下载成功')
  } else {
    message.warning('文件内容不可用')
  }
}

// 同步外部文件列表
watch(
  () => props.fileList,
  (newFileList) => {
    if (newFileList && newFileList.length > 0) {
      internalFileList.value = newFileList.map((file) => ({
        id: file.id,
        name: file.name,
        status: file.status || 'finished',
        type: file.type,
      }))
    } else {
      internalFileList.value = []
    }
  },
  { immediate: true, deep: true }
)
</script>

<style scoped lang="scss">
.g-file-upload-card {
  height: 100%;

  :deep(.n-card-header) {
    padding: 12px 16px;
  }

  :deep(.n-card__content) {
    padding: 12px;
  }
}

.g-file-upload {
  width: 100%;
}

.upload-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 20px;
}

// 下载按钮样式
:deep(.n-button--text-type) {
  color: #18a058;
  transition: all 0.2s ease;

  &:hover {
    color: #0c7a43;
    background-color: rgba(24, 160, 88, 0.1);
  }
}
</style>

