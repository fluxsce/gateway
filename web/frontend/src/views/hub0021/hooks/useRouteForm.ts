/**
 * 路由配置对话框表单状态：字段、校验、匹配类型说明与提交快照。
 */
import type { RsFormRules, RsFormValidationResult, RsSelectModelValue } from '@/ui'
import { computed, reactive, ref } from 'vue'
import type { RouteConfig } from '../types'
import { MatchType } from '../types'

/** RsForm 暴露的校验方法 */
interface RsFormExpose {
  validate: () => Promise<RsFormValidationResult>
  clearValidation: (names?: string | string[]) => void
  resetFields: (names?: string | string[]) => void
}

/** 对话框表单字段 */
export interface RouteFormData {
  routeName: string
  matchType: MatchType
  routePath: string
  allowedMethods: string[]
  allowedHosts: string
  routePriority: number
  serviceDefinitionId: string
  logConfigId: string
  activeFlag: 'Y' | 'N'
  routeMetadata: Record<string, unknown>
  noteText: string
}

const HTTP_METHODS = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'] as const

function createDefaultForm(): RouteFormData {
  return {
    routeName: '',
    matchType: MatchType.PREFIX,
    routePath: '',
    allowedMethods: ['GET'],
    allowedHosts: '',
    routePriority: 100,
    serviceDefinitionId: '',
    logConfigId: '',
    activeFlag: 'Y',
    routeMetadata: {},
    noteText: '',
  }
}

function parseAllowedMethods(value: RouteConfig['allowedMethods']): string[] {
  if (Array.isArray(value)) return value.map(String)
  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) return parsed.map(String)
    } catch {
      return value.split(',').map((item) => item.trim()).filter(Boolean)
    }
  }
  return ['GET']
}

/**
 * 路由新建/编辑表单逻辑，供 useRouteConfigDialog 组合。
 */
export function useRouteForm() {
  const formRef = ref<RsFormExpose | null>(null)
  const formData = reactive<RouteFormData>(createDefaultForm())
  const editingRouteId = ref('')

  const isEditMode = computed(() => Boolean(editingRouteId.value))

  const httpMethodOptions = HTTP_METHODS.map((method) => ({
    label: method,
    value: method,
  }))

  const matchTypeOptions = [
    { label: '精确匹配', value: MatchType.EXACT },
    { label: '前缀匹配', value: MatchType.PREFIX },
    { label: '正则匹配', value: MatchType.REGEX },
  ]

  const getMatchTypeDescription = computed(() => {
    switch (formData.matchType) {
      case MatchType.EXACT:
        return '精确匹配：请求路径必须完全匹配配置的路径'
      case MatchType.PREFIX:
        return '前缀匹配：请求路径以配置的路径为前缀即可匹配'
      case MatchType.REGEX:
        return '正则匹配：使用正则表达式匹配请求路径'
      default:
        return '请选择匹配类型'
    }
  })

  const getPathExample = computed(() => {
    switch (formData.matchType) {
      case MatchType.EXACT:
        return '示例: /api/users/123'
      case MatchType.PREFIX:
        return '示例: /api/users（匹配 /api/users/*）'
      case MatchType.REGEX:
        return '示例: ^/api/users/\\d+$'
      default:
        return '精确匹配示例：/api/users；前缀匹配示例：/api/；正则匹配示例：^/api/(users|orders)/?$'
    }
  })

  const formRules: RsFormRules = {
    routeName: [{ required: true, message: '请输入路由名称', trigger: ['blur', 'change'] }],
    routePath: [
      { required: true, message: '请输入路由路径', trigger: ['blur', 'change'] },
      {
        validator: (value: unknown) => {
          if (typeof value !== 'string' || !value) return true
          if (formData.matchType === MatchType.REGEX || value.startsWith('^')) {
            try {
              new RegExp(value)
            } catch {
              return '请输入有效的正则表达式'
            }
            return true
          }
          if (!value.startsWith('/')) return '路由路径必须以 / 开头'
          return true
        },
        trigger: ['blur', 'change'],
      },
    ],
    matchType: [{ required: true, message: '请选择匹配类型', trigger: 'change' }],
  }

  const resetForm = () => {
    Object.assign(formData, createDefaultForm())
    editingRouteId.value = ''
    formRef.value?.clearValidation()
  }

  const fillFormData = (route: RouteConfig) => {
    editingRouteId.value = route.routeConfigId
    formData.routeName = route.routeName || ''
    formData.matchType = route.matchType ?? MatchType.PREFIX
    formData.routePath = route.routePath || ''
    formData.allowedMethods = parseAllowedMethods(route.allowedMethods)
    formData.allowedHosts = route.allowedHosts || ''
    formData.routePriority = route.routePriority ?? 100
    formData.serviceDefinitionId = route.serviceDefinitionId || ''
    formData.logConfigId = route.logConfigId || ''
    formData.activeFlag = route.activeFlag === 'N' ? 'N' : 'Y'
    formData.routeMetadata = { ...(route.routeMetadata || {}) }
    formData.noteText = route.noteText || ''
  }

  const validateForm = async (): Promise<boolean> => {
    const result = await formRef.value?.validate()
    return result?.valid ?? false
  }

  const getFormData = (): RouteFormData => ({
    routeName: formData.routeName,
    matchType: formData.matchType,
    routePath: formData.routePath,
    allowedMethods: [...formData.allowedMethods],
    allowedHosts: formData.allowedHosts,
    routePriority: formData.routePriority,
    serviceDefinitionId: formData.serviceDefinitionId,
    logConfigId: formData.logConfigId,
    activeFlag: formData.activeFlag,
    routeMetadata: { ...formData.routeMetadata },
    noteText: formData.noteText,
  })

  /**
   * 同步匹配类型；非法值忽略。
   */
  const handleMatchTypeChange = (value: RsSelectModelValue) => {
    const raw = Array.isArray(value) ? value[0] : value
    const token =
      typeof raw === 'object' && raw && 'value' in raw ? raw.value : raw
    const matchType = Number(token)
    if (
      matchType === MatchType.EXACT ||
      matchType === MatchType.PREFIX ||
      matchType === MatchType.REGEX
    ) {
      formData.matchType = matchType
    }
  }

  /**
   * 路径输入：精确/前缀补前导斜杠；正则不自动加前缀。
   */
  const handlePathInput = (value: string) => {
    if (formData.matchType === MatchType.REGEX) {
      formData.routePath = value
      return
    }
    if (value && !value.startsWith('/')) {
      formData.routePath = `/${value}`
    } else {
      formData.routePath = value
    }
  }

  return {
    formRef,
    formData,
    formRules,
    isEditMode,
    editingRouteId,
    httpMethodOptions,
    matchTypeOptions,
    getPathExample,
    getMatchTypeDescription,
    resetForm,
    fillFormData,
    validateForm,
    getFormData,
    handleMatchTypeChange,
    handlePathInput,
  }
}
