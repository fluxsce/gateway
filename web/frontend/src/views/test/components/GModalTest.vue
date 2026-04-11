<template>
  <div class="gmodal-test-page">
    <div class="page-header">
      <h1>GModal 弹窗测试</h1>
      <p>
        默认仅内容区滚动；支持 <code>width</code> / <code>height</code>；全屏；东 / 南 / 东南 边框拖拽缩放（<code>resizable</code>）。
        GModal 默认 <code>trap-focus=false</code>，避免 vueuc 焦点陷阱占位 div 与 <code>aria-hidden</code> 在 Chrome 下打控制台警告；需要键盘焦点锁在弹窗内时可传 <code>:trap-focus="true"</code>。
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">01</span>
          <div>
            <h2>默认（仅内容滚动）</h2>
            <p>未传 <code>height</code> 时外壳 <code>max-height: 80vh</code>，头尾固定，中间 <code>.g-modal__body</code> 滚动。</p>
          </div>
        </div>
        <n-button @click="defaultScroll.visible = true">打开</n-button>
        <GModal
          v-model:visible="defaultScroll.visible"
          title="默认滚动"
          :width="520"
          :show-footer="false"
          :mask="true"
        >
          <div class="scroll-demo">
            <p v-for="n in 40" :key="n">段落 {{ n }}：请只在此区域滚动，标题栏不随内容滚动。</p>
          </div>
        </GModal>
      </section>

      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">02</span>
          <div>
            <h2>固定宽高</h2>
            <p><code>:width="720"</code> <code>:height="400"</code>，内容超出时仍只有 body 滚动。</p>
          </div>
        </div>
        <n-button @click="fixedSize.visible = true">打开</n-button>
        <GModal
          v-model:visible="fixedSize.visible"
          title="固定 720×400"
          :width="720"
          :height="400"
          :show-footer="false"
          :mask="true"
        >
          <div class="scroll-demo">
            <p v-for="n in 30" :key="n">行 {{ n }} — 固定高度弹窗内的滚动区域。</p>
          </div>
        </GModal>
      </section>

      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">03</span>
          <div>
            <h2>边框缩放</h2>
            <p>
              <code>resizable</code>：鼠标移到右缘、下缘、右下角透明边缘时出现调整光标，拖拽可改宽高；松手触发 <code>@resize</code>。为避免与标题栏拖拽混淆，本例关闭 <code>draggable</code>。
            </p>
          </div>
        </div>
        <n-button @click="resizableModal.visible = true">打开</n-button>
        <p v-if="lastResizeText" class="hint">{{ lastResizeText }}</p>
        <GModal
          v-model:visible="resizableModal.visible"
          title="可缩放"
          :width="560"
          :height="360"
          :resizable="true"
          :resize-min-width="360"
          :resize-min-height="240"
          :show-footer="false"
          :mask="true"
          :draggable="false"
          @resize="onResize"
        >
          <div class="scroll-demo scroll-demo--resize-demo">
            <p class="scroll-demo__resize-hint">
              将鼠标移到窗口<strong>右边缘、下边缘、右下角</strong>，指针变为双向箭头后按住拖拽即可缩放（非标题栏拖拽）。
            </p>
            <p v-for="n in 12" :key="n">段落 {{ n }}：内容区可滚动，用于观察缩放后 body 高度变化。</p>
          </div>
        </GModal>
      </section>

      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">04</span>
          <div>
            <h2>全屏</h2>
            <p>标题栏右侧展开图标；全屏后内容区纵向铺满。</p>
          </div>
        </div>
        <n-button @click="fullscreenModal.visible = true">打开</n-button>
        <GModal
          v-model:visible="fullscreenModal.visible"
          title="全屏与列表"
          :width="900"
          :height="480"
          :show-footer="false"
          :mask="true"
        >
          <div class="table-wrap">
            <table class="demo-table">
              <thead>
                <tr>
                  <th>列 A</th>
                  <th>列 B</th>
                  <th>列 C</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="i in 24" :key="i">
                  <td>单元 {{ i }}-1</td>
                  <td>单元 {{ i }}-2</td>
                  <td>单元 {{ i }}-3</td>
                </tr>
              </tbody>
            </table>
          </div>
        </GModal>
      </section>

      <section class="test-section">
        <div class="section-header">
          <span class="section-badge">05</span>
          <div>
            <h2>带底部按钮</h2>
            <p>有 footer 时仅内容区滚动，底部操作栏固定。</p>
          </div>
        </div>
        <n-button @click="withFooter.visible = true">打开</n-button>
        <GModal
          v-model:visible="withFooter.visible"
          title="确认操作"
          :width="480"
          :height="420"
          :mask="true"
          @confirm="withFooter.visible = false"
          @cancel="withFooter.visible = false"
        >
          <div class="scroll-demo">
            <p v-for="n in 25" :key="n">说明文案 {{ n }} …</p>
          </div>
        </GModal>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { GModal } from '@/components/gmodal'
