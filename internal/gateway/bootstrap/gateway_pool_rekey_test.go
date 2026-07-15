package bootstrap

import "testing"

func TestGatewayPoolRekeyDoesNotStopRunningGateway(t *testing.T) {
	pool := newGatewayPool()
	gateway := &Gateway{running: true}
	if err := pool.Add("old-id", gateway); err != nil {
		t.Fatal(err)
	}

	if err := pool.rekey("old-id", "new-id", gateway); err != nil {
		t.Fatal(err)
	}
	if !gateway.IsRunning() {
		t.Fatal("rekey stopped the running gateway")
	}
	if pool.Exists("old-id") || !pool.Exists("new-id") {
		t.Fatal("rekey did not atomically replace the pool key")
	}
}
