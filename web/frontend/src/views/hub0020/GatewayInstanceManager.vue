<template>
  <div class="gateway-instance-manager" :id="service.model.moduleId">
    <RsSplitPane
      class="gateway-instance-manager__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <!-- 上部：搜索表单（auto 随内容） -->
      <template #search>
        <div class="gateway-instance-manager__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
            @toolbar-click="handleToolbarClick"
          />
        </div>
      </template>

      <!-- 下部：表格占满剩余高度 -->
      <template #grid>
        <div class="gateway-instance-manager__grid">
          <RsGrid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.instanceList"
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

    <!-- 实例对话框（新增/编辑/查看共用） -->
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

    <!-- 日志配置对话框 -->
    <RsDataFormModal
      v-model:visible="logConfigDialogVisible"
      :module-id="service.model.moduleId"
      :mode="logConfigDialogMode"
      :title="logConfigDialogMode === 'edit' ? '编辑日志配置' : '查看日志配置'"
      :to="`#${service.model.moduleId}`"
      :form-tabs="service.model.logConfigFormConfig.tabs"
      :form-fields="service.model.logConfigFormConfig.fields"
      :initial-data="currentLogConfig || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="logConfigSubmitting"
      @submit="handleLogConfigSubmit"
    />

    <!-- IP访问控制配置对话框 -->
    <IpAccessConfigListModal
      v-model:visible="ipAccessControlDialogVisible"
      module-id="hub0020:ipAccessControl"
      :security-config-id="ipAccessControlSecurityConfigId"
      :title="'IP访问控制配置'"
      :width="1200"
      :to="`#${service.model.moduleId}`"
    />

    <!-- User-Agent访问控制配置对话框 -->
    <UserAgentAccessConfigListModal
      v-model:visible="userAgentAccessControlDialogVisible"
      module-id="hub0020:userAgentAccessControl"
      :security-config-id="userAgentAccessControlSecurityConfigId"
      :title="'User-Agent访问控制配置'"
      :width="1200"
      :to="`#${service.model.moduleId}`"
    />

    <!-- API访问控制配置对话框 -->
    <ApiAccessConfigListModal
      v-model:visible="apiAccessControlDialogVisible"
      module-id="hub0020:apiAccessControl"
      :security-config-id="apiAccessControlSecurityConfigId"
      :title="'API访问控制配置'"
      :width="1200"
      :to="`#${service.model.moduleId}`"
    />

    <!-- 域名访问控制配置对话框 -->
    <DomainAccessConfigListModal
      v-model:visible="domainAccessControlDialogVisible"
      module-id="hub0020:domainAccessControl"
      :security-config-id="domainAccessControlSecurityConfigId"
      :title="'域名访问控制配置'"
      :width="1200"
      :to="`#${service.model.moduleId}`"
    />

    <!-- 跨域配置对话框 -->
    <CorsConfigFormModal
      v-model:visible="corsConfigDialogVisible"
      :gateway-instance-id="corsConfigGatewayInstanceId"
      :width="800"
      :to="`#${service.model.moduleId}`"
      module-id="hub0020:corsConfig"
    />

    <!-- 认证配置对话框 -->
    <AuthConfigFormModal
      v-model:visible="authConfigDialogVisible"
      :gateway-instance-id="authConfigGatewayInstanceId"
      :width="800"
      :to="`#${service.model.moduleId}`"
      module-id="hub0020:authConfig"
    />

    <!-- 限流配置对话框 -->
    <RateLimitConfigFormModal
      v-model:visible="rateLimitConfigDialogVisible"
      :gateway-instance-id="rateLimitConfigGatewayInstanceId"
      :width="800"
      :to="`#${service.model.moduleId}`"
      module-id="hub0020:rateLimitConfig"
    />

    <!-- 导出：只传参数，visible=true 时自动开始导出 -->
    <GExport
      v-model:visible="exportVisible"
      :module-id="service.model.moduleId"
      :url="exportUrl"
      :params="exportParams"
      :filename="exportFilename"
      dialog-title="导出网关实例配置"
    />

    <!-- 导入 -->
    <GImport
      v-model:visible="importVisible"
      :module-id="service.model.moduleId"
      :url="importUrl"
      @success="handleImportSuccess"
      @error="(e) => message.error(e.message)"
    />
  </div>
</template>

<script lang="ts" setup>
import { moduleApiPrefix, requestPathHelper } from '@/api/requestPath'
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { GExport, GImport } from '@/components/gexport-import'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import UserAgentAccessConfigListModal from '@/views/common/common002/agent-config/UserAgentAccessConfigListModal.vue'
import ApiAccessConfigListModal from '@/views/common/common002/api-config/ApiAccessConfigListModal.vue'
import AuthConfigFormModal from '@/views/common/common002/auth-config/AuthConfigFormModal.vue'
import CorsConfigFormModal from '@/views/common/common002/cors-config/CorsConfigFormModal.vue'
import DomainAccessConfigListModal from '@/views/common/common002/domain-config/DomainAccessConfigListModal.vue'
import IpAccessConfigListModal from '@/views/common/common002/ip-config/IpAccessConfigListModal.vue'
import RateLimitConfigFormModal from '@/views/common/common002/limit-config/RateLimitConfigFormModal.vue'
import { computed, ref } from 'vue'
import { useGatewayInstancePage } from './hooks'

defineOptions({
  name: 'GatewayInstanceManager',
})

/** 上方面板内容自适应，下方吃满剩余空间；disabled 禁止拖拽 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)
const message = useAppMessage()

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditInstance,
  submitting,
  handleFormSubmit,
  logConfigDialogVisible,
  logConfigDialogMode,
  currentLogConfig,
  logConfigSubmitting,
  handleLogConfigSubmit,
  ipAccessControlDialogVisible,
  ipAccessControlSecurityConfigId,
  userAgentAccessControlDialogVisible,
  userAgentAccessControlSecurityConfigId,
  apiAccessControlDialogVisible,
  apiAccessControlSecurityConfigId,
  domainAccessControlDialogVisible,
  domainAccessControlSecurityConfigId,
  corsConfigDialogVisible,
  corsConfigGatewayInstanceId,
  authConfigDialogVisible,
  authConfigGatewayInstanceId,
  rateLimitConfigDialogVisible,
  rateLimitConfigGatewayInstanceId,
  exportVisible,
  exportInstanceId,
  importVisible,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useGatewayInstancePage(gridRef, searchFormRef)

const exportParams = computed(() => ({ gatewayInstanceId: exportInstanceId.value }))
const exportFilename = computed(() => `网关实例配置_${exportInstanceId.value}`)
const exportUrl = requestPathHelper.join(moduleApiPrefix('hub0020'), 'exportGatewayInstance')
const importUrl = requestPathHelper.join(moduleApiPrefix('hub0020'), 'importGatewayInstance')

const handleImportSuccess = () => {
  message.success('导入成功，数据已刷新')
  handleSearch()
}
</script>

<style scoped>
.gateway-instance-manager {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.gateway-instance-manager__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.gateway-instance-manager__search {
  width: 100%;
}

.gateway-instance-manager__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
