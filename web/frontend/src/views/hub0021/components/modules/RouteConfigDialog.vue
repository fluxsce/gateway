<template>
  <RsDialog
    :open="visible"
    layout="window"
    :title="isEditMode ? '编辑路由' : '新建路由'"
    :width="900"
    :draggable="true"
    :fullscreenable="true"
    @update:open="handleUpdateOpen"
  >
    <template #body>
      <RsForm
        ref="formRef"
        :rules="formRules"
        label-position="left"
        label-width="7.5rem"
        gap="md"
      >
        <RsTabs
          v-model="activeTab"
          :items="tabItems"
          variant="line"
          size="md"
          borderless
          content-gap="md"
        >
          <template #basic>
            <RsCard title="路由基本信息" class="tab-card">
              <div class="form-grid">
                <RsInput
                  v-model="formData.routeName"
                  name="routeName"
                  label="路由名称"
                  placeholder="请输入路由名称"
                  :maxlength="100"
                  show-count
                />

                <div class="field-block">
                  <RsLabel>匹配类型</RsLabel>
                  <RsSelect
                    v-model="formData.matchType"
                    name="matchType"
                    :options="matchTypeOptions"
                    placeholder="请选择匹配类型"
                    block
                    match-trigger-width
                    @update:model-value="handleMatchTypeChange"
                  />
                </div>

                <div class="form-grid__full">
                  <RsInput
                    v-model="formData.routePath"
                    name="routePath"
                    label="路由路径"
                    placeholder="请输入路由路径"
                    @update:model-value="handlePathInput"
                  />
                  <p class="field-hint">{{ getMatchTypeDescription }}</p>
                  <p class="field-hint">{{ getPathExample }}</p>
                </div>

                <div class="form-grid__full field-block">
                  <RsLabel>HTTP方法</RsLabel>
                  <div class="checkbox-row">
                    <RsCheckbox
                      v-for="method in httpMethodOptions"
                      :key="String(method.value)"
                      :model-value="isMethodChecked(String(method.value))"
                      @update:model-value="(checked) => toggleMethod(String(method.value), checked)"
                    >
                      {{ method.label }}
                    </RsCheckbox>
                  </div>
                </div>

                <div>
                  <RsInput
                    v-model="formData.allowedHosts"
                    name="allowedHosts"
                    label="允许的主机"
                    placeholder="留空表示允许所有主机"
                  />
                  <p class="field-hint">多个主机用逗号分隔，如：api.example.com,www.example.com</p>
                </div>

                <div>
                  <RsInputNumber
                    v-model="formData.routePriority"
                    name="routePriority"
                    label="路由优先级"
                    :min="1"
                    :max="999"
                    placeholder="数值越小优先级越高"
                  />
                  <p class="field-hint">数值越小优先级越高，建议范围：1-999</p>
                </div>

                <div class="form-grid__full">
                  <ServiceDefinitionSelector
                    v-model="formData.serviceDefinitionId"
                    :gateway-instance-id="gatewayInstanceId"
                  />
                </div>

                <RsInput
                  v-model="formData.logConfigId"
                  name="logConfigId"
                  label="日志配置"
                  placeholder="请输入日志配置ID（可选）"
                />

                <div class="field-row form-grid__full">
                  <RsLabel class="field-row__label">启用状态</RsLabel>
                  <div class="switch-with-text">
                    <RsSwitch
                      :model-value="activeSwitch === 'Y'"
                      @update:model-value="(v) => (activeSwitch = v ? 'Y' : 'N')"
                    />
                    <span>{{ activeSwitch === 'Y' ? '启用' : '禁用' }}</span>
                  </div>
                </div>
              </div>
            </RsCard>
          </template>

          <template #metadata>
            <RsCard title="元数据配置" class="tab-card">
              <div class="metadata-section">
                <div class="field-block">
                  <RsLabel>路由元数据</RsLabel>
                  <div class="metadata-list">
                    <div
                      v-for="(item, index) in metadataList"
                      :key="index"
                      class="metadata-row"
                    >
                      <RsInput v-model="item.key" placeholder="键" class="metadata-key" />
                      <RsInput v-model="item.value" placeholder="值" class="metadata-value" />
                      <RsButton variant="text" tone="danger" @click="removeMetadataItem(index)">
                        删除
                      </RsButton>
                    </div>
                    <RsButton variant="secondary" size="sm" @click="addMetadataItem">
                      添加元数据
                    </RsButton>
                  </div>
                  <p class="field-hint">用于存储路由的自定义元数据信息</p>
                </div>

                <div class="field-block">
                  <RsLabel>备注信息</RsLabel>
                  <textarea
                    v-model="formData.noteText"
                    class="form-textarea"
                    rows="4"
                    maxlength="500"
                    placeholder="请输入备注信息"
                  />
                  <p class="field-hint">{{ (formData.noteText || '').length }} / 500</p>
                </div>
              </div>
            </RsCard>
          </template>
        </RsTabs>

        <RsCard v-if="!isEditMode" class="tip-card">
          <RsAlert type="info" title="高级配置说明">
            路由创建成功后，您可以通过"路由配置管理"功能来配置：
            <ul class="tip-list">
              <li>断言设置（路由匹配条件）</li>
              <li>过滤器配置（请求处理逻辑）</li>
              <li>CORS跨域配置</li>
              <li>认证授权配置</li>
              <li>限流策略配置</li>
            </ul>
            这样可以确保在路由存在的基础上进行精确的配置管理。
          </RsAlert>
        </RsCard>
      </RsForm>
    </template>

    <template #footer>
      <div class="dialog-footer">
        <span v-if="isEditMode" class="footer-hint">
          提示：高级配置请使用"路由配置管理"功能
        </span>
        <div class="footer-actions">
          <RsButton variant="secondary" @click="closeDialog">取消</RsButton>
          <RsButton variant="primary" :loading="submitting" @click="handleSubmit">
            {{ isEditMode ? '更新路由' : '创建路由' }}
          </RsButton>
        </div>
      </div>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import {
  RsAlert,
  RsButton,
  RsCard,
  RsCheckbox,
  RsDialog,
  RsForm,
  RsInput,
  RsInputNumber,
  RsLabel,
  RsSelect,
  RsSwitch,
  RsTabs,
  type RsTabItem,
} from '@/ui'
import { computed, ref } from 'vue'
import { useRouteConfigDialog } from '../../hooks/useRouteConfigDialog'
import type { RouteConfig } from '../../types'
import { ServiceDefinitionSelector } from '../services'

