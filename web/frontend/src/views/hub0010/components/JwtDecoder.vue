<template>
  <ToolPage
    icon="key-round"
    :title="t('tools.jwtDecoder.name')"
    :description="t('tools.jwtDecoder.description')"
    :error="errorText"
    :loading="loading"
    :denied="denied"
    :failed="failed"
  >
    <template #actions>
      <RsButton variant="primary" icon="scan-text" @click="decode">{{ t('work.decode') }}</RsButton>
    </template>

    <div class="jwt-decoder">
      <ToolEditor v-model="input" class="jwt-decoder__token" :label="t('work.input')" :hint="t('work.jwtHint')" />
      <RsSplitPane class="jwt-decoder__split" fill with-handle :panes="panes">
        <template #header>
          <ToolEditor v-model="header" readonly :label="t('work.header')" />
        </template>
        <template #payload>
          <ToolEditor v-model="payload" readonly :label="t('work.payload')" />
        </template>
      </RsSplitPane>
    </div>
  </ToolPage>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsButton, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ref } from 'vue'
import { useGrantedTool } from '../useGrantedTool'
import ToolEditor from './ToolEditor.vue'
import ToolPage from './ToolPage.vue'

defineOptions({ name: 'JwtDecoder' })

const panes: RsSplitPaneItem[] = [
  { key: 'header', min: 24 },
  { key: 'payload', min: 32 },
]

const { t } = useModuleI18n('hub0010')
const { loading, denied, failed } = useGrantedTool('jwt-decoder')
const input = ref('')
const header = ref('')
const payload = ref('')
const errorText = ref('')

const decodePart = (part: string) => {
  const normalized = part.replace(/-/g, '+').replace(/_/g, '/')
  const padded = normalized + '='.repeat((4 - (normalized.length % 4)) % 4)
  const binary = atob(padded)
  const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0))
  return JSON.stringify(JSON.parse(new TextDecoder().decode(bytes)), null, 2)
}

const decode = () => {
  const parts = input.value.trim().split('.')
  if (parts.length < 2 || !parts[0] || !parts[1]) {
    header.value = ''
    payload.value = ''
    errorText.value = t('work.invalidJwt')
    return
  }
  try {
    header.value = decodePart(parts[0])
    payload.value = decodePart(parts[1])
    errorText.value = ''
  } catch {
    header.value = ''
    payload.value = ''
    errorText.value = t('work.invalidJwt')
  }
}
</script>

<style lang="scss" scoped>
.jwt-decoder {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
  min-height: 0;
}

.jwt-decoder__token {
  flex: 0 0 132px;
}

.jwt-decoder__split {
  flex: 1;
  width: 100%;
  min-height: 0;
}
</style>
