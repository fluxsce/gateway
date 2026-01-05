<template>
    <n-modal :show="show" @update:show="(val: boolean) => emit('update:show', val)" preset="dialog" title="工具配置">
        <div v-if="tool" class="config-dialog">
            <n-form>
                <n-form-item label="工具名称">
                    <n-input :value="tool.displayName || tool.name" readonly />
                </n-form-item>

                <n-form-item label="版本">
                    <n-input :value="tool.version" readonly />
                </n-form-item>

                <n-form-item label="状态">
                    <n-tag :type="tool.status === 'installed' ? 'success' : 'default'">
                        {{ tool.status }}
                    </n-tag>
                </n-form-item>
            </n-form>
        </div>

        <template #action>
            <n-space>
                <n-button @click="emit('update:show', false)">取消</n-button>
                <n-button type="primary" @click="handleSave">保存</n-button>
            </n-space>
        </template>
    </n-modal>
</template>

<script setup lang="ts">
import type { Tool } from '../../types/toolMarketplace'

interface Props {
    show: boolean
    tool?: Tool | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
    'update:show': [value: boolean]
    save: [config: Record<string, any>]
}>()

const handleSave = () => {
    emit('save', {})
    emit('update:show', false)
}
</script>