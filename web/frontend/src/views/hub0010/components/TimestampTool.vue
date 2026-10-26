<template>
  <ToolPage
    icon="clock"
    :title="t('tools.timestamp.name')"
    :description="t('tools.timestamp.description')"
    :error="errorText"
    :loading="loading"
    :denied="denied"
    :failed="failed"
  >
    <template #actions>
      <RsButton variant="primary" icon="arrow-right-left" @click="convert">{{ t('work.format') }}</RsButton>
      <RsButton variant="ghost" icon="timer" @click="useNow">{{ t('work.now') }}</RsButton>
    </template>

    <div class="timestamp-tool">
      <label class="timestamp-tool__composer">
        <span>{{ t('work.input') }}</span>
        <RsInput v-model="input" size="lg" clearable :placeholder="t('work.timeHint')" @keyup.enter="convert" />
      </label>

      <div class="timestamp-tool__grid">
        <RsStatCard
          v-for="card in cards"
          :key="card.key"
          :label="card.label"
          :value="card.value || '—'"
          :accent="card.accent"
          size="lg"
        >
          <RsButton
            class="timestamp-tool__copy"
            variant="text"
            size="sm"
            icon="copy"
            icon-only
            :disabled="!card.value"
            :tooltip="t('work.copy')"
            :aria-label="t('work.copy')"
            @click="copy(card.value)"
          />
        </RsStatCard>
      </div>
    </div>
  </ToolPage>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsButton, RsInput, RsStatCard, type RsStatAccent } from '@/ui'
import { computed, ref } from 'vue'
import { copyToClipboardAsync } from '@/utils/clipboard'
import { useGrantedTool } from '../useGrantedTool'
import ToolPage from './ToolPage.vue'

defineOptions({ name: 'TimestampTool' })

const { t } = useModuleI18n('hub0010')
const { loading, denied, failed } = useGrantedTool('timestamp')
const input = ref('')
const seconds = ref('')
const millis = ref('')
const local = ref('')
const utc = ref('')
const errorText = ref('')

const cards = computed(() => [
  { key: 'seconds', label: t('work.seconds'), value: seconds.value, accent: 'primary' as RsStatAccent },
  { key: 'millis', label: t('work.millis'), value: millis.value, accent: 'info' as RsStatAccent },
  { key: 'local', label: t('work.local'), value: local.value, accent: 'success' as RsStatAccent },
  { key: 'utc', label: t('work.utc'), value: utc.value, accent: 'warning' as RsStatAccent },
])

const apply = (date: Date) => {
  if (Number.isNaN(date.getTime())) {
    errorText.value = t('work.invalidTime')
    return
  }
  errorText.value = ''
  seconds.value = String(Math.floor(date.getTime() / 1000))
  millis.value = String(date.getTime())
  local.value = date.toLocaleString()
  utc.value = date.toISOString()
}

const convert = () => {
  const text = input.value.trim()
  if (!text) {
    errorText.value = t('work.invalidTime')
    return
  }
  if (/^\d+$/.test(text)) {
    const n = Number(text)
    apply(new Date(text.length >= 13 ? n : n * 1000))
    return
  }
  apply(new Date(text))
}

const useNow = () => {
  const now = new Date()
  input.value = String(now.getTime())
  apply(now)
}

const copy = (value: string) => {
  if (!value) return
  void copyToClipboardAsync(value, {
    successMessage: t('work.copied'),
    errorMessage: t('work.copyFail'),
  })
}
</script>

<style lang="scss" scoped>
.timestamp-tool {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 960px;
  container-type: inline-size;
}

.timestamp-tool__composer {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
  border: 1px solid var(--rs-border);
  border-radius: 12px;
  background: var(--rs-surface);
  box-shadow: var(--rs-shadow-sm);

  span {
    font-size: 12px;
    font-weight: 650;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--rs-muted);
  }
}

.timestamp-tool__grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.timestamp-tool__copy {
  margin-top: 8px;
}

@container (max-width: 640px) {
  .timestamp-tool__grid {
    grid-template-columns: 1fr;
  }
}
</style>
