/**
 * 路由本机目录托管配置
 * 对应数据库表：HUB_GW_STATIC_HOST_CONFIG
 */
export interface StaticHostConfig {
  tenantId?: string
  staticHostConfigId?: string
  routeConfigId: string
  configName?: string
  rootDirectory: string
  stripRoutePrefix: 'Y' | 'N'
  indexFiles?: string
  rewriteRules?: string
  spaFallback: 'Y' | 'N'
  cacheControlMaxAge: number
  allowedExtensions?: string
  maxFileSizeBytes?: number
  followSymlinks?: 'Y' | 'N'
  enablePrecompress?: 'Y' | 'N'
  redirectDirectorySlash?: 'Y' | 'N'
  rootTokenExact?: 'Y' | 'N'
  fallbackRoots?: string
  cacheControlByExt?: string
  enableGzip?: 'Y' | 'N'
  securityHeaders?: string
  errorPage404?: string
  errorPage403?: string
  noteText?: string
}
