<template>
  <div class="service-list" :id="service.model.moduleId">
    <!-- 列表视图 -->
    <div v-if="!showDetailView" class="service-list-view">
      <GPane direction="vertical" default-size="300px">
        <!-- 上部：命名空间列表 -->
        <template #1>
          <GCard>
            <NamespaceList
              ref="namespaceListRef"
              moduleId="hub0042:namespace"
              :show-dialog="true"
              :auto-load="true"
              @row-click="handleNamespaceRowClick"
              @namespace-select="handleNamespaceSelect"
            />
          </GCard>
        </template>

        <!-- 下部：服务列表 -->
        <template #2>
          <GCard>
            <GPane direction="vertical" default-size="80px">
              <!-- 服务搜索表单 -->
              <template #1>
                <search-form
                  ref="serviceSearchFormRef"
                  :module-id="service.model.moduleId"
                  v-bind="service.model.searchFormConfig"
                  @search="handleServiceSearch"
                  @toolbar-click="handleServiceToolbarClick"
                />
              </template>

              <!-- 服务数据表格 -->
              <template #2>
                <g-grid
                  ref="serviceGridRef"
                  :module-id="service.model.moduleId"
                  :data="service.model.serviceList"
                  :loading="service.model.loading"
                  v-bind="service.model.gridConfig"
                  @page-change="handleServicePageChange"
                  @menu-click="handleServiceMenuClick"
                >
                  <!-- 分组名称多彩显示 -->
                  <template #groupName="{ row }">
                    <n-tag :type="getGroupTagType(row.groupName)" size="small" :bordered="false">
                      {{ row.groupName }}
                    </n-tag>
                  </template>

                  <!-- 服务名称多彩显示 -->
                  <template #serviceName="{ row }">
                    <n-tag :type="getServiceTagType(row.serviceName)" size="small" round>
                      <template #icon>
                        <n-icon :component="ServerOutline" />
                      </template>
                      {{ row.serviceName }}
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
          </GCard>
        </template>
      </GPane>
    </div>

    <!-- 详情视图 -->
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

    <!-- 服务对话框（新增/编辑共用） -->
    <GdataFormModal
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
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import { GCard } from '@/components/gcard'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { ServerOutline } from '@vicons/ionicons5'
import { NIcon, NTag, useMessage } from 'naive-ui'
import { ref } from 'vue'
import { NamespaceList } from '../hub0041/components'
import type { Namespace } from '../hub0041/types'
import ServiceDetail from './components/ServiceDetail.vue'
import { useServicePage } from './hooks'
import type { Service, ServiceNode } from './types'

// 定义组件名称
defineOptions({
  name: 'ServiceList'
})

// ============= Refs =============

const namespaceListRef = ref()
const serviceSearchFormRef = ref()
const serviceGridRef = ref()

// 选中的命名空间
const selectedNamespace = ref<Namespace | null>(null)

// 详情视图状态
const showDetailView = ref(false)
const currentDetailService = ref<Service | null>(null)
const detailLoading = ref(false)

// 消息提示
const message = useMessage()

// ============= 多颜色标签配置 =============

// 分组名称颜色映射（基于名称哈希生成稳定颜色）
const groupColorTypes: ('default' | 'primary' | 'info' | 'success' | 'warning' | 'error')[] = [
  'primary', 'info', 'success', 'warning', 'error'
]

// 服务名称颜色映射
const serviceColorTypes: ('default' | 'primary' | 'info' | 'success' | 'warning' | 'error')[] = [
  'success', 'info', 'primary', 'warning', 'error'
]

/**
 * 根据字符串生成稳定的哈希值
 */
const hashString = (str: string): number => {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = ((hash << 5) - hash) + char
    hash = hash & hash // Convert to 32bit integer
  }
  return Math.abs(hash)
}

/**
 * 获取分组名称的标签类型（多颜色）
 */
const getGroupTagType = (groupName: string): 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' => {
  if (!groupName) return 'default'
  // DEFAULT_GROUP 使用默认颜色
  if (groupName === 'DEFAULT_GROUP') return 'default'
  const hash = hashString(groupName)
  return groupColorTypes[hash % groupColorTypes.length]
}

/**
 * 获取服务名称的标签类型（多颜色）
 */
const getServiceTagType = (serviceName: string): 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' => {
  if (!serviceName) return 'default'
  const hash = hashString(serviceName)
  return serviceColorTypes[hash % serviceColorTypes.length]
}

// ============= 服务相关 =============

