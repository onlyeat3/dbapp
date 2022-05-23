package performance

import (
	"github.com/siddontang/go-log/log"
	"time"
)

type Monitor struct {
	Name      string
	BeginTime time.Time
	timeUnit  time.Duration
}

func (h Monitor) Start() Monitor {
	h.BeginTime = time.Now()
	return h
}

func (h Monitor) End() {
	duration := time.Now().Sub(h.BeginTime)
	durationInTimeUnit := duration.Nanoseconds()
	timeUnitStr := "Nanosecond"
	switch h.timeUnit {
	case time.Millisecond:
		durationInTimeUnit = duration.Milliseconds()
		timeUnitStr = "Millisecond"
		break
	case time.Microsecond:
		timeUnitStr = "Microsecond"
		durationInTimeUnit = duration.Microseconds()
	}
	log.Infof("[%v]duration %s:%v", h.Name, timeUnitStr, durationInTimeUnit)
}

func StartNewMonitor(name string) *Monitor {
	return &Monitor{BeginTime: time.Now(), Name: name}
}

func StartNewMonitorWithTimeUnit(name string, timeUnit time.Duration) *Monitor {
	return &Monitor{BeginTime: time.Now(), Name: name, timeUnit: timeUnit}
}
