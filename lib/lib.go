package lib

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"sms_timetosend_task/log"
	"sort"
	"strconv"
	"strings"
)

const Server = "https://api.mysubmail.com"

func Get(requesturl string, data map[string]string) (string, error) {
	u, _ := url.Parse(requesturl)
	str_list := make([]string, 0, 32)
	for key, val := range data {
		str_list = append(str_list, key+"="+val)
	}
	sigstr := strings.Join(str_list, "&")
	retstr, err := http.Get(u.String() + "?" + sigstr)
	if err != nil {
		return "", err
	}
	result, err := ioutil.ReadAll(retstr.Body)
	retstr.Body.Close()
	if err != nil {
		return "", err
	}
	return string(result), nil
}
func GetWithoutData(requesturl string) string {
	u, _ := url.Parse(requesturl)
	retstr, err := http.Get(u.String())
	if err != nil {
		return err.Error()
	}
	result, err := ioutil.ReadAll(retstr.Body)
	retstr.Body.Close()
	if err != nil {
		return err.Error()
	}
	return string(result)
}
func GetTimestamp() string {
	resp := GetWithoutData(Server + "/service/timestamp")
	var dict map[string]interface{}
	err := json.Unmarshal([]byte(resp), &dict)
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(int(dict["timestamp"].(float64)))
}
func Post(requesturl string, postdata map[string]string) (string, error) {
	var r http.Request
	r.ParseForm()
	// str_list := make([]string, 0, 32)
	for key, val := range postdata {
		// str_list = append(str_list,  key + "=" + val)
		r.Form.Add(key, val)
	}
	// sigstr := strings.Join(str_list, "&")

	//body := bytes.NewBuffer([]byte(sigstr))
	body := strings.NewReader(r.Form.Encode())

	//打印请求体
	log.Logger.Debug("request:%s", r.Form.Encode())

	retstr, err := http.Post(requesturl, "application/x-www-form-urlencoded", body)

	if err != nil {
		return "", err
	}
	result, err := ioutil.ReadAll(retstr.Body)
	retstr.Body.Close()
	if err != nil {
		return "", err
	}
	log.Logger.Debug("request res:%s", string(result))
	return string(result), nil
}

func PostByJson(queryurl string, postdata map[string]string) (string, error) {
	data, err := json.Marshal(postdata)
	if err != nil {
		return "", err
	}
	body := bytes.NewBuffer([]byte(data))

	retstr, err := http.Post(queryurl, "application/json;charset=utf-8", body)

	if err != nil {
		return "", err
	}
	result, err := ioutil.ReadAll(retstr.Body)
	retstr.Body.Close()
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func MultipartPost(requesturl string, postdata map[string]string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range postdata {
		_ = writer.WriteField(key, val)
	}
	err := writer.Close()
	if err != nil {
		return "", err
	}
	contentType := writer.FormDataContentType()
	writer.Close()
	//打印请求体
	//fmt.Println("request:",string(body.Bytes()))

	resp, err := http.Post(requesturl, contentType, body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func CreateSignature(request map[string]string, config map[string]string) string {
	appkey := config["appkey"]
	appid := config["appid"]
	signtype := config["signType"]

	keys := make([]string, 0, 32)
	for key, _ := range request {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	str_list := make([]string, 0, 32)
	for _, key := range keys {
		str_list = append(str_list, key+"="+request[key])
	}
	sigstr := strings.Join(str_list, "&")
	sigstr = appid + appkey + sigstr + appid + appkey
	if signtype == "md5" {
		mymd5 := md5.New()
		io.WriteString(mymd5, sigstr)
		return hex.EncodeToString(mymd5.Sum(nil))
	} else if signtype == "sha1" {
		mysha1 := sha1.New()
		io.WriteString(mysha1, sigstr)
		return hex.EncodeToString(mysha1.Sum(nil))
	} else {
		return appkey
	}
}
