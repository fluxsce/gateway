// ─── 公共类型 ─────────────────────────────────────────────────────────────────

/**
 * 按钮尺寸
 */
export type GExportImportSize = 'tiny' | 'small' | 'medium' | 'large'

// ─── GExport ─────────────────────────────────────────────────────────────────

/**
 * GExport 组件 Props
 *
 * 权限码约定：{moduleId}:export
 *
 * @example
 * ```vue
 * <GExport
 *   module-id="hub0020"
 *   url="/hub0020/instance/exportGatewayInstance"
 *   :params="{ gatewayInstanceId: id }"
 *   filename="网关实例配置"
 * />
 * ```
 */
export interface GExportProps {
  /**
   * 模块ID，用于权限控制。
   * 导出权限码为 `{moduleId}:export`。
   * 不传则跳过权限检查。
   */
  moduleId?: string

  /**
   * 控制弹窗显隐（v-model:visible）。
   * 设为 true 时立即打开弹窗并开始导出；关闭时组件自动回写 false。
   */
  visible?: boolean

  /**
   * 导出请求地址（POST）
   */
  url: string

  /**
   * 请求参数，作为 JSON body 发送
   */
  params?: Record<string, any>

  /**
   * 导出文件的默认文件名（不含扩展名）
   * @default 'export'
   */
  filename?: string

  /**
   * 请求超时（ms），0 表示不限制
   * @default 0
   */
  timeout?: number

  /**
   * 进度弹窗标题
   * @default '导出'
   */
  dialogTitle?: string
}

/**
 * GExport 组件 Emits
 */
export interface GExportEmits {
  /** v-model:visible 回写（弹窗关闭时置 false） */
  (event: 'update:visible', value: boolean): void
  /** 导出成功 */
  (event: 'success'): void
  /** 导出失败 */
  (event: 'error', error: Error): void
}

// ─── GImport ─────────────────────────────────────────────────────────────────

/**
 * GImport 组件 Props
 *
 * 权限码约定：{moduleId}:import
 *
 * @example
 * ```vue
 * <GImport
 *   v-model:visible="importVisible"
 *   module-id="hub0020"
 *   url="/hub0020/instance/importGatewayInstance"
 *   :params="{ gatewayInstanceId: id }"
 *   @success="handleImportSuccess"
 * />
 * ```
 */
export interface GImportProps {
  /**
   * 控制弹窗显隐（v-model:visible）。
   * 设为 true 时打开弹窗，用户在弹窗内选择文件并手动点击「开始导入」触发上传。
   */
  visible?: boolean

  /**
   * 模块ID，用于权限控制。
   * 导入权限码为 `{moduleId}:import`。
   * 不传则跳过权限检查。
   */
  moduleId?: string

  /**
   * 导入请求地址（POST multipart/form-data）
   */
  url: string

  /**
   * 上传字段名
   * @default 'file'
   */
  fieldName?: string

  /**
   * 额外表单参数（随文件一起上传）
   */
  params?: Record<string, any>

  /**
   * 允许的文件类型
   * @default '.xlsx,.xls'
   */
  accept?: string

  /**
   * 最大文件大小（字节）
   * @default 20971520 (20MB)
   */
  maxSize?: number

  /**
   * 弹窗标题
   * @default '导入'
   */
  dialogTitle?: string
}

/**
 * GImport 组件 Emits
 */
export interface GImportEmits {
  /** v-model:visible 回写（弹窗关闭时置 false） */
  (event: 'update:visible', value: boolean): void
  /** 导入成功，payload 为服务端返回的完整响应 */
  (event: 'success', data: any): void
  /** 导入失败 */
  (event: 'error', error: Error): void
}
