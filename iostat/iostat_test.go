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
	"io"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/snap-plugin-collector-iostat/iostat/parser"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

// type Mock struct {
// 	key   string
// 	value float64
// }

// var ns_prefix = []string{parser.NsVendor, parser.NsClass, parser.NsType}

// var mockKV = []Mock{
// 	{"/intel/linux/iostat/avg-cpu/%user", 0.50},
// 	{"/intel/linux/iostat/avg-cpu/%nice", 0.00},
// 	{"/intel/linux/iostat/avg-cpu/%system", 0.13},
// 	{"/intel/linux/iostat/avg-cpu/%iowait", 0.00},
// 	{"/intel/linux/iostat/avg-cpu/%steal", 0.00},
// 	{"/intel/linux/iostat/avg-cpu/%idle", 99.37},
// 	{"/intel/linux/iostat/device/sda/rrqm_per_sec", 0.00},
// 	{"/intel/linux/iostat/device/sdb/wrqm_per_sec", 0.33},
// 	{"/intel/linux/iostat/device/sda1/r_per_sec", 0.00},
// 	{"/intel/linux/iostat/device/sdb1/w_per_sec", 0.08},
// 	{"/intel/linux/iostat/device/sda2/rkB_per_sec", 0.00},
// 	{"/intel/linux/iostat/device/sdb2/wkB_per_sec", 4.55},
// 	{"/intel/linux/iostat/device/sda3/avgrq-sz", 8.00},
// 	{"/intel/linux/iostat/device/sda4/avgqu-sz", 0.00},
// 	{"/intel/linux/iostat/device/sdb/await", 1.83},
// 	{"/intel/linux/iostat/device/sdb/r_await", 0.94},
// 	{"/intel/linux/iostat/device/sdb/w_await", 2.00},
// 	{"/intel/linux/iostat/device/sdb/svctm", 0.06},
// 	{"/intel/linux/iostat/device/sdb/%util", 0.00},
// 	{"/intel/linux/iostat/device/ALL/rrqm_per_sec", 0.05},
// 	{"/intel/linux/iostat/device/ALL/wkB_per_sec", 30.68},
// }

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
		Namespace_: []string{"intel", "linux", "iostat", "device", "sda", "rkB_per_sec"},
	},

	plugin.PluginMetricType{
		Namespace_: []string{"intel", "linux", "iostat", "device", "sdb", "wkB_per_sec"},
	},
}

type mockCmdRunner struct{}

func (c *mockCmdRunner) Run(cmd string, args []string) (io.Reader, error) {
	return strings.NewReader(mockCmdOut), nil
}

//////////////////////////////////////////////////////////////////////////////
//	***						TESTS										***	//
//////////////////////////////////////////////////////////////////////////////

func TestIostat(t *testing.T) {
	iostat := &IOSTAT{parser: parser.New(), cmd: &mockCmdRunner{}}

	Convey("Given invalid metric namespace collect metrics", t, func() {
		badMetrics := []plugin.PluginMetricType{
			plugin.PluginMetricType{
				Namespace_: []string{"intel", "linux", "iostat", "device", "sda", "bad"},
			},
		}
		So(func() { iostat.CollectMetrics(badMetrics) }, ShouldNotPanic)
		result, err := iostat.CollectMetrics(badMetrics)
		So(len(result), ShouldEqual, len(badMetrics))
		So(err, ShouldBeNil)

		for _, r := range result {
			So(r.Namespace(), ShouldBeEmpty)
			So(r.Data(), ShouldBeNil)
			So(r.Source(), ShouldBeEmpty)
		}
	})

	Convey("Given valid metric namespace collect metrics", t, func() {
		So(func() { iostat.CollectMetrics(mockMts) }, ShouldNotPanic)
		result, err := iostat.CollectMetrics(mockMts)
		So(len(result), ShouldEqual, len(mockMts))
		So(err, ShouldBeNil)
		So(result[0].Data(), ShouldEqual, 0)
		So(result[1].Data(), ShouldEqual, 0)
		So(result[2].Data(), ShouldEqual, 0)
		So(result[3].Data(), ShouldEqual, 0.01)
		So(result[4].Data(), ShouldEqual, 15.34)
	})

	Convey("Get metric types", t, func() {
		mts, err := iostat.GetMetricTypes(plugin.PluginConfigType{})
		So(err, ShouldBeNil)
		So(len(mts), ShouldEqual, 123)
	})

	Convey("Get config policy", t, func() {
		policy, err := iostat.GetConfigPolicy()
		So(err, ShouldBeNil)
		So(policy, ShouldResemble, cpolicy.New())
	})
}
