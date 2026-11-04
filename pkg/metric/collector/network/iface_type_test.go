package network

import "testing"

func TestInterfaceType(t *testing.T) {
	collector := NewNetworkCollector()
	cases := map[string]string{
		"lo":                          "loopback",
		"lo0":                         "loopback",
		"Loopback Pseudo-Interface 1": "loopback",
		"Local Area Connection":       "ethernet",
		"Ethernet":                    "ethernet",
		"以太网":                         "ethernet",
		"eth0":                        "ethernet",
		"ens33":                       "ethernet",
		"Wi-Fi":                       "wifi",
		"wlan0":                       "wifi",
		"wlp2s0":                      "wifi",
		"docker0":                     "bridge",
		"tun0":                        "vpn",
	}
	for name, want := range cases {
		if got := collector.getInterfaceType(name); got != want {
			t.Errorf("%s: got %s, want %s", name, got, want)
		}
	}
}
