package log

import (
	"fmt"
)

type MsgbusLogger struct {
	defaultFields Fields
}

func NewMsgbusLogger(defaultFields Fields) *MsgbusLogger {
	return &MsgbusLogger{defaultFields}
}

func NewDefaultMsgbusLogger() *MsgbusLogger {
	return &MsgbusLogger{Fields{}}
}

func (l *MsgbusLogger) Print(v ...interface{}) {
	s := fmt.Sprint(v...)
	WithFields(l.defaultFields).WithField("_type", "msgbus").Warn(s)
}

func (l *MsgbusLogger) Printf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	WithFields(l.defaultFields).WithField("_type", "msgbus").Warn(s)
}

func (l *MsgbusLogger) Println(v ...interface{}) {
	s := fmt.Sprintln(v...)
	WithFields(l.defaultFields).WithField("_type", "msgbus").Warn(s)
}
