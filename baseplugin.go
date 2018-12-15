package main

import (
	"github.com/gus-maurizio/sre-agent/types"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)

func pluginMaker(context types.Context, duration time.Duration, pName string, plugin types.FPlugin, measure  func() ([]uint8, []uint8, float64)) {
        log.WithFields(log.Fields{"duration": duration, "name": pName}).Debug("pluginMaker")
	pRuntime := types.PluginRuntime{Ticker: time.NewTicker(duration), PluginName: pName}
	PluginSlice = append(PluginSlice, pRuntime)
	go plugin(context, pRuntime.PluginName, pRuntime.Ticker, measure)
}

func basePlugin(myContext types.Context, myName string, ticker *time.Ticker, measure types.FuncMeasure) {
	traceid 		:= uuid.New().String()
	pluginLogger 	:= log.WithFields(log.Fields{"pluginname": myName, "context": myContext})
	jsonContext, _ 	:= json.Marshal(myContext)

	pluginLogger.WithFields(log.Fields{"timestamp": float64(time.Now().UnixNano()) / 1e9}).Debug("started")
	defer ticker.Stop()
	
	for t := range ticker.C {
		var myMeasure interface{}
		measuredata, _, mymeasuretime := measure()
		if PluginMap[myName].AlertFunction {
			PluginMap[myName].AlertMsg, PluginMap[myName].AlertLvl, PluginMap[myName].Alert, PluginMap[myName].AlertError = PluginMap[myName].PluginAlert(measuredata)
		}
		// update the measure count and state	
		PluginMap[myName].MeasureCount += 1
		if PluginMap[myName].Alert { 
			PluginMap[myName].AlertCount =+ 1 
		}

		logformat := "{\"timestamp\": %f, \"plugin\": \"%s\", \"measure\": %s, \"context\": %s}\n"
		if PluginMap[myName].MeasureFile {
			fmt.Fprintf(PluginMap[myName].MeasureHandle, logformat, mymeasuretime, myName, measuredata, jsonContext)
		} else {
			fmt.Fprintf(PluginMap[myName].MeasureConn,   logformat, mymeasuretime, myName, measuredata, jsonContext)
		}
		
		err := json.Unmarshal(measuredata, &myMeasure)
		if err != nil { log.Fatal("unmarshall err %+v",err) }
        	myModuleContext := &types.ModuleContext{ModuleName: myName, RequestId: uuid.New().String(), TraceId: traceid, RunId: myContext.RunId}
		// build the ModuleData answer
		myModuleData    := &types.ModuleData{
			RunId: 			myContext.RunId, 
			Timestamp: 		float64(t.UnixNano()) / 1e9,
		 	ModContext: 	*myModuleContext, 
			Measure:		myMeasure,
			TimeOverhead: 	(mymeasuretime - float64(t.UnixNano()) / 1e9) * 1e3,
		} 

		// Good idea to log
		pluginLogger.WithFields(log.Fields{"myModuleData": myModuleData}).Info("tick")
		// Update metrics related to the plugin
		overheadMetric.With(prometheus.Labels{"plugin":myName}).Set(myModuleData.TimeOverhead)
		messageMetric.With(prometheus.Labels{"plugin":myName}).Inc()
		bytesMetric.With(prometheus.Labels{"plugin":myName}).Add(float64(len(measuredata)))
	}
        pluginLogger.WithFields(log.Fields{"timestamp": float64(time.Now().UnixNano()) / 1e9}).Info("ended")
}

