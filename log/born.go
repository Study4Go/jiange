package log

import "github.com/sirupsen/logrus"

func NewBornHook(fields Fields) *BornHook {
	return &BornHook{fields}
}

type BornHook struct {
	fields Fields
}

func (h *BornHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *BornHook) Fire(entry *logrus.Entry) error {
	for k, v := range h.fields {
		entry.Data[k] = v
	}
	return nil
}
