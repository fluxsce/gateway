<template>
    <div class="tool-marketplace">
        <!-- 工具市场头部 -->
        <div class="marketplace-header">
            <div class="header-title">
                <n-icon size="24" class="title-icon">
                    <Apps />
                </n-icon>
                <div class="title-content">
                    <h2>{{ t('title') }}</h2>
                    <p class="subtitle">发现和安装强大的工具来提升你的工作效率</p>
                </div>
            </div>

            <div class="header-actions">
                <!-- 搜索框 -->
                <n-input v-model:value="searchQuery" :placeholder="t('searchPlaceholder')" style="width: 300px;"
                    size="large" clearable round @input="handleSearch">
                    <template #prefix>
                        <n-icon size="18">
                            <Search />
                        </n-icon>
                    </template>
                </n-input>

                <!-- 视图切换 -->
                <n-button-group size="large">
                    <n-button :type="viewMode === 'grid' ? 'primary' : 'default'" @click="setViewMode('grid')"
                        :ghost="viewMode !== 'grid'">
                        <template #icon>
                            <n-icon size="18">
                                <Grid />
                            </n-icon>
                        </template>
                        网格视图
                    </n-button>
                    <n-button :type="viewMode === 'list' ? 'primary' : 'default'" @click="setViewMode('list')"
                        :ghost="viewMode !== 'list'">
                        <template #icon>
                            <n-icon size="18">
                                <List />
                            </n-icon>
                        </template>
                        列表视图
                    </n-button>
                </n-button-group>
            </div>
        </div>

        <!-- 工具分类标签 -->
        <div class="category-section">
            <n-tabs v-model:value="activeCategory" type="card" size="large" @update:value="handleCategoryChange"
                animated>
                <n-tab-pane v-for="category in categories" :key="category.key" :name="category.key"
                    :tab="category.label" />
            </n-tabs>
        </div>

        <!-- 工具统计信息 -->
        <div class="stats-bar" v-if="!loading">
            <n-text class="stats-text">
                找到 <n-text strong type="primary">{{ filteredTools.length }}</n-text> 个工具
            </n-text>
            <div class="stats-actions">
                <n-button text @click="refreshTools">
                    <template #icon>
                        <n-icon>
                            <Search />
                        </n-icon>
                    </template>
                    刷新
                </n-button>
            </div>
        </div>

        <!-- 工具列表 -->
        <div class="tools-section">
            <n-spin :show="loading" size="large">
                <div class="tools-grid" v-if="viewMode === 'grid'">
                    <tool-card v-for="tool in filteredTools" :key="tool.id" :tool="tool" @install="handleInstallTool"
                        @uninstall="handleUninstallTool" @configure="handleConfigureTool"
                        @preview="handlePreviewTool" />
                </div>

                <div class="tools-list" v-else>
                    <tool-list-item v-for="tool in filteredTools" :key="tool.id" :tool="tool"
                        @install="handleInstallTool" @uninstall="handleUninstallTool" @configure="handleConfigureTool"
                        @preview="handlePreviewTool" />
                </div>

                <!-- 空状态 -->
                <div v-if="!loading && filteredTools.length === 0" class="empty-state">
                    <n-empty :description="t('noToolsFound')" size="large">
                        <template #icon>
                            <n-icon size="64" color="var(--text-color-disabled)">
                                <Apps />
                            </n-icon>
                        </template>
                        <template #extra>
                            <n-button @click="clearSearch" v-if="searchQuery">
                                清除搜索条件
                            </n-button>
                        </template>
                    </n-empty>
                </div>
            </n-spin>
        </div>

        <!-- 工具预览抽屉 -->
        <tool-preview-drawer v-model:show="previewDrawerVisible" :tool="selectedTool" @install="handleInstallTool"
            @configure="handleConfigureTool" />

        <!-- 工具配置对话框 -->
        <tool-config-dialog v-model:show="configDialogVisible" :tool="selectedTool" @save="handleSaveToolConfig" />
    </div>
</template>

<script setup lang="ts">
import { Apps, Search, Grid, List } from '@vicons/ionicons5'
import ToolCard from './modules/ToolCard.vue'
import ToolListItem from './modules/ToolListItem.vue'
import ToolPreviewDrawer from './modules/ToolPreviewDrawer.vue'
import ToolConfigDialog from './modules/ToolConfigDialog.vue'
import { useToolMarketplace } from '@/views/layout/hooks/useToolMarketplace'
import { useModuleI18n } from '@/hooks/useModuleI18n'

// 国际化
const { t } = useModuleI18n('toolMarket')

// 使用工具市场Hook
const {
    filteredTools,
    categories,
    selectedTool,
    loading,

    // 界面状态
    searchQuery,
    activeCategory,
    viewMode,
    previewDrawerVisible,
    configDialogVisible,

    // 方法
    setViewMode,
    handleSearch,
    handleCategoryChange,
    handleInstallTool,
    handleUninstallTool,
    handleConfigureTool,
    handlePreviewTool,
    handleSaveToolConfig,

    // 初始化
    initializeMarketplace
} = useToolMarketplace()

// 清除搜索条件
const clearSearch = () => {
    searchQuery.value = ''
}

// 刷新工具列表
const refreshTools = () => {
    initializeMarketplace()
}

// 初始化
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
    }

    .category-section {
        margin-bottom: 24px;

        :deep(.n-tabs) {
            .n-tabs-nav {
                margin-bottom: 0;
            }

            .n-tabs-tab {
                padding: 12px 20px;
                font-weight: 500;
            }

            .n-tab-pane {
                padding: 0;
            }
        }
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

        .stats-actions {
            display: flex;
            gap: 12px;
        }
    }

    .tools-section {
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

        // 优化滚动条样式
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

// 响应式设计
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

                .n-input {
                    flex: 1;
                    min-width: 240px;
                }
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

                .n-input {
                    min-width: 200px;
                }
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

                .n-input {
                    width: 100%;
                    min-width: unset;
                }

                .n-button-group {
                    width: 100%;

                    .n-button {
                        flex: 1;
                    }
                }
            }
        }
    }
}
</style>