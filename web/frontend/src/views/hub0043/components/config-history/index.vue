<template>
  <div class="config-history-page" :id="page.service.model.moduleId">
    <!-- 列表视图 -->
    <template v-if="currentView === 'list'">
      <RsSplitPane
        class="config-history-page__split"
        orientation="vertical"
        :panes="splitPanes"
        disabled
      >
        <template #search>
          <div class="config-history-page__search">
            <RsSearchForm
              ref="searchFormRef"
              :module-id="page.service.model.moduleId"
              v-bind="page.service.model.searchFormConfig"
              @search="page.handleSearch"
              @toolbar-click="handleToolbarClick"
            />
          </div>
        </template>

        <template #grid>
          <div class="config-history-page__grid">
            <RsGrid
              ref="gridRef"
              :module-id="page.service.model.moduleId"
              :data="page.service.model.historyList"
              :loading="page.service.model.loading"
              :columns="page.service.model.gridConfig.columns"
              :selectable="page.service.model.gridConfig.selectable"
              :row-key="page.service.model.gridConfig.rowKey"
              height="100%"
              :pagination-config="page.service.model.gridConfig.paginationConfig"
              :menu-config="page.service.model.gridConfig.menuConfig"
              @page-change="page.handlePageChange"
              @menu-click="page.handleMenuClick"
            />
          </div>
        </template>
      </RsSplitPane>
    </template>

    <!-- 详情视图 -->
    <template v-else-if="currentView === 'detail'">
      <div class="config-history-detail-view">
        <div class="config-history-detail-header">
          <RsButton size="sm" icon="arrow-left" @click="page.handleBackToList">
            返回列表
          </RsButton>
        </div>

        <RsDataForm
          v-if="currentHistoryDetailPlain"
          ref="detailFormRef"
          mode="view"
          :form-fields="page.service.model.detailFormConfig.fields"
          :form-tabs="page.service.model.detailFormConfig.tabs"
          :initial-data="currentHistoryDetailPlain"
        />
      </div>
    </template>

    <!-- 回滚确认对话框 -->
    <RollbackDialog
      v-model:visible="rollbackDialogVisible"
      :history="currentRollbackHistory"
      :submitting="submitting"
      @confirm="page.handleRollbackConfirm"
      @cancel="page.closeRollbackDialog"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataForm, type RsDataFormExpose } from '@/components/form/rs-data'
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsButton, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { computed, onMounted, ref, toRaw, watch } from 'vue'
import { useConfigHistoryPage } from './hooks'
import RollbackDialog from './RollbackDialog.vue'

defineOptions({
  name: 'ConfigHistoryPage',
})

interface Props {
  /** 初始查询条件 */
  initialQuery?: {
    namespaceId: string
    groupName: string
    configDataId: string
  } | null
}

interface Emits {
  /** 返回配置列表事件 */
  (e: 'back'): void
}

const props = withDefaults(defineProps<Props>(), {
  initialQuery: null,
})

const emit = defineEmits<Emits>()

/** 上方搜索区随内容自适应，下方表格占满剩余高度 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref<RsSearchFormExpose | null>(null)
const gridRef = ref<RsGridExpose | null>(null)
const detailFormRef = ref<RsDataFormExpose | null>(null)

const page = useConfigHistoryPage(searchFormRef)

/**
 * 工具栏点击：返回配置列表由父组件处理，其余交给 page hook。
 */
const handleToolbarClick = (key: string) => {
  if (key === 'back') {
    emit('back')
  } else {
    page.handleToolbarClick(key)
  }
}

const currentView = computed(() => page.currentView.value)

const rollbackDialogVisible = computed({
  get: () => page.rollbackDialogVisible.value,
  set: (val: boolean) => { page.rollbackDialogVisible.value = val }
})

const currentRollbackHistory = computed(() => page.currentRollbackHistory.value)
const submitting = computed(() => page.submitting.value)

// 将响应式对象转换为普通对象，避免 JSON.stringify 循环引用错误
const currentHistoryDetailPlain = computed(() => {
  const detail = page.currentHistoryDetail.value
  return detail ? toRaw(detail) : null
})

/**
 * 填充初始查询条件到搜索表单
 */
const fillInitialQuery = () => {
  if (props.initialQuery && searchFormRef.value?.setFormData) {
    searchFormRef.value.setFormData({
      namespaceId: props.initialQuery.namespaceId,
      groupName: props.initialQuery.groupName,
      configDataId: props.initialQuery.configDataId,
      limit: 50,
    })
    page.handleSearch()
  }
}

watch(() => props.initialQuery, (newQuery) => {
  if (newQuery) {
    fillInitialQuery()
  }
}, { immediate: true })

onMounted(() => {
  if (props.initialQuery) {
    fillInitialQuery()
  }
})

defineExpose({
  searchFormRef,
  gridRef,
  service: page.service
})
</script>

<style lang="scss" scoped>
.config-history-page {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.config-history-page__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.config-history-page__search {
  width: 100%;
  box-sizing: border-box;
}

.config-history-page__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.config-history-detail-view {
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.config-history-detail-header {
  display: flex;
  align-items: center;
  padding: var(--g-space-sm) var(--g-space-md);
  border-bottom: 1px solid var(--g-border-primary);
  background-color: var(--g-bg-color);
}

.config-history-detail-view :deep(.rs-data-form) {
  flex: 1;
  overflow: auto;
  padding: var(--g-space-md);
}
</style>
