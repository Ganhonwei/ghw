package main

import (
	"bufio"
	"fmt"
	"goserver/pkg/game/algo"
	"log"
	"math/rand/v2"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

func removeMultipleIndexes(arr []uint32, indexes []int) []uint32 {
	// 先将索引位置按升序排序
	sort.Ints(indexes)

	// 创建一个新的切片，只包含要保留的元素
	result := make([]uint32, 0, len(arr)-len(indexes))
	currentIndex := 0

	for i, v := range arr {
		// 如果当前索引需要保留，则添加到新切片中
		if currentIndex < len(indexes) && i == indexes[currentIndex] {
			currentIndex++
		} else {
			result = append(result, v)
		}
	}

	return result
}

func judge1(c1, c2 []uint32) bool {
	if algo.HuaType(c1) == algo.GaoPai && algo.HuaType(c2) == algo.GaoPai {
		return true
	}
	return false
}

func judge2(c1, c2 []uint32) bool {
	if algo.Duizi28Or9A(c1) == algo.Duizi28 && (algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}

	c1, c2 = c2, c1
	if algo.Duizi28Or9A(c1) == algo.Duizi28 && (algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}

	return false
}
func judge3(c1, c2 []uint32) bool {
	if algo.Duizi28Or9A(c1) == algo.Duizi9A && (algo.Duizi28Or9A(c2) == algo.Duizi9A || algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}

	c1, c2 = c2, c1
	if algo.Duizi28Or9A(c1) == algo.Duizi9A && (algo.Duizi28Or9A(c2) == algo.Duizi9A || algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}
	return false
}
func judge4(c1, c2 []uint32) bool {
	if algo.HuaType(c1) == algo.TongHua && (algo.HuaType(c2) == algo.TongHua || algo.Duizi28Or9A(c2) == algo.Duizi9A || algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}
	c1, c2 = c2, c1
	if algo.HuaType(c1) == algo.TongHua && (algo.HuaType(c2) == algo.TongHua || algo.Duizi28Or9A(c2) == algo.Duizi9A || algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}
	return false
}
func judge5(c1, c2 []uint32) bool {
	if algo.HuaType(c1) == algo.ShunZi && (algo.HuaType(c2) == algo.ShunZi || algo.HuaType(c2) == algo.TongHua || algo.Duizi28Or9A(c2) == algo.Duizi9A || algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}
	c1, c2 = c2, c1
	if algo.HuaType(c1) == algo.ShunZi && (algo.HuaType(c2) == algo.ShunZi || algo.HuaType(c2) == algo.TongHua || algo.Duizi28Or9A(c2) == algo.Duizi9A || algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}
	return false
}
func judge6(c1, c2 []uint32) bool {
	if algo.HuaType(c1) == algo.TongHuaShun && (algo.HuaType(c2) == algo.TongHuaShun || algo.HuaType(c2) == algo.ShunZi || algo.HuaType(c2) == algo.TongHua || algo.Duizi28Or9A(c2) == algo.Duizi9A || algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}
	c1, c2 = c2, c1
	if algo.HuaType(c1) == algo.TongHuaShun && (algo.HuaType(c2) == algo.TongHuaShun || algo.HuaType(c2) == algo.ShunZi || algo.HuaType(c2) == algo.TongHua || algo.Duizi28Or9A(c2) == algo.Duizi9A || algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}
	return false
}
func judge7(c1, c2 []uint32) bool {
	if algo.HuaType(c1) == algo.BaoZi && (algo.HuaType(c2) == algo.BaoZi || algo.HuaType(c2) == algo.TongHuaShun || algo.HuaType(c2) == algo.ShunZi || algo.HuaType(c2) == algo.TongHua || algo.Duizi28Or9A(c2) == algo.Duizi9A || algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}
	c1, c2 = c2, c1
	if algo.HuaType(c1) == algo.BaoZi && (algo.HuaType(c2) == algo.BaoZi || algo.HuaType(c2) == algo.TongHuaShun || algo.HuaType(c2) == algo.ShunZi || algo.HuaType(c2) == algo.TongHua || algo.Duizi28Or9A(c2) == algo.Duizi9A || algo.Duizi28Or9A(c2) == algo.Duizi28 || algo.HuaType(c2) == algo.GaoPai) {
		return true
	}
	return false
}

func mainx() {
	cards := append([]uint32{}, algo.RMCARDS...)

	cards_length := len(cards)

	count1 := 0
	count2 := 0
	count3 := 0
	count4 := 0
	count5 := 0
	count6 := 0
	count7 := 0

	// type pair struct {
	// 	c1 []uint32
	// 	c2 []uint32
	// }

	// pairs := []pair{}
	count := 0

	for i := 0; i < cards_length; i++ {
		for j := i + 1; j < cards_length; j++ {
			for k := j + 1; k < cards_length; k++ {
				// c1 := []uint32{cards[i], cards[j], cards[k]}

				for m := k + 1; m < cards_length; m++ {
					for n := m + 1; n < cards_length; n++ {
						for o := n + 1; o < cards_length; o++ {
							// c2 := []uint32{cards[m], cards[n], cards[o]}
							// pairs = append(pairs, pair{c1, c2})
							// fmt.Println(len(pairs))

							new_cards := []uint32{cards[i], cards[j], cards[k], cards[m], cards[n], cards[o]}
							new_cards_length := len(new_cards)

							for p := 0; p < new_cards_length; p++ {
								for q := p + 1; q < new_cards_length; q++ {
									for r := q + 1; r < new_cards_length; r++ {
										c1 := []uint32{new_cards[p], new_cards[q], new_cards[r]}
										c2 := removeMultipleIndexes(new_cards, []int{p, q, r})
										if judge1(c1, c2) {
											count1++
										} else if judge2(c1, c2) {
											count2++
										} else if judge3(c1, c2) {
											count3++
										} else if judge4(c1, c2) {
											count4++
										} else if judge5(c1, c2) {
											count5++
										} else if judge6(c1, c2) {
											count6++
										} else if judge7(c1, c2) {
											count7++
										}
										// if judge4(c1, c2) {
										// 	count4++
										// }
										// pairs = append(pairs, pair{c1, c2})
										count++
										// fmt.Println(len(pairs))

									}
								}
							}
						}
					}
				}

			}
		}
	}
	fmt.Printf("count:%d,count1:%d,count2:%d,count3:%d,count4:%d,count5:%d,count6:%d,count7:%d,count:%d\n", count, count1, count2, count3, count4, count5, count6, count7, count)

}

// func gosperHack(n, k int) [][]int {
// 	var result [][]int

// 	// 初始状态：最右边的k位为1
// 	x := (1 << k) - 1

// 	for x < (1 << n) {
// 		// 将当前状态转换为组合
// 		combination := make([]int, 0, k)
// 		for i := 0; i < n; i++ {
// 			if x&(1<<i) != 0 {
// 				combination = append(combination, i)
// 			}
// 		}
// 		result = append(result, combination)

// 		// 计算下一个状态
// 		c := x & -x
// 		r := x + c
// 		x = (((r ^ x) >> 2) / c) | r
// 	}

// 	return result
// }

func gosperHack(n, k int, excluded []int) [][]int {
	var result [][]int

	// 创建一个映射来快速检查一张牌是否被排除
	excludedMap := make(map[int]bool)
	for _, e := range excluded {
		excludedMap[e] = true
	}

	// 初始状态：最右边的k位为1
	x := (1 << k) - 1

	for x < (1 << n) {
		// 将当前状态转换为组合
		combination := make([]int, 0, k)
		validCount := 0
		for i := 0; i < n; i++ {
			if x&(1<<i) != 0 {
				if !excludedMap[i] {
					combination = append(combination, i)
					validCount++
				}
			}
		}

		// 只有当组合中的有效元素数量等于k时，才添加到结果中
		if validCount == k {
			result = append(result, combination)
		}

		// 计算下一个状态
		c := x & -x
		r := x + c
		x = (((r ^ x) >> 2) / c) | r
	}

	return result
}

func main2() {

	res := gosperHack(6, 3, []int{})

	fmt.Println(len(res))

	count := 0
	for _, v := range res {
		res2 := gosperHack(6, 3, v)
		for _, v2 := range res2 {
			_ = v2
			count++
			fmt.Printf("%v,%v\n", v, v2)
		}
	}
	fmt.Println(count)
}

func writeIntArrayToFile(filename string, arr [][]int) error {
	// 创建或打开文件
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("无法创建文件: %v", err)
	}
	defer file.Close()

	// 创建一个带缓冲的writer
	writer := bufio.NewWriter(file)

	// 遍历二维数组
	for _, row := range arr {
		typ := algo.HuaTypeUpOrDown(algo.IndexsToCards([]int{row[0], row[1], row[2]}))

		// 将每个int转换为string
		strRow := make([]string, len(row))
		for i, num := range row {
			strRow[i] = fmt.Sprintf("%d", num)
		}

		// 将这一行的数字用空格连接，并添加换行符
		line := strings.Join(strRow, " ")
		line2 := fmt.Sprintf("%s:%d\n", line, typ)
		// 写入这一行
		_, err := writer.WriteString(line2)
		if err != nil {
			return fmt.Errorf("写入文件时出错: %v", err)
		}
	}

	// 确保所有缓冲的数据都写入文件
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("刷新writer时出错: %v", err)
	}

	return nil
}

