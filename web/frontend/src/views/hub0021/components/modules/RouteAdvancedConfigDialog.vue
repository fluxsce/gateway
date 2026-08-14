<template>
  <RsDialog
    :open="visible"
    layout="window"
    :title="`路由配置管理 - ${routeName}`"
    :width="1400"
    :draggable="true"
    :fullscreenable="true"
    @update:open="handleUpdateOpen"
  >
    <template #body>
      <div v-if="routeConfigId" class="route-advanced-body">
        <RsAlert type="info" title="配置说明" class="info-alert">
          <div class="meta-row">
            <span class="meta-text">路由ID：{{ routeConfigId }}</span>
            <RsDivider orientation="vertical" />
            <span class="meta-text">网关实例：{{ gatewayInstanceId }}</span>
            <RsDivider orientation="vertical" />
            <span class="meta-text">配置级别：路由级（优先级高于实例级配置）</span>
          </div>
        </RsAlert>

        <RsTabs
          v-model="activeTab"
          :items="tabItems"
          variant="line"
          size="md"
          borderless
          content-gap="md"
          class="config-tabs"
        >
          <template #predicates>
            <div class="config-section">
              <RsEmpty description="断言配置组件暂不可用（RouteAssertionList 缺失）" />
            </div>
          </template>

          <template #cors>
            <div class="config-section">
              <RsEmpty description="CORS配置组件暂不可用" />
            </div>
          </template>

          <template #security>
            <div class="config-section">
              <RsEmpty description="安全配置组件暂不可用" />
              <RsAlert type="info" title="路由安全配置说明" class="section-alert">
                路由安全配置仅作用于当前路由，支持多种安全策略：
                <ul class="tip-list">
                  <li><strong>IP访问控制</strong>：基于客户端IP地址的白名单/黑名单策略</li>
                  <li><strong>域名访问控制</strong>：限制特定域名的访问权限</li>
                  <li><strong>User-Agent过滤</strong>：基于浏览器标识的访问控制</li>
                  <li><strong>API访问控制</strong>：基于API密钥的访问验证</li>
                  <li>路由级安全配置优先级高于网关实例级配置</li>
                </ul>
                支持的配置类型：IP_ACCESS、DOMAIN_ACCESS、USER_AGENT_ACCESS、API_ACCESS
              </RsAlert>
            </div>
          </template>

          <template #auth>
            <div class="config-section">
              <RsEmpty description="认证配置组件暂不可用" />
            </div>
          </template>

          <template #rateLimit>
            <div class="config-section">
              <RsEmpty description="限流配置组件暂不可用" />
            </div>
          </template>

          <template #filters>
            <div class="config-section">
              <div class="filter-toolbar">
                <div class="toolbar-left">
                  <div class="toolbar-titles">
                    <span class="toolbar-title">路由过滤器管理</span>
                    <span class="toolbar-desc">管理作用于当前路由的后置处理过滤器</span>
                  </div>
                </div>
                <div class="toolbar-right">
                  <RsButton variant="primary" :disabled="!routeConfigId" @click="handleCreateFilter">
                    <GIcon :icon="Add" />
                    新建过滤器
                  </RsButton>
                  <RsButton variant="secondary" :loading="filtersLoading" @click="handleRefreshFilters">
                    <GIcon :icon="Refresh" />
                    刷新
                  </RsButton>
                </div>
              </div>

              <RsTable
                :columns="filterColumns"
                :data="routeFilters"
                row-key="filterConfigId"
                :loading="filtersLoading"
                size="sm"
                striped
              />
              <RsEmpty
                v-if="!filtersLoading && routeFilters.length === 0"
                description="暂无路由过滤器"
                class="filter-empty"
              />

              <RsAlert type="info" title="路由过滤器说明" class="section-alert">
                路由过滤器配置仅作用于当前路由，支持多种执行时机：
                <ul class="tip-list">
                  <li><strong>后置处理</strong>：在路由匹配后、转发到后端服务前执行</li>
                  <li><strong>响应前处理</strong>：在返回响应给客户端前执行</li>
                  <li>支持多个过滤器，按执行顺序依次处理</li>
                  <li>路由级过滤器优先级高于全局过滤器</li>
                </ul>
                支持的过滤器类型：header、query-param、body、strip、rewrite、method、cookie、response
              </RsAlert>
            </div>
          </template>
        </RsTabs>
      </div>
      <div v-else>
        <RsEmpty description="请选择一个路由进行配置管理" />
      </div>
    </template>

    <template #footer>
      <div class="dialog-footer">
        <RsButton variant="secondary" @click="closeDialog">关闭</RsButton>
      </div>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import { GIcon } from '@/components/gicon'
