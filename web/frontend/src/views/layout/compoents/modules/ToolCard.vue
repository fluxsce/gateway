<template>
  <RsCard class="tool-card" elevated>
    <!-- 工具图标 -->
    <div class="tool-header">
      <div class="tool-icon">
        <GIcon v-if="tool.icon" size="24" :icon="tool.icon" />
        <GIcon v-else size="24">
          <Apps />
        </GIcon>
      </div>

      <!-- 工具状态标识 -->
      <div class="tool-status">
        <RsTag v-if="tool.status === ToolStatus.INSTALLED" variant="success" size="sm">
          已安装
        </RsTag>
        <RsTag v-else-if="tool.status === ToolStatus.INSTALLING" variant="info" size="sm">
          安装中
        </RsTag>
        <RsTag v-else variant="default" size="sm">
          可安装
        </RsTag>
      </div>
    </div>

    <!-- 工具信息 -->
    <div class="tool-content">
      <h4 class="tool-title">{{ tool.displayName || tool.name }}</h4>
      <p class="tool-description">{{ tool.description }}</p>

      <!-- 工具标签 -->
      <div class="tool-tags" v-if="tool.tags && tool.tags.length">
        <RsTag
          v-for="tag in tool.tags.slice(0, 2)"
          :key="tag"
          size="sm"
          variant="default"
          class="tag-item"
        >
          {{ tag }}
        </RsTag>
        <span v-if="tool.tags.length > 2" class="more-tags">
          +{{ tool.tags.length - 2 }}
        </span>
      </div>

      <!-- 工具元信息 -->
      <div class="tool-meta">
        <span class="meta-item">
          <GIcon size="12">
            <Person />
          </GIcon>
          {{ tool.author }}
        </span>
        <span class="meta-item">
          <GIcon size="12">
            <Code />
          </GIcon>
          v{{ tool.version }}
        </span>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="tool-actions">
      <RsButton
        v-if="tool.status === ToolStatus.AVAILABLE || tool.status === ToolStatus.INSTALLING"
        variant="primary"
        size="ssm"
        :loading="tool.status === ToolStatus.INSTALLING"
        @click="$emit('install', tool)"
      >
        <GIcon size="12">
          <Download />
        </GIcon>
        安装
      </RsButton>

      <RsButton
        v-else-if="tool.status === ToolStatus.INSTALLED"
        variant="danger"
        size="ssm"
        @click="$emit('uninstall', tool)"
      >
        <GIcon size="12">
          <Trash />
        </GIcon>
        卸载
      </RsButton>

      <RsButton
        v-if="tool.status === ToolStatus.INSTALLED"
        variant="default"
        size="ssm"
        @click="$emit('configure', tool)"
      >
        <GIcon size="12">
          <Settings />
        </GIcon>
        配置
      </RsButton>

      <RsButton variant="ghost" size="ssm" @click="$emit('preview', tool)">
        <GIcon size="12">
          <Eye />
        </GIcon>
        预览
      </RsButton>
    </div>
  </RsCard>
</template>

<script setup lang="ts">
import GIcon from '@/components/gicon/GIcon.vue'
import { RsButton, RsCard, RsTag } from '@/ui'
import { Apps, Person, Code, Download, Trash, Settings, Eye } from '@vicons/ionicons5'
import { ToolStatus, type Tool } from '../../types/toolMarketplace'

interface Props {
  tool: Tool
}

defineProps<Props>()

defineEmits<{
  install: [tool: Tool]
  uninstall: [tool: Tool]
  configure: [tool: Tool]
  preview: [tool: Tool]
}>()
</script>

<style lang="scss" scoped>
.tool-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  transition: transform 0.2s ease, box-shadow 0.2s ease;

  :deep(.rs-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 12px;
    box-sizing: border-box;
  }

  &:hover {
    transform: translateY(-1px);
  }

  .tool-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 8px;

    .tool-icon {
      color: var(--primary-color);
    }

    .tool-status {
      flex-shrink: 0;
    }
  }

  .tool-content {
    flex: 1;

    .tool-title {
      margin: 0 0 6px 0;
      font-size: 14px;
      font-weight: 600;
      color: var(--text-color-primary);
      line-height: 1.3;
    }

    .tool-description {
      margin: 0 0 8px 0;
      font-size: 12px;
      color: var(--text-color-secondary);
      line-height: 1.4;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }

    .tool-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 3px;
      margin-bottom: 8px;

      .tag-item {
        font-size: 11px;
      }

      .more-tags {
        font-size: 11px;
        color: var(--text-color-tertiary);
        align-self: center;
      }
    }

    .tool-meta {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-bottom: 10px;

      .meta-item {
        display: flex;
        align-items: center;
        gap: 3px;
        font-size: 11px;
        color: var(--text-color-tertiary);
      }
    }
  }

  .tool-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: auto;
  }
}

@media (max-width: 768px) {
  .tool-card {
    .tool-actions {
      flex-direction: column;

      :deep(.rs-btn) {
        width: 100%;
      }
    }
  }
}
</style>
