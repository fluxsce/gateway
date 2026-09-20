<template>
  <div class="service-list" :id="service.model.moduleId">
    <div v-show="!showDetailView" class="service-list-view">
      <RsSplitPane
        class="service-list__split"
        orientation="vertical"
        :panes="splitPanes"
        disabled
      >
        <template #search>
          <div class="service-list__search">
            <RsSearchForm
              ref="serviceSearchFormRef"
              :module-id="service.model.moduleId"
              v-bind="service.model.searchFormConfig"
              @search="handleServiceSearch"
              @reset="handleServiceReset"
              @toolbar-click="(key) => handleServiceToolbarClick(key)"
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

    <div v-show="showDetailView" class="service-detail-view">
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
      :module-id="service.model.moduleId"
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
import ServiceDetail from './components/ServiceDetail.vue'
import { useServicePage } from './hooks'
import type { Service, ServiceNode } from './types'

defineOptions({
  name: 'ServiceList',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const serviceSearchFormRef = ref()
const serviceGridRef = ref<RsGridExpose | null>(null)

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
  handleServiceFormSubmit,
  handleToolbarClick: handleServiceToolbarClick,
  handleMenuClick: handleServiceMenuClickBase,
  handleSearch: handleServiceSearch,
  handleReset: handleServiceReset,
} = useServicePage(serviceGridRef, serviceSearchFormRef)

const handleServicePageChange = (params: { currentPage: number; pageSize: number }) => {
  service.handlePageChange(params.currentPage, params.pageSize)
}

const handleServiceMenuClick = async (params: { key: string; row?: Service }) => {
  if (params.key === 'view' && params.row) {
    await openServiceDetail(params.row)
    return
  }
  await handleServiceMenuClickBase(params)
}

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
  } catch {
    message.error('获取服务详情失败')
  } finally {
    detailLoading.value = false
  }
}

const handleDetailBack = () => {
  showDetailView.value = false
  currentDetailService.value = null
}

const handleDetailEdit = () => {
  if (!currentDetailService.value) return
  showDetailView.value = false
  handleServiceMenuClick({
    key: 'edit',
    row: currentDetailService.value,
  })
}

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
  } catch {
    message.error('刷新服务详情失败')
  } finally {
    detailLoading.value = false
  }
}

const handleClusterConfig = () => {
  message.info('集群配置功能开发中')
}

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

.service-list__split {
  width: 100%;
  height: 100%;
  min-height: 0;
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
