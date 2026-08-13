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

    <!-- 分页（RsPagination，替代已弃用的 GPagination） -->
    <div
      v-if="showPagination"
      class="g-grid__pagination"
      :class="`g-grid__pagination--${paginationConfig?.align || 'right'}`"
    >
      <RsPagination
        :page="paginationCurrentPage"
        :page-size="paginationPageSize"
        :total="paginationTotal"
        :page-size-options="paginationPageSizes"
        show-page-size
        show-quick-jumper
        show-summary
        size="sm"
        @update:page="handlePageUpdate"
        @update:page-size="handlePageSizeUpdate"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import GToolbar from '@/components/toolbar/GToolbar.vue'
import { RsPagination } from '@/ui'
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
  autoResize: true,
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

const showPagination = computed(() => props.paginationConfig?.show === true)

const paginationCurrentPage = computed(() => {
  const pageInfo = toValue(props.paginationConfig?.pageInfo)
  if (pageInfo) return pageInfo.pageIndex || 1
  return props.paginationConfig?.currentPage || 1
})

const paginationPageSize = computed(() => {
  const pageInfo = toValue(props.paginationConfig?.pageInfo)
  if (pageInfo) return pageInfo.pageSize || 20
  return props.paginationConfig?.pageSize || 20
})

const paginationTotal = computed(() => {
  const pageInfo = toValue(props.paginationConfig?.pageInfo)
  if (pageInfo) return pageInfo.totalCount || 0
  return props.paginationConfig?.total || 0
})

const paginationPageSizes = computed(
  () => props.paginationConfig?.pageSizes || [10, 20, 50, 100, 200]
)

const handlePageUpdate = (page: number) => {
  emit('page-change', { currentPage: page, pageSize: paginationPageSize.value })
}

const handlePageSizeUpdate = (pageSize: number) => {
  emit('page-change', { currentPage: 1, pageSize })
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
  box-sizing: border-box;
  display: flex;
  align-items: center;
  width: 100%;
  /* 与侧栏底部折叠条同壳：高度含 border-top（--g-footer-height） */
  height: var(--g-footer-height);
  min-height: var(--g-footer-height);
  padding: 0 var(--rs-space-sm, 8px);
  border-top: 1px solid var(--rs-border);

  &--left {
    justify-content: flex-start;
  }

  &--center {
    justify-content: center;
  }

  &--right {
    justify-content: flex-end;
  }
}
</style>

