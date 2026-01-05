/**
 * 静态服务统计面板类型定义
 * 与后端 models.StaticServerStats 对应
 */

/** 静态服务统计数据 */
export interface StaticServerStats {
  /** 总服务数 */
  totalServers: number
  /** 运行中服务数 */
  runningServers: number
  /** 已停止服务数 */
  stoppedServers: number
  /** 总连接数 */
  totalConnections: number
  /** 总接收字节数 */
  totalBytesReceived: number
  /** 总发送字节数 */
  totalBytesSent: number
}
