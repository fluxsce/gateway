<template>
  <div class="env-settings">
    <RsLoading :loading="loading" overlay block size="lg" />
    <RsCard :title="t('moduleName')" variant="outlined">
      <RsTabs
        v-model="activeTab"
        :items="tabItems"
        variant="line"
        size="md"
        borderless
        content-gap="lg"
      >
        <template #retention>
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

        <template #retentionJob>
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

        <template #webTimeout>
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
      </RsTabs>
    </RsCard>
  </div>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import {
  RsAlert,
  RsButton,
  RsCard,
  RsForm,
  RsFormItem,
  RsInputNumber,
  RsLoading,
  RsSwitch,
  RsTabs,
  RsTimePicker,
  RsTooltip,
  type RsTabItem,
} from '@/ui'
import { computed, onMounted, ref } from 'vue'
import { useEnvironmentSettings } from './hooks'
import type { RetentionSettings, WebTimeoutSettings } from './types'

defineOptions({ name: 'EnvironmentSettings' })

type RetentionDayKey = Exclude<keyof RetentionSettings, 'currentVersion'>
type WebTimeoutKey = Exclude<keyof WebTimeoutSettings, 'currentVersion'>

const { t } = useModuleI18n('hub0009')
const activeTab = ref('retention')
const tabItems = computed<RsTabItem[]>(() => [
  { value: 'retention', label: t('tabs.retention') },
  { value: 'retentionJob', label: t('tabs.retentionJob') },
  { value: 'webTimeout', label: t('tabs.webTimeout') },
])

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
  canEdit,
  savingRetention,
  savingRetentionJob,
  savingWebTimeout,
  fetchSettings,
  saveGroup,
} = useEnvironmentSettings()

onMounted(() => {
  fetchSettings()
})
</script>

<style lang="scss" scoped>
.env-settings {
  position: relative;
  padding: 16px;
  max-width: 880px;
  margin: 0 auto;
}

.hint {
  margin-bottom: 16px;
}

.settings-form {
  max-width: 560px;
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
