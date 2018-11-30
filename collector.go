package main

import (
	"sre-agent/types"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
        log "github.com/sirupsen/logrus"
	"net"
	"os"
	"os/user"
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


var myContext    types.Context 
var myModContext types.ModuleContext 
var myNets       types.Nets

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

	// need to get ALL ip addresses from all interfaces
        ifaces, _ := net.Interfaces()
        for _, i := range ifaces {
                if i.Flags&net.FlagLoopback     != 0 { continue }
                if i.Flags&net.FlagPointToPoint != 0 { continue }
                addrs, _ := i.Addrs()
                if len(addrs) == 0 { continue }
		myNet := &types.NetInfo{Net: i.Name, MTU: i.MTU, MAC: fmt.Sprintf("%v",i.HardwareAddr)}
                for  _, addr := range addrs {
                        var ip net.IP
                        switch v := addr.(type) {
                        case *net.IPNet:
                                ip = v.IP
                        case *net.IPAddr:
                                ip = v.IP
                        }
			myNet.IP = append(myNet.IP, ip.String())
                }
		myNets.Items = append(myNets.Items,*myNet)
        }
}
