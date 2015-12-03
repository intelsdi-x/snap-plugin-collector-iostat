/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package iostat

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/source"
	"github.com/intelsdi-x/snap/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

type Mock struct {
	key   string
	value float64
}

var ns_prefix = []string{ns_vendor, ns_class, ns_type}

var mockKV = []Mock{
	{"/intel/linux/iostat/avg-cpu/%user", 0.50},
	{"/intel/linux/iostat/avg-cpu/%nice", 0.00},
	{"/intel/linux/iostat/avg-cpu/%system", 0.13},
	{"/intel/linux/iostat/avg-cpu/%iowait", 0.00},
	{"/intel/linux/iostat/avg-cpu/%steal", 0.00},
	{"/intel/linux/iostat/avg-cpu/%idle", 99.37},
	{"/intel/linux/iostat/device/sda/rrqm_per_sec", 0.00},
	{"/intel/linux/iostat/device/sdb/wrqm_per_sec", 0.33},
	{"/intel/linux/iostat/device/sda1/r_per_sec", 0.00},
	{"/intel/linux/iostat/device/sdb1/w_per_sec", 0.08},
	{"/intel/linux/iostat/device/sda2/rkB_per_sec", 0.00},
	{"/intel/linux/iostat/device/sdb2/wkB_per_sec", 4.55},
	{"/intel/linux/iostat/device/sda3/avgrq-sz", 8.00},
	{"/intel/linux/iostat/device/sda4/avgqu-sz", 0.00},
	{"/intel/linux/iostat/device/sdb/await", 1.83},
	{"/intel/linux/iostat/device/sdb/r_await", 0.94},
	{"/intel/linux/iostat/device/sdb/w_await", 2.00},
	{"/intel/linux/iostat/device/sdb/svctm", 0.06},
	{"/intel/linux/iostat/device/sdb/%util", 0.00},
	{"/intel/linux/iostat/device/ALL/rrqm_per_sec", 0.05},
	{"/intel/linux/iostat/device/ALL/wkB_per_sec", 30.68},
}

var mockCmdOut = `Linux 3.10.0-229.11.1.el7.x86_64 (gklab-108-166) 0/26/2015      _x86_64_        (8 CPU)
				
			10/26/2015 06:36:57 AM
			avg-cpu:  %user   %nice %system %iowait  %steal   %idle
			           0.50    0.00    0.13    0.00    0.00   99.37
			
			Device:         rrqm/s   wrqm/s     r/s     w/s    rkB/s    wkB/s avgrq-sz avgqu-sz   await r_await w_await  svctm  %util
			sda               0.00     0.00    0.00    0.00     0.01     0.00     8.06     0.00    0.10    0.10    0.00   0.04   0.00
			sda1              0.00     0.00    0.00    0.00     0.00     0.00     8.19     0.00    0.12    0.12    0.00   0.12   0.00
			sda2              0.00     0.00    0.00    0.00     0.00     0.00     7.80     0.00    0.08    0.08    0.00   0.08   0.00
			sda3              0.00     0.00    0.00    0.00     0.00     0.00     8.00     0.00    0.12    0.12    0.00   0.12   0.00
			sda4              0.00     0.00    0.00    0.00     0.00     0.00     8.00     0.00    0.11    0.11    0.00   0.11   0.00
			sdb               0.02     0.33    0.13    0.64     2.08    15.34    45.70     0.00    1.83    0.94    2.00   0.06   0.00
			sdb1              0.00     0.07    0.04    0.08     0.26    10.79   185.22     0.00    9.81    0.23   14.21   0.25   0.00
			sdb2              0.02     0.26    0.09    0.55     1.81     4.55    19.87     0.00    0.34    1.24    0.20   0.03   0.00
			 ALL              0.05     0.66    0.26    1.27     4.17    30.68    45.65     0.00    1.82    0.92    2.00   0.06   0.00
				
		`

