package myorderid

import (
	"fmt"
	"testing"
	"time"
)

func TestOrderID(t *testing.T) {
	InitOrderID("localhost:6379", 0)
	for {
		orderID, err := GenerateOrderID("0223")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(orderID)
		time.Sleep(2 * time.Second)
	}
}
