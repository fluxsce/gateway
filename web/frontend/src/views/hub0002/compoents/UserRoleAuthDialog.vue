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
    <div class="user-role-auth-wrap">
      <RsLoading :loading="loading" overlay block size="lg" />
      <div class="user-role-auth-content">
        <RsCard class="user-info" variant="outlined">
          <template #header>
            <div class="info-header">
              <GIcon :size="18" color="success">
                <PersonOutline />
              </GIcon>
              <span>用户信息</span>
            </div>
          </template>

          <RsDescriptions :columns="2" size="sm" label-placement="left" bordered>
            <RsDescriptionsItem label="用户ID">
              {{ currentUser?.userId }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="用户名">
              {{ currentUser?.userName }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="真实姓名">
              {{ currentUser?.realName }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="状态">
              <RsTag
                :variant="currentUser?.statusFlag === 'Y' ? 'success' : 'danger'"
                size="sm"
              >
                {{ currentUser?.statusFlag === 'Y' ? '启用' : '禁用' }}
              </RsTag>
            </RsDescriptionsItem>
          </RsDescriptions>
        </RsCard>

        <RsCard class="role-selection" variant="outlined">
          <template #header>
            <div class="config-header">
              <GIcon :size="18" color="warning">
                <PeopleCircleOutline />
              </GIcon>
              <span>角色选择</span>
            </div>
          </template>

          <div class="role-selection-body">
            <RsInput
              v-model="roleSearchKeyword"
              placeholder="搜索角色名称或描述"
              clearable
              size="sm"
            >
              <template #prefix>
                <GIcon :size="16">
                  <SearchOutline />
                </GIcon>
              </template>
            </RsInput>

            <div class="role-list-container">
              <RsTree
                v-if="!loading && treeData.length > 0"
                :nodes="treeData"
                v-model:checked-keys="checkedKeys"
                v-model:expanded-keys="expandedKeys"
                :filter="roleSearchKeyword"
                checkable
                check-strictly
                :selectable="false"
                show-line
                block-node
                virtual
                height="100%"
              />
              <RsEmpty
                v-else-if="!loading && treeData.length === 0"
                description="暂无角色数据"
              />
            </div>
          </div>
        </RsCard>
      </div>
    </div>
  </GModal>
</template>

<script lang="ts" setup>
import { GIcon } from '@/components/gicon'
import { GModal } from '@/components/gmodal'
import { useAppMessage } from '@/composables/useAppMessage'
import {
  RsCard,
  RsDescriptions,
  RsDescriptionsItem,
  RsEmpty,
  RsInput,
  RsLoading,
  RsTag,
  RsTree,
} from '@/ui'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { PeopleCircleOutline, PersonOutline, SearchOutline } from '@vicons/ionicons5'
import { computed, onMounted, ref, watch } from 'vue'
import * as userApi from '../api'
import type { User } from '../types'

/** RsTree 节点结构 */
interface RoleTreeNode {
  key: string
  label: string
  children?: RoleTreeNode[]
}

/** 角色树节点（接口返回结构） */
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
  user: undefined,
})

const emit = defineEmits<Emits>()

const message = useAppMessage()

const dialogVisible = computed({
  get: () => props.visible,
  set: (value: boolean) => emit('update:visible', value),
})

const loading = ref(false)
const saving = ref(false)
const roleSearchKeyword = ref('')
const currentUser = ref<User | undefined>(props.user)

const treeData = ref<RoleTreeNode[]>([])
const expandedKeys = ref<string[]>([])
const checkedKeys = ref<string[]>([])

const handleClose = () => {
  emit('update:visible', false)
  emit('close')
}

/**
 * 加载用户角色列表（含全部角色与已授权勾选状态）。
 */
async function loadUserRoles() {
  if (!props.userId) return

  loading.value = true
  try {
    const response = await userApi.getUserRoles(props.userId)

    if (isApiSuccess(response)) {
      const roles = parseJsonData<RoleItem[]>(response, [])

      treeData.value = convertToTreeData(roles)
      checkedKeys.value = extractCheckedKeys(roles)
      expandedKeys.value = extractAllKeys(roles)
    } else {
      message.error(getApiMessage(response) || '加载角色列表失败')
    }
  } catch (error: unknown) {
    const errMsg = error instanceof Error ? error.message : '加载角色列表失败'
    message.error(errMsg)
  } finally {
    loading.value = false
  }
}

/**
 * 将角色列表转换为 RsTree 节点结构。
 */
function convertToTreeData(roles: RoleItem[]): RoleTreeNode[] {
  return roles.map((role) => ({
    key: role.roleId,
    label: role.roleName,
    children:
      role.children && role.children.length > 0
        ? convertToTreeData(role.children)
        : undefined,
  }))
}

/**
 * 提取已授权角色 ID（checked === true）。
 */
function extractCheckedKeys(roles: RoleItem[]): string[] {
  const keys: string[] = []

  function traverse(items: RoleItem[]) {
    for (const item of items) {
      if (item.checked === true) {
        keys.push(item.roleId)
      }
      if (item.children && item.children.length > 0) {
        traverse(item.children)
      }
    }
  }

  traverse(roles)
  return keys
}

/**
 * 提取全部角色 ID，用于默认展开。
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
 * 保存用户角色授权。
 */
async function handleSave() {
  if (!props.userId) {
    message.error('用户ID不能为空')
    return
  }

  if (checkedKeys.value.length === 0) {
    message.warning('请至少选择一个角色')
    return
  }

  saving.value = true
  try {
    const roleIdsString = checkedKeys.value.join(',')

    const response = await userApi.assignUserRoles({
      userId: props.userId,
      roleIds: roleIdsString,
    })

    if (isApiSuccess(response)) {
      message.success('保存成功')
      emit('saved')
      await loadUserRoles()
      emit('update:visible', false)
    } else {
      message.error(getApiMessage(response) || '保存失败')
    }
  } catch (error: unknown) {
    const errMsg = error instanceof Error ? error.message : '保存失败'
    message.error(errMsg)
  } finally {
    saving.value = false
  }
}

watch(
  () => props.visible,
  (newVal) => {
    if (newVal && props.userId) {
      currentUser.value = props.user
      loadUserRoles()
    } else if (!newVal) {
      treeData.value = []
      checkedKeys.value = []
      expandedKeys.value = []
      roleSearchKeyword.value = ''
    }
  },
  { immediate: true },
)

watch(
  () => props.userId,
  (newVal) => {
    if (newVal && props.visible) {
      loadUserRoles()
    }
  },
)

onMounted(() => {
  if (props.visible && props.userId) {
    loadUserRoles()
  }
})
</script>

<style scoped lang="scss">
.user-role-auth-wrap {
  position: relative;
  min-height: 200px;
}

.user-role-auth-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.user-info,
.role-selection {
  .info-header,
  .config-header {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 500;
  }
}

.role-selection-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.role-list-container {
  border: 1px solid var(--rs-border);
  border-radius: var(--rs-radius-sm, 4px);
  padding: 12px;
  background-color: var(--rs-surface);
  height: 400px;
  overflow: hidden;
}
</style>