var mockMts = []plugin.PluginMetricType{
	plugin.PluginMetricType{
		Namespace_: []string{"intel", "linux", "iostat", "device", "sda", "%util"},
	},
	plugin.PluginMetricType{
		Namespace_: []string{"intel", "linux", "iostat", "device", "sdb", "%util"},
	},
	plugin.PluginMetricType{
		Namespace_: []string{"intel", "linux", "iostat", "device", "ALL", "%util"},
	},

	plugin.PluginMetricType{
		Namespace_: []string{"intel", "linux", "iostat", "device", "sda", "sda1", "rkB_per_sec"},
	},

	plugin.PluginMetricType{
		Namespace_: []string{"intel", "linux", "iostat", "device", "sda", "sdb1", "w_per_sec"},
	},
}

//////////////////////////////////////////////////////////////////////////////
//	***					MOCKING FUNCTIONS								***	//
// 	***		mocking: exec.Command, os.Getenv, exec.LookPath				***	//
//////////////////////////////////////////////////////////////////////////////

type MockCmdCall struct {
	getEnvOut   struct{ env string }
	lookPathOut struct {
		path string
		err  error
	}
	execCommandOut struct {
		name string
		args []string
	}
}

func (m *MockCmdCall) getEnv(key string) string {
	return m.getEnvOut.env
}

func (m *MockCmdCall) lookPath(file string) (string, error) {
	return m.lookPathOut.path, m.lookPathOut.err
}

func (m *MockCmdCall) execCommand(src source.Source, data chan interface{}, err chan error) {
	mockSrc := source.Source{
		Command: m.execCommandOut.name,
		Args:    m.execCommandOut.args,
	}
	mockSrc.Generate(data, err)
}

//////////////////////////////////////////////////////////////////////////////
//	***						TESTS										***	//
//////////////////////////////////////////////////////////////////////////////

func TestGetMetricTypes(t *testing.T) {
	var conf plugin.PluginConfigType
	iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{},
		cmdCall: &MockCmdCall{}}
	iostat.keys = make([]string, len(mockMts))

	Convey("no metrics available", t, func() {
		So(func() { iostat.GetMetricTypes(conf) }, ShouldNotPanic)
		result, err := iostat.GetMetricTypes(conf)
		So(len(result), ShouldEqual, len(iostat.keys))
		So(err, ShouldBeNil)

		for _, r := range result {
			So(r.Namespace(), ShouldResemble, []string{""})
			So(r.Data(), ShouldBeNil)
			So(r.Source(), ShouldBeEmpty)
		}
	})

	// mocking iostat keys
	for i, m := range mockMts {
		iostat.keys[i] = joinNamespace(m.Namespace())
	}

	Convey("getMetricsTypes returns no error", t, func() {
		//		var conf plugin.PluginConfigType
		So(func() { iostat.GetMetricTypes(conf) }, ShouldNotPanic)
		result, err := iostat.GetMetricTypes(conf)
		So(len(result), ShouldEqual, len(iostat.keys))
		So(err, ShouldBeNil)
		for i, r := range result {
			So(r.Namespace(), ShouldResemble, strings.Split(strings.TrimPrefix(iostat.keys[i], "/"), "/"))
		}
	})
}

func TestCollectMetrics(t *testing.T) {
	iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: &MockCmdCall{}}

	Convey("no iostat metric available", t, func() {
		So(func() { iostat.CollectMetrics(mockMts) }, ShouldNotPanic)
		result, err := iostat.CollectMetrics(mockMts)
		So(len(result), ShouldEqual, len(mockMts))
		So(err, ShouldBeNil)

		for _, r := range result {
			So(r.Namespace(), ShouldBeEmpty)
			So(r.Data(), ShouldBeNil)
			So(r.Source(), ShouldBeEmpty)
		}
	})

	Convey("invalid metric namespace", t, func() {
		iostat.keys = []string{"intel/linux/iostat/device/sda/%wrong"}
		iostat.data[iostat.keys[0]] = 1

		So(func() { iostat.CollectMetrics(mockMts) }, ShouldNotPanic)
		result, err := iostat.CollectMetrics(mockMts)
		So(len(result), ShouldEqual, len(mockMts))
		So(err, ShouldBeNil)

		for _, r := range result {
			So(r.Namespace(), ShouldBeEmpty)
			So(r.Data(), ShouldBeNil)
			So(r.Source(), ShouldBeEmpty)
		}
	})

	Convey("collect metrics with no errors", t, func() {
		tstamp := time.Now()
		hostname, _ := os.Hostname()
		iostat.keys = make([]string, len(mockMts))
		iostat.timestamp = tstamp

		for i, m := range mockMts {
			iostat.keys[i] = joinNamespace(m.Namespace())
			iostat.data[iostat.keys[i]] = float32(i) * 1.0

		}

		So(func() { iostat.CollectMetrics(mockMts) }, ShouldNotPanic)
		result, err := iostat.CollectMetrics(mockMts)
		So(len(result), ShouldEqual, len(mockMts))
		So(err, ShouldBeNil)

		for _, r := range result {
			So(iostat.keys, ShouldContain, joinNamespace(r.Namespace()))
			So(r.Timestamp(), ShouldResemble, tstamp)
			So(r.Source(), ShouldEqual, hostname)
			So(r.Data(), ShouldEqual, iostat.data[joinNamespace(r.Namespace())])
		}
	})
}

