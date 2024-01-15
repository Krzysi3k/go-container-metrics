package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func storeHostMetrics(interval int) {
	cli := newRedisCli(redisAddr, context.Background())
	for {
		time.Sleep(time.Second * time.Duration(interval))
		percent, _ := cpu.Percent(time.Second, false)
		memstat, _ := mem.VirtualMemory()
		if percent == nil || memstat.Used == 0 {
			log.Println("cannot get values from gopsutil")
			continue
		}
		ts := time.Now().Unix()
		cpuval := fmt.Sprintf("%.2f", percent[0])
		newrow := fmt.Sprintf("%s,%d,%d\n", cpuval, memstat.Used/1024/1024, ts)
		vals, err := cli.r.Get(cli.ctx, "docker:metrics:host.usage").Result()
		var sb strings.Builder
		if err == redis.Nil {
			sb.WriteString("cpu,mem,ts\n" + newrow)
		} else if err != nil {
			log.Println("redis error: ", err)
			continue
		} else {
			sb.WriteString(vals + newrow)
		}

		cli.r.Set(cli.ctx, "docker:metrics:host.usage", sb.String(), 0)
	}
}
