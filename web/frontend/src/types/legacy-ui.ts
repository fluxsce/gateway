/**
 * 迁移期占位类型：原 naive-ui 类型已移除，业务侧逐步收紧为具体结构。
 */
export type TreeOption = {
  key?: string | number
  label?: string
  children?: TreeOption[]
  disabled?: boolean
  isLeaf?: boolean
  [key: string]: unknown
}

export type ButtonProps = Record<string, unknown>
export type CardProps = Record<string, unknown>
export type FormRules = Record<string, unknown>
export type FormItemRule = Record<string, unknown>
export type UploadFileInfo = {
  id?: string
  name: string
  status?: string
  file?: File
  url?: string
  [key: string]: unknown
}
export type UploadCustomRequestOptions = {
  file: UploadFileInfo
  onFinish: () => void
  onError: () => void
  onProgress?: (e: { percent: number }) => void
}
