<template>
  <div
    :class="['g-richtext', props.class, { 'g-richtext--readonly': props.readonly }]"
    :style="containerStyle"
  >
    <!-- 工具栏 -->
    <div v-if="showToolbar" class="g-richtext__toolbar">
      <!-- 字体样式组 -->
      <div v-if="toolbarOptions.fontStyle" class="g-richtext__toolbar-group">
        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :type="editor?.isActive('bold') ? 'primary' : 'default'"
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().toggleBold().run()"
            >
              <template #icon>
                <n-icon><TextOutline /></n-icon>
              </template>
            </n-button>
          </template>
          粗体 (Ctrl+B)
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :type="editor?.isActive('italic') ? 'primary' : 'default'"
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().toggleItalic().run()"
            >
              <template #icon>
                <n-icon><TextOutline /></n-icon>
              </template>
            </n-button>
          </template>
          斜体 (Ctrl+I)
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :type="editor?.isActive('underline') ? 'primary' : 'default'"
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().toggleUnderline().run()"
            >
              <template #icon>
                <n-icon><TextOutline /></n-icon>
              </template>
            </n-button>
          </template>
          下划线 (Ctrl+U)
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :type="editor?.isActive('strike') ? 'primary' : 'default'"
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().toggleStrike().run()"
            >
              <template #icon>
                <n-icon><RemoveOutline /></n-icon>
              </template>
            </n-button>
          </template>
          删除线
        </n-tooltip>

      </div>

      <div v-if="toolbarOptions.fontStyle && toolbarOptions.fontFamily" class="g-richtext__divider" />

      <!-- 字体族 -->
      <div v-if="toolbarOptions.fontFamily" class="g-richtext__toolbar-group">
        <n-select
          v-model:value="currentFontFamily"
          :options="fontFamilyOptions"
          :disabled="!editor"
          size="small"
          style="width: 120px"
          placeholder="字体"
          @update:value="handleFontFamilyChange"
        />
      </div>

      <div v-if="toolbarOptions.fontFamily && toolbarOptions.textColor" class="g-richtext__divider" />

      <!-- 文本颜色 -->
      <div v-if="toolbarOptions.textColor" class="g-richtext__toolbar-group">
        <n-color-picker
          :value="currentTextColor"
          :show-alpha="false"
          :modes="['hex']"
          :actions="['confirm']"
          :disabled="!editor"
          size="small"
          @update:value="handleTextColorChange"
        />
      </div>

      <div v-if="toolbarOptions.textColor && toolbarOptions.textAlign" class="g-richtext__divider" />

      <!-- 对齐方式 -->
      <div v-if="toolbarOptions.textAlign" class="g-richtext__toolbar-group">
        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :type="editor?.isActive({ textAlign: 'left' }) ? 'primary' : 'default'"
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().setTextAlign('left').run()"
            >
              <template #icon>
                <n-icon><OptionsOutline /></n-icon>
              </template>
            </n-button>
          </template>
          左对齐
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :type="editor?.isActive({ textAlign: 'center' }) ? 'primary' : 'default'"
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().setTextAlign('center').run()"
            >
              <template #icon>
                <n-icon><OptionsOutline /></n-icon>
              </template>
            </n-button>
          </template>
          居中
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :type="editor?.isActive({ textAlign: 'right' }) ? 'primary' : 'default'"
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().setTextAlign('right').run()"
            >
              <template #icon>
                <n-icon><OptionsOutline /></n-icon>
              </template>
            </n-button>
          </template>
          右对齐
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :type="editor?.isActive({ textAlign: 'justify' }) ? 'primary' : 'default'"
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().setTextAlign('justify').run()"
            >
              <template #icon>
                <n-icon><MenuOutline /></n-icon>
              </template>
            </n-button>
          </template>
          两端对齐
        </n-tooltip>
      </div>

      <div v-if="toolbarOptions.textAlign && toolbarOptions.list" class="g-richtext__divider" />

      <!-- 列表 -->
      <div v-if="toolbarOptions.list" class="g-richtext__toolbar-group">
        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :type="editor?.isActive('bulletList') ? 'primary' : 'default'"
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().toggleBulletList().run()"
            >
              <template #icon>
                <n-icon><ListOutline /></n-icon>
              </template>
            </n-button>
          </template>
          无序列表
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :type="editor?.isActive('orderedList') ? 'primary' : 'default'"
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().toggleOrderedList().run()"
            >
              <template #icon>
                <n-icon><ListOutline /></n-icon>
              </template>
            </n-button>
          </template>
          有序列表
        </n-tooltip>
      </div>

      <div v-if="toolbarOptions.list && (toolbarOptions.heading || toolbarOptions.link || toolbarOptions.image || toolbarOptions.table)" class="g-richtext__divider" />

      <!-- 标题 -->
      <div v-if="toolbarOptions.heading" class="g-richtext__toolbar-group">
        <n-select
          v-model:value="currentHeading"
          :options="headingOptions"
          :disabled="!editor"
          size="small"
          style="width: 100px"
          placeholder="标题"
          @update:value="handleHeadingChange"
        />
      </div>

      <!-- 链接 -->
      <div v-if="toolbarOptions.link" class="g-richtext__toolbar-group">
        <n-popover trigger="click" :show="showLinkDialog" @update:show="showLinkDialog = $event">
          <template #trigger>
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button
                  :type="editor?.isActive('link') ? 'primary' : 'default'"
                  :disabled="!editor"
                  quaternary
                  size="small"
                  @click="handleLinkClick"
                >
                  <template #icon>
                    <n-icon><LinkOutline /></n-icon>
                  </template>
                </n-button>
              </template>
              插入链接
            </n-tooltip>
          </template>
          <div class="g-richtext__link-dialog">
            <n-input
              v-model:value="linkUrl"
              placeholder="输入链接地址"
              size="small"
              style="margin-bottom: 8px"
              @keyup.enter="handleLinkSubmit"
            />
            <n-space justify="end">
              <n-button size="small" @click="handleLinkRemove">移除链接</n-button>
              <n-button type="primary" size="small" @click="handleLinkSubmit">确定</n-button>
            </n-space>
          </div>
        </n-popover>
      </div>

      <!-- 图片 -->
      <div v-if="toolbarOptions.image" class="g-richtext__toolbar-group">
        <n-popover trigger="click" :show="showImageDialog" @update:show="showImageDialog = $event">
          <template #trigger>
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button
                  :disabled="!editor"
                  quaternary
                  size="small"
                  @click="showImageDialog = !showImageDialog"
                >
                  <template #icon>
                    <n-icon><ImageOutline /></n-icon>
                  </template>
                </n-button>
              </template>
              插入图片
            </n-tooltip>
          </template>
          <div class="g-richtext__image-dialog">
            <n-tabs v-model:value="imageTab" size="small">
              <n-tab-pane name="url" tab="图片地址">
                <n-input
                  v-model:value="imageUrl"
                  placeholder="输入图片地址"
                  size="small"
                  style="margin-bottom: 8px"
                  @keyup.enter="handleImageSubmit"
                />
              </n-tab-pane>
              <n-tab-pane name="upload" tab="本地上传">
                <n-upload
                  :file-list="imageFileList"
                  :max="1"
                  accept="image/*"
                  :show-file-list="false"
                  @change="handleImageFileChange"
                >
                  <n-button size="small" style="width: 100%">选择图片</n-button>
                </n-upload>
                <div v-if="imagePreview" style="margin-top: 8px; text-align: center;">
                  <img :src="imagePreview" alt="预览" style="max-width: 100%; max-height: 150px; border-radius: 4px;" />
                </div>
              </n-tab-pane>
            </n-tabs>
            <n-space justify="end" style="margin-top: 12px">
              <n-button size="small" @click="handleImageCancel">取消</n-button>
              <n-button type="primary" size="small" @click="handleImageSubmit" :disabled="!imageUrl && !imagePreview">确定</n-button>
            </n-space>
          </div>
        </n-popover>
      </div>

      <!-- 表格 -->
      <div v-if="toolbarOptions.table" class="g-richtext__toolbar-group">
        <n-popover trigger="click" placement="bottom" :show="showTableDialog" @update:show="showTableDialog = $event">
          <template #trigger>
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button
                  :type="editor?.isActive('table') ? 'primary' : 'default'"
                  :disabled="!editor"
                  quaternary
                  size="small"
                  @click="showTableDialog = !showTableDialog"
                >
                  <template #icon>
                    <n-icon><AppsOutline /></n-icon>
                  </template>
                </n-button>
              </template>
              插入表格
            </n-tooltip>
          </template>
          <div class="g-richtext__table-dialog">
            <div style="margin-bottom: 8px; font-size: 12px; color: #666;">选择表格大小</div>
            <div class="g-richtext__table-grid">
              <div
                v-for="row in 10"
                :key="row"
                class="g-richtext__table-grid-row"
              >
                <div
                  v-for="col in 10"
                  :key="col"
                  :class="[
                    'g-richtext__table-grid-cell',
                    { 'g-richtext__table-grid-cell--selected': row <= tableRows && col <= tableCols }
                  ]"
                  @mouseenter="tableRows = row; tableCols = col"
                  @click="handleInsertTable(row, col)"
                />
              </div>
            </div>
            <div style="margin-top: 8px; text-align: center; font-size: 12px; color: #666;">
              {{ tableRows }} × {{ tableCols }}
            </div>
          </div>
        </n-popover>

        <!-- 表格操作按钮（仅在表格内显示） -->
        <template v-if="editor?.isActive('table')">
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button
                :disabled="!editor"
                quaternary
                size="small"
                @click="editor?.chain().focus().addRowBefore().run()"
              >
                <template #icon>
                  <n-icon><ArrowUndoOutline /></n-icon>
                </template>
              </n-button>
            </template>
            在上方插入行
          </n-tooltip>

          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button
                :disabled="!editor"
                quaternary
                size="small"
                @click="editor?.chain().focus().addRowAfter().run()"
              >
                <template #icon>
                  <n-icon><ArrowRedoOutline /></n-icon>
                </template>
              </n-button>
            </template>
            在下方插入行
          </n-tooltip>

          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button
                :disabled="!editor"
                quaternary
                size="small"
                @click="editor?.chain().focus().deleteRow().run()"
              >
                <template #icon>
                  <n-icon><RemoveOutline /></n-icon>
                </template>
              </n-button>
            </template>
            删除行
          </n-tooltip>

          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button
                :disabled="!editor"
                quaternary
                size="small"
                @click="editor?.chain().focus().addColumnBefore().run()"
              >
                <template #icon>
                  <n-icon><OptionsOutline /></n-icon>
                </template>
              </n-button>
            </template>
            在左侧插入列
          </n-tooltip>

          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button
                :disabled="!editor"
                quaternary
                size="small"
                @click="editor?.chain().focus().addColumnAfter().run()"
              >
                <template #icon>
                  <n-icon><OptionsOutline /></n-icon>
                </template>
              </n-button>
            </template>
            在右侧插入列
          </n-tooltip>

          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button
                :disabled="!editor"
                quaternary
                size="small"
                @click="editor?.chain().focus().deleteColumn().run()"
              >
                <template #icon>
                  <n-icon><RemoveOutline /></n-icon>
                </template>
              </n-button>
            </template>
            删除列
          </n-tooltip>

          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button
                :disabled="!editor"
                quaternary
                size="small"
                @click="editor?.chain().focus().deleteTable().run()"
              >
                <template #icon>
                  <n-icon><CodeOutline /></n-icon>
                </template>
              </n-button>
            </template>
            删除表格
          </n-tooltip>
        </template>
      </div>

      <div v-if="toolbarOptions.table && toolbarOptions.other" class="g-richtext__divider" />

      <!-- 其他工具 -->
      <template v-if="toolbarOptions.other">
        <div class="g-richtext__divider" />

        <div class="g-richtext__toolbar-group">
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button
                :disabled="!editor"
                quaternary
                size="small"
                @click="editor?.chain().focus().setHorizontalRule().run()"
              >
              <template #icon>
                <n-icon><RemoveOutline /></n-icon>
              </template>
            </n-button>
          </template>
          插入水平线
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().clearNodes().unsetAllMarks().run()"
            >
              <template #icon>
                <n-icon><CodeOutline /></n-icon>
              </template>
            </n-button>
          </template>
          清除格式
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().undo().run()"
            >
              <template #icon>
                <n-icon><ArrowUndoOutline /></n-icon>
              </template>
            </n-button>
          </template>
          撤销 (Ctrl+Z)
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              :disabled="!editor"
              quaternary
              size="small"
              @click="editor?.chain().focus().redo().run()"
            >
              <template #icon>
                <n-icon><ArrowRedoOutline /></n-icon>
              </template>
              </n-button>
            </template>
            重做 (Ctrl+Y)
          </n-tooltip>
        </div>
      </template>
    </div>

    <!-- 编辑器内容区域 -->
    <EditorContent
      v-if="editor"
      :editor="editor"
      class="g-richtext__editor"
      :class="{ 'g-richtext__editor--readonly': props.readonly }"
    />
  </div>
