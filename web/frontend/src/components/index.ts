/**
 * 公共组件临时精简入口：仅保留 Rs 表单封装与 GFieldset。
 * 其它组件请按子路径直连导入（如 `@/components/gicon`），避免 barrel 连带加载旧表单。
 */

// 数据编辑表单（niuma-ui 新实现）
export { RsDataForm, RsDataFormModal } from './form/rs-data'
export type {
  RsDataFormField,
  RsDataFormFieldsEmits,
  RsDataFormFieldsProps,
  RsDataFormFieldType,
  RsDataFormModalEmits,
  RsDataFormModalProps,
  RsDataFormProps,
  RsDataFormSelectOption,
  RsDataFormTab,
  RsDataModalMode,
} from './form/rs-data'

// Fieldset 字段分组组件
export { GFieldset } from './gfieldset'
export * from './gfieldset/types'
