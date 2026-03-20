<template>
  <div class="gdialog-test-page">

    <div class="page-header">
      <h1>GDialog 对话框测试</h1>
      <p>标题/副标题、图标、渐变头部、宽度、滚动、confirmLoading、自定义插槽</p>
    </div>

    <div class="test-sections">

      <!-- ① 基础 -->
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">01</span>
          <div>
            <h2>基础对话框</h2>
            <p>默认配置：有标题、取消 + 确定按钮，点击遮罩不关闭</p>
          </div>
        </div>
        <div class="button-group">
          <NButton @click="basic.show = true">打开基础对话框</NButton>
        </div>
        <GDialog
          v-model:show="basic.show"
          title="基础对话框"
          :width="480"
          @confirm="onConfirm('基础')"
          @cancel="onCancel('基础')"
          @close="onClose('基础')"
        >
          这是一个基础对话框，包含默认的取消和确认按钮。
        </GDialog>
      </section>

      <!-- ② 副标题位置 -->
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">02</span>
          <div>
            <h2>副标题位置</h2>
            <p><code>subtitle-position="header"</code> 显示在标题下方，<code>footer</code> 显示在底部左侧</p>
          </div>
        </div>
        <div class="button-group">
          <NButton @click="subtitleHeader.show = true">副标题在头部</NButton>
          <NButton @click="subtitleFooter.show = true">副标题在底部</NButton>
        </div>
        <GDialog v-model:show="subtitleHeader.show" title="主标题" subtitle="副标题显示在头部区域" subtitle-position="header" :width="480">
          副标题紧跟在标题下方（header 区域）。
        </GDialog>
        <GDialog v-model:show="subtitleFooter.show" title="主标题" subtitle="副标题显示在底部左侧" subtitle-position="footer" :width="480">
          副标题显示在 footer 左侧，按钮在右侧。
        </GDialog>
      </section>

      <!-- ③ 头部图标 -->
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">03</span>
          <div>
            <h2>头部图标</h2>
            <p>通过 <code>:icon</code> 传入 Ionicons5 组件，图标显示在标题左侧</p>
          </div>
        </div>
        <div class="button-group">
          <NButton @click="iconInfo.show    = true">信息</NButton>
          <NButton @click="iconSuccess.show = true">成功</NButton>
          <NButton @click="iconWarning.show = true">警告</NButton>
          <NButton @click="iconError.show   = true">错误</NButton>
        </div>
        <GDialog v-model:show="iconInfo.show"    title="信息提示" :icon="InformationCircleOutline" :width="440">这是一条信息提示内容。</GDialog>
        <GDialog v-model:show="iconSuccess.show" title="操作成功" :icon="CheckmarkCircleOutline"   :width="440">操作已成功完成。</GDialog>
        <GDialog v-model:show="iconWarning.show" title="注意"     :icon="WarningOutline"           :width="440">请注意此操作的潜在风险。</GDialog>
        <GDialog v-model:show="iconError.show"   title="操作失败" :icon="CloseCircleOutline"       :width="440">操作执行过程中发生了错误。</GDialog>
      </section>

      <!-- ④ 渐变头部 -->
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">04</span>
          <div>
            <h2>渐变头部</h2>
            <p><code>header-style="gradient"</code> — 头部背景变为品牌渐变色，文字白色</p>
          </div>
        </div>
        <div class="button-group">
          <NButton type="primary" @click="gradient.show    = true">仅渐变</NButton>
          <NButton type="primary" @click="gradientIcon.show = true">渐变 + 图标</NButton>
          <NButton type="primary" @click="gradientSub.show  = true">渐变 + 副标题</NButton>
        </div>
        <GDialog v-model:show="gradient.show"     title="渐变头部"            header-style="gradient"                                          :width="500">header-style="gradient" 纯文字标题。</GDialog>
        <GDialog v-model:show="gradientIcon.show"  title="渐变 + 图标"         header-style="gradient" :icon="StarOutline"                      :width="500">渐变头部配合图标，图标自动变为白色。</GDialog>
        <GDialog v-model:show="gradientSub.show"   title="渐变 + 副标题"       header-style="gradient" :icon="RocketOutline" subtitle="副标题文字" subtitle-position="header" :width="500">渐变 + 图标 + 副标题全部白色显示。</GDialog>
      </section>

      <!-- ⑤ 宽度 -->
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">05</span>
          <div>
            <h2>宽度配置</h2>
            <p><code>width</code> 支持数字（px）或字符串（如 <code>'80vw'</code>）</p>
          </div>
        </div>
        <div class="button-group">
          <NButton @click="widthSm.show = true">300px</NButton>
          <NButton @click="widthMd.show = true">600px</NButton>
          <NButton @click="widthLg.show = true">900px</NButton>
          <NButton @click="widthVw.show = true">80vw</NButton>
        </div>
        <GDialog v-model:show="widthSm.show" title="宽 300px" :width="300">内容区域随宽度自适应。</GDialog>
        <GDialog v-model:show="widthMd.show" title="宽 600px" :width="600">内容区域随宽度自适应。</GDialog>
        <GDialog v-model:show="widthLg.show" title="宽 900px" :width="900">内容区域随宽度自适应。</GDialog>
        <GDialog v-model:show="widthVw.show" title="宽 80vw"  width="80vw">内容区域随宽度自适应。</GDialog>
      </section>

      <!-- ⑥ 关闭行为 -->
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">06</span>
          <div>
            <h2>关闭行为</h2>
            <p><code>maskClosable</code>、<code>closeOnEsc</code>、<code>closable</code> 的组合控制</p>
          </div>
        </div>
        <div class="button-group">
          <NButton @click="closeMask.show = true">点遮罩关闭</NButton>
          <NButton @click="closeEsc.show  = true">ESC 关闭</NButton>
          <NButton @click="closeNone.show = true">仅按钮关闭</NButton>
          <NButton @click="noX.show       = true">无右上角 ×</NButton>
        </div>
        <GDialog v-model:show="closeMask.show" title="点遮罩可关闭"     :mask-closable="true"  :close-on-esc="false" :width="440">点击遮罩区域即可关闭。</GDialog>
        <GDialog v-model:show="closeEsc.show"  title="ESC 可关闭"      :mask-closable="false" :close-on-esc="true"  :width="440">按下 Esc 键即可关闭。</GDialog>
        <GDialog v-model:show="closeNone.show" title="仅按钮可关闭"    :mask-closable="false" :close-on-esc="false" :width="440">只能点底部按钮关闭。</GDialog>
        <GDialog v-model:show="noX.show"       title="无右上角关闭按钮" :closable="false"      :width="440">closable=false 时右上角 × 不显示。</GDialog>
      </section>

      <!-- ⑦ 长内容滚动 -->
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">07</span>
          <div>
            <h2>长内容 + 滚动</h2>
            <p><code>show-scrollbar</code> + <code>content-max-height</code> 限制可视高度</p>
          </div>
        </div>
        <div class="button-group">
          <NButton @click="scroll.show = true">打开长内容</NButton>
        </div>
        <GDialog v-model:show="scroll.show" title="长内容对话框" :show-scrollbar="true" content-max-height="280px" :width="520">
          <div v-for="i in 30" :key="i" class="scroll-row">
            第 {{ i }} 行 — Lorem ipsum dolor sit amet, consectetur adipiscing elit.
          </div>
        </GDialog>
      </section>

      <!-- ⑧ confirmLoading -->
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">08</span>
          <div>
            <h2>confirmLoading</h2>
            <p>点击确认后按钮进入 loading，模拟 2 秒异步操作后自动关闭</p>
          </div>
        </div>
        <div class="button-group">
          <NButton type="primary" @click="loadingDialog.show = true">模拟异步提交</NButton>
        </div>
        <GDialog
          v-model:show="loadingDialog.show"
          title="异步提交示例"
          subtitle="提交需要一点时间，请耐心等待"
          subtitle-position="header"
          :icon="CloudUploadOutline"
          :confirm-loading="loadingDialog.loading"
          :auto-close-on-confirm="false"
          :width="480"
          @confirm="handleLoadingConfirm"
          @cancel="loadingDialog.show = false"
        >
          点击确认按钮，模拟 2 秒异步操作，完成后自动关闭对话框。
        </GDialog>
      </section>

      <!-- ⑨ 无 footer -->
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">09</span>
          <div>
            <h2>无底部操作区</h2>
            <p><code>show-footer="false"</code> — 适合纯展示场景，点击遮罩或 × 关闭</p>
          </div>
        </div>
        <div class="button-group">
          <NButton @click="noFooter.show = true">无 Footer</NButton>
        </div>
        <GDialog v-model:show="noFooter.show" title="纯展示对话框" :show-footer="false" :mask-closable="true" :width="480">
          该对话框没有底部按钮，点击遮罩或右上角 × 关闭。
        </GDialog>
      </section>

      <!-- ⑩ 自定义 footer 插槽 -->
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">10</span>
          <div>
            <h2>自定义 footer 插槽</h2>
            <p>通过 <code>#footer</code> 插槽完全替换底部按钮布局</p>
          </div>
        </div>
        <div class="button-group">
          <NButton @click="customFooter.show = true">自定义 Footer</NButton>
        </div>
        <GDialog v-model:show="customFooter.show" title="自定义底部" :show-footer="true" :width="520">
          通过 <code>#footer</code> 插槽可以完全自定义底部操作区的按钮数量和样式。
          <template #footer>
            <div class="custom-footer">
              <NButton size="small" @click="customFooter.show = false">取消</NButton>
              <NButton size="small" type="warning" @click="customFooter.show = false">暂存草稿</NButton>
              <NButton size="small" type="primary" @click="customFooter.show = false">立即发布</NButton>
            </div>
          </template>
        </GDialog>
      </section>

      <!-- ⑪ header-extra 插槽 -->
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">11</span>
          <div>
            <h2>header-extra 插槽</h2>
            <p>在右上角关闭按钮前插入自定义操作（如分享、全屏等）</p>
          </div>
        </div>
        <div class="button-group">
          <NButton @click="headerExtra.show = true">header-extra 插槽</NButton>
        </div>
        <GDialog v-model:show="headerExtra.show" title="自定义头部右侧" :width="520">
          头部右侧插入了分享图标按钮。
          <template #header-extra>
            <NButton quaternary circle size="tiny" @click="onHeaderExtraClick">
              <template #icon><NIcon :size="16"><ShareOutline /></NIcon></template>
            </NButton>
          </template>
        </GDialog>
      </section>

    </div>

    <!-- 事件日志 -->
    <div class="event-log">
      <div class="event-log__header">
        <span>事件日志</span>
        <NButton size="tiny" quaternary @click="eventLog = []">清空</NButton>
      </div>
      <div class="event-log__body">
        <div v-if="eventLog.length === 0" class="event-log__empty">暂无事件，打开并操作对话框后此处显示记录</div>
        <div
          v-for="(log, i) in eventLog"
          :key="i"
          class="event-log__item"
          :class="`event-log__item--${log.type}`"
        >
          <span class="event-log__time">{{ log.time }}</span>
          <span class="event-log__msg">{{ log.msg }}</span>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { GDialog } from '@/components/gdialog'
