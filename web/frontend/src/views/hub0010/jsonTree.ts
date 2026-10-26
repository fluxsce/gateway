export type JsonKind = 'object' | 'array' | 'string' | 'number' | 'boolean' | 'null'

export interface JsonTreeNode {
  id: string
  label: string
  kind: JsonKind
  preview: string
  title: string
  children: JsonTreeNode[]
}

const PREVIEW_MAX = 72

function kindOf(value: unknown): JsonKind {
  if (value === null) return 'null'
  if (Array.isArray(value)) return 'array'
  if (typeof value === 'string') return 'string'
  if (typeof value === 'number') return 'number'
  if (typeof value === 'boolean') return 'boolean'
  if (typeof value === 'object') return 'object'
  return 'null'
}

function displayKey(key: string, parent: 'object' | 'array'): string {
  if (parent === 'array') return key
  if (/^[A-Za-z_$][A-Za-z0-9_$]*$/.test(key)) return key
  return JSON.stringify(key)
}

function previewOf(kind: JsonKind, value: unknown, size: number): { preview: string; title: string } {
  if (kind === 'object') return { preview: `{${size}}`, title: '' }
  if (kind === 'array') return { preview: `[${size}]`, title: '' }
  if (kind === 'string') {
    const full = JSON.stringify(value)
    if (full.length <= PREVIEW_MAX) return { preview: full, title: '' }
    return { preview: `${full.slice(0, PREVIEW_MAX - 1)}…`, title: full }
  }
  if (kind === 'null') return { preview: 'null', title: '' }
  return { preview: String(value), title: '' }
}

function build(key: string, value: unknown, path: string, parent: 'object' | 'array' | 'root'): JsonTreeNode {
  const kind = kindOf(value)
  const label = parent === 'root' ? '' : displayKey(key, parent)
  if (kind === 'object') {
    const record = value as Record<string, unknown>
    const children = Object.keys(record).map((name) => build(name, record[name], `${path}.${name}`, 'object'))
    const shown = previewOf(kind, value, children.length)
    return { id: path, label, kind, ...shown, children }
  }
  if (kind === 'array') {
    const items = value as unknown[]
    const children = items.map((item, index) => build(String(index), item, `${path}[${index}]`, 'array'))
    const shown = previewOf(kind, value, children.length)
    return { id: path, label, kind, ...shown, children }
  }
  const shown = previewOf(kind, value, 0)
  return { id: path, label, kind, ...shown, children: [] }
}

/** 把 JSON 值收成可展开的对象树。对象键保持解析后的顺序。 */
export function buildJsonTree(value: unknown): JsonTreeNode {
  return build('', value, '$', 'root')
}

function countNodes(node: JsonTreeNode): number {
  let count = 1
  for (const child of node.children) count += countNodes(child)
  return count
}

/** 默认展开前两层；节点不多时全部展开。 */
export function initialOpenIds(node: JsonTreeNode): string[] {
  const total = countNodes(node)
  const ids: string[] = []
  const walk = (current: JsonTreeNode, depth: number) => {
    if (current.children.length === 0) return
    if (total <= 60 || depth < 2) ids.push(current.id)
    for (const child of current.children) walk(child, depth + 1)
  }
  walk(node, 0)
  return ids
}

/** 按键、值和路径过滤。命中节点保留其子节点，未命中则只留下命中的后代。 */
export function filterJsonTree(node: JsonTreeNode, query: string): JsonTreeNode | null {
  const needle = query.trim().toLowerCase()
  if (!needle) return node
  const self = `${node.label}\n${node.preview}\n${node.id}`.toLowerCase().includes(needle)
  if (self) return node
  const children = node.children
    .map((child) => filterJsonTree(child, needle))
    .filter((child): child is JsonTreeNode => child !== null)
  if (children.length === 0) return null
  return { ...node, children }
}

export function allBranchIds(node: JsonTreeNode): string[] {
  const ids: string[] = []
  const walk = (current: JsonTreeNode) => {
    if (current.children.length === 0) return
    ids.push(current.id)
    for (const child of current.children) walk(child)
  }
  walk(node)
  return ids
}
