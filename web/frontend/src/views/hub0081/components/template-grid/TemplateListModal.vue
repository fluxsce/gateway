<template>
  <GModal
    :visible="modalVisible"
    :title="props.title || '选择模板'"
    :width="props.width || 1200"
    :to="props.to"
    :show-footer="false"
    @update:visible="handleUpdateVisible"
    @after-leave="handleAfterLeave"
  >
    <div class="template-list-modal" :id="service.model.moduleId">
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
            :data="service.model.templateList"
            :loading="service.model.loading"
            v-bind="service.model.gridConfig"
            @page-change="handlePageChange"
            @row-click="handleRowClick"
          >
            <template #channelType="{ row }">
              <n-tag size="small" type="info">
                {{ service.model.getChannelTypeLabel(row.channelType) || '通用' }}
              </n-tag>
            </template>

            <template #displayFormat="{ row }">
              <n-tag size="small" :type="row.displayFormat === 'table' ? 'warning' : 'default'">
                {{ service.model.getDisplayFormatLabel(row.displayFormat) }}
              </n-tag>
            </template>

            <template #activeFlag="{ row }">
              <n-tag size="small" :type="row.activeFlag === 'Y' ? 'success' : 'default'">
                {{ row.activeFlag === 'Y' ? '启用' : '禁用' }}
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
import { useAlertTemplateListPage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'TemplateListModal'
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
  /** 选中的模板名称（v-model） */
  modelValue?: string
  /** 渠道类型（可选，用于过滤模板） */
  channelType?: string
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  title: '',
  width: 1200,
  to: undefined,
  modelValue: '',
  channelType: undefined,
})

// ============= Emits =============

interface Emits {
  (e: 'update:visible', visible: boolean): void
  (e: 'after-leave'): void
  (e: 'select', template: any): void
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

// ============= 页面级 Hook（只包含查询功能） =============

const {
  service,
  handleSearch,
  handlePageChange,
} = useAlertTemplateListPage(gridRef, searchFormRef, props.channelType)

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
 * 处理行点击事件（选择模板）
 */
const handleRowClick = ({ row }: { row: any }) => {
  if (row) {
    // 获取模板名称
    const templateName = row.templateName || ''
    
    // 更新 v-model 值
    emit('update:modelValue', templateName)
    
    // 触发 select 事件（保留向后兼容）
    emit('select', row)
    
    // 选择后关闭弹窗
    handleUpdateVisible(false)
  }
}

// ============= 生命周期 =============

onBeforeUnmount(() => {
  stopVisibleWatch()
})
</script>

<style lang="scss" scoped>
.template-list-modal {
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

