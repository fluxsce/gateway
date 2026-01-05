/**
 * 角色管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import {
  AddOutline,
  CreateOutline,
  TrashOutline
} from '@vicons/ionicons5'
import { ref } from 'vue'
import type { Role } from '../types/index'
import { BuiltInFlag, FlagEnum, RoleStatus } from '../types/index'

/**
 * 角色管理 Model
 */
export function useRoleModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0005'
  /** 加载状态 */
  const loading = ref(false)

  /** 角色列表数据 */
  const roleList = ref<Role[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'roleName',
        label: '角色名称',
        type: 'input',
        placeholder: '请输入角色名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'roleStatus',
        label: '状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '启用', value: RoleStatus.ENABLED },
          { label: '禁用', value: RoleStatus.DISABLED },
        ],
      },
      {
        field: 'builtInFlag',
        label: '类型',
        type: 'select',
        placeholder: '请选择类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '内置', value: BuiltInFlag.BUILT_IN },
          { label: '自定义', value: BuiltInFlag.CUSTOM },
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新增角色',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: CreateOutline,
        tooltip: '编辑选中的角色',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '删除选中的角色',
      }
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 数据编辑表单字段配置（供 GdataFormModal 使用） =============
  const formTabs = [
    { key: 'basic', label: '基本信息' },
    { key: 'other', label: '其他信息' },
  ]

  const formFields: DataFormField[] = [
    {
      field: 'roleId',
      label: '角色ID',
      type: 'input',
      placeholder: '请输入角色ID',
      span: 8,
      tabKey: 'basic',
      required: true,
      primary: true,
    },
    {
      field: 'roleName',
      label: '角色名称',
      type: 'input',
      placeholder: '请输入角色名称',
      span: 8,
      tabKey: 'basic',
      required: true,
    },
    {
      field: 'roleDescription',
      label: '角色描述',
      type: 'textarea',
      placeholder: '请输入角色描述',
      span: 24,
      tabKey: 'basic',
    },
    {
      field: 'roleStatus',
      label: '角色状态',
      type: 'select',
      placeholder: '请选择状态',
      span: 8,
      tabKey: 'basic',
      defaultValue: RoleStatus.ENABLED,
      options: [
        { label: '启用', value: RoleStatus.ENABLED },
        { label: '禁用', value: RoleStatus.DISABLED },
      ],
    },
    {
      field: 'builtInFlag',
      label: '内置标记',
      type: 'select',
      placeholder: '请选择类型',
      span: 8,
      tabKey: 'basic',
      defaultValue: BuiltInFlag.CUSTOM,
      options: [
        { label: '内置', value: BuiltInFlag.BUILT_IN },
        { label: '自定义', value: BuiltInFlag.CUSTOM },
      ],
    },
    {
      field: 'dataScope',
      label: '数据权限范围',
      type: 'textarea',
      placeholder: '请输入数据权限范围配置（JSON格式）',
      span: 24,
      tabKey: 'basic',
    },
    {
      field: 'addTime',
      label: '创建时间',
      type: 'datetime',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'addWho',
      label: '创建人',
      type: 'input',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'editTime',
      label: '修改时间',
      type: 'datetime',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'editWho',
      label: '修改人',
      type: 'input',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'oprSeqFlag',
      label: '操作序列标识',
      type: 'input',
      span: 8,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'currentVersion',
      label: '当前版本号',
      type: 'number',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'activeFlag',
      label: '活动标记',
      type: 'select',
      span: 8,
      tabKey: 'basic',
      defaultValue: FlagEnum.YES,
      options: [
        { label: '活动', value: FlagEnum.YES },
        { label: '非活动', value: FlagEnum.NO },
      ],
    },
    {
      field: 'noteText',
      label: '备注',
      type: 'textarea',
      placeholder: '请输入备注信息',
      span: 24,
      tabKey: 'basic',
    },
  ]

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'roleId',
        title: '角色ID',
        showOverflow: true,
      },
      {
        field: 'roleName',
        title: '角色名称',
        sortable: true,
        showOverflow: true,
      },
      {
        field: 'roleDescription',
        title: '角色描述',
        showOverflow: 'tooltip',
      },
      {
        field: 'roleStatus',
        title: '状态',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue === RoleStatus.ENABLED ? '启用' : '禁用'
        },
      },
      {
        field: 'builtInFlag',
        title: '类型',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue === BuiltInFlag.BUILT_IN ? '内置' : '自定义'
        },
      },
      {
        field: 'activeFlag',
        title: '活动标记',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue === FlagEnum.YES ? '活动' : '非活动'
        },
      },
      {
        field: 'addTime',
        title: '创建时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'addWho',
        title: '创建人',
        showOverflow: true,
      },
      {
        field: 'editTime',
        title: '修改时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'editWho',
        title: '修改人',
        showOverflow: true,
      },
      {
        field: 'noteText',
        title: '备注',
        showOverflow: 'tooltip',
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
      showCopyCell: true,
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
          code: 'delete',
          name: '删除',
          prefixIcon: 'vxe-icon-delete',
        },
        {
          code: 'roleAuth',
          name: '角色授权',
          prefixIcon: 'vxe-icon-user',
        },
      ],
    },
    height: '100%',
  }

  // ============= 辅助方法 =============

  /**
   * 重置分页
   */
  const resetPagination = () => {
    pageInfo.value = undefined
  }

  /**
   * 更新分页信息（接收后端 PageInfoObj）
   */
  const updatePagination = (newPageInfo: Partial<PageInfoObj>) => {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  /**
   * 设置角色列表
   */
  const setRoleList = (list: Role[]) => {
    roleList.value = list
  }

  /**
   * 清空角色列表
   */
  const clearRoleList = () => {
    roleList.value = []
  }

  /**
   * 添加角色到列表
   */
  const addRoleToList = (role: Role) => {
    roleList.value.unshift(role)
  }

  /**
   * 更新列表中的角色
   */
  const updateRoleInList = (roleId: string, tenantId: string, updatedRole: Partial<Role>) => {
    const index = roleList.value.findIndex((r) => r.roleId === roleId && r.tenantId === tenantId)
    if (index !== -1) {
      Object.assign(roleList.value[index], updatedRole)
    }
  }

  /**
   * 从列表中删除角色
   */
  const removeRoleFromList = (roleId: string, tenantId: string) => {
    const index = roleList.value.findIndex((r) => r.roleId === roleId && r.tenantId === tenantId)
    if (index !== -1) {
      roleList.value.splice(index, 1)
    }
  }

  /**
   * 批量删除角色
   */
  const removeRolesFromList = (roles: Role[]) => {
    roles.forEach((role) => {
      removeRoleFromList(role.roleId, role.tenantId)
    })
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    roleList,
    pageInfo,

    // 配置
    searchFormConfig,
    formTabs,
    formFields,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setRoleList,
    clearRoleList,
    addRoleToList,
    updateRoleInList,
    removeRoleFromList,
    removeRolesFromList,
  }
}

/**
 * Model 返回类型
 */
export type RoleModel = ReturnType<typeof useRoleModel>

