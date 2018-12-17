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

func pluginMaker(context types.Context, tick *time.Ticker, pName string, plugin types.FPlugin, measure  func() ([]uint8, []uint8, float64)) {
	go plugin(context, pName, tick, measure)
}

func basePlugin(myContext types.Context, myName string, ticker *time.Ticker, measure types.FuncMeasure) {
	traceid 		:= uuid.New().String()
	pluginLogger 	:= log.WithFields(log.Fields{"pluginname": myName, "context": myContext})
	jsonContext, _ 	:= json.Marshal(myContext)

	pluginLogger.WithFields(log.Fields{"timestamp": float64(time.Now().UnixNano()) / 1e9}).Debug("started")
	defer ticker.Stop()
	
	for t := range ticker.C {
		var myMeasure 	interface{}

		// Just in case there is no Alert function defined, initialize to all is ok
		MapPlugState[myName].AlertMsg	= ""
		MapPlugState[myName].AlertLvl	= ""
		MapPlugState[myName].Alert 		= false
		MapPlugState[myName].Warning 	= false
		
		MapPlugState[myName].AlertError = "n/a"

		// Now do the measurements
		measuredata, _, mymeasuretime := measure()

		if MapPlugState[myName].AlertFunction {
			var myerr error
			MapPlugState[myName].AlertMsg, MapPlugState[myName].AlertLvl, MapPlugState[myName].Alert, myerr = MapPlugState[myName].PluginAlert(measuredata)
			MapPlugState[myName].AlertError = myerr.Error()
		}
		// update the measure count and state, make sure it does not go beyond limits
		MapPlugState[myName].MeasureCount += 1
		if MapPlugState[myName].MeasureCount == 2147483647 {MapPlugState[myName].MeasureCount = 0}
		// Did we get an alert
		if MapPlugState[myName].Alert {
			alertformat := "{\"timestamp\": %f, \"plugin\": \"%s\", \"alertmsg\": %s, \"alertlvl\": %s, \"error\": %s, \"measure\": %s, \"context\": %s}\n"
			if MapPlugState[myName].AlertLvl == "warn" {
				// it is a warning, so clear the alert flag and post to warning
				MapPlugState[myName].Alert 		= false
				MapPlugState[myName].Warning 	= true
				MapPlugState[myName].WarnCount  += 1
				if MapPlugState[myName].WarnFile {
					fmt.Fprintf(MapPlugState[myName].WarnHandle, alertformat, mymeasuretime, myName, MapPlugState[myName].AlertMsg, 
								MapPlugState[myName].AlertLvl, MapPlugState[myName].AlertError, measuredata, jsonContext)
				} else {
					fmt.Fprintf(MapPlugState[myName].WarnConn, alertformat, mymeasuretime, myName, MapPlugState[myName].AlertMsg,
								MapPlugState[myName].AlertLvl, MapPlugState[myName].AlertError, measuredata, jsonContext)
				}
			} else {
				// this is a real alert, so post to alert
				MapPlugState[myName].AlertCount += 1
				if MapPlugState[myName].AlertFile {
					fmt.Fprintf(MapPlugState[myName].AlertHandle, alertformat, mymeasuretime, myName, MapPlugState[myName].AlertMsg, 
								MapPlugState[myName].AlertLvl, MapPlugState[myName].AlertError, measuredata, jsonContext)
				} else {
					fmt.Fprintf(MapPlugState[myName].AlertConn, alertformat, mymeasuretime, myName, MapPlugState[myName].AlertMsg,
								MapPlugState[myName].AlertLvl, MapPlugState[myName].AlertError, measuredata, jsonContext)
				}
			}
		}
		// Time to send to measure destination
		logformat := "{\"timestamp\": %f, \"plugin\": \"%s\", \"measure\": %s, \"context\": %s}\n"
		if MapPlugState[myName].MeasureFile {
			fmt.Fprintf(MapPlugState[myName].MeasureHandle, logformat, mymeasuretime, myName, measuredata, jsonContext)
		} else {
			fmt.Fprintf(MapPlugState[myName].MeasureConn,   logformat, mymeasuretime, myName, measuredata, jsonContext)
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
			Measuretime:	mymeasuretime,
			TimeOverhead: 	(mymeasuretime - float64(t.UnixNano()) / 1e9) * 1e3,
			PState:			*MapPlugState[myName],
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

