package server

import (
	"errors"
	"io/ioutil"
	"jiange/log"
	"jiange/utils"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	fieldTs        string = "Timestamp" // timestamp
	fieldSign      string = "Sign"      // sign result
	fieldAppID     string = "AppId"     // app id
	fieldScreteKey string = "Secret"    // secret key field name
	fieldURL       string = "UrlPath"   // url path in sign
	fieldPlatform  string = "platform"
)

// Gateway struct
type Gateway struct {

	// RequestIP is original ip
	RequestIP string

	// RequestURL defines the URL to request,like /asset/query
	RequestURL string

	// RequestParam is the param sending to real service supportor
	RequestParam url.Values

	// RequestBody us the body from request
	RequestBody []byte

	// Method is request method:GET,POST,PUT,DELETE
	Method string

	// Ts is timestamp that gateway needs
	Ts int64

	// Sign is sign for all params
	Sign string

	// AppID is request's id
	AppID string

	// RealServiceHost is real service host
	RealServiceHost string

	// Platform
	Platform string

	// channelid
	Channeld string

	// gin context
	Context *gin.Context
}

// NewGateway analysis struct from request
func NewGateway(c *gin.Context) (*Gateway, error) {
	req := c.Request
	tsStr := req.Header.Get(fieldTs)
	sign := req.Header.Get(fieldSign)
	appID := req.Header.Get(fieldAppID)
	platform := c.PostForm(fieldPlatform)
	requestIP := req.Header.Get("X-Forwarded-For")
	if requestIP == "" {
		requestIP = strings.Split(req.RemoteAddr, ":")[0]
	}
	if tsStr == "" || sign == "" || appID == "" || requestIP == "" {
		log.WithFields(log.Fields{
			"tsStr":   tsStr,
			"sign":    sign,
			"appID":   appID,
			"request": requestIP,
		}).Error()
		return nil, errors.New("request error")
	}
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return nil, errors.New("request error")
	}
	// path: /jiange/xxx/xxx  realPath: /xxx/xxx
	gateway := &Gateway{
		RequestIP:  requestIP,
		RequestURL: req.URL.Path[7:],
		Method:     req.Method,
		Ts:         ts,
		Sign:       sign,
		AppID:      appID,
		Context:    c,
		Platform:   platform,
	}
	err = gateway.AnalysisParam()
	if err != nil {
		return nil, err
	}
	return gateway, nil
}

// AnalysisParam param
func (gateway *Gateway) AnalysisParam() error {
	requestInfo := gateway.Context.Request
	// params in url
	urlParams := requestInfo.URL.Query()
	// param in form
	requestInfo.ParseForm()
	formParams := requestInfo.PostForm
	var bodyParams []byte
	var err error
	if requestInfo.Body != nil {
		bodyParams, err = ioutil.ReadAll(requestInfo.Body)
		if err != nil {
			log.WithError(err).Error("body read error")
			return errors.New("read body error")
		}
	}
	switch requestInfo.Method {
	case "GET":
		gateway.RequestParam = urlParams
	case "POST":
		gateway.RequestParam = formParams
		gateway.RequestBody = bodyParams
	default:
		log.WithFields(log.Fields{
			"method": requestInfo.Method,
		}).Error("error request method")
		return errors.New("error request method")
	}
	return nil
}

// Handler deal request and response
func (gateway *Gateway) Handler() {
	// sign check
	checkRes, err := gateway.SignCheck()
	if err != nil {
		// todo
		gateway.Response400(err)
		return
	}
	if !checkRes {
		gateway.Response400(errors.New("illegal sign"))
		return
	}
	// appid and url check
	flag, err := Devingidine(gateway.RequestURL, gateway.AppID)
	if err != nil {
		gateway.Response400(err)
		return
	}
	// query real service host
	realServiceHost, err := QueryRealServiceHost(gateway.RequestURL)
	if err != nil {
		gateway.Response400(err)
		return
	}
	gateway.RealServiceHost = realServiceHost
	// params rule check
	flag, err = gateway.Verify()
	if err != nil {
		gateway.Response400(err)
		return
	}
	if !flag {
		gateway.Response403(errors.New("rule verify fail"))
		return
	}
	// statistics url
	acc, err := SaveCount(gateway.AppID, gateway.RequestURL)
	if err != nil {
		gateway.Response403(errors.New("access count error"))
		return
	}
	if !acc {
		gateway.Response403(errors.New("access limit"))
		return
	}
	// request and response
	gateway.httpRequestAndResponse()
}

