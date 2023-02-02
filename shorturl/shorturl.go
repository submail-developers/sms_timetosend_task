package shorturl

import (
	lib "sms_timetosend_task/lib"
)

type ShorturlCreate struct {
	appid    string
	appkey   string
	signType string
	group    string
	url      string
	item     string
	domain   string
	tag      string
}

const shorturlCreateURL = "https://service.mysubmail.com/shorturl"

func CreateShorturl(config map[string]string) *ShorturlCreate {
	return &ShorturlCreate{config["appid"], config["appkey"], config["signType"], "", "", "", "", ""}
}
func (this *ShorturlCreate) SetURL(url string) {
	this.url = url
}
func (this *ShorturlCreate) SetItem(item string) {
	this.item = item
}
func (this *ShorturlCreate) SetDomain(domain string) {
	this.domain = domain
}
func (this *ShorturlCreate) SetTag(tag string) {
	this.tag = tag
}
func (this *ShorturlCreate) SetGroup(group string) {
	this.group = group
}
func (this *ShorturlCreate) Create() (string, error) {
	request := make(map[string]string)
	request["appid"] = this.appid
	request["signature"] = this.appkey
	request["url"] = this.url
	if this.group != "" {
		request["group"] = this.group
	}
	if this.item != "" {
		request["item"] = this.item
	}
	if this.domain != "" {
		request["domain"] = this.domain
	}
	if this.tag != "" {
		request["tag"] = this.tag
	}
	return lib.Post(shorturlCreateURL, request)
}
