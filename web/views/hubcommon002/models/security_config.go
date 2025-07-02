package models

import (
	"time"
)

// SecurityConfig 安全配置模型，对应数据库HUB_GW_SECURITY_CONFIG表
type SecurityConfig struct {
	TenantId          string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID，联合主键
	SecurityConfigId  string     `json:"securityConfigId" form:"securityConfigId" query:"securityConfigId" db:"securityConfigId"`   // 安全配置ID，联合主键
	GatewayInstanceId *string    `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"` // 网关实例ID(实例级安全配置)
	RouteConfigId     *string    `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId" db:"routeConfigId"`                 // 路由配置ID(路由级安全配置)
	ConfigName        string     `json:"configName" form:"configName" query:"configName" db:"configName"`                           // 安全配置名称
	ConfigDesc        *string    `json:"configDesc" form:"configDesc" query:"configDesc" db:"configDesc"`                           // 安全配置描述
	ConfigPriority    int        `json:"configPriority" form:"configPriority" query:"configPriority" db:"configPriority"`           // 配置优先级，数值越小优先级越高
	CustomConfigJson  *string    `json:"customConfigJson" form:"customConfigJson" query:"customConfigJson" db:"customConfigJson"`   // 自定义配置参数，JSON格式
	Reserved1         *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                               // 预留字段1
	Reserved2         *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                               // 预留字段2
	Reserved3         *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                               // 预留字段3
	Reserved4         *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                               // 预留字段4
	Reserved5         *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                               // 预留字段5
	ExtProperty       *string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`                       // 扩展属性，JSON格式
	AddTime           time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                       // 创建时间
	AddWho            string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                           // 创建人ID
	EditTime          time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                   // 最后修改时间
	EditWho           string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                       // 最后修改人ID
	OprSeqFlag        string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                           // 操作序列标识
	CurrentVersion    int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`           // 当前版本号
	ActiveFlag        string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                           // 活动状态标记(N非活动,Y活动)
	NoteText          *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                   // 备注信息
}

// TableName 返回表名
func (SecurityConfig) TableName() string {
	return "HUB_GW_SECURITY_CONFIG"
}

// IpAccessConfig IP访问控制配置模型，对应数据库HUB_GW_IP_ACCESS_CONFIG表
type IpAccessConfig struct {
	TenantId           string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                           // 租户ID，联合主键
	IpAccessConfigId   string     `json:"ipAccessConfigId" form:"ipAccessConfigId" query:"ipAccessConfigId" db:"ipAccessConfigId"`         // IP访问配置ID，联合主键
	SecurityConfigId   string     `json:"securityConfigId" form:"securityConfigId" query:"securityConfigId" db:"securityConfigId"`         // 关联的安全配置ID
	ConfigName         string     `json:"configName" form:"configName" query:"configName" db:"configName"`                                 // IP访问配置名称
	DefaultPolicy      string     `json:"defaultPolicy" form:"defaultPolicy" query:"defaultPolicy" db:"defaultPolicy"`                     // 默认策略(allow允许,deny拒绝)
	WhitelistIps       *string    `json:"whitelistIps" form:"whitelistIps" query:"whitelistIps" db:"whitelistIps"`                         // IP白名单，JSON数组格式
	BlacklistIps       *string    `json:"blacklistIps" form:"blacklistIps" query:"blacklistIps" db:"blacklistIps"`                         // IP黑名单，JSON数组格式
	WhitelistCidrs     *string    `json:"whitelistCidrs" form:"whitelistCidrs" query:"whitelistCidrs" db:"whitelistCidrs"`                 // CIDR白名单，JSON数组格式
	BlacklistCidrs     *string    `json:"blacklistCidrs" form:"blacklistCidrs" query:"blacklistCidrs" db:"blacklistCidrs"`                 // CIDR黑名单，JSON数组格式
	TrustXForwardedFor string     `json:"trustXForwardedFor" form:"trustXForwardedFor" query:"trustXForwardedFor" db:"trustXForwardedFor"` // 是否信任X-Forwarded-For头(N否,Y是)
	TrustXRealIp       string     `json:"trustXRealIp" form:"trustXRealIp" query:"trustXRealIp" db:"trustXRealIp"`                         // 是否信任X-Real-IP头(N否,Y是)
	Reserved1          *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                                     // 预留字段1
	Reserved2          *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                                     // 预留字段2
	Reserved3          *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                                     // 预留字段3
	Reserved4          *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                                     // 预留字段4
	Reserved5          *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                                     // 预留字段5
	ExtProperty        *string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`                             // 扩展属性，JSON格式
	AddTime            time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                             // 创建时间
	AddWho             string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                                 // 创建人ID
	EditTime           time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                         // 最后修改时间
	EditWho            string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                             // 最后修改人ID
	OprSeqFlag         string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                                 // 操作序列标识
	CurrentVersion     int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`                 // 当前版本号
	ActiveFlag         string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                                 // 活动状态标记(N非活动,Y活动)
	NoteText           *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                         // 备注信息
}