func TestInit(t *testing.T) {
	Convey("successful init iostat plugin", t, func() {
		mcc := &MockCmdCall{}
		mcc.getEnvOut.env = "path/to/iostat/bin"
		mcc.execCommandOut.name = "echo"
		mcc.execCommandOut.args = []string{mockCmdOut}

		iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: mcc}
		So(func() { iostat.init() }, ShouldNotPanic)
		result, err := iostat.init()
		So(result, ShouldNotBeNil)
		So(len(result.keys), ShouldEqual, len(result.data))
		So(err, ShouldBeNil)

		for _, m := range mockKV {
			So(result.keys, ShouldContain, m.key)
			So(result.data[m.key], ShouldNotBeNil)
			So(result.data[m.key], ShouldEqual, m.value)
		}
	})
}

func TestInvalidInit(t *testing.T) {
	Convey("executable file not found in $PATH", t, func() {
		mcc := &MockCmdCall{}
		mcc.getEnvOut.env = ""
		mcc.lookPathOut.path = ""
		mcc.lookPathOut.err = errors.New("executable file not found in $PATH")

		iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: mcc}
		So(func() { iostat.init() }, ShouldPanic)
	})

	Convey("invalid command - error starting of execution", t, func() {
		mcc := &MockCmdCall{}
		mcc.lookPathOut.path = "path/to/iostat/bin"
		mcc.lookPathOut.err = nil
		mcc.execCommandOut.name = ""
		mcc.execCommandOut.args = []string{}

		iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: mcc}
		So(func() { iostat.init() }, ShouldNotPanic)
		result, err := iostat.init()
		So(result, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("output is null, err: timeout for waiting for metrics", t, func() {
		mcc := &MockCmdCall{}
		mcc.getEnvOut.env = "path/to/iostat/bin"
		mcc.execCommandOut.name = "echo"
		mcc.execCommandOut.args = []string{""}

		iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: mcc}
		So(func() { iostat.init() }, ShouldNotPanic)
		result, err := iostat.init()
		So(result, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("invalid output, err: timeout for waiting for metrics ", t, func() {
		mcc := &MockCmdCall{}
		mcc.getEnvOut.env = "path/to/iostat/bin"
		mcc.execCommandOut.name = "echo"
		mcc.execCommandOut.args = []string{"a", "b"}

		iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: mcc}
		So(func() { iostat.init() }, ShouldNotPanic)
		result, err := iostat.init()
		So(result, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("empty tokens > defaultEmptyTokenAcceptance, err: timeout for waiting for metrics", t, func() {
		mcc := &MockCmdCall{}
		mcc.getEnvOut.env = "path/to/iostat/bin"
		mcc.execCommandOut.name = "echo"

		for i := 0; i <= defaultEmptyTokenAcceptance; i++ {
			mcc.execCommandOut.args = append(mcc.execCommandOut.args, fmt.Sprint("\n"))
		}

		iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: mcc}
		So(func() { iostat.init() }, ShouldNotPanic)
		result, err := iostat.init()
		So(result, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("incorrect last section, err: timeout for waiting for metrics", t, func() {
		mcc := &MockCmdCall{}
		mcc.getEnvOut.env = "path/to/iostat/bin"
		mcc.execCommandOut.name = "echo"

		// command output where "ALL" section does not occure
		mcc.execCommandOut.args = []string{strings.Replace(mockCmdOut, "ALL", "", 1)}

		iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: mcc}
		So(func() { iostat.init() }, ShouldNotPanic)
		result, err := iostat.init()
		So(result, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("no statType defined, err: invalid format of iostat output data", t, func() {
		mcc := &MockCmdCall{}
		mcc.getEnvOut.env = "path/to/iostat/bin"
		mcc.execCommandOut.name = "echo"

		//command output where cannot recognize statType (removed ":")
		mcc.execCommandOut.args = []string{strings.Replace(mockCmdOut, "avg-cpu:", "avg-cpu", 1)}

		iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: mcc}
		So(func() { iostat.init() }, ShouldNotPanic)
		result, err := iostat.init()
		So(result, ShouldBeNil)
		So(err, ShouldNotBeNil)

	})

	Convey("less metric's values than expected for avg-cpu stats", t, func() {
		mcc := &MockCmdCall{}
		mcc.getEnvOut.env = "path/to/iostat/bin"
		mcc.execCommandOut.name = "echo"

		// command output where for avg-cpu stats is less values than expected like:
		// avg-cpu:  %user   %nice %system %iowait  %steal   %idle
		//	          0.50    0.13    0.00    0.00   99.37

		mcc.execCommandOut.args = []string{strings.Replace(mockCmdOut, "0.00", "", 1)}
		iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: mcc}
		So(func() { iostat.init() }, ShouldNotPanic)
		result, err := iostat.init()
		So(result, ShouldNotBeEmpty)
		So(err, ShouldBeNil)

		for _, key := range result.keys {
			// invalid format for avg-cpu statistics, avg-cpu metrics were omitted
			So(key, ShouldNotContainSubstring, "avg-cpu")
			So(key, ShouldContainSubstring, "device")
		}

	})

	Convey("more values of metrics than expected for device stats", t, func() {
		mcc := &MockCmdCall{}
		mcc.getEnvOut.env = "path/to/iostat/bin"
		mcc.execCommandOut.name = "echo"

		// command output where for sda1 stats is more values than expected like:
		//Device:         rrqm/s   wrqm/s     r/s     w/s    rkB/s    wkB/s avgrq-sz avgqu-sz   await r_await w_await  svctm  %util
		//	sda1           1.11     0.00     0.00    0.00    0.00     0.00     0.00     8.19     0.00    0.12    0.12    0.00   0.12   0.00

		mcc.execCommandOut.args = []string{strings.Replace(mockCmdOut, "sda1 ", "sda 1.11 ", 1)}
		iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: mcc}
		So(func() { iostat.init() }, ShouldNotPanic)
		result, err := iostat.init()
		So(result, ShouldNotBeEmpty)
		So(err, ShouldBeNil)

		for _, key := range result.keys {
			// invalid format of sda1 stats, device metrics omitted
			So(key, ShouldNotContainSubstring, "device/sda1/")
		}
	})

	Convey("invalid values of metrics, err: strconv.ParseFloat invalid syntax", t, func() {
		mcc := &MockCmdCall{}
		mcc.getEnvOut.env = "path/to/iostat/bin"
		mcc.execCommandOut.name = "echo"

		// command output where for sda stats value is not floating-point number
		mcc.execCommandOut.args = []string{
			`Linux 3.10.0-229.11.1.el7.x86_64 (gklab-108-166) 0/26/2015      _x86_64_        (8 CPU)
				
			10/26/2015 06:36:57 AM
			Device:         rrqm/s   wrqm/s     r/s        
			sda              err   	  n/a    	null    
			ALL              none     nil    	zero   
				
			`,
		}

		iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: mcc}
		So(func() { iostat.init() }, ShouldNotPanic)
		result, err := iostat.init()
		So(result, ShouldNotBeEmpty)
		So(err, ShouldBeNil)

		for _, key := range result.keys {
			So(result.data[key], ShouldBeNil)
		}
	})
}