import {
  CheckmarkCircleOutline,
  CloseCircleOutline,
  CloudUploadOutline,
  InformationCircleOutline,
  RocketOutline,
  ShareOutline,
  StarOutline,
  WarningOutline,
} from '@vicons/ionicons5'
import { NButton, NIcon, useMessage } from 'naive-ui'
import { reactive, ref } from 'vue'

defineOptions({ name: 'GDialogTest' })

const message = useMessage()

// ─── 对话框状态 ────────────────────────────────────────────────────────────────

const basic          = reactive({ show: false })
const subtitleHeader = reactive({ show: false })
const subtitleFooter = reactive({ show: false })
const iconInfo       = reactive({ show: false })
const iconSuccess    = reactive({ show: false })
const iconWarning    = reactive({ show: false })
const iconError      = reactive({ show: false })
const gradient       = reactive({ show: false })
const gradientIcon   = reactive({ show: false })
const gradientSub    = reactive({ show: false })
const widthSm        = reactive({ show: false })
const widthMd        = reactive({ show: false })
const widthLg        = reactive({ show: false })
const widthVw        = reactive({ show: false })
const closeMask      = reactive({ show: false })
const closeEsc       = reactive({ show: false })
const closeNone      = reactive({ show: false })
const noX            = reactive({ show: false })
const scroll         = reactive({ show: false })
const loadingDialog  = reactive({ show: false, loading: false })
const noFooter       = reactive({ show: false })
const customFooter   = reactive({ show: false })
const headerExtra    = reactive({ show: false })

