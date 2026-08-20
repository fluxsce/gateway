package session

import (
	"context"

	"gateway/pkg/logger"
)

// InvalidateUserSessions 在权限或身份变更后强制指定用户重新登录。
// 删除失败只记日志，不回滚已成功的业务写入。
func InvalidateUserSessions(ctx context.Context, userId string) {
	if userId == "" {
		return
	}
	sm := GetGlobalSessionManager()
	if sm == nil {
		return
	}
	if err := sm.DeleteUserSessions(ctx, userId); err != nil {
		logger.ErrorWithTrace(ctx, "权限变更后清理用户session失败", "error", err, "userId", userId)
	}
}

// InvalidateUsersSessions 批量强制用户重新登录。
// exceptUserId 不为空时跳过该用户，避免管理员保存角色授权时把自己踢下线。
func InvalidateUsersSessions(ctx context.Context, userIds []string, exceptUserId string) {
	for _, userId := range userIds {
		if userId == "" || userId == exceptUserId {
			continue
		}
		InvalidateUserSessions(ctx, userId)
	}
}
