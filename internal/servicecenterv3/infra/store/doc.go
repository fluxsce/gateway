// Package store 是持久化适配层，把领域对象映射到既有 HUB_SERVICE* 表。
//
// 本包不是稳定面：表结构演进、SQL 方言与行映射可以改，但不得把 ERP 表模型
// 泄漏到 contract / model。上层只应依赖 Store 的方法语义，不要依赖导出的
// 行结构（未导出的 *Row 类型）。
package store
