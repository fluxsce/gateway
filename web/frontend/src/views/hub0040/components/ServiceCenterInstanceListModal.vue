<template>
  <GModal
    :visible="modalVisible"
    :title="props.title || '选择服务中心实例'"
    :width="props.width || 1200"
    :to="props.to"
    :show-footer="false"
    @update:visible="handleUpdateVisible"
    @after-leave="handleAfterLeave"
  >
    <div class="service-center-instance-list-modal" :id="service.model.moduleId">
      <GPane direction="vertical" :no-resize="true">
        <!-- 上部：搜索表单 -->
        <template #1>
          <search-form
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
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
            @row-click="handleRowClick"
          >
            <!-- 实例状态自定义渲染 -->
            <template #instanceStatus="{ row }">
              <n-tag
                :type="getInstanceStatusType(row.instanceStatus)"
                size="small"
              >
                <template #icon>
                  <n-icon>
                    <CheckmarkCircleOutline v-if="row.instanceStatus === 'RUNNING'" />
                    <AlertCircleOutline v-else-if="row.instanceStatus === 'ERROR'" />
                    <HourglassOutline v-else-if="row.instanceStatus === 'STARTING' || row.instanceStatus === 'STOPPING'" />
                    <StopCircleOutline v-else />
                  </n-icon>
                </template>
                {{ getInstanceStatusText(row.instanceStatus) }}
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
    </div>
  </GModal>
</template>

<script lang="ts" setup>
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GModal } from '@/components/gmodal'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import {
    AlertCircleOutline,
    CheckmarkCircleOutline,
    HourglassOutline,
    StopCircleOutline
} from '@vicons/ionicons5'
import { NIcon, NTag } from 'naive-ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { useServiceCenterInstanceService } from '../hooks/useServiceCenterInstanceService'
import type { ServiceCenterInstance } from '../types'

// 定义组件名称
defineOptions({
  name: 'ServiceCenterInstanceListModal'
})

// ============= Props =============

interface Props {
  /** 是否显示弹窗 */
  visible?: boolean
  /** 弹窗标题 */
  title?: string
  /** 弹窗宽度 */
  width?: number | string
  /** 弹窗挂载目标 */
  to?: string
  /** 选中的实例名称（v-model） */
  modelValue?: string
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  title: '',
  width: 1200,
  to: undefined,
  modelValue: '',
})

// ============= Emits =============

interface Emits {
  (e: 'update:visible', visible: boolean): void
  (e: 'after-leave'): void
  (e: 'select', instance: ServiceCenterInstance): void
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 模态框可见性 =============

const modalVisible = ref(props.visible)

// 监听 props.visible 变化，同步到本地状态
const stopVisibleWatch = watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
  if (newVal) {
    // 弹窗打开时，自动执行查询
    handleSearch()
  }
})

// ============= 服务（只包含查询功能） =============

const service = useServiceCenterInstanceService(searchFormRef)

// 初始化加载数据
service.loadInstances()

// ============= 事件处理 =============

/**
 * 处理弹窗可见性更新
 */
const handleUpdateVisible = (visible: boolean) => {
  modalVisible.value = visible
  emit('update:visible', visible)
}

/**
 * 处理弹窗关闭后事件
 */
const handleAfterLeave = () => {
  emit('after-leave')
}

/**
 * 处理搜索
 */
const handleSearch = () => {
  service.handleSearch()
}

/**
 * 处理行点击事件（选择实例）
 */
const handleRowClick = ({ row }: { row: ServiceCenterInstance }) => {
  if (row) {
    // 更新 v-model 值
    emit('update:modelValue', row.instanceName)
    
    // 触发 select 事件
    emit('select', row)
    
    // 选择后关闭弹窗
    handleUpdateVisible(false)
  }
}

/**
 * 获取实例状态类型
 */
function getInstanceStatusType(status: string): 'default' | 'success' | 'error' | 'warning' | 'info' {
  const statusMap: Record<string, 'default' | 'success' | 'error' | 'warning' | 'info'> = {
    'RUNNING': 'success',
    'STOPPED': 'default',
    'STARTING': 'info',
    'STOPPING': 'warning',
    'ERROR': 'error',
  }
  return statusMap[status] || 'default'
}

/**
 * 获取实例状态文本
 */
function getInstanceStatusText(status: string): string {
  const statusMap: Record<string, string> = {
    'RUNNING': '运行中',
    'STOPPED': '停止',
    'STARTING': '启动中',
    'STOPPING': '停止中',
    'ERROR': '异常',
  }
  return statusMap[status] || status
}

// ============= 生命周期 =============

onBeforeUnmount(() => {
  stopVisibleWatch()
})
</script>

<style lang="scss" scoped>
.service-center-instance-list-modal {
  width: 100%;
  height: 100%;
  min-height: 500px;

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

