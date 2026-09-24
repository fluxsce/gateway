package types

import "time"

// ClusterNode 集群节点。
// 对应数据库表 HUB_CLUSTER_NODE。进程用与采集表相同的节点 ID 登记自身。
type ClusterNode struct {
	NodeId       string    `json:"nodeId" db:"nodeId"`
	TenantId     string    `json:"tenantId" db:"tenantId"`
	NodeIp       string    `json:"nodeIp" db:"nodeIp"`
	Hostname     string    `json:"hostname" db:"hostname"`
	StartedTime  time.Time `json:"startedTime" db:"startedTime"`
	LastSeenTime time.Time `json:"lastSeenTime" db:"lastSeenTime"`

	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       string    `json:"noteText" db:"noteText"`
	ExtProperty    string    `json:"extProperty" db:"extProperty"`
	Reserved1      string    `json:"reserved1" db:"reserved1"`
	Reserved2      string    `json:"reserved2" db:"reserved2"`
	Reserved3      string    `json:"reserved3" db:"reserved3"`
	Reserved4      string    `json:"reserved4" db:"reserved4"`
	Reserved5      string    `json:"reserved5" db:"reserved5"`
}
