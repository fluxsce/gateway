<template>
  <g-modal
    v-model:visible="visible"
    title="选择注册服务"
    :header-icon="SearchOutline"
    :width="1200"
    :mask-closable="false"
    :show-fullscreen-toggle="false"
    :to="to"
    @close="handleClose"
    @cancel="handleClose"
    @confirm="confirmSelection"
    :show-confirm="true"
    :show-cancel="true"
    confirm-text="确认选择"
    cancel-text="取消"
  >
    <div class="service-selector-modal-content">
      <!-- 搜索表单 -->
      <div class="search-section">
        <search-form
          ref="searchFormRef"
          :module-id="moduleId"
          :fields="selectorSearchFormConfig.fields"
          :show-search-button="true"
          :show-reset-button="true"
          @search="handleSearch"
        />
      </div>

      <!-- 服务列表 -->
      <div class="grid-section">
        <g-grid
          ref="gridRef"
          :module-id="moduleId"
          :data="service.model.serviceList"
          :loading="service.model.loading"
          :columns="selectorGridConfig.columns"
          :height="400"
          :show-overflow="true"
          :stripe="true"
          :border="true"
          :row-config="selectorGridConfig.rowConfig"
          :radio-config="selectorGridConfig.radioConfig"
          :pager-config="selectorGridConfig.pagerConfig"
          :page-info="service.model.pageInfo"
          @page-change="handlePageChange"
          @row-click="handleRowClick"
        >
          <!-- 活动状态自定义渲染 -->
          <template #activeFlag="{ row }">
            <n-tag :type="row.activeFlag === 'Y' ? 'success' : 'default'" size="small">
              {{ row.activeFlag === 'Y' ? '启用' : '禁用' }}
            </n-tag>
          </template>
        </g-grid>
      </div>
    </div>
  </g-modal>
</template>

<script setup lang="ts">
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GModal } from '@/components/gmodal'
import { GGrid } from '@/components/grid'
import { formatDate } from '@/utils/format'
import { SearchOutline } from '@vicons/ionicons5'
import { NTag, useMessage } from 'naive-ui'
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useServiceService } from '../hooks'
import type { Service } from '../types'

// Props
interface Props {
  visible: boolean
  to?: string
}

const props = withDefaults(defineProps<Props>(), {
  to: 'body'
})

// Emits
const emit = defineEmits<{
  'update:visible': [value: boolean]
  'select': [service: Service]
  'close': []
}>()

// 模块ID
const moduleId = 'hub0042-selector'

// 响应式状态
const message = useMessage()
const searchFormRef = ref()
const gridRef = ref()
const selectedService = ref<Service | null>(null)

// 使用服务 Service Hook（复用公共业务逻辑）
const service = useServiceService(searchFormRef)

// 选择器专用的搜索表单配置（简化版，不需要工具栏按钮）
const selectorSearchFormConfig = {
  fields: [
    {
      field: 'serviceName',
      label: '服务名称',
      type: 'input' as const,
      placeholder: '请输入服务名称',
      span: 6,
      clearable: true,
    },
    {
      field: 'groupName',
      label: '分组名称',
      type: 'input' as const,
      placeholder: '请输入分组名称',
      span: 6,
      clearable: true,
    },
    {
      field: 'namespaceId',
      label: '命名空间',
      type: 'input' as const,
      placeholder: '请输入命名空间',
      span: 6,
      clearable: true,
    },
    {
      field: 'activeFlag',
      label: '状态',
      type: 'select' as const,
      placeholder: '请选择状态',
      span: 6,
      clearable: true,
      options: [
        { label: '全部', value: '' },
        { label: '启用', value: 'Y' },
        { label: '禁用', value: 'N' },
      ],
    },
  ],
}

// 选择器专用的表格配置（支持单选）
const selectorGridConfig = {
  columns: [
    {
      type: 'radio',
      width: 50,
      fixed: 'left',
    },
    {
      field: 'serviceName',
      title: '服务名称',
      minWidth: 180,
      sortable: true,
      fixed: 'left',
    },
    {
      field: 'namespaceId',
      title: '命名空间',
      minWidth: 120,
    },
    {
      field: 'groupName',
      title: '分组',
      minWidth: 120,
    },
    {
      field: 'serviceType',
      title: '服务类型',
      minWidth: 100,
    },
    {
      field: 'nodeCount',
      title: '节点数',
      minWidth: 80,
      align: 'center',
      formatter: ({ cellValue }: { cellValue: number }) => cellValue ?? 0,
    },
    {
      field: 'healthyNodeCount',
      title: '健康节点',
      minWidth: 90,
      align: 'center',
      formatter: ({ cellValue }: { cellValue: number }) => cellValue ?? 0,
    },
    {
      field: 'unhealthyNodeCount',
      title: '不健康节点',
      minWidth: 100,
      align: 'center',
      formatter: ({ cellValue }: { cellValue: number }) => cellValue ?? 0,
    },
    {
      field: 'serviceDescription',
      title: '服务描述',
      minWidth: 180,
      showOverflow: 'tooltip',
    },
    {
      field: 'activeFlag',
      title: '状态',
      minWidth: 80,
      align: 'center',
      slots: { default: 'activeFlag' },
    },
    {
      field: 'addTime',
      title: '创建时间',
      minWidth: 160,
      formatter: ({ cellValue }: { cellValue: string }) => formatDate(cellValue),
    },
  ] as any[],
  rowConfig: {
    keyField: 'serviceName',
    isHover: true,
    isCurrent: true,
  },
  radioConfig: {
    highlight: true,
    trigger: 'row',
  },
  pagerConfig: {
    enabled: true,
    pageSize: 10,
    pageSizes: [10, 20, 50],
  },
}

// 计算属性
const visible = computed({
  get: () => props.visible,
  set: (value: boolean) => emit('update:visible', value)
})

// 监听弹窗显示状态
const stopVisibleWatch = watch(() => props.visible, (show) => {
  if (show) {
    selectedService.value = null
    // 加载服务列表
    service.loadServices({ tenantId: 'default' })
  } else {
    selectedService.value = null
    // 重置搜索表单
    if (searchFormRef.value?.resetForm) {
      searchFormRef.value.resetForm()
    }
  }
})

// 组件卸载时清理监听器
onBeforeUnmount(() => {
  stopVisibleWatch()
})

// 搜索处理
function handleSearch() {
  service.handleSearch()
}

// 分页处理
function handlePageChange(params: { currentPage: number; pageSize: number }) {
  service.handlePageChange(params.currentPage, params.pageSize)
}

// 行点击处理（选择服务）
function handleRowClick({ row }: { row: Service }) {
  selectedService.value = row
  // 设置表格单选选中状态
  if (gridRef.value?.setRadioRow) {
    gridRef.value.setRadioRow(row)
  }
}

// 确认选择
function confirmSelection() {
  // 优先从表格获取选中的行
  const radioRecord = gridRef.value?.getRadioRecord?.()
  const finalSelection = radioRecord || selectedService.value
  
  if (!finalSelection) {
    message.warning('请选择一个服务')
    return
  }
  
  emit('select', finalSelection)
  handleClose()
}

// 关闭弹窗
function handleClose() {
  emit('update:visible', false)
  emit('close')
}
</script>

<style scoped lang="scss">
.service-selector-modal-content {
  display: flex;
  flex-direction: column;
  gap: 16px;

  .search-section {
    flex-shrink: 0;
  }

  .grid-section {
    flex: 1;
    min-height: 0;
  }
}
</style>
