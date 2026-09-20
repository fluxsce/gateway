// Package model 存放跨层共享的领域对象，不含 I/O 与业务编排。
//
// 这些类型同时被 contract、naming、config、store、stream 使用，字段变更
// 属于兼容面：只允许新增字段，既有字段的含义与零值语义保持不变。
// JSON/DB 列名不在本包承诺；持久化映射由 infra/store 单独维护。
package model
