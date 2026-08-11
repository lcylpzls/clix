package core

import "testing"

// TestFlagDisplayUint64 覆盖 Uint64 flag 的帮助显示。
func TestFlagDisplayUint64(t *testing.T) {
	got := flagDisplay(Uint64Flag("dc", "数据中心 ID"))
	if got != "--dc uint64" {
		t.Fatalf("显示不符：%s", got)
	}
}
