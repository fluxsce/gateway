<template>
  <div class="service-list" :id="service.model.moduleId">
    <div v-if="!showDetailView" class="service-list-view">
      <RsSplitPane
        class="service-list__outer"
        orientation="vertical"
        :panes="outerPanes"
      >
        <template #namespace>
          <div class="service-list__pane">
            <NamespaceList
              ref="namespaceListRef"
              moduleId="hub0042:namespace"
              :show-dialog="true"
              :auto-load="true"
              @row-click="handleNamespaceRowClick"
              @namespace-select="handleNamespaceSelect"
            />
          </div>
        </template>

        <template #services>
          <div class="service-list__pane">
            <RsSplitPane
              class="service-list__inner"
              orientation="vertical"
              :panes="innerPanes"
              disabled
            >
              <template #search>
                <div class="service-list__search">
                  <RsSearchForm
                    ref="serviceSearchFormRef"
                    :module-id="service.model.moduleId"
                    v-bind="service.model.searchFormConfig"
                    @search="handleServiceSearch"
                    @toolbar-click="handleServiceToolbarClick"
                  />
                </div>
              </template>

              <template #grid>
                <div class="service-list__grid">
                  <RsGrid
                    ref="serviceGridRef"
                    :module-id="service.model.moduleId"
                    :data="service.model.serviceList"
                    :loading="service.model.loading"
                    :columns="service.model.gridConfig.columns"
                    :selectable="service.model.gridConfig.selectable"
                    :row-key="service.model.gridConfig.rowKey"
                    height="100%"
                    :pagination-config="service.model.gridConfig.paginationConfig"
                    :menu-config="service.model.gridConfig.menuConfig"
                    @page-change="handleServicePageChange"
                    @menu-click="handleServiceMenuClick"
                  />
                </div>
              </template>
            </RsSplitPane>
          </div>
        </template>
      </RsSplitPane>
    </div>

    <div v-else class="service-detail-view">
      <ServiceDetail
        :service="currentDetailService"
        :loading="detailLoading"
        @back="handleDetailBack"
        @edit="handleDetailEdit"
        @refresh="handleDetailRefresh"
        @cluster-config="handleClusterConfig"
        @edit-node="handleEditNode"
      />
    </div>

    <RsDataFormModal
      v-model:visible="serviceFormDialogVisible"
      :mode="serviceFormDialogMode"
      :title="serviceFormDialogMode === 'create' ? '新增服务' : '编辑服务'"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.serviceFormConfig.fields"
      :form-tabs="service.model.serviceFormConfig.tabs"
      :initial-data="currentEditService || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="serviceSubmitting"
      @submit="handleServiceFormSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ref } from 'vue'
import { NamespaceList } from '../hub0041/components'
import type { Namespace } from '../hub0041/types'
import ServiceDetail from './components/ServiceDetail.vue'
import { useServicePage } from './hooks'
import type { Service, ServiceNode } from './types'

defineOptions({
  name: 'ServiceList',
})

const outerPanes: RsSplitPaneItem[] = [
  { key: 'namespace', size: 35, min: 20 },
  { key: 'services' },
]

const innerPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const namespaceListRef = ref()
const serviceSearchFormRef = ref()
const serviceGridRef = ref<RsGridExpose | null>(null)

const selectedNamespace = ref<Namespace | null>(null)
const showDetailView = ref(false)
const currentDetailService = ref<Service | null>(null)
const detailLoading = ref(false)
const message = useAppMessage()

const {
  service,
  formDialogVisible: serviceFormDialogVisible,
  formDialogMode: serviceFormDialogMode,
  currentEditService,
  submitting: serviceSubmitting,
  handleServiceFormSubmit: handleServiceFormSubmitBase,
  handleToolbarClick: handleServiceToolbarClickBase,
  handleMenuClick: handleServiceMenuClickBase,
} = useServicePage(serviceGridRef, serviceSearchFormRef)

/**
 * 命名空间行点击 - 选择命名空间并加载服务列表
 */
