package main

import (
	"encoding/json"
	"fmt"
	"github.com/gus-maurizio/sreagent/types"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"time"
)

func cleanup() {
	log.Info("Program Cleanup Started")
	jsonContext, _ 	:= json.Marshal(myContext)
	ts := float64(time.Now().UnixNano())/1e9
	for pluginIdx, PluginPtr := range(MapRuntime) { 
		log.Info(  fmt.Sprintf("Stopping plugin %20s %20s ticker %#v\n", pluginIdx, PluginPtr.PluginName, PluginPtr.Ticker) )
		PluginPtr.Ticker.Stop()
		logformat := "{\"timestamp\": %f, \"plugin\": \"%s\", \"measure\": %s, \"context\": %s}\n"
		measuredata := "plugin stopped"
		if MapPlugState[pluginIdx].MeasureFile {
			fmt.Fprintf(MapPlugState[pluginIdx].MeasureHandle, logformat, ts, pluginIdx, measuredata, jsonContext)
		} else {
			fmt.Fprintf(MapPlugState[pluginIdx].MeasureConn,   logformat, ts, pluginIdx, measuredata, jsonContext)
		}
	}
}


func processConfig() {
	// See if LogFormat is needed
	if Config.LogDest   == "STDERR" { log.SetOutput(os.Stderr) }
	if Config.LogFormat == "JSON"   { log.SetFormatter(&log.JSONFormatter{ DisableTimestamp: Config.DisableTimestamp, PrettyPrint: Config.PrettyPrint}) }



	log.Info( fmt.Sprintf("Program %s [from %s] Started", myName, myExecDir) )
	log.Info( fmt.Sprintf("config: %+v\n", Config) )

	//--------------------------------------------------------------------------//
	// Complete the Context values with non-changing information

    myContext.AccountId      = "000000000000"
    myContext.ApplicationId  = Config.ApplicationId
    myContext.ModuleId       = Config.ModuleId
    myContext.VersionId      = Config.VersionId
    myContext.EnvironmentId  = Config.EnvironmentId
    myContext.ComputeId      = "iMac"
    myContext.RegionId       = "US-EAST"
    myContext.ZoneId         = "Reston"
    myContext.RunId          = uuid.New().String()

	// Set the context in the logger as default
	contextLogger = log.WithFields(log.Fields{"name": myName, "context": myContext})
    contextLogger.WithFields(log.Fields{"staticinfo": myStaticInfo}).Info( "STATIC" )

	// Scan the configuration to load all the plugins
	for i := range Config.Plugins {
		// initialize the state machine
		var mConn,nConn,oConn,pConn net.Conn
		var fConn,gConn,hConn,iConn *os.File
		var err 					error

		contextLogger.WithFields(log.Fields{"plugin_entry": Config.Plugins[i]}).Debug("plugin")
		//--- Ensure Dests are defined at plugin level or get them to be the default ones
		if len(Config.Plugins[i].MeasureDest) != 2 { Config.Plugins[i].MeasureDest = Config.DefMeasureDest }
		if len(Config.Plugins[i].AlertDest)   != 2 { Config.Plugins[i].AlertDest   = Config.DefAlertDest }
		if len(Config.Plugins[i].WarnDest)    != 2 { Config.Plugins[i].WarnDest    = Config.DefWarnDest }
		if len(Config.Plugins[i].PageDest)    != 2 { Config.Plugins[i].PageDest    = Config.DefPageDest }

		//--- define files for destinations
		fConn, gConn, hConn, iConn = nil, nil, nil, nil

		if Config.Plugins[i].MeasureDest[0] == "file" {
			fConn, err	= os.OpenFile(Config.Plugins[i].MeasureDest[1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		} else {
			mConn, err 	= net.Dial(Config.Plugins[i].MeasureDest[0], Config.Plugins[i].MeasureDest[1])	
		}
        if err != nil {
			contextLogger.WithFields(log.Fields{"plugin_entry": Config.Plugins[i], "error": err}).Fatal("Error dialing measurement function destination")
			os.Exit(16)
        }
        //--- Alert Dest
        if Config.Plugins[i].AlertDest[0] == "file" {
			gConn, err      = os.OpenFile(Config.Plugins[i].AlertDest[1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        } else {
			nConn, err      = net.Dial(Config.Plugins[i].AlertDest[0], Config.Plugins[i].AlertDest[1])
        }
        if err != nil {
			contextLogger.WithFields(log.Fields{"plugin_entry": Config.Plugins[i], "error": err}).Fatal("Error dialing alert function destination")
			os.Exit(16)
        }
        //--- Warning Dest
        if Config.Plugins[i].WarnDest[0] == "file" {
			hConn, err      = os.OpenFile(Config.Plugins[i].WarnDest[1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        } else {
			oConn, err      = net.Dial(Config.Plugins[i].WarnDest[0], Config.Plugins[i].WarnDest[1])
        }
        if err != nil {
			contextLogger.WithFields(log.Fields{"plugin_entry": Config.Plugins[i], "error": err}).Fatal("Error dialing warning function destination")
			os.Exit(16)
		}	

        //--- Page Dest - This happens when Thresholds are exceeded and is BAD
        if Config.Plugins[i].PageDest[0] == "file" {
			iConn, err      = os.OpenFile(Config.Plugins[i].PageDest[1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        } else {
			pConn, err      = net.Dial(Config.Plugins[i].PageDest[0], Config.Plugins[i].PageDest[1])
        }
        if err != nil {
			contextLogger.WithFields(log.Fields{"plugin_entry": Config.Plugins[i], "error": err}).Fatal("Error dialing pagingfunction destination")
			os.Exit(16)
		}	
		//--------------------------------------------------------------------------//
		// Compute the TICK between measurements
		if Config.Plugins[i].PluginTick == "" { Config.Plugins[i].PluginTick = Config.DefaultTick }
		plugintick, err := time.ParseDuration(Config.Plugins[i].PluginTick)
		if err != nil { plugintick, _ = time.ParseDuration(Config.DefaultTick) }

		//--- Ensure rolling windows are defined or get them to be the default ones
		if len(Config.Plugins[i].PluginRollW) == 0 { Config.Plugins[i].PluginRollW = Config.DefaultRollW }
		if len(Config.Plugins[i].PluginErrT)  == 0 { Config.Plugins[i].PluginErrT  = Config.DefaultErrT  }
		if len(Config.Plugins[i].PluginWarnT) == 0 { Config.Plugins[i].PluginWarnT = Config.DefaultWarnT }

		wRcount 	:=	make([]int, len(Config.Plugins[i].PluginRollW))
		wAcount 	:=	make([]int, len(Config.Plugins[i].PluginRollW))
		wWcount 	:=	make([]int, len(Config.Plugins[i].PluginRollW))

		// we need to initialize and compute each rolling window in number of ticks for the plugin
		// also set the count of alerts and warns to 0
		// find the largest number of ticks across all windows to set the RollW circular list max size
		rollWsize := 0
		for winIdx, wLength :=  range(Config.Plugins[i].PluginRollW) {
			wDuration, _ 	:=  time.ParseDuration(wLength)
			wRcount[winIdx]  =  int(wDuration / plugintick)
			wAcount[winIdx]  =  0
			wWcount[winIdx]  =  0
			if rollWsize < wRcount[winIdx] { rollWsize = wRcount[winIdx] }
		}
		//--------------------------------------------------------------------------//
		// Initialize the Plugin State
		MapPlugState[Config.Plugins[i].PluginName]	= &types.PluginState{	
			Alert:			false,
			AlertFunction:	false,
			MeasureCount:	0,
			MeasureFile:    Config.Plugins[i].MeasureDest[0] == "file",
			MeasureConn:	mConn,
			MeasureHandle:	fConn,
			AlertCount:		0,
            AlertFile:    	Config.Plugins[i].AlertDest[0] == "file",
            AlertConn:    	nConn,
            AlertHandle:  	gConn,

            WarnCount:      0,
            WarnFile:       Config.Plugins[i].WarnDest[0] == "file",
            WarnConn:       oConn,
            WarnHandle:     hConn,

            PageCount:      0,
            PageFile:       Config.Plugins[i].PageDest[0] == "file",
            PageConn:       pConn,
            PageHandle:     iConn,

            RollWcount: 	wRcount,
            WAlerts:		wAcount,
            WWarns:			wWcount,

            TAlerts:		Config.Plugins[i].PluginErrT,
            TWarns:			Config.Plugins[i].PluginWarnT,

            PConfig:		nil,
            PData:			nil,
			PluginAlert:	nil,
       	}
       	//--------------------------------------------------------------------------//
		MapHistory[Config.Plugins[i].PluginName] = &types.PluginHistory{}
		// initialize circular buffers, they will hold an exact number of elements each
		MapHistory[Config.Plugins[i].PluginName].Metric.Init(Config.MetricHistory, nil)
		MapHistory[Config.Plugins[i].PluginName].RollW.Init(rollWsize, uint8(0))
		//--------------------------------------------------------------------------//
		MapRuntime[Config.Plugins[i].PluginName] = &types.PluginRuntime{
			Ticker: 		time.NewTicker(plugintick), 
			PluginName: 	Config.Plugins[i].PluginName,
			PState:			MapPlugState[Config.Plugins[i].PluginName],
		}
	}
	// --- end --- //
}


