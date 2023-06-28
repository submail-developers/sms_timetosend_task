package main

import (
	"encoding/json"
	"fmt"
	"sms_timetosend_task/database"
	"sms_timetosend_task/log"
	"sms_timetosend_task/redis"
	"strconv"
	"time"
)

const TIME_LAYOUT = "2006-01-02 15:04:05"
const MaxRoutineNum = 20

func init() {
}

// 短信定时发送服务----------------------------------------
func smsOnlineTimetoSendService(ch chan string) {
	for {
		result, err := redis.Conn.Brpop("sms_timetosend_task", 30)
		if err == nil {
			// go smsTimetosendHandler(result[1].([]byte))
			sendlist := string(result[1].([]byte))
			item := &SMSSendlistFromDB{}
			db := database.DbPublic.Table("message_send_list").Select("timetosend").Where("id = ?", sendlist).Find(item)
			if db.Error == nil {
				timetosend, _ := time.Parse(TIME_LAYOUT, item.Timetosend)
				if timetosend.Unix()-time.Now().Unix() > 4*3660 {
					go smsTimetosendHandler(result[1].([]byte))
				} else {
					go func() {
						for {
							result, err := redis.Conn.Brpop("sms_timetosend_queue:"+sendlist, 5)
							if err == nil {
								tempItem := &TimetosendItem{}
								err = json.Unmarshal(result[1].([]byte), tempItem)
								if err == nil {
									tmpdata := &tmpSmsMessageStruct{}
									tmpdata.Address = tempItem.Address
									if tempItem.Vars != "" {
										bs, _ := json.Marshal(tempItem.Vars)
										json.Unmarshal(bs, &tmpdata.Vars)

									}
									r, _ := json.Marshal(tmpdata)
									log.Logger.Debug("submit data", string(r))
									sendlist := "sms_send:" + tempItem.Account + ":" + sendlist
									redis.Conn.Lpush(sendlist, string(r))
								}
								if err != nil {
									log.Logger.Error("数据解析错误", string(result[1].([]byte)), err)
								}
							} else {
								break
							}
						}
					}()
				}
			}
		}
	}
}

type SMSSendlistFromDB struct {
	Timetosend string `gorm:"column(timetosend)"`
}

type TimetosendItem struct {
	Sendlist string      `gorm:"column(sendlist)" json:"sendlist"`
	Account  string      `gorm:"column(account)" json:"account"`
	Appid    interface{} `gorm:"column(appid)" json:"appid"`
	Project  string      `gorm:"column(project)" json:"project"`
	Address  string      `gorm:"column(address)" json:"address"`
	Send     string      `gorm:"column(send)" json:"send"`
	Vars     interface{} `gorm:"vars" json:"vars"`
}

type TimetosendSql struct {
	Sendlist string `gorm:"column(sendlist)" json:"sendlist"`
	Account  string `gorm:"column(account)" json:"account"`
	Appid    string `gorm:"column(appid)" json:"appid"`
	Project  string `gorm:"column(project)" json:"project"`
	Address  string `gorm:"column(address)" json:"address"`
	Send     string `gorm:"column(send)" json:"send"`
	Vars     string `gorm:"vars" json:"vars"`
}

type tmpSmsMessageStruct struct {
	Address interface{}
	Vars    map[string]interface{}
}

func smsTimetosendHandler(data []byte) {
	sendlist := string(data)
	ch := make(chan int, MaxRoutineNum)
	counter := 0
	pack := []TimetosendSql{}
	for {
		result, err := redis.Conn.Brpop("sms_timetosend_queue:"+sendlist, 1)
		if err == nil {
			tempItem := &TimetosendItem{}
			err = json.Unmarshal(result[1].([]byte), tempItem)
			tempItem.Sendlist = sendlist

			var appid string
			switch tempItem.Appid.(type) {
			case string:
				appid = tempItem.Appid.(string)
			case int:
				appid = strconv.Itoa(tempItem.Appid.(int))
			}
			//bs, _ := json.Marshal(tempItem.Vars)
			s := &TimetosendSql{
				Sendlist: tempItem.Sendlist,
				Account:  tempItem.Account,
				Appid:    appid,
				Project:  tempItem.Project,
				Address:  tempItem.Address,
				Send:     tempItem.Send,
			}
			if tempItem.Vars == "" {
				s.Vars = ""
			} else {
				bs, _ := json.Marshal(tempItem.Vars)
				s.Vars = string(bs)
			}
			if err != nil {
				log.Logger.Error("数据解析错误", string(result[1].([]byte)), err)
			}
			pack = append(pack, *s)
			if len(pack) == 1000 {
				temp := pack
				pack = []TimetosendSql{}
				counter += 1000
				ch <- 1
				go func(ch chan int, temp []TimetosendSql) {
					database.DbQueue.Table("message_send_queue").Create(temp)
					<-ch
				}(ch, temp)
			}
		} else {
			fmt.Println(pack)
			if len(pack) > 0 {
				database.DbQueue.Table("message_send_queue").Create(pack)
				pack = []TimetosendSql{}
			}
			for {
				if len(ch) == 0 {
					return
				}
			}
		}
	}

}

// 彩信定时发送服务----------------------------------------
func mmsOnlineTimetoSendService(ch chan string) {
	signal := true
	log.Logger.Warning("SMS TimeToSend Service Starting")
	for {
		select {
		case <-ch:
			log.Logger.Warning("SMS TimeToSend Service Exiting ...")
			signal = false
			break
		default:
		}
		if !signal {
			break
		}
		result, err := redis.Conn.Brpop("mms_timetosend_task", 30)
		if err == nil {
			log.Logger.Info(string(result[1].([]byte)))
			go mmsTimetosendHandler(result[1].([]byte))
		}
	}
}

func mmsTimetosendHandler(data []byte) {
	sendlist := string(data)
	ch := make(chan int, MaxRoutineNum)
	counter := 0
	pack := []TimetosendSql{}
	for {
		result, err := redis.Conn.Brpop("mms_timetosend_queue:"+sendlist, 1)
		if err == nil {
			tempItem := &TimetosendItem{}
			err = json.Unmarshal(result[1].([]byte), tempItem)
			tempItem.Sendlist = sendlist
			bs, _ := json.Marshal(tempItem.Vars)
			var appid string
			switch tempItem.Appid.(type) {
			case string:
				appid = tempItem.Appid.(string)
			case int:
				appid = strconv.Itoa(tempItem.Appid.(int))
			}
			s := &TimetosendSql{
				Sendlist: tempItem.Sendlist,
				Account:  tempItem.Account,
				Appid:    appid,
				Project:  tempItem.Project,
				Address:  tempItem.Address,
				Send:     tempItem.Send,
				Vars:     string(bs),
			}
			if err != nil {
				log.Logger.Error("数据解析错误", string(result[1].([]byte)), err)
			}
			pack = append(pack, *s)
			if len(pack) == 1000 {
				temp := pack
				pack = []TimetosendSql{}
				counter += 1000
				go func() {
					ch <- 1
					database.DbQueue.Table("mms_send_queue").Create(temp)
					<-ch
				}()
			}
		} else {
			if len(pack) > 0 {
				database.DbQueue.Table("mms_send_queue").Create(pack)
				pack = []TimetosendSql{}
			}
			for {
				if len(ch) == 0 {
					return
				}
			}
		}
	}

}