</template>

<script setup lang="ts">
import Bold from '@tiptap/extension-bold'
import Color from '@tiptap/extension-color'
import FontFamily from '@tiptap/extension-font-family'
import Heading from '@tiptap/extension-heading'
import HorizontalRule from '@tiptap/extension-horizontal-rule'
import Image from '@tiptap/extension-image'
import Italic from '@tiptap/extension-italic'
import Link from '@tiptap/extension-link'
import Placeholder from '@tiptap/extension-placeholder'
import Strike from '@tiptap/extension-strike'
import { Table } from '@tiptap/extension-table'
import { TableCell } from '@tiptap/extension-table-cell'
import { TableHeader } from '@tiptap/extension-table-header'
import { TableRow } from '@tiptap/extension-table-row'
import TextAlign from '@tiptap/extension-text-align'
import { TextStyle } from '@tiptap/extension-text-style'
import Underline from '@tiptap/extension-underline'
import StarterKit from '@tiptap/starter-kit'
import { EditorContent, useEditor } from '@tiptap/vue-3'
import {
  AppsOutline,
  ArrowRedoOutline,
  ArrowUndoOutline,
  CodeOutline,
  ImageOutline,
  LinkOutline,
  ListOutline,
  MenuOutline,
  OptionsOutline,
  RemoveOutline,
  TextOutline
} from '@vicons/ionicons5'
import {
  NButton,
  NColorPicker,
  NIcon,
  NInput,
  NPopover,
  NSelect,
  NSpace,
  NTabPane,
  NTabs,
  NTooltip,
  NUpload
} from 'naive-ui'
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import type { GRichTextEmits, GRichTextExpose, GRichTextProps } from './types'

