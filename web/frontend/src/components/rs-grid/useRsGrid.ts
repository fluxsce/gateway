/**
 * RsGrid 逻辑层：列映射、选择、右键菜单、暴露方法（仅面向 niuma-ui RsTable）。
 */

import type { ToolbarButton } from '@/components/toolbar'
import { useAppMessage } from '@/composables/useAppMessage'
import { store } from '@/stores'
import {
  type RsContextMenuItem,
  type RsTableColumn,
  type RsTableRowData,
  type RsTableSortState
} from '@/ui'
import { ExpandOutline, RefreshOutline, SettingsOutline } from '@vicons/ionicons5'
import { computed, ref, toValue, watch } from 'vue'
import { mapRsGridMenuIcon } from './mapMenuIcon'
import type { RsGridColumn, RsGridEmits, RsGridExpose, RsGridMenuItem, RsGridProps } from './types'

/**
 * useRsGrid 入参
 */
export interface UseRsGridOptions {
  props: RsGridProps
  emit: RsGridEmits
}

/**
 * 将 RsGridColumn 映射为 RsTableColumn（无旧版兼容分支）。
 */
export function mapColumn(col: RsGridColumn): RsTableColumn<RsTableRowData> {
  const key = col.key
  return {
    key,
    title: col.title,
    dataIndex: col.dataIndex ?? key,
    width: col.width,
    minWidth: col.minWidth ?? 120,
    align: col.align,
    fixed: col.fixed,
    sortable: col.sortable,
    filterable: col.filterable,
    ellipsis: col.ellipsis,
    formatter: col.formatter,
    render: col.render
  }
}

/**
 * RsGrid 组件主逻辑
 */
