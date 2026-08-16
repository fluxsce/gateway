<template>
  <div class="route-config-list" id="hub0021-route-config-list">
    <RsSplitPane
      class="route-config-list__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="route-config-list__search">
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
        <div class="route-config-list__grid">
          <RsGrid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.routeList"
            :loading="service.model.loading"
            :columns="service.model.gridConfig.columns"
            :selectable="service.model.gridConfig.selectable"
            :row-key="service.model.gridConfig.rowKey"
            height="100%"
            :pagination-config="service.model.gridConfig.paginationConfig"
            :menu-config="service.model.gridConfig.menuConfig"
            @page-change="service.handlePageChange"
            @menu-click="handleMenuClick"
          />
        </div>
      </template>
    </RsSplitPane>

    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="
        formDialogMode === 'create'
          ? '新增路由配置'
          : formDialogMode === 'edit'
            ? '编辑路由配置'
            : '查看路由配置详情'
      "
      to="#hub0021-route-config-list"
      :form-fields="routeFormConfig.fields"
      :form-tabs="routeFormConfig.tabs"
      :initial-data="getRouteFormInitialData()"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />

    <AssertConfigListModal
      v-model:visible="assertConfigDialogVisible"
      :route-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
    />

    <IpAccessConfigListModal
      v-model:visible="ipAccessControlDialogVisible"
      module-id="hub0021:ipAccessControl"
      :security-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
    />

    <UserAgentAccessConfigListModal
      v-model:visible="userAgentAccessControlDialogVisible"
      module-id="hub0021:userAgentAccessControl"
      :security-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
    />

    <ApiAccessConfigListModal
      v-model:visible="apiAccessControlDialogVisible"
      module-id="hub0021:apiAccessControl"
      :security-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
    />

    <DomainAccessConfigListModal
      v-model:visible="domainAccessControlDialogVisible"
      module-id="hub0021:domainAccessControl"
      :security-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
    />

    <CorsConfigFormModal
      v-model:visible="corsConfigDialogVisible"
      module-id="hub0021:corsConfig"
      :route-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
    />

    <AuthConfigFormModal
      v-model:visible="authConfigDialogVisible"
      module-id="hub0021:authConfig"
      :route-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
    />

    <RateLimitConfigFormModal
      v-model:visible="rateLimitConfigDialogVisible"
      module-id="hub0021:rateLimitConfig"
      :route-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
    />

    <StaticHostConfigFormModal
      v-model:visible="staticHostConfigDialogVisible"
      :route-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
      @saved="service.loadRouteList()"
    />

    <FilterConfigListModal
      v-model:visible="filterConfigDialogVisible"
      module-id="hub0021:filters"
      :route-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import UserAgentAccessConfigListModal from '@/views/common/common002/agent-config/UserAgentAccessConfigListModal.vue'
import ApiAccessConfigListModal from '@/views/common/common002/api-config/ApiAccessConfigListModal.vue'
import AuthConfigFormModal from '@/views/common/common002/auth-config/AuthConfigFormModal.vue'
import CorsConfigFormModal from '@/views/common/common002/cors-config/CorsConfigFormModal.vue'
import DomainAccessConfigListModal from '@/views/common/common002/domain-config/DomainAccessConfigListModal.vue'
import IpAccessConfigListModal from '@/views/common/common002/ip-config/IpAccessConfigListModal.vue'
import RateLimitConfigFormModal from '@/views/common/common002/limit-config/RateLimitConfigFormModal.vue'
import { onBeforeUnmount, ref, watch } from 'vue'
import { AssertConfigListModal } from '../assert-config'
import { FilterConfigListModal } from '../filter-config'
import { StaticHostConfigFormModal } from '../static-host'
import { useRouteConfigPage } from './hooks/page'

defineOptions({
  name: 'RouteConfigList',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

interface Props {
  /** 网关实例ID */
  gatewayInstanceId?: string
}

const props = withDefaults(defineProps<Props>(), {
  gatewayInstanceId: undefined,
})

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)

const {
  service,
  formDialogVisible,
  formDialogMode,
  routeFormConfig,
  handleFormSubmit,
  getRouteFormInitialData,
  handleToolbarClick,
  handleMenuClick,
  handleSearch: pageHandleSearch,
  currentRouteConfigId,
  assertConfigDialogVisible,
  ipAccessControlDialogVisible,
  userAgentAccessControlDialogVisible,
  apiAccessControlDialogVisible,
  domainAccessControlDialogVisible,
  corsConfigDialogVisible,
  authConfigDialogVisible,
  rateLimitConfigDialogVisible,
  staticHostConfigDialogVisible,
  filterConfigDialogVisible,
} = useRouteConfigPage(props.gatewayInstanceId, searchFormRef, gridRef)

const stopGatewayInstanceIdWatch = watch(
  () => props.gatewayInstanceId,
  (newId, oldId) => {
    if (newId && newId !== oldId) {
      service.loadRouteList({ gatewayInstanceId: newId })
    } else if (!newId && oldId) {
      service.model.routeList.value = []
    }
  },
  { immediate: false },
)

onBeforeUnmount(() => {
  stopGatewayInstanceIdWatch()
})

function handleSearch(formData?: Record<string, any>) {
  if (!props.gatewayInstanceId) {
    return
  }
  const searchParams = formData
    ? {
        ...formData,
        ...(props.gatewayInstanceId ? { gatewayInstanceId: props.gatewayInstanceId } : {}),
      }
    : props.gatewayInstanceId
      ? { gatewayInstanceId: props.gatewayInstanceId }
      : undefined
  pageHandleSearch(searchParams)
}

defineExpose({
  refresh: () => {
    service.loadRouteList()
  },
})
</script>

<style lang="scss" scoped>
.route-config-list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.route-config-list__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.route-config-list__search {
  width: 100%;
  box-sizing: border-box;
}

.route-config-list__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
