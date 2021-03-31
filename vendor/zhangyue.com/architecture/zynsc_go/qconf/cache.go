//带缓存的Qconf
package qconf

import (
	"reflect"
	"sync"
	"time"

	"zhangyue.com/architecture/zynsc_go/utils"
)

type cacheStruct struct {
	items map[string]*item
}

type item struct {
	value interface{}
	time  time.Time
}

var rw sync.RWMutex

//缓存qconf信息
var qonfCache cacheStruct

const (
	cacheExpired = time.Minute //缓存有效期60s
)

//对全局变量qonfCache做初始化
func init() {
	if qonfCache.items == nil {
		qonfCache.items = make(map[string]*item)
	}
}

//获取qconf的path信息具有缓存
//path 是zookeeper的路径
//idc qconf中的环境配置
func GetConfByCache(path, idc string) (string, error) {
	val, err := commonCall(path, idc, "GetConf", reflect.ValueOf(GetConf))
	if err != nil {
		return "", err
	}
	return utils.String(val)
}

//批量获取qconf的path下的key
//由于interface是引用类型，直接返回ArrStr，如果外部有修改会导致缓存的结果被改变
func GetBatchKeysByCache(path string, idc string) ([]string, error) {
	val, err := commonCall(path, idc, "GetBatchKeys", reflect.ValueOf(GetBatchKeys))
	if err != nil {
		return nil, err
	}

	ArrStr, err := utils.ArrayString(val)
	if err != nil {
		return nil, err
	}
	ret := make([]string, len(ArrStr))
	for _, v := range ArrStr {
		ret = append(ret, v)
	}
	return ret, nil
}

//批量获取path信息下的信息
//由于interface是引用类型，直接返回mapStrStr，如果外部有修改会导致缓存的结果被改变
func GetBatchConfByCache(path string, idc string) (map[string]string, error) {
	val, err := commonCall(path, idc, "GetBatchConf", reflect.ValueOf(GetBatchConf))

	if err != nil {
		return nil, err
	}

	mapStrStr, err := utils.MapStringString(val)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]string)
	for k, v := range mapStrStr {
		ret[k] = v
	}
	return ret, nil
}

//统一函调Qconf里面的函数
//由于函数的类型不一致，这里使用反射统一调用
func commonCall(path, idc, funName string, fun reflect.Value) (interface{}, error) {
	//key是路径+环境+请求的方法
	key := path + "_" + idc + "_" + funName
	now := time.Now()
	rw.RLock()
	it, keyIsExist := qonfCache.items[key]
	rw.RUnlock()
	//key存在，并且值是有效的，并且在有效期内的
	if keyIsExist {
		if it.value != nil && cacheExpired >= now.Sub(it.time) {
			return it.value, nil
		}
	}

	//去请求qconf对应的方法，获取数据
	params := []string{path, idc}
	//封装入参
	in := make([]reflect.Value, 2)
	for i := range in {
		in[i] = reflect.ValueOf(params[i])
	}
	//回调函数的返回值
	out := fun.Call(in)
	//qconf的第一个返回值
	val := out[0].Interface()
	//qconf的第二个返回值，error类型
	err := out[1].Interface()

	if err != nil {
		return val, err.(error)
	}

	//是否存在key
	if keyIsExist {
		it.value = val
		it.time = now
	} else {

		it = new(item)
		it.value = val
		it.time = now
		//由于key对应的值还不存在，使用全局锁
		rw.Lock()
		qonfCache.items[key] = it
		rw.Unlock()
	}

	return it.value, nil
}
