package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

func getContainersMetrics() []RedisEntry {
	cmdStr := "docker stats --format \"{{.Name}} ; {{.CPUPerc}} ; {{.MemUsage}}\" --no-stream"
	out, err := exec.Command("/usr/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		log.Println(err)
		return nil
	}
	ts := time.Now().Unix()
	output := string(out)
	lines := strings.Split(output, "\n")
	var entries []RedisEntry
	for _, ln := range lines {
		if ln != "" {
			columns := strings.Split(ln, ";")
			cntName, cpuUsg, memUsg := columns[0], columns[1], columns[2]
			memTmp := strings.Trim(strings.Split(memUsg, "/")[0], " ")
			nameFmt := strings.Trim(cntName, " ")
			memFmt := strings.ReplaceAll(memTmp, "MiB", "")
			cpuFmt := strings.Trim(strings.ReplaceAll(cpuUsg, "%", ""), " ")
			entry := fmt.Sprintf("%v,%v,%v", cpuFmt, memFmt, ts)
			entries = append(entries, RedisEntry{
				keyName:  "docker:metrics:" + nameFmt,
				keyvalue: entry,
			})
		}
	}
	return entries
}

func storeContainerMetrics(interval int) {
	cli := newRedisCli(redisAddr, context.Background())
	for {
		if entries := getContainersMetrics(); entries != nil {
			existingKeys, err := cli.r.Keys(cli.ctx, "docker:metrics:*").Result()
			if err != nil {
				log.Println(err)
			}

			var mkeys []string
			for _, i := range entries {
				if !isInSlice(i.keyName, existingKeys) {
					cli.r.Set(cli.ctx, i.keyName, "cpu,mem,ts", 0)
				}
				mkeys = append(mkeys, i.keyName)
			}

			keysContent, _ := cli.r.MGet(cli.ctx, mkeys...).Result()
			pairs := make(map[string]interface{})
			for i := 0; i < len(keysContent); i++ {
				payload := keysContent[i].(string) + "\n" + entries[i].keyvalue
				pairs[entries[i].keyName] = payload
			}

			cli.r.MSet(cli.ctx, pairs)
		} else {
			log.Println("no containers metrics found")
		}
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func isInSlice(val string, arr []string) bool {
	for _, a := range arr {
		if val == a {
			return true
		}
	}
	return false
}
