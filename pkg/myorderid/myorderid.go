package myorderid

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"goserver/pkg/myredis"
)

const (
	Part1Key = "order:part1"
	Part2Key = "order:part2"
	Part3Key = "order:part3"
)

func InitOrderID(addr string, db int) {
	myredis.InitRedis(addr, db)
}

func padLeft(num int64, width int) string {
	return fmt.Sprintf("%0*d", width, num)
}

func GenerateOrderID(suffix string) (string, error) {
	ctx := context.Background()
	luaScript := `
local part1 = redis.call('GET', KEYS[1])
if not part1 then
    part1 = 17724
    redis.call('SET', KEYS[1], part1)
else
    part1 = tonumber(part1)
end

local step2 = tonumber(ARGV[1])
local part2 = redis.call('INCRBY', KEYS[2], step2)
if part2 > 999999 then
    part2 = 0
    redis.call('SET', KEYS[2], part2)
    part1 = part1 + 1
    redis.call('SET', KEYS[1], part1)
end

local step3 = tonumber(ARGV[2])
local part3 = redis.call('INCRBY', KEYS[3], step3)
if part3 > 9999 then
    part3 = 0
    redis.call('SET', KEYS[3], part3)
end

return {part1, part2, part3}
`
	part1Key := "order:part1"
	part2Key := "order:part2"
	part3Key := "order:part3"

	rand.Seed(time.Now().UnixNano())
	step2 := rand.Intn(101) + 100 // 100~300
	step3 := rand.Intn(101) + 100 // 100~300

	result, err := myredis.Redis().Eval(ctx, luaScript, []string{part1Key, part2Key, part3Key}, step2, step3).Result()
	if err != nil {
		return "", err
	}
	vals := result.([]interface{})
	part1 := vals[0].(int64)
	part2 := vals[1].(int64)
	part3 := vals[2].(int64)

	orderID := fmt.Sprintf("%05d-%06d-%04d%s", part1, part2, part3, suffix)
	return orderID, nil
}
