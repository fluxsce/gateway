<template>
  <GDrawer
    v-model:show="drawerVisible"
    :title="drawerTitle"
    :width="600"
    :to="'#hub0005'"
    :mask="false"
    :show-footer="true"
    :show-cancel="true"
    :show-confirm="true"
    :confirm-loading="loading"
    cancel-text="取消"
    confirm-text="保存"
    @confirm="handleSave"
    @cancel="handleCancel"
    @close="handleClose"
  >
    <template #default>
      <div class="role-resource-drawer">
        <!-- 加载状态 -->
        <n-spin v-if="loading && !treeData.length" :show="true">
          <template #description>加载资源列表...</template>
        </n-spin>

        <template v-else>
          <!-- 搜索框 -->
          <div class="search-container">
            <n-input
              v-model:value="searchKeyword"
              placeholder="搜索资源名称"
              clearable
              size="small"
            >
              <template #prefix>
                <n-icon><SearchOutline /></n-icon>
              </template>
            </n-input>
          </div>

          <!-- 资源树 -->
          <GTree
            :data="filteredTreeData"
            :checkable="true"
            :cascade="true"
            :check-strategy="'child'"
            :show-line="true"
            :show-icon="true"
            :default-expanded-keys="expandedKeys"
            :checked-keys="checkedKeys"
            :virtual-scroll="true"
            :render-label="renderLabel"
            @update:checkedKeys="handleCheckedKeysChange"
          />
        </template>
      </div>
    </template>
  </GDrawer>
</template>

<script setup lang="ts">
import { GDrawer } from '@/components/gdrawer'
import { GTree } from '@/components/gtree'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import type { Resource } from '@/views/hub0006/types'
import { CheckmarkCircleOutline, FolderOutline, GridOutline, ListOutline, SearchOutline, ServerOutline } from '@vicons/ionicons5'
import type { TreeOption } from 'naive-ui'
import { NIcon, NInput, NSpin, useMessage } from 'naive-ui'
import { computed, h, markRaw, ref, watch } from 'vue'
import { getRoleResources, saveRoleResources } from '../api'

defineOptions({
  name: 'RoleResourceDrawer'
})

interface Props {
  /** 是否显示抽屉 */
  show?: boolean
  /** 角色ID */
  roleId?: string
  /** 角色名称 */
  roleName?: string
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  roleId: '',
  roleName: ''
})

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'close'): void
  (e: 'success'): void
}

const emit = defineEmits<Emits>()

const message = useMessage()

// 抽屉显示状态
const drawerVisible = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

// 抽屉标题
const drawerTitle = computed(() => {
  if (props.roleName) {
    return `角色授权 - ${props.roleName}`
  }
  return '角色授权'
})

// 加载状态
const loading = ref(false)

// 树形数据
const treeData = ref<TreeOption[]>([])

// 默认展开的节点
const defaultExpandedKeys = ref<string[]>([])

// 选中的节点（已授权的资源）
const checkedKeys = ref<string[]>([])

// 当前选中的节点（用于保存）
const currentCheckedKeys = ref<string[]>([])

// 搜索关键词
const searchKeyword = ref('')

// 计算属性 - 过滤后的树形数据
const filteredTreeData = computed(() => {
  if (!searchKeyword.value) {
    return treeData.value
  }
  const keyword = searchKeyword.value.toLowerCase()
  return filterTreeData(treeData.value, keyword)
})

// 计算属性 - 展开的节点（有搜索关键词时，展开所有匹配的节点）
const expandedKeys = computed(() => {
  if (!searchKeyword.value) {
    return defaultExpandedKeys.value
  }
  // 有搜索关键词时，展开所有过滤后的节点
  return extractAllKeysFromTreeData(filteredTreeData.value)
})

// 方法 - 过滤树形数据
function filterTreeData(data: TreeOption[], keyword: string): TreeOption[] {
  const result: TreeOption[] = []
  
  for (const item of data) {
    const label = (item.label as string) || ''
    const matches = label.toLowerCase().includes(keyword)
    
    let children: TreeOption[] | undefined
    if (item.children && item.children.length > 0) {
      children = filterTreeData(item.children, keyword)
    }
    
    if (matches || (children && children.length > 0)) {
      result.push({
        ...item,
        children: children && children.length > 0 ? children : undefined
      })
    }
  }
  
  return result
}

// 监听抽屉显示状态，只在抽屉打开时加载数据
watch(
  () => props.show,
  (newShow) => {
    if (newShow && props.roleId) {
      // 抽屉打开且有角色ID时，加载资源列表
      loadRoleResources()
    } else if (!newShow) {
      // 关闭时重置数据
      treeData.value = []
      checkedKeys.value = []
      currentCheckedKeys.value = []
      defaultExpandedKeys.value = []
      searchKeyword.value = ''
    }
  }
)

/**
 * 加载角色授权的资源列表
 */
async function loadRoleResources() {
  if (!props.roleId) {
    return
  }

  loading.value = true
  try {
    const response = await getRoleResources(props.roleId)

    if (isApiSuccess(response)) {
      // parseJsonData 会从 response.bizData 中解析 JSON 字符串，直接返回 Resource[] 数组
      const resources = parseJsonData<Resource[]>(response, [])
      
      // 转换为树形数据
      treeData.value = convertToTreeData(resources)
      
      // 提取已授权的资源ID
      const authorizedIds = extractCheckedKeys(resources)
      checkedKeys.value = authorizedIds
      currentCheckedKeys.value = [...authorizedIds]
      
      // 默认不展开节点（需要时用户可手动展开）
      defaultExpandedKeys.value = []
    } else {
      message.error(getApiMessage(response) || '加载资源列表失败')
    }
  } catch (error: any) {
    message.error(error.message || '加载资源列表失败')
  } finally {
    loading.value = false
  }
}

