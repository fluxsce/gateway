<template>
  <div
    class="rs-grid"
    :class="{
      'rs-grid--fullscreen': isFullscreen,
      'rs-grid--fixed': hasFixedHeight,
      'rs-grid--auto': isAutoHeight
    }"
    :style="rootStyle"
  >
    <g-toolbar
      v-if="showToolbar"
      :module-id="moduleId"
      :buttons="toolbarButtonsComputed"
      :align="toolbarConfig?.align || 'right'"
      :bordered="false"
      :compact="true"
      class="rs-grid__toolbar"
      @button-click="handleToolbarClick"
    />

    <div class="rs-grid__table-wrapper">
      <RsTable
        class="rs-grid__table"
        :columns="columnsComputed"
        :data="localData"
        :loading="loadingComputed"
        :row-key="rowKeyField"
        :bordered="border"
        column-bordered
        :striped="stripe"
        :show-index="showIndex"
        :selectable="selectable"
        selection-type="checkbox"
        :selected-row-keys="selectedRowKeys"
        :height="tableHeight"
        :fill="fillTable"
        size="sm"
        compact
        resizable
        :min-column-width="120"
        highlight-row
        highlight-row-on-click
        :highlighted-row-key="currentRowKey ?? undefined"
        :context-menu="contextMenuEnabled"
        :context-menu-items="buildContextMenuItems"
        :remote-sort="remoteSort"
        :tree-config="treeConfig"
        :expanded-row-keys="expandedRowKeys"
        :default-expanded-row-keys="defaultExpandedRowKeys"
        v-model:sort="sortState"
        @update:selected-row-keys="handleSelectionChange"
        @update:highlighted-row-key="handleHighlightChange"
        @update:expanded-row-keys="handleExpandedRowKeysChange"
        @expand-change="handleExpandChange"
        @row-click="handleRowClick"
        @row-dblclick="handleRowDblclick"
        @cell-view="handleCellView"
        @update:sort="handleSortChange"
        @column-filters-change="handleFilterChange"
        @context-menu-select="handleContextMenuSelect"
      >
        <template v-for="(_, name) in $slots" #[name]="slotProps">
          <slot :name="name" v-bind="slotProps || {}" />
        </template>
      </RsTable>
    </div>

    <div
      v-if="showPagination"
      class="rs-grid__pagination"
      :class="`rs-grid__pagination--${paginationConfig?.align || 'right'}`"
    >
      <RsPagination
        :page="paginationCurrentPage"
        :page-size="paginationPageSize"
        :total="paginationTotal"
        :page-size-options="paginationPageSizes"
        show-page-size
        :show-quick-jumper="paginationShowJumper"
        :show-jump-confirm="paginationShowJumpConfirm"
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
import { RsPagination, RsTable } from '@/ui'
import { computed, toValue } from 'vue'
import type { RsGridEmits, RsGridExpose, RsGridProps } from './types'
import { useRsGrid } from './useRsGrid'

defineOptions({
  name: 'RsGrid'
})

const props = withDefaults(defineProps<RsGridProps>(), {
  loading: false,
  border: true,
  stripe: true,
  fill: true,
  rowKey: 'id',
  selectable: false,
  showIndex: true,
  remoteSort: false,
})

const emit = defineEmits<RsGridEmits>()

const loadingComputed = computed(() => toValue(props.loading) || false)

const {
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
} = useRsGrid({ props, emit })

function handleExpandedRowKeysChange(keys: string[]) {
  emit('update:expandedRowKeys', keys)
}

function handleExpandChange(keys: string[]) {
  emit('expand-change', keys)
}

/** 明确像素/长度高度（非 100%）时约束根节点，避免 height:100% 撑破父级 */
const hasFixedHeight = computed(
  () => props.height != null && props.height !== '' && props.height !== '100%'
)

/** fill=false 且无固定高度：随内容增高，不抢父容器 100% */
const isAutoHeight = computed(
  () => props.fill === false && !hasFixedHeight.value && props.height !== '100%'
)

const rootStyle = computed(() => {
  if (!hasFixedHeight.value) return undefined
  const h = props.height as string | number
  return { height: typeof h === 'number' ? `${h}px` : h }
})

/**
 * 表格区域填充策略：
 * - 根节点已固定高度时，内部用 fill 吃满剩余空间（含工具栏/分页）
 * - height 为 100% / 未指定时跟随 props.fill
 */
const fillTable = computed(() => {
  if (hasFixedHeight.value) return true
  if (props.height == null || props.height === '100%') {
    return props.fill !== false
  }
  return false
})

const tableHeight = computed(() => {
  if (fillTable.value) return undefined
  if (props.maxHeight) return props.maxHeight
  return undefined
})

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

/** 默认开启页码跳转；paginationConfig.showJumper === false 时关闭 */
const paginationShowJumper = computed(() => props.paginationConfig?.showJumper !== false)

/** 默认不显示跳转「确定」；paginationConfig.showJumpConfirm === true 时开启（回车 / 失焦即可跳转） */
const paginationShowJumpConfirm = computed(
  () => props.paginationConfig?.showJumpConfirm === true
)

const handlePageUpdate = (page: number) => {
  emit('page-change', { currentPage: page, pageSize: paginationPageSize.value })
}

const handlePageSizeUpdate = (pageSize: number) => {
  emit('page-change', { currentPage: 1, pageSize })
}

defineExpose<RsGridExpose>(gridMethods)
</script>

<style lang="scss" scoped>
.rs-grid {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  background: var(--rs-surface, var(--g-bg-primary));

  &--fixed,
  &--auto {
    /* 覆盖默认 height:100%，避免无父高时被内容/百分比撑破 */
    flex: none;
  }

  &--auto {
    height: auto;
    overflow: visible;
  }

  &--fullscreen {
    position: fixed;
    inset: 0;
    z-index: 1000;
    padding: var(--rs-space-sm, 8px);
    overflow: hidden;
  }
}

.rs-grid__toolbar {
  flex: 0 0 auto;
}

.rs-grid__table-wrapper {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.rs-grid__table {
  width: 100%;
  flex: 1 1 auto;
  min-height: 0;
}

.rs-grid__pagination {
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
