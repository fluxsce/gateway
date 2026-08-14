package connection

import (
	"context"
	"testing"

	pb "gateway/internal/servicecenter/server/proto"

	"google.golang.org/grpc/metadata"
)

// recordConnectStream 记录服务端发出的消息，用于验证 SERVER_CLOSE 广播。
type recordConnectStream struct {
	ctx  context.Context
	sent []*pb.ServerMessage
}

func (s *recordConnectStream) Send(msg *pb.ServerMessage) error {
	s.sent = append(s.sent, msg)
	return nil
}

func (s *recordConnectStream) Recv() (*pb.ClientMessage, error) {
	<-s.ctx.Done()
	return nil, s.ctx.Err()
}

func (s *recordConnectStream) SetHeader(metadata.MD) error  { return nil }
func (s *recordConnectStream) SendHeader(metadata.MD) error { return nil }
func (s *recordConnectStream) SetTrailer(metadata.MD)       {}
func (s *recordConnectStream) Context() context.Context     { return s.ctx }
func (s *recordConnectStream) SendMsg(interface{}) error    { return nil }
func (s *recordConnectStream) RecvMsg(interface{}) error    { return nil }

// TestConnectionManager_Close_SendsServerClose 验证 ConnectionManager.Close 向连接广播 SERVER_CLOSE。
func TestConnectionManager_Close_SendsServerClose(t *testing.T) {
	mgr := NewConnectionManager()
	stream := &recordConnectStream{ctx: context.Background()}
	conn := mgr.CreateConnection(context.Background(), stream, "127.0.0.1")
	if conn == nil {
		t.Fatal("CreateConnection returned nil")
	}

	mgr.Close()

	found := false
	for _, msg := range stream.sent {
		if msg.GetMessageType() == pb.ServerMessageType_SERVER_CLOSE && msg.GetClose() != nil {
			found = true
			if msg.GetClose().GetReason() != "server_shutdown" {
				t.Fatalf("unexpected reason: %s", msg.GetClose().GetReason())
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected SERVER_CLOSE, sent=%d", len(stream.sent))
	}
}
