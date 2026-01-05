/**
 * GFileUpload 组件类型定义
 */

/**
 * 文件上传模式
 */
export type FileUploadMode = 'text' | 'base64' | 'binary'

/**
 * 文件上传配置
 */
export interface FileUploadConfig {
  /** 接受的文件类型（MIME类型或扩展名，如 ".crt,.pem"） */
  accept?: string

  /** 最大文件数量 */
  max?: number

  /** 最大文件大小（字节） */
  maxSize?: number

  /** 文件读取模式 */
  mode?: FileUploadMode

  /** 是否显示文件列表 */
  showFileList?: boolean

  /** 是否使用拖拽上传 */
  draggable?: boolean

  /** 上传提示文本 */
  uploadText?: string

  /** 上传提示描述 */
  uploadDescription?: string
}

/**
 * 文件信息
 */
export interface FileInfo {
  /** 文件ID */
  id: string

  /** 文件名 */
  name: string

  /** 文件大小（字节） */
  size?: number

  /** 文件类型 */
  type?: string

  /** 文件状态 */
  status?: 'pending' | 'uploading' | 'finished' | 'error'

  /** 文件内容（文本模式） */
  content?: string

  /** 文件内容（Base64模式） */
  base64?: string

  /** 文件对象（Binary模式） */
  file?: File
}

/**
 * GFileUpload 组件回调函数
 */
export interface GFileUploadCallbacks {
  /** 文件变化回调 */
  onChange?: (file: FileInfo | null) => void

  /** 文件下载回调 */
  onDownload?: (file: FileInfo) => void

  /** 文件移除回调 */
  onRemove?: () => void

  /** 文件上传错误回调 */
  onError?: (error: Error) => void
}

/**
 * GFileUpload 组件 Props
 */
export interface GFileUploadProps {
  /** 文件列表（v-model） */
  fileList?: FileInfo[]

  /** 文件上传配置 */
  config?: FileUploadConfig

  /** 是否禁用 */
  disabled?: boolean

  /** 标题 */
  title?: string

  /** 标题图标 */
  titleIcon?: any

  /** 标题图标颜色 */
  titleIconColor?: string

  /** 是否显示下载按钮（当文件存在时） */
  showDownload?: boolean

  /** 下载按钮文本 */
  downloadText?: string

  /** 回调函数配置 */
  callbacks?: GFileUploadCallbacks
}

/**
 * GFileUpload 组件事件
 */
export interface GFileUploadEmits {
  /** 文件列表更新 */
  (event: 'update:fileList', value: FileInfo[]): void

  /** 文件变化 */
  (event: 'change', file: FileInfo | null): void

  /** 文件移除 */
  (event: 'remove'): void

  /** 文件下载 */
  (event: 'download', file: FileInfo): void

  /** 文件上传错误 */
  (event: 'error', error: Error): void
}