const {
  service,
  formDialogVisible: serviceFormDialogVisible,
  formDialogMode: serviceFormDialogMode,
  currentEditService,
  submitting: serviceSubmitting,
  handleServiceFormSubmit: handleServiceFormSubmitBase,
  handleToolbarClick: handleServiceToolbarClickBase,
  handleMenuClick: handleServiceMenuClickBase,
  handleSearch: handleServiceSearchBase,
} = useServicePage(serviceGridRef, serviceSearchFormRef)

// ============= 事件处理 =============

/**
 * 命名空间行点击 - 选择命名空间并加载服务列表
 */
const handleNamespaceRowClick = async (row: Namespace) => {
  if (row) {
    selectedNamespace.value = row
    // 加载该命名空间下的服务列表（强制使用选中的命名空间ID）
    await service.loadServices({}, row.namespaceId)
  }
}

/**
 * 命名空间选择变化
 */
const handleNamespaceSelect = (namespace: Namespace | null) => {
  selectedNamespace.value = namespace
  // 如果取消选择命名空间，清空服务列表
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
  // 直接调用 service.handleSearch，传入命名空间ID
  service.handleSearch(selectedNamespace.value.namespaceId)
}

/**
 * 服务工具栏点击（必须选择命名空间后才能操作）
 */
const handleServiceToolbarClick = (key: string) => {
  // 新增服务时，如果没有选择命名空间，提示用户
  if (key === 'add' && !selectedNamespace.value) {
    message.warning('请先在上方命名空间列表中选择一个命名空间')
    return
  }
  // 传递选中的命名空间，用于新增时预填充 namespaceId
  handleServiceToolbarClickBase(key, selectedNamespace.value)
}

/**
 * 服务分页变化处理（必须选择命名空间后才能分页）
 */
const handleServicePageChange = (params: { currentPage: number; pageSize: number }) => {
  if (!selectedNamespace.value) {
    return
  }
  service.handlePageChange(params.currentPage, params.pageSize)
}

/**
 * 服务表单提交（自动填充命名空间ID）
 */
const handleServiceFormSubmit = (formData?: Record<string, any>) => {
  handleServiceFormSubmitBase(formData, selectedNamespace.value)
}

/**
 * 服务右键菜单点击处理（覆盖 hook 中的方法，添加查看详情功能）
 */
const handleServiceMenuClick = async (params: { code: string; row?: any }) => {
  if (!params.row) {
    return
  }
  const row = params.row as Service
  
  // 处理查看详情操作
  if (params.code === 'view') {
    await openServiceDetail(row)
    return
  }
  
  // 其他操作（编辑、删除等）使用 hook 中的默认处理
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
      serviceItem.serviceName
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
  if (currentDetailService.value) {
    showDetailView.value = false
    // 打开编辑对话框
    handleServiceMenuClick({
      code: 'edit',
      row: currentDetailService.value,
    })
  }
}

/**
 * 刷新服务详情
 */
const handleDetailRefresh = async () => {
  if (!currentDetailService.value) {
    return
  }
  
  detailLoading.value = true
  try {
    const detailService = await service.getServiceDetail(
      currentDetailService.value.namespaceId,
      currentDetailService.value.groupName,
      currentDetailService.value.serviceName
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
const handleEditNode = (node: ServiceNode) => {
  message.info('节点编辑功能开发中')
}

</script>

<style lang="scss" scoped>
.service-list {
  width: 100%;
  height: 100%;
  overflow: hidden;

  :deep(.n-split) {
    height: 100%;
  }

  /* 上半区：命名空间列表 */
  :deep(.n-split-pane:first-child) {
    overflow: hidden;
    padding: var(--g-space-sm);

    .g-card {
      height: 100%;
      overflow: hidden;

      :deep(.n-card__content) {
        height: 100%;
        overflow: hidden;
      }
    }
  }

  /* 下半区：服务列表 */
  :deep(.n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);

    .g-card {
      height: 100%;
      overflow: hidden;

      :deep(.n-card__content) {
        height: 100%;
        overflow: hidden;
      }
    }
  }

  /* 搜索表单区域 */
  :deep(.n-split-pane .n-split-pane:first-child) {
    overflow: auto;
    padding: var(--g-space-sm);
  }

  /* 表格区域 */
  :deep(.n-split-pane .n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);
    display: flex;
    flex-direction: column;
  }

  .service-list-view {
    width: 100%;
    height: 100%;
  }

  .service-detail-view {
    width: 100%;
    height: 100%;
    overflow: hidden;
  }
}
</style>

