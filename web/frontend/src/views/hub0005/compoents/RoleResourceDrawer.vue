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

        <!-- 资源树 -->
        <GTree
          v-else
          :data="treeData"
          :checkable="true"
          :cascade="true"
          :check-strategy="'all'"
          :show-line="true"
          :default-expanded-keys="defaultExpandedKeys"
          :checked-keys="checkedKeys"
          :virtual-scroll="true"
          @update:checkedKeys="handleCheckedKeysChange"
        />
      </div>
    </template>
  </GDrawer>
</template>

<script setup lang="ts">
import { GDrawer } from '@/components/gdrawer'
import { GTree } from '@/components/gtree'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import type { Resource } from '@/views/hub0006/types'
import type { TreeOption } from 'naive-ui'
import { NSpin, useMessage } from 'naive-ui'
import { computed, ref, watch } from 'vue'
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
      
      // 默认展开所有节点
      defaultExpandedKeys.value = extractAllKeys(resources)
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
 * 将资源数据转换为树形数据格式
 */
function convertToTreeData(resources: Resource[]): TreeOption[] {
  return resources.map((resource) => {
    const option: TreeOption = {
      key: resource.resourceId,
      label: resource.resourceName,
      children: resource.children && resource.children.length > 0
        ? convertToTreeData(resource.children)
        : undefined
    }
    return option
  })
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

  :deep(.n-spin-container) {
    min-height: 400px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
</style>

