package main

import (
	"sre-agent/types"
	//"fmt"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/net"
	//"github.com/shirou/gopsutil/process"

	log "github.com/sirupsen/logrus"
	"os"
	"os/user"
	"time"
)

//Define the metrics we wish to expose

var overheadMetric = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
                Name: "agent_plugin_overhead",
		Help: "Plugin measure overhead in microseconds",
        },
        []string{"plugin"},
)

var messageMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "agent_plugin_ticks",
		Help: "Number of times plugin has executed.",
	},
	[]string{"plugin"},
)

var bytesMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "agent_bytes_sent",
		Help: "Number of bytes plugin has generated.",
	},
	[]string{"plugin"},
)


var myContext     types.Context 
var myModContext  types.ModuleContext 
var myStaticInfo  []interface{}
var myDynamicInfo map[string]interface{}

func init() {
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

	// Register metrics with prometheus
	prometheus.MustRegister(overheadMetric)
	prometheus.MustRegister(messageMetric)
	prometheus.MustRegister(bytesMetric)


	//Get all the components needed to populate Context.
	osUser, _ := user.Current()
	myContext.UserId  = osUser.Username 
	myContext.UserUID = osUser.Uid 
	myContext.ExecuteId, _ = os.Hostname()
	myContext.AccountId = "000000000000"

	// Get all the static information about this instance
	cpu.Percent(0, true)	// this will initialize for future calls!
	s1, _ := host.Info()
	s2, _ := net.Interfaces()	
	s3, _ := disk.Partitions(true)
	s4, _ := cpu.Info()
	myStaticInfo = append(myStaticInfo, s1, s2, s3, s4)
}

func getInfo() {
        // Get all the static information about this instance

        if myDynamicInfo == nil {
                myDynamicInfo = make(map[string]interface{},20)
        }

	myDynamicInfo["mem"]           , _ = mem.VirtualMemory()
	myDynamicInfo["cputimes"]      , _ = cpu.Times(false)
	myDynamicInfo["cputimes_i"]    , _ = cpu.Times(true)
	myDynamicInfo["cpupercent"]    , _ = cpu.Percent(10 * time.Millisecond, false)
	myDynamicInfo["cpupercent_i"]  , _ = cpu.Percent(10 * time.Millisecond, true)

}
