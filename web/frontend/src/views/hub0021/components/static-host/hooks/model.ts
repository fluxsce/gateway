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
      label: '1. 查找目录',
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
          tips: '先在这个目录里找文件。多个站点目录只差一段名字时，可写 .../app-{v1,v2}/dist。请求不得读出该目录。',
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
          field: 'fallbackRoots',
          label: '备用目录',
          type: 'textarea',
          span: 24,
          placeholder: '本目录找不到时再找，一行一个，最多 3 个\n例如 /var/www/wms/dist',
          tips: '主目录没有这个文件时，用同一相对路径按顺序到这些目录再找。适合以前 nginx try_files 写两个根的情况。这里不能写 {v1,v2}，也不能用路由变量拼路径。',
          props: { rows: 3 },
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
          tips: '将请求 URI 映射为根目录内的相对路径。开启后，路由 /app 的 /app/index.html 对应根目录下 index.html；正则路由剥字面前缀。路由上的「转发剥前缀」打开时，这里即使关闭也会剥前缀，避免两套开关不一致。关闭且路由也未剥前缀时，保留完整 URI 再拼接。不修改浏览器地址。',
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
        {
          field: 'redirectDirectorySlash',
          label: '目录补尾斜杠',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '默认关闭：/docs 直接出索引。打开后，目录请求会 301 到 /docs/，适合页面里用相对路径引用资源的老站点。.js/.css 即使落到目录也不会补斜杠。',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'rootTokenExact',
          label: '占位符精确匹配',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '根目录写成 app-{d10,d12}/dist 时，默认 d10app 会命中 d10。打开后必须整段相等（只要 /d10 不要 /d10app）。datahub 子应用一般保持关闭。',
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
      field: 'cacheFieldset',
      label: '4. 浏览器缓存',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'cacheControlMaxAge',
          label: '普通文件缓存（秒）',
          type: 'number',
          span: 12,
          defaultValue: 3600,
          tips: '图片、脚本等普通文件告诉浏览器记住多久。首页和 HTML 不会用这个值，避免发版后还看到旧页面。填 0 表示普通文件也不缓存。',
          props: { min: 0, max: 31536000 },
        },
        {
          field: 'cacheControlByExt',
          label: '按文件类型覆盖',
          type: 'textarea',
          span: 12,
          placeholder: '一行一个，例如\n.js=86400\n.woff2=31536000',
          tips: '只改这类后缀的缓存时间。没写的后缀仍用上面的秒数。带内容哈希的 js/css 默认缓存一年；这里写了 .js 则以这里为准。HTML 始终不缓存。',
          props: { rows: 3 },
        },
      ],
    },
    {
      field: 'compressFieldset',
      label: '5. 压缩',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'enablePrecompress',
          label: '使用已压好的文件',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '目录里已有同名 .br 或 .gz 时直接发给浏览器，最快。没有这些文件时不受影响。建议保持打开。',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'enableGzip',
          label: '没有预压缩时再压',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          tips: '只压 HTML/JS/CSS/JSON 等文本。已有 .gz/.br 时仍用预压缩。带 Range 的下载、小于 1KB 或大于 1MB 的文件不压，避免和分段下载冲突。默认关。',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
      ],
    },
    {
      field: 'headerFieldset',
      label: '6. 页面安全头',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'securityHeaders',
          label: '额外响应头',
          type: 'textarea',
          span: 24,
          placeholder: '一行一个，例如\nX-Frame-Options: SAMEORIGIN\nReferrer-Policy: no-referrer',
          tips: '只能写页面安全相关的头：X-Frame-Options、Content-Security-Policy、Referrer-Policy、Permissions-Policy、Strict-Transport-Security 等。不能改 Location、Cookie、缓存。网页和文本会自动带 charset=utf-8，不用在这里写。跨域请用路由上的 CORS。',
          props: { rows: 4 },
        },
      ],
    },
    {
      field: 'errorFieldset',
      label: '7. 错误页面',
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
      field: 'limitFieldset',
      label: '8. 访问限制',
      type: 'fieldset',
      span: 24,
      children: [
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
