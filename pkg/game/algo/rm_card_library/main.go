package main

import (
	"fmt"
	"goserver/pkg/game/algo"
	"goserver/pkg/utils"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Element struct {
	joker uint32
	seats [6][13]uint32
	index string
}

type Group struct {
	num    int32
	ele    []*Element
	cindex string
}

var (
	cardLibrary   map[string]*Group
	cards         []uint32
	mux           sync.Mutex
	wg            sync.WaitGroup
	goroutineNum  int
	goroutineTask int
)

func init() {
	goroutineNum = 10

	goroutineTask = 1_000_000

	cardLibrary = make(map[string]*Group)
	for i := 0; i < 4; i++ {
		cards = append(cards, algo.RMCARDS...)
		cards = append(cards, algo.RMJOKERS...)
	}
}

func getCards() []uint32 {
	new := append([]uint32{}, cards...)
	return new
}

func shuffle(cards []uint32) {
	for i := len(cards) - 1; i > 0; i-- {
		j := utils.RandIntN(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

func deal(cards []uint32) (joker uint32, seats [6][13]uint32) {
	joker = cards[0]
	cards = cards[1:]

	for i := 0; i < 13; i++ {
		for j := 0; j < 6; j++ {
			seats[j][i] = cards[0]
			cards = cards[1:]
		}
	}
	return
}

func getLevel(score int64) string {
	if score == 0 {
		return "2"
	} else if score >= 1 && score <= 19 {
		return "3"
	} else if score >= 20 && score <= 39 {
		return "4"
	} else if score >= 40 && score <= 59 {
		return "5"
	} else if score >= 60 && score <= 79 {
		return "6"
	} else if score >= 80 {
		return "7"
	}
	return ""
}

func getIndex(joker uint32, seats [6][13]uint32) string {
	var index string

	for _, cards := range seats {
		sort := algo.SortCards3(cards, joker)
		score := algo.CalcScore(sort, joker)
		level := getLevel(score)
		index += level
	}
	return index
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

func insertElement(joker uint32, seats [6][13]uint32) {
	index := getIndex(joker, seats)
	cindex := convertIndex(index)

	mux.Lock()
	defer mux.Unlock()

	if group, ok := cardLibrary[cindex]; ok {
		group.num++
		group.ele = append(group.ele, &Element{joker, seats, index})
	} else {
		group := &Group{}
		group.num++
		group.cindex = cindex
		group.ele = append(group.ele, &Element{joker, seats, index})
		cardLibrary[index] = group
	}
}

func generator(done chan struct{}) {
	defer wg.Done()
	for i := 0; i < goroutineTask; i++ {
		cards := getCards()
		shuffle(cards)
		joker, seats := deal(cards)
		insertElement(joker, seats)
		done <- struct{}{}
		// time.Sleep(1 * time.Microsecond)
		// fmt.Printf("%d\n", i)
	}
}

func count(done chan struct{}) {
	defer wg.Done()
	count := goroutineNum * goroutineTask

	for {
		select {
		case <-done:
			count--
			fmt.Printf("%d\n", count)
			if count == 0 {
				return
			}
		}
	}
}

func main() {
	file, err := os.Create("output.txt")
	if err != nil {
		fmt.Println("无法创建文件:", err)
		return
	}
	defer file.Close()

	startTime := time.Now()
	done := make(chan struct{})

	// generator()
	wg.Add(goroutineNum + 1)
	for i := 0; i < goroutineNum; i++ {
		go generator(done)
	}
	go count(done)

	wg.Wait()

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)

	var cardLibrarySlice []*Group
	for _, group := range cardLibrary {
		cardLibrarySlice = append(cardLibrarySlice, group)
	}

	sort.Slice(cardLibrarySlice, func(i, j int) bool {
		// 按值进行降序排序
		return cardLibrarySlice[i].num > cardLibrarySlice[j].num
	})

	for _, group := range cardLibrarySlice {
		fmt.Fprintf(file, "cindex:%s, group num:%d\n", group.cindex, group.num)
	}

	fmt.Fprintf(file, "程序执行时间：%s\n", elapsedTime)
}
