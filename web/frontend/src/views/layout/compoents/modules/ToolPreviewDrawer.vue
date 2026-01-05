<template>
    <n-drawer :show="show" @update:show="(val: boolean) => $emit('update:show', val)" :width="600" placement="right">
        <n-drawer-content title="工具预览" closable>
            <div v-if="tool" class="tool-preview">
                <div class="preview-header">
                    <div class="tool-icon">
                        <n-icon size="48" :name="tool.icon || 'apps-outline'" />
                    </div>
                    <div class="tool-info">
                        <h2>{{ tool.displayName || tool.name }}</h2>
                        <p>{{ tool.description }}</p>
                    </div>
                </div>

                <n-divider />

                <div class="preview-content">
                    <n-space vertical>
                        <n-card title="基本信息">
                            <n-descriptions :column="2">
                                <n-descriptions-item label="版本">{{ tool.version }}</n-descriptions-item>
                                <n-descriptions-item label="作者">{{ tool.author }}</n-descriptions-item>
                                <n-descriptions-item label="大小">{{ tool.size }}KB</n-descriptions-item>
                                <n-descriptions-item label="状态">{{ tool.status }}</n-descriptions-item>
                            </n-descriptions>
                        </n-card>

                        <n-card title="标签">
                            <n-space>
                                <n-tag v-for="tag in tool.tags" :key="tag" type="primary">
                                    {{ tag }}
                                </n-tag>
                            </n-space>
                        </n-card>
                    </n-space>
                </div>

                <!-- 操作按钮 -->
                <div class="preview-actions">
                    <n-space>
                        <n-button type="primary" @click="$emit('install', tool)">
                            安装
                        </n-button>
                        <n-button @click="$emit('configure', tool)">
                            配置
                        </n-button>
                    </n-space>
                </div>
            </div>
        </n-drawer-content>
    </n-drawer>
</template>

<script setup lang="ts">
import type { Tool } from '../../types/toolMarketplace'

interface Props {
    show: boolean
    tool?: Tool | null
}

defineProps<Props>()

defineEmits<{
    'update:show': [value: boolean]
    install: [tool: Tool]
    configure: [tool: Tool]
}>()
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
        margin: 16px 0;
    }

    .preview-actions {
        margin-top: 24px;
        padding-top: 16px;
        border-top: 1px solid var(--border-color);
    }
}
</style>