defineOptions({
  name: 'GRichText'
})

const props = withDefaults(defineProps<GRichTextProps>(), {
  modelValue: '',
  readonly: false,
  placeholder: '请输入内容...',
  showToolbar: true,
  toolbarOptions: () => ({
    fontStyle: true,
    fontFamily: true,
    textColor: true,
    textAlign: true,
    list: true,
    link: true,
    image: true,
    heading: true,
    table: true,
    other: true
  })
})

const emit = defineEmits<GRichTextEmits>()

const editor = useEditor({
  extensions: [
    StarterKit.configure({
      // 禁用已在下面单独配置的扩展，避免重复
      bold: false,
      italic: false,
      strike: false,
      heading: false,
      horizontalRule: false,
      link: false,
      // 注意：StarterKit 通常不包含 underline，但为了保险起见也禁用它
      underline: false,
      bulletList: {
        keepMarks: true,
        keepAttributes: false
      },
      orderedList: {
        keepMarks: true,
        keepAttributes: false
      }
    }),
    Bold,
    Italic,
    Underline,
    Strike,
    TextStyle,
    Color,
    FontFamily,
    Heading.configure({
      levels: [1, 2, 3, 4, 5, 6]
    }),
    TextAlign.configure({
      types: ['heading', 'paragraph']
    }),
    Link.configure({
      openOnClick: false,
      HTMLAttributes: {
        target: '_blank',
        rel: 'noopener noreferrer nofollow'
      }
    }),
    Image.configure({
      inline: false,
      allowBase64: true,
      HTMLAttributes: {
        class: 'g-richtext-image'
      }
    }),
    HorizontalRule,
    Table.configure({
      resizable: true
    }),
    TableRow,
    TableHeader,
    TableCell,
    Placeholder.configure({
      placeholder: props.placeholder
    }),
    // 过滤掉可能重复的扩展（通过名称去重）
    ...(props.extensions?.filter((ext: any) => {
      const extName = ext.name
      // 避免重复添加已配置的扩展
      const existingExtNames = [
        'bold', 'italic', 'strike', 'underline', 'heading',
        'horizontalRule', 'link', 'image', 'textStyle', 'color',
        'fontFamily', 'textAlign', 'bulletList', 'orderedList',
        'table', 'tableRow', 'tableHeader', 'tableCell', 'placeholder'
      ]
      return !existingExtNames.includes(extName)
    }) || [])
  ],
  content: props.modelValue,
  editable: !props.readonly,
  onUpdate: ({ editor }) => {
    const html = editor.getHTML()
    emit('update:modelValue', html)
    emit('change', html)
  },
  onFocus: () => {
    emit('focus')
  },
  onBlur: () => {
    emit('blur')
  },
  onCreate: ({ editor }) => {
    emit('ready', editor)
  }
})

