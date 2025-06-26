package crash

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/game/handler"
	"goserver/pkg/utils"

	console "github.com/AsynkronIT/goconsole"
	"github.com/AsynkronIT/protoactor-go/remote"
)

func TestNode(t *testing.T) {
	remote.Start("127.0.0.1:0")
	activate("hello")
	<-time.After(time.Minute)
	activate("hello")
	console.ReadLine()
}

func activate(name string) {
	timeout := 1 * time.Second
	pid, err := remote.SpawnNamed("127.0.0.1:8081", "remote1", name, timeout)
	if err != nil {
		//fmt.Println(err)
		//return
	}
	res, _ := pid.GetPid().RequestFuture(new(pb.Request), timeout).Result()
	fmt.Println("res ", res)
	response := res.(*pb.Response)
	fmt.Println(response)
	//pid.Stop()
	//
	//pid, _ = remote.SpawnNamed("127.0.0.1:8080", "remote2", name, timeout)
	res, _ = pid.GetPid().RequestFuture(new(pb.Request), timeout).Result()
	response = res.(*pb.Response)
	fmt.Println(response)
}

func TestBoom(t *testing.T) {
	// m := math.Pow10(18)
	// fmt.Println(m)
	// fmt.Println(int64(m))

	for i := 0; i < 100; i++ {
		var mulpitle int64 = 100
		m := math.Pow10(18)
		var rate int64
		for {
			if mulpitle == 101 {
				rate = int64(m - (97.0 / 101.0 * m))
			} else {
				rate = int64(m - (float64(mulpitle-1) / float64(mulpitle) * m))
			}

			if utils.RandInt64N(int64(m))+1 <= rate {
				fmt.Println(i, mulpitle)
				break
			}

			mulpitle += 1
		}
	}
}

func TestBoomTime(t *testing.T) {
	// mulpitle := math.Pow(math.E, 0.0821*60)
	mulpitle := 137.82
	result := math.Log(mulpitle)

	fmt.Println(result)
	ts := result / 0.0821
	fmt.Println(ts)
}

func TestTackoffMultiple(t *testing.T) {
	tackoff := time.Now()
	now := tackoff.Add(time.Millisecond * 300)
	m := handler.CrashTimeMultiple(tackoff, now)
	fmt.Println(m)
}

func TestBoomMultiples(t *testing.T) {
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// m := handler.CrashBoomMultiple(r)
	// fmt.Println(m)
	// fmt.Println("51514assd41s1212a23s1aa27a10aas21aaaas56s5")

	a := time.Now().UnixMilli()
	ranges := []int{0, 0, 0, 0, 0}
	for i := 0; i < 1000; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(i)))
		m := handler.CrashBoomMultiple(r)
		if m < 120 {
			ranges[0]++
		} else if m < 150 {
			ranges[1]++
		} else if m < 200 {
			ranges[2]++
		} else if m < 300 {
			ranges[3]++
		} else {
			ranges[4]++
		}
		fmt.Println(m)
	}
	fmt.Println("+---+----+----+--+--s", ranges, time.Now().UnixMilli()-a)
}
