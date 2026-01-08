<template>
  <GModal
    :visible="modalVisible"
    :title="props.title || `静态节点管理 - ${props.serverName || ''}`"
    :width="props.width || 1400"
    :to="props.to"
    :show-footer="false"
    @update:visible="handleUpdateVisible"
    @after-leave="handleAfterLeave"
  >
    <div class="static-node-list-modal" :id="htmlId">
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
            @page-change="handlePageChange"
            @menu-click="({ code, row }) => handleMenuClick({ menu: { code }, row })"
          >
            <!-- 代理类型自定义渲染 -->
            <template #proxyType="{ row }">
              <n-tag :type="row.proxyType === 'tcp' ? 'primary' : 'info'" size="small">
                {{ service.model.getProxyTypeLabel(row.proxyType) }}
              </n-tag>
            </template>

            <!-- 节点状态自定义渲染 -->
            <template #nodeStatus="{ row }">
              <n-tag :type="service.model.getNodeStatusTagType(row.nodeStatus)" size="small">
                {{ service.model.getNodeStatusLabel(row.nodeStatus) }}
              </n-tag>
            </template>

            <!-- 健康检查状态自定义渲染 -->
            <template #healthCheckStatus="{ row }">
              <n-tag :type="service.model.getHealthCheckStatusTagType(row.healthCheckStatus)" size="small">
                {{ service.model.getHealthCheckStatusLabel(row.healthCheckStatus) }}
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

      <!-- 节点配置对话框（新增/编辑/查看共用） -->
      <GdataFormModal
        v-model:visible="formDialogVisible"
        :mode="formDialogMode"
        :title="formDialogMode === 'create' ? '新增静态节点' : formDialogMode === 'edit' ? '编辑静态节点' : '查看静态节点详情'"
        :to="`#${htmlId}`"
        :form-fields="service.model.formFields"
        :form-tabs="service.model.formTabs"
        :initial-data="currentEditNode || undefined"
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
import { useStaticNodePage } from './hooks'
import type { StaticNodeListModalEmits, StaticNodeListModalProps } from './hooks/types'

// 定义组件名称
defineOptions({
  name: 'StaticNodeListModal'
})

// ============= Props =============

const props = withDefaults(defineProps<StaticNodeListModalProps>(), {
  visible: false,
  title: '',
  width: 1400,
  to: undefined,
  tunnelStaticServerId: '',
  serverName: '',
})

// ============= Emits =============

const emit = defineEmits<StaticNodeListModalEmits>()

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 模态框可见性 =============

const modalVisible = ref(props.visible)

// 监听 props.visible 变化，同步到本地状态
const stopVisibleWatch = watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
})

// ============= 静态服务器ID =============

const tunnelStaticServerId = ref<string>(props.tunnelStaticServerId)

// 监听 props 变化
const stopServerIdWatch = watch(() => props.tunnelStaticServerId, (newVal) => {
  tunnelStaticServerId.value = newVal
})

// ============= 资源清理 =============

// 组件卸载时清理所有监听器
onBeforeUnmount(() => {
  stopVisibleWatch()
  stopServerIdWatch()
})

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditNode,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handlePageChange,
  handleToggleStatus,
} = useStaticNodePage(gridRef, tunnelStaticServerId, searchFormRef)

// ============= HTML ID（用于 DOM，符合 HTML 规范） =============

// 固定的 HTML id（符合 HTML 规范，无特殊字符）
// 注意：权限校验仍使用原始 moduleId（service.model.moduleId）
const htmlId = 'hub0061-static-nodes'

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
    service.loadNodeList()
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
    currentEditNode.value = null
    // 清空列表数据
    service.model.nodeList.value = []
    service.model.resetPagination()
  }
}

</script>

<style scoped>
.static-node-list-modal {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>

