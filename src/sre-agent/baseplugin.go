package main

import (
	"sre-agent/types"
        "github.com/google/uuid"
        "github.com/prometheus/client_golang/prometheus"
	"log"
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
	logrecord := p.Sprintf("PluginMaker context [%#v] with duration %v name: %s and function %#v with function_measure %#v\n", context, duration, pName, plugin, measure)
	log.Print(logrecord)
	pRuntime := types.PluginRuntime{Ticker: time.NewTicker(duration), PluginName: pName}
	pContext := context
	PluginSlice = append(PluginSlice, pRuntime)
	go plugin(pContext, pRuntime.PluginName, pRuntime.Ticker, measure)
}

func basePlugin(myParentContext types.Context, myName string, ticker *time.Ticker, measure types.FuncMeasure) {
	// make sure we Stop at end
	myContext := myParentContext
        myContext.TraceId  = uuid.New().String()
	defer ticker.Stop()
	log.Printf("%s started context: [%v]\n", myName, myContext)
	for t := range ticker.C {
		myContext.Timestamp = float64(t.UTC().UnixNano())/1e9
        	myContext.RequestId = uuid.New().String()
		myContext.ParentId  = ""
		myMeasure := measure()
		log.Printf("%20s Tick at %f measure: [%#v] context: [%#v]\n", myName, float64(t.UTC().UnixNano())/1e9, myMeasure, myContext)
		messageMetric.With(prometheus.Labels{"plugin":myName}).Inc()
		bytesMetric.With(prometheus.Labels{"plugin":myName}).Add(float64(len(myMeasure)))
	}
}

