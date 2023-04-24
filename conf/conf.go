package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	config "github.com/astaxie/beego/config"
)

type Database struct {
	Host     string
	Port     int
	Username string
	Password string
	DB       string
}

type Redis struct {
	Host string
	Port int
	Auth string
}

type Logs struct {
	Filename string
	Maxlines int
	Daily    bool
	Maxdays  int
	Level    int
	Separate string
}

var Public *Database
var Te *Database
var Queue *Database
var DefautRedis *Redis
var AccountRedis *Redis
var SRedis *Redis
var Log *Logs
var DefautSpeed chan int
var KAFKA_HOST string

var MailConfig map[string]string
var SmsConfig map[string]string

func init() {
	Reload()
}

func Reload() {
	cf, err := config.NewConfig("ini", GetCurrentDirectory()+"/config/app.conf")
	if err != nil {
		panic(err)
	}

	KAFKA_HOST = cf.String("KAFKA::host")
	PublicPort, _ := cf.Int("DBPublic::port")
	teport, _ := cf.Int("DBTe::port")
	queueport, _ := cf.Int("DBQueue::port")
	redisport, _ := cf.Int("Redis::port")
	AccountRedisPort, _ := cf.Int("AccountRedis::port")
	AccountRedis = &Redis{cf.String("AccountRedis::host"), AccountRedisPort, cf.String("AccountRedis::auth")}

	SRedisPort, _ := cf.Int("SRedis::port")
	SRedis = &Redis{cf.String("SRedis::host"), SRedisPort, cf.String("SRedis::auth")}

	Public = &Database{cf.String("DBPublic::host"), PublicPort, cf.String("DBPublic::username"), cf.String("DBPublic::password"), cf.String("DBPublic::database")}
	Te = &Database{cf.String("DBTe::host"), teport, cf.String("DBTe::username"), cf.String("DBTe::password"), cf.String("DBTe::database")}
	Queue = &Database{cf.String("DBQueue::host"), queueport, cf.String("DBQueue::username"), cf.String("DBQueue::password"), cf.String("DBQueue::database")}

	DefautRedis = &Redis{cf.String("Redis::host"), redisport, cf.String("Redis::auth")}
	logMaxlines, _ := cf.Int("log::maxlines")
	logMaxdays, _ := cf.Int("log::maxdays")
	logDaily, _ := cf.Bool("log::daily")
	logLevel, _ := cf.Int("log::level")
	Log = &Logs{GetCurrentDirectory() + "/" + cf.String("log::filename"), logMaxlines, logDaily, logMaxdays, logLevel, cf.String("log::separate")}
	speed, _ := cf.Int("chn::speed")
	DefautSpeed = make(chan int, speed)

	MailConfig = make(map[string]string)
	MailConfig["appid"] = cf.String("mailconfig::appid")
	MailConfig["appkey"] = cf.String("mailconfig::appkey")
	MailConfig["signType"] = cf.String("mailconfig::signtype")
	SmsConfig = make(map[string]string)
	SmsConfig["appid"] = cf.String("smsconfig::appid")
	SmsConfig["appkey"] = cf.String("smsconfig::appkey")
	SmsConfig["signType"] = cf.String("smsconfig::signtype")
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}
