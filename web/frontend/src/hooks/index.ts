/**
 * 应用Hooks索引
 * 导出所有自定义钩子函数，便于统一导入
 */
import useRequest from './useRequest'
import useAsync, { useAsyncImmediate } from './useAsync'
import useUnwrappedRefs, { unwrapRefs } from './useRefUnwrapper'
import { nextTick, computed, type Ref } from 'vue'

export { useRequest, useAsync, useAsyncImmediate, useUnwrappedRefs, unwrapRefs }

/**
 * 安全的模态框管理Hook
 * 防止Naive UI z-index-manager错误
 */
export function useModalManager() {
  const createModalManager = (modals: Record<string, Ref<boolean>>) => {
    const isAnyModalOpen = computed(() => Object.values(modals).some((modal) => modal.value))

    // 安全地关闭所有模态框
    const closeAll = async () => {
      for (const [, modal] of Object.entries(modals)) {
        if (modal.value) {
          modal.value = false
          await nextTick()
        }
      }
    }

    // 安全地打开特定模态框
    const openModal = async (modalKey: string) => {
      if (!modals[modalKey]) {
        console.warn(`Modal "${modalKey}" not found in modal manager`)
        return
      }

      // 先关闭所有其他模态框
      await closeAll()
      await nextTick()

      // 打开指定模态框
      modals[modalKey].value = true
    }

    // 安全地关闭特定模态框
    const closeModal = async (modalKey: string) => {
      if (!modals[modalKey]) {
        console.warn(`Modal "${modalKey}" not found in modal manager`)
        return
      }

      modals[modalKey].value = false
      await nextTick()
    }

    return {
      isAnyModalOpen,
      closeAll,
      openModal,
      closeModal,
      modals,
    }
  }

  return {
    createModalManager,
  }
}
