// 配置数据基础类型 - 严格按照HUB_SERVICE_CONFIG_DATA表结构定义
// 对应数据库表：HUB_SERVICE_CONFIG_DATA
export interface Config {
  // 主键和租户信息
  configDataId: string // 配置数据ID，主键
  tenantId: string // 租户ID，用于多租户数据隔离

  // 关联命名空间和分组
  namespaceId: string // 命名空间ID，关联HUB_SERVICE_NAMESPACE表
  groupName: string // 分组名称，如DEFAULT_GROUP

  // 配置基本信息
  configContent: string // 配置内容，支持大文本
  contentType: 'text' | 'json' | 'xml' | 'yaml' | 'properties' // 内容类型(text:文本,json:JSON,xml:XML,yaml:YAML,properties:Properties)
  configDescription: string // 配置描述
  encrypted: string // 是否加密存储(N否,Y是)

  // 版本信息
  version: number // 配置版本号（BIGINT），每次修改递增

  // MD5校验值（用于配置变更检测）
  md5Value: string // 配置内容的MD5值，用于快速比较配置是否变更

  // 通用字段（对应数据库 DATETIME/DATE 类型）
  addTime: string // 创建时间（DATETIME/DATE NOT NULL）
  addWho: string // 创建人ID
  editTime: string // 最后修改时间（DATETIME/DATE NOT NULL）
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: string // 活动状态标记(N非活动,Y活动)
  noteText: string // 备注信息
  extProperty: string // 扩展属性，JSON格式

  // 前端显示字段（可选，用于兼容）
  configVersion?: number // 配置版本号（前端显示用）
  contentMd5?: string // 配置内容的MD5值（前端显示用）
  configDesc?: string // 配置描述（前端显示用）
}

// 配置历史类型
// 对应数据库表：HUB_SERVICE_CONFIG_HISTORY
export interface ConfigHistory {
  // 主键和租户信息
  configHistoryId: string // 配置历史ID，主键
  tenantId: string // 租户ID，用于多租户数据隔离

  // 关联配置数据
  configDataId: string // 配置数据ID，关联HUB_SERVICE_CONFIG_DATA表
  namespaceId: string // 命名空间ID，冗余字段便于查询
  groupName: string // 分组名称，冗余字段便于查询

  // 变更信息
  changeType: 'CREATE' | 'UPDATE' | 'DELETE' | 'ROLLBACK' // 变更类型(CREATE:创建,UPDATE:更新,DELETE:删除,ROLLBACK:回滚)
  oldContent: string // 旧配置内容
  newContent: string // 新配置内容
  oldVersion: number // 旧版本号（BIGINT）
  newVersion: number // 新版本号（BIGINT）
  oldMd5Value: string // 旧配置MD5值
  newMd5Value: string // 新配置MD5值

  // 变更原因和操作人
  changeReason: string // 变更原因
  changedBy: string // 变更人ID
  changedAt: string // 变更时间（DATETIME/DATE NOT NULL）

  // 通用字段（对应数据库 DATETIME/DATE 类型）
  addTime: string // 创建时间（DATETIME/DATE NOT NULL）
  addWho: string // 创建人ID
  editTime: string // 最后修改时间（DATETIME/DATE NOT NULL）
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: string // 活动状态标记(N非活动,Y活动)
  noteText: string // 备注信息
  extProperty: string // 扩展属性，JSON格式

  // 前端显示字段（可选，用于兼容）
  contentType?: string // 内容类型
  configContent?: string // 配置内容（前端显示用，通常使用 newContent）
  contentMd5?: string // MD5值（前端显示用，通常使用 newMd5Value）
  configVersion?: number // 配置版本号（前端显示用，通常使用 newVersion）
  changeTime?: string // 变更时间（前端显示用，通常使用 changedAt）
}

// 配置查询条件
export interface ConfigQuery {
  namespaceId: string // 命名空间ID（必填）
  groupName?: string // 分组名称
  configDataId?: string // 配置数据ID（模糊查询）
  contentType?: string // 内容类型
  activeFlag?: 'Y' | 'N' // 活动状态
}

// 配置历史查询请求
export interface ConfigHistoryRequest {
  namespaceId: string // 命名空间ID（必填）
  groupName: string // 分组名称（必填）
  configDataId: string // 配置数据ID（必填）
  limit?: number // 限制数量，默认50
}

// 配置回滚请求
export interface RollbackRequest {
  configHistoryId: string // 配置历史ID（必填）
  changeReason?: string // 变更原因
}

