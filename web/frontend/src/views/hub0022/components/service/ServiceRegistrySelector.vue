<template>
  <g-modal
    v-model:visible="visible"
    title="选择注册服务"
    :header-icon="SearchOutline"
    :width="1200"
    :mask-closable="false"
    :show-fullscreen-toggle="false"
    to="#hub0022-service-definition-list"
    @close="handleClose"
    @cancel="handleClose"
    @confirm="confirmSelection"
    :show-confirm="true"
    :show-cancel="true"
    confirm-text="确认选择"
    cancel-text="取消"
  >

    <!-- 搜索区域 -->
    <div class="service-search-section">
      <div class="search-form">
        <div class="search-row">
          <n-input 
            v-model:value="searchText"
            placeholder="搜索服务名称"
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
            v-model:value="searchGroupName"
            placeholder="搜索分组名称"
            clearable
            @clear="handleSearch"
            @keyup.enter="handleSearch"
            style="width: 240px"
          >
            <template #prefix>
              <n-icon><SearchOutline /></n-icon>
            </template>
          </n-input>
          <div class="search-actions">
            <n-button @click="handleSearch" type="primary">
              搜索
            </n-button>
            <n-button @click="resetSearch">
              重置
            </n-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 当前选择的服务 -->
    <div v-if="selectedService" class="selected-service-section">
      <div class="section-title">当前选择的服务</div>
      <div class="selected-service-card">
        <div class="service-info">
          <div class="service-name">{{ selectedService.serviceName }}</div>
          <div class="service-meta">
            <n-tag type="info" size="small">{{ selectedService.groupName }}</n-tag>
            <n-tag type="success" size="small">{{ selectedService.protocolType }}</n-tag>
            <span class="service-path">{{ selectedService.contextPath }}</span>
          </div>
        </div>
        <n-button text type="error" @click="clearSelection">
          <template #icon>
            <n-icon><CloseOutline /></n-icon>
          </template>
          取消选择
        </n-button>
      </div>
    </div>

    <!-- 服务列表 -->
    <div class="services-list-section">
      <div class="section-header">
        <div class="section-title">可选择的服务</div>
        <div class="section-info" v-if="pageInfo && pageInfo.totalCount > 0">
          共 {{ pageInfo.totalCount }} 个服务
        </div>
      </div>
      
      <n-spin :show="loading">
        <template v-if="services.length > 0">
          <div class="service-cards-container">
            <div 
              v-for="service in services" 
              :key="service.serviceName"
              class="service-card"
              :class="{ selected: selectedService?.serviceName === service.serviceName }"
              @click="selectService(service)"
            >
              <div class="card-header">
                <div class="card-title-row">
                  <n-radio 
                    :checked="selectedService?.serviceName === service.serviceName"
                    @click.stop
                    class="card-radio"
                  />
                  <div class="service-name">{{ service.serviceName }}</div>
                  <div class="service-status">
                    <n-tag 
                      :type="service.activeFlag === 'Y' ? 'success' : 'error'" 
                      size="small"
                      round
                    >
                      {{ service.activeFlag === 'Y' ? '启用' : '禁用' }}
                    </n-tag>
                  </div>
                </div>
              </div>
              
              <div class="card-content">
                <div class="service-info-grid">
                  <div class="info-item">
                    <span class="info-label">分组</span>
                    <n-tag type="info" size="small">
                      {{ service.groupName || 'DEFAULT' }}
                    </n-tag>
                  </div>
                  <div class="info-item">
                    <span class="info-label">协议</span>
                    <n-tag type="primary" size="small">
                      {{ service.protocolType }}
                    </n-tag>
                  </div>
                  <div class="info-item">
                    <span class="info-label">路径</span>
                    <span class="info-value">{{ service.contextPath || '/' }}</span>
                  </div>
                  <div class="info-item">
                    <span class="info-label">实例数</span>
                    <span class="info-value">
                      {{ service.instances ? service.instances.length : '0' }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <!-- 分页组件 -->
          <div class="pagination-wrapper" v-if="pageInfo && pageInfo.totalCount > 0">
            <g-pagination 
              :page-info="pageInfo"
              :page-sizes="[10, 20, 50]"
              @page-change="handlePageChange"
            />
          </div>
        </template>
        
        <template v-else-if="!loading">
          <n-empty
            description="暂无可用的注册服务"
            style="padding: 40px 0;"
          />
        </template>
      </n-spin>
    </div>
  </g-modal>
</template>

<script setup lang="ts">
import { GModal } from '@/components/gmodal'
import { GPagination } from '@/components/gpage'
import type { PageInfoObj } from '@/types/api'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import {
  CloseOutline,
  SearchOutline
} from '@vicons/ionicons5'
import { NButton, NEmpty, NIcon, NInput, NRadio, NSpin, NTag, useMessage } from 'naive-ui'
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { registServiceQuery } from '../../api'

// Props
interface Props {
  show: boolean
  currentService?: any // 当前已选择的服务
}

const props = defineProps<Props>()

// Emits
const emit = defineEmits<{
  'update:show': [value: boolean]
  'select': [service: any]
}>()

// 响应式状态
const message = useMessage()
const loading = ref(false)
const services = ref<any[]>([])
const selectedService = ref<any | null>(null)
const searchText = ref('')
const searchGroupName = ref('')
const pageInfo = ref<PageInfoObj | undefined>(undefined)

// 计算属性
const visible = computed({
  get: () => props.show,
  set: (value: boolean) => emit('update:show', value)
})

// 监听弹窗显示状态
const stopShowWatch = watch(() => props.show, (show) => {
  if (show) {
    // 打开时设置当前选择的服务
    selectedService.value = props.currentService || null
    resetPagination()
    fetchServices()
  } else {
    // 关闭时重置状态
    selectedService.value = null
    searchText.value = ''
    searchGroupName.value = ''
    resetPagination()
  }
})

// ============= 资源清理 =============

// 组件卸载时清理所有监听器
onBeforeUnmount(() => {
  stopShowWatch()
})

// 重置分页
function resetPagination() {
  pageInfo.value = undefined
}

// 获取服务列表
async function fetchServices(customParams: any = {}) {
  try {
    loading.value = true
    
    // 合并分页参数和搜索参数
    const requestParams: any = {
      tenantId: 'default',
      pageIndex: pageInfo.value?.pageIndex || 1,
      pageSize: pageInfo.value?.pageSize || 10,
      ...customParams
    }
    
    const response = await registServiceQuery(requestParams)
    
    if (isApiSuccess(response)) {
      // 使用format工具类解析业务数据
      const serviceList = parseJsonData<any[]>(response, [])
      services.value = Array.isArray(serviceList) ? serviceList : []
      
      // 解析分页信息
      try {
        const parsedPageInfo = parsePageInfo(response)
        pageInfo.value = parsedPageInfo
      } catch (error) {
        console.warn('解析分页信息失败:', error)
        // 如果分页信息解析失败，创建默认分页信息
        pageInfo.value = {
          pageIndex: requestParams.pageIndex || 1,
          pageSize: requestParams.pageSize || 10,
          totalCount: serviceList?.length || 0,
          totalPageIndex: Math.ceil((serviceList?.length || 0) / (requestParams.pageSize || 10)),
          baseData: '',
          curPageCount: serviceList?.length || 0,
          dbsId: '',
          mainKey: '',
          orderByList: '',
          otherData: '',
          paramObjectsJson: '',
          timeTypeFieldNames: '',
        } as PageInfoObj
      }
      
      console.log('获取服务列表成功:', services.value.length, '总数:', pageInfo.value?.totalCount)
    } else {
      const errorMsg = getApiMessage(response, '获取服务列表失败')
      message.error(errorMsg)
      services.value = []
      pageInfo.value = undefined
    }
  } catch (error: any) {
    console.error('获取服务列表失败:', error)
    message.error('获取服务列表失败: ' + (error?.message || '未知错误'))
    services.value = []
    pageInfo.value = undefined
  } finally {
    loading.value = false
  }
}

// 处理分页变化
function handlePageChange({ currentPage, pageSize }: { currentPage: number; pageSize: number }) {
  fetchServices({ pageIndex: currentPage, pageSize })
}

// 搜索服务
function handleSearch() {
  const searchParams: any = {}
  if (searchText.value) {
    searchParams.serviceName = searchText.value
  }
  if (searchGroupName.value) {
    searchParams.groupName = searchGroupName.value
  }
  
  // 重置到第一页并搜索
  resetPagination()
  fetchServices({ ...searchParams, pageIndex: 1, pageSize: 10 })
}

// 重置搜索
function resetSearch() {
  searchText.value = ''
  searchGroupName.value = ''
  resetPagination()
  fetchServices({ pageIndex: 1, pageSize: 10 })
}

// 选择服务
function selectService(service: any) {
  selectedService.value = service
  console.log('选择服务:', service.serviceName)
}

// 清除选择
function clearSelection() {
  selectedService.value = null
}

// 确认选择
function confirmSelection() {
  if (!selectedService.value) {
    message.warning('请选择一个服务')
    return
  }
  
  emit('select', selectedService.value)
  message.success(`已选择服务：${selectedService.value.serviceName}`)
  handleClose()
}

// 关闭弹窗
function handleClose() {
  emit('update:show', false)
}
</script>

<style scoped lang="scss">

.service-search-section {
  padding: 16px;
  border-bottom: 1px solid var(--n-divider-color);
  margin-bottom: 16px;

  .search-form {
    .search-row {
      display: flex;
      gap: 16px;
      align-items: center;
      flex-wrap: wrap;
    }

    .search-actions {
      display: flex;
      gap: 12px;
      margin-left: auto;
    }
  }
}

.selected-service-section {
  margin-bottom: 16px;
  padding: 0 16px;

  .section-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--n-text-color-1);
    margin-bottom: 12px;
  }

  .selected-service-card {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px;
    background: var(--n-color-hover);
    border-radius: 6px;
    border: 2px solid var(--n-color-primary);

    .service-info {
      flex: 1;

      .service-name {
        font-weight: 600;
        color: var(--n-text-color-1);
        margin-bottom: 4px;
        font-size: 14px;
      }

      .service-meta {
        display: flex;
        align-items: center;
        gap: 8px;

        .service-path {
          font-size: 12px;
          color: var(--n-text-color-3);
        }
      }
    }
  }
}