// TableName 返回表名
func (IpAccessConfig) TableName() string {
	return "HUB_GW_IP_ACCESS_CONFIG"
}

// UseragentAccessConfig User-Agent访问控制配置模型，对应数据库HUB_GW_UA_ACCESS_CONFIG表
type UseragentAccessConfig struct {
	TenantId                 string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                                                 // 租户ID，联合主键
	UseragentAccessConfigId  string     `json:"useragentAccessConfigId" form:"useragentAccessConfigId" query:"useragentAccessConfigId" db:"useragentAccessConfigId"` // User-Agent访问配置ID，联合主键
	SecurityConfigId         string     `json:"securityConfigId" form:"securityConfigId" query:"securityConfigId" db:"securityConfigId"`                             // 关联的安全配置ID
	ConfigName               string     `json:"configName" form:"configName" query:"configName" db:"configName"`                                                     // User-Agent访问配置名称
	DefaultPolicy            string     `json:"defaultPolicy" form:"defaultPolicy" query:"defaultPolicy" db:"defaultPolicy"`                                         // 默认策略(allow允许,deny拒绝)
	WhitelistPatterns        *string    `json:"whitelistPatterns" form:"whitelistPatterns" query:"whitelistPatterns" db:"whitelistPatterns"`                       // User-Agent白名单模式，JSON数组格式，支持正则表达式
	BlacklistPatterns        *string    `json:"blacklistPatterns" form:"blacklistPatterns" query:"blacklistPatterns" db:"blacklistPatterns"`                       // User-Agent黑名单模式，JSON数组格式，支持正则表达式
	BlockEmptyUserAgent      string     `json:"blockEmptyUserAgent" form:"blockEmptyUserAgent" query:"blockEmptyUserAgent" db:"blockEmptyUserAgent"`               // 是否阻止空User-Agent(N否,Y是)
	Reserved1                *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                                                         // 预留字段1
	Reserved2                *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                                                         // 预留字段2
	Reserved3                *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                                                         // 预留字段3
	Reserved4                *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                                                         // 预留字段4
	Reserved5                *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                                                         // 预留字段5
	ExtProperty              *string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`                                                 // 扩展属性，JSON格式
	AddTime                  time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                                                 // 创建时间
	AddWho                   string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                                                     // 创建人ID
	EditTime                 time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                                             // 最后修改时间
	EditWho                  string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                                                 // 最后修改人ID
	OprSeqFlag               string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                                                     // 操作序列标识
	CurrentVersion           int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`                                     // 当前版本号
	ActiveFlag               string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                                                     // 活动状态标记(N非活动,Y活动)
	NoteText                 *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                                             // 备注信息
}

// TableName 返回表名
func (UseragentAccessConfig) TableName() string {
	return "HUB_GW_UA_ACCESS_CONFIG"
}

