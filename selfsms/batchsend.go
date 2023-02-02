package selfsms

import (
	"encoding/json"
	lib "sms_timetosend_task/lib"
)

type batchsend struct {
	appid    string
	appkey   string
	signType string
	content  string
	multi    []map[string]interface{}
	tag      string
}

const batchsendURL = lib.Server + "/batchsms/multisend"

func CreateBatchSend(config map[string]string) *batchsend {
	return &batchsend{config["appid"], config["appkey"], config["signType"], "", nil, ""}
}

func (this *batchsend) SetContent(content string) {
	this.content = content
}

func (this *batchsend) AddMulti(multi map[string]interface{}) {
	this.multi = append(this.multi, multi)
}

func (this *batchsend) SetTag(tag string) {
	this.tag = tag
}

func (this *batchsend) BatchSend() (string, error) {
	config := make(map[string]string)
	config["appid"] = this.appid
	config["appkey"] = this.appkey
	config["signType"] = this.signType

	request := make(map[string]string)
	request["appid"] = this.appid
	request["content"] = this.content
	if this.signType != "normal" {
		request["sign_type"] = this.signType
		request["timestamp"] = lib.GetTimestamp()
		request["sign_version"] = "2"
	}
	if this.tag != "" {
		request["tag"] = this.tag
	}
	request["signature"] = lib.CreateSignature(request, config)
	//v2 数字签名 multi 不参与计算

	data, err := json.Marshal(this.multi)
	if err == nil {
		request["multi"] = string(data)
	}

	return lib.Post(batchsendURL, request)
}
