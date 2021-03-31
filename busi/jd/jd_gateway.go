package jd

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
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	fieldTs        string = "timestamp" // timestamp
	fieldSign      string = "sign"      // sign result
	fieldScreteKey string = "Secret"    // secret key field name
	fieldSignType  string = "signType"  //  sign type
	fieldVersion   string = "version"   //  sign type
	fieldJdOrderNo string = "jdOrderNo" //  jdOrderNo
)

//ResponseMsg struct
type ResponseMsg struct {
	Success   string `json:"isSuccess"`
	ErrorCode string `json:"errorCode"`
}

// JDGateway struct
type JDGateway struct {
	server.Gateway

	JdOrderNo string

	Version string

	SignType string
}

// NewGateway analysis struct from request
func NewGateway(c *gin.Context) (*JDGateway, error) {
	req := c.Request
	tsStr := c.PostForm(fieldTs)
	sign := c.PostForm(fieldSign)
	signType := c.PostForm(fieldSignType)
	version := c.PostForm(fieldVersion)
	jdOrderNo := c.PostForm(fieldJdOrderNo)
	requestIP := req.Header.Get("X-Forwarded-For")
	if requestIP == "" {
		requestIP = strings.Split(req.RemoteAddr, ":")[0]
	}
	// path: /jiange/xxx/xxx  realPath: /xxx/xxx
	var gateway JDGateway
	gateway.RequestIP = requestIP
	gateway.RequestURL = req.URL.Path[7:]
	gateway.Method = req.Method
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	gateway.Ts = ts
	gateway.Sign = sign
	gateway.Version = version
	gateway.SignType = signType
	gateway.JdOrderNo = jdOrderNo
	gateway.AppID = config.Config.JdAppID
	gateway.Context = c
	err = gateway.AnalysisParam()
	if err != nil {
		return nil, err
	}
	return &gateway, nil
}

// Handler deal request and response
func (gateway *JDGateway) Handler() {
	// sign check
	checkRes, err := gateway.SignCheck()
	if err != nil {
		// todo
		gateway.Response400(err)
		return
	}
	if !checkRes {
		msg := new(ResponseMsg)
		msg.Success = "F"
		msg.ErrorCode = constant.JdCheckSignError
		gateway.Response200(msg)
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
func (gateway *JDGateway) FullParams() map[string]string {
	allParams := make(map[string]string, len(gateway.RequestParam))
	for key, val := range gateway.RequestParam {
		if len(val) != 0 && val[0] != "" {
			allParams[key] = val[0]
		}
	}
	delete(allParams, fieldSign)
	delete(allParams, fieldSignType)
	return allParams
}

// httpRequestAndResponse send http request and do response
func (gateway *JDGateway) httpRequestAndResponse() {
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
func (gateway *JDGateway) respSign(data []byte, flag bool) ([]byte, error) {
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
	allParams := make(map[string]string, 6)
	allParams["isSuccess"] = "F"
	allParams["timestamp"] = strconv.FormatInt(gateway.Ts, 10)
	allParams["version"] = gateway.Version
	allParams["jdOrderNo"] = gateway.JdOrderNo
	allParams["agentPrice"] = ""
	if unionResponse.Code == constant.CostPriceError {
		//成本价错误
		allParams["errorCode"] = constant.JdCostPriceError
	} else if unionResponse.Code == constant.ParmeterError {
		//参数错误
		allParams["errorCode"] = constant.JdParameterError
	} else if unionResponse.Code == constant.SubIDError ||
		unionResponse.Code == constant.ProductTypeError {
		//产品信息不存在
		allParams["errorCode"] = constant.JdProductNotExist
	} else if unionResponse.Code == constant.Default &&
		unionResponse.Body.Status == constant.DefaultStatus {
		//处理成功 & 订单已处理
		allParams["isSuccess"] = "T"
		allParams["agentOrderNo"] = unionResponse.Body.ZyOrder
	} else if unionResponse.Code == constant.Default &&
		unionResponse.Body.Status == constant.OrderHasDealed {
		// 查询订单状态
		allParams["isSuccess"] = "T"
		allParams["agentOrderNo"] = unionResponse.Body.ZyOrder
		allParams["quantity"] = "1"
		allParams["status"] = "1"
	} else if unionResponse.Code == constant.OrderNotExist {
		//订单不存在
		allParams["errorCode"] = constant.JdOrderNotExist
		allParams["status"] = "2"
	} else {
		//系统错误
		allParams["errorCode"] = constant.SystemError
	}
	signStrAfter := sign(allParams)
	allParams["sign"] = signStrAfter
	allParams["signType"] = "MD5"
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
