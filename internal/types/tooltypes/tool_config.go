package tooltypes

import (
	"encoding/json"
	"errors"
	"time"
)

// ToolConfig 定义工具配置信息，对应数据库表 HUB_TOOL_CONFIG
type ToolConfig struct {
	// 主键信息
	ToolConfigId      string    `json:"toolConfigId" db:"toolConfigId"`
	TenantId          string    `json:"tenantId" db:"tenantId"`
	
	// 工具基础信息
	ToolName          string    `json:"toolName" db:"toolName"`
	ToolType          string    `json:"toolType" db:"toolType"`
	ToolVersion       *string   `json:"toolVersion" db:"toolVersion"`
	ConfigName        string    `json:"configName" db:"configName"`
	ConfigDescription *string   `json:"configDescription" db:"configDescription"`
	
	// 分组信息
	ConfigGroupId     *string   `json:"configGroupId" db:"configGroupId"`
	ConfigGroupName   *string   `json:"configGroupName" db:"configGroupName"`
	
	// 连接配置
	HostAddress       *string   `json:"hostAddress" db:"hostAddress"`
	PortNumber        *int      `json:"portNumber" db:"portNumber"`
	ProtocolType      *string   `json:"protocolType" db:"protocolType"`
	
	// 认证配置
	AuthType          *string   `json:"authType" db:"authType"`
	UserName          *string   `json:"userName" db:"userName"`
	PasswordEncrypted *string   `json:"passwordEncrypted" db:"passwordEncrypted"`
	KeyFilePath       *string   `json:"keyFilePath" db:"keyFilePath"`
	KeyFileContent    *string   `json:"keyFileContent" db:"keyFileContent"`
	
	// 配置参数
	ConfigParameters   *string   `json:"configParameters" db:"configParameters"`
	EnvironmentVariables *string `json:"environmentVariables" db:"environmentVariables"`
	CustomSettings     *string   `json:"customSettings" db:"customSettings"`
	
	// 状态和控制
	ConfigStatus      string    `json:"configStatus" db:"configStatus"`
	DefaultFlag       string    `json:"defaultFlag" db:"defaultFlag"`
	PriorityLevel     *int      `json:"priorityLevel" db:"priorityLevel"`
	
	// 安全和加密
	EncryptionType    *string   `json:"encryptionType" db:"encryptionType"`
	EncryptionKey     *string   `json:"encryptionKey" db:"encryptionKey"`
	
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
func (ToolConfig) TableName() string {
	return "HUB_TOOL_CONFIG"
}

// Validate 验证工具配置是否有效
func (t *ToolConfig) Validate() error {
	// 验证必填字段
	if t.ToolName == "" {
		return errors.New("工具名称不能为空")
	}
	
	if t.ToolType == "" {
		return errors.New("工具类型不能为空")
	}
	
	if t.ConfigName == "" {
		return errors.New("配置名称不能为空")
	}
	
	// 验证连接信息
	if t.HostAddress != nil && *t.HostAddress != "" {
		if t.PortNumber == nil {
			return errors.New("指定主机地址时，端口号不能为空")
		}
	}
	
	// 验证认证信息
	if t.AuthType != nil {
		switch *t.AuthType {
		case AuthTypePassword:
			if t.UserName == nil || *t.UserName == "" || t.PasswordEncrypted == nil || *t.PasswordEncrypted == "" {
				return errors.New("密码认证方式下，用户名和密码不能为空")
			}
		case AuthTypePublicKey:
			if (t.KeyFilePath == nil || *t.KeyFilePath == "") && (t.KeyFileContent == nil || *t.KeyFileContent == "") {
				return errors.New("公钥认证方式下，密钥文件路径或内容不能同时为空")
			}
		}
	}
	
	return nil
}

// IsActive 检查配置是否处于活动状态
func (t *ToolConfig) IsActive() bool {
	return t.ActiveFlag == ActiveFlagYes
}

// IsEnabled 检查配置是否启用
func (t *ToolConfig) IsEnabled() bool {
	return t.ConfigStatus == ConfigStatusEnabled
}

// IsDefault 检查是否为默认配置
func (t *ToolConfig) IsDefault() bool {
	return t.DefaultFlag == DefaultFlagYes
}

// GetConfigParameters 获取配置参数的JSON对象
func (t *ToolConfig) GetConfigParameters() (map[string]interface{}, error) {
	if t.ConfigParameters == nil || *t.ConfigParameters == "" {
		return make(map[string]interface{}), nil
	}
	
	var result map[string]interface{}
	err := json.Unmarshal([]byte(*t.ConfigParameters), &result)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// GetEnvironmentVariables 获取环境变量的JSON对象
func (t *ToolConfig) GetEnvironmentVariables() (map[string]string, error) {
	if t.EnvironmentVariables == nil || *t.EnvironmentVariables == "" {
		return make(map[string]string), nil
	}
	
	var result map[string]string
	err := json.Unmarshal([]byte(*t.EnvironmentVariables), &result)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// GetCustomSettings 获取自定义设置的JSON对象
func (t *ToolConfig) GetCustomSettings() (map[string]interface{}, error) {
	if t.CustomSettings == nil || *t.CustomSettings == "" {
		return make(map[string]interface{}), nil
	}
	
	var result map[string]interface{}
	err := json.Unmarshal([]byte(*t.CustomSettings), &result)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// GetExtProperty 获取扩展属性的JSON对象
func (t *ToolConfig) GetExtProperty() (map[string]interface{}, error) {
	if t.ExtProperty == nil || *t.ExtProperty == "" {
		return make(map[string]interface{}), nil
	}
	
	var result map[string]interface{}
	err := json.Unmarshal([]byte(*t.ExtProperty), &result)
	if err != nil {
		return nil, err
	}
	
	return result, nil
} 