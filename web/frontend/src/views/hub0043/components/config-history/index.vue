<template>
  <div class="config-history-page" :id="page.service.model.moduleId">
    <!-- 列表视图 -->
    <template v-if="currentView === 'list'">
      <GPane direction="vertical" default-size="80px">
        <!-- 上部：搜索表单 -->
        <template #1>
          <search-form
            ref="searchFormRef"
            :module-id="page.service.model.moduleId"
            v-bind="page.service.model.searchFormConfig"
            @search="page.handleSearch"
            @toolbar-click="handleToolbarClick"
          />
        </template>

        <!-- 下部：数据表格 -->
        <template #2>
          <g-grid
            ref="gridRef"
            :module-id="page.service.model.moduleId"
            :data="page.service.model.historyList"
            :loading="page.service.model.loading"
            v-bind="page.service.model.gridConfig"
            @page-change="page.handlePageChange"
            @menu-click="page.handleMenuClick"
          />
        </template>
      </GPane>
    </template>

    <!-- 详情视图 -->
    <template v-else-if="currentView === 'detail'">
      <div class="config-history-detail-view">
        <!-- 返回按钮 -->
        <div class="config-history-detail-header">
          <n-button size="small" @click="page.handleBackToList">
            <template #icon>
              <n-icon><ArrowBackOutline /></n-icon>
            </template>
            返回列表
          </n-button>
        </div>

        <!-- 详情表单 -->
        <GDataForm
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
import { GDataForm } from '@/components'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { ArrowBackOutline } from '@vicons/ionicons5'
import { NButton, NIcon } from 'naive-ui'
import { computed, onMounted, ref, toRaw, watch } from 'vue'
import { useConfigHistoryPage } from './hooks'
import RollbackDialog from './RollbackDialog.vue'

// 定义组件名称
defineOptions({
  name: 'ConfigHistoryPage'
})

// ============= Props & Emits =============
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

// ============= Refs =============
const searchFormRef = ref()
const gridRef = ref()
const detailFormRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============
const page = useConfigHistoryPage(searchFormRef)

// ============= 工具栏点击处理 =============
const handleToolbarClick = (key: string) => {
  if (key === 'back') {
    emit('back')
  } else {
    page.handleToolbarClick(key)
  }
}

// ============= 计算属性（用于模板绑定） =============
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

// ============= 初始化查询条件 =============
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
    // 自动执行搜索
    page.handleSearch()
  }
}

// 监听 initialQuery 变化
watch(() => props.initialQuery, (newQuery) => {
  if (newQuery) {
    fillInitialQuery()
  }
}, { immediate: true })

// 组件挂载后填充查询条件
onMounted(() => {
  if (props.initialQuery) {
    fillInitialQuery()
  }
})

// 暴露 refs 给父组件（如果需要）
defineExpose({
  searchFormRef,
  gridRef,
  service: page.service
})
</script>

<style scoped>
.config-history-page {
  height: 100%;
  width: 100%;
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

.config-history-detail-view :deep(.g-data-form) {
  flex: 1;
  overflow: auto;
  padding: var(--g-space-md);
}
</style>
