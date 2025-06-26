package moneycoming

import (
	"fmt"
	"goserver/pkg/table"
	"os"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestGetElementsByConfig(t *testing.T) {
	err := os.Chdir("/root/work/goserver/run")
	if err != nil {
		t.Fatal(err)
	}

	//加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	elements := GetElementsByConfig(4, 5)
	fmt.Println(elements)
}
func SymbolToString(s Symbol) string {
	switch s {
	case Empty:
		return "空"
	case TWOX:
		return "2x"
	case FIVEX:
		return "5x"
	case TENX:
		return "10x"
	case RESPIN:
		return "RES"
	case SCATTER1:
		return "SCATTER1"
	case SCATTER2:
		return "SCATTER2"
	case TEN:
		return "10"
	case ZEROZERO:
		return "00"
	case FIVE:
		return "5"
	case ONE:
		return "1"
	case ZERO:
		return "0"
	case SCATTER3:
		return "SCATTER3"
	case SCATTER4:
		return "SCATTER4"
	default:
		return "?"
	}
}

func pad(s string, width int) string {
	w := runewidth.StringWidth(s)
	if w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s
}

func PrintReelsWithMultiplier(reels map[int32]Symbol) {
	fmt.Println("转轴结果: ")
	fmt.Println("+--------+--------+--------+--------+")
	fmt.Printf("|%s|%s|%s|%s|\n",
		pad(" 轴1", 8), pad(" 轴2", 8), pad(" 轴3", 8), pad(" 倍数轴", 8))
	fmt.Println("+--------+--------+--------+--------+")
	for row := 0; row < 3; row++ {
		fmt.Printf("|%s|%s|%s|%s|\n",
			pad(SymbolToString(reels[int32(row+1)]), 8),
			pad(SymbolToString(reels[int32(row+4)]), 8),
			pad(SymbolToString(reels[int32(row+7)]), 8),
			pad(SymbolToString(reels[int32(row+10)]), 8),
		)
	}
	fmt.Println("+--------+--------+--------+--------+")
}

func TestFortuneGems2(t *testing.T) {

	err := os.Chdir("/root/work/goserver/run")
	if err != nil {
		t.Fatal(err)
	}

	//加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	bet := int64(10000) // 假设押注100分（1元）
	score, reels, specificSymbol, winLines, winMultiplier, normalWin, wheelMultiplier, extraMultipliers, wheelWin, _ := PlaySlot(bet, map[int32]Symbol{}, "")
	PrintReelsWithMultiplier(reels)
	fmt.Printf("押注：%d\n", bet)
	fmt.Printf("扣分：%d\n", score)
	fmt.Printf("连线：%v\n", winLines)
	fmt.Printf("特定符号：%s\n", SymbolToString(specificSymbol))
	fmt.Printf("连线倍数：%d\n", winMultiplier)
	fmt.Printf("连线赢得：%d\n", normalWin)
	fmt.Printf("幸运转盘倍数：%d\n", wheelMultiplier)
	fmt.Printf("幸运转盘额外倍数：%v\n", extraMultipliers)
	fmt.Printf("幸运转盘赢得：%d\n", wheelWin)
	fmt.Printf("本局总赢得：%d\n", normalWin+wheelWin)
}

func TestGetWinSymbol(t *testing.T) {
	tests := []struct {
		name    string
		symbols []Symbol
		want    int64
	}{
		{
			name:    "10 00 5",
			symbols: []Symbol{TEN, ZEROZERO, FIVE},
			want:    10005,
		},
		{
			name:    "1 0 0",
			symbols: []Symbol{ONE, ZERO, ZERO},
			want:    100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetWinSymbol(tt.symbols)
			if got != tt.want {
				t.Errorf("GetWinSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}
