<template>
  <div class="router-config-manager">
    <!-- 配置不存在时的提示 -->
    <RsEmpty v-if="!configExists && !loading" description="当前网关实例暂无Router配置">
      <RsButton variant="primary" :loading="saving" @click="handleCreateConfig">
        创建Router配置
      </RsButton>
    </RsEmpty>

    <!-- 配置存在时显示配置表单 -->
    <div v-else-if="configExists">
      <div class="header">
        <div class="header-left">
          <div class="header-titles">
            <span class="header-title">Router配置</span>
            <span class="header-desc">配置当前网关实例的Router级别设置</span>
          </div>
        </div>
        <div class="header-right">
          <RsButton
            variant="primary"
            :loading="saving"
            :disabled="!props.gatewayInstanceId"
            @click="handleSave"
          >
            <GIcon :icon="SaveOutline" />
            保存配置
          </RsButton>
          <RsButton variant="secondary" :loading="loading" @click="handleRefresh">
            <GIcon :icon="Refresh" />
            刷新
          </RsButton>
          <RsButton variant="secondary" :disabled="!hasChanges" @click="handleReset">
            <GIcon :icon="RefreshOutline" />
            重置
          </RsButton>
        </div>
      </div>

      <RsCard title="Router基础配置" class="config-card">
        <RsForm
          ref="formRef"
          :rules="rules"
          label-position="left"
          label-width="140px"
          gap="md"
        >
          <div class="form-grid">
            <RsInput
              v-model="formData.routerName"
              name="routerName"
              label="Router名称"
              placeholder="请输入Router名称"
            />
            <RsInputNumber
              v-model="formData.defaultPriority"
              name="defaultPriority"
              label="默认优先级"
              :min="0"
              :max="9999"
              placeholder="默认路由优先级"
            />
          </div>
          <div class="field-block">
            <RsLabel>Router描述</RsLabel>
            <textarea
              v-model="formData.routerDesc"
              class="form-textarea"
              rows="2"
              placeholder="请输入Router描述信息"
            />
          </div>
        </RsForm>
      </RsCard>

      <RsCard title="路由缓存配置" class="config-card">
        <RsForm label-position="left" label-width="140px" gap="md">
          <div class="form-grid">
            <div class="field-row">
              <RsLabel class="field-row__label">启用路由缓存</RsLabel>
              <RsSwitch
                :model-value="formData.enableRouteCache === 'Y'"
                @update:model-value="(v) => (formData.enableRouteCache = v ? 'Y' : 'N')"
              />
            </div>
            <RsInputNumber
              v-if="formData.enableRouteCache === 'Y'"
              v-model="formData.routeCacheTtlSeconds"
              name="routeCacheTtlSeconds"
              label="缓存TTL(秒)"
              :min="1"
              :max="86400"
              placeholder="缓存存活时间"
            />
          </div>
          <div v-if="formData.enableRouteCache === 'Y'" class="form-grid">
            <RsInputNumber
              v-model="formData.maxRoutes"
              label="最大路由数"
              :min="1"
              :max="10000"
              placeholder="最大缓存路由数"
            />
            <RsInputNumber
              v-model="formData.routeMatchTimeout"
              label="路由匹配超时(ms)"
              :min="100"
              :max="30000"
              placeholder="路由匹配超时时间"
            />
          </div>
        </RsForm>
      </RsCard>

      <RsCard title="全局过滤器配置" class="config-card">
        <RsForm label-position="left" label-width="140px" gap="md">
          <div class="form-grid">
            <div class="field-row">
              <RsLabel class="field-row__label">启用全局过滤器</RsLabel>
              <RsSwitch
                :model-value="formData.enableGlobalFilters === 'Y'"
                @update:model-value="(v) => (formData.enableGlobalFilters = v ? 'Y' : 'N')"
              />
            </div>
            <div v-if="formData.enableGlobalFilters === 'Y'" class="field-row field-row--stack">
              <RsLabel class="field-row__label">过滤器执行模式</RsLabel>
              <RsSelect
                v-model="formData.filterExecutionMode"
                :options="executionModeOptions"
                placeholder="选择执行模式"
                block
                match-trigger-width
              />
            </div>
          </div>
          <div v-if="formData.enableGlobalFilters === 'Y'" class="form-grid">
            <RsInputNumber
              v-model="formData.maxFilterChainDepth"
              label="最大过滤器链深度"
              :min="1"
              :max="100"
              placeholder="过滤器链最大深度"
            />
            <div class="field-row">
              <RsLabel class="field-row__label">启用异步处理</RsLabel>
              <RsSwitch
                :model-value="formData.enableAsyncProcessing === 'Y'"
                @update:model-value="(v) => (formData.enableAsyncProcessing = v ? 'Y' : 'N')"
              />
            </div>
          </div>
        </RsForm>
      </RsCard>

      <RsCard title="性能优化配置" class="config-card">
        <RsForm label-position="left" label-width="140px" gap="md">
          <div class="form-grid">
            <div class="field-row">
              <RsLabel class="field-row__label">启用严格模式</RsLabel>
              <RsSwitch
                :model-value="formData.enableStrictMode === 'Y'"
                @update:model-value="(v) => (formData.enableStrictMode = v ? 'Y' : 'N')"
              />
            </div>
            <div class="field-row">
              <RsLabel class="field-row__label">启用路由池</RsLabel>
              <RsSwitch
                :model-value="formData.enableRoutePooling === 'Y'"
                @update:model-value="(v) => (formData.enableRoutePooling = v ? 'Y' : 'N')"
              />
            </div>
          </div>
          <div class="form-grid">
            <RsInputNumber
              v-if="formData.enableRoutePooling === 'Y'"
              v-model="formData.routePoolSize"
              label="路由池大小"
              :min="10"
              :max="1000"
              placeholder="路由对象池大小"
            />
            <div class="field-row">
              <RsLabel class="field-row__label">大小写敏感</RsLabel>
              <RsSwitch
                :model-value="formData.caseSensitive === 'Y'"
                @update:model-value="(v) => (formData.caseSensitive = v ? 'Y' : 'N')"
              />
            </div>
          </div>
          <div class="form-grid">
            <div class="field-row">
              <RsLabel class="field-row__label">移除尾部斜杠</RsLabel>
              <RsSwitch
                :model-value="formData.removeTrailingSlash === 'Y'"
                @update:model-value="(v) => (formData.removeTrailingSlash = v ? 'Y' : 'N')"
              />
            </div>
            <div class="field-row">
              <RsLabel class="field-row__label">启用监控指标</RsLabel>
              <RsSwitch
                :model-value="formData.enableMetrics === 'Y'"
                @update:model-value="(v) => (formData.enableMetrics = v ? 'Y' : 'N')"
              />
            </div>
          </div>
        </RsForm>
      </RsCard>

      <RsCard title="错误处理配置" class="config-card">
        <RsForm label-position="left" label-width="140px" gap="md">
          <div class="form-grid">
            <div class="field-row">
              <RsLabel class="field-row__label">启用降级处理</RsLabel>
              <RsSwitch
                :model-value="formData.enableFallback === 'Y'"
                @update:model-value="(v) => (formData.enableFallback = v ? 'Y' : 'N')"
              />
            </div>
            <RsInput
              v-if="formData.enableFallback === 'Y'"
              v-model="formData.fallbackRoute"
              label="降级路由"
              placeholder="降级路由路径"
            />
          </div>
          <div class="form-grid">
            <RsInputNumber
              v-model="formData.notFoundStatusCode"
              name="notFoundStatusCode"
              label="404状态码"
              :min="400"
              :max="599"
              placeholder="未找到路由的状态码"
            />
            <RsInput
              v-model="formData.notFoundMessage"
              name="notFoundMessage"
              label="404消息"
              placeholder="未找到路由的错误消息"
            />
          </div>
        </RsForm>
      </RsCard>

      <RsCard title="其他配置">
        <RsForm label-position="left" label-width="140px" gap="md">
          <div class="form-grid">
            <div class="field-row">
              <RsLabel class="field-row__label">启用链路追踪</RsLabel>
              <RsSwitch
                :model-value="formData.enableTracing === 'Y'"
                @update:model-value="(v) => (formData.enableTracing = v ? 'Y' : 'N')"
              />
            </div>
            <div class="field-row">
              <RsLabel class="field-row__label">配置状态</RsLabel>
              <RsSwitch
                :model-value="formData.activeFlag === 'Y'"
                @update:model-value="(v) => (formData.activeFlag = v ? 'Y' : 'N')"
              />
            </div>
          </div>
          <div class="field-block">
            <RsLabel>备注信息</RsLabel>
            <textarea
              v-model="formData.noteText"
              class="form-textarea"
              rows="3"
              placeholder="请输入备注信息"
            />
          </div>
        </RsForm>
      </RsCard>
    </div>

    <!-- 加载状态 -->
    <div v-else-if="loading" class="loading-wrap">
      <RsLoading size="lg" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { GIcon } from '@/components/gicon'
