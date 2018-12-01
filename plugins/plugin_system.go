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
	"fmt"
	log "github.com/sirupsen/logrus"
	//"github.com/prometheus/client_golang/prometheus"
	"time"
)


func PluginMeasure() ([]byte, float64) {
        timenow := float64(time.Now().UnixNano())/1e9
        return []byte(fmt.Sprintf(`[{"measuretime": %f}]`, timenow)), timenow
}

func InitPlugin(config string) () {
	var myconfig interface{}
	err := json.Unmarshal([]byte(config), &myconfig)
	if err != nil {
		log.WithFields(log.Fields{"config": config}).Error("failed to unmarshal config")
	}
	log.WithFields(log.Fields{"jsonconfig": myconfig}).Info("InitPlugin")
}

func main() {
	// for testing purposes only, can safely not exist!
	arraybyte, timenow := PluginMeasure()
	fmt.Printf("%#v\n%#v\n", timenow, arraybyte)
}
