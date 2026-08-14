<template>
  <RsDialog
    :open="modalVisible"
    :title="props.title || '选择告警渠道'"
    layout="window"
    :width="props.width || 1200"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :modal="false"
    :show-overlay="false"
    :close-on-overlay-click="false"
    class="hub0080-alert-channel-list-dialog"
    @update:open="handleUpdateVisible"
    @after-close="handleAfterLeave"
  >
    <template #body>
      <div class="alert-channel-list-modal" :id="service.model.moduleId">
        <RsSplitPane
          class="alert-channel-list-modal__split"
          orientation="vertical"
          :panes="splitPanes"
          disabled
        >
          <template #search>
            <div class="alert-channel-list-modal__search">
              <RsSearchForm
                ref="searchFormRef"
                :module-id="service.model.moduleId"
                v-bind="service.model.searchFormConfig"
                @search="handleSearch"
              />
            </div>
          </template>

          <template #grid>
            <div class="alert-channel-list-modal__grid">
              <RsGrid
                ref="gridRef"
                :module-id="service.model.moduleId"
                :data="service.model.configList"
                :loading="service.model.loading"
                :columns="service.model.gridConfig.columns"
                :selectable="false"
                :row-key="service.model.gridConfig.rowKey"
                height="100%"
                :pagination-config="service.model.gridConfig.paginationConfig"
                :menu-config="{ enabled: false }"
                @page-change="handlePageChange"
                @row-click="handleRowClick"
              />
            </div>
          </template>
        </RsSplitPane>
      </div>
    </template>
  </RsDialog>
</template>

<script lang="ts" setup>
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsDialog, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { useAlertConfigPage } from '../hooks'
import type { AlertConfig } from '../types'

defineOptions({
  name: 'AlertChannelListModal',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

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

interface Emits {
  (e: 'update:visible', visible: boolean): void
  (e: 'after-leave'): void
  (e: 'select', channel: AlertConfig): void
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

const searchFormRef = ref<RsSearchFormExpose | null>(null)
const gridRef = ref<RsGridExpose | null>(null)
const modalVisible = ref(props.visible)

const stopVisibleWatch = watch(
  () => props.visible,
  (newVal) => {
    modalVisible.value = newVal
    if (newVal) {
      handleSearch()
    }
  },
)

const { service, handleSearch, handlePageChange } = useAlertConfigPage(gridRef, searchFormRef)

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
 * 处理行点击（选择告警渠道）
 */
const handleRowClick = ({ row }: { row: AlertConfig }) => {
  if (row && row.channelName) {
    emit('update:modelValue', row.channelName)
    emit('select', row)
    handleUpdateVisible(false)
  }
}

onBeforeUnmount(() => {
  stopVisibleWatch()
})
</script>

<style scoped>
.alert-channel-list-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}

.alert-channel-list-modal__split {
  flex: 1;
  min-height: 0;
  height: 100%;
}

.alert-channel-list-modal__search {
  width: 100%;
}

.alert-channel-list-modal__grid {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
