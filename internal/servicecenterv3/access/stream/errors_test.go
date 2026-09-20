package stream

import (
	"testing"

	"gateway/internal/servicecenterv3/contract"
)

func TestErrorCode(t *testing.T) {
	code, _ := errorCode(errHandshakeRequired)
	if code != "PROTOCOL_HANDSHAKE_REQUIRED" {
		t.Fatalf("got %s", code)
	}
	code, _ = errorCode(contract.ErrConfigNotFound)
	if code != "CONFIG_NOT_FOUND" {
		t.Fatalf("got %s", code)
	}
}
