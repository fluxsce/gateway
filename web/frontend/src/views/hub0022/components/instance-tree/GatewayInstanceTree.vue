<template>
  <div class="gateway-instance-tree">
    <!-- 上部分：标题、刷新、过滤 -->
    <div class="instance-tree-header">
      <div class="instance-header">
        <GIcon :icon="ServerOutline" :size="20" color="var(--g-primary)" />
        <span>网关实例列表</span>
        <div class="flex-spacer"></div>
        <RsButton
          variant="text"
          size="sm"
          icon-only
          :bordered="false"
          :disabled="page.model.loading.value"
          :loading="page.model.loading.value"
          @click="page.handleRefresh"
        >
          <GIcon :icon="RefreshOutline" size="sm" />
        </RsButton>
      </div>
      <div class="instance-filter">
        <RsInput
          v-model="page.model.filterKeyword.value"
          placeholder="搜索实例名称、地址或端口"
          clearable
          size="sm"
        >
          <template #prefix>
            <GIcon :icon="SearchOutline" size="sm" />
          </template>
        </RsInput>
      </div>
    </div>

    <!-- 中部：树区域 -->
    <RsContextMenu
      :items="page.model.contextMenuItems"
      @select="page.handleContextMenuSelect"
    >
      <div class="instance-tree-container">
        <RsLoading :loading="page.model.loading.value" overlay block size="md" />
        <RsTree
          v-if="page.model.treeData.value.length > 0"
          v-model="selectedKeys"
          class="instance-tree"
          :nodes="page.model.treeData.value"
          block-node
          show-line
          selectable
          height="100%"
          @node-click="(node, key) => page.handleNodeClick(node, key, emit)"
        >
          <template #title="{ node, label }">
            <span
              class="tree-node-title"
              @contextmenu="page.setContextNode(node)"
            >
              <span class="tree-node-label">{{ label }}</span>
              <RsTag
                v-if="asInstanceNode(node).instance"
                :variant="asInstanceNode(node).instance!.healthStatus === 'Y' ? 'success' : 'warning'"
                size="sm"
                class="tree-node-tag"
              >
                {{ asInstanceNode(node).instance!.healthStatus === 'Y' ? '健康' : '异常' }}
              </RsTag>
            </span>
          </template>
        </RsTree>
        <RsEmpty v-else description="暂无可用的网关实例" style="padding: 40px 0;">
          <template #icon>
            <GIcon :icon="ServerOutline" :size="40" color="var(--g-primary)" />
          </template>
        </RsEmpty>
      </div>
    </RsContextMenu>

    <!-- 底部：分页 -->
    <div class="instance-pagination" v-if="page.model.totalCount.value > 0">
      <RsPagination
        :page="page.model.currentPage.value"
        :page-size="page.model.pageSize.value"
        :total="page.model.totalCount.value"
        size="sm"
        @update:page="(p) => page.handlePageChange({ currentPage: p, pageSize: page.model.pageSize.value })"
        @update:page-size="(s) => page.handlePageChange({ currentPage: 1, pageSize: s })"
      />
    </div>

    <!-- 代理配置对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="page.proxyFormDialogVisible.value"
      :mode="page.proxyFormDialogMode.value"
      :title="page.proxyFormDialogMode.value === 'create' ? '新增代理配置' : page.proxyFormDialogMode.value === 'edit' ? '编辑代理配置' : '查看代理配置详情'"
      :to="`#${props.parentModuleId}`"
      :form-fields="page.model.proxyFormConfig.fields"
      :form-tabs="page.model.proxyFormConfig.tabs"
      :initial-data="page.getProxyFormInitialData()"
      :auto-close-on-confirm="false"
      :confirm-loading="page.proxySubmitting.value"
      @submit="page.handleProxyFormSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import { GIcon } from '@/components/gicon'
import {
  RsButton,
  RsContextMenu,
  RsEmpty,
  RsInput,
  RsLoading,
  RsPagination,
  RsTag,
  RsTree,
  type RsTreeNode,
} from '@/ui'
import {
  RefreshOutline,
  SearchOutline,
  ServerOutline,
} from '@vicons/ionicons5'
import { onMounted, ref } from 'vue'
import { useGatewayInstanceTreePage } from './hooks/page'
import type { GatewayInstance, InstanceTreeNode } from './types'

defineOptions({
  name: 'GatewayInstanceTree',
})

const props = withDefaults(defineProps<{
  /** 父容器模块ID，用于 GdataFormModal 的 :to 属性 */
  parentModuleId?: string
}>(), {
  parentModuleId: 'proxy-management',
})

const emit = defineEmits<{
  (e: 'select', instanceId: string, instance: GatewayInstance): void
}>()

const page = useGatewayInstanceTreePage()
const selectedKeys = ref<string | string[]>('')

/**
 * 将通用 RsTreeNode 收窄为带实例信息的节点。
 */
function asInstanceNode(node: RsTreeNode): InstanceTreeNode {
  return node as InstanceTreeNode
}

onMounted(() => {
  page.service.loadGatewayInstances()
})

defineExpose({
  refresh: page.service.loadGatewayInstances,
  resetPage: page.model.resetPage,
  clearFilter: page.model.clearFilter,
})
</script>

<style lang="scss" scoped>
.gateway-instance-tree {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.instance-tree-header {
  flex-shrink: 0;
  padding: 12px;
  border-bottom: 1px solid var(--g-border-color, var(--rs-border, #e5e7eb));
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

.instance-tree-container {
  position: relative;
  flex: 1;
  overflow: hidden;
  min-height: 0;
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
}

.instance-tree {
  flex: 1;
  min-height: 0;
}

.tree-node-title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  max-width: 100%;
}

.tree-node-label {
  display: inline-block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tree-node-tag {
  flex-shrink: 0;
}

.instance-pagination {
  flex-shrink: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: var(--rs-control-height-md);
  padding: 0 var(--rs-space-sm, 8px);
  border-top: 1px solid var(--g-border-color, var(--rs-border, #e5e7eb));
  box-sizing: border-box;
}
</style>
