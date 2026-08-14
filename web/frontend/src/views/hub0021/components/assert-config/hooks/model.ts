/**
 * 断言配置列表 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsDataFormField, RsDataFormTab } from '@/components/form/rs-data'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { RsTag } from '@/ui'
import { h, ref } from 'vue'
import type { AssertConfig } from './types'
import { ASSERTION_OPERATOR_OPTIONS, ASSERTION_TYPE_OPTIONS } from './types'

/**
 * 断言配置表格配置（对齐 RsGrid Props 子集）。
 */
export interface AssertConfigGridConfig {
  columns: RsGridColumn<AssertConfig>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * 断言配置列表 Model
 */
export function useAssertConfigModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0021:assertConfig'

  /** 加载状态 */
  const loading = ref(false)

  /** 断言配置列表数据 */
  const assertList = ref<AssertConfig[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
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
          ...ASSERTION_TYPE_OPTIONS.map((opt) => ({ label: opt.label, value: opt.value })),
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
        icon: 'AddOutline',
        type: 'primary',
        tooltip: '新增断言配置',
      },
      {
        key: 'delete',
        label: '删除',
        icon: 'TrashOutline',
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
    const option = ASSERTION_TYPE_OPTIONS.find((opt) => opt.value === assertionType)
    return option?.label || assertionType
  }

  /** 获取断言类型标签颜色 */
  const getAssertionTypeTagType = (
    assertionType: string,
  ): 'default' | 'success' | 'danger' | 'warning' | 'primary' | 'info' => {
    const typeColorMap: Record<
      string,
      'default' | 'success' | 'danger' | 'warning' | 'primary' | 'info'
    > = {
      PATH: 'primary',
      HEADER: 'info',
      QUERY: 'success',
      COOKIE: 'warning',
      IP: 'danger',
      BODY_CONTENT: 'success',
    }
    return typeColorMap[assertionType] || 'default'
  }

  /** 获取操作符标签 */
  const getOperatorLabel = (operator: string) => {
    const option = ASSERTION_OPERATOR_OPTIONS.find((opt) => opt.value === operator)
    return option?.label || operator
  }

