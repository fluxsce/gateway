<template>
  <n-modal
    v-model:show="visible"
    class="service-group-selection-dialog"
    :mask-closable="false"
    preset="dialog"
    :title="t('selectServiceGroup')"
    :style="{ width: '800px' }"
  >
    <div class="dialog-content">
      <!-- 搜索区域 -->
      <div class="search-section">
        <n-space :size="12">
          <n-input
            v-model:value="searchKeyword"
            :placeholder="t('searchGroupName')"
            clearable
            style="width: 240px"
            @input="handleSearch"
          >
            <template #prefix>
              <n-icon><SearchOutline /></n-icon>
            </template>
          </n-input>
          <n-select
            v-model:value="searchGroupType"
            :placeholder="t('selectGroupType')"
            :options="groupTypeOptions"
            clearable
            style="width: 140px"
            @update:value="handleSearch"
          />
          <n-button @click="handleRefresh" :loading="loading">
            <template #icon>
              <n-icon><RefreshOutline /></n-icon>
            </template>
            {{ t('actions.refresh') }}
          </n-button>
        </n-space>
      </div>

      <!-- 分组列表 -->
      <div class="group-list">
        <n-spin :show="loading">
          <n-scrollbar style="max-height: 400px">
            <n-empty v-if="filteredGroups.length === 0 && !loading" :description="t('noServiceGroups')" />
            <div v-else class="group-items">
              <div
                v-for="group in filteredGroups"
                :key="group.serviceGroupId"
                class="group-item"
                :class="{ selected: selectedGroup?.serviceGroupId === group.serviceGroupId }"
                @click="handleSelectGroup(group)"
              >
                <div class="group-header">
                  <div class="group-name">
                    <!-- 选中状态标识 -->
                    <n-icon 
                      v-if="selectedGroup?.serviceGroupId === group.serviceGroupId"
                      color="#18a058" 
                      size="18"
                      style="margin-right: 8px"
                    >
                      <CheckmarkCircleOutline />
                    </n-icon>
                    <n-text strong>{{ group.groupName }}</n-text>
                    <n-tag
                      :type="getGroupTypeTagType(group.groupType)"
                      size="small"
                      style="margin-left: 8px"
                    >
                      {{ getGroupTypeLabel(group.groupType) }}
                    </n-tag>
                  </div>
                  <n-tag
                    :type="group.activeFlag === 'Y' ? 'success' : 'error'"
                    size="small"
                  >
                    {{ group.activeFlag === 'Y' ? t('status.Y') : t('status.N') }}
                  </n-tag>
                </div>
                <div class="group-description">
                  <n-text depth="3">{{ group.groupDescription || t('noDescription') }}</n-text>
                </div>
              </div>
            </div>
          </n-scrollbar>
        </n-spin>
      </div>


    </div>

    <template #action>
      <n-space>
        <n-button @click="handleCancel">
          {{ t('cancel') }}
        </n-button>
        <n-button type="primary" @click="handleConfirm" :disabled="!selectedGroup">
          {{ t('actions.confirm') }}
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { 
  NModal, NInput, NSelect, NButton, NSpace, NIcon, NSpin, NScrollbar,
  NEmpty, NText, NTag, NAlert, useMessage 
} from 'naive-ui'
import { SearchOutline, RefreshOutline, CheckmarkCircleOutline } from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { getServiceGroups } from '../api'

// 定义 ServiceGroup 接口
interface ServiceGroup {
  serviceGroupId: string
  tenantId: string
  groupName: string
  groupDescription: string
  groupType: string
  ownerUserId: string
  adminUserIds?: string
  readUserIds?: string
  accessControlEnabled: string
  defaultProtocolType: string
  defaultLoadBalanceStrategy: string
  defaultHealthCheckUrl: string
  defaultHealthCheckIntervalSeconds: number
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: string
  noteText?: string
  extProperty?: string
  reserved1?: string
  reserved2?: string
  reserved3?: string
  reserved4?: string
  reserved5?: string
  reserved6?: string
  reserved7?: string
  reserved8?: string
  reserved9?: string
  reserved10?: string
}

interface Props {
  show: boolean
  selectedGroupId?: string
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'confirm', group: ServiceGroup): void
}

const props = withDefaults(defineProps<Props>(), {
  selectedGroupId: ''
})
const emit = defineEmits<Emits>()

// 国际化
const { t } = useModuleI18n('hub0041')

// 消息提示
const message = useMessage()

// 控制显示
const visible = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

