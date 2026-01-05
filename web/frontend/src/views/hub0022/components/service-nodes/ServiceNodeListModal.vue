<template>
  <GModal
    :visible="modalVisible"
    :title="props.title || '服务节点管理'"
    :width="props.width || 1200"
    :to="props.to"
    :show-footer="false"
    @update:visible="handleUpdateVisible"
    @after-leave="handleAfterLeave"
  >
    <div class="service-node-list-modal" id="hub0022-service-node-list">
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
            :data="service.model.nodeList"
            :loading="service.model.loading"
            v-bind="service.model.gridConfig"
            @page-change="service.handlePageChange"
            @menu-click="handleMenuClick"
          >
            <!-- 协议自定义渲染 -->
            <template #nodeProtocol="{ row }">
              <n-tag :type="row.nodeProtocol === 'HTTPS' ? 'success' : 'info'" size="small">
                {{ row.nodeProtocol }}
              </n-tag>
            </template>

            <!-- 健康状态自定义渲染 -->
            <template #healthStatus="{ row }">
              <n-tag :type="row.healthStatus === 'Y' ? 'success' : 'error'" size="small">
                {{ row.healthStatus === 'Y' ? '健康' : '不健康' }}
              </n-tag>
            </template>

            <!-- 运行状态自定义渲染 -->
            <template #nodeStatus="{ row }">
              <n-tag
                :type="
                  row.nodeStatus === 1
                    ? 'success'
                    : row.nodeStatus === 0
                      ? 'error'
                      : 'warning'
                "
                size="small"
              >
                {{ row.nodeStatus === 1 ? '在线' : row.nodeStatus === 0 ? '下线' : '维护' }}
              </n-tag>
            </template>

            <!-- 启用状态自定义渲染 -->
            <template #activeFlag="{ row }">
              <n-tag :type="row.activeFlag === 'Y' ? 'success' : 'default'" size="small">
                {{ row.activeFlag === 'Y' ? '启用' : '禁用' }}
              </n-tag>
            </template>
          </g-grid>
        </template>
      </GPane>

      <!-- 服务节点对话框（新增/编辑/查看共用） -->
      <GdataFormModal
        v-model:visible="formDialogVisible"
        :mode="formDialogMode"
        :title="
          formDialogMode === 'create'
            ? '新增服务节点'
            : formDialogMode === 'edit'
              ? '编辑服务节点'
              : '查看服务节点详情'
        "
        to="#hub0022-service-node-list"
        :form-fields="service.model.nodeFormConfig.fields"
        :form-tabs="service.model.nodeFormConfig.tabs"
        :initial-data="getFormInitialData()"
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
import { NTag } from 'naive-ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { useServiceNodePage } from './hooks'
import type { ServiceNodeListModalEmits, ServiceNodeListModalProps } from './types'

// 定义组件名称
defineOptions({
  name: 'ServiceNodeListModal',
})

// ============= Props =============

const props = withDefaults(defineProps<ServiceNodeListModalProps>(), {
  visible: false,
  title: '服务节点管理',
  width: 1200,
  to: undefined,
  serviceDefinitionId: undefined,
})

// ============= Emits =============

const emit = defineEmits<ServiceNodeListModalEmits>()

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 模态框可见性 =============

const modalVisible = ref(props.visible)

// 监听 props.visible 变化，同步到本地状态
const stopVisibleWatch = watch(
  () => props.visible,
  (newVal) => {
    modalVisible.value = newVal
  }
)

// ============= 服务定义ID =============

const serviceDefinitionId = ref<string | undefined>(props.serviceDefinitionId)

// 监听 props.serviceDefinitionId 变化
const stopServiceDefinitionIdWatch = watch(
  () => props.serviceDefinitionId,
  (newVal) => {
    serviceDefinitionId.value = newVal
    // 如果模态框已打开且 serviceDefinitionId 变化，重新加载数据
    if (modalVisible.value && newVal) {
      service.handleRefresh()
    }
  }
)

// ============= 资源清理 =============

// 组件卸载时清理所有监听器
onBeforeUnmount(() => {
  stopVisibleWatch()
  stopServiceDefinitionIdWatch()
})

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  getFormInitialData,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useServiceNodePage(gridRef, serviceDefinitionId, searchFormRef)

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
    // 模态框打开时触发刷新事件
    emit('refresh')
    // 如果 serviceDefinitionId 存在，加载数据
    if (serviceDefinitionId.value) {
      service.handleRefresh()
    }
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
    // 清空列表数据
    service.model.nodeList.value = []
    service.model.resetPagination()
  }
}
</script>

<style scoped>
.service-node-list-modal {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>

