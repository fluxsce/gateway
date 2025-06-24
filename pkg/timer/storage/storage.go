// Package storage 提供任务存储功能
package storage

import "gohub/pkg/timer"

// Storage 任务存储接口的别名
// 使用别名避免循环依赖
type Storage = timer.TaskStorage 