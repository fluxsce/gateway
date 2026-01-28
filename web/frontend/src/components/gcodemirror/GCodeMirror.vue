<template>
  <div
    ref="containerRef"
    :class="['g-codemirror', props.class]"
    :style="containerStyle"
  />
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/user'
import { closeBrackets } from '@codemirror/autocomplete'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { bracketMatching, defaultHighlightStyle, foldGutter, syntaxHighlighting } from '@codemirror/language'
import { highlightSelectionMatches, searchKeymap } from '@codemirror/search'
import { Compartment, EditorState, type Extension } from '@codemirror/state'
import { oneDark } from '@codemirror/theme-one-dark'
import type { EditorView as EditorViewType } from '@codemirror/view'
import { EditorView, highlightActiveLine, highlightActiveLineGutter, keymap, lineNumbers, type ViewUpdate } from '@codemirror/view'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { CodeMirrorLanguage, GCodeMirrorEmits, GCodeMirrorProps } from './types'

// 语言包缓存（避免重复加载）
const languageCache = new Map<CodeMirrorLanguage, Extension>()

/**
 * 动态加载语言包（按需加载，减少初始内存开销）
 */
const loadLanguageExtension = async (lang: CodeMirrorLanguage): Promise<Extension> => {
  // 如果已缓存，直接返回
  if (languageCache.has(lang)) {
    return languageCache.get(lang)!
  }

  // 如果不是 plaintext，动态导入对应的语言包
  if (lang === 'plaintext') {
    languageCache.set(lang, [])
    return []
  }

  try {
    let extension: Extension

    switch (lang) {
      case 'javascript':
        extension = (await import('@codemirror/lang-javascript')).javascript()
        break
      case 'typescript':
        extension = (await import('@codemirror/lang-javascript')).javascript({ typescript: true })
        break
      case 'json':
        extension = (await import('@codemirror/lang-json')).json()
        break
      case 'html':
        extension = (await import('@codemirror/lang-html')).html()
        break
      case 'css':
        extension = (await import('@codemirror/lang-css')).css()
        break
      case 'xml':
        extension = (await import('@codemirror/lang-xml')).xml()
        break
      case 'sql':
        extension = (await import('@codemirror/lang-sql')).sql()
        break
      case 'yaml':
        extension = (await import('@codemirror/lang-yaml')).yaml()
        break
      case 'markdown':
        extension = (await import('@codemirror/lang-markdown')).markdown()
        break
      case 'python':
        extension = (await import('@codemirror/lang-python')).python()
        break
      case 'java':
        extension = (await import('@codemirror/lang-java')).java()
        break
      case 'go':
        extension = (await import('@codemirror/lang-go')).go()
        break
      case 'rust':
        extension = (await import('@codemirror/lang-rust')).rust()
        break
      case 'shell':
        // Shell 脚本使用 JavaScript 语言包作为替代（CodeMirror 6 没有专门的 shell 语言包）
        extension = (await import('@codemirror/lang-javascript')).javascript()
        break
      case 'properties':
        // Properties 文件格式，使用 legacy-modes 插件
        // 根据官网文档：https://codemirror.net/docs/legacy-modes/
        const { properties } = await import('@codemirror/legacy-modes/mode/properties')
        const { StreamLanguage } = await import('@codemirror/language')
        extension = StreamLanguage.define(properties)
        break
      default:
        extension = []
    }

    // 缓存加载的语言包
    languageCache.set(lang, extension)
    return extension
  } catch (error) {
    console.warn(`Failed to load language package for ${lang}:`, error)
    return []
  }
}

// 定义组件名称
defineOptions({
  name: 'GCodeMirror'
})

// Props
const props = withDefaults(defineProps<GCodeMirrorProps>(), {
  modelValue: '',
  language: 'plaintext',
  theme: 'auto',
  readonly: false,
  lineNumbers: true,
  foldGutter: true,
  lineWrapping: false,
  highlightActiveLine: true,
  bracketMatching: true,
  autoCloseBrackets: true,
  searchKeymap: true
})

// Emits
const emit = defineEmits<GCodeMirrorEmits>()

// 容器引用
const containerRef = ref<HTMLDivElement>()

// 编辑器实例
let editorView: EditorViewType | null = null

// 语言配置
const languageCompartment = new Compartment()

// 主题配置
const themeCompartment = new Compartment()

// 只读配置
const readonlyCompartment = new Compartment()

// 用户 store
const userStore = useUserStore()

// 检测主题
const isDark = computed(() => {
  const theme = userStore.theme
  if (theme === 'system') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  return theme === 'dark'
})

