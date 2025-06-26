package main

import (
	"context"
	"encoding/json"
	"fmt"
	"goserver/pkg/game/algo"
	"goserver/pkg/utils"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// type Element struct {
// 	joker uint32
// 	seats [][]uint32
// 	index string //原始索引
// }

// type Group struct {
// 	num    int32
// 	ele    []*Element
// 	cindex string //排序后的索引
// }

type Group struct {
	Task   string     `json:"index" bson:"index"`
	Joker  uint32     `json:"joker" bson:"joker"`
	Seats  [][]uint32 `json:"seats" bson:"seats"`
	Remain []uint32   `json:"remain" bson:"remain"`
}

var (
	cardLibrary  []*Group
	cards        []uint32
	goroutineNum int
	indexSize    int
	mux          sync.Mutex
	wg           sync.WaitGroup
	client       *redis.Client
)

func init() {
	goroutineNum = 100000
	indexSize = 1000

	for i := 0; i < 4; i++ {
		cards = append(cards, algo.RMCARDS...)
	}

	client = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       2,
	})

}

func convertIndex(index string) string {
	chars := strings.Split(index, "")

	nums := []int{}
	for _, char := range chars {
		num, _ := strconv.Atoi(char)
		nums = append(nums, num)
	}
	sort.Ints(nums)

	var ret string
	for _, num := range nums {
		ret += strconv.Itoa(num)
	}
	return ret
}

func getCombine() (ret []string) {
	combine := make(map[string]struct{})
	m := 2
	n := 7
	for a := m; a <= n; a++ {
		for b := m; b <= n; b++ {
			for c := m; c <= n; c++ {
				for d := m; d <= n; d++ {
					for e := m; e <= n; e++ {
						for f := m; f <= n; f++ {
							index := fmt.Sprintf("%d%d%d%d%d%d", a, b, c, d, e, f)
							cindex := convertIndex(index)
							if _, ok := combine[cindex]; !ok {
								combine[cindex] = struct{}{}
							}
						}
					}
				}
			}
		}
	}
	for k, _ := range combine {
		ret = append(ret, k)
	}

	for i := len(ret) - 1; i > 0; i-- {
		j := utils.RandIntN(i + 1)
		ret[i], ret[j] = ret[j], ret[i]
	}

	return
}

func getChunks(combine []string, chunkSize int) [][]string {
	var result [][]string
	length := len(combine) / chunkSize

	for i := 0; i < chunkSize-1; i++ {
		result = append(result, combine[:length])
		combine = combine[length:]
	}

	result = append(result, combine)

	return result
}

func getTask(combine []string) []string {
	var ret []string
	for _, v := range combine {
		for i := 0; i < indexSize; i++ {
			task := fmt.Sprintf("%s:%d", v, i)
			ret = append(ret, task)
		}
	}
	return ret
}

// func count(total int, done chan struct{}) {
// 	defer wg.Done()
// 	fmt.Printf("total:%d\n", total)
// 	for {
// 		select {
// 		case <-done:
// 			total--
// 			fmt.Printf("remain:%d\n", total)
// 			if total == 0 {
// 				fmt.Println("all done")
// 				return
// 			}
// 		}
// 	}
// }

func getLevel(score int64) string {
	if score == 0 {
		return "7"
	} else if score >= 1 && score <= 19 {
		return "6"
	} else if score >= 20 && score <= 39 {
		return "5"
	} else if score >= 40 && score <= 59 {
		return "4"
	} else if score >= 60 && score <= 79 {
		return "3"
	} else if score >= 80 {
		return "2"
	}
	return ""
}

func getIndex(joker uint32, seats [][]uint32) string {
	var index string

	for _, cards := range seats {
		sort := algo.SortCards4(cards, joker)
		score := algo.CalcScore(sort, joker)
		level := getLevel(score)
		index += level
	}
	return index
}

func insertGroup(task string, joker uint32, seats [][]uint32, remain []uint32) {
	mux.Lock()
	defer mux.Unlock()

	group := &Group{task, joker, seats, remain}
	cardLibrary = append(cardLibrary, group)
	// fmt.Printf("cardLibrary:%d", len(cardLibrary))
	// jsonData, err := json.Marshal(group)
	// if err != nil {
	// panic(err)
	// }

	// err = client.Set(context.Background(), task, jsonData, 0).Err()
	// if err != nil {
	// panic(err)
	// }
}

