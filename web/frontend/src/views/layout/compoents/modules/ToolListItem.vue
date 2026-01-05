<template>
    <n-list-item class="tool-list-item">
        <div class="tool-item-content">
            <!-- 左侧图标和基本信息 -->
            <div class="tool-basic-info">
                <div class="tool-icon">
                    <n-icon v-if="tool.icon" size="32" :name="tool.icon" />
                    <n-icon v-else size="32">
                        <Apps />
                    </n-icon>
                </div>

                <div class="tool-info">
                    <h4 class="tool-title">{{ tool.displayName || tool.name }}</h4>
                    <p class="tool-description">{{ tool.description }}</p>

                    <!-- 工具元信息 -->
                    <div class="tool-meta">
                        <span class="meta-item">
                            <n-icon size="12">
                                <Person />
                            </n-icon>
                            {{ tool.author }}
                        </span>
                        <span class="meta-item">
                            <n-icon size="12"><Code /></n-icon>
                            v{{ tool.version }}
                        </span>
                        <span v-if="tool.size" class="meta-item">
                            <n-icon size="12">
                                <Archive />
                            </n-icon>
                            {{ formatSize(tool.size) }}
                        </span>
                    </div>
                </div>
            </div>

            <!-- 右侧状态和操作 -->
            <div class="tool-actions-area">
                <!-- 状态标签 -->
                <div class="tool-status">
                    <n-tag v-if="tool.status === ToolStatus.INSTALLED" type="success" size="small">
                        已安装
                    </n-tag>
                    <n-tag v-else-if="tool.status === ToolStatus.INSTALLING" type="warning" size="small">
                        安装中
                    </n-tag>
                    <n-tag v-else type="default" size="small">
                        可安装
                    </n-tag>
                </div>

                <!-- 操作按钮 -->
                <div class="tool-actions">
                    <n-button v-if="tool.status === ToolStatus.AVAILABLE || tool.status === ToolStatus.INSTALLING"
                        type="primary" size="small" :loading="tool.status === ToolStatus.INSTALLING"
                        @click="$emit('install', tool)">
                        安装
                    </n-button>

                    <n-button v-else-if="tool.status === ToolStatus.INSTALLED" type="error" size="small"
                        @click="$emit('uninstall', tool)">
                        卸载
                    </n-button>

                    <n-button v-if="tool.status === ToolStatus.INSTALLED" size="small"
                        @click="$emit('configure', tool)">
                        配置
                    </n-button>

                    <n-button size="small" @click="$emit('preview', tool)">
                        预览
                    </n-button>
                </div>
            </div>
        </div>
    </n-list-item>
</template>

<script setup lang="ts">
import { Apps, Person, Code, Archive } from '@vicons/ionicons5'
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

/**
 * 格式化文件大小
 */
const formatSize = (size: number): string => {
    if (size < 1024) return `${size}KB`
    if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)}MB`
    return `${(size / (1024 * 1024)).toFixed(1)}GB`
}
</script>

<style lang="scss" scoped>
.tool-list-item {
    .tool-item-content {
        display: flex;
        justify-content: space-between;
        align-items: center;
        width: 100%;
        padding: 8px 0;

        .tool-basic-info {
            display: flex;
            align-items: center;
            gap: 12px;
            flex: 1;

            .tool-icon {
                color: var(--primary-color);
                flex-shrink: 0;
            }

            .tool-info {
                flex: 1;
                min-width: 0;

                .tool-title {
                    margin: 0 0 4px 0;
                    font-size: 16px;
                    font-weight: 600;
                    color: var(--text-color-primary);
                    white-space: nowrap;
                    overflow: hidden;
                    text-overflow: ellipsis;
                }

                .tool-description {
                    margin: 0 0 8px 0;
                    font-size: 14px;
                    color: var(--text-color-secondary);
                    display: -webkit-box;
                    -webkit-line-clamp: 1;
                    -webkit-box-orient: vertical;
                    overflow: hidden;
                }

                .tool-meta {
                    display: flex;
                    gap: 16px;

                    .meta-item {
                        display: flex;
                        align-items: center;
                        gap: 4px;
                        font-size: 12px;
                        color: var(--text-color-tertiary);
                    }
                }
            }
        }

        .tool-actions-area {
            display: flex;
            align-items: center;
            gap: 12px;
            flex-shrink: 0;

            .tool-status {
                // 状态标签样式已由n-tag处理
            }

            .tool-actions {
                display: flex;
                gap: 8px;
            }
        }
    }
}

@media (max-width: 768px) {
    .tool-list-item {
        .tool-item-content {
            flex-direction: column;
            align-items: flex-start;
            gap: 12px;

            .tool-actions-area {
                width: 100%;
                justify-content: space-between;
            }
        }
    }
}
</style>