  /** 表格配置（符合 RsGrid 结构） */
  const gridConfig: AssertConfigGridConfig = {
    columns: [
      {
        key: 'routeAssertionId',
        title: '断言ID',
        visible: false,
      },
      {
        key: 'assertionOrder',
        title: '执行顺序',
        align: 'center',
        width: 120,
        render: (row) =>
          h(
            'span',
            { style: { fontWeight: 'bold', color: '#0066cc' } },
            String(row.assertionOrder ?? ''),
          ),
      },
      {
        key: 'assertionName',
        title: '断言名称',
        align: 'center',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'assertionType',
        title: '断言类型',
        align: 'center',
        width: 120,
        render: (row) =>
          h(
            RsTag,
            {
              variant: getAssertionTypeTagType(row.assertionType),
              size: 'sm',
            },
            () => getAssertionTypeLabel(row.assertionType),
          ),
      },
      {
        key: 'assertionOperator',
        title: '操作符',
        align: 'center',
        width: 120,
        render: (row) =>
          h(RsTag, { variant: 'info', size: 'sm' }, () =>
            getOperatorLabel(row.assertionOperator),
          ),
      },
      {
        key: 'fieldName',
        title: '字段名称',
        align: 'center',
        ellipsis: true,
        width: 150,
      },
      {
        key: 'expectedValue',
        title: '期望值/模式',
        align: 'center',
        ellipsis: true,
        width: 200,
        formatter: (_v, row) => row.expectedValue || '-',
      },
      {
        key: 'isRequired',
        title: '必须匹配',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.isRequired === 'Y' ? 'danger' : 'default',
              size: 'sm',
            },
            () => (row.isRequired === 'Y' ? '必须' : '可选'),
          ),
      },
      {
        key: 'activeFlag',
        title: '状态',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.activeFlag === 'Y' ? 'success' : 'danger',
              size: 'sm',
            },
            () => (row.activeFlag === 'Y' ? '启用' : '禁用'),
          ),
      },
      {
        key: 'assertionDesc',
        title: '描述',
        align: 'center',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'addTime',
        title: '创建时间',
        align: 'center',
        width: 180,
        formatter: (_v, row) => formatDate(row.addTime),
      },
      {
        key: 'addWho',
        title: '创建人',
        align: 'center',
        ellipsis: true,
        width: 120,
      },
      {
        key: 'editTime',
        title: '修改时间',
        align: 'center',
        width: 180,
        formatter: (_v, row) => formatDate(row.editTime),
      },
      {
        key: 'editWho',
        title: '修改人',
        align: 'center',
        ellipsis: true,
        width: 120,
      },
    ],
    selectable: true,
    rowKey: 'routeAssertionId',
    height: '100%',
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      items: [
        { key: 'view', label: '查看详情', icon: 'eye' },
        { key: 'edit', label: '编辑', icon: 'pencil' },
        { key: 'toggle-status', label: '切换状态', icon: 'circle-check' },
        { key: 'delete', label: '删除', icon: 'trash-2', danger: true },
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
    assertList.value.sort((a, b) => (a.assertionOrder || 0) - (b.assertionOrder || 0))
  }

  /**
   * 更新列表中的断言
   */
  function updateAssertInList(
    routeAssertionId: string,
    tenantId: string | undefined,
    updatedAssert: Partial<AssertConfig>,
  ) {
    const index = assertList.value.findIndex(
      (a) => a.routeAssertionId === routeAssertionId && (!tenantId || a.tenantId === tenantId),
    )
    if (index !== -1) {
      Object.assign(assertList.value[index], updatedAssert)
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
    assertList.value = assertList.value.filter(
      (a) => !routeAssertionIds.includes(a.routeAssertionId),
    )
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
  ] as RsDataFormTab[]

  /** 断言表单配置（用于 RsDataFormModal） */
  const formFields: RsDataFormField[] = [
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
    {
      field: 'assertionName',
      label: '断言名称',
      type: 'input' as const,
      placeholder: '请输入断言名称',
      span: 12,
      tabKey: 'basic',
      required: true,
      rules: [
        { required: true, message: '请输入断言名称', trigger: ['blur', 'change'] },
        { min: 2, max: 100, message: '断言名称长度应在2-100字符之间', trigger: ['blur', 'change'] },
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
      options: ASSERTION_TYPE_OPTIONS.map((opt) => ({ label: opt.label, value: opt.value })),
      rules: [{ required: true, message: '请选择断言类型', trigger: ['blur', 'change'] }],
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
      options: ASSERTION_OPERATOR_OPTIONS.map((opt) => ({ label: opt.label, value: opt.value })),
      rules: [{ required: true, message: '请选择操作符', trigger: ['blur', 'change'] }],
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
      rules: [{ required: true, message: '请选择是否必须匹配', trigger: ['change'] }],
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
          required: true,
          rules: [
            {
              // 字段仅在 HEADER/QUERY/COOKIE 时展示（v-if），隐藏时不会参与校验
              validator: (value: unknown) => {
                if (typeof value !== 'string' || !value.trim()) {
                  return '请输入字段名称'
                }
                return true
              },
              trigger: ['blur', 'change'],
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
            const noValueOperators = ['EXISTS', 'NOT_EXISTS']
            return !noValueOperators.includes(formData.assertionOperator)
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
          type: 'select' as const,
          placeholder: '请选择路径匹配模式（仅路径断言使用）',
          span: 24,
          show: (formData: Record<string, any>) => {
            return formData.assertionType === 'PATH'
          },
          options: [
            { label: '精确匹配（exact）', value: 'exact' },
            { label: '前缀匹配（prefix）', value: 'prefix' },
            { label: '正则匹配（regex）', value: 'regex' },
            { label: '参数匹配（param）', value: 'param' },
          ],
          defaultValue: 'exact',
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
