/**
 * Grid 组件逻辑
 * 统一管理 Grid 的配置、事件和方法
 */

import type { ToolbarButton } from '@/components/toolbar'
import { store } from '@/stores'
import { copyToClipboard } from '@/utils'
import { ExpandOutline, RefreshOutline, SettingsOutline } from '@vicons/ionicons5'
import { computed, type Ref } from 'vue'
import type { VxeGridInstance } from 'vxe-table'
import type { GridColumn, GridEmits, GridExpose, GridProps } from './types'

export interface UseGridOptions {
  props: GridProps
  emit: GridEmits
  gridRef: Ref<VxeGridInstance | undefined>
}

/**
 * Grid 组件主逻辑
 * 统一管理配置、事件和方法
 */
export function useGrid(options: UseGridOptions) {
  const { props, emit, gridRef } = options

  // ============= 事件处理 =============

  const handleToolbarClick = (key: string) => {
    emit('toolbar-button-click', key)
  }

  const handleRefresh = () => {
    emit('refresh')
  }

  const handleFullscreen = () => {
    if (gridRef.value) {
      gridRef.value.zoom()
    }
  }

  const handleCheckboxChange = () => {
    if (gridRef.value) {
      const selection = gridRef.value.getCheckboxRecords()
      emit('checkbox-change', selection)
    }
  }

  const handleCellClick = (params: any) => {
    emit('cell-click', params)
  }

  const handleCellDblclick = (params: any) => {
    emit('cell-dblclick', params)
  }

  const handleRowClick = (params: any) => {
    emit('row-click', params)
  }

  const handleSortChange = (params: any) => {
    emit('sort-change', params)
  }

  const handleFilterChange = (params: any) => {
    emit('filter-change', params)
  }

  const handleMenuClick = ({ menu, row, column }: any) => {
    const code = menu?.code

    // 处理默认菜单项（复制行 / 复制单元格）
    if (code === 'copyRow') {
      const rowData = JSON.stringify(row, null, 2)
      copyToClipboard(rowData)
    } else if (code === 'copyCell') {
      const cellValue = row?.[column?.field]
      copyToClipboard(String(cellValue ?? ''))
    }

    emit('menu-click', { code, row, column })

    if (props.menuConfig?.onMenuClick) {
      props.menuConfig.onMenuClick({ code, row, column })
    }
  }

  const handleEditActived = (params: any) => {
    emit('edit-actived', params)
  }

  const handleEditClosed = (params: any) => {
    emit('edit-closed', params)
  }

  // ============= 配置计算 =============

  const showToolbar = computed(() => {
    // 默认不显示工具栏，只有显式配置 toolbarConfig.show === true 时才显示
    return props.toolbarConfig?.show === true
  })

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

  const columnsComputed = computed<GridColumn[]>(() => {
    const columns: GridColumn[] = []

    // 先添加复选框列（如果启用）
    if (props.showCheckbox) {
      columns.push({
        type: 'checkbox',
        field: '__checkbox',
        width: 50,
        align: 'center',
        // 复选框列固定在左侧，确保在序号列之前
        fixed: 'left',
        // 勾选列一般不需要拖拽宽度，关闭拖拽以提升交互稳定性
        resizable: false
      } as GridColumn)
    }

    // 再添加序号列（如果启用），放在复选框后面
    if (props.showSeq) {
      columns.push({
        type: 'seq',
        field: '__seq',
        title: '序号',
        width: 60,
        align: 'center',
        // 序号列不固定，让它紧跟在复选框后面
        // fixed: 'left',
        // 序号列一般不需要拖拽宽度，避免误操作
        resizable: false,
        ...props.seqConfig
      } as GridColumn)
    }

    columns.push(...props.columns.filter((col) => col.visible !== false))

    return columns
  })

  const menuConfigComputed = computed(() => {
    // 未配置或显式关闭时，不启用右键菜单
    if (!props.menuConfig || props.menuConfig.enabled === false) {
      return undefined
    }

    // 检查菜单项权限
    const checkMenuPermission = (menuCode: string): boolean => {
      if (!props.moduleId) {
        return true // 没有 moduleId 时默认允许
      }
      const permissionCode = `${props.moduleId}:${menuCode}`
      return store.user.hasButton(permissionCode)
    }

    const defaultMenus: any[] = []

    // 默认复制整行（不需要权限校验）
    if (props.menuConfig.showCopyRow !== false) {
      defaultMenus.push({
        code: 'copyRow',
        name: '复制行数据',
        prefixIcon: 'vxe-icon-copy',
        visible: true
      })
    }

    // 默认复制单元格（不需要权限校验）
    if (props.menuConfig.showCopyCell !== false) {
      defaultMenus.push({
        code: 'copyCell',
        name: '复制单元格',
        prefixIcon: 'vxe-icon-copy',
        visible: true
      })
    }

    // 处理自定义菜单，添加权限检查
    const customMenus = (props.menuConfig.customMenus || []).map(menu => {
      const hasPermission = checkMenuPermission(menu.code)
      return {
        ...menu,
        disabled: menu.disabled || !hasPermission,
        // 如果有子菜单，也需要检查权限
        children: menu.children?.map(child => {
          const childHasPermission = checkMenuPermission(child.code)
          return {
            ...child,
            disabled: child.disabled || !childHasPermission
          }
        })
      }
    })

    // vxe-menu 要求二维数组，每一项是一个分组
    const groups: any[] = []
    const defaultGroup = [...defaultMenus]
    if (defaultGroup.length) {
      groups.push(defaultGroup)
    }
    if (customMenus.length) {
      groups.push(customMenus)
    }

    return {
      // 将菜单渲染到 body，避免被父容器的 overflow 或 z-index 遮挡
      transfer: true,
      body: {
        options: groups
      },
      visibleMethod: ({ options }: any) => {
        // 没有可见菜单时不显示右键菜单
        return Array.isArray(options) && options.some((group) => Array.isArray(group) && group.length > 0)
      }
    } as any
  })

  const rowConfigComputed = computed(() => {
    // 从 gridOptions 中获取 rowConfig，如果没有则使用默认值
    const userRowConfig = props.gridOptions?.rowConfig
    return {
      keyField: props.rowId,
      isHover: true,
      // 默认支持选中行高亮（点击行时高亮当前行），除非用户显式关闭
      isCurrent: userRowConfig?.isCurrent !== false,
      ...userRowConfig
    }
  })

  const checkboxConfigComputed = computed(() =>
    props.showCheckbox
      ? {
          checkField: '__checked',
          reserve: true,
          highlight: true
        }
      : undefined
  )

  const seqConfigComputed = computed(() =>
    props.showSeq
      ? {
          startIndex: 0,
          ...props.seqConfig
        }
      : undefined
  )

  const gridPropsComputed = computed(() => {
    const userOptions = props.gridOptions || {}

    // 默认列配置：支持拖拽调节列宽，并设置一个合理的最小宽度，避免表头换行太严重
    const columnConfig = {
      resizable: true,
      minWidth: 120,
      ...(userOptions as any).columnConfig
    }

    // 默认虚拟纵向滚动配置：默认开启虚拟滚动，提升大数据量时的性能
    // 如果用户没有自定义 virtualYConfig，则使用默认配置
    const defaultVirtualYConfig = {
      enabled: true, // 启用虚拟滚动
      gt: 0, // 总是启用虚拟滚动（0 表示总是启用）
      oSize: 5, // 每次渲染的数据偏移量，平衡渲染次数和性能
      immediate: true, // 开启实时渲染
      scrollToTopOnChange: true, // 数据源更改时自动滚动到顶部
    }

    // 如果 gridOptions 中有 expandConfig，确保它被包含在返回的对象中
    const result: any = {
      // 表头文字溢出时显示省略号 + 悬浮提示，避免换行挤压
      showHeaderOverflow: userOptions.showHeaderOverflow ?? true,
      showOverflow: userOptions.showOverflow ?? 'tooltip',
      ...userOptions,
      columnConfig,
      // 虚拟纵向滚动配置：如果用户没有配置，使用默认配置；如果用户配置了，则合并配置
      virtualYConfig: (userOptions as any).virtualYConfig
        ? {
            ...defaultVirtualYConfig,
            ...(userOptions as any).virtualYConfig,
          }
        : defaultVirtualYConfig,
    }

    // 确保 expandConfig 被正确传递（如果存在）
    if ((userOptions as any).expandConfig) {
      result.expandConfig = (userOptions as any).expandConfig
    }

    return result
  })

  // ============= 暴露方法 =============

  const gridMethods: GridExpose = {
    getGridInstance: () => gridRef.value,
    refresh: handleRefresh,
    getCheckboxRecords: () => {
      return gridRef.value?.getCheckboxRecords() || []
    },
    getCurrentRecord: () => {
      return gridRef.value?.getCurrentRecord() || null
    },
    setCheckboxRow: (rows: any[], checked: boolean) => {
      if (gridRef.value) {
        gridRef.value.setCheckboxRow(rows, checked)
      }
    },
    clearCheckboxRow: () => {
      if (gridRef.value) {
        gridRef.value.clearCheckboxRow()
      }
    },
    getTableData: () => {
      return gridRef.value?.getTableData().fullData || []
    },
    insert: async (record: any) => {
      if (gridRef.value) {
        return await gridRef.value.insert(record)
      }
    },
    insertAt: async (record: any, row: any) => {
      if (gridRef.value) {
        return await gridRef.value.insertAt(record, row)
      }
    },
    remove: async (row: any) => {
      if (gridRef.value) {
        return await gridRef.value.remove(row)
      }
    },
    removeCheckboxRow: async () => {
      if (gridRef.value) {
        const records = gridRef.value.getCheckboxRecords()
        return await gridRef.value.remove(records)
      }
    },
    getRecordset: () => {
      if (gridRef.value) {
        return gridRef.value.getRecordset()
      }
      return {
        insertRecords: [],
        removeRecords: [],
        updateRecords: []
      }
    },
    clearData: async () => {
      if (gridRef.value) {
        await gridRef.value.clearData()
      }
    },
    reloadData: async (data: any[]) => {
      if (gridRef.value) {
        await gridRef.value.reloadData(data)
      }
    },
    exportData: async (options?: any) => {
      if (gridRef.value) {
        return await gridRef.value.exportData(options)
      }
    },
    print: (options?: any) => {
      if (gridRef.value) {
        gridRef.value.print(options)
      }
    },
    zoom: () => {
      if (gridRef.value) {
        gridRef.value.zoom()
      }
    },
    validate: async (callback?: (valid: boolean) => void): Promise<boolean> => {
      if (gridRef.value) {
        const result = await gridRef.value.validate(callback)
        if (typeof result === 'object' && result !== null) {
          return false
        }
        return result as boolean
      }
      return Promise.resolve(false)
    }
  }

  return {
    // 配置
    showToolbar,
    toolbarButtonsComputed,
    columnsComputed,
    menuConfigComputed,
    rowConfigComputed,
    checkboxConfigComputed,
    seqConfigComputed,
    gridPropsComputed,
    // 事件处理
    handleToolbarClick,
    handleRefresh,
    handleFullscreen,
    handleCheckboxChange,
    handleCellClick,
    handleCellDblclick,
    handleRowClick,
    handleSortChange,
    handleFilterChange,
    handleMenuClick,
    handleEditActived,
    handleEditClosed,
    // 暴露方法
    gridMethods
  }
}