interface Props {
  editingRoute?: RouteConfig | null
  gatewayInstanceId?: string
}

interface Emits {
  (e: 'success'): void
}

const props = withDefaults(defineProps<Props>(), {
  editingRoute: null,
  gatewayInstanceId: '',
})

const emit = defineEmits<Emits>()

const activeTab = ref('basic')
const tabItems: RsTabItem[] = [
  { value: 'basic', label: '基本信息' },
  { value: 'metadata', label: '元数据配置' },
]

const {
  visible,
  formRef,
  formData,
  formRules,
  isEditMode,
  httpMethodOptions,
  matchTypeOptions,
  getPathExample,
  getMatchTypeDescription,
  activeSwitch,
  metadataList,
  createMetadataItem,
  submitting,
  openDialog,
  closeDialog,
  handleSubmit,
  handleMatchTypeChange,
  handlePathInput,
  gatewayInstanceId: dialogGatewayInstanceId,
} = useRouteConfigDialog({
  onSuccess: () => emit('success'),
})

/** 优先用 openDialog 写入的实例 ID，其次用 props */
const gatewayInstanceId = computed(
  () => dialogGatewayInstanceId.value || props.gatewayInstanceId || '',
)

const isMethodChecked = (method: string) => {
  const methods = formData.allowedMethods
  return Array.isArray(methods) && methods.includes(method)
}

const toggleMethod = (method: string, checked: boolean) => {
  const current = Array.isArray(formData.allowedMethods) ? [...formData.allowedMethods] : []
  if (checked) {
    if (!current.includes(method)) current.push(method)
  } else {
    const idx = current.indexOf(method)
    if (idx >= 0) current.splice(idx, 1)
  }
  formData.allowedMethods = current
}

const addMetadataItem = () => {
  metadataList.value.push(createMetadataItem())
}

const removeMetadataItem = (index: number) => {
  metadataList.value.splice(index, 1)
}

const handleUpdateOpen = (open: boolean) => {
  if (!open) {
    closeDialog()
  } else {
    visible.value = true
  }
}

defineExpose({
  openDialog,
  closeDialog,
})
</script>

<style scoped>
.tab-card {
  margin-bottom: 8px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 24px;
}

.form-grid__full {
  grid-column: 1 / -1;
}

.field-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.field-row__label {
  width: 7.5rem;
  flex-shrink: 0;
}

.switch-with-text {
  display: flex;
  align-items: center;
  gap: 8px;
}

.checkbox-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 16px;
}

.field-hint {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--g-text-tertiary, var(--rs-text-tertiary));
  line-height: 1.4;
}

.metadata-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.metadata-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.metadata-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.metadata-key {
  flex: 1;
}

.metadata-value {
  flex: 2;
}

.form-textarea {
  width: 100%;
  box-sizing: border-box;
  padding: 8px 12px;
  border: 1px solid var(--g-border-primary, var(--rs-border));
  border-radius: 6px;
  background: var(--g-bg-primary, var(--rs-surface));
  color: var(--g-text-primary, var(--rs-text));
  font: inherit;
  resize: vertical;
}

.form-textarea:focus {
  outline: none;
  border-color: var(--g-primary, var(--rs-primary));
}

.tip-card {
  margin-top: 16px;
}

.tip-list {
  margin: 8px 0;
  padding-left: 20px;
}

.tip-list li {
  margin: 4px 0;
}

.dialog-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}

.footer-hint {
  font-size: 12px;
  color: var(--g-text-tertiary, var(--rs-text-tertiary));
}

.footer-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}
</style>