import { useAppMessage } from '@/composables/useAppMessage'
import {
  rsConfirm,
  RsAlert,
  RsButton,
  RsDialog,
  RsDivider,
  RsEmpty,
  RsTable,
  RsTabs,
  type RsTabItem,
  type RsTableColumn,
} from '@/ui'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { Add, Refresh } from '@vicons/ionicons5'
import { computed, h, ref, watch } from 'vue'
import {
  addFilterConfig,
  deleteFilterConfig,
  editFilterConfig,
  queryFilterConfigs,
} from '../../api'
import type { FilterAction, FilterConfig, FilterType } from '../filter-config/hooks/types'
import type { RouteConfig } from '../../types'

/** 遗留编辑表单数据（FilterConfigDialog 恢复后使用） */
interface FilterFormData {
  filterConfigId?: string
  filterName: string
  filterType: FilterType
  filterAction: FilterAction
  filterOrder: number
  filterDesc?: string
  activeFlag: 'Y' | 'N'
  config?: Record<string, unknown>
}

defineOptions({
  name: 'RouteAdvancedConfigDialog',
})

interface Props {
  route?: RouteConfig | null
}

const props = withDefaults(defineProps<Props>(), {
  route: null,
})

const emit = defineEmits<{
  (e: 'update:visible', visible: boolean): void
}>()

const message = useAppMessage()

const visible = ref(false)
const activeTab = ref('predicates')

const routeFilters = ref<FilterConfig[]>([])
const filtersLoading = ref(false)
const filterDialogVisible = ref(false)
const currentFilter = ref<FilterConfig | null>(null)

const routeConfigId = computed(() => props.route?.routeConfigId || '')
const routeName = computed(() => props.route?.routeName || '未知路由')
const gatewayInstanceId = computed(() => props.route?.gatewayInstanceId || '')

const ROUTE_FILTER_ACTIONS = new Set<FilterAction>(['post-routing', 'pre-response'])

const tabItems: RsTabItem[] = [
  { value: 'predicates', label: '断言配置' },
  { value: 'cors', label: 'CORS配置' },
  { value: 'security', label: '安全设置' },
  { value: 'auth', label: '认证配置' },
  { value: 'rateLimit', label: '限流配置' },
  { value: 'filters', label: '过滤器配置' },
]

const filterColumns: RsTableColumn<FilterConfig>[] = [
  { title: '名称', key: 'filterName', width: 160 },
  { title: '类型', key: 'filterType', width: 120 },
  { title: '执行时机', key: 'filterAction', width: 140 },
  { title: '顺序', key: 'filterOrder', width: 80 },
  {
    title: '状态',
    key: 'activeFlag',
    width: 100,
    render: (row) =>
      h(
        'span',
        { style: { color: row.activeFlag === 'Y' ? '#18a058' : '#d03050' } },
        row.activeFlag === 'Y' ? '启用' : '禁用',
      ),
  },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    render: (row) =>
      h('div', { class: 'filter-actions', style: { display: 'flex', gap: '8px' } }, [
        h(
          RsButton,
          { size: 'sm', variant: 'text', tone: 'primary', onClick: () => handleEditFilter(row) },
          () => '编辑',
        ),
        h(
          RsButton,
          { size: 'sm', variant: 'text', onClick: () => handleToggleFilterStatus(row) },
          () => (row.activeFlag === 'Y' ? '禁用' : '启用'),
        ),
        h(
          RsButton,
          { size: 'sm', variant: 'text', onClick: () => handleMoveFilterUp(row) },
          () => '上移',
        ),
        h(
          RsButton,
          { size: 'sm', variant: 'text', onClick: () => handleMoveFilterDown(row) },
          () => '下移',
        ),
        h(
          RsButton,
          { size: 'sm', variant: 'text', tone: 'danger', onClick: () => handleDeleteFilter(row) },
          () => '删除',
        ),
      ]),
  },
]

