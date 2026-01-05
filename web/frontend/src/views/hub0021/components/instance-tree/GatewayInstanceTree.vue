<template>
  <div class="gateway-instance-tree">
    <!-- 上部分：标题、刷新、过滤 -->
    <div class="instance-tree-header">
      <div class="instance-header">
        <n-icon size="20" color="var(--g-primary)">
          <ServerOutline />
        </n-icon>
        <span>网关实例列表</span>
        <div class="flex-spacer"></div>
        <n-button 
          text 
          size="small" 
          @click="page.handleRefresh" 
          :disabled="page.model.loading.value"
          :loading="page.model.loading.value"
        >
          <template #icon>
            <n-icon :component="RefreshOutline" />
          </template>
        </n-button>
      </div>
      <!-- 搜索过滤框 -->
      <div class="instance-filter">
        <n-input
          v-model:value="page.model.filterKeyword.value"
          placeholder="搜索实例名称、地址或端口"
          clearable
          size="small"
        >
          <template #prefix>
            <n-icon :component="SearchOutline" />
          </template>
        </n-input>
      </div>
    </div>

    <!-- 中部：树区域（允许滚动） -->
    <div class="instance-tree-container">
      <n-spin :show="page.model.loading.value">
        <GTree
          v-if="page.model.treeData.value.length > 0"
          :data="page.model.treeData.value"
          :show-icon="true"
          :block-line="true"
          :node-height="32"
          :show-line="true"
          :ellipsis="true"
          :ellipsis-line-clamp="1"
          :ellipsis-tooltip="true"
          :module-id="page.model.moduleId"
          :menu-config="page.model.treeMenuConfig"
          :render-prefix="page.renderNodePrefix"
          :render-label="page.renderNodeLabel"
          :render-suffix="page.renderNodeSuffix"
          @select="(keys, option) => page.handleTreeSelect(keys, option, emit)"
          @menu-click="page.handleMenuClick"
        />
        <n-empty v-else description="暂无可用的网关实例" style="padding: 40px 0;">
          <template #icon>
            <n-icon size="40" color="var(--g-primary)">
              <ServerOutline />
            </n-icon>
          </template>
        </n-empty>
      </n-spin>
    </div>

    <!-- 底部：分页 -->
    <div class="instance-pagination" v-if="page.model.totalCount.value > 0">
      <GPagination
        :current-page="page.model.currentPage.value"
        :page-size="page.model.pageSize.value"
        :total="page.model.totalCount.value"
        :layouts="['PrevPage', 'NextPage']"
        align="center"
        @page-change="page.handlePageChange"
      />
    </div>

    <!-- Router配置对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="page.routerFormDialogVisible.value"
      :mode="page.routerFormDialogMode.value"
      :title="page.routerFormDialogMode.value === 'create' ? '新增Router配置' : page.routerFormDialogMode.value === 'edit' ? '编辑Router配置' : '查看Router配置详情'"
      :to="`#${props.parentModuleId}`"
      :form-fields="page.model.routerFormConfig.fields"
      :form-tabs="page.model.routerFormConfig.tabs"
      :initial-data="page.getRouterFormInitialData()"
      :auto-close-on-confirm="false"
      :confirm-loading="page.routerSubmitting.value"
      @submit="page.handleRouterFormSubmit"
    />

    <!-- 全局过滤器配置对话框 -->
    <FilterConfigListModal
      v-model:visible="page.globalFilterConfigDialogVisible.value"
      :gateway-instance-id="page.currentGatewayInstanceId.value"
      :filter-scope="'global'"
      :to="`#${props.parentModuleId}`"
    />
  </div>
</template>

<script setup lang="ts">
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import { GPagination } from '@/components/gpage'
import { GTree } from '@/components/gtree'
import {
  RefreshOutline,
  SearchOutline,
  ServerOutline
} from '@vicons/ionicons5'
import {
  NButton,
  NEmpty,
  NIcon,
  NInput,
  NSpin
} from 'naive-ui'
import { onMounted } from 'vue'
import FilterConfigListModal from '../filter-config/FilterConfigListModal.vue'
import { useGatewayInstanceTreePage } from './hooks/page'
import type { GatewayInstance } from './types'

// 定义组件名称
defineOptions({
  name: 'GatewayInstanceTree'
})

// ============= Props =============

const props = withDefaults(defineProps<{
  /** 父容器模块ID，用于 GdataFormModal 的 :to 属性 */
  parentModuleId?: string
}>(), {
  parentModuleId: 'route-management'
})

// ============= Emits =============

const emit = defineEmits<{
  /** 实例选择变化 */
  (e: 'select', instanceId: string, instance: GatewayInstance): void
}>()

// ============= 使用 Page Hook =============

const page = useGatewayInstanceTreePage()

// ============= 生命周期 =============

// 组件挂载时加载数据
onMounted(() => {
  page.service.loadGatewayInstances()
})

// ============= 暴露方法 =============

defineExpose({
  /** 刷新实例列表 */
  refresh: page.service.loadGatewayInstances,
  /** 重置分页到第一页 */
  resetPage: page.model.resetPage,
  /** 清空搜索关键词 */
  clearFilter: page.model.clearFilter
})
</script>

<style lang="scss" scoped>
.gateway-instance-tree {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

/* 上部分：标题、刷新、过滤 */
.instance-tree-header {
  flex-shrink: 0;
  padding: 12px;
  border-bottom: 1px solid var(--n-border-color);
}

.instance-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  margin-bottom: 12px;
}

.flex-spacer {
  flex: 1;
}

.instance-filter {
  margin-top: 8px;
}

/* 中部：树区域（允许滚动） */
.instance-tree-container {
  flex: 1;
  overflow: auto;
  min-height: 0;
  padding: 8px 12px;
}

:deep(.tree-node-label) {
  &:hover {
    background-color: var(--n-color-hover);
  }
}

/* 底部：分页 */
.instance-pagination {
  flex-shrink: 0;
  border-top: 1px solid var(--n-border-color);
}
</style>