const handleNamespaceRowClick = async (row: Namespace) => {
  if (!row) return
  selectedNamespace.value = row
  await service.loadServices({}, row.namespaceId)
}

/**
 * 命名空间选择变化
 */
const handleNamespaceSelect = (namespace: Namespace | null) => {
  selectedNamespace.value = namespace
  if (!namespace) {
    service.model.setServiceList([])
  }
}

/**
 * 服务搜索（必须选择命名空间后才能搜索）
 */
const handleServiceSearch = () => {
  if (!selectedNamespace.value) {
    message.warning('请先在上方命名空间列表中选择一个命名空间')
    return
  }
  service.handleSearch(selectedNamespace.value.namespaceId)
}

/**
 * 服务工具栏点击（必须选择命名空间后才能操作）
 */
const handleServiceToolbarClick = (key: string) => {
  if (key === 'add' && !selectedNamespace.value) {
    message.warning('请先在上方命名空间列表中选择一个命名空间')
    return
  }
  handleServiceToolbarClickBase(key, selectedNamespace.value)
}

/**
 * 服务分页变化处理
 */
const handleServicePageChange = (params: { currentPage: number; pageSize: number }) => {
  if (!selectedNamespace.value) return
  service.handlePageChange(params.currentPage, params.pageSize)
}

/**
 * 服务表单提交（自动填充命名空间ID）
 */
const handleServiceFormSubmit = (formData?: Record<string, any>) => {
  handleServiceFormSubmitBase(formData, selectedNamespace.value)
}

/**
 * 服务右键菜单点击处理
 */
const handleServiceMenuClick = async (params: { key: string; row?: Service }) => {
  if (!params.row) return
  if (params.key === 'view') {
    await openServiceDetail(params.row)
    return
  }
  await handleServiceMenuClickBase(params)
}

/**
 * 打开服务详情视图
 */
const openServiceDetail = async (serviceItem: Service) => {
  detailLoading.value = true
  try {
    const detailService = await service.getServiceDetail(
      serviceItem.namespaceId,
      serviceItem.groupName,
      serviceItem.serviceName,
    )
    if (detailService) {
      currentDetailService.value = detailService
      showDetailView.value = true
    } else {
      message.error('获取服务详情失败')
    }
  } catch (error) {
    message.error('获取服务详情失败')
  } finally {
    detailLoading.value = false
  }
}

/**
 * 返回列表视图
 */
const handleDetailBack = () => {
  showDetailView.value = false
  currentDetailService.value = null
}

/**
 * 从详情视图编辑服务
 */
const handleDetailEdit = () => {
  if (!currentDetailService.value) return
  showDetailView.value = false
  handleServiceMenuClick({
    key: 'edit',
    row: currentDetailService.value,
  })
}

/**
 * 刷新服务详情
 */
const handleDetailRefresh = async () => {
  if (!currentDetailService.value) return

  detailLoading.value = true
  try {
    const detailService = await service.getServiceDetail(
      currentDetailService.value.namespaceId,
      currentDetailService.value.groupName,
      currentDetailService.value.serviceName,
    )
    if (detailService) {
      currentDetailService.value = detailService
      message.success('服务详情已刷新')
    } else {
      message.error('刷新服务详情失败')
    }
  } catch (error) {
    message.error('刷新服务详情失败')
  } finally {
    detailLoading.value = false
  }
}

/**
 * 集群配置
 */
const handleClusterConfig = () => {
  message.info('集群配置功能开发中')
}

/**
 * 编辑节点
 */
const handleEditNode = (_node: ServiceNode) => {
  message.info('节点编辑功能开发中')
}
</script>

<style lang="scss" scoped>
.service-list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.service-list-view,
.service-detail-view {
  flex: 1;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.service-list__outer,
.service-list__inner {
  width: 100%;
  height: 100%;
  min-height: 0;
}

.service-list__pane {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.service-list__search {
  width: 100%;
  box-sizing: border-box;
}

.service-list__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