/** 加载路由过滤器列表 */
const loadRouteFilters = async () => {
  if (!routeConfigId.value) return

  filtersLoading.value = true
  try {
    const response = await queryFilterConfigs({
      routeConfigId: routeConfigId.value,
      pageIndex: 1,
      pageSize: 10000,
    })

    if (isApiSuccess(response)) {
      const parseData = parseJsonData<FilterConfig[] | { list?: FilterConfig[]; data?: FilterConfig[] }>(
        response,
        [],
      )
      const filterList = Array.isArray(parseData)
        ? parseData
        : parseData?.list || parseData?.data || []
      routeFilters.value = filterList
        .filter((filter) => ROUTE_FILTER_ACTIONS.has(filter.filterAction))
        .sort((a, b) => (a.filterOrder || 0) - (b.filterOrder || 0))
    } else {
      routeFilters.value = []
      message.error(getApiMessage(response, '加载路由过滤器失败'))
    }
  } catch (error) {
    console.error('加载路由过滤器失败:', error)
    message.error('加载路由过滤器失败')
    routeFilters.value = []
  } finally {
    filtersLoading.value = false
  }
}

/** 处理创建过滤器 */
const handleCreateFilter = () => {
  currentFilter.value = null
  filterDialogVisible.value = true
  message.warning('过滤器编辑对话框组件暂不可用')
}

/** 处理编辑过滤器 */
const handleEditFilter = (filter: FilterConfig) => {
  currentFilter.value = filter
  filterDialogVisible.value = true
  message.warning('过滤器编辑对话框组件暂不可用')
}

/** 处理删除过滤器 */
const handleDeleteFilter = async (filter: FilterConfig) => {
  const confirmed = await rsConfirm.warning({
    title: '确认删除',
    description: `确定要删除过滤器"${filter.filterName}"吗？此操作不可恢复。`,
    confirmText: '删除',
    cancelText: '取消',
  })
  if (!confirmed) return

  try {
    filtersLoading.value = true
    const response = await deleteFilterConfig(filter.filterConfigId)

    if (isApiSuccess(response)) {
      message.success(getApiMessage(response, '删除过滤器成功'))
      await loadRouteFilters()
    } else {
      message.error(getApiMessage(response, '删除过滤器失败'))
    }
  } catch (error) {
    console.error('删除过滤器失败:', error)
    message.error('删除过滤器失败')
  } finally {
    filtersLoading.value = false
  }
}

/** 处理过滤器状态切换 */
const handleToggleFilterStatus = async (filter: FilterConfig) => {
  try {
    const newStatus = filter.activeFlag === 'Y' ? 'N' : 'Y'
    const response = await editFilterConfig({
      ...filter,
      activeFlag: newStatus,
    })

    if (isApiSuccess(response)) {
      message.success(`过滤器已${newStatus === 'Y' ? '启用' : '禁用'}`)
      await loadRouteFilters()
    } else {
      message.error(getApiMessage(response, '更新过滤器状态失败'))
    }
  } catch (error) {
    console.error('更新过滤器状态失败:', error)
    message.error('更新过滤器状态失败')
  }
}

/** 处理向上移动过滤器 */
const handleMoveFilterUp = async (filter: FilterConfig) => {
  const currentIndex = routeFilters.value.findIndex((f) => f.filterConfigId === filter.filterConfigId)
  if (currentIndex <= 0) return

  const targetFilter = routeFilters.value[currentIndex - 1]
  await swapFilterOrder(filter, targetFilter)
}

/** 处理向下移动过滤器 */
const handleMoveFilterDown = async (filter: FilterConfig) => {
  const currentIndex = routeFilters.value.findIndex((f) => f.filterConfigId === filter.filterConfigId)
  if (currentIndex < 0 || currentIndex >= routeFilters.value.length - 1) return

  const targetFilter = routeFilters.value[currentIndex + 1]
  await swapFilterOrder(filter, targetFilter)
}