import { NButton } from 'naive-ui'
import { reactive, ref } from 'vue'

defineOptions({ name: 'GModalTest' })

const defaultScroll = reactive({ visible: false })
const fixedSize = reactive({ visible: false })
const resizableModal = reactive({ visible: false })
const fullscreenModal = reactive({ visible: false })
const withFooter = reactive({ visible: false })

const lastResizeText = ref('')

function onResize(payload: { width: number; height: number }) {
  lastResizeText.value = `上次 resize：${payload.width} × ${payload.height} px`
}
</script>

<style scoped lang="scss">
.gmodal-test-page {
  padding: var(--g-padding-xxl, 24px);
  max-width: 960px;
  margin: 0 auto;
  min-height: 100%;
}

.page-header {
  margin-bottom: var(--g-space-xl);

  h1 {
    font-size: 24px;
    font-weight: 600;
    margin: 0 0 var(--g-space-sm);
  }

  p {
    margin: 0;
    font-size: var(--g-font-size-sm);
    color: var(--g-text-secondary);
    line-height: 1.55;

    code {
      font-size: 12px;
      padding: 0 4px;
      border-radius: 4px;
      background: var(--g-bg-tertiary, rgba(0, 0, 0, 0.06));
    }
  }
}

.test-sections {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-xl);
}

.test-section {
  padding: var(--g-padding-lg);
  background: var(--g-bg-secondary);
  border: 1px solid var(--g-border-primary);
  border-radius: var(--g-radius-lg);
}

.section-header {
  display: flex;
  gap: var(--g-space-md);
  margin-bottom: var(--g-space-md);
}

.section-badge {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: var(--g-primary-light);
  color: var(--g-primary);
  font-size: 12px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.section-header h2 {
  margin: 0 0 4px;
  font-size: var(--g-font-size-lg);
}

.section-header p {
  margin: 0;
  font-size: var(--g-font-size-sm);
  color: var(--g-text-secondary);
  line-height: 1.5;

  code {
    font-size: 12px;
    padding: 0 3px;
    border-radius: 3px;
    background: var(--g-bg-tertiary, rgba(0, 0, 0, 0.06));
  }
}

.scroll-demo {
  line-height: 1.6;
  font-size: 13px;
  color: var(--g-text-primary);

  p {
    margin: 0 0 8px;
  }

  &--short p {
    margin: 0;
  }

  &--resize-demo {
    min-height: 120px;
  }

  &__resize-hint {
    margin: 0 0 12px !important;
    padding: 8px 10px;
    border-radius: 6px;
    background: var(--g-bg-tertiary, rgba(0, 0, 0, 0.06));
    border: 1px dashed var(--g-border-primary);
    font-size: 12px;
    color: var(--g-text-secondary);
    line-height: 1.5;
  }
}

.table-wrap {
  overflow: auto;
  max-height: 100%;
}

.demo-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;

  th,
  td {
    border: 1px solid var(--g-border-primary);
    padding: 8px 10px;
    text-align: left;
  }

  th {
    background: var(--g-bg-tertiary, rgba(0, 0, 0, 0.04));
  }
}

.hint {
  margin-top: var(--g-space-sm);
  font-size: 12px;
  color: var(--g-text-secondary);
}
</style>
