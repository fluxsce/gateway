import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { onMounted, ref } from 'vue'
import { queryTools, type ToolCatalogItem } from './api'

/**
 * 进入工具页时再查一次目录。不在返回列表里的工具视为当前角色不可用。
 */
export function useGrantedTool(toolId: string) {
  const loading = ref(true)
  const denied = ref(false)
  const failed = ref('')
  const item = ref<ToolCatalogItem | null>(null)

  onMounted(async () => {
    try {
      const result = await queryTools()
      if (!isApiSuccess(result)) {
        failed.value = getApiMessage(result, '')
        return
      }
      const list = parseJsonData<ToolCatalogItem[]>(result, [])
      const found = list.find((row) => row.id === toolId) ?? null
      item.value = found
      denied.value = !found
    } catch (err) {
      failed.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  })

  return { loading, denied, failed, item }
}
