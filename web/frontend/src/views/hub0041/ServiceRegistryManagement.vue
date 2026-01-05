<template>
  <div class="service-registry-management">
    <!-- 列表页面 -->
    <div v-show="currentView === 'list'" class="list-view">

      <!-- 搜索和筛选 -->
      <div class="search-section">
        <n-space :size="16" justify="space-between">
          <n-space :size="16">
            <n-input 
              v-model:value="searchParams.serviceName"
              :placeholder="t('searchServiceName')"
              clearable
              @clear="handleSearch"
              @keyup.enter="handleSearch"
              style="width: 240px"
            >
              <template #prefix>
                <n-icon><SearchOutline /></n-icon>
              </template>
            </n-input>
            
            <n-input 
              v-model:value="searchParams.groupName"
              :placeholder="t('search.placeholder.groupName')"
              clearable
              @clear="handleSearch"
              @keyup.enter="handleSearch"
              style="width: 180px"
            >
              <template #prefix>
                <n-icon><FolderOutline /></n-icon>
              </template>
            </n-input>
            
            <n-select
              v-model:value="searchParams.protocolType"
              :options="protocolOptions"
              :placeholder="t('selectProtocol')"
              clearable
              style="width: 140px"
              @update:value="handleSearch"
            />
            
            <n-select
              v-model:value="searchParams.activeFlag"
              :options="statusOptions"
              :placeholder="t('selectStatus')"
              clearable
              style="width: 120px"
              @update:value="handleSearch"
            />
            
            <n-button @click="handleSearch" type="primary">
              {{ t('actions.search') }}
            </n-button>
            
            <n-button @click="resetSearch">
              {{ t('actions.reset') }}
            </n-button>
            
            <n-button 
              type="primary" 
              @click="refreshServices"
              :loading="loading"
            >
              <template #icon>
                <n-icon><RefreshOutline /></n-icon>
              </template>
              {{ t('actions.refresh') }}
            </n-button>
          </n-space>
          
          <n-button 
            type="primary" 
            @click="handleAddService"
          >
            <template #icon>
              <n-icon><AddOutline /></n-icon>
            </template>
            {{ t('actions.add') }}
          </n-button>
        </n-space>
      </div>

      <!-- 服务卡片列表 -->
      <div class="services-section">
        <n-spin :show="loading">
          <template v-if="services.length > 0">
            <n-grid :cols="4" :x-gap="16" :y-gap="16">
              <n-gi v-for="service in services" :key="service.serviceName">
                <ServiceCard 
                  :service="service" 
                  @view-detail="handleViewDetail"
                  @edit="handleEditService"
                  @refresh="handleRefreshService"
                  @delete="confirmDeleteService"
                  @view-events="handleViewEvents"
                />
              </n-gi>
            </n-grid>
          </template>
          
          <template v-else-if="!loading">
            <n-empty
              :description="t('noServices')"
              style="padding: 60px 0;"
            />
          </template>
        </n-spin>
      </div>

      <!-- 服务详情抽屉 -->
      <ServiceDetailDrawer
        v-model:show="detailDrawerVisible"
        :service-name="selectedServiceName"
      />
    </div>

    <!-- 表单页面 -->
    <div v-show="currentView === 'form'" class="form-view">
      <ServiceFormPage
        v-if="currentView === 'form'"
        :edit-data="editingService"
        @back="handleBackToList"
        @success="handleFormSuccess"
      />
    </div>

    <!-- 事件日志页面 -->
    <div v-show="currentView === 'events'" class="events-view">
      <div class="events-header">
        <n-space justify="space-between" align="center">
          <div class="events-title">
            <n-button text @click="handleBackToList">
              <template #icon>
                <n-icon><ArrowBackOutline /></n-icon>
              </template>
              {{ t('actions.back') }}
            </n-button>
            <n-divider vertical />
            <h3>{{ t('serviceEventLog') }} - {{ selectedServiceForEvents }}</h3>
          </div>
        </n-space>
      </div>
      <ServiceEventPage
        v-if="currentView === 'events'"
        :service-id="selectedServiceForEvents"
      />
    </div>
    
    <!-- 分页 - 容器底部 -->
    <div v-show="currentView === 'list'" class="bottom-pagination">
      <n-pagination v-bind="naiveConfig" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { 
  NButton, NIcon, NSpace, NInput, NSelect, NGrid, NGi, 
  NSpin, NPagination, NEmpty, NDivider, useDialog, useMessage
} from 'naive-ui'
import { RefreshOutline, SearchOutline, AddOutline, ArrowBackOutline, FolderOutline } from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { useServiceRegistry } from './hooks'

