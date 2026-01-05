<template>
  <GModal
    :visible="modalVisible"
    :title="props.title || '选择路由'"
    :width="props.width || 1200"
    :to="props.to"
    :show-footer="false"
    @update:visible="handleUpdateVisible"
    @after-leave="handleAfterLeave"
  >
    <div class="route-list-modal" :id="service.model.moduleId">
      <GPane direction="vertical" :no-resize="true">
        <!-- 上部：搜索表单 -->
        <template #1>
          <search-form
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
          />
        </template>

        <!-- 下部：数据表格 -->
        <template #2>
          <g-grid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.routeList"
            :loading="service.model.loading"
            v-bind="service.model.gridConfig"
            @page-change="service.handlePageChange"
            @row-click="handleRowClick"
          >
            <!-- 匹配类型自定义渲染 -->
            <template #matchType="{ row }">
              <n-tag :type="service.model.getMatchTypeTagType(row.matchType)" size="small">
                {{ service.model.getMatchTypeLabel(row.matchType) }}
              </n-tag>
            </template>

            <!-- 状态自定义渲染 -->
            <template #activeFlag="{ row }">
              <n-tag :type="row.activeFlag === 'Y' ? 'success' : 'error'" size="small">
                {{ row.activeFlag === 'Y' ? '启用' : '禁用' }}
              </n-tag>
            </template>
          </g-grid>
        </template>
      </GPane>
    </div>
  </GModal>
</template>

<script lang="ts" setup>
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GModal } from '@/components/gmodal'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { NTag } from 'naive-ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { useRouteListPage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'RouteListModal'
})

// ============= Props =============

interface Props {
  /** 是否显示弹窗 */
  visible?: boolean
  /** 弹窗标题 */
  title?: string
  /** 弹窗宽度 */
  width?: number | string
  /** 弹窗挂载目标 */
  to?: string
  /** 选中的路由名称（v-model） */
  modelValue?: string
  /** 网关实例ID（可选，用于过滤路由） */
  gatewayInstanceId?: string
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  title: '',
  width: 1200,
  to: undefined,
  modelValue: '',
  gatewayInstanceId: undefined,
})

// ============= Emits =============

interface Emits {
  (e: 'update:visible', visible: boolean): void
  (e: 'after-leave'): void
  (e: 'select', route: any): void // 保留 select 事件以保持向后兼容
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 模态框可见性 =============

const modalVisible = ref(props.visible)

// 监听 props.visible 变化，同步到本地状态
const stopVisibleWatch = watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
  if (newVal) {
    // 弹窗打开时，自动执行查询
    handleSearch()
  }
})

// ============= 页面级 Hook（只包含查询功能） =============

const {
  service,
  handleSearch,
} = useRouteListPage(props.gatewayInstanceId, gridRef, searchFormRef)

// ============= 事件处理 =============

/**
 * 处理弹窗可见性更新
 */
const handleUpdateVisible = (visible: boolean) => {
  modalVisible.value = visible
  emit('update:visible', visible)
}

/**
 * 处理弹窗关闭后事件
 */
const handleAfterLeave = () => {
  emit('after-leave')
}

/**
 * 处理行点击事件（选择路由）
 */
const handleRowClick = ({ row }: { row: any }) => {
  if (row) {
    const routeName = row.routeName || ''
    emit('update:modelValue', routeName) // 更新 v-model
    emit('select', row) // 触发 select 事件（兼容旧用法）
    // 选择后关闭弹窗
    handleUpdateVisible(false)
  }
}

// ============= 生命周期 =============

onBeforeUnmount(() => {
  stopVisibleWatch()
})
</script>

<style lang="scss" scoped>
.route-list-modal {
  width: 100%;
  height: 100%;
  min-height: 500px;

  :deep(.n-split) {
    height: 100%;
  }

  /* 上半区：搜索表单，内容较少，允许自身滚动 */
  :deep(.n-split-pane:first-child) {
    overflow: auto;
    padding: var(--g-space-sm);
  }

  /* 下半区：表格区域，高度由 GGrid 占满，滚动全部交给 vxe-grid */
  :deep(.n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);
    display: flex;
    flex-direction: column;
  }
}
</style>

