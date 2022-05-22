package performance

import (
	"github.com/siddontang/go-log/log"
	"time"
)

type Monitor struct {
	BeginTime time.Time
}

func (h Monitor) Start() Monitor {
	h.BeginTime = time.Now()
	return h
}

func (h Monitor) End() {
	duration := time.Now().Sub(h.BeginTime).Nanoseconds()
	log.Infof("duration nanos:%v", duration)
}

func StartNewMonitor() *Monitor {
	return &Monitor{BeginTime: time.Now()}
}
