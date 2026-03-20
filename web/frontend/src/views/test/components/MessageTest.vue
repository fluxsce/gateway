<template>
  <div class="message-test-page">
    <div class="page-header">
      <h1>GMessage / GDialog 测试</h1>
      <p class="page-description">
        消息：$gMessage（info / success / error / warning / loading）。对话框：$gDialog（info / success / error / warning / confirm / create）
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>函数式调用</h2>
        <p class="section-description">
          使用 <code>$gMessage</code> 或 <code>window.$gMessage</code> 调用
        </p>
        <div class="button-group">
          <NButton type="success" @click="handleSuccess">成功消息</NButton>
          <NButton type="error" @click="handleError">错误消息</NButton>
          <NButton type="warning" @click="handleWarning">警告消息</NButton>
          <NButton type="info" @click="handleInfo">信息消息</NButton>
          <NButton type="primary" @click="handleLoading">加载消息</NButton>
        </div>
      </section>

      <section class="test-section">
        <h2>位置测试</h2>
        <p class="section-description">
          不同 position 的消息显示位置（需在对应 Provider 下展示）
        </p>
        <div class="button-group">
          <NButton @click="handlePositionTop">顶部居中</NButton>
          <NButton @click="handlePositionTopLeft">顶部左侧</NButton>
          <NButton @click="handlePositionTopRight">顶部右侧</NButton>
          <NButton @click="handlePositionBottom">底部居中</NButton>
          <NButton @click="handlePositionBottomLeft">底部左侧</NButton>
          <NButton @click="handlePositionBottomRight">底部右侧</NButton>
        </div>
      </section>

      <section class="test-section">
        <h2>消息高级</h2>
        <p class="section-description">
          自定义 duration、closable、onClose；关闭全部
        </p>
        <div class="button-group">
          <NButton @click="handleCustomDuration">自定义持续时间（10 秒）</NButton>
          <NButton @click="handleWithCallback">带 onClose 回调</NButton>
          <NButton type="error" @click="handleDestroyAll">关闭所有消息</NButton>
        </div>
      </section>

      <section class="test-section">
        <h2>对话框（$gDialog）</h2>
        <p class="section-description">
          使用 <code>$gDialog</code> 或 <code>window.$gDialog</code>，返回 Promise&lt;boolean&gt;（确定 true / 取消 false）
        </p>
        <div class="button-group">
          <NButton type="info" @click="handleDialogInfo">info 提示</NButton>
          <NButton type="success" @click="handleDialogSuccess">success 成功</NButton>
          <NButton type="error" @click="handleDialogError">error 错误</NButton>
          <NButton type="warning" @click="handleDialogWarning">warning 警告</NButton>
          <NButton @click="handleDialogConfirm">confirm 确认</NButton>
          <NButton @click="handleDialogCreate">create 自定义</NButton>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { NButton } from 'naive-ui'
import { getCurrentInstance } from 'vue'

defineOptions({ name: 'MessageTest' })

const instance = getCurrentInstance()
const $gMessage = instance?.appContext.config.globalProperties.$gMessage
const $gDialog = instance?.appContext.config.globalProperties.$gDialog

function getMessageApi() {
  return $gMessage ?? (typeof window !== 'undefined' ? (window as any).$gMessage : null)
}

function getDialogApi() {
  return $gDialog ?? (typeof window !== 'undefined' ? (window as any).$gDialog : null)
}

function handleSuccess() {
  getMessageApi()?.success('操作成功！这是一个成功消息示例')
}

function handleError() {
  getMessageApi()?.error('操作失败！这是一个错误消息示例')
}

function handleWarning() {
  getMessageApi()?.warning('请注意！这是一个警告消息示例')
}

function handleInfo() {
  getMessageApi()?.info('提示信息：这是一个信息消息示例')
}

function handleLoading() {
  getMessageApi()?.loading('正在加载中...', { duration: 0 })
  setTimeout(() => {
    getMessageApi()?.success('加载完成！')
  }, 2000)
}

