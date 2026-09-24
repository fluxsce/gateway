import { createApi } from '@/api/request'
import { moduleApiPrefix } from '@/api/requestPath'
import type { JsonDataObj } from '@/types/api'

/**
 * 查询集群节点，并探测调用方给出的监听口。
 * moduleId 决定打到哪个模块的接口，权限按该模块的按钮校验。
 * params 由各页面自己组，共用函数不规定字段名。
 */
export async function queryClusterTopology(
  moduleId: string,
  params: Record<string, unknown>,
): Promise<JsonDataObj> {
  return createApi(moduleApiPrefix(moduleId)).post('/queryClusterTopology', params)
}
