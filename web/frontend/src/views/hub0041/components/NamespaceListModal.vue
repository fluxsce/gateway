<template>
  <RsDialog
    :open="modalVisible"
    :title="props.title || '选择命名空间'"
    layout="window"
    :width="props.width || 1200"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :modal="false"
    :show-overlay="false"
    :close-on-overlay-click="false"
    class="hub0041-namespace-list-dialog"
    @update:open="handleUpdateVisible"
    @after-close="handleAfterLeave"
  >
    <template #body>
      <div class="namespace-list-modal" :id="service.model.moduleId">
        <RsSplitPane
          class="namespace-list-modal__split"
          orientation="vertical"
          :panes="splitPanes"
          disabled
        >
          <template #search>
            <div class="namespace-list-modal__search">
              <RsSearchForm
                ref="searchFormRef"
                :module-id="service.model.moduleId"
                :fields="pickerSearchFields"
                :toolbar-buttons="pickerToolbarButtons"
                :show-search-button="false"
                :show-reset-button="false"
                @reset="handlePickerReset"
              />
            </div>
          </template>

          <template #grid>
            <div class="namespace-list-modal__grid">
              <RsGrid
                :module-id="service.model.moduleId"
                :data="service.model.namespaceList"
                :loading="service.model.loading"
                :columns="pickerColumns"
                :selectable="false"
                row-key="namespaceId"
                height="100%"
                :pagination-config="service.model.gridConfig.paginationConfig"
                :menu-config="{ enabled: false, items: [] }"
                @page-change="service.handlePageChange"
                @row-click="handleRowClick"
                @row-dblclick="handleRowClick"
              />
            </div>
          </template>
        </RsSplitPane>
      </div>
    </template>
  </RsDialog>
</template>

<script lang="ts" setup>
import { RsSearchForm, type RsSearchFormProps, type RsSearchFormRenderContext } from '@/components/form/rs-search'
import { RsGrid, type RsGridColumn } from '@/components/rs-grid'
import type { ToolbarButton } from '@/components/toolbar'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsDialog, RsSplitPane, RsTag, type RsSplitPaneItem } from '@/ui'
import { h, onBeforeUnmount, ref, watch } from 'vue'
import { ServiceCenterInstanceNameSelector } from '../../hub0040/components'
import type { ServiceCenterInstance } from '../../hub0040/types'
import { useNamespaceService } from '../hooks/useNamespaceService'
import type { Namespace } from '../types'

