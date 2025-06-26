package tooltypes

import (
	"encoding/json"
	"errors"
	"time"
)

// ToolConfigGroup 定义工具配置分组信息，对应数据库表 HUB_TOOL_CONFIG_GROUP
type ToolConfigGroup struct {
	// 主键信息
	ConfigGroupId     string    `json:"configGroupId" db:"configGroupId"`
	TenantId          string    `json:"tenantId" db:"tenantId"`
	
	// 分组信息
	GroupName         string    `json:"groupName" db:"groupName"`
	GroupDescription  *string   `json:"groupDescription" db:"groupDescription"`
	ParentGroupId     *string   `json:"parentGroupId" db:"parentGroupId"`
	GroupLevel        *int      `json:"groupLevel" db:"groupLevel"`
	GroupPath         *string   `json:"groupPath" db:"groupPath"`
	
	// 分组属性
	GroupType         *string   `json:"groupType" db:"groupType"`
	SortOrder         *int      `json:"sortOrder" db:"sortOrder"`
	GroupIcon         *string   `json:"groupIcon" db:"groupIcon"`
	GroupColor        *string   `json:"groupColor" db:"groupColor"`
	
	// 权限控制
	AccessLevel       *string   `json:"accessLevel" db:"accessLevel"`
	AllowedUsers      *string   `json:"allowedUsers" db:"allowedUsers"`
	AllowedRoles      *string   `json:"allowedRoles" db:"allowedRoles"`
	
	// 通用字段
	AddTime           time.Time `json:"addTime" db:"addTime"`
	AddWho            string    `json:"addWho" db:"addWho"`
	EditTime          time.Time `json:"editTime" db:"editTime"`
	EditWho           string    `json:"editWho" db:"editWho"`
	OprSeqFlag        string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion    int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag        string    `json:"activeFlag" db:"activeFlag"`
	NoteText          *string   `json:"noteText" db:"noteText"`
	ExtProperty       *string   `json:"extProperty" db:"extProperty"`
	Reserved1         *string   `json:"reserved1" db:"reserved1"`
	Reserved2         *string   `json:"reserved2" db:"reserved2"`
	Reserved3         *string   `json:"reserved3" db:"reserved3"`
	Reserved4         *string   `json:"reserved4" db:"reserved4"`
	Reserved5         *string   `json:"reserved5" db:"reserved5"`
	Reserved6         *string   `json:"reserved6" db:"reserved6"`
	Reserved7         *string   `json:"reserved7" db:"reserved7"`
	Reserved8         *string   `json:"reserved8" db:"reserved8"`
	Reserved9         *string   `json:"reserved9" db:"reserved9"`
	Reserved10        *string   `json:"reserved10" db:"reserved10"`
}

// TableName 返回数据库表名
func (ToolConfigGroup) TableName() string {
	return "HUB_TOOL_CONFIG_GROUP"
}

// Validate 验证配置分组是否有效
func (g *ToolConfigGroup) Validate() error {
	// 验证必填字段
	if g.GroupName == "" {
		return errors.New("分组名称不能为空")
	}
	
	// 验证父分组关系
	if g.ParentGroupId != nil && *g.ParentGroupId != "" {
		if g.GroupLevel == nil || *g.GroupLevel <= 1 {
			return errors.New("指定父分组时，分组层级必须大于1")
		}
	} else {
		// 根分组层级应为1
		if g.GroupLevel != nil && *g.GroupLevel != 1 {
			return errors.New("根分组的层级必须为1")
		}
	}
	
	return nil
}

// IsActive 检查分组是否处于活动状态
func (g *ToolConfigGroup) IsActive() bool {
	return g.ActiveFlag == ActiveFlagYes
}

// GetAllowedUsers 获取允许访问的用户列表
func (g *ToolConfigGroup) GetAllowedUsers() ([]string, error) {
	if g.AllowedUsers == nil || *g.AllowedUsers == "" {
		return []string{}, nil
	}
	
	var result []string
	err := json.Unmarshal([]byte(*g.AllowedUsers), &result)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// GetAllowedRoles 获取允许访问的角色列表
func (g *ToolConfigGroup) GetAllowedRoles() ([]string, error) {
	if g.AllowedRoles == nil || *g.AllowedRoles == "" {
		return []string{}, nil
	}
	
	var result []string
	err := json.Unmarshal([]byte(*g.AllowedRoles), &result)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// GetExtProperty 获取扩展属性的JSON对象
func (g *ToolConfigGroup) GetExtProperty() (map[string]interface{}, error) {
	if g.ExtProperty == nil || *g.ExtProperty == "" {
		return make(map[string]interface{}), nil
	}
	
	var result map[string]interface{}
	err := json.Unmarshal([]byte(*g.ExtProperty), &result)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// IsPublic 检查分组是否为公开访问
func (g *ToolConfigGroup) IsPublic() bool {
	return g.AccessLevel != nil && *g.AccessLevel == AccessLevelPublic
}

// IsPrivate 检查分组是否为私有访问
func (g *ToolConfigGroup) IsPrivate() bool {
	return g.AccessLevel == nil || *g.AccessLevel == AccessLevelPrivate
}

// IsRestricted 检查分组是否为受限访问
func (g *ToolConfigGroup) IsRestricted() bool {
	return g.AccessLevel != nil && *g.AccessLevel == AccessLevelRestricted
} 