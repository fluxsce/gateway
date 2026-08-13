/**
 * @deprecated 请改用 RsUpload（File[]）。表单 type=file 的值为 File[]。
 */

/** @deprecated */
export type FileUploadMode = 'text' | 'base64' | 'binary'

/** @deprecated 映射到 RsUpload props */
export interface FileUploadConfig {
  accept?: string
  max?: number
  maxSize?: number
  mode?: FileUploadMode
  showFileList?: boolean
  uploadText?: string
  uploadDescription?: string
}

/** @deprecated 现为 File，保留别名避免旧导入报错 */
export type FileInfo = File

/** @deprecated */
export interface GFileUploadCallbacks {
  onChange?: (file: File | null) => void
  onDownload?: (file: File) => void
  onRemove?: () => void
  onError?: (error: Error) => void
}

/** @deprecated */
export interface GFileUploadProps {
  fileList?: File[]
  config?: FileUploadConfig
  disabled?: boolean
  title?: string
  titleIcon?: any
  titleIconColor?: string
  showDownload?: boolean
  downloadText?: string
  callbacks?: GFileUploadCallbacks
}

/** @deprecated */
export interface GFileUploadEmits {
  (event: 'update:fileList', value: File[]): void
  (event: 'change', file: File | null): void
  (event: 'remove'): void
  (event: 'download', file: File): void
  (event: 'error', error: Error): void
}
