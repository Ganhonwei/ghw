package rm2

import (
	"errors"
	"fmt"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/handler"
	"goserver/pkg/utils"
	"strings"
	"testing"
)

func TestA(t *testing.T) {
	count := 10
	for i := 0; i < count; i++ {
		fmt.Println("4")
		var d []uint32
		// d := make([]uint32, algo.RMNumCard*2, algo.RMNumCard*2)
		d = append(d, algo.RMCARDS...)
		d = append(d, algo.RMCARDS...)

		// copy(d, algo.RMCARDS)
		//测试暂时去掉洗牌
		for i := len(d) - 1; i > 0; i-- {
			j := utils.RandIntN(i + 1)
			d[i], d[j] = d[j], d[i]
		}
		wildCard := d[0]
		cards := d[2:]

		// 插入 joker
		cards = append(cards, algo.RMJOKERS...)
		cards = append(cards, algo.RMJOKERS...)
		for i := len(cards) - 1; i > 0; i-- {
			j := utils.RandIntN(i + 1)
			cards[i], cards[j] = cards[j], cards[i]
		}

		ct := "N01,N01,N01,N01,E01,F01,E02,F0E,E03,F03,N02,N02,G01"
		cardType := strings.Split(ct, ",")

		hand, _ := algo.RMGetCard(cardType, cards, wildCard)
		groups := algo.GroupTheCards(hand, wildCard)
		fmt.Println("hand: ", hand)
		// fmt.Println("remain:", remain)
		fmt.Print("group: ")
		for _, g := range groups {
			fmt.Print(algo.CardsString(g))
		}
		fmt.Println("\nwild :", algo.CardsString([]uint32{wildCard}))
	}
}

func TestB(t *testing.T) {
	for i := 0; i < 1000; i++ {

		var d []uint32
		d = append(d, algo.RMCARDS...)
		d = append(d, algo.RMCARDS...)

		//测试暂时去掉洗牌
		for i := len(d) - 1; i > 0; i-- {
			j := utils.RandIntN(i + 1)
			d[i], d[j] = d[j], d[i]
		}

		wildCard := d[0] //万能牌
		_ = wildCard
		qiCards := d[1] //弃牌堆
		_ = qiCards
		cards := d[2:]

		cards = append(cards, algo.RMJOKERS...)
		cards = append(cards, algo.RMJOKERS...)
		for i := len(cards) - 1; i > 0; i-- {
			j := utils.RandIntN(i + 1)
			cards[i], cards[j] = cards[j], cards[i]
		}

		// 随机各发13张牌
		var cardsCount = make(map[uint32]int, 16)
		// var hand = 13
		// for i := 0; i < 6; i++ {
		// 	myHand := make([]uint32, 0, hand)
		// 	myHand = append(myHand, cards[:hand]...)
		// 	cards = cards[hand:]
		// 	fmt.Println(algo.CardsString(myHand))
		// 	checkRepeat(myHand, cardsCount)
		// }

		huCt := []string{"N01", "N01", "N01", "N01", "E01", "F01", "E02", "F0E", "E03", "F03", "N02", "N02", "G04"}
		for i := 0; i < 4; i++ {
			// hand := make([]uint32, len(huCt))
			var hand []uint32
			hand, cards = algo.RMGetCard(huCt, cards, wildCard)
			// t.DeskGame.Cards = remain
			// t.deal3Draw(v.Cards, t.DeskGame.Cards, drawNum)
			fmt.Println(algo.CardsString(hand))
			deal3Draw(hand, cards, 4)
			checkRepeat(hand, cardsCount)
		}
		fmt.Println("=======================================")
	}
}

// 从手牌中抽牌替换
func deal3Draw(hand, cards []uint32, drawNum int) {
	for i := 0; i < drawNum; i++ {
		m, n := utils.RandIntN(len(hand)), utils.RandIntN(len(cards))
		hand[m], cards[n] = cards[n], hand[m]
	}
}

func checkRepeat(cards []uint32, counter map[uint32]int) {
	for _, card := range cards {
		counter[card]++
		if counter[card] > 2 {
			fmt.Println("card repeate")
			panic(errors.New("card repeate"))
		}
	}
}

func TestRoi(t *testing.T) {
	// rm/desk_coin.go:779	roi生效判定: 5248452, rc=2, diamond=11600, winRate=0.22625, dieRate=116
	// rm/desk_coin.go:797	roi生效limit判定: 5248452, 测试1, limit=0, dayLimit=0
	fmt.Println(handler.AnalysisFreeStrategyParam("(0,2]", float64(2)))
	fmt.Println(handler.AnalysisFreeStrategyParam("(,50000]", float64(11600)))
	fmt.Println(handler.AnalysisFreeStrategyParam("(0,1.2]", float64(0.22625)))
	fmt.Println(handler.AnalysisFreeStrategyParam("(,120]", float64(116)))
	// fmt.Println(handler.AnalysisFreeStrategyParam("(0,2]", float64(0)))
	// fmt.Println(handler.AnalysisFreeStrategyParam("(0,2]", float64(0)))
}