defineOptions({
  name: 'NamespaceListModal',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const envMap: Record<string, string> = {
  DEVELOPMENT: '开发',
  STAGING: '预发布',
  PRODUCTION: '生产',
}

interface Props {
  visible?: boolean
  title?: string
  width?: number | string
  to?: string
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
  (e: 'select', namespace: Namespace): void
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

const searchFormRef = ref()
const modalVisible = ref(props.visible)
const selectedNamespaceId = ref(props.modelValue || '')
const message = useAppMessage()

const service = useNamespaceService(searchFormRef, 'hub0041:picker')

const readInstanceName = (formData?: Record<string, any>) => {
  const data = formData || searchFormRef.value?.getFormData?.() || {}
  return String(data.instanceName || '').trim()
}

const handlePickerSearch = async (formData?: Record<string, any>) => {
  const data = formData || searchFormRef.value?.getFormData?.() || {}
  const instanceName = readInstanceName(data)
  if (!instanceName) {
    message.warning('请选择服务中心实例')
    try {
      await searchFormRef.value?.getFormRef?.()?.validate?.()
    } catch {
      // 表单展示必填错误
    }
    return
  }
  await service.handleSearch(data)
}

const handlePickerReset = () => {
  service.model.resetPagination()
  service.model.setNamespaceList([])
  service.model.updatePagination({ pageIndex: 1, totalCount: 0 })
}

const pickerSearchFields: RsSearchFormProps['fields'] = [
  {
    field: 'instanceName',
    label: '服务中心实例',
    type: 'custom',
    span: 8,
    required: true,
    clearable: false,
    placeholder: '请选择服务中心实例',
    rules: { required: true, message: '请选择服务中心实例' },
    render: (_formData: Record<string, any>, ctx: RsSearchFormRenderContext) => {
      return h(ServiceCenterInstanceNameSelector, {
        modelValue: (ctx.value as string) || '',
        clearable: false,
        placeholder: '请选择服务中心实例',
        'onUpdate:modelValue': (value: string) => ctx.onUpdate(value),
        onSelect: (instance: ServiceCenterInstance) => {
          ctx.onUpdate(instance.instanceName)
          handlePickerSearch({
            ...(searchFormRef.value?.getFormData?.() || {}),
            instanceName: instance.instanceName,
          })
        },
      })
    },
  },
  {
    field: 'namespaceName',
    label: '命名空间名称',
    type: 'input',
    placeholder: '请输入命名空间名称',
    span: 8,
    clearable: true,
  },
  {
    field: 'namespaceId',
    label: '命名空间ID',
    type: 'input',
    placeholder: '请输入命名空间ID',
    span: 8,
    clearable: true,
  },
]

const pickerToolbarButtons: ToolbarButton[] = [
  {
    key: 'search',
    label: '查询',
    icon: 'SearchOutline',
    type: 'primary',
    skipPermission: true,
    onClick: () => {
      void handlePickerSearch()
    },
  },
  {
    key: 'reset',
    label: '重置',
    icon: 'RefreshOutline',
    skipPermission: true,
    onClick: () => {
      searchFormRef.value?.resetForm?.()
    },
  },
]

const pickerColumns: RsGridColumn<Namespace>[] = [
  { key: 'namespaceName', title: '命名空间名称', ellipsis: true, minWidth: 160 },
  { key: 'namespaceId', title: '命名空间ID', ellipsis: true, minWidth: 160 },
  { key: 'instanceName', title: '服务中心实例', ellipsis: true, minWidth: 140 },
  {
    key: 'environment',
    title: '环境',
    width: 100,
    formatter: (value) => envMap[String(value || '')] || String(value || '-'),
  },
  {
    key: 'activeFlag',
    title: '状态',
    width: 90,
    align: 'center',
    render: (row) =>
      h(
        RsTag,
        { variant: row.activeFlag === 'Y' ? 'success' : 'default', size: 'sm' },
        () => (row.activeFlag === 'Y' ? '活动' : '非活动'),
      ),
  },
]

const stopVisibleWatch = watch(
  () => props.visible,
  (newVal) => {
    modalVisible.value = newVal
    if (!newVal) return
    selectedNamespaceId.value = props.modelValue || ''
    const formData = searchFormRef.value?.getFormData?.() || {}
    if (formData.instanceName) {
      service.handleSearch(formData)
    } else {
      service.model.setNamespaceList([])
      service.model.updatePagination({ pageIndex: 1, totalCount: 0 })
    }
  },
)

const handleUpdateVisible = (visible: boolean) => {
  modalVisible.value = visible
  emit('update:visible', visible)
}

const handleAfterLeave = () => {
  emit('after-leave')
}

const handleRowClick = (params: { row: Namespace }) => {
  const namespace = params?.row
  if (!namespace) return
  selectedNamespaceId.value = namespace.namespaceId
  emit('update:modelValue', namespace.namespaceId)
  emit('select', namespace)
  handleUpdateVisible(false)
}

onBeforeUnmount(() => {
  stopVisibleWatch()
})
</script>

<style scoped>
.namespace-list-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}

.namespace-list-modal__split {
  flex: 1;
  min-height: 0;
  height: 100%;
}

.namespace-list-modal__search {
  width: 100%;
  overflow: visible;
}

.namespace-list-modal__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}
</style>
