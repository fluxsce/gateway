<template>
  <GModal
    :visible="modalVisible"
    :title="props.title || 'IP访问控制配置列表'"
    :width="props.width || 1200"
    :to="props.to"
    :show-footer="false"
    @update:visible="handleUpdateVisible"
    @after-leave="handleAfterLeave"
  >
    <div class="ip-access-config-list-modal" :id="service.model.moduleId">
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
            :data="service.model.configList"
            :loading="service.model.loading"
            v-bind="service.model.gridConfig"
            @page-change="service.handlePageChange"
            @menu-click="handleMenuClick"
          >
            <!-- 默认策略自定义渲染 -->
            <template #defaultPolicy="{ row }">
              <n-tag :type="row.defaultPolicy === 'allow' ? 'success' : 'error'" size="small">
                {{ row.defaultPolicy === 'allow' ? '允许' : '拒绝' }}
              </n-tag>
            </template>

            <!-- 活动状态自定义渲染 -->
            <template #activeFlag="{ row }">
              <n-tag :type="row.activeFlag === 'Y' ? 'success' : 'default'" size="small">
                {{ row.activeFlag === 'Y' ? '活动' : '非活动' }}
              </n-tag>
            </template>

            <!-- 信任X-Forwarded-For自定义渲染 -->
            <template #trustXForwardedFor="{ row }">
              <n-tag
                :type="row.trustXForwardedFor === 'Y' ? 'success' : 'default'"
                size="small"
              >
                {{ row.trustXForwardedFor === 'Y' ? '是' : '否' }}
              </n-tag>
            </template>

            <!-- 信任X-Real-IP自定义渲染 -->
            <template #trustXRealIp="{ row }">
              <n-tag
                :type="row.trustXRealIp === 'Y' ? 'success' : 'default'"
                size="small"
              >
                {{ row.trustXRealIp === 'Y' ? '是' : '否' }}
              </n-tag>
            </template>
          </g-grid>
        </template>
      </GPane>

      <!-- IP访问控制配置对话框（新增/编辑/查看共用） -->
      <GdataFormModal
        v-model:visible="formDialogVisible"
        :mode="formDialogMode"
        :title="formDialogMode === 'create' ? '新增IP访问控制配置' : formDialogMode === 'edit' ? '编辑IP访问控制配置' : '查看IP访问控制配置详情'"
        :to="`#${service.model.moduleId}`"
        :form-fields="service.model.formFields"
        :initial-data="currentEditConfig || undefined"
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
import { NTag } from 'naive-ui'
import { ref, watch } from 'vue'
import { useIpAccessConfigPage } from './hooks'
import type { IpAccessConfigListModalEmits, IpAccessConfigListModalProps } from './hooks/types'

// 定义组件名称
defineOptions({
  name: 'IpAccessConfigListModal'
})

// ============= Props =============

const props = withDefaults(defineProps<IpAccessConfigListModalProps>(), {
  visible: false,
  title: 'IP访问控制配置列表',
  width: 1200,
  to: undefined,
  securityConfigId: undefined,
})

// ============= Emits =============

const emit = defineEmits<IpAccessConfigListModalEmits>()

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 模态框可见性 =============

const modalVisible = ref(props.visible)

// 监听 props.visible 变化，同步到本地状态
watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
})

// ============= 安全配置ID =============

const securityConfigId = ref<string | undefined>(props.securityConfigId)

// 监听 props.securityConfigId 变化
watch(() => props.securityConfigId, (newVal) => {
  securityConfigId.value = newVal
})

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditConfig,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useIpAccessConfigPage(gridRef, securityConfigId, searchFormRef)

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
    // 模态框打开时触发刷新事件
    emit('refresh')
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
    currentEditConfig.value = null
    // 清空列表数据
    service.model.configList.value = []
    service.model.resetPagination()
  }
}

</script>

<style scoped>
.ip-access-config-list-modal {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>

