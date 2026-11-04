package disk

import "testing"

func TestSkipDiskIODevice(t *testing.T) {
	cases := []struct {
		device string
		skip   bool
	}{
		{"sda", false},
		{"sda1", true},
		{"nvme0n1", false},
		{"nvme0n1p1", true},
		{"mmcblk0", false},
		{"mmcblk0p1", true},
		{"dm-0", false},
		{"md0", false},
		{"loop0", true},
		{"/dev/loop1", true},
		{"ram0", true},
		{"zram0", true},
		{"C:", false},
	}
	for _, tc := range cases {
		if got := skipDiskIODevice(tc.device); got != tc.skip {
			t.Errorf("%s: got skip=%v, want %v", tc.device, got, tc.skip)
		}
	}
}
