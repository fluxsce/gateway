/**
 * 应用配置文件
 * 用于加载和管理环境变量，提供全局配置访问
 */

/**
 * 系统配置接口
 */
export interface SystemConfig {
  /** 应用标题 */
  appTitle: string
  /** 应用版本号 */
  appVersion: string
  /** 应用基础路径 */
  baseUrl: string
  /** API基础URL */
  apiBaseUrl: string
  /** API请求超时时间(ms) */
  apiTimeout: number
  /** 是否启用Mock数据 */
  useMock: boolean
  /** 是否为开发环境 */
  isDevelopment: boolean
  /** 是否为生产环境 */
  isProduction: boolean
}

/**
 * 获取环境变量
 * @param key 环境变量键名
 * @param defaultValue 默认值
 * @returns 环境变量值
 */
function getEnv(key: string, defaultValue: string = ''): string {
  return import.meta.env[key] || defaultValue
}

/**
 * 获取布尔类型环境变量
 * @param key 环境变量键名
 * @param defaultValue 默认值
 * @returns 布尔值
 */
function getEnvBoolean(key: string, defaultValue: boolean = false): boolean {
  const value = getEnv(key)
  if (value === '') return defaultValue
  return value === 'true' || value === '1'
}

/**
 * 获取数字类型环境变量
 * @param key 环境变量键名
 * @param defaultValue 默认值
 * @returns 数字值
 */
function getEnvNumber(key: string, defaultValue: number = 0): number {
  const value = getEnv(key)
  if (value === '') return defaultValue
  const num = Number(value)
  return isNaN(num) ? defaultValue : num
}

/**
 * 系统配置对象
 */
export const config: SystemConfig = {
  appTitle: getEnv('VITE_APP_TITLE', 'Gateway Web'),
  appVersion: getEnv('VITE_APP_VERSION', '1.0.0'),
  baseUrl: getEnv('VITE_BASE_URL', '/'),
  apiBaseUrl: getEnv('VITE_API_BASE_URL', '/api'),
  apiTimeout: getEnvNumber('VITE_API_TIMEOUT', 30000),
  useMock: getEnvBoolean('VITE_USE_MOCK', false),
  isDevelopment: import.meta.env.DEV,
  isProduction: import.meta.env.PROD,
}

/**
 * 导出默认配置
 */
export default config
