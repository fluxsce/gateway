/**
 * 静态服务管理相关 hooks
 */

// Model + Service + Page 架构
export { useStaticServerModel, type StaticServerModel } from './model'
export { useStaticServerPage, type StaticServerPage } from './page'
export { useStaticServerService, type StaticServerService } from './service'

// 常量导出
export {
    ACTIVE_FLAG_OPTIONS,
    HEALTH_CHECK_TYPE_OPTIONS,
    LOAD_BALANCE_OPTIONS,
    SERVER_STATUS_OPTIONS,
    SERVER_TYPE_OPTIONS
} from './model'

