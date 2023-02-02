package main

import (
	"encoding/json"
	"fmt"
	"sms_timetosend_task/conf"
	"sms_timetosend_task/log"
	"sms_timetosend_task/redis"
)

var Status string

func main() {
	//启动smsOnlineTimetoSendService服务
	ch := make(chan string)
	go smsOnlineTimetoSendService(ch)

	ch2 := make(chan string)
	go mmsOnlineTimetoSendService(ch2)

	Status = "runing"
	for {
		result, err := redis.Conn.Brpop("go:jobs:smstimetosend", 5)
		if err == nil {
			var datajson map[string]interface{}
			err := json.Unmarshal(result[1].([]byte), &datajson)
			if err != nil {
				fmt.Println(err)
			} else {
				if datajson["signal"].(string) == "shutdown" {
					if Status == "runing" {
						fmt.Println("SMS Timetosend Service CLoseing")
						ch <- "exit"
						close(ch)

						ch2 <- "exit"
						close(ch2)
					}
					log.Logger.Warning("SMS Timetosend Service Now is Shuting Dwon")
					break
				} else {
					signal := datajson["signal"].(string)
					switch signal {
					case "stop":
						if Status == "runing" {
							fmt.Println("SMS Timetosend Service  CLoseing")
							ch <- "exit"
							close(ch)
							ch2 <- "exit"
							close(ch2)
							Status = "stopped"
						} else {
							log.Logger.Warning("SMS Timetosend Service Current state is Stopped")
						}
						break
					case "start":
						if Status == "stopped" {
							fmt.Println("SMS Timetosend Service Starting")
							ch := make(chan string)
							go smsOnlineTimetoSendService(ch)
							ch2 := make(chan string)
							go mmsOnlineTimetoSendService(ch2)
							Status = "runing"
						} else {
							log.Logger.Warning("SMS Timetosend Service Current state is Runing")
						}
						break
					case "reload":
						log.Logger.Warning("SMS Timetosend Service Loading Configures")
						conf.Reload()
						break
					default:
					}
				}
			}
		}
	}
}