import { useAppMessage } from '@/composables/useAppMessage'
import {
  RsButton,
  RsCard,
  RsEmpty,
  RsForm,
  RsInput,
  RsInputNumber,
  RsLabel,
  RsLoading,
  RsSelect,
  RsSwitch,
  type RsFormRules,
  type RsFormValidationResult,
  type RsSelectOption,
} from '@/ui'
import { Refresh, RefreshOutline, SaveOutline } from '@vicons/ionicons5'
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { addRouterConfig, editRouterConfig, getRouterConfigsByInstance } from '../api'
import type { RouterConfigForm } from '../types'

interface Props {
  gatewayInstanceId: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'router-status-change': [exists: boolean]
  'router-config-ready': [ready: boolean]
}>()

if (!props.gatewayInstanceId) {
  console.warn('RouterConfigManager: gatewayInstanceId is required')
}

const message = useAppMessage()

type RsFormExpose = {
  validate: () => Promise<RsFormValidationResult>
}

const loading = ref(false)
const saving = ref(false)
const formRef = ref<RsFormExpose>()
const originalData = ref<RouterConfigForm | null>(null)
const isUnmounted = ref(false)
const configExists = ref(false)
const configReady = ref(false)

const formData = reactive({
  gatewayInstanceId: '',
  routerConfigId: '',
  routerName: '默认Router',
  routerDesc: '',
  defaultPriority: 100,
  enableRouteCache: 'Y' as 'Y' | 'N',
  routeCacheTtlSeconds: 300,
  maxRoutes: 1000,
  routeMatchTimeout: 5000,
  enableStrictMode: 'N' as 'Y' | 'N',
  enableMetrics: 'Y' as 'Y' | 'N',
  enableTracing: 'N' as 'Y' | 'N',
  caseSensitive: 'N' as 'Y' | 'N',
  removeTrailingSlash: 'Y' as 'Y' | 'N',
  enableGlobalFilters: 'Y' as 'Y' | 'N',
  filterExecutionMode: 'SEQUENTIAL',
  maxFilterChainDepth: 10,
  enableRoutePooling: 'N' as 'Y' | 'N',
  routePoolSize: 100,
  enableAsyncProcessing: 'N' as 'Y' | 'N',
  enableFallback: 'N' as 'Y' | 'N',
  fallbackRoute: '',
  notFoundStatusCode: 404,
  notFoundMessage: 'Route not found',
  activeFlag: 'Y' as 'Y' | 'N',
  noteText: '',
})