// 工具栏状态
const showLinkDialog = ref(false)
const showImageDialog = ref(false)
const showTableDialog = ref(false)
const linkUrl = ref('')
const imageUrl = ref('')
const imageTab = ref<'url' | 'upload'>('url')
const imageFileList = ref<any[]>([])
const imagePreview = ref<string>('')
const tableRows = ref(3)
const tableCols = ref(3)

// 字体族选项
const fontFamilyOptions = [
  { label: '默认', value: '' },
  { label: 'Arial', value: 'Arial' },
  { label: 'Comic Sans MS', value: 'Comic Sans MS' },
  { label: 'Courier New', value: 'Courier New' },
  { label: 'Georgia', value: 'Georgia' },
  { label: 'Helvetica', value: 'Helvetica' },
  { label: 'Impact', value: 'Impact' },
  { label: 'Lucida Console', value: 'Lucida Console' },
  { label: 'Tahoma', value: 'Tahoma' },
  { label: 'Times New Roman', value: 'Times New Roman' },
  { label: 'Trebuchet MS', value: 'Trebuchet MS' },
  { label: 'Verdana', value: 'Verdana' },
  { label: '宋体', value: 'SimSun' },
  { label: '黑体', value: 'SimHei' },
  { label: '微软雅黑', value: 'Microsoft YaHei' },
  { label: '楷体', value: 'KaiTi' }
]

