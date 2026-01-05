<template>
    <n-card class="tool-card" hoverable>
        <!-- 工具图标 -->
        <div class="tool-header">
            <div class="tool-icon">
                <n-icon v-if="tool.icon" size="24" :name="tool.icon" />
                <n-icon v-else size="24">
                    <Apps />
                </n-icon>
            </div>

            <!-- 工具状态标识 -->
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
        </div>

        <!-- 工具信息 -->
        <div class="tool-content">
            <h4 class="tool-title">{{ tool.displayName || tool.name }}</h4>
            <p class="tool-description">{{ tool.description }}</p>

            <!-- 工具标签 -->
            <div class="tool-tags" v-if="tool.tags && tool.tags.length">
                <n-tag v-for="tag in tool.tags.slice(0, 2)" :key="tag" size="small" type="default" class="tag-item">
                    {{ tag }}
                </n-tag>
                <span v-if="tool.tags.length > 2" class="more-tags">
                    +{{ tool.tags.length - 2 }}
                </span>
            </div>

            <!-- 工具元信息 -->
            <div class="tool-meta">
                <span class="meta-item">
                    <n-icon size="12">
                        <Person />
                    </n-icon>
                    {{ tool.author }}
                </span>
                <span class="meta-item">
                    <n-icon size="12">
                        <Code />
                    </n-icon>
                    v{{ tool.version }}
                </span>
            </div>
        </div>

        <!-- 操作按钮 -->
        <div class="tool-actions">
            <n-space size="small">
                <n-button v-if="tool.status === ToolStatus.AVAILABLE || tool.status === ToolStatus.INSTALLING"
                    type="primary" size="tiny" :loading="tool.status === ToolStatus.INSTALLING"
                    @click="$emit('install', tool)">
                    <template #icon>
                        <n-icon size="12">
                            <Download />
                        </n-icon>
                    </template>
                    安装
                </n-button>

                <n-button v-else-if="tool.status === ToolStatus.INSTALLED" type="error" size="tiny"
                    @click="$emit('uninstall', tool)">
                    <template #icon>
                        <n-icon size="12">
                            <Trash />
                        </n-icon>
                    </template>
                    卸载
                </n-button>

                <n-button v-if="tool.status === ToolStatus.INSTALLED" size="tiny" @click="$emit('configure', tool)">
                    <template #icon>
                        <n-icon size="12">
                            <Settings />
                        </n-icon>
                    </template>
                    配置
                </n-button>

                <n-button size="tiny" quaternary @click="$emit('preview', tool)">
                    <template #icon>
                        <n-icon size="12">
                            <Eye />
                        </n-icon>
                    </template>
                    预览
                </n-button>
            </n-space>
        </div>
    </n-card>
</template>

<script setup lang="ts">
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
.tool-card {
    height: 100%;
    display: flex;
    flex-direction: column;
    transition: all 0.3s ease;

    :deep(.n-card__content) {
        padding: 12px;
    }

    &:hover {
        transform: translateY(-1px);
        box-shadow: var(--n-box-shadow-hover);
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
        margin-top: auto;
    }
}

@media (max-width: 768px) {
    .tool-card {
        .tool-actions {
            .n-space {
                flex-direction: column;
                width: 100%;

                .n-button {
                    width: 100%;
                }
            }
        }
    }
}
</style>