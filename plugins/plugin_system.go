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

import (
        "encoding/json"
	"errors"
        "fmt"
        log "github.com/sirupsen/logrus"
        //"github.com/prometheus/client_golang/prometheus"
        "time"
)


var MyConfig  interface{}
var MyMeasure []byte

func PluginMeasure() ([]byte, float64) {
        timenow := float64(time.Now().UnixNano())/1e9
	MyMeasure = []byte(fmt.Sprintf(`{"measuretime": %f, "myconfig": "%+v"}`, timenow, MyConfig))
        return MyMeasure, timenow
}


func InitPlugin(config string) () {
        err := json.Unmarshal([]byte(config), &MyConfig)
        if err != nil {
                log.WithFields(log.Fields{"config": config}).Error("failed to unmarshal config")
        }
        log.WithFields(log.Fields{"jsonconfig": MyConfig}).Info("InitPlugin")
}


func PluginAlert(measure []byte) (string, string, bool, error) {
        log.WithFields(log.Fields{"MyMeasure": string(MyMeasure[:]), "measure": string(measure[:])}).Info("PluginAlert")
	return "", "", false, errors.New("nothing")
}


func main() {
        // for testing purposes only, can safely not exist!
	config := " { \"alert\": {    \"blue\":   [0,  3], \"green\":  [3,  60], \"yellow\": [60, 80], \"orange\": [80, 90], \"red\":    [90, 100] } } "
	InitPlugin(config)
	log.WithFields(log.Fields{"MyConfig": MyConfig}).Info("InitPlugin")
	tickd := 1* time.Second
	for i := 1; i <= 5; i++ {
		tick := time.Now().UnixNano()
		measure, timestamp := PluginMeasure()
		alertmsg, alertlvl, isAlert, err := PluginAlert(measure)
		fmt.Printf("Iteration #%d tick %d \n", i, tick)
		log.WithFields(log.Fields{"timestamp": timestamp, 
					  "measure": string(measure[:]),
					  "alertmsg": alertmsg,
					  "alertlvl": alertlvl,
					  "isAlert":  isAlert,
					  "err":      err,
		}).Info("Tick")
		time.Sleep(tickd)
	}
}