// 标题选项
const headingOptions = [
  { label: '正文', value: 0 },
  { label: '标题 1', value: 1 },
  { label: '标题 2', value: 2 },
  { label: '标题 3', value: 3 },
  { label: '标题 4', value: 4 },
  { label: '标题 5', value: 5 },
  { label: '标题 6', value: 6 }
]

// 当前字体族
const currentFontFamily = computed({
  get: () => {
    if (!editor.value) return undefined
    return editor.value.getAttributes('textStyle').fontFamily || undefined
  },
  set: (value) => {
    if (value) {
      editor.value?.chain().focus().setFontFamily(value).run()
    } else {
      editor.value?.chain().focus().unsetFontFamily().run()
    }
  }
})

// 当前文本颜色
const currentTextColor = computed({
  get: () => {
    if (!editor.value) return '#000000'
    return editor.value.getAttributes('textStyle').color || '#000000'
  },
  set: (value) => {
    // handled by handleTextColorChange
  }
})

// 当前标题级别
const currentHeading = computed({
  get: () => {
    if (!editor.value) return 0
    if (editor.value.isActive('heading', { level: 1 })) return 1
    if (editor.value.isActive('heading', { level: 2 })) return 2
    if (editor.value.isActive('heading', { level: 3 })) return 3
    if (editor.value.isActive('heading', { level: 4 })) return 4
    if (editor.value.isActive('heading', { level: 5 })) return 5
    if (editor.value.isActive('heading', { level: 6 })) return 6
    return 0
  },
  set: (value) => {
    if (!editor.value) return
    if (value === 0) {
      editor.value.chain().focus().setParagraph().run()
    } else {
      editor.value.chain().focus().toggleHeading({ level: value as 1 | 2 | 3 | 4 | 5 | 6 }).run()
    }
  }
})

// 处理字体族变化
const handleFontFamilyChange = (value: string) => {
  if (value) {
    editor.value?.chain().focus().setFontFamily(value).run()
  } else {
    editor.value?.chain().focus().unsetFontFamily().run()
  }
}

// 处理文本颜色变化
const handleTextColorChange = (value: string) => {
  if (!value) return
  // 确保颜色值是有效的十六进制格式
  let colorValue = value
  if (typeof value === 'string') {
    // 如果已经是 # 开头的 hex 格式，直接使用
    if (value.startsWith('#')) {
      colorValue = value.toUpperCase()
    } else {
      // 其他格式，尝试添加 #
      colorValue = `#${value}`.toUpperCase()
    }
  }
  editor.value?.chain().focus().setColor(colorValue).run()
}

