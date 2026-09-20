package stream

import (
	"context"

	"gateway/internal/servicecenterv3/model"
	pb "gateway/internal/servicecenterv3/proto"
)

// getConfig 返回已发布配置，不含草稿。
func (h *Handler) getConfig(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetGetConfig()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	rel, err := h.config.GetPublished(ctx, cc, in.GetGroupName(), in.GetConfigDataId())
	if err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_GET_CONFIG,
		Message:     &pb.ServerMessage_GetConfig{GetConfig: &pb.GetConfigResponse{Success: true, Config: releaseToPB(rel)}},
	})
}

// getDraft 读取未发布草稿。
func (h *Handler) getDraft(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetGetDraft()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	draft, err := h.config.GetDraft(ctx, cc, in.GetGroupName(), in.GetConfigDataId())
	if err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_GET_DRAFT,
		Message:     &pb.ServerMessage_GetDraft{GetDraft: &pb.GetConfigResponse{Success: true, Config: draftToPB(draft)}},
	})
}

// saveDraft 只写库中草稿，不改已发布、不推 SDK。
func (h *Handler) saveDraft(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetSaveConfig()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	if err := h.config.SaveDraft(ctx, cc, draftFromPB(cc.NamespaceID, in)); err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_SAVE_CONFIG,
		Message:     &pb.ServerMessage_SaveConfig{SaveConfig: &pb.SaveConfigResponse{Success: true}},
	})
}

// publish 将当前草稿发布为新版本并推送。
func (h *Handler) publish(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetPublishConfig()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	rel, err := h.config.Publish(ctx, cc, in.GetGroupName(), in.GetConfigDataId(), in.GetChangeReason())
	if err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_PUBLISH_CONFIG,
		Message:     &pb.ServerMessage_PublishConfig{PublishConfig: &pb.SaveConfigResponse{Success: true, Version: rel.Version, ContentMd5: rel.MD5}},
	})
}

// deleteConfig 删除已发布配置及草稿。
func (h *Handler) deleteConfig(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetDeleteConfig()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	if err := h.config.Delete(ctx, cc, in.GetGroupName(), in.GetConfigDataId()); err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_DELETE_CONFIG,
		Message:     &pb.ServerMessage_DeleteConfig{DeleteConfig: &pb.ConfigResponse{Success: true}},
	})
}

// listConfigs 列出已发布配置。
func (h *Handler) listConfigs(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetListConfigs()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	list, err := h.config.ListPublished(ctx, cc, in.GetGroupName())
	if err != nil {
		return err
	}
	all := make([]*pb.ConfigData, 0, len(list))
	for _, rel := range list {
		all = append(all, releaseToPB(rel))
	}
	out, next, total := pageOf(all, in.GetPageSize(), in.GetPageToken())
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_LIST_CONFIGS,
		Message:     &pb.ServerMessage_ListConfigs{ListConfigs: &pb.ListConfigsResponse{Success: true, Configs: out, NextPageToken: next, Total: total}},
	})
}

// watchConfig 订阅已发布变更，绑定连接并回 ACK。
func (h *Handler) watchConfig(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetWatchConfig()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	if err := h.config.Watch(ctx, cc, in.GetGroupName(), in.GetConfigDataIds(), st.configCh); err != nil {
		return err
	}
	h.config.BindClient(st.configCh, st.id)
	return h.sendAck(st, stream, msg.GetRequestId())
}

// unwatch 取消本连接全部 Watch 并回 ACK。
func (h *Handler) unwatch(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	cc := h.callContext(ctx, "", st)
	h.config.Unwatch(ctx, cc, st.configCh)
	return h.sendAck(st, stream, msg.GetRequestId())
}

// unwatchConfigs 取消本连接上指定配置的 Watch 并回 ACK。
func (h *Handler) unwatchConfigs(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetUnwatchConfigs()
	ns, group := "", ""
	var ids []string
	if in != nil {
		ns, group, ids = in.GetNamespaceId(), in.GetGroupName(), in.GetConfigDataIds()
	}
	cc := h.callContext(ctx, ns, st)
	h.config.UnwatchConfigs(ctx, cc, group, ids, st.configCh)
	return h.sendAck(st, stream, msg.GetRequestId())
}

// configHistory 倒序返回发布历史。
func (h *Handler) configHistory(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetGetConfigHistory()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	limit := int(in.GetPageSize())
	if limit <= 0 {
		limit = int(in.GetLimit())
	}
	list, err := h.config.ListReleases(ctx, cc, in.GetGroupName(), in.GetConfigDataId(), limit)
	if err != nil {
		return err
	}
	all := make([]*pb.ConfigHistory, 0, len(list))
	for _, rel := range list {
		all = append(all, &pb.ConfigHistory{
			NamespaceId:   rel.NamespaceID,
			GroupName:     rel.GroupName,
			ConfigDataId:  rel.DataID,
			ConfigContent: rel.Content,
			ContentMd5:    rel.MD5,
			ConfigVersion: rel.Version,
			ChangeType:    configChangeTypeToPB("UPDATE"),
			ChangeReason:  rel.Reason,
			ChangedBy:     rel.PublishedBy,
			ChangeTime:    rel.PublishedAt.UnixMilli(),
		})
	}
	out, next, total := pageOf(all, in.GetPageSize(), in.GetPageToken())
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_GET_CONFIG_HISTORY,
		Message:     &pb.ServerMessage_GetConfigHistory{GetConfigHistory: &pb.GetConfigHistoryResponse{Success: true, History: out, NextPageToken: next, Total: total}},
	})
}

// rollback 将目标版本重新发布为最新版本。
func (h *Handler) rollback(ctx context.Context, stream pb.ServiceCenterStream_ConnectServer, st *connState, msg *pb.ClientMessage) error {
	in := msg.GetRollbackConfig()
	cc := h.callContext(ctx, in.GetNamespaceId(), st)
	rel, err := h.config.Rollback(ctx, cc, in.GetGroupName(), in.GetConfigDataId(), in.GetTargetVersion(), in.GetChangeReason())
	if err != nil {
		return err
	}
	return h.send(st, stream, &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_ROLLBACK_CONFIG,
		Message:     &pb.ServerMessage_RollbackConfig{RollbackConfig: &pb.RollbackConfigResponse{Success: true, NewVersion: rel.Version, ContentMd5: rel.MD5}},
	})
}

func draftToPB(d *model.ConfigDraft) *pb.ConfigData {
	if d == nil {
		return nil
	}
	return &pb.ConfigData{
		NamespaceId:   d.NamespaceID,
		GroupName:     d.GroupName,
		ConfigDataId:  d.DataID,
		ContentType:   d.ContentType,
		ConfigContent: d.Content,
		ConfigDesc:    d.Description,
	}
}
