/**
 * 审计日志模块类型定义
 * 与后端 web/views/hub0004/models/audit_log.go 保持一致
 */

/** 审计动作 */
export type AuditAction =
  | 'CREATE'
  | 'UPDATE'
  | 'DELETE'
  | 'ROLLBACK'
  | 'GRANT'
  | 'EXPORT'
  | 'LOGIN'
  | 'LOGIN_FAIL'
  | 'KICK'

/** 审计结果：Y 成功 / N 失败 */
export type AuditResult = 'Y' | 'N'

/** 权限审计日志 - 对应 AuthAuditLog */
export interface AuthAuditLog {
  auditId: string
  tenantId: string
  userId: string
  userName?: string | null
  action: AuditAction | string
  moduleCode?: string | null
  targetType?: string | null
  targetId?: string | null
  targetName?: string | null
  resourceCode?: string | null
  requestPath?: string | null
  requestMethod?: string | null
  clientIP?: string | null
  result: AuditResult | string
  detail?: string | null
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: string
}

/** 审计日志查询参数 */
export interface AuthAuditLogQueryParams {
  pageIndex: number
  pageSize: number
  auditId?: string
  userId?: string
  userName?: string
  action?: string
  moduleCode?: string
  targetType?: string
  targetId?: string
  targetName?: string
  resourceCode?: string
  clientIP?: string
  result?: string
  startTime?: string
  endTime?: string
}
