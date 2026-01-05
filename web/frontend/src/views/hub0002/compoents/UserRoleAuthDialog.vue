<template>
  <GModal
    v-model:visible="dialogVisible"
    title="用户角色授权"
    preset="dialog"
    width="800px"
    :mask-closable="false"
    :closable="true"
    :show-footer="true"
    :show-cancel="true"
    :show-confirm="true"
    cancel-text="取消"
    confirm-text="保存"
    :confirm-loading="saving"
    @close="handleClose"
    @cancel="handleClose"
    @confirm="handleSave"
  >
    <n-spin :show="loading">
      <div class="user-role-auth-content">
        <!-- 用户信息 -->
        <n-card size="small" class="user-info">
          <template #header>
            <div class="info-header">
              <n-icon size="18" color="#18a058">
                <PersonOutline />
              </n-icon>
              <span>用户信息</span>
            </div>
          </template>

          <n-descriptions :column="2" size="small">
            <n-descriptions-item label="用户ID">
              {{ currentUser?.userId }}
            </n-descriptions-item>
            <n-descriptions-item label="用户名">
              {{ currentUser?.userName }}
            </n-descriptions-item>
            <n-descriptions-item label="真实姓名">
              {{ currentUser?.realName }}
            </n-descriptions-item>
            <n-descriptions-item label="状态">
              <n-tag :type="currentUser?.statusFlag === 'Y' ? 'success' : 'error'" size="small">
                {{ currentUser?.statusFlag === 'Y' ? '启用' : '禁用' }}
              </n-tag>
            </n-descriptions-item>
          </n-descriptions>
        </n-card>

        <!-- 角色选择区域 -->
        <n-card size="small" class="role-selection">
          <template #header>
            <div class="config-header">
              <n-icon size="18" color="#f0a020">
                <PeopleCircleOutline />
              </n-icon>
              <span>角色选择</span>
            </div>
          </template>

          <n-space vertical :size="16">
            <!-- 角色搜索 -->
            <n-input
              v-model:value="roleSearchKeyword"
              placeholder="搜索角色名称或描述"
              clearable
              @input="handleRoleSearch"
            >
              <template #prefix>
                <n-icon><SearchOutline /></n-icon>
              </template>
            </n-input>

            <!-- 角色树 -->
            <div class="role-list-container">
              <GTree
                v-if="!loading && filteredTreeData.length > 0"
                :data="filteredTreeData"
                :checkable="true"
                :cascade="false"
                :check-strategy="'all'"
                :show-line="true"
                :default-expanded-keys="defaultExpandedKeys"
                :checked-keys="checkedKeys"
                :virtual-scroll="true"
                @update:checkedKeys="handleCheckedKeysChange"
              />
              <n-empty v-else-if="!loading && filteredTreeData.length === 0" description="暂无角色数据" />
            </div>
          </n-space>
        </n-card>

      </div>
    </n-spin>
  </GModal>
</template>

<script lang="ts" setup>
import { GModal } from '@/components/gmodal'
import { GTree } from '@/components/gtree'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { PeopleCircleOutline, PersonOutline, SearchOutline } from '@vicons/ionicons5'
import type { TreeOption } from 'naive-ui'
import { NEmpty, useMessage } from 'naive-ui'
import { computed, onMounted, ref, watch } from 'vue'
import * as userApi from '../api'
import type { User } from '../types'

interface RoleItem {
  roleId: string
  roleName: string
  roleDescription: string
  roleStatus: string
  builtInFlag: string
  checked: boolean
  children?: RoleItem[]
}

interface Props {
  visible: boolean
  userId?: string
  user?: User
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'saved'): void
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  userId: '',
  user: undefined
})

const emit = defineEmits<Emits>()

const message = useMessage()

// 对话框显示状态（计算属性，用于 v-model:visible）
const dialogVisible = computed({
  get: () => props.visible,
  set: (value: boolean) => emit('update:visible', value)
})

// 状态
const loading = ref(false)
const saving = ref(false)
const roleSearchKeyword = ref('')
const currentUser = ref<User | undefined>(props.user)

// 树形数据
const treeData = ref<TreeOption[]>([])

// 默认展开的节点
const defaultExpandedKeys = ref<string[]>([])

// 选中的节点（已授权的角色）
const checkedKeys = ref<string[]>([])

// 当前选中的节点（用于保存）
const currentCheckedKeys = ref<string[]>([])

// 原始角色数据
const allRoles = ref<RoleItem[]>([])

