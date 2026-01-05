<template>
  <!-- This component doesn't render anything visible -->
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useLoadingBar, useDialog } from 'naive-ui'
import { initRequestTools } from '@/api/request'

// 组件挂载后初始化请求工具
onMounted(() => {
  try {
    // 获取loading bar实例
    const loadingBar = useLoadingBar()
    const dialog = useDialog()
    // 初始化请求工具
    if (loadingBar) {
      initRequestTools(loadingBar, dialog)
      console.log('请求工具已成功初始化')
    } else {
      console.warn('loadingBar不可用，请求工具初始化失败')
    }
  } catch (error) {
    console.error('RequestInitializer初始化失败:', error)
  }
})
</script>