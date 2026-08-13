<template>
  <!--
    @deprecated 请直接使用 RsUpload（@/ui）。表单 type=file 已内置 RsUpload。
  -->
  <RsUpload
    :model-value="fileList"
    :disabled="disabled"
    :accept="config.accept"
    :max-count="config.max ?? 1"
    :max-size="config.maxSize"
    :label="config.uploadText"
    :hint="config.uploadDescription"
    :show-file-list="config.showFileList !== false"
    :show-download="showDownload"
    :hide-dropzone-when-full="(config.max ?? 1) === 1"
    @update:model-value="onUpdate"
    @reject="onReject"
    @remove="onRemove"
  />
</template>

<script setup lang="ts">
/**
 * @deprecated 使用 `RsUpload`。表单场景用 RsDataForm `type: 'file'`（值为 File[]）。
 */
import { useAppMessage } from '@/composables/useAppMessage'
import { RsUpload } from '@/ui'
import { onMounted } from 'vue'
import type { FileUploadConfig, GFileUploadEmits, GFileUploadProps } from './types'

defineOptions({
  name: 'GFileUpload',
})

const props = withDefaults(defineProps<GFileUploadProps>(), {
  fileList: () => [],
  config: () =>
    ({
      max: 1,
      maxSize: 10 * 1024 * 1024,
      showFileList: true,
    }) as FileUploadConfig,
  disabled: false,
  showDownload: false,
})

const emit = defineEmits<GFileUploadEmits>()
const message = useAppMessage()

onMounted(() => {
  if (import.meta.env.DEV) {
    console.warn('[GFileUpload] 已弃用：请改用 RsUpload；表单请用 type: "file"（File[]）。')
  }
})

const onUpdate = (value: File[]) => {
  emit('update:fileList', value)
  emit('change', value[0] ?? null)
  props.callbacks?.onChange?.(value[0] ?? null)
}

const onRemove = () => {
  emit('remove')
  props.callbacks?.onRemove?.()
}

const onReject = (errors: { reason: string }[]) => {
  const first = errors[0]
  if (!first) return
  const text =
    first.reason === 'accept'
      ? '文件类型不符合要求'
      : first.reason === 'maxSize'
        ? '文件大小超出限制'
        : '文件数量超出限制'
  message.error(text)
  props.callbacks?.onError?.(new Error(text))
  emit('error', new Error(text))
}
</script>
