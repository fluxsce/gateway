<template>
  <GModal
    :visible="modalVisible"
    :title="props.title || '路由断言配置列表'"
    :width="props.width || 1200"
    :to="props.to"
    :show-footer="false"
    @update:visible="handleUpdateVisible"
    @after-leave="handleAfterLeave"
  >
    <div class="assert-config-list-modal" :id="service.model.moduleId">
      <GPane direction="vertical" :no-resize="true">
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
            :data="service.model.assertList"
            :loading="service.model.loading"
            v-bind="service.model.gridConfig"
            @page-change="service.handlePageChange"
            @menu-click="({ code, row }) => handleMenuClick({ code, row })"
          >
            <!-- 执行顺序自定义渲染 -->
            <template #assertionOrder="{ row }">
              <span style="font-weight: bold; color: #0066cc;">{{ row.assertionOrder }}</span>
            </template>

            <!-- 断言类型自定义渲染 -->
            <template #assertionType="{ row }">
              <n-tag :type="service.model.getAssertionTypeTagType(row.assertionType)" size="small">
                {{ service.model.getAssertionTypeLabel(row.assertionType) }}
              </n-tag>
            </template>

            <!-- 操作符自定义渲染 -->
            <template #assertionOperator="{ row }">
              <n-tag type="info" size="small">
                {{ service.model.getOperatorLabel(row.assertionOperator) }}
              </n-tag>
            </template>

            <!-- 必须匹配自定义渲染 -->
            <template #isRequired="{ row }">
              <n-tag :type="row.isRequired === 'Y' ? 'error' : 'default'" size="small">
                {{ row.isRequired === 'Y' ? '必须' : '可选' }}
              </n-tag>
            </template>

            <!-- 状态自定义渲染 -->
            <template #activeFlag="{ row }">
              <n-switch
                :value="row.activeFlag === 'Y'"
                @update:value="() => handleToggleStatus(row)"
                size="small"
              />
            </template>
          </g-grid>
        </template>
      </GPane>

      <!-- 断言配置对话框（新增/编辑/查看共用） -->
      <GdataFormModal
        v-model:visible="formDialogVisible"
        :mode="formDialogMode"
        :title="formDialogMode === 'create' ? '新增断言配置' : formDialogMode === 'edit' ? '编辑断言配置' : '查看断言配置详情'"
        :to="`#${service.model.moduleId}`"
        :form-fields="service.model.formFields"
        :form-tabs="service.model.formTabs"
        :initial-data="currentEditAssert || undefined"
        :auto-close-on-confirm="false"
        :confirm-loading="service.model.loading.value"
        @submit="handleFormSubmit"
      />
    </div>
  </GModal>
</template>

<script lang="ts" setup>
import { GModal } from '@/components/gmodal'
import { GPane } from '@/components/gpane'
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GGrid } from '@/components/grid'
import { NSwitch, NTag } from 'naive-ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { useAssertConfigPage } from './hooks'
import type { AssertConfigListModalEmits, AssertConfigListModalProps } from './hooks/types'

// 定义组件名称
defineOptions({
  name: 'AssertConfigListModal'
})

// ============= Props =============

const props = withDefaults(defineProps<AssertConfigListModalProps>(), {
  visible: false,
  title: '',
  width: 1200,
  to: undefined,
  routeConfigId: '',
})

// ============= Emits =============

const emit = defineEmits<AssertConfigListModalEmits>()

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 模态框可见性 =============

const modalVisible = ref(props.visible)

// 监听 props.visible 变化，同步到本地状态
const stopVisibleWatch = watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
})

// ============= 路由配置ID =============

const routeConfigId = ref<string | undefined>(props.routeConfigId)

// 监听 props 变化
const stopRouteConfigIdWatch = watch(() => props.routeConfigId, (newVal) => {
  routeConfigId.value = newVal
})

// ============= 资源清理 =============

// 组件卸载时清理所有监听器
onBeforeUnmount(() => {
  stopVisibleWatch()
  stopRouteConfigIdWatch()
})

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditAssert,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handleToggleStatus,
} = useAssertConfigPage(routeConfigId, gridRef, searchFormRef)

// ============= 事件处理 =============

/**
 * 处理模态框可见性变化
 */
const handleUpdateVisible = (value: boolean) => {
  // 更新本地状态
  modalVisible.value = value
  // 通知父组件
  emit('update:visible', value)
  if (!value) {
    emit('close')
  } else {
    // 模态框打开时触发刷新事件并加载数据
    emit('refresh')
    service.loadAssertList()
  }
}

/**
 * 处理模态框关闭动画完成后的回调
 * 重置业务状态
 */
const handleAfterLeave = () => {
  if (!modalVisible.value) {
    // 重置表单对话框状态
    formDialogVisible.value = false
    formDialogMode.value = 'create'
    currentEditAssert.value = null
    // 清空列表数据
    service.model.assertList.value = []
    service.model.resetPagination()
  }
}

</script>

<style scoped>
.assert-config-list-modal {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>

