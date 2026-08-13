<template>
  <div class="tool-marketplace">
    <!-- 工具市场头部 -->
    <div class="marketplace-header">
      <div class="header-title">
        <GIcon size="24" class="title-icon">
          <Apps />
        </GIcon>
        <div class="title-content">
          <h2>{{ t('title') }}</h2>
          <p class="subtitle">发现和安装强大的工具来提升你的工作效率</p>
        </div>
      </div>

      <div class="header-actions">
        <RsInput
          v-model="searchQuery"
          class="marketplace-search"
          :placeholder="t('searchPlaceholder')"
          size="lg"
          clearable
          radius="full"
        >
          <template #prefix>
            <GIcon size="18">
              <Search />
            </GIcon>
          </template>
        </RsInput>

        <div class="view-mode-group">
          <RsButton
            :variant="viewMode === 'grid' ? 'primary' : 'ghost'"
            size="sm"
            @click="setViewMode('grid')"
          >
            <template #default>
              <GIcon size="18">
                <Grid />
              </GIcon>
              网格视图
            </template>
          </RsButton>
          <RsButton
            :variant="viewMode === 'list' ? 'primary' : 'ghost'"
            size="sm"
            @click="setViewMode('list')"
          >
            <template #default>
              <GIcon size="18">
                <List />
              </GIcon>
              列表视图
            </template>
          </RsButton>
        </div>
      </div>
    </div>

    <!-- 工具分类标签（仅导航，内容区自行按 activeCategory 过滤） -->
    <div class="category-section">
      <RsTabs
        v-model="activeCategoryModel"
        :items="categoryTabItems"
        variant="card"
        size="md"
        panelless
      />
    </div>

    <!-- 工具统计信息 -->
    <div class="stats-bar" v-if="!loading">
      <span class="stats-text">
        找到 <strong class="stats-count">{{ filteredTools.length }}</strong> 个工具
      </span>
      <div class="stats-actions">
        <RsButton variant="ghost" @click="refreshTools">
          <GIcon>
            <Search />
          </GIcon>
          刷新
        </RsButton>
      </div>
    </div>

    <!-- 工具列表 -->
    <div class="tools-section">
      <RsLoading v-if="loading" block size="lg" />

      <template v-else>
        <div class="tools-grid" v-if="viewMode === 'grid' && filteredTools.length > 0">
          <tool-card
            v-for="tool in filteredTools"
            :key="tool.id"
            :tool="tool"
            @install="handleInstallTool"
            @uninstall="handleUninstallTool"
            @configure="handleConfigureTool"
            @preview="handlePreviewTool"
          />
        </div>

        <div class="tools-list" v-else-if="viewMode === 'list' && filteredTools.length > 0">
          <tool-list-item
            v-for="tool in filteredTools"
            :key="tool.id"
            :tool="tool"
            @install="handleInstallTool"
            @uninstall="handleUninstallTool"
            @configure="handleConfigureTool"
            @preview="handlePreviewTool"
          />
        </div>

        <div v-else class="empty-state">
          <RsEmpty :description="t('noToolsFound')" fill>
            <template #icon>
              <GIcon size="64" color="var(--text-color-disabled)">
                <Apps />
              </GIcon>
            </template>
            <RsButton v-if="searchQuery" @click="clearSearch">
              清除搜索条件
            </RsButton>
          </RsEmpty>
        </div>
      </template>
    </div>

    <!-- 工具预览抽屉 -->
    <tool-preview-drawer
      v-model:show="previewDrawerVisible"
      :tool="selectedTool"
      @install="handleInstallTool"
      @configure="handleConfigureTool"
    />

    <!-- 工具配置对话框 -->
    <tool-config-dialog
      v-model:show="configDialogVisible"
      :tool="selectedTool"
      @save="handleSaveToolConfig"
    />
  </div>
</template>

<script setup lang="ts">
import GIcon from '@/components/gicon/GIcon.vue'
import { RsButton, RsEmpty, RsInput, RsLoading, RsTabs } from '@/ui'
import { Apps, Search, Grid, List } from '@vicons/ionicons5'
import { computed } from 'vue'
import ToolCard from './modules/ToolCard.vue'
import ToolListItem from './modules/ToolListItem.vue'
import ToolPreviewDrawer from './modules/ToolPreviewDrawer.vue'
import ToolConfigDialog from './modules/ToolConfigDialog.vue'
import { useToolMarketplace } from '@/views/layout/hooks/useToolMarketplace'
import { useModuleI18n } from '@/hooks/useModuleI18n'

const { t } = useModuleI18n('toolMarket')

const {
  filteredTools,
  categories,
  selectedTool,
  loading,
  searchQuery,
  activeCategory,
  viewMode,
  previewDrawerVisible,
  configDialogVisible,
  setViewMode,
  handleInstallTool,
  handleUninstallTool,
  handleConfigureTool,
  handlePreviewTool,
  handleSaveToolConfig,
  initializeMarketplace,
} = useToolMarketplace()