// 处理标题变化
const handleHeadingChange = (value: number) => {
  if (value === 0) {
    editor.value?.chain().focus().setParagraph().run()
  } else {
    editor.value?.chain().focus().toggleHeading({ level: value as 1 | 2 | 3 | 4 | 5 | 6 }).run()
  }
}

// 处理链接点击
const handleLinkClick = () => {
  if (editor.value?.isActive('link')) {
    const attrs = editor.value.getAttributes('link')
    linkUrl.value = attrs.href || ''
  }
  showLinkDialog.value = !showLinkDialog.value
}

// 处理链接提交
const handleLinkSubmit = () => {
  if (linkUrl.value) {
    editor.value?.chain().focus().extendMarkRange('link').setLink({ href: linkUrl.value }).run()
  }
  showLinkDialog.value = false
  linkUrl.value = ''
}

// 处理移除链接
const handleLinkRemove = () => {
  editor.value?.chain().focus().unsetLink().run()
  showLinkDialog.value = false
  linkUrl.value = ''
}

// 处理图片文件选择
const handleImageFileChange = (options: { file: any; fileList: any[] }) => {
  const { file } = options
  if (file.file) {
    // 验证文件类型
    if (!file.file.type.startsWith('image/')) {
      return
    }
    
    // 读取文件为 Data URL
    const reader = new FileReader()
    reader.onload = (e) => {
      const result = e.target?.result as string
      imagePreview.value = result
    }
    reader.readAsDataURL(file.file)
  }
}

// 处理图片取消
const handleImageCancel = () => {
  showImageDialog.value = false
  imageUrl.value = ''
  imagePreview.value = ''
  imageFileList.value = []
  imageTab.value = 'url'
}

// 处理图片提交
const handleImageSubmit = () => {
  let src = ''
  
  if (imageTab.value === 'url') {
    // URL 方式
    src = imageUrl.value.trim()
  } else if (imageTab.value === 'upload' && imagePreview.value) {
    // 本地上传方式，使用 Data URL
    src = imagePreview.value
  }
  
  if (src) {
    // 确保图片能正确渲染，添加 alt 属性
    editor.value?.chain().focus().setImage({ 
      src,
      alt: '图片',
      title: '图片'
    }).run()
  }
  
  handleImageCancel()
}

// 处理插入表格
const handleInsertTable = (rows: number, cols: number) => {
  editor.value?.chain().focus().insertTable({ rows, cols, withHeaderRow: true }).run()
  showTableDialog.value = false
  tableRows.value = 3
  tableCols.value = 3
}

// 容器样式
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
    if (editor.value && editor.value.getHTML() !== newValue) {
      editor.value.commands.setContent(newValue || '', { emitUpdate: false })
    }
  }
)

// 监听只读状态
watch(
  () => props.readonly,
  (readonly) => {
    editor.value?.setEditable(!readonly)
  }
)

// 组件挂载时编辑器已经通过 useEditor 创建

// 组件卸载时销毁编辑器
onBeforeUnmount(() => {
  editor.value?.destroy()
})

// 暴露方法
defineExpose<GRichTextExpose>({
  getEditor: () => editor.value ?? null,
  getHTML: () => editor.value?.getHTML() || '',
  setHTML: (html: string) => editor.value?.commands.setContent(html, { emitUpdate: false }),
  getText: () => editor.value?.getText() || '',
  setText: (text: string) => editor.value?.commands.setContent(text, { emitUpdate: false }),
  getJSON: () => editor.value?.getJSON() || null,
  setJSON: (json: any) => editor.value?.commands.setContent(json, { emitUpdate: false }),
  focus: () => editor.value?.commands.focus(),
  blur: () => editor.value?.commands.blur(),
  clear: () => editor.value?.commands.clearContent()
})
</script>

