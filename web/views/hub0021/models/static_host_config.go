package models

import (
	"time"
)

// StaticHostConfig 本机目录托管配置模型，对应 HUB_GW_STATIC_HOST_CONFIG。
// 通过 routeConfigId 关联路由；命中后数据面从 rootDirectory 出文件，不再转发上游。
// 两层路径：路由过滤器改 Request.URL；本配置只做文件映射（剥前缀 -> rewriteRules -> 文件/索引/SPA）。
// 路由 stripPathPrefix 为 Y 时，静态查找也会剥前缀；rewritePath 仍只作用于反向代理。
type StaticHostConfig struct {
	TenantId               string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                                         // 租户ID，联合主键
	StaticHostConfigId     string `json:"staticHostConfigId" form:"staticHostConfigId" query:"staticHostConfigId" db:"staticHostConfigId"`                 // 静态托管配置ID，联合主键
	RouteConfigId          string `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId" db:"routeConfigId"`                                     // 关联的路由配置ID，一条路由对应一条活动配置
	ConfigName             string `json:"configName" form:"configName" query:"configName" db:"configName"`                                                 // 配置名称，便于日志与管理端识别
	RootDirectory          string `json:"rootDirectory" form:"rootDirectory" query:"rootDirectory" db:"rootDirectory"`                                     // 本机根目录；可用 {v1,v2} 允许名单，从剥前缀后第一段展开
	StripRoutePrefix       string `json:"stripRoutePrefix" form:"stripRoutePrefix" query:"stripRoutePrefix" db:"stripRoutePrefix"`                         // 是否按路径段剥离已匹配路由前缀(N否,Y是)，默认Y
	IndexFiles             string `json:"indexFiles" form:"indexFiles" query:"indexFiles" db:"indexFiles"`                                                 // 目录索引文件，入库为JSON数组，如["index.html"]
	RewriteRules           string `json:"rewriteRules" form:"rewriteRules" query:"rewriteRules" db:"rewriteRules"`                                         // 路径重写规则，入库为JSON数组[{mode,from,to}]；mode=prefix按段前缀、exact按字面字符串、regex正则，按顺序命中第一条
	SpaFallback            string `json:"spaFallback" form:"spaFallback" query:"spaFallback" db:"spaFallback"`                                             // 无扩展名的缺失路径是否回退到索引文件(N否,Y是)，供history路由
	CacheControlMaxAge     int    `json:"cacheControlMaxAge" form:"cacheControlMaxAge" query:"cacheControlMaxAge" db:"cacheControlMaxAge"`                 // 普通文件Cache-Control max-age(秒)；索引与SPA回退固定no-cache；0表示普通文件也不缓存
	AllowedExtensions      string `json:"allowedExtensions" form:"allowedExtensions" query:"allowedExtensions" db:"allowedExtensions"`                     // 允许的文件扩展名，入库为JSON数组；空表示不额外限制
	MaxFileSizeBytes       int64  `json:"maxFileSizeBytes" form:"maxFileSizeBytes" query:"maxFileSizeBytes" db:"maxFileSizeBytes"`                         // 单文件大小上限(字节)，0表示不限制
	FollowSymlinks         string `json:"followSymlinks" form:"followSymlinks" query:"followSymlinks" db:"followSymlinks"`                                 // 是否跟随符号链接(N否,Y是)，默认N，开启后仍须落在根目录内
	EnablePrecompress      string `json:"enablePrecompress" form:"enablePrecompress" query:"enablePrecompress" db:"enablePrecompress"`                     // 是否直接出预压缩.br/.gz(N否,Y是)，默认Y
	RedirectDirectorySlash string `json:"redirectDirectorySlash" form:"redirectDirectorySlash" query:"redirectDirectorySlash" db:"redirectDirectorySlash"` // 目录缺斜杠是否301到path/(N否,Y是)，默认N
	RootTokenExact         string `json:"rootTokenExact" form:"rootTokenExact" query:"rootTokenExact" db:"rootTokenExact"`                                 // 占位符是否只精确匹配第一段(N否前缀命中,Y是)，默认N
	FallbackRoots          string `json:"fallbackRoots" form:"fallbackRoots" query:"fallbackRoots" db:"fallbackRoots"`                                     // 备用查找目录，一行一个，最多3个；本根找不到再用同一相对路径
	CacheControlByExt      string `json:"cacheControlByExt" form:"cacheControlByExt" query:"cacheControlByExt" db:"cacheControlByExt"`                     // 按扩展名覆盖缓存秒数，如 .js=86400，HTML仍不缓存
	EnableGzip             string `json:"enableGzip" form:"enableGzip" query:"enableGzip" db:"enableGzip"`                                                 // 无预压缩文件时是否现场gzip(N否,Y是)，默认N
	SecurityHeaders        string `json:"securityHeaders" form:"securityHeaders" query:"securityHeaders" db:"securityHeaders"`                             // 页面安全头，一行一个「名: 值」，仅白名单
	ErrorPage404           string `json:"errorPage404" form:"errorPage404" query:"errorPage404" db:"errorPage404"`                                         // 根目录内自定义404页面，如 /404.html
	ErrorPage403           string `json:"errorPage403" form:"errorPage403" query:"errorPage403" db:"errorPage403"`                                         // 根目录内自定义403页面，如 /403.html
	ConfigPriority         int    `json:"configPriority" form:"configPriority" query:"configPriority" db:"configPriority"`                                 // 同路由多行时的优先级，数值越小越优先

	// 预留字段
	Reserved1 string     `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"` // 预留字段1
	Reserved2 string     `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"` // 预留字段2
	Reserved3 *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"` // 预留字段3
	Reserved4 *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"` // 预留字段4
	Reserved5 *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"` // 预留字段5

	// 扩展属性
	ExtProperty string `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"` // 扩展属性,JSON格式

	// 标准字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号，乐观锁
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动/禁用,Y活动/启用)
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// TableName 返回表名。
func (StaticHostConfig) TableName() string {
	return "HUB_GW_STATIC_HOST_CONFIG"
}
