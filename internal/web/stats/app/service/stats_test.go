package service

import (
	"fmt"
	"goserver/pkg/utils"
	"testing"
	"time"
)

func TestTimeZone(t *testing.T) {
	now := time.Now()
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		panic(err)
	}
	india := now.In(loc)
	fmt.Println(india)
	fmt.Println(now.Format(utils.FORMAT))
	fmt.Println(india.Format(utils.FORMAT))

	// utils.Time2Stamp(startTime)
	fmt.Println(fmt.Sprintf("%c", 65))
}
