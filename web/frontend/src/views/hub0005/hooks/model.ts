/**
 * 角色管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态（RsSearchForm / RsGrid / RsDataForm）
 */

import type { RsDataFormField } from '@/components/form/rs-data'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { PageInfoObj } from '@/types/api'
import { RsTag } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, reactive, ref, watch } from 'vue'
import type { Role } from '../types/index'
import { BuiltInFlag, FlagEnum, RoleStatus } from '../types/index'

/**
 * 角色管理表格配置（对齐 RsGrid Props 子集）。
 */
export interface RoleGridConfig {
  columns: RsGridColumn<Role>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * 角色管理 Model
 */
export function useRoleModel() {
  const { t, locale } = useModuleI18n('hub0005')

  const moduleId = 'hub0005'
  const loading = ref(false)
  const roleList = ref<Role[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const searchFormConfig = reactive<Omit<RsSearchFormProps, 'moduleId'>>({
    fields: [],
    toolbarButtons: [],
    showSearchButton: true,
    showResetButton: true,
  })

  const formTabs = reactive<{ key: string; label: string }[]>([])
  const formFields = reactive<RsDataFormField[]>([])
  const gridConfig = reactive<RoleGridConfig>({
    columns: [],
    selectable: true,
    rowKey: 'roleId',
    height: '100%',
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      items: [],
    },
  })

  /**
   * 按当前语言刷新表单 / 表格文案。
   */
  function applyI18n() {
    searchFormConfig.fields = [
      {
        field: 'roleName',
        label: t('role.search.roleName'),
        type: 'input',
        placeholder: t('role.search.roleNamePlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'roleStatus',
        label: t('role.search.status'),
        type: 'select',
        placeholder: t('role.search.statusPlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('role.search.all'), value: '' },
          { label: t('role.status.enabled'), value: RoleStatus.ENABLED },
          { label: t('role.status.disabled'), value: RoleStatus.DISABLED },
        ],
      },
      {
        field: 'builtInFlag',
        label: t('role.search.type'),
        type: 'select',
        placeholder: t('role.search.typePlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('role.search.all'), value: '' },
          { label: t('role.columns.builtin'), value: BuiltInFlag.BUILT_IN },
          { label: t('role.columns.custom'), value: BuiltInFlag.CUSTOM },
        ],
      },
    ]
    searchFormConfig.toolbarButtons = [
      {
        key: 'add',
        label: t('role.toolbar.add'),
        icon: 'AddOutline',
        type: 'primary',
        tooltip: t('role.toolbar.addTooltip'),
      },
      {
        key: 'edit',
        label: t('role.toolbar.edit'),
        icon: 'CreateOutline',
        tooltip: t('role.toolbar.editTooltip'),
      },
      {
        key: 'delete',
        label: t('role.toolbar.delete'),
        icon: 'TrashOutline',
        type: 'error',
        tooltip: t('role.toolbar.deleteTooltip'),
      },
    ]

    formTabs.splice(
      0,
      formTabs.length,
      { key: 'basic', label: t('role.dialog.tabBasic') },
      { key: 'other', label: t('role.dialog.tabOther') },
    )

    formFields.splice(
      0,
      formFields.length,
      {
        field: 'roleId',
        label: t('role.formExtra.roleId'),
        type: 'input',
        placeholder: t('role.formExtra.roleIdPlaceholder'),
        span: 8,
        tabKey: 'basic',
        required: true,
        primary: true,
      },
      {
        field: 'roleName',
        label: t('role.form.roleName'),
        type: 'input',
        placeholder: t('role.form.placeholder.roleName'),
        span: 8,
        tabKey: 'basic',
        required: true,
      },
      {
        field: 'roleDescription',
        label: t('role.form.roleDescription'),
        type: 'textarea',
        placeholder: t('role.form.placeholder.roleDescription'),
        span: 24,
        tabKey: 'basic',
      },
      {
        field: 'roleStatus',
        label: t('role.form.roleStatus'),
        type: 'select',
        placeholder: t('role.formExtra.statusPlaceholder'),
        span: 8,
        tabKey: 'basic',
        defaultValue: RoleStatus.ENABLED,
        options: [
          { label: t('role.status.enabled'), value: RoleStatus.ENABLED },
          { label: t('role.status.disabled'), value: RoleStatus.DISABLED },
        ],
      },
      {
        field: 'builtInFlag',
        label: t('role.formExtra.builtInFlag'),
        type: 'select',
        placeholder: t('role.formExtra.typePlaceholder'),
        span: 8,
        tabKey: 'basic',
        defaultValue: BuiltInFlag.CUSTOM,
        options: [
          { label: t('role.columns.builtin'), value: BuiltInFlag.BUILT_IN },
          { label: t('role.columns.custom'), value: BuiltInFlag.CUSTOM },
        ],
      },
      {
        field: 'dataScope',
        label: t('role.form.dataScope'),
        type: 'textarea',
        placeholder: t('role.formExtra.dataScopePlaceholder'),
        span: 24,
        tabKey: 'basic',
      },
      {
        field: 'addTime',
        label: t('role.formExtra.addTime'),
        type: 'datetime',
        span: 8,
        disabled: true,
        tabKey: 'other',
      },
      {
        field: 'addWho',
        label: t('role.formExtra.addWho'),
        type: 'input',
        span: 8,
        disabled: true,
        tabKey: 'other',
      },
      {
        field: 'editTime',
        label: t('role.formExtra.editTime'),
        type: 'datetime',
        span: 8,
        disabled: true,
        tabKey: 'other',
      },
      {
        field: 'editWho',
        label: t('role.formExtra.editWho'),
        type: 'input',
        span: 8,
        disabled: true,
        tabKey: 'other',
      },
      {
        field: 'oprSeqFlag',
        label: t('role.formExtra.oprSeqFlag'),
        type: 'input',
        span: 8,
        tabKey: 'other',
        disabled: true,
      },
      {
        field: 'currentVersion',
        label: t('role.formExtra.currentVersion'),
        type: 'number',
        span: 8,
        disabled: true,
        tabKey: 'other',
      },
      {
        field: 'activeFlag',
        label: t('role.formExtra.activeFlag'),
        type: 'select',
        span: 8,
        tabKey: 'basic',
        defaultValue: FlagEnum.YES,
        options: [
          { label: t('role.columns.active'), value: FlagEnum.YES },
          { label: t('role.columns.inactive'), value: FlagEnum.NO },
        ],
      },
      {
        field: 'noteText',
        label: t('role.formExtra.noteText'),
        type: 'textarea',
        placeholder: t('role.formExtra.noteTextPlaceholder'),
        span: 24,
        tabKey: 'basic',
      },
    )

    gridConfig.columns = [
      {
        key: 'roleId',
        title: t('role.columns.roleId'),
        ellipsis: true,
      },
      {
        key: 'roleName',
        title: t('role.columns.roleName'),
        sortable: true,
        ellipsis: true,
      },
      {
        key: 'roleDescription',
        title: t('role.columns.roleDescription'),
        ellipsis: true,
      },
      {
        key: 'roleStatus',
        title: t('role.columns.status'),
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.roleStatus === RoleStatus.ENABLED ? 'success' : 'danger',
              size: 'sm',
            },
            () =>
              row.roleStatus === RoleStatus.ENABLED
                ? t('role.status.enabled')
                : t('role.status.disabled'),
          ),
      },
      {
        key: 'builtInFlag',
        title: t('role.columns.type'),
        align: 'center',
        formatter: (value) =>
          value === BuiltInFlag.BUILT_IN ? t('role.columns.builtin') : t('role.columns.custom'),
      },
      {
        key: 'activeFlag',
        title: t('role.columns.activeFlag'),
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.activeFlag === FlagEnum.YES ? 'success' : 'default',
              size: 'sm',
            },
            () =>
              row.activeFlag === FlagEnum.YES
                ? t('role.columns.active')
                : t('role.columns.inactive'),
          ),
      },
      {
        key: 'addTime',
        title: t('role.columns.addTime'),
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : ''),
      },
      {
        key: 'addWho',
        title: t('role.columns.addWho'),
        ellipsis: true,
      },
      {
        key: 'editTime',
        title: t('role.columns.editTime'),
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : ''),
      },
      {
        key: 'editWho',
        title: t('role.columns.editWho'),
        ellipsis: true,
      },
      {
        key: 'noteText',
        title: t('role.columns.noteText'),
        ellipsis: true,
      },
    ]
    gridConfig.menuConfig = {
      enabled: true,
      items: [
        { key: 'view', label: t('role.contextMenu.view'), icon: 'EyeOutline' },
        { key: 'edit', label: t('role.contextMenu.edit'), icon: 'CreateOutline' },
        { key: 'delete', label: t('role.contextMenu.delete'), icon: 'TrashOutline', danger: true },
        { key: 'roleAuth', label: t('role.contextMenu.roleAuth'), icon: 'PersonOutline' },
      ],
    }
  }

  watch(locale, applyI18n, { immediate: true })

  const resetPagination = () => {
    pageInfo.value = undefined
  }

  const updatePagination = (newPageInfo: Partial<PageInfoObj>) => {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  const setRoleList = (list: Role[]) => {
    roleList.value = list
  }

  const clearRoleList = () => {
    roleList.value = []
  }

  const addRoleToList = (role: Role) => {
    roleList.value.unshift(role)
  }

  const updateRoleInList = (roleId: string, tenantId: string, updatedRole: Partial<Role>) => {
    const index = roleList.value.findIndex((r) => r.roleId === roleId && r.tenantId === tenantId)
    if (index !== -1) {
      Object.assign(roleList.value[index], updatedRole)
    }
  }

  const removeRoleFromList = (roleId: string, tenantId: string) => {
    const index = roleList.value.findIndex((r) => r.roleId === roleId && r.tenantId === tenantId)
    if (index !== -1) {
      roleList.value.splice(index, 1)
    }
  }

  const removeRolesFromList = (roles: Role[]) => {
    roles.forEach((role) => {
      removeRoleFromList(role.roleId, role.tenantId)
    })
  }

  return {
    moduleId,
    loading,
    roleList,
    pageInfo,
    searchFormConfig,
    formTabs,
    formFields,
    gridConfig,
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