const executionModeOptions: RsSelectOption[] = [
  { label: '顺序执行', value: 'SEQUENTIAL' },
  { label: '并行执行', value: 'PARALLEL' },
]

const rules: RsFormRules = {
  routerName: [{ required: true, message: '请输入Router名称', trigger: 'blur' }],
  defaultPriority: [{ required: true, type: 'number', message: '请输入默认优先级', trigger: 'blur' }],
  routeCacheTtlSeconds: [{ required: true, type: 'number', message: '请输入缓存TTL', trigger: 'blur' }],
  notFoundStatusCode: [{ required: true, type: 'number', message: '请输入404状态码', trigger: 'blur' }],
  notFoundMessage: [{ required: true, message: '请输入404消息', trigger: 'blur' }],
}

/**
 * 加载Router配置
 */
const loadRouterConfig = async () => {
  if (!props.gatewayInstanceId || isUnmounted.value) return

  loading.value = true
  configExists.value = false
  configReady.value = false
  try {
    console.log('开始加载Router配置，网关实例ID:', props.gatewayInstanceId)
    const response = await getRouterConfigsByInstance(props.gatewayInstanceId)
    console.log('API响应:', response)

    if (isUnmounted.value) return

    if (response?.oK) {
      if (!response.bizData) {
        console.warn('No bizData in response')
        configExists.value = false
        originalData.value = JSON.parse(JSON.stringify(formData))
        return
      }

      let configs
      try {
        configs = JSON.parse(response.bizData)
      } catch (parseError) {
        console.error('Error parsing bizData:', parseError)
        message.error('解析Router配置数据失败')
        configExists.value = false
        return
      }

      const config = configs && typeof configs === 'object' && !Array.isArray(configs) ? configs : null

      if (config) {
        configExists.value = true
        configReady.value = true

        console.log('Loading router config:', config)

        Object.assign(formData, {
          routerConfigId: config.routerConfigId || '',
          routerName: config.routerName || '默认Router',
          routerDesc: config.routerDesc || '',
          defaultPriority: config.defaultPriority || 100,
          enableRouteCache: config.enableRouteCache || 'Y',
          routeCacheTtlSeconds: config.routeCacheTtlSeconds || 300,
          maxRoutes: config.maxRoutes || 1000,
          routeMatchTimeout: config.routeMatchTimeout || 5000,
          enableStrictMode: config.enableStrictMode || 'N',
          enableMetrics: config.enableMetrics || 'Y',
          enableTracing: config.enableTracing || 'N',
          caseSensitive: config.caseSensitive || 'N',
          removeTrailingSlash: config.removeTrailingSlash || 'Y',
          enableGlobalFilters: config.enableGlobalFilters || 'Y',
          filterExecutionMode: config.filterExecutionMode || 'SEQUENTIAL',
          maxFilterChainDepth: config.maxFilterChainDepth || 10,
          enableRoutePooling: config.enableRoutePooling || 'N',
          routePoolSize: config.routePoolSize || 100,
          enableAsyncProcessing: config.enableAsyncProcessing || 'N',
          enableFallback: config.enableFallback || 'N',
          fallbackRoute: config.fallbackRoute || '',
          notFoundStatusCode: config.notFoundStatusCode || 404,
          notFoundMessage: config.notFoundMessage || 'Route not found',
          activeFlag: config.activeFlag || 'Y',
          noteText: config.noteText || '',
        })

        console.log('Router config loaded successfully. configExists:', configExists.value)
        console.log('Form data updated:', {
          routerConfigId: formData.routerConfigId,
          routerName: formData.routerName,
        })
      } else {
        console.log('No router configs found for instance:', props.gatewayInstanceId)
        configExists.value = false
      }

      originalData.value = JSON.parse(JSON.stringify(formData))
    } else {
      const errorMsg = response?.errMsg || '加载Router配置失败'
      console.error('API response error:', response)
      message.error(errorMsg)
    }
  } catch (error) {
    console.error('Error loading router config:', error)
    message.error('加载Router配置时发生错误: ' + (error instanceof Error ? error.message : String(error)))
  } finally {
    loading.value = false
    emit('router-status-change', configExists.value)
    emit('router-config-ready', configReady.value)
  }
}

