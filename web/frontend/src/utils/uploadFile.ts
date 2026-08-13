import { createUploadFileFromContent } from 'niuma-ui'

/** 将数据库中的文本内容还原为 RsUpload 可展示的 File[]（编辑/查看回填） */
export function filesFromTextContent(
  name: string,
  content: string,
  type = 'text/plain',
): File[] {
  if (!content) return []
  return [createUploadFileFromContent(name || 'file.txt', content, type)]
}

/** 读取本地 File 为文本（提交前写入 certContent 等字段） */
export async function readFileAsText(file: File): Promise<string> {
  return file.text()
}

/**
 * 从表单 File[] 字段提取文本内容与文件名到业务字段，并删除临时列表字段。
 * 编辑态若列表为空，则保留 fallbackPath（与旧逻辑一致）。
 */
export async function consumeTextFileField(
  data: Record<string, any>,
  listKey: string,
  contentKey: string,
  pathKey: string,
  options?: { fallbackPath?: string },
): Promise<void> {
  const list = data[listKey]
  delete data[listKey]

  const file = Array.isArray(list) ? list.find((item) => item instanceof File) : null
  if (file instanceof File) {
    data[contentKey] = await readFileAsText(file)
    data[pathKey] = file.name
    return
  }

  if (options?.fallbackPath && !data[pathKey]) {
    data[pathKey] = options.fallbackPath
  }
}
