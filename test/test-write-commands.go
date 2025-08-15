package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// 连接到Redis
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer client.Close()

	ctx := context.Background()

	// 测试所有写命令
	writeCommands := []string{
		"HSET", "LPUSH", "RPUSH", "SADD", "ZADD", "SETBIT", "PFADD", "GEOADD", "XADD", "MSET",
	}

	fmt.Println("测试所有写命令...")
	fmt.Println("==================")

	for _, cmd := range writeCommands {
		fmt.Printf("测试命令: %s\n", cmd)
		
		start := time.Now()
		var err error
		
		switch cmd {
		case "HSET":
			err = client.HSet(ctx, "test:hash", "field1", "value1").Err()
		case "LPUSH":
			err = client.LPush(ctx, "test:list", "value1").Err()
		case "RPUSH":
			err = client.RPush(ctx, "test:list2", "value1").Err()
		case "SADD":
			err = client.SAdd(ctx, "test:set", "value1").Err()
		case "ZADD":
			err = client.ZAdd(ctx, "test:zset", redis.Z{Score: 1.0, Member: "value1"}).Err()
		case "SETBIT":
			err = client.SetBit(ctx, "test:bitmap", 0, 1).Err()
		case "PFADD":
			err = client.PFAdd(ctx, "test:hll", "value1").Err()
		case "GEOADD":
			err = client.GeoAdd(ctx, "test:geo", &redis.GeoLocation{Name: "point1", Longitude: 0, Latitude: 0}).Err()
		case "XADD":
			err = client.XAdd(ctx, &redis.XAddArgs{Stream: "test:stream", Values: map[string]any{"field": "value1"}}).Err()
		case "MSET":
			err = client.MSet(ctx, "key1", "value1", "key2", "value2").Err()
		}
		
		elapsed := time.Since(start)
		
		if err != nil {
			fmt.Printf("  ❌ 错误: %v\n", err)
		} else {
			fmt.Printf("  ✅ 成功 (耗时: %v)\n", elapsed)
		}
		
		fmt.Println()
	}

	fmt.Println("所有测试完成!")
} 