.services-list-section {
  padding: 0 16px;
  margin-bottom: 16px;

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;

    .section-title {
      font-size: 14px;
      font-weight: 600;
      color: var(--n-text-color-1);
      margin: 0;
    }

    .section-info {
      font-size: 12px;
      color: var(--n-text-color-3);
    }
  }

  .service-cards-container {
    max-height: 450px;
    overflow-y: auto;
    padding: 8px;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
    gap: 16px;

    .service-card {
      background: var(--n-card-color);
      border: 1px solid var(--n-border-color);
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.3s ease;
      overflow: hidden;

      &:hover {
        border-color: var(--n-color-primary);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        transform: translateY(-2px);
      }

      &.selected {
        border-color: var(--n-color-primary);
        box-shadow: 0 0 0 2px rgba(24, 160, 88, 0.2);
        
        .card-header {
          background: var(--n-color-primary-hover);
        }
      }

      .card-header {
        padding: 12px 16px;
        background: var(--n-color-hover);
        border-bottom: 1px solid var(--n-divider-color);

        .card-title-row {
          display: flex;
          align-items: center;
          gap: 12px;

          .card-radio {
            flex-shrink: 0;
          }

          .service-name {
            flex: 1;
            font-weight: 600;
            font-size: 14px;
            color: var(--n-text-color-1);
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }

          .service-status {
            flex-shrink: 0;
          }
        }
      }

      .card-content {
        padding: 16px;

        .service-info-grid {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 12px;

          .info-item {
            display: flex;
            align-items: center;
            gap: 8px;

            .info-label {
              font-size: 12px;
              color: var(--n-text-color-3);
              min-width: 40px;
            }

            .info-value {
              font-size: 12px;
              color: var(--n-text-color-2);
              font-family: monospace;
            }
          }
        }
      }
    }
  }

  .pagination-wrapper {
    display: flex;
    justify-content: center;
    padding: 16px 0;
    border-top: 1px solid var(--n-divider-color);
    margin-top: 16px;
  }
}
</style>