// 数据状态
const loading = ref(false)
const serviceGroups = ref<ServiceGroup[]>([])
const selectedGroup = ref<ServiceGroup | null>(null)

// 搜索状态
const searchKeyword = ref('')
const searchGroupType = ref('')

// 分组类型选项
const groupTypeOptions = [
  { label: t('groupType.BUSINESS'), value: 'BUSINESS' },
  { label: t('groupType.SYSTEM'), value: 'SYSTEM' },
  { label: t('groupType.TEST'), value: 'TEST' }
]

// 过滤后的分组列表
const filteredGroups = computed(() => {
  let filtered = serviceGroups.value

  // 按关键词搜索
  if (searchKeyword.value.trim()) {
    const keyword = searchKeyword.value.toLowerCase()
    filtered = filtered.filter(group =>
      group.groupName.toLowerCase().includes(keyword) ||
      group.groupDescription?.toLowerCase().includes(keyword)
    )
  }

  // 按类型筛选
  if (searchGroupType.value) {
    filtered = filtered.filter(group => group.groupType === searchGroupType.value)
  }

  // 只显示活动状态的分组
  filtered = filtered.filter(group => group.activeFlag === 'Y')

  return filtered
})

// 监听选中的分组ID变化
watch(() => props.selectedGroupId, (groupId) => {
  if (groupId && serviceGroups.value.length > 0) {
    selectedGroup.value = serviceGroups.value.find(group => group.serviceGroupId === groupId) || null
  } else {
    selectedGroup.value = null
  }
}, { immediate: true })

// 监听对话框显示状态
watch(() => props.show, (show) => {
  if (show) {
    fetchServiceGroups()
  }
})

// 获取分组类型标签类型
const getGroupTypeTagType = (groupType: string) => {
  switch (groupType) {
    case 'BUSINESS': return 'success'
    case 'SYSTEM': return 'warning'
    case 'TEST': return 'info'
    default: return 'default'
  }
}

// 获取分组类型标签
const getGroupTypeLabel = (groupType: string) => {
  return t(`groupType.${groupType}`) || groupType
}

// 获取服务分组列表
const fetchServiceGroups = async () => {
  try {
    loading.value = true
    const response = await getServiceGroups()
    
    if (response.oK) {
      let responseData: any = {}
      try {
        responseData = typeof response.bizData === 'string' 
          ? JSON.parse(response.bizData) 
          : response.bizData
        
        serviceGroups.value = responseData || []
        
        // 如果有预选的分组ID，设置选中状态
        if (props.selectedGroupId) {
          selectedGroup.value = serviceGroups.value.find(
            group => group.serviceGroupId === props.selectedGroupId
          ) || null
        }
      } catch (error) {
        console.error('Failed to parse service groups data:', error)
        serviceGroups.value = []
      }
    } else {
      message.error(t('fetchServiceGroupsFailed'))
      serviceGroups.value = []
    }
  } catch (error) {
    console.error('Failed to fetch service groups:', error)
    message.error(t('fetchServiceGroupsFailed'))
    serviceGroups.value = []
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  // 搜索逻辑已在计算属性中处理
}

// 刷新列表
const handleRefresh = () => {
  fetchServiceGroups()
}

// 选择分组
const handleSelectGroup = (group: ServiceGroup) => {
  selectedGroup.value = group
}

// 确认选择
const handleConfirm = () => {
  if (selectedGroup.value) {
    emit('confirm', selectedGroup.value)
    visible.value = false
  }
}

// 取消操作
const handleCancel = () => {
  visible.value = false
}

// 初始化
onMounted(() => {
  if (props.show) {
    fetchServiceGroups()
  }
})
</script>

<style scoped lang="scss">
.service-group-selection-dialog {
  .dialog-content {
    padding: 4px;
  }

  .search-section {
    padding: 16px;
    border-bottom: 1px solid var(--border-color);
    margin-bottom: 16px;
  }

  .group-list {
    .group-items {
      .group-item {
        padding: 16px;
        border: 1px solid var(--border-color);
        border-radius: 8px;
        margin-bottom: 12px;
        cursor: pointer;
        transition: all 0.2s ease;

        &:hover {
          border-color: var(--primary-color);
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        }

        &.selected {
          border-color: #18a058;
          background-color: #f0f9f0;
          box-shadow: 0 2px 12px rgba(24, 160, 88, 0.15);
          transform: translateY(-1px);
        }

        .group-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 8px;

          .group-name {
            display: flex;
            align-items: center;
          }
        }

        .group-description {
          font-size: 13px;
        }
      }
    }
  }


}
</style>
