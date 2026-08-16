/**
 * 路由本机目录托管表单字段
 */

import type { RsDataFormField } from '@/components/form/rs-data'
import { h } from 'vue'
import RewriteRulesEditor from '../RewriteRulesEditor.vue'

export type { StaticHostConfig } from './types'

export function useStaticHostConfigModel() {
  const formFields: RsDataFormField[] = [
    {
      field: 'rootFieldset',
      label: '1. 站点根目录',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'rootDirectory',
          label: '网站文件目录',
          type: 'input',
          span: 24,
          required: true,
          placeholder: '例如 /var/www/app-{v1,v2}/dist',
          tips: '文件系统访问边界。请求不得读出该目录。多个站点目录仅名称一段不同时，可写 .../app-{v1,v2}/dist：剥除路由前缀后的首段须匹配允许名单（最长优先），展开结果须落在花括号前的父目录内；名单外返回 403。后续重写不得变更已选定的根目录。',
          rules: [
            {
              required: true,
              message: '请填写网站文件目录',
              trigger: ['blur', 'change'],
              validator: (value: unknown) => {
                if (typeof value !== 'string' || !value.trim()) {
                  return '请填写网站文件目录'
                }
                return true
              },
            },
          ],
        },
      ],
    },
    {
      field: 'lookupFieldset',
      label: '2. 路径映射',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'stripRoutePrefix',
          label: '按路由路径找文件',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '将请求 URI 映射为根目录内的相对路径。开启后，路由 /app 的 /app/index.html 对应根目录下 index.html；正则路由剥字面前缀，例如 ^/datahub01webVue/(d10|d12) 的 /datahub01webVue/d10app 对应根目录下 d10app。关闭则保留完整 URI 再拼接。不修改浏览器地址。根目录含 {v1,v2} 时，以剥除前缀后的首段选择目录。',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'indexFiles',
          label: '打开目录时的默认文件',
          type: 'input',
          span: 12,
          defaultValue: 'index.html',
          placeholder: 'index.html',
          tips: '请求指向目录时依次尝试的索引文件，逗号分隔。默认 index.html。',
        },
        {
          field: 'spaFallback',
          label: '单页应用刷新不报错',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '无扩展名且未命中文件时，回退至索引文件，供前端 History 路由。带 .js / .css 的缺失请求仍返回 404。',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
      ],
    },
    {
      field: 'rewriteFieldset',
      label: '3. 重写规则',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'rewriteRules',
          label: '查找路径重写',
          type: 'custom',
          span: 24,
          tips: '根目录确定后，调整目录内的查找路径。不变更根目录，也不修改浏览器地址。仅当请求 URI 与目录内相对路径不一致时配置。',
          render: (formData, ctx) =>
            h(RewriteRulesEditor, {
              modelValue: String(ctx?.value ?? formData.rewriteRules ?? ''),
              'onUpdate:modelValue': (value: string) => {
                if (ctx?.onUpdate) {
                  ctx.onUpdate(value)
                  return
                }
                formData.rewriteRules = value
              },
            }),
        },
      ],
    },
    {
      field: 'errorFieldset',
      label: '4. 错误页面',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'errorPage404',
          label: '找不到页面时显示',
          type: 'input',
          span: 12,
          placeholder: '例如 /404.html，可留空',
          tips: '未命中文件时返回的页面，路径相对于站点根目录。HTTP 状态码保持 404。留空则返回默认响应。',
        },
        {
          field: 'errorPage403',
          label: '无权访问时显示',
          type: 'input',
          span: 12,
          placeholder: '例如 /403.html，可留空',
          tips: '路径被拒绝时返回的页面，路径相对于站点根目录。HTTP 状态码保持 403。留空则返回默认响应。',
        },
      ],
    },
    {
      field: 'advancedFieldset',
      label: '高级：缓存与安全',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'cacheControlMaxAge',
          label: '浏览器缓存（秒）',
          type: 'number',
          span: 12,
          defaultValue: 3600,
          tips: '普通图片、脚本可以在浏览器里记住多久。首页和 HTML 不会用这个值，避免发版后还看到旧页面。填 0 表示普通文件也不缓存。一般保持 3600 即可。',
          props: { min: 0, max: 31536000 },
        },
        {
          field: 'enablePrecompress',
          label: '使用已压好的文件',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '目录里如果已经有同名的 .br 或 .gz 文件，就直接发给浏览器，页面会更快。没有这些文件时不受影响。建议保持打开。',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'allowedExtensions',
          label: '只允许这些文件类型',
          type: 'input',
          span: 12,
          placeholder: '留空表示不限制，例如 .html,.js,.css,.png',
          tips: '留空即可。若只想提供网页和资源，可填写 .html,.js,.css,.png,.woff2。隐藏文件（以点开头）始终不提供。',
        },
        {
          field: 'maxFileSizeBytes',
          label: '单个文件最大字节数',
          type: 'number',
          span: 12,
          defaultValue: 0,
          tips: '超过这个大小会拒绝下载。0 表示不限制。一般不用改。',
          props: { min: 0, max: 1073741824 },
        },
        {
          field: 'followSymlinks',
          label: '跟随软链接',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '默认关闭更安全。只有目录里用了软链接（快捷方式）指向其他文件时才需要打开，且目标仍必须在网站目录内。',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
      ],
    },
  ]

  return { formFields }
}
