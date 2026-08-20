/**
 * 权限资源管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态（RsSearchForm / RsGrid / RsDataForm）
 */

import type { RsDataFormField } from '@/components/form/rs-data'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type {
  RsGridColumn,
  RsGridMenuConfig,
  RsGridPaginationConfig,
} from '@/components/rs-grid'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { PageInfoObj } from '@/types/api'
import { RsTag, type RsTableTreeConfig } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, reactive, ref, watch } from 'vue'
import type { Resource } from '../types/index'
import { BuiltInFlag, FlagEnum, ResourceStatus, ResourceType } from '../types/index'

/**
 * 资源管理表格配置（对齐 RsGrid Props 子集）。
 */
export interface ResourceGridConfig {
  columns: RsGridColumn<Resource>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
  treeConfig: RsTableTreeConfig<Resource>
}

/**
 * 资源管理 Model
 */
export function useResourceModel() {
  const { t, locale } = useModuleI18n('hub0006')

  const moduleId = 'hub0006'
  const loading = ref(false)
  const resourceList = ref<Resource[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const searchFormConfig = reactive<Omit<RsSearchFormProps, 'moduleId'>>({
    fields: [],
    toolbarButtons: [],
    showSearchButton: true,
    showResetButton: true,
  })

  const formTabs = reactive<{ key: string; label: string }[]>([])
  const formFields = reactive<RsDataFormField[]>([])
  const gridConfig = reactive<ResourceGridConfig>({
    columns: [],
    selectable: false,
    rowKey: 'resourceId',
    height: '100%',
    paginationConfig: {
      show: false,
    },
    menuConfig: {
      enabled: true,
      items: [],
    },
    treeConfig: {
      childrenField: 'children',
      expandColumnKey: 'resourceId',
      defaultExpandAll: false,
      indent: 16,
    },
  })

  /**
   * 按当前语言刷新表单 / 表格文案。
   */
  function applyI18n() {
    searchFormConfig.fields = [
      {
        field: 'resourceName',
        label: t('resource.search.resourceName'),
        type: 'input',
        placeholder: t('resource.search.resourceNamePlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'resourceCode',
        label: t('resource.search.resourceCode'),
        type: 'input',
        placeholder: t('resource.search.resourceCodePlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'resourceType',
        label: t('resource.search.resourceType'),
        type: 'select',
        placeholder: t('resource.search.resourceTypePlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('resource.search.all'), value: '' },
          { label: t('resource.type.module'), value: ResourceType.MODULE },
          { label: t('resource.type.menu'), value: ResourceType.MENU },
          { label: t('resource.type.button'), value: ResourceType.BUTTON },
          { label: t('resource.type.api'), value: ResourceType.API },
        ],
      },
      {
        field: 'resourceStatus',
        label: t('resource.search.status'),
        type: 'select',
        placeholder: t('resource.search.statusPlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('resource.search.all'), value: '' },
          { label: t('resource.status.enabled'), value: ResourceStatus.ENABLED },
          { label: t('resource.status.disabled'), value: ResourceStatus.DISABLED },
        ],
      },
      {
        field: 'builtInFlag',
        label: t('resource.search.type'),
        type: 'select',
        placeholder: t('resource.search.typePlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('resource.search.all'), value: '' },
          { label: t('resource.columns.builtin'), value: BuiltInFlag.BUILT_IN },
          { label: t('resource.columns.custom'), value: BuiltInFlag.CUSTOM },
        ],
      },
    ]
    searchFormConfig.toolbarButtons = [
      {
        key: 'add',
        label: t('resource.toolbar.add'),
        icon: 'AddOutline',
        type: 'primary',
        tooltip: t('resource.toolbar.addTooltip'),
      },
      {
        key: 'edit',
        label: t('resource.toolbar.edit'),
        icon: 'CreateOutline',
        tooltip: t('resource.toolbar.editTooltip'),
      },
      {
        key: 'delete',
        label: t('resource.toolbar.delete'),
        icon: 'TrashOutline',
        type: 'error',
        tooltip: t('resource.toolbar.deleteTooltip'),
      },
      {
        key: 'view',
        label: t('resource.toolbar.view'),
        icon: 'EyeOutline',
        tooltip: t('resource.toolbar.viewTooltip'),
      },
    ]

    formTabs.splice(
      0,
      formTabs.length,
      { key: 'basic', label: t('resource.dialog.tabBasic') },
      { key: 'hierarchy', label: t('resource.dialog.tabHierarchy') },
      { key: 'other', label: t('resource.dialog.tabOther') },
    )

    formFields.splice(
      0,
      formFields.length,
      {
        field: 'resourceId',
        label: t('resource.form.resourceId'),
        type: 'input',
        placeholder: t('resource.form.resourceIdPlaceholder'),
        span: 8,
        tabKey: 'basic',
        required: true,
        primary: true,
      },
      {
        field: 'resourceName',
        label: t('resource.form.resourceName'),
        type: 'input',
        placeholder: t('resource.form.resourceNamePlaceholder'),
        span: 8,
        tabKey: 'basic',
        required: true,
      },
      {
        field: 'resourceCode',
        label: t('resource.form.resourceCode'),
        type: 'input',
        placeholder: t('resource.form.resourceCodePlaceholder'),
        span: 8,
        tabKey: 'basic',
        required: true,
      },
      {
        field: 'resourceType',
        label: t('resource.form.resourceType'),
        type: 'select',
        placeholder: t('resource.form.resourceTypePlaceholder'),
        span: 8,
        tabKey: 'basic',
        required: true,
        options: [
          { label: t('resource.type.module'), value: ResourceType.MODULE },
          { label: t('resource.type.menu'), value: ResourceType.MENU },
          { label: t('resource.type.button'), value: ResourceType.BUTTON },
          { label: t('resource.type.api'), value: ResourceType.API },
        ],
      },
      {
        field: 'resourcePath',
        label: t('resource.form.resourcePath'),
        type: 'input',
        placeholder: t('resource.form.resourcePathPlaceholder'),
        span: 12,
        tabKey: 'basic',
      },
      {
        field: 'resourceMethod',
        label: t('resource.form.resourceMethod'),
        type: 'select',
        placeholder: t('resource.form.resourceMethodPlaceholder'),
        span: 6,
        tabKey: 'basic',
        options: [
          { label: 'GET', value: 'GET' },
          { label: 'POST', value: 'POST' },
          { label: 'PUT', value: 'PUT' },
          { label: 'DELETE', value: 'DELETE' },
          { label: 'PATCH', value: 'PATCH' },
        ],
      },
      {
        field: 'displayName',
        label: t('resource.form.displayName'),
        type: 'input',
        placeholder: t('resource.form.displayNamePlaceholder'),
        span: 8,
        tabKey: 'basic',
      },
      {
        field: 'iconClass',
        label: t('resource.form.iconClass'),
        type: 'input',
        placeholder: t('resource.form.iconClassPlaceholder'),
        span: 8,
        tabKey: 'basic',
      },
      {
        field: 'description',
        label: t('resource.form.description'),
        type: 'textarea',
        placeholder: t('resource.form.descriptionPlaceholder'),
        span: 24,
        tabKey: 'basic',
      },
      {
        field: 'resourceStatus',
        label: t('resource.form.resourceStatus'),
        type: 'select',
        placeholder: t('resource.form.resourceStatusPlaceholder'),
        span: 8,
        tabKey: 'basic',
        defaultValue: ResourceStatus.ENABLED,
        options: [
          { label: t('resource.status.enabled'), value: ResourceStatus.ENABLED },
          { label: t('resource.status.disabled'), value: ResourceStatus.DISABLED },
        ],
      },
      {
        field: 'builtInFlag',
        label: t('resource.form.builtInFlag'),
        type: 'select',
        placeholder: t('resource.form.builtInFlagPlaceholder'),
        span: 8,
        tabKey: 'basic',
        defaultValue: BuiltInFlag.CUSTOM,
        options: [
          { label: t('resource.columns.builtin'), value: BuiltInFlag.BUILT_IN },
          { label: t('resource.columns.custom'), value: BuiltInFlag.CUSTOM },
        ],
      },
      {
        field: 'parentResourceId',
        label: t('resource.form.parentResourceId'),
        type: 'input',
        placeholder: t('resource.form.parentResourceIdPlaceholder'),
        span: 8,
        tabKey: 'hierarchy',
      },
      {
        field: 'resourceLevel',
        label: t('resource.form.resourceLevel'),
        type: 'number',
        placeholder: t('resource.form.resourceLevelPlaceholder'),
        span: 8,
        tabKey: 'hierarchy',
        defaultValue: 1,
      },
      {
        field: 'sortOrder',
        label: t('resource.form.sortOrder'),
        type: 'number',
        placeholder: t('resource.form.sortOrderPlaceholder'),
        span: 8,
        tabKey: 'hierarchy',
        defaultValue: 0,
      },
      {
        field: 'addTime',
        label: t('resource.form.addTime'),
        type: 'datetime',
        span: 8,
        disabled: true,
        tabKey: 'other',
      },
      {
        field: 'addWho',
        label: t('resource.form.addWho'),
        type: 'input',
        span: 8,
        disabled: true,
        tabKey: 'other',
      },
      {
        field: 'editTime',
        label: t('resource.form.editTime'),
        type: 'datetime',
        span: 8,
        disabled: true,
        tabKey: 'other',
      },
      {
        field: 'editWho',
        label: t('resource.form.editWho'),
        type: 'input',
        span: 8,
        disabled: true,
        tabKey: 'other',
      },
      {
        field: 'oprSeqFlag',
        label: t('resource.form.oprSeqFlag'),
        type: 'input',
        span: 8,
        disabled: true,
        tabKey: 'other',
      },
      {
        field: 'currentVersion',
        label: t('resource.form.currentVersion'),
        type: 'number',
        span: 8,
        disabled: true,
        tabKey: 'other',
      },
      {
        field: 'activeFlag',
        label: t('resource.form.activeFlag'),
        type: 'select',
        span: 8,
        tabKey: 'basic',
        defaultValue: FlagEnum.YES,
        options: [
          { label: t('resource.columns.active'), value: FlagEnum.YES },
          { label: t('resource.columns.inactive'), value: FlagEnum.NO },
        ],
      },
      {
        field: 'noteText',
        label: t('resource.form.noteText'),
        type: 'textarea',
        placeholder: t('resource.form.noteTextPlaceholder'),
        span: 24,
        tabKey: 'basic',
      },
    )

    gridConfig.columns = [
      {
        key: 'resourceId',
        title: t('resource.columns.resourceId'),
        width: 200,
        ellipsis: true,
      },
      {
        key: 'resourceName',
        title: t('resource.columns.resourceName'),
        sortable: true,
        ellipsis: true,
      },
      {
        key: 'resourceCode',
        title: t('resource.columns.resourceCode'),
        sortable: true,
        width: 200,
        ellipsis: true,
      },
      {
        key: 'resourceType',
        title: t('resource.columns.resourceType'),
        align: 'center',
        render: (row) => {
          const typeMap: Record<
            string,
            { variant: 'info' | 'success' | 'warning' | 'primary' | 'default'; label: string }
          > = {
            [ResourceType.MODULE]: { variant: 'info', label: t('resource.type.module') },
            [ResourceType.MENU]: { variant: 'success', label: t('resource.type.menu') },
            [ResourceType.BUTTON]: { variant: 'warning', label: t('resource.type.button') },
            [ResourceType.API]: { variant: 'primary', label: t('resource.type.api') },
          }
          const meta = typeMap[row.resourceType] || {
            variant: 'default' as const,
            label: row.resourceType,
          }
          return h(RsTag, { variant: meta.variant, size: 'sm' }, () => meta.label)
        },
      },
      {
        key: 'resourcePath',
        title: t('resource.columns.resourcePath'),
        ellipsis: true,
      },
      {
        key: 'resourceMethod',
        title: t('resource.columns.resourceMethod'),
        align: 'center',
        ellipsis: true,
      },
      {
        key: 'resourceLevel',
        title: t('resource.columns.resourceLevel'),
        align: 'center',
        sortable: true,
      },
      {
        key: 'sortOrder',
        title: t('resource.columns.sortOrder'),
        align: 'center',
        sortable: true,
      },
      {
        key: 'resourceStatus',
        title: t('resource.columns.status'),
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.resourceStatus === ResourceStatus.ENABLED ? 'success' : 'danger',
              size: 'sm',
            },
            () =>
              row.resourceStatus === ResourceStatus.ENABLED
                ? t('resource.status.enabled')
                : t('resource.status.disabled'),
          ),
      },
      {
        key: 'builtInFlag',
        title: t('resource.columns.type'),
        align: 'center',
        formatter: (value) =>
          value === BuiltInFlag.BUILT_IN
            ? t('resource.columns.builtin')
            : t('resource.columns.custom'),
      },
      {
        key: 'activeFlag',
        title: t('resource.columns.activeFlag'),
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
                ? t('resource.columns.active')
                : t('resource.columns.inactive'),
          ),
      },
      {
        key: 'addTime',
        title: t('resource.columns.addTime'),
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : ''),
      },
      {
        key: 'addWho',
        title: t('resource.columns.addWho'),
        ellipsis: true,
      },
      {
        key: 'editTime',
        title: t('resource.columns.editTime'),
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : ''),
      },
      {
        key: 'editWho',
        title: t('resource.columns.editWho'),
        ellipsis: true,
      },
      {
        key: 'description',
        title: t('resource.columns.description'),
        ellipsis: true,
      },
    ]
    gridConfig.menuConfig = {
      enabled: true,
      items: [
        { key: 'view', label: t('resource.contextMenu.view'), icon: 'EyeOutline' },
        { key: 'edit', label: t('resource.contextMenu.edit'), icon: 'CreateOutline' },
        { key: 'delete', label: t('resource.contextMenu.delete'), icon: 'TrashOutline' },
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

  const setResourceList = (list: Resource[]) => {
    resourceList.value = list
  }

  const clearResourceList = () => {
    resourceList.value = []
  }

  const addResourceToList = (resource: Resource) => {
    resourceList.value.unshift(resource)
  }

  const updateResourceInList = (
    resourceId: string,
    tenantId: string,
    updatedResource: Partial<Resource>,
  ) => {
    const index = resourceList.value.findIndex(
      (r) => r.resourceId === resourceId && r.tenantId === tenantId,
    )
    if (index !== -1) {
      Object.assign(resourceList.value[index], updatedResource)
    }
  }

  const removeResourceFromList = (resourceId: string, tenantId: string) => {
    const index = resourceList.value.findIndex(
      (r) => r.resourceId === resourceId && r.tenantId === tenantId,
    )
    if (index !== -1) {
      resourceList.value.splice(index, 1)
    }
  }

  const removeResourcesFromList = (resources: Resource[]) => {
    resources.forEach((resource) => {
      removeResourceFromList(resource.resourceId, resource.tenantId)
    })
  }

  return {
    moduleId,
    loading,
    resourceList,
    pageInfo,
    searchFormConfig,
    formTabs,
    formFields,
    gridConfig,
    resetPagination,
    updatePagination,
    setResourceList,
    clearResourceList,
    addResourceToList,
    updateResourceInList,
    removeResourceFromList,
    removeResourcesFromList,
  }
}

/**
 * Model 返回类型
 */
export type ResourceModel = ReturnType<typeof useResourceModel>