/**
 * 根据资源类型获取对应的图标组件
 */
function getResourceTypeIcon(resourceType: string) {
  switch (resourceType) {
    case 'MODULE':
      return markRaw(GridOutline) // 网格图标，表示模块
    case 'GROUP':
      return markRaw(FolderOutline) // 文件夹图标，表示分组
    case 'MENU':
      return markRaw(ListOutline) // 列表图标，表示菜单
    case 'BUTTON':
      return markRaw(CheckmarkCircleOutline) // 圆形勾选图标，表示按钮
    case 'API':
      return markRaw(ServerOutline) // 服务器图标，表示API接口
    default:
      return markRaw(FolderOutline)
  }
}

/**
 * 将资源数据转换为树形数据格式
 */
function convertToTreeData(resources: Resource[]): TreeOption[] {
  return resources.map((resource) => {
    const option: TreeOption = {
      key: resource.resourceId,
      label: resource.resourceName,
      children: resource.children && resource.children.length > 0
        ? convertToTreeData(resource.children)
        : undefined,
      // 保存资源类型和图标组件，供 renderLabel 使用
      resourceType: resource.resourceType,
      iconComponent: getResourceTypeIcon(resource.resourceType)
    }
    return option
  })
}

/**
 * 自定义标签渲染函数，添加图标前缀
 */
function renderLabel({ option }: { option: TreeOption & { resourceType?: string; iconComponent?: any } }) {
  const IconComponent = option.iconComponent || markRaw(GridOutline)
  
  return h('span', { style: { display: 'flex', alignItems: 'center' } }, [
    h(NIcon, { size: 16, style: { marginRight: '6px', flexShrink: 0 } }, {
      default: () => h(IconComponent)
    }),
    h('span', option.label as string)
  ])
}

/**
 * 提取已授权的资源ID（checked为true的资源）
 */
function extractCheckedKeys(resources: Resource[]): string[] {
  const keys: string[] = []
  
  function traverse(items: Resource[]) {
    for (const item of items) {
      // 如果资源已授权，添加到选中列表
      if ((item as any).checked === true) {
        keys.push(item.resourceId)
      }
      // 递归处理子资源
      if (item.children && item.children.length > 0) {
        traverse(item.children)
      }
    }
  }
  
  traverse(resources)
  return keys
}

/**
 * 提取所有资源ID（用于默认展开）
 */
function extractAllKeys(resources: Resource[]): string[] {
  const keys: string[] = []
  
  function traverse(items: Resource[]) {
    for (const item of items) {
      keys.push(item.resourceId)
      if (item.children && item.children.length > 0) {
        traverse(item.children)
      }
    }
  }
  
  traverse(resources)
  return keys
}

/**
 * 从树形数据中提取所有节点的 key（用于展开）
 */
function extractAllKeysFromTreeData(treeData: TreeOption[]): string[] {
  const keys: string[] = []
  
  function traverse(items: TreeOption[]) {
    for (const item of items) {
      keys.push(item.key as string)
      if (item.children && item.children.length > 0) {
        traverse(item.children)
      }
    }
  }
  
  traverse(treeData)
  return keys
}

/**
 * 处理选中节点变化
 * Naive UI 的 cascade 和 check-strategy 属性已经自动处理级联选择
 * 父节点选中时，所有子节点自动选中；父节点取消时，所有子节点自动取消
 */
function handleCheckedKeysChange(keys: string[]) {
  // 直接更新选中状态，级联逻辑由 Naive UI 的 cascade 属性自动处理
  checkedKeys.value = keys
  currentCheckedKeys.value = keys
}

/**
 * 保存角色授权
 */
async function handleSave() {
  if (!props.roleId) {
    message.warning('角色ID不能为空')
    return
  }

  loading.value = true
  try {
    // 将资源ID数组转换为逗号分割的字符串
    const resourceIdsString = currentCheckedKeys.value.join(',')
    
    const response = await saveRoleResources({
      roleId: props.roleId,
      resourceIds: resourceIdsString,
      permissionType: 'ALLOW'
    })
    
    // 直接检查原始响应，而不是解析后的数据
    if (isApiSuccess(response)) {
      message.success('保存成功')
      emit('success')
      handleClose()
    } else {
      message.error(getApiMessage(response) || '保存失败')
    }
  } catch (error: any) {
    message.error(error.message || '保存失败')
  } finally {
    loading.value = false
  }
}

/**
 * 取消操作
 */
function handleCancel() {
  handleClose()
}

/**
 * 关闭抽屉
 */
function handleClose() {
  emit('update:show', false)
  emit('close')
}
</script>

<style scoped lang="scss">
.role-resource-drawer {
  width: 100%;
  height: 100%;
  min-height: 400px;
  display: flex;
  flex-direction: column;

  .search-container {
    margin-bottom: 12px;
    flex-shrink: 0;
  }

  :deep(.n-spin-container) {
    min-height: 400px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  :deep(.g-tree-wrapper) {
    flex: 1;
    overflow: auto;
  }
}
</style>