/** 交换过滤器顺序 */
const swapFilterOrder = async (filter1: FilterConfig, filter2: FilterConfig) => {
  try {
    filtersLoading.value = true

    const tempOrder = filter1.filterOrder
    const [response1, response2] = await Promise.all([
      editFilterConfig({ ...filter1, filterOrder: filter2.filterOrder }),
      editFilterConfig({ ...filter2, filterOrder: tempOrder }),
    ])

    if (isApiSuccess(response1) && isApiSuccess(response2)) {
      message.success('调整执行顺序成功')
      await loadRouteFilters()
    } else {
      message.error('调整执行顺序失败')
    }
  } catch (error) {
    console.error('调整执行顺序失败:', error)
    message.error('调整执行顺序失败')
  } finally {
    filtersLoading.value = false
  }
}

/** 处理刷新过滤器 */
const handleRefreshFilters = async () => {
  await loadRouteFilters()
  message.success('刷新完成')
}

/** 处理保存过滤器（供 FilterConfigDialog 恢复后调用） */
const handleSaveFilter = async (filterData: FilterFormData) => {
  try {
    const routeLevelFilterAction = filterData.filterConfigId
      ? filterData.filterAction
      : 'post-routing'

    const saveData = {
      tenantId: 'default',
      routeConfigId: routeConfigId.value,
      filterName: filterData.filterName,
      filterType: filterData.filterType,
      filterAction: routeLevelFilterAction,
      filterOrder: filterData.filterOrder,
      filterConfig: JSON.stringify(filterData.config ?? {}),
      filterDesc: filterData.filterDesc,
      activeFlag: filterData.activeFlag,
      addWho: 'admin',
      editWho: 'admin',
    }

    const response = filterData.filterConfigId
      ? await editFilterConfig({
          ...saveData,
          filterConfigId: filterData.filterConfigId,
        })
      : await addFilterConfig(saveData)

    if (isApiSuccess(response)) {
      message.success(
        getApiMessage(response, `${filterData.filterConfigId ? '编辑' : '创建'}过滤器成功`),
      )
      filterDialogVisible.value = false
      await loadRouteFilters()
    } else {
      message.error(
        getApiMessage(response, `${filterData.filterConfigId ? '编辑' : '创建'}过滤器失败`),
      )
    }
  } catch (error) {
    console.error('保存过滤器失败:', error)
    message.error('保存过滤器失败')
  }
}

/** 打开对话框 */
const openDialog = (route: RouteConfig) => {
  if (!route) {
    console.warn('无法打开路由配置管理：缺少路由信息')
    return
  }
  visible.value = true
  void loadRouteFilters()
}

/** 关闭对话框 */
const closeDialog = () => {
  visible.value = false
  emit('update:visible', false)
  routeFilters.value = []
  filterDialogVisible.value = false
  currentFilter.value = null
}

const handleUpdateOpen = (open: boolean) => {
  if (!open) {
    closeDialog()
  } else {
    visible.value = true
  }
}

/** 处理配置变更 */
const handleConfigChange = () => {
  console.log('配置已变更')
}

watch(
  () => visible.value,
  (newVisible) => {
    emit('update:visible', newVisible)
  },
)

watch(
  () => props.route,
  (newRoute) => {
    if (newRoute && visible.value) {
      void loadRouteFilters()
    }
  },
)

defineExpose({
  openDialog,
  closeDialog,
  handleSaveFilter,
  handleConfigChange,
})
</script>

<style scoped>
.route-advanced-body {
  min-height: 400px;
}

.info-alert {
  margin-bottom: 16px;
}

.meta-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.meta-text {
  font-size: 13px;
  color: var(--g-text-tertiary, var(--rs-text-tertiary));
}

.config-section {
  min-height: 400px;
  padding: 8px 0 16px;
}

.section-alert {
  margin-top: 16px;
}

.tip-list {
  margin: 8px 0;
  padding-left: 20px;
}

.tip-list li {
  margin: 4px 0;
}

.filter-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  gap: 12px;
}

.toolbar-left {
  flex: 1;
}

.toolbar-titles {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.toolbar-title {
  font-weight: 600;
  color: var(--g-text-primary, var(--rs-text));
}

.toolbar-desc {
  font-size: 13px;
  color: var(--g-text-tertiary, var(--rs-text-tertiary));
}

.toolbar-right {
  display: flex;
  gap: 8px;
}

.filter-empty {
  margin-top: 12px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
}
</style>
