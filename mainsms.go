package mainsms

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"

	"github.com/buger/jsonparser"
)

type SmsSender struct {
	apiKey   string
	projName string
	sender   string
}

func NewSMSSender(apiKey, projName, sender string) *SmsSender {
	return &SmsSender{
		apiKey:   apiKey,
		projName: projName,
		sender:   sender,
	}
}

type Param struct {
	Name string
	Val  string
}

type Params []Param

func (s Params) Len() int {
	return len(s)
}
func (s Params) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Params) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (c *SmsSender) SendSMS(params ...Param) error {
	url := c.makeURL(params...)
	resp, err := http.Get(url)
	if err != nil {
		return errors.New("Ошибка отправки запроса: " + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	value, _, _, err := jsonparser.Get(body, "status")
	if err != nil {
		return errors.New("Ошибка при чтении статуса из ответа: " + err.Error() + ", " + string(body))
	}
	if !bytes.Equal(value, []byte("success")) {
		return errors.New("Ответ от mainsms.ru с ошибкой: " + string(body))
	}
	return nil
}

func (c *SmsSender) makeURL(params ...Param) string {
	var urlname string = "http://mainsms.ru/api/mainsms/message/send?project=" + c.projName + "&sender=" + c.sender + "&"
	for _, param := range params {
		urlname += url.QueryEscape(param.Name) + "=" + url.QueryEscape(param.Val) + "&"
	}
	params = append(params, Param{"project", c.projName}, Param{"sender", c.sender})
	sort.Sort(Params(params))
	var joined string
	for _, v := range params {
		joined += v.Val + ";"
	}
	joined += c.apiKey
	h := sha1.New()
	io.WriteString(h, joined)
	md5h := md5.New()
	io.WriteString(md5h, hex.EncodeToString(h.Sum(nil)))
	sign := hex.EncodeToString(md5h.Sum(nil))
	urlname += "sign=" + sign
	return urlname
}
