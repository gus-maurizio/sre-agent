package types

import (
	"net"
	"os"
	"plugin"
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
	RunId		string   		`json:"runid"`
	Timestamp	float64  		`json:"timestamp"`
	ModContext	ModuleContext	`json:"modulecontext"`
	Measure		interface{}		`json:"measure"`
	TimeOverhead	float64		`json:"overhead"`
}

// This is what gets loaded from the -f .yaml configuration file
type Config struct {
    ApplicationId    string 	`yaml:"applicationid"`
    ModuleId         string 	`yaml:"moduleid"`
    VersionId        string 	`yaml:"versionid"`
    EnvironmentId    string 	`yaml:"environmentid"`

    LogFormat        string 	`yaml:"logformat"`
    LogDest          string 	`yaml:"logdest"`
	PrettyPrint	 bool			`yaml:"prettyprint"`
	DisableTimestamp bool		`yaml:"disabletimestamp"`

	PrometheusPort   int    	`yaml:"prometheusport"`
	MetricHandle     string 	`yaml:"metrichandle"`
	DetailHandle     string 	`yaml:"detailhandle"`
	HealthHandle     string 	`yaml:"healthhandle"`

    DefaultTick      string 	`yaml:"defaulttimetick"`
    DefMeasureDest   []string 	`yaml:"defaultmeasuredest"`
    DefAlertDest     []string 	`yaml:"defaultalertdest"`
    DefWarnDest      []string 	`yaml:"defaultwarndest"`

	DefaultRollW1	string 		`yaml:"defaultrollingwindow1"`
	DefaultRollW2	string 		`yaml:"defaultrollingwindow2"`

	DefaultErrT1	int			`yaml:"defaulterrorthresh1"`
	DefaultWarnT1	int			`yaml:"defaultwarnthresh1"`
    DefaultErrT2    int     	`yaml:"defaulterrorthresh2"`
    DefaultWarnT2   int     	`yaml:"defaultwarnthresh12`
	Plugins []struct {
		PluginName   string 	`yaml:"pluginname"`
		PluginModule string 	`yaml:"pluginmodule"`
		MeasureDest  []string 	`yaml:"measuredest"`
		AlertDest    []string 	`yaml:"alertdest"`
		WarnDest     []string 	`yaml:"warndest"`
		PluginTick   string 	`yaml:"plugintimetick"`
		PluginRollW1 string 	`yaml:"plugintrollingwindow1"`
		PluginRollW2 string 	`yaml:"plugintrollingwindow2"`
    	PluginErrT1  int    	`yaml:"pluginerrorthresh1"`
    	PluginWarnT1 int    	`yaml:"pluginwarnthresh1"`
        PluginErrT2  int    	`yaml:"pluginerrorthresh2"`
        PluginWarnT2 int    	`yaml:"pluginwarnthresh2"`
		PluginConfig string 	`yaml:"pluginconfig"`
	}
}

type PluginRuntime struct {
	Ticker     		*time.Ticker
	PluginName 		string
	PState 			*PluginState
}

type PluginState struct {
	AlertFunction	bool		`json:"alertfunction"`
	AlertMsg		string		`json:"alertmsg"`
	AlertLvl		string		`json:"alertlvl"`
	AlertError		string		`json:"alerterror"`

	MeasureCount	int			`json:"measurecount"`
	MeasureFile 	bool		`json:"measurefile"`
	MeasureHandle	*os.File 	`json:"-"`
	MeasureConn		net.Conn 	`json:"-"`

	Alert			bool		`json:"alert"`
	AlertCount		int        	`json:"alertcount"`
	AlertFile       bool		`json:"alertfile"`
	AlertHandle     *os.File	`json:"-"`
	AlertConn		net.Conn 	`json:"-"`

    Warning         bool		`json:"warning"`
    WarnCount       int			`json:"warncount"`
    WarnFile        bool 		`json:"warnfile"`
    WarnHandle      *os.File	`json:"-"`
	WarnConn		net.Conn 	`json:"-"`

	RollW1count		int			`json:"rollw1"`
	RollW2count		int			`json:"rollw2"`
	W1Alerts		int 		`json:"w1alerts"`
	W1Warns			int 		`json:"w1warns"`
	W2Alerts		int 		`json:"w2alerts"`
	W2Warns			int 		`json:"w2warns"`

	PConfig 		plugin.Symbol	`json:"pluginconfig"`
	PData 			plugin.Symbol	`json:"plugindata"`

	PluginAlert	func([]byte) (string, string, bool, error)	`json:"-"`
}

type FuncMeasure func() ([]byte, []byte, float64)

type FuncPlugin func(Context, string, *time.Ticker, FuncMeasure)

type FPlugin    func(Context, string, *time.Ticker, FuncMeasure)
