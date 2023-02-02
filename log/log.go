package log

import (
	"fmt"
	"sms_timetosend_task/conf"

	log "github.com/astaxie/beego/logs"
)

var Logger *log.BeeLogger

func init() {
	Logger = log.NewLogger()
	logSetting := fmt.Sprintf(`{"filename":"%s","maxlines":%d,"separate":%s,"daily":%t,"maxdays":%d,"level":%d}`, conf.Log.Filename, conf.Log.Maxlines, conf.Log.Separate, conf.Log.Daily, conf.Log.Maxdays, conf.Log.Level)
	Logger.SetLogger(log.AdapterMultiFile, logSetting)
}
