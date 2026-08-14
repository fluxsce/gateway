import { useAppMessage } from '@/composables/useAppMessage'
import { computed, getCurrentInstance, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  addRouteAssertion,
  deleteRouteAssertion,
  editRouteAssertion,
  queryRouteAssertions,
} from '../api'
import {
  AssertionOperator,
  AssertionType,
  type RouteAssertion,
  type RouteAssertionForm,
} from '../types'

interface UseRouteAssertionListOptions {
  routeConfigId: string
}

export function useRouteAssertionList(options: UseRouteAssertionListOptions) {
  const message = useAppMessage()

  // 获取当前组件实例，用于检查组件是否还存在
  const instance = getCurrentInstance()

  // 组件状态
  const loading = ref(false)
  const dialogVisible = ref(false)
  const currentAssertion = ref<RouteAssertion | null>(null)
  const assertions = ref<RouteAssertion[]>([])
  const isUnmounted = ref(false)

  // 安全的消息提示函数
  const safeMessage = {
    success: (msg: string) => {
      if (!isUnmounted.value && instance && !instance.isUnmounted) {
        nextTick(() => {
          if (!isUnmounted.value) {
            message.success(msg)
          }
        })
      }
    },
    error: (msg: string) => {
      if (!isUnmounted.value && instance && !instance.isUnmounted) {
        nextTick(() => {
          if (!isUnmounted.value) {
            message.error(msg)
          }
        })
      }
    },
    warning: (msg: string) => {
      if (!isUnmounted.value && instance && !instance.isUnmounted) {
        nextTick(() => {
          if (!isUnmounted.value) {
            message.warning(msg)
          }
        })
      }
    },
  }

  // 按执行顺序排序的断言列表
  const sortedAssertions = computed(() => {
    return [...assertions.value].sort((a, b) => a.assertionOrder - b.assertionOrder)
  })

  /**
   * 获取断言类型颜色
   */
  const getAssertionTypeColor = (
    type: AssertionType,
  ): 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' => {
    const colorMap: Record<
      AssertionType,
      'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'
    > = {
      [AssertionType.PATH]: 'primary',
      [AssertionType.HEADER]: 'info',
      [AssertionType.QUERY]: 'success',
      [AssertionType.COOKIE]: 'warning',
      [AssertionType.IP]: 'error',
      [AssertionType.BODY_CONTENT]: 'success',
    }
    return colorMap[type] || 'default'
  }

  /**
   * 获取断言类型标签
   */
  const getAssertionTypeLabel = (type: AssertionType): string => {
    const labelMap: Record<AssertionType, string> = {
      [AssertionType.PATH]: '路径',
      [AssertionType.HEADER]: '请求头',
      [AssertionType.QUERY]: '查询参数',
      [AssertionType.COOKIE]: 'Cookie',
      [AssertionType.IP]: 'IP地址',
      [AssertionType.BODY_CONTENT]: '请求体内容',
    }
    return labelMap[type] || type
  }

  /**
   * 获取操作符标签
   */
  const getOperatorLabel = (operator: AssertionOperator): string => {
    const labelMap: Record<AssertionOperator, string> = {
      [AssertionOperator.EQUAL]: '等于',
      [AssertionOperator.NOT_EQUAL]: '不等于',
      [AssertionOperator.CONTAINS]: '包含',
      [AssertionOperator.NOT_CONTAINS]: '不包含',
      [AssertionOperator.MATCHES]: '正则匹配',
      [AssertionOperator.NOT_MATCHES]: '正则不匹配',
      [AssertionOperator.STARTS_WITH]: '开头匹配',
      [AssertionOperator.ENDS_WITH]: '结尾匹配',
      [AssertionOperator.IN]: '在列表中',
      [AssertionOperator.NOT_IN]: '不在列表中',
    }
    return labelMap[operator] || operator
  }

  /**
   * 判断是否为第一个断言
   */
  const isFirstAssertion = (assertion: RouteAssertion): boolean => {
    const sorted = sortedAssertions.value
    return sorted.length > 0 && sorted[0].routeAssertionId === assertion.routeAssertionId
  }

  /**
   * 判断是否为最后一个断言
   */
  const isLastAssertion = (assertion: RouteAssertion): boolean => {
    const sorted = sortedAssertions.value
    return (
      sorted.length > 0 && sorted[sorted.length - 1].routeAssertionId === assertion.routeAssertionId
    )
  }

  /**
   * 加载路由断言列表
   */
  const loadRouteAssertions = async () => {
    if (!options.routeConfigId || isUnmounted.value) return

    loading.value = true
    try {
      const response = await queryRouteAssertions({ routeConfigId: options.routeConfigId })

      // 检查组件是否还存在
      if (isUnmounted.value) return

      if (response?.oK && response.bizData) {
        const parseData = JSON.parse(response.bizData) || []
        assertions.value = parseData.sort(
          (a: RouteAssertion, b: RouteAssertion) => a.assertionOrder - b.assertionOrder,
        )
      } else {
        assertions.value = []
      }
    } catch (error) {
      if (!isUnmounted.value) {
        console.error('加载路由断言失败:', error)
        safeMessage.error('加载路由断言失败')
        assertions.value = []
      }
    } finally {
      if (!isUnmounted.value) {
        loading.value = false
      }
    }
  }

  /**
   * 处理创建
   */
  const handleCreate = () => {
    if (isUnmounted.value) return
    currentAssertion.value = null
    dialogVisible.value = true
  }

  /**
   * 处理编辑
   */
  const handleEdit = (assertion: RouteAssertion) => {
    if (isUnmounted.value) return
    currentAssertion.value = assertion
    dialogVisible.value = true
  }

  /**
   * 处理删除
   */
  const handleDelete = async (assertion: RouteAssertion) => {
    if (isUnmounted.value) return

    try {
      loading.value = true
      const response = await deleteRouteAssertion(assertion.routeAssertionId)

      if (isUnmounted.value) return

      if (response?.oK) {
        safeMessage.success('删除断言成功')
        await loadRouteAssertions()
      } else {
        safeMessage.error(response?.errMsg || '删除断言失败')
      }
    } catch (error) {
      if (!isUnmounted.value) {
        console.error('删除断言失败:', error)
        safeMessage.error('删除断言失败')
      }
    } finally {
      if (!isUnmounted.value) {
        loading.value = false
      }
    }
  }

  /**
   * 处理状态切换
   */
  const handleToggleStatus = async (assertion: RouteAssertion) => {
    if (isUnmounted.value) return

    try {
      const newStatus = assertion.activeFlag === 'Y' ? 'N' : 'Y'
      const updateData: Partial<RouteAssertion> = {
        ...assertion,
        activeFlag: newStatus,
      }

      const response = await editRouteAssertion({
        routeAssertionId: updateData.routeAssertionId!,
        routeConfigId: updateData.routeConfigId!,
        assertionName: updateData.assertionName!,
        assertionType: updateData.assertionType!,
        assertionOperator: updateData.assertionOperator!,
        fieldName: updateData.fieldName || '',
        expectedValue: updateData.expectedValue || '',
        patternValue: updateData.patternValue || '',
        caseSensitive: updateData.caseSensitive!,
        assertionOrder: updateData.assertionOrder!,
        isRequired: updateData.isRequired!,
        assertionDesc: updateData.assertionDesc || '',
        activeFlag: newStatus,
        noteText: updateData.noteText || '',
      })

      if (isUnmounted.value) return

      if (response?.oK) {
        safeMessage.success(`断言已${newStatus === 'Y' ? '启用' : '禁用'}`)
        await loadRouteAssertions()
      } else {
        safeMessage.error(response?.errMsg || '更新断言状态失败')
      }
    } catch (error) {
      if (!isUnmounted.value) {
        console.error('更新断言状态失败:', error)
        safeMessage.error('更新断言状态失败')
      }
    }
  }

  /**
   * 处理向上移动
   */
  const handleMoveUp = async (assertion: RouteAssertion) => {
    if (isUnmounted.value) return

    const sorted = sortedAssertions.value
    const currentIndex = sorted.findIndex((a) => a.routeAssertionId === assertion.routeAssertionId)
    if (currentIndex <= 0) return

    const targetAssertion = sorted[currentIndex - 1]
    await swapAssertionOrder(assertion, targetAssertion)
  }

  /**
   * 处理向下移动
   */
  const handleMoveDown = async (assertion: RouteAssertion) => {
    if (isUnmounted.value) return

    const sorted = sortedAssertions.value
    const currentIndex = sorted.findIndex((a) => a.routeAssertionId === assertion.routeAssertionId)
    if (currentIndex >= sorted.length - 1) return

    const targetAssertion = sorted[currentIndex + 1]
    await swapAssertionOrder(assertion, targetAssertion)
  }

  /**
   * 交换断言顺序
   */
  const swapAssertionOrder = async (assertion1: RouteAssertion, assertion2: RouteAssertion) => {
    if (isUnmounted.value) return

    try {
      loading.value = true

      const tempOrder = assertion1.assertionOrder
      const updateAssertion1 = { ...assertion1, assertionOrder: assertion2.assertionOrder }
      const updateAssertion2 = { ...assertion2, assertionOrder: tempOrder }

      // 并行更新两个断言的顺序
      const [response1, response2] = await Promise.all([
        editRouteAssertion({
          routeAssertionId: updateAssertion1.routeAssertionId,
          routeConfigId: updateAssertion1.routeConfigId,
          assertionName: updateAssertion1.assertionName,
          assertionType: updateAssertion1.assertionType,
          assertionOperator: updateAssertion1.assertionOperator,
          fieldName: updateAssertion1.fieldName || '',
          expectedValue: updateAssertion1.expectedValue || '',
          patternValue: updateAssertion1.patternValue || '',
          caseSensitive: updateAssertion1.caseSensitive,
          assertionOrder: updateAssertion1.assertionOrder,
          isRequired: updateAssertion1.isRequired,
          assertionDesc: updateAssertion1.assertionDesc || '',
          activeFlag: updateAssertion1.activeFlag,
          noteText: updateAssertion1.noteText || '',
        }),
        editRouteAssertion({
          routeAssertionId: updateAssertion2.routeAssertionId,
          routeConfigId: updateAssertion2.routeConfigId,
          assertionName: updateAssertion2.assertionName,
          assertionType: updateAssertion2.assertionType,
          assertionOperator: updateAssertion2.assertionOperator,
          fieldName: updateAssertion2.fieldName || '',
          expectedValue: updateAssertion2.expectedValue || '',
          patternValue: updateAssertion2.patternValue || '',
          caseSensitive: updateAssertion2.caseSensitive,
          assertionOrder: updateAssertion2.assertionOrder,
          isRequired: updateAssertion2.isRequired,
          assertionDesc: updateAssertion2.assertionDesc || '',
          activeFlag: updateAssertion2.activeFlag,
          noteText: updateAssertion2.noteText || '',
        }),
      ])

      if (isUnmounted.value) return

      if (response1?.oK && response2?.oK) {
        safeMessage.success('调整执行顺序成功')
        await loadRouteAssertions()
      } else {
        safeMessage.error('调整执行顺序失败')
      }
    } catch (error) {
      if (!isUnmounted.value) {
        console.error('调整执行顺序失败:', error)
        safeMessage.error('调整执行顺序失败')
      }
    } finally {
      if (!isUnmounted.value) {
        loading.value = false
      }
    }
  }

  /**
   * 处理保存断言
   */
  const handleSaveAssertion = async (assertionData: RouteAssertionForm) => {
    if (isUnmounted.value) return

    try {
      const saveData = {
        routeConfigId: options.routeConfigId,
        assertionName: assertionData.assertionName,
        assertionType: assertionData.assertionType,
        assertionOperator: assertionData.assertionOperator,
        fieldName: assertionData.fieldName || '',
        expectedValue: assertionData.expectedValue || '',
        patternValue: assertionData.patternValue || '',
        caseSensitive: assertionData.caseSensitive,
        assertionOrder: assertionData.assertionOrder,
        isRequired: assertionData.isRequired,
        assertionDesc: assertionData.assertionDesc || '',
        activeFlag: assertionData.activeFlag,
        addWho: 'admin',
        editWho: 'admin',
        oprSeqFlag: Date.now().toString(),
        currentVersion: 1,
        noteText: assertionData.noteText || '',
      }

      let response
      const isEdit = !!currentAssertion.value?.routeAssertionId

      if (isEdit) {
        // 编辑模式
        response = await editRouteAssertion({
          ...saveData,
          routeAssertionId: currentAssertion.value!.routeAssertionId,
        })
      } else {
        // 新建模式
        response = await addRouteAssertion(saveData)
      }

      if (isUnmounted.value) return

      if (response?.oK) {
        const actionText = isEdit ? '编辑' : '创建'
        safeMessage.success(`${actionText}断言成功`)
        dialogVisible.value = false

        if (isEdit) {
          // 编辑模式：重新加载列表以确保数据一致性
          await loadRouteAssertions()
        } else {
          // 新建模式：直接将后端返回的数据添加到列表中
          if (response.bizData) {
            try {
              const returnedData = JSON.parse(response.bizData)

              // 构建完整的断言对象
              const newAssertion: RouteAssertion = {
                tenantId: returnedData.tenantId || 'default',
                routeAssertionId: returnedData.routeAssertionId,
                routeConfigId: returnedData.routeConfigId,
                assertionName: returnedData.assertionName || assertionData.assertionName,
                assertionType: assertionData.assertionType,
                assertionOperator: assertionData.assertionOperator,
                fieldName: assertionData.fieldName || '',
                expectedValue: assertionData.expectedValue || '',
                patternValue: assertionData.patternValue || '',
                caseSensitive: assertionData.caseSensitive ?? 'N',
                assertionOrder: assertionData.assertionOrder ?? 0,
                isRequired: assertionData.isRequired ?? 'N',
                assertionDesc: assertionData.assertionDesc || '',
                reserved1: returnedData.reserved1 || '',
                reserved2: returnedData.reserved2 || '',
                extProperty: returnedData.extProperty || '',
                activeFlag: assertionData.activeFlag ?? 'Y',
                addTime: new Date().toISOString(),
                addWho: 'admin',
                editTime: new Date().toISOString(),
                editWho: 'admin',
                oprSeqFlag: saveData.oprSeqFlag,
                currentVersion: 1,
                noteText: assertionData.noteText || '',
              }

              // 将新断言添加到列表中
              assertions.value.push(newAssertion)

              // 按执行顺序重新排序
              assertions.value.sort((a, b) => a.assertionOrder - b.assertionOrder)

              console.log('新断言已添加到列表:', newAssertion)
            } catch (error) {
              console.error('解析返回数据失败，回退到重新加载:', error)
              // 如果解析失败，回退到重新加载列表
              await loadRouteAssertions()
            }
          } else {
            // 如果没有返回数据，回退到重新加载列表
            console.warn('后端未返回完整数据，回退到重新加载')
            await loadRouteAssertions()
          }
        }
      } else {
        safeMessage.error(response?.errMsg || `${isEdit ? '编辑' : '创建'}断言失败`)
      }
    } catch (error) {
      if (!isUnmounted.value) {
        console.error('保存断言失败:', error)
        safeMessage.error('保存断言失败')
      }
    }
  }

  /**
   * 处理刷新
   */
  const handleRefresh = async () => {
    if (isUnmounted.value) return

    await loadRouteAssertions()
    safeMessage.success('刷新完成')
  }

  /**
   * 暴露给父组件的刷新方法
   */
  const refresh = async () => {
    if (isUnmounted.value) return
    await loadRouteAssertions()
  }

  // 监听路由配置ID变化
  watch(
    () => options.routeConfigId,
    async (newId) => {
      if (newId && !isUnmounted.value) {
        await loadRouteAssertions()
      }
    },
    { immediate: true },
  )

  // 初始化
  onMounted(() => {
    if (options.routeConfigId) {
      console.log('路由断言管理器初始化，路由ID:', options.routeConfigId)
    }
  })

  // 组件卸载时清理
  onUnmounted(() => {
    isUnmounted.value = true
  })

  return {
    // 响应式数据
    loading,
    dialogVisible,
    currentAssertion,
    assertions,
    sortedAssertions,

    // 计算属性和方法
    getAssertionTypeColor,
    getAssertionTypeLabel,
    getOperatorLabel,
    isFirstAssertion,
    isLastAssertion,

    // 操作方法
    loadRouteAssertions,
    handleCreate,
    handleEdit,
    handleDelete,
    handleToggleStatus,
    handleMoveUp,
    handleMoveDown,
    handleSaveAssertion,
    handleRefresh,
    refresh,
  }
}