/**
 * 获取语言扩展（同步版本，用于初始化）
 * 使用空扩展作为占位符，实际语言包在异步加载后更新
 */
const getLanguageExtensionSync = (lang: CodeMirrorLanguage): Extension => {
  // 如果已缓存，直接返回
  if (languageCache.has(lang)) {
    return languageCache.get(lang)!
  }
  // 否则返回空扩展（作为占位符）
  return []
}

/**
 * 获取主题扩展
 */
const getThemeExtension = (theme: 'light' | 'dark' | 'auto'): Extension => {
  const shouldUseDark = theme === 'dark' || (theme === 'auto' && isDark.value)
  return shouldUseDark ? oneDark : []
}

/**
 * 构建编辑器扩展（异步版本，确保语言包已加载）
 */
const buildExtensions = async (): Promise<Extension[]> => {
  const extensions: Extension[] = []

  // 行号
  if (props.lineNumbers) {
    extensions.push(lineNumbers())
  }

  // 折叠
  if (props.foldGutter) {
    extensions.push(foldGutter())
  }

  // 括号匹配
  if (props.bracketMatching) {
    extensions.push(bracketMatching())
  }

  // 自动闭合括号
  if (props.autoCloseBrackets) {
    extensions.push(closeBrackets())
  }

  // 历史记录
  extensions.push(history())

  // 高亮选中匹配
  extensions.push(highlightSelectionMatches())

  // 搜索快捷键
  if (props.searchKeymap) {
    extensions.push(keymap.of(searchKeymap))
  }

  // 默认快捷键
  extensions.push(keymap.of(defaultKeymap))
  extensions.push(keymap.of(historyKeymap))

  // 高亮当前行
  if (props.highlightActiveLine) {
    extensions.push(highlightActiveLine())
    extensions.push(highlightActiveLineGutter())
  }

  // 语言支持（异步加载，确保使用正确的实例）
  const languageExtension = await loadLanguageExtension(props.language)
  extensions.push(languageCompartment.of(languageExtension))

  // 语法高亮（确保语法高亮功能被启用）
  // 注意：
  // 1. plaintext 模式不支持语法高亮（纯文本没有语法结构）
  // 2. 其他语言包已经包含了语法高亮，但我们需要确保默认高亮样式被应用
  // 3. 如果使用暗色主题，oneDark 主题会提供自己的高亮样式
  if (props.language !== 'plaintext') {
    const shouldUseDark = props.theme === 'dark' || (props.theme === 'auto' && isDark.value)
    if (!shouldUseDark) {
      // 浅色主题使用默认高亮样式
      // 注意：语言包已经包含了语法高亮功能，这里只是确保默认样式被应用
      extensions.push(syntaxHighlighting(defaultHighlightStyle))
    }
    // 暗色主题的语法高亮由 oneDark 主题自动提供，无需额外添加
  }

  // 主题
  extensions.push(themeCompartment.of(getThemeExtension(props.theme)))

  // 只读
  extensions.push(readonlyCompartment.of(EditorState.readOnly.of(props.readonly)))

  // 自动换行
  if (props.lineWrapping) {
    extensions.push(EditorView.lineWrapping)
  }

  // 占位符（通过 CSS 实现）
  if (props.placeholder) {
    extensions.push(
      EditorView.contentAttributes.of({
        'data-placeholder': props.placeholder
      })
    )
  }

  // 内容变化监听
  extensions.push(
    EditorView.updateListener.of((update: ViewUpdate) => {
      if (update.docChanged) {
        const content = update.state.doc.toString()
        emit('update:modelValue', content)
        emit('change', content)
      }
    })
  )

  // 自定义扩展
  if (props.extensions && props.extensions.length > 0) {
    extensions.push(...props.extensions)
  }

  return extensions
}

/**
 * 初始化编辑器
 */
const initEditor = async () => {
  if (!containerRef.value) return

  // 构建扩展（异步，确保语言包已加载）
  const extensions = await buildExtensions()

  // 创建编辑器状态
  const state = EditorState.create({
    doc: props.modelValue || '',
    extensions
  })

  // 创建编辑器视图
  editorView = new EditorView({
    state,
    parent: containerRef.value
  })

  // 监听焦点事件
  editorView.dom.addEventListener('focus', () => {
    emit('focus')
  })

  editorView.dom.addEventListener('blur', () => {
    emit('blur')
  })

  // 触发 ready 事件
  emit('ready', editorView)
}

/**
 * 更新编辑器内容
 */