// ─── 事件日志 ─────────────────────────────────────────────────────────────────

interface LogEntry { time: string; msg: string; type: 'confirm' | 'cancel' | 'close' | 'info' }
const eventLog = ref<LogEntry[]>([])

function addLog(msg: string, type: LogEntry['type'] = 'info') {
  const now = new Date()
  const hh = now.getHours().toString().padStart(2, '0')
  const mm = now.getMinutes().toString().padStart(2, '0')
  const ss = now.getSeconds().toString().padStart(2, '0')
  eventLog.value.unshift({ time: `${hh}:${mm}:${ss}`, msg, type })
  if (eventLog.value.length > 40) eventLog.value.pop()
}

function onConfirm(name: string) { addLog(`[${name}] confirm 触发`, 'confirm') }
function onCancel(name: string)  { addLog(`[${name}] cancel 触发`,  'cancel')  }
function onClose(name: string)   { addLog(`[${name}] close 触发`,   'close')   }

// ─── confirmLoading 模拟 ──────────────────────────────────────────────────────

function handleLoadingConfirm() {
  loadingDialog.loading = true
  addLog('[异步提交] 开始提交…', 'info')
  setTimeout(() => {
    loadingDialog.loading = false
    loadingDialog.show = false
    message.success('提交成功')
    addLog('[异步提交] 成功，对话框已关闭', 'confirm')
  }, 2000)
}

