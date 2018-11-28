package main

import (
	"agent/types"
	"github.com/prometheus/client_golang/prometheus"
	"net"
	"os"
	"os/user"
)

//Define the metrics we wish to expose
var fooMetric = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "agent_foometric",
	Help: "Shows whether a foo has occurred in our cluster",
})

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


var myContext types.Context 


func init() {
	//Register metrics with prometheus
	prometheus.MustRegister(fooMetric)
	prometheus.MustRegister(messageMetric)
	prometheus.MustRegister(bytesMetric)

	//Set fooMetric to 1
	fooMetric.Set(0)

	//Get all the components needed to populate Context.
	osUser, _ := user.Current()
	myContext.UserId  = osUser.Username 
	myContext.UserUID = osUser.Uid 
	myContext.ExecuteId, _ = os.Hostname()
	myContext.AccountId = "000000000000"
	// need to get ALL ip addresses from all interfaces
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// process IP address
			myContext.IPaddress = append(myContext.IPaddress, ip.String())
		}
	}
}
