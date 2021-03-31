package log

import (
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

func NewStackHook(levels []logrus.Level) *StackHook {
	return &StackHook{
		file:     true,
		line:     true,
		function: true,
		levels:   levels,
	}
}

type StackHook struct {
	file     bool
	line     bool
	function bool
	levels   []logrus.Level
}

func (h *StackHook) Levels() []logrus.Level {
	return h.levels
}

func (h *StackHook) Fire(entry *logrus.Entry) error {
	pc := make([]uintptr, 64)
	cnt := runtime.Callers(skipFrameCnt, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i])
		name := fu.Name()
		if !strings.Contains(name, "github.com/sirupsen/logrus") {
			file, line := fu.FileLine(pc[i] - 1)
			if h.file {
				entry.Data["file"] = path.Base(file)
			}
			if h.function {
				entry.Data["func"] = path.Base(name)
			}
			if h.line {
				entry.Data["line"] = line
			}
			break
		}
	}

	return nil
}
