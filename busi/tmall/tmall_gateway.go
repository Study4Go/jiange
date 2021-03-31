package tmall

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
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
	fieldTBOrderNo string = "tbOrderNo" // tbOrderNo
	fieldAppKey    string = "app_key"   // app_key
	fieldTs        string = "timestamp" // timestamp
	fieldSign      string = "sign"      // sign result
	fieldScreteKey string = "Secret"    // secret key field name
)

// TMallOrder tmall order
type TMallOrder struct {
	XMLName              xml.Name
	TbOrderNo            string `xml:"tbOrderNo"`
	CoopOrderSuccessTime string `xml:"coopOrderSuccessTime"`
	CoopOrderStatus      string `xml:"coopOrderStatus"` //商户订单状态
	FailedReason         string `xml:"failedReason"`
	CoopOrderNo          string `xml:"coopOrderNo"` //商户订单号
	FailedCode           string `xml:"failedCode"`
	CoopOrderSnap        string `xml:"coopOrderSnap"` //商户订单快照
}

// TMallGateway struct
type TMallGateway struct {
	server.Gateway
	TimeStamp    string
	TMallOrderID string
}

// NewGateway analysis struct from request
func NewGateway(c *gin.Context) (*TMallGateway, error) {
	req := c.Request
	appKey := c.Query(fieldAppKey)
	tsStr := c.Query(fieldTs)
	sign := c.Query(fieldSign)
	tmallOrderID := c.Query(fieldTBOrderNo)
	requestIP := req.Header.Get("X-Forwarded-For")
	if requestIP == "" {
		requestIP = strings.Split(req.RemoteAddr, ":")[0]
	}
	// path: /jiange/xxx/xxx  realPath: /xxx/xxx
	var gateway TMallGateway
	gateway.RequestIP = requestIP
	gateway.RequestURL = req.URL.Path[7:]
	gateway.Method = req.Method
	gateway.TimeStamp = tsStr
	gateway.Sign = sign
	gateway.AppID = appKey
	gateway.Channeld = config.Config.TMallChanneld
	gateway.TMallOrderID = tmallOrderID
	gateway.Context = c
	err := gateway.AnalysisParam()
	if err != nil {
		return nil, err
	}
	return &gateway, nil
}

// getXMLName get xml name
func getXMLName(requestURL string) string {
	temp := "gamezctoporder"
	if strings.Compare(requestURL, "/vp/tmall/queryOrderStatus") == 0 {
		temp = "gamezctopquery"
	}
	return temp
}

// Handler deal request and response
func (gateway *TMallGateway) Handler() {
	// sign check
	isSuccess, err := gateway.SignCheck()
	if err != nil {
		// todo
		gateway.Response400(err)
		return
	}
	if !isSuccess {
		gateway.Context.Header("Content-Type", "application/xml; charset=GBK")
		gateway.Context.XML(http.StatusOK, TMallOrder{XMLName: xml.Name{Local: getXMLName(gateway.RequestURL)}, TbOrderNo: gateway.TMallOrderID,
			CoopOrderSuccessTime: time.Now().Format(constant.YYYYMMDDHHMMSS),
			CoopOrderStatus:      "GENERAL_ERROR",
			FailedReason:         "sign error", CoopOrderNo: "-1",
			FailedCode: constant.TmallCheckSignError, CoopOrderSnap: "VIP"})
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
func (gateway *TMallGateway) FullParams() map[string]string {
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
func (gateway *TMallGateway) httpRequestAndResponse() {
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
		tmallOrderResp := &TMallOrder{}
		tmallOrderResp.XMLName = xml.Name{Local: getXMLName(gateway.RequestURL)}
		err = gateway.respSign(body, gzip == "gzip", tmallOrderResp)
		if err != nil {
			gateway.Response400(errors.New("respSign body error"))
		}
		gateway.Context.Header("Content-Type", "application/xml; charset=GBK")
		gateway.Context.XML(http.StatusOK, tmallOrderResp)

	}
}

// respSign response sign
func (gateway *TMallGateway) respSign(data []byte, flag bool, tmallOrderResp *TMallOrder) error {
	//check gzip
	if flag {
		gzipData, err := GzipDecode(data)
		if err != nil {
			log.WithFields(log.Fields{"GzipDecode": "msg"}).Error(err)
			return err
		}
		data = gzipData
	}
	unionResponse := &constant.UnionResponse{}
	err := json.Unmarshal(data, &unionResponse)
	if err != nil {
		log.WithFields(log.Fields{"json.Unmarshal": "error"}).Error(err)
		return err
	}
	log.WithFields(log.Fields{}).Info("unionResponse result", *unionResponse)

	tmallOrderResp.TbOrderNo = gateway.TMallOrderID
	tmallOrderResp.CoopOrderSuccessTime = time.Now().Format(constant.YYYYMMDDHHMMSS)
	tmallOrderResp.CoopOrderSnap = "VIP"
	if unionResponse.Code == constant.ParmeterError {
		//参数错误
		tmallOrderResp.CoopOrderStatus = constant.TmallGeneralError
		tmallOrderResp.FailedCode = constant.TmallParameterError
		tmallOrderResp.FailedReason = "parmeter error"
	} else if unionResponse.Code == constant.SubIDError ||
		unionResponse.Code == constant.ProductTypeError {
		//产品信息不存在
		tmallOrderResp.CoopOrderStatus = constant.TmallOrderFalied
		tmallOrderResp.FailedCode = constant.TmallProductNotExist
		tmallOrderResp.FailedReason = "product information does not exist"
	} else if (unionResponse.Code == constant.Default &&
		unionResponse.Body.Status == constant.DefaultStatus) ||
		unionResponse.Body.Status == constant.OrderHasDealed {
		//订单充值成功
		tmallOrderResp.CoopOrderStatus = constant.TmallOrderSuccess
		tmallOrderResp.CoopOrderNo = unionResponse.Body.ZyOrder
	} else if unionResponse.Code == constant.OrderNotExist {
		//订单不存在
		tmallOrderResp.CoopOrderStatus = constant.TamllReuestFalied
		tmallOrderResp.FailedCode = constant.TmallOrderNotExist
		tmallOrderResp.FailedReason = "order info does not exist"
	} else if unionResponse.Code == constant.OrderDealing {
		// 订单处理中
		tmallOrderResp.CoopOrderStatus = constant.TmallUnderWay
		tmallOrderResp.FailedCode = ""
		tmallOrderResp.FailedReason = "order processing"
	} else if unionResponse.Code == constant.OrderError {
		//订单错误
		tmallOrderResp.CoopOrderStatus = constant.TmallOrderFalied
		tmallOrderResp.FailedCode = constant.TmallOrderError
		tmallOrderResp.FailedReason = "order creation error"
	} else {
		//系统错误
		tmallOrderResp.CoopOrderStatus = constant.TmallOrderFalied
		tmallOrderResp.FailedCode = constant.TmallSystemError
		tmallOrderResp.FailedReason = "system error"
	}
	return nil
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
