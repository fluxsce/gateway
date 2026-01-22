<template>
  <div class="gateway-instance-manager" :id="service.model.moduleId">
    <GPane direction="vertical" default-size="80px">
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
          :data="service.model.instanceList"
          :loading="service.model.loading"
          v-bind="service.model.gridConfig"
          @page-change="service.handlePageChange"
          @menu-click="handleMenuClick"
        >
          <!-- TLS 状态自定义渲染 -->
          <template #tlsEnabled="{ row }">
            <n-tag :type="row.tlsEnabled === 'Y' ? 'success' : 'default'" size="small">
              {{ row.tlsEnabled === 'Y' ? '启用' : '禁用' }}
            </n-tag>
          </template>

          <!-- 健康状态自定义渲染 -->
          <template #healthStatus="{ row }">
            <n-tag :type="row.healthStatus === 'Y' ? 'success' : 'error'" size="small">
              <template #icon>
                <n-icon>
                  <CheckmarkCircleOutline v-if="row.healthStatus === 'Y'" />
                  <AlertCircleOutline v-else />
                </n-icon>
              </template>
              {{ row.healthStatus === 'Y' ? '在线' : '离线' }}
            </n-tag>
          </template>

          <!-- 活动状态自定义渲染 -->
          <template #activeFlag="{ row }">
            <n-tag :type="row.activeFlag === 'Y' ? 'success' : 'default'" size="small">
              {{ row.activeFlag === 'Y' ? '活动' : '非活动' }}
            </n-tag>
          </template>
        </g-grid>
      </template>
    </GPane>

    <!-- 实例对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="formDialogVisible"
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
    <GdataFormModal
      v-model:visible="logConfigDialogVisible"
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
  </div>
</template>

<script lang="ts" setup>
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import UserAgentAccessConfigListModal from '@/views/common/common002/agent-config/UserAgentAccessConfigListModal.vue'
import ApiAccessConfigListModal from '@/views/common/common002/api-config/ApiAccessConfigListModal.vue'
import AuthConfigFormModal from '@/views/common/common002/auth-config/AuthConfigFormModal.vue'
import CorsConfigFormModal from '@/views/common/common002/cors-config/CorsConfigFormModal.vue'
import DomainAccessConfigListModal from '@/views/common/common002/domain-config/DomainAccessConfigListModal.vue'
import IpAccessConfigListModal from '@/views/common/common002/ip-config/IpAccessConfigListModal.vue'
import RateLimitConfigFormModal from '@/views/common/common002/limit-config/RateLimitConfigFormModal.vue'
import { AlertCircleOutline, CheckmarkCircleOutline } from '@vicons/ionicons5'
import { NIcon, NTag } from 'naive-ui'
import { ref } from 'vue'
import { useGatewayInstancePage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'GatewayInstanceManager'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

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
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
} = useGatewayInstancePage(gridRef, searchFormRef)

// 数据由搜索表单的"查询"按钮触发加载
</script>

<style lang="scss" scoped>
.gateway-instance-manager {
  width: 100%;
  height: 100%;
  overflow: hidden;

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
