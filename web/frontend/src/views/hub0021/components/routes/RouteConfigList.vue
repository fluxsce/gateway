<template>
  <div class="route-config-list" id="hub0021-route-config-list">
    <GPane direction="vertical" :default-size="0.1" :min="0.1" :max="0.5">
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
          :data="service.model.routeList"
          :loading="service.model.loading"
          v-bind="service.model.gridConfig"
          @page-change="service.handlePageChange"
          @menu-click="handleMenuClick"
        >
          <!-- 路由名称自定义渲染 -->
          <template #routeName="{ row }">
            <span class="route-name-text">{{ row.routeName }}</span>
          </template>

          <!-- 匹配类型自定义渲染 -->
          <template #matchType="{ row }">
            <n-tag :type="getMatchTypeTagType(row.matchType)" size="small">
              {{ getMatchTypeLabel(row.matchType) }}
            </n-tag>
          </template>

          <!-- HTTP方法自定义渲染 -->
          <template #allowedMethods="{ row }">
            <div class="http-methods-container">
              <template v-if="getAllowedMethods(row.allowedMethods).length > 0">
                <span
                  v-for="(method, index) in getDisplayMethods(row.allowedMethods)"
                  :key="method"
                  class="http-method-tag"
                  :class="getMethodClass(method)"
                >
                  {{ method }}
                </span>
                <span v-if="getRemainingMethodsCount(row.allowedMethods) > 0" class="http-method-more">
                  +{{ getRemainingMethodsCount(row.allowedMethods) }}
                </span>
              </template>
              <span v-else class="http-method-empty">全部</span>
            </div>
          </template>

          <!-- WebSocket自定义渲染 -->
          <template #enableWebsocket="{ row }">
            <n-tag :type="row.enableWebsocket === 'Y' ? 'success' : 'default'" size="small">
              {{ row.enableWebsocket === 'Y' ? '支持' : '不支持' }}
            </n-tag>
          </template>

          <!-- 关联服务自定义渲染 -->
          <template #serviceName="{ row }">
            <div class="service-name-container">
              <template v-if="row.serviceName">
                <!-- 后端返回了服务名称，使用 tag 显示 -->
                <n-tag size="small" type="success">
                  {{ row.serviceName }}
                </n-tag>
              </template>
              <template v-else-if="row.serviceDefinitionId">
                <!-- 后端未返回服务名称，根据 serviceDefinitionId 显示 -->
                <template v-if="isMultipleServices(row.serviceDefinitionId)">
                  <!-- 多个服务：显示服务ID列表 -->
                  <n-tag
                    v-for="(serviceId, index) in getServiceIds(row.serviceDefinitionId)"
                    :key="serviceId"
                    size="small"
                    type="info"
                    style="margin-right: 4px; margin-bottom: 2px;"
                  >
                    {{ serviceId }}
                  </n-tag>
                  <n-tag size="small" type="default" style="margin-left: 4px;">
                    {{ getServiceIds(row.serviceDefinitionId).length }}个服务
                  </n-tag>
                </template>
                <template v-else>
                  <!-- 单个服务：使用 tag 显示服务ID -->
                  <n-tag size="small" type="info">
                    {{ row.serviceDefinitionId }}
                  </n-tag>
                </template>
              </template>
              <template v-else>
                <n-tag size="small" type="default">
                  未关联
                </n-tag>
              </template>
            </div>
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

    <!-- 路由配置对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增路由配置' : formDialogMode === 'edit' ? '编辑路由配置' : '查看路由配置详情'"
      to="#hub0021-route-config-list"
      :form-fields="routeFormConfig.fields"
      :form-tabs="routeFormConfig.tabs"
      :initial-data="getRouteFormInitialData() || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />

    <!-- 路由断言配置对话框 -->
    <AssertConfigListModal
      v-model:visible="assertConfigDialogVisible"
      :route-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
    />

    <!-- 路由级配置对话框 -->
    <IpAccessConfigListModal
      v-model:visible="ipAccessControlDialogVisible"
      module-id="hub0021:ipAccessControl"
      :security-config-id="currentRouteConfigId"
      :title="'IP访问控制配置'"
      :width="1200"
      :to="'#hub0021-route-config-list'"
    />

    <UserAgentAccessConfigListModal
      v-model:visible="userAgentAccessControlDialogVisible"
      module-id="hub0021:userAgentAccessControl"
      :security-config-id="currentRouteConfigId"
      :title="'User-Agent访问控制配置'"
      :width="1200"
      :to="'#hub0021-route-config-list'"
    />

    <ApiAccessConfigListModal
      v-model:visible="apiAccessControlDialogVisible"
      module-id="hub0021:apiAccessControl"
      :security-config-id="currentRouteConfigId"
      :title="'API访问控制配置'"
      :width="1200"
      :to="'#hub0021-route-config-list'"
    />

    <DomainAccessConfigListModal
      v-model:visible="domainAccessControlDialogVisible"
      module-id="hub0021:domainAccessControl"
      :security-config-id="currentRouteConfigId"
      :title="'域名访问控制配置'"
      :width="1200"
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

    <!-- 路由过滤器配置对话框 -->
    <FilterConfigListModal
      v-model:visible="filterConfigDialogVisible"
      module-id="hub0021:filters"
      :route-config-id="currentRouteConfigId"
      :to="'#hub0021-route-config-list'"
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
import { NTag } from 'naive-ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { AssertConfigListModal } from '../assert-config'
import { FilterConfigListModal } from '../filter-config'
import { useRouteConfigPage } from './hooks/page'
import { MatchType } from './types'