const hasChanges = computed(() => {
  if (!originalData.value) return false
  return JSON.stringify(formData) !== JSON.stringify(originalData.value)
})

watch(
  () => props.gatewayInstanceId,
  async (newId) => {
    if (newId) {
      formData.gatewayInstanceId = newId
      try {
        await loadRouterConfig()
      } catch (error) {
        console.error('Error in watcher callback:', error)
        message.error('加载Router配置时发生错误')
      }
    }
  },
  { immediate: true },
)

/**
 * 保存配置
 */
const handleSave = async () => {
  if (!formRef.value) return

  try {
    const result = await formRef.value.validate()
    if (!result?.valid) {
      message.warning('请检查表单输入')
      return
    }

    saving.value = true

    const submitData = {
      routerConfigId: formData.routerConfigId,
      gatewayInstanceId: formData.gatewayInstanceId,
      routerName: formData.routerName,
      routerDesc: formData.routerDesc,
      defaultPriority: formData.defaultPriority,
      enableRouteCache: formData.enableRouteCache,
      routeCacheTtlSeconds: formData.routeCacheTtlSeconds,
      maxRoutes: formData.maxRoutes,
      routeMatchTimeout: formData.routeMatchTimeout,
      enableStrictMode: formData.enableStrictMode,
      enableMetrics: formData.enableMetrics,
      enableTracing: formData.enableTracing,
      caseSensitive: formData.caseSensitive,
      removeTrailingSlash: formData.removeTrailingSlash,
      enableGlobalFilters: formData.enableGlobalFilters,
      filterExecutionMode: formData.filterExecutionMode,
      maxFilterChainDepth: formData.maxFilterChainDepth,
      enableRoutePooling: formData.enableRoutePooling,
      routePoolSize: formData.routePoolSize,
      enableAsyncProcessing: formData.enableAsyncProcessing,
      enableFallback: formData.enableFallback,
      fallbackRoute: formData.fallbackRoute,
      notFoundStatusCode: formData.notFoundStatusCode,
      notFoundMessage: formData.notFoundMessage,
      activeFlag: formData.activeFlag,
      noteText: formData.noteText,
      routerMetadata: {},
      customConfig: {},
    }

    let response
    if (formData.routerConfigId) {
      response = await editRouterConfig(submitData as any)
    } else {
      response = await addRouterConfig(submitData as any)
    }

    if (response.oK) {
      message.success('Router配置保存成功')

      if (response.bizData) {
        try {
          const updatedConfig = JSON.parse(response.bizData)
          if (updatedConfig) {
            Object.assign(formData, {
              routerConfigId: updatedConfig.routerConfigId || formData.routerConfigId,
              routerName: updatedConfig.routerName || formData.routerName,
              routerDesc: updatedConfig.routerDesc || formData.routerDesc,
              defaultPriority: updatedConfig.defaultPriority || formData.defaultPriority,
              enableRouteCache: updatedConfig.enableRouteCache || formData.enableRouteCache,
              routeCacheTtlSeconds: updatedConfig.routeCacheTtlSeconds || formData.routeCacheTtlSeconds,
              maxRoutes: updatedConfig.maxRoutes || formData.maxRoutes,
              routeMatchTimeout: updatedConfig.routeMatchTimeout || formData.routeMatchTimeout,
              enableStrictMode: updatedConfig.enableStrictMode || formData.enableStrictMode,
              enableMetrics: updatedConfig.enableMetrics || formData.enableMetrics,
              enableTracing: updatedConfig.enableTracing || formData.enableTracing,
              caseSensitive: updatedConfig.caseSensitive || formData.caseSensitive,
              removeTrailingSlash: updatedConfig.removeTrailingSlash || formData.removeTrailingSlash,
              enableGlobalFilters: updatedConfig.enableGlobalFilters || formData.enableGlobalFilters,
              filterExecutionMode: updatedConfig.filterExecutionMode || formData.filterExecutionMode,
              maxFilterChainDepth: updatedConfig.maxFilterChainDepth || formData.maxFilterChainDepth,
              enableRoutePooling: updatedConfig.enableRoutePooling || formData.enableRoutePooling,
              routePoolSize: updatedConfig.routePoolSize || formData.routePoolSize,
              enableAsyncProcessing: updatedConfig.enableAsyncProcessing || formData.enableAsyncProcessing,
              enableFallback: updatedConfig.enableFallback || formData.enableFallback,
              fallbackRoute: updatedConfig.fallbackRoute || formData.fallbackRoute,
              notFoundStatusCode: updatedConfig.notFoundStatusCode || formData.notFoundStatusCode,
              notFoundMessage: updatedConfig.notFoundMessage || formData.notFoundMessage,
              activeFlag: updatedConfig.activeFlag || formData.activeFlag,
              noteText: updatedConfig.noteText || formData.noteText,
            })

            originalData.value = JSON.parse(JSON.stringify(formData))
          }
        } catch (parseError) {
          console.error('Error parsing updated config:', parseError)
          await loadRouterConfig()
        }
      } else {
        await loadRouterConfig()
      }
    } else {
      message.error('保存Router配置失败')
    }
  } catch (error) {
    console.error('Error saving router config:', error)
    message.error('保存Router配置时发生错误')
  } finally {
    saving.value = false
  }
}