<style scoped lang="scss">
.g-richtext {
  width: 100%;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--n-border-color, #e0e0e6);
  border-radius: var(--n-border-radius, 4px);
  background-color: var(--n-color, #ffffff);
  overflow: hidden;

  &--readonly {
    .g-richtext__toolbar {
      opacity: 0.6;
      pointer-events: none;
    }
  }

  &__toolbar {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 8px;
    border-bottom: 1px solid var(--n-border-color, #e0e0e6);
    background-color: var(--n-color, #ffffff);
    flex-wrap: wrap;

    &-group {
      display: flex;
      align-items: center;
      gap: 2px;
    }
  }

  // 颜色选择器样式优化 - 让触发器更清晰可见
  :deep(.g-richtext__toolbar-group .n-color-picker) {
    .n-base-selection {
      min-width: 80px;
    }

    .n-color-picker-trigger {
      min-width: 80px;
      
      .n-color-picker-trigger__value {
        font-size: 12px;
      }
    }
  }

  &__divider {
    width: 1px;
    height: 20px;
    background-color: var(--n-border-color, #e0e0e6);
    margin: 0 4px;
  }

  &__link-dialog,
  &__image-dialog {
    padding: 8px;
    min-width: 280px;
  }

  &__table-dialog {
    padding: 8px;
    min-width: 240px;
  }

  &__table-grid {
    display: grid;
    grid-template-columns: repeat(10, 1fr);
    gap: 2px;
    width: 100%;
  }

  &__table-grid-row {
    display: contents;
  }

  &__table-grid-cell {
    width: 16px;
    height: 16px;
    border: 1px solid var(--n-border-color, #e0e0e6);
    background-color: var(--n-color, #ffffff);
    cursor: pointer;
    transition: background-color 0.2s;

    &:hover {
      background-color: var(--n-color-hover, #f5f5f5);
    }

    &--selected {
      background-color: var(--n-primary-color, #18a058);
      border-color: var(--n-primary-color, #18a058);
    }
  }

  &__editor {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    min-height: 200px;
    max-height: 600px;

    &--readonly {
      background-color: var(--n-color-disabled, #f5f5f5);
    }

    :deep(.ProseMirror) {
      outline: none;
      min-height: 200px;

      p {
        margin: 0.5em 0;

        &.is-editor-empty:first-child::before {
          content: attr(data-placeholder);
          float: left;
          color: var(--n-placeholder-color, #aaa);
          pointer-events: none;
          height: 0;
        }
      }

      h1, h2, h3, h4, h5, h6 {
        margin: 0.5em 0;
        font-weight: 600;
        line-height: 1.4;
      }

      h1 {
        font-size: 2em;
      }

      h2 {
        font-size: 1.5em;
      }

      h3 {
        font-size: 1.25em;
      }

      h4 {
        font-size: 1.1em;
      }

      h5 {
        font-size: 1em;
      }

      h6 {
        font-size: 0.9em;
      }

      ul, ol {
        padding-left: 1.5em;
        margin: 0.5em 0;
      }

      li {
        margin: 0.25em 0;
      }

      blockquote {
        border-left: 3px solid var(--n-border-color, #e0e0e6);
        padding-left: 1em;
        margin: 0.5em 0;
        color: var(--n-text-color-2, #666);
      }

      hr {
        border: none;
        border-top: 2px solid var(--n-border-color, #e0e0e6);
        margin: 1em 0;
      }

      a {
        color: var(--n-primary-color, #18a058);
        text-decoration: underline;
        cursor: pointer;

        &:hover {
          color: var(--n-primary-color-hover, #36ad6a);
        }
      }

      img {
        max-width: 100%;
        height: auto;
        border-radius: 4px;
        margin: 0.5em 0;
        display: block;
        
        &.g-richtext-image {
          display: block;
          margin: 0.5em auto;
        }
      }

      table {
        border-collapse: collapse;
        margin: 0.5em 0;
        width: 100%;

        td, th {
          border: 1px solid var(--n-border-color, #e0e0e6);
          padding: 8px;
          min-width: 1em;
        }

        th {
          background-color: var(--n-color-disabled, #f5f5f5);
          font-weight: 600;
        }
      }

      code {
        background-color: var(--n-color-disabled, #f5f5f5);
        padding: 2px 4px;
        border-radius: 3px;
        font-family: 'Courier New', monospace;
        font-size: 0.9em;
      }

      pre {
        background-color: var(--n-color-disabled, #f5f5f5);
        padding: 1em;
        border-radius: 4px;
        overflow-x: auto;
        margin: 0.5em 0;

        code {
          background: none;
          padding: 0;
        }
      }
    }
  }
}
</style>
