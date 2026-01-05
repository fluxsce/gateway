import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid/types'
import type { ToolbarProps } from '@/components/toolbar/types'

/**
 * 网格表单配置
 * 用于统一配置搜索表单、数据表单、表格和工具栏
 */
export interface GridFormConfig {
  /**
   * 搜索表单配置
   */
  searchForm?: Partial<SearchFormProps>

  /**
   * 编辑表单配置（用于新增/编辑弹窗）
   */
  editForm?: {
    fields?: DataFormField[]
    labelWidth?: number | string
    labelPlacement?: 'left' | 'top'
    showResetButton?: boolean
    showSubmitButton?: boolean
    [key: string]: any
  }

  /**
   * 表格配置
   */
  grid?: Partial<GridProps>

  /**
   * 工具栏配置
   */
  toolbar?: {
    buttons?: ToolbarProps['buttons']
    [key: string]: any
  }
}

