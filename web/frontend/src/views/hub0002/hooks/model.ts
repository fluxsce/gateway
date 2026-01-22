/**
 * 用户管理模块 Model
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
  LockClosedOutline,
  TrashOutline
} from '@vicons/ionicons5'
import { ref } from 'vue'
import type { User } from '../types/index'
import { FlagEnum, Gender, UserStatus } from '../types/index'

/**
 * 用户管理 Model
 */
export function useUserModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0002'
  /** 加载状态 */
  const loading = ref(false)

  /** 用户列表数据 */
  const userList = ref<User[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'userName',
        label: '用户名',
        type: 'input',
        placeholder: '请输入用户名',
        span: 6,
        clearable: true,
      },
      {
        field: 'realName',
        label: '姓名',
        type: 'input',
        placeholder: '请输入姓名',
        span: 6,
        clearable: true,
      },
      {
        field: 'statusFlag',
        label: '状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '启用', value: UserStatus.ENABLED },
          { label: '禁用', value: UserStatus.DISABLED },
        ],
      },
    ],
    moreFields: [
      {
        field: 'mobile',
        label: '手机号',
        type: 'input',
        placeholder: '请输入手机号',
        span: 6,
        clearable: true,
      },
      {
        field: 'email',
        label: '邮箱',
        type: 'input',
        placeholder: '请输入邮箱',
        span: 6,
        clearable: true,
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新增用户',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: CreateOutline,
        tooltip: '编辑选中的用户',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '删除选中的用户',
      },
      {
        key: 'resetPassword',
        label: '重置密码',
        icon: LockClosedOutline,
        tooltip: '重置选中用户的密码',
      }
    ],
    showSearchButton:true,
    showResetButton:true,
  }

  // ============= 数据编辑表单字段配置（供 GdataFormModal 使用） =============
  // 与 User 类型字段保持一致（租户ID tenantId 由系统处理，这里不做输入）
  // 表单页签配置（主信息 / 自定义 / 其他）
  const formTabs = [
    { key: 'basic', label: '主信息' },
    { key: 'custom', label: '自定义' },
    { key: 'other', label: '其他' },
  ]

  const formFields: DataFormField[] = [
    {
      field: 'userId',
      label: '用户ID',
      type: 'input',
      placeholder: '请输入用户ID',
      span: 8,
      tabKey: 'basic',
      required: true,
      primary: true,
    },
    {
      field: 'userName',
      label: '用户名',
      type: 'input',
      placeholder: '请输入用户名',
      span: 8,
      tabKey: 'basic',
      required: true,
    },
    {
      field: 'password',
      label: '密码',
      type: 'input',
      placeholder: '请输入密码',
      span: 8,
      tabKey: 'basic',
      // 仅在新增模式下显示密码字段，编辑模式下隐藏（有专门的密码修改功能）
      show: (formData: Record<string, any>) => !formData.userId,
      required: true, // 仅在新增时显示，所以始终是必填的
      props: {
        type: 'password',
        showPasswordOn: 'click',
      },
    },
    {
      field: 'realName',
      label: '姓名',
      type: 'input',
      placeholder: '请输入姓名',
      span: 8,
      tabKey: 'basic',
      required: true,
    },
    {
      field: 'deptId',
      label: '部门ID',
      type: 'input',
      placeholder: '请输入部门ID',
      span: 8,
      tabKey: 'basic',
    },
    {
      field: 'email',
      label: '邮箱',
      type: 'input',
      placeholder: '请输入邮箱',
      span: 8,
      tabKey: 'basic',
    },
    {
      field: 'mobile',
      label: '手机号',
      type: 'input',
      placeholder: '请输入手机号',
      span: 8,
      tabKey: 'basic',
    },
    {
      field: 'avatar',
      label: '头像地址',
      type: 'input',
      placeholder: '请输入头像URL',
      span: 8,
      tabKey: 'custom',
      show: false,
    },
    {
      field: 'gender',
      label: '性别',
      type: 'select',
      placeholder: '请选择性别',
      span: 8,
      tabKey: 'basic',
      options: [
        { label: '未知', value: Gender.UNKNOWN },
        { label: '男', value: Gender.MALE },
        { label: '女', value: Gender.FEMALE },
      ],
    },
    {
      field: 'roles',
      label: '角色',
      type: 'custom',
      span: 8,
      tabKey: 'custom',
      show: false,
    },
    {
      field: 'statusFlag',
      label: '状态',
      type: 'select',
      placeholder: '请选择状态',
      span: 8,
      tabKey: 'basic',
      defaultValue: UserStatus.ENABLED,
      options: [
        { label: '启用', value: UserStatus.ENABLED },
        { label: '禁用', value: UserStatus.DISABLED },
      ],
    },
    {
      field: 'deptAdminFlag',
      label: '部门管理员',
      type: 'switch',
      span: 8,
      tabKey: 'basic',
      defaultValue: FlagEnum.NO,
      props: {
        checkedValue: FlagEnum.YES,
        uncheckedValue: FlagEnum.NO,
      },
    },
    {
      field: 'tenantAdminFlag',
      label: '租户管理员',
      type: 'switch',
      span: 8,
      tabKey: 'basic',
      defaultValue: FlagEnum.NO,
      props: {
        checkedValue: FlagEnum.YES,
        uncheckedValue: FlagEnum.NO,
      },
    },
    {
      field: 'userExpireDate',
      label: '用户过期时间',
      type: 'datetime',
      placeholder: '请选择过期时间',
      span: 8,
      tabKey: 'basic',
      required: true,
      rules: [
        { required: true, message: '请选择用户过期时间', trigger: ['blur', 'change'] },
      ],
    },
    {
      field: 'lastLoginTime',
      label: '最后登录时间',
      type: 'datetime',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'lastLoginIp',
      label: '最后登录IP',
      type: 'input',
      span: 8,
      disabled: true,
      tabKey: 'other',
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
      tabKey: 'custom',
    },
  ]

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'userId',
        title: '用户ID',
        showOverflow: true,
      },
      {
        field: 'userName',
        title: '用户名',
        sortable: true,
        showOverflow: true,
      },
      {
        field: 'realName',
        title: '姓名',
        sortable: true,
        showOverflow: true,
      },
      {
        field: 'deptId',
        title: '部门ID',
        showOverflow: true,
      },
      {
        field: 'email',
        title: '邮箱',
        showOverflow: true,
      },
      {
        field: 'mobile',
        title: '手机号',
        showOverflow: true,
      },
      {
        field: 'gender',
        title: '性别',
        align: 'center',
        formatter: ({ cellValue }) => {
          switch (cellValue) {
            case Gender.MALE:
              return '男'
            case Gender.FEMALE:
              return '女'
            default:
              return '未知'
          }
        },
      },
      {
        field: 'statusFlag',
        title: '状态',
        align: 'center',
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => ({
            type: row.statusFlag === UserStatus.ENABLED ? 'success' : 'danger',
            content: row.statusFlag === UserStatus.ENABLED ? '启用' : '禁用',
          }),
        },
      },
      {
        field: 'deptAdminFlag',
        title: '部门管理员',
        align: 'center',
        formatter: ({ cellValue }) => (cellValue === 'Y' ? '是' : '否'),
      },
      {
        field: 'tenantAdminFlag',
        title: '租户管理员',
        align: 'center',
        formatter: ({ cellValue }) => (cellValue === 'Y' ? '是' : '否'),
      },
      {
        field: 'userExpireDate',
        title: '用户过期时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'lastLoginTime',
        title: '最后登录时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'lastLoginIp',
        title: '最后登录IP',
        showOverflow: true,
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
        field: 'oprSeqFlag',
        title: '操作序列标识',
        showOverflow: true,
      },
      {
        field: 'currentVersion',
        title: '当前版本号',
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'activeFlag',
        title: '活动标记',
        align: 'center',
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => ({
            type: row.activeFlag === FlagEnum.YES ? 'success' : 'default',
            content: row.activeFlag === FlagEnum.YES ? '活动' : '非活动',
          }),
        },
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
      pageInfo: pageInfo as any, // GridPaginationConfig.pageInfo 接受 Ref<PageInfoObj | undefined>
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
          // 使用 vxe-pc-ui 内置的图标类名
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
          code: 'resetPassword',
          name: '重置密码',
          prefixIcon: 'vxe-icon-lock',
        },
        {
          code: 'roleAuth',
          name: '用户授权',
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
   * 设置用户列表
   */
  const setUserList = (list: User[]) => {
    userList.value = list
  }

  /**
   * 清空用户列表
   */
  const clearUserList = () => {
    userList.value = []
  }

  /**
   * 添加用户到列表
   */
  const addUserToList = (user: User) => {
    userList.value.unshift(user)
  }

  /**
   * 更新列表中的用户
   */
  const updateUserInList = (userId: string, tenantId: string, updatedUser: Partial<User>) => {
    const index = userList.value.findIndex((u) => u.userId === userId && u.tenantId === tenantId)
    if (index !== -1) {
      // 使用 Object.assign 更新对象属性，确保 Vue 响应式系统能够检测到变化
      Object.assign(userList.value[index], updatedUser)
    }
  }

  /**
   * 从列表中删除用户
   */
  const removeUserFromList = (userId: string, tenantId: string) => {
    const index = userList.value.findIndex((u) => u.userId === userId && u.tenantId === tenantId)
    if (index !== -1) {
      userList.value.splice(index, 1)
    }
  }

  /**
   * 批量删除用户
   */
  const removeUsersFromList = (users: User[]) => {
    users.forEach((user) => {
      removeUserFromList(user.userId, user.tenantId)
    })
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    userList,
    pageInfo,

    // 配置
    searchFormConfig,
    formTabs,
    formFields,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setUserList,
    clearUserList,
    addUserToList,
    updateUserInList,
    removeUserFromList,
    removeUsersFromList,
  }
}

/**
 * Model 返回类型
 */
export type UserModel = ReturnType<typeof useUserModel>