export function useRsGrid(options: UseRsGridOptions) {
  const { props, emit } = options

  const selectedRowKeys = ref<string[]>([])
  const currentRowKey = ref<string | null>(null)
  const localData = ref<any[]>([])
  const sortState = ref<RsTableSortState | null>(null)
  const isFullscreen = ref(false)

  const rowKeyField = computed(() => props.rowKey || 'id')

  const resolveRowKey = (row: any): string => {
    const key = row?.[rowKeyField.value]
    return key == null ? '' : String(key)
  }

  const pruneSelectionAgainstData = () => {
    const keySet = new Set(localData.value.map(resolveRowKey).filter(Boolean))
    selectedRowKeys.value = selectedRowKeys.value.filter((k) => keySet.has(k))
    if (currentRowKey.value && !keySet.has(currentRowKey.value)) {
      currentRowKey.value = null
    }
  }

  /**
   * 同步外部 data → 本地副本。
   * 仅浅监听数组引用：就地 Object.assign / splice 由业务直接改同一引用时，
   * 表格通过行对象响应式更新，避免 deep watch 触发整表浅拷贝。
   */
  watch(
    () => toValue(props.data),
    (value) => {
      localData.value = Array.isArray(value) ? value : []
      pruneSelectionAgainstData()
    },
    { immediate: true },
  )

  const handleToolbarClick = (key: string) => {
    emit('toolbar-button-click', key)
  }

  const handleRefresh = () => {
    emit('refresh')
  }

  const handleFullscreen = () => {
    isFullscreen.value = !isFullscreen.value
  }

  const emitSelectionChange = () => {
    emit('selection-change', getSelectedRows())
  }

  const handleSelectionChange = (keys: string[]) => {
    selectedRowKeys.value = keys
    emitSelectionChange()
  }

  const handleRowClick = (row: any) => {
    const key = resolveRowKey(row)
    if (key) currentRowKey.value = key
    emit('row-click', { row })
  }

  const handleHighlightChange = (key: string | undefined) => {
    currentRowKey.value = key ?? null
  }

  const handleRowDblclick = (row: any) => {
    emit('row-dblclick', { row })
  }

  const handleCellView = (row: any, column: any, index: number) => {
    emit('cell-click', { row, column, index })
  }

  const handleSortChange = (sort: RsTableSortState | null) => {
    sortState.value = sort
    emit('sort-change', sort)
  }

  const handleFilterChange = (filters: Record<string, string>) => {
    emit('filter-change', filters)
  }

  const message = useAppMessage()

  const handleContextMenuSelect = (key: string, row: any | null) => {
    // 内置复制由 RsTable 写入剪贴板；此处补齐与旧 GGrid 一致的成功提示
    if (key === 'copyCell' || key === 'copyRow') {
      message.success('已复制到剪贴板')
    }
    emit('menu-click', { key, row: row ?? undefined })
    props.menuConfig?.onMenuClick?.({ key, row: row ?? undefined })
  }

  const showToolbar = computed(() => props.toolbarConfig?.show === true)

  const toolbarButtonsComputed = computed<ToolbarButton[]>(() => {
    const buttons: ToolbarButton[] = []

    if (props.toolbarConfig?.buttons) {
      buttons.push(...props.toolbarConfig.buttons)
    }

    if (props.toolbarConfig?.showRefresh !== false) {
      buttons.push({
        key: 'refresh',
        label: '刷新',
        icon: RefreshOutline,
        tooltip: '刷新数据',
        onClick: handleRefresh
      })
    }

    if (props.toolbarConfig?.showColumnSetting) {
      buttons.push({
        key: 'column-setting',
        label: '列设置',
        icon: SettingsOutline,
        tooltip: '列设置'
      })
    }

    if (props.toolbarConfig?.showFullscreen) {
      buttons.push({
        key: 'fullscreen',
        label: '全屏',
        icon: ExpandOutline,
        tooltip: '全屏显示',
        onClick: handleFullscreen
      })
    }

    return buttons
  })

  const columnsComputed = computed<RsTableColumn<RsTableRowData>[]>(() => {
    return props.columns.filter((col) => col.visible !== false).map(mapColumn)
  })

  const contextMenuEnabled = computed(() => {
    if (!props.menuConfig) return true
    return props.menuConfig.enabled !== false
  })

  /**
   * 构建业务右键菜单；内置复制由 RsTable 处理。
   * 支持 separator / children 嵌套分组。
   */
  const buildContextMenuItems = (row: any | null): RsContextMenuItem[] => {
    if (!props.menuConfig || props.menuConfig.enabled === false || !row) {
      return []
    }

    const checkPermission = (menuKey: string): boolean => {
      if (!props.moduleId) return true
      return store.user.hasPermission(`${props.moduleId}:${menuKey}`)
    }

    const mapItem = (item: RsGridMenuItem): RsContextMenuItem | null => {
      if (item.separator) {
        return { key: `sep-${item.key}`, label: '', separator: true }
      }

      const children = (item.children || [])
        .map(mapItem)
        .filter((child): child is RsContextMenuItem => child != null)

      // 分组项：有子菜单时不按父 key 鉴权（避免文件夹 key 无权限导致整组不可用）
      if (children.length > 0) {
        const actionable = children.filter((child) => !child.separator)
        if (actionable.length === 0) return null
        const allDisabled = actionable.every((child) => child.disabled)
        return {
          key: item.key,
          label: item.label,
          icon: mapRsGridMenuIcon(item.icon),
          disabled: Boolean(item.disabled || allDisabled),
          danger: item.danger,
          children,
        }
      }

      const allowed = checkPermission(item.key)
      return {
        key: item.key,
        label: item.label,
        icon: mapRsGridMenuIcon(item.icon),
        disabled: Boolean(item.disabled || !allowed),
        danger: item.danger ?? item.key === 'delete',
        shortcut: item.shortcut,
      }
    }

    return (props.menuConfig.items || [])
      .map(mapItem)
      .filter((item): item is RsContextMenuItem => item != null)
  }

  /** 当前勾选行（按 localData 顺序） */
  function getSelectedRows(): any[] {
    const keySet = new Set(selectedRowKeys.value)
    return localData.value.filter((row) => keySet.has(resolveRowKey(row)))
  }

  /** 当前高亮行（单击行写入的 currentRowKey） */
  function getCurrentRow(): any | null {
    if (!currentRowKey.value) return null
    return localData.value.find((row) => resolveRowKey(row) === currentRowKey.value) || null
  }

  /**
   * 业务「当前操作行」：优先首条勾选，否则高亮行。
   * 用于工具栏/右键在未勾选时仍能拿到上下文行。
   */
  function getActiveRow(): any | null {
    const selected = getSelectedRows()
    if (selected.length > 0) return selected[0]
    return getCurrentRow()
  }

  /** 勾选全部；无勾选时返回高亮行 0～1 条 */
  function getActiveRows(): any[] {
    const selected = getSelectedRows()
    if (selected.length > 0) return selected
    const current = getCurrentRow()
    return current ? [current] : []
  }

  const gridMethods: RsGridExpose = {
    refresh: handleRefresh,
    getSelectedRows,
    getCurrentRow,
    getActiveRow,
    /** 兼容旧 GGrid API 名 */
    getSelectedOrCurrentRecord: getActiveRow,
    getActiveRows,
    setSelectedRows: (rows: any[], selected: boolean) => {
      const keys = rows.map(resolveRowKey).filter(Boolean)
      if (selected) {
        selectedRowKeys.value = Array.from(new Set([...selectedRowKeys.value, ...keys]))
      } else {
        const removeSet = new Set(keys)
        selectedRowKeys.value = selectedRowKeys.value.filter((k) => !removeSet.has(k))
      }
      emitSelectionChange()
    },
    clearSelection: () => {
      selectedRowKeys.value = []
      emitSelectionChange()
    },
    getData: () => [...localData.value],
    clearData: async () => {
      localData.value = []
      selectedRowKeys.value = []
      currentRowKey.value = null
    },
    reloadData: async (data: any[]) => {
      localData.value = Array.isArray(data) ? [...data] : []
      selectedRowKeys.value = []
      currentRowKey.value = null
    },
    /**
     * 按 rowKey 合并补丁到本地行（Object.assign，不替换数组引用）。
     * 配合浅监听 props.data：就地回写时不必整表 reload / 浅拷贝。
     */
    patchRows: (patches: Array<{ key: string; data: Record<string, any> }>) => {
      if (!patches.length) return
      const map = new Map(patches.map((item) => [item.key, item.data]))
      for (const row of localData.value) {
        const key = resolveRowKey(row)
        const patch = map.get(key)
        if (patch) Object.assign(row, patch)
      }
    },
    /**
     * 替换整表数据（采用传入数组引用，不再深拷贝）。
     * 会 prune 已失效的 selectedRowKeys / currentRowKey。
     */
    replaceRows: (rows: any[]) => {
      localData.value = Array.isArray(rows) ? rows : []
      pruneSelectionAgainstData()
    },
    toggleFullscreen: handleFullscreen
  }

  return {
    selectedRowKeys,
    currentRowKey,
    localData,
    sortState,
    isFullscreen,
    rowKeyField,
    showToolbar,
    toolbarButtonsComputed,
    columnsComputed,
    contextMenuEnabled,
    buildContextMenuItems,
    handleToolbarClick,
    handleSelectionChange,
    handleRowClick,
    handleHighlightChange,
    handleRowDblclick,
    handleCellView,
    handleSortChange,
    handleFilterChange,
    handleContextMenuSelect,
    gridMethods
  }
}
