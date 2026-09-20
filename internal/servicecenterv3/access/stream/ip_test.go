package stream

import "testing"

func TestMatchIPListExactAndCIDR(t *testing.T) {
	if !matchIPList("10.0.0.8", "10.0.0.8,192.168.1.1") {
		t.Fatal("exact IP should match")
	}
	if matchIPList("10.0.0.9", "10.0.0.8") {
		t.Fatal("other IP should not match")
	}
	if !matchIPList("10.1.2.3", `["10.1.0.0/16"]`) {
		t.Fatal("CIDR should match")
	}
	if matchIPList("11.0.0.1", "10.0.0.0/8") {
		t.Fatal("outside CIDR should not match")
	}
}
