<template>
  <ToolPage
    icon="binary"
    :title="t('tools.base64.name')"
    :description="t('tools.base64.description')"
    :error="errorText"
    :loading="loading"
    :denied="denied"
    :failed="failed"
  >
    <template #actions>
      <RsButton variant="primary" icon="arrow-right" @click="encode">{{ t('work.encode') }}</RsButton>
      <RsButton variant="ghost" icon="arrow-left" @click="decode">{{ t('work.decode') }}</RsButton>
      <RsButton variant="text" icon="arrow-left-right" :disabled="!output" @click="swap">
        {{ t('work.swap') }}
      </RsButton>
    </template>

    <RsSplitPane class="base64-tool__split" fill with-handle :panes="panes">
      <template #input>
        <ToolEditor v-model="input" :label="t('work.input')" />
      </template>
      <template #output>
        <ToolEditor v-model="output" readonly :label="t('work.output')" />
      </template>
    </RsSplitPane>
  </ToolPage>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsButton, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ref } from 'vue'
import { useGrantedTool } from '../useGrantedTool'
import ToolEditor from './ToolEditor.vue'
import ToolPage from './ToolPage.vue'

defineOptions({ name: 'Base64Tool' })

const panes: RsSplitPaneItem[] = [
  { key: 'input', min: 28 },
  { key: 'output', min: 28 },
]

const { t } = useModuleI18n('hub0010')
const { loading, denied, failed } = useGrantedTool('base64')
const input = ref('')
const output = ref('')
const errorText = ref('')

const encode = () => {
  errorText.value = ''
  const bytes = new TextEncoder().encode(input.value)
  let binary = ''
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte)
  })
  output.value = btoa(binary)
}

const decode = () => {
  try {
    const binary = atob(input.value.trim())
    const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0))
    output.value = new TextDecoder().decode(bytes)
    errorText.value = ''
  } catch {
    output.value = ''
    errorText.value = t('work.invalidBase64')
  }
}

const swap = () => {
  input.value = output.value
  output.value = ''
  errorText.value = ''
}
</script>

<style lang="scss" scoped>
.base64-tool__split {
  width: 100%;
  height: 100%;
  min-height: 0;
}
</style>
