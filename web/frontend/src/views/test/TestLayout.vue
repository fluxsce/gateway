<template>
  <div class="test-layout">
    <router-view />
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'TestLayout' })
</script>

<style scoped lang="scss">
.test-layout {
  width: 100%;
  min-height: 0;
  overflow: auto;
}
</style>
