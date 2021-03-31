package constant

// redis key
const (
	AppKeyRedisKey      string = "app_key_config"      // hash: (appID, secretKey)
	Platform            string = "platform_"           // hash: (platform_appID, platform)
	SupportHostRedisKey string = "service_namespace_"  // string: serviceName -> namespace
	AppUrlsRredisKey    string = "app_urls_redis_key_" // set: app-> url set
	URLRuleRedisKey     string = "url_rule_config"     // hash: (url, rule)
	URLVersionRedisKey  string = "url_version"         // string: version
	URLListRedisKey     string = "url_list"            // set: all url list
	URLAccessCountKey   string = "url_access_count_"   // hash:(appID, count)
	URLAccessLimitKey   string = "url_access_limit_"   // hash:(day or month, count)
)

// time
const (
	TimeLayout     string = "2006-01-02 15:04:05" // time format layout
	YYYYMMDDHHMMSS string = "20060102150405"
)

// zhangyue vip error code
const (
	ParmeterError    int = 20099 // 参数错误
	OrderDealing     int = 30001 //订单处理中
	SubIDError       int = 30002 //档位信息错误
	OrderError       int = 30004 //订单号生成错误
	OrderHasDealed   int = 30006 //订单已处理
	CostPriceError   int = 30010 // 成本价错误
	ProductTypeError int = 30011 // 产品类型错误
	OrderNotExist    int = 30012 // 订单不存在
	Default          int = 0
	DefaultStatus    int = 1
)

// jd error code
const (
	JdParameterError  string = "JDI_00001" //参数不正确
	JdCheckSignError  string = "JDI_00002" //签名失败
	JdProductNotExist string = "JDI_00003" //没有对应商品
	JdCostPriceError  string = "JDI_00005" //成本价不正确
	JdOrderNotExist   string = "JDI_00007" //没有对应订单
	SystemError       string = "JDI_00010" //系统错误
)

// pdd error code
const (
	PddOrderSuccess    int = 0  //订单成功
	PddParameterError  int = 13 //参数不正确
	PddCheckSignError  int = 12 //签名失败
	PddProductNotExist int = 15 //没有对应商品
	PddOrderNotExist   int = 16 //订单不存在
	PddOrderHasDealed  int = 17 //订单已存在
	PddOrderError      int = 18 //订单错误
	PddSystemError     int = -1 //系统错误
)

// tmall error code
const (
	TmallParameterError  string = "0101" //参数不正确
	TmallCheckSignError  string = "0102" //签名失败
	TmallProductNotExist string = "0305" //没有对应商品
	TmallOrderNotExist   string = "0104" //订单不存在
	TmallOrderError      string = "0503" //订单错误
	TmallSystemError     string = "9999" //系统错误
	TmallOrderFalied     string = "ORDER_FAILED"
	TamllReuestFalied    string = "REQUEST_FAILED"
	TmallOrderSuccess    string = "SUCCESS" //订单成功(订单已处理)
	TmallGeneralError    string = "GENERAL_ERROR"
	TmallUnderWay        string = "UNDERWAY" //订单处理中
)

// UnionResponse vip& assets response
type UnionResponse struct {
	Code int       `json:"code"`
	Body UnionResp `json:"body"`
	Msg  string    `json:"msg"`
}

// UnionResp unionResp
type UnionResp struct {
	OutOrder string `json:"outOrder"`
	ZyOrder  string `json:"zyOrder"`
	Status   int    `json:"status"`
}
