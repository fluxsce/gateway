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
      field: 'basicFieldset',
      label: '基本设置',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'rootDirectory',
          label: '网站文件目录',
          type: 'input',
          span: 24,
          required: true,
          placeholder: '例如 D:/www/shop/dist 或 /var/www/shop/dist',
          tips: '填打包后的文件夹，也就是里面有 index.html 的那个目录。网关只从这里读文件，不会读到外面。',
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
        {
          field: 'stripRoutePrefix',
          label: '按路由路径找文件',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '建议打开。路由是 /app 时，访问 /app/index.html 会去目录里找 index.html，而不是找 app/index.html。浏览器地址栏不会变。',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'spaFallback',
          label: '单页应用刷新不报错',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: 'Vue / React 这类前端自己管页面跳转时请打开。刷新 /user/12 这类没有文件后缀的地址时，会打开首页交给前端处理。找不到的 .js / .css 仍会提示文件不存在。',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'indexFiles',
          label: '打开目录时的默认文件',
          type: 'input',
          span: 24,
          defaultValue: 'index.html',
          placeholder: 'index.html',
          tips: '访问 / 或某个文件夹时，依次尝试这些文件名，多个用逗号分开。一般填 index.html 即可。',
        },
      ],
    },
    {
      field: 'optionalFieldset',
      label: '可选：旧地址与错误页',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'rewriteRules',
          label: '旧地址对应到新文件',
          type: 'custom',
          span: 24,
          tips: '文件改名或换目录后，把旧访问路径指到新文件。多数站点不用填。只改找哪个文件，不改浏览器地址。',
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
        {
          field: 'errorPage404',
          label: '找不到页面时显示',
          type: 'input',
          span: 12,
          placeholder: '例如 /404.html，可留空',
          tips: '填网站目录里的页面，如 /404.html。留空则返回简短错误说明。',
        },
        {
          field: 'errorPage403',
          label: '无权访问时显示',
          type: 'input',
          span: 12,
          placeholder: '例如 /403.html，可留空',
          tips: '填网站目录里的页面，如 /403.html。留空则返回简短错误说明。',
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
