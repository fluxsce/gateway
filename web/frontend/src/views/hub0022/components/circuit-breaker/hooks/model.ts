/**
 * 服务熔断配置表单字段
 */

import type { RsDataFormField } from '@/components/form/rs-data'

export function useCircuitBreakerConfigModel() {
  const formFields: RsDataFormField[] = [
    {
      field: 'errorRatePercent',
      label: '错误率阈值(%)',
      type: 'number',
      span: 12,
      defaultValue: 50,
      tips: '这台机器连不上、超时或返回 5xx 的比例达到该值，就暂时不再把请求打过去。客户参数错误（4xx）不算。建议 50。',
      props: { min: 1, max: 100 },
    },
    {
      field: 'minimumRequests',
      label: '最小请求数',
      type: 'number',
      span: 12,
      defaultValue: 10,
      tips: '统计窗口内至少要有这么多次访问才判断。次数不够即使全失败也不踢，避免刚上线或流量很小误伤。建议 10。',
      props: { min: 1, max: 10000 },
    },
    {
      field: 'windowSizeSeconds',
      label: '统计窗口(秒)',
      type: 'number',
      span: 12,
      defaultValue: 60,
      tips: '只看最近这段时间的访问。更早的失败会过期，不会一直记仇。建议 60 秒。',
      props: { min: 1, max: 3600 },
    },
    {
      field: 'openTimeoutSeconds',
      label: '停用等待(秒)',
      type: 'number',
      span: 12,
      defaultValue: 60,
      tips: '机器被摘掉后，先休息这么久再试探几笔。建议 60 秒。',
      props: { min: 1, max: 3600 },
    },
    {
      field: 'halfOpenMaxRequests',
      label: '恢复试探次数',
      type: 'number',
      span: 12,
      defaultValue: 3,
      tips: '休息结束后先放这么多笔试探。全部成功才恢复使用；有一笔失败就再停用一轮。建议 3。',
      props: { min: 1, max: 100 },
    },
    {
      field: 'slowCallThreshold',
      label: '慢调用阈值(毫秒)',
      type: 'number',
      span: 12,
      defaultValue: 60000,
      tips: '这一跳转发（从网关发出到读完上游响应）超过该时间算慢。请按接口正常耗时填写，正常要 60 秒就填 60000，不要填 1000，否则正常请求也会被当成慢调用。',
      props: { min: 1, max: 600000 },
    },
    {
      field: 'slowCallRatePercent',
      label: '慢调用率(%)',
      type: 'number',
      span: 12,
      defaultValue: 50,
      tips: '统计窗口内慢请求占比达到该值，即使没有报错也会摘掉这台机器。建议 50。',
      props: { min: 1, max: 100 },
    },
  ]

  return { formFields }
}
