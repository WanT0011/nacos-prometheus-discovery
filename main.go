package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"nacos-prometheus-discovery/model"
	"nacos-prometheus-discovery/service"
	"os"
	"os/signal"
	"strconv"
	"time"
)

const (
	DefaultConfigPath = "conf/config.json"

	MODE_CONFIG  = "config"
	MODE_SERVICE = "service"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	log.Println("Nacos Prometheus Discovery Starting ...")
	filename := DefaultConfigPath
	if len(os.Args) > 1 {
		argsWithoutProg := os.Args[1:2]
		log.Println("config file path:", argsWithoutProg)
		filename = argsWithoutProg[0]
	}

	// read config file
	configJson, configErr := ioutil.ReadFile(filename)
	if configErr != nil {
		log.Fatal("read config file error.", configErr)
	}
	config := model.Config{}
	json.Unmarshal(configJson, &config)

	// start timer
	// 从环境变量中获取配置间隔,单位秒
	intervalInSecondStr := os.Getenv("INTERVAL_INSECOND")
	intervalInSecond, _ := strconv.Atoi(intervalInSecondStr)
	ticker := time.NewTicker(time.Second * time.Duration(intervalInSecond))
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		// listen terminate sig and stop
		s := <-c
		fmt.Println("terminated.", s)
		done <- true
	}()
	for {
		select {
		case <-done:
			fmt.Println("Exit!")
			return
		case t := <-ticker.C:
			mode := config.Mode
			log.Printf("Current time:%s mode: %s", t, mode)
			if mode == MODE_CONFIG {
				service.FetchPrometheusConfig(config)
			}
			if mode == MODE_SERVICE {
				service.GeneratePrometheusTarget(config)
			}
		}
	}
}
