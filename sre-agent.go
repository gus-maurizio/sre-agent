// Copyright 2018 Gustavo Maurizio
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
// OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.
//

package main

import _ "net/http/pprof"

import (
	"sre-agent/types"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v2"
	//"html"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"plugin"
	"syscall"
	"time"
	"strconv"
)

var PluginSlice []types.PluginRuntime

var p = message.NewPrinter(language.English)

func cleanup() {
	log.Info("Program Cleanup Started")
	for pluginIdx, PluginPtr := range PluginSlice {
		log.Debug(  p.Sprintf("Stopping plugin %3d %20s ticker %#v\n", pluginIdx, PluginPtr.PluginName, PluginPtr.Ticker) )
		PluginPtr.Ticker.Stop()
	}
}

func main() {
	// get the program name and directory where it is loaded from
	// also create a properly formatted (language aware) printer object
	myName    := filepath.Base(os.Args[0])
	myExecDir := filepath.Dir(os.Args[0])

	//--------------------------------------------------------------------------//
	// good practice to initialize what we want and read the command line options
	rand.Seed(time.Now().UTC().UnixNano())

	yamlPtr  := flag.String("f", "./config/agent.yaml", "Agent configuration YAML file")
	debugPtr := flag.Bool("d", false, "Agent debug mode - verbose")
	flag.Parse()

	//--------------------------------------------------------------------------//
	// read the yaml configuration into the Config structure
	config := types.Config{}
	yamlFile, err := ioutil.ReadFile(*yamlPtr)
	if err != nil {
		log.Fatalf("config YAML file Get err  #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// See if LogFormat is needed
	if config.LogDest   == "STDERR" { log.SetOutput(os.Stderr) }
	if config.LogFormat == "JSON"   { log.SetFormatter(&log.JSONFormatter{ DisableTimestamp: config.DisableTimestamp, PrettyPrint: config.PrettyPrint}) }
	if *debugPtr { log.SetLevel(log.DebugLevel) }


	log.Info( p.Sprintf("Program %s [from %s] Started", myName, myExecDir) )
	log.Debug( p.Sprintf("config: %+v\n", config) )

	//--------------------------------------------------------------------------//
	// Complete the Context values with non-changing information (while we are alive!)

        myContext.AccountId      = "000000000000"
        myContext.ApplicationId  = config.ApplicationId
        myContext.ModuleId       = config.ModuleId
        myContext.VersionId      = config.VersionId
        myContext.EnvironmentId  = config.EnvironmentId
        myContext.ComputeId      = "iMac"
        myContext.RegionId       = "US-EAST"
        myContext.ZoneId         = "Reston"
        myContext.RunId          = uuid.New().String()

	// Set the context in the logger as default
	contextLogger := log.WithFields(log.Fields{"name": myName, "context": myContext})
        contextLogger.WithFields(log.Fields{"staticinfo": myStaticInfo}).Info( "STATIC" )
	//--------------------------------------------------------------------------//
	// time to start a prometheus metrics server
	// and export any metrics on the /metrics endpoint.
	http.Handle(config.MetricHandle, promhttp.Handler())
	// we now add a health function!
	http.HandleFunc(config.HealthHandle, func(w http.ResponseWriter, r *http.Request) {
		//fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
		answer := struct {
				Timestamp	float64
				ContextData	types.Context
				Staticinfo	[]interface{}
			} { 	float64(time.Now().UnixNano())/1e9,
				myContext,
				myStaticInfo,
			}
		jsonAnswer, err := json.MarshalIndent(answer, "", "\t")
		if err != nil { contextLogger.Fatal("Cannot json marshal info. Err %s", err) }
		fmt.Fprintf(w, "%s\n", jsonAnswer)
	})

        // we now add a details function!
        http.HandleFunc(config.DetailHandle, func(w http.ResponseWriter, r *http.Request) {
                //fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
		switch r.URL.Path {
		case config.DetailHandle + "all":
			getDetailInfo()
			myDynamicDetailInfo["timestamp"] = float64(time.Now().UnixNano())/1e9
			myDynamicDetailInfo["context"]   = myContext
			infoAnswer, ierr := json.MarshalIndent(myDynamicDetailInfo, "", "\t") 
			if ierr != nil { contextLogger.Fatal("Cannot json marshal info. Err %s", ierr) }
			fmt.Fprintf(w, "%s\n", infoAnswer)
                case config.DetailHandle + "summary":
                        getInfo()
                        myDynamicInfo["timestamp"] = float64(time.Now().UnixNano())/1e9
                        myDynamicInfo["context"]   = myContext
                        infoAnswer, ierr := json.MarshalIndent(myDynamicInfo, "", "\t")
                        if ierr != nil { contextLogger.Fatal("Cannot json marshal info. Err %s", ierr) }
                        fmt.Fprintf(w, "%s\n", infoAnswer)
		default:	
                        fmt.Fprintf(w, "%s\n", "must specify /all or /summary")
		}
        })

	// Launch the Prometheus server that will answer to the /metrics requests
	go func() {
		contextLogger.WithFields(log.Fields{"prometheusport": config.PrometheusPort, "prometheuspath": config.MetricHandle}).Debug("Beginning metrics")
		contextLogger.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.PrometheusPort), nil))
	}()

	//--------------------------------------------------------------------------//
	// Start the base plugin
	// set the timer
	//pluginMaker(myContext, 200*time.Millisecond, "baseChannelPlugin", basePlugin, baseMeasure)

	// Scan the configuration to load all the plugins
	for i := range config.Plugins {
		contextLogger.WithFields(log.Fields{"plugin_entry": config.Plugins[i]}).Debug("plugin")
		// load the plugin
		plug, lerr := plugin.Open(config.Plugins[i].PluginModule)
		if lerr != nil {
			contextLogger.WithFields(log.Fields{"plugin_entry": config.Plugins[i], "error": lerr}).Fatal("Error loading plugin")
			os.Exit(16)
		}
		// Identify the main needed function exported as symbol PluginMeasure
		pluginMeasure, perr := plug.Lookup("PluginMeasure")
                if perr != nil {
                        contextLogger.WithFields(log.Fields{"plugin_entry": config.Plugins[i], "error": perr}).Fatal("Error loading measure function")
			continue
                }
		// It is possible that the plugin needs a ONE TIME initialization via function exported as symbol InitPlugin
		// and then pass the config parameter pluginconfig, a string that usually is a json element
                pluginInit, ierr := plug.Lookup("InitPlugin")
                if ierr == nil {
                        contextLogger.WithFields(log.Fields{"plugin_entry": config.Plugins[i]}).Info("about to initialize plugin")
			pluginInit.(func(string) ())(config.Plugins[i].PluginConfig)
                }
		// Compute the TICK between measurements by identifying the unit (Millisecond or Second, if not use Minute)
		var plugintick time.Duration
		if config.Plugins[i].PluginUnit == "" { config.Plugins[i].PluginUnit = config.DefaultUnit }
		switch config.Plugins[i].PluginUnit {
		case "Second":
			plugintick = time.Second  
		case "Millisecond":
			plugintick = time.Millisecond  
		default:
			plugintick = time.Minute  
		}
		if config.Plugins[i].PluginTick == 0 {config.Plugins[i].PluginTick = config.DefaultTick}
		// Now we have all the elements to call the pluginMaker and pass the parameters
                contextLogger.WithFields(log.Fields{"plugin_entry": config.Plugins[i]}).Info("about to create the plugin")
		pluginMaker(myContext, time.Duration(config.Plugins[i].PluginTick) * plugintick, config.Plugins[i].PluginName, basePlugin, pluginMeasure.(func() ([]uint8, float64)))
	}

	//--------------------------------------------------------------------------//
	// now get ready to finish if some signals are received
	contextLogger.Debug("Setting signal handlers")
	csignal := make(chan os.Signal, 3)
	signal.Notify(csignal, syscall.SIGINT)
	signal.Notify(csignal, syscall.SIGTERM)
	contextLogger.Debug("Waiting for a signal to end")
	s := <-csignal
	contextLogger.Debug("Got signal:", s)
	cleanup()
	contextLogger.Info("Program Ended")
	os.Exit(4)
}
