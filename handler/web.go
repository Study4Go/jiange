package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"jiange/busi/jd"
	"jiange/busi/pdd"
	"jiange/busi/tmall"
	"jiange/config"
	"jiange/log"
	"jiange/server"

	"github.com/gin-gonic/gin"
)

const sleepTime = 30 * time.Second

// Start is web service
func Start() {
	//此处不打印日志,文件就不会创建,下面路由日志就是失败
	log.MainLogger.Info("server start ....")
	gin.DisableConsoleColor()
	// gin.SetMode(gin.ReleaseMode)
	f, _ := os.OpenFile(config.Config.AccLogPath, os.O_RDWR|os.O_APPEND|os.O_SYNC, 0666)
	os.Chmod(config.Config.AccLogPath, 0744)
	gin.DefaultWriter = f

	r := gin.New()
	c := gin.LoggerConfig{
		Output:    f,
		SkipPaths: []string{},
		Formatter: func(param gin.LogFormatterParams) string {
			jsIndent, _ := json.MarshalIndent(param.Request.Header, "", "\t")
			// your custom format
			return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s %s\" \n%s\n",
				param.ClientIP,
				time.Now().Format("2006-01-02 15:04:05"),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.ErrorMessage,
				string(jsIndent),
			)
		},
	}

	r.Use(gin.LoggerWithConfig(c))
	r.Use(gin.Recovery())

	go func(r *gin.Engine) {
		for {
			refreshRouter(r)
			time.Sleep(sleepTime)
		}
	}(r)
	server.InitialNamespace()
	// r.POST("/jiange/vp/third/giftVip", handler)
	// r.Run(fmt.Sprintf("%s:%d", config.Config.Server.IP, config.Config.Server.Port))
	//优雅地重启或停止
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Config.Server.IP, config.Config.Server.Port),
		Handler: r,
	}

	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Error("listen error")
		}
	}()

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.WithFields(log.Fields{}).Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Server Shutdown:")
	}
	log.WithFields(log.Fields{}).Info("Server exiting")

}

func handler(c *gin.Context) {
	//以/vp/third/开始走另外一套签名算法
	log.MainLogger.Info("http start ....")
	if strings.HasPrefix(c.Request.URL.Path[7:], "/vp/third/") {
		gateway, err := jd.NewGateway(c)
		if err != nil {
			log.WithError(err).Error()
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		gateway.Handler()
	} else if strings.HasPrefix(c.Request.URL.Path[7:], "/vp/pdd/") {
		gateway, err := pdd.NewGateway(c)
		if err != nil {
			log.WithError(err).Error()
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		gateway.Handler()
	} else if strings.HasPrefix(c.Request.URL.Path[7:], "/vp/tmall/") {
		gateway, err := tmall.NewGateway(c)
		if err != nil {
			log.WithError(err).Error()
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		gateway.Handler()
	} else {
		// request analysis
		gateway, err := server.NewGateway(c)
		if err != nil {
			log.WithError(err).Error()
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		gateway.Handler()
	}
}

func refreshRouter(r *gin.Engine) {
	defer func() {
		if err := recover(); err != nil {
			log.WithError(errors.New(err.(string))).Error()
		}
	}()
	routeInfo, ok := server.GetURLChange()
	if !ok {
		return
	}
	// initial or refresh
	for _, newURL := range routeInfo.New {
		isNew := true
		for _, oldURL := range routeInfo.Old {
			if newURL == oldURL {
				isNew = false
				break
			}
		}
		// not in old list,need to add to route tree
		if isNew {
			r.Any("/jiange"+newURL, handler)
		}
	}
}
