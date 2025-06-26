package up7

import (
	"fmt"
	"goserver/pkg/utils"
	"testing"
)

func TestMultiple(t *testing.T) {
	choices := []utils.ChoiceI64[int]{
		{Item: 1, Weight: parsePctMultiple("85.714285770%", 100_000_000_000)},
		{Item: 2, Weight: parsePctMultiple("10.702413820%", 100_000_000_000)},
		{Item: 3, Weight: parsePctMultiple("1.728811880%", 100_000_000_000)},
		{Item: 4, Weight: parsePctMultiple("0.676935640%", 100_000_000_000)},
		{Item: 5, Weight: parsePctMultiple("0.358438740%", 100_000_000_000)},
	}

	for range 1000 {
		c, err := utils.WeightedChoiceI64(choices)
		if err != nil {
			t.Error(err)
			return
		}
		if c.Item != 1 {
			fmt.Println(c.Item, c.Weight)
		}
	}
}
