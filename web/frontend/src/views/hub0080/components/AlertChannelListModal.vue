<template>
  <GModal
    :visible="modalVisible"
    :title="props.title || '选择告警渠道'"
    :width="props.width || 1200"
    :to="props.to"
    :show-footer="false"
    @update:visible="handleUpdateVisible"
    @after-leave="handleAfterLeave"
  >
    <div class="alert-channel-list-modal" :id="service.model.moduleId">
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
            :data="service.model.configList"
            :loading="service.model.loading"
            v-bind="service.model.gridConfig"
            @page-change="handlePageChange"
            @row-click="handleRowClick"
          >
            <template #channelType="{ row }">
              <n-tag size="small" :type="getChannelTypeTagType(row.channelType)">
                {{ getChannelTypeLabel(row.channelType) }}
              </n-tag>
            </template>

            <template #activeFlag="{ row }">
              <n-tag size="small" :type="row.activeFlag === 'Y' ? 'success' : 'default'">
                {{ row.activeFlag === 'Y' ? '启用' : '禁用' }}
              </n-tag>
            </template>

            <template #defaultFlag="{ row }">
              <n-tag size="small" :type="row.defaultFlag === 'Y' ? 'warning' : 'default'">
                {{ row.defaultFlag === 'Y' ? '默认' : '-' }}
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
import { NTag } from 'naive-ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { useAlertConfigPage } from '../hooks'
import { CHANNEL_TYPE_OPTIONS } from '../types'

// 定义组件名称
defineOptions({
  name: 'AlertChannelListModal'
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
  /** 选中的告警渠道名称（v-model） */
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
  (e: 'select', channel: any): void
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

// ============= 使用告警配置 Page Hook =============

const { service, handleSearch: handleSearchInternal, handlePageChange: handlePageChangeInternal } = useAlertConfigPage(gridRef, searchFormRef)

// ============= 工具函数 =============

/**
 * 获取渠道类型显示标签
 */
const getChannelTypeLabel = (channelType: string) => {
  const option = CHANNEL_TYPE_OPTIONS.find(opt => opt.value === channelType)
  return option?.label || channelType
}

/**
 * 获取渠道类型标签颜色
 */
const getChannelTypeTagType = (channelType: string): "default" | "success" | "error" | "warning" | "primary" | "info" => {
  const typeMap: Record<string, "default" | "success" | "error" | "warning" | "primary" | "info"> = {
    email: 'primary',
    qq: 'info',
    wechat_work: 'success',
    dingtalk: 'warning',
    webhook: 'default',
    sms: 'error',
  }
  return typeMap[channelType] || 'default'
}

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
const handleSearch = async () => {
  await handleSearchInternal()
}

/**
 * 处理分页变化
 */
const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
  await handlePageChangeInternal({ currentPage, pageSize })
}

/**
 * 处理行点击（选择告警渠道）
 */
const handleRowClick = ({ row }: { row: any }) => {
  if (row && row.channelName) {
    emit('update:modelValue', row.channelName)
    emit('select', row)
    handleUpdateVisible(false)
  }
}

// ============= 清理 =============

onBeforeUnmount(() => {
  stopVisibleWatch()
})
</script>

<style lang="scss" scoped>
.alert-channel-list-modal {
  width: 100%;
  height: 600px;
  overflow: hidden;

  :deep(.n-split) {
    height: 100%;
  }

  /* 上半区：搜索表单 */
  :deep(.n-split-pane:first-child) {
    overflow: auto;
    padding: var(--g-space-sm);
  }

  /* 下半区：表格区域 */
  :deep(.n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);
    display: flex;
    flex-direction: column;
  }
}
</style>

