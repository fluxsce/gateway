<template>
  <div class="g-text-show" :class="props.class" :style="computedStyle">
    <!-- 工具栏 -->
    <div v-if="showToolbar" class="g-text-show__toolbar">
      <div class="g-text-show__toolbar-left">
        <n-tag size="small" :type="formatTagType">
          {{ formatLabel }}
        </n-tag>
        <n-tag v-if="isLargeContent" size="small" type="warning">
          超大内容（{{ contentSizeLabel }}）
        </n-tag>
        <span v-if="isLargeContent && !enableHighlight" class="performance-tip">
          已禁用语法高亮以提升性能
        </span>
      </div>
      <div class="g-text-show__toolbar-right">
        <n-button
          v-if="showCopyButton"
          size="small"
          quaternary
          @click="handleCopy"
        >
          <template #icon>
            <n-icon><CopyOutline /></n-icon>
          </template>
          复制
        </n-button>
        <n-button
          v-if="canFormat"
          size="small"
          quaternary
          @click="handleFormat"
        >
          <template #icon>
            <n-icon><CodeOutline /></n-icon>
          </template>
          {{ (isManuallyFormatted !== null ? isManuallyFormatted : props.autoFormat) ? '取消格式化' : '格式化' }}
        </n-button>
      </div>
    </div>

    <!-- 文本内容区域 -->
    <div class="g-text-show__content" :style="contentStyle">
      <n-code
        v-if="enableHighlight"
        :code="formattedContent"
        :language="codeLanguage"
        :show-line-numbers="showLineNumbers"
        :hljs="hljsInstance"
        class="g-text-show__code"
      />
      <pre
        v-else
        class="g-text-show__plain-text"
      ><code>{{ formattedContent }}</code></pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import hljs from '@/utils/highlight'
import { CodeOutline, CopyOutline } from '@vicons/ionicons5'
import { NButton, NCode, NIcon, NTag, useMessage } from 'naive-ui'
import type { Hljs } from 'naive-ui/es/_mixins'
import { computed, ref, watch } from 'vue'
import type { GTextShowEmits, GTextShowProps, TextFormat } from './types'

// 定义组件名称
defineOptions({
  name: 'GTextShow'
})

// Props
const props = withDefaults(defineProps<GTextShowProps>(), {
  content: '',
  format: 'auto',
  showLineNumbers: false,
  showCopyButton: true,
  autoFormat: true
})

// Emits
const emit = defineEmits<GTextShowEmits>()

// Message
const message = useMessage()

// 性能优化配置
const LARGE_CONTENT_THRESHOLD = 500 * 1024 // 500KB，超过此大小视为超大内容
const MAX_FORMAT_SIZE = 2 * 1024 * 1024 // 2MB，超过此大小禁用格式化

// 是否启用高亮（超大内容时可禁用）
const enableHighlight = ref(true)

// 是否手动触发了格式化（null 表示未手动操作，使用 autoFormat；true/false 表示手动设置的状态）
const isManuallyFormatted = ref<boolean | null>(null)

// hljs 实例
const hljsInstance: Hljs = {
  highlight: hljs.highlight.bind(hljs),
  getLanguage: hljs.getLanguage.bind(hljs)
}

/**
 * 内容大小（字节）
 */
const contentSize = computed(() => {
  return props.content ? new Blob([props.content]).size : 0
})

/**
 * 是否为超大内容
 */
const isLargeContent = computed(() => {
  return contentSize.value > LARGE_CONTENT_THRESHOLD
})

/**
 * 内容大小标签
 */
