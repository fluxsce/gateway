<template>
  <RsDrawer
    v-model:open="drawerVisible"
    :title="drawerTitle"
    side="right"
    size="lg"
    :teleport-to="'#hub0005'"
    :show-overlay="false"
    :close-on-overlay-click="false"
  >
    <div class="role-resource-drawer">
      <RsLoading v-if="loading && !treeData.length" block size="lg" />

      <template v-else>
        <div class="search-container">
          <RsInput
            v-model="searchKeyword"
            :placeholder="t('role.auth.searchPlaceholder')"
            clearable
            size="sm"
          >
            <template #prefix>
              <RsIcon name="search" size="sm" />
            </template>
          </RsInput>
        </div>

        <div class="tree-container">
          <RsTree
            v-if="treeData.length > 0"
            :nodes="treeData"
            v-model:checked-keys="checkedKeys"
            v-model:expanded-keys="expandedKeys"
            :filter="searchKeyword"
            checkable
            :selectable="false"
            show-line
            block-node
            virtual
            height="100%"
            @check="handleCheck"
          />
          <RsEmpty v-else :description="t('role.auth.empty')" />
        </div>
      </template>
    </div>

    <template #footer>
      <RsButton variant="default" size="sm" @click="handleCancel">
        {{ t('common.action.cancel') }}
      </RsButton>
      <RsButton
        variant="primary"
        size="sm"
        :loading="loading"
        @click="handleSave"
      >
        {{ t('common.action.save') }}
      </RsButton>
    </template>
  </RsDrawer>
</template>

<script setup lang="ts">
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import {
  RsButton,
  RsDrawer,
  RsEmpty,
  RsIcon,
  RsInput,
  RsLoading,
  RsTree,
} from '@/ui'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import type { Resource } from '@/views/hub0006/types'
import type { RsTreeNode } from 'niuma-ui'
import { computed, ref, watch } from 'vue'
import { getRoleResources, saveRoleResources } from '../api'

defineOptions({
  name: 'RoleResourceDrawer',
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
  roleName: '',
})

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'close'): void
  (e: 'success'): void
}

const emit = defineEmits<Emits>()

const { t } = useModuleI18n('hub0005')
const message = useAppMessage()

const drawerVisible = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value),
})

const drawerTitle = computed(() => {
  if (props.roleName) {
    return t('role.auth.titleWithName', { name: props.roleName })
  }
  return t('role.auth.title')
})

const loading = ref(false)
const treeData = ref<RsTreeNode[]>([])
const expandedKeys = ref<string[]>([])
const checkedKeys = ref<string[]>([])
const currentCheckedKeys = ref<string[]>([])
const searchKeyword = ref('')

watch(
  () => props.show,
  (newShow) => {
    if (newShow && props.roleId) {
      loadRoleResources()
    } else if (!newShow) {
      treeData.value = []
      checkedKeys.value = []
      currentCheckedKeys.value = []
      expandedKeys.value = []
      searchKeyword.value = ''
    }
  },
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
      const resources = parseJsonData<Resource[]>(response, [])

      treeData.value = convertToTreeData(resources)

      const authorizedIds = extractCheckedKeys(resources)
      checkedKeys.value = authorizedIds
      currentCheckedKeys.value = [...authorizedIds]
      expandedKeys.value = []
    } else {
      message.error(getApiMessage(response) || t('role.auth.loadFailed'))
    }
  } catch (error: unknown) {
    const errMsg = error instanceof Error ? error.message : t('role.auth.loadFailed')
    message.error(errMsg)
  } finally {
    loading.value = false
  }
}

/**
 * 根据资源类型映射 Lucide 图标名
 */
function getResourceTypeIcon(resourceType: string): string {
  switch (resourceType) {
    case 'MODULE':
      return 'layout-grid'
    case 'GROUP':
      return 'folder'
    case 'MENU':
      return 'list'
    case 'BUTTON':
      return 'circle-check'
    case 'API':
      return 'server'
    default:
      return 'folder'
  }
}

/**
 * 将资源数据转换为 RsTree 节点
 */
function convertToTreeData(resources: Resource[]): RsTreeNode[] {
  return resources.map((resource) => {
    const node: RsTreeNode = {
      key: resource.resourceId,
      label: resource.resourceName,
      icon: getResourceTypeIcon(resource.resourceType),
      children:
        resource.children && resource.children.length > 0
          ? convertToTreeData(resource.children)
          : undefined,
    }
    return node
  })
}

/**
 * 提取已授权的资源ID（checked 为 true 的资源）
 */
function extractCheckedKeys(resources: Resource[]): string[] {
  const keys: string[] = []

  function traverse(items: Resource[]) {
    for (const item of items) {
      if ((item as Resource & { checked?: boolean }).checked === true) {
        keys.push(item.resourceId)
      }
      if (item.children && item.children.length > 0) {
        traverse(item.children)
      }
    }
  }

  traverse(resources)
  return keys
}

/**
 * 勾选变化时同步待保存的资源 ID
 */
function handleCheck(keys: string[]) {
  checkedKeys.value = keys
  currentCheckedKeys.value = keys
}

/**
 * 保存角色授权
 */
async function handleSave() {
  if (!props.roleId) {
    message.warning(t('role.auth.roleIdRequired'))
    return
  }

  loading.value = true
  try {
    const resourceIdsString = currentCheckedKeys.value.join(',')

    const response = await saveRoleResources({
      roleId: props.roleId,
      resourceIds: resourceIdsString,
      permissionType: 'ALLOW',
    })

    if (isApiSuccess(response)) {
      message.success(t('role.auth.saveSuccess'))
      emit('success')
      handleClose()
    } else {
      message.error(getApiMessage(response) || t('role.auth.saveFailed'))
    }
  } catch (error: unknown) {
    const errMsg = error instanceof Error ? error.message : t('role.auth.saveFailed')
    message.error(errMsg)
  } finally {
    loading.value = false
  }
}

function handleCancel() {
  handleClose()
}

function handleClose() {
  emit('update:show', false)
  emit('close')
}
</script>

<style scoped>
.role-resource-drawer {
  box-sizing: border-box;
  width: 100%;
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.search-container {
  flex-shrink: 0;
}

.tree-container {
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--rs-border);
  border-radius: var(--rs-radius-sm, 4px);
  padding: 8px;
  background-color: var(--rs-surface);
}

.tree-container :deep(.rs-tree) {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
}
</style>