import ServiceCard from './components/ServiceCard.vue'
import ServiceDetailDrawer from './components/ServiceDetailDrawer.vue'
import ServiceFormPage from './components/ServiceFormPage.vue'
import ServiceEventPage from './components/ServiceEventPage.vue'
import type { Service, ProtocolType } from './types'

// 国际化
const { t } = useModuleI18n('hub0041')

// 消息提示
const message = useMessage()

// 对话框
const dialog = useDialog()



// 服务注册管理
const {
  services,
  loading,
  searchParams,
  naiveConfig,
  totalCount,
  fetchServices,
  refreshService,
  handleSearch,
  handleReset,
  handleRefresh,
  handleDeleteService
} = useServiceRegistry()

// 协议类型选项
const protocolOptions = [
  { label: 'HTTP', value: 'HTTP' },
  { label: 'HTTPS', value: 'HTTPS' },
  { label: 'TCP', value: 'TCP' },
  { label: 'UDP', value: 'UDP' },
  { label: 'GRPC', value: 'GRPC' }
]

// 状态选项
const statusOptions = [
  { label: t('status.Y'), value: 'Y' },
  { label: t('status.N'), value: 'N' }
]

// 页面视图状态
const currentView = ref<'list' | 'form' | 'events'>('list')

// 详情抽屉
const detailDrawerVisible = ref(false)
const selectedServiceName = ref('')

// 表单页面
const editingService = ref<Service | null>(null)

// 事件日志页面
const selectedServiceForEvents = ref('')

// 初始化
onMounted(() => {
  fetchServices()
})

// 重新映射方法名以保持向后兼容
const resetSearch = handleReset
const refreshServices = handleRefresh

// 查看详情
const handleViewDetail = (service: Service) => {
  selectedServiceName.value = service.serviceName
  detailDrawerVisible.value = true
}

// 刷新单个服务
const handleRefreshService = async (serviceName: string) => {
  await refreshService(serviceName)
}



// 新增服务
const handleAddService = () => {
  editingService.value = null
  currentView.value = 'form'
}

// 编辑服务
const handleEditService = (service: Service) => {
  editingService.value = service
  currentView.value = 'form'
}

// 查看服务事件日志
const handleViewEvents = (service: Service) => {
  selectedServiceForEvents.value = service.serviceName
  currentView.value = 'events'
}

// 返回列表页面
const handleBackToList = () => {
  currentView.value = 'list'
  editingService.value = null
  selectedServiceForEvents.value = ''
}

// 表单操作成功
const handleFormSuccess = () => {
  // 刷新服务列表
  fetchServices()
}

// 删除服务
const confirmDeleteService = (service: Service) => {
  dialog.warning({
    title: t('actions.delete'),
    content: t('messages.confirmDeleteService'),
    positiveText: t('actions.confirm'),
    negativeText: t('actions.cancel'),
    onPositiveClick: async () => {
      await handleDeleteService(service.serviceName)
    }
  })
}
</script>

<style scoped lang="scss">
.service-registry-management {
  min-height: 100vh;
  display: flex;
  flex-direction: column;

  .list-view, .form-view, .events-view {
    width: 100%;
    flex: 1;
  }

  .events-view {
    .events-header {
      background: var(--card-color);
      padding: 16px;
      border-radius: 8px;
      margin-bottom: 24px;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
      
      .events-title {
        display: flex;
        align-items: center;
        gap: 12px;
        
        h3 {
          margin: 0;
          color: var(--text-color-1);
          font-size: 18px;
          font-weight: 600;
        }
      }
    }
  }

  .search-section {
    background: var(--card-color);
    padding: 16px;
    border-radius: 8px;
    margin-bottom: 24px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  }



  .services-section {
    margin-bottom: 20px;
  }
  
  .bottom-pagination {
    display: flex;
    justify-content: center;
    padding: 16px 0;
    margin-top: auto; /* 将分页推到底部 */
    margin-bottom: 20px;
    background-color: var(--card-color);
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  }
}
</style>
