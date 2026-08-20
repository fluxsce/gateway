import { useAppMessage } from '@/composables/useAppMessage'
import { computed, ref } from 'vue'
import {
  addRouteAssertion,
  deleteRouteAssertion,
  editRouteAssertion,
  queryRouteAssertions,
} from '../api'
import type { RouteAssertion, RouteAssertionForm } from '../types'

/**
 * 路由断言管理Hook
 */
export function useRouteAssertionManagement(routeConfigId: string) {
  const message = useAppMessage()

  // 状态管理
  const assertions = ref<RouteAssertion[]>([])
  const loading = ref(false)
  const dialogVisible = ref(false)
  const currentAssertion = ref<RouteAssertion | null>(null)

  // 计算属性
  const sortedAssertions = computed(() => {
    return [...assertions.value].sort((a, b) => a.assertionOrder - b.assertionOrder)
  })

  const hasAssertions = computed(() => assertions.value.length > 0)

  const enabledAssertions = computed(() => {
    return assertions.value.filter((assertion) => assertion.activeFlag === 'Y')
  })

  const disabledAssertions = computed(() => {
    return assertions.value.filter((assertion) => assertion.activeFlag === 'N')
  })

  /**
   * 加载路由断言列表
   */
  const loadAssertions = async () => {
    if (!routeConfigId) {
      return
    }

    loading.value = true
    try {
      const response = await queryRouteAssertions({ routeConfigId })

      if (response?.oK && response.bizData) {
        const parseData = JSON.parse(response.bizData) || []
        assertions.value = parseData.sort(
          (a: RouteAssertion, b: RouteAssertion) => a.assertionOrder - b.assertionOrder,
        )
      } else {
        assertions.value = []
      }
    } catch (error) {
      message.error('加载路由断言失败')
      assertions.value = []
    } finally {
      loading.value = false
    }
  }

  /**
   * 创建路由断言
   */
  const createAssertion = async (assertionData: RouteAssertionForm): Promise<boolean> => {
    try {
      const saveData = {
        routeConfigId,
        ...assertionData,
        addWho: 'admin',
        editWho: 'admin',
        oprSeqFlag: Date.now().toString(),
        currentVersion: 1,
      }

      const response = await addRouteAssertion(saveData)

      if (response?.oK) {
        message.success('创建断言成功')
        await loadAssertions()
        return true
      } else {
        message.error(response?.errMsg || '创建断言失败')
        return false
      }
    } catch (error) {
      message.error('创建断言失败')
      return false
    }
  }

  /**
   * 更新路由断言
   */
  const updateAssertion = async (
    routeAssertionId: string,
    assertionData: Partial<RouteAssertionForm>,
  ): Promise<boolean> => {
    try {
      const updateData = {
        routeAssertionId,
        routeConfigId,
        ...assertionData,
        editWho: 'admin',
      }

      const response = await editRouteAssertion(updateData)

      if (response?.oK) {
        message.success('更新断言成功')
        await loadAssertions()
        return true
      } else {
        message.error(response?.errMsg || '更新断言失败')
        return false
      }
    } catch (error) {
      message.error('更新断言失败')
      return false
    }
  }

  /**
   * 删除路由断言
   */
  const removeAssertion = async (routeAssertionId: string): Promise<boolean> => {
    try {
      const response = await deleteRouteAssertion(routeAssertionId)

      if (response?.oK) {
        message.success('删除断言成功')
        await loadAssertions()
        return true
      } else {
        message.error(response?.errMsg || '删除断言失败')
        return false
      }
    } catch (error) {
      message.error('删除断言失败')
      return false
    }
  }

  /**
   * 切换断言状态
   */
  const toggleAssertionStatus = async (assertion: RouteAssertion): Promise<boolean> => {
    const newStatus = assertion.activeFlag === 'Y' ? 'N' : 'Y'
    const success = await updateAssertion(assertion.routeAssertionId, {
      activeFlag: newStatus,
    })

    if (success) {
      message.success(`断言已${newStatus === 'Y' ? '启用' : '禁用'}`)
    }

    return success
  }

  /**
   * 调整断言顺序
   */
  const swapAssertionOrder = async (
    assertion1: RouteAssertion,
    assertion2: RouteAssertion,
  ): Promise<boolean> => {
    try {
      loading.value = true

      const tempOrder = assertion1.assertionOrder

      // 并行更新两个断言的顺序
      const [response1, response2] = await Promise.all([
        updateAssertion(assertion1.routeAssertionId, {
          assertionOrder: assertion2.assertionOrder,
        }),
        updateAssertion(assertion2.routeAssertionId, {
          assertionOrder: tempOrder,
        }),
      ])

      if (response1 && response2) {
        message.success('调整执行顺序成功')
        return true
      } else {
        message.error('调整执行顺序失败')
        return false
      }
    } catch (error) {
      message.error('调整执行顺序失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 向上移动断言
   */
  const moveAssertionUp = async (assertion: RouteAssertion): Promise<boolean> => {
    const sorted = sortedAssertions.value
    const currentIndex = sorted.findIndex((a) => a.routeAssertionId === assertion.routeAssertionId)

    if (currentIndex <= 0) {
      message.warning('已经是第一个断言，无法继续向上移动')
      return false
    }

    const targetAssertion = sorted[currentIndex - 1]
    return await swapAssertionOrder(assertion, targetAssertion)
  }

  /**
   * 向下移动断言
   */
  const moveAssertionDown = async (assertion: RouteAssertion): Promise<boolean> => {
    const sorted = sortedAssertions.value
    const currentIndex = sorted.findIndex((a) => a.routeAssertionId === assertion.routeAssertionId)

    if (currentIndex >= sorted.length - 1) {
      message.warning('已经是最后一个断言，无法继续向下移动')
      return false
    }

    const targetAssertion = sorted[currentIndex + 1]
    return await swapAssertionOrder(assertion, targetAssertion)
  }

  /**
   * 打开创建对话框
   */
  const openCreateDialog = () => {
    currentAssertion.value = null
    dialogVisible.value = true
  }

  /**
   * 打开编辑对话框
   */
  const openEditDialog = (assertion: RouteAssertion) => {
    currentAssertion.value = assertion
    dialogVisible.value = true
  }

  /**
   * 关闭对话框
   */
  const closeDialog = () => {
    dialogVisible.value = false
    currentAssertion.value = null
  }

  /**
   * 保存断言（创建或更新）
   */
  const saveAssertion = async (assertionData: RouteAssertionForm): Promise<boolean> => {
    let success = false

    if (currentAssertion.value?.routeAssertionId) {
      // 编辑模式
      success = await updateAssertion(currentAssertion.value.routeAssertionId, assertionData)
    } else {
      // 新建模式
      success = await createAssertion(assertionData)
    }

    if (success) {
      closeDialog()
    }

    return success
  }

  /**
   * 刷新断言列表
   */
  const refreshAssertions = async () => {
    await loadAssertions()
    message.success('刷新完成')
  }

  /**
   * 获取断言统计信息
   */
  const getAssertionStats = () => {
    return {
      total: assertions.value.length,
      enabled: enabledAssertions.value.length,
      disabled: disabledAssertions.value.length,
      byType: {
        PATH: assertions.value.filter((a) => a.assertionType === 'PATH').length,
        HEADER: assertions.value.filter((a) => a.assertionType === 'HEADER').length,
        QUERY: assertions.value.filter((a) => a.assertionType === 'QUERY').length,
        COOKIE: assertions.value.filter((a) => a.assertionType === 'COOKIE').length,
        IP: assertions.value.filter((a) => a.assertionType === 'IP').length,
      },
    }
  }

  /**
   * 验证断言配置
   */
  const validateAssertion = (assertionData: RouteAssertionForm): string[] => {
    const errors: string[] = []

    if (!assertionData.assertionName?.trim()) {
      errors.push('断言名称不能为空')
    }

    if (!assertionData.assertionType) {
      errors.push('请选择断言类型')
    }

    if (!assertionData.assertionOperator) {
      errors.push('请选择操作符')
    }

    // 检查字段名称是否必需
    const needsFieldName = ['HEADER', 'QUERY', 'COOKIE'].includes(assertionData.assertionType)
    if (needsFieldName && !assertionData.fieldName?.trim()) {
      errors.push('字段名称不能为空')
    }

    // 检查期望值或匹配模式
    const needsExpectedValue = [
      'EQUAL',
      'NOT_EQUAL',
      'CONTAINS',
      'NOT_CONTAINS',
      'STARTS_WITH',
      'ENDS_WITH',
      'IN',
      'NOT_IN',
    ].includes(assertionData.assertionOperator)
    const needsPatternValue = ['MATCHES', 'NOT_MATCHES'].includes(assertionData.assertionOperator)

    if (needsExpectedValue && !assertionData.expectedValue?.trim()) {
      errors.push('期望值不能为空')
    }

    if (needsPatternValue && !assertionData.patternValue?.trim()) {
      errors.push('匹配模式不能为空')
    }

    // 验证正则表达式
    if (needsPatternValue && assertionData.patternValue?.trim()) {
      try {
        new RegExp(assertionData.patternValue.trim())
      } catch {
        errors.push('匹配模式不是有效的正则表达式')
      }
    }

    return errors
  }

  return {
    // 状态
    assertions: sortedAssertions,
    loading,
    dialogVisible,
    currentAssertion,

    // 计算属性
    hasAssertions,
    enabledAssertions,
    disabledAssertions,

    // 方法
    loadAssertions,
    createAssertion,
    updateAssertion,
    removeAssertion,
    toggleAssertionStatus,
    moveAssertionUp,
    moveAssertionDown,
    swapAssertionOrder,
    openCreateDialog,
    openEditDialog,
    closeDialog,
    saveAssertion,
    refreshAssertions,
    getAssertionStats,
    validateAssertion,
  }
}