const contentSizeLabel = computed(() => {
  const size = contentSize.value
  if (size < 1024) {
    return `${size}B`
  } else if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(2)}KB`
  } else {
    return `${(size / 1024 / 1024).toFixed(2)}MB`
  }
})

/**
 * 自动检测文本格式（对超大内容进行优化）
 */
const detectFormat = (text: string): TextFormat => {
  if (!text || !text.trim()) {
    return 'txt'
  }

  const trimmed = text.trim()
  const size = new Blob([trimmed]).size

  // 超大内容时，简化检测逻辑，避免 JSON.parse 卡顿
  if (size > MAX_FORMAT_SIZE) {
    // 超大内容只做简单的前缀检测
    if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
      return 'json'
    }
    if (trimmed.startsWith('<?xml') || trimmed.startsWith('<')) {
      if (trimmed.includes('soap:Envelope') || trimmed.includes('soapenv:Envelope')) {
        return 'soap'
      }
      return 'xml'
    }
    return 'txt'
  }

  // 检测 JSON（仅对小内容进行完整解析）
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    try {
      JSON.parse(trimmed)
      return 'json'
    } catch {
      // 不是有效的 JSON
    }
  }

  // 检测 XML/SOAP
  if (trimmed.startsWith('<?xml') || trimmed.startsWith('<')) {
    if (trimmed.includes('soap:Envelope') || trimmed.includes('soapenv:Envelope')) {
      return 'soap'
    }
    return 'xml'
  }

  // 检测 YAML
  if (trimmed.includes('---') || (trimmed.includes(':') && trimmed.split('\n').length > 1)) {
    return 'yaml'
  }

  // 检测 SQL
  if (/^\s*(SELECT|INSERT|UPDATE|DELETE|CREATE|ALTER|DROP)\s+/i.test(trimmed)) {
    return 'sql'
  }

  // 检测 HTML
  if (/<html[\s>]|<body[\s>]|<div[\s>]/i.test(trimmed)) {
    return 'html'
  }

  // 检测 JavaScript/TypeScript
  if (trimmed.includes('function') || trimmed.includes('const ') || trimmed.includes('let ')) {
    if (trimmed.includes(':') && trimmed.includes('interface') || trimmed.includes('type ')) {
      return 'typescript'
    }
    return 'javascript'
  }

  // 检测 CSS
  if (trimmed.includes('{') && trimmed.includes('}') && trimmed.includes(':')) {
    return 'css'
  }

  return 'txt'
}

/**
 * 检测到的格式
 */
const detectedFormat = computed<TextFormat>(() => {
  if (props.format === 'auto') {
    return detectFormat(props.content)
  }
  return props.format
})

/**
 * 代码高亮语言
 */
const codeLanguage = computed(() => {
  const format = detectedFormat.value
  // 将格式映射到 highlight.js 支持的语言
  const languageMap: Record<TextFormat, string> = {
    json: 'json',
    xml: 'xml',
    soap: 'xml', // SOAP 使用 XML 高亮
    txt: 'plaintext',
    yaml: 'yaml',
    sql: 'sql',
    javascript: 'javascript',
    typescript: 'typescript',
    css: 'css',
    html: 'html',
    auto: 'plaintext'
  }
  return languageMap[format] || 'plaintext'
})

/**
 * 格式化 JSON（对超大内容进行优化）
 */
const formatJson = (jsonString: string): string => {
  const size = new Blob([jsonString]).size
  
  // 超大内容禁用格式化，避免卡顿
  if (size > MAX_FORMAT_SIZE) {
    return jsonString
  }
  
  try {
    const parsed = JSON.parse(jsonString)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return jsonString
  }
}

/**
 * 格式化后的内容
 */
const formattedContent = computed(() => {
  if (!props.content) {
    return ''
  }

  const format = detectedFormat.value
  const size = contentSize.value

  // 如果手动触发了格式化，优先使用手动格式化状态
  // 否则使用 autoFormat 设置
  const shouldFormat = isManuallyFormatted.value !== null 
    ? isManuallyFormatted.value 
    : props.autoFormat

  // 超大内容时，如果未手动格式化且 autoFormat 为 false，则禁用格式化
  if (size > MAX_FORMAT_SIZE && !shouldFormat) {
    return props.content
  }

  // JSON 格式化
  if (format === 'json' && shouldFormat) {
    return formatJson(props.content)
  }

  // XML/SOAP 格式化（简单处理）
  if ((format === 'xml' || format === 'soap') && shouldFormat) {
    // 简单的 XML 格式化（可以后续扩展更完善的格式化）
    return props.content
  }

  return props.content
})

// 监听内容大小变化，自动调整高亮设置
watch(
  () => props.content,
  (newContent) => {
    if (!newContent) {
      enableHighlight.value = true
      isManuallyFormatted.value = null
      return
    }
    
    const size = new Blob([newContent]).size
    // 超大内容时禁用高亮以提升性能
    enableHighlight.value = size <= LARGE_CONTENT_THRESHOLD
    // 内容变化时重置手动格式化状态（使用 null 表示未手动操作）
    isManuallyFormatted.value = null
  },
  { immediate: true }
)

/**
 * 格式标签类型
 */
const formatTagType = computed(() => {
  const format = detectedFormat.value
  const typeMap: Record<TextFormat, 'default' | 'info' | 'success' | 'warning' | 'error' | 'primary'> = {
    json: 'success',
    xml: 'info',
    soap: 'primary',
    txt: 'default',
    yaml: 'warning',
    sql: 'info',
    javascript: 'success',
    typescript: 'primary',
    css: 'info',
    html: 'info',
    auto: 'default'
  }
  return typeMap[format] || 'default'
})

/**
 * 格式显示标签
 */
const formatLabel = computed(() => {
  const format = detectedFormat.value
  const labelMap: Record<TextFormat, string> = {
    json: 'JSON',
    xml: 'XML',
    soap: 'SOAP',
    txt: 'TXT',
    yaml: 'YAML',
    sql: 'SQL',
    javascript: 'JavaScript',
    typescript: 'TypeScript',
    css: 'CSS',
    html: 'HTML',
    auto: 'AUTO'
  }
  return labelMap[format] || 'TXT'
})

/**
 * 是否显示工具栏
 */
const showToolbar = computed(() => {
  return props.showCopyButton || (canFormat.value && !props.autoFormat)
})

/**
 * 是否可以格式化
 */
const canFormat = computed(() => {
  return detectedFormat.value === 'json' || detectedFormat.value === 'xml' || detectedFormat.value === 'soap'
})

/**
 * 计算样式
 */
const computedStyle = computed(() => {
  const style: Record<string, string | number> = {}
  if (props.style) {
    if (typeof props.style === 'string') {
      return props.style
    }
    Object.assign(style, props.style)
  }
  return style
})

/**
 * 内容区域样式
 */
const contentStyle = computed(() => {
  const style: Record<string, string | number> = {}
  if (props.maxHeight) {
    style.maxHeight = typeof props.maxHeight === 'number' ? `${props.maxHeight}px` : props.maxHeight
    style.overflow = 'auto'
  }
  if (props.minHeight) {
    style.minHeight = typeof props.minHeight === 'number' ? `${props.minHeight}px` : props.minHeight
  }
  return style
})

/**
 * 处理复制
 */
const handleCopy = async () => {
  try {
    const text = formattedContent.value || props.content
    await navigator.clipboard.writeText(text)
    message.success('复制成功')
    emit('copy', text)
  } catch (error) {
    message.error('复制失败')
    console.error('复制失败:', error)
  }
}

/**
 * 处理格式化
 */
const handleFormat = () => {
  const format = detectedFormat.value
  
  // 检查是否支持格式化
  if (format !== 'json' && format !== 'xml' && format !== 'soap') {
    message.warning('当前格式不支持格式化')
    return
  }
  
  // 获取当前格式化状态
  const currentFormatted = isManuallyFormatted.value !== null 
    ? isManuallyFormatted.value 
    : props.autoFormat
  
  // 切换手动格式化状态（取反）
  isManuallyFormatted.value = !currentFormatted
  
  if (isManuallyFormatted.value) {
    message.success('已格式化')
  } else {
    message.info('已取消格式化')
  }
}
</script>

<style scoped lang="scss">
.g-text-show {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;

  &__toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--g-space-xs) var(--g-space-sm);
    border-bottom: 1px solid var(--g-border-primary);
    background-color: var(--g-bg-secondary);

    &-left {
      display: flex;
      align-items: center;
      gap: var(--g-space-xs);
    }

    &-right {
      display: flex;
      align-items: center;
      gap: var(--g-space-xs);
    }
  }

  &__content {
    flex: 1;
    overflow: auto;
    position: relative;
  }

  &__code {
    width: 100%;
    height: 100%;
    margin: 0;
    border: none;
    border-radius: 0;

    :deep(.n-code) {
      height: 100%;
      overflow: auto;
    }
  }

  &__plain-text {
    width: 100%;
    height: 100%;
    margin: 0;
    padding: 12px;
    border: none;
    border-radius: 0;
    background-color: var(--n-code-color);
    color: var(--n-code-text-color);
    font-family: var(--n-font-family-mono);
    font-size: var(--n-font-size);
    line-height: 1.6;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-word;

    code {
      display: block;
      width: 100%;
      height: 100%;
    }
  }
}

.performance-tip {
  font-size: 12px;
  color: var(--n-warning-color);
  margin-left: 8px;
}
</style>

