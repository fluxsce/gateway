// 构建标签说明：
// no_oracle 表示禁用Oracle数据库支持
// 当项目不需要Oracle数据库支持时，使用此文件中的空实现
// 使用方法：
//   1. 使用 go build -tags no_oracle 命令编译时，此文件会被使用
//   2. 可以减少最终二进制文件的大小，避免引入Oracle相关依赖
//   3. 适用于不需要Oracle支持的开发环境或轻量级部署
//
//go:build no_oracle
// +build no_oracle

package sqlutils

import "reflect"

// IsOracleSpecificType Oracle特有类型检查的禁用版本
//
// 参数：
//   - value: 要检查的值（在禁用版本中不会被使用）
//
// 返回：
//   - bool: 始终返回false
//
// 说明：
//   - 这是Oracle支持被禁用时的替代实现
//   - 用于在不需要Oracle支持时编译使用
//   - 通过build tag "no_oracle" 控制是否启用
//   - 始终返回false，表示没有Oracle特有类型需要处理
//   - 这个实现可以减少不必要的依赖和编译体积
func (fm *FieldMapper) IsOracleSpecificType(value interface{}) bool {
    return false
}

// convertOracleValue Oracle类型转换的禁用版本
//
// 参数：
//   - field: 目标字段的反射值（在禁用版本中不会被使用）
//   - value: 要转换的值（在禁用版本中不会被使用）
//
// 返回：
//   - error: 始终返回nil
//
// 说明：
//   - 这是Oracle支持被禁用时的替代实现
//   - 在不需要Oracle支持的环境中使用
//   - 直接返回nil，表示不进行任何Oracle相关的类型转换
//   - 允许代码在没有Oracle支持的环境中正常编译和运行
func (fm *FieldMapper) convertOracleValue(field reflect.Value, value interface{}) error {
    return nil
}