func main3() {
	res := gosperHack(52, 3, []int{})

	sort.Slice(res, func(i, j int) bool {
		c1 := algo.IndexsToCards(res[i])
		c2 := algo.IndexsToCards(res[j])
		return algo.HuaCompare(c1, c2)
	})

	writeIntArrayToFile("res.txt", res)
}

func mergeSlices(slices ...[]int) []int {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]int, 0, totalLen)
	for _, s := range slices {
		result = append(result, s...)
	}

	return result
}

func main4() {

	n, k := 52, 3
	// count := 0

	cardProbMap := make(map[string]int)

	increCardProb := func(card []int) {
		key := fmt.Sprintf("%02d,%02d,%02d", card[0], card[1], card[2])
		if _, ok := cardProbMap[key]; !ok {
			cardProbMap[key] = 1
		} else {
			cardProbMap[key]++
		}
	}

	res1 := gosperHack(n, k, []int{})
	for _, v1 := range res1 {
		res2 := gosperHack(n, k, v1)
		for _, v2 := range res2 {
			if !algo.HuaCompare(algo.IndexsToCards(v1), algo.IndexsToCards(v2)) {
				continue
			}
			increCardProb(v1)
			// res3 := gosperHack(n, k, mergeSlices(v1, v2))
			// for _, v3 := range res3 {
			// 	if !algo.HuaCompare(algo.IndexsToCards(v1), algo.IndexsToCards(v3)) {
			// 		break
			// 	}
			// 	res4 := gosperHack(n, k, mergeSlices(v1, v2, v3))
			// 	for _, v4 := range res4 {
			// 		if !algo.HuaCompare(algo.IndexsToCards(v1), algo.IndexsToCards(v4)) {
			// 			break
			// 		}
			// 		res5 := gosperHack(n, k, mergeSlices(v1, v2, v3, v4))
			// 		for _, v5 := range res5 {
			// 			if !algo.HuaCompare(algo.IndexsToCards(v1), algo.IndexsToCards(v5)) {
			// 				break
			// 			}
			// 			fmt.Println(v1, v2, v3, v4, v5)
			// 			count++
			// 		}
			// 	}
			// }
		}
	}
	fmt.Println(cardProbMap["00,13,26"])
}

