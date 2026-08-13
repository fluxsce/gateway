<template>
  <div class="rs-data-form-test-page">
    <div class="page-header">
      <h1>RsDataForm / Modal 数据表单测试</h1>
      <p class="page-description">
        niuma-ui 实现：页签、fieldset、主键禁用、mode、校验；弹窗基于 RsDialog + RsDataForm
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>扁平表单（create）</h2>
          <RsDataForm
            ref="formRef"
            mode="create"
            label-placement="top"
            :form-fields="formFields"
            :form-tabs="formTabs"
            :show-footer="true"
            :show-submit="true"
            submit-text="提交校验"
            @submit="onFormSubmit"
          />
        <pre class="result">{{ lastSubmitJson }}</pre>
      </section>

      <section class="test-section">
        <h2>弹窗模式</h2>
        <div class="row">
          <RsButton size="sm" variant="primary" @click="openModal('create')">新增</RsButton>
          <RsButton size="sm" @click="openModal('edit')">编辑</RsButton>
          <RsButton size="sm" @click="openModal('view')">查看</RsButton>
        </div>
        <RsDataFormModal
          v-model:visible="modalVisible"
          :mode="modalMode"
          :title="modalTitle"
          :form-fields="formFields"
          :form-tabs="formTabs"
          :initial-data="modalInitial || undefined"
          :auto-close-on-confirm="false"
          :confirm-loading="confirmLoading"
          @submit="onModalSubmit"
          @cancel="message.info('已取消')"
        />
      </section>

      <section class="test-section">
        <h2>暴露方法</h2>
        <div class="row">
          <RsButton size="sm" @click="validateFlat">校验扁平表单</RsButton>
          <RsButton size="sm" @click="resetFlat">重置扁平表单</RsButton>
          <RsButton size="sm" @click="dumpFlat">打印 getFormData</RsButton>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  RsDataForm,
  RsDataFormModal,
  type RsDataFormExpose,
  type RsDataFormField,
  type RsDataFormTab,
  type RsDataModalMode,
} from '@/components/form/rs-data'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsButton } from '@/ui'
import { computed, ref } from 'vue'

defineOptions({ name: 'RsDataFormTest' })

const message = useAppMessage()
const formRef = ref<RsDataFormExpose | null>(null)
const lastSubmit = ref<Record<string, any> | null>(null)
const modalVisible = ref(false)
const modalMode = ref<RsDataModalMode>('create')
const modalInitial = ref<Record<string, any> | null>(null)
const confirmLoading = ref(false)

const lastSubmitJson = computed(() =>
  lastSubmit.value ? JSON.stringify(lastSubmit.value, null, 2) : '（尚未提交）',
)

const modalTitle = computed(() => {
  switch (modalMode.value) {
    case 'create':
      return '新增演示'
    case 'edit':
      return '编辑演示'
    default:
      return '查看演示'
  }
})

const formTabs: RsDataFormTab[] = [
  { key: 'basic', label: '主信息' },
  { key: 'extra', label: '扩展' },
]

const formFields: RsDataFormField[] = [
  {
    field: 'userId',
    label: '用户ID',
    type: 'input',
    span: 8,
    tabKey: 'basic',
    required: true,
    primary: true,
    tips: '编辑模式下主键自动禁用',
  },
  {
    field: 'userName',
    label: '用户名',
    type: 'input',
    span: 8,
    tabKey: 'basic',
    required: true,
  },
  {
    field: 'password',
    label: '密码',
    type: 'input',
    span: 8,
    tabKey: 'basic',
    show: (d) => d._mode === 'create',
    required: true,
    props: { type: 'password', visibilityToggle: true },
  },
  {
    field: 'gender',
    label: '性别',
    type: 'select',
    span: 8,
    tabKey: 'basic',
    options: [
      { label: '未知', value: 0 },
      { label: '男', value: 1 },
      { label: '女', value: 2 },
    ],
  },
  {
    field: 'statusFlag',
    label: '状态',
    type: 'select',
    span: 8,
    tabKey: 'basic',
    defaultValue: 1,
    options: [
      { label: '启用', value: 1 },
      { label: '禁用', value: 0 },
    ],
  },
  {
    field: 'adminFlag',
    label: '管理员',
    type: 'switch',
    span: 8,
    tabKey: 'basic',
    defaultValue: 'N',
    props: { checkedValue: 'Y', uncheckedValue: 'N' },
  },
  {
    field: 'expireAt',
    label: '过期时间',
    type: 'datetime',
    span: 12,
    tabKey: 'basic',
    required: true,
  },
  {
    field: 'profileGroup',
    label: '资料分组',
    type: 'fieldset',
    tabKey: 'extra',
    props: {
      borderStyle: 'dashed',
      titleSize: 'normal',
    },
    children: [
      {
        field: 'email',
        label: '邮箱',
        type: 'input',
        span: 12,
        placeholder: 'name@example.com',
      },
      {
        field: 'remark',
        label: '备注',
        type: 'textarea',
        span: 12,
        props: { rows: 3 },
      },
      {
        field: 'age',
        label: '年龄',
        type: 'number',
        span: 8,
      },
    ],
  },
]

const onFormSubmit = (data?: Record<string, any>) => {
  lastSubmit.value = data || null
  message.success('扁平表单提交成功')
}

const openModal = (mode: RsDataModalMode) => {
  modalMode.value = mode
  if (mode === 'create') {
    modalInitial.value = null
  } else {
    modalInitial.value = {
      userId: 'U1001',
      userName: 'demo',
      gender: 1,
      statusFlag: 1,
      adminFlag: 'Y',
      expireAt: '2030-12-31 23:59:59',
      email: 'demo@example.com',
      remark: '编辑/查看初始数据',
      age: 28,
    }
  }
  modalVisible.value = true
}

const onModalSubmit = async (data?: Record<string, any>) => {
  confirmLoading.value = true
  lastSubmit.value = data || null
  await new Promise((r) => setTimeout(r, 600))
  confirmLoading.value = false
  modalVisible.value = false
  message.success(`${modalMode.value} 提交成功`)
}

const validateFlat = async () => {
  const ok = await formRef.value?.validate()
  if (ok) message.success('校验通过')
  else message.error('校验未通过')
}

const resetFlat = () => {
  formRef.value?.reset()
  message.info('已重置')
}

const dumpFlat = () => {
  lastSubmit.value = formRef.value?.getFormData() || null
  message.info('已写入 getFormData 结果')
}
</script>

<style scoped lang="scss">
.rs-data-form-test-page {
  padding: var(--g-padding-lg);
  max-width: 1100px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: var(--g-space-xl);

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
  }
}

.test-sections {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-xl);
}

.test-section {
  padding: var(--g-padding-md);
  background: var(--g-bg-secondary);
  border: 1px solid var(--g-border-primary);
  border-radius: var(--g-radius-lg);

  h2 {
    font-size: var(--g-font-size-lg);
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 var(--g-space-md);
  }
}

.row {
  display: flex;
  gap: var(--g-space-sm);
  flex-wrap: wrap;
  margin-bottom: var(--g-space-md);
}

.result {
  margin: var(--g-space-md) 0 0;
  padding: var(--g-padding-sm);
  background: var(--g-bg-tertiary, #f5f5f5);
  border-radius: var(--g-radius-md);
  font-size: 12px;
  color: var(--g-text-secondary);
  overflow: auto;
  max-height: 240px;
}
</style>
