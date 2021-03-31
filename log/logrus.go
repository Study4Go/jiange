package log

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

const skipFrameCnt = 3

var (
	WithField = logrus.WithField
	WithError = logrus.WithError
)

type Fields map[string]interface{}

func WithFields(fields Fields) *logrus.Entry {
	return logrus.WithFields(logrus.Fields(fields))
}

func checkFile(filePath string) error {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func InitPath(filePath string, format string, levelName string, defaultFields Fields) error {
	level, err := logrus.ParseLevel(levelName)
	if err != nil {
		return err
	}

	if err := checkFile(filePath); err != nil {
		return err
	}

	writer, err := rotatelogs.New(
		filePath+".%Y%m%d",
		rotatelogs.WithLinkName(filePath),
		rotatelogs.WithMaxAge(time.Duration(24*30)*time.Hour),         // 保留30天
		rotatelogs.WithRotationTime(time.Duration(86400)*time.Second), // 每天滚一次
	)

	if err != nil {
		return err
	}

	// 设置日志输出格式
	var formatter logrus.Formatter
	if format == "json" {
		formatter = &logrus.JSONFormatter{}
	} else {
		formatter = &logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05.000"}
	}

	// 设置文件输出路径
	writerMap := lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
	}

	logrus.SetLevel(level)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(formatter)
	// 在日志里增加文件名和行号
	logrus.AddHook(NewStackHook(logrus.AllLevels))
	// 默认字段
	if len(defaultFields) > 0 {
		logrus.AddHook(NewBornHook(defaultFields))
	}
	// 加上文件输出
	logrus.AddHook(lfshook.NewHook(writerMap, formatter))
	return nil
}

func Silent() {
	logrus.SetOutput(ioutil.Discard)
}
