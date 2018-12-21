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
	"flag"
	"github.com/gus-maurizio/sreagent/types"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"os/user"
	"os/signal"
	"path/filepath"
	"plugin"
	"syscall"
	"time"
)

var MapRuntime  			map[string]*types.PluginRuntime
var MapPlugState   			map[string]*types.PluginState
var MapHistory				map[string]*types.PluginHistory

var Config 					types.Config = types.Config{}

var myContext            	types.Context 

var myModContext         	types.ModuleContext 
var myName					string
var myExecDir				string
var contextLogger			*log.Entry

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
    // Setup logging
    //log.SetFormatter(&log.JSONFormatter{})
    log.SetOutput(os.Stdout)
    log.SetLevel(log.InfoLevel)
    log.SetFormatter(&log.TextFormatter{
            DisableColors: false,
            FullTimestamp: true,
            })
    // This can be removed if CPU overhead is too high
    //log.SetReportCaller(true)

	//Get all the components needed to populate Context.
	osUser, _ 				:= user.Current()
	myContext.UserId  		= osUser.Username 
	myContext.UserUID 		= osUser.Uid 
	myContext.ExecuteId, _ 	= os.Hostname()
	myContext.AccountId 	= "000000000000"

}

func main() {
	// get the program name and directory where it is loaded from
	myName    = filepath.Base(os.Args[0])
	myExecDir = filepath.Dir(os.Args[0])

	//--------------------------------------------------------------------------//
	// Read the command line options and store in Config global variable

	yamlPtr  := flag.String("f", "./config/agent.yaml", "Agent configuration YAML file")
	debugPtr := flag.Bool("d", false, "Agent debug mode - verbose")
	flag.Parse()
	if *debugPtr { log.SetLevel(log.DebugLevel) }
	// read the yaml configuration into the Config structure
	yamlFile, err := ioutil.ReadFile(*yamlPtr)
	if err != nil {
		log.Fatalf("config YAML file Get err  #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &Config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	//--------------------------------------------------------------------------//
	// Create the state machine
	MapPlugState = make(map[string]*types.PluginState,  len(Config.Plugins))
	MapHistory   = make(map[string]*types.PluginHistory,len(Config.Plugins))

	MapRuntime   = make(map[string]*types.PluginRuntime,len(Config.Plugins))
	//--------------------------------------------------------------------------//
	// Process all configuration items and setup REST api and Prometheus Metrics
	log.Info("before config")
	processConfig()
	log.Info("before resApi")
	setupRestAPIs()

	log.Info("after resApi")

	// Load the plugins:
	// Plugin MUST: 	1) exist, 2) have PluginMeasure exported,
	// 		  OPTIONAL: 3) have PluginConfig and PluginData exported
	// 					4) InitPlugin and PluginAlert functions
	for i := range Config.Plugins {
		plug, lerr := plugin.Open(Config.Plugins[i].PluginModule)
		if lerr != nil {
			contextLogger.WithFields(log.Fields{"plugin_entry": Config.Plugins[i], "error": lerr}).Fatal("Error loading plugin")
			os.Exit(16)
		}
		// Identify the main needed function exported as symbol PluginMeasure
		pluginMeasure, perr := plug.Lookup("PluginMeasure")
        if perr != nil {
			contextLogger.WithFields(log.Fields{"plugin_entry": Config.Plugins[i], "error": perr}).Fatal("Error loading measure function")
			os.Exit(16)
		}
        // Plugin might export PluginConfig and PluginData
        ptrConfig, pcerr := plug.Lookup("PluginConfig")
        if pcerr != nil { ptrConfig = nil }
        ptrData, pcerr   := plug.Lookup("PluginData")
        if pcerr != nil { ptrData = nil }
        pluginInit, ierr := plug.Lookup("InitPlugin")
        if ierr == nil {
			contextLogger.WithFields(log.Fields{"plugin_entry": Config.Plugins[i]}).Info("about to initialize plugin")
			pluginInit.(func(string) ())(Config.Plugins[i].PluginConfig)
        }
        pluginAlert, aerr := plug.Lookup("PluginAlert")
        if aerr == nil {
			contextLogger.Info("There is an Alert defined")
			MapPlugState[Config.Plugins[i].PluginName].AlertFunction = true
			MapPlugState[Config.Plugins[i].PluginName].PluginAlert   = pluginAlert.( func([]byte) (string, string, bool, error) )
        }

		MapPlugState[Config.Plugins[i].PluginName].PConfig 			= ptrConfig
		MapPlugState[Config.Plugins[i].PluginName].PData 			= ptrData

		// Now we have all the elements to call the pluginMaker and pass the parameters
		contextLogger.WithFields(log.Fields{"plugin_entry": Config.Plugins[i]}).Info("about to create the plugin")
		go pluginLauncher(
						Config.Plugins[i].PluginName, 
						myContext, 
						MapRuntime[Config.Plugins[i].PluginName].Ticker, 
						pluginMeasure.(func() ([]uint8, []uint8, float64)),
						)
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
