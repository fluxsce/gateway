<template>
  <RsDrawer
    v-model:open="drawerVisible"
    :title="drawerTitle"
    side="right"
    size="lg"
    :teleport-to="`#${moduleId}`"
    :show-overlay="false"
    :close-on-overlay-click="false"
  >
    <div class="center-token-drawer">
      <p class="center-token-drawer__lead">
        令牌由服务端颁发并绑定当前实例。明文只出现一次，请立即复制到 SDK 的 setAuthToken。
      </p>
      <RsAlert
        v-if="instance && instance.enableAuth !== 'Y'"
        type="warning"
        title="当前未启用认证"
      >
        已颁发的令牌不会被校验。请先在接入设置中打开「启用认证」。
      </RsAlert>

      <section v-if="canIssue" class="center-token-drawer__section">
        <h3>颁发新令牌</h3>
        <div class="center-token-drawer__form">
          <RsInput v-model="tokenName" placeholder="令牌名称，如 订单服务" clearable />
          <RsInputNumber v-model="expireDays" :min="0" :max="3650" placeholder="有效天数" />
          <RsButton variant="primary" :loading="issuing" @click="emitIssue">
            颁发
          </RsButton>
        </div>
        <p class="center-token-drawer__hint">有效天数填 0 表示长期有效。</p>
      </section>

      <section v-if="issuedToken" class="center-token-drawer__issued">
        <h3>刚刚颁发，请立即复制</h3>
        <code>{{ issuedToken.tokenValue }}</code>
        <div class="center-token-drawer__issued-actions">
          <RsButton size="sm" @click="copyToken">复制令牌</RsButton>
          <RsButton size="sm" variant="ghost" @click="copySdk">复制 SDK 代码</RsButton>
        </div>
      </section>

      <section class="center-token-drawer__section">
        <div class="center-token-drawer__section-head">
          <h3>已颁发令牌</h3>
          <RsTag variant="info" size="sm">{{ tokens.length }} 条</RsTag>
        </div>
        <RsLoading v-if="loading" block size="lg" />
        <RsEmpty v-else-if="!tokens.length" description="还没有颁发过访问令牌" />
        <RsGrid
          v-else
          module-id="hub0040:token"
          :data="tokens"
          :columns="columns"
          :selectable="false"
          row-key="tokenId"
          height="360px"
        />
      </section>
    </div>
  </RsDrawer>
</template>

<script lang="ts" setup>
import { RsGrid, type RsGridColumn } from '@/components/rs-grid'
import { copyToClipboardAsync } from '@/utils/clipboard'
import { formatDate } from '@/utils/format'
import { RsAlert, RsButton, RsDrawer, RsEmpty, RsInput, RsInputNumber, RsLoading, RsTag } from '@/ui'
import { computed, h, ref, watch } from 'vue'
import type { CenterAuthToken, CenterIssuedAuthToken, ServiceCenterInstance } from '../types'

defineOptions({
  name: 'CenterTokenDrawer',
})

const props = withDefaults(defineProps<{
  visible: boolean
  loading?: boolean
  issuing?: boolean
  canIssue?: boolean
  instance?: ServiceCenterInstance | null
  tokens?: CenterAuthToken[]
  issuedToken?: CenterIssuedAuthToken | null
  moduleId?: string
}>(), {
  loading: false,
  issuing: false,
  canIssue: false,
  instance: null,
  tokens: () => [],
  issuedToken: null,
  moduleId: 'hub0040',
})

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'issue', payload: { tokenName: string; expireDays: number }): void
  (e: 'revoke', token: CenterAuthToken): void
}>()

const tokenName = ref('访问令牌')
const expireDays = ref(0)

watch(
  () => props.visible,
  (open) => {
    if (open) {
      tokenName.value = '访问令牌'
      expireDays.value = 0
    }
  },
)

const drawerVisible = computed({
  get: () => props.visible,
  set: (value: boolean) => emit('update:visible', value),
})

const drawerTitle = computed(() => {
  const name = props.instance?.instanceName
  const env = props.instance?.environment
  if (name && env) return `访问令牌 · ${name} · ${env}`
  if (name) return `访问令牌 · ${name}`
  return '访问令牌'
})

const columns: RsGridColumn<CenterAuthToken>[] = [
  { key: 'tokenName', title: '名称', ellipsis: true },
  { key: 'tokenPreview', title: '令牌', width: 160 },
  { key: 'statusText', title: '状态', width: 90 },
  {
    key: 'expireTime',
    title: '到期',
    width: 170,
    formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '长期'),
  },
  {
    key: 'addTime',
    title: '颁发时间',
    width: 170,
    formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-'),
  },
  {
    key: 'actions',
    title: '操作',
    width: 80,
    render: (row) => {
      if (!props.canIssue || row.statusFlag !== 'Y') return h('span', { class: 'center-token-drawer__muted' }, '—')
      return h(RsButton, {
        size: 'sm',
        variant: 'danger',
        onClick: () => emit('revoke', row),
      }, () => '吊销')
    },
  },
]

function emitIssue() {
  emit('issue', {
    tokenName: tokenName.value,
    expireDays: Number(expireDays.value || 0),
  })
}

function copyToken() {
  const value = props.issuedToken?.tokenValue
  if (!value) return
  void copyToClipboardAsync(value, { successMessage: '令牌已复制' })
}

function copySdk() {
  const value = props.issuedToken?.tokenValue
  if (!value) return
  void copyToClipboardAsync(`.setAuthToken("${value}")`, { successMessage: 'SDK 代码已复制' })
}
</script>

<style lang="scss" scoped>
.center-token-drawer {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-lg);
  min-height: 0;
}

.center-token-drawer__lead,
.center-token-drawer__hint {
  margin: 0;
  font-size: 13px;
  line-height: 20px;
  color: var(--g-text-secondary);
}

.center-token-drawer__section h3 {
  margin: 0 0 var(--g-space-sm);
  font-size: 14px;
  font-weight: 600;
}

.center-token-drawer__section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--g-space-sm);
}

.center-token-drawer__section-head h3 {
  margin: 0;
}

.center-token-drawer__form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 120px auto;
  gap: 8px;
  align-items: center;
}

.center-token-drawer__issued {
  padding: 12px;
  border-radius: 12px;
  background: rgba(48, 209, 88, 0.1);
}

.center-token-drawer__issued h3 {
  margin: 0 0 8px;
  font-size: 14px;
  font-weight: 600;
}

.center-token-drawer__issued code {
  display: block;
  overflow-wrap: anywhere;
  font-size: 13px;
  line-height: 20px;
}

.center-token-drawer__issued-actions {
  display: flex;
  gap: 8px;
  margin-top: 10px;
}

.center-token-drawer__muted {
  color: var(--g-text-tertiary);
}

@media (max-width: 720px) {
  .center-token-drawer__form {
    grid-template-columns: 1fr;
  }
}
</style>