function handlePositionTop() {
  getMessageApi()?.info('顶部居中', { position: 'top' })
}

function handlePositionTopLeft() {
  getMessageApi()?.info('顶部左侧', { position: 'top-left' })
}

function handlePositionTopRight() {
  getMessageApi()?.info('顶部右侧', { position: 'top-right' })
}

function handlePositionBottom() {
  getMessageApi()?.info('底部居中', { position: 'bottom' })
}

function handlePositionBottomLeft() {
  getMessageApi()?.info('底部左侧', { position: 'bottom-left' })
}

function handlePositionBottomRight() {
  getMessageApi()?.info('底部右侧', { position: 'bottom-right' })
}

function handleCustomDuration() {
  getMessageApi()?.info('这条消息将在 10 秒后自动关闭', { duration: 10000 })
}

function handleWithCallback() {
  getMessageApi()?.success('带关闭回调的消息', {
    closable: true,
    onClose: () => {
      getMessageApi()?.info('消息已关闭（onClose 已触发）')
    },
  })
}

function handleDestroyAll() {
  getMessageApi()?.destroyAll()
}

// ——— 对话框测试 ———
function handleDialogInfo() {
  getDialogApi()?.info('这是一条提示内容').then((ok: boolean) => {
    getMessageApi()?.info(ok ? '你点击了确定' : '你点击了取消/关闭')
  })
}

function handleDialogSuccess() {
  getDialogApi()?.success('操作已完成').then((ok: boolean) => {
    if (ok) getMessageApi()?.success('已确认')
  })
}

function handleDialogError() {
  getDialogApi()?.error('发生错误，请重试').then((ok: boolean) => {
    if (ok) getMessageApi()?.info('已确认错误')
  })
}

function handleDialogWarning() {
  getDialogApi()?.warning('请注意当前操作').then((ok: boolean) => {
    getMessageApi()?.info(ok ? '已确认' : '已取消')
  })
}

function handleDialogConfirm() {
  getDialogApi()?.confirm('确定要执行该操作吗？').then((ok: boolean) => {
    getMessageApi()?.[ok ? 'success' : 'info'](ok ? '已确认' : '已取消')
  })
}

function handleDialogCreate() {
  getDialogApi()?.create({
    title: '自定义对话框',
    subtitle: '副标题示例',
    content: '支持自定义 title、subtitle、content、宽度等。',
    width: 480,
    positiveText: '知道了',
    negativeText: '关闭',
  }).then((ok: boolean) => {
    getMessageApi()?.info(ok ? '点击了「知道了」' : '点击了「关闭」或遮罩')
  })
}
</script>

<style scoped lang="scss">
.message-test-page {
  padding: var(--g-padding-xxl, 24px);
  max-width: 900px;
  margin: 0 auto;
  background: var(--g-bg-primary);
  min-height: 100%;
}

.page-header {
  margin-bottom: var(--g-space-xl);
  padding-bottom: var(--g-space-lg);
  border-bottom: 1px solid var(--g-border-primary);

  h1 {
    font-size: 24px;
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 var(--g-space-sm);
  }

  .page-description {
    font-size: var(--g-font-size-sm);
    color: var(--g-text-secondary);
    margin: 0;
    line-height: 1.6;
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
  border-radius: 12px;

  h2 {
    font-size: var(--g-font-size-base);
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 var(--g-space-xs);
  }

  .section-description {
    font-size: var(--g-font-size-sm);
    color: var(--g-text-secondary);
    margin: 0 0 var(--g-space-md);
    line-height: 1.5;

    code {
      padding: 2px 6px;
      border-radius: 4px;
      background: var(--g-bg-tertiary);
      font-family: ui-monospace, monospace;
    }
  }

  .button-group {
    display: flex;
    flex-wrap: wrap;
    gap: var(--g-space-sm);
  }
}
</style>
