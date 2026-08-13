/**
 * @deprecated 请直接使用 `@/ui` 的 `RsSplitPane`。本封装仅作过渡，后续将删除。
 *
 * 迁移示例：
 * ```vue
 * <RsSplitPane
 *   orientation="vertical"
 *   :panes="[{ key: 'top', size: 'auto' }, { key: 'bottom' }]"
 *   disabled
 * >
 *   <template #top>...</template>
 *   <template #bottom>...</template>
 * </RsSplitPane>
 * ```
 */
export { default as GPane } from './GPane.vue'
export * from './types'
