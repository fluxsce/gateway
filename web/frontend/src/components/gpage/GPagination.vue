<template>
  <div class="g-pagination">
    <vxe-pager
      :current-page="computedCurrentPage"
      :page-size="computedPageSize"
      :total="computedTotal"
      :page-sizes="pageSizes"
      :layouts="computedLayouts"
      :align="align"
     
      size="mini"
      @page-change="handlePageChange"
    >
    </vxe-pager>
  </div>
</template>

<script setup lang="ts">
import { computed, unref } from 'vue';
import { VxeUI } from 'vxe-pc-ui';
import type { PaginationEmits, PaginationProps } from './types';

// 获取 VxePager 组件
const VxePager = VxeUI.getComponent('VxePager')

// 定义组件名称
defineOptions({
  name: 'GPagination'
})

// Props
const props = withDefaults(defineProps<PaginationProps>(), {
  currentPage: 1,
  pageSize: 20,
  total: 0,
  pageSizes: () => [10, 20, 50, 100, 200],
  layouts: () => ['PrevJump', 'PrevPage', 'Number', 'NextPage', 'NextJump', 'Sizes', 'FullJump', 'Total'],
  align: 'right',
  background: true,
  perfect: true,
  showJumper: true,
  showTotal: true
})

// Emits
const emit = defineEmits<PaginationEmits>()

// 计算实际的分页属性（如果传入了 pageInfo，优先使用）
// 使用 unref 正确处理 ref 或普通对象
const computedCurrentPage = computed(() => {
  const pageInfo = unref(props.pageInfo)
  return pageInfo ? pageInfo.pageIndex : props.currentPage
})

const computedPageSize = computed(() => {
  const pageInfo = unref(props.pageInfo)
  return pageInfo ? pageInfo.pageSize : props.pageSize
})

const computedTotal = computed(() => {
  const pageInfo = unref(props.pageInfo)
  return pageInfo ? pageInfo.totalCount : props.total
})

// 计算布局配置
const computedLayouts = computed(() => {
  const layouts = [...props.layouts]
  
  // 如果不显示跳转，移除相关布局
  if (!props.showJumper) {
    const index = layouts.indexOf('FullJump')
    if (index > -1) {
      layouts.splice(index, 1)
    }
  }
  
  return layouts
})

// 注意：pagerOptions 暂时不使用，避免传递内部属性导致错误
// 如果需要传递额外的 vxe-pager 属性，可以在这里显式添加

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(computedTotal.value / computedPageSize.value)
})

// 处理分页变化
const handlePageChange = ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
  // 发出事件
  emit('page-change', { currentPage, pageSize })
  emit('update:currentPage', currentPage)
  emit('update:pageSize', pageSize)
}

// 暴露方法
defineExpose({
  /**
   * 获取总页数
   */
  getTotalPages: () => totalPages.value,
  
  /**
   * 跳转到指定页
   */
  goToPage: (page: number) => {
    if (page >= 1 && page <= totalPages.value) {
      handlePageChange({ currentPage: page, pageSize: computedPageSize.value })
    }
  },
  
  /**
   * 上一页
   */
  prevPage: () => {
    if (computedCurrentPage.value > 1) {
      handlePageChange({ currentPage: computedCurrentPage.value - 1, pageSize: computedPageSize.value })
    }
  },
  
  /**
   * 下一页
   */
  nextPage: () => {
    if (computedCurrentPage.value < totalPages.value) {
      handlePageChange({ currentPage: computedCurrentPage.value + 1, pageSize: computedPageSize.value })
    }
  }
})
</script>

<style lang="scss" scoped>
.g-pagination {
  width: 100%;
}
</style>

