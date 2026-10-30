package proxy

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"gateway/internal/gateway/constants"
)

func TestProxyCancelStatus(t *testing.T) {
	alive, cancelAlive := context.WithCancel(context.Background())
	defer cancelAlive()

	callerDone, cancelCaller := context.WithCancel(context.Background())
	cancelCaller()

	proxyDone, cancelProxy := context.WithCancel(context.Background())
	cancelProxy()

	proxyDeadline, cancelDeadline := context.WithTimeout(context.Background(), 0)
	defer cancelDeadline()
	<-proxyDeadline.Done()

	postErr := &url.Error{Op: "Post", URL: "http://upstream/datahubjson/", Err: context.Canceled}

	tests := []struct {
		name     string
		incoming context.Context
		outbound context.Context
		err      error
		want     int
	}{
		{name: "caller cancel", incoming: callerDone, outbound: proxyDone, err: postErr, want: constants.GatewayStatusClientClosed},
		{name: "proxy timer", incoming: alive, outbound: proxyDone, err: postErr, want: constants.GatewayStatusGatewayTimeout},
		{name: "proxy deadline", incoming: alive, outbound: proxyDeadline, err: context.DeadlineExceeded, want: constants.GatewayStatusGatewayTimeout},
		{name: "other error", incoming: callerDone, outbound: proxyDone, err: errors.New("connection refused"), want: 0},
		{name: "canceled without either context", incoming: alive, outbound: alive, err: context.Canceled, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proxyCancelStatus(tt.incoming, tt.outbound, tt.err); got != tt.want {
				t.Fatalf("proxyCancelStatus() = %d, want %d", got, tt.want)
			}
		})
	}
	if proxyCancelStatus(callerDone, proxyDone, nil) != 0 {
		t.Fatal("nil error should stay unclassified")
	}
}

func TestProxyFailureSummary(t *testing.T) {
	if proxyFailureSummary(constants.GatewayStatusClientClosed) != "client closed request" {
		t.Fatal("499 summary")
	}
	if proxyFailureSummary(constants.GatewayStatusGatewayTimeout) != "gateway timeout" {
		t.Fatal("504 summary")
	}
	if proxyFailureSummary(constants.GatewayStatusBadGateway) != "proxy request failed" {
		t.Fatal("502 summary")
	}
}
