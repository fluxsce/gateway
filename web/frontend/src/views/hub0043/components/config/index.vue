<template>
  <div class="config-management" :id="service.model.moduleId">
    <!-- 列表视图 -->
    <template v-if="currentView === 'list'">
      <GPane direction="vertical" default-size="80px">
        <!-- 上部：搜索表单 -->
        <template #1>
          <search-form
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
            @toolbar-click="handleToolbarClick"
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
            @page-change="service.handlePageChange"
            @menu-click="handleMenuClick"
          />
        </template>
      </GPane>
    </template>

    <!-- 表单视图（新增/编辑/查看共用） -->
    <template v-else-if="currentView === 'form'">
      <div class="config-form-view">
        <!-- 返回按钮 -->
        <div class="config-form-header">
          <n-button size="small" @click="handleBackToList">
            <template #icon>
              <n-icon><ArrowBackOutline /></n-icon>
            </template>
            返回列表
          </n-button>
        </div>

        <!-- 扁平表单 -->
        <GDataForm
          ref="formRef"
          :mode="formDialogMode"
          :form-fields="service.model.configFormConfig.fields"
          :initial-data="currentEditConfig || undefined"
          :show-footer="true"
          :show-submit="formDialogMode !== 'view'"
          :submit-text="formDialogMode === 'create' ? '发布' : '保存'"
          :submit-loading="submitting"
          @submit="handleFormSubmit"
        />
      </div>
    </template>
  </div>
</template>

<script lang="ts" setup>
import { GDataForm } from '@/components'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { ArrowBackOutline } from '@vicons/ionicons5'
import { NButton, NIcon } from 'naive-ui'
import { ref } from 'vue'
import type { Config } from '../../types'
import { useConfigPage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'ConfigManagement'
})

// ============= Props & Emits =============
interface Props {
  /** 是否显示历史按钮（用于跳转到历史页面） */
  showHistoryButton?: boolean
}

interface Emits {
  /** 查看历史事件 */
  (e: 'view-history', config: Config): void
}

const props = withDefaults(defineProps<Props>(), {
  showHistoryButton: true,
})

const emit = defineEmits<Emits>()

// ============= Refs =============
const searchFormRef = ref()
const gridRef = ref()
const formRef = ref()

// ============= 页面级 Hook（包含服务、视图切换、事件处理等） =============
const {
  service,
  currentView,
  formDialogMode,
  currentEditConfig,
  submitting,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handleBackToList,
} = useConfigPage(
  gridRef,
  searchFormRef,
  formRef,
  {
    useViewMode: true, // 使用视图切换模式
    onViewChange: (view) => {
      // 视图切换回调（如果需要额外处理可以在这里添加）
    },
    onHistoryClick: (config) => {
      // 历史事件回调
      if (props.showHistoryButton) {
        emit('view-history', config)
      }
    },
  }
)

// ============= 暴露方法 =============
defineExpose({
  refresh: service.handleRefresh,
  loadConfigs: service.loadConfigs,
})
</script>

<style scoped>
.config-management {
  height: 100%;
  width: 100%;
}

.config-form-view {
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.config-form-header {
  display: flex;
  align-items: center;
  padding: var(--g-space-sm) var(--g-space-md);
  border-bottom: 1px solid var(--g-border-primary);
  background-color: var(--g-bg-color);
}

.config-form-view :deep(.g-data-form) {
  flex: 1;
  overflow: auto;
  padding: var(--g-space-md);
}
</style>

