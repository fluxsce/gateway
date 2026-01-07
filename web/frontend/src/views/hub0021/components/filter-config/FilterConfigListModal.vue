<template>
  <GModal
    :visible="modalVisible"
    :title="props.title || (props.filterScope === 'global' ? '全局过滤器配置列表' : '路由过滤器配置列表')"
    :width="props.width || 1200"
    :to="props.to"
    :show-footer="false"
    @update:visible="handleUpdateVisible"
    @after-leave="handleAfterLeave"
  >
    <div class="filter-config-list-modal" id="filter-config-list-modal">
      <GPane direction="vertical" :no-resize="true">
        <!-- 上部：搜索表单 -->
        <template #1>
          <search-form
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
            @toolbar-click="handleToolbarClick"
          />
        </template>

        <!-- 下部：数据表格 -->
        <template #2>
          <g-grid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.filterList"
            :loading="service.model.loading"
            v-bind="service.model.gridConfig"
            @page-change="handlePageChange"
            @menu-click="({ code, row }) => handleMenuClick({ menu: { code }, row })"
          >
            <!-- 执行顺序自定义渲染 -->
            <template #filterOrder="{ row }">
              <span style="font-weight: bold; color: #0066cc;">{{ row.filterOrder }}</span>
            </template>

            <!-- 过滤器类型自定义渲染 -->
            <template #filterType="{ row }">
              <n-tag :type="service.model.getFilterTypeTagType(row.filterType)" size="small">
                {{ service.model.getFilterTypeLabel(row.filterType) }}
              </n-tag>
            </template>

            <!-- 执行时机自定义渲染 -->
            <template #filterAction="{ row }">
              <n-tag :type="service.model.getFilterActionTagType(row.filterAction)" size="small">
                {{ service.model.getFilterActionLabel(row.filterAction) }}
              </n-tag>
            </template>

            <!-- 状态自定义渲染 -->
            <template #activeFlag="{ row }">
              <n-switch
                :value="row.activeFlag === 'Y'"
                @update:value="() => handleToggleStatus(row)"
                size="small"
              />
            </template>
          </g-grid>
        </template>
      </GPane>

      <!-- 过滤器配置对话框（新增/编辑/查看共用） -->
      <!-- 注意：过滤器配置表单比较复杂，这里暂时使用基础表单，后续可以根据需要扩展 -->
      <GdataFormModal
        v-model:visible="formDialogVisible"
        :mode="formDialogMode"
        :title="formDialogMode === 'create' ? '新增过滤器配置' : formDialogMode === 'edit' ? '编辑过滤器配置' : '查看过滤器配置详情'"
        to="#filter-config-list-modal"
        :form-fields="service.model.formFields"
        :form-tabs="service.model.formTabs"
        :initial-data="currentEditFilter || undefined"
        :auto-close-on-confirm="false"
        :confirm-loading="service.model.loading.value"
        @submit="handleFormSubmit"
      />
    </div>
  </GModal>
</template>

<script lang="ts" setup>
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GModal } from '@/components/gmodal'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { NSwitch, NTag } from 'naive-ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { useFilterConfigPage } from './hooks'
import type { FilterConfigListModalEmits, FilterConfigListModalProps } from './hooks/types'

// 定义组件名称
defineOptions({
  name: 'FilterConfigListModal'
})

// ============= Props =============

const props = withDefaults(defineProps<FilterConfigListModalProps>(), {
  visible: false,
  title: '',
  width: 1200,
  to: undefined,
  gatewayInstanceId: undefined,
  routeConfigId: undefined,
  filterScope: 'global',
})

// ============= Emits =============

const emit = defineEmits<FilterConfigListModalEmits>()

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()
const moduleIdRef = ref<string>(props.moduleId)

// ============= 模态框可见性 =============

const modalVisible = ref(props.visible)

// 监听 props.visible 变化，同步到本地状态
const stopVisibleWatch = watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
})

// ============= 网关实例ID和路由配置ID =============

const gatewayInstanceId = ref<string | undefined>(props.gatewayInstanceId)
const routeConfigId = ref<string | undefined>(props.routeConfigId)

// 监听 props 变化
const stopGatewayInstanceIdWatch = watch(() => props.gatewayInstanceId, (newVal) => {
  gatewayInstanceId.value = newVal
})

const stopRouteConfigIdWatch = watch(() => props.routeConfigId, (newVal) => {
  routeConfigId.value = newVal
})

// ============= 资源清理 =============

// 组件卸载时清理所有监听器
onBeforeUnmount(() => {
  stopVisibleWatch()
  stopGatewayInstanceIdWatch()
  stopRouteConfigIdWatch()
})

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditFilter,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handlePageChange,
  handleToggleStatus,
} = useFilterConfigPage(moduleIdRef, gridRef, gatewayInstanceId, routeConfigId, searchFormRef)

// ============= 事件处理 =============

/**
 * 处理模态框可见性变化
 */
const handleUpdateVisible = (value: boolean) => {
  // 更新本地状态
  modalVisible.value = value
  // 通知父组件
  emit('update:visible', value)
  if (!value) {
    emit('close')
  } else {
    // 模态框打开时触发刷新事件并加载数据
    emit('refresh')
    service.loadFilterList()
  }
}

/**
 * 处理模态框关闭动画完成后的回调
 * 重置业务状态
 */
const handleAfterLeave = () => {
  if (!modalVisible.value) {
    // 重置表单对话框状态
    formDialogVisible.value = false
    formDialogMode.value = 'create'
    currentEditFilter.value = null
    // 清空列表数据
    service.model.filterList.value = []
    service.model.resetPagination()
  }
}

</script>

<style scoped>
.filter-config-list-modal {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>

