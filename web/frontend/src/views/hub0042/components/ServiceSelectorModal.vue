<template>
  <RsDialog
    :open="visible"
    title="选择注册服务"
    layout="window"
    :width="1200"
    :teleport-to="to"
    :draggable="true"
    :fullscreenable="true"
    :modal="false"
    :show-overlay="false"
    :close-on-overlay-click="false"
    class="hub0042-service-selector-dialog"
    @update:open="handleUpdateVisible"
  >
    <template #body>
      <div class="service-selector-modal" id="hub0042-service-selector">
        <RsSplitPane
          class="service-selector-modal__split"
          orientation="vertical"
          :panes="splitPanes"
          disabled
        >
          <template #search>
            <div class="service-selector-modal__search">
              <RsSearchForm
                ref="searchFormRef"
                :module-id="moduleId"
                :fields="selectorSearchFormConfig.fields"
                :show-search-button="true"
                :show-reset-button="true"
                @search="handleSearch"
              />
            </div>
          </template>

          <template #grid>
            <div class="service-selector-modal__grid">
              <RsGrid
                ref="gridRef"
                :module-id="moduleId"
                :data="service.model.serviceList"
                :loading="service.model.loading"
                :columns="selectorGridColumns"
                :selectable="false"
                :row-key="service.model.gridConfig.rowKey"
                height="100%"
                :pagination-config="selectorPaginationConfig"
                @page-change="handlePageChange"
                @row-click="handleRowClick"
              />
            </div>
          </template>
        </RsSplitPane>
      </div>
    </template>
    <template #footer>
      <RsButton variant="secondary" @click="handleClose">取消</RsButton>
      <RsButton variant="primary" @click="confirmSelection">确认选择</RsButton>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridColumn, type RsGridExpose, type RsGridPaginationConfig } from '@/components/rs-grid'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsButton, RsDialog, RsSplitPane, RsTag, type RsSplitPaneItem } from '@/ui'
import { formatDate } from '@/utils/format'
import { computed, h, onBeforeUnmount, ref, watch } from 'vue'
import { NamespaceNameSelector } from '../../hub0041/components'
import type { Namespace } from '../../hub0041/types'
import { useServiceService } from '../hooks'
import type { Service } from '../types'

defineOptions({
  name: 'ServiceSelectorModal',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

interface Props {
  visible: boolean
  to?: string
}

const props = withDefaults(defineProps<Props>(), {
  to: 'body',
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  select: [service: Service]
  close: []
}>()

const moduleId = 'hub0042'
const message = useAppMessage()
const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)
const selectedService = ref<Service | null>(null)

const service = useServiceService(searchFormRef)

const selectorSearchFormConfig = {
  fields: [
    {
      field: 'namespaceId',
      label: '命名空间',
      type: 'custom' as const,
      span: 8,
      required: true,
      clearable: false,
      placeholder: '请选择命名空间',
      rules: { required: true, message: '请选择命名空间' },
      render: (_formData: Record<string, any>, ctx: { value: unknown; onUpdate: (value: unknown) => void }) => {
        return h(NamespaceNameSelector, {
          modelValue: (ctx.value as string) || '',
          clearable: false,
          placeholder: '请选择命名空间',
          'onUpdate:modelValue': (value: string) => ctx.onUpdate(value),
          onSelect: (namespace: Namespace) => ctx.onUpdate(namespace.namespaceId),
        })
      },
    },
    {
      field: 'serviceName',
      label: '服务名称',
      type: 'input' as const,
      placeholder: '请输入服务名称',
      span: 8,
      clearable: true,
    },
    {
      field: 'groupName',
      label: '分组名称',
      type: 'input' as const,
      placeholder: '请输入分组名称',
      span: 8,
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

const selectorGridColumns: RsGridColumn<Service>[] = [
  {
    key: 'serviceName',
    title: '服务名称',
    minWidth: 180,
    sortable: true,
    fixed: 'left',
  },
  {
    key: 'namespaceId',
    title: '命名空间',
    minWidth: 120,
  },
  {
    key: 'groupName',
    title: '分组',
    minWidth: 120,
  },
  {
    key: 'serviceType',
    title: '服务类型',
    minWidth: 100,
  },
  {
    key: 'nodeCount',
    title: '节点数',
    minWidth: 80,
    align: 'center',
    formatter: (value) => String(value ?? 0),
  },
  {
    key: 'healthyNodeCount',
    title: '健康节点',
    minWidth: 90,
    align: 'center',
    formatter: (value) => String(value ?? 0),
  },
  {
    key: 'unhealthyNodeCount',
    title: '不健康节点',
    minWidth: 100,
    align: 'center',
    formatter: (value) => String(value ?? 0),
  },
  {
    key: 'serviceDescription',
    title: '服务描述',
    minWidth: 180,
    ellipsis: true,
  },
  {
    key: 'activeFlag',
    title: '状态',
    minWidth: 80,
    align: 'center',
    render: (row) =>
      h(
        RsTag,
        { variant: row.activeFlag === 'Y' ? 'success' : 'default', size: 'sm' },
        () => (row.activeFlag === 'Y' ? '启用' : '禁用'),
      ),
  },
  {
    key: 'addTime',
    title: '创建时间',
    minWidth: 160,
    formatter: (value) => (value ? formatDate(value as string) : ''),
  },
]

const selectorPaginationConfig: RsGridPaginationConfig = {
  show: true,
  pageInfo: service.model.pageInfo as any,
  align: 'right',
}

const visible = computed({
  get: () => props.visible,
  set: (value: boolean) => emit('update:visible', value),
})

const stopVisibleWatch = watch(
  () => props.visible,
  (show) => {
    if (show) {
      selectedService.value = null
      service.model.setServiceList([])
    } else {
      selectedService.value = null
      searchFormRef.value?.resetForm?.()
    }
  },
)

onBeforeUnmount(() => {
  stopVisibleWatch()
})

function handleSearch(formData?: Record<string, any>) {
  const namespaceId = String(formData?.namespaceId || searchFormRef.value?.getFormData?.()?.namespaceId || '').trim()
  if (!namespaceId) {
    message.warning('请先选择命名空间')
    return
  }
  service.handleSearch(namespaceId, formData)
}

function handlePageChange(params: { currentPage: number; pageSize: number }) {
  service.handlePageChange(params.currentPage, params.pageSize)
}

function handleRowClick({ row }: { row: Service }) {
  selectedService.value = row
}

function confirmSelection() {
  const fromGrid = gridRef.value?.getSelectedOrCurrentRecord?.() as Service | null | undefined
  const finalSelection = fromGrid ?? selectedService.value

  if (!finalSelection) {
    message.warning('请选择一个服务')
    return
  }

  emit('select', finalSelection)
  handleClose()
}

function handleUpdateVisible(open: boolean) {
  visible.value = open
  if (!open) {
    emit('close')
  }
}

function handleClose() {
  emit('update:visible', false)
  emit('close')
}
</script>

<style scoped>
.service-selector-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}

.service-selector-modal__split {
  flex: 1;
  min-height: 0;
  height: 100%;
}

.service-selector-modal__search {
  width: 100%;
}

.service-selector-modal__grid {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