// 计算属性 - 过滤后的树形数据
const filteredTreeData = computed(() => {
  if (!roleSearchKeyword.value) {
    return treeData.value
  }
  const keyword = roleSearchKeyword.value.toLowerCase()
  return filterTreeData(treeData.value, keyword)
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

const handleRoleSearch = () => {
  // 搜索逻辑已在计算属性中处理
}

const handleClose = () => {
  emit('update:visible', false)
  emit('close')
}

/**
 * 处理选中节点变化
 */
function handleCheckedKeysChange(keys: string[]) {
  checkedKeys.value = keys
  currentCheckedKeys.value = keys
}

/**
 * 加载用户角色列表（包含所有角色和选中状态）
 */
async function loadUserRoles() {
  if (!props.userId) return

  loading.value = true
  try {
    const response = await userApi.getUserRoles(props.userId)

    if (isApiSuccess(response)) {
      // parseJsonData 会从 response.bizData 中解析 JSON 字符串，直接返回 RoleItem[] 数组
      const roles = parseJsonData<RoleItem[]>(response, [])
      
      // 保存原始数据
      allRoles.value = roles
      
      // 转换为树形数据
      treeData.value = convertToTreeData(roles)
      
      // 提取已授权的角色ID（checked为true的角色）
      const authorizedIds = extractCheckedKeys(roles)
      checkedKeys.value = authorizedIds
      currentCheckedKeys.value = [...authorizedIds]
      
      // 默认展开所有节点
      defaultExpandedKeys.value = extractAllKeys(roles)
    } else {
      message.error(getApiMessage(response) || '加载角色列表失败')
    }
  } catch (error: any) {
    message.error(error.message || '加载角色列表失败')
  } finally {
    loading.value = false
  }
}

/**
 * 将角色数据转换为树形数据格式
 */
function convertToTreeData(roles: RoleItem[]): TreeOption[] {
  return roles.map((role) => {
    const option: TreeOption = {
      key: role.roleId,
      label: role.roleName,
      children: role.children && role.children.length > 0
        ? convertToTreeData(role.children)
        : undefined
    }
    return option
  })
}

/**
 * 提取已授权的角色ID（checked为true的角色）
 */
function extractCheckedKeys(roles: RoleItem[]): string[] {
  const keys: string[] = []
  
  function traverse(items: RoleItem[]) {
    for (const item of items) {
      // 如果角色已授权，添加到选中列表
      if (item.checked === true) {
        keys.push(item.roleId)
      }
      // 递归处理子角色
      if (item.children && item.children.length > 0) {
        traverse(item.children)
      }
    }
  }
  
  traverse(roles)
  return keys
}

/**
 * 提取所有角色ID（用于默认展开）
 */
function extractAllKeys(roles: RoleItem[]): string[] {
  const keys: string[] = []
  
  function traverse(items: RoleItem[]) {
    for (const item of items) {
      keys.push(item.roleId)
      if (item.children && item.children.length > 0) {
        traverse(item.children)
      }
    }
  }
  
  traverse(roles)
  return keys
}

/**
 * 保存用户角色授权
 */
async function handleSave() {
  if (!props.userId) {
    message.error('用户ID不能为空')
    return
  }

  if (currentCheckedKeys.value.length === 0) {
    message.warning('请至少选择一个角色')
    return
  }

  saving.value = true
  try {
    // 将角色ID数组转换为逗号分割的字符串
    const roleIdsString = currentCheckedKeys.value.join(',')
    
    const response = await userApi.assignUserRoles({
      userId: props.userId,
      roleIds: roleIdsString
    })
    
    // 直接检查原始响应，而不是解析后的数据
    if (isApiSuccess(response)) {
    message.success('保存成功')
    emit('saved')
    await loadUserRoles()
    emit('update:visible', false)
    } else {
      message.error(getApiMessage(response) || '保存失败')
    }
  } catch (error: any) {
    message.error(error.message || '保存失败')
  } finally {
    saving.value = false
  }
}

// 监听visible变化
watch(
  () => props.visible,
  (newVal) => {
    if (newVal && props.userId) {
      currentUser.value = props.user
      loadUserRoles()
    } else if (!newVal) {
      // 关闭时重置数据
      treeData.value = []
      checkedKeys.value = []
      currentCheckedKeys.value = []
      defaultExpandedKeys.value = []
      roleSearchKeyword.value = ''
      allRoles.value = []
    }
  },
  { immediate: true }
)

// 监听userId变化
watch(
  () => props.userId,
  (newVal) => {
    if (newVal && props.visible) {
      loadUserRoles()
    }
  }
)

onMounted(() => {
  if (props.visible && props.userId) {
    loadUserRoles()
  }
})
</script>

<style scoped lang="scss">
.user-role-auth-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.user-info,
.role-selection,
.assigned-roles {
  .info-header,
  .config-header {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 500;
  }
}

.role-list-container {
  border: 1px solid var(--n-border-color);
  border-radius: 4px;
  padding: 12px;
  background-color: var(--n-color);
  min-height: 400px;
  max-height: 500px;
  overflow: auto;
}
</style>

