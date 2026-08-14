/**
 * 将业务侧菜单图标名规范为 RsIcon（Lucide kebab-case）。
 * 兼容历史 ionicons5 PascalCase（如 EyeOutline）。
 */

const ICON_ALIASES: Record<string, string> = {
  add: 'plus',
  addoutline: 'plus',
  create: 'pencil',
  createoutline: 'pencil',
  edit: 'pencil',
  editoutline: 'pencil',
  trash: 'trash-2',
  trashoutline: 'trash-2',
  delete: 'trash-2',
  deleteoutline: 'trash-2',
  eye: 'eye',
  eyeoutline: 'eye',
  lock: 'lock',
  lockclosed: 'lock',
  lockclosedoutline: 'lock',
  person: 'user',
  personoutline: 'user',
  people: 'users',
  peopleoutline: 'users',
  peoplecircleoutline: 'users',
  refresh: 'refresh-cw',
  refreshoutline: 'refresh-cw',
  settings: 'settings',
  settingsoutline: 'settings',
  search: 'search',
  searchoutline: 'search',
  history: 'clock',
  historyoutline: 'clock',
  time: 'clock',
  timeoutline: 'clock',
  copy: 'copy',
  copyoutline: 'copy',
  chevronforward: 'chevron-right',
  chevronforwardoutline: 'chevron-right',
  key: 'key-round',
  keyoutline: 'key-round',
}

function pascalToKebab(name: string): string {
  return name
    .replace(/(Outline|Sharp|Filled|OutlineSharp)$/i, '')
    .replace(/([a-z0-9])([A-Z])/g, '$1-$2')
    .replace(/([A-Z])([A-Z][a-z])/g, '$1-$2')
    .replace(/[_\s]+/g, '-')
    .toLowerCase()
}

/**
 * 映射菜单图标到 Lucide 名；无法识别时原样返回（便于直接传 kebab 名）。
 */
export function mapRsGridMenuIcon(icon?: string): string | undefined {
  if (!icon) return undefined
  const trimmed = icon.trim()
  if (!trimmed) return undefined

  const compact = trimmed.replace(/[-_\s]/g, '').toLowerCase()
  if (ICON_ALIASES[compact]) return ICON_ALIASES[compact]

  // 已是 kebab-case（含数字，如 trash-2）
  if (/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(trimmed)) {
    return ICON_ALIASES[trimmed.replace(/-/g, '')] ?? trimmed
  }

  const kebab = pascalToKebab(trimmed)
  const kebabCompact = kebab.replace(/-/g, '')
  return ICON_ALIASES[kebabCompact] ?? kebab
}