const updateContent = (content: string) => {
  if (!editorView) return

  const currentContent = editorView.state.doc.toString()
  if (currentContent !== content) {
    editorView.dispatch({
      changes: {
        from: 0,
        to: editorView.state.doc.length,
        insert: content || ''
      }
    })
  }
}

/**
 * 更新语言（异步加载语言包）
 */
const updateLanguage = async () => {
  if (!editorView) return

  // 异步加载语言包
  const extension = await loadLanguageExtension(props.language)
  
  editorView.dispatch({
    effects: languageCompartment.reconfigure(extension)
  })
}

/**
 * 更新主题
 */
const updateTheme = () => {
  if (!editorView) return

  editorView.dispatch({
    effects: themeCompartment.reconfigure(getThemeExtension(props.theme))
  })
}

/**
 * 更新只读状态
 */
const updateReadonly = () => {
  if (!editorView) return

  editorView.dispatch({
    effects: readonlyCompartment.reconfigure(EditorState.readOnly.of(props.readonly))
  })
}

/**
 * 容器样式
 */
const containerStyle = computed(() => {
  const style: Record<string, string | number> = {}

  if (props.width) {
    style.width = typeof props.width === 'number' ? `${props.width}px` : props.width
  }

  if (props.height) {
    style.height = typeof props.height === 'number' ? `${props.height}px` : props.height
  } else {
    if (props.minHeight) {
      style.minHeight = typeof props.minHeight === 'number' ? `${props.minHeight}px` : props.minHeight
    }
    if (props.maxHeight) {
      style.maxHeight = typeof props.maxHeight === 'number' ? `${props.maxHeight}px` : props.maxHeight
    }
  }

  if (props.style) {
    if (typeof props.style === 'string') {
      return props.style
    }
    Object.assign(style, props.style)
  }

  return style
})

// 监听内容变化
watch(
  () => props.modelValue,
  (newValue) => {
    if (editorView) {
      updateContent(newValue || '')
    }
  }
)

// 监听语言变化
watch(
  () => props.language,
  async () => {
    await updateLanguage()
  }
)

// 监听主题变化
watch(
  [() => props.theme, isDark],
  () => {
    updateTheme()
  }
)

// 监听只读变化
watch(
  () => props.readonly,
  () => {
    updateReadonly()
  }
)

// 组件挂载时初始化
onMounted(() => {
  initEditor()
})

// 组件卸载时销毁编辑器
onBeforeUnmount(() => {
  if (editorView) {
    editorView.destroy()
    editorView = null
  }
})

// 暴露编辑器实例供外部使用
defineExpose({
  editorView: () => editorView,
  getValue: () => editorView?.state.doc.toString() || '',
  setValue: (value: string) => updateContent(value),
  focus: () => editorView?.focus(),
  blur: () => editorView?.contentDOM.blur()
})
</script>

<style scoped lang="scss">
.g-codemirror {
  width: 100%;
  height: 100%;
  overflow: hidden;
  border: 1px solid var(--g-border-primary, #e0e0e6);
  border-radius: var(--g-border-radius, 4px);
  background-color: var(--g-bg-color, #ffffff);

  :deep(.cm-editor) {
    height: 100%;
    font-size: 14px;
    font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  }

  :deep(.cm-scroller) {
    overflow: auto;
  }

  :deep(.cm-content) {
    padding: var(--g-space-sm, 8px);
    min-height: 100%;
    //内容样式和序号保持一致
    background-color: var(--g-bg-secondary, #f5f5f5);
  }

  :deep(.cm-content[data-placeholder]:empty::before) {
    content: attr(data-placeholder);
    color: var(--g-text-tertiary, #a3a3a3);
    pointer-events: none;
  }

  :deep(.cm-gutters) {
    background-color: var(--g-bg-secondary, #f5f5f5);
    border-right: 1px solid var(--g-border-primary, #e0e0e6);
  }

  :deep(.cm-lineNumbers) {
    color: var(--g-text-secondary, #999);
  }

  :deep(.cm-activeLine) {
    background-color: var(--g-hover-overlay, rgba(0, 0, 0, 0.04));
  }

  :deep(.cm-activeLineGutter) {
    background-color: var(--g-hover-overlay, rgba(0, 0, 0, 0.04));
  }

  // 只读状态样式
  &.g-codemirror--readonly {
    :deep(.cm-editor) {
      cursor: default;
    }

    :deep(.cm-content) {
      background-color: var(--g-bg-secondary, #f5f5f5);
    }
  }
}

// 暗色主题样式（通过 CodeMirror 的 oneDark 主题处理）
</style>

