<template>
  <div class="env-settings">
    <RsLoading :loading="loading" overlay block size="lg" />

    <aside class="env-settings__nav">
      <div class="env-settings__nav-title">{{ t('moduleName') }}</div>
      <RsMenu v-model="activeTab" :items="navItems" class="env-settings__menu" />
    </aside>

    <section class="env-settings__main">
      <header class="env-settings__header">
        <h2 class="env-settings__title">{{ activeLabel }}</h2>
      </header>

      <div class="env-settings__body" :class="{ 'env-settings__body--fill': activeTab === 'envVars' }">
        <template v-if="activeTab === 'retention'">
          <RsAlert type="info" class="hint">{{ t('retention.hint') }}</RsAlert>
          <RsForm
            class="settings-form"
            label-position="left"
            label-width="10rem"
            size="md"
            gap="md"
          >
            <RsFormItem v-for="field in retentionFields" :key="field">
              <template #label>
                <RsTooltip :content="t(`retention.${field}Desc`)" icon>
                  {{ t(`retention.${field}`) }}
                </RsTooltip>
              </template>
              <div class="input-with-unit">
                <RsInputNumber
                  v-model="retention[field]"
                  :name="field"
                  label-position="top"
                  :min="1"
                  :max="3650"
                  :disabled="!canEdit"
                />
                <span class="field-unit">{{ t('common.days') }}</span>
              </div>
            </RsFormItem>
            <div class="form-actions">
              <RsButton
                variant="primary"
                :loading="savingRetention"
                :disabled="!canEdit"
                @click="saveGroup('retention')"
              >
                {{ t('common.save') }}
              </RsButton>
            </div>
          </RsForm>
        </template>

        <template v-else-if="activeTab === 'retentionJob'">
          <RsAlert type="info" class="hint">{{ t('retentionJob.hint') }}</RsAlert>
          <RsForm
            class="settings-form"
            label-position="left"
            label-width="10rem"
            size="md"
            gap="md"
          >
            <RsFormItem>
              <template #label>
                <RsTooltip :content="t('retentionJob.enabledDesc')" icon>
                  {{ t('retentionJob.enabled') }}
                </RsTooltip>
              </template>
              <RsSwitch v-model="retentionJob.enabled" :disabled="!canEdit" />
            </RsFormItem>
            <RsFormItem>
              <template #label>
                <RsTooltip :content="t('retentionJob.intervalMinutesDesc')" icon>
                  {{ t('retentionJob.intervalMinutes') }}
                </RsTooltip>
              </template>
              <div class="input-with-unit">
                <RsInputNumber
                  v-model="retentionJob.intervalMinutes"
                  name="intervalMinutes"
                  label-position="top"
                  :min="1"
                  :max="10080"
                  :disabled="!canEdit"
                />
                <span class="field-unit">{{ t('common.minutes') }}</span>
              </div>
            </RsFormItem>
            <RsFormItem>
              <template #label>
                <RsTooltip :content="t('retentionJob.startTimeDesc')" icon>
                  {{ t('retentionJob.startTime') }}
                </RsTooltip>
              </template>
              <div class="input-with-unit">
                <RsTimePicker
                  v-model="retentionJob.startTime"
                  name="startTime"
                  label-position="top"
                  :placeholder="t('retentionJob.startTimePlaceholder')"
                  :disabled="!canEdit"
                />
              </div>
            </RsFormItem>
            <div class="form-actions">
              <RsButton
                variant="primary"
                :loading="savingRetentionJob"
                :disabled="!canEdit"
                @click="saveGroup('retentionJob')"
              >
                {{ t('common.save') }}
              </RsButton>
            </div>
          </RsForm>
        </template>

        <template v-else-if="activeTab === 'webTimeout'">
          <RsAlert type="info" class="hint">{{ t('webTimeout.hint') }}</RsAlert>
          <RsForm
            class="settings-form"
            label-position="left"
            label-width="10rem"
            size="md"
            gap="md"
          >
            <RsFormItem v-for="field in webTimeoutFields" :key="field.key">
              <template #label>
                <RsTooltip :content="t(`webTimeout.${field.key}Desc`)" icon>
                  {{ t(`webTimeout.${field.key}`) }}
                </RsTooltip>
              </template>
              <div class="input-with-unit">
                <RsInputNumber
                  v-model="webTimeout[field.key]"
                  :name="field.key"
                  label-position="top"
                  :min="field.min"
                  :max="field.max"
                  :disabled="!canEdit"
                />
                <span class="field-unit">{{ t(`common.${field.unit}`) }}</span>
              </div>
            </RsFormItem>
            <div class="form-actions">
              <RsButton
                variant="primary"
                :loading="savingWebTimeout"
                :disabled="!canEdit"
                @click="saveGroup('webTimeout')"
              >
                {{ t('common.save') }}
              </RsButton>
            </div>
          </RsForm>
        </template>

        <EnvVarsPanel
          v-else
          ref="envVarsPanelRef"
          class="env-settings__vars"
          :env-vars="envVars"
          :can-edit="canEdit"
          :saving-env-var="savingEnvVar"
          @save="onSaveVariable"
          @remove="onRemoveVariable"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import {
  RsAlert,
  RsButton,
  RsForm,
  RsFormItem,
  RsInputNumber,
  RsLoading,
  RsMenu,
  RsSwitch,
  RsTimePicker,
  RsTooltip,
  type RsMenuItem,
} from '@/ui'
import { computed, onMounted, ref } from 'vue'
import EnvVarsPanel from './components/EnvVarsPanel.vue'
import { useEnvironmentSettings } from './hooks'
import type { RetentionSettings, WebTimeoutSettings } from './types'

