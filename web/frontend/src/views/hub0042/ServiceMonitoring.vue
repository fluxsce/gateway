<template>
  <div class="service-monitoring-page">
    <div class="content-wrapper">
      <n-tabs v-model:value="mainTab" type="line" animated>
        <!-- 实例列表 -->
        <n-tab-pane name="list" :tab="t('jvmInstanceList')">
          <jvm-resource-list @select="handleSelectJvm" />
        </n-tab-pane>

        <!-- 监控详情（选中实例后显示） -->
        <n-tab-pane 
          v-if="selectedJvmResourceId" 
          name="detail" 
          :tab="t('monitoringDetail')"
        >
          <!-- 实例信息头部 -->
          <n-card :bordered="false" class="detail-header">
            <n-space align="center">
              <n-tag type="info" size="large">
                {{ currentJvmInfo?.applicationName || t('unknownApplication') }}
              </n-tag>
              <n-divider vertical />
              <span>{{ t('host') }}: {{ currentJvmInfo?.hostName || '-' }}</span>
              <n-divider vertical />
              <span>{{ t('ip') }}: {{ currentJvmInfo?.hostIpAddress || '-' }}</span>
              <n-divider vertical />
              <n-tag 
                :type="currentJvmInfo?.healthyFlag === 'Y' ? 'success' : 'error'"
                size="medium"
              >
                {{ currentJvmInfo?.healthyFlag === 'Y' ? t('healthy') : t('unhealthy') }}
              </n-tag>
              <n-divider vertical />
              <n-button text type="primary" @click="handleBackToList">
                <template #icon>
                  <n-icon :component="ArrowBackOutline" />
                </template>
                {{ t('backToList') }}
              </n-button>
            </n-space>
          </n-card>

          <!-- 监控类型标签页 -->
          <n-tabs v-model:value="monitorType" type="card" animated style="margin-top: 16px">
            <!-- JVM监控 -->
            <n-tab-pane name="jvm" :tab="t('jvmMonitoring')">
              <n-tabs v-model:value="jvmDetailTab" type="line" animated>
                <!-- 内存监控 -->
                <n-tab-pane name="memory" :tab="t('memoryMonitor')">
                  <memory-monitor :jvm-resource-id="selectedJvmResourceId" />
                </n-tab-pane>

                <!-- GC监控 -->
                <n-tab-pane name="gc" :tab="t('gcMonitor')">
                  <gc-monitor :jvm-resource-id="selectedJvmResourceId" />
                </n-tab-pane>

                <!-- 线程监控 -->
                <n-tab-pane name="thread" :tab="t('threadMonitor')">
                  <thread-monitor :jvm-resource-id="selectedJvmResourceId" />
                </n-tab-pane>
              </n-tabs>
            </n-tab-pane>

            <!-- 应用组件监控 -->
            <n-tab-pane name="component" :tab="t('componentMonitoring')">
              <app-monitor :jvm-resource-id="selectedJvmResourceId" />
            </n-tab-pane>
          </n-tabs>
        </n-tab-pane>
      </n-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ArrowBackOutline } from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import JvmResourceList from './components/JvmResourceList.vue'
import MemoryMonitor from './components/MemoryMonitor.vue'
import GcMonitor from './components/GcMonitor.vue'
import ThreadMonitor from './components/ThreadMonitor.vue'
import AppMonitor from './components/AppMonitor.vue'
import { useJvmResourceManagement } from './hooks'
import type { JvmResource } from './types'

// 使用模块化多语言
const { t } = useModuleI18n('hub0042')

const { getJvmResource } = useJvmResourceManagement()

// 主标签页（列表 / 详情）
const mainTab = ref('list')
// 监控类型（JVM / 应用组件）
const monitorType = ref('jvm')
// JVM监控详情子标签
const jvmDetailTab = ref('memory')

// 选中的JVM资源
const selectedJvmResourceId = ref<string>('')
const currentJvmInfo = ref<JvmResource | null>(null)

// 选择JVM实例
const handleSelectJvm = async (jvmResource: JvmResource) => {
  selectedJvmResourceId.value = jvmResource.jvmResourceId
  currentJvmInfo.value = jvmResource
  mainTab.value = 'detail'
  monitorType.value = 'jvm'
  jvmDetailTab.value = 'memory'
}

// 返回列表
const handleBackToList = () => {
  mainTab.value = 'list'
  selectedJvmResourceId.value = ''
  currentJvmInfo.value = null
}

// 加载JVM详情（预留）
const loadJvmDetail = async (jvmResourceId: string) => {
  const data = await getJvmResource(jvmResourceId)
  if (data) {
    currentJvmInfo.value = data
  }
}
</script>

<style scoped lang="scss">
.service-monitoring-page {

  .content-wrapper {
    margin-top: 16px;
  }

  .detail-header {
    margin-bottom: 16px;
    
   
  }
}
</style>