// ApiAccessConfig API访问控制配置模型，对应数据库HUB_GW_API_ACCESS_CONFIG表
type ApiAccessConfig struct {
	TenantId         string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID，联合主键
	ApiAccessConfigId string    `json:"apiAccessConfigId" form:"apiAccessConfigId" query:"apiAccessConfigId" db:"apiAccessConfigId"` // API访问配置ID，联合主键
	SecurityConfigId string     `json:"securityConfigId" form:"securityConfigId" query:"securityConfigId" db:"securityConfigId"`   // 关联的安全配置ID
	ConfigName       string     `json:"configName" form:"configName" query:"configName" db:"configName"`                           // API访问配置名称
	DefaultPolicy    string     `json:"defaultPolicy" form:"defaultPolicy" query:"defaultPolicy" db:"defaultPolicy"`               // 默认策略(allow允许,deny拒绝)
	WhitelistPaths   *string    `json:"whitelistPaths" form:"whitelistPaths" query:"whitelistPaths" db:"whitelistPaths"`           // API路径白名单，JSON数组格式，支持通配符
	BlacklistPaths   *string    `json:"blacklistPaths" form:"blacklistPaths" query:"blacklistPaths" db:"blacklistPaths"`           // API路径黑名单，JSON数组格式，支持通配符
	AllowedMethods   *string    `json:"allowedMethods" form:"allowedMethods" query:"allowedMethods" db:"allowedMethods"`           // 允许的HTTP方法，逗号分隔
	BlockedMethods   *string    `json:"blockedMethods" form:"blockedMethods" query:"blockedMethods" db:"blockedMethods"`           // 禁止的HTTP方法，逗号分隔
	Reserved1        *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                               // 预留字段1
	Reserved2        *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                               // 预留字段2
	Reserved3        *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                               // 预留字段3
	Reserved4        *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                               // 预留字段4
	Reserved5        *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                               // 预留字段5
	ExtProperty      *string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`                       // 扩展属性，JSON格式
	AddTime          time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                       // 创建时间
	AddWho           string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                           // 创建人ID
	EditTime         time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                   // 最后修改时间
	EditWho          string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                       // 最后修改人ID
	OprSeqFlag       string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                           // 操作序列标识
	CurrentVersion   int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`           // 当前版本号
	ActiveFlag       string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                           // 活动状态标记(N非活动,Y活动)
	NoteText         *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                   // 备注信息
}

// TableName 返回表名
func (ApiAccessConfig) TableName() string {
	return "HUB_GW_API_ACCESS_CONFIG"
}

// DomainAccessConfig 域名访问控制配置模型，对应数据库HUB_GW_DOMAIN_ACCESS_CONFIG表
type DomainAccessConfig struct {
	TenantId             string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                                     // 租户ID，联合主键
	DomainAccessConfigId string     `json:"domainAccessConfigId" form:"domainAccessConfigId" query:"domainAccessConfigId" db:"domainAccessConfigId"` // 域名访问配置ID，联合主键
	SecurityConfigId     string     `json:"securityConfigId" form:"securityConfigId" query:"securityConfigId" db:"securityConfigId"`                 // 关联的安全配置ID
	ConfigName           string     `json:"configName" form:"configName" query:"configName" db:"configName"`                                         // 域名访问配置名称
	DefaultPolicy        string     `json:"defaultPolicy" form:"defaultPolicy" query:"defaultPolicy" db:"defaultPolicy"`                             // 默认策略(allow允许,deny拒绝)
	WhitelistDomains     *string    `json:"whitelistDomains" form:"whitelistDomains" query:"whitelistDomains" db:"whitelistDomains"`                 // 域名白名单，JSON数组格式
	BlacklistDomains     *string    `json:"blacklistDomains" form:"blacklistDomains" query:"blacklistDomains" db:"blacklistDomains"`                 // 域名黑名单，JSON数组格式
	AllowSubdomains      string     `json:"allowSubdomains" form:"allowSubdomains" query:"allowSubdomains" db:"allowSubdomains"`                     // 是否允许子域名(N否,Y是)
	Reserved1            *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                                             // 预留字段1
	Reserved2            *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                                             // 预留字段2
	Reserved3            *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                                             // 预留字段3
	Reserved4            *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                                             // 预留字段4
	Reserved5            *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                                             // 预留字段5
	ExtProperty          *string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`                                     // 扩展属性，JSON格式
	AddTime              time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                                     // 创建时间
	AddWho               string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                                         // 创建人ID
	EditTime             time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                                 // 最后修改时间
	EditWho              string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                                     // 最后修改人ID
	OprSeqFlag           string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                                         // 操作序列标识
	CurrentVersion       int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`                         // 当前版本号
	ActiveFlag           string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                                         // 活动状态标记(N非活动,Y活动)
	NoteText             *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                                 // 备注信息
}

// TableName 返回表名
func (DomainAccessConfig) TableName() string {
	return "HUB_GW_DOMAIN_ACCESS_CONFIG"
} 