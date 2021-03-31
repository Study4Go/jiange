package utils

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// send type
const (
	SendTypeForm = "from"
	SendTypeJSON = "json"
)

// HTTPSend is http sender
type HTTPSend struct {
	Link     string
	SendType string
	Header   map[string]string
	Form     map[string]string
	Body     []byte
	sync.RWMutex
}

// NewHTTPSend construct new http sender
func NewHTTPSend(link string) *HTTPSend {
	return &HTTPSend{
		Link:     link,
		SendType: SendTypeForm,
	}
}

// SetForm set http request form
func (h *HTTPSend) SetForm(form map[string]string) {
	h.Lock()
	defer h.Unlock()
	h.Form = form
}

// SetFormV2 set http request form
func (h *HTTPSend) SetFormV2(form map[string][]string) {
	h.Lock()
	defer h.Unlock()
	formMap := make(map[string]string, len(form))
	for key, val := range form {
		formMap[key] = val[0]
	}
	h.Form = formMap
}

// SetBody set http request body
func (h *HTTPSend) SetBody(body []byte) {
	h.Lock()
	defer h.Unlock()
	h.Body = body
}

// SetHeader set http request header
func (h *HTTPSend) SetHeader(header map[string]string) {
	h.Lock()
	defer h.Unlock()
	h.Header = header
}

// SetSendType set http request send type
func (h *HTTPSend) SetSendType(sendType string) {
	h.Lock()
	defer h.Unlock()
	h.SendType = sendType
}

// Get send get http request
func (h *HTTPSend) Get() (*http.Response, error) {
	return h.send("GET")
}

// Post send post http request
func (h *HTTPSend) Post() (*http.Response, error) {
	return h.send("POST")
}

// GetURLBuild build full get url
func GetURLBuild(link string, data url.Values) string {
	u, _ := url.Parse(link)
	q := u.Query()
	for k, v := range data {
		q.Set(k, v[0])
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (h *HTTPSend) send(method string) (*http.Response, error) {
	var (
		req      *http.Request
		client   http.Client
		sendData string
		err      error
	)

	if method == "POST" {
		if strings.ToLower(h.SendType) == SendTypeJSON {
			sendBody, err := json.Marshal(h.Body)
			if err != nil {
				return nil, err
			}
			sendData = string(sendBody)
		} else {
			sendBody := http.Request{}
			sendBody.ParseForm()
			for k, v := range h.Form {
				sendBody.Form.Add(k, v)
			}
			sendData = sendBody.Form.Encode()
		}
	}

	//忽略https的证书
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	req, err = http.NewRequest(method, h.Link, strings.NewReader(sendData))
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	//设置默认header
	if len(h.Header) == 0 {
		//json
		if strings.ToLower(h.SendType) == SendTypeJSON {
			h.Header = map[string]string{
				"Content-Type": "application/json; charset=utf-8",
			}
		} else { //form
			h.Header = map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			}
		}
	}

	for k, v := range h.Header {
		if strings.ToLower(k) == "host" {
			req.Host = v
		} else {
			req.Header.Add(k, v)
		}
	}

	return client.Do(req)
}
