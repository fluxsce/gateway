/**
 * 断言配置列表 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField, DataFormTab } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { AddOutline, TrashOutline } from '@vicons/ionicons5'
import { ref } from 'vue'
import type { AssertConfig } from './types'
import { ASSERTION_OPERATOR_OPTIONS, ASSERTION_TYPE_OPTIONS } from './types'

/**
 * 断言配置列表 Model
 */
export function useAssertConfigModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0021-assert-config'
  
  /** 加载状态 */
  const loading = ref(false)

  /** 断言配置列表数据 */
  const assertList = ref<AssertConfig[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'assertionName',
        label: '断言名称',
        type: 'input',
        placeholder: '请输入断言名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'assertionType',
        label: '断言类型',
        type: 'select',
        placeholder: '请选择断言类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...ASSERTION_TYPE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
        ],
      },
      {
        field: 'activeFlag',
        label: '状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        options: [
          { label: '全部', value: '' },
          { label: '启用', value: 'Y' },
          { label: '禁用', value: 'N' },
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增断言',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新增断言配置',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '批量删除选中的断言配置',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 获取断言类型显示标签 */
  const getAssertionTypeLabel = (assertionType: string) => {
    const option = ASSERTION_TYPE_OPTIONS.find(opt => opt.value === assertionType)
    return option?.label || assertionType
  }

  /** 获取断言类型标签颜色 */
  const getAssertionTypeTagType = (assertionType: string): "default" | "success" | "error" | "warning" | "primary" | "info" => {
    const typeColorMap: Record<string, "default" | "success" | "error" | "warning" | "primary" | "info"> = {
      'PATH': 'primary',
      'HEADER': 'info',
      'QUERY': 'success',
      'COOKIE': 'warning',
      'IP': 'error',
    }
    return typeColorMap[assertionType] || 'default'
  }

  /** 获取操作符标签 */
  const getOperatorLabel = (operator: string) => {
    const option = ASSERTION_OPERATOR_OPTIONS.find(opt => opt.value === operator)
    return option?.label || operator
  }

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'routeAssertionId',
        title: '断言ID',
        visible: false,
        width: 0,
      },
      {
        field: 'assertionOrder',
        title: '执行顺序',
        align: 'center',
        width: 120,
        slots: { default: 'assertionOrder' },
      },
      {
        field: 'assertionName',
        title: '断言名称',
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
      },
      {
        field: 'assertionType',
        title: '断言类型',
        align: 'center',
        width: 120,
        slots: { default: 'assertionType' },
      },
      {
        field: 'assertionOperator',
        title: '操作符',
        align: 'center',
        width: 120,
        slots: { default: 'assertionOperator' },
      },
      {
        field: 'fieldName',
        title: '字段名称',
        align: 'center',
        showOverflow: 'tooltip',
        width: 150,
      },
      {
        field: 'expectedValue',
        title: '期望值/模式',
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
        formatter: ({ row }) => {
          return row.expectedValue || row.patternValue || '-'
        },
      },
      {
        field: 'isRequired',
        title: '必须匹配',
        align: 'center',
        width: 100,
        slots: { default: 'isRequired' },
      },
      {
        field: 'activeFlag',
        title: '状态',
        align: 'center',
        width: 100,
        slots: { default: 'activeFlag' },
      },
      {
        field: 'assertionDesc',
        title: '描述',
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
      },
      {
        field: 'addTime',
        title: '创建时间',
        align: 'center',
        formatter: ({ row }) => formatDate(row.addTime),
        width: 180,
      },
      {
        field: 'addWho',
        title: '创建人',
        align: 'center',
        showOverflow: 'tooltip',
        width: 120,
      },
      {
        field: 'editTime',
        title: '修改时间',
        align: 'center',
        formatter: ({ row }) => formatDate(row.editTime),
        width: 180,
      },
      {
        field: 'editWho',
        title: '修改人',
        align: 'center',
        showOverflow: 'tooltip',
        width: 120,
      },
    ],
    showCheckbox: true,
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      showCopyRow: true,
      customMenus: [
        {
          code: 'view',
          name: '查看详情',
          prefixIcon: 'vxe-icon-eye-fill',
        },
        {
          code: 'edit',
          name: '编辑',
          prefixIcon: 'vxe-icon-edit',
        },
        {
          code: 'toggle-status',
          name: '切换状态',
          prefixIcon: 'vxe-icon-check',
        },
        {
          code: 'delete',
          name: '删除',
          prefixIcon: 'vxe-icon-delete',
        },
      ],
    },
  }

  // ============= 状态更新方法 =============

  /**
   * 设置断言列表
   */
  function setAssertList(list: AssertConfig[]) {
    assertList.value = list
  }

  /**
   * 设置加载状态
   */
  function setLoading(value: boolean) {
    loading.value = value
  }

  /**
   * 重置分页信息
   */
  function resetPagination() {
    pageInfo.value = undefined
  }

  /**
   * 更新分页信息（接收后端 PageInfoObj）
   */
  function updatePagination(newPageInfo: Partial<PageInfoObj>) {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  /**
   * 添加断言到列表
   */
  function addAssertToList(assert: AssertConfig) {
    assertList.value.push(assert)
    // 按 assertionOrder 排序
    assertList.value.sort((a, b) => (a.assertionOrder || 0) - (b.assertionOrder || 0))
  }

  /**
   * 更新列表中的断言
   */
  function updateAssertInList(routeAssertionId: string, tenantId: string | undefined, updatedAssert: Partial<AssertConfig>) {
    const index = assertList.value.findIndex(
      (a) => a.routeAssertionId === routeAssertionId && (!tenantId || a.tenantId === tenantId)
    )
    if (index !== -1) {
      Object.assign(assertList.value[index], updatedAssert)
      // 按 assertionOrder 排序
      assertList.value.sort((a, b) => (a.assertionOrder || 0) - (b.assertionOrder || 0))
    }
  }

  /**
   * 从列表中移除断言
   */
  function removeAssertFromList(routeAssertionId: string) {
    const index = assertList.value.findIndex((a) => a.routeAssertionId === routeAssertionId)
    if (index >= 0) {
      assertList.value.splice(index, 1)
    }
  }

  /**
   * 从列表中批量移除断言
   */
  function removeAssertsFromList(routeAssertionIds: string[]) {
    assertList.value = assertList.value.filter((a) => !routeAssertionIds.includes(a.routeAssertionId))
  }

  // ============= 表单配置 =============

  /** 表单页签配置 */
  const formTabs = [
    {
      key: 'basic',
      label: '基本信息',
    },
    {
      key: 'other',
      label: '其他信息',
    },
  ] as DataFormTab[]

  /** 断言表单配置（用于 GdataFormModal） */
  const formFields: DataFormField[] = [
    // 主键字段（隐藏，但必须存在用于编辑）
    {
      field: 'routeAssertionId',
      label: '断言ID',
      type: 'input' as const,
      span: 12,
      primary: true,
      show: false,
    },
    {
      field: 'routeConfigId',
      label: '路由配置ID',
      type: 'input' as const,
      span: 12,
      show: false,
    },
    // 基本信息
    {
      field: 'assertionName',
      label: '断言名称',
      type: 'input' as const,
      placeholder: '请输入断言名称',
      span: 12,
      tabKey: 'basic',
      required: true,
      rules: [
        { required: true, message: '请输入断言名称', trigger: ['blur', 'input'] },
        { min: 2, max: 100, message: '断言名称长度应在2-100字符之间', trigger: ['blur', 'input'] },
      ],
    },
    {
      field: 'assertionType',
      label: '断言类型',
      type: 'select' as const,
      placeholder: '请选择断言类型',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 'HEADER',
      options: ASSERTION_TYPE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
      rules: [
        { required: true, message: '请选择断言类型', trigger: ['blur', 'change'] },
      ],
    },
    {
      field: 'assertionOperator',
      label: '操作符',
      type: 'select' as const,
      placeholder: '请选择操作符',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 'EQUAL',
      options: ASSERTION_OPERATOR_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
      rules: [
        { required: true, message: '请选择操作符', trigger: ['blur', 'change'] },
      ],
    },
    {
      field: 'assertionOrder',
      label: '执行顺序',
      type: 'number' as const,
      placeholder: '数值越小优先级越高',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 100,
      props: {
        min: 0,
        max: 9999,
        precision: 0,
      },
      rules: [
        { 
          required: true, 
          type: 'number',
          message: '请输入执行顺序', 
          trigger: ['blur', 'change'],
        },
      ],
    },
    {
      field: 'isRequired',
      label: '是否必须匹配',
      type: 'select' as const,
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 'Y',
      options: [
        { label: '必须匹配', value: 'Y' },
        { label: '可选匹配', value: 'N' },
      ],
      rules: [
        { required: true, message: '请选择是否必须匹配', trigger: ['change'] },
      ],
    },
    {
      field: 'activeFlag',
      label: '启用状态',
      type: 'switch' as const,
      span: 12,
      tabKey: 'basic',
      defaultValue: 'Y',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
    // ============= 断言配置（使用 fieldset 包裹） =============
    {
      field: 'assertionConfigFieldset',
      label: '断言配置',
      type: 'fieldset' as const,
      span: 24,
      tabKey: 'basic',
      children: [
        {
          field: 'fieldName',
          label: '字段名称',
          type: 'input' as const,
          placeholder: '请输入字段名称（HEADER/QUERY/COOKIE类型必填）',
          span: 12,
          show: (formData: Record<string, any>) => {
            return ['HEADER', 'QUERY', 'COOKIE'].includes(formData.assertionType)
          },
          rules: [
            {
              validator: (_rule: any, value: any, formData: Record<string, any>) => {
                const needsField = ['HEADER', 'QUERY', 'COOKIE'].includes(formData.assertionType)
                if (needsField && (!value || !value.trim())) {
                  return new Error('请输入字段名称')
                }
                return true
              },
              trigger: ['blur', 'input'],
            },
          ],
        },
        {
          field: 'expectedValue',
          label: '期望值',
          type: 'input' as const,
          placeholder: '请输入期望值',
          span: 24,
          show: (formData: Record<string, any>) => {
            const valueOperators = [
              'EQUAL',
              'NOT_EQUAL',
              'CONTAINS',
              'NOT_CONTAINS',
              'STARTS_WITH',
              'ENDS_WITH',
              'IN',
              'NOT_IN',
            ]
            return valueOperators.includes(formData.assertionOperator)
          },
          props: {
            type: 'textarea',
            rows: 3,
            maxlength: 500,
            showCount: true,
          },
        },
        {
          field: 'patternValue',
          label: '匹配模式',
          type: 'input' as const,
          placeholder: '请输入正则表达式或匹配模式',
          span: 24,
          show: (formData: Record<string, any>) => {
            return ['MATCHES', 'NOT_MATCHES'].includes(formData.assertionOperator)
          },
          props: {
            type: 'textarea',
            rows: 3,
            maxlength: 500,
            showCount: true,
          },
        },
        {
          field: 'caseSensitive',
          label: '区分大小写',
          type: 'select' as const,
          span: 12,
          defaultValue: 'Y',
          show: (formData: Record<string, any>) => {
            return formData.assertionType !== 'IP'
          },
          options: [
            { label: '区分大小写', value: 'Y' },
            { label: '不区分大小写', value: 'N' },
          ],
        },
        {
          field: 'assertionDesc',
          label: '断言描述',
          type: 'input' as const,
          placeholder: '请输入断言描述（可选）',
          span: 24,
          props: {
            type: 'textarea',
            rows: 3,
            maxlength: 200,
            showCount: true,
          },
        },
      ],
    },
    // 其他信息
    {
      field: 'noteText',
      label: '备注信息',
      type: 'input' as const,
      placeholder: '请输入备注信息',
      span: 24,
      tabKey: 'other',
      props: {
        type: 'textarea',
        rows: 3,
        maxlength: 500,
        showCount: true,
      },
    },
    {
      field: 'addTime',
      label: '创建时间',
      type: 'datetime' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'addWho',
      label: '创建人',
      type: 'input' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'editTime',
      label: '修改时间',
      type: 'datetime' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'editWho',
      label: '修改人',
      type: 'input' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
  ]

  return {
    // 状态
    moduleId,
    loading,
    assertList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,
    formFields,
    formTabs,

    // 工具函数
    getAssertionTypeLabel,
    getAssertionTypeTagType,
    getOperatorLabel,

    // 方法
    setAssertList,
    setLoading,
    resetPagination,
    updatePagination,
    addAssertToList,
    updateAssertInList,
    removeAssertFromList,
    removeAssertsFromList,
  }
}

/**
 * 断言配置列表 Model 类型
 */
export type AssertConfigModel = ReturnType<typeof useAssertConfigModel>

