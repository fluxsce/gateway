<template>
  <div class="config-management-wrapper">
    <!-- 配置列表页面 -->
    <ConfigManagementComponent
      v-show="currentView === 'config'"
      @view-history="handleViewHistory"
    />
    
    <!-- 配置历史页面 -->
    <ConfigHistoryPage
      v-show="currentView === 'history'"
      :initial-query="historyQuery"
      @back="handleBackToConfig"
    />
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import ConfigHistoryPage from './components/config-history/index.vue'
import ConfigManagementComponent from './components/config/index.vue'
import type { Config } from './types'

// 定义组件名称
defineOptions({
  name: 'ConfigManagement'
})

// ============= 页面状态 =============
const currentView = ref<'config' | 'history'>('config')
const historyQuery = ref<{
  namespaceId: string
  groupName: string
  configDataId: string
} | null>(null)

/**
 * 处理查看历史事件
 */
const handleViewHistory = (config: Config) => {
  // 填充查询条件并切换到历史页面
  historyQuery.value = {
    namespaceId: config.namespaceId,
    groupName: config.groupName || 'DEFAULT_GROUP',
    configDataId: config.configDataId,
  }
  currentView.value = 'history'
}

/**
 * 返回配置列表页面
 */
const handleBackToConfig = () => {
  currentView.value = 'config'
  historyQuery.value = null
}
</script>

<style scoped>
.config-management-wrapper {
  height: 100%;
  width: 100%;
  overflow: hidden;
  position: relative;
}

.config-management-wrapper > * {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}
</style>

