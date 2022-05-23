package performance

import (
	"github.com/siddontang/go-log/log"
	"time"
)

type Monitor struct {
	Name      string
	BeginTime time.Time
}

func (h Monitor) Start() Monitor {
	h.BeginTime = time.Now()
	return h
}

func (h Monitor) End() {
	duration := time.Now().Sub(h.BeginTime).Nanoseconds()
	log.Infof("[%v]duration nanos:%v", h.Name, duration)
}

func StartNewMonitor(name string) *Monitor {
	return &Monitor{BeginTime: time.Now(), Name: name}
}
