package database

import (
	"fmt"
	"sms_timetosend_task/conf"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DbPublic *gorm.DB
var DbQueue *gorm.DB

func init() {
	publicDBUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", conf.Public.Username, conf.Public.Password, conf.Public.Host, conf.Public.Port, conf.Public.DB)
	queueDBUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", conf.Queue.Username, conf.Queue.Password, conf.Queue.Host, conf.Queue.Port, conf.Queue.DB)

	var err error
	DbPublic, err = gorm.Open(mysql.Open(publicDBUrl), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DbQueue, err = gorm.Open(mysql.Open(queueDBUrl), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}
