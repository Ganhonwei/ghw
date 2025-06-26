package algo

import (
	"testing"
)

// 测试
func TestLhd(t *testing.T) {
	cs1 := []uint32{0x0a}
	cs2 := []uint32{0x0a}
	t.Log(Lhd(cs1, cs2))
}