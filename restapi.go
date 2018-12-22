package main

import (
	"encoding/json"
	"fmt"
	"github.com/gus-maurizio/sre-agent/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
	//"net"
	"net/http"
	"time"
	"strconv"
)


var myStaticInfo         	[]interface{}
var myDynamicInfo        	map[string]interface{}
var myDynamicDetailInfo  	map[string]interface{}


//Define the metrics we wish to expose

var overheadMetric = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
                Name: "agent_plugin_overhead_ms",
		Help: "Plugin measure overhead in milliseconds",
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


func getInfo() {
	// Get all the static information about this instance

	if myDynamicInfo == nil {
		myDynamicInfo = make(map[string]interface{},20)
	}

	myDynamicInfo["mem"]           , _ = mem.VirtualMemory()
	myDynamicInfo["cputimes"]      , _ = cpu.Times(false)
	myDynamicInfo["cpupercent"]    , _ = cpu.Percent(200 * time.Millisecond, false)
	myDynamicInfo["netcounters"]   , _ = net.IOCounters(false)

}

func getDetailInfo() {
    // Get all the static information about this instance

    if myDynamicDetailInfo == nil {
            myDynamicDetailInfo = make(map[string]interface{},20)
    }

    myDynamicDetailInfo["cputimes"]    , _ 	= cpu.Times(true)
    myDynamicDetailInfo["cpupercent"]  , _ 	= cpu.Percent(200 * time.Millisecond, true)
    myDynamicDetailInfo["users"]        , _ = host.Users()
    myDynamicDetailInfo["netcounters"] , _ 	= net.IOCounters(true)
    myDynamicDetailInfo["netconn"], _ 		= net.Connections("all")

    f, _ := disk.Partitions(true)
    for _, part := range f { myDynamicDetailInfo[part.Device], _ = disk.Usage(part.Mountpoint) }
    p, _ := process.Processes()
    for _, proc := range p {
            q, _ := proc.Connections()
            if len(q) == 0 {continue}
            myDynamicDetailInfo["proc_" + strconv.Itoa(int(proc.Pid))] = proc
            myDynamicDetailInfo["proc_" + strconv.Itoa(int(proc.Pid)) + "_connections"] = q
    }
}

func setupRestAPIs() {
	// Register metrics with prometheus
	prometheus.MustRegister(overheadMetric)
	prometheus.MustRegister(messageMetric)
	prometheus.MustRegister(bytesMetric)

	//--------------------------------------------------------------------------//
	// time to start a prometheus metrics server
	// and export any metrics on the /metrics endpoint.
	http.Handle(Config.MetricHandle, promhttp.Handler())

	// REST API: health (HealthHandle)
	http.HandleFunc(Config.HealthHandle, func(w http.ResponseWriter, r *http.Request) {
		//fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
		tnow 	:= time.Now()
		// Get all the static information about this instance
		s1, _ 			:= host.Info()
		s2, _ 			:= net.Interfaces()	
		s3, _ 			:= disk.Partitions(true)
		s4, _ 			:= cpu.Info()
		myStaticInfo 	= append(myStaticInfo, s1, s2, s3, s4)
		answer 	:= struct {
				TimeNow		string
				TimeUnix	float64
				ContextData	types.Context
				Staticinfo	[]interface{}
			} { fmt.Sprintf("%s",tnow),
				float64(tnow.UnixNano())/1e9,
				myContext,
				myStaticInfo,
			}
		jsonAnswer, err := json.MarshalIndent(answer, "", "\t")
		if err != nil { contextLogger.Fatal("Cannot json marshal info. Err %s", err) }
		fmt.Fprintf(w, "%s\n", jsonAnswer)
	})

    // REST API: detail (DetailHandle)
    http.HandleFunc(Config.DetailHandle, func(w http.ResponseWriter, r *http.Request) {
		tnow 	:= time.Now()
    	switch r.URL.Path {
		// detail/all
		case Config.DetailHandle + "all":
			getDetailInfo()
			myDynamicDetailInfo["timenow"]  = fmt.Sprintf("%s",tnow)
			myDynamicDetailInfo["timeunix"] = float64(tnow.UnixNano())/1e9
			myDynamicDetailInfo["context"]   = myContext
			// infoAnswer, ierr := json.MarshalIndent(myDynamicDetailInfo, "", "\t") 
			infoAnswer, ierr := json.Marshal(myDynamicDetailInfo) 
			if ierr != nil { contextLogger.Fatal("Cannot json marshal info. Err %s", ierr) }
			fmt.Fprintf(w, "%s\n", infoAnswer)

		// detail/state
        case Config.DetailHandle + "state":	
            infoAnswer, serr := json.MarshalIndent(MapPlugState, "", "\t")
            if serr != nil { contextLogger.Fatal("Cannot json marshal info. Err %s", serr) }
            fmt.Fprintf(w, "%s\n", infoAnswer)

		// detail/history
        case Config.DetailHandle + "history":	
            infoHistory, herr := json.Marshal(MapHistory)
            if herr != nil { contextLogger.Fatal("Cannot json marshal info. Err %s", herr) }
            fmt.Fprintf(w, "%s\n", infoHistory)

		// detail/plugin
        case Config.DetailHandle + "plugin":	
        	for pname, phistory := range(MapHistory) {
				fmt.Fprintf(w, "### %s:\n", pname)
				phistory.Metric.Do( func(m interface{}) { fmt.Fprintf(w,"%+v\n",m) })
				fmt.Fprintf(w, "\n")
        	}
		// detail/summary
        case Config.DetailHandle + "summary":
            getInfo()
  			myDynamicInfo["timenow"]  = fmt.Sprintf("%s",tnow)
			myDynamicInfo["timeunix"] = float64(tnow.UnixNano())/1e9
            myDynamicInfo["context"]   = myContext
            infoAnswer, ierr := json.MarshalIndent(myDynamicInfo, "", "\t")
            if ierr != nil { contextLogger.Fatal("Cannot json marshal info. Err %s", ierr) }
            fmt.Fprintf(w, "%s\n", infoAnswer)
		default:	
			fmt.Fprintf(w, "%s\n", "must specify /all /state /summary /history /plugin")
		}
	})

	// Launch the Prometheus server that will answer to the /metrics requests
	go func() {
		contextLogger.WithFields(log.Fields{"prometheusport": Config.PrometheusPort, "prometheuspath": Config.MetricHandle}).Debug("Beginning metrics")
		contextLogger.Fatal(http.ListenAndServe(":"+strconv.Itoa(Config.PrometheusPort), nil))
	}()

}