// FullParams is method to add gateway's special param,like ts,appID,body
func (gateway *Gateway) FullParams(secretKey string) map[string]string {
	allParams := make(map[string]string, len(gateway.RequestParam)+4)
	for key, val := range gateway.RequestParam {
		allParams[key] = val[0]
	}
	allParams[fieldTs] = strconv.FormatInt(gateway.Ts, 10)
	allParams[fieldAppID] = gateway.AppID
	allParams[fieldScreteKey] = secretKey
	allParams[fieldURL] = "/jiange" + gateway.RequestURL
	return allParams
}

// FullRequestURL get full request url
func (gateway *Gateway) FullRequestURL() string {
	return gateway.RealServiceHost + gateway.RequestURL
}

// httpRequestAndResponse send http request and do response
func (gateway *Gateway) httpRequestAndResponse() {
	var (
		resp *http.Response
		err  error
	)
	req := gateway.Context.Request
	switch gateway.Method {
	case "GET":
		// new sender
		sender := utils.NewHTTPSend(utils.GetURLBuild(gateway.FullRequestURL(), req.URL.Query()))
		// set header
		sender.SetHeader(gateway.GetNewHeader())
		resp, err = sender.Get()
	case "POST":
		// new sender
		sender := utils.NewHTTPSend(gateway.FullRequestURL())
		// set header
		sender.SetHeader(gateway.GetNewHeader())
		// post request only send post form and body params
		if len(gateway.RequestParam) != 0 {
			sender.SetFormV2(gateway.RequestParam)
		} else {
			sender.SetSendType(utils.SendTypeJSON)
			sender.SetBody(gateway.RequestBody)
		}
		resp, err = sender.Post()
	default:
		gateway.Response400(errors.New("error request method"))
	}
	if err != nil {
		gateway.Response400(errors.New("request struct error"))
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		gateway.Response400(errors.New("read body error"))
	} else {
		for key, val := range resp.Header {
			gateway.Context.Header(key, val[0])
		}
		gateway.Context.Data(resp.StatusCode, resp.Header.Get("content-type"), body)
	}
}

// GetNewHeader get header
func (gateway *Gateway) GetNewHeader() map[string]string {
	header := gateway.Context.Request.Header
	header.Del(fieldTs)
	header.Del(fieldSign)
	var newHeader = map[string]string{}
	for key, val := range header {
		newHeader[key] = val[0]
	}
	newHeader["X-Forwarded-For"] = gateway.RequestIP
	return newHeader
}

// Response200 response 200
func (gateway *Gateway) Response200(msg interface{}) {
	gateway.Context.JSON(http.StatusOK, msg)
}

// Response400 is 400 response
func (gateway *Gateway) Response400(err error) {
	gateway.Context.String(http.StatusBadRequest, err.Error())
}

// Response403 is 403 response
func (gateway *Gateway) Response403(err error) {
	gateway.Context.String(http.StatusForbidden, err.Error())
}

// Response404 is 404 response
func (gateway *Gateway) Response404(err error) {
	gateway.Context.String(http.StatusNotFound, err.Error())
}

// Response500 is 500 response
func (gateway *Gateway) Response500(err error) {
	gateway.Context.String(http.StatusInternalServerError, err.Error())
}

// Response502 is 502 response
func (gateway *Gateway) Response502(err error) {
	gateway.Context.String(http.StatusBadGateway, err.Error())
}
