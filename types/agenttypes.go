package types

import (
	"time"
)


type NetInfo struct {
	Net   string   `json:"net"`
	MTU   int      `json:"mtu"`
	MAC   string   `json:"mac"`
	IP    []string `json:"ip"`
}

type Nets    struct {
	Items []NetInfo `json:"netifs"`
}

// Context is designed to store relevant information for observability and
// tracing that will be needed to identify what is going on.

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

type ModuleContext struct {
        TraceId       string   `json:"traceid"`
        RequestId     string   `json:"requestid"`
        ParentId      string   `json:"parentid"`
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
	PrometheusHandle string `yaml:"prometheushandle"`

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

type FuncMeasure func() string

type FuncPlugin func(Context, string, *time.Ticker, FuncMeasure)

type FPlugin    func(Context, string, *time.Ticker, FuncMeasure)
