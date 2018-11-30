package main

import (
	"sre-agent/types"
        "github.com/google/uuid"
        "github.com/prometheus/client_golang/prometheus"
        //      "log"
        log "github.com/sirupsen/logrus"
	"runtime"
	"time"
)

func baseMeasure() string {
	caller := "not available"
	whoami := "not available"

	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		caller = details.Name()
	}

	me, _, _, ok := runtime.Caller(0)
	mydetails := runtime.FuncForPC(me)
	if ok && mydetails != nil {
		whoami = mydetails.Name()
	}
	return (p.Sprintf("sample %20s called by %20s at %f", whoami, caller, float64(time.Now().UnixNano())/1e9))
}

func pluginMaker(context types.Context, duration time.Duration, pName string, plugin types.FPlugin, measure types.FuncMeasure) {
	pRuntime := types.PluginRuntime{Ticker: time.NewTicker(duration), PluginName: pName}
	PluginSlice = append(PluginSlice, pRuntime)
	go plugin(context, pRuntime.PluginName, pRuntime.Ticker, measure)
}

func basePlugin(myContext types.Context, myName string, ticker *time.Ticker, measure types.FuncMeasure) {
	traceid := uuid.New().String()
        pluginLogger := log.WithFields(log.Fields{"pluginname": myName, "context": myContext})
        pluginLogger.WithFields(log.Fields{"timestamp": float64(time.Now().UnixNano()) / 1e9}).Debug("started")
	defer ticker.Stop()
	for t := range ticker.C {
		myMeasure := measure()
        	myModuleContext := &types.ModuleContext{RequestId: uuid.New().String(), TraceId: traceid}
		pluginLogger.WithFields(log.Fields{"mycontext": myModuleContext, "timestamp": float64(t.UnixNano()) / 1e9, "measure": myMeasure}).Info("tick")

		messageMetric.With(prometheus.Labels{"plugin":myName}).Inc()
		bytesMetric.With(prometheus.Labels{"plugin":myName}).Add(float64(len(myMeasure)))
	}
        pluginLogger.WithFields(log.Fields{"timestamp": float64(time.Now().UnixNano()) / 1e9}).Info("ended")
}

