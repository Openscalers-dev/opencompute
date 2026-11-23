package hardware

import "testing"

func TestSystemDiskPath(t *testing.T) {
	for _, tc := range []struct {
		drive string
		path  string
	}{
		{drive: "", path: "C:\\"},
		{drive: "D:", path: "D:\\"},
	} {
		t.Setenv("SystemDrive", tc.drive)
		if got := SystemDiskPath(); got != tc.path {
			t.Errorf("SystemDiskPath() with SystemDrive %q = %q, want %q", tc.drive, got, tc.path)
		}
	}
}
