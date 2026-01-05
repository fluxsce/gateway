<template>
  <div class="g-grid">
    <!-- 工具栏 -->
    <g-toolbar
      v-if="showToolbar"
      :module-id="moduleId"
      :buttons="toolbarButtonsComputed"
      :align="toolbarConfig?.align || 'right'"
      :bordered="false"
      :compact="true"
      class="g-grid__toolbar"
      @button-click="handleToolbarClick"
    />

    <!-- 表格区域：包装一层，用 flex 固定高度，内部 vxe-grid 自己滚动 -->
    <div class="g-grid__table-wrapper">
      <vxe-grid
        ref="gridRef"
        class="g-grid__table"
        v-bind="gridPropsComputed"
        :data="dataComputed"
        :columns="columnsComputed"
        :loading="loadingComputed"
        :menu-config="menuConfigComputed"
        :border="border"
        :stripe="stripeComputed"
        :height="height || '100%'"
        :max-height="maxHeight"
        :auto-resize="autoResize"
        :row-config="rowConfigComputed"
        :checkbox-config="checkboxConfigComputed"
        :seq-config="seqConfigComputed"
        :sort-config="props.sortConfig || (props.gridOptions as any)?.sortConfig || {}"
        :filter-config="props.filterConfig || (props.gridOptions as any)?.filterConfig"
        :edit-config="props.editConfig || (props.gridOptions as any)?.editConfig"
        :tree-config="props.treeConfig || (props.gridOptions as any)?.treeConfig"
        :expand-config="props.expandConfig || (props.gridOptions as any)?.expandConfig"
        :export-config="props.exportConfig || (props.gridOptions as any)?.exportConfig"
        :print-config="props.printConfig || (props.gridOptions as any)?.printConfig"
        :show-footer="showFooter"
        :footer-data="footerData"
        :footer-method="footerMethod"
        size="mini"
        @checkbox-change="handleCheckboxChange"
        @checkbox-all="handleCheckboxChange"
        @cell-click="handleCellClick"
        @cell-dblclick="handleCellDblclick"
        @current-change="handleRowClick"
        @sort-change="handleSortChange"
        @filter-change="handleFilterChange"
        @edit-actived="handleEditActived"
        @edit-closed="handleEditClosed"
        @menu-click="handleMenuClick"
      >
        <!-- 传递所有插槽 -->
        <template v-for="(_, name) in $slots" #[name]="slotProps">
          <slot :name="name" v-bind="slotProps || {}" />
        </template>
      </vxe-grid>
    </div>

    <!-- 分页 -->
    <g-pagination
      v-if="showPagination"
      class="g-grid__pagination"
      :current-page="paginationCurrentPage"
      :page-size="paginationPageSize"
      :total="paginationTotal"
      :page-sizes="paginationConfig?.pageSizes"
      :align="paginationConfig?.align"
      :page-info="paginationConfig?.pageInfo"
      @page-change="handlePaginationChange"
    />
  </div>
</template>

<script setup lang="ts">
import GPagination from '@/components/gpage/GPagination.vue'
import GToolbar from '@/components/toolbar/GToolbar.vue'
import { computed, ref, toValue } from 'vue'
import type { VxeGridInstance } from 'vxe-table'
import { VxeGrid } from 'vxe-table'
import type { GridEmits, GridExpose, GridProps } from './types'
import { useGrid } from './useGrid'

// 定义组件名称
defineOptions({
  name: 'GGrid'
})

// Props
const props = withDefaults(defineProps<GridProps>(), {
  loading: false,
  border: true,
  stripe: true,
  autoResize: false,
  rowId: 'id',
  showCheckbox: false,
  showSeq: true,
  showFooter: false
})

// Emits
const emit = defineEmits<GridEmits>()

// 表格引用
const gridRef = ref<VxeGridInstance>()

// 解包 ref props（保持响应式）
const dataComputed = computed(() => {
  return toValue(props.data) || []
})

const loadingComputed = computed(() => {
  return toValue(props.loading) || false
})

// 树形结构不支持 stripe，自动禁用
const stripeComputed = computed(() => {
  // 如果有 treeConfig，则禁用 stripe
  if (props.treeConfig) {
    return false
  }
  return props.stripe
})

// 使用 Grid 逻辑
const {
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
  handleCheckboxChange,
  handleCellClick,
  handleCellDblclick,
  handleRowClick,
  handleSortChange,
  handleFilterChange,
  handleMenuClick,
  handleEditActived,
  handleEditClosed,
  // 方法
  gridMethods
} = useGrid({ props, emit, gridRef })

// ============= 分页逻辑 =============

// 是否显示分页
const showPagination = computed(() => {
  return props.paginationConfig?.show === true
})

// 分页当前页（只读，从 props.paginationConfig 中派生）
const paginationCurrentPage = computed(() => {
  return props.paginationConfig?.currentPage || 1
})

// 分页每页大小（只读，从 props.paginationConfig 中派生）
const paginationPageSize = computed(() => {
  return props.paginationConfig?.pageSize || 20
})

// 分页总数
const paginationTotal = computed(() => {
  return props.paginationConfig?.total || 0
})

// 处理分页变化
const handlePaginationChange = ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
  emit('page-change', { currentPage, pageSize })
}

// 暴露方法
defineExpose<GridExpose>(gridMethods)
</script>

<style lang="scss" scoped>
.g-grid {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.g-grid__toolbar {
  flex: 0 0 auto;
}

.g-grid__table-wrapper {
  flex: 1 1 auto;
  min-height: 0; /* 允许表格内容在内部滚动，而不是撑开整个 pane */
}

.g-grid__table {
  width: 100%;
  height: 100%;
}

.g-grid__pagination {
  flex: 0 0 auto;
}
</style>

