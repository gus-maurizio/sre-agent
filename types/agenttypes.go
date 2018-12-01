package types

import (
	"time"
)


// Context is designed to store relevant information for observability and
// tracing that will be needed to identify what is going on.
// It basically identify the host this agent will be running on. The RunId
// is a unique identifier generated once and passed to all plugins to be 
// able to report and correlate to the Context. Each plugin at each tick
// should update the ModuleContext RequestId. 

type Context struct {
	UserId        string   `json:"userid"`
	UserUID       string   `json:"uid"`
	AccountId     string   `json:"accountid"`
	ExecuteId     string   `json:"executeid"`
	ApplicationId string   `json:"applicationid"`
	ModuleId      string   `json:"moduleid"`
	VersionId     string   `json:"versionid"`
	EnvironmentId string   `json:"environmentid"`
	ComputeId     string   `json:"computeid"`
	RegionId      string   `json:"regionid"`
	ZoneId        string   `json:"zoneid"`
        RunId         string   `json:"runid"`
}

// Each Plugin will keep the RunId from the agent.
// Once the plugin is created and activated, it will
// create a unique TraceId to identify his data.
// The plugin will loop and each 'tick' will use
// a new RequestId that can be passed down to the measurement
// functions. 

type ModuleContext struct {
        ModuleName    string   `json:"modulename"`
        RunId         string   `json:"runid"`
        TraceId       string   `json:"traceid"`
        RequestId     string   `json:"requestid"`
        ParentId      string   `json:"parentid"`
}

// The information needs to be packed in a simple way.
// The ModuleData type provides the base elements.
// Each plugin will provide a Measure() method that
// will return a json structure. 

type ModuleData struct {
	RunId		string   	`json:"runid"`
	Timestamp	float64  	`json:"timestamp"`
	ModContext	ModuleContext	`json:"modulecontext"`
	Measure		interface{}	`json:"measure"`
	TimeOverhead	float64		`json:"overhead"`
}

// This is what gets loaded from the -f .yaml configuration file
type Config struct {
        ApplicationId    string `yaml:"applicationid"`
        ModuleId         string `yaml:"moduleid"`
        VersionId        string `yaml:"versionid"`
        EnvironmentId    string `yaml:"environmentid"`

        LogFormat        string `yaml:"logformat"`
        LogDest          string `yaml:"logdest"`
	PrettyPrint	 bool	`yaml:"prettyprint"`
	DisableTimestamp bool	`yaml:"disabletimestamp"`

	PrometheusPort   int    `yaml:"prometheusport"`
	MetricHandle     string `yaml:"metrichandle"`
	DetailHandle     string `yaml:"detailhandle"`
	HealthHandle     string `yaml:"healthhandle"`

        DefaultUnit      string `yaml:"defaulttimeunit"`
        DefaultTick      int    `yaml:"defaulttimetick"`

	Plugins []struct {
		PluginName   string `yaml:"pluginname"`
		PluginModule string `yaml:"pluginmodule"`
		PluginUnit   string `yaml:"plugintimeunit"`
		PluginTick   int    `yaml:"plugintimetick"`
	}
}

type PluginRuntime struct {
	Ticker     *time.Ticker
	PluginName string
}

type FuncMeasure func() ([]byte, float64)

type FuncPlugin func(Context, string, *time.Ticker, FuncMeasure)

type FPlugin    func(Context, string, *time.Ticker, FuncMeasure)
