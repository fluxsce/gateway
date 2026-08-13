<template>
  <span class="g-ellipsis" :title="text">
    <slot>{{ text }}</slot>
  </span>
</template>

<script setup lang="ts">
defineOptions({ name: 'GEllipsis' })
withDefaults(defineProps<{ text?: string }>(), { text: '' })
</script>

<style scoped>
.g-ellipsis {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
}
</style>
