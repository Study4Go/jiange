package pdd

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io/ioutil"
	"jiange/config"
	"jiange/constant"
	"jiange/log"
	"jiange/server"
	"jiange/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	fieldTs        string = "timestamp" // timestamp
	fieldSign      string = "sign"      // sign result
	fieldScreteKey string = "key"       // secret key field name
	fieldSignType  string = "signType"  //  sign type
	fieldVersion   string = "version"   //  sign type
)

//ResponseMsg struct
type ResponseMsg struct {
	ResultMsg  string      `json:"resultMsg"`
	ResultCode int         `json:"resultCode"`
	ResultData interface{} `json:"resultData"`
}

// PddGateway struct
type PddGateway struct {
	server.Gateway

	Version string

	SignType string
}

// NewGateway analysis struct from request
func NewGateway(c *gin.Context) (*PddGateway, error) {
	req := c.Request
	sign := c.Query(fieldSign)
	signType := c.Query(fieldSignType)
	requestIP := req.Header.Get("X-Forwarded-For")
	if requestIP == "" {
		requestIP = strings.Split(req.RemoteAddr, ":")[0]
	}
	// path: /jiange/xxx/xxx  realPath: /xxx/xxx

	var gateway PddGateway
	gateway.RequestIP = requestIP
	gateway.RequestURL = req.URL.Path[7:]
	gateway.Method = req.Method
	gateway.Sign = sign
	gateway.SignType = signType
	gateway.AppID = config.Config.PddAppID
	gateway.Channeld = config.Config.Channeld
	gateway.Context = c
	err := gateway.AnalysisParam()
	if err != nil {
		return nil, err
	}
	return &gateway, nil
}

// Handler deal request and response
func (gateway *PddGateway) Handler() {
	// sign check
	checkRes, err := gateway.SignCheck()
	if err != nil {
		// todo
		gateway.Response400(err)
		return
	}
	if !checkRes {
		msg := new(ResponseMsg)
		msg.ResultCode = constant.PddCheckSignError
		msg.ResultMsg = "签名失败"
		msg.ResultData = nil
		gateway.Response200(msg)
		return
	}
	// appid and url check
	flag, err := server.Devingidine(gateway.RequestURL, gateway.AppID)
	if err != nil {
		gateway.Response400(err)
		return
	}
	if !flag {
		gateway.Response403(errors.New("rule verify fail"))
		return
	}
	// query real service host
	realServiceHost, err := server.QueryRealServiceHost(gateway.RequestURL)
	if err != nil {
		gateway.Response400(err)
		return
	}
	gateway.RealServiceHost = realServiceHost
	// statistics url
	acc, err := server.SaveCount(gateway.AppID, gateway.RequestURL)
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
func (gateway *PddGateway) FullParams() map[string]string {
	allParams := make(map[string]string, len(gateway.RequestParam))
	for key, val := range gateway.RequestParam {
		if len(val) != 0 && val[0] != "" {
			allParams[key] = val[0]
		}
	}
	delete(allParams, fieldSign)
	return allParams
}

// httpRequestAndResponse send http request and do response
func (gateway *PddGateway) httpRequestAndResponse() {
	var (
		resp *http.Response
		err  error
	)
	req := gateway.Context.Request
	switch gateway.Method {
	case "GET":
		// new sender
		values := req.URL.Query()
		//request add appid
		values.Set("appId", gateway.AppID)
		values.Set("channeld", gateway.Channeld)
		sender := utils.NewHTTPSend(utils.GetURLBuild(gateway.FullRequestURL(), values))
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
		gzip := resp.Header.Get("Content-Encoding")
		//response sign
		body, err = gateway.respSign(body, gzip == "gzip")
		if err != nil {
			gateway.Response400(errors.New("respSign body error"))
		}
		gateway.Context.Data(resp.StatusCode, resp.Header.Get("content-type"), body)
	}
}

// respSign response sign
func (gateway *PddGateway) respSign(data []byte, flag bool) ([]byte, error) {
	//check gzip
	if flag {
		gzipData, err := GzipDecode(data)
		if err != nil {
			log.WithFields(log.Fields{"GzipDecode": "msg"}).Error(err)
			return nil, err
		}
		data = gzipData
	}
	unionResponse := &constant.UnionResponse{}
	err := json.Unmarshal(data, &unionResponse)
	if err != nil {
		log.WithFields(log.Fields{"json.Unmarshal": "error"}).Error(err)
		return nil, err
	}
	log.WithFields(log.Fields{}).Info("unionResponse result", *unionResponse)
	allParams := make(map[string]interface{}, 6)
	allParams["resultCode"] = constant.PddOrderSuccess
	allParams["resultMsg"] = unionResponse.Msg
	if unionResponse.Code == constant.ParmeterError {
		//参数错误
		allParams["resultCode"] = constant.PddParameterError
	} else if unionResponse.Code == constant.SubIDError ||
		unionResponse.Code == constant.ProductTypeError {
		//产品信息不存在
		allParams["errorCode"] = constant.PddProductNotExist
	} else if unionResponse.Code == constant.Default &&
		unionResponse.Body.Status == constant.DefaultStatus {
		//处理成功 & 订单已处理
		secParams := make(map[string]interface{}, 6)
		secParams["status"] = "ACCEPT"
		secParams["outOrderNo"] = unionResponse.Body.OutOrder
		secParams["orderNo"] = unionResponse.Body.ZyOrder
		secParams["createTime"] = time.Now().UnixNano() / 1e6
		allParams["resultData"] = secParams

	} else if unionResponse.Code == constant.OrderNotExist {
		//订单不存在
		allParams["resultCode"] = constant.PddOrderNotExist
	} else if unionResponse.Code == constant.OrderDealing ||
		unionResponse.Code == constant.OrderHasDealed {
		// 订单处理中 订单已处理
		allParams["resultCode"] = constant.PddOrderHasDealed
	} else if unionResponse.Code == constant.OrderError {
		//订单错误
		allParams["resultCode"] = constant.PddOrderError
	} else {
		//系统错误
		allParams["resultCode"] = constant.PddSystemError
	}
	jsonByte, err := json.Marshal(allParams)
	if err != nil {
		log.WithFields(log.Fields{"json marshal": "error"}).Error(err)
		return nil, err
	}
	if !flag {
		return jsonByte, nil
	}
	return GzipEncode(jsonByte)
}

// GzipEncode encode
func GzipEncode(in []byte) ([]byte, error) {
	var (
		buffer bytes.Buffer
		out    []byte
		err    error
	)
	writer := gzip.NewWriter(&buffer)
	_, err = writer.Write(in)
	if err != nil {
		writer.Close()
		return out, err
	}
	err = writer.Close()
	if err != nil {
		return out, err
	}
	return buffer.Bytes(), nil
}

// GzipDecode decode
func GzipDecode(in []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		var out []byte
		return out, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}
