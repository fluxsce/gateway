<template>
  <RsDrawer
    :open="show"
    title="工具预览"
    side="right"
    size="lg"
    @update:open="(val: boolean) => $emit('update:show', val)"
  >
    <div v-if="tool" class="tool-preview">
      <div class="preview-header">
        <div class="tool-icon">
          <GIcon size="48" :icon="tool.icon || 'apps-outline'" />
        </div>
        <div class="tool-info">
          <h2>{{ tool.displayName || tool.name }}</h2>
          <p>{{ tool.description }}</p>
        </div>
      </div>

      <RsDivider />

      <div class="preview-content">
        <RsCard title="基本信息">
          <RsDescriptions :columns="2" :items="basicInfoItems" />
        </RsCard>

        <RsCard title="标签" class="tags-card">
          <div class="tag-list">
            <RsTag v-for="tag in tool.tags" :key="tag" variant="primary" size="sm">
              {{ tag }}
            </RsTag>
          </div>
        </RsCard>
      </div>

      <div class="preview-actions">
        <RsButton variant="primary" @click="$emit('install', tool)">
          安装
        </RsButton>
        <RsButton variant="default" @click="$emit('configure', tool)">
          配置
        </RsButton>
      </div>
    </div>
  </RsDrawer>
</template>

<script setup lang="ts">
import GIcon from '@/components/gicon/GIcon.vue'
import {
  RsButton,
  RsCard,
  RsDescriptions,
  RsDivider,
  RsDrawer,
  RsTag,
  type RsDescriptionsItemData,
} from '@/ui'
import { computed } from 'vue'
import type { Tool } from '../../types/toolMarketplace'

interface Props {
  show: boolean
  tool?: Tool | null
}

const props = defineProps<Props>()

defineEmits<{
  'update:show': [value: boolean]
  install: [tool: Tool]
  configure: [tool: Tool]
}>()

const basicInfoItems = computed<RsDescriptionsItemData[]>(() => {
  const tool = props.tool
  if (!tool) return []
  return [
    { label: '版本', value: tool.version },
    { label: '作者', value: tool.author },
    { label: '大小', value: `${tool.size ?? 0}KB` },
    { label: '状态', value: tool.status },
  ]
})
</script>

<style lang="scss" scoped>
.tool-preview {
  .preview-header {
    display: flex;
    gap: 16px;
    align-items: center;

    .tool-icon {
      color: var(--primary-color);
    }

    .tool-info {
      h2 {
        margin: 0 0 8px 0;
        color: var(--text-color-primary);
      }

      p {
        margin: 0;
        color: var(--text-color-secondary);
      }
    }
  }

  .preview-content {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin: 16px 0;
  }

  .tag-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .preview-actions {
    display: flex;
    gap: 12px;
    margin-top: 24px;
    padding-top: 16px;
    border-top: 1px solid var(--border-color);
  }
}
</style>
