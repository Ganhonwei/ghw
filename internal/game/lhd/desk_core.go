package lhd

// 自然概率开奖
func (t *Desk) NatureWinner() {
	// 指定房间
	// if t.Dtype != int32(pb.DESK_TYPE_NORMAL) {
	// 	return
	// }

	t.DeskFree.LHCards = make([]uint32, 0)
	for {
		cards := make([]uint32, 1)
		tmp := t.DeskGame.Cards[:1]
		copy(cards, tmp)

		t.DeskGame.Cards = t.DeskGame.Cards[1:]
		t.DeskFree.LHCards = append(t.DeskFree.LHCards, cards[0])

		if len(t.DeskFree.LHCards) >= 2 {
			break
		}
	}
	t.Power[1] = t.DeskFree.LHCards[0]
	t.Power[2] = t.DeskFree.LHCards[1]

	// 中奖上限
}