func getCards() []uint32 {
	new := append([]uint32{}, cards...)
	return new
}

func insertJoker(cards []uint32) []uint32 {
	new := append([]uint32{}, cards...)
	for i := 0; i < 4; i++ {
		new = append(new, algo.RMJOKERS...)
	}
	return new
}

func shuffle(cards []uint32) {
	for i := len(cards) - 1; i > 0; i-- {
		j := utils.RandIntN(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

func judgeCards(char string, joker uint32, cards []uint32) bool {
	sort := algo.SortCards4(cards, joker)
	score := algo.CalcScore(sort, joker)
	level := getLevel(score)
	if level == char {
		return true
	} else {
		return false
	}
}

func _getLevelCards(char string, joker uint32, cards []uint32) (ok bool, ret_cards []uint32, remain_cards []uint32) {
	news := append([]uint32{}, cards...)
	shuffle(news)

	ret_cards = news[:13]
	if judgeCards(char, joker, ret_cards) {
		remain_cards, _ = algo.RemoveCards(cards, ret_cards)
		ok = true
		return
	}

	remain_cards = news[13:]
	for _, v := range remain_cards {
		ret_cards = append(ret_cards, v)
		_, discardCard := algo.GetOneCardToDiscard(ret_cards, joker)
		ret_cards = algo.RemoveCard(ret_cards, discardCard)
		if judgeCards(char, joker, ret_cards) {
			remain_cards, _ = algo.RemoveCards(cards, ret_cards)
			ok = true
			return
		}
	}

	return
}

func getLevelCards(char string, joker uint32, cards []uint32) (ret_cards []uint32, remain_cards []uint32) {
	for i := 0; i < 10000; i++ {
		var ok bool
		ok, ret_cards, remain_cards = _getLevelCards(char, joker, cards)
		if ok {
			return
		}
	}
	panic("too many for loop")
}

func __generator(task string) (joker uint32, seats [][]uint32, remain []uint32) {
	index := strings.Split(task, ":")[0]
	chars := strings.Split(index, "")

	cards := getCards()
	shuffle(cards)
	joker = cards[0]
	cards = cards[1:]
	cards = insertJoker(cards)
	shuffle(cards)

	for _, char := range chars {
		var ret_cards []uint32
		ret_cards, cards = getLevelCards(char, joker, cards)
		seats = append(seats, ret_cards)
	}

	remain = cards
	return
}

func _generator(task string) {
	joker, seats, remain := __generator(task)
	insertGroup(task, joker, seats, remain)
}

func generator(taskChan <-chan string) {
	defer wg.Done()

	for task := range taskChan { //遍历所有索引
		_generator(task)
	}
}

func test() {
	main()
}

// func getGOMAXPROCS() int {
// 	// cpu := runtime.NumCPU()
// 	return runtime.GOMAXPROCS(0)
// }

func main() {
	// file, err := os.Create("output.txt")
	// if err != nil {
	// 	fmt.Println("无法创建文件:", err)
	// 	return
	// }
	// defer file.Close()
	// fmt.Println(getGOMAXPROCS())

	startTime := time.Now()

	taskChan := make(chan string)

	combine := getCombine()  //获取所有索引组合
	task := getTask(combine) //获取所有任务组合（索引:序号构成一个任务）

	// chunks := getChunks(task, goroutineNum) //将所有任务分组

	for i := 0; i < goroutineNum; i++ {
		wg.Add(1)
		go generator(taskChan)
	}

	len := len(task)
	for _, t := range task {
		taskChan <- t
		len--
		fmt.Printf("remain:%d\n", len)
	}
	close(taskChan)

	wg.Wait()

	fmt.Println("start write data to redis")
	pipe := client.Pipeline()
	count := 0
	for _, group := range cardLibrary {
		jsonData, err := json.Marshal(group)
		if err != nil {
			panic(err)
		}
		pipe.Set(context.Background(), group.Task, jsonData, 0)
		count++
		if count == 100000 {
			_, err := pipe.Exec(context.Background())
			if err != nil {
				panic(err)
			}
			count = 0
			fmt.Println("write 100000 record")
		}
	}
	_, err := pipe.Exec(context.Background())
	if err != nil {
		panic(err)
	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	fmt.Printf("程序执行时间：%s\n", elapsedTime)
}
