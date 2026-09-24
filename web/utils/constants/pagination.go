package constants

// 分页相关常量

// DefaultPage 默认页码
const DefaultPage = "1"

// DefaultPageSize 未配置租户分页时的默认每页记录数，与环境设置缺省一致。
const DefaultPageSize = "20"

// MinPageSize 最小每页记录数
const MinPageSize = 1

// MaxPageSize 最大每页记录数，与 syssetting.PageSizeCeiling 一致。
const MaxPageSize = 200

// DefaultSortField 默认排序字段
const DefaultSortField = "addTime"

// DefaultSortOrder 默认排序方向
const DefaultSortOrder = "desc"