// 定义组件名称
defineOptions({
  name: 'RouteConfigList'
})

// ============= Props =============

interface Props {
  /** 网关实例ID */
  gatewayInstanceId?: string
}

const props = withDefaults(defineProps<Props>(), {
  gatewayInstanceId: undefined,
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditRoute,
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
  filterConfigDialogVisible,
} = useRouteConfigPage(props.gatewayInstanceId, searchFormRef, gridRef)

// ============= 监听器 =============

// 监听 gatewayInstanceId 变化，重新加载数据
const stopGatewayInstanceIdWatch = watch(
  () => props.gatewayInstanceId,
  (newId, oldId) => {
    if (newId && newId !== oldId) {
      // 当实例ID变化时，重新加载路由配置列表
      service.loadRouteList({ gatewayInstanceId: newId })
    } else if (!newId && oldId) {
      // 当实例ID被清空时，清空列表
      service.model.routeList.value = []
    }
  },
  { immediate: false }
)

// 组件卸载时清理监听器
onBeforeUnmount(() => {
  stopGatewayInstanceIdWatch()
})

// ============= 方法 =============

/**
 * 处理搜索（确保使用最新的 gatewayInstanceId）
 */
function handleSearch(formData?: Record<string, any>) {
  // 校验是否已选择实例
  if (!props.gatewayInstanceId) {
    return
  }
  // 使用最新的 props.gatewayInstanceId 来合并查询参数
  const searchParams = formData
    ? {
        ...formData,
        ...(props.gatewayInstanceId ? { gatewayInstanceId: props.gatewayInstanceId } : {}),
      }
    : props.gatewayInstanceId
      ? { gatewayInstanceId: props.gatewayInstanceId }
      : undefined
  // 调用 service 的 handleSearch，传入合并后的参数
  pageHandleSearch(searchParams)
}

/**
 * 获取匹配类型标签类型
 */
function getMatchTypeTagType(matchType: number): 'success' | 'info' | 'warning' | 'default' {
  const typeMap: Record<number, 'success' | 'info' | 'warning' | 'default'> = {
    [MatchType.EXACT]: 'success',
    [MatchType.PREFIX]: 'info',
    [MatchType.REGEX]: 'warning',
  }
  return typeMap[matchType] || 'default'
}

/**
 * 获取匹配类型标签
 */
function getMatchTypeLabel(matchType: number): string {
  const labelMap: Record<number, string> = {
    [MatchType.EXACT]: '精确匹配',
    [MatchType.PREFIX]: '前缀匹配',
    [MatchType.REGEX]: '正则匹配',
  }
  return labelMap[matchType] || '未知'
}

/**
 * 获取允许的HTTP方法数组
 */
function getAllowedMethods(allowedMethods?: string[] | string): string[] {
  if (!allowedMethods) {
    return []
  }
  if (Array.isArray(allowedMethods)) {
    return allowedMethods
  }
  if (typeof allowedMethods === 'string') {
    try {
      const parsed = JSON.parse(allowedMethods)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }
  return []
}

/**
 * 获取要显示的方法（最多显示2个，避免换行）
 */
function getDisplayMethods(allowedMethods?: string[] | string): string[] {
  const methods = getAllowedMethods(allowedMethods)
  return methods.slice(0, 2)
}

/**
 * 获取剩余方法数量
 */
function getRemainingMethodsCount(allowedMethods?: string[] | string): number {
  const methods = getAllowedMethods(allowedMethods)
  return Math.max(0, methods.length - 2)
}

/**
 * 获取HTTP方法的样式类
 */
function getMethodClass(method: string): string {
  const methodMap: Record<string, string> = {
    GET: 'method-get',
    POST: 'method-post',
    PUT: 'method-put',
    DELETE: 'method-delete',
    PATCH: 'method-patch',
    HEAD: 'method-head',
    OPTIONS: 'method-options',
  }
  return methodMap[method.toUpperCase()] || 'method-default'
}

/**
 * 判断是否为多个服务（根据 serviceDefinitionId 是否包含逗号）
 */
function isMultipleServices(serviceDefinitionId?: string): boolean {
  if (!serviceDefinitionId) return false
  return serviceDefinitionId.includes(',')
}

/**
 * 获取服务ID列表（从逗号分隔的字符串中解析）
 */
function getServiceIds(serviceDefinitionId?: string): string[] {
  if (!serviceDefinitionId) return []
  return serviceDefinitionId.split(',').map(id => id.trim()).filter(id => id)
}

// 暴露刷新方法供父组件调用
defineExpose({
  refresh: () => {
    service.loadRouteList()
  }
})
</script>

<style lang="scss" scoped>
.route-config-list {
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

/* 路由名称突出显示样式 */
.route-name-text {
  color: var(--g-primary, #7c3aed);
}


/* HTTP方法显示样式 */
.http-methods-container {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
}

.http-method-tag {
  display: inline-block;
  padding: 2px 6px;
  border-radius: var(--g-radius-sm);
  font-size: 11px;
  font-weight: 500;
  line-height: 1.4;
  white-space: nowrap;
  flex-shrink: 0;
  transition: all var(--g-transition-base) var(--g-transition-ease);
  
  /* 默认样式 */
  background-color: var(--g-bg-tertiary, #f5f5f5);
  color: var(--g-text-secondary, #666);
  border: 1px solid var(--g-border-primary, #e0e0e0);
  
  /* GET方法 - 蓝色系 */
  &.method-get {
    background-color: rgba(96, 165, 250, 0.1);
    color: var(--g-info, #60a5fa);
    border-color: rgba(96, 165, 250, 0.3);
  }
  
  /* POST方法 - 绿色系 */
  &.method-post {
    background-color: rgba(52, 211, 153, 0.1);
    color: var(--g-success, #34d399);
    border-color: rgba(52, 211, 153, 0.3);
  }
  
  /* PUT方法 - 橙色系 */
  &.method-put {
    background-color: rgba(251, 191, 36, 0.1);
    color: var(--g-warning, #fbbf24);
    border-color: rgba(251, 191, 36, 0.3);
  }
  
  /* DELETE方法 - 红色系 */
  &.method-delete {
    background-color: rgba(248, 113, 113, 0.1);
    color: var(--g-error, #f87171);
    border-color: rgba(248, 113, 113, 0.3);
  }
  
  /* PATCH方法 - 紫色系 */
  &.method-patch {
    background-color: rgba(129, 140, 248, 0.1);
    color: var(--g-primary, #818cf8);
    border-color: rgba(129, 140, 248, 0.3);
  }
  
  /* HEAD方法 - 灰色系 */
  &.method-head {
    background-color: var(--g-bg-tertiary, #f5f5f5);
    color: var(--g-text-tertiary, #999);
    border-color: var(--g-border-secondary, #d0d0d0);
  }
  
  /* OPTIONS方法 - 灰色系 */
  &.method-options {
    background-color: var(--g-bg-tertiary, #f5f5f5);
    color: var(--g-text-tertiary, #999);
    border-color: var(--g-border-secondary, #d0d0d0);
  }
  
  /* 默认方法 */
  &.method-default {
    background-color: var(--g-bg-tertiary, #f5f5f5);
    color: var(--g-text-secondary, #666);
    border-color: var(--g-border-primary, #e0e0e0);
  }
}

.http-method-more {
  display: inline-block;
  padding: 2px 6px;
  border-radius: var(--g-radius-sm);
  font-size: 11px;
  font-weight: 500;
  line-height: 1.4;
  white-space: nowrap;
  flex-shrink: 0;
  background-color: var(--g-bg-tertiary, #f5f5f5);
  color: var(--g-text-tertiary, #999);
  border: 1px solid var(--g-border-secondary, #d0d0d0);
  cursor: help;
}

.http-method-empty {
  display: inline-block;
  padding: 2px 6px;
  border-radius: var(--g-radius-sm);
  font-size: 11px;
  font-weight: 500;
  line-height: 1.4;
  white-space: nowrap;
  color: var(--g-text-tertiary, #999);
  font-style: italic;
}

/* 深色主题适配 */
[data-theme='dark'] {
  .http-method-tag {
    /* GET方法 */
    &.method-get {
      background-color: rgba(96, 165, 250, 0.15);
      color: var(--g-info, #60a5fa);
      border-color: rgba(96, 165, 250, 0.4);
    }
    
    /* POST方法 */
    &.method-post {
      background-color: rgba(52, 211, 153, 0.15);
      color: var(--g-success, #34d399);
      border-color: rgba(52, 211, 153, 0.4);
    }
    
    /* PUT方法 */
    &.method-put {
      background-color: rgba(251, 191, 36, 0.15);
      color: var(--g-warning, #fbbf24);
      border-color: rgba(251, 191, 36, 0.4);
    }
    
    /* DELETE方法 */
    &.method-delete {
      background-color: rgba(248, 113, 113, 0.15);
      color: var(--g-error, #f87171);
      border-color: rgba(248, 113, 113, 0.4);
    }
    
    /* PATCH方法 */
    &.method-patch {
      background-color: rgba(129, 140, 248, 0.15);
      color: var(--g-primary, #818cf8);
      border-color: rgba(129, 140, 248, 0.4);
    }
    
    /* HEAD和OPTIONS方法 */
    &.method-head,
    &.method-options {
      background-color: var(--g-bg-tertiary, #262626);
      color: var(--g-text-tertiary, #a3a3a3);
      border-color: var(--g-border-secondary, #525252);
    }
    
    /* 默认方法 */
    &.method-default {
      background-color: var(--g-bg-tertiary, #262626);
      color: var(--g-text-secondary, #d4d4d4);
      border-color: var(--g-border-primary, #404040);
    }
  }
  
  .http-method-more {
    background-color: var(--g-bg-tertiary, #262626);
    color: var(--g-text-tertiary, #a3a3a3);
    border-color: var(--g-border-secondary, #525252);
  }
  
  .http-method-empty {
    color: var(--g-text-tertiary, #a3a3a3);
  }
}

/* 关联服务显示样式 */
.service-name-container {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  justify-content: center;
}
</style>