/**
 * 刷新配置
 */
const handleRefresh = async () => {
  await loadRouterConfig()
  message.success('刷新完成')
}

/**
 * 重置配置
 */
const handleReset = () => {
  if (originalData.value) {
    Object.assign(formData, JSON.parse(JSON.stringify(originalData.value)))
    message.success('配置已重置')
  }
}

/**
 * 创建新的Router配置
 */
const handleCreateConfig = async () => {
  if (!props.gatewayInstanceId) {
    message.error('网关实例ID不能为空')
    return
  }

  configReady.value = true
  emit('router-config-ready', true)

  saving.value = true
  try {
    const submitData = {
      gatewayInstanceId: props.gatewayInstanceId,
      routerName: formData.routerName,
      routerDesc: formData.routerDesc,
      defaultPriority: formData.defaultPriority,
      enableRouteCache: formData.enableRouteCache,
      routeCacheTtlSeconds: formData.routeCacheTtlSeconds,
      maxRoutes: formData.maxRoutes,
      routeMatchTimeout: formData.routeMatchTimeout,
      enableStrictMode: formData.enableStrictMode,
      enableMetrics: formData.enableMetrics,
      enableTracing: formData.enableTracing,
      caseSensitive: formData.caseSensitive,
      removeTrailingSlash: formData.removeTrailingSlash,
      enableGlobalFilters: formData.enableGlobalFilters,
      filterExecutionMode: formData.filterExecutionMode,
      maxFilterChainDepth: formData.maxFilterChainDepth,
      enableRoutePooling: formData.enableRoutePooling,
      routePoolSize: formData.routePoolSize,
      enableAsyncProcessing: formData.enableAsyncProcessing,
      enableFallback: formData.enableFallback,
      fallbackRoute: formData.fallbackRoute,
      notFoundStatusCode: formData.notFoundStatusCode,
      notFoundMessage: formData.notFoundMessage,
      activeFlag: formData.activeFlag,
      noteText: formData.noteText,
      routerMetadata: {},
      customConfig: {},
    }

    const response = await addRouterConfig(submitData as any)

    if (response.oK) {
      message.success('Router配置创建成功')
      configExists.value = true

      if (response.bizData) {
        try {
          const newConfig = JSON.parse(response.bizData)
          if (newConfig) {
            formData.routerConfigId = newConfig.routerConfigId
            originalData.value = JSON.parse(JSON.stringify(formData))
          }
        } catch (parseError) {
          console.error('Error parsing new config:', parseError)
          await loadRouterConfig()
        }
      } else {
        await loadRouterConfig()
      }

      emit('router-status-change', true)
    } else {
      message.error('创建Router配置失败')
    }
  } catch (error) {
    console.error('Error creating router config:', error)
    message.error('创建Router配置时发生错误')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  if (props.gatewayInstanceId) {
    console.log('Router配置管理器初始化，网关实例ID:', props.gatewayInstanceId)
  }
})

onBeforeUnmount(() => {
  isUnmounted.value = true
})
</script>

<style scoped>
.router-config-manager {
  width: 100%;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 16px;
  background: var(--g-bg-secondary, var(--rs-surface-hover, #fafafa));
  border-radius: 6px;
}

.header-left {
  flex: 1;
}

.header-titles {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.header-title {
  font-weight: 600;
  color: var(--g-text-primary, var(--rs-text));
}

.header-desc {
  font-size: 13px;
  color: var(--g-text-tertiary, var(--rs-text-tertiary));
}

.header-right {
  display: flex;
  gap: 8px;
}

.config-card {
  margin-bottom: 16px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 24px;
}

.field-row {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 32px;
}

.field-row--stack {
  align-items: flex-start;
}

.field-row--stack .field-row__label {
  padding-top: 6px;
}

.field-row__label {
  width: 140px;
  flex-shrink: 0;
}

.field-row :deep(.rs-select) {
  flex: 1;
  min-width: 0;
}

.field-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
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

.loading-wrap {
  width: 100%;
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
