/**
 * 预警模板管理 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import { GCodeMirror, GRichText } from '@/components'
import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { AddOutline, TrashOutline } from '@vicons/ionicons5'
import { NTable, NTag } from 'naive-ui'
import { h, ref } from 'vue'
import type { AlertTemplate, ChannelType, DisplayFormat } from '../types'
import { ACTIVE_FLAG_OPTIONS, CHANNEL_TYPE_OPTIONS, DISPLAY_FORMAT_OPTIONS } from '../types'

export function useAlertTemplateModel() {
  const moduleId = 'hub0081:alert-template'

  const loading = ref(false)
  const templateList = ref<AlertTemplate[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const getDefaultTitleTemplate = (channelType?: ChannelType | string | null) => {
    // 默认：邮件
    const ct = (channelType || 'email') as string
    if (ct === 'sms') return '告警短信 - {{title}}'
    if (ct === 'webhook') return '告警Webhook - {{title}}'
    if (ct === 'dingtalk') return '告警通知(钉钉) - {{title}}'
    if (ct === 'wechat_work') return '告警通知(企微) - {{title}}'
    if (ct === 'qq') return '告警通知(QQ) - {{title}}'
    return '告警通知 - {{title}}'
  }

  const getDefaultContentTemplate = (channelType?: ChannelType | string | null) => {
    const ct = (channelType || 'email') as string
    // 默认：邮件 HTML（更美观的邮件模板）
    if (ct === 'email') {
      return [
        '<!DOCTYPE html>',
        '<html>',
        '<head>',
        '  <meta charset="UTF-8">',
        '  <meta name="viewport" content="width=device-width, initial-scale=1.0">',
        '</head>',
        '<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, Helvetica, Arial, PingFang SC, Microsoft YaHei, sans-serif; background-color: #f5f5f5;">',
        '  <div style="max-width: 600px; margin: 20px auto; background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); overflow: hidden;">',
        '    <!-- 头部 -->',
        '    <div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 24px; color: #ffffff;">',
        '      <h1 style="margin: 0; font-size: 24px; font-weight: 600;">🚨 告警通知</h1>',
        '      <div style="margin-top: 8px; font-size: 14px; opacity: 0.9;">{{title}}</div>',
        '    </div>',
        '    <!-- 时间信息 -->',
        '    <div style="padding: 16px 24px; background-color: #f8f9fa; border-bottom: 1px solid #e9ecef;">',
        '      <div style="display: flex; align-items: center; color: #6c757d; font-size: 13px;">',
        '        <span style="margin-right: 8px;">⏰</span>',
        '        <span>{{timestamp}}（{{timestamp_iso}}）</span>',
        '      </div>',
        '    </div>',
        '    <!-- 内容区域 -->',
        '    <div style="padding: 24px;">',
        '      <div style="margin-bottom: 20px;">',
        '        <h3 style="margin: 0 0 12px; font-size: 16px; color: #212529; font-weight: 600;">📋 告警内容</h3>',
        '        <div style="padding: 16px; background-color: #f8f9fa; border-left: 4px solid #667eea; border-radius: 4px; white-space: pre-wrap; line-height: 1.6; color: #495057;">{{content}}</div>',
        '      </div>',
        '      <!-- 标签信息 -->',
        '      <div style="margin-bottom: 20px;">',
        '        <h3 style="margin: 0 0 12px; font-size: 16px; color: #212529; font-weight: 600;">🏷️ 标签信息</h3>',
        '        <div style="display: flex; flex-wrap: wrap; gap: 8px;">',
        '          <span style="display: inline-block; padding: 4px 12px; background-color: #e7f3ff; color: #0066cc; border-radius: 12px; font-size: 12px;">标签：{{tags}}</span>',
        '          <span style="display: inline-block; padding: 4px 12px; background-color: #fff3cd; color: #856404; border-radius: 12px; font-size: 12px;">严重级别：{{tag.severity}}</span>',
        '          <span style="display: inline-block; padding: 4px 12px; background-color: #d1ecf1; color: #0c5460; border-radius: 12px; font-size: 12px;">来源：{{tag.source}}</span>',
        '        </div>',
        '      </div>',
        '      <!-- 表格字段 -->',
        '      <div>',
        '        <h3 style="margin: 0 0 12px; font-size: 16px; color: #212529; font-weight: 600;">📊 详细信息</h3>',
        '        <table cellspacing="0" cellpadding="12" border="0" style="width: 100%; border-collapse: collapse; border: 1px solid #e9ecef; border-radius: 4px; overflow: hidden;">',
        '          <tr style="background-color: #f8f9fa;">',
        '            <td style="font-weight: 600; color: #495057; border-bottom: 1px solid #e9ecef;">service_name</td>',
        '            <td style="color: #212529; border-bottom: 1px solid #e9ecef;">{{table.service_name}}</td>',
        '          </tr>',
        '          <tr>',
        '            <td style="font-weight: 600; color: #495057; border-bottom: 1px solid #e9ecef;">instance_id</td>',
        '            <td style="color: #212529; border-bottom: 1px solid #e9ecef;">{{table.instance_id}}</td>',
        '          </tr>',
        '        </table>',
        '      </div>',
        '    </div>',
        '    <!-- 底部 -->',
        '    <div style="padding: 16px 24px; background-color: #f8f9fa; border-top: 1px solid #e9ecef; text-align: center; color: #6c757d; font-size: 12px;">',
        '      此邮件由系统自动发送，请勿回复',
        '    </div>',
        '  </div>',
        '</body>',
        '</html>',
      ].join('\n')
    }
    // 其它渠道：文本/Markdown（优化格式）
    return [
      '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━',
      '🚨 告警通知',
      '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━',
      '',
      '📌 告警标题',
      '   {{title}}',
      '',
      '⏰ 告警时间',
      '   {{timestamp}}',
      '   ISO格式：{{timestamp_iso}}',
      '',
      '📋 告警内容',
      '   {{content}}',
      '',
      '🏷️ 标签信息',
      '   标签汇总：{{tags}}',
      '   严重级别：{{tag.severity}}',
      '   来源：{{tag.source}}',
      '',
      '📊 详细信息',
      '   服务名称：{{table.service_name}}',
      '   实例ID：{{table.instance_id}}',
      '',
      '💡 自定义字段',
      '   {{extra.custom_field}}',
      '',
      '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━',
    ].join('\n')
  }

  // 搜索表单
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'templateName',
        label: '模板名称',
        type: 'input',
        placeholder: '请输入模板名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'channelType',
        label: '渠道类型',
        type: 'select',
        placeholder: '请选择渠道类型',
        span: 6,
        clearable: true,
        options: [{ label: '全部', value: '' }, ...CHANNEL_TYPE_OPTIONS.map(o => ({ label: o.label, value: o.value }))],
      },
      {
        field: 'displayFormat',
        label: '显示格式',
        type: 'select',
        placeholder: '请选择显示格式',
        span: 6,
        clearable: true,
        options: [{ label: '全部', value: '' }, ...DISPLAY_FORMAT_OPTIONS.map(o => ({ label: o.label, value: o.value }))],
      },
      {
        field: 'activeFlag',
        label: '启用状态',
        type: 'select',
        placeholder: '请选择启用状态',
        span: 6,
        clearable: true,
        options: [{ label: '全部', value: '' }, ...ACTIVE_FLAG_OPTIONS.map(o => ({ label: o.label, value: o.value }))],
      },
    ],
    toolbarButtons: [
      { key: 'add', label: '新增模板', icon: AddOutline, type: 'primary', tooltip: '新增预警模板' },
      { key: 'delete', label: '删除', icon: TrashOutline, type: 'error', tooltip: '批量删除选中的模板' },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  const getChannelTypeLabel = (channelType?: ChannelType | string | null) => {
    if (!channelType) return ''
    const option = CHANNEL_TYPE_OPTIONS.find(opt => opt.value === channelType)
    return option?.label || String(channelType)
  }

  const getDisplayFormatLabel = (format?: DisplayFormat | string | null) => {
    if (!format) return ''
    const option = DISPLAY_FORMAT_OPTIONS.find(opt => opt.value === format)
    return option?.label || String(format)
  }

  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      { field: 'templateName', title: '模板名称', align: 'center', showOverflow: 'tooltip', width: 180 },
      { field: 'channelType', title: '渠道类型', align: 'center', width: 120, slots: { default: 'channelType' } },
      { field: 'displayFormat', title: '显示格式', align: 'center', width: 120, slots: { default: 'displayFormat' } },
      { field: 'activeFlag', title: '启用状态', align: 'center', width: 100, slots: { default: 'activeFlag' } },
      { field: 'templateDesc', title: '模板描述', align: 'center', showOverflow: 'tooltip', width: 240 },
      { field: 'addTime', title: '创建时间', align: 'center', width: 160, formatter: ({ row }) => formatDate(row.addTime) },
      { field: 'addWho', title: '创建人', align: 'center', width: 120, showOverflow: 'tooltip' },
      { field: 'editTime', title: '修改时间', align: 'center', width: 160, formatter: ({ row }) => formatDate(row.editTime) },
      { field: 'editWho', title: '修改人', align: 'center', width: 120, showOverflow: 'tooltip' },
    ],
    showCheckbox: true,
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      showCopyRow: true,
      options: [
        { code: 'view', name: '查看详情', prefixIcon: 'vxe-icon-eye-fill' },
        { code: 'edit', name: '编辑', prefixIcon: 'vxe-icon-edit' },
        { code: 'delete', name: '删除', prefixIcon: 'vxe-icon-delete' },
      ],
    },
  }

  // 表单页签
  const formTabs = [
    { key: 'basic', label: '基本信息' },
    { key: 'template', label: '模板内容' },
    { key: 'other', label: '其他信息' },
  ]

  // 表单字段（CodeMirror 由页面 slots 渲染）
  const formFields: DataFormField[] = [
    {
      field: 'templateName',
      label: '模板名称',
      type: 'input' as const,
      span: 12,
      primary: true,
      tabKey: 'basic',
      required: true,
      rules: [
        { required: true, message: '请输入模板名称', trigger: ['blur', 'input'] },
        { max: 64, message: '模板名称不能超过64个字符', trigger: ['blur', 'input'] },
        { pattern: /^[a-zA-Z0-9_]+$/, message: '模板名称只能包含英文字母、数字和下划线', trigger: ['blur', 'input'] },
      ],
    },
    { field: 'tenantId', label: '租户ID', type: 'input' as const, span: 12, show: false },
    {
      field: 'channelType',
      label: '渠道类型',
      type: 'select' as const,
      span: 12,
      tabKey: 'basic',
      placeholder: '请选择渠道类型（可选，空为通用模板）',
      clearable: true,
      defaultValue: 'email',
      options: [{ label: '通用(不限制)', value: '' }, ...CHANNEL_TYPE_OPTIONS.map(o => ({ label: o.label, value: o.value }))],
      props: {
        onUpdateValue: (_value: any, formData: Record<string, any>) => {
          // 仅在“未填写模板内容”时自动填充，避免覆盖用户已编辑内容
          const titleEmpty = formData.titleTemplate === null || formData.titleTemplate === undefined || formData.titleTemplate === ''
          const contentEmpty = formData.contentTemplate === null || formData.contentTemplate === undefined || formData.contentTemplate === ''
          if (titleEmpty) formData.titleTemplate = getDefaultTitleTemplate(formData.channelType)
          if (contentEmpty) formData.contentTemplate = getDefaultContentTemplate(formData.channelType)
        },
      },
    },
    {
      field: 'displayFormat',
      label: '显示格式',
      type: 'select' as const,
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 'table',
      options: DISPLAY_FORMAT_OPTIONS.map(o => ({ label: o.label, value: o.value })),
      rules: [{ required: true, message: '请选择显示格式', trigger: ['change', 'blur'] }],
    },
    {
      field: 'activeFlag',
      label: '启用状态',
      type: 'switch' as const,
      span: 12,
      tabKey: 'basic',
      defaultValue: 'Y',
      props: { checkedValue: 'Y', uncheckedValue: 'N' },
    },
    {
      field: 'templateDesc',
      label: '模板描述',
      type: 'input' as const,
      span: 24,
      tabKey: 'basic',
      props: { type: 'textarea', rows: 2, maxlength: 500, showCount: true },
    },
    // 模板内容：使用 custom 渲染 CodeMirror（GDataFormModal 支持 custom）
    {
      field: 'titleTemplate',
      label: '标题模板',
      type: 'custom' as const,
      span: 24,
      tabKey: 'template',
      defaultValue: getDefaultTitleTemplate('email'),
      tips: '支持变量占位符，例如：{{title}} / {{tag.severity}}',
      render: (formData) =>
        h(GCodeMirror as any, {
          modelValue: formData.titleTemplate || '',
          'onUpdate:modelValue': (v: string) => (formData.titleTemplate = v),
          language: 'plaintext',
          height: 120,
          lineNumbers: true,
          lineWrapping: true,
          placeholder: '例如：告警通知 - {{title}}',
        }),
    },
    {
      field: 'contentTemplate',
      label: '内容模板',
      type: 'custom' as const,
      span: 24,
      tabKey: 'template',
      tips: '支持变量占位符，建议使用多行编辑',
      defaultValue: getDefaultContentTemplate('email'),
      render: (formData) => {
        // 邮件：优先使用富文本编辑器（所见即所得），输出 HTML
        if (formData.channelType === 'email') {
          return h(GRichText as any, {
            modelValue: formData.contentTemplate || '',
            'onUpdate:modelValue': (v: string) => (formData.contentTemplate = v),
            placeholder: '请输入邮件内容（支持占位符，如 {{title}} / {{content}}）',
            minHeight: 260,
            showToolbar: true,
            readonly: false,
          })
        }
        // 其它渠道：继续使用 CodeMirror（markdown/text）
        return h(GCodeMirror as any, {
          modelValue: formData.contentTemplate || '',
          'onUpdate:modelValue': (v: string) => (formData.contentTemplate = v),
          language: 'markdown',
          height: 260,
          lineNumbers: true,
          lineWrapping: true,
          placeholder: '请输入内容模板（支持变量占位符）',
        })
      },
    },
    {
      field: '__templateHelp',
      label: '占位符说明',
      type: 'custom' as const,
      span: 24,
      tabKey: 'template',
      tips: '与后端占位符替换规则一致：使用 {{...}} 引用字段；找不到字段时将保留原样',
      render: () => {
        const rows: Array<{ key: string; desc: string; example?: string }> = [
          { key: 'title', desc: '消息标题', example: '{{title}}' },
          { key: 'content', desc: '消息内容', example: '{{content}}' },
          { key: 'timestamp', desc: '时间戳（2006-01-02 15:04:05）', example: '{{timestamp}}' },
          { key: 'timestamp_iso', desc: '时间戳（RFC3339/ISO）', example: '{{timestamp_iso}}' },
          { key: 'timestamp_unix', desc: 'Unix 秒时间戳', example: '{{timestamp_unix}}' },
          { key: 'timestamp_unix_ms', desc: 'Unix 毫秒时间戳', example: '{{timestamp_unix_ms}}' },
          { key: 'tags', desc: '所有标签（key: value | ...）', example: '{{tags}}' },
          { key: 'tag.<key>', desc: '指定标签值（如 severity）', example: '{{tag.severity}}' },
          { key: 'extra.<key>', desc: '额外字段（来自 Extra，跳过 send_config）', example: '{{extra.custom_field}}' },
          { key: 'table', desc: '所有表格字段（key: value | ...）', example: '{{table}}' },
          { key: 'table.<key>', desc: '指定表格字段（如 service_name）', example: '{{table.service_name}}' },
        ]
        return h('div', { style: 'width: 100%;' }, [
          h(
            'div',
            { style: 'display:flex; gap:8px; align-items:center; margin-bottom:8px;' },
            [
              h(NTag, { size: 'small', type: 'info' }, { default: () => '占位符格式：{{field}} / {{tag.key}} / {{extra.key}} / {{table.key}}' }),
              h(NTag, { size: 'small', type: 'warning' }, { default: () => '未匹配到字段：原样保留' }),
            ]
          ),
          h(
            NTable,
            { bordered: true, striped: true, size: 'small' } as any,
            {
              default: () => [
                h('thead', [
                  h('tr', [
                    h('th', { style: 'width: 240px;' }, '字段'),
                    h('th', '说明'),
                    h('th', { style: 'width: 220px;' }, '示例'),
                  ]),
                ]),
                h(
                  'tbody',
                  rows.map((r) =>
                    h('tr', [
                      h('td', { style: 'font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;' }, r.key),
                      h('td', r.desc),
                      h('td', { style: 'font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;' }, r.example || ''),
                    ])
                  )
                ),
              ],
            }
          ),
        ])
      },
    },
    {
      field: 'attachmentConfig',
      label: '附件配置(JSON)',
      type: 'custom' as const,
      span: 24,
      tabKey: 'template',
      tips: '邮件附件等配置（可选，JSON）',
      render: (formData) =>
        h(GCodeMirror as any, {
          modelValue: formData.attachmentConfig || '',
          'onUpdate:modelValue': (v: string) => (formData.attachmentConfig = v),
          language: 'json',
          height: 160,
          lineNumbers: true,
          lineWrapping: true,
          placeholder: '可选：附件配置 JSON',
        }),
    },
    // 其他信息
    { field: 'noteText', label: '备注信息', type: 'input' as const, span: 24, tabKey: 'other', props: { type: 'textarea', rows: 3, maxlength: 500, showCount: true } },
    { field: 'addTime', label: '创建时间', type: 'datetime' as const, span: 12, tabKey: 'other', disabled: true },
    { field: 'addWho', label: '创建人', type: 'input' as const, span: 12, tabKey: 'other', disabled: true },
    { field: 'editTime', label: '修改时间', type: 'datetime' as const, span: 12, tabKey: 'other', disabled: true },
    { field: 'editWho', label: '修改人', type: 'input' as const, span: 12, tabKey: 'other', disabled: true },
  ]

  function setTemplateList(list: AlertTemplate[]) {
    templateList.value = list
  }
  function setLoading(value: boolean) {
    loading.value = value
  }
  function resetPagination() {
    pageInfo.value = undefined
  }
  function updatePagination(newPageInfo: Partial<PageInfoObj>) {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  return {
    moduleId,
    loading,
    templateList,
    pageInfo,

    searchFormConfig,
    gridConfig,
    formTabs,
    formFields,

    getChannelTypeLabel,
    getDisplayFormatLabel,

    setTemplateList,
    setLoading,
    resetPagination,
    updatePagination,
  }
}

export type AlertTemplateModel = ReturnType<typeof useAlertTemplateModel>


