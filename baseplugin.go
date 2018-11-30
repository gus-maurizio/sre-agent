package main

import (
	"sre-agent/types"
	"encoding/json"
	"fmt"
        "github.com/google/uuid"
        "github.com/prometheus/client_golang/prometheus"
        //      "log"
        log "github.com/sirupsen/logrus"
	"runtime"
	"time"
)

func baseMeasure() []byte {
	caller := "not available"
	whoami := "not available"

	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		caller = details.Name()
	}

	me, _, _, mok := runtime.Caller(0)
	mydetails := runtime.FuncForPC(me)
	if mok && mydetails != nil {
		whoami = mydetails.Name()
	}
	return([]byte(fmt.Sprintf(`[{"mcaller": "%s", "mwho": "%s", "mtime": %f}]`, caller, whoami, float64(time.Now().UnixNano())/1e9)))
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
		var myMeasure interface{}
		measuredata := measure()
		err := json.Unmarshal(measuredata, &myMeasure)
		if err != nil { log.Fatal("unmarshall err %+v",err) }
        	myModuleContext := &types.ModuleContext{ModuleName: myName, RequestId: uuid.New().String(), TraceId: traceid, RunId: myContext.RunId}
		pluginLogger.WithFields(log.Fields{"mycontext": myModuleContext, "timestamp": float64(t.UnixNano()) / 1e9, "measure": myMeasure}).Info("tick")

		messageMetric.With(prometheus.Labels{"plugin":myName}).Inc()
		bytesMetric.With(prometheus.Labels{"plugin":myName}).Add(float64(len(measuredata)))
	}
        pluginLogger.WithFields(log.Fields{"timestamp": float64(time.Now().UnixNano()) / 1e9}).Info("ended")
}