func writeMapToFile(m map[string]int, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for key, value := range m {
		cards := strings.Split(key, ",")

		card1, _ := strconv.Atoi(cards[0])
		card2, _ := strconv.Atoi(cards[1])
		card3, _ := strconv.Atoi(cards[2])

		typ := algo.HuaTypeUpOrDown(algo.IndexsToCards([]int{card1, card2, card3}))

		_, err := fmt.Fprintf(writer, "%s:%d:%d\n", key, value, typ)
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

// func main() {
// 	//核算

// 	counter := make(map[uint32]int)
// 	little := [][]int{}
// 	// a := []int{1, 2, 3}
// 	a := []int{37, 38, 39}
// 	combine := gosperHack(52, 3, a)
// 	for _, v := range combine {
// 		if algo.HuaCompare(algo.IndexsToCards(a), algo.IndexsToCards(v)) {
// 			continue
// 		}
// 		little = append(little, v)
// 		typ := algo.HuaType(algo.IndexsToCards(v))
// 		counter[typ]++
// 	}

// 	total := 0
// 	for k, v := range counter {
// 		fmt.Print(k, ":", v, "\n")
// 		total += v
// 	}

// 	fmt.Println("total:", total)
// 	writeIntArrayToFile("res.txt", little)
// }

func main5() {
	f, _ := os.OpenFile("logfile.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	log.SetOutput(f)
	//计算2人
	n, k := 52, 3
	// count := 0

	counter := struct {
		sync.Mutex
		m     map[string]int
		count int
	}{
		m: make(map[string]int),
	}

	// increCardProb := func(card []int) {
	// 	counter.Lock()
	// 	defer counter.Unlock()
	// 	key := fmt.Sprintf("%02d,%02d,%02d", card[0], card[1], card[2])
	// 	if _, ok := counter.m[key]; !ok {
	// 		counter.m[key] = 1
	// 	} else {
	// 		counter.m[key]++
	// 	}
	// }

	start := time.Now()

	var wg sync.WaitGroup

	runTask := func(i interface{}, localCounter map[string]int) int {
		// fmt.Printf("Running task %v\n", i.([]int))
		count := 0
		v1 := i.([]int)
		res2 := gosperHack(n, k, v1)
		for _, v2 := range res2 {
			count++
			if !algo.HuaCompare(algo.IndexsToCards(v1), algo.IndexsToCards(v2)) {
				continue
			}
			// count++
			// increCardProb(v1)
			key := fmt.Sprintf("%02d,%02d,%02d", v1[0], v1[1], v1[2])
			localCounter[key]++
		}
		return count
	}

	p, _ := ants.NewPoolWithFunc(16, func(i interface{}) {
		localCounter := make(map[string]int)
		count := runTask(i, localCounter)

		counter.Lock()
		counter.count += count
		for key, count := range localCounter {
			counter.m[key] += count
		}
		counter.Unlock()
		log.Println(float64(counter.count) / float64(407170400))
		wg.Done()
	})
	defer p.Release()

	res1 := gosperHack(n, k, []int{})

	for _, v1 := range res1 {
		key := fmt.Sprintf("%02d,%02d,%02d", v1[0], v1[1], v1[2])
		counter.m[key] = 0
	}

	for _, v1 := range res1 {
		wg.Add(1)
		p.Invoke(v1)
	}

	wg.Wait()

	// fmt.Println(count)

	end := time.Now()
	fmt.Println(end.Sub(start))

	// fmt.Println(counter.m["00,13,26"])
	writeMapToFile(counter.m, "output2.txt")

}

func main6() {

	f, _ := os.OpenFile("logfile.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	log.SetOutput(f)
	//计算3人
	n, k := 13*4, 3
	thread := 192 * 8
	// count := 0

	counter := struct {
		sync.Mutex
		m     map[string]int
		count uint64
	}{
		m: make(map[string]int),
	}

	// increCardProb := func(card []int) {
	// 	counter.Lock()
	// 	defer counter.Unlock()
	// 	key := fmt.Sprintf("%02d,%02d,%02d", card[0], card[1], card[2])
	// 	if _, ok := counter.m[key]; !ok {
	// 		counter.m[key] = 1
	// 	} else {
	// 		counter.m[key]++
	// 	}
	// }

	start := time.Now()

	var wg sync.WaitGroup

	var p *ants.PoolWithFunc
	var p1 *ants.PoolWithFunc

	runTask := func(i interface{}, localCounter map[string]int) uint64 {
		// fmt.Printf("Running task %v\n", i.([]int))
		var count uint64 = 0
		v1 := i.([]int)
		length := len(v1)
		res2 := gosperHack(n, k, v1)
		for _, v2 := range res2 {
			count++
			if !algo.HuaCompare(algo.IndexsToCards(v1), algo.IndexsToCards(v2)) {
				continue
			}

			if length == 6 {
				key := fmt.Sprintf("%02d,%02d,%02d", v1[0], v1[1], v1[2])
				localCounter[key]++
				// fmt.Printf("%v,%v\n", v1, v2)
			} else {
				wg.Add(1)
				p1.Invoke(mergeSlices(v1, v2))
			}
		}
		return count
	}

	poolFunc := func(i interface{}) {
		localCounter := make(map[string]int)
		count := runTask(i, localCounter)

		counter.Lock()
		counter.count += count
		for key, count := range localCounter {
			counter.m[key] += count
		}
		counter.Unlock()

		if rand.Float32() < 0.001 {
			log.Println(counter.count)
		}

		wg.Done()

	}

	p, _ = ants.NewPoolWithFunc(thread, poolFunc)
	defer p.Release()

	p1, _ = ants.NewPoolWithFunc(thread, poolFunc)
	defer p1.Release()

	res1 := gosperHack(n, k, []int{})

	for _, v1 := range res1 {
		key := fmt.Sprintf("%02d,%02d,%02d", v1[0], v1[1], v1[2])
		counter.m[key] = 0
	}

	for _, v1 := range res1 {
		wg.Add(1)
		p.Invoke(v1)
	}

	wg.Wait()

	// fmt.Println(count)

	end := time.Now()
	fmt.Println(end.Sub(start))

	// fmt.Println(counter.m["00,13,26"])
	writeMapToFile(counter.m, "output3.txt")

}

func main7() {

	f, _ := os.OpenFile("logfile.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	log.SetOutput(f)
	//计算4人
	n, k := 13*4, 3
	thread := 192 * 8
	// count := 0

	counter := struct {
		sync.Mutex
		m     map[string]int
		count uint64
	}{
		m: make(map[string]int),
	}

	start := time.Now()

	var wg sync.WaitGroup

	var p *ants.PoolWithFunc
	var p1 *ants.PoolWithFunc
	var p2 *ants.PoolWithFunc

	runTask := func(i interface{}, localCounter map[string]int) uint64 {
		// fmt.Printf("Running task %v\n", i.([]int))
		var count uint64 = 0
		v1 := i.([]int)
		length := len(v1)
		res2 := gosperHack(n, k, v1)
		for _, v2 := range res2 {
			count++
			if !algo.HuaCompare(algo.IndexsToCards(v1), algo.IndexsToCards(v2)) {
				continue
			}

			if length == 9 {
				key := fmt.Sprintf("%02d,%02d,%02d", v1[0], v1[1], v1[2])
				localCounter[key]++
				// fmt.Printf("%v,%v\n", v1, v2)
			} else if length == 6 {
				wg.Add(1)
				p2.Invoke(mergeSlices(v1, v2))
			} else {
				wg.Add(1)
				p1.Invoke(mergeSlices(v1, v2))
			}
		}
		return count
	}

	poolFunc := func(i interface{}) {
		localCounter := make(map[string]int)
		count := runTask(i, localCounter)

		counter.Lock()
		counter.count += count
		for key, count := range localCounter {
			counter.m[key] += count
		}
		counter.Unlock()

		if rand.Float32() < 0.001 {
			log.Println(counter.count)
		}

		wg.Done()

	}

	p, _ = ants.NewPoolWithFunc(thread, poolFunc)
	defer p.Release()

	p1, _ = ants.NewPoolWithFunc(thread, poolFunc)
	defer p1.Release()

	p2, _ = ants.NewPoolWithFunc(thread, poolFunc)
	defer p1.Release()

	res1 := gosperHack(n, k, []int{})

	for _, v1 := range res1 {
		key := fmt.Sprintf("%02d,%02d,%02d", v1[0], v1[1], v1[2])
		counter.m[key] = 0
	}

	for _, v1 := range res1 {
		wg.Add(1)
		p.Invoke(v1)
	}

	wg.Wait()

	// fmt.Println(count)

	end := time.Now()
	fmt.Println(end.Sub(start))

	// fmt.Println(counter.m["00,13,26"])
	writeMapToFile(counter.m, "output4.txt")

}

func main() {

	t1 := time.Now()
	n, k := 52, 3

	count := 0
	res1 := gosperHack(n, k, []int{})
	for _, v1 := range res1 {
		res2 := gosperHack(n, k, v1)
		for _, v2 := range res2 {
			// fmt.Printf("%v,%v\n", v1, v2)
			_ = v2
			count++
		}
	}

	println(count)
	t2 := time.Now()
	fmt.Println(t2.Sub(t1))
}
