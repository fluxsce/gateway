package stream

import "gateway/internal/servicecenterv3/model"

// Session 是数据面连接登记，由 center.SessionHub 实现。
type Session interface {
	// Add 登记或覆盖一条连接。
	Add(info *model.ConnectionInfo)
	// Touch 刷新最近活跃时间。
	Touch(connectionID string)
	// Remove 删除连接记录。
	Remove(connectionID string)
}