const categoryTabItems = computed(() =>
  categories.value.map((category) => ({
    value: category.key,
    label: category.label,
  })),
)

/** RsTabs v-model 为 string，与 ToolCategory 枚举对齐 */
const activeCategoryModel = computed({
  get: () => activeCategory.value as string,
  set: (value: string) => {
    activeCategory.value = value as typeof activeCategory.value
  },
})

const clearSearch = () => {
  searchQuery.value = ''
}

const refreshTools = () => {
  initializeMarketplace()
}

initializeMarketplace()
</script>

<style lang="scss" scoped>
.tool-marketplace {
  padding: 8px;

  .marketplace-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 24px 0;
    margin-bottom: 24px;
    border-bottom: 1px solid var(--border-color);

    .header-title {
      display: flex;
      align-items: flex-start;
      gap: 16px;

      .title-icon {
        color: var(--primary-color);
        margin-top: 4px;
      }

      .title-content {
        h2 {
          margin: 0 0 8px 0;
          font-size: 24px;
          font-weight: 700;
          color: var(--text-color-primary);
          line-height: 1.2;
        }

        .subtitle {
          margin: 0;
          font-size: 14px;
          color: var(--text-color-secondary);
          line-height: 1.4;
        }
      }
    }

    .header-actions {
      display: flex;
      align-items: center;
      gap: 20px;
      flex-shrink: 0;
    }

    .marketplace-search {
      width: 300px;
    }

    .view-mode-group {
      display: inline-flex;
      gap: 8px;
    }
  }

  .category-section {
    margin-bottom: 24px;
  }

  .stats-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 0;
    margin-bottom: 20px;
    border-bottom: 1px solid var(--border-color-light, rgba(0, 0, 0, 0.06));

    .stats-text {
      font-size: 15px;
      color: var(--text-color-secondary);
    }

    .stats-count {
      color: var(--primary-color);
      font-weight: 600;
    }

    .stats-actions {
      display: flex;
      gap: 12px;
    }
  }

  .tools-section {
    position: relative;
    min-height: 500px;
    max-height: 65vh;
    overflow-y: auto;
    padding: 12px 8px;
    border-radius: 8px;
    background-color: var(--bg-color-hover, rgba(0, 0, 0, 0.02));

    .tools-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: 20px;
      padding: 8px;
    }

    .tools-list {
      display: flex;
      flex-direction: column;
      gap: 12px;
      padding: 8px;
    }

    .empty-state {
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 300px;
      padding: 40px;
    }

    &::-webkit-scrollbar {
      width: 10px;
    }

    &::-webkit-scrollbar-track {
      background: var(--bg-color-container);
      border-radius: 5px;
    }

    &::-webkit-scrollbar-thumb {
      background: var(--scrollbar-color, #d9d9d9);
      border-radius: 5px;
      border: 2px solid var(--bg-color-container);

      &:hover {
        background: var(--scrollbar-hover-color, #bfbfbf);
      }
    }
  }
}

@media (max-width: 1024px) {
  .tool-marketplace {
    .marketplace-header {
      flex-direction: column;
      gap: 20px;
      align-items: stretch;

      .header-actions {
        justify-content: space-between;
        flex-wrap: wrap;
        gap: 16px;
      }

      .marketplace-search {
        flex: 1;
        min-width: 240px;
        width: auto;
      }
    }

    .tools-section {
      max-height: 55vh;

      .tools-grid {
        grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
        gap: 16px;
      }
    }
  }
}

@media (max-width: 768px) {
  .tool-marketplace {
    padding: 4px;

    .marketplace-header {
      padding: 16px 0;
      margin-bottom: 16px;

      .header-title {
        .title-content {
          h2 {
            font-size: 20px;
          }

          .subtitle {
            font-size: 13px;
          }
        }
      }

      .header-actions {
        gap: 12px;
      }

      .marketplace-search {
        min-width: 200px;
      }
    }

    .category-section {
      margin-bottom: 16px;
    }

    .stats-bar {
      padding: 12px 0;
      margin-bottom: 16px;
    }

    .tools-section {
      max-height: 50vh;

      .tools-grid {
        grid-template-columns: 1fr;
        gap: 12px;
      }
    }
  }
}

@media (max-width: 480px) {
  .tool-marketplace {
    .marketplace-header {
      .header-actions {
        flex-direction: column;
        align-items: stretch;

        .marketplace-search {
          width: 100%;
          min-width: unset;
        }

        .view-mode-group {
          width: 100%;

          :deep(.rs-btn) {
            flex: 1;
          }
        }
      }
    }
  }
}
</style>
