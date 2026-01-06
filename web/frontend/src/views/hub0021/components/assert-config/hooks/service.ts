/**
 * 断言配置列表服务层 Hook
 * 纯业务逻辑：数据获取、增删改查等操作
 */

import { createBackendPaginationParams } from '@/components/gpage'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import {
  addRouteAssertion,
  deleteRouteAssertion,
  editRouteAssertion,
  queryRouteAssertions
} from '../../../api'
import { useAssertConfigModel } from './model'
import type { AssertConfig } from './types'

/**
 * 断言配置列表服务层 Hook
 * @param routeConfigId 路由配置ID（必填）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useAssertConfigService(
  routeConfigId: Ref<string | undefined> | string,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()

  // 获取 routeConfigId 的实际值
  const getRouteConfigId = () => {
    return typeof routeConfigId === 'string' ? routeConfigId : routeConfigId?.value
  }

  // 使用 model
  const model = useAssertConfigModel()
  
  // 解构 model 中的列表操作方法
  const {
    pageInfo,
    addAssertToList,
    updateAssertInList,
    removeAssertFromList,
    removeAssertsFromList,
    updatePagination,
    resetPagination,
  } = model

  /**
   * 加载断言列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadAssertList = async (searchParams?: Record<string, any>) => {
    try {
      model.setLoading(true)

      const routeId = getRouteConfigId()
      if (!routeId) {
        model.setAssertList([])
        model.resetPagination()
        return
      }

      // 如果没有传入查询参数，从搜索表单获取
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      // 过滤掉空字符串、null 和 undefined 的查询条件
      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(
            Object.entries(finalSearchParams).filter(
              ([, value]) => value !== '' && value !== null && value !== undefined
            )
          )
        : {}

      // 构建请求参数：合并查询条件和分页参数
      const params = {
        // 路由配置ID（必填）
        routeConfigId: routeId,
        // 查询条件
        ...effectiveSearchParams,
        // 分页参数
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        ),
      }

      // 调用 API 分页查询断言列表
      const response = await queryRouteAssertions(params)

      if (isApiSuccess(response)) {
        // 解析业务数据
        const asserts = parseJsonData<AssertConfig[]>(response, []) || []
        
        // 按执行顺序排序
        if (Array.isArray(asserts) && asserts.length > 0) {
          asserts.sort((a, b) => (a.assertionOrder || 0) - (b.assertionOrder || 0))
        }

        model.setAssertList(asserts)

        // 解析分页信息
        if (response.pageQueryData) {
          try {
            const backendPageInfo = JSON.parse(response.pageQueryData)
            updatePagination(backendPageInfo)
          } catch (error) {
            console.error('解析分页信息失败:', error)
            model.resetPagination()
          }
        } else {
          model.resetPagination()
        }
      } else {
        message.error(getApiMessage(response, '加载断言列表失败'))
        model.setAssertList([])
        model.resetPagination()
      }
    } catch (error) {
      console.error('加载断言列表失败:', error)
      message.error('加载断言列表失败')
      model.setAssertList([])
      model.resetPagination()
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 添加断言
   * @param assertData 断言数据
   */
  const addAssert = async (assertData: any) => {
    try {
      model.setLoading(true)

      const routeId = getRouteConfigId()
      if (!routeId) {
        message.error('路由配置ID不能为空')
        return false
      }

      const saveData = {
        routeConfigId: routeId,
        assertionName: assertData.assertionName,
        assertionType: assertData.assertionType,
        assertionOperator: assertData.assertionOperator,
        fieldName: assertData.fieldName || '',
        expectedValue: assertData.expectedValue || '',
        // patternValue 仅用于路径断言（PATH），对应后端的 Pattern 字段
        // 用于选择路径匹配模式：exact（精确匹配）、prefix（前缀匹配）、regex（正则匹配）、param（参数匹配）
        patternValue: assertData.assertionType === 'PATH' ? (assertData.patternValue || '') : '',
        caseSensitive: assertData.caseSensitive || 'Y',
        assertionOrder: assertData.assertionOrder || 100,
        isRequired: assertData.isRequired || 'Y',
        assertionDesc: assertData.assertionDesc || '',
        activeFlag: assertData.activeFlag || 'Y',
        noteText: assertData.noteText || '',
      }

      const response = await addRouteAssertion(saveData)

      if (isApiSuccess(response)) {
        message.success('添加断言成功')
        
        // 如果返回了完整数据，直接添加到列表
        if (response.bizData) {
          try {
            const newAssert = parseJsonData<AssertConfig>(response)
            if (newAssert) {
              addAssertToList(newAssert)
              return true
            }
          } catch (error) {
            console.error('解析返回数据失败:', error)
          }
        }
        
        // 否则重新加载列表
        await loadAssertList()
        return true
      } else {
        message.error(getApiMessage(response, '添加断言失败'))
        return false
      }
    } catch (error) {
      console.error('添加断言失败:', error)
      message.error('添加断言失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 编辑断言
   * @param assertData 断言数据
   */
  const editAssert = async (assertData: any) => {
    try {
      model.setLoading(true)

      const routeId = getRouteConfigId()
      if (!routeId) {
        message.error('路由配置ID不能为空')
        return false
      }

      if (!assertData.routeAssertionId) {
        message.error('断言ID不能为空')
        return false
      }

      const updateData = {
        routeAssertionId: assertData.routeAssertionId,
        routeConfigId: routeId,
        assertionName: assertData.assertionName,
        assertionType: assertData.assertionType,
        assertionOperator: assertData.assertionOperator,
        fieldName: assertData.fieldName || '',
        expectedValue: assertData.expectedValue || '',
        // patternValue 仅用于路径断言（PATH），对应后端的 Pattern 字段
        // 用于选择路径匹配模式：exact（精确匹配）、prefix（前缀匹配）、regex（正则匹配）、param（参数匹配）
        patternValue: assertData.assertionType === 'PATH' ? (assertData.patternValue || '') : '',
        caseSensitive: assertData.caseSensitive || 'Y',
        assertionOrder: assertData.assertionOrder || 100,
        isRequired: assertData.isRequired || 'Y',
        assertionDesc: assertData.assertionDesc || '',
        activeFlag: assertData.activeFlag || 'Y',
        noteText: assertData.noteText || '',
      }

      const response = await editRouteAssertion(updateData)

      if (isApiSuccess(response)) {
        message.success('编辑断言成功')
        
        // 如果返回了完整数据，直接更新列表
        if (response.bizData) {
          try {
            const updatedAssert = parseJsonData<AssertConfig>(response)
            if (updatedAssert) {
              updateAssertInList(updatedAssert.routeAssertionId, updatedAssert.tenantId, updatedAssert)
              return true
            }
          } catch (error) {
            console.error('解析返回数据失败:', error)
          }
        }
        
        // 否则重新加载列表
        await loadAssertList()
        return true
      } else {
        message.error(getApiMessage(response, '编辑断言失败'))
        return false
      }
    } catch (error) {
      console.error('编辑断言失败:', error)
      message.error('编辑断言失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 删除断言
   * @param routeAssertionId 断言ID
   */
  const deleteAssert = async (routeAssertionId: string) => {
    try {
      model.setLoading(true)

      const response = await deleteRouteAssertion(routeAssertionId)

      if (isApiSuccess(response)) {
        message.success('删除断言成功')
        removeAssertFromList(routeAssertionId)
        return true
      } else {
        message.error(getApiMessage(response, '删除断言失败'))
        return false
      }
    } catch (error) {
      console.error('删除断言失败:', error)
      message.error('删除断言失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 批量删除断言
   * @param routeAssertionIds 断言ID数组
   */
  const deleteAsserts = async (routeAssertionIds: string[]) => {
    if (!routeAssertionIds || routeAssertionIds.length === 0) {
      message.warning('请选择要删除的断言')
      return false
    }

    try {
      model.setLoading(true)

      // 并行删除
      const results = await Promise.all(
        routeAssertionIds.map(id => deleteRouteAssertion(id))
      )

      const successCount = results.filter(r => isApiSuccess(r)).length
      const failCount = results.length - successCount

      if (failCount === 0) {
        message.success(`成功删除 ${successCount} 个断言`)
        removeAssertsFromList(routeAssertionIds)
        return true
      } else if (successCount > 0) {
        message.warning(`部分删除成功：成功 ${successCount} 个，失败 ${failCount} 个`)
        // 只移除成功删除的
        const successIds = routeAssertionIds.filter((id, index) => isApiSuccess(results[index]))
        removeAssertsFromList(successIds)
        return true
      } else {
        message.error('批量删除失败')
        return false
      }
    } catch (error) {
      console.error('批量删除断言失败:', error)
      message.error('批量删除断言失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 切换断言状态
   * @param assert 断言对象
   */
  const toggleAssertStatus = async (assert: AssertConfig) => {
    try {
      const newStatus = assert.activeFlag === 'Y' ? 'N' : 'Y'
      const updateData = {
        ...assert,
        activeFlag: newStatus,
      }

      const result = await editAssert(updateData)
      return result
    } catch (error) {
      console.error('切换断言状态失败:', error)
      message.error('切换断言状态失败')
      return false
    }
  }

  /**
   * 搜索断言
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    // 搜索时重置到第一页
    model.resetPagination()
    await loadAssertList(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadAssertList()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    await loadAssertList()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadAssertList()
  }

  return {
    model,
    loadAssertList,
    addAssert,
    editAssert,
    deleteAssert,
    deleteAsserts,
    toggleAssertStatus,
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
  }
}

/**
 * 断言配置列表服务层 Hook 类型
 */
export type AssertConfigService = ReturnType<typeof useAssertConfigService>