defineOptions({ name: 'EnvironmentSettings' })

type RetentionDayKey = Exclude<keyof RetentionSettings, 'currentVersion'>
type WebTimeoutKey = Exclude<keyof WebTimeoutSettings, 'currentVersion'>
type SettingTab = 'retention' | 'retentionJob' | 'webTimeout' | 'envVars'

const { t } = useModuleI18n('hub0009')
const activeTab = ref<SettingTab>('retention')
const navItems = computed<RsMenuItem[]>(() => [
  { key: 'retention', label: t('tabs.retention'), icon: 'archive' },
  { key: 'retentionJob', label: t('tabs.retentionJob'), icon: 'clock' },
  { key: 'webTimeout', label: t('tabs.webTimeout'), icon: 'timer' },
  { key: 'envVars', label: t('tabs.envVars'), icon: 'key' },
])
const activeLabel = computed(
  () => navItems.value.find((item) => item.key === activeTab.value)?.label || t('moduleName'),
)

const retentionFields: RetentionDayKey[] = [
  'auditLogDays',
  'taskLogDays',
  'alertLogDays',
  'clusterEventDays',
  'metricsDays',
  'gatewayLogDefaultDays',
]

const webTimeoutFields: {
  key: WebTimeoutKey
  min: number
  max: number
  unit: 'seconds' | 'hours'
}[] = [
  { key: 'requestTimeoutSeconds', min: 10, max: 600, unit: 'seconds' },
  { key: 'sessionExpireHours', min: 1, max: 168, unit: 'hours' },
]

const {
  loading,
  retention,
  retentionJob,
  webTimeout,
  envVars,
  canEdit,
  savingRetention,
  savingRetentionJob,
  savingWebTimeout,
  savingEnvVar,
  fetchSettings,
  saveGroup,
  saveVariable,
  removeVariable,
} = useEnvironmentSettings()

const envVarsPanelRef = ref<{ closeDialog?: () => void } | null>(null)

const onSaveVariable = async (payload: {
  name: string
  originalName?: string
  value: string
  secret: boolean
  note: string
}) => {
  const ok = await saveVariable(payload)
  if (ok) {
    envVarsPanelRef.value?.closeDialog?.()
  }
}

const onRemoveVariable = async (name: string) => {
  await removeVariable(name)
}

onMounted(() => {
  fetchSettings()
})
</script>

<style lang="scss" scoped>
.env-settings {
  --env-nav-width: 13.5rem;

  position: relative;
  box-sizing: border-box;
  display: flex;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  background: var(--rs-bg);
}

.env-settings__nav {
  flex: 0 0 var(--env-nav-width);
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: var(--rs-space-md) var(--rs-space-sm);
  background: var(--rs-surface);
  border-right: 1px solid var(--rs-border);
}

.env-settings__nav-title {
  flex-shrink: 0;
  padding: 0 var(--rs-space-sm) var(--rs-space-md);
  color: var(--rs-muted);
  font-size: var(--rs-font-size-xs);
  font-weight: 600;
  letter-spacing: 0.04em;
}

.env-settings__menu {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

.env-settings__main {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

.env-settings__header {
  flex-shrink: 0;
  padding: var(--rs-space-md) var(--rs-space-lg) 0;
}

.env-settings__title {
  margin: 0;
  color: var(--rs-fg);
  font-size: var(--rs-font-size-lg);
  font-weight: 600;
  line-height: 1.4;
}

.env-settings__body {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
  padding: var(--rs-space-md) var(--rs-space-lg) var(--rs-space-lg);
}

.env-settings__body--fill {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.env-settings__vars {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.hint {
  margin-bottom: 16px;
}

.settings-form {
  max-width: 36rem;
}

.input-with-unit {
  display: flex;
  align-items: center;
  gap: 8px;
  max-width: 16rem;
}

.input-with-unit :deep(.rs-field) {
  flex: 1;
  min-width: 0;
  width: auto;
}

.field-unit {
  flex-shrink: 0;
  color: var(--rs-muted);
  font-size: var(--rs-font-size-sm);
  white-space: nowrap;
}

.form-actions {
  display: flex;
  gap: 12px;
  padding-left: 10rem;
  margin-top: 8px;
}
</style>
