<template>
  <div class="service-center-instance-manager" :id="service.model.moduleId">
    <RsSplitPane
      class="service-center-instance-manager__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="service-center-instance-manager__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
            @toolbar-click="handleToolbarClick"
          />
        </div>
      </template>

      <template #grid>
        <div class="service-center-instance-manager__list">
          <RsLoading v-if="loading" block size="lg" />
          <RsEmpty
            v-else-if="!instanceList.length"
            description="暂无服务中心实例"
          >
            <template #icon>
              <GIcon :icon="ServerOutline" :size="32" color="var(--g-primary)" />
            </template>
          </RsEmpty>
          <div v-else class="service-center-instance-manager__cards">
            <ServiceCenterInstanceCard
              v-for="instance in instanceList"
              :key="cardKey(instance)"
              :instance="instance"
              :selected="isSelected(instance)"
              @select="selectInstance(instance)"
              @action="(key) => handleCardAction(key, instance)"
            />
          </div>
          <div
            v-if="totalCount > 0"
            class="service-center-instance-manager__pagination"
          >
            <RsPagination
              :page="currentPage"
              :page-size="pageSize"
              :total="totalCount"
              size="sm"
              :show-summary="true"
              @update:page="(p) => service.handlePageChange({ currentPage: p, pageSize })"
              @update:page-size="(s) => service.handlePageChange({ currentPage: 1, pageSize: s })"
            />
          </div>
        </div>
      </template>
    </RsSplitPane>

    <CenterRuntimeDrawer
      v-model:visible="runtimeDrawerVisible"
      :loading="runtimeLoading"
      :instance="runtimeInstance"
      :overview="runtimeOverview"
      :connections="runtimeConnections"
      :module-id="service.model.moduleId"
    />

    <CenterTokenDrawer
      v-model:visible="tokenDrawerVisible"
      :loading="tokenLoading"
      :issuing="tokenIssuing"
      :can-issue="canIssueToken"
      :instance="tokenInstance"
      :tokens="tokenList"
      :issued-token="issuedToken"
      :module-id="service.model.moduleId"
      @issue="handleIssueToken"
      @revoke="handleRevokeToken"
    />

    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :module-id="service.model.moduleId"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增实例' : formDialogMode === 'edit' ? '编辑实例' : '查看实例详情'"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.instanceFormConfig.fields"
      :form-tabs="service.model.instanceFormConfig.tabs"
      :initial-data="currentEditInstance || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="submitting"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { GIcon } from '@/components/gicon'
import { RsEmpty, RsLoading, RsPagination, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ServerOutline } from '@vicons/ionicons5'
import { getDefaultPageSize } from '@/utils/pagination'
import { computed, ref } from 'vue'
import { CenterRuntimeDrawer, CenterTokenDrawer, ServiceCenterInstanceCard } from './components'
import { canInstanceAction, useServiceCenterInstancePage } from './hooks'
import type { ServiceCenterInstance } from './types'

defineOptions({
  name: 'ServiceCenterInstanceManager',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditInstance,
  submitting,
  handleFormSubmit,
  handleToolbarClick,
  handleCardAction,
  handleSearch,
  selectedInstance,
  selectInstance,
  runtimeDrawerVisible,
  runtimeLoading,
  runtimeInstance,
  runtimeOverview,
  runtimeConnections,
  tokenDrawerVisible,
  tokenLoading,
  tokenIssuing,
  tokenInstance,
  tokenList,
  issuedToken,
  handleIssueToken,
  handleRevokeToken,
} = useServiceCenterInstancePage(searchFormRef)

const canIssueToken = computed(() => canInstanceAction('edit'))

const instanceList = computed(() => service.model.instanceList.value)
const loading = computed(() => service.model.loading.value)
const totalCount = computed(() => service.model.pageInfo.value?.totalCount || 0)
const currentPage = computed(() => service.model.pageInfo.value?.pageIndex || 1)
const pageSize = computed(() => service.model.pageInfo.value?.pageSize || getDefaultPageSize())

function cardKey(instance: ServiceCenterInstance) {
  return instance.oprSeqFlag || `${instance.tenantId}:${instance.instanceName}:${instance.environment}`
}

function isSelected(instance: ServiceCenterInstance) {
  const current = selectedInstance.value
  return Boolean(
    current
    && current.tenantId === instance.tenantId
    && current.instanceName === instance.instanceName
    && current.environment === instance.environment,
  )
}
</script>

<style lang="scss" scoped>
.service-center-instance-manager {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.service-center-instance-manager__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.service-center-instance-manager__search {
  width: 100%;
  box-sizing: border-box;
}

.service-center-instance-manager__list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 12px;
  gap: 12px;
  background: var(--g-bg-primary);
}

.service-center-instance-manager__list :deep(.rs-empty) {
  flex: 1;
}

.service-center-instance-manager__cards {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(312px, 1fr));
  align-content: start;
  gap: 18px;
  padding: 4px 6px 12px;
}

.service-center-instance-manager__pagination {
  flex: none;
  display: flex;
  justify-content: flex-end;
}
</style>
