<template>
  <div class="gateway-instance-tree">
    <!-- 工具条：标题 + 数量 + 刷新（企业控制台侧栏标准头） -->
    <div class="instance-tree-toolbar">
      <div class="instance-tree-toolbar__left">
        <span class="instance-tree-toolbar__title">网关实例</span>
        <span
          v-if="page.model.totalCount.value > 0"
          class="instance-tree-toolbar__count"
        >
          {{ page.model.totalCount.value }}
        </span>
      </div>
      <!-- icon-only 必须用 icon prop；默认插槽只会进 tooltip，界面上看不见 -->
      <RsButton
        variant="text"
        size="sm"
        icon="refresh-cw"
        icon-only
        tooltip="刷新"
        :bordered="false"
        :disabled="page.model.loading.value"
        :loading="page.model.loading.value"
        @click="page.handleRefresh"
      />
    </div>

    <div class="instance-tree-filter">
      <RsInput
        v-model="page.model.filterKeyword.value"
        placeholder="搜索名称、地址或端口"
        clearable
        size="sm"
      >
        <template #prefix>
          <GIcon :icon="SearchOutline" size="sm" />
        </template>
      </RsInput>
    </div>

    <div class="instance-tree-body">
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
            selectable
            size="md"
            :item-height="60"
            height="100%"
            @node-click="(node, key) => page.handleNodeClick(node, key, emit)"
          >
            <template #title="{ node, selected }">
              <div
                class="tree-node"
                :class="{ 'tree-node--selected': selected }"
                @contextmenu="page.setContextNode(node)"
              >
                <div class="tree-node__icon" aria-hidden="true">
                  <GIcon :icon="ServerOutline" :size="16" />
                </div>
                <div class="tree-node__body">
                  <div class="tree-node__row">
                    <span class="tree-node__name" :title="instanceName(node)">
                      {{ instanceName(node) }}
                    </span>
                    <span
                      class="tree-node__health"
                      :class="isHealthy(node) ? 'is-ok' : 'is-bad'"
                    >
                      {{ isHealthy(node) ? '健康' : '异常' }}
                    </span>
                  </div>
                  <div class="tree-node__meta">
                    <span class="tree-node__addr" :title="instanceAddress(node)">
                      {{ instanceAddress(node) }}
                    </span>
                    <span v-if="isTls(node)" class="tree-node__proto">TLS</span>
                  </div>
                </div>
              </div>
            </template>
          </RsTree>
          <RsEmpty v-else description="暂无可用的网关实例" class="instance-tree-empty">
            <template #icon>
              <GIcon :icon="ServerOutline" :size="32" color="var(--g-primary)" />
            </template>
          </RsEmpty>
        </div>
      </RsContextMenu>
    </div>

    <div
      v-if="page.model.totalCount.value > 0"
      class="instance-pagination"
    >
      <RsPagination
        :page="page.model.currentPage.value"
        :page-size="page.model.pageSize.value"
        :total="page.model.totalCount.value"
        size="sm"
        :show-summary="false"
        @update:page="(p) => page.handlePageChange({ currentPage: p, pageSize: page.model.pageSize.value })"
        @update:page-size="(s) => page.handlePageChange({ currentPage: 1, pageSize: s })"
      />
    </div>

    <RsDataFormModal
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
import { RsDataFormModal } from '@/components/form/rs-data'
import { GIcon } from '@/components/gicon'
import {
  RsButton,
  RsContextMenu,
  RsEmpty,
  RsInput,
  RsLoading,
  RsPagination,
  RsTree,
  type RsTreeNode,
} from '@/ui'
import {
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
  /** 父容器模块ID，用于 RsDataFormModal 的 :to 属性 */
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

/** 实例显示名称 */
function instanceName(node: RsTreeNode): string {
  return asInstanceNode(node).instance?.instanceName || '未命名'
}

/** 绑定地址与端口（TLS 时用 httpsPort） */
function instanceAddress(node: RsTreeNode): string {
  const instance = asInstanceNode(node).instance
  if (!instance) return '-'
  const port = instance.tlsEnabled === 'Y' ? instance.httpsPort : instance.httpPort
  return `${instance.bindAddress || '-'}:${port || '-'}`
}

function isHealthy(node: RsTreeNode): boolean {
  return asInstanceNode(node).instance?.healthStatus === 'Y'
}

function isTls(node: RsTreeNode): boolean {
  return asInstanceNode(node).instance?.tlsEnabled === 'Y'
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
  min-height: 0;
  overflow: hidden;
  background: var(--rs-surface, var(--g-bg-primary, #fff));
  color: var(--rs-text, var(--g-text-primary));
}

/* ── 顶栏工具条 ── */
.instance-tree-toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  height: var(--g-toolbar-height, 36px);
  padding: 0 12px;
  border-bottom: 1px solid var(--g-border-color, var(--rs-border-subtle, #e5e7eb));
  background: var(--rs-surface, var(--g-bg-primary, #fff));
  box-sizing: border-box;
}

.instance-tree-toolbar__left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.instance-tree-toolbar__title {
  font-size: var(--rs-font-size-sm, 13px);
  font-weight: var(--rs-font-weight-medium, 500);
  line-height: 1;
  color: var(--rs-text, var(--g-text-primary));
}

.instance-tree-toolbar__count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 6px;
  border-radius: 9px;
  font-size: 11px;
  font-weight: var(--rs-font-weight-medium, 500);
  line-height: 1;
  color: var(--rs-text-secondary, var(--g-text-secondary));
  background: var(--rs-surface-hover, var(--g-bg-tertiary, #f3f4f6));
  font-variant-numeric: tabular-nums;
}

/* ── 搜索区：与面板同色，避免灰底+输入圆角在顶角露出白缝 ── */
.instance-tree-filter {
  flex-shrink: 0;
  padding: 8px 12px;
  border-bottom: 1px solid var(--g-border-color, var(--rs-border-subtle, #e5e7eb));
  background: var(--rs-surface, var(--g-bg-primary, #fff));
}

/* ── 列表区 ── */
.instance-tree-body {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--rs-surface, var(--g-bg-primary, #fff));
}

.instance-tree-body > :deep(*) {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.instance-tree-container {
  position: relative;
  flex: 1 1 auto;
  width: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.instance-tree {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  border-radius: 0;
}

/* 去掉 RsTree 默认圆角/边距，避免选中底与直角边框形成白角割裂 */
.instance-tree :deep(.rs-tree),
.instance-tree :deep(.rs-tree__viewport),
.instance-tree :deep(.rs-tree__viewport--scroll) {
  border-radius: 0;
}

.instance-tree :deep(.rs-tree__spacer),
.instance-tree :deep(.rs-tree__node-icon) {
  display: none;
}

.instance-tree :deep(.rs-tree__list) {
  padding: 0;
}

.instance-tree :deep(.rs-tree__row) {
  align-items: stretch;
  gap: 0;
  margin: 0 !important;
  padding: 0;
  border-radius: 0 !important;
  box-shadow: none !important;
  border-bottom: 1px solid var(--g-border-color, var(--rs-border-subtle, #eef0f3));
}

.instance-tree :deep(.rs-tree__row:hover:not(.rs-tree__row--disabled):not(.rs-tree__row--selected)) {
  background: var(--rs-item-hover, var(--g-hover-overlay, rgba(0, 0, 0, 0.04)));
}

.instance-tree :deep(.rs-tree__row--selected),
.instance-tree :deep(.rs-tree__row--selected.rs-tree__row--focused) {
  background: color-mix(in srgb, var(--g-primary, var(--rs-primary)) 8%, transparent);
  box-shadow: none;
}

.instance-tree :deep(.rs-tree__row--focused:not(.rs-tree__row--selected)) {
  background: var(--rs-item-hover, var(--g-hover-overlay, rgba(0, 0, 0, 0.04)));
  box-shadow: none;
}

.instance-tree :deep(.rs-tree__label--block) {
  display: block;
  width: 100%;
  min-width: 0;
  padding: 0;
  box-sizing: border-box;
}

/* ── 节点：图标井 + 信息 + 状态文案 ── */
.tree-node {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  width: 100%;
  height: 60px;
  padding: 0 12px 0 14px;
  box-sizing: border-box;
}

.tree-node--selected::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: var(--g-primary, var(--rs-primary));
}

.tree-node__icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--rs-radius-xs, 4px);
  color: var(--rs-text-secondary, var(--g-text-secondary));
  background: var(--rs-surface-hover, var(--g-bg-tertiary, #f3f4f6));
  border: 1px solid var(--g-border-color, var(--rs-border-subtle, #e5e7eb));
}

.tree-node--selected .tree-node__icon {
  color: var(--g-primary, var(--rs-primary));
  background: color-mix(in srgb, var(--g-primary, var(--rs-primary)) 12%, transparent);
  border-color: color-mix(in srgb, var(--g-primary, var(--rs-primary)) 28%, transparent);
}

.tree-node__body {
  flex: 1 1 auto;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.tree-node__row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.tree-node__name {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--rs-font-size-sm, 13px);
  font-weight: var(--rs-font-weight-medium, 500);
  line-height: 1.3;
  color: var(--rs-text, var(--g-text-primary));
}

.tree-node--selected .tree-node__name {
  color: var(--g-primary, var(--rs-primary));
}

.tree-node__health {
  flex-shrink: 0;
  font-size: 11px;
  font-weight: var(--rs-font-weight-medium, 500);
  line-height: 1;
  letter-spacing: 0.02em;

  &.is-ok {
    color: var(--rs-success, var(--g-success, #16a34a));
  }

  &.is-bad {
    color: var(--rs-warning, var(--g-warning, #d97706));
  }
}

.tree-node__meta {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.tree-node__addr {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  line-height: 1.3;
  color: var(--rs-text-tertiary, var(--g-text-tertiary, #8b93a1));
  font-family: var(--rs-font-mono, var(--g-font-family-mono, ui-monospace, monospace));
  font-variant-numeric: tabular-nums;
}

.tree-node__proto {
  flex-shrink: 0;
  padding: 0 4px;
  border-radius: 2px;
  font-size: 10px;
  font-weight: var(--rs-font-weight-medium, 500);
  line-height: 16px;
  color: var(--rs-text-secondary, var(--g-text-secondary));
  background: var(--rs-surface-hover, var(--g-bg-tertiary, #f3f4f6));
  border: 1px solid var(--g-border-color, var(--rs-border-subtle, #e5e7eb));
}

.instance-tree-empty {
  padding: 40px 0;
}

.instance-pagination {
  flex: 0 0 auto;
  box-sizing: border-box;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: var(--g-footer-height);
  min-height: var(--g-footer-height);
  padding: 0 8px;
  border-top: 1px solid var(--g-border-color, var(--rs-border-subtle, #e5e7eb));
  background: var(--rs-surface, var(--g-bg-primary, #fff));
  overflow: hidden;
}
</style>
