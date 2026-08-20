package audit

import (
	"fmt"
	"strings"

	"gateway/web/utils/session"

	"github.com/gin-gonic/gin"
)

// KickUserSessions 踢掉指定用户全部会话并写 KICK 审计。
func KickUserSessions(c *gin.Context, userId string) {
	session.InvalidateUserSessions(c, userId)
	writeKick(c, []string{userId}, "")
}

// KickUsersSessions 批量踢会话并写 KICK 审计。exceptUserId 不为空时跳过该用户。
func KickUsersSessions(c *gin.Context, userIds []string, exceptUserId string) {
	session.InvalidateUsersSessions(c, userIds, exceptUserId)
	writeKick(c, userIds, exceptUserId)
}

// writeKick 将会话失效当场记为独立 KICK 记录，不放入主事件上下文，以免覆盖业务 SetEvent。
func writeKick(c *gin.Context, userIds []string, exceptUserId string) {
	kicked := make([]string, 0, len(userIds))
	for _, userId := range userIds {
		if userId == "" || userId == exceptUserId {
			continue
		}
		kicked = append(kicked, userId)
	}
	if len(kicked) == 0 {
		return
	}
	targetId := kicked[0]
	targetName := ""
	detail := "session invalidated"
	if len(kicked) > 1 {
		targetId = "batch"
		targetName = fmt.Sprintf("%d users", len(kicked))
		joined := strings.Join(kicked, ",")
		if len(joined) > 1500 {
			joined = joined[:1500]
		}
		detail = "userIds=" + joined
	}
	writeFromGin(c, &AuditEvent{
		Action:     AuditActionKick,
		ModuleCode: ModuleCodeFromResourceOrPath("", requestPath(c)),
		TargetType: "USER",
		TargetId:   targetId,
		TargetName: targetName,
		Detail:     detail,
	})
}
