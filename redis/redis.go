package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"sms_timetosend_task/conf"
	"strconv"
	"strings"
	"time"

	redigo "github.com/gomodule/redigo/redis"
)

type Cache struct {
	p        *redigo.Pool // redis connection pool
	conninfo string
	dbNum    int
	password string
	maxIdle  int
}

var Conn *Cache
var AccountConn *Cache
var SConn *Cache
var DefaultKey string = "mm"

func NewRedisCache() *Cache {
	return &Cache{}
}

func (rc *Cache) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("missing required arguments")
	}
	c := rc.p.Get()
	defer c.Close()

	return c.Do(commandName, args...)
}

func (rc *Cache) Get(key string) interface{} {
	v, err := rc.do("GET", key)
	if err == nil {
		return v
	}
	return nil
}

func (rc *Cache) GetString(key string) string {
	v, err := redigo.String(rc.do("GET", key))
	if err == nil {
		return v
	}
	return ""
}

func (rc *Cache) Rename(old, new string) error {
	_, err := rc.do("RENAME", old, new)
	return err
}

func (rc *Cache) Hget(hashkey, key string) interface{} {
	if v, err := rc.do("HGET", hashkey, key); err == nil {
		return v
	}
	return nil
}

func (rc *Cache) Put(key string, val interface{}, timeout time.Duration) error {
	_, err := rc.do("SETEX", key, int64(timeout/time.Second), val)
	return err
}

func (rc *Cache) Set(key string, val interface{}) error {
	_, err := rc.do("SET", key, val)
	return err
}

func (rc *Cache) IsExist(key string) bool {
	v, err := redigo.Bool(rc.do("EXISTS", key))
	if err != nil {
		return false
	}
	return v
}

func (rc *Cache) Delete(key string) error {
	_, err := rc.do("DEL", key)
	return err
}

func (rc *Cache) Incr(key string) error {
	_, err := redigo.Bool(rc.do("INCRBY", key, 1))
	return err
}

func (rc *Cache) Decr(key string) error {
	_, err := redigo.Bool(rc.do("INCRBY", key, -1))
	return err
}

func (rc *Cache) Lpush(key string, val interface{}) error {
	_, err := rc.do("lpush", key, val)
	return err
}

func (rc *Cache) Llen(key string) int {
	v, _ := rc.do("llen", key)
	var error error
	ret, _ := redigo.Int(v, error)
	return ret
}

func (rc *Cache) Rpush(key string, val interface{}) error {
	_, err := rc.do("rpush", key, val)
	return err
}

func (rc *Cache) Blpop(key string, timeout int) ([]interface{}, error) {
	result, err := redigo.Values(rc.do("blpop", key, timeout))
	return result, err
}

func (rc *Cache) Brpop(key string, timeout int) ([]interface{}, error) {
	result, err := redigo.Values(rc.do("brpop", key, timeout))
	return result, err
}

func (rc *Cache) Sismember(set, key string) bool {
	v, err := redigo.Bool(rc.do("SISMEMBER", set, key))
	if err != nil {
		return false
	}
	return v
}

func (rc *Cache) Smembers(set string) ([]string, error) {
	v, err := redigo.Values(rc.do("SMEMBERS", set))
	if err == nil {
		var val []string
		for _, item := range v {
			itemStr, _ := redigo.String(item, err)
			val = append(val, itemStr)
		}
		return val, err
	}
	return nil, err
}

func (rc *Cache) Sadd(set, key string) error {
	_, err := rc.do("SADD", set, key)
	return err
}

func (rc *Cache) Srem(key, item string) error {
	_, err := rc.do("SREM", key, item)
	return err
}

func (rc *Cache) StartAndGC(config string) error {
	var cf map[string]string
	json.Unmarshal([]byte(config), &cf)

	if _, ok := cf["conn"]; !ok {
		return errors.New("config has no conn key")
	}

	// Format redis://<password>@<host>:<port>
	cf["conn"] = strings.Replace(cf["conn"], "redis://", "", 1)
	if i := strings.Index(cf["conn"], "@"); i > -1 {
		cf["password"] = cf["conn"][0:i]
		cf["conn"] = cf["conn"][i+1:]
	}

	if _, ok := cf["dbNum"]; !ok {
		cf["dbNum"] = "0"
	}
	if _, ok := cf["password"]; !ok {
		cf["password"] = ""
	}
	if _, ok := cf["maxIdle"]; !ok {
		cf["maxIdle"] = "3"
	}
	rc.conninfo = cf["conn"]
	rc.dbNum, _ = strconv.Atoi(cf["dbNum"])
	rc.password = cf["password"]
	rc.maxIdle, _ = strconv.Atoi(cf["maxIdle"])

	rc.connectInit()

	c := rc.p.Get()
	defer c.Close()

	return c.Err()
}

// connect to redis.
func (rc *Cache) connectInit() {
	dialFunc := func() (c redigo.Conn, err error) {
		c, err = redigo.Dial("tcp", rc.conninfo)
		if err != nil {
			return nil, err
		}

		if rc.password != "" {
			if _, err := c.Do("AUTH", rc.password); err != nil {
				c.Close()
				return nil, err
			}
		}

		_, selecterr := c.Do("SELECT", rc.dbNum)
		if selecterr != nil {
			c.Close()
			return nil, selecterr
		}
		return
	}
	// initialize a new pool
	rc.p = &redigo.Pool{
		MaxIdle:     rc.maxIdle,
		IdleTimeout: 180 * time.Second,
		Dial:        dialFunc,
	}
}

func init() {
	redis_auth := ""
	if conf.DefautRedis.Auth != "" {
		redis_auth = `,"password":"` + conf.DefautRedis.Auth + `"`
	}
	redis_config := `{"conn":"` + conf.DefautRedis.Host + `:` + fmt.Sprintf("%d", conf.DefautRedis.Port) + `","dbNum":"0"` + redis_auth + `}`
	Conn = &Cache{}
	err := Conn.StartAndGC(redis_config)
	if err != nil {
		fmt.Println(err)
	}

	account_redis_auth := ""
	if conf.AccountRedis.Auth != "" {
		account_redis_auth = `,"password":"` + conf.AccountRedis.Auth + `"`
	}
	account_redis_config := `{"conn":"` + conf.AccountRedis.Host + `:` + fmt.Sprintf("%d", conf.AccountRedis.Port) + `","dbNum":"0"` + account_redis_auth + `}`
	//fmt.Println(account_redis_config)
	AccountConn = &Cache{}
	err2 := AccountConn.StartAndGC(account_redis_config)
	if err2 != nil {
		fmt.Println(err2)
	}

	s_redis_auth := ""
	if conf.SRedis.Auth != "" {
		s_redis_auth = `,"password":"` + conf.SRedis.Auth + `"`
	}
	s_redis_config := `{"conn":"` + conf.SRedis.Host + `:` + fmt.Sprintf("%d", conf.SRedis.Port) + `","dbNum":"0"` + s_redis_auth + `}`
	//fmt.Println(account_redis_config)
	SConn = &Cache{}
	err3 := SConn.StartAndGC(s_redis_config)
	if err3 != nil {
		fmt.Println(err3)
	}
}
