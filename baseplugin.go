package main

import (
	"agent/types"
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

func pluginMaker(duration time.Duration, pName string, plugin types.FuncPlugin, measure types.FuncMeasure) {
	logrecord := p.Sprintf("PluginMaker with duration %v name: %s and function %#v with function_measure %#v\n", duration, pName, plugin, measure)
	log.Print(logrecord)
	pRuntime := types.PluginRuntime{Ticker: time.NewTicker(duration), PluginName: pName}
	PluginSlice = append(PluginSlice, pRuntime)
	go plugin(pRuntime.PluginName, pRuntime.Ticker, measure)
}

func basePlugin(myName string, ticker *time.Ticker, measure types.FuncMeasure) {
	// make sure we Stop at end
	defer ticker.Stop()
	log.Printf("%s started", myName)
	for t := range ticker.C {
		myMeasure := measure()
		log.Printf("%20s Tick at %f measure: [%v]\n", myName, float64(t.UTC().UnixNano())/1e9, myMeasure)
		messageMetric.With(prometheus.Labels{"plugin":myName}).Inc()
		bytesMetric.With(prometheus.Labels{"plugin":myName}).Add(float64(len(myMeasure)))
	}
}