// ─── header-extra ─────────────────────────────────────────────────────────────

function onHeaderExtraClick() {
  message.info('点击了分享按钮')
  addLog('[header-extra] 分享按钮点击', 'info')
}
</script>

<style scoped lang="scss">
.gdialog-test-page {
  padding: 24px;
  max-width: 1000px;
  margin: 0 auto;
  background: var(--g-bg-primary);
  min-height: 100%;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

// ─── 页头 ─────────────────────────────────────────────────────────────────────

.page-header {
  text-align: center;
  padding-bottom: 20px;
  border-bottom: 2px solid var(--g-border-primary);

  h1 {
    font-size: 24px;
    font-weight: 700;
    margin: 0 0 6px;
    background: linear-gradient(135deg, var(--g-primary) 0%, #8b5cf6 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  p {
    font-size: 13px;
    color: var(--g-text-secondary);
    margin: 0;
  }
}

// ─── 测试区块 ─────────────────────────────────────────────────────────────────

.test-sections {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.test-section {
  background: var(--g-bg-secondary);
  border: 1px solid var(--g-border-primary);
  border-radius: 10px;
  padding: 16px 20px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

// ─── section 头部 ─────────────────────────────────────────────────────────────

.section-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;

  h2 {
    font-size: 14px;
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 2px;
    line-height: 1.4;
  }

  p {
    font-size: 12px;
    color: var(--g-text-secondary);
    margin: 0;
    line-height: 1.5;

    code {
      font-size: 11px;
      background: var(--g-bg-tertiary, rgba(0,0,0,.05));
      border-radius: 3px;
      padding: 1px 4px;
      color: var(--g-primary);
    }
  }
}

.section-badge {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: linear-gradient(135deg, var(--g-primary) 0%, #8b5cf6 100%);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  margin-top: 1px;
  user-select: none;
}

// ─── 按钮组 ───────────────────────────────────────────────────────────────────

.button-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

// ─── 滚动内容行 ───────────────────────────────────────────────────────────────

.scroll-row {
  padding: 5px 0;
  border-bottom: 1px solid var(--g-border-secondary, #f0f0f0);
  font-size: 13px;
  color: var(--g-text-primary);

  &:last-child { border-bottom: none; }
}

// ─── 自定义 footer ────────────────────────────────────────────────────────────

.custom-footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 0 24px;
}

// ─── 事件日志 ─────────────────────────────────────────────────────────────────

.event-log {
  border: 1px solid var(--g-border-primary);
  border-radius: 10px;
  overflow: hidden;
  background: var(--g-bg-secondary);

  &__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 14px;
    background: var(--g-bg-tertiary, rgba(0,0,0,.03));
    border-bottom: 1px solid var(--g-border-primary);
    font-size: 12px;
    font-weight: 600;
    color: var(--g-text-secondary);
    letter-spacing: 0.3px;
  }

  &__body {
    max-height: 180px;
    overflow-y: auto;
    padding: 4px 0;
  }

  &__empty {
    padding: 16px;
    text-align: center;
    font-size: 12px;
    color: var(--g-text-tertiary, #bbb);
  }

  &__item {
    display: flex;
    align-items: baseline;
    gap: 10px;
    padding: 4px 14px;
    font-size: 12px;
    border-left: 3px solid transparent;
    transition: background 0.1s;

    &:hover { background: var(--g-hover-overlay, rgba(0,0,0,.02)); }

    &--confirm { border-left-color: var(--g-success, #18a058); }
    &--cancel  { border-left-color: var(--g-text-secondary, #999); }
    &--close   { border-left-color: var(--g-warning, #f0a020); }
    &--info    { border-left-color: var(--g-primary); }
  }

  &__time {
    flex-shrink: 0;
    font-size: 11px;
    color: var(--g-text-tertiary, #bbb);
    font-variant-numeric: tabular-nums;
    min-width: 60px;
  }

  &__msg {
    color: var(--g-text-primary);
  }
}